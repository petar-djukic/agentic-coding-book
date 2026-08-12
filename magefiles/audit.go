// Copyright (c) 2026 Petar Djukic. All rights reserved.
// SPDX-License-Identifier: MIT

package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Audit checks that the specification is internally consistent, that the
// drafted prose conforms to the checkable subset of the voice constitution,
// and that the book still builds. Findings accumulate and report together;
// the audit fails once, with everything it found.
//
// Findings recorded in docs/audit-baseline.yaml are accepted debt: they are
// printed but do not fail the audit, and each names the issue that clears it.
// A baseline entry that no longer matches a finding is itself a finding, so
// the file cannot rot into a blanket exemption.
func Audit() error {
	return audit(".", Figures, All)
}

// audit is the testable body. figures renders diagrams before the reference
// check so a fresh clone audits clean; build compiles the book as a
// diagnostic.
func audit(root string, figures, build func() error) error {
	spec, findings := loadSpec(root)
	if spec != nil {
		findings = append(findings, checkSpec(spec)...)
		findings = append(findings, checkProse(root, spec)...)
		findings = append(findings, checkSidebarAuthorship(root, spec)...)
		findings = append(findings, checkSectionNumbering(root)...)
		findings = append(findings, checkChapterBinding(root, spec)...)
	}

	if figures != nil {
		if err := figures(); err != nil {
			findings = append(findings, finding{File: "figures", Rule: "build", Detail: err.Error()})
		}
	}
	findings = append(findings, checkFigureReferences(root)...)
	if build != nil {
		if err := build(); err != nil {
			findings = append(findings, finding{File: "build", Rule: "build", Detail: err.Error()})
		}
	}

	accepted, baselineFindings := loadBaseline(root)
	findings = append(findings, baselineFindings...)
	return reportFindings(findings, accepted)
}

// finding is one audit result: where it is, which rule or document field it
// comes from, and what is wrong. The triple is also its baseline key.
type finding struct {
	File   string
	Rule   string
	Detail string
}

func (f finding) key() string { return f.File + "|" + f.Rule + "|" + f.Detail }

func (f finding) String() string {
	return fmt.Sprintf("%s [%s] %s", f.File, f.Rule, f.Detail)
}

func reportFindings(findings []finding, accepted map[string]string) error {
	sort.Slice(findings, func(i, j int) bool { return findings[i].key() < findings[j].key() })

	var blocking []string
	var known []string
	seen := map[string]bool{}
	matched := map[string]bool{}
	for _, f := range findings {
		if seen[f.key()] {
			continue
		}
		seen[f.key()] = true
		if issue, ok := accepted[f.key()]; ok {
			matched[f.key()] = true
			known = append(known, fmt.Sprintf("%s (accepted, %s)", f, issue))
			continue
		}
		blocking = append(blocking, f.String())
	}
	for _, stale := range staleBaselineFindings(accepted, matched) {
		blocking = append(blocking, stale.String())
	}

	for _, k := range known {
		fmt.Printf("audit: %s\n", k)
	}
	if len(blocking) > 0 {
		return fmt.Errorf("audit failed:\n  - %s", strings.Join(blocking, "\n  - "))
	}
	fmt.Printf("audit: OK (%d accepted finding(s) in docs/audit-baseline.yaml)\n", len(known))
	return nil
}

// ---------------------------------------------------------------- loading

type spec struct {
	outline     outlineDoc
	arch        archDoc
	roadmap     roadmapDoc
	vision      visionDoc
	argument    argumentDoc
	voice       voiceDoc
	srds        []srdDoc
	ruleIDs     map[string]bool
	definitions map[string]bool
	references  map[string]bool
}

type outlineDoc struct {
	Parts []struct {
		ID       string `yaml:"id"`
		Chapters []struct {
			Title string `yaml:"title"`
		} `yaml:"chapters"`
	} `yaml:"parts"`
}

type archDoc struct {
	Structure struct {
		Parts []struct {
			ID       string `yaml:"id"`
			Chapters []struct {
				ID  string `yaml:"id"`
				SRD string `yaml:"srd"`
			} `yaml:"chapters"`
		} `yaml:"parts"`
	} `yaml:"structure"`
}

type roadmapDoc struct {
	Parts []struct {
		ID       string `yaml:"id"`
		Chapters []struct {
			ID     string `yaml:"id"`
			Status string `yaml:"status"`
		} `yaml:"chapters"`
	} `yaml:"parts"`
}

type visionDoc struct {
	Goals []struct {
		ID string `yaml:"id"`
	} `yaml:"goals"`
}

