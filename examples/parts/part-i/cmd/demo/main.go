// Copyright (c) 2026 Petar Djukic. All rights reserved.
// SPDX-License-Identifier: MIT

// Command demo runs every Part I snapshot end to end and prints its
// transcript. Each snapshot gets a fresh temporary directory, so a run
// writes nothing into the repository and two runs produce identical output.
//
// The entry point lives in the part module rather than in examples/magefiles
// because each snapshot's Demo returns that snapshot's own transcript type;
// the build discovers the part from the manifest and runs this, which keeps
// the build from having to import every part it may ever have to demonstrate.
package main

import (
	"fmt"
	"os"
	"strings"

	c11 "github.com/petar-djukic/agentic-coding-book/examples/parts/part-i/c1.1"
	c14 "github.com/petar-djukic/agentic-coding-book/examples/parts/part-i/c1.4"
	c16 "github.com/petar-djukic/agentic-coding-book/examples/parts/part-i/c1.6"
)

// snapshot pairs a chapter id with the demo it runs. The order is reading
// order, so the output shows the runtime gaining the write, the gate, and
// the memory in the order the book adds them.
var snapshots = []struct {
	chapter string
	shows   string
	run     func(root string) ([]string, error)
}{
	{"c1.1", "the skeleton: it reads, and a declaration the library does not know is refused", c11.Demo},
	{"c1.4", "the write and the gate: a wrong answer fails the compiler, an escaping path is refused", c14.Demo},
	{"c1.6", "memory: two sessions, and what the first wrote is what the second knows", c16.Demo},
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "demo:", err)
		os.Exit(1)
	}
}

func run() error {
	for _, s := range snapshots {
		dir, err := os.MkdirTemp("", "part-i-demo")
		if err != nil {
			return err
		}
		transcript, err := s.run(dir)
		os.RemoveAll(dir)
		if err != nil {
			return fmt.Errorf("%s: %w", s.chapter, err)
		}
		if len(transcript) == 0 {
			return fmt.Errorf("%s: the run produced no transcript", s.chapter)
		}

		fmt.Printf("=== %s -- %s\n", s.chapter, s.shows)
		for i, entry := range transcript {
			fmt.Printf("%3d | %s\n", i, strings.ReplaceAll(strings.TrimRight(entry, "\n"), "\n", "\n    | "))
		}
		fmt.Println()
	}
	return nil
}
