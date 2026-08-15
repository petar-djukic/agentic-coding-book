// Copyright (c) 2026 Petar Djukic. All rights reserved.
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// Test runs each part module's Go tests. Parts are discovered from
// MANIFEST.yaml, so a directory under parts/ that no manifest entry claims is
// not exercised here -- mage audit reports it instead.
func Test() error {
	return test(".")
}

// test is the testable body, rooted at root.
func test(root string) error {
	parts, err := runnableParts(root, os.Stdout)
	if err != nil {
		return err
	}
	if len(parts) == 0 {
		return fmt.Errorf("no part is %s: nothing to test", statusImplemented)
	}
	for _, p := range parts {
		dir := filepath.Join(root, p.Path)
		fmt.Printf("go test ./... in %s\n", p.Path)
		if err := runIn(dir, "go", "vet", "./..."); err != nil {
			return err
		}
		if err := runIn(dir, "go", "test", "./..."); err != nil {
			return err
		}
	}
	return nil
}
