// Copyright (c) 2026 Petar Djukic. All rights reserved.
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// Audit enforces the constraints in docs/ARCHITECTURE.yaml. Findings
// accumulate and report together; the audit fails once, with everything it
// found, and each finding names the constraint id it comes from.
//
// Findings recorded in docs/audit-baseline.yaml are accepted debt: they are
// printed but do not fail the audit, and each names the issue that clears
// it. A baseline entry matching nothing is itself a finding, so the file
// cannot rot into a blanket exemption.
//
// The listing-extraction check is deliberately absent. It reads the book's
// chapters as well as the part source, so it lives in the book's own
// magefiles; the root audit shells in here and reports what this returns as
// one finding of its own.
func Audit() error {
	return audit(".")
}

// audit is the testable body, rooted at root -- "." for a target, ".." for a
// test running from magefiles/.
func audit(root string) error {
	book, err := bookRoot(root)
	if err != nil {
		return err
	}

	m, err := loadManifest(root)
	if err != nil {
		return err
	}

	var findings []finding
	findings = append(findings, checkChapterIDs(m, book)...)
	findings = append(findings, checkBuildSectionCoverage(m, book)...)
	findings = append(findings, checkProvenance(m)...)
	findings = append(findings, checkCatalogHeaders(root)...)
	findings = append(findings, checkListingRegions(root, m)...)
	findings = append(findings, checkUpstreamIndependence(root, m)...)
	findings = append(findings, checkSnapshotDiffs(root, m)...)
	findings = append(findings, checkRealizes(root, m, book)...)

	accepted, baselineFindings := loadBaseline(root)
	findings = append(findings, baselineFindings...)
	return reportFindings(findings, accepted)
}

// bookRoot resolves the repository the examples live in, so a run from the
// wrong directory says so rather than reporting every chapter missing.
func bookRoot(root string) (string, error) {
	book := filepath.Join(root, "..")
	for _, want := range []string{
		filepath.Join("docs", "ARCHITECTURE.yaml"),
		filepath.Join("docs", "srd"),
	} {
		if _, err := os.Stat(filepath.Join(book, want)); err != nil {
			return "", fmt.Errorf("examples/ audit must run inside the book repository: %s not found at %s",
				want, book)
		}
	}
	return book, nil
}

// finding is one audit result: where it is, which constraint it comes from,
// and what is wrong. The triple is also its baseline key. The shape is the
// book's own (magefiles/audit.go); it is reimplemented rather than imported
// because examples/magefiles is a separate module and a dependency between
// two build trees would be worse than sixty duplicated lines.
type finding struct {
	File   string
	Rule   string
	Detail string
}

func (f finding) key() string { return f.File + "|" + f.Rule + "|" + f.Detail }

func (f finding) String() string {
	return fmt.Sprintf("%s [%s] %s", f.File, f.Rule, f.Detail)
}

func reportFindings(findings []finding, accepted map[string]string) error {
	sort.Slice(findings, func(i, j int) bool { return findings[i].key() < findings[j].key() })

	var blocking, known []string
	seen := map[string]bool{}
	matched := map[string]bool{}
	for _, f := range findings {
		if seen[f.key()] {
			continue
		}
		seen[f.key()] = true
		if issue, ok := accepted[f.key()]; ok {
			matched[f.key()] = true
			known = append(known, fmt.Sprintf("%s (accepted, %s)", f, issue))
			continue
		}
		blocking = append(blocking, f.String())
	}

	keys := make([]string, 0, len(accepted))
	for k := range accepted {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if !matched[k] {
			blocking = append(blocking, finding{
				File:   baselineFile,
				Rule:   "baseline",
				Detail: fmt.Sprintf("accepted finding %q no longer occurs; delete the entry", k),
			}.String())
		}
	}

	for _, k := range known {
		fmt.Printf("examples audit: %s\n", k)
	}
	if len(blocking) > 0 {
		return fmt.Errorf("examples audit failed:\n  - %s", strings.Join(blocking, "\n  - "))
	}
	fmt.Printf("examples audit: OK (%d accepted finding(s) in %s)\n", len(known), baselineFile)
	return nil
}

// ------------------------------------------------------------------ E-C6

