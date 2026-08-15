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

func repo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := fixture.Write(dir); err != nil {
		t.Fatal(err)
	}
	return dir
}

func declared(t *testing.T) Profile {
	t.Helper()
	p, err := Declared()
	if err != nil {
		t.Fatalf("Declared: %v", err)
	}
	return p
}

func TestDeclaredProfileAddsTheWriteAndTheFailRoute(t *testing.T) {
	p := declared(t)
	var names []string
	for _, d := range p.Tools {
		names = append(names, d.Name)
	}
	if strings.Join(names, ",") != "read_file,write_file" {
		t.Errorf("tools = %v", names)
	}
	var fail bool
	for _, tr := range p.Transitions {
		if tr.From == "verifying" && tr.On == "fail" {
			fail = true
			if tr.To != "deciding" {
				t.Errorf("fail routes to %q; section 4.6 sends both verdicts to deciding", tr.To)
			}
		}
	}
	if !fail {
		t.Error("no fail transition out of verifying")
	}
}

func TestFailingGateAppendsTheBuildOutputAndRoutesFail(t *testing.T) {
	dir := repo(t)
	if err := os.WriteFile(filepath.Join(dir, fixture.Missing), []byte(fixture.Broken), 0o644); err != nil {
		t.Fatal(err)
	}
	r := New(declared(t), &Canned{}, dir)
	out, ok := r.verify()
	if ok {
		t.Fatal("a module that does not type-check must not pass the gate")
	}
	if !strings.HasPrefix(out, "build failed:") || !strings.Contains(out, "greeting.go") {
		t.Errorf("verdict = %q, want the compiler's own output", out)
	}

	r.state = "verifying"
	r.next("fail")
	if r.state != "deciding" {
		t.Errorf("fail routed to %q, want deciding", r.state)
	}
}

func TestPassingGateAppendsAndRoutesPass(t *testing.T) {
	dir := repo(t)
	if err := os.WriteFile(filepath.Join(dir, fixture.Missing), []byte(fixture.Fixed), 0o644); err != nil {
		t.Fatal(err)
	}
	r := New(declared(t), &Canned{}, dir)
	out, ok := r.verify()
	if !ok {
		t.Fatalf("a module that compiles must pass the gate, got %q", out)
	}
	if out != "build ok" {
		t.Errorf("verdict = %q", out)
	}
}

func TestTheDifferenceTravelsInTheTranscript(t *testing.T) {
	dir := repo(t)
	r := New(declared(t), &Canned{calls: []Call{
		{Tool: "write_file", Args: map[string]string{
			"path": fixture.Missing, "content": fixture.Broken}},
		{Tool: "write_file", Args: map[string]string{
			"path": fixture.Missing, "content": fixture.Fixed}},
	}}, dir)
	transcript := r.Run("make it compile")

	joined := strings.Join(transcript, "\n")
	if !strings.Contains(joined, "build failed:") {
		t.Errorf("first attempt should have failed the gate:\n%s", joined)
	}
	if !strings.Contains(joined, "build ok") {
		t.Errorf("second attempt should have passed:\n%s", joined)
	}
	if strings.Index(joined, "build failed:") > strings.Index(joined, "build ok") {
		t.Error("the failure must precede the pass -- the gate is what tells them apart")
	}
}

func TestRootIsSharedWithTheTools(t *testing.T) {
	dir := repo(t)
	r := New(declared(t), &Canned{calls: []Call{
		{Tool: "write_file", Args: map[string]string{
			"path": fixture.Missing, "content": fixture.Fixed}},
	}}, dir)
	r.Run("write it")
	if _, err := os.Stat(filepath.Join(dir, fixture.Missing)); err != nil {
		t.Fatalf("the tool did not write under the runtime's root: %v", err)
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
	for _, want := range []string{"build failed:", "write refused:", "build ok"} {
		if !strings.Contains(joined, want) {
			t.Errorf("demo transcript missing %q:\n%s", want, joined)
		}
	}
}
