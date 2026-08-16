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

// How a fence in a Build section announces itself. Every fence carries one of
// the two, and a fence carrying neither is the hole GH-139 closed: a code
// block nobody registered, printed in a book whose whole claim is that its
// code is checked.
var (
	// listingLabelRe is the bold label the voice constitution requires above a
	// numbered listing (V-S7).
	listingLabelRe = regexp.MustCompile(`^\*\*(Listing [0-9]+\.[0-9]+)\*\*`)

	// snippetMarkerRe is how an unnumbered fence names itself. An HTML comment
	// rather than a visible label, for the reason the chapter marker is one:
	// it renders to nothing in every output format, so a fence the book shows
	// without a listing number stays unnumbered on the page while still being
	// something the audit can resolve.
	snippetMarkerRe = regexp.MustCompile(`^<!--\s*snippet:\s*(\S+)\s*-->`)

	// buildHeadingRe opens a Build section and sectionHeadingRe closes it.
	buildHeadingRe = regexp.MustCompile(`^## [0-9.]+ Build:`)
	anySectionRe   = regexp.MustCompile(`^## `)

	fenceLineRe = regexp.MustCompile("^```([a-z]*)\\s*$")
)

// fence is one fenced code block inside a Build section, with whatever
// announced it. Exactly one of label and snippet is set on a well-formed
// fence; neither being set is what the check reports.
type fence struct {
	label   string
	snippet string
	body    string
	line    int
}

// buildSectionFences returns every fenced block inside a Build section, in
// document order. Fences elsewhere in the chapter are not the build thread's
// business and are left alone.
func buildSectionFences(data string) []fence {
	var out []fence
	var pendingLabel, pendingSnippet string
	inBuild, inFence := false, false
	var body []string
	fenceLine := 0

	for i, line := range strings.Split(data, "\n") {
		switch {
		case inFence:
			if fenceLineRe.MatchString(line) {
				out = append(out, fence{
					label:   pendingLabel,
					snippet: pendingSnippet,
					body:    strings.Join(body, "\n"),
					line:    fenceLine,
				})
				pendingLabel, pendingSnippet = "", ""
				inFence, body = false, nil
				continue
			}
			body = append(body, line)
		case buildHeadingRe.MatchString(line):
			inBuild = true
			pendingLabel, pendingSnippet = "", ""
		case anySectionRe.MatchString(line):
			inBuild = false
		case !inBuild:
			// nothing to track outside a Build section
		case fenceLineRe.MatchString(line):
			inFence, fenceLine = true, i+1
		default:
			if m := listingLabelRe.FindStringSubmatch(line); m != nil {
				pendingLabel, pendingSnippet = m[1], ""
			}
			if m := snippetMarkerRe.FindStringSubmatch(line); m != nil {
				pendingSnippet, pendingLabel = m[1], ""
			}
		}
	}
	return out
}

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
				ID      string          `yaml:"id"`
				Label   string          `yaml:"label"`
				Regions []exampleRegion `yaml:"regions"`
			} `yaml:"listings"`
			Snippets []struct {
				ID      string          `yaml:"id"`
				Prose   string          `yaml:"prose"`
				Regions []exampleRegion `yaml:"regions"`
			} `yaml:"snippets"`
		} `yaml:"snapshots"`
	} `yaml:"examples"`
}

type exampleRegion struct {
	File   string `yaml:"file"`
	Marker string `yaml:"marker"`
}

// registeredListing is one fence's resolved source, keyed by chapter id so a
// label repeated across chapters -- "Listing 4.1" exists once per part in a
// book numbered by section -- stays distinct.
//
// prose is set for a snippet the manifest declares as prose rather than as an
// extract. That fence is not compared against anything, and the reason it
// cannot be is written down where a reader of the manifest will find it. It
// is a declared exception, which is the difference between an unchecked fence
// and an unnoticed one.
type registeredListing struct {
	id      string
	chapter string
	label   string
	source  string
	origin  string
	prose   string
}

