# nfnt/resize — Revival-Check & Repo-Analyse

**Datum:** 2026-10-25
**Analysiertes Repo:** https://github.com/nfnt/resize (Stand `master`, Commit `83c6a99`, 2018-02-21)
**Dieser Fork:** https://github.com/schankst/resize
**Angewandter Skill:** `engineering/revival-check-golang` — passt, weil das Projekt archiviert/unmaintained ist und die Frage „lohnt Revival des Forks?" beantwortet werden muss. Prinzip des Skills: **verifizieren durch Ausführen, nicht behaupten.**

---

## 1. Status quo (Fakten)

| Merkmal | Wert |
|---|---|
| Archiviert | **ja** (`isArchived: true`, seit ca. April 2022) |
| Letzter Commit `master` | 2018-02-21 (`83c6a99`, Merge PR #64) |
| Letzter Push | 2022-04-02 (nur Branch/Archiv-Aktion, kein Code) |
| Commits gesamt | 147 |
| Stars / Forks | 3.039 / 322 |
| Tags / Releases | **keine** (Issue #53 seit 2016 offen) |
| go.mod | **fehlt** (PR #70 offen, nie gemerged) |
| Lizenz | ISC (LICENSE ohne Titelzeile, PR #48 offen) |
| Issues / PRs | 45 Issues (7 offen), 26 PRs (5 offen, 13 gemerged) |
| Größe | 2.671 Zeilen Go in 6 Dateien + 5 Testdateien (1.961 Nicht-Test-Zeilen) |
| Abhängigkeiten | **keine** — nur `image`, `image/color`, `math`, `runtime`, `sync` |
| Öffentliche API | `Resize()`, `Thumbnail()`, `InterpolationFunction` + 6 Konstanten |

Der Maintainer hat sich in Issue #63 (2018-02-21) explizit verabschiedet und auf `golang.org/x/image/draw` bzw. ImageMagick-Bindings verwiesen. Seitdem keine funktionalen Änderungen. Das Archivieren 2022 hat alle offenen PRs faktisch eingefroren („mergeable_state: blocked").

---

## 2. Ausgeführte Checks (Revival-Skill, Phase A–F)

| # | Check | Kommando | Ergebnis |
|---|---|---|---|
| 1 | Inventar | `find . -name '*.go' \| xargs wc -l` | 2.671 LOC, 5 Testdateien, 0 Vendor |
| 2 | Modul-Setup | `go mod init github.com/nfnt/resize` | **nötig**, Pre-Modules-Ära; `go mod tidy` zieht nichts nach |
| 3 | Build | `go build ./...` (Go 1.26.10) | **OK**, keine Warnungen |
| 4 | Vet | `go vet ./...` | **OK**, 0 Findings |
| 5 | Staticcheck | `staticcheck ./...` (2026er Build) | **0 Findings** — für 8 Jahre alten Code bemerkenswert |
| 6 | Vulns | `govulncheck ./...` | **No vulnerabilities found** (keine Deps, Stdlib-Pfade sauber) |
| 7 | Security | `gosec -fmt=json ./...` | **30 Issues, alle G115** (Integer-Conversions) — ausschließlich `converter.go` (17) und `nearest.go` (13); siehe 5.2 |
| 8 | Cross-Compile | `GOOS=windows\|darwin\|linux CGO_ENABLED=0 go build` | **3/3 OK** (reines Go, kein cgo) |
| 9 | Dep-Health | entfällt | **0 externe Abhängigkeiten** — beste mögliche Antwort |
| 10 | Runtime-Smoke | Probe-Programm (siehe `probe/`) | **funktioniert**; Verhalten unten dokumentiert |
| — | Tests | `go test ./...` / `-cover` | **OK, 63,0 % Coverage**; `ycc.go` fast 100 %, `converter.go`-Pfade teils ungetestet |

**Fazit der Checks:** Das Ding ist kein Sanierungsfall, sondern ein *stabiles, abhängigkeitsfreies, portables* Paket — es ist nur eingefroren und hat keinen Modul-Support.

---

## 3. Issue-Triage (45 Issues, 7 offen)

### Offen

| # | Titel | Bewertung |
|---|---|---|
| #71 | „can you supper tiff" | **Kein Library-Bug.** TIFF dekodiert man selbst (`golang.org/x/image/tiff`), `Resize` arbeitet auf `image.Image`. Dokumentationslücke. |
| #69 | Hoher Speicherverbrauch bei vielen hochauflösenden Bildern | **Echt, by design** (siehe 5.1). Größter inhaltlicher Schwachpunkt. |
| #68 | PNG-Format nicht erkannt | **Kein Bug.** Antwort im Thread ist korrekt: fehlender Blank-Import `_ "image/png"`. Reine Doku-Falle. |
| #65 | „Error output is generated" | Unbrauchbar (nur Screenshot, keine Repro). Zu schließen. |
| #63 | Maintainer-Abschied | **Erledigt** — durch Fork/Archiv obsolet. |
| #56 | Bild auf Mobile-Browser unsichtbar | Vermutlich EXIF-Orientierung/Decoder-Thema, nicht Resampling. Nicht reproduzierbar ohne Asset. |
| #53 | „Please version this repository" | **Berechtigt und leicht zu lösen** — Tags + go.mod. Debian-Maintainer-Anfrage. |

### Geschlossen, aber strategisch relevant

- **#41 „Don't reverse alpha premultiplication, it's fundamentally wrong"** (2016, 6 Kommentare) — Alpha-/Prämultiplikations-Semantik ist der historisch umstrittenste Punkt. Heute durch `Test_ResizeWithPremultipliedAlpha` und `Test_ResizeWithTranslucentColor` abgedeckt, aber die Semantik (Rückgabe von prämultipliziertem RGBA statt NRGBA) bleibt eine API-Falle.
- **#19 / #26 / #31 / #49 / #52 / #58** — Artefakte und Panics: fast alle auf Bounds-/YCbCr-Subsampling-Fehler zurückzuführen, die zwischen 2014 und 2017 gefixt wurden. Diese Fixes sind der Grund, warum der Fork überhaupt sinnvoll ist.
- **#44 / #13 / #10** — Metadaten (EXIF, DPI) gehen verloren. **Kein Bug**, sondern Natur eines Pixel-Resamplers; wer Metadaten braucht, muss sie separat durchreichen.
- **#32 / #12 / #24 / #33 / #8** — Crop-/Subset-Wünsche, alle abgelehnt. Bewusste Scope-Entscheidung (Single Responsibility).

---

## 4. PR-Triage (26 PRs, 5 offen)

### Offene PRs — einzeln bewertet

| PR | Inhalt | Verdict |
|---|---|---|
| **#70** | `go.mod` (3 Zeilen, `go 1.18`) | **Übernehmen.** Ist bei einem Revival Pflicht; trivial, risikofrei. |
| **#67** | „Fix issue when checking the image size": `Bounds().Dx()` → `Bounds().Max.X` | **Ablehnen — der Patch ist falsch.** `Max.X` ist bei Bildern mit `Min != (0,0)` nicht die Breite. Der bestehende Code ist korrekt (empirisch bestätigt: SubImage mit Min(50,60) liefert identisches Ergebnis wie der zugeschnittene Kopie-Vergleich). |
| **#66** | ppc64le-CI in `.travis.yml` | **Ablehnen.** Travis ist tot; Cross-Compile-Check via GitHub Actions ist der richtige Weg. Der eigentliche Code ist architekturunabhängig und braucht nichts. |
| **#48** | „ISC License"-Titelzeile in LICENSE | **Übernehmen**, kosmetisch, aber korrekt und von choosealicense empfohlen. |
| **#43** | Benchmarks in README | **Nicht übernehmen** (Werte von 2016, nicht reproduzierbar). Stattdessen eigene Benchmarks. |

### Gemergte PRs, die die Codebasis prägen

#16 (Refactor, 1.3× Speedup, YCbCr+Nearest), #18 (YCbCr SubImage/PixOffset), #28 (Bounds-Checks + GOMAXPROCS-Limit), #29 (Dst-Pixel-Koordinate), #34 (RGBA-Farbe), #14 (Min != 0), #55 (Zero-Size-Schutz), #7 (Thumbnail), #64 (README).

### Nie gemergte, aber inhaltlich relevante geschlossene PRs

- **#51 / #57 / #39 / #38 / #60** — YCbCr-Subsampling 4:1:1 und 4:1:0. Der Code enthält diese Pfade heute (`ycc.go` mit eigenen Konstanten-Kopien, Zeilen 91–100). `#60` („Use sub sample ratio from the input image") wurde **nicht** übernommen — der Ausgabepfad forciert 4:4:4, was korrekt, aber speicherintensiv ist.
- **#42 / #41** — Alpha-Prämultiplikation: #34 wurde gemerged, #42 nicht. Die Semantik blieb dadurch inkonsistent zwischen RGBA- und NRGBA-Pfad.
- **#2** (Gamma/sRGB), **#4** (LUT für Lanczos), **#8** (Crop) — nicht übernommen; #2 ist der Grund, warum Farben bei extremen Skalierungen leicht daneben liegen können (kein Gamma-korrektes Resampling).

---

## 5. Code-Befunde (verifiziert, nicht behauptet)

### 5.1 Speicherverhalten — Issue #69 ist real

Gemessen (Probe, 2 vCPU):

| Szenario | Peak-RSS | Zeit |
|---|---|---|
| 24-MP-JPEG (6000×4000 YCbCr) dekodieren + →1200×800 Lanczos3 | **~130 MB** | 282 ms |
| 20× 4000×3000 RGBA → 800×600 Lanczos3 | ~119 MB | 2,47 s |
| 8 parallele Goroutinen × 3 Resizes (3000×2000 → 600×400) | ~143 MB | 1,46 s |

Ursache, drei parallel wirksame Effekte:
1. **Vollkopie bei YCbCr:** `ycc.go:127` alloziert `3*w*h` Bytes für das *ganze* Bild, zusätzlich zum dekodierten Original (`imageYCbCrToYCC`, `ycc.go:124`). Bei 24 MP sind das 72 MB Extra.
2. **Transponierter Zwischenpuffer:** `resize.go:114` legt `Dx(input) × width` an — also ein volles Bild in Zielbreite, plus Ergebnis.
3. **Kein Tiling/Streaming:** es gibt keinen Weg, ein Bild in Streifen zu verarbeiten; jede Anfrage braucht den vollen Puffer. Kein Pooling, keine Wiederverwendung der Koeffiziententabellen.

### 5.2 gosec: 30× G115, kein echtes Sicherheitsproblem

Alle Findings sind `uint16 -> uint8` / `int32 -> uint8` in `clampUint8`/`clampUint16` und `nearest.go`. Das ist **absichtliche Saturierung** (Kommentar in `converter.go:25-27` erklärt den Trick). Nicht ausnutzbar: das Paket hat **keine I/O, kein Netzwerk, kein Crypto, keine Dateizugriffe** (Importe: `image`, `image/color`, `math`, `runtime`, `sync`). Das reale Risiko ist **DoS über Speicher/CPU**, nicht Injection.

### 5.3 Weitere Befunde

| Befund | Ort | Schwere |
|---|---|---|
| `int(width)`/`int(height)` ohne Overflow-Prüfung: auf 32-bit kann `width > 2^31` zu negativen Dimensionen → `image.NewRGBA` panict | `resize.go:92,114-115` | mittel (nur 32-bit / absurde Eingaben) |
| Division durch `sum` ohne Null-Check; theoretisch `sum == 0`, wenn alle Koeffizienten auf 0 runden | `converter.go:129` u. a. | niedrig (nicht reproduzierbar getestet bis 100× Downscale) |
| `var blur = 1.0` global, **nicht exportiert** → toter Konfigurationspfad („sharpen" nicht nutzbar) | `resize.go:74` | niedrig, aber API-Missstand |
| Goroutine-Zahl = `GOMAXPROCS(0)` pro Durchlauf, unabhängig von Bildgröße → bei vielen parallelen Aufrufen Oversubscription | `resize.go:106,351` | niedrig (Messung: 1,46 s für 24 Resizes auf 2 Kernen, akzeptabel) |
| `image.Paletted` und `image.CMYK` fallen in den **generischen 16-bit-Pfad** (`in.At()`), d. h. massiv langsamer und ohne Palette-Erhalt | `resize.go` default-Zweig | mittel für GIF/PNG-8-Workloads |
| Kein Kontext/Abbruch: ein 100-MP-Resize läuft bis zum Ende | gesamtes Paket | niedrig |
| Kein Gamma-korrektes Resampling (PR #2 nie übernommen) | `filters.go` | kosmetisch |
| `README` behauptet „needs at least Go 1.1"; `.travis.yml` testet Go 1.1/1.4/1.x (tot) | README, `.travis.yml` | Doku |
| Keine Versionierung → `go get` liefert Pseudo-Version `v0.0.0-…`; Debian-Packaging hängt | fehlende Tags | mittel für Konsumenten |

### 5.4 Was überraschend gut ist

- **`go vet`, `staticcheck`, `govulncheck` komplett sauber** — für 2018er Code ungewöhnlich.
- **`ycc.go` hat ~100 % Testabdeckung** inkl. 4:1:1/4:1:0-Pfaden.
- **Bounds-Behandlung funktioniert** (SubImage mit `Min != (0,0)` verifiziert korrekt, inkl. Nearest-Pfad).
- **Null Abhängigkeiten** → keine Supply-Chain-Fläche, kein CVE-Risiko über Deps.

---

## 6. Empfehlung für den Fork

### Kurzfassung

Der Fork ist **revival-fähig mit sehr geringem Aufwand** (~2–4 Dev-Tage agentengestützt). Der Code ist nicht kaputt — er ist eingefroren. Der eigentliche Wert liegt in: go.mod + Tags, Speicher-Effizienz, moderne API, CI.

### Konkrete Roadmap

| Priorität | Maßnahme | Aufwand |
|---|---|---|
| P0 | `go.mod` übernehmen (#70), Tag `v1.0.0`, `LICENSE`-Titel (#48) | 1 h |
| P0 | GitHub Actions: `go test -race`, `staticcheck`, `govulncheck`, Cross-Compile-Matrix (ersetzt `.travis.yml`, macht #66 überflüssig) | 2 h |
| P1 | Größen-/Overflow-Validierung + dokumentierte Fehler-API (statt panic) — erfordert neue Signatur oder `ResizeE` | 0,5 d |
| P1 | Speicher: YCbCr nur bei Bedarf nach `ycc` kopieren, Zwischenpuffer wiederverwenden (oder gar nicht transponieren), `sync.Pool` für Koeffizienten | 1–1,5 d |
| P2 | Optionale Konfiguration: `blur`/Sharpen exportieren (funktionale Optionen statt globaler `var`) | 0,5 d |
| P2 | `image.Paletted`/`CMYK` schnell pfaden bzw. dokumentiert konvertieren | 0,5 d |
| P3 | Doku: PNG/TIFF/EXIF-Fallen (aus #68/#71/#44) in README; Benchmarks statt 2016er-Zahlen | 0,5 d |

### Alternative, die man ehrlich benennen muss

`golang.org/x/image/draw` (Empfehlung des Originalautors) ist semantisch verwandt, aber ohne Komfort-API (`Thumbnail`, 6 benannte Interpolationen). Wer das Rad nicht neu drehen will: `github.com/disintegration/imaging` (gepflegt, wrappt `x/image/draw`) oder `github.com/h2non/bimg` (libvips, cgo). **Der Fork lohnt sich nur, wenn** (a) API-Kompatibilität zu bestehenden `nfnt/resize`-Aufrufen gewünscht ist oder (b) man die Lern-/Ownership-Übung will. Für reine Neuprojekte ist `imaging` die pragmatischere Wahl.

### Entscheidung zum Fork-Status

Der Fork ist seit 2026-10-25 **public**, getaggt mit `v1.0.0` und über einen `replace`-Block
verifiziert (Testmodul, das `github.com/nfnt/resize` importiert und gegen
`github.com/schankst/resize v1.0.0` ersetzt — Build und Laufzeit OK über den Module Proxy).
Ein direktes `go get github.com/schankst/resize` schlägt bewusst fehl, weil der Modulpfad
`github.com/nfnt/resize` bleibt.

### Umsetzungsstand (Stand 2026-10-25)

| Phase | Inhalt | Status |
|---|---|---|
| P0 | `go.mod`, Tag `v1.0.0`, ISC-Titel, CI (test/race/staticcheck/govulncheck/cross-compile), README, Changelog, Doku | **erledigt** |
| P1 | Panic-/Overflow-Härtung mit additiver Fehler-API, Speicherreduktion, Benchmarks | offen |
| P2 | Konfigurierbares `blur` (funktionale Optionen), schnelle Paletted/CMYK-Pfade | offen |
| P3 | Doku-Feinschliff, Gamma-Frage bewerten | offen |

---

## 7. Checkliste des Skills — Status

- [x] In separatem Scratch-Verzeichnis geklont, Original unangetastet
- [x] LOC + Testdateien erfasst (2.671 / 5)
- [x] `go build ./...` pro Paket (ein Paket, OK)
- [x] `go vet ./...` (0)
- [x] `staticcheck ./...` (0)
- [x] `govulncheck ./...` (keine; keine Deps, kein Ausschluss nötig)
- [x] `gosec` gelaufen (30× G115, erklärt)
- [x] AUTH/Hash/TLS manuell geprüft — **nicht anwendbar**, keine Auth/Crypto/Netzwerk im Paket
- [x] Cross-Compile windows/darwin/linux (3/3)
- [x] Dependency-Tabelle — leer, 0 Deps
- [x] Runtime-Smoke-Test mit echten Bildern, Speicher-/Zeitmessung
- [x] Refactoring-Aufwandsschätzung erstellt
- [x] Keine Behauptung ohne ausgeführtes Kommando

**Reproduzierbarkeit:** Alle Befunde wurden mit Go 1.26.10 auf einer 2-vCPU-Maschine erzeugt. Die Messungen stammen aus einem separaten Wegwerf-Programm (Dekodierung + Resize + Peak-RSS-Sampling über `/proc/<pid>/status`); die Zahlen sind als Größenordnung zu lesen, nicht als Benchmark.
