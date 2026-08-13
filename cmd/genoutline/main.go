// Copyright (c) 2026 Petar Djukic. All rights reserved.
// SPDX-License-Identifier: MIT

// Command genoutline renders the book's outline from its specification.
//
// It reads docs/ARCHITECTURE.yaml for the part and chapter order,
// docs/road-map.yaml for per-chapter status, and docs/srd/*.yaml for each
// chapter's drafting contract, then writes a markdown outline on stdout:
// part, chapter, goal, objective, subgoals, the derivation-chain links the
// chapter owns, its content spine, figures, links, and gaps.
//
// The outline exists so the book's structure is reviewable as one artifact
// before any chapter is drafted. Chapters with no SRD are listed with their
// road-map status, so the document doubles as a coverage report.
//
//	go run ./cmd/genoutline > outline.md
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// chapterRef is named rather than anonymous so the render helpers can take a
// part's chapters as a parameter (GH-71).
type chapterRef struct {
	ID    string `yaml:"id"`
	Title string `yaml:"title"`
	SRD   string `yaml:"srd"`
}

type architecture struct {
	Meta struct {
		Name string `yaml:"name"`
	} `yaml:"meta"`
	Structure struct {
		Parts []struct {
			ID       string       `yaml:"id"`
			Title    string       `yaml:"title"`
			Role     string       `yaml:"role"`
			Chapters []chapterRef `yaml:"chapters"`
		} `yaml:"parts"`
	} `yaml:"structure"`
}

type vision struct {
	Meta struct {
		Artifact string `yaml:"artifact"`
	} `yaml:"meta"`
}

type roadmap struct {
	Releases []struct {
		ID   string `yaml:"id"`
		Gate string `yaml:"gate"`
	} `yaml:"releases"`
	Parts []struct {
		Chapters []struct {
			ID     string `yaml:"id"`
			Status string `yaml:"status"`
		} `yaml:"chapters"`
	} `yaml:"parts"`
}

type srd struct {
	Meta struct {
		Chapter string `yaml:"chapter"`
		Title   string `yaml:"title"`
		Parent  string `yaml:"parent"`
		Status  string `yaml:"status"`
	} `yaml:"meta"`
	SectionGoal string `yaml:"section_goal"`
	Goals       []struct {
		ID   string `yaml:"id"`
		Goal string `yaml:"goal"`
	} `yaml:"goals"`
	Chain []struct {
		ID   string `yaml:"id"`
		Owns string `yaml:"owns"`
	} `yaml:"chain"`
	Objective string `yaml:"objective"`
	Content   []struct {
		Say   string   `yaml:"say"`
		Cites []string `yaml:"cites"`
	} `yaml:"content"`
	Figures []struct {
		Shows  string `yaml:"shows"`
		Status string `yaml:"status"`
	} `yaml:"figures"`
	Links struct {
		Requires []string `yaml:"requires"`
		Supports []string `yaml:"supports"`
	} `yaml:"links"`
	Gaps []string `yaml:"gaps"`
}

func main() {
	docs := flag.String("docs", "docs", "path to the docs directory")
	flag.Parse()

	if err := run(*docs, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "genoutline: %v\n", err)
		os.Exit(1)
	}
}

func run(docs string, out *os.File) error {
	var arch architecture
	if err := load(filepath.Join(docs, "ARCHITECTURE.yaml"), &arch); err != nil {
		return err
	}
	var vis vision
	if err := load(filepath.Join(docs, "VISION.yaml"), &vis); err != nil {
		return err
	}

	// road-map.yaml supplies status for chapters that have no SRD; a missing
	// road-map costs the status column, not the outline.
	status := map[string]string{}
	// gatedBy maps a chapter id to the releases whose gate text names it. A gap
	// in such a chapter is not merely worth filling; it is holding a release,
	// and the background-needed appendix says so (GH-71).
	gatedBy := map[string][]string{}
	var rm roadmap
	if err := load(filepath.Join(docs, "road-map.yaml"), &rm); err == nil {
		for _, p := range rm.Parts {
			for _, c := range p.Chapters {
				status[c.ID] = c.Status
			}
		}
		for _, p := range rm.Parts {
			for _, c := range p.Chapters {
				for _, rel := range rm.Releases {
					if c.ID != "" && strings.Contains(rel.Gate, c.ID) {
						gatedBy[c.ID] = append(gatedBy[c.ID], rel.ID)
					}
				}
			}
		}
	}

	srds, err := loadSRDs(filepath.Join(docs, "srd"))
	if err != nil {
		return err
	}
	if len(srds) == 0 {
		return fmt.Errorf("no SRDs found in %s: write one before running mage outline (see %s/README.md)",
			filepath.Join(docs, "srd"), filepath.Join(docs, "srd"))
	}

	var b strings.Builder
	writeHeader(&b, vis.Meta.Artifact, len(srds))
	for _, part := range arch.Structure.Parts {
		fmt.Fprintf(&b, "\n# %s — %s\n\n", part.ID, part.Title)
		if part.Role != "" {
			fmt.Fprintf(&b, "*%s*\n", oneLine(part.Role))
		}
		writeBuildThread(&b, part.Chapters, srds)
		for _, ch := range part.Chapters {
			s, ok := srds[ch.ID]
			if !ok {
				writeChapterWithoutSRD(&b, ch.ID, ch.Title, status[ch.ID])
				continue
			}
			writeChapter(&b, ch.ID, ch.Title, status[ch.ID], s)
		}
	}
	writeBackgroundNeeded(&b, arch, srds, gatedBy, status)
	_, err = out.WriteString(b.String())
	return err
}

