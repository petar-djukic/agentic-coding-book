// Copyright (c) 2026 Petar Djukic. All rights reserved.
// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// Test runs this build's own tests and then each part module's. Parts are
// discovered from MANIFEST.yaml, so a directory under parts/ that no manifest
// entry claims is not exercised here -- mage audit reports it instead.
//
// The build's own suite is included because leaving it out is how it stops
// running. It was unreachable from every target until GH-139, and in that
// window an assertion about the manifest went stale and failed on main with
// nothing to notice. A suite no target invokes is not a gate.
func Test() error {
	return test(".")
}

// test is the testable body, rooted at root. The self-test runs only for the
// real root: a test calling test() would otherwise re-enter its own suite.
func test(root string) error {
	if root == "." {
		fmt.Println("go test ./... in magefiles")
		if err := runIn(filepath.Join(root, "magefiles"), "go", "test", "./..."); err != nil {
			return err
		}
	}

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
