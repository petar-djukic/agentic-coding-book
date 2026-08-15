// Copyright (c) 2026 Petar Djukic. All rights reserved.
// SPDX-License-Identifier: MIT

package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestDemoRunsPartOneEndToEnd(t *testing.T) {
	out, err := demoOutput(repoRoot, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("demoOutput: %v", err)
	}
	for _, want := range []string{
		"=== c1.1", "=== c1.4", "=== c1.6",
		"no such tool",   // c1.1: a declaration the library does not know
		"write refused:", // c1.4: the policy check, in the transcript
		"build failed:",  // c1.4: the compiler's verdict re-entering context
		"build ok",       // c1.4: and the verdict that ends the loop
	} {
		if !strings.Contains(out, want) {
			t.Errorf("demo output missing %q", want)
		}
	}
	if strings.Contains(out, os.TempDir()) {
		t.Error("the transcript names a temporary path, so two runs cannot match")
	}
}

func TestDemoIsDeterministic(t *testing.T) {
	first, err := demoOutput(repoRoot, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("demoOutput: %v", err)
	}
	second, err := demoOutput(repoRoot, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("demoOutput: %v", err)
	}
	if first != second {
		t.Errorf("two runs differ:\n--- first\n%s\n--- second\n%s", first, second)
	}
}

func TestDemoLeavesTheTreeClean(t *testing.T) {
	before := gitStatus(t)
	if _, err := demoOutput(repoRoot, &bytes.Buffer{}); err != nil {
		t.Fatalf("demoOutput: %v", err)
	}
	if after := gitStatus(t); after != before {
		t.Errorf("the demo wrote into the repository:\n%s", diffLines(before, after))
	}
}

func TestDemoRefusesAPartWithNoEntryPoint(t *testing.T) {
	root := fakeRoot(t, `schema_version: 1
examples:
  - {id: part-i, kind: part, status: implemented, path: parts/part-i}
`)
	dir := filepath.Join(root, "parts", "part-i")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := demoOutput(root, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), demoCommand) {
		t.Fatalf("err = %v, want a complaint naming %s", err, demoCommand)
	}
}

func gitStatus(t *testing.T) string {
	t.Helper()
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = repoRoot
	out, err := cmd.Output()
	if err != nil {
		t.Skipf("git unavailable: %v", err)
	}
	return string(out)
}

func diffLines(before, after string) string {
	was := map[string]bool{}
	for _, l := range strings.Split(before, "\n") {
		was[l] = true
	}
	var added []string
	for _, l := range strings.Split(after, "\n") {
		if !was[l] {
			added = append(added, l)
		}
	}
	return strings.Join(added, "\n")
}
