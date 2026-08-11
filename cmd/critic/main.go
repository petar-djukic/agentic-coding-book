// Copyright (c) 2026 Petar Djukic. All rights reserved.
// SPDX-License-Identifier: MIT

// Command critic runs an LLM critic over a drafted chapter and reports
// findings against the book's constitutions and the chapter's SRD.
//
// It reads the chapter markdown, the binding rule sets under
// docs/constitutions/, and the chapter's Section Requirements Document when
// one can be paired, then asks the model for structured findings: a
// constitution rule id, a line anchor, the offending text, and what is wrong.
// It never edits the chapter — the output is a report, and revision is a
// separate human step (docs/constitutions/process.yaml: drafting_pipeline).
//
// The model call is gated on ANTHROPIC_API_KEY. Without the key the command
// reports that it is skipping and exits zero, so a CI run with no credentials
// does not fail the build.
//
//	go run ./cmd/critic 05-how-the-machine-works.md
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"gopkg.in/yaml.v3"
)

const (
	defaultModel  = "claude-opus-5"
	maxTokens     = 16000
	apiKeyEnv     = "ANTHROPIC_API_KEY"
	skipExitCode  = 0
	foundExitCode = 1
)

type finding struct {
	RuleID     string `json:"rule_id"`
	Severity   string `json:"severity"`
	Location   string `json:"location"`
	Quote      string `json:"quote"`
	Issue      string `json:"issue"`
	Suggestion string `json:"suggestion"`
}

type report struct {
	Findings []finding `json:"findings"`
	Summary  string    `json:"summary"`
}

type srd struct {
	Meta struct {
		Chapter string `yaml:"chapter"`
		Title   string `yaml:"title"`
	} `yaml:"meta"`
	raw string
}

func main() {
	docs := flag.String("docs", "docs", "path to the docs directory")
	srdPath := flag.String("srd", "", "SRD to check against; inferred from the chapter title when omitted")
	model := flag.String("model", defaultModel, "model id")
	flag.Parse()

	if err := run(*docs, *srdPath, *model, flag.Args()); err != nil {
		fmt.Fprintf(os.Stderr, "critic: %v\n", err)
		os.Exit(foundExitCode)
	}
}

func run(docs, srdPath, model string, args []string) error {
	chapters, err := selectChapters(args)
	if err != nil {
		return err
	}
	if len(chapters) == 0 {
		fmt.Println("no drafted chapters found: nothing to critique")
		os.Exit(skipExitCode)
	}

	if os.Getenv(apiKeyEnv) == "" {
		fmt.Printf("%s is not set: skipping the critic for %d chapter(s)\n", apiKeyEnv, len(chapters))
		os.Exit(skipExitCode)
	}

	rules, err := loadConstitutions(filepath.Join(docs, "constitutions"))
	if err != nil {
		return err
	}
	srds, err := loadSRDs(filepath.Join(docs, "srd"))
	if err != nil {
		return err
	}

	client := anthropic.NewClient()
	blocking := 0

	for _, chapter := range chapters {
		body, err := os.ReadFile(chapter)
		if err != nil {
			return fmt.Errorf("read %s: %w", chapter, err)
		}
		contract, note := pairSRD(chapter, string(body), srdPath, srds)

		fmt.Printf("\n=== %s ===\n", chapter)
		if note != "" {
			fmt.Printf("%s\n", note)
		}

		rep, err := critique(context.Background(), &client, model, rules, contract, chapter, string(body))
		if err != nil {
			return fmt.Errorf("%s: %w", chapter, err)
		}
		blocking += printReport(rep)
	}

	if blocking > 0 {
		return fmt.Errorf("%d blocking finding(s)", blocking)
	}
	fmt.Println("\nno blocking findings")
	return nil
}

// selectChapters resolves the chapter files to critique. Explicit paths win;
// otherwise every numbered markdown file that is a chapter rather than a part
// divider or front matter is critiqued.
func selectChapters(args []string) ([]string, error) {
	if len(args) > 0 {
		for _, a := range args {
			if _, err := os.Stat(a); err != nil {
				return nil, fmt.Errorf("read %s: %w", a, err)
			}
		}
		return args, nil
	}
	paths, err := filepath.Glob("[0-9][0-9]-*.md")
	if err != nil {
		return nil, err
	}
	sort.Strings(paths)
	var out []string
	for _, p := range paths {
		body, err := os.ReadFile(p)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", p, err)
		}
		if isChapter(string(body)) {
			out = append(out, p)
		}
	}
	return out, nil
}