// checkChapterIDs resolves every chapter id the manifest names against the
// book's own architecture. The manifest is the binding table; a binding to a
// chapter that does not exist binds nothing.
func checkChapterIDs(m *manifest, book string) []finding {
	known, err := bookChapterIDs(book)
	if err != nil {
		return []finding{{File: manifestFile, Rule: "E-C6", Detail: err.Error()}}
	}
	var out []finding
	for _, e := range m.Examples {
		for _, s := range e.Snapshots {
			if !known[s.Chapter] {
				out = append(out, finding{File: manifestFile, Rule: "E-C6",
					Detail: fmt.Sprintf("%s names chapter %s, which docs/ARCHITECTURE.yaml does not define", e.ID, s.Chapter)})
			}
		}
		for _, c := range e.Consumers {
			if !known[c] {
				out = append(out, finding{File: manifestFile, Rule: "E-C6",
					Detail: fmt.Sprintf("%s lists consumer %s, which docs/ARCHITECTURE.yaml does not define", e.ID, c)})
			}
		}
	}
	return out
}

type bookArch struct {
	Structure struct {
		Parts []struct {
			ID       string `yaml:"id"`
			Chapters []struct {
				ID  string `yaml:"id"`
				SRD string `yaml:"srd"`
			} `yaml:"chapters"`
		} `yaml:"parts"`
	} `yaml:"structure"`
}

func loadBookArch(book string) (*bookArch, error) {
	path := filepath.Join(book, "docs", "ARCHITECTURE.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var a bookArch
	if err := yaml.Unmarshal(data, &a); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return &a, nil
}

func bookChapterIDs(book string) (map[string]bool, error) {
	a, err := loadBookArch(book)
	if err != nil {
		return nil, err
	}
	ids := map[string]bool{}
	for _, p := range a.Structure.Parts {
		for _, c := range p.Chapters {
			ids[c.ID] = true
		}
	}
	if len(ids) == 0 {
		return nil, fmt.Errorf("docs/ARCHITECTURE.yaml defines no chapters")
	}
	return ids, nil
}

// ------------------------------------------------------------------ E-C7

var (
	buildSectionRe  = regexp.MustCompile(`(?m)^## [0-9.]+ Build:`)
	chapterMarkerRe = regexp.MustCompile(`<!--\s*chapter:\s*([A-Za-z0-9.]+)\s*-->`)
	chapterFileRe   = regexp.MustCompile(`^\d\d-.*\.md$`)
	releaseTagRe    = regexp.MustCompile(`^v[0-9]`)
)

// checkBuildSectionCoverage reports a drafted Build section with no snapshot
// behind it. Without this a section drafted later inherits no artifact and no
// check, and the omission looks like a decision.
func checkBuildSectionCoverage(m *manifest, book string) []finding {
	covered := map[string]bool{}
	for _, e := range m.Examples {
		for _, s := range e.Snapshots {
			covered[s.Chapter] = true
		}
	}

	entries, err := os.ReadDir(book)
	if err != nil {
		return []finding{{File: "docs", Rule: "E-C7", Detail: err.Error()}}
	}
	var out []finding
	for _, entry := range entries {
		name := entry.Name()
		if !chapterFileRe.MatchString(name) {
			continue
		}
		data, err := os.ReadFile(filepath.Join(book, name))
		if err != nil {
			out = append(out, finding{File: name, Rule: "E-C7", Detail: err.Error()})
			continue
		}
		if !buildSectionRe.Match(data) {
			continue
		}
		mark := chapterMarkerRe.FindSubmatch(data)
		if mark == nil {
			out = append(out, finding{File: name, Rule: "E-C7",
				Detail: "has a Build section but no <!-- chapter: --> marker, so nothing can bind an artifact to it"})
			continue
		}
		if id := string(mark[1]); !covered[id] {
			out = append(out, finding{File: name, Rule: "E-C7",
				Detail: fmt.Sprintf("%s has a Build section with no snapshot entry in %s", id, manifestFile)})
		}
	}
	return out
}

// ------------------------------------------------------------------ E-C4

// checkProvenance reports a copy that cannot be diffed against its source. An
// unpinned copy drifts without anyone being able to tell.
func checkProvenance(m *manifest) []finding {
	var out []finding
	for _, kind := range copiedKinds {
		for _, e := range m.entries(kind) {
			p := e.Provenance
			if p == nil {
				out = append(out, finding{File: manifestFile, Rule: "E-C4",
					Detail: fmt.Sprintf("%s is copied material with no provenance block", e.ID)})
				continue
			}
			for _, f := range []struct{ name, value string }{
				{"upstream", p.Upstream}, {"path", p.Path}, {"release", p.Release},
				{"license", p.License}, {"holder", p.Holder}, {"simplified", p.Simplified},
			} {
				if strings.TrimSpace(f.value) == "" {
					out = append(out, finding{File: manifestFile, Rule: "E-C4",
						Detail: fmt.Sprintf("%s provenance has no %s", e.ID, f.name)})
				}
			}
			if p.Release != "" && !releaseTagRe.MatchString(p.Release) {
				out = append(out, finding{File: manifestFile, Rule: "E-C4",
					Detail: fmt.Sprintf("%s release %q does not look like an upstream tag", e.ID, p.Release)})
			}
		}
	}
	return out
}

// ------------------------------------------------------------------ E-C5

const (
	catalogDir     = "catalog"
	upstreamHolder = "Nokia"
	upstreamSPDX   = "BSD-3-Clause"
)

// checkCatalogHeaders reports a copied file whose notice was stripped. The
// header staying in place is what discharges BSD-3 clause 1, and it is why
// the repository needs no NOTICE file.
func checkCatalogHeaders(root string) []finding {
	dir := filepath.Join(root, catalogDir)
	if _, err := os.Stat(dir); err != nil {
		return nil
	}
	var out []finding
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		// The directory's own README is this repository's prose, not a copy.
		if strings.EqualFold(info.Name(), "README.md") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		text := string(data)
		rel := mustRel(root, path)
		if !strings.Contains(text, upstreamHolder) {
			out = append(out, finding{File: rel, Rule: "E-C5",
				Detail: "copied file does not name " + upstreamHolder})
		}
		if !strings.Contains(text, upstreamSPDX) {
			out = append(out, finding{File: rel, Rule: "E-C5",
				Detail: "copied file does not carry the " + upstreamSPDX + " identifier"})
		}
		return nil
	})
	if err != nil {
		out = append(out, finding{File: catalogDir, Rule: "E-C5", Detail: err.Error()})
	}
	return out
}

