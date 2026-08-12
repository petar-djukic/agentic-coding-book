// Copyright (c) 2026 Petar Djukic. All rights reserved.
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// tree is a scratch repository: a path-to-content map that each test mutates
// in exactly one place, so a failure names the check that caught it.
type tree map[string]string

func cleanTree() tree {
	return tree{
		"outline.yaml": `
parts:
  - id: part-i
    chapters:
      - title: What is an agent?
      - title: Second chapter
`,
		"docs/ARCHITECTURE.yaml": `
structure:
  parts:
    - id: P1
      chapters:
        - id: C1.1
          srd: docs/srd/srd-1.1-what-is-an-agent.yaml
        - id: C1.2
`,
		"docs/road-map.yaml": `
parts:
  - id: P1
    chapters:
      - id: C1.1
        status: outline
      - id: C1.2
        status: outline
`,
		"docs/VISION.yaml": `
goals:
  - id: G1
    goal: Orient.
`,
		"docs/definitions.yaml": `
agent:
  definition: A state machine plus tools.
`,
		"docs/constitutions/argument.yaml": `
derivation_chain:
  - id: Q1
    question: What writes the code?
    owners: [C1.1]
part_obligations:
  parts:
    - id: P1
      owes: Establishes Q1.
claims_register:
  rules:
    - id: A-C1
      rule: Every claim carries evidence.
`,
		"docs/constitutions/voice.yaml": `
structure_rules:
  - id: V-S1
    name: learning_objectives
    rule: Every chapter opens with learning objectives.
forbidden_terms:
  terms: [powerful, just]
sidebars:
  types:
    - id: V-B1
      label: Common Error
      former_label: Common Clauding Error
`,
		"docs/srd/srd-1.1-what-is-an-agent.yaml": `
meta:
  chapter: C1.1
  title: What is an agent?
goals:
  - id: G1.1
    goal: The reader can define an agent.
chain:
  - id: Q1
constitutions:
  - file: ../constitutions/voice.yaml
    rules: [V-S1]
citations:
  - id: hunt1999
    role: anchor
apparatus:
  key_terms: [agent]
`,
		"references.yaml": `
- id: hunt1999
  title: The Pragmatic Programmer
`,
		"03-what-is-an-agent.md": `<!-- chapter: C1.1 -->

# What is an agent?

## Learning Objectives

1. Define an agent.

Prose that cites something [@hunt1999].

## Summary

It is a state machine plus tools.

## Key Terms

| Term | Definition |
|---|---|
| **Agent** | A state machine plus tools |
`,
	}
}

