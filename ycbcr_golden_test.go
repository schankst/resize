// Copyright (c) 2026, Stefan Schank <github@schank.it>
//
// SPDX-License-Identifier: ISC
//
// See the LICENSE file in the repository root for the full license text.

package resize

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"testing"
)

var updateGolden = flag.Bool("update", false, "rewrite testdata/ycbcr_golden.txt")

// goldenCases are the YCbCr resize results whose pixel hashes must never change. The file exists to
// prove that optimisations keep output byte-identical, in particular the removal of the full
// YCbCr-to-ycc copy.
func goldenCases() map[string]string {
	out := make(map[string]string)
	for _, c := range imageCases() {
		if _, ok := c.img.(*image.YCbCr); !ok {
			continue
		}
		for _, in := range interpolationFunctions() {
			for _, dir := range []struct {
				name string
				w, h uint
			}{
				{"down", 4, 4},
				{"up", 16, 16},
				{"wide", 16, 4},
				{"tall", 4, 16},
			} {
				img := Resize(dir.w, dir.h, c.img, in.fn)
				key := fmt.Sprintf("%s/%s/%s", c.name, in.name, dir.name)
				out[key] = pixelHash(img)
			}
		}
	}
	return out
}

func pixelHash(img image.Image) string {
	h := sha256.New()
	fmt.Fprintf(h, "%v\n", img.Bounds())
	b := img.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, bl, a := img.At(x, y).RGBA()
			fmt.Fprintf(h, "%d,%d,%d,%d;", r, g, bl, a)
		}
	}
	return hex.EncodeToString(h.Sum(nil))
}

// Covers [ARCH-3.7].
func TestYCbCrGolden(t *testing.T) {
	const path = "testdata/ycbcr_golden.txt"
	got := goldenCases()

	if *updateGolden {
		f, err := os.Create(path)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		w := bufio.NewWriter(f)
		keys := make([]string, 0, len(got))
		for k := range got {
			keys = append(keys, k)
		}
		// Stable order keeps diffs readable.
		for i := 0; i < len(keys); i++ {
			for j := i + 1; j < len(keys); j++ {
				if keys[j] < keys[i] {
					keys[i], keys[j] = keys[j], keys[i]
				}
			}
		}
		for _, k := range keys {
			fmt.Fprintf(w, "%s %s\n", k, got[k])
		}
		if err := w.Flush(); err != nil {
			t.Fatal(err)
		}
		t.Logf("wrote %d cases to %s", len(got), path)
		return
	}

	f, err := os.Open(filepath.Clean(path))
	if err != nil {
		t.Fatalf("%v (run go test -run TestYCbCrGolden -update)", err)
	}
	defer f.Close()

	want := make(map[string]string)
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		var key, hash string
		if _, err := fmt.Sscan(sc.Text(), &key, &hash); err != nil {
			t.Fatalf("bad line %q: %v", sc.Text(), err)
		}
		want[key] = hash
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}

	for key, hash := range want {
		gotHash, ok := got[key]
		if !ok {
			t.Errorf("%s: missing from actual results", key)
			continue
		}
		if gotHash != hash {
			t.Errorf("%s: pixels changed (got %s, want %s)", key, gotHash, hash)
		}
	}
	for key := range got {
		if _, ok := want[key]; !ok {
			t.Errorf("%s: not present in %s", key, path)
		}
	}
}
