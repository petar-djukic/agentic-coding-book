// Copyright (c) 2026 Petar Djukic. All rights reserved.
// SPDX-License-Identifier: MIT

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// bookArchYAML is the smallest architecture the audit can resolve chapters
// against: three chapters, matching the three Part I Build sections.
const bookArchYAML = `structure:
  parts:
    - id: P1
      chapters:
        - {id: C1.1, srd: docs/srd/srd-1.1-what-is-an-agent.yaml}
        - {id: C1.4, srd: docs/srd/srd-1.4-harness-touches-code.yaml}
        - {id: C1.6, srd: docs/srd/srd-1.6-externalizing-memory.yaml}
`

// fakeTree builds a book repository with an examples/ inside it and returns
// the examples path, so a check can be pointed at a shape the real tree does
// not have.
func fakeTree(t *testing.T, manifestBody string) string {
	t.Helper()
	book := t.TempDir()
	write(t, filepath.Join(book, "docs", "ARCHITECTURE.yaml"), bookArchYAML)
	for _, name := range []string{
		"srd-1.1-what-is-an-agent.yaml",
		"srd-1.4-harness-touches-code.yaml",
		"srd-1.6-externalizing-memory.yaml",
	} {
		write(t, filepath.Join(book, "docs", "srd", name), "meta: {}\n")
	}
	root := filepath.Join(book, "examples")
	write(t, filepath.Join(root, manifestFile), manifestBody)
	return root
}

