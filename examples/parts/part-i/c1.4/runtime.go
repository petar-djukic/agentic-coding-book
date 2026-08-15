// Copyright (c) 2026 Petar Djukic. All rights reserved.
// SPDX-License-Identifier: MIT

// Package agent is the runtime as section 4.6 leaves it: the c1.1 skeleton
// with a write that sits behind a policy check and a verifying state that
// runs the compiler and routes its verdict. It still forgets everything when
// the process exits; section 6.6 closes that in the c1.6 snapshot.
//
// Listings 4.1 and 4.2 extract from this package.
package agent

import (
	"os/exec"
	"path/filepath"
)

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
	root       string
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
		// example:begin c1.4-2b
		case "verifying":
			out, ok := r.verify()
			r.transcript = append(r.transcript, out)
			if ok {
				r.next("pass")
			} else {
				r.next("fail")
			}
			// example:end c1.4-2b
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

// example:begin c1.4-2a
func (r *Runtime) verify() (string, bool) {
	cmd := exec.Command("go", "build", "./...")
	cmd.Dir = r.root
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "build failed:\n" + string(out), false
	}
	return "build ok", true
}

// example:end c1.4-2a

// New builds a runtime from a profile and the repository root it works in.
// The root is the field section 4.6 says the runtime gains: it is shared with
// every tool built from the profile's declarations, which is how a tool's
// path check knows what it is checking against.
func New(profile Profile, model Model, root string) *Runtime {
	r := &Runtime{profile: profile, model: model, tools: map[string]Tool{}, root: root}
	for _, d := range profile.Tools {
		if t := newTool(d.Name, filepath.Join(root, d.Root)); t != nil {
			r.tools[t.Name()] = t
		}
	}
	return r
}