func write(t *testing.T, files tree) string {
	t.Helper()
	root := t.TempDir()
	for path, content := range files {
		full := filepath.Join(root, path)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

// auditTree runs the audit with the figure render and the book build skipped;
// those need pandoc and are covered by mage all.
func auditTree(t *testing.T, files tree) error {
	t.Helper()
	return audit(write(t, files), nil, nil)
}

func TestAuditPassesOnACleanSpec(t *testing.T) {
	if err := auditTree(t, cleanTree()); err != nil {
		t.Fatalf("clean tree should audit clean, got:\n%v", err)
	}
}

// Each case breaks exactly one thing and names the text the audit must report.
func TestAuditCatchesEachBreakage(t *testing.T) {
	cases := []struct {
		name   string
		break_ func(tree)
		want   string
	}{
		{
			name: "SRD cites a rule that does not exist",
			break_: func(f tree) {
				f["docs/srd/srd-1.1-what-is-an-agent.yaml"] = replace(f, "rules: [V-S1]", "rules: [V-S9]")
			},
			want: "cites unknown rule V-S9",
		},
		{
			name:   "SRD claims a chain link it does not own",
			break_: func(f tree) { f["docs/constitutions/argument.yaml"] = replace2(f, "owners: [C1.1]", "owners: [C1.2]") },
			want:   "argument.yaml owners are",
		},
		{
			name:   "chain owner is not a chapter",
			break_: func(f tree) { f["docs/constitutions/argument.yaml"] = replace2(f, "owners: [C1.1]", "owners: [C9.9]") },
			want:   "owns unknown chapter C9.9",
		},
		{
			name: "a part carries no obligation",
			break_: func(f tree) {
				f["docs/constitutions/argument.yaml"] = replace2(f, "    - id: P1\n      owes: Establishes Q1.\n", "")
			},
			want: "no obligation for part P1",
		},
		{
			name:   "subgoal hangs off no book goal",
			break_: func(f tree) { f["docs/srd/srd-1.1-what-is-an-agent.yaml"] = replace(f, "id: G1.1", "id: G7.1") },
			want:   "has no VISION goal G7",
		},
		{
			name: "citation does not resolve",
			break_: func(f tree) {
				f["docs/srd/srd-1.1-what-is-an-agent.yaml"] = replace(f, "id: hunt1999", "id: nosuch2026")
			},
			want: "cites nosuch2026, which does not resolve",
		},
		{
			name: "key term is not defined",
			break_: func(f tree) {
				f["docs/srd/srd-1.1-what-is-an-agent.yaml"] = replace(f, "key_terms: [agent]", "key_terms: [harness]")
			},
			want: "harness is not in docs/definitions.yaml",
		},
		{
			name: "SRD governs a chapter that does not exist",
			break_: func(f tree) {
				f["docs/srd/srd-1.1-what-is-an-agent.yaml"] = replace(f, "chapter: C1.1", "chapter: C4.9")
			},
			want: "C4.9 is not a chapter in ARCHITECTURE",
		},
		{
			name: "two SRDs govern the same chapter",
			break_: func(f tree) {
				f["docs/srd/srd-1.1-duplicate.yaml"] = f["docs/srd/srd-1.1-what-is-an-agent.yaml"]
			},
			want: "already governed by",
		},
		{
			name:   "road-map chapter ids drift from ARCHITECTURE",
			break_: func(f tree) { f["docs/road-map.yaml"] = replace(f, "id: C1.1", "id: C1.2") },
			want:   "chapter ids or order differ",
		},
		{
			name:   "outline chapter count drifts from ARCHITECTURE",
			break_: func(f tree) { f["outline.yaml"] = replace(f, "      - title: What is an agent?\n", "") },
			want:   "chapters, ARCHITECTURE has",
		},
		{
			name:   "chapter cites a key that does not resolve",
			break_: func(f tree) { f["03-what-is-an-agent.md"] = replace(f, "[@hunt1999]", "[@ghost2026]") },
			want:   "cites [@ghost2026]",
		},
		{
			name: "chapter uses a forbidden term",
			break_: func(f tree) {
				f["03-what-is-an-agent.md"] = replace(f, "Prose that cites", "A powerful thing that cites")
			},
			want: `uses "powerful"`,
		},
		{
			name: "chapter uses a retired sidebar label",
			break_: func(f tree) {
				f["03-what-is-an-agent.md"] = replace(f, "Prose that", "> **Common Clauding Error:** prose that")
			},
			want: "retired sidebar label",
		},
		{
			name:   "chapter has no learning objectives",
			break_: func(f tree) { f["03-what-is-an-agent.md"] = replace(f, "## Learning Objectives", "## Overview") },
			want:   "no Learning Objectives",
		},
		{
			name:   "chapter has no summary",
			break_: func(f tree) { f["03-what-is-an-agent.md"] = replace(f, "## Summary", "## Wrapping Up") },
			want:   "no Summary",
		},
		{
			name:   "chapter references a figure that does not exist",
			break_: func(f tree) { f["03-what-is-an-agent.md"] += "\n![skeleton](figures/fig-1-1.png)\n" },
			want:   "missing figure",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			files := cleanTree()
			c.break_(files)
			err := auditTree(t, files)
			if err == nil {
				t.Fatalf("audit passed; expected a finding containing %q", c.want)
			}
			if !strings.Contains(err.Error(), c.want) {
				t.Fatalf("finding %q not reported; got:\n%v", c.want, err)
			}
		})
	}
}

func TestBaselineAcceptsAKnownFinding(t *testing.T) {
	files := cleanTree()
	files["03-what-is-an-agent.md"] = strings.Replace(
		files["03-what-is-an-agent.md"], "Prose that cites", "A powerful thing that cites", 1)
	files["docs/audit-baseline.yaml"] = `
accepted:
  - file: 03-what-is-an-agent.md
    rule: 'voice.yaml: forbidden_terms'
    detail: 'uses "powerful"'
    issue: GH-28
`
	if err := auditTree(t, files); err != nil {
		t.Fatalf("baselined finding should not fail the audit, got:\n%v", err)
	}
}

func TestBaselineRejectsAnEntryWithNoIssue(t *testing.T) {
	files := cleanTree()
	files["03-what-is-an-agent.md"] = strings.Replace(
		files["03-what-is-an-agent.md"], "Prose that cites", "A powerful thing that cites", 1)
	files["docs/audit-baseline.yaml"] = `
accepted:
  - file: 03-what-is-an-agent.md
    rule: 'voice.yaml: forbidden_terms'
    detail: 'uses "powerful"'
`
	err := auditTree(t, files)
	if err == nil || !strings.Contains(err.Error(), "has no issue") {
		t.Fatalf("an accepted entry with no issue should be a finding, got:\n%v", err)
	}
}

func TestBaselineReportsAStaleEntry(t *testing.T) {
	files := cleanTree()
	files["docs/audit-baseline.yaml"] = `
accepted:
  - file: 03-what-is-an-agent.md
    rule: 'voice.yaml: forbidden_terms'
    detail: 'uses "powerful"'
    issue: GH-28
`
	err := auditTree(t, files)
	if err == nil || !strings.Contains(err.Error(), "no longer occurs") {
		t.Fatalf("a stale accepted entry should be a finding, got:\n%v", err)
	}
}

func TestForbiddenTermSkipsComparativeConstructions(t *testing.T) {
	benign := []string{
		"the loops split apart, just as the two kinds of intent do",
		"not just the inner loop, but all five steps",
		"it reads just like the surrounding code",
	}
	for _, s := range benign {
		if usesForbiddenTerm(s, "just") {
			t.Errorf("flagged a comparative use: %q", s)
		}
	}
	if !usesForbiddenTerm("just add a verification step", "just") {
		t.Error("failed to flag the minimizing use")
	}
}

func TestIntroductionIsExemptFromChapterApparatus(t *testing.T) {
	files := cleanTree()
	files["01-introduction.md"] = "# Introduction\n\nFront matter, not a chapter.\n"
	if err := auditTree(t, files); err != nil {
		t.Fatalf("the introduction is front matter and owes no apparatus, got:\n%v", err)
	}
}

func TestPartDividersAreNotChapters(t *testing.T) {
	files := cleanTree()
	files["02-part-i.md"] = "# Part I — Agents and Harnesses\n\nA divider, not a chapter.\n"
	if err := auditTree(t, files); err != nil {
		t.Fatalf("a part divider owes no chapter apparatus, got:\n%v", err)
	}
}

// replace edits the SRD fixture; replace2 edits the argument constitution.
// Both exist so the table cases read as one-line mutations.
func replace(f tree, old, new string) string {
	for _, key := range []string{
		"docs/srd/srd-1.1-what-is-an-agent.yaml",
		"docs/road-map.yaml",
		"outline.yaml",
		"03-what-is-an-agent.md",
	} {
		if strings.Contains(f[key], old) {
			return strings.Replace(f[key], old, new, 1)
		}
	}
	return ""
}

func replace2(f tree, old, new string) string {
	return strings.Replace(f["docs/constitutions/argument.yaml"], old, new, 1)
}

// ------------------------------------------- V-B4 sidebar authorship (GH-46)

// gitTree writes files into a scratch repository and commits them, so blame
// has something to report. The existing tree fixture is a plain directory,
// which is deliberate: it proves the check stays quiet outside a repository.
func gitTree(t *testing.T, files tree, commitMessage string) string {
	t.Helper()
	root := write(t, files)
	run := func(args ...string) {
		t.Helper()
		if out, err := gitOutput(root, args...); err != nil {
			t.Fatalf("git %s: %v (%s)", strings.Join(args, " "), err, out)
		}
	}
	run("init", "--quiet")
	run("config", "user.email", "author@example.com")
	run("config", "user.name", "Author")
	run("add", "-A")
	run("commit", "--quiet", "-m", commitMessage)
	return root
}

// voiceWithAuthorOnly is the fixture's voice.yaml with From the Field marked
// author_only, matching the real constitution.
const voiceWithAuthorOnly = `
structure_rules:
  - id: V-S1
    name: learning_objectives
    rule: Every chapter opens with learning objectives.
forbidden_terms:
  terms: [powerful, just]
sidebars:
  types:
    - id: V-B1
      label: Common Error
      former_label: Common Clauding Error
    - id: V-B4
      label: From the Field
      authorship: author_only
`

func chapterWithSidebar(label string) string {
	return `<!-- chapter: C1.1 -->

# What is an agent?

## Learning Objectives

1. Define an agent.

Prose that cites something [@hunt1999].

> **` + label + `:** Something that happened once.

## Summary

It is a state machine plus tools.

## Key Terms

| Term | Definition |
|---|---|
| **Agent** | A state machine plus tools |
`
}

func TestSidebarAuthorshipFlagsAgentAuthoredFromTheField(t *testing.T) {
	files := cleanTree()
	files["docs/constitutions/voice.yaml"] = voiceWithAuthorOnly
	files["03-what-is-an-agent.md"] = chapterWithSidebar("From the Field")

	root := gitTree(t, files, "Add the chapter\n\nSkill: do-work\nCalled-by: gh-issue-pop")
	err := audit(root, nil, nil)
	if err == nil {
		t.Fatal("a From the Field sidebar committed by a skill should be a finding")
	}
	for _, want := range []string{"V-B4 authorship", "do-work", "author_only"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("finding should mention %q, got:\n%v", want, err)
		}
	}
}

