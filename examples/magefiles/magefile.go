// Copyright (c) 2026 Petar Djukic. All rights reserved.
// SPDX-License-Identifier: MIT

// The build for examples/, run as `mage -d examples <target>`. The book's
// root build shells in here and reports the result as one finding; the
// checking code lives here rather than there, with one exception -- the
// listing-extraction check reads both the chapters and the part source, so it
// stays in the book's own magefiles (docs/ARCHITECTURE.yaml:
// listing_extraction).
package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

// Default is what `mage -d examples` runs with no target named. The audit is
// the cheap check and the one the book's root build dispatches to; `mage -d
// examples audit test` runs both.
var Default = Audit

// runIn runs a command with its working directory set, streaming output.
// mage's sh package has no directory option and every target here works
// inside a part module, so the build uses exec directly.
func runIn(dir, name string, args ...string) error {
	return streamIn(dir, os.Stdout, name, args...)
}

func streamIn(dir string, out io.Writer, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = out
	cmd.Stderr = out
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s %v in %s: %w", name, args, dir, err)
	}
	return nil
}

// outputIn runs a command and returns what it wrote, so a caller can compare
// two runs rather than only watch them.
func outputIn(dir, name string, args ...string) (string, error) {
	var buf bytes.Buffer
	err := streamIn(dir, &buf, name, args...)
	return buf.String(), err
}

// runnableParts returns the part entries whose artifact exists, printing a
// note for each one that does not. A part is listed in the manifest from the
// moment its chapters are planned, so skipping is the normal case rather than
// an error -- Parts II to V are registered and unwritten.
func runnableParts(root string, note io.Writer) ([]example, error) {
	m, err := loadManifest(root)
	if err != nil {
		return nil, err
	}
	var out []example
	for _, p := range m.entries(kindPart) {
		if p.Status != statusImplemented {
			fmt.Fprintf(note, "skipping %s: status %s\n", p.ID, p.Status)
			continue
		}
		dir := filepath.Join(root, p.Path)
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err != nil {
			return nil, fmt.Errorf("%s is %s but has no module at %s", p.ID, statusImplemented, p.Path)
		}
		out = append(out, p)
	}
	return out, nil
}
