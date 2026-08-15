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

// Demo runs the c1.4 runtime end to end in root, which must be an empty
// directory it may write into, and returns the transcript. The agent writes
// the missing file wrongly, fails the gate, writes it correctly, and passes:
// the compiler's verdict is what tells the second attempt from the first.
func Demo(root string) ([]string, error) {
	if err := fixture.Write(root); err != nil {
		return nil, err
	}
	profile, err := Declared()
	if err != nil {
		return nil, err
	}
	model := &Canned{calls: []Call{
		{Tool: "read_file", Args: map[string]string{"path": fixture.Task}},
		{Tool: "write_file", Args: map[string]string{
			"path": fixture.Missing, "content": fixture.Broken}},
		{Tool: "write_file", Args: map[string]string{
			"path": "../escape.go", "content": fixture.Fixed}},
		{Tool: "write_file", Args: map[string]string{
			"path": fixture.Missing, "content": fixture.Fixed}},
	}}
	return New(profile, model, root).Run("make the module compile"), nil
}
