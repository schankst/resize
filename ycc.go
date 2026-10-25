/*
Copyright (c) 2014, Charlie Vieth <charlie.vieth@gmail.com>

Permission to use, copy, modify, and/or distribute this software for any purpose
with or without fee is hereby granted, provided that the above copyright notice
and this permission notice appear in all copies.

THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES WITH
REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF MERCHANTABILITY AND
FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR ANY SPECIAL, DIRECT,
INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES WHATSOEVER RESULTING FROM LOSS
OF USE, DATA OR PROFITS, WHETHER IN AN ACTION OF CONTRACT, NEGLIGENCE OR OTHER
TORTIOUS ACTION, ARISING OUT OF OR IN CONNECTION WITH THE USE OR PERFORMANCE OF
THIS SOFTWARE.
*/

package resize

import (
	"image"
	"image/color"
)

// ycc is an in memory YCbCr image.  The Y, Cb and Cr samples are held in a
// single slice to increase resizing performance.
type ycc struct {
	// Pix holds the image's pixels, in Y, Cb, Cr order. The pixel at
	// (x, y) starts at Pix[(y-Rect.Min.Y)*Stride + (x-Rect.Min.X)*3].
	Pix []uint8
	// Stride is the Pix stride (in bytes) between vertically adjacent pixels.
	Stride int
	// Rect is the image's bounds.
	Rect image.Rectangle
	// SubsampleRatio is the subsample ratio of the original YCbCr image.
	SubsampleRatio image.YCbCrSubsampleRatio
}

// PixOffset returns the index of the first element of Pix that corresponds to
// the pixel at (x, y).
func (p *ycc) PixOffset(x, y int) int {
	return (y-p.Rect.Min.Y)*p.Stride + (x-p.Rect.Min.X)*3
}

// yccRow returns row y of the image as w*3 interleaved Y, Cb, Cr bytes. The returned slice aliases
// Pix, so it must be treated as read-only.
func (p *ycc) yccRow(y int, _ []byte) []byte {
	w := p.Bounds().Dx()
	start := (y - p.Rect.Min.Y) * p.Stride
	return p.Pix[start : start+w*3]
}

func (p *ycc) width() int { return p.Bounds().Dx() }

// yccRowSource yields one interleaved Y, Cb, Cr row at a time. Streaming the rows keeps the
// resizing passes from materialising a full-size copy of a YCbCr image.
type yccRowSource interface {
	// yccRow returns row y (relative to the image origin) as width*3 interleaved bytes. The
	// result may alias internal memory; callers must not modify it.
	yccRow(y int, scratch []byte) []byte
	// width reports the number of pixels per row.
	width() int
}

// ycbcrRowSource reads rows straight out of an *image.YCbCr, expanding subsampled chroma on the
// fly. It produces exactly the bytes imageYCbCrToYCC used to write into a full-size buffer.
type ycbcrRowSource struct {
	img *image.YCbCr
}

func (s ycbcrRowSource) width() int { return s.img.Bounds().Dx() }

func (s ycbcrRowSource) yccRow(y int, scratch []byte) []byte {
	m := s.img
	w := m.Bounds().Dx()
	if cap(scratch) < 3*w {
		scratch = make([]byte, 3*w)
	}
	row := scratch[:3*w]

	// Index math matches the old convertToYCC* helpers: offsets are relative to the image origin.
	var cyShift, cxShift uint
	switch m.SubsampleRatio {
	case image.YCbCrSubsampleRatio422:
		cxShift = 1
	case image.YCbCrSubsampleRatio420:
		cyShift, cxShift = 1, 1
	case image.YCbCrSubsampleRatio440:
		cyShift = 1
	case image.YCbCrSubsampleRatio411:
		cxShift = 2
	case image.YCbCrSubsampleRatio410:
		cyShift, cxShift = 1, 2
	}

	yy := y * m.YStride
	cy := (y >> cyShift) * m.CStride
	for x := 0; x < w; x++ {
		ci := cy + (x >> cxShift)
		row[x*3+0] = m.Y[yy+x]
		row[x*3+1] = m.Cb[ci]
		row[x*3+2] = m.Cr[ci]
	}
	return row
}

