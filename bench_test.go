// Copyright (c) 2026, Stefan Schank <github@schank.it>
//
// SPDX-License-Identifier: ISC
//
// See the LICENSE file in the repository root for the full license text.

package resize

import (
	"fmt"
	"image"
	"testing"
)

// benchSink keeps the result alive so the compiler cannot optimise the work away.
var benchSink image.Image

const (
	benchSrcW = 512
	benchSrcH = 512
)

func benchRGBA(w, h int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			i := y*img.Stride + x*4
			img.Pix[i+0] = uint8(x*7 + y*3)
			img.Pix[i+1] = uint8(x*5 ^ y)
			img.Pix[i+2] = uint8(x + y*11)
			img.Pix[i+3] = 0xff
		}
	}
	return img
}

func benchNRGBA(w, h int) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			i := y*img.Stride + x*4
			img.Pix[i+0] = uint8(x*7 + y*3)
			img.Pix[i+1] = uint8(x*5 ^ y)
			img.Pix[i+2] = uint8(x + y*11)
			img.Pix[i+3] = uint8(x ^ y)
		}
	}
	return img
}

func benchGray(w, h int) *image.Gray {
	img := image.NewGray(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Pix[y*img.Stride+x] = uint8(x ^ y)
		}
	}
	return img
}

func benchYCbCr(w, h int) *image.YCbCr {
	img := image.NewYCbCr(image.Rect(0, 0, w, h), image.YCbCrSubsampleRatio420)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Y[y*img.YStride+x] = uint8(x ^ y)
		}
	}
	for i := range img.Cb {
		img.Cb[i] = uint8(i * 3)
		img.Cr[i] = uint8(i * 7)
	}
	return img
}

func benchImages() []struct {
	name string
	img  image.Image
} {
	return []struct {
		name string
		img  image.Image
	}{
		{"RGBA", benchRGBA(benchSrcW, benchSrcH)},
		{"NRGBA", benchNRGBA(benchSrcW, benchSrcH)},
		{"Gray", benchGray(benchSrcW, benchSrcH)},
		{"YCbCr420", benchYCbCr(benchSrcW, benchSrcH)},
	}
}

// BenchmarkResize measures the full resize pipeline for every image type, kernel and direction.
func BenchmarkResize(b *testing.B) {
	interps := []struct {
		name string
		fn   InterpolationFunction
	}{
		{"NearestNeighbor", NearestNeighbor},
		{"Bilinear", Bilinear},
		{"Lanczos3", Lanczos3},
	}

	directions := []struct {
		name string
		w, h uint
	}{
		{"down-128x128", 128, 128},
		{"up-1024x1024", 1024, 1024},
	}

	for _, im := range benchImages() {
		for _, dir := range directions {
			for _, in := range interps {
				name := fmt.Sprintf("%s/%s/%s", im.name, dir.name, in.name)
				b.Run(name, func(b *testing.B) {
					b.ReportAllocs()
					for i := 0; i < b.N; i++ {
						benchSink = Resize(dir.w, dir.h, im.img, in.fn)
					}
				})
			}
		}
	}
}

// BenchmarkThumbnail measures the aspect-ratio preserving downscale helper.
func BenchmarkThumbnail(b *testing.B) {
	for _, im := range benchImages() {
		b.Run(im.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				benchSink = Thumbnail(128, 128, im.img, Lanczos3)
			}
		})
	}
}
