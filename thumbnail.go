/*
Copyright (c) 2012, Jan Schlicht <jan.schlicht@gmail.com>

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
	"fmt"
	"image"
	"math"
)

// Thumbnail will downscale provided image to max width and height preserving
// original aspect ratio and using the interpolation function interp.
// It will return original image, without processing it, if original sizes
// are already smaller than provided constraints.
//
// Thumbnail panics if img is nil or the resulting dimensions cannot be allocated; use ThumbnailE
// if you need to handle those cases as errors.
func Thumbnail(maxWidth, maxHeight uint, img image.Image, interp InterpolationFunction) image.Image {
	out, err := ThumbnailE(maxWidth, maxHeight, img, interp)
	if err != nil {
		panic(err)
	}
	return out
}

// ThumbnailE is like Thumbnail but reports invalid input instead of panicking. For every input
// that Thumbnail can handle, the result is identical.
func ThumbnailE(maxWidth, maxHeight uint, img image.Image, interp InterpolationFunction) (image.Image, error) {
	if img == nil {
		return nil, ErrNilImage
	}

	origBounds := img.Bounds()
	origWidth := uint(origBounds.Dx())
	origHeight := uint(origBounds.Dy())
	newWidth, newHeight := origWidth, origHeight

	// Return original image if it have same or smaller size as constraints
	if maxWidth >= origWidth && maxHeight >= origHeight {
		return img, nil
	}

	// The multiplications below stay exact as long as the source dimensions fit into 32 bits. An
	// image that large cannot be allocated on any platform Go supports, so this only guards
	// against integer wraparound in the arithmetic.
	if uint64(origWidth) > math.MaxUint32 || uint64(origHeight) > math.MaxUint32 {
		return nil, fmt.Errorf("%w: source %dx%d is too large", ErrInvalidDimensions, origWidth, origHeight)
	}

	// Preserve aspect ratio
	if origWidth > maxWidth {
		newHeight = uint(uint64(origHeight) * uint64(maxWidth) / uint64(origWidth))
		if newHeight < 1 {
			newHeight = 1
		}
		newWidth = maxWidth
	}

	if newHeight > maxHeight {
		newWidth = uint(uint64(newWidth) * uint64(maxHeight) / uint64(newHeight))
		if newWidth < 1 {
			newWidth = 1
		}
		newHeight = maxHeight
	}

	out, err := ResizeE(newWidth, newHeight, img, interp)
	if err != nil {
		return nil, err
	}
	return out, nil
}
