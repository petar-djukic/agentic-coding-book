// Copyright (c) 2026 Petar Djukic. All rights reserved.
// SPDX-License-Identifier: MIT

// Package fixture writes the repository the demos and tests run against: a
// Go module that does not compile until one named file exists, and a task
// file describing what to do about it.
//
// The repository is written rather than copied from testdata so a run needs
// no path arithmetic, works from any directory, and leaves nothing outside
// the directory it is given. Nothing here is book code; no listing extracts
// from this package.
package fixture

import (
	"os"
	"path/filepath"
)

// Names the runtime and the demos refer to.
const (
	// Task is the file the agent reads to find out what it is doing.
	Task = "TASK.md"
	// Missing is the file whose absence keeps the module from compiling.
	Missing = "greeting.go"
	// Notes is the memory file the section 6.6 profile names.
	Notes = "NOTES.md"
)

// Broken is a first attempt that compiles the way a plausible wrong answer
// does: it parses, and the type checker rejects it. The demos write it once
// so the verification gate has something real to fail on.
const Broken = `package main

func greeting() string {
	return 42
}
`

// Fixed is the same file, correct. Writing it makes the module build.
const Fixed = `package main

func greeting() string {
	return "hello"
}
`

// Write creates the fixture module in dir, which must already exist.
func Write(dir string) error {
	files := map[string]string{
		"go.mod": "module fixture\n\ngo 1.26\n",
		"main.go": `package main

func main() {
	println(greeting())
}
`,
		Task: "Add " + Missing + " so that the module compiles. " +
			"greeting() returns a string.\n",
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			return err
		}
	}
	return nil
}
