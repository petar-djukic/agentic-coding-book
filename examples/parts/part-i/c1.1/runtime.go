// Copyright (c) 2026 Petar Djukic. All rights reserved.
// SPDX-License-Identifier: MIT

// Package agent is the runtime as section 1.5 leaves it: a profile that
// declares the machine, and an interpreter that performs each state. Nothing
// writes, nothing checks the work, and nothing survives the process. Sections
// 4.6 and 6.6 close those gaps in the c1.4 and c1.6 snapshots.
//
// Listings 1.1 and 1.2 extract from this package.
package agent

// example:begin c1.1-2
type Profile struct {
	Agent       string
	States      []string
	Start       string
	Transitions []struct{ From, On, To string }
	Tools       []struct{ Name, Root string }
}

// Tool is the only path from the process to anything outside it.
// The runtime turns each declaration in the profile's tools list
// into one of these from its library.
type Tool interface {
	Name() string
	Run(args map[string]string) (string, error)
}

// Model supplies one signal per step: the next call, or done.
type Model interface {
	Decide(transcript []string) (call Call, done bool)
}

type Call struct {
	Tool string
	Args map[string]string
}

type Runtime struct {
	profile    Profile
	state      string
	transcript []string
	pending    Call
	tools      map[string]Tool
	model      Model
}

// next consults the profile, not a switch: the wiring is data.
func (r *Runtime) next(signal string) {
	for _, t := range r.profile.Transitions {
		if t.From == r.state && t.On == signal {
			r.state = t.To
			return
		}
	}
	r.state = "done" // no declared transition: stop rather than guess
}

func (r *Runtime) Run(task string) []string {
	r.transcript = append(r.transcript, task)
	r.state = r.profile.Start
	for r.state != "done" {
		switch r.state {
		case "deciding":
			call, done := r.model.Decide(r.transcript)
			if done {
				r.next("done")
				break
			}
			r.pending = call
			r.next("tool_call")
		case "executing":
			r.transcript = append(r.transcript, r.execute(r.pending))
			r.next("result")
		case "verifying":
			// Nothing checks the work yet; Chapter 4's Build section
			// declares the gate this state is reserved for.
			r.next("pass")
		}
	}
	return r.transcript
}

func (r *Runtime) execute(c Call) string {
	tool, ok := r.tools[c.Tool]
	if !ok {
		return "no such tool: " + c.Tool
	}
	result, err := tool.Run(c.Args)
	if err != nil {
		return "error: " + err.Error()
	}
	return result
}

// example:end c1.1-2

// New builds a runtime from a profile. Section 1.5 says the runtime turns
// each declaration in the profile's tools list into a Tool from its library
// without showing where that happens; this is where. The listing stops at
// the line above because the book's subject is the machine, not its wiring.
func New(profile Profile, model Model) *Runtime {
	r := &Runtime{profile: profile, model: model, tools: map[string]Tool{}}
	for _, d := range profile.Tools {
		if t := newTool(d.Name, d.Root); t != nil {
			r.tools[t.Name()] = t
		}
	}
	return r
}