// writeBuildThread renders what the reader builds in this part. The build
// thread is specified as G6 subgoals distributed across the chapters that
// deliver them, so it is derived from the SRDs rather than restated in a
// second place. A part with no G6 subgoal renders nothing (GH-71).
func writeBuildThread(b *strings.Builder, chapters []chapterRef, srds map[string]srd) {
	type item struct{ id, chapter, goal string }
	var items []item
	for _, ch := range chapters {
		s, ok := srds[ch.ID]
		if !ok {
			continue
		}
		for _, g := range s.Goals {
			if strings.HasPrefix(g.ID, "G6.") {
				items = append(items, item{g.ID, ch.ID, oneLine(g.Goal)})
			}
		}
	}
	if len(items) == 0 {
		return
	}
	b.WriteString("\n**Build thread.** By the end of this part the reader has built:\n\n")
	for _, it := range items {
		fmt.Fprintf(b, "- **%s** (%s) %s\n", it.id, it.chapter, it.goal)
	}
	b.WriteString("\n")
}

// writeBackgroundNeeded collects every SRD's gaps into one appendix. Scattered
// across 39 files they are unreadable; collected they are the book's
// pre-drafting work list. A gap in a chapter a release gate names is marked,
// because that one is holding a release rather than merely waiting (GH-71).
func writeBackgroundNeeded(b *strings.Builder, arch architecture, srds map[string]srd, gatedBy map[string][]string, status map[string]string) {
	var any bool
	for _, s := range srds {
		if len(s.Gaps) > 0 {
			any = true
			break
		}
	}
	if !any {
		return
	}
	b.WriteString("\n# Appendix — Background needed\n\n")
	b.WriteString("Every chapter contract records what must exist before the chapter can be " +
		"drafted. Collected here, they are the book's pre-drafting work list. Items marked " +
		"**gating** sit in a chapter a release gate names, so they hold a release rather " +
		"than merely waiting. A chapter already drafted carries its status here too, because " +
		"its entries are mostly decisions taken and recorded rather than work outstanding.\n")
	for _, part := range arch.Structure.Parts {
		var wrote bool
		for _, ch := range part.Chapters {
			s, ok := srds[ch.ID]
			if !ok || len(s.Gaps) == 0 {
				continue
			}
			if !wrote {
				fmt.Fprintf(b, "\n## %s — %s\n", part.ID, part.Title)
				wrote = true
			}
			mark := ""
			if st := status[ch.ID]; st != "" {
				mark = fmt.Sprintf(" — %s", st)
			}
			if rels := gatedBy[ch.ID]; len(rels) > 0 {
				mark += fmt.Sprintf(" — **gating** %s", strings.Join(rels, ", "))
			}
			fmt.Fprintf(b, "\n**%s %s**%s\n\n", ch.ID, ch.Title, mark)
			for _, g := range s.Gaps {
				fmt.Fprintf(b, "- %s\n", oneLine(g))
			}
		}
	}
	b.WriteString("\n")
}

func writeHeader(b *strings.Builder, artifact string, n int) {
	date := time.Now().Format("2006-01-02")
	fmt.Fprintf(b, `---
title: "Agentic Coding — Outline"
subtitle: "Generated from docs/srd/ (%d chapter contracts)"
author: "Petar Djukic"
date: %s
titlepage: true
toc: true
toc-own-page: true
toc-depth: 2
---

`, n, date)
	fmt.Fprintf(b, "%s\n\n", oneLine(artifact))
	b.WriteString("This outline is generated from the specification, not written by hand. " +
		"Each chapter shows the contract its draft is checked against: the goal, the " +
		"objective, the subgoals, the derivation-chain links it owns, and the content " +
		"spine. Chapters with no SRD carry their road-map status instead, so gaps in " +
		"coverage are visible rather than silent.\n")
}

