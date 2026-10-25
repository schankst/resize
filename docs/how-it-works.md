# How it works

This document describes the inside of the package: which function does what, in which order, and
why the implementation looks the way it does. Comments in the source and tests refer to the numbered
sections here as `[ARCH-x.y]`. A test (`docAnchor_test.go`) fails if a referenced anchor does not
exist in this file, so the references cannot rot silently.

If you only want to use the package, the README is the place to start.

---

## [ARCH-1] Overview

`resize` is a separable convolution resampler. "Separable" means the 2-D filter is applied as two
1-D passes: first along x, then along y. For a kernel with `taps` samples per output pixel this turns
`taps²` operations per output pixel into `2 × taps`.

Everything works on already decoded images; the package has no decoders and no encoders, and no
external dependencies.

### [ARCH-1.1] Public API

| Function | Purpose |
|---|---|
| `Resize(w, h, img, kernel)` | scale to exact dimensions, `0` means "derive from aspect ratio" |
| `ResizeE(w, h, img, kernel)` | same, returns an error instead of panicking |
| `Thumbnail(maxW, maxH, img, kernel)` | downscale preserving aspect ratio, never upscale |
| `ThumbnailE(maxW, maxH, img, kernel)` | same, returns an error instead of panicking |
| `InterpolationFunction` | kernel selector (six constants) |
| `MaxDimension`, `ErrNilImage`, `ErrInvalidDimensions` | limits and error values |

### [ARCH-1.2] File map

| File | Contains |
|---|---|
| `resize.go` | entry points, size resolution, dispatch, parallel splitting |
| `thumbnail.go` | aspect-ratio-preserving downscale, delegates to `ResizeE` |
| `filters.go` | the six kernels and the coefficient tables |
| `converter.go` | the filtering loops for the concrete image types |
| `nearest.go` | the same loops specialised for nearest neighbour (no multiplication) |
| `ycc.go` | interleaved Y/Cb/Cr buffer, Chroma expansion, conversion back to `image.YCbCr` |

---

## [ARCH-2] Pipeline

```
 image.Image
      │
      ▼
 ┌──────────────────────────────────────────────────────────────┐
 │ [ARCH-3.1] ResizeE / ThumbnailE                              │
 │   nil image            → ErrNilImage                         │
 │   width/height == 0    → resolveSize (aspect ratio, min 1)   │
 │   too large / overflow → ErrInvalidDimensions                │
 └───────────────────────────┬──────────────────────────────────┘
                             │ same size, or empty image?
                             ├──────────────────► input returned unchanged
                             ▼
 ┌──────────────────────────────────────────────────────────────┐
 │ [ARCH-3.2] dispatch on the concrete type                     │
 └───┬──────────────────────┬────────────────────────┬──────────┘
     ▼                      ▼                        ▼
 NearestNeighbor      RGBA, NRGBA, RGBA64,       YCbCr
 [ARCH-3.8]           NRGBA64, Gray, Gray16      [ARCH-3.7]
 one pass             [ARCH-3.4] → [ARCH-3.6]    row streaming
                      (filtered kernels)         + [ARCH-3.4] → [ARCH-3.6]

 any other image.Image ──► generic 16-bit path inside [ARCH-3.4]
```

---

## [ARCH-3] Stations

### [ARCH-3.1] Entry and validation

`ResizeE` is the real entry point; `Resize` calls it and panics on error. `ThumbnailE` computes the
target size from the aspect ratio and then delegates to `ResizeE`, so there is exactly one code path
that allocates pixels.

* `resolveSize` fills in a `0` width or height and clamps both to at least 1, because the
  aspect-ratio arithmetic can round down to zero for extreme shapes.
* `validateSize` rejects dimensions above `MaxDimension` and combinations whose buffers would
  overflow `int`. It is called before anything is allocated.
* Empty images and same-size requests return the input unchanged, without allocating.

### [ARCH-3.2] Type dispatch

`resizeImage` is one large `switch` over `img.(type)`. The concrete type decides three things: the
pixel format of the intermediate buffer, the precision of the coefficients (8- or 16-bit), and which
loop function is used. The `default` branch handles every other `image.Image` through `At()`.

### [ARCH-3.3] Coefficient tables

`InterpolationFunction.kernel()` maps a kernel constant to its support:

| Kernel | taps |
|---|---|
| `NearestNeighbor` | 2 |
| `Bilinear` | 2 |
| `Bicubic` | 4 |
| `MitchellNetravali` | 4 |
| `Lanczos2` | 4 |
| `Lanczos3` | 6 |

`createWeights8` and `createWeights16` turn a kernel plus a scale factor into three arrays:

```
 createWeights8(dy, taps, blur, scale, kernel)

   filterLength = taps × max(⌈blur·scale⌉, 1)     wider window when downscaling
   start[y]     = int(scale·(y+0.5) − 0.5) − filterLength/2 + 1
   coeff[y][i]  = int16(kernel(…) × 256)         8-bit path
                = int32(kernel(…) × 65536)       16-bit path

   dy = number of output positions along the filtered axis
```