type argumentDoc struct {
	DerivationChain []struct {
		ID     string   `yaml:"id"`
		Owners []string `yaml:"owners"`
	} `yaml:"derivation_chain"`
	PartObligations struct {
		Parts []struct {
			ID string `yaml:"id"`
		} `yaml:"parts"`
	} `yaml:"part_obligations"`
}

type voiceDoc struct {
	ForbiddenTerms struct {
		Terms []string `yaml:"terms"`
	} `yaml:"forbidden_terms"`
	Sidebars struct {
		Types []struct {
			Label       string `yaml:"label"`
			FormerLabel string `yaml:"former_label"`
			Authorship  string `yaml:"authorship"`
		} `yaml:"types"`
	} `yaml:"sidebars"`
}

type srdDoc struct {
	path string
	Meta struct {
		Chapter string `yaml:"chapter"`
		Title   string `yaml:"title"`
	} `yaml:"meta"`
	Goals []struct {
		ID string `yaml:"id"`
	} `yaml:"goals"`
	Chain []struct {
		ID string `yaml:"id"`
	} `yaml:"chain"`
	Constitutions []struct {
		File  string   `yaml:"file"`
		Rules []string `yaml:"rules"`
	} `yaml:"constitutions"`
	Citations []struct {
		ID   string `yaml:"id"`
		Role string `yaml:"role"`
	} `yaml:"citations"`
	Apparatus struct {
		KeyTerms []string `yaml:"key_terms"`
	} `yaml:"apparatus"`
}

func loadSpec(root string) (*spec, []finding) {
	var findings []finding
	s := &spec{
		ruleIDs:     map[string]bool{},
		definitions: map[string]bool{},
		references:  map[string]bool{},
	}
	fail := func(path string, err error) {
		findings = append(findings, finding{File: path, Rule: "spec", Detail: err.Error()})
	}

	docs := filepath.Join(root, "docs")
	for _, load := range []struct {
		path string
		into any
	}{
		{filepath.Join(root, "outline.yaml"), &s.outline},
		{filepath.Join(docs, "ARCHITECTURE.yaml"), &s.arch},
		{filepath.Join(docs, "road-map.yaml"), &s.roadmap},
		{filepath.Join(docs, "VISION.yaml"), &s.vision},
		{filepath.Join(docs, "constitutions", "argument.yaml"), &s.argument},
		{filepath.Join(docs, "constitutions", "voice.yaml"), &s.voice},
	} {
		if err := readYAML(load.path, load.into); err != nil {
			fail(rel(root, load.path), err)
		}
	}

	// Rule ids come from every constitution, whatever its shape: any mapping
	// with an `id` alongside a rule-ish field declares a rule.
	consPaths, _ := filepath.Glob(filepath.Join(docs, "constitutions", "*.yaml"))
	sort.Strings(consPaths)
	if len(consPaths) == 0 {
		fail("docs/constitutions", fmt.Errorf("no constitutions found; the audit checks against them"))
	}
	for _, p := range consPaths {
		var raw any
		if err := readYAML(p, &raw); err != nil {
			fail(rel(root, p), err)
			continue
		}
		collectRuleIDs(raw, s.ruleIDs)
	}

	var defs map[string]struct {
		Definition string `yaml:"definition"`
		Introduced string `yaml:"introduced"`
		Status     string `yaml:"status"`
	}
	if err := readYAML(filepath.Join(docs, "definitions.yaml"), &defs); err != nil {
		fail("docs/definitions.yaml", err)
	}
	for k, v := range defs {
		s.definitions[k] = true
		// Every defined term records the chapter or part that introduces it.
		// That field is what arbitrates where a term belongs -- GH-50 used it to
		// settle every disputed key term -- so an entry without one leaves the
		// question unanswerable. A retired term is exempt: it is recorded so the
		// vocabulary of earlier drafts still resolves, and nothing introduces it.
		if v.Definition != "" && v.Introduced == "" && v.Status != "retired" {
			findings = append(findings, finding{
				File:   "docs/definitions.yaml",
				Rule:   "definitions: introduced",
				Detail: fmt.Sprintf("%s has no introduced field naming the chapter or part that defines it", k),
			})
		}
	}

	for _, id := range referenceIDs(filepath.Join(root, "references.yaml"), fail) {
		s.references[id] = true
	}

	srdPaths, _ := filepath.Glob(filepath.Join(docs, "srd", "srd-*.yaml"))
	sort.Strings(srdPaths)
	for _, p := range srdPaths {
		var d srdDoc
		if err := readYAML(p, &d); err != nil {
			fail(rel(root, p), err)
			continue
		}
		d.path = rel(root, p)
		s.srds = append(s.srds, d)
	}
	return s, findings
}

func readYAML(path string, into any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, into)
}

func rel(root, path string) string {
	if r, err := filepath.Rel(root, path); err == nil {
		return r
	}
	return path
}

