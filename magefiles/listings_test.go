// Copyright (c) 2026 Petar Djukic. All rights reserved.
// SPDX-License-Identifier: MIT

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// listingTree writes a book with one chapter and one registered listing, and
// returns the root. chapterBody and sourceBody are what the two sides say, so
// a test can make them agree or disagree.
func listingTree(t *testing.T, chapterBody, sourceBody string) string {
	t.Helper()
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "06-how-a-harness-touches-your-code.md"), chapterBody)
	writeFile(t, filepath.Join(root, examplesDir, "MANIFEST.yaml"), `schema_version: 1
examples:
  - id: part-i
    kind: part
    path: parts/part-i
    snapshots:
      - chapter: C1.4
        path: c1.4
        listings:
          - id: c1.4-1
            label: "Listing 4.1"
            language: go
            regions:
              - {file: c1.4/tools.go, marker: c1.4-1}
`)
	writeFile(t, filepath.Join(root, examplesDir, "parts", "part-i", "c1.4", "tools.go"), sourceBody)
	return root
}

func writeFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

const listingChapter = "<!-- chapter: C1.4 -->\n\n# How a harness touches your code\n\n" +
	"## 4.6 Build: The Write and the Gate\n\n" +
	"**Listing 4.1** Edit application.\n\n" +
	"```go\ntype WriteFile struct{ root string }\n```\n\n" +
	"*The refusal happens in the harness.*\n"

const listingSource = "package agent\n\n" +
	"// example:begin c1.4-1\ntype WriteFile struct{ root string }\n// example:end c1.4-1\n"

func listingFindings(t *testing.T, root string) []finding {
	t.Helper()
	return checkListings(root)
}

func TestListingMatchingItsRegionIsNotAFinding(t *testing.T) {
	root := listingTree(t, listingChapter, listingSource)
	if found := listingFindings(t, root); len(found) != 0 {
		t.Fatalf("expected no findings, got %v", found)
	}
}

func TestListingDriftIsAFinding(t *testing.T) {
	drifted := strings.Replace(listingSource,
		"type WriteFile struct{ root string }",
		"type WriteFile struct{ base string }", 1)
	root := listingTree(t, listingChapter, drifted)

	found := listingFindings(t, root)
	if len(found) != 1 {
		t.Fatalf("expected one finding, got %v", found)
	}
	for _, want := range []string{"Listing 4.1 does not match", "the code is the authority", "line 1"} {
		if !strings.Contains(found[0].Detail, want) {
			t.Errorf("finding does not say %q: %s", want, found[0].Detail)
		}
	}
}

func TestAMissingRegionIsAFinding(t *testing.T) {
	root := listingTree(t, listingChapter, "package agent\n\ntype WriteFile struct{ root string }\n")
	found := listingFindings(t, root)
	if len(found) == 0 || !strings.Contains(found[0].Detail, "not delimited by begin and end markers") {
		t.Fatalf("got %v, want a complaint about the absent markers", found)
	}
}

func TestAnUnregisteredListingIsAFinding(t *testing.T) {
	extra := listingChapter + "\n**Listing 4.9** Something nobody registered.\n\n```go\nfunc x() {}\n```\n\n*Gloss.*\n"
	root := listingTree(t, extra, listingSource)

	found := listingFindings(t, root)
	if len(found) != 1 || !strings.Contains(found[0].Detail, "Listing 4.9 (C1.4) is not registered") {
		t.Fatalf("got %v, want the unregistered listing reported", found)
	}
}

func TestARegisteredListingNoChapterPrintsIsAFinding(t *testing.T) {
	withoutListing := "<!-- chapter: C1.4 -->\n\n# How a harness touches your code\n\n## 4.6 Build: nothing here\n"
	root := listingTree(t, withoutListing, listingSource)

	found := listingFindings(t, root)
	if len(found) != 1 || !strings.Contains(found[0].Detail, "which that chapter does not print") {
		t.Fatalf("got %v, want the orphaned registration reported", found)
	}
}

func TestListingsAreComparedByteForByte(t *testing.T) {
	// One leading space on the printed line. A whitespace-tolerant compare
	// would pass this, and indentation is what makes a Go listing readable.
	reindented := strings.Replace(listingChapter,
		"```go\ntype WriteFile struct{ root string }\n```",
		"```go\n type WriteFile struct{ root string }\n```", 1)
	root := listingTree(t, reindented, listingSource)

	if found := listingFindings(t, root); len(found) != 1 {
		t.Fatalf("a one-space indentation change must be a finding, got %v", found)
	}
}

func TestNoExamplesDirectoryIsNotAFinding(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "06-chapter.md"), listingChapter)
	if found := checkListings(root); len(found) != 0 {
		t.Fatalf("the check is inert without examples/, got %v", found)
	}
	if found := checkExamplesBuild(root); len(found) != 0 {
		t.Fatalf("the dispatch is inert without examples/, got %v", found)
	}
}