func writeChapter(b *strings.Builder, id, title, status string, s srd) {
	fmt.Fprintf(b, "\n## %s %s\n\n", id, title)
	if status != "" {
		fmt.Fprintf(b, "**Status.** %s", status)
		if s.Meta.Status != "" && s.Meta.Status != status {
			fmt.Fprintf(b, " (SRD records %s)", s.Meta.Status)
		}
		b.WriteString("\n\n")
	}
	if s.SectionGoal != "" {
		fmt.Fprintf(b, "**Goal.** The goal of this chapter is to %s.\n\n", trimPeriod(oneLine(s.SectionGoal)))
	}
	if s.Objective != "" {
		fmt.Fprintf(b, "**Objective.** %s\n\n", oneLine(s.Objective))
	}
	if len(s.Chain) > 0 {
		var links []string
		for _, c := range s.Chain {
			if c.Owns != "" {
				links = append(links, fmt.Sprintf("%s (%s)", c.ID, c.Owns))
				continue
			}
			links = append(links, c.ID)
		}
		fmt.Fprintf(b, "**Answers.** %s\n\n", strings.Join(links, ", "))
	}
	if len(s.Goals) > 0 {
		b.WriteString("**Subgoals**\n\n")
		for _, g := range s.Goals {
			fmt.Fprintf(b, "- **%s** %s\n", g.ID, oneLine(g.Goal))
		}
		b.WriteString("\n")
	}
	if len(s.Content) > 0 {
		b.WriteString("**Content**\n\n")
		for i, c := range s.Content {
			fmt.Fprintf(b, "%d. %s", i+1, oneLine(c.Say))
			if len(c.Cites) > 0 {
				fmt.Fprintf(b, " `[%s]`", strings.Join(c.Cites, ", "))
			}
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}
	for _, f := range s.Figures {
		fmt.Fprintf(b, "**Figure** (%s). %s\n\n", orUnknown(f.Status), oneLine(f.Shows))
	}
	if len(s.Links.Requires) > 0 || len(s.Links.Supports) > 0 {
		b.WriteString("**Links.**")
		if len(s.Links.Requires) > 0 {
			fmt.Fprintf(b, " Requires %s.", strings.Join(s.Links.Requires, ", "))
		}
		if len(s.Links.Supports) > 0 {
			fmt.Fprintf(b, " Supports %s.", strings.Join(s.Links.Supports, ", "))
		}
		b.WriteString("\n\n")
	}
	if len(s.Gaps) > 0 {
		b.WriteString("**Gaps**\n\n")
		for _, g := range s.Gaps {
			fmt.Fprintf(b, "- %s\n", oneLine(g))
		}
		b.WriteString("\n")
	}
}

func writeChapterWithoutSRD(b *strings.Builder, id, title, status string) {
	fmt.Fprintf(b, "\n## %s %s\n\n", id, title)
	fmt.Fprintf(b, "**Status.** %s — no SRD written.\n\n", orUnknown(status))
}

func loadSRDs(dir string) (map[string]srd, error) {
	paths, err := filepath.Glob(filepath.Join(dir, "srd-*.yaml"))
	if err != nil {
		return nil, err
	}
	sort.Strings(paths)
	out := make(map[string]srd, len(paths))
	for _, p := range paths {
		var s srd
		if err := load(p, &s); err != nil {
			return nil, err
		}
		if s.Meta.Chapter == "" {
			return nil, fmt.Errorf("%s: meta.chapter is empty; every SRD names the chapter it governs", p)
		}
		if prev, dup := out[s.Meta.Chapter]; dup {
			return nil, fmt.Errorf("%s: chapter %s already governed by an SRD titled %q",
				p, s.Meta.Chapter, prev.Meta.Title)
		}
		out[s.Meta.Chapter] = s
	}
	return out, nil
}

func load(path string, into interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	if err := yaml.Unmarshal(data, into); err != nil {
		return fmt.Errorf("parse %s: %w", path, err)
	}
	return nil
}

// oneLine collapses a folded YAML scalar onto a single line so it survives
// markdown, where a stray newline inside a bold run breaks the emphasis.
func oneLine(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func trimPeriod(s string) string {
	return strings.TrimRight(s, ".")
}

func orUnknown(s string) string {
	if s == "" {
		return "unknown"
	}
	return s
}
