package resize

import (
	"image"
	"image/color"
	"testing"
)

// fuzzImage builds a small image of the requested kind filled from data. Sizes are capped by the
// caller so the fuzzer cannot allocate gigabytes.
func fuzzImage(kind uint8, w, h int, data []byte) image.Image {
	pick := func(i int) uint8 {
		if len(data) == 0 {
			return uint8(i * 37)
		}
		return data[i%len(data)]
	}

	switch kind % 8 {
	case 0:
		m := image.NewRGBA(image.Rect(0, 0, w, h))
		for i := range m.Pix {
			m.Pix[i] = pick(i)
		}
		return m
	case 1:
		m := image.NewNRGBA(image.Rect(0, 0, w, h))
		for i := range m.Pix {
			m.Pix[i] = pick(i)
		}
		return m
	case 2:
		m := image.NewRGBA64(image.Rect(0, 0, w, h))
		for i := range m.Pix {
			m.Pix[i] = pick(i)
		}
		return m
	case 3:
		m := image.NewGray(image.Rect(0, 0, w, h))
		for i := range m.Pix {
			m.Pix[i] = pick(i)
		}
		return m
	case 4:
		m := image.NewGray16(image.Rect(0, 0, w, h))
		for i := range m.Pix {
			m.Pix[i] = pick(i)
		}
		return m
	case 5:
		ratios := []image.YCbCrSubsampleRatio{
			image.YCbCrSubsampleRatio444,
			image.YCbCrSubsampleRatio422,
			image.YCbCrSubsampleRatio420,
			image.YCbCrSubsampleRatio440,
			image.YCbCrSubsampleRatio411,
			image.YCbCrSubsampleRatio410,
		}
		m := image.NewYCbCr(image.Rect(0, 0, w, h), ratios[int(kind/8)%len(ratios)])
		for i := range m.Y {
			m.Y[i] = pick(i)
		}
		for i := range m.Cb {
			m.Cb[i] = pick(i + 1)
			m.Cr[i] = pick(i + 2)
		}
		return m
	case 6:
		m := image.NewPaletted(image.Rect(0, 0, w, h), color.Palette{
			color.Black, color.White, color.RGBA{0x10, 0x20, 0x30, 0xff},
		})
		for i := range m.Pix {
			// Colour indices must stay inside the palette, otherwise image.Paletted.At panics.
			m.Pix[i] = pick(i) % uint8(len(m.Palette))
		}
		return m
	default:
		m := image.NewCMYK(image.Rect(0, 0, w, h))
		for i := range m.Pix {
			m.Pix[i] = pick(i)
		}
		return m
	}
}

// FuzzResize checks that no input combination panics or hangs. Decoded images are adversarial, so
// the invariant is "always returns an image with the requested size", not pixel equality.
func FuzzResize(f *testing.F) {
	f.Add(uint8(0), uint8(8), uint8(8), uint8(4), uint8(4), uint8(2), []byte{1, 2, 3, 4})
	f.Add(uint8(5), uint8(1), uint8(1), uint8(31), uint8(31), uint8(5), []byte{})
	f.Add(uint8(9), uint8(17), uint8(3), uint8(0), uint8(9), uint8(0), []byte{7})
	f.Add(uint8(6), uint8(31), uint8(31), uint8(1), uint8(1), uint8(4), []byte{0})

	f.Fuzz(func(t *testing.T, kind, sw, sh, dw, dh, interp uint8, data []byte) {
		w := int(sw%32) + 1
		h := int(sh%32) + 1
		ow := int(dw % 32)
		oh := int(dh % 32)

		img := fuzzImage(kind, w, h, data)
		fn := InterpolationFunction(int(interp) % 6)

		out := Resize(uint(ow), uint(oh), img, fn)
		if out == nil {
			t.Fatal("Resize returned nil")
		}
		switch {
		case ow == 0 && oh == 0:
			if out != img {
				t.Fatalf("Resize(0,0) should return the input, got %v", out.Bounds())
			}
		case ow == 0:
			if out.Bounds().Dy() != oh {
				t.Fatalf("bounds = %v, want height %d", out.Bounds(), oh)
			}
		case oh == 0:
			if out.Bounds().Dx() != ow {
				t.Fatalf("bounds = %v, want width %d", out.Bounds(), ow)
			}
		default:
			if want := image.Rect(0, 0, ow, oh); out.Bounds() != want {
				t.Fatalf("bounds = %v, want %v", out.Bounds(), want)
			}
		}
		out.At(out.Bounds().Min.X, out.Bounds().Min.Y)

		thumb := Thumbnail(uint(ow), uint(oh), img, fn)
		if thumb == nil {
			t.Fatal("Thumbnail returned nil")
		}
		if thumb.Bounds().Empty() {
			t.Fatalf("Thumbnail returned empty bounds")
		}
	})
}
