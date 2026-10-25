# Changelog

All notable changes to this fork are documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project
adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Fixed

- `Resize` no longer returns an empty image when the aspect-ratio calculation rounds a dimension
  down to zero (for example a 4×32 input resized with `height` only). Such dimensions are clamped
  to `1`. Found by `FuzzResize`.
- `README.md` documents that `NearestNeighbor` preserves `*image.NRGBA`/`*image.NRGBA64`, while the
  filtered kernels return premultiplied `*image.RGBA`/`*image.RGBA64`.

### Changed

- `README.md` shows the unversioned `go get github.com/schankst/resize` and explains which releases
  declare which module path: `v1.0.0` was moved after publication, `v1.0.1` was withdrawn (the
  module proxy still serves its cached copy), and `v1.0.2` is the last release with the original
  module path.
- `docs/revival-analysis.md` is now in English.
- Benchmark helpers moved out of `resize_test.go` into `bench_test.go`; the old single-size
  `Benchmark_*` functions are superseded by the table-driven sub-benchmarks.

### Added

- Table-driven test matrix covering every input type (`RGBA`, `NRGBA`, `RGBA64`, `NRGBA64`, `Gray`,
  `Gray16`, all six `YCbCr` subsampling ratios, `Paletted`, `CMYK` and a generic `image.Image`) with
  every kernel and four resize directions, pinning the concrete result types, aspect-ratio
  behaviour, non-zero `Bounds().Min`, the interpolation kernels and the internal `ycc` accessors.
  Statement coverage is now 99 % (was 63 %).
- `FuzzResize`: 2.2 million executions over 90 s without a crash; the seed that exposed the
  degenerate aspect-ratio case is kept in `testdata/fuzz/FuzzResize/`.
- `bench_test.go`: `BenchmarkResize` covers `RGBA`, `NRGBA`, `Gray` and `YCbCr 4:2:0` with
  `NearestNeighbor`, `Bilinear` and `Lanczos3` in both directions, plus `BenchmarkThumbnail`.
  `docs/benchmarks.md` records the baseline and explains how to compare runs with `benchstat`.
  Replaces the unreproducible 2016 numbers from upstream PR #43.
- `docs/revival-analysis.md` now records what happened to each of the five open upstream pull
  requests: nothing upstream (the repository is archived and read-only), with the decision and the
  corresponding commit in this fork documented instead.
- CI cross-compiles only the targets this fork supports: `linux/amd64`, `linux/arm64`,
  `darwin/arm64`, `windows/amd64`, `windows/arm64`, plus 32-bit ARM Linux with `GOARM=6` (Raspberry
  Pi 1, Zero) and `GOARM=7` (Pi 2 v1.2+, Pi 3/4/5 with a 32-bit OS). Exotic architectures are no
  longer built; the code remains architecture-independent.

## [1.0.3] - 2026-10-25

### Changed

- **Module path is now `github.com/schankst/resize`** (was `github.com/nfnt/resize`). This makes the
  fork installable with a plain `go get github.com/schankst/resize` — no `replace` directive, no
  path mismatch. The API is unchanged; only the import path differs, so migrating is
  `github.com/nfnt/resize` → `github.com/schankst/resize`.
- `v1.0.2` remains the last release that declares the original module path, for consumers that
  cannot change imports. Note that Go requires the import path to match the module path, so the two
  cannot be offered by the same tag series.
- The pkg.go.dev badge is back, now that the module path resolves to this repository.

## [1.0.2] - 2026-10-25

Documentation only; the resizing code is unchanged.

### Fixed

- `README.md` no longer suggests `go get github.com/nfnt/resize` as the installation step. That
  command fetches the archived upstream, not this fork. The installation section now shows the
  `go mod edit -replace` flow for new and existing modules, how to verify the replacement with
  `go list -m all`, and the two consequences of the `replace` approach (`go get -u` does not
  upgrade replaced modules; a direct `go get github.com/schankst/resize` fails by design).
- Removed the pkg.go.dev badge: it resolves the module path to the archived upstream, so it never
  linked to these docs.

## [1.0.1] - 2026-10-25

Documentation and CI only; the resizing code is unchanged from v1.0.0.

### Fixed

- `README.md` no longer claims that `go get github.com/schankst/resize` works. It does not and
  cannot: the module deliberately declares the path `github.com/nfnt/resize`, so the `replace`
  directive is the only way to depend on this fork. Verified against the module proxy.
- `docs/revival-analysis.md` records the P0 status and the verified drop-in setup.

### Changed

- CI cross-compile matrix targets `windows/arm64` instead of the 32-bit `windows/arm` port, which
  the Go toolchain removed in 1.26. All other targets and the stable toolchain requirement are
  unchanged.

## [1.0.0] - 2026-10-25

First release of the maintained fork. Behaviour is identical to the final upstream state
(`nfnt/resize@83c6a99`, 2018-02-21); the release is about making the package usable again, not
about changing pixels.

### Added

- `go.mod`, declaring module path `github.com/nfnt/resize` so the fork can be used with a
  `replace` directive (supersedes upstream PR #70).
- GitHub Actions CI: tests across Go versions, race detector, `staticcheck`, `govulncheck` and a
  cross-compile matrix (`linux`, `darwin`, `windows`, `freebsd` × `amd64`, `arm64`, `386`, `arm`).
  Replaces the dead Travis configuration (supersedes upstream PR #66).
- `docs/revival-analysis.md`: full audit of the upstream repository — build, vet, static analysis,
  vulnerability scan, memory measurements, and a triage of all 45 issues and 26 pull requests.
- This changelog.

### Changed

- `LICENSE` now carries the "ISC License" title line, as recommended by the ISC template
  (supersedes upstream PR #48).
- `README.md` rewritten: installation, decoding-format gotchas, metadata and alpha semantics,
  memory/concurrency characteristics, drop-in usage via `replace`, and maintained alternatives.
- Repository is versioned with Git tags (upstream issue #53).

### Removed

- `.travis.yml`.

### Notes

- No source file of the resizing implementation was modified in this release.
- Pull requests from upstream were reviewed but not merged; see the triage table in
  `docs/revival-analysis.md`. Notably upstream PR #67 (`Bounds().Max` instead of `Bounds().Dx()`)
  is incorrect and was intentionally not applied.

[Unreleased]: https://github.com/schankst/resize/compare/v1.0.3...HEAD
[1.0.3]: https://github.com/schankst/resize/compare/v1.0.2...v1.0.3
[1.0.2]: https://github.com/schankst/resize/compare/v1.0.1...v1.0.2
[1.0.1]: https://github.com/schankst/resize/compare/v1.0.0...v1.0.1
[1.0.0]: https://github.com/schankst/resize/releases/tag/v1.0.0
