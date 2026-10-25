// Copyright (c) 2026, Stefan Schank <github@schank.it>
//
// SPDX-License-Identifier: ISC
//
// See the LICENSE file in the repository root for the full license text.

package resize

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// anchors in source comments look like [ARCH-3.4]; every one of them must be a heading in
// docs/how-it-works.md. The literal prefix is assembled so this file does not match itself.
var (
	anchorPrefix  = "[" + "ARCH-"
	codeAnchorRe  = regexp.MustCompile(regexp.QuoteMeta(anchorPrefix) + `\d+(?:\.\d+)?\]`)
	docHeadingRe  = regexp.MustCompile(`(?m)^#{1,6}[ \t]+` + regexp.QuoteMeta(anchorPrefix) + `\d+(?:\.\d+)?\]`)
	docAnchorOnly = regexp.MustCompile(regexp.QuoteMeta(anchorPrefix) + `\d+(?:\.\d+)?\]`)
)

const architectureDoc = "docs/how-it-works.md"

// Covers [ARCH-1]: source comments reference sections of docs/how-it-works.md, and those sections
// must exist. Renaming a heading without updating the code fails here instead of rotting silently.
// The map in [ARCH-6] lists which test covers which stage.
func TestDocAnchorsExist(t *testing.T) {
	headings, err := documentedAnchors(architectureDoc)
	if err != nil {
		t.Fatalf("read %s: %v", architectureDoc, err)
	}
	if len(headings) == 0 {
		t.Fatalf("%s contains no [ARCH-…] headings", architectureDoc)
	}

	used := map[string][]string{}
	err = filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			switch path {
			case ".git", "testdata":
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for _, anchor := range codeAnchorRe.FindAllString(string(data), -1) {
			used[anchor] = append(used[anchor], path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	var missing []string
	for anchor, files := range used {
		if _, ok := headings[anchor]; ok {
			continue
		}
		sort.Strings(files)
		missing = append(missing, fmt.Sprintf("%s (referenced in %s)", anchor, strings.Join(unique(files), ", ")))
	}
	sort.Strings(missing)
	if len(missing) > 0 {
		t.Fatalf("anchors without a heading in %s:\n  %s", architectureDoc, strings.Join(missing, "\n  "))
	}

	// The other direction is reported, not enforced: overview sections may only be linked from the
	// document itself.
	var unused []string
	for anchor := range headings {
		if len(used[anchor]) == 0 {
			unused = append(unused, anchor)
		}
	}
	sort.Strings(unused)
	if len(unused) > 0 {
		t.Logf("documented but not referenced from code: %s", strings.Join(unused, ", "))
	}
}

func documentedAnchors(path string) (map[string]struct{}, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	out := make(map[string]struct{})
	for _, line := range docHeadingRe.FindAllString(string(data), -1) {
		if anchor := docAnchorOnly.FindString(line); anchor != "" {
			out[anchor] = struct{}{}
		}
	}
	return out, nil
}

func unique(in []string) []string {
	out := in[:0]
	for i, v := range in {
		if i == 0 || v != in[i-1] {
			out = append(out, v)
		}
	}
	return out
}