// collectRuleIDs walks a decoded constitution and records every id that
// belongs to a rule-bearing mapping.
func collectRuleIDs(node any, into map[string]bool) {
	switch n := node.(type) {
	case map[string]any:
		if id, ok := n["id"].(string); ok {
			for _, k := range []string{"rule", "name", "question", "pattern", "label"} {
				if _, has := n[k]; has {
					into[id] = true
					break
				}
			}
		}
		for _, v := range n {
			collectRuleIDs(v, into)
		}
	case []any:
		for _, v := range n {
			collectRuleIDs(v, into)
		}
	}
}

// referenceIDs reads the CSL-YAML bibliography, which may be a bare list or a
// mapping with a `references` key.
func referenceIDs(path string, fail func(string, error)) []string {
	var raw any
	if err := readYAML(path, &raw); err != nil {
		fail("references.yaml", err)
		return nil
	}
	entries, ok := raw.([]any)
	if !ok {
		if m, isMap := raw.(map[string]any); isMap {
			entries, _ = m["references"].([]any)
		}
	}
	var ids []string
	for _, e := range entries {
		if m, ok := e.(map[string]any); ok {
			if id, ok := m["id"].(string); ok {
				ids = append(ids, id)
			}
		}
	}
	return ids
}

// ------------------------------------------------------------- spec checks

func checkSpec(s *spec) []finding {
	var out []finding
	add := func(file, rule, format string, args ...any) {
		out = append(out, finding{File: file, Rule: rule, Detail: fmt.Sprintf(format, args...)})
	}

	// Chapter ids and their order must agree across the three structural views.
	var archIDs, roadIDs []string
	chapterIDs := map[string]bool{}
	archParts := map[string]bool{}
	for _, p := range s.arch.Structure.Parts {
		archParts[p.ID] = true
		for _, c := range p.Chapters {
			archIDs = append(archIDs, c.ID)
			chapterIDs[c.ID] = true
		}
	}
	for _, p := range s.roadmap.Parts {
		for _, c := range p.Chapters {
			roadIDs = append(roadIDs, c.ID)
		}
	}
	if strings.Join(archIDs, ",") != strings.Join(roadIDs, ",") {
		add("docs/road-map.yaml", "structure", "chapter ids or order differ from docs/ARCHITECTURE.yaml")
	}
	if len(s.outline.Parts) != len(s.arch.Structure.Parts) {
		add("outline.yaml", "structure", "part count %d differs from ARCHITECTURE %d",
			len(s.outline.Parts), len(s.arch.Structure.Parts))
	} else {
		for i, op := range s.outline.Parts {
			ap := s.arch.Structure.Parts[i]
			if len(op.Chapters) != len(ap.Chapters) {
				add("outline.yaml", "structure", "%s has %d chapters, ARCHITECTURE has %d",
					ap.ID, len(op.Chapters), len(ap.Chapters))
			}
		}
	}

	// The derivation chain must be owned by real chapters, and every part must
	// carry an obligation.
	chainOwners := map[string][]string{}
	for _, q := range s.argument.DerivationChain {
		chainOwners[q.ID] = q.Owners
		for _, owner := range q.Owners {
			if !chapterIDs[owner] {
				add("docs/constitutions/argument.yaml", q.ID, "owns unknown chapter %s", owner)
			}
		}
	}
	oblParts := map[string]bool{}
	for _, p := range s.argument.PartObligations.Parts {
		oblParts[p.ID] = true
	}
	for id := range archParts {
		if !oblParts[id] {
			add("docs/constitutions/argument.yaml", "part_obligations", "no obligation for part %s", id)
		}
	}

	visionGoals := map[string]bool{}
	for _, g := range s.vision.Goals {
		visionGoals[g.ID] = true
	}

	seenChapter := map[string]string{}
	for _, d := range s.srds {
		if d.Meta.Chapter == "" {
			add(d.path, "meta.chapter", "empty; every SRD names the chapter it governs")
		} else {
			if prev, dup := seenChapter[d.Meta.Chapter]; dup {
				add(d.path, "meta.chapter", "chapter %s already governed by %s", d.Meta.Chapter, prev)
			}
			seenChapter[d.Meta.Chapter] = d.path
			if !chapterIDs[d.Meta.Chapter] {
				add(d.path, "meta.chapter", "%s is not a chapter in ARCHITECTURE", d.Meta.Chapter)
			}
		}
		if len(d.Constitutions) == 0 {
			add(d.path, "constitutions", "names no rule set; a chapter with no rule set has nothing to fail")
		}
		for _, c := range d.Constitutions {
			for _, r := range c.Rules {
				if !s.ruleIDs[r] {
					add(d.path, "constitutions", "cites unknown rule %s in %s", r, c.File)
				}
			}
		}
		if len(d.Chain) == 0 {
			add(d.path, "chain", "names no derivation-chain link")
		}
		for _, link := range d.Chain {
			owners, known := chainOwners[link.ID]
			if !known {
				add(d.path, "chain", "claims unknown link %s", link.ID)
				continue
			}
			if !contains(owners, d.Meta.Chapter) {
				add(d.path, "chain", "claims %s but argument.yaml owners are %v", link.ID, owners)
			}
		}
		for _, g := range d.Goals {
			base := baseGoal(g.ID)
			if base == "" || !visionGoals[base] {
				add(d.path, "goals", "subgoal %s has no VISION goal %s", g.ID, base)
			}
		}
		for _, c := range d.Citations {
			if !s.references[c.ID] {
				add(d.path, "citations", "cites %s, which does not resolve in references.yaml", c.ID)
			}
		}
		for _, k := range d.Apparatus.KeyTerms {
			if !s.definitions[k] {
				add(d.path, "apparatus.key_terms", "%s is not in docs/definitions.yaml", k)
			}
		}
	}
	return out
}

