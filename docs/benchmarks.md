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

Measured after the YCbCr copy removal; see the memory section below for the before/after comparison.

| Benchmark | ms/op | B/op | allocs/op |
|---|---:|---:|---:|
| `RGBA/down-128x128/NearestNeighbor` | 0.97 | 332,504 | 20 |
| `RGBA/down-128x128/Bilinear` | 1.30 | 334,537 | 20 |
| `RGBA/down-128x128/Lanczos3` | 3.17 | 342,730 | 20 |
| `RGBA/up-1024x1024/NearestNeighbor` | 7.72 | 6,312,651 | 20 |
| `RGBA/up-1024x1024/Bilinear` | 10.30 | 6,316,748 | 20 |
| `RGBA/up-1024x1024/Lanczos3` | 19.28 | 6,333,132 | 20 |
| `NRGBA/down-128x128/NearestNeighbor` | 0.94 | 332,488 | 20 |
| `NRGBA/down-128x128/Bilinear` | 1.58 | 334,536 | 20 |
| `NRGBA/down-128x128/Lanczos3` | 3.75 | 342,728 | 20 |
| `NRGBA/up-1024x1024/NearestNeighbor` | 7.62 | 6,312,651 | 20 |
| `NRGBA/up-1024x1024/Bilinear` | 10.58 | 6,316,748 | 20 |
| `NRGBA/up-1024x1024/Lanczos3` | 21.35 | 6,333,131 | 20 |
| `Gray/down-128x128/NearestNeighbor` | 0.75 | 86,728 | 20 |
| `Gray/down-128x128/Bilinear` | 0.92 | 88,776 | 20 |
| `Gray/down-128x128/Lanczos3` | 2.16 | 96,968 | 20 |
| `Gray/up-1024x1024/NearestNeighbor` | 4.72 | 1,594,058 | 20 |
| `Gray/up-1024x1024/Bilinear` | 5.32 | 1,598,154 | 20 |
| `Gray/up-1024x1024/Lanczos3` | 11.16 | 1,614,537 | 20 |
| `YCbCr420/down-128x128/NearestNeighbor` | 1.13 | 306,088 | 26 |
| `YCbCr420/down-128x128/Bilinear` | 1.42 | 308,136 | 26 |
| `YCbCr420/down-128x128/Lanczos3` | 2.96 | 316,328 | 26 |
| `YCbCr420/up-1024x1024/NearestNeighbor` | 8.92 | 7,891,885 | 26 |
| `YCbCr420/up-1024x1024/Bilinear` | 10.92 | 7,895,981 | 26 |
| `YCbCr420/up-1024x1024/Lanczos3` | 18.66 | 7,912,365 | 26 |
| `Thumbnail/RGBA` | 3.23 | 342,728 | 20 |
| `Thumbnail/NRGBA` | 3.68 | 342,728 | 20 |
| `Thumbnail/Gray` | 2.17 | 96,968 | 20 |
| `Thumbnail/YCbCr420` | 3.58 | 316,328 | 26 |

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
- **`YCbCr420` no longer pays for a full copy.** Decoded JPEGs arrive as subsampled YCbCr. Rows are
  expanded on demand, so a 512×512 source needs ~0.31 MB per operation instead of the ~1.09 MB the
  old full-image `ycc` buffer required. Downscaling is now in the same range as RGBA.
- **Upscaling costs `source + 2 × destination`.** At 1024×1024 the transposed intermediate and the
  result dominate: ~7.9 MB for `YCbCr420`, ~6.3 MB for RGBA, ~1.6 MB for `Gray`. Removing that
  intermediate is the next memory target.
- **Allocations are constant** — 20 per call for the 8-bit formats, 26 for YCbCr (two scratch rows
  for the on-demand chroma expansion) — because each call allocates its buffers once, independent
  of `b.N`. Peak memory, not allocation count, is what matters.
