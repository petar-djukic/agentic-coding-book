// Copyright (c) 2026 Petar Djukic. All rights reserved.
// SPDX-License-Identifier: MIT

package agent

import (
	"strings"
	"testing"
)

// rewire returns the declared profile with one transition's destination
// changed, which is the only way to test that next reads the profile: if the
// runtime held the wiring in a switch, this edit would change nothing.
func rewire(t *testing.T, from, on, to string) Profile {
	t.Helper()
	p, err := Declared()
	if err != nil {
		t.Fatalf("Declared: %v", err)
	}
	for i := range p.Transitions {
		if p.Transitions[i].From == from && p.Transitions[i].On == on {
			p.Transitions[i].To = to
			return p
		}
	}
	t.Fatalf("no transition %s on %s", from, on)
	return p
}

func TestDeclaredProfileMatchesListing(t *testing.T) {
	p, err := Declared()
	if err != nil {
		t.Fatalf("Declared: %v", err)
	}
	if p.Agent != "executor" || p.Start != "deciding" {
		t.Errorf("agent %q start %q", p.Agent, p.Start)
	}
	if len(p.States) != 4 || len(p.Transitions) != 4 {
		t.Errorf("%d states, %d transitions", len(p.States), len(p.Transitions))
	}
	if len(p.Tools) != 1 || p.Tools[0].Name != "read_file" || p.Tools[0].Root != "." {
		t.Errorf("tools = %+v", p.Tools)
	}
}

func TestNextFollowsTheProfileNotASwitch(t *testing.T) {
	p := rewire(t, "executing", "result", "done")
	r := New(p, &Canned{calls: []Call{{Tool: "read_file"}}})
	r.state = "executing"
	r.next("result")
	if r.state != "done" {
		t.Fatalf("state = %q, want done -- the runtime is not reading the profile", r.state)
	}
}

func TestUndeclaredSignalStopsTheMachine(t *testing.T) {
	p, err := Declared()
	if err != nil {
		t.Fatal(err)
	}
	r := New(p, &Canned{})
	r.state = "deciding"
	r.next("no_such_signal")
	if r.state != "done" {
		t.Fatalf("state = %q, want done -- an undeclared signal must stop rather than guess", r.state)
	}
}

func TestUnknownToolBecomesTranscriptNotAnError(t *testing.T) {
	p, err := Declared()
	if err != nil {
		t.Fatal(err)
	}
	r := New(p, &Canned{calls: []Call{{Tool: "delete_everything"}}})
	transcript := r.Run("try it")
	joined := strings.Join(transcript, "\n")
	if !strings.Contains(joined, "no such tool: delete_everything") {
		t.Fatalf("transcript did not record the refusal:\n%s", joined)
	}
	if r.state != "done" {
		t.Errorf("state = %q, want done -- the loop must continue past a missing tool", r.state)
	}
}

func TestRunReturnsTranscriptStartingWithTheTask(t *testing.T) {
	dir := t.TempDir()
	if err := writeFixture(t, dir); err != nil {
		t.Fatal(err)
	}
	p, err := Declared()
	if err != nil {
		t.Fatal(err)
	}
	for i := range p.Tools {
		p.Tools[i].Root = dir
	}
	r := New(p, &Canned{calls: []Call{
		{Tool: "read_file", Args: map[string]string{"path": "TASK.md"}},
	}})
	transcript := r.Run("read the task")
	if transcript[0] != "read the task" {
		t.Errorf("transcript[0] = %q, want the task", transcript[0])
	}
	if len(transcript) != 2 || !strings.Contains(transcript[1], "greeting()") {
		t.Errorf("transcript = %q", transcript)
	}
	if r.state != "done" {
		t.Errorf("state = %q, want done", r.state)
	}
}

func TestVerifyingPassesThrough(t *testing.T) {
	p, err := Declared()
	if err != nil {
		t.Fatal(err)
	}
	r := New(p, &Canned{calls: []Call{{Tool: "read_file", Args: map[string]string{"path": "nope"}}}})
	transcript := r.Run("t")
	// Three entries and no verdict among them: this snapshot's verifying
	// state appends nothing, which is the gap section 4.6 closes.
	if len(transcript) != 2 {
		t.Fatalf("transcript = %q, want the task and one tool result", transcript)
	}
	for _, line := range transcript {
		if strings.HasPrefix(line, "build ") {
			t.Errorf("c1.1 must not verify anything, got %q", line)
		}
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
	if !strings.Contains(strings.Join(first, "\n"), "no such tool: list_files") {
		t.Errorf("demo should show a declaration the library does not know:\n%q", first)
	}
}