// baseGoal maps a subgoal id such as G1.4 onto its book goal G1.
func baseGoal(id string) string {
	if !strings.HasPrefix(id, "G") {
		return ""
	}
	major, _, found := strings.Cut(strings.TrimPrefix(id, "G"), ".")
	if !found || major == "" {
		return ""
	}
	return "G" + major
}

func contains(haystack []string, needle string) bool {
	for _, h := range haystack {
		if h == needle {
			return true
		}
	}
	return false
}

// ------------------------------------------------------------ prose checks

var (
	fenceRe    = regexp.MustCompile("(?s)```.*?```")
	citationRe = regexp.MustCompile(`\[@([A-Za-z0-9_.:/-]+)`)
	figureRe   = regexp.MustCompile(`\*\*Figure ([0-9]+\.[0-9]+)\*\*`)
	headingRe  = regexp.MustCompile(`(?m)^#{2,} *(.+?) *$`)
)

// checkProse applies the mechanically checkable subset of voice.yaml to every
// chapter present. A chapter file exists only once it is drafted, so presence
// is the trigger; the apparatus rules do not apply to the introduction, which
// is front matter rather than a chapter (venue.yaml: chapter_arc).
func checkProse(root string, s *spec) []finding {
	var out []finding
	add := func(file, rule, format string, args ...any) {
		out = append(out, finding{File: file, Rule: rule, Detail: fmt.Sprintf(format, args...)})
	}

	labels := map[string]bool{}
	retired := map[string]string{}
	for _, t := range s.voice.Sidebars.Types {
		labels[t.Label] = true
		if t.FormerLabel != "" {
			retired[t.FormerLabel] = t.Label
		}
	}

	for _, path := range chapterFiles(root) {
		data, err := os.ReadFile(filepath.Join(root, path))
		if err != nil {
			add(path, "prose", "%v", err)
			continue
		}
		body := string(data)
		prose := fenceRe.ReplaceAllString(body, "")

		for _, term := range s.voice.ForbiddenTerms.Terms {
			if usesForbiddenTerm(prose, term) {
				add(path, "voice.yaml: forbidden_terms", "uses %q", term)
			}
		}

		for former, now := range retired {
			if strings.Contains(body, former) {
				add(path, "V-B", "uses the retired sidebar label %q; the canonical label is %q", former, now)
			}
		}

		for _, key := range citationRe.FindAllStringSubmatch(body, -1) {
			if !s.references[key[1]] {
				add(path, "V-V3", "cites [@%s], which does not resolve in references.yaml", key[1])
			}
		}

		for _, m := range figureRe.FindAllStringSubmatch(body, -1) {
			if len(figureRe.FindAllString(body, -1)) > 0 &&
				strings.Count(body, "Figure "+m[1]) < 2 {
				add(path, "voice.yaml: figures.checks", "Figure %s is never referenced from the prose", m[1])
			}
		}

		if isIntroduction(body) {
			continue
		}
		headings := map[string]bool{}
		for _, h := range headingRe.FindAllStringSubmatch(body, -1) {
			headings[strings.ToLower(strings.TrimSpace(h[1]))] = true
		}
		if !headings["learning objectives"] {
			add(path, "V-S1", "no Learning Objectives section")
		}
		if !headings["summary"] {
			add(path, "V-S6", "no Summary section")
		}
	}
	return out
}

