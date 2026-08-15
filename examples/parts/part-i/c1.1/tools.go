// Copyright (c) 2026 Petar Djukic. All rights reserved.
// SPDX-License-Identifier: MIT

package agent

import (
	"os"
	"path/filepath"
)

// newTool is the library the profile's declarations are resolved against. A
// name the library does not know yields no tool, so the capability is absent
// and execute reports it as "no such tool" in the transcript.
func newTool(name, root string) Tool {
	switch name {
	case "read_file":
		return ReadFile{root: root}
	}
	return nil
}

// ReadFile implements the read_file declaration. It carries no path check:
// section 4.6 adds one to the write and says the read needs the same test,
// leaving it where the chapter's argument leaves it.
type ReadFile struct{ root string }

func (t ReadFile) Name() string { return "read_file" }

func (t ReadFile) Run(args map[string]string) (string, error) {
	data, err := os.ReadFile(filepath.Join(t.root, filepath.Clean(args["path"])))
	if err != nil {
		return "", err
	}
	return string(data), nil
}
