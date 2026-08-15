// Copyright (c) 2026 Petar Djukic. All rights reserved.
// SPDX-License-Identifier: MIT

package agent

import (
	"github.com/petar-djukic/agentic-coding-book/examples/parts/part-i/fixture"
)

// Canned is a Model that returns a fixed sequence of calls and then stops.
// Section 1.4 says the interface works identically whether the implementation
// asks a vendor API or reads a table; this is the table. It is the whole
// reason the demos need no model and no credentials.
type Canned struct {
	calls []Call
	at    int
}

func (m *Canned) Decide(transcript []string) (Call, bool) {
	if m.at >= len(m.calls) {
		return Call{}, true
	}
	c := m.calls[m.at]
	m.at++
	return c, false
}

// Demo runs the c1.6 runtime end to end in root, twice. The first session
// starts with no notes and writes one at the end; the second reads that note
// back before its task, which is the whole of what crosses the gap. It
// returns both transcripts in order.
func Demo(root string) ([]string, error) {
	if err := fixture.Write(root); err != nil {
		return nil, err
	}
	profile, err := Declared()
	if err != nil {
		return nil, err
	}

	first := New(profile, &Canned{calls: []Call{
		{Tool: "read_file", Args: map[string]string{"path": fixture.Task}},
		{Tool: "write_file", Args: map[string]string{
			"path": fixture.Missing, "content": fixture.Broken}},
		{Tool: "write_file", Args: map[string]string{
			"path": fixture.Missing, "content": fixture.Fixed}},
	}}, root)
	first.Start()
	transcript := first.Run("make the module compile")
	if err := first.End("greeting() returns a string, not an int"); err != nil {
		return nil, err
	}

	second := New(profile, &Canned{}, root)
	second.Start()
	return append(transcript, second.Run("continue where the last session left off")...), nil
}
