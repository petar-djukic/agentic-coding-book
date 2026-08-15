// Copyright (c) 2026 Petar Djukic. All rights reserved.
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// manifestFile is the binding table every target reads. Nothing here walks
// the directory tree: a part present on disk but absent from the manifest is
// not exercised, and its absence is a finding rather than a silent pass
// (docs/ARCHITECTURE.yaml: dependency).
const manifestFile = "MANIFEST.yaml"

// Statuses the manifest's status field takes. It gates the build and nothing
// else -- whether a chapter's prose has been converted to extract from a part
// is carried by docs/road-map.yaml on a scale of its own.
const (
	statusPlanned     = "planned"
	statusImplemented = "implemented"
)

// Kinds an entry takes.
const (
	kindPart    = "part"
	kindCatalog = "catalog-family"
)

type manifest struct {
	SchemaVersion int        `yaml:"schema_version"`
	ThirdParty    thirdParty `yaml:"third_party"`
	Examples      []example  `yaml:"examples"`
}

type example struct {
	ID         string      `yaml:"id"`
	Kind       string      `yaml:"kind"`
	Status     string      `yaml:"status"`
	Path       string      `yaml:"path"`
	Module     string      `yaml:"module"`
	SRD        string      `yaml:"srd"`
	BookPart   string      `yaml:"book_part"`
	Summary    string      `yaml:"summary"`
	Consumers  []string    `yaml:"consumers"`
	Snapshots  []snapshot  `yaml:"snapshots"`
	Provenance *provenance `yaml:"provenance"`
}

type snapshot struct {
	Chapter  string    `yaml:"chapter"`
	Path     string    `yaml:"path"`
	Leaves   string    `yaml:"leaves"`
	Adds     []string  `yaml:"adds"`
	Listings []listing `yaml:"listings"`
}

type listing struct {
	ID       string   `yaml:"id"`
	Label    string   `yaml:"label"`
	Language string   `yaml:"language"`
	Regions  []region `yaml:"regions"`
	Note     string   `yaml:"note"`
}

type region struct {
	File   string `yaml:"file"`
	Marker string `yaml:"marker"`
}

type provenance struct {
	Upstream   string `yaml:"upstream"`
	Path       string `yaml:"path"`
	Release    string `yaml:"release"`
	License    string `yaml:"license"`
	Holder     string `yaml:"holder"`
	Simplified string `yaml:"simplified"`
}

type thirdParty struct {
	Upstream    string `yaml:"upstream"`
	License     string `yaml:"license"`
	Holder      string `yaml:"holder"`
	Obligation  string `yaml:"obligation"`
	Boundary    string `yaml:"boundary"`
	Attribution string `yaml:"attribution"`
}

// loadManifest reads the manifest rooted at root, which is "." for a mage
// target and ".." for a test running from magefiles/.
func loadManifest(root string) (*manifest, error) {
	path := filepath.Join(root, manifestFile)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var m manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if m.SchemaVersion != 1 {
		return nil, fmt.Errorf("%s: schema_version %d, this build understands 1", path, m.SchemaVersion)
	}
	return &m, nil
}

// entries returns the manifest entries of one kind, in the order the file
// lists them.
func (m *manifest) entries(kind string) []example {
	var out []example
	for _, e := range m.Examples {
		if e.Kind == kind {
			out = append(out, e)
		}
	}
	return out
}
