// Copyright (c) 2026 Petar Djukic. All rights reserved.
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// demoCommand is the entry point each part module ships. The build runs it
// rather than importing the part, because a part's demo returns that part's
// own transcript types and the build should not have to know them.
const demoCommand = "./cmd/demo"

// Demo runs each part end to end on its canned fixture and prints the
// transcripts. No model, no credentials, no network: every run drives the
// runtime with a table of responses behind the Model interface, in a
// temporary directory the part creates and removes. Two runs produce
// identical output (docs/ARCHITECTURE.yaml: E-C9).
func Demo() error {
	out, err := demoOutput(".", os.Stdout)
	if err != nil {
		return err
	}
	fmt.Print(out)
	return nil
}

// demoOutput returns what the demos printed, so a caller can compare two runs
// rather than only watch them. Progress notes go to note; the returned string
// is the transcripts alone.
func demoOutput(root string, note io.Writer) (string, error) {
	parts, err := runnableParts(root, note)
	if err != nil {
		return "", err
	}
	if len(parts) == 0 {
		return "", fmt.Errorf("no part is %s: nothing to demonstrate", statusImplemented)
	}
	var b strings.Builder
	for _, p := range parts {
		dir := filepath.Join(root, p.Path)
		if _, err := os.Stat(filepath.Join(dir, "cmd", "demo")); err != nil {
			return "", fmt.Errorf("%s ships no %s: %w", p.ID, demoCommand, err)
		}
		fmt.Fprintf(note, "%s in %s\n", demoCommand, p.Path)
		out, err := outputIn(dir, "go", "run", demoCommand)
		if err != nil {
			return "", err
		}
		b.WriteString(out)
	}
	return b.String(), nil
}
