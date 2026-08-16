// Copyright (c) 2026 Petar Djukic. All rights reserved.
// SPDX-License-Identifier: MIT

package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// repoRoot is examples/, one level up from this package.
const repoRoot = ".."

// fakeRoot writes a manifest into a temporary directory and returns its path,
// so a test can exercise a shape the real tree does not have yet.
func fakeRoot(t *testing.T, body string) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, manifestFile), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestLoadManifestReadsTheRealFile(t *testing.T) {
	m, err := loadManifest(repoRoot)
	if err != nil {
		t.Fatalf("loadManifest: %v", err)
	}
	if m.SchemaVersion != 1 {
		t.Errorf("schema_version = %d", m.SchemaVersion)
	}
	if m.ThirdParty.License != "BSD-3-Clause" {
		t.Errorf("third_party.license = %q", m.ThirdParty.License)
	}
	parts := m.entries(kindPart)
	if len(parts) != 1 || parts[0].ID != "part-i" {
		t.Fatalf("parts = %+v", parts)
	}
	if n := len(parts[0].Snapshots); n != 3 {
		t.Errorf("part-i has %d snapshots, want one per Build section", n)
	}
	// Six listings: C1.1 prints two, C1.4 three since GH-138 split the gate
	// from the state that routes it, C1.6 one. Two snippets: the section 4.6
	// profile abridgement, declared prose, and the section 6.6 memory block,
	// which extracts (GH-139).
	var listings, snippets, prose int
	for _, s := range parts[0].Snapshots {
		listings += len(s.Listings)
		snippets += len(s.Snippets)
		for _, sn := range s.Snippets {
			if sn.Prose != "" {
				prose++
			}
		}
	}
	if listings != 6 {
		t.Errorf("%d listings registered, want 6", listings)
	}
	if snippets != 2 || prose != 1 {
		t.Errorf("%d snippets registered (%d declared prose), want 2 and 1", snippets, prose)
	}
	families := m.entries(kindCatalog)
	if len(families) != 1 || families[0].ID != "executor" {
		t.Errorf("catalog families = %+v, want executor alone", families)
	}
	tools := m.entries(kindTools)
	if len(tools) != 1 || tools[0].ID != "filesystem-tools" {
		t.Errorf("catalog tool sets = %+v, want filesystem-tools alone", tools)
	}
}

func TestEveryCopiedEntryIsPinned(t *testing.T) {
	m, err := loadManifest(repoRoot)
	if err != nil {
		t.Fatal(err)
	}
	for _, kind := range copiedKinds {
		for _, e := range m.entries(kind) {
			p := e.Provenance
			if p == nil {
				t.Errorf("%s carries no provenance", e.ID)
				continue
			}
			if p.Upstream == "" || p.Path == "" || p.Release == "" {
				t.Errorf("%s provenance = %+v, want upstream, path, and a release", e.ID, p)
			}
			if p.Simplified == "" {
				t.Errorf("%s does not say what the copy dropped", e.ID)
			}
			if p.License != "BSD-3-Clause" || p.Holder != "Nokia" {
				t.Errorf("%s license = %q holder = %q", e.ID, p.License, p.Holder)
			}
		}
	}
}

func TestLoadManifestRejectsAnUnknownSchema(t *testing.T) {
	root := fakeRoot(t, "schema_version: 99\nexamples: []\n")
	_, err := loadManifest(root)
	if err == nil || !strings.Contains(err.Error(), "schema_version 99") {
		t.Fatalf("err = %v, want a refusal naming the version", err)
	}
}

func TestLoadManifestReportsAMissingFile(t *testing.T) {
	if _, err := loadManifest(t.TempDir()); err == nil {
		t.Fatal("a missing manifest must be an error, not an empty build")
	}
}

func TestRunnablePartsSkipsWhatIsNotImplemented(t *testing.T) {
	root := fakeRoot(t, `schema_version: 1
examples:
  - {id: part-i,  kind: part, status: implemented, path: parts/part-i}
  - {id: part-ii, kind: part, status: planned,     path: parts/part-ii}
  - {id: executor, kind: catalog-family, status: implemented, path: catalog/agents/executor}
`)
	if err := os.MkdirAll(filepath.Join(root, "parts", "part-i"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "parts", "part-i", "go.mod"), []byte("module x\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var note bytes.Buffer
	parts, err := runnableParts(root, &note)
	if err != nil {
		t.Fatalf("runnableParts: %v", err)
	}
	if len(parts) != 1 || parts[0].ID != "part-i" {
		t.Fatalf("parts = %+v, want part-i alone", parts)
	}
	if !strings.Contains(note.String(), "skipping part-ii: status planned") {
		t.Errorf("the skip must be printed, not silent; note = %q", note.String())
	}
	if strings.Contains(note.String(), "executor") {
		t.Error("a catalog family is not a part and must not be reported as skipped")
	}
}

func TestRunnablePartsRejectsAnImplementedPartWithNoModule(t *testing.T) {
	root := fakeRoot(t, `schema_version: 1
examples:
  - {id: part-i, kind: part, status: implemented, path: parts/part-i}
`)
	_, err := runnableParts(root, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "no module") {
		t.Fatalf("err = %v, want a complaint about the missing module", err)
	}
}
