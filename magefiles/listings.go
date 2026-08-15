// Copyright (c) 2026 Petar Djukic. All rights reserved.
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// The code-to-prose direction (examples/docs/VISION.yaml: A1). A listing in a
// Build section is an extracted region of examples/parts/ source, never
// retyped, and this is where the two are compared. It is the one check that
// straddles the boundary -- it reads both the chapters and the part source --
// so it lives here rather than in examples/magefiles with the rest.
//
// Where the compiled source and the printed draft disagree, the prose is
// corrected. The code is the authority, the same way references.yaml is the
// authority for a citation.

const examplesDir = "examples"

// listingRe finds a listing and the fenced block beneath it: the bold label
// the voice constitution requires (V-S7), then the fence with its language
// tag. The label is what binds the fence to a manifest entry.
var listingRe = regexp.MustCompile("(?s)\\*\\*(Listing [0-9]+\\.[0-9]+)\\*\\*.*?\n```[a-z]*\n(.*?)\n```")

// exampleManifest is the part of examples/MANIFEST.yaml this check reads.
// examples/magefiles owns the full schema; duplicating four structs beats a
// dependency between two build trees, and the audit there checks the fields
// this one does not.
type exampleManifest struct {
	Examples []struct {
		ID        string `yaml:"id"`
		Kind      string `yaml:"kind"`
		Path      string `yaml:"path"`
		Snapshots []struct {
			Chapter  string `yaml:"chapter"`
			Listings []struct {
				ID      string `yaml:"id"`
				Label   string `yaml:"label"`
				Regions []struct {
					File   string `yaml:"file"`
					Marker string `yaml:"marker"`
				} `yaml:"regions"`
			} `yaml:"listings"`
		} `yaml:"snapshots"`
	} `yaml:"examples"`
}

// registeredListing is one label's resolved source, keyed by chapter id so a
// label repeated across chapters -- "Listing 4.1" exists once per part in a
// book numbered by section -- stays distinct.
type registeredListing struct {
	id      string
	chapter string
	label   string
	source  string
	origin  string
}

// checkListings compares every listing printed in a drafted Build section
// against the region it is registered to extract from. A mismatch, an
// unregistered listing, and a registered listing no chapter prints are each
// findings: the manifest and the chapters must agree in both directions.
func checkListings(root string) []finding {
	registered, findings := loadRegisteredListings(root)
	if registered == nil {
		return findings
	}

	printed := map[string]bool{}
	for _, name := range chapterFiles(root) {
		data, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			findings = append(findings, finding{File: name, Rule: "listing extraction", Detail: err.Error()})
			continue
		}
		chapter := ""
		if m := chapterMarkerRe.FindSubmatch(data); m != nil {
			chapter = string(m[1])
		}
		for _, match := range listingRe.FindAllStringSubmatch(string(data), -1) {
			label, fence := match[1], match[2]
			key := chapter + "|" + label
			printed[key] = true

			reg, ok := registered[key]
			if !ok {
				findings = append(findings, finding{File: name, Rule: "listing extraction",
					Detail: fmt.Sprintf("%s (%s) is not registered in %s/MANIFEST.yaml, so nothing checks it",
						label, chapter, examplesDir)})
				continue
			}
			if fence != reg.source {
				findings = append(findings, finding{File: name, Rule: "listing extraction",
					Detail: fmt.Sprintf("%s does not match %s; the code is the authority, so correct the prose%s",
						label, reg.origin, firstDifference(reg.source, fence))})
			}
		}
	}

	for key, reg := range registered {
		if !printed[key] {
			findings = append(findings, finding{File: examplesDir + "/MANIFEST.yaml", Rule: "listing extraction",
				Detail: fmt.Sprintf("%s registers %s for %s, which that chapter does not print",
					reg.id, reg.label, reg.chapter)})
		}
	}
	return findings
}

