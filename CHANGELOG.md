# Changelog

All notable changes to this fork are documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project
adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- `ResizeE` and `ThumbnailE`: same behaviour as `Resize` and `Thumbnail`, but they return an error
  instead of panicking. `Resize`/`Thumbnail` are thin wrappers that panic with that error, so
  existing call sites behave as before while new code can handle the cases. Includes the
  `ErrNilImage` and `ErrInvalidDimensions` sentinels and the `MaxDimension` limit.
- `examples/resize`: a complete program that decodes, resizes and re-encodes an image, including
  flattening transparency for JPEG output. Built by CI as part of `go build ./...`.
- `cmd/resize-measure`: reports peak RSS (`VmHWM`) and time for a resize scenario, so memory changes
  can be compared between checkouts. Documented in `docs/benchmarks.md`; not part of CI.
- `bench_test.go` and `docs/benchmarks.md`: a reproducible baseline across four image types, three
  kernels and both directions, plus how to compare runs with `benchstat`. Replaces the
  unreproducible 2016 numbers from upstream PR #43.
- Table-driven test matrix covering every input type (`RGBA`, `NRGBA`, `RGBA64`, `NRGBA64`, `Gray`,
  `Gray16`, all six `YCbCr` subsampling ratios, `Paletted`, `CMYK` and a generic `image.Image`) with
  every kernel and four directions. Statement coverage went from 63 % to 98 %.
- `FuzzResize`, which already found one real defect (see below). The triggering seed is kept in
  `testdata/fuzz/FuzzResize/`.
- `testdata/ycbcr_golden.txt`: pixel hashes for every YCbCr subsample ratio × kernel × direction,
  used to prove that the memory work did not change a single pixel.

### Changed

- YCbCr inputs are no longer expanded into a full-size interleaved `ycc` buffer (3 bytes per pixel
  of the source) before filtering; rows are expanded on demand. Peak RSS for a 24 MP JPEG resized to
  1200×800 drops from 125.6 MiB to 56.9 MiB, and from 737.7 MiB to 193.0 MiB with eight parallel
  calls. The parallel pass is now split along source rows, so each row is expanded exactly once.
- License headers in all source files reduced to copyright, `SPDX-License-Identifier: ISC` and a
  pointer to `LICENSE`, replacing eight copies of the full ISC text. `LICENSE` now also credits
  Charlie Vieth (`ycc.go`, `nearest.go`) and Stefan Schank (this fork).
- `README.md` rewritten for the fork: unversioned `go get`, which releases declare which module
  path, supported and unsupported formats, fast paths versus the generic fallback, decoder import
  gotchas, metadata and alpha semantics, error handling, memory characteristics and alternatives.
- `docs/revival-analysis.md` is now in English and records what happened to each open upstream pull
  request.
- CI builds only the targets this fork supports: `linux/amd64`, `linux/arm64`, `darwin/arm64`,
  `windows/amd64`, `windows/arm64`, plus 32-bit ARM Linux with `GOARM=6` (Raspberry Pi 1, Zero) and
  `GOARM=7` (Pi 2 v1.2+, Pi 3/4/5 with a 32-bit OS). Benchmark helpers moved out of
  `resize_test.go` into `bench_test.go`.

### Fixed

- `Resize` no longer returns an empty image when the aspect-ratio calculation rounds a dimension
  down to zero (for example a 4×32 input resized with `height` only). Such dimensions are clamped
  to `1`. Found by `FuzzResize`.
- Documented that `NearestNeighbor` preserves `*image.NRGBA`/`*image.NRGBA64`, while the filtered
  kernels return premultiplied `*image.RGBA`/`*image.RGBA64`; the README claimed the latter for all
  kernels.

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

## [1.0.1] - 2026-10-25 — withdrawn

Withdrawn: release and tag were removed because all its changes are contained in v1.0.2. The Go
module proxy still serves the cached copy.

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