Two details matter for reading the filtering loops:

1. The table is built for one axis at a time, so pass 1 and pass 2 use different tables (`scaleX`
   versus `scaleY`).
2. The coefficients are **not** normalised upfront. The loop accumulates `Σ coeff × sample` and
   divides by `Σ coeff` at the end, which keeps the intermediate values in integer arithmetic.

`createWeightsNearest` produces `[]bool` instead: exactly one tap per output position is selected.

### [ARCH-3.4] The filtering loop

All loops in `converter.go` share one shape: they filter **along a row of the input** and write to a
**column of the output**.

```
 for x := out.Min.X; x < out.Max.X; x++ {     // x = row index in `in`
     row := in.Pix[x*in.Stride:]              // contiguous input row
     for y := out.Min.Y; y < out.Max.Y; y++ { // y = position along the row
         sum, total := 0, 0
         for i := 0; i < filterLength; i++ {
             sample := row[clamp(start[y]+i) * pixelsize]
             sum   += coeff[y][i] * sample
             total += coeff[y][i]
         }
         out.Pix[y*out.Stride + x*pixelsize] = clamp(sum / total)
     }
 }
```

The clamped index (`clamp(start[y]+i)`) is what extends the image at the borders: samples outside
the row are replaced by the nearest edge pixel.

Special cases in this station:

* `resizeNRGBA` / `resizeNRGBA64` premultiply alpha before filtering, so transparent pixels do not
  bleed their colour into the result.
* `resizeGeneric` uses `in.At(x, y)` instead of direct pixel access and works at 16-bit precision.
  It is the fallback for `Paletted`, `CMYK`, `Alpha`, `NYCbCrA`, `Uniform` and custom types.
* `clampUint8` / `clampUint16` saturate the result. They exploit unsigned wraparound instead of
  branching, which is why they contain the only integer conversions that `gosec` reports.

### [ARCH-3.5] Why there is a transposed intermediate

> **Status: scheduled for replacement.** This section describes the current implementation. The
> transposed buffer is the last large allocation in the package; removing it by streaming the
> vertical pass is the next planned memory improvement. When that lands, this section is rewritten
> and `[ARCH-3.6]` changes with it.

Pass 1 processes one source row per iteration and writes one column of the intermediate. Storing the
result column-wise means the horizontal pass has a contiguous input *and* a contiguous output.
Pass 2 then reads that buffer row-wise.

```
 source 4×3                            intermediate 4×3 (transposed)
 ┌────────────────┐                    ┌────────────────┐
 │ s00 s01 s02 s03│    [ARCH-3.4]      │ h00 h10 h20 h30│  row 0    = target x 0
 │ s10 s11 s12 s13│ ───────────────►   │ h01 h11 h21 h31│  column 0 = source row 0
 │ s20 s21 s22 s23│                    │ h02 h12 h22 h32│
 └────────────────┘                    └────────────────┘
                                              │
                                              │ [ARCH-3.6] filter each row
                                              ▼
                                       result 4×3
```

Without the transpose, pass 2 would have to walk columns with a stride of one row, which is a cache
miss factory. The cost is one full extra buffer (`source height × target width`) — the allocation
this section is scheduled to lose.

### [ARCH-3.6] The second pass and the result

Pass 2 calls the same function as pass 1 with the intermediate as input. Because the buffer is
transposed, "filter along the row" now means "filter along y". The output is the final image, or —
for `YCbCr` — a `ycc` buffer that still has to be converted back ([ARCH-3.9]).

There is a `sync.WaitGroup` between the passes: pass 2 needs every column of the intermediate to be
complete.

### [ARCH-3.7] YCbCr row streaming

JPEGs decode to `*image.YCbCr`, with chroma subsampled (4:2:0 and friends). Historically the whole
image was first expanded into an interleaved `ycc` buffer with three bytes per pixel, which was the
largest allocation in the package. Now `ycbcrRowSource` expands **one row at a time** while filtering:

```
 *image.YCbCr                     ycbcrRowSource.yccRow(y)
 ┌───────────┐                   ┌────────────────────────┐
 │ Y  plane  │──┐                │ Y  Y  Y  Y  …  (w)     │      ┌────────────┐
 │ Cb plane  │──┼── one row ───► │ Cb Cb Cb …  (w/2 x)    │─────►│ Σ coeff×…  │
 │ Cr plane  │──┘   at a time    │ Cr Cr Cr …  (w/2 x)    │      └────────────┘
 └───────────┘                   └────────────────────────┘
```

The parallel split uses `makeColSlice`, i.e. one goroutine per range of **source rows**, so each row
is expanded exactly once. `yccRow` serves `*ycc` input by returning a slice of the existing buffer,
so pass 2 pays nothing for this abstraction.

### [ARCH-3.8] Nearest neighbour

`NearestNeighbor` takes a separate route (`resizeNearest` → `nearest.go`). The structure is the same
two-pass pipeline, but the coefficient table is `[]bool` and the inner loop copies one sample instead
of accumulating a weighted sum. It is the only path that preserves `*image.NRGBA` and
`*image.NRGBA64` as output types; the filtered kernels premultiply and therefore return `RGBA`.