func TestMultiRegionListingsConcatenateInOrder(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "06-chapter.md"),
		"<!-- chapter: C1.4 -->\n\n# Chapter\n\n## 4.6 Build: x\n\n"+
			"**Listing 4.2** Two places at once.\n\n```go\nfirst\nsecond\n```\n\n*Gloss.*\n")
	writeFile(t, filepath.Join(root, examplesDir, "MANIFEST.yaml"), `schema_version: 1
examples:
  - id: part-i
    kind: part
    path: parts/part-i
    snapshots:
      - chapter: C1.4
        path: c1.4
        listings:
          - id: c1.4-2
            label: "Listing 4.2"
            language: go
            regions:
              - {file: c1.4/a.go, marker: one}
              - {file: c1.4/b.go, marker: two}
`)
	writeFile(t, filepath.Join(root, examplesDir, "parts", "part-i", "c1.4", "a.go"),
		"package agent\n\n// example:begin one\nfirst\n// example:end one\n")
	writeFile(t, filepath.Join(root, examplesDir, "parts", "part-i", "c1.4", "b.go"),
		"package agent\n\n// example:begin two\nsecond\n// example:end two\n")

	if found := checkListings(root); len(found) != 0 {
		t.Fatalf("regions must concatenate in the order given, got %v", found)
	}
}

func TestExtractRegionExcludesTheMarkers(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "runtime.go")
	writeFile(t, path, "package agent\n\n// example:begin r\nbody line\n// example:end r\n\nfunc after() {}\n")

	got, err := extractRegion(path, "r")
	if err != nil {
		t.Fatal(err)
	}
	if got != "body line" {
		t.Errorf("extractRegion = %q, want the body without its delimiters", got)
	}
}

func TestMarkerDirectiveStripsTheHostComment(t *testing.T) {
	for line, want := range map[string]string{
		"// example:begin c1.1-2": "begin c1.1-2",
		"# example:end c1.1-1":    "end c1.1-1",
		"\t\t// example:begin x":  "begin x",
		"type Profile struct {":   "",
		"// an ordinary comment":  "",
		"//example:begin tight":   "begin tight",
	} {
		if got := markerDirective(line); got != want {
			t.Errorf("markerDirective(%q) = %q, want %q", line, got, want)
		}
	}
}

// ------------------------------------------------------ unnumbered fences

// buildTree writes a chapter whose Build section holds whatever fences the
// test supplies, plus a manifest snapshot body, and returns the root.
func buildTree(t *testing.T, sectionBody, manifestSnapshot string) string {
	t.Helper()
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "06-chapter.md"),
		"<!-- chapter: C1.4 -->\n\n# Chapter\n\n## 4.6 Build: The Write and the Gate\n\n"+
			sectionBody+"\n## Summary\n\nAfterwards.\n")
	writeFile(t, filepath.Join(root, examplesDir, "MANIFEST.yaml"),
		"schema_version: 1\nexamples:\n  - id: part-i\n    kind: part\n    path: parts/part-i\n"+
			"    snapshots:\n      - chapter: C1.4\n        path: c1.4\n"+manifestSnapshot)
	return root
}

const bareFence = "```yaml\nmemory:\n  notes: NOTES.md\n```\n"

func TestAFenceWithNoLabelAndNoMarkerIsAFinding(t *testing.T) {
	root := buildTree(t, bareFence, "        listings: []\n")
	found := checkListings(root)
	if len(found) != 1 || !strings.Contains(found[0].Detail, "carries no **Listing N.M** label and no") {
		t.Fatalf("got %v, want the unannounced fence reported", found)
	}
}

func TestAMarkedSnippetExtracts(t *testing.T) {
	root := buildTree(t,
		"<!-- snippet: c1.4-p1 -->\n\n"+bareFence,
		"        snippets:\n          - id: c1.4-p1\n            regions:\n"+
			"              - {file: c1.4/profile.yaml, marker: c1.4-p1}\n")
	writeFile(t, filepath.Join(root, examplesDir, "parts", "part-i", "c1.4", "profile.yaml"),
		"agent: executor\n# example:begin c1.4-p1\nmemory:\n  notes: NOTES.md\n# example:end c1.4-p1\n")

	if found := checkListings(root); len(found) != 0 {
		t.Fatalf("a marked snippet matching its region must pass, got %v", found)
	}
}

func TestAMarkedSnippetDriftsLikeAListing(t *testing.T) {
	root := buildTree(t,
		"<!-- snippet: c1.4-p1 -->\n\n"+bareFence,
		"        snippets:\n          - id: c1.4-p1\n            regions:\n"+
			"              - {file: c1.4/profile.yaml, marker: c1.4-p1}\n")
	writeFile(t, filepath.Join(root, examplesDir, "parts", "part-i", "c1.4", "profile.yaml"),
		"agent: executor\n# example:begin c1.4-p1\nmemory:\n  notes: MEMORY.md\n# example:end c1.4-p1\n")

	found := checkListings(root)
	if len(found) != 1 || !strings.Contains(found[0].Detail, "snippet c1.4-p1 does not match") {
		t.Fatalf("got %v, want the snippet drift reported", found)
	}
}

