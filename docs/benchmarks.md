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

## Memory

### How to measure

`cmd/resize-measure` reports peak resident memory (`VmHWM` from `/proc/self/status`, Linux only)
and wall time for one scenario:

```bash
go run ./cmd/resize-measure -src 6000x4000 -dst 1200x800 -type ycbcr420
go run ./cmd/resize-measure -src 6000x4000 -dst 1200x800 -type ycbcr420 -parallel 8
```

Run the same command against two checkouts to compare. It is not part of CI; the numbers below were
taken on the machine described above.

### Before / after removing the full YCbCr copy

The YCbCr path used to expand the whole source into an interleaved `ycc` buffer (3 bytes per pixel)
before filtering. Rows are now expanded on demand while filtering, so only the transposed
intermediate and the result remain.

| Scenario | peak RSS before | peak RSS after | heap after | time before | time after |
|---|---|---|---|---|---|
| 6000×4000 YCbCr 4:2:0 → 1200×800 Lanczos3 | 125.6 MiB | **56.9 MiB** | 122.4 → 53.8 MiB | 304 ms | 237 ms |
| same, 8 parallel calls | 737.7 MiB | **193.0 MiB** | 738.7 → 189.8 MiB | 2.64 s | 1.79 s |
| 6000×4000 RGBA → 1200×800 Lanczos3 | 116.7 MiB | 116.8 MiB | unchanged | 231 ms | 231 ms |

Allocation per operation for the 512×512 benchmark source dropped accordingly:

| Benchmark | B/op before | B/op after |
|---|---:|---:|
| `YCbCr420/down-128x128/NearestNeighbor` | 1,086,481 | 306,134 |
| `YCbCr420/down-128x128/Lanczos3` | 1,096,698 | 316,329 |
| `YCbCr420/up-1024x1024/Lanczos3` | 8,692,733 | 7,912,365 |

Output is unchanged: `testdata/ycbcr_golden.txt` pins pixel hashes for all six subsample ratios ×
six kernels × four directions, and a one-off comparison of 108 further cases (97×61 source, up,
down and extreme aspect ratios) produced identical hashes before and after.

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
