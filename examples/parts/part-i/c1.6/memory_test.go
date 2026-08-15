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

func TestProfileNamesTheNotesFile(t *testing.T) {
	if got := declared(t).Memory.Notes; got != fixture.Notes {
		t.Errorf("memory.notes = %q, want %q", got, fixture.Notes)
	}
}

func TestStartOnAMissingNotesFileIsNotAnEvent(t *testing.T) {
	r := New(declared(t), &Canned{}, repo(t))
	r.Start()
	if len(r.transcript) != 0 {
		t.Errorf("transcript = %q, want empty -- the first session has no notes", r.transcript)
	}
}

func TestStartPutsTheNotesBeforeTheTask(t *testing.T) {
	dir := repo(t)
	if err := os.WriteFile(filepath.Join(dir, fixture.Notes), []byte("earlier session said this\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	r := New(declared(t), &Canned{}, dir)
	r.Start()
	transcript := r.Run("this session's task")
	if len(transcript) < 2 {
		t.Fatalf("transcript = %q", transcript)
	}
	if !strings.Contains(transcript[0], "earlier session said this") {
		t.Errorf("transcript[0] = %q, want the notes", transcript[0])
	}
	if transcript[1] != "this session's task" {
		t.Errorf("transcript[1] = %q, want the task after the notes", transcript[1])
	}
}

func TestEndCreatesTheFileAndThenAppends(t *testing.T) {
	dir := repo(t)
	r := New(declared(t), &Canned{}, dir)
	if err := r.End("first decision"); err != nil {
		t.Fatalf("End: %v", err)
	}
	if err := r.End("second decision"); err != nil {
		t.Fatalf("End: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(dir, fixture.Notes))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "first decision\nsecond decision\n" {
		t.Errorf("notes = %q -- the second End truncated rather than appended", got)
	}
}

func TestOnlyWhatPassesThroughEndCrossesTheGap(t *testing.T) {
	dir := repo(t)
	first := New(declared(t), &Canned{calls: []Call{
		{Tool: "read_file", Args: map[string]string{"path": fixture.Task}},
	}}, dir)
	first.Start()
	first.Run("session one")
	if err := first.End("the decision worth carrying"); err != nil {
		t.Fatal(err)
	}

	second := New(declared(t), &Canned{}, dir)
	second.Start()
	transcript := second.Run("session two")

	joined := strings.Join(transcript, "\n")
	if !strings.Contains(joined, "the decision worth carrying") {
		t.Errorf("the decision did not cross the gap:\n%s", joined)
	}
	if strings.Contains(joined, "session one") {
		t.Errorf("the first transcript should have died with the process:\n%s", joined)
	}
}