func write(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// wants asserts that exactly the findings whose rule and detail substring are
// named were produced, and reports everything found when they were not.
func wants(t *testing.T, found []finding, rule, substring string) {
	t.Helper()
	for _, f := range found {
		if f.Rule == rule && strings.Contains(f.Detail, substring) {
			return
		}
	}
	t.Fatalf("no %s finding containing %q; got %v", rule, substring, found)
}

func wantsNone(t *testing.T, found []finding) {
	t.Helper()
	if len(found) != 0 {
		t.Fatalf("expected no findings, got %v", found)
	}
}

// ---------------------------------------------------------- the real tree

func TestAuditPassesOnTheRealTree(t *testing.T) {
	if err := audit(repoRoot); err != nil {
		t.Fatalf("audit: %v", err)
	}
}

func TestAuditRefusesToRunOutsideTheBook(t *testing.T) {
	err := audit(t.TempDir())
	if err == nil || !strings.Contains(err.Error(), "must run inside the book repository") {
		t.Fatalf("err = %v, want a clear refusal rather than a pile of missing chapters", err)
	}
}

// ------------------------------------------------------------------ E-C6

func TestChapterIDsMustResolve(t *testing.T) {
	book := t.TempDir()
	write(t, filepath.Join(book, "docs", "ARCHITECTURE.yaml"), bookArchYAML)
	m := &manifest{Examples: []example{{
		ID: "part-i", Kind: kindPart,
		Snapshots: []snapshot{{Chapter: "C9.9"}},
		Consumers: []string{"C8.8"},
	}}}
	found := checkChapterIDs(m, book)
	wants(t, found, "E-C6", "chapter C9.9")
	wants(t, found, "E-C6", "consumer C8.8")
}

func TestChapterIDsAcceptWhatArchitectureDefines(t *testing.T) {
	book := t.TempDir()
	write(t, filepath.Join(book, "docs", "ARCHITECTURE.yaml"), bookArchYAML)
	m := &manifest{Examples: []example{{
		ID: "part-i", Kind: kindPart,
		Snapshots: []snapshot{{Chapter: "C1.1"}, {Chapter: "C1.4"}},
		Consumers: []string{"C1.6"},
	}}}
	wantsNone(t, checkChapterIDs(m, book))
}

// ------------------------------------------------------------------ E-C7

const chapterWithBuild = `<!-- chapter: C1.4 -->

# How a harness touches your code

## 4.6 Build: The Write and the Gate

Prose.
`

func TestBuildSectionMustHaveASnapshot(t *testing.T) {
	book := t.TempDir()
	write(t, filepath.Join(book, "docs", "ARCHITECTURE.yaml"), bookArchYAML)
	write(t, filepath.Join(book, "06-how-a-harness-touches-your-code.md"), chapterWithBuild)

	m := &manifest{Examples: []example{{ID: "part-i", Kind: kindPart}}}
	wants(t, checkBuildSectionCoverage(m, book), "E-C7", "C1.4 has a Build section with no snapshot entry")

	covered := &manifest{Examples: []example{{
		ID: "part-i", Kind: kindPart, Snapshots: []snapshot{{Chapter: "C1.4"}},
	}}}
	wantsNone(t, checkBuildSectionCoverage(covered, book))
}

func TestBuildSectionMustCarryAChapterMarker(t *testing.T) {
	book := t.TempDir()
	write(t, filepath.Join(book, "docs", "ARCHITECTURE.yaml"), bookArchYAML)
	write(t, filepath.Join(book, "06-unmarked.md"), "# Untitled\n\n## 4.6 Build: something\n")
	wants(t, checkBuildSectionCoverage(&manifest{}, book), "E-C7", "no <!-- chapter: --> marker")
}

func TestAChapterWithoutABuildSectionNeedsNoSnapshot(t *testing.T) {
	book := t.TempDir()
	write(t, filepath.Join(book, "docs", "ARCHITECTURE.yaml"), bookArchYAML)
	write(t, filepath.Join(book, "04-the-agents-you-already-use.md"), "<!-- chapter: C1.2 -->\n\n## 2.1 Prose\n")
	wantsNone(t, checkBuildSectionCoverage(&manifest{}, book))
}

// ------------------------------------------------------------------ E-C4

func TestProvenanceMustBePinned(t *testing.T) {
	m := &manifest{Examples: []example{
		{ID: "no-block", Kind: kindCatalog},
		{ID: "half-filled", Kind: kindTools, Provenance: &provenance{
			Upstream: "Nokia-Bell-Labs/declarative-agents",
			Path:     "agent-core/tools/builtin",
			License:  "BSD-3-Clause",
			Holder:   "Nokia",
		}},
		{ID: "unpinned", Kind: kindCatalog, Provenance: &provenance{
			Upstream:   "Nokia-Bell-Labs/declarative-agents",
			Path:       "applications/catalog/agents/executor",
			Release:    "main",
			License:    "BSD-3-Clause",
			Holder:     "Nokia",
			Simplified: "dropped the variants",
		}},
	}}
	found := checkProvenance(m)
	wants(t, found, "E-C4", "no-block is copied material with no provenance block")
	wants(t, found, "E-C4", "half-filled provenance has no release")
	wants(t, found, "E-C4", "half-filled provenance has no simplified")
	wants(t, found, "E-C4", `unpinned release "main" does not look like an upstream tag`)
}

func TestAPartNeedsNoProvenance(t *testing.T) {
	wantsNone(t, checkProvenance(&manifest{Examples: []example{{ID: "part-i", Kind: kindPart}}}))
}

// ------------------------------------------------------------------ E-C5

func TestCatalogHeadersMustSurvive(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, catalogDir, "agents", "executor", "profile.yaml"),
		"# Copyright (c) 2026 Nokia\n# SPDX-License-Identifier: BSD-3-Clause\nname: executor\n")
	write(t, filepath.Join(root, catalogDir, "agents", "executor", "stripped.yaml"), "name: executor\n")
	write(t, filepath.Join(root, catalogDir, "README.md"), "# catalog\n\nThis repository's own prose.\n")

	found := checkCatalogHeaders(root)
	wants(t, found, "E-C5", "does not name Nokia")
	wants(t, found, "E-C5", "does not carry the BSD-3-Clause identifier")
	for _, f := range found {
		if strings.Contains(f.File, "README.md") {
			t.Error("the directory's own README is not copied material")
		}
		if strings.Contains(f.File, "profile.yaml") {
			t.Errorf("a file with both notices must pass: %v", f)
		}
	}
}

func TestNoCatalogDirectoryIsNotAFinding(t *testing.T) {
	wantsNone(t, checkCatalogHeaders(t.TempDir()))
}

// ------------------------------------------------------------------ E-C3

