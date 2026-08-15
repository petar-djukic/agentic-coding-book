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

// Demo runs the c1.1 runtime end to end in root, which must be an empty
// directory it may write into, and returns the transcript. The agent reads
// the task file and stops: this snapshot has no way to write, so reading is
// the whole of what it can do.
func Demo(root string) ([]string, error) {
	if err := fixture.Write(root); err != nil {
		return nil, err
	}
	profile, err := Declared()
	if err != nil {
		return nil, err
	}
	for i := range profile.Tools {
		profile.Tools[i].Root = root
	}
	model := &Canned{calls: []Call{
		{Tool: "read_file", Args: map[string]string{"path": fixture.Task}},
		{Tool: "list_files", Args: map[string]string{"path": "."}},
	}}
	return New(profile, model).Run("read the task"), nil
}