// isChapter distinguishes a chapter from a part divider or the references
// list by its first heading. Part dividers open with "# Part ...".
func isChapter(body string) bool {
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "# ") {
			continue
		}
		title := strings.TrimSpace(strings.TrimPrefix(line, "# "))
		return !strings.HasPrefix(title, "Part ") && !strings.EqualFold(title, "References")
	}
	return false
}

func loadConstitutions(dir string) (string, error) {
	paths, err := filepath.Glob(filepath.Join(dir, "*.yaml"))
	if err != nil {
		return "", err
	}
	if len(paths) == 0 {
		return "", fmt.Errorf("no constitutions found in %s: the critic checks against them and cannot run without them", dir)
	}
	sort.Strings(paths)
	var b strings.Builder
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			return "", fmt.Errorf("read %s: %w", p, err)
		}
		fmt.Fprintf(&b, "===== %s =====\n%s\n", filepath.Base(p), data)
	}
	return b.String(), nil
}

func loadSRDs(dir string) ([]srd, error) {
	paths, err := filepath.Glob(filepath.Join(dir, "srd-*.yaml"))
	if err != nil {
		return nil, err
	}
	sort.Strings(paths)
	out := make([]srd, 0, len(paths))
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", p, err)
		}
		var s srd
		if err := yaml.Unmarshal(data, &s); err != nil {
			return nil, fmt.Errorf("parse %s: %w", p, err)
		}
		s.raw = string(data)
		out = append(out, s)
	}
	return out, nil
}

// pairSRD resolves which drafting contract a chapter is checked against. An
// explicit -srd wins. Otherwise the chapter's first heading is matched against
// SRD titles; the book binds chapters to SRDs by id rather than by filename,
// so a title match is the only automatic pairing available and a miss is
// reported rather than guessed at.
func pairSRD(path, body, explicit string, srds []srd) (string, string) {
	if explicit != "" {
		data, err := os.ReadFile(explicit)
		if err != nil {
			return "", fmt.Sprintf("could not read %s (%v); checking against the constitutions only", explicit, err)
		}
		return string(data), fmt.Sprintf("contract: %s", explicit)
	}
	title := firstHeading(body)
	for _, s := range srds {
		if normalize(s.Meta.Title) == normalize(title) {
			return s.raw, fmt.Sprintf("contract: SRD %s (%s)", s.Meta.Chapter, s.Meta.Title)
		}
	}
	return "", fmt.Sprintf("no SRD titled %q; checking %s against the constitutions only (pass -srd to pair one)", title, filepath.Base(path))
}

func firstHeading(body string) string {
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "# "))
		}
	}
	return ""
}

func normalize(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		}
	}
	return b.String()
}

const systemPrompt = `You are a critic for a practitioner book on agentic coding.

You are given the book's binding constitutions, optionally the Section
Requirements Document (SRD) that is the chapter's drafting contract, and the
chapter itself with line numbers.

Report findings where the chapter breaches a constitution rule or fails its
SRD. Every finding must name the rule id it breaches (for example V-S1, V-P4,
A-C4, A-X2, VN-6) or, for an SRD breach, the SRD field it fails (for example
content, acceptance, apparatus.key_terms). Anchor every finding to a line
number or a section heading from the numbered chapter text, and quote the
text you are objecting to.

Mark a finding blocking when it breaches a rule the chapter cannot ship
against: a missing required apparatus element, an unsupported load-bearing
claim, a prescription with no Part I fact behind it, or a forbidden term.
Mark it advisory when it is a judgment call or a matter of degree.

Do not rewrite the chapter and do not output revised prose. Report only.
Report nothing you cannot tie to a rule id or an SRD field.`

