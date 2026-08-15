// Copyright (c) 2026 Petar Djukic. All rights reserved.
// SPDX-License-Identifier: MIT

package agent

import (
	"os"
	"path/filepath"
)

// example:begin c1.6-1
// Start runs before Run: externalized state enters the transcript
// first, so the notes precede the task Run appends.
func (r *Runtime) Start() {
	notes, err := os.ReadFile(filepath.Join(r.root, r.profile.Memory.Notes))
	if err == nil {
		r.transcript = append(r.transcript, string(notes))
	}
}

// End writes what the next session is allowed to know.
func (r *Runtime) End(decision string) error {
	f, err := os.OpenFile(filepath.Join(r.root, r.profile.Memory.Notes),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(decision + "\n")
	return err
}

// example:end c1.6-1