func TestSidebarAuthorshipAcceptsAuthorWrittenFromTheField(t *testing.T) {
	files := cleanTree()
	files["docs/constitutions/voice.yaml"] = voiceWithAuthorOnly
	files["03-what-is-an-agent.md"] = chapterWithSidebar("From the Field")

	root := gitTree(t, files, "Add the chapter")
	if err := audit(root, nil, nil); err != nil {
		t.Fatalf("a commit with no Skill trailer is the author's; got:\n%v", err)
	}
}

// Only types marked author_only are governed; the other three are fair game
// for an agent, which is why the marking lives in voice.yaml per type.
func TestSidebarAuthorshipIgnoresUnmarkedSidebarTypes(t *testing.T) {
	files := cleanTree()
	files["docs/constitutions/voice.yaml"] = voiceWithAuthorOnly
	files["03-what-is-an-agent.md"] = chapterWithSidebar("Common Error")

	root := gitTree(t, files, "Add the chapter\n\nSkill: do-work")
	if err := audit(root, nil, nil); err != nil {
		t.Fatalf("Common Error is not author_only; got:\n%v", err)
	}
}

// Without the marking the check is inert, so adopting it is opt-in per type.
func TestSidebarAuthorshipInertWithoutTheMarking(t *testing.T) {
	files := cleanTree()
	files["03-what-is-an-agent.md"] = chapterWithSidebar("From the Field")

	root := gitTree(t, files, "Add the chapter\n\nSkill: do-work")
	if err := audit(root, nil, nil); err != nil {
		t.Fatalf("unmarked voice.yaml should not govern authorship; got:\n%v", err)
	}
}

