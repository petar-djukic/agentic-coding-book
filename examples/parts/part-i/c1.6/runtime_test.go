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

func TestPartOneEndsWithTheWholeMachine(t *testing.T) {
	p := declared(t)
	if len(p.Tools) != 2 {
		t.Errorf("tools = %+v, want the read and the write", p.Tools)
	}
	if p.Memory.Notes == "" {
		t.Error("no memory declared")
	}
	var pass, fail bool
	for _, tr := range p.Transitions {
		if tr.From == "verifying" && tr.On == "pass" {
			pass = true
		}
		if tr.From == "verifying" && tr.On == "fail" {
			fail = true
		}
	}
	if !pass || !fail {
		t.Error("both verdicts must be routed by the profile")
	}
}

func TestGateStillRoutesBothVerdicts(t *testing.T) {
	dir := repo(t)
	r := New(declared(t), &Canned{}, dir)
	if out, ok := r.verify(); ok {
		t.Errorf("the fixture does not compile until greeting.go exists, got %q", out)
	}
	if err := os.WriteFile(filepath.Join(dir, fixture.Missing), []byte(fixture.Fixed), 0o644); err != nil {
		t.Fatal(err)
	}
	if out, ok := r.verify(); !ok {
		t.Errorf("should pass once the file is written, got %q", out)
	}
}

func TestWriteIsStillRefusedOutsideTheRoot(t *testing.T) {
	if _, err := (WriteFile{root: t.TempDir()}).Run(map[string]string{
		"path": "../escape.go", "content": "x"}); err == nil {
		t.Error("the policy check must survive into the last snapshot")
	}
}

func TestDemoIsDeterministic(t *testing.T) {
	first, err := Demo(t.TempDir())
	if err != nil {
		t.Fatalf("Demo: %v", err)
	}
	second, err := Demo(t.TempDir())
	if err != nil {
		t.Fatalf("Demo: %v", err)
	}
	if strings.Join(first, "\n") != strings.Join(second, "\n") {
		t.Errorf("two runs differ:\n%q\n%q", first, second)
	}
	joined := strings.Join(first, "\n")
	for _, want := range []string{"build failed:", "build ok", "greeting() returns a string, not an int"} {
		if !strings.Contains(joined, want) {
			t.Errorf("demo transcript missing %q:\n%s", want, joined)
		}
	}
}