func TestListingsMustNotExtractFromTheCatalog(t *testing.T) {
	root := t.TempDir()
	m := &manifest{Examples: []example{{
		ID: "part-i", Kind: kindPart, Path: ".",
		Snapshots: []snapshot{{Chapter: "C1.1", Listings: []listing{{
			ID: "c1.1-1", Regions: []region{{File: "catalog/agents/executor/profile.yaml", Marker: "c1.1-1"}},
		}}}},
	}}}
	wants(t, checkListingRegions(root, m), "E-C3", "listings resolve into parts/ only")
}

func TestListingsMustNameARegionThatExists(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "parts", "part-i", "c1.1", "runtime.go"),
		"package agent\n\n// example:begin c1.1-2\ntype Profile struct{}\n// example:end c1.1-2\n")

	base := example{ID: "part-i", Kind: kindPart, Path: filepath.Join("parts", "part-i")}

	missingFile := base
	missingFile.Snapshots = []snapshot{{Listings: []listing{{
		ID: "c1.1-2", Regions: []region{{File: "c1.1/absent.go", Marker: "c1.1-2"}},
	}}}}
	wants(t, checkListingRegions(root, &manifest{Examples: []example{missingFile}}),
		"E-C3", "cannot be read")

	missingMarker := base
	missingMarker.Snapshots = []snapshot{{Listings: []listing{{
		ID: "c1.1-2", Regions: []region{{File: "c1.1/runtime.go", Marker: "c1.1-9"}},
	}}}}
	wants(t, checkListingRegions(root, &manifest{Examples: []example{missingMarker}}),
		"E-C3", "which the file does not delimit")

	noRegion := base
	noRegion.Snapshots = []snapshot{{Listings: []listing{{ID: "c1.1-2"}}}}
	wants(t, checkListingRegions(root, &manifest{Examples: []example{noRegion}}),
		"E-C3", "names no region")

	good := base
	good.Snapshots = []snapshot{{Listings: []listing{{
		ID: "c1.1-2", Regions: []region{{File: "c1.1/runtime.go", Marker: "c1.1-2"}},
	}}}}
	wantsNone(t, checkListingRegions(root, &manifest{Examples: []example{good}}))
}

func TestARegionMustBeOneSpan(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "parts", "part-i", "c1.1", "runtime.go"),
		"package agent\n\n// example:begin dup\nA\n// example:end dup\n\n// example:begin dup\nB\n// example:end dup\n")
	m := &manifest{Examples: []example{{
		ID: "part-i", Kind: kindPart, Path: filepath.Join("parts", "part-i"),
		Snapshots: []snapshot{{Listings: []listing{{
			ID: "dup", Regions: []region{{File: "c1.1/runtime.go", Marker: "dup"}},
		}}}},
	}}}
	wants(t, checkListingRegions(root, m), "E-C3", "a region is one span")
}

func TestMarkerOfStripsTheHostComment(t *testing.T) {
	for line, want := range map[string]string{
		"// example:begin c1.1-2":    "begin c1.1-2",
		"# example:end c1.1-1":       "end c1.1-1",
		"\t\t// example:end c1.4-2b": "end c1.4-2b",
		"type Profile struct {":      "",
		"// ordinary comment":        "",
	} {
		if got := markerOf(line); got != want {
			t.Errorf("markerOf(%q) = %q, want %q", line, got, want)
		}
	}
}

// ------------------------------------------------------------------ E-C1

func TestPartsMustNotRequireUpstream(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, "parts", "part-i", "go.mod"),
		"module example\n\ngo 1.26\n\nrequire github.com/Nokia-Bell-Labs/declarative-agents/applications/catalog v0.20260814.4\n")
	m := &manifest{Examples: []example{{
		ID: "part-i", Kind: kindPart, Status: statusImplemented, Path: filepath.Join("parts", "part-i"),
	}}}
	wants(t, checkUpstreamIndependence(root, m), "E-C1", "parts compile standalone")
}

func TestAnImplementedPartNeedsAModule(t *testing.T) {
	m := &manifest{Examples: []example{{
		ID: "part-i", Kind: kindPart, Status: statusImplemented, Path: filepath.Join("parts", "part-i"),
	}}}
	wants(t, checkUpstreamIndependence(t.TempDir(), m), "E-C1", "has no module file")
}

func TestAPlannedPartNeedsNoModule(t *testing.T) {
	m := &manifest{Examples: []example{{
		ID: "part-ii", Kind: kindPart, Status: statusPlanned, Path: filepath.Join("parts", "part-ii"),
	}}}
	wantsNone(t, checkUpstreamIndependence(t.TempDir(), m))
}

