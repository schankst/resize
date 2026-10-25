// Copyright (c) 2026, Stefan Schank <github@schank.it>
//
// SPDX-License-Identifier: ISC
//
// See the LICENSE file in the repository root for the full license text.

package resize

import (
	"image"
	"image/color"
	"testing"
)

// probeImage is deliberately not one of the concrete types that has a fast path, so it exercises
// the generic converter (which goes through image.Image.At).
type probeImage struct{ r image.Rectangle }

func (p probeImage) ColorModel() color.Model { return color.RGBA64Model }
func (p probeImage) Bounds() image.Rectangle { return p.r }
func (p probeImage) At(x, y int) color.Color {
	return color.RGBA64{R: uint16(x * 257), G: uint16(y * 257), B: 0x8000, A: 0xffff}
}

// fill writes a deterministic pattern into img, so results are not uniform.
func fill(img image.Image, w, h int) {
	rect := image.Rect(0, 0, w, h)
	switch m := img.(type) {
	case *image.RGBA:
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				m.SetRGBA(rect.Min.X+x, rect.Min.Y+y, color.RGBA{uint8(x * 9), uint8(y * 7), uint8(x ^ y), 0xff})
			}
		}
	case *image.NRGBA:
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				m.SetNRGBA(rect.Min.X+x, rect.Min.Y+y, color.NRGBA{uint8(x * 9), uint8(y * 7), uint8(x ^ y), uint8(128 + x)})
			}
		}
	case *image.RGBA64:
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				m.SetRGBA64(rect.Min.X+x, rect.Min.Y+y, color.RGBA64{uint16(x * 1000), uint16(y * 1000), 0x8000, 0xffff})
			}
		}
	case *image.NRGBA64:
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				m.SetNRGBA64(rect.Min.X+x, rect.Min.Y+y, color.NRGBA64{uint16(x * 1000), uint16(y * 1000), 0x8000, 0xc000})
			}
		}
	case *image.Gray:
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				m.SetGray(rect.Min.X+x, rect.Min.Y+y, color.Gray{uint8(x * 11)})
			}
		}
	case *image.Gray16:
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				m.SetGray16(rect.Min.X+x, rect.Min.Y+y, color.Gray16{uint16(x * 2000)})
			}
		}
	case *image.YCbCr:
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				yi := m.YOffset(rect.Min.X+x, rect.Min.Y+y)
				ci := m.COffset(rect.Min.X+x, rect.Min.Y+y)
				m.Y[yi] = uint8(x*13 + y)
				m.Cb[ci] = uint8(x * 5)
				m.Cr[ci] = uint8(y * 5)
			}
		}
	case *image.Paletted:
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				m.SetColorIndex(rect.Min.X+x, rect.Min.Y+y, uint8(x+y)%uint8(len(m.Palette)))
			}
		}
	case *image.CMYK:
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				m.SetCMYK(rect.Min.X+x, rect.Min.Y+y, color.CMYK{uint8(x), uint8(y), 0, 0})
			}
		}
	}
}

type imageCase struct {
	name string
	img  image.Image
	// wantFiltered and wantNearest are the concrete result types. They differ for NRGBA inputs:
	// the filtered paths premultiply into RGBA, the nearest path keeps NRGBA.
	wantFiltered string
	wantNearest  string
}

func imageCases() []imageCase {
	rect := image.Rect(0, 0, 8, 8)
	cases := []imageCase{
		{"RGBA", image.NewRGBA(rect), "*image.RGBA", "*image.RGBA"},
		{"NRGBA", image.NewNRGBA(rect), "*image.RGBA", "*image.NRGBA"},
		{"RGBA64", image.NewRGBA64(rect), "*image.RGBA64", "*image.RGBA64"},
		{"NRGBA64", image.NewNRGBA64(rect), "*image.RGBA64", "*image.NRGBA64"},
		{"Gray", image.NewGray(rect), "*image.Gray", "*image.Gray"},
		{"Gray16", image.NewGray16(rect), "*image.Gray16", "*image.Gray16"},
		{"Paletted", image.NewPaletted(rect, color.Palette{color.Black, color.White, color.RGBA{0, 0, 0xff, 0xff}}), "*image.RGBA64", "*image.RGBA64"},
		{"CMYK", image.NewCMYK(rect), "*image.RGBA64", "*image.RGBA64"},
		{"Generic", probeImage{rect}, "*image.RGBA64", "*image.RGBA64"},
	}
	for _, ratio := range []struct {
		name string
		r    image.YCbCrSubsampleRatio
	}{
		{"YCbCr444", image.YCbCrSubsampleRatio444},
		{"YCbCr422", image.YCbCrSubsampleRatio422},
		{"YCbCr420", image.YCbCrSubsampleRatio420},
		{"YCbCr440", image.YCbCrSubsampleRatio440},
		{"YCbCr411", image.YCbCrSubsampleRatio411},
		{"YCbCr410", image.YCbCrSubsampleRatio410},
	} {
		cases = append(cases, imageCase{ratio.name, image.NewYCbCr(rect, ratio.r), "*image.YCbCr", "*image.YCbCr"})
	}
	for _, c := range cases {
		fill(c.img, 8, 8)
	}
	return cases
}

