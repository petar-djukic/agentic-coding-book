// Copyright (c) 2026 Petar Djukic. All rights reserved.
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/magefile/mage/sh"
)

const (
	bookSlug   = "agentic-coding"
	outputDir  = "generated-files"
	figuresDir = "figures"
	csl        = "templates/ieee.csl"
	template   = "templates/eisvogel.latex"
)

// All generates figures then builds the PDF (default target).
func All() error {
	if err := Figures(); err != nil {
		return err
	}
	return PDF()
}

// Figures renders all PlantUML diagrams to PNG.
func Figures() error {
	pumls, err := filepath.Glob(filepath.Join(figuresDir, "*.puml"))
	if err != nil {
		return err
	}
	if len(pumls) == 0 {
		fmt.Println("no .puml files found")
		return nil
	}
	sort.Strings(pumls)
	for _, puml := range pumls {
		png := strings.TrimSuffix(puml, ".puml") + ".png"
		pumlInfo, err := os.Stat(puml)
		if err != nil {
			return err
		}
		if pngInfo, err := os.Stat(png); err == nil && pngInfo.ModTime().After(pumlInfo.ModTime()) {
			continue
		}
		fmt.Printf("plantuml %s\n", puml)
		if err := sh.Run("plantuml", "-tpng", "-o", ".", puml); err != nil {
			return fmt.Errorf("plantuml %s: %w", puml, err)
		}
	}
	return nil
}

// PDF compiles the numbered markdown chapters into a PDF via the
// Eisvogel pandoc template.
func PDF() error {
	if err := Figures(); err != nil {
		return err
	}

	mds, err := discoverMarkdownChapters()
	if err != nil {
		return err
	}
	if len(mds) == 0 {
		return fmt.Errorf("no [0-9][0-9]-*.md chapter files found")
	}

	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", outputDir, err)
	}

	date := time.Now().Format("2006-01-02")
	out := filepath.Join(outputDir, fmt.Sprintf("%s-v%s.pdf", bookSlug, date))

	args := []string{
		"--citeproc",
		"--csl=" + csl,
		"--bibliography=references.yaml",
		"--template=" + template,
		"--from", "markdown",
		"--pdf-engine=xelatex",
		"--listings",
	}
	args = append(args, mds...)
	args = append(args, "-o", out)

	fmt.Printf("generating %s from %d chapters\n", out, len(mds))
	if err := sh.Run("pandoc", args...); err != nil {
		return fmt.Errorf("pandoc: %w", err)
	}
	return nil
}

// Outline renders the book's outline from docs/srd/ to a PDF, so the
// structure is reviewable before any chapter is drafted.
func Outline() error {
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", outputDir, err)
	}

	md, err := sh.Output("go", "run", "./cmd/genoutline")
	if err != nil {
		return fmt.Errorf("genoutline: %w", err)
	}

	date := time.Now().Format("2006-01-02")
	src := filepath.Join(outputDir, fmt.Sprintf("%s-outline-%s.md", bookSlug, date))
	out := filepath.Join(outputDir, fmt.Sprintf("%s-outline-%s.pdf", bookSlug, date))

	if err := os.WriteFile(src, []byte(md), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", src, err)
	}

	// The outline quotes citation ids as literal text rather than pandoc
	// [@key] markers, so it needs no bibliography pass.
	args := []string{
		"--template=" + template,
		"--from", "markdown",
		"--pdf-engine=xelatex",
		"--listings",
		src, "-o", out,
	}

	fmt.Printf("generating %s\n", out)
	if err := sh.Run("pandoc", args...); err != nil {
		return fmt.Errorf("pandoc: %w", err)
	}
	return nil
}

// Critic runs an LLM critic over a drafted chapter and reports findings
// against the constitutions and the chapter's SRD. Pass a chapter file, or
// "all" to critique every drafted chapter. It never edits the prose, and it
// skips rather than fails when ANTHROPIC_API_KEY is unset.
func Critic(chapter string) error {
	args := []string{"run", "./cmd/critic"}
	if chapter != "" && chapter != "all" {
		args = append(args, chapter)
	}
	if err := sh.RunV("go", args...); err != nil {
		return fmt.Errorf("critic: %w", err)
	}
	return nil
}

// Clean removes generated PNGs and PDFs.
func Clean() error {
	pngs, _ := filepath.Glob(filepath.Join(figuresDir, "*.png"))
	for _, f := range pngs {
		fmt.Printf("rm %s\n", f)
		os.Remove(f)
	}
	entries, err := os.ReadDir(outputDir)
	if err == nil {
		for _, e := range entries {
			path := filepath.Join(outputDir, e.Name())
			fmt.Printf("rm %s\n", path)
			os.Remove(path)
		}
	}
	return nil
}

func discoverMarkdownChapters() ([]string, error) {
	entries, err := os.ReadDir(".")
	if err != nil {
		return nil, err
	}
	var mds []string
	for _, e := range entries {
		name := e.Name()
		if len(name) >= 3 && name[0] >= '0' && name[0] <= '9' && name[1] >= '0' && name[1] <= '9' && name[2] == '-' && strings.HasSuffix(name, ".md") {
			mds = append(mds, name)
		}
	}
	sort.Strings(mds)
	return mds, nil
}