func TestADeclaredProseSnippetIsNotCompared(t *testing.T) {
	root := buildTree(t,
		"<!-- snippet: c1.4-p1 -->\n\n"+bareFence,
		"        snippets:\n          - id: c1.4-p1\n            prose: \"an abridgement the source cannot carry\"\n")

	if found := checkListings(root); len(found) != 0 {
		t.Fatalf("declared prose is unchecked on purpose, got %v", found)
	}
}

func TestAProseDeclarationMustNotAlsoNameRegions(t *testing.T) {
	root := buildTree(t,
		"<!-- snippet: c1.4-p1 -->\n\n"+bareFence,
		"        snippets:\n          - id: c1.4-p1\n            prose: \"a reason\"\n            regions:\n"+
			"              - {file: c1.4/profile.yaml, marker: c1.4-p1}\n")
	found := checkListings(root)
	if len(found) == 0 || !strings.Contains(found[0].Detail, "declares prose and names regions") {
		t.Fatalf("got %v, want the contradiction reported", found)
	}
}

func TestASnippetMustSaySomething(t *testing.T) {
	root := buildTree(t,
		"<!-- snippet: c1.4-p1 -->\n\n"+bareFence,
		"        snippets:\n          - id: c1.4-p1\n            language: yaml\n")
	found := checkListings(root)
	if len(found) == 0 || !strings.Contains(found[0].Detail, "names no regions and declares no prose reason") {
		t.Fatalf("got %v, want the empty registration reported", found)
	}
}

func TestAnUnregisteredSnippetIsAFinding(t *testing.T) {
	root := buildTree(t, "<!-- snippet: c1.4-p9 -->\n\n"+bareFence, "        listings: []\n")
	found := checkListings(root)
	if len(found) != 1 || !strings.Contains(found[0].Detail, "snippet c1.4-p9 (C1.4) is not registered") {
		t.Fatalf("got %v, want the unregistered snippet reported", found)
	}
}

func TestFencesOutsideBuildSectionsAreNotGoverned(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "06-chapter.md"),
		"<!-- chapter: C1.4 -->\n\n# Chapter\n\n## 4.2 Assembling the Context\n\n"+
			"```go\nfunc illustrative() {}\n```\n\n"+
			"## 4.6 Build: x\n\nNo fences here.\n\n## Summary\n\nDone.\n")
	writeFile(t, filepath.Join(root, examplesDir, "MANIFEST.yaml"), "schema_version: 1\nexamples: []\n")

	if found := checkListings(root); len(found) != 0 {
		t.Fatalf("only Build-section fences are governed, got %v", found)
	}
}

func TestBuildSectionFencesReadsWhatAnnouncedEach(t *testing.T) {
	doc := "<!-- chapter: C1.4 -->\n\n## 4.2 Prose\n\n```go\noutside\n```\n\n" +
		"## 4.6 Build: x\n\n" +
		"**Listing 4.1** Labelled.\n\n```go\nfirst\n```\n\n*Gloss.*\n\n" +
		"<!-- snippet: p1 -->\n\n```yaml\nsecond\n```\n\n" +
		"```yaml\nthird\n```\n\n" +
		"## Summary\n\n```go\nafter\n```\n"

	got := buildSectionFences(doc)
	if len(got) != 3 {
		t.Fatalf("got %d fences, want the three inside the Build section: %+v", len(got), got)
	}
	for i, want := range []fence{
		{label: "Listing 4.1", body: "first"},
		{snippet: "p1", body: "second"},
		{body: "third"},
	} {
		if got[i].label != want.label || got[i].snippet != want.snippet || got[i].body != want.body {
			t.Errorf("fence %d = %+v, want label=%q snippet=%q body=%q",
				i, got[i], want.label, want.snippet, want.body)
		}
	}
}

func TestALabelIsConsumedByOneFenceOnly(t *testing.T) {
	doc := "## 4.6 Build: x\n\n**Listing 4.1** One.\n\n```go\nfirst\n```\n\n```go\nsecond\n```\n"
	got := buildSectionFences(doc)
	if len(got) != 2 {
		t.Fatalf("got %d fences, want 2", len(got))
	}
	if got[0].label != "Listing 4.1" {
		t.Errorf("first fence label = %q", got[0].label)
	}
	if got[1].label != "" || got[1].snippet != "" {
		t.Errorf("the label must not carry to the next fence: %+v", got[1])
	}
}

// ------------------------------------------------- the book's own listings

func TestEveryPartOneListingExtracts(t *testing.T) {
	if found := checkListings(".."); len(found) != 0 {
		t.Fatalf("the book's listings must extract from examples/parts/:\n%v", found)
	}
}

func TestTheExamplesBuildIsGreen(t *testing.T) {
	if testing.Short() {
		t.Skip("runs the examples build")
	}
	if found := checkExamplesBuild(".."); len(found) != 0 {
		t.Fatalf("mage -d examples audit test must pass:\n%v", found)
	}
}
