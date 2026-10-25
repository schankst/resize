# Benchmarks

Baseline for this fork. The numbers are a regression reference, not a competition entry: they were
taken on a small shared VM, so treat absolute values as an order of magnitude and compare runs on
the same machine only.

## How to run

```bash
go test -run '^$' -bench . -benchmem
```

The suite runs in about 40 seconds with the default `-benchtime=1s`. It is **not** part of CI: a
shared runner produces too much noise for a meaningful gate.

To compare two commits, save the output and use
[`benchstat`](https://pkg.go.dev/golang.org/x/perf/cmd/benchstat):

```bash
go test -run '^$' -bench . -benchmem -count=6 > old.txt
git switch my-change
go test -run '^$' -bench . -benchmem -count=6 > new.txt
benchstat old.txt new.txt
```

## Setup

| | |
|---|---|
| Machine | 2 vCPU, Intel Xeon (Skylake), Linux x86-64 |
| Go | 1.26.10 |
| Source image | 512×512, synthetic gradient pattern |
| Downscale | 128×128 (16× fewer pixels) |
| Upscale | 1024×1024 (4× more pixels) |
| Thumbnail | max 128×128, `Lanczos3` |

## Results

| Benchmark | ms/op | B/op | allocs/op |
|---|---:|---:|---:|
| `RGBA/down-128x128/NearestNeighbor` | 0.94 | 332,506 | 20 |
| `RGBA/down-128x128/Bilinear` | 1.33 | 334,537 | 20 |
| `RGBA/down-128x128/Lanczos3` | 3.29 | 342,729 | 20 |
| `RGBA/up-1024x1024/NearestNeighbor` | 7.61 | 6,312,651 | 20 |
| `RGBA/up-1024x1024/Bilinear` | 9.77 | 6,316,750 | 20 |
| `RGBA/up-1024x1024/Lanczos3` | 19.29 | 6,333,132 | 20 |
| `NRGBA/down-128x128/NearestNeighbor` | 0.92 | 332,488 | 20 |
| `NRGBA/down-128x128/Bilinear` | 1.56 | 334,537 | 20 |
| `NRGBA/down-128x128/Lanczos3` | 3.92 | 342,729 | 20 |
| `NRGBA/up-1024x1024/NearestNeighbor` | 7.75 | 6,312,652 | 20 |
| `NRGBA/up-1024x1024/Bilinear` | 10.64 | 6,316,748 | 20 |
| `NRGBA/up-1024x1024/Lanczos3` | 20.82 | 6,333,132 | 20 |
| `Gray/down-128x128/NearestNeighbor` | 0.78 | 86,728 | 20 |
| `Gray/down-128x128/Bilinear` | 0.83 | 88,776 | 20 |
| `Gray/down-128x128/Lanczos3` | 2.33 | 96,968 | 20 |
| `Gray/up-1024x1024/NearestNeighbor` | 4.92 | 1,594,058 | 20 |
| `Gray/up-1024x1024/Bilinear` | 5.42 | 1,598,154 | 20 |
| `Gray/up-1024x1024/Lanczos3` | 11.62 | 1,614,538 | 20 |
| `YCbCr420/down-128x128/NearestNeighbor` | 1.43 | 1,086,458 | 24 |
| `YCbCr420/down-128x128/Bilinear` | 1.84 | 1,088,507 | 24 |
| `YCbCr420/down-128x128/Lanczos3` | 3.48 | 1,096,697 | 24 |
| `YCbCr420/up-1024x1024/NearestNeighbor` | 9.13 | 8,672,253 | 24 |
| `YCbCr420/up-1024x1024/Bilinear` | 10.99 | 8,676,349 | 24 |
| `YCbCr420/up-1024x1024/Lanczos3` | 18.48 | 8,692,732 | 24 |
| `BenchmarkThumbnail/RGBA` | 3.07 | 342,728 | 20 |
| `BenchmarkThumbnail/NRGBA` | 3.87 | 342,728 | 20 |
| `BenchmarkThumbnail/Gray` | 2.36 | 96,968 | 20 |
| `BenchmarkThumbnail/YCbCr420` | 3.26 | 1,096,698 | 24 |

## What the numbers say

- **Kernel cost dominates.** `Lanczos3` is roughly 3–3.5× slower than `NearestNeighbor` and about
  2–2.5× slower than `Bilinear` at every size. That is the expected trade-off for a 6-tap kernel.
- **Downscaling is cheap, upscaling is not.** Halving the output size costs 1–4 ms; producing
  1024×1024 costs 5–21 ms, because the work scales with destination pixels.
- **`Gray` is the cheapest path** (1 byte per pixel, fewer cache lines): 0.8 ms versus 0.9–1.4 ms
  for 4-byte formats on downscale.
- **`YCbCr420` pays for its conversion.** Decoded JPEGs arrive as subsampled YCbCr, and the package
  copies the whole image into an interleaved `ycc` buffer: ~1.09 MB allocated for a 512×512 source
  versus ~0.33 MB for RGBA. This is the memory issue tracked in issue #2.
- **Allocations are constant** (20 per call for 8-bit formats, 24 for YCbCr) because each call
  allocates the intermediate and result buffers once, independent of `b.N`. Peak memory, not
  allocation count, is the thing to improve.