// chapterMarkerRe matches the binding a chapter file carries to name itself.
// An HTML comment rather than YAML front matter: pandoc merges metadata across
// every concatenated input, so a `chapter:` key would share a namespace with
// the book's real metadata, while a comment renders to nothing in any output
// format (GH-25).
var chapterMarkerRe = regexp.MustCompile(`(?m)^<!--\s*chapter:\s*(\S+)\s*-->`)

// keyTermsRe isolates the Key Terms table, which is where apparatus.key_terms
// promises the terms will be. keyTermRowRe then picks out the bolded term from
// each row, and parenGlossRe strips a trailing gloss like "(LLM)".
//
// Matching is against the term column rather than the section text, because a
// term mentioned inside another term's definition cell is not a term the table
// defines -- that leniency hid three real gaps until GH-50.
var (
	keyTermsRe   = regexp.MustCompile(`(?ms)^## Key Terms\s*$(.*)`)
	keyTermRowRe = regexp.MustCompile(`(?m)^\|\s*\*\*(.+?)\*\*`)
	parenGlossRe = regexp.MustCompile(`\s*\([^)]*\)\s*$`)
)

// termKeyRe collapses the separators a term may be written with. A glossary key
// is `file_system_access`, a Key Terms row is "File-system access", and prose
// may hyphenate either way; all three name one term (GH-52).
var termKeyRe = regexp.MustCompile(`[-_\s]+`)

// normalizeTerm renders a term in one comparable form.
func normalizeTerm(s string) string {
	return strings.TrimSpace(termKeyRe.ReplaceAllString(strings.ToLower(s), " "))
}

// tableDefines reports whether the Key Terms table has a row defining term.
func tableDefines(rows []string, term string) bool {
	want := normalizeTerm(term)
	for _, r := range rows {
		if r == want {
			return true
		}
	}
	return false
}

// checkChapterBinding resolves each drafted chapter to the SRD that governs it
// and checks the mechanically checkable half of that contract.
//
// Chapters name themselves; nothing infers the pairing from a title or a
// filename. Title matching was the old mechanism and it broke silently when a
// chapter was retitled, while a filename map in the spec goes stale every time
// the book is renumbered -- which happened six times while Part I was drafted
// (GH-25).
func checkChapterBinding(root string, s *spec) []finding {
	var out []finding
	add := func(file, rule, format string, args ...any) {
		out = append(out, finding{File: file, Rule: rule, Detail: fmt.Sprintf(format, args...)})
	}

	srdByChapter := map[string]string{}
	archChapters := map[string]bool{}
	for _, p := range s.arch.Structure.Parts {
		for _, c := range p.Chapters {
			archChapters[c.ID] = true
			if c.SRD != "" {
				srdByChapter[c.ID] = c.SRD
			}
		}
	}
	defined := map[string]bool{}
	for k := range s.definitions {
		defined[normalizeTerm(k)] = true
	}

	srdByPath := map[string]*srdDoc{}
	for i := range s.srds {
		srdByPath[rel(root, s.srds[i].path)] = &s.srds[i]
	}

	fileOf := map[string]string{}
	for _, path := range chapterFiles(root) {
		data, err := os.ReadFile(filepath.Join(root, path))
		if err != nil {
			continue
		}
		body := string(data)
		if isIntroduction(body) {
			continue
		}
		m := chapterMarkerRe.FindStringSubmatch(body)
		if m == nil {
			add(path, "GH-25 binding", "carries no `<!-- chapter: ID -->` marker, so no SRD governs it")
			continue
		}
		id := m[1]
		if prev, dup := fileOf[id]; dup {
			add(path, "GH-25 binding", "claims chapter %s, already claimed by %s", id, prev)
			continue
		}
		fileOf[id] = path
		if !archChapters[id] {
			add(path, "GH-25 binding", "claims chapter %s, which ARCHITECTURE does not define", id)
			continue
		}

		// Every term a chapter's table defines has to reach the glossary,
		// whether or not the chapter has an SRD yet. This is the reverse of the
		// key_terms check below: that one catches a contract promising a term the
		// chapter never defines, this one catches a term coined while drafting
		// that never reached definitions.yaml (GH-52).
		var rows []string
		if km := keyTermsRe.FindStringSubmatch(body); km != nil {
			for _, m := range keyTermRowRe.FindAllStringSubmatch(km[1], -1) {
				rows = append(rows, normalizeTerm(parenGlossRe.ReplaceAllString(m[1], "")))
			}
		}
		for _, r := range rows {
			if !defined[r] {
				add(path, "V-S4", "Key Terms defines %q, which is not in docs/definitions.yaml", r)
			}
		}

		srdPath, ok := srdByChapter[id]
		if !ok {
			continue // no SRD written yet; the outline reports that gap
		}
		d := srdByPath[srdPath]
		if d == nil {
			add(path, "GH-25 binding", "ARCHITECTURE points %s at %s, which did not load", id, srdPath)
			continue
		}
		for _, t := range d.Apparatus.KeyTerms {
			if !tableDefines(rows, t) {
				add(path, "GH-25 binding", "%s lists key term %q, absent from the Key Terms table", srdPath, t)
			}
		}
		for _, c := range d.Citations {
			if c.Role == "anchor" && !strings.Contains(body, "[@"+c.ID) {
				add(path, "GH-25 binding", "%s names %s as an anchor citation, which the chapter never cites", srdPath, c.ID)
			}
		}
	}

	// A chapter the road map calls drafted must have a file claiming it.
	for _, p := range s.roadmap.Parts {
		for _, c := range p.Chapters {
			if c.Status == "drafted" && fileOf[c.ID] == "" {
				add("docs/road-map.yaml", "GH-25 binding", "%s is marked drafted but no chapter file claims it", c.ID)
			}
		}
	}
	return out
}