func (p *ycc) Bounds() image.Rectangle {
	return p.Rect
}

func (p *ycc) ColorModel() color.Model {
	return color.YCbCrModel
}

func (p *ycc) At(x, y int) color.Color {
	if !(image.Point{x, y}.In(p.Rect)) {
		return color.YCbCr{}
	}
	i := p.PixOffset(x, y)
	return color.YCbCr{
		p.Pix[i+0],
		p.Pix[i+1],
		p.Pix[i+2],
	}
}

func (p *ycc) Opaque() bool {
	return true
}

// SubImage returns an image representing the portion of the image p visible
// through r. The returned value shares pixels with the original image.
func (p *ycc) SubImage(r image.Rectangle) image.Image {
	r = r.Intersect(p.Rect)
	if r.Empty() {
		return &ycc{SubsampleRatio: p.SubsampleRatio}
	}
	i := p.PixOffset(r.Min.X, r.Min.Y)
	return &ycc{
		Pix:            p.Pix[i:],
		Stride:         p.Stride,
		Rect:           r,
		SubsampleRatio: p.SubsampleRatio,
	}
}

// newYCC returns a new ycc with the given bounds and subsample ratio.
func newYCC(r image.Rectangle, s image.YCbCrSubsampleRatio) *ycc {
	w, h := r.Dx(), r.Dy()
	buf := make([]uint8, 3*w*h)
	return &ycc{Pix: buf, Stride: 3 * w, Rect: r, SubsampleRatio: s}
}

// Copy of image.YCbCrSubsampleRatio constants - this allows us to support
// older versions of Go where these constants are not defined (i.e. Go 1.4)
const (
	ycbcrSubsampleRatio444 image.YCbCrSubsampleRatio = iota
	ycbcrSubsampleRatio422
	ycbcrSubsampleRatio420
	ycbcrSubsampleRatio440
	ycbcrSubsampleRatio411
	ycbcrSubsampleRatio410
)

// YCbCr converts ycc to a YCbCr image with the same subsample ratio
// as the YCbCr image that ycc was generated from.
func (p *ycc) YCbCr() *image.YCbCr {
	ycbcr := image.NewYCbCr(p.Rect, p.SubsampleRatio)
	switch ycbcr.SubsampleRatio {
	case ycbcrSubsampleRatio422:
		return p.ycbcr422(ycbcr)
	case ycbcrSubsampleRatio420:
		return p.ycbcr420(ycbcr)
	case ycbcrSubsampleRatio440:
		return p.ycbcr440(ycbcr)
	case ycbcrSubsampleRatio444:
		return p.ycbcr444(ycbcr)
	case ycbcrSubsampleRatio411:
		return p.ycbcr411(ycbcr)
	case ycbcrSubsampleRatio410:
		return p.ycbcr410(ycbcr)
	}
	return ycbcr
}

func (p *ycc) ycbcr422(ycbcr *image.YCbCr) *image.YCbCr {
	var off int
	Pix := p.Pix
	Y := ycbcr.Y
	Cb := ycbcr.Cb
	Cr := ycbcr.Cr
	for y := 0; y < ycbcr.Rect.Max.Y-ycbcr.Rect.Min.Y; y++ {
		yy := y * ycbcr.YStride
		cy := y * ycbcr.CStride
		for x := 0; x < ycbcr.Rect.Max.X-ycbcr.Rect.Min.X; x++ {
			ci := cy + x/2
			Y[yy+x] = Pix[off+0]
			Cb[ci] = Pix[off+1]
			Cr[ci] = Pix[off+2]
			off += 3
		}
	}
	return ycbcr
}