// A scratch export or a fresh unpack is not a repository. The check yields
// nothing there rather than failing, which is what keeps every other test in
// this file working on a plain directory.
func TestSidebarAuthorshipSilentOutsideAGitRepository(t *testing.T) {
	files := cleanTree()
	files["docs/constitutions/voice.yaml"] = voiceWithAuthorOnly
	files["03-what-is-an-agent.md"] = chapterWithSidebar("From the Field")

	if err := auditTree(t, files); err != nil {
		t.Fatalf("no repository means no blame and no finding; got:\n%v", err)
	}
}

// An uncommitted sidebar has no commit to judge, so it is not yet a finding.
func TestSidebarAuthorshipSkipsUncommittedLines(t *testing.T) {
	files := cleanTree()
	files["docs/constitutions/voice.yaml"] = voiceWithAuthorOnly

	root := gitTree(t, files, "Add the spec\n\nSkill: do-work")
	chapter := filepath.Join(root, "03-what-is-an-agent.md")
	if err := os.WriteFile(chapter, []byte(chapterWithSidebar("From the Field")), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := audit(root, nil, nil); err != nil {
		t.Fatalf("an uncommitted sidebar has no commit to judge; got:\n%v", err)
	}
}

// ------------------------------------------------ section numbering (GH-35)

// numberedChapter is a minimal chapter whose sections carry the given number.
func numberedChapter(id, title string, n int) string {
	return fmt.Sprintf(`<!-- chapter: %s -->

# %s

## Learning Objectives

1. Do the thing.

## %d.1 First

Prose that cites something [@hunt1999].

### %d.1.1 A subsection

More prose.

## Summary

Done.

## Key Terms

| Term | Definition |
|---|---|
| **Agent** | A state machine plus tools |
`, id, title, n, n)
}

// twoChapterTree adds a second chapter so positions can actually be wrong.
func twoChapterTree(firstNum, secondNum int) tree {
	files := cleanTree()
	files["03-what-is-an-agent.md"] = numberedChapter("C1.1", "What is an agent?", firstNum)
	files["04-second-chapter.md"] = numberedChapter("C1.2", "Second chapter", secondNum)
	return files
}

func TestSectionNumberingAcceptsBookWidePositions(t *testing.T) {
	if err := auditTree(t, twoChapterTree(1, 2)); err != nil {
		t.Fatalf("chapters 1 and 2 numbered 1.x and 2.x should pass, got:\n%v", err)
	}
}

// The defect GH-35 fixed: a second chapter numbered as though it restarted in
// a new part, which is what put two 2.1s in one PDF.
func TestSectionNumberingRejectsPerPartRestart(t *testing.T) {
	err := auditTree(t, twoChapterTree(1, 1))
	if err == nil {
		t.Fatal("a second chapter numbered 1.x should be a finding")
	}
	for _, want := range []string{"section_numbering", "numbered 1.x", "chapter 2"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("finding should mention %q, got:\n%v", want, err)
		}
	}
}

