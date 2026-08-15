// Copyright (c) 2026 Petar Djukic. All rights reserved.
// SPDX-License-Identifier: MIT

package agent

import (
	"testing"

	"github.com/petar-djukic/agentic-coding-book/examples/parts/part-i/fixture"
)

func writeFixture(t *testing.T, dir string) error {
	t.Helper()
	return fixture.Write(dir)
}

func TestReadFileReadsUnderItsRoot(t *testing.T) {
	dir := t.TempDir()
	if err := writeFixture(t, dir); err != nil {
		t.Fatal(err)
	}
	out, err := ReadFile{root: dir}.Run(map[string]string{"path": fixture.Task})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if out == "" {
		t.Error("read returned nothing")
	}
}

func TestReadFileReportsAMissingFile(t *testing.T) {
	if _, err := (ReadFile{root: t.TempDir()}).Run(map[string]string{"path": "absent"}); err == nil {
		t.Error("reading a missing file should return an error the runtime turns into transcript")
	}
}

func TestLibraryResolvesOnlyWhatItKnows(t *testing.T) {
	if newTool("read_file", ".") == nil {
		t.Error("read_file should resolve")
	}
	if newTool("write_file", ".") != nil {
		t.Error("write_file must not resolve in c1.1 -- section 4.6 adds it")
	}
}
