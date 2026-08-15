// Copyright (c) 2026 Petar Djukic. All rights reserved.
// SPDX-License-Identifier: MIT

package agent

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/petar-djukic/agentic-coding-book/examples/parts/part-i/fixture"
)

func TestWriteFileRefusesPathsOutsideTheRoot(t *testing.T) {
	for _, path := range []string{"/etc/passwd", "..", "../escape.go", "../../escape.go"} {
		t.Run(path, func(t *testing.T) {
			// The root is a directory inside the temp dir, so an escaping
			// write has somewhere real to land if the check lets it through.
			base := t.TempDir()
			root := filepath.Join(base, "repo")
			if err := os.Mkdir(root, 0o755); err != nil {
				t.Fatal(err)
			}
			before := tree(t, base)

			out, err := WriteFile{root: root}.Run(map[string]string{
				"path": path, "content": "package main\n"})
			if err == nil {
				t.Fatalf("write of %q was allowed, returned %q", path, out)
			}
			if !strings.Contains(err.Error(), "outside the root") {
				t.Errorf("err = %v, want a refusal naming the root", err)
			}
			if after := tree(t, base); after != before {
				t.Errorf("the refused write still touched the disk:\n%s\n%s", before, after)
			}
		})
	}
}

func TestWriteFileWritesUnderTheRoot(t *testing.T) {
	root := t.TempDir()
	out, err := WriteFile{root: root}.Run(map[string]string{
		"path": "sub/../greeting.go", "content": fixture.Fixed})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if out != "wrote greeting.go" {
		t.Errorf("out = %q, want the cleaned path", out)
	}
	got, err := os.ReadFile(filepath.Join(root, "greeting.go"))
	if err != nil || string(got) != fixture.Fixed {
		t.Errorf("file = %q, err = %v", got, err)
	}
}

func TestLibraryResolvesBothTools(t *testing.T) {
	if newTool("read_file", ".") == nil || newTool("write_file", ".") == nil {
		t.Error("c1.4 declares read_file and write_file; both must resolve")
	}
	if newTool("delete_file", ".") != nil {
		t.Error("an undeclared name must not resolve")
	}
}

// tree renders every path under dir with its size, so a test can assert that
// a refused write changed nothing anywhere.
func tree(t *testing.T, dir string) string {
	t.Helper()
	var b strings.Builder
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		b.WriteString(path)
		if !info.IsDir() {
			b.WriteString(" " + info.Mode().String())
		}
		b.WriteString("\n")
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return b.String()
}
