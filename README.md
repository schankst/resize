# resize

[![CI](https://github.com/schankst/resize/actions/workflows/ci.yml/badge.svg)](https://github.com/schankst/resize/actions/workflows/ci.yml)
[![License: ISC](https://img.shields.io/badge/license-ISC-blue.svg)](LICENSE)

Image resizing for the Go programming language with common interpolation methods.

This repository is a **maintained fork** of the original [`nfnt/resize`](https://github.com/nfnt/resize),
which was archived by its author in 2022. The fork keeps the same import path and API so it can be
used as a drop-in replacement.

## Features

* Six interpolation kernels, from fast to high quality
* Works with any `image.Image`; specialized fast paths for `RGBA`, `NRGBA`, `RGBA64`, `NRGBA64`,
  `Gray`, `Gray16` and `YCbCr` (all subsampling ratios)
* Parallelized across `GOMAXPROCS`
* **Zero external dependencies** — only the Go standard library
* Pure Go: cross-compiles to every target without cgo

## Installation

**`go get github.com/nfnt/resize` does not install this fork.** That path resolves to the archived
upstream repository. This fork deliberately keeps the module path `github.com/nfnt/resize` so that
existing code compiles unchanged, which means it is consumed through a `replace` directive.

In a module that already depends on `github.com/nfnt/resize`:

```bash
go mod edit -replace=github.com/nfnt/resize=github.com/schankst/resize@v1.0.2
go mod tidy
```

In a new module:

```bash
go mod init example.com/myapp
go mod edit -require=github.com/nfnt/resize@v1.0.2
go mod edit -replace=github.com/nfnt/resize=github.com/schankst/resize@v1.0.2
go mod tidy
```

The equivalent `go.mod`:

```
require github.com/nfnt/resize v1.0.2

replace github.com/nfnt/resize => github.com/schankst/resize v1.0.2
```

Imports stay `github.com/nfnt/resize`; no source change is needed. Requires Go 1.18 or newer.

Check that the fork is actually in use:

```bash
go list -m all | grep nfnt
# github.com/nfnt/resize v1.0.2 => github.com/schankst/resize v1.0.2
```

Two consequences of the `replace` approach:

* `go get -u` does **not** upgrade a replaced module. Bump the version on the right-hand side
  yourself.
* `go get github.com/schankst/resize@v1.0.2` fails with
  `module declares its path as: github.com/nfnt/resize`. That is expected, not a bug.

## Usage

The package provides two functions:

```go
resize.Resize(width, height uint, img image.Image, interp resize.InterpolationFunction) image.Image
resize.Thumbnail(maxWidth, maxHeight uint, img image.Image, interp resize.InterpolationFunction) image.Image
```

`Resize` scales to the given dimensions. If either `width` or `height` is `0`, the value is
calculated from the aspect ratio of the input. `Thumbnail` downscales preserving the aspect ratio,
never upscales, and returns the input unchanged if it already fits.

```go
package main

import (
	"image/jpeg"
	"log"
	"os"

	"github.com/nfnt/resize"
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
* `Resize` returns `*image.RGBA`/`*image.RGBA64` for those inputs, i.e. the result carries
  premultiplied alpha. Call `draw.Draw` onto an `image.NRGBA` if you need straight alpha again.
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

## Why the module path stays `github.com/nfnt/resize`

Keeping the original module path is what makes this fork a drop-in replacement: existing imports,
`go.mod` entries and call sites keep working, and the fork can be swapped in or out with a single
`replace` line. The trade-offs are the awkward install shown above and the fact that
[pkg.go.dev](https://pkg.go.dev/github.com/nfnt/resize) documents the module path, which resolves
to the archived upstream rather than to this repository.

If you would rather depend on the fork directly, rename the module path in your own copy
(`module github.com/schankst/resize`) and adjust every import — but then it is no longer a drop-in
replacement for code that imports the original path.

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
