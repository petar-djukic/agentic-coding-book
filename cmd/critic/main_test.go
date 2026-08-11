// Copyright (c) 2026 Petar Djukic. All rights reserved.
// SPDX-License-Identifier: MIT

package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseReportTolerantOfFences(t *testing.T) {
	raw := "```json\n{\"findings\":[{\"rule_id\":\"V-S1\",\"severity\":\"blocking\"," +
		"\"location\":\"line 3\",\"quote\":\"x\",\"issue\":\"y\",\"suggestion\":\"z\"}]," +
		"\"summary\":\"one finding\"}\n```"
	r, err := parseReport(raw)
	if err != nil {
		t.Fatalf("parseReport: %v", err)
	}
	if len(r.Findings) != 1 || r.Findings[0].RuleID != "V-S1" {
		t.Fatalf("got %+v", r)
	}
}

func TestParseReportRejectsNonJSON(t *testing.T) {
	if _, err := parseReport("I could not review this chapter."); err == nil {
		t.Fatal("expected an error for a non-JSON response")
	}
}

func TestPrintReportCountsBlocking(t *testing.T) {
	r := report{Findings: []finding{
		{RuleID: "V-P2", Severity: "advisory", Location: "line 1", Issue: "a"},
		{RuleID: "A-C4", Severity: "blocking", Location: "line 2", Issue: "b"},
		{RuleID: "VN-6", Severity: "BLOCKING", Location: "line 3", Issue: "c"},
	}}
	if got := printReport(r); got != 2 {
		t.Fatalf("blocking count = %d, want 2", got)
	}
	if got := printReport(report{}); got != 0 {
		t.Fatalf("empty report blocking count = %d, want 0", got)
	}
}

func TestIsChapterExcludesPartDividersAndReferences(t *testing.T) {
	cases := []struct {
		body string
		want bool
	}{
		{"# What is a harness?\n\ntext\n", true},
		{"# Part I — Agents and Harnesses\n\ntext\n", false},
		{"# References\n\ntext\n", false},
		{"---\ntitle: front matter\n---\n", false},
	}
	for _, c := range cases {
		if got := isChapter(c.body); got != c.want {
			t.Errorf("isChapter(%q) = %v, want %v", c.body, got, c.want)
		}
	}
}

func TestSelectChaptersSkipsDividers(t *testing.T) {
	dir := t.TempDir()
	write := func(name, body string) {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("02-part-i.md", "# Part I — Agents and Harnesses\n")
	write("03-what-is-a-harness.md", "# What is a harness?\n")
	write("12-references.md", "# References\n")

	wd, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(wd) })
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	got, err := selectChapters(nil)
	if err != nil {
		t.Fatalf("selectChapters: %v", err)
	}
	if len(got) != 1 || got[0] != "03-what-is-a-harness.md" {
		t.Fatalf("got %v, want [03-what-is-a-harness.md]", got)
	}
}

func TestPairSRDMatchesByTitle(t *testing.T) {
	srds := []srd{{raw: "contract: yes"}}
	srds[0].Meta.Chapter = "C1.3"
	srds[0].Meta.Title = "What is a harness?"

	contract, note := pairSRD("03-x.md", "# What is a harness?\n", "", srds)
	if contract != "contract: yes" {
		t.Fatalf("contract = %q, want the paired SRD", contract)
	}
	if note == "" {
		t.Fatal("expected a note naming the paired contract")
	}

	contract, note = pairSRD("03-x.md", "# Something Else\n", "", srds)
	if contract != "" {
		t.Fatal("expected no contract for an unmatched title")
	}
	if note == "" {
		t.Fatal("expected a note explaining the miss")
	}
}

// The critic reports; it never edits. Guard the property by construction:
// the package must open no file for writing.
func TestCriticNeverWritesFiles(t *testing.T) {
	src, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"os.WriteFile", "os.Create", "os.OpenFile", "os.Remove", "os.Rename"} {
		if containsToken(string(src), forbidden) {
			t.Errorf("main.go calls %s; the critic must never modify a chapter", forbidden)
		}
	}
}

func containsToken(hay, needle string) bool {
	return len(hay) >= len(needle) && (func() bool {
		for i := 0; i+len(needle) <= len(hay); i++ {
			if hay[i:i+len(needle)] == needle {
				return true
			}
		}
		return false
	})()
}