// reportSchema constrains the response so the report is valid JSON by
// construction rather than by asking the model nicely for it.
var reportSchema = map[string]any{
	"type": "object",
	"properties": map[string]any{
		"findings": map[string]any{
			"type": "array",
			"items": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"rule_id":    map[string]any{"type": "string", "description": "constitution rule id, or the SRD field the chapter fails"},
					"severity":   map[string]any{"type": "string", "enum": []string{"blocking", "advisory"}},
					"location":   map[string]any{"type": "string", "description": "line number or section heading from the numbered chapter"},
					"quote":      map[string]any{"type": "string", "description": "the text being objected to"},
					"issue":      map[string]any{"type": "string", "description": "what is wrong, in one or two sentences"},
					"suggestion": map[string]any{"type": "string", "description": "what would resolve it, without rewriting the prose"},
				},
				"required":             []string{"rule_id", "severity", "location", "quote", "issue", "suggestion"},
				"additionalProperties": false,
			},
		},
		"summary": map[string]any{"type": "string"},
	},
	"required":             []string{"findings", "summary"},
	"additionalProperties": false,
}

func critique(ctx context.Context, client *anthropic.Client, model, rules, contract, path, body string) (report, error) {
	var b strings.Builder
	fmt.Fprintf(&b, "# Constitutions\n\n%s\n", rules)
	if contract != "" {
		fmt.Fprintf(&b, "# Drafting contract (SRD)\n\n%s\n", contract)
	} else {
		b.WriteString("# Drafting contract (SRD)\n\nNone paired. Check against the constitutions only, and do not report SRD findings.\n\n")
	}
	fmt.Fprintf(&b, "# Chapter: %s\n\n%s\n", path, numberLines(body))

	msg, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.Model(model),
		MaxTokens: maxTokens,
		System:    []anthropic.TextBlockParam{{Text: systemPrompt}},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(b.String())),
		},
		OutputConfig: anthropic.OutputConfigParam{
			Effort: anthropic.OutputConfigEffortHigh,
			Format: anthropic.JSONOutputFormatParam{Schema: reportSchema},
		},
	})
	if err != nil {
		return report{}, err
	}
	if msg.StopReason == anthropic.StopReasonMaxTokens {
		return report{}, fmt.Errorf("model output hit the %d token cap before finishing; critique the chapter in sections", maxTokens)
	}

	var text strings.Builder
	for _, block := range msg.Content {
		if t, ok := block.AsAny().(anthropic.TextBlock); ok {
			text.WriteString(t.Text)
		}
	}
	return parseReport(text.String())
}

// parseReport reads the model's JSON report, tolerating a fenced code block
// around it.
func parseReport(s string) (report, error) {
	s = strings.TrimSpace(s)
	if i := strings.Index(s, "{"); i > 0 {
		s = s[i:]
	}
	if j := strings.LastIndex(s, "}"); j >= 0 {
		s = s[:j+1]
	}
	var r report
	if err := json.Unmarshal([]byte(s), &r); err != nil {
		return report{}, fmt.Errorf("model did not return a usable report: %w", err)
	}
	return r, nil
}

func numberLines(body string) string {
	var b strings.Builder
	for i, line := range strings.Split(body, "\n") {
		fmt.Fprintf(&b, "%d\t%s\n", i+1, line)
	}
	return b.String()
}

func printReport(r report) int {
	if len(r.Findings) == 0 {
		fmt.Println("no findings")
		return 0
	}
	blocking := 0
	for _, f := range r.Findings {
		marker := "advisory"
		if strings.EqualFold(f.Severity, "blocking") {
			marker = "BLOCKING"
			blocking++
		}
		fmt.Printf("\n[%s] %s  %s\n", marker, f.RuleID, f.Location)
		if f.Quote != "" {
			fmt.Printf("  > %s\n", collapse(f.Quote))
		}
		fmt.Printf("  %s\n", collapse(f.Issue))
		if f.Suggestion != "" {
			fmt.Printf("  suggestion: %s\n", collapse(f.Suggestion))
		}
	}
	if r.Summary != "" {
		fmt.Printf("\n%s\n", collapse(r.Summary))
	}
	fmt.Printf("\n%d finding(s), %d blocking\n", len(r.Findings), blocking)
	return blocking
}

func collapse(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