// ------------------------------------------------------------------ E-C3

// checkListingRegions resolves every registered listing and snippet and
// reports one that points at nothing, and one that points into catalog/. The
// boundary is not a style rule: a BSD-3-covered file reproduced verbatim in
// the built PDF would pull clause 2's notice obligation onto the book itself,
// and an unnumbered snippet reproduces bytes exactly as a listing does.
func checkListingRegions(root string, m *manifest) []finding {
	var out []finding
	for _, e := range m.entries(kindPart) {
		for _, s := range e.Snapshots {
			for _, l := range s.Listings {
				if len(l.Regions) == 0 {
					out = append(out, finding{File: manifestFile, Rule: "E-C3",
						Detail: fmt.Sprintf("listing %s names no region", l.ID)})
					continue
				}
				out = append(out, resolveRegionFindings(root, e.Path, "listing "+l.ID, l.Regions)...)
			}
			for _, sn := range s.Snippets {
				switch {
				case sn.Prose != "" && len(sn.Regions) > 0:
					out = append(out, finding{File: manifestFile, Rule: "E-C3",
						Detail: fmt.Sprintf("snippet %s declares prose and names regions; it is one or the other", sn.ID)})
				case sn.Prose == "" && len(sn.Regions) == 0:
					out = append(out, finding{File: manifestFile, Rule: "E-C3",
						Detail: fmt.Sprintf("snippet %s names no region and declares no prose reason", sn.ID)})
				case sn.Prose != "":
					// Declared prose. Unchecked by design, and the manifest
					// carries the reason, which is what makes it a decision
					// rather than an omission.
				default:
					out = append(out, resolveRegionFindings(root, e.Path, "snippet "+sn.ID, sn.Regions)...)
				}
			}
		}
	}
	return out
}

// resolveRegionFindings checks every region a listing or snippet names: that
// it stays inside parts/, and that the file delimits it exactly once.
func resolveRegionFindings(root, partPath, what string, regions []region) []finding {
	var out []finding
	for _, r := range regions {
		rel := filepath.Join(partPath, r.File)
		if within(rel, catalogDir) {
			out = append(out, finding{File: manifestFile, Rule: "E-C3",
				Detail: fmt.Sprintf("%s extracts from %s; listings resolve into parts/ only", what, rel)})
			continue
		}
		out = append(out, regionFindings(root, rel, what, r.Marker)...)
	}
	return out
}

// regionFindings reports a region whose file or markers are missing, so a
// manifest entry pointing at nothing fails here rather than surfacing later
// as an unexplained extraction mismatch.
func regionFindings(root, rel, listingID, marker string) []finding {
	data, err := os.ReadFile(filepath.Join(root, rel))
	if err != nil {
		return []finding{{File: rel, Rule: "E-C3",
			Detail: fmt.Sprintf("%s names this file, which cannot be read: %v", listingID, err)}}
	}
	begin, end := 0, 0
	for _, line := range strings.Split(string(data), "\n") {
		switch markerOf(line) {
		case "begin " + marker:
			begin++
		case "end " + marker:
			end++
		}
	}
	switch {
	case begin == 0 || end == 0:
		return []finding{{File: rel, Rule: "E-C3",
			Detail: fmt.Sprintf("%s names marker %q, which the file does not delimit", listingID, marker)}}
	case begin > 1 || end > 1:
		return []finding{{File: rel, Rule: "E-C3",
			Detail: fmt.Sprintf("marker %q is opened %d and closed %d times; a region is one span", marker, begin, end)}}
	}
	return nil
}