// The introduction is front matter. Counting it would shift every chapter by
// one, which is the bug this test pins.
func TestSectionNumberingDoesNotCountTheIntroduction(t *testing.T) {
	files := twoChapterTree(1, 2)
	files["01-introduction.md"] = `# Introduction

Front matter, not a chapter.
`
	if err := auditTree(t, files); err != nil {
		t.Fatalf("the introduction must not consume a chapter number, got:\n%v", err)
	}
}

// Part dividers and the references are not chapters either.
func TestSectionNumberingSkipsPartDividersAndReferences(t *testing.T) {
	files := twoChapterTree(1, 2)
	files["02-part-i.md"] = "# Part I -- Agents\n\nA divider.\n"
	files["05-references.md"] = "# References\n"
	if err := auditTree(t, files); err != nil {
		t.Fatalf("dividers and references must not consume chapter numbers, got:\n%v", err)
	}
}

// Figures share the chapter number, so a stale figure is caught even when the
// section headings are right.
func TestSectionNumberingCatchesAStaleFigureNumber(t *testing.T) {
	files := twoChapterTree(1, 2)
	files["04-second-chapter.md"] = strings.Replace(
		numberedChapter("C1.2", "Second chapter", 2),
		"## Summary",
		"**Figure 3.1** A caption.\n\n![](figures/fig-3-1-x.png)\n\nFigure 3.1 is referenced here.\n\n## Summary", 1)
	err := auditTree(t, files)
	if err == nil {
		t.Fatal("Figure 3.1 in chapter 2 should be a finding")
	}
	if !strings.Contains(err.Error(), "Figure 3.1 is numbered for chapter 3") {
		t.Errorf("finding should name the mismatch, got:\n%v", err)
	}
}

// One finding per wrong chapter number, not one per heading.
func TestSectionNumberingReportsOncePerChapter(t *testing.T) {
	err := auditTree(t, twoChapterTree(1, 1))
	if err == nil {
		t.Fatal("expected a finding")
	}
	if got := strings.Count(err.Error(), "voice.yaml: section_numbering"); got != 1 {
		t.Errorf("a chapter with several wrong headings should report once, got %d", got)
	}
}

