package resize

import (
	"image"
	"image/color"
	"runtime"
	"testing"
)

var img = image.NewGray16(image.Rect(0, 0, 3, 3))

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	img.Set(1, 1, color.White)
}

func Test_Param1(t *testing.T) {
	m := Resize(0, 0, img, NearestNeighbor)
	if m.Bounds() != img.Bounds() {
		t.Fail()
	}
}

func Test_Param2(t *testing.T) {
	m := Resize(100, 0, img, NearestNeighbor)
	if m.Bounds() != image.Rect(0, 0, 100, 100) {
		t.Fail()
	}
}

func Test_ZeroImg(t *testing.T) {
	zeroImg := image.NewGray16(image.Rect(0, 0, 0, 0))

	m := Resize(0, 0, zeroImg, NearestNeighbor)
	if m.Bounds() != zeroImg.Bounds() {
		t.Fail()
	}
}

func Test_HalfZeroImg(t *testing.T) {
	zeroImg := image.NewGray16(image.Rect(0, 0, 0, 100))

	m := Resize(0, 1, zeroImg, NearestNeighbor)
	if m.Bounds() != zeroImg.Bounds() {
		t.Fail()
	}

	m = Resize(1, 0, zeroImg, NearestNeighbor)
	if m.Bounds() != zeroImg.Bounds() {
		t.Fail()
	}
}

func Test_CorrectResize(t *testing.T) {
	zeroImg := image.NewGray16(image.Rect(0, 0, 256, 256))

	m := Resize(60, 0, zeroImg, NearestNeighbor)
	if m.Bounds() != image.Rect(0, 0, 60, 60) {
		t.Fail()
	}
}

func Test_SameColorWithRGBA(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 20, 20))
	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
			img.SetRGBA(x, y, color.RGBA{0x80, 0x80, 0x80, 0xFF})
		}
	}
	out := Resize(10, 10, img, Lanczos3)
	for y := out.Bounds().Min.Y; y < out.Bounds().Max.Y; y++ {
		for x := out.Bounds().Min.X; x < out.Bounds().Max.X; x++ {
			color := out.At(x, y).(color.RGBA)
			if color.R != 0x80 || color.G != 0x80 || color.B != 0x80 || color.A != 0xFF {
				t.Errorf("%+v", color)
			}
		}
	}
}

func Test_SameColorWithNRGBA(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 20, 20))
	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
			img.SetNRGBA(x, y, color.NRGBA{0x80, 0x80, 0x80, 0xFF})
		}
	}
	out := Resize(10, 10, img, Lanczos3)
	for y := out.Bounds().Min.Y; y < out.Bounds().Max.Y; y++ {
		for x := out.Bounds().Min.X; x < out.Bounds().Max.X; x++ {
			color := out.At(x, y).(color.RGBA)
			if color.R != 0x80 || color.G != 0x80 || color.B != 0x80 || color.A != 0xFF {
				t.Errorf("%+v", color)
			}
		}
	}
}

func Test_SameColorWithRGBA64(t *testing.T) {
	img := image.NewRGBA64(image.Rect(0, 0, 20, 20))
	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
			img.SetRGBA64(x, y, color.RGBA64{0x8000, 0x8000, 0x8000, 0xFFFF})
		}
	}
	out := Resize(10, 10, img, Lanczos3)
	for y := out.Bounds().Min.Y; y < out.Bounds().Max.Y; y++ {
		for x := out.Bounds().Min.X; x < out.Bounds().Max.X; x++ {
			color := out.At(x, y).(color.RGBA64)
			if color.R != 0x8000 || color.G != 0x8000 || color.B != 0x8000 || color.A != 0xFFFF {
				t.Errorf("%+v", color)
			}
		}
	}
}

