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

func TestTestRunsThePartModule(t *testing.T) {
	if err := test(repoRoot); err != nil {
		t.Fatalf("test: %v", err)
	}
}

func TestTestFailsWhenNothingIsRunnable(t *testing.T) {
	root := fakeRoot(t, `schema_version: 1
examples:
  - {id: part-ii, kind: part, status: planned, path: parts/part-ii}
`)
	err := test(root)
	if err == nil || !strings.Contains(err.Error(), "nothing to test") {
		t.Fatalf("err = %v, want a refusal rather than a silent pass", err)
	}
}

func TestTestReportsAFailingPart(t *testing.T) {
	root := fakeRoot(t, `schema_version: 1
examples:
  - {id: part-x, kind: part, status: implemented, path: parts/part-x}
`)
	dir := filepath.Join(root, "parts", "part-x")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	write := func(name, body string) {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("go.mod", "module partx\n\ngo 1.26\n")
	write("x.go", "package partx\n\nfunc Answer() int { return 41 }\n")
	write("x_test.go", `package partx

import "testing"

func TestAnswer(t *testing.T) {
	if Answer() != 42 {
		t.Fatal("wrong")
	}
}
`)
	if err := test(root); err == nil {
		t.Fatal("a failing part test must fail the target")
	}
}

// The manifest registers a part once its Build sections are drafted, so the
// real tree has one part and reports no skips. Parts II to V are tracked in
// docs/road-map.yaml until then; the skip path is exercised against a fake
// root in TestRunnablePartsSkipsWhatIsNotImplemented.
func TestTheRealTreeReportsNoSkips(t *testing.T) {
	var note bytes.Buffer
	parts, err := runnableParts(repoRoot, &note)
	if err != nil {
		t.Fatal(err)
	}
	if len(parts) != 1 {
		t.Errorf("runnable parts = %d, want part-i alone", len(parts))
	}
	if note.String() != "" {
		t.Errorf("unexpected skip: %q", note.String())
	}
}
