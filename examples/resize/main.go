// Copyright (c) 2026, Stefan Schank <github@schank.it>
//
// SPDX-License-Identifier: ISC
//
// See the LICENSE file in the repository root for the full license text.

// Command resize scales an image and re-encodes it.
//
//	go run ./examples/resize -in photo.jpg -out thumb.png -width 800
//	go run ./examples/resize -in photo.jpg -out thumb.jpg -width 400 -height 300
//	go run ./examples/resize -in logo.png -out logo.png -width 0 -height 64 -kernel bilinear
package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"strings"

	// Register the decoders used by image.Decode.
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	"github.com/schankst/resize"
)

func main() {
	in := flag.String("in", "", "input image (JPEG, PNG or GIF)")
	out := flag.String("out", "", "output image; the extension picks the encoder: .jpg, .png or .gif")
	width := flag.Uint("width", 800, "target width in pixels; 0 derives it from the aspect ratio")
	height := flag.Uint("height", 0, "target height in pixels; 0 derives it from the aspect ratio")
	kernel := flag.String("kernel", "lanczos3", "nearest, bilinear, bicubic, mitchell, lanczos2 or lanczos3")
	flag.Parse()

	if *in == "" || *out == "" {
		fmt.Fprintln(os.Stderr, "usage: resize -in <input> -out <output> [-width N] [-height N] [-kernel K]")
		flag.PrintDefaults()
		os.Exit(2)
	}

	interp, err := parseKernel(*kernel)
	if err != nil {
		log.Fatal(err)
	}

	src, format, err := decode(*in)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("input  %s: %T %v", format, src, src.Bounds())

	dst, err := resize.ResizeE(*width, *height, src, interp)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("output %T %v (%s)", dst, dst.Bounds(), *kernel)

	if err := encode(*out, dst); err != nil {
		log.Fatal(err)
	}
}

func decode(path string) (image.Image, string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, "", err
	}
	defer f.Close()

	img, format, err := image.Decode(f)
	if err != nil {
		return nil, "", fmt.Errorf("decode %s: %w", path, err)
	}
	return img, format, nil
}

func encode(path string, img image.Image) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}

	switch strings.ToLower(filepath.Ext(path)) {
	case ".jpg", ".jpeg":
		// JPEG has no alpha channel: flatten onto white first.
		flat := image.NewRGBA(img.Bounds())
		draw.Draw(flat, flat.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Src)
		draw.Draw(flat, flat.Bounds(), img, img.Bounds().Min, draw.Over)
		err = jpeg.Encode(f, flat, &jpeg.Options{Quality: 90})
	case ".png":
		err = png.Encode(f, img)
	case ".gif":
		err = gif.Encode(f, img, nil)
	default:
		f.Close()
		return fmt.Errorf("unsupported output extension %q: use .jpg, .png or .gif", filepath.Ext(path))
	}

	if closeErr := f.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return fmt.Errorf("encode %s: %w", path, err)
	}
	return nil
}

func parseKernel(name string) (resize.InterpolationFunction, error) {
	switch strings.ToLower(name) {
	case "nearest":
		return resize.NearestNeighbor, nil
	case "bilinear":
		return resize.Bilinear, nil
	case "bicubic":
		return resize.Bicubic, nil
	case "mitchell":
		return resize.MitchellNetravali, nil
	case "lanczos2":
		return resize.Lanczos2, nil
	case "lanczos3":
		return resize.Lanczos3, nil
	}
	return 0, fmt.Errorf("unknown kernel %q", name)
}