// markerOf returns the marker directive a line carries, or "". A marker is a
// comment in the host language -- // in Go, # in YAML -- so the prefix is
// stripped rather than matched (docs/ARCHITECTURE.yaml: listing_extraction).
func markerOf(line string) string {
	s := strings.TrimSpace(line)
	s = strings.TrimLeft(s, "/#")
	s = strings.TrimSpace(s)
	rest, ok := strings.CutPrefix(s, "example:")
	if !ok {
		return ""
	}
	return strings.TrimSpace(rest)
}

// ------------------------------------------------------------------ E-C1

const upstreamModule = "declarative-agents"

// checkUpstreamIndependence reports a part that cannot compile on its own.
// The reader builds their own runtime; a part that needs upstream code has
// abandoned the premise, and nothing but a check would notice.
func checkUpstreamIndependence(root string, m *manifest) []finding {
	var out []finding
	for _, e := range m.entries(kindPart) {
		path := filepath.Join(root, e.Path, "go.mod")
		data, err := os.ReadFile(path)
		if err != nil {
			if e.Status == statusImplemented {
				out = append(out, finding{File: filepath.Join(e.Path, "go.mod"), Rule: "E-C1",
					Detail: fmt.Sprintf("%s is %s but has no module file", e.ID, statusImplemented)})
			}
			continue
		}
		if strings.Contains(string(data), upstreamModule) {
			out = append(out, finding{File: filepath.Join(e.Path, "go.mod"), Rule: "E-C1",
				Detail: fmt.Sprintf("%s requires %s; parts compile standalone", e.ID, upstreamModule)})
		}
	}
	return out
}

// ------------------------------------------------------------------ E-C2

// supportFile reports whether a snapshot file is scaffolding rather than the
// thing the reader is building. These differ between every pair of snapshots
// by construction, so a diff rule covering them fires always and gets turned
// off (docs/ARCHITECTURE.yaml: E-C2 scope).
func supportFile(rel string) bool {
	base := filepath.Base(rel)
	return strings.HasSuffix(base, "_test.go") ||
		base == "demo.go" ||
		strings.HasPrefix(rel, "fixture"+string(filepath.Separator))
}

// checkSnapshotDiffs reports book source that changed between consecutive
// snapshots without the manifest saying so. A snapshot claims the runtime
// changed exactly as the prose says it did; undeclared drift makes that claim
// false silently.
func checkSnapshotDiffs(root string, m *manifest) []finding {
	var out []finding
	for _, e := range m.entries(kindPart) {
		for i := 1; i < len(e.Snapshots); i++ {
			prev, cur := e.Snapshots[i-1], e.Snapshots[i]
			changed, err := changedBookSource(
				filepath.Join(root, e.Path, prev.Path),
				filepath.Join(root, e.Path, cur.Path))
			if err != nil {
				out = append(out, finding{File: filepath.Join(e.Path, cur.Path), Rule: "E-C2", Detail: err.Error()})
				continue
			}
			declared := declaredFiles(cur.Adds)
			for _, name := range changed {
				if !declared[name] {
					out = append(out, finding{File: filepath.Join(e.Path, cur.Path), Rule: "E-C2",
						Detail: fmt.Sprintf("%s differs from %s but %s does not declare it in adds:",
							name, prev.Path, manifestFile)})
				}
			}
			for name := range declared {
				if !contains(changed, name) {
					out = append(out, finding{File: manifestFile, Rule: "E-C2",
						Detail: fmt.Sprintf("%s declares %s as added at %s, but it does not differ from %s",
							e.ID, name, cur.Path, prev.Path)})
				}
			}
		}
	}
	return out
}

// declaredFiles reads the file names out of an adds: list. An entry states
// the file it changed before a colon, so the declaration names something a
// check can find rather than describing it in prose.
func declaredFiles(adds []string) map[string]bool {
	out := map[string]bool{}
	for _, a := range adds {
		name, _, ok := strings.Cut(a, ":")
		if ok {
			out[strings.TrimSpace(name)] = true
		}
	}
	return out
}