// sectionHeadingRe matches a numbered section or subsection heading and
// captures the leading chapter number.
var sectionHeadingRe = regexp.MustCompile(`(?m)^#{2,3} +(\d+)\.(\d+(?:\.\d+)?) `)

// checkSectionNumbering enforces voice.yaml's section_numbering rule: a
// chapter's sections carry that chapter's book-wide position, counting from
// the first chapter after the front matter and ignoring part boundaries.
//
// The rule exists because any scheme keyed on position *within a part*
// collides across parts -- the second chapter of Part I and the first chapter
// of Part II both reduce to 2 -- which is how the book ended up with two
// section 2.1s and two section 4.1s in one PDF (GH-35). Figure numbers share
// the chapter number, so they are checked against the same position.
func checkSectionNumbering(root string) []finding {
	// The introduction is front matter, not chapter 1, so it is dropped before
	// positions are assigned rather than skipped inside the loop -- skipping it
	// in place would still consume a number and shift every chapter by one.
	type chapter struct {
		path string
		body string
	}
	var chapters []chapter
	for _, path := range chapterFiles(root) {
		data, err := os.ReadFile(filepath.Join(root, path))
		if err != nil {
			continue
		}
		if isIntroduction(string(data)) {
			continue
		}
		chapters = append(chapters, chapter{path, string(data)})
	}

	var out []finding
	for i, ch := range chapters {
		path, body := ch.path, ch.body
		want := i + 1
		reported := map[string]bool{}
		for _, m := range sectionHeadingRe.FindAllStringSubmatch(body, -1) {
			if m[1] == fmt.Sprint(want) || reported[m[1]] {
				continue
			}
			reported[m[1]] = true
			out = append(out, finding{
				File:   path,
				Rule:   "voice.yaml: section_numbering",
				Detail: fmt.Sprintf("sections are numbered %s.x but this is chapter %d", m[1], want),
			})
		}
		for _, m := range figureRe.FindAllStringSubmatch(body, -1) {
			chapter := strings.SplitN(m[1], ".", 2)[0]
			if chapter == fmt.Sprint(want) || reported["fig"+chapter] {
				continue
			}
			reported["fig"+chapter] = true
			out = append(out, finding{
				File:   path,
				Rule:   "voice.yaml: section_numbering",
				Detail: fmt.Sprintf("Figure %s is numbered for chapter %s but this is chapter %d", m[1], chapter, want),
			})
		}
	}
	return out
}

// sidebarLabelRe matches a sidebar's opening line and captures its label, the
// same `> **Label:**` shape checkProse relies on.
var sidebarLabelRe = regexp.MustCompile(`^\s*>\s*\*\*([^:*]+):\*\*`)

// skillTrailerRe finds the git trailer that gh-issue-pop and do-work stamp on
// every commit they author.
var skillTrailerRe = regexp.MustCompile(`(?m)^Skill:\s*\S`)