func Test_SameColorWithNRGBA64(t *testing.T) {
	img := image.NewNRGBA64(image.Rect(0, 0, 20, 20))
	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
			img.SetNRGBA64(x, y, color.NRGBA64{0x8000, 0x8000, 0x8000, 0xFFFF})
		}
	}
	out := Resize(10, 10, img, Lanczos3)
	for y := out.Bounds().Min.Y; y < out.Bounds().Max.Y; y++ {
		for x := out.Bounds().Min.X; x < out.Bounds().Max.X; x++ {
			color := out.At(x, y).(color.RGBA64)
			if color.R != 0x8000 || color.G != 0x8000 || color.B != 0x8000 || color.A != 0xFFFF {
				t.Errorf("%+v", color)
			}
		}
	}
}

func Test_SameColorWithGray(t *testing.T) {
	img := image.NewGray(image.Rect(0, 0, 20, 20))
	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
			img.SetGray(x, y, color.Gray{0x80})
		}
	}
	out := Resize(10, 10, img, Lanczos3)
	for y := out.Bounds().Min.Y; y < out.Bounds().Max.Y; y++ {
		for x := out.Bounds().Min.X; x < out.Bounds().Max.X; x++ {
			color := out.At(x, y).(color.Gray)
			if color.Y != 0x80 {
				t.Errorf("%+v", color)
			}
		}
	}
}

func Test_SameColorWithGray16(t *testing.T) {
	img := image.NewGray16(image.Rect(0, 0, 20, 20))
	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
			img.SetGray16(x, y, color.Gray16{0x8000})
		}
	}
	out := Resize(10, 10, img, Lanczos3)
	for y := out.Bounds().Min.Y; y < out.Bounds().Max.Y; y++ {
		for x := out.Bounds().Min.X; x < out.Bounds().Max.X; x++ {
			color := out.At(x, y).(color.Gray16)
			if color.Y != 0x8000 {
				t.Errorf("%+v", color)
			}
		}
	}
}

func Test_Bounds(t *testing.T) {
	img := image.NewRGBA(image.Rect(20, 10, 200, 99))
	out := Resize(80, 80, img, Lanczos2)
	out.At(0, 0)
}

func Test_SameSizeReturnsOriginal(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	out := Resize(0, 0, img, Lanczos2)

	if img != out {
		t.Fail()
	}

	out = Resize(10, 10, img, Lanczos2)

	if img != out {
		t.Fail()
	}
}

func Test_PixelCoordinates(t *testing.T) {
	checkers := image.NewGray(image.Rect(0, 0, 4, 4))
	checkers.Pix = []uint8{
		255, 0, 255, 0,
		0, 255, 0, 255,
		255, 0, 255, 0,
		0, 255, 0, 255,
	}

	resized := Resize(12, 12, checkers, NearestNeighbor).(*image.Gray)

	if resized.Pix[0] != 255 || resized.Pix[1] != 255 || resized.Pix[2] != 255 {
		t.Fail()
	}

	if resized.Pix[3] != 0 || resized.Pix[4] != 0 || resized.Pix[5] != 0 {
		t.Fail()
	}
}

func Test_ResizeWithPremultipliedAlpha(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 1, 4))
	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		// 0x80 = 0.5 * 0xFF.
		img.SetRGBA(0, y, color.RGBA{0x80, 0x80, 0x80, 0x80})
	}

	out := Resize(1, 2, img, MitchellNetravali)

	outputColor := out.At(0, 0).(color.RGBA)
	if outputColor.R != 0x80 {
		t.Fail()
	}
}

func Test_ResizeWithTranslucentColor(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 1, 2))

	// Set the pixel colors to an "invisible green" and white.
	// After resizing, the green shouldn't be visible.
	img.SetNRGBA(0, 0, color.NRGBA{0x00, 0xFF, 0x00, 0x00})
	img.SetNRGBA(0, 1, color.NRGBA{0x00, 0x00, 0x00, 0xFF})

	out := Resize(1, 1, img, Bilinear)

	_, g, _, _ := out.At(0, 0).RGBA()
	if g != 0x00 {
		t.Errorf("%+v", g)
	}
}
