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
	var listings int
	for _, s := range parts[0].Snapshots {
		listings += len(s.Listings)
	}
	if listings != 5 {
		t.Errorf("%d listings registered, want 5", listings)
	}
	if families := m.entries(kindCatalog); len(families) != 2 {
		t.Errorf("catalog families = %d, want executor and applier", len(families))
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
