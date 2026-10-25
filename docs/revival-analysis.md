# nfnt/resize — revival check and repository analysis

**Date:** 2026-10-25
**Analysed repository:** https://github.com/nfnt/resize (`master`, commit `83c6a99`, 2018-02-21)
**This fork:** https://github.com/schankst/resize
**Skill applied:** `engineering/revival-check-golang` — a fit, because the project is archived and
unmaintained and the question is whether reviving it is worth the effort. The skill's principle:
**verify by executing, never assert on paper.**

---

## 1. Status quo (facts)

| Property | Value |
|---|---|
| Archived | **yes** (`isArchived: true`, since about April 2022) |
| Last commit on `master` | 2018-02-21 (`83c6a99`, merge of PR #64) |
| Last push | 2022-04-02 (branch/archive housekeeping only, no code) |
| Total commits | 147 |
| Stars / forks | 3,039 / 322 |
| Tags / releases | **none** (issue #53 open since 2016) |
| go.mod | **missing** (PR #70 open, never merged) |
| License | ISC (`LICENSE` without the title line, PR #48 open) |
| Issues / PRs | 45 issues (7 open), 26 PRs (5 open, 13 merged) |
| Size | 2,671 lines of Go in 6 files + 5 test files (1,961 non-test lines) |
| Dependencies | **none** — only `image`, `image/color`, `math`, `runtime`, `sync` |
| Public API | `Resize()`, `Thumbnail()`, `InterpolationFunction` + 6 constants |

The maintainer explicitly signed off in issue #63 (2018-02-21) and pointed at
`golang.org/x/image/draw` and ImageMagick bindings. No functional changes since then. Archiving in
2022 effectively froze every open PR (`mergeable_state: blocked`).

---

## 2. Checks executed (revival skill, phases A–F)

| # | Check | Command | Result |
|---|---|---|---|
| 1 | Inventory | `find . -name '*.go' \| xargs wc -l` | 2,671 LOC, 5 test files, 0 vendored |
| 2 | Module setup | `go mod init github.com/nfnt/resize` | **required**, pre-modules era; `go mod tidy` pulls nothing in |
| 3 | Build | `go build ./...` (Go 1.26.10) | **OK**, no warnings |
| 4 | Vet | `go vet ./...` | **OK**, 0 findings |
| 5 | Staticcheck | `staticcheck ./...` (2026 build) | **0 findings** — remarkable for 8-year-old code |
| 6 | Vulnerability scan | `govulncheck ./...` | **No vulnerabilities found** (no deps, clean stdlib paths) |
| 7 | Security | `gosec -fmt=json ./...` | **30 findings, all G115** (integer conversions) — only `converter.go` (17) and `nearest.go` (13); see 5.2 |
| 8 | Cross-compile | `GOOS=windows\|darwin\|linux CGO_ENABLED=0 go build` | **3/3 OK** (pure Go, no cgo) |
| 9 | Dependency health | not applicable | **0 external dependencies** — the best possible answer |
| 10 | Runtime smoke test | throwaway probe program | **works**; behaviour documented below |
| — | Tests | `go test ./...` / `-cover` | **OK, 63.0 % coverage**; `ycc.go` near 100 %, several `converter.go` paths untested |

**Conclusion of the checks:** this is not a restoration case but a *stable, dependency-free,
portable* package — it is merely frozen and has no module support.

---

## 3. Issue triage (45 issues, 7 open)

### Open

| # | Title | Assessment |
|---|---|---|
| #71 | "can you supper tiff" | **Not a library bug.** You decode TIFF yourself (`golang.org/x/image/tiff`); `Resize` operates on `image.Image`. Documentation gap. |
| #69 | High memory usage when processing many high-resolution images | **Real, by design** (see 5.1). The most substantial open weakness. |
| #68 | PNG format not recognized | **Not a bug.** The answer in the thread is correct: missing blank import `_ "image/png"`. A pure documentation trap. |
| #65 | "Error output is generated" | Unusable (screenshot only, no repro). Would be closed. |
| #63 | Maintainer's farewell | **Settled** — obsolete through the fork and the archive. |
| #56 | Image not shown in mobile browser after resizing | Probably EXIF orientation or a decoder issue, not resampling. Not reproducible without the asset. |
| #53 | "Please version this repository" | **Valid and easy to fix** — tags plus go.mod. A Debian maintainer's request. |

### Closed but strategically relevant

- **#41 "Don't reverse alpha premultiplication, it's fundamentally wrong"** (2016, 6 comments) —
  alpha/premultiplication semantics is the most contested topic historically. Covered today by
  `Test_ResizeWithPremultipliedAlpha` and `Test_ResizeWithTranslucentColor`, but the semantics
  (returning premultiplied RGBA instead of NRGBA) remain an API trap.
- **#19 / #26 / #31 / #49 / #52 / #58** — artifacts and panics: almost all trace back to bounds and
  YCbCr subsampling bugs fixed between 2014 and 2017. Those fixes are the very reason a fork makes
  sense.
- **#44 / #13 / #10** — metadata (EXIF, DPI) is lost. **Not a bug**, but the nature of a pixel
  resampler; anyone who needs metadata has to carry it through separately.
- **#32 / #12 / #24 / #33 / #8** — crop/subset requests, all rejected. A deliberate scope decision
  (single responsibility).

---

## 4. PR triage (26 PRs, 5 open)

### Open PRs, assessed individually

| PR | Content | Verdict |
|---|---|---|
| **#70** | `go.mod` (3 lines, `go 1.18`) | **Adopt.** Mandatory for a revival; trivial and risk-free. |
| **#67** | "Fix issue when checking the image size": `Bounds().Dx()` → `Bounds().Max.X` | **Reject — the patch is wrong.** `Max.X` is not the width when `Min != (0,0)`. The existing code is correct (empirically confirmed: a SubImage with Min(50,60) produces exactly the same result as the equivalent cropped copy). |
| **#66** | ppc64le CI in `.travis.yml` | **Reject.** Travis is dead; a cross-compile check via GitHub Actions is the right approach. The code itself is architecture-independent and needs nothing. |
| **#48** | "ISC License" title line in LICENSE | **Adopt** — cosmetic, but correct and recommended by choosealicense. |
| **#43** | Benchmarks in the README | **Do not adopt** (2016 numbers, not reproducible). Write our own benchmarks instead. |

### What happened to the five open PRs

Up front: **nothing happened upstream.** The repository is archived — nothing gets merged there,
and the author stated in issue #63 that he no longer looks at PRs or issues. All five are still
`open` to this day (verified via the GitHub API). The decision therefore lives here, and adoption
happens as a commit in this fork.

| PR | Upstream status | Adoption here |
|---|---|---|
| #70 Add go.mod | open, unmerged | Implemented in `v1.0.0`; since `v1.0.3` additionally with its own module path, i.e. beyond what the PR did |
| #67 `Bounds().Max` | open, unmerged | **Rejected** — the patch is factually wrong (see table above); the reasoning belongs in this document, not in the PR |
| #66 ppc64le CI | open, unmerged | Travis part obsolete (`.travis.yml` removed). `linux/ppc64le` is **not** adopted — it is not a target platform of this fork. The code is architecture-independent; anyone who needs ppc64le can build it themselves (verified). |
| #48 ISC title line | open, unmerged | **Adopted** in `v1.0.0` |
| #43 Benchmarks | open, unmerged | Numbers not adopted; superseded by issue #3 (own, reproducible benchmarks) |

### Why #48 ("Add license title")?

The PR adds a first line `ISC License` to `LICENSE` and nothing else. Background: the official ISC
template (choosealicense.com, opensource.org, SPDX) starts with that title line, while the upstream
file only contained the copyright line and the permission text.

Benefits and limits:

- **It changes no license terms.** Pure text, identical rights and obligations.
- **The file becomes self-explanatory** when distributed on its own (a vendor directory, an
  attachment).
- **SPDX/template matching becomes unambiguous**, because the file then matches the canonical
  template.
- **GitHub already detected the license before**: `licenseInfo.key` was `isc` all along, because the
  permission text is what gets matched. So the PR is cosmetic — but risk-free and verifiable in
  seconds (`gh repo view --json licenseInfo`).

In short: no functional effect, but the file then matches the standard text. That is why it was
adopted rather than rejected.

### Merged PRs that shape the codebase

#16 (refactor, 1.3× speedup, YCbCr + nearest), #18 (YCbCr SubImage/PixOffset), #28 (bounds checks +
GOMAXPROCS limit), #29 (destination pixel coordinate), #34 (RGBA colour), #14 (Min != 0), #55
(zero-size guard), #7 (Thumbnail), #64 (README).

### Never merged, but substantively relevant closed PRs

- **#51 / #57 / #39 / #38 / #60** — YCbCr subsampling 4:1:1 and 4:1:0. The code contains those paths
  today (`ycc.go`, with its own copies of the ratio constants). **#60** ("Use sub sample ratio from
  the input image") was **not** adopted — the output path forces 4:4:4, which is correct but
  memory-hungry.
- **#42 / #41** — alpha premultiplication: #34 was merged, #42 was not. The semantics therefore
  remained inconsistent between the RGBA and NRGBA paths.
- **#2** (gamma/sRGB), **#4** (lookup table for Lanczos), **#8** (crop) — not adopted; #2 is the
  reason colours can drift slightly on extreme scaling (no gamma-correct resampling).

---

## 5. Code findings (verified, not asserted)

### 5.1 Memory behaviour — issue #69 is real

Measured (probe, 2 vCPU):

| Scenario | Peak RSS | Time |
|---|---|---|
| Decode a 24 MP JPEG (6000×4000 YCbCr) + resize → 1200×800 Lanczos3 | **~130 MB** | 282 ms |
| 20× 4000×3000 RGBA → 800×600 Lanczos3 | ~119 MB | 2.47 s |
| 8 parallel goroutines × 3 resizes (3000×2000 → 600×400) | ~143 MB | 1.46 s |

Three effects cause this, all active at once:

1. **Full copy for YCbCr:** `ycc.go:127` allocates `3*w*h` bytes for the *entire* image, on top of
   the decoded original (`imageYCbCrToYCC`, `ycc.go:124`). At 24 MP that is 72 MB extra.
2. **Transposed intermediate buffer:** `resize.go:114` allocates `Dx(input) × width`, i.e. a full
   image at the target width, plus the result.
3. **No tiling/streaming:** there is no way to process an image in strips; every call needs the
   full buffer. No pooling, no reuse of the coefficient tables.

**Addressed:** points 1 and 3 for YCbCr were fixed after this analysis. Rows are expanded on demand
while filtering instead of building a `3 × w × h` copy, which halves peak RSS for JPEG inputs
(see `docs/benchmarks.md`, "Before / after"). The transposed intermediate (point 2) is still there
and is the remaining target.

### 5.2 gosec: 30× G115, no real security problem

Every finding is a `uint16 -> uint8` or `int32 -> uint8` conversion in `clampUint8`/`clampUint16`
and in `nearest.go`. That is **deliberate saturation** (the comment in `converter.go:25-27` explains
the trick). Not exploitable: the package has **no I/O, no network, no crypto, no file access**
(imports: `image`, `image/color`, `math`, `runtime`, `sync`). The real risk is **DoS through memory
and CPU**, not injection.

### 5.3 Further findings

| Finding | Location | Severity |
|---|---|---|
| `int(width)`/`int(height)` without an overflow check: on 32-bit, `width > 2^31` can yield negative dimensions → `image.NewRGBA` panics | `resize.go:92,114-115` | medium (32-bit only / absurd input) |
| Division by `sum` without a zero check; theoretically `sum == 0` if all coefficients round to 0 | `converter.go:129` and others | low (not reproducible up to 100× downscale) |
| `var blur = 1.0` global and **not exported** → dead configuration path ("sharpen" unusable) | `resize.go:74` | low, but an API gap |
| Goroutine count = `GOMAXPROCS(0)` per pass, independent of image size → oversubscription with many concurrent calls | `resize.go:106,351` | low (measured: 1.46 s for 24 resizes on 2 cores, acceptable) |
| `image.Paletted` and `image.CMYK` fall into the **generic 16-bit path** (`in.At()`), i.e. far slower and without palette preservation | `resize.go` default branch | medium for GIF/PNG-8 workloads |
| No context/cancellation: a 100 MP resize runs to completion | whole package | low |
| No gamma-correct resampling (PR #2 never adopted) | `filters.go` | cosmetic |
| README claims "needs at least Go 1.1"; `.travis.yml` tests Go 1.1/1.4/1.x (dead) | README, `.travis.yml` | documentation |
| No versioning → `go get` returns a pseudo-version `v0.0.0-…`; Debian packaging is blocked | missing tags | medium for consumers |

### 5.4 What is surprisingly good

- **`go vet`, `staticcheck` and `govulncheck` are completely clean** — unusual for 2018 code.
- **`ycc.go` has ~100 % test coverage**, including the 4:1:1/4:1:0 paths.
- **Bounds handling works** (SubImage with `Min != (0,0)` verified correct, including the nearest
  path).
- **Zero dependencies** → no supply-chain surface and no CVE exposure through dependencies.

---

## 6. Recommendation for the fork

### Summary

The fork is **revivable at very low cost** (~2–4 dev-days, agent-assisted). The code is not broken —
it is frozen. The real value lies in: go.mod plus tags, memory efficiency, a modern API, and CI.

### Concrete roadmap

| Priority | Measure | Effort |
|---|---|---|
| P0 | Adopt `go.mod` (#70), tag `v1.0.0`, LICENSE title (#48) | 1 h |
| P0 | GitHub Actions: `go test -race`, `staticcheck`, `govulncheck`, cross-compile matrix (replaces `.travis.yml`, makes #66 moot) | 2 h |
| P1 | Size/overflow validation plus a documented error API (instead of panic) — needs a new signature or `ResizeE` | 0.5 d |
| P1 | Memory: copy YCbCr to `ycc` only when needed, reuse the intermediate buffer (or stop transposing), `sync.Pool` for coefficients | 1–1.5 d |
| P2 | Optional configuration: export `blur`/sharpen (functional options instead of a package-level `var`) | 0.5 d |
| P2 | Fast paths for `image.Paletted`/`CMYK`, or documented conversion | 0.5 d |
| P3 | Documentation: PNG/TIFF/EXIF traps (from #68/#71/#44) in the README; benchmarks instead of the 2016 numbers | 0.5 d |

### The alternative we have to name honestly

`golang.org/x/image/draw` (the original author's recommendation) is semantically related but has no
convenience API (`Thumbnail`, six named interpolations). Anyone who does not want to rebuild the
wheel can use `github.com/disintegration/imaging` (maintained, wraps `x/image/draw`) or
`github.com/h2non/bimg` (libvips, cgo). **The fork is worth it only if** (a) API compatibility with
existing `nfnt/resize` call sites matters, or (b) you want the learning/ownership exercise. For new
projects, `imaging` is the more pragmatic choice.

### Decision on the fork's status

The fork has been **public** and tagged since 2026-10-25. From `v1.0.3` on, the module path is
`github.com/schankst/resize`, so installation is a single command:
`go get github.com/schankst/resize`. The API is unchanged; only import paths differ.

Up to `v1.0.2` the module kept the path `github.com/nfnt/resize` so a `replace` block could act as a
drop-in. That was technically correct but impractical: `go get github.com/nfnt/resize` fetches the
archived original, and a direct `go get` of the fork is impossible under that path. `v1.0.2` remains
available as the last release with the original path, in case someone cannot touch imports.

A side note on SemVer: `v2.0.0` would be the wrong move, because Go would then require the module
path `github.com/schankst/resize/v2` and corresponding `/v2` imports — the opposite of simple.

### Implementation status (as of 2026-10-25)

| Phase | Content | Status |
|---|---|---|
| P0 | `go.mod` with its own module path, tags up to `v1.0.3`, ISC title, CI (test/race/staticcheck/govulncheck/cross-compile), README, changelog, docs | **done** |
| P1 | Panic/overflow hardening with an additive error API, memory reduction, benchmarks | done except the transposed intermediate buffer |
| P2 | Configurable `blur` (functional options), fast Paletted/CMYK paths | open |
| P3 | Documentation polish, evaluate the gamma question | open |

---

## 7. Skill checklist — status

- [x] Cloned into a separate scratch directory, original untouched
- [x] LOC and test files recorded (2,671 / 5)
- [x] `go build ./...` per package (one package, OK)
- [x] `go vet ./...` (0)
- [x] `staticcheck ./...` (0)
- [x] `govulncheck ./...` (none; no deps, no exclusions needed)
- [x] `gosec` run (30× G115, explained)
- [x] AUTH/hash/TLS checked manually — **not applicable**, no auth/crypto/network in the package
- [x] Cross-compile windows/darwin/linux (3/3)
- [x] Dependency table — empty, 0 deps
- [x] Runtime smoke test with real images, memory/time measurement
- [x] Refactoring effort estimate produced
- [x] No claim without a command that was actually run

**Reproducibility:** all findings were produced with Go 1.26.10 on a 2-vCPU machine. The measurements
come from a separate throwaway program (decode + resize + peak-RSS sampling via `/proc/<pid>/status`);
the numbers are orders of magnitude, not a benchmark.