// loadRegisteredListings resolves every listing the manifest registers to the
// bytes of its regions, concatenated in the order given. A part with no
// examples directory yields no findings: the check is inert until the
// directory exists.
func loadRegisteredListings(root string) (map[string]registeredListing, []finding) {
	path := filepath.Join(root, examplesDir, "MANIFEST.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, []finding{{File: rel(root, path), Rule: "listing extraction", Detail: err.Error()}}
	}
	var m exampleManifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, []finding{{File: rel(root, path), Rule: "listing extraction", Detail: err.Error()}}
	}

	out := map[string]registeredListing{}
	var findings []finding
	for _, e := range m.Examples {
		if e.Kind != "part" {
			continue
		}
		for _, s := range e.Snapshots {
			for _, l := range s.Listings {
				var parts []string
				var origins []string
				failed := false
				for _, r := range l.Regions {
					file := filepath.Join(root, examplesDir, e.Path, r.File)
					body, err := extractRegion(file, r.Marker)
					if err != nil {
						findings = append(findings, finding{File: rel(root, file), Rule: "listing extraction",
							Detail: fmt.Sprintf("%s: %v", l.ID, err)})
						failed = true
						break
					}
					parts = append(parts, body)
					origins = append(origins, filepath.Join(e.Path, r.File)+":"+r.Marker)
				}
				if failed {
					continue
				}
				out[s.Chapter+"|"+l.Label] = registeredListing{
					id:      l.ID,
					chapter: s.Chapter,
					label:   l.Label,
					source:  strings.Join(parts, "\n"),
					origin:  strings.Join(origins, " + "),
				}
			}
		}
	}
	return out, findings
}

// extractRegion returns the bytes between a region's markers, delimiters
// excluded. Comparison is byte for byte and never normalized: a
// whitespace-tolerant match would let indentation drift, and indentation is
// what makes a Go listing readable on the page.
func extractRegion(path, marker string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	var body []string
	inside, closed := false, false
	for _, line := range strings.Split(string(data), "\n") {
		switch markerDirective(line) {
		case "begin " + marker:
			inside = true
			continue
		case "end " + marker:
			if inside {
				inside, closed = false, true
			}
			continue
		}
		if inside {
			body = append(body, line)
		}
	}
	if !closed {
		return "", fmt.Errorf("region %q is not delimited by begin and end markers", marker)
	}
	return strings.Trim(strings.Join(body, "\n"), "\n"), nil
}

// markerDirective returns the marker a line carries, or "". The prefix is the
// host language's comment -- // in Go, # in YAML -- so it is stripped rather
// than matched (examples/docs/ARCHITECTURE.yaml: listing_extraction).
func markerDirective(line string) string {
	s := strings.TrimSpace(line)
	s = strings.TrimLeft(s, "/#")
	s = strings.TrimSpace(s)
	rest, ok := strings.CutPrefix(s, "example:")
	if !ok {
		return ""
	}
	return strings.TrimSpace(rest)
}

// firstDifference names the first line that differs, so a finding points at
// the edit rather than at the whole listing.
func firstDifference(source, fence string) string {
	src := strings.Split(source, "\n")
	prt := strings.Split(fence, "\n")
	for i := 0; i < len(src) && i < len(prt); i++ {
		if src[i] != prt[i] {
			return fmt.Sprintf(" (line %d: source %q, book %q)", i+1, src[i], prt[i])
		}
	}
	if len(src) != len(prt) {
		return fmt.Sprintf(" (source has %d lines, book has %d)", len(src), len(prt))
	}
	return ""
}

// checkExamplesBuild runs the examples build and reports a failure as one
// finding. The checks themselves live in examples/magefiles; this is the
// dispatch, so the two trees stay independent and the book's audit still
// fails when the examples tree does.
func checkExamplesBuild(root string) []finding {
	dir := filepath.Join(root, examplesDir)
	if _, err := os.Stat(dir); err != nil {
		return nil
	}
	cmd := exec.Command("mage", "-d", examplesDir, "audit", "test")
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err == nil {
		return nil
	}
	detail := strings.TrimSpace(string(out))
	if detail == "" {
		detail = err.Error()
	}
	return []finding{{File: examplesDir, Rule: "examples build",
		Detail: fmt.Sprintf("mage -d %s audit test failed:\n%s", examplesDir, indentLines(detail))}}
}

func indentLines(s string) string {
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		lines[i] = "      " + l
	}
	return strings.Join(lines, "\n")
}