// checkSidebarAuthorship enforces voice.yaml's author_only marking on sidebar
// types. A first-person note about the author's own experience is the one kind
// of content an agent cannot supply, and an invented incident that reads
// plausibly is indistinguishable from a real one to everyone but the author.
//
// The discriminator is the `Skill:` git trailer, not the commit author: agent
// commits in this repository carry the author's own name and address, so
// identity proves nothing, while the skill-tracing trailers are stamped only
// by the skills. Blame reports the commit that *last* touched the line, which
// is the semantics wanted here -- an agent editing the author's sidebar needs
// the same look as an agent inventing one.
//
// Outside a git repository the check yields nothing rather than failing, so
// scratch trees and fresh exports audit clean.
func checkSidebarAuthorship(root string, s *spec) []finding {
	authorOnly := map[string]bool{}
	for _, t := range s.voice.Sidebars.Types {
		if t.Authorship == "author_only" {
			authorOnly[t.Label] = true
		}
	}
	if len(authorOnly) == 0 || !isGitRepo(root) {
		return nil
	}

	var out []finding
	for _, path := range chapterFiles(root) {
		data, err := os.ReadFile(filepath.Join(root, path))
		if err != nil {
			continue
		}
		// Findings are located by the sidebar's ordinal within its file rather
		// than by line number, because the triple is also the baseline key: a
		// line number would make an accepted finding go stale on any edit
		// above it, and the churn would train readers to ignore the file.
		seen := map[string]int{}
		for i, line := range strings.Split(string(data), "\n") {
			m := sidebarLabelRe.FindStringSubmatch(line)
			if m == nil {
				continue
			}
			label := strings.TrimSpace(m[1])
			if !authorOnly[label] {
				continue
			}
			seen[label]++
			skill, ok := blameSkill(root, path, i+1)
			if !ok || skill == "" {
				continue
			}
			out = append(out, finding{
				File:   path,
				Rule:   "voice.yaml: sidebars V-B4 authorship",
				Detail: fmt.Sprintf("%s sidebar #%d was last touched by the %s skill; V-B4 is author_only", strconv.Quote(label), seen[label], skill),
			})
		}
	}
	return out
}

// isGitRepo reports whether root sits inside a working tree.
func isGitRepo(root string) bool {
	out, err := gitOutput(root, "rev-parse", "--is-inside-work-tree")
	return err == nil && strings.TrimSpace(out) == "true"
}

// blameSkill returns the Skill trailer of the commit that last touched line of
// path. ok is false when the line has no commit yet -- uncommitted or
// untracked -- which is not a finding, since nothing has been recorded to
// judge.
func blameSkill(root, path string, line int) (skill string, ok bool) {
	out, err := gitOutput(root, "blame", "--porcelain", "-L", fmt.Sprintf("%d,%d", line, line), "--", path)
	if err != nil {
		return "", false
	}
	fields := strings.Fields(out)
	if len(fields) == 0 {
		return "", false
	}
	sha := fields[0]
	// An all-zero sha is blame's marker for a not-yet-committed line.
	if strings.Trim(sha, "0") == "" {
		return "", false
	}
	msg, err := gitOutput(root, "log", "-1", "--format=%B", sha)
	if err != nil {
		return "", false
	}
	if !skillTrailerRe.MatchString(msg) {
		return "", true
	}
	name, err := gitOutput(root, "log", "-1", "--format=%(trailers:key=Skill,valueonly,separator=%x20)", sha)
	if err != nil {
		return "", false
	}
	return strings.TrimSpace(name), true
}

func gitOutput(root string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = root
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return stdout.String(), nil
}

// benignConstructions records phrasings in which a forbidden term is doing
// ordinary comparative or contrastive work rather than the minimizing work
// voice.yaml bars. The constitution already says a term of art that trips the
// lexical scan stays; this makes that allowance mechanical, so the check stays
// worth reading rather than being tuned out.
var benignConstructions = map[string][]string{
	"just": {"not just", "just as", "just like"},
}

// usesForbiddenTerm reports whether a forbidden term appears in a sense the
// voice constitution actually bars.
func usesForbiddenTerm(prose, term string) bool {
	re := regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(term) + `\b`)
	benign := benignConstructions[strings.ToLower(term)]
	quoted := quotedSpans(prose)
	for _, loc := range re.FindAllStringIndex(prose, -1) {
		if inBenignConstruction(prose, loc, benign) || inAnySpan(quoted, loc) {
			continue
		}
		return true
	}
	return false
}

// quotedSpans returns the half-open ranges covered by double-quoted passages.
//
// voice.yaml governs the author's voice. A forbidden term inside a quotation is
// someone else's word -- C1.5 quotes the naive belief that a model "just
// guesses randomly" in order to refute it, and rewriting that to satisfy a
// grep would damage the sentence to no purpose (GH-57). Straight and curly
// pairs both count; an unclosed quote covers nothing, so a stray apostrophe
// cannot silence the rest of a chapter.
func quotedSpans(prose string) [][2]int {
	var spans [][2]int
	var open = -1
	for i, r := range prose {
		switch r {
		case '"', '\u201c', '\u201d':
			if open < 0 {
				open = i
			} else {
				spans = append(spans, [2]int{open, i + len(string(r))})
				open = -1
			}
		case '\n':
			// A quotation does not span a blank line; reset at paragraph breaks
			// so an unmatched quote cannot swallow the document.
			if open >= 0 && i+1 < len(prose) && prose[i+1] == '\n' {
				open = -1
			}
		}
	}
	return spans
}