// ------------------------------------------------------------------ E-C2

// snapshotPair writes two snapshots differing in the files named, and returns
// the root they were written under.
func snapshotPair(t *testing.T, differing map[string]string) string {
	t.Helper()
	root := t.TempDir()
	base := filepath.Join(root, "parts", "part-i")
	for _, snap := range []string{"c1.1", "c1.4"} {
		write(t, filepath.Join(base, snap, "runtime.go"), "package agent\n")
		write(t, filepath.Join(base, snap, "profile.yaml"), "agent: executor\n")
		write(t, filepath.Join(base, snap, "runtime_test.go"), "package agent\n// "+snap+"\n")
		write(t, filepath.Join(base, snap, "demo.go"), "package agent\n// "+snap+"\n")
	}
	for name, body := range differing {
		write(t, filepath.Join(base, "c1.4", name), body)
	}
	return root
}

func partWithSnapshots(adds []string) *manifest {
	return &manifest{Examples: []example{{
		ID: "part-i", Kind: kindPart, Path: filepath.Join("parts", "part-i"),
		Snapshots: []snapshot{
			{Chapter: "C1.1", Path: "c1.1"},
			{Chapter: "C1.4", Path: "c1.4", Adds: adds},
		},
	}}}
}

func TestUndeclaredSnapshotDriftIsAFinding(t *testing.T) {
	root := snapshotPair(t, map[string]string{"runtime.go": "package agent\n\nfunc verify() {}\n"})
	wants(t, checkSnapshotDiffs(root, partWithSnapshots(nil)),
		"E-C2", "runtime.go differs from c1.1 but MANIFEST.yaml does not declare it")
}

func TestDeclaredSnapshotDriftIsAccepted(t *testing.T) {
	root := snapshotPair(t, map[string]string{"runtime.go": "package agent\n\nfunc verify() {}\n"})
	wantsNone(t, checkSnapshotDiffs(root, partWithSnapshots(
		[]string{"runtime.go: the verify gate replacing the pass-through case"})))
}

func TestANewFileCountsAsDrift(t *testing.T) {
	root := snapshotPair(t, map[string]string{"tools.go": "package agent\n\ntype WriteFile struct{}\n"})
	wants(t, checkSnapshotDiffs(root, partWithSnapshots(nil)), "E-C2", "tools.go differs")
}

func TestSupportFilesAreOutsideTheDiffRule(t *testing.T) {
	// The snapshots differ only in a test file and a demo, which every pair
	// of snapshots does by construction.
	root := snapshotPair(t, nil)
	wantsNone(t, checkSnapshotDiffs(root, partWithSnapshots(nil)))
}

func TestDeclaringSomethingThatDidNotChangeIsAFinding(t *testing.T) {
	root := snapshotPair(t, nil)
	wants(t, checkSnapshotDiffs(root, partWithSnapshots(
		[]string{"memory.go: Start and End"})),
		"E-C2", "declares memory.go as added at c1.4, but it does not differ")
}

func TestSupportFileClassification(t *testing.T) {
	for path, want := range map[string]bool{
		"runtime_test.go":                       true,
		"demo.go":                               true,
		filepath.Join("fixture", "fixture.go"):  true,
		"runtime.go":                            false,
		"profile.yaml":                          false,
		filepath.Join("cmd", "demo", "main.go"): false,
	} {
		if got := supportFile(path); got != want {
			t.Errorf("supportFile(%q) = %v, want %v", path, got, want)
		}
	}
}

// ------------------------------------------------------------------ E-C8

func TestRealizesMustResolve(t *testing.T) {
	root := fakeTree(t, "schema_version: 1\nexamples: []\n")
	book := filepath.Join(root, "..")
	write(t, filepath.Join(root, "docs", "srd", "srd-part-i.yaml"),
		"realizes: [srd-1.1, srd-9.9]\n")
	m := &manifest{Examples: []example{{
		ID: "part-i", Kind: kindPart, SRD: filepath.Join("docs", "srd", "srd-part-i.yaml"),
	}}}
	wants(t, checkRealizes(root, m, book), "E-C8", "realizes srd-9.9, which resolves to no file")
}