### [ARCH-3.9] From `ycc` back to `image.YCbCr`

The filtered `YCbCr` path produces a `*ycc` with 4:4:4 sampling. `(*ycc).YCbCr` allocates the final
`image.YCbCr` and writes the three planes with the ratio-specific helpers (`ycbcr420`, `ycbcr444`, …).
Output is always 4:4:4, so resizing never introduces chroma-subsampling artefacts, at the cost of
33 % more bytes for the result.

---

## [ARCH-4] Concurrency

Each pass fans out to `runtime.GOMAXPROCS(0)` goroutines, and the two passes are separated by a
`WaitGroup`.

```
 intermediate / result
 ┌──────────────┬──────────────┐     makeSlice (default): split output rows.
 │ goroutine 0  │ goroutine 1  │     Every goroutine reads all input rows
 │ rows 0..n/2  │ rows n/2..n  │     and writes only its own output rows.
 └──────────────┴──────────────┘

 YCbCr pass 1
 ┌──────────────┬──────────────┐     makeColSlice: split source rows, so
 │ source rows  │ source rows  │     every source row is expanded exactly
 │   0..m/2     │   m/2..m     │     once instead of once per goroutine.
 └──────────────┴──────────────┘
```

Calls are independent: there is no shared state between two `Resize` calls. The package-level `blur`
variable is read-only in practice, which is why concurrent calls are safe.

The fan-out is not limited by the image size, so many concurrent calls on small images oversubscribe
the CPU. That is a caller-side concern: limit your own worker pool.

---

## [ARCH-5] Memory

```
 peak ≈ source + intermediate + result

 source        what the caller already holds (1.5 B/px for YCbCr, 4 B/px for RGBA)
 intermediate  source height × target width × bytes per pixel   ← transposed buffer
 result        target width × target height × bytes per pixel   (ycc uses 3 B/px,
                                                                  the final image.YCbCr 1.5 B/px)
```

Measured for a 24 MP JPEG (6000×4000) resized to 1200×800 with `Lanczos3`: about 57 MiB peak RSS,
of which 34 MiB is the decoded source. `docs/benchmarks.md` has the numbers and the tool that
produces them.

---

## [ARCH-6] Test map

Every test function starts with a `Covers [ARCH-x.y].` line; this table is generated from those
annotations. If you add a test, annotate it — the map is only as good as the annotations, and
`TestDocAnchorsExist` guarantees that every referenced stage exists.

| Tests | Stage |
|---|---|
| `Test_Param1`, `Test_Param2`, `Test_ZeroImg`, `Test_HalfZeroImg`, `Test_SameSizeReturnsOriginal`, `TestThumbnail`, `TestResizeAspectRatio`, `TestThumbnailMatrix`, `TestResizeENilImage`, `TestResizeEPanicsOnInvalidInput`, `TestResizeEInvalidDimensions`, `TestResizeEMatchesResize`, `TestResizeEEdgeCases`, `TestThumbnailEMatchesThumbnail` | [ARCH-3.1] |
| `Test_Bounds`, `TestResizeMatrix`, `FuzzResize` | [ARCH-3.2] |
| `TestKernels` | [ARCH-3.3] |
| `Test_ClampUint8`, `Test_ClampUint16`, `Test_CorrectResize`, `Test_SameColorWithRGBA`, `Test_SameColorWithNRGBA`, `Test_SameColorWithRGBA64`, `Test_SameColorWithNRGBA64`, `Test_SameColorWithGray`, `Test_SameColorWithGray16`, `Test_PixelCoordinates`, `Test_ResizeWithPremultipliedAlpha`, `Test_ResizeWithTranslucentColor`, `TestResizeSubImageBounds` | [ARCH-3.4] |
| `TestConvertYCbCr`, `TestYCbCrGolden` | [ARCH-3.7] |
| `Test_FloatToUint8`, `Test_FloatToUint16` | [ARCH-3.8] |
| `TestImage`, `TestYCbCr`, `TestYCCAccessors` | [ARCH-3.9] |
| `BenchmarkResize`, `BenchmarkThumbnail` | [ARCH-5] |

`testdata/ycbcr_golden.txt` pins pixel hashes for every YCbCr subsample ratio × kernel × direction.
It exists so that performance work can prove it changed no pixels.

---

## [ARCH-7] Glossary

| Term | Meaning |
|---|---|
| **taps** | number of samples a kernel needs per output position (2 for linear, 6 for Lanczos3) |
| **filterLength** | `taps × ⌈scale⌉`: the actual window after widening for downscaling |
| **scale** | `source size / target size`; > 1 means downscaling |
| **premultiplied alpha** | colour components already multiplied by alpha, which is what filtering requires |
| **ycc** | internal interleaved Y, Cb, Cr buffer, three bytes per pixel |
| **transposed intermediate** | buffer written column-wise by pass 1 so pass 2 can read it row-wise |
| **kernel** | the weight function, e.g. `lanczos3` |