func inAnySpan(spans [][2]int, loc []int) bool {
	for _, s := range spans {
		if loc[0] >= s[0] && loc[1] <= s[1] {
			return true
		}
	}
	return false
}

func inBenignConstruction(prose string, loc []int, benign []string) bool {
	const window = 12
	start := max(0, loc[0]-window)
	end := min(len(prose), loc[1]+window)
	around := strings.ToLower(prose[start:end])
	for _, phrase := range benign {
		if strings.Contains(around, phrase) {
			return true
		}
	}
	return false
}

// chapterFiles lists the numbered markdown files that carry chapter prose,
// excluding part dividers, the references list, and files with no heading.
func chapterFiles(root string) []string {
	paths, _ := filepath.Glob(filepath.Join(root, "[0-9][0-9]-*.md"))
	sort.Strings(paths)
	var out []string
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		if isChapterBody(string(data)) {
			out = append(out, rel(root, p))
		}
	}
	return out
}

func isChapterBody(body string) bool {
	title := firstHeading(body)
	if title == "" {
		return false
	}
	return !strings.HasPrefix(title, "Part ") && !strings.EqualFold(title, "References")
}

func isIntroduction(body string) bool {
	return strings.EqualFold(firstHeading(body), "Introduction")
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

// ------------------------------------------------------------ build checks

var imageRe = regexp.MustCompile(`!\[[^\]]*\]\(([^)]+)\)|\\includegraphics(?:\[[^\]]*\])?\{([^}]+)\}`)

// checkFigureReferences resolves every image a chapter references, the way
// autogenic's audit resolves \includegraphics against fig/.
func checkFigureReferences(root string) []finding {
	var out []finding
	for _, path := range chapterFiles(root) {
		data, err := os.ReadFile(filepath.Join(root, path))
		if err != nil {
			continue
		}
		for _, m := range imageRe.FindAllStringSubmatch(string(data), -1) {
			ref := m[1]
			if ref == "" {
				ref = m[2]
			}
			if ref == "" || strings.HasPrefix(ref, "http") {
				continue
			}
			if !figureExists(root, ref) {
				out = append(out, finding{
					File:   path,
					Rule:   "build",
					Detail: fmt.Sprintf("missing figure %q (build with `mage figures`)", ref),
				})
			}
		}
	}
	return out
}

func figureExists(root, name string) bool {
	candidates := []string{name}
	if filepath.Ext(name) == "" {
		for _, ext := range []string{".png", ".pdf", ".jpg", ".jpeg", ".svg"} {
			candidates = append(candidates, name+ext)
		}
	}
	for _, c := range candidates {
		if _, err := os.Stat(filepath.Join(root, c)); err == nil {
			return true
		}
		if _, err := os.Stat(filepath.Join(root, figuresDir, c)); err == nil {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------- baseline

type baselineDoc struct {
	Accepted []struct {
		File   string `yaml:"file"`
		Rule   string `yaml:"rule"`
		Detail string `yaml:"detail"`
		Issue  string `yaml:"issue"`
	} `yaml:"accepted"`
}

// loadBaseline reads the accepted-findings file. Every entry names the issue
// that clears it. An entry matching nothing is reported so the file cannot
// drift into a blanket exemption.
func loadBaseline(root string) (map[string]string, []finding) {
	path := filepath.Join(root, "docs", "audit-baseline.yaml")
	accepted := map[string]string{}
	var doc baselineDoc
	if err := readYAML(path, &doc); err != nil {
		if os.IsNotExist(err) {
			return accepted, nil
		}
		return accepted, []finding{{File: "docs/audit-baseline.yaml", Rule: "baseline", Detail: err.Error()}}
	}
	var findings []finding
	for _, e := range doc.Accepted {
		if e.Issue == "" {
			findings = append(findings, finding{
				File:   "docs/audit-baseline.yaml",
				Rule:   "baseline",
				Detail: fmt.Sprintf("entry %s|%s has no issue; accepted debt names the issue that clears it", e.File, e.Rule),
			})
		}
		accepted[finding{File: e.File, Rule: e.Rule, Detail: e.Detail}.key()] = e.Issue
	}
	return accepted, findings
}

// staleBaselineFindings reports accepted entries that matched nothing this
// run, so a fixed defect does not leave a permanent exemption behind.
func staleBaselineFindings(accepted map[string]string, matched map[string]bool) []finding {
	var out []finding
	keys := make([]string, 0, len(accepted))
	for k := range accepted {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if !matched[k] {
			out = append(out, finding{
				File:   "docs/audit-baseline.yaml",
				Rule:   "baseline",
				Detail: fmt.Sprintf("accepted finding %q no longer occurs; delete the entry", k),
			})
		}
	}
	return out
}