// -------------------------------------------- chapter <-> SRD binding (GH-25)

func TestBindingRejectsAChapterWithNoMarker(t *testing.T) {
	files := cleanTree()
	files["03-what-is-an-agent.md"] = strings.Replace(
		files["03-what-is-an-agent.md"], "<!-- chapter: C1.1 -->\n\n", "", 1)
	err := auditTree(t, files)
	if err == nil {
		t.Fatal("an unmarked chapter should be a finding, not a silent pass")
	}
	if !strings.Contains(err.Error(), "carries no") {
		t.Errorf("finding should say the marker is missing, got:\n%v", err)
	}
}

// The old mechanism inferred the pairing from the chapter's title, so a
// retitle silently unpaired it. The marker has to survive that.
func TestBindingSurvivesARetitle(t *testing.T) {
	files := cleanTree()
	files["03-what-is-an-agent.md"] = strings.Replace(
		files["03-what-is-an-agent.md"], "# What is an agent?", "# Something Else Entirely", 1)
	if err := auditTree(t, files); err != nil {
		t.Fatalf("the binding is the marker, not the title, got:\n%v", err)
	}
}

// ...and a rename, which is what broke six times while Part I was drafted.
func TestBindingSurvivesARenumber(t *testing.T) {
	files := cleanTree()
	files["07-what-is-an-agent.md"] = files["03-what-is-an-agent.md"]
	delete(files, "03-what-is-an-agent.md")
	if err := auditTree(t, files); err != nil {
		t.Fatalf("the binding lives in the file, not its name, got:\n%v", err)
	}
}

func TestBindingRejectsAnUnknownChapterID(t *testing.T) {
	files := cleanTree()
	files["03-what-is-an-agent.md"] = strings.Replace(
		files["03-what-is-an-agent.md"], "chapter: C1.1", "chapter: C9.9", 1)
	err := auditTree(t, files)
	if err == nil || !strings.Contains(err.Error(), "ARCHITECTURE does not define") {
		t.Fatalf("an id absent from ARCHITECTURE should be a finding, got:\n%v", err)
	}
}

func TestBindingRejectsTwoChaptersClaimingOneID(t *testing.T) {
	files := cleanTree()
	files["04-second-chapter.md"] = files["03-what-is-an-agent.md"]
	err := auditTree(t, files)
	if err == nil || !strings.Contains(err.Error(), "already claimed by") {
		t.Fatalf("two files claiming C1.1 should be a finding, got:\n%v", err)
	}
}

// The contract half: apparatus.key_terms promises the Key Terms table.
func TestBindingCatchesAKeyTermMissingFromTheTable(t *testing.T) {
	files := cleanTree()
	files["03-what-is-an-agent.md"] = strings.Replace(
		files["03-what-is-an-agent.md"], "| **Agent** | A state machine plus tools |", "", 1)
	err := auditTree(t, files)
	if err == nil || !strings.Contains(err.Error(), `key term "agent"`) {
		t.Fatalf("a declared key term absent from the table should be a finding, got:\n%v", err)
	}
}

// An anchor citation is the source the SRD says carries the chapter.
func TestBindingCatchesAnUncitedAnchor(t *testing.T) {
	files := cleanTree()
	files["03-what-is-an-agent.md"] = strings.Replace(
		files["03-what-is-an-agent.md"], "[@hunt1999]", "with no citation", 1)
	err := auditTree(t, files)
	if err == nil || !strings.Contains(err.Error(), "anchor citation") {
		t.Fatalf("an uncited anchor should be a finding, got:\n%v", err)
	}
}

// Only anchors are required; a context or evidence citation may go unused.
func TestBindingIgnoresUncitedNonAnchors(t *testing.T) {
	files := cleanTree()
	files["docs/srd/srd-1.1-what-is-an-agent.yaml"] = strings.Replace(
		files["docs/srd/srd-1.1-what-is-an-agent.yaml"],
		"  - id: hunt1999\n    role: anchor",
		"  - id: hunt1999\n    role: anchor\n  - id: hunt1999\n    role: context", 1)
	if err := auditTree(t, files); err != nil {
		t.Fatalf("only anchors are required to appear, got:\n%v", err)
	}
}