func interpolationFunctions() []struct {
	name string
	fn   InterpolationFunction
} {
	return []struct {
		name string
		fn   InterpolationFunction
	}{
		{"NearestNeighbor", NearestNeighbor},
		{"Bilinear", Bilinear},
		{"Bicubic", Bicubic},
		{"MitchellNetravali", MitchellNetravali},
		{"Lanczos2", Lanczos2},
		{"Lanczos3", Lanczos3},
	}
}

// TestResizeMatrix pins the behaviour of every input type with every kernel and direction,
// including the concrete result type.
//
// Covers [ARCH-3.2].
func TestResizeMatrix(t *testing.T) {
	directions := []struct {
		name string
		w, h uint
	}{
		{"down", 4, 4},
		{"up", 16, 16},
		{"wide", 16, 4},
		{"tall", 4, 16},
	}

	for _, c := range imageCases() {
		for _, in := range interpolationFunctions() {
			for _, dir := range directions {
				name := c.name + "/" + in.name + "/" + dir.name
				t.Run(name, func(t *testing.T) {
					out := Resize(dir.w, dir.h, c.img, in.fn)
					if out == nil {
						t.Fatal("Resize returned nil")
					}
					want := image.Rect(0, 0, int(dir.w), int(dir.h))
					if out.Bounds() != want {
						t.Fatalf("bounds = %v, want %v", out.Bounds(), want)
					}
					gotType := typeName(out)
					expect := c.wantFiltered
					if in.fn == NearestNeighbor {
						expect = c.wantNearest
					}
					if gotType != expect {
						t.Fatalf("result type = %s, want %s", gotType, expect)
					}
					// At must work on every pixel of the result.
					for y := out.Bounds().Min.Y; y < out.Bounds().Max.Y; y++ {
						for x := out.Bounds().Min.X; x < out.Bounds().Max.X; x++ {
							out.At(x, y)
						}
					}
				})
			}
		}
	}
}

// TestResizeSubImageBounds covers inputs whose Bounds().Min is not (0,0). The pixel data of a
// SubImage starts at that offset, which the fast paths index into directly.
//
// Covers [ARCH-3.4].
func TestResizeSubImageBounds(t *testing.T) {
	base := image.Rect(0, 0, 32, 32)
	inner := image.Rect(5, 7, 27, 25)
	img := image.NewRGBA(base)
	fill(img, 32, 32)

	sub := img.SubImage(inner).(*image.RGBA)
	out := Resize(11, 9, sub, Lanczos3)
	if want := image.Rect(0, 0, 11, 9); out.Bounds() != want {
		t.Fatalf("bounds = %v, want %v", out.Bounds(), want)
	}

	// Same result as resizing an equivalent copy with Min == (0,0).
	crop := image.NewRGBA(image.Rect(0, 0, inner.Dx(), inner.Dy()))
	for y := 0; y < inner.Dy(); y++ {
		start := (inner.Min.Y+y-img.Rect.Min.Y)*img.Stride + (inner.Min.X-img.Rect.Min.X)*4
		copy(crop.Pix[y*crop.Stride:(y+1)*crop.Stride], img.Pix[start:start+inner.Dx()*4])
	}
	ref := Resize(11, 9, crop, Lanczos3)
	if !imagesEqual(out, ref) {
		t.Fatal("resizing a SubImage differs from resizing an equivalent copy")
	}
}