// changedBookSource returns the book-source files that differ between two
// snapshot directories, present in one and not the other included.
func changedBookSource(prev, cur string) ([]string, error) {
	prevFiles, err := bookSource(prev)
	if err != nil {
		return nil, err
	}
	curFiles, err := bookSource(cur)
	if err != nil {
		return nil, err
	}
	names := map[string]bool{}
	for n := range prevFiles {
		names[n] = true
	}
	for n := range curFiles {
		names[n] = true
	}
	var out []string
	for n := range names {
		if prevFiles[n] != curFiles[n] {
			out = append(out, n)
		}
	}
	sort.Strings(out)
	return out, nil
}

// bookSource maps each book-source file in dir to its contents.
func bookSource(dir string) (map[string]string, error) {
	out := map[string]string{}
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		rel, relErr := filepath.Rel(dir, path)
		if relErr != nil {
			return relErr
		}
		if supportFile(rel) {
			return nil
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		out[rel] = string(data)
		return nil
	})
	return out, err
}

// ------------------------------------------------------------------ E-C8

// checkRealizes resolves every chapter SRD a part SRD cites. The id is a
// citation, so a renumbered chapter breaks it loudly rather than pointing at
// the wrong contract.
func checkRealizes(root string, m *manifest, book string) []finding {
	var out []finding
	for _, e := range m.entries(kindPart) {
		if e.SRD == "" {
			out = append(out, finding{File: manifestFile, Rule: "E-C8",
				Detail: fmt.Sprintf("%s names no SRD", e.ID)})
			continue
		}
		path := filepath.Join(root, e.SRD)
		data, err := os.ReadFile(path)
		if err != nil {
			out = append(out, finding{File: e.SRD, Rule: "E-C8", Detail: err.Error()})
			continue
		}
		var doc struct {
			Realizes []string `yaml:"realizes"`
		}
		if err := yaml.Unmarshal(data, &doc); err != nil {
			out = append(out, finding{File: e.SRD, Rule: "E-C8", Detail: err.Error()})
			continue
		}
		if len(doc.Realizes) == 0 {
			out = append(out, finding{File: e.SRD, Rule: "E-C8",
				Detail: "realizes: is empty; a part SRD cites the chapter contracts it implements"})
		}
		for _, id := range doc.Realizes {
			hits, err := filepath.Glob(filepath.Join(book, "docs", "srd", id+"-*.yaml"))
			if err != nil {
				out = append(out, finding{File: e.SRD, Rule: "E-C8", Detail: err.Error()})
				continue
			}
			switch len(hits) {
			case 1:
			case 0:
				out = append(out, finding{File: e.SRD, Rule: "E-C8",
					Detail: fmt.Sprintf("realizes %s, which resolves to no file under docs/srd/", id)})
			default:
				out = append(out, finding{File: e.SRD, Rule: "E-C8",
					Detail: fmt.Sprintf("realizes %s, which resolves to %d files under docs/srd/", id, len(hits))})
			}
		}
	}
	return out
}

// --------------------------------------------------------------- baseline

const baselineFile = "docs/audit-baseline.yaml"

type baselineDoc struct {
	Accepted []struct {
		File   string `yaml:"file"`
		Rule   string `yaml:"rule"`
		Detail string `yaml:"detail"`
		Issue  string `yaml:"issue"`
	} `yaml:"accepted"`
}

func loadBaseline(root string) (map[string]string, []finding) {
	accepted := map[string]string{}
	data, err := os.ReadFile(filepath.Join(root, baselineFile))
	if err != nil {
		if os.IsNotExist(err) {
			return accepted, nil
		}
		return accepted, []finding{{File: baselineFile, Rule: "baseline", Detail: err.Error()}}
	}
	var doc baselineDoc
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return accepted, []finding{{File: baselineFile, Rule: "baseline", Detail: err.Error()}}
	}
	var findings []finding
	for _, e := range doc.Accepted {
		if e.Issue == "" {
			findings = append(findings, finding{File: baselineFile, Rule: "baseline",
				Detail: fmt.Sprintf("entry %s|%s has no issue; accepted debt names the issue that clears it", e.File, e.Rule)})
		}
		accepted[finding{File: e.File, Rule: e.Rule, Detail: e.Detail}.key()] = e.Issue
	}
	return accepted, findings
}

// ----------------------------------------------------------------- shared

func mustRel(root, path string) string {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return path
	}
	return rel
}

// within reports whether rel sits under dir.
func within(rel, dir string) bool {
	return rel == dir || strings.HasPrefix(rel, dir+string(filepath.Separator))
}

func contains(haystack []string, needle string) bool {
	for _, h := range haystack {
		if h == needle {
			return true
		}
	}
	return false
}
