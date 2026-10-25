# resize

[![CI](https://github.com/schankst/resize/actions/workflows/ci.yml/badge.svg)](https://github.com/schankst/resize/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/schankst/resize.svg)](https://pkg.go.dev/github.com/schankst/resize)
[![License: ISC](https://img.shields.io/badge/license-ISC-blue.svg)](LICENSE)

Image resizing for the Go programming language with common interpolation methods.

This repository is a **maintained fork** of the original [`nfnt/resize`](https://github.com/nfnt/resize),
which was archived by its author in 2022. The API is unchanged; the import path is
`github.com/schankst/resize`.

## Features

* Six interpolation kernels, from fast to high quality
* Works with any `image.Image`; fast paths for the common concrete types — see
  [Supported image formats](#supported-image-formats)
* Parallelized across `GOMAXPROCS`
* **Zero external dependencies** — only the Go standard library
* Pure Go: cross-compiles to every target without cgo

## Installation

```bash
go get github.com/schankst/resize
```

Requires Go 1.18 or newer.

### A note on releases

Only `v1.0.3` and newer declare the module path `github.com/schankst/resize`. Earlier releases
declared `github.com/nfnt/resize`, so `go get` of this path fails for them; `v1.0.2` is kept as the
last release that still works with a `replace` directive. The `v1.0.0` tag was moved after
publication and `v1.0.1` was withdrawn — the Go module proxy keeps serving its cached copies under
the original contents.

## Migrating from `nfnt/resize`

The API is identical, so only the import path changes:

```bash
go get github.com/schankst/resize
grep -rl 'github.com/nfnt/resize' --include='*.go' . | xargs sed -i 's|github.com/nfnt/resize|github.com/schankst/resize|g'
go mod tidy
```

That is `github.com/nfnt/resize` → `github.com/schankst/resize` in imports and call sites; nothing
else.

## Usage

The package provides two operations, each in a panicking and an error-returning flavour:

```go
resize.Resize(width, height uint, img image.Image, interp resize.InterpolationFunction) image.Image
resize.ResizeE(width, height uint, img image.Image, interp resize.InterpolationFunction) (image.Image, error)

resize.Thumbnail(maxWidth, maxHeight uint, img image.Image, interp resize.InterpolationFunction) image.Image
resize.ThumbnailE(maxWidth, maxHeight uint, img image.Image, interp resize.InterpolationFunction) (image.Image, error)
```

`Resize` scales to the given dimensions. If either `width` or `height` is `0`, the value is
calculated from the aspect ratio of the input; if that calculation rounds down to zero, the
dimension is clamped to `1` so the result is never empty. `Thumbnail` downscales preserving the
aspect ratio, never upscales, and returns the input unchanged if it already fits.

```go
package main

import (
	"image/jpeg"
	"log"
	"os"

	"github.com/schankst/resize"
)

func main() {
	in, err := os.Open("input.jpg")
	if err != nil {
		log.Fatal(err)
	}
	defer in.Close()

	img, err := jpeg.Decode(in)
	if err != nil {
		log.Fatal(err)
	}

	out := resize.Resize(800, 0, img, resize.Lanczos3)

	f, err := os.Create("output.jpg")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	if err := jpeg.Encode(f, out, nil); err != nil {
		log.Fatal(err)
	}
}
```

### Error handling

`Resize` and `Thumbnail` panic on invalid input — a `nil` image, or dimensions that cannot be
represented or allocated. Use the `E` variants if you want an error instead:

```go
out, err := resize.ResizeE(800, 0, src, resize.Lanczos3)
switch {
case errors.Is(err, resize.ErrNilImage):
	// src was nil
case errors.Is(err, resize.ErrInvalidDimensions):
	// dimensions overflow this platform or exceed resize.MaxDimension
case err != nil:
	return err
}
```

For every input the panicking functions accept, the `E` variants return exactly the same image.
`MaxDimension` (`math.MaxInt32`) is a representability limit for cross-architecture consistency, not
a memory limit: an image that fits the limit can still be far too large for the machine.

### Interpolation functions

From fastest to slowest:

| Constant | Notes |
|---|---|
| `NearestNeighbor` | No filtering. Only sane for pixel art and masks. |
| `Bilinear` | Cheap, soft results. Good for previews. |
| `Bicubic` | Sharper than bilinear, slight ringing. |
| `MitchellNetravali` | Balanced, works well for photos. |
| `Lanczos2` | Sharper, some ringing. |
| `Lanczos3` | Sharpest of the built-in kernels, strongest ringing. |

For downscaling photos, `Lanczos3` or `MitchellNetravali` are the usual choices.

### Supported image formats

`resize` neither decodes nor encodes: it operates on an already decoded `image.Image`. What you get
therefore depends on the decoder you use. In practice:

| Format | Decoder | Decoded type |
|---|---|---|
| JPEG | `image/jpeg` (standard library) | `*image.YCbCr`; grayscale files become `*image.Gray`, some CMYK files `*image.CMYK` |
| PNG | `image/png` | `*image.RGBA` (truecolour), `*image.NRGBA` (with alpha), `*image.RGBA64`, `*image.NRGBA64`, `*image.Gray`, `*image.Gray16`, or `*image.Paletted` for indexed PNGs |
| GIF | `image/gif` | `*image.Paletted` |
| TIFF | `golang.org/x/image/tiff` | `*image.RGBA`, `*image.NRGBA`, `*image.Gray` or `*image.Paletted`, depending on the file |
| BMP | `golang.org/x/image/bmp` | `*image.RGBA` or `*image.Paletted` |
| WebP | `golang.org/x/image/webp` (decode only) | `*image.YCbCr` (lossy) or `*image.NRGBA` (lossless) |

Decoders must be registered through blank imports, otherwise `image.Decode` fails with
`image: unknown format`:

```go
import (
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/webp"
)
```

Upstream issues [`#68`](https://github.com/nfnt/resize/issues/68) and
[`#71`](https://github.com/nfnt/resize/issues/71) were exactly this: a missing import, not a library
bug.

#### Fast paths

These concrete types are filtered directly on their pixel buffers:

`*image.RGBA`, `*image.NRGBA`, `*image.RGBA64`, `*image.NRGBA64`, `*image.Gray`, `*image.Gray16`,
and `*image.YCbCr` with any of the six subsample ratios (4:4:4, 4:2:2, 4:2:0, 4:4:0, 4:1:1, 4:1:0).

Everything else — `*image.Paletted`, `*image.CMYK`, `*image.Alpha`, `*image.Alpha16`,
`*image.NYCbCrA`, `image.Uniform` and any custom `image.Image` — is read pixel by pixel through
`At()` and processed at 16-bit precision. That is correct but slower, and the result is always
`*image.RGBA64`.

#### Not supported

* **No decoders of its own.** Formats without a Go decoder cannot be used as input at all: HEIC/HEIF,
  AVIF, JPEG XL, camera RAW and similar. Convert to PNG or JPEG first, or use a cgo binding.
* **Vector formats.** SVG is not a raster format; rasterise it first.
* **Animated or multi-page files.** One frame per call. For an animated GIF, decode with
  `gif.DecodeAll`, resize each `gif.GIF.Image[i]` yourself, and reassemble the animation. For
  multi-page TIFF, `x/image/tiff` decodes the first page only.
* **Metadata.** EXIF orientation is not applied, and EXIF/ICC/DPI are not carried over — see below.

#### Encoding

The same applies on the way out. `image/jpeg`, `image/png`, `image/gif`,
`golang.org/x/image/tiff` and `golang.org/x/image/bmp` all provide encoders; WebP is decode-only.
`image/jpeg` cannot store alpha, so composite the result onto a background first if the source has
transparency.

A complete program is in [`examples/resize`](examples/resize).

### Metadata

Resizing always produces a new pixel buffer. **EXIF, ICC profiles and DPI information are not
carried over.** If you need to preserve them, copy them over yourself or use a decoder/encoder
that exposes them.

### Colors and alpha

* Inputs with non-premultiplied alpha (`NRGBA`, `NRGBA64`) are premultiplied for filtering.
  The filtered kernels therefore return `*image.RGBA`/`*image.RGBA64`, i.e. premultiplied alpha.
  Call `draw.Draw` onto an `image.NRGBA` if you need straight alpha again.
* `NearestNeighbor` is the exception: it keeps the input format and returns
  `*image.NRGBA`/`*image.NRGBA64` unchanged in that regard, with no premultiplication.
* `image.Paletted` and `image.CMYK` inputs take the generic 16-bit path: correct, but slower and
  the result is always `*image.RGBA64`.
* Resampling is not gamma-correct, matching the behaviour of the original package.

### YCbCr

JPEGs decode to `*image.YCbCr`. Subsample ratios 4:4:4, 4:2:2, 4:2:0, 4:4:0, 4:1:1 and 4:1:0 are
read row by row and never copied as a whole; output is always 4:4:4, so resizing cannot introduce
chroma-subsampling artifacts. Peak memory stays close to the decoded image plus the destination.

### Memory and concurrency

Each call allocates a transposed intermediate image and the result, so peak memory is roughly
`source + 2 × destination`. Calls are independent and safe to run concurrently; each call fans out
to `GOMAXPROCS` goroutines. When resizing many images in
parallel, consider limiting your own worker count so the two levels of parallelism do not
oversubscribe the CPU.

## Alternatives

If you do not need this exact API, prefer a maintained library:

* [`golang.org/x/image/draw`](https://pkg.go.dev/golang.org/x/image/draw) — the recommendation of
  the original author; primitive scalers (`ApproxBiLinear`, `BiLinear`, `CatmullRom`), no
  `Thumbnail` helper.
* [`github.com/disintegration/imaging`](https://github.com/disintegration/imaging) — higher-level
  API built on `x/image/draw`, includes filters and EXIF-aware helpers.
* `libvips` bindings (cgo) for high-throughput server workloads.

## Documentation

* [`docs/benchmarks.md`](docs/benchmarks.md) — measured baseline, how to reproduce it, and the
  before/after numbers of the YCbCr memory work.
* [`docs/revival-analysis.md`](docs/revival-analysis.md) — audit of the archived upstream
  repository, including the triage of its 45 issues and 26 pull requests.

## License

ISC — see [LICENSE](LICENSE).

Copyright (c) 2012 Jan Schlicht, (c) 2014 Charlie Vieth, (c) 2026 Stefan Schank (this fork).