// TestResizeAspectRatio covers the width==0 and height==0 shorthand.
//
// Covers [ARCH-3.1].
func TestResizeAspectRatio(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 64, 32))
	fill(img, 64, 32)

	for _, tc := range []struct {
		name string
		w, h uint
		want image.Rectangle
	}{
		{"width only", 32, 0, image.Rect(0, 0, 32, 16)},
		{"height only", 0, 16, image.Rect(0, 0, 32, 16)},
		{"both zero returns input", 0, 0, image.Rect(0, 0, 64, 32)},
		{"same size returns input", 64, 32, image.Rect(0, 0, 64, 32)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out := Resize(tc.w, tc.h, img, Bilinear)
			if out.Bounds() != tc.want {
				t.Fatalf("bounds = %v, want %v", out.Bounds(), tc.want)
			}
		})
	}
}

// TestKernels exercises every interpolation kernel, including their zero crossings.
//
// Covers [ARCH-3.3].
func TestKernels(t *testing.T) {
	for _, in := range interpolationFunctions() {
		taps, k := in.fn.kernel()
		if taps <= 0 {
			t.Fatalf("%s: taps = %d", in.name, taps)
		}
		if k(0) == 0 {
			t.Fatalf("%s: kernel(0) = 0", in.name)
		}
		for _, x := range []float64{-3, -1.5, -0.5, 0, 0.5, 1.5, 3} {
			k(x)
		}
	}

	// An unknown value falls back to nearest neighbour.
	taps, _ := InterpolationFunction(42).kernel()
	if taps != 2 {
		t.Fatalf("fallback taps = %d, want 2", taps)
	}
}

// TestYCCAccessors covers the image.Image implementation of the internal ycc type.
//
// Covers [ARCH-3.9].
func TestYCCAccessors(t *testing.T) {
	p := newYCC(image.Rect(0, 0, 4, 3), image.YCbCrSubsampleRatio420)
	if p.Bounds() != image.Rect(0, 0, 4, 3) {
		t.Fatalf("bounds = %v", p.Bounds())
	}
	if p.ColorModel() != color.YCbCrModel {
		t.Fatalf("color model = %v", p.ColorModel())
	}
	if !p.Opaque() {
		t.Fatal("ycc should be opaque")
	}
	if off := p.PixOffset(2, 1); off != 1*p.Stride+2*3 {
		t.Fatalf("PixOffset = %d", off)
	}
	_ = p.At(0, 0)
	// Out-of-bounds At must not panic.
	_ = p.At(-1, -1)
	_ = p.At(4, 3)

	sub := p.SubImage(image.Rect(1, 1, 3, 3))
	if sub.Bounds() != image.Rect(1, 1, 3, 3) {
		t.Fatalf("sub bounds = %v", sub.Bounds())
	}
	if empty := p.SubImage(image.Rect(10, 10, 10, 10)); !empty.Bounds().Empty() {
		t.Fatalf("empty sub bounds = %v", empty.Bounds())
	}
}

// TestThumbnailMatrix covers the thumbnail helper for every input type.
//
// Covers [ARCH-3.1].
func TestThumbnailMatrix(t *testing.T) {
	for _, c := range imageCases() {
		t.Run(c.name, func(t *testing.T) {
			out := Thumbnail(4, 4, c.img, Lanczos3)
			if b := out.Bounds(); b.Dx() > 4 || b.Dy() > 4 || b.Empty() {
				t.Fatalf("bounds = %v", b)
			}
		})
	}
}

func typeName(v any) string {
	switch v.(type) {
	case *image.RGBA:
		return "*image.RGBA"
	case *image.NRGBA:
		return "*image.NRGBA"
	case *image.RGBA64:
		return "*image.RGBA64"
	case *image.NRGBA64:
		return "*image.NRGBA64"
	case *image.Gray:
		return "*image.Gray"
	case *image.Gray16:
		return "*image.Gray16"
	case *image.YCbCr:
		return "*image.YCbCr"
	default:
		return "other"
	}
}

func imagesEqual(a, b image.Image) bool {
	if a.Bounds() != b.Bounds() {
		return false
	}
	for y := a.Bounds().Min.Y; y < a.Bounds().Max.Y; y++ {
		for x := a.Bounds().Min.X; x < a.Bounds().Max.X; x++ {
			ar, ag, ab, aa := a.At(x, y).RGBA()
			br, bg, bb, ba := b.At(x, y).RGBA()
			if ar != br || ag != bg || ab != bb || aa != ba {
				return false
			}
		}
	}
	return true
}