func TestRealizesMustNotBeEmpty(t *testing.T) {
	root := fakeTree(t, "schema_version: 1\nexamples: []\n")
	write(t, filepath.Join(root, "docs", "srd", "srd-part-i.yaml"), "meta: {part: part-i}\n")
	m := &manifest{Examples: []example{{
		ID: "part-i", Kind: kindPart, SRD: filepath.Join("docs", "srd", "srd-part-i.yaml"),
	}}}
	wants(t, checkRealizes(root, m, filepath.Join(root, "..")), "E-C8", "realizes: is empty")
}

func TestAmbiguousRealizesIsAFinding(t *testing.T) {
	root := fakeTree(t, "schema_version: 1\nexamples: []\n")
	book := filepath.Join(root, "..")
	write(t, filepath.Join(book, "docs", "srd", "srd-1.1-duplicate.yaml"), "meta: {}\n")
	write(t, filepath.Join(root, "docs", "srd", "srd-part-i.yaml"), "realizes: [srd-1.1]\n")
	m := &manifest{Examples: []example{{
		ID: "part-i", Kind: kindPart, SRD: filepath.Join("docs", "srd", "srd-part-i.yaml"),
	}}}
	wants(t, checkRealizes(root, m, book), "E-C8", "resolves to 2 files")
}

func TestAPartMustNameAnSRD(t *testing.T) {
	root := fakeTree(t, "schema_version: 1\nexamples: []\n")
	m := &manifest{Examples: []example{{ID: "part-i", Kind: kindPart}}}
	wants(t, checkRealizes(root, m, filepath.Join(root, "..")), "E-C8", "names no SRD")
}

// --------------------------------------------------------------- baseline

func TestBaselineAcceptsAMatchingFinding(t *testing.T) {
	found := []finding{{File: "a.yaml", Rule: "E-C4", Detail: "x is unpinned"}}
	accepted := map[string]string{found[0].key(): "GH-1"}
	if err := reportFindings(found, accepted); err != nil {
		t.Fatalf("an accepted finding must not fail the audit: %v", err)
	}
}

func TestBaselineDoesNotAcceptAnythingElse(t *testing.T) {
	found := []finding{{File: "a.yaml", Rule: "E-C4", Detail: "x is unpinned"}}
	accepted := map[string]string{"a.yaml|E-C4|y is unpinned": "GH-1"}
	err := reportFindings(found, accepted)
	if err == nil || !strings.Contains(err.Error(), "x is unpinned") {
		t.Fatalf("err = %v, want the unmatched finding to block", err)
	}
}

func TestAStaleBaselineEntryIsItselfAFinding(t *testing.T) {
	accepted := map[string]string{"gone.yaml|E-C4|fixed long ago": "GH-1"}
	err := reportFindings(nil, accepted)
	if err == nil || !strings.Contains(err.Error(), "no longer occurs") {
		t.Fatalf("err = %v, want the stale entry reported", err)
	}
}

func TestABaselineEntryMustNameAnIssue(t *testing.T) {
	root := t.TempDir()
	write(t, filepath.Join(root, baselineFile),
		"accepted:\n  - {file: a.yaml, rule: E-C4, detail: x, issue: \"\"}\n")
	_, findings := loadBaseline(root)
	wants(t, findings, "baseline", "has no issue")
}

func TestNoBaselineFileIsNotAFinding(t *testing.T) {
	_, findings := loadBaseline(t.TempDir())
	wantsNone(t, findings)
}

// ------------------------------------------------------- accumulation

func TestFindingsAccumulateAndReportTogether(t *testing.T) {
	root := fakeTree(t, `schema_version: 1
examples:
  - id: part-i
    kind: part
    status: implemented
    path: parts/part-i
    snapshots:
      - {chapter: C9.9, path: c9.9}
  - id: executor
    kind: catalog-family
    status: implemented
    path: catalog/agents/executor
`)
	err := audit(root)
	if err == nil {
		t.Fatal("expected findings")
	}
	for _, want := range []string{"E-C6", "E-C4", "E-C1", "E-C8"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the audit reported no %s finding; it must fail once with everything it found:\n%v", want, err)
		}
	}
}
