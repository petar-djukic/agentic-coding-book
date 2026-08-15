// Copyright (c) 2026 Petar Djukic. All rights reserved.
// SPDX-License-Identifier: MIT

package agent

import (
	_ "embed"
	"os"

	"gopkg.in/yaml.v3"
)

//go:embed profile.yaml
var profileSource []byte

// Load reads a profile from disk. The decoder matches YAML keys to struct
// fields by lowercasing the field name, so Listing 1.2's Profile decodes
// Listing 1.1's file with no struct tags and nothing else in between -- which
// is what section 1.5 means when it says a workflow change is an edit to the
// profile and a recompile of nothing.
func Load(path string) (Profile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Profile{}, err
	}
	return parse(data)
}

// Declared returns the profile this snapshot ships, embedded at build time so
// a test or a demo finds it without knowing where it was run from.
func Declared() (Profile, error) { return parse(profileSource) }

func parse(data []byte) (Profile, error) {
	var p Profile
	if err := yaml.Unmarshal(data, &p); err != nil {
		return Profile{}, err
	}
	return p, nil
}
