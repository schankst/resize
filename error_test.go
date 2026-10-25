// Copyright (c) 2026, Stefan Schank <github@schank.it>
//
// SPDX-License-Identifier: ISC
//
// See the LICENSE file in the repository root for the full license text.

package resize

import (
	"errors"
	"image"
	"math"
	"strconv"
	"testing"
)

// TestResizeENilImage covers the nil guard for both entry points.
//
// Covers [ARCH-3.1].
func TestResizeENilImage(t *testing.T) {
	if _, err := ResizeE(8, 8, nil, Bilinear); !errors.Is(err, ErrNilImage) {
		t.Fatalf("ResizeE error = %v, want ErrNilImage", err)
	}
	if _, err := ThumbnailE(8, 8, nil, Bilinear); !errors.Is(err, ErrNilImage) {
		t.Fatalf("ThumbnailE error = %v, want ErrNilImage", err)
	}
}

// TestResizeEPanicsOnInvalidInput documents that the compatibility wrappers still panic.
//
// Covers [ARCH-3.1].
func TestResizeEPanicsOnInvalidInput(t *testing.T) {
	for _, tc := range []struct {
		name string
		call func()
	}{
		{"Resize nil image", func() { Resize(8, 8, nil, Bilinear) }},
		{"Resize huge width", func() { Resize(^uint(0), 8, image.NewRGBA(image.Rect(0, 0, 8, 8)), Bilinear) }},
		{"Thumbnail nil image", func() { Thumbnail(8, 8, nil, Bilinear) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("expected a panic")
				}
			}()
			tc.call()
		})
	}
}

// TestResizeEInvalidDimensions covers dimensions that cannot be represented or allocated. Nothing
// here may panic or attempt an allocation.
//
// Covers [ARCH-3.1].
func TestResizeEInvalidDimensions(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 8, 8))

	cases := []struct {
		name string
		w, h uint
	}{
		{"max uint", ^uint(0), 8},
		{"max uint32", uint(math.MaxUint32), 8},
		{"max uint32 both", uint(math.MaxUint32), uint(math.MaxUint32)},
	}
	if strconv.IntSize == 64 {
		// Both factors fit into uint on this platform, but their product overflows int.
		big := uint64(1) << 40
		cases = append(cases, struct {
			name string
			w, h uint
		}{"product overflow", 1 << 24, uint(big)})
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := ResizeE(tc.w, tc.h, img, Bilinear); !errors.Is(err, ErrInvalidDimensions) {
				t.Fatalf("ResizeE(%d, %d) error = %v, want ErrInvalidDimensions", tc.w, tc.h, err)
			}
		})
	}
}

// TestResizeEMatchesResize verifies that the error-returning API is a drop-in for the panicking
// one: same image, same pixels.
//
// Covers [ARCH-3.1].
func TestResizeEMatchesResize(t *testing.T) {
	directions := []struct {
		name string
		w, h uint
	}{
		{"down", 4, 4},
		{"up", 16, 16},
		{"width only", 4, 0},
		{"height only", 0, 16},
	}
	for _, c := range imageCases() {
		for _, in := range interpolationFunctions() {
			for _, dir := range directions {
				t.Run(c.name+"/"+in.name+"/"+dir.name, func(t *testing.T) {
					want := Resize(dir.w, dir.h, c.img, in.fn)
					got, err := ResizeE(dir.w, dir.h, c.img, in.fn)
					if err != nil {
						t.Fatalf("ResizeE: %v", err)
					}
					if got.Bounds() != want.Bounds() {
						t.Fatalf("bounds = %v, want %v", got.Bounds(), want.Bounds())
					}
					if typeName(got) != typeName(want) {
						t.Fatalf("type = %s, want %s", typeName(got), typeName(want))
					}
					if !imagesEqual(got, want) {
						t.Fatal("pixels differ between Resize and ResizeE")
					}
				})
			}
		}
	}
}

// TestResizeEEdgeCases covers inputs that must not be treated as errors.
//
// Covers [ARCH-3.1].
func TestResizeEEdgeCases(t *testing.T) {
	empty := image.NewRGBA(image.Rect(0, 0, 0, 0))
	if out, err := ResizeE(8, 8, empty, Bilinear); err != nil || out != image.Image(empty) {
		t.Fatalf("empty image: out=%v err=%v, want the input and no error", out, err)
	}

	img := image.NewRGBA(image.Rect(0, 0, 8, 8))
	if out, err := ResizeE(8, 8, img, Bilinear); err != nil || out != image.Image(img) {
		t.Fatalf("same size: out=%v err=%v, want the input and no error", out, err)
	}
	if out, err := ResizeE(0, 0, img, Bilinear); err != nil || out != image.Image(img) {
		t.Fatalf("0x0: out=%v err=%v, want the input and no error", out, err)
	}

	// Aspect-ratio rounding must not produce an empty image.
	out, err := ResizeE(0, 1, image.NewRGBA(image.Rect(0, 0, 4, 32)), Bilinear)
	if err != nil {
		t.Fatalf("degenerate aspect ratio: %v", err)
	}
	if out.Bounds().Empty() {
		t.Fatalf("degenerate aspect ratio returned empty bounds: %v", out.Bounds())
	}

	// Large but representable dimensions pass validation; they are not allocated here, so the
	// test stays fast.
	if err := validateSize(1<<16, 1<<10, img.Bounds()); err != nil {
		t.Fatalf("large but valid dimensions rejected: %v", err)
	}
}

// TestThumbnailEMatchesThumbnail verifies parity and the unconstrained-max shortcut.
//
// Covers [ARCH-3.1].
func TestThumbnailEMatchesThumbnail(t *testing.T) {
	for _, c := range imageCases() {
		t.Run(c.name, func(t *testing.T) {
			want := Thumbnail(4, 4, c.img, Lanczos3)
			got, err := ThumbnailE(4, 4, c.img, Lanczos3)
			if err != nil {
				t.Fatalf("ThumbnailE: %v", err)
			}
			if got.Bounds() != want.Bounds() || !imagesEqual(got, want) {
				t.Fatalf("ThumbnailE = %v, Thumbnail = %v", got.Bounds(), want.Bounds())
			}

			// Unconstrained maximum: no arithmetic on huge values, return the input.
			unbounded, err := ThumbnailE(^uint(0), ^uint(0), c.img, Lanczos3)
			if err != nil {
				t.Fatalf("unconstrained ThumbnailE: %v", err)
			}
			if unbounded != c.img {
				t.Fatal("unconstrained ThumbnailE should return the input")
			}
		})
	}
}