// A chapter the road map calls drafted must have a file claiming it.
func TestBindingCatchesADraftedChapterWithNoFile(t *testing.T) {
	files := cleanTree()
	files["docs/road-map.yaml"] = strings.Replace(
		files["docs/road-map.yaml"], "      - id: C1.2\n        status: outline",
		"      - id: C1.2\n        status: drafted", 1)
	err := auditTree(t, files)
	if err == nil || !strings.Contains(err.Error(), "marked drafted but no chapter file claims it") {
		t.Fatalf("a drafted chapter with no file should be a finding, got:\n%v", err)
	}
}

// The leniency GH-50 removed: a term is "present" only if the table has a row
// defining it, not because it appears inside another term's definition cell.
func TestBindingIgnoresATermMentionedOnlyInsideADefinition(t *testing.T) {
	files := cleanTree()
	files["03-what-is-an-agent.md"] = strings.Replace(
		files["03-what-is-an-agent.md"],
		"| **Agent** | A state machine plus tools |",
		"| **Harness** | The software around a model, distinct from an agent |", 1)
	err := auditTree(t, files)
	if err == nil {
		t.Fatal(`"agent" appearing only inside another row's definition should still be a finding`)
	}
	if !strings.Contains(err.Error(), `key term "agent"`) {
		t.Errorf("finding should name the absent term, got:\n%v", err)
	}
}

// ...without flagging a table term that carries a parenthetical gloss, which
// is how the LLM chapter writes its first row.
func TestBindingAcceptsATableTermWithAGloss(t *testing.T) {
	files := cleanTree()
	files["03-what-is-an-agent.md"] = strings.Replace(
		files["03-what-is-an-agent.md"],
		"| **Agent** | A state machine plus tools |",
		"| **Agent (autonomous)** | A state machine plus tools |", 1)
	if err := auditTree(t, files); err != nil {
		t.Fatalf("a trailing gloss should not unmatch a key term, got:\n%v", err)
	}
}

// ------------------------------------------------ glossary coverage (GH-52)

// The reverse of the key_terms check: a term coined while drafting must reach
// definitions.yaml, whether or not the chapter has an SRD.
func TestBindingCatchesATableTermMissingFromTheGlossary(t *testing.T) {
	files := cleanTree()
	files["03-what-is-an-agent.md"] = strings.Replace(
		files["03-what-is-an-agent.md"],
		"| **Agent** | A state machine plus tools |",
		"| **Agent** | A state machine plus tools |\n| **Tracer bullet** | A thin end-to-end slice |", 1)
	err := auditTree(t, files)
	if err == nil {
		t.Fatal("a table term absent from definitions.yaml should be a finding")
	}
	if !strings.Contains(err.Error(), "not in docs/definitions.yaml") {
		t.Errorf("finding should name the glossary, got:\n%v", err)
	}
}

// It applies to chapters with no SRD too, which is where coined terms hide.
func TestGlossaryCheckAppliesWithoutAnSRD(t *testing.T) {
	files := cleanTree()
	files["04-second-chapter.md"] = `<!-- chapter: C1.2 -->

# Second chapter

## Learning Objectives

1. Do the thing.

## Summary

Done.

## Key Terms

| Term | Definition |
|---|---|
| **Undefined coinage** | Something nobody wrote down |
`
	err := auditTree(t, files)
	if err == nil || !strings.Contains(err.Error(), "undefined coinage") {
		t.Fatalf("C1.2 has no SRD but its table still needs the glossary, got:\n%v", err)
	}
}

// A glossary key and a table row may spell the separator differently. The
// keys use underscores, the tables hyphenate, and both name one term.
func TestTermMatchingIgnoresSeparatorStyle(t *testing.T) {
	for _, c := range []struct{ key, row string }{
		{"file_system_access", "File-system access"},
		{"file_system_access", "File system access"},
		{"fsm_constrained_decoding", "FSM-constrained decoding"},
	} {
		if normalizeTerm(c.key) != normalizeTerm(c.row) {
			t.Errorf("%q and %q should normalize alike, got %q and %q",
				c.key, c.row, normalizeTerm(c.key), normalizeTerm(c.row))
		}
	}
}
