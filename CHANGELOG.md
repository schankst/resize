# Changelog

All notable changes to this fork are documented here.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project
adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.0.0] - TBD

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

[Unreleased]: https://github.com/schankst/resize/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/schankst/resize/releases/tag/v1.0.0
