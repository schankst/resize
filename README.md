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
* Works with any `image.Image`; specialized fast paths for `RGBA`, `NRGBA`, `RGBA64`, `NRGBA64`,
  `Gray`, `Gray16` and `YCbCr` (all subsampling ratios)
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

The package provides two functions:

```go
resize.Resize(width, height uint, img image.Image, interp resize.InterpolationFunction) image.Image
resize.Thumbnail(maxWidth, maxHeight uint, img image.Image, interp resize.InterpolationFunction) image.Image
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

### Decoding formats

`resize` operates on already decoded images; it does not contain image decoders. Import the
formats you need for their side effects, otherwise `image.Decode` fails with
`image: unknown format`:

```go
import (
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/tiff" // optional, TIFF lives outside the standard library
	_ "golang.org/x/image/bmp"  // optional
)
```

See [`nfnt/resize#68`](https://github.com/nfnt/resize/issues/68) and
[`#71`](https://github.com/nfnt/resize/issues/71) — those reports were missing imports, not
library bugs.

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

JPEGs decode to `*image.YCbCr`. For these inputs the package converts the whole image into an
interleaved `ycc` buffer (3 bytes per pixel) and always emits 4:4:4 output. That is fast and
avoids chroma-subsampling artifacts, but it costs extra memory — roughly `3 × width × height`
bytes on top of the decoded image.

### Memory and concurrency

Each call allocates a transposed intermediate image and the result, so peak memory is roughly
`source + 2 × destination` (more for `YCbCr`). Calls are independent and safe to run
concurrently; each call fans out to `GOMAXPROCS` goroutines. When resizing many images in
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

## License

ISC — see [LICENSE](LICENSE). Original copyright (c) 2012 Jan Schlicht.