func (p *ycc) ycbcr420(ycbcr *image.YCbCr) *image.YCbCr {
	var off int
	Pix := p.Pix
	Y := ycbcr.Y
	Cb := ycbcr.Cb
	Cr := ycbcr.Cr
	for y := 0; y < ycbcr.Rect.Max.Y-ycbcr.Rect.Min.Y; y++ {
		yy := y * ycbcr.YStride
		cy := (y / 2) * ycbcr.CStride
		for x := 0; x < ycbcr.Rect.Max.X-ycbcr.Rect.Min.X; x++ {
			ci := cy + x/2
			Y[yy+x] = Pix[off+0]
			Cb[ci] = Pix[off+1]
			Cr[ci] = Pix[off+2]
			off += 3
		}
	}
	return ycbcr
}

func (p *ycc) ycbcr440(ycbcr *image.YCbCr) *image.YCbCr {
	var off int
	Pix := p.Pix
	Y := ycbcr.Y
	Cb := ycbcr.Cb
	Cr := ycbcr.Cr
	for y := 0; y < ycbcr.Rect.Max.Y-ycbcr.Rect.Min.Y; y++ {
		yy := y * ycbcr.YStride
		cy := (y / 2) * ycbcr.CStride
		for x := 0; x < ycbcr.Rect.Max.X-ycbcr.Rect.Min.X; x++ {
			ci := cy + x
			Y[yy+x] = Pix[off+0]
			Cb[ci] = Pix[off+1]
			Cr[ci] = Pix[off+2]
			off += 3
		}
	}
	return ycbcr
}

func (p *ycc) ycbcr444(ycbcr *image.YCbCr) *image.YCbCr {
	var off int
	Pix := p.Pix
	Y := ycbcr.Y
	Cb := ycbcr.Cb
	Cr := ycbcr.Cr
	for y := 0; y < ycbcr.Rect.Max.Y-ycbcr.Rect.Min.Y; y++ {
		yy := y * ycbcr.YStride
		cy := y * ycbcr.CStride
		for x := 0; x < ycbcr.Rect.Max.X-ycbcr.Rect.Min.X; x++ {
			ci := cy + x
			Y[yy+x] = Pix[off+0]
			Cb[ci] = Pix[off+1]
			Cr[ci] = Pix[off+2]
			off += 3
		}
	}
	return ycbcr
}

func (p *ycc) ycbcr411(ycbcr *image.YCbCr) *image.YCbCr {
	var off int
	Pix := p.Pix
	Y := ycbcr.Y
	Cb := ycbcr.Cb
	Cr := ycbcr.Cr
	for y := 0; y < ycbcr.Rect.Max.Y-ycbcr.Rect.Min.Y; y++ {
		yy := y * ycbcr.YStride
		cy := y * ycbcr.CStride
		for x := 0; x < ycbcr.Rect.Max.X-ycbcr.Rect.Min.X; x++ {
			ci := cy + x/4
			Y[yy+x] = Pix[off+0]
			Cb[ci] = Pix[off+1]
			Cr[ci] = Pix[off+2]
			off += 3
		}
	}
	return ycbcr
}

func (p *ycc) ycbcr410(ycbcr *image.YCbCr) *image.YCbCr {
	var off int
	Pix := p.Pix
	Y := ycbcr.Y
	Cb := ycbcr.Cb
	Cr := ycbcr.Cr
	for y := 0; y < ycbcr.Rect.Max.Y-ycbcr.Rect.Min.Y; y++ {
		yy := y * ycbcr.YStride
		cy := (y / 2) * ycbcr.CStride
		for x := 0; x < ycbcr.Rect.Max.X-ycbcr.Rect.Min.X; x++ {
			ci := cy + x/4
			Y[yy+x] = Pix[off+0]
			Cb[ci] = Pix[off+1]
			Cr[ci] = Pix[off+2]
			off += 3
		}
	}
	return ycbcr
}