// checkListings compares every fenced code block in a drafted Build section
// against the region it is registered to extract from. Four things are
// findings: a fence that announces itself as neither a listing nor a snippet,
// a fence whose registration is missing, a fence that does not match its
// source, and a registration no chapter prints. The manifest and the chapters
// must agree in both directions, and no fence may sit outside the agreement.
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
		for _, f := range buildSectionFences(string(data)) {
			what, key := f.label, chapter+"|"+f.label
			if f.label == "" {
				what, key = "snippet "+f.snippet, chapter+"|snippet:"+f.snippet
			}
			if f.label == "" && f.snippet == "" {
				findings = append(findings, finding{File: name, Rule: "listing extraction",
					Detail: fmt.Sprintf("the fenced block at line %d carries no **Listing N.M** label and no "+
						"<!-- snippet: id --> marker, so nothing binds it to %s/MANIFEST.yaml", f.line, examplesDir)})
				continue
			}
			printed[key] = true

			reg, ok := registered[key]
			if !ok {
				findings = append(findings, finding{File: name, Rule: "listing extraction",
					Detail: fmt.Sprintf("%s (%s) is not registered in %s/MANIFEST.yaml, so nothing checks it",
						what, chapter, examplesDir)})
				continue
			}
			if reg.prose != "" {
				continue // declared prose: unchecked on purpose, and the manifest says why
			}
			if f.body != reg.source {
				findings = append(findings, finding{File: name, Rule: "listing extraction",
					Detail: fmt.Sprintf("%s does not match %s; the code is the authority, so correct the prose%s",
						what, reg.origin, firstDifference(reg.source, f.body))})
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
				reg, fs := resolveRegions(root, e.Path, l.ID, l.Label, l.Regions)
				findings = append(findings, fs...)
				if reg == nil {
					continue
				}
				reg.chapter = s.Chapter
				out[s.Chapter+"|"+l.Label] = *reg
			}
			for _, sn := range s.Snippets {
				key := s.Chapter + "|snippet:" + sn.ID
				if sn.Prose != "" {
					if len(sn.Regions) > 0 {
						findings = append(findings, finding{File: examplesDir + "/MANIFEST.yaml", Rule: "listing extraction",
							Detail: fmt.Sprintf("snippet %s declares prose and names regions; it is one or the other", sn.ID)})
						continue
					}
					out[key] = registeredListing{id: sn.ID, chapter: s.Chapter, label: "snippet " + sn.ID, prose: sn.Prose}
					continue
				}
				if len(sn.Regions) == 0 {
					findings = append(findings, finding{File: examplesDir + "/MANIFEST.yaml", Rule: "listing extraction",
						Detail: fmt.Sprintf("snippet %s names no regions and declares no prose reason", sn.ID)})
					continue
				}
				reg, fs := resolveRegions(root, e.Path, sn.ID, "snippet "+sn.ID, sn.Regions)
				findings = append(findings, fs...)
				if reg == nil {
					continue
				}
				reg.chapter = s.Chapter
				out[key] = *reg
			}
		}
	}
	return out, findings
}

// resolveRegions reads a fence's regions and concatenates them in the order
// given. A region that cannot be read yields a finding and no registration,
// so the fence is reported once rather than twice.
func resolveRegions(root, partPath, id, label string, regions []exampleRegion) (*registeredListing, []finding) {
	var parts, origins []string
	for _, r := range regions {
		file := filepath.Join(root, examplesDir, partPath, r.File)
		body, err := extractRegion(file, r.Marker)
		if err != nil {
			return nil, []finding{{File: rel(root, file), Rule: "listing extraction",
				Detail: fmt.Sprintf("%s: %v", id, err)}}
		}
		parts = append(parts, body)
		origins = append(origins, filepath.Join(partPath, r.File)+":"+r.Marker)
	}
	return &registeredListing{
		id:     id,
		label:  label,
		source: strings.Join(parts, "\n"),
		origin: strings.Join(origins, " + "),
	}, nil
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
