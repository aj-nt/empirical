# Koiné → Empirical Unification Plan

> **For Hermes:** Use subagent-driven-development skill to implement this plan task-by-task.

**Goal:** Make the `koine` repo import `empirical` for all shared computation, deleting ~25 duplicated files from koine. Koine becomes a thin synthesis/interpretation layer on top of empirical's compute engine.

**Architecture:** Empirical is the canonical compute engine (`BaseChart`, `ComputeBaseChart`, all `*FromBase` functions). Koine imports empirical and keeps only its unique synthesis/interpretation code (`synthesis.go`, `planet_in_sign.go`, `interpretation.go`, `evolutionary/`). The koine server and CLI are rewritten to use `BaseChart` instead of the local `chartData` struct.

**Tech Stack:** Go 1.21+, both repos use `github.com/aj-nt/koine` and `github.com/aj-nt/empirical` module paths.

---

## Current State

| Category | Count | Files |
|----------|-------|-------|
| **IDENTICAL** (delete from koine) | 8 | `draconic.go`, `traditional.go`, `synastry.go`, `astrocartography.go`, `directions.go`, `batch.go`, `lunar_mansion.go`, `planet_in_house.go` |
| **TRIVIAL DIFF** (empirical canonical) | 4 | `zodiac.go`, `house.go`, `aspect.go`, `timing.go` |
| **EMPIRICAL AHEAD** (empirical canonical) | 3 | `dignity.go`, `stars.go`, `node.go` |
| **KOINE AHEAD** (port to empirical first) | 3 | `pattern.go` (star support), `transit.go` (ScanTransitPatterns), `arabic_parts.go` |
| **BOTH DIVERGED** (reconcile) | 4 | `composite.go`, `progressed.go`, `api.go`, `interpretation.go` |
| **ARCHITECTURAL** (different designs) | 2 | `chart.go` (SVG renderer vs compute), `server/server.go` (24 params vs ServerConfig) |
| **KOINE-ONLY** (keep in koine) | 5 | `synthesis.go`, `planet_in_sign.go`, `interpretation_stars_test.go`, `evolutionary/`, `cmd/bazi/` |
| **EMPIRICAL-ONLY** (already canonical) | 7 | `base_chart.go`, `koine_from_base.go`, `vedic_from_base.go`, `bazi_from_base.go`, `draconic_from_base.go`, `planets.go`, `mundane/` |

---

## Phase 1: Port koine improvements into empirical

### Task 1: Port `DetectPatternsWithStars` to empirical

**Objective:** Add fixed-star-aware pattern detection from koine's `pattern.go` to empirical.

**Files:**
- Modify: `empirical/internal/dignity/pattern.go`
- Modify: `empirical/internal/dignity/pattern_test.go` (if exists)

**Step 1: Add `DetectPatternsWithStars` function**

Copy the function from koine's `pattern.go` (lines 65-112) into empirical's `pattern.go`. The function takes `(planets, stars map[string]float64, orbDeg float64)` and returns `*PatternReport`. It uses `buildEdges` with `includeStars=true`.

**Step 2: Add `buildEdges` with star support**

The current empirical `buildEdges` doesn't support stars. Add the `includeStars` parameter and star-aspect logic from koine. Stars aspect planets via conjunction, opposition, square, trine only. Stars don't aspect each other.

**Step 3: Add `DefaultPatternOrb` constant**

Koine defines `DefaultPatternOrb = 3.0`. Add this to empirical's `pattern.go`.

**Step 4: Run tests**

```bash
cd /Users/aj/Documents/repos/empirical && go test ./internal/dignity/ -run Pattern -v -count=1
```

### Task 2: Port `ScanTransitPatterns` to empirical

**Objective:** Add transit pattern scanning from koine's `transit.go` to empirical.

**Files:**
- Modify: `empirical/internal/dignity/transit.go`

**Step 1: Add `TransitPatternHit` struct**

```go
type TransitPatternHit struct {
    Date    string  `json:"date"`
    Pattern Pattern `json:"pattern"`
}
```

**Step 2: Add `realTransitPlanets` helper**

Returns the 10 real planets (Sun-Pluto) as `[]planetSpec`.

**Step 3: Add `ScanTransitPatterns` function**

```go
func ScanTransitPatterns(
    startDate, endDate string,
    orbDeg float64,
    compute ComputeFunc,
) ([]TransitPatternHit, error)
```

**Step 4: Run tests**

```bash
cd /Users/aj/Documents/repos/empirical && go test ./internal/dignity/ -run Transit -v -count=1
```

### Task 3: Reconcile `arabic_parts.go`

**Objective:** Check what koine has that empirical doesn't, port if needed.

**Files:**
- Check: diff between koine and empirical `arabic_parts.go`

**Step 1: Diff the files**

```bash
diff /Users/aj/Documents/repos/koine/internal/dignity/arabic_parts.go /Users/aj/Documents/repos/empirical/internal/dignity/arabic_parts.go
```

**Step 2: Port any missing functionality to empirical**

**Step 3: Run tests**

```bash
cd /Users/aj/Documents/repos/empirical && go test ./internal/dignity/ -run Arabic -v -count=1
```

### Task 4: Reconcile `composite.go`, `progressed.go`, `interpretation.go`

**Objective:** Diff each pair, port koine improvements to empirical.

**Step 1: Diff each file**

```bash
for f in composite.go progressed.go interpretation.go; do
    echo "=== $f ==="
    diff koine/internal/dignity/$f empirical/internal/dignity/$f
done
```

**Step 2: Port any koine-only improvements to empirical**

**Step 3: Run tests**

```bash
cd /Users/aj/Documents/repos/empirical && go test ./internal/dignity/ -count=1
```

### Task 5: Reconcile `api.go`

**Objective:** Koine's `api.go` has `FullReport` + `ComputeFullReport`. Empirical's `api.go` has different content. Check if `FullReport` is still used in koine and whether it should move to empirical.

**Step 1: Check koine usage of FullReport**

```bash
cd /Users/aj/Documents/repos/koine && grep -r "FullReport" --include="*.go"
```

**Step 2: If still used, port to empirical. If not, mark for deletion.**

### Task 6: Handle `chart.go` divergence

**Objective:** Koine's `chart.go` is an SVG renderer (366 lines). Empirical's `chart.go` is different (539 lines). These serve different purposes — keep both but rename to avoid collision.

**Step 1: Check what empirical's chart.go contains**

```bash
head -50 /Users/aj/Documents/repos/empirical/internal/dignity/chart.go
```

**Step 2: Rename koine's SVG renderer to `chart_svg.go` in empirical (or keep in koine as koine-only)**

Decision: The SVG renderer is koine-specific (it's a presentation layer). Keep it in koine as `chart_svg.go`. Empirical's `chart.go` stays as-is.

---

## Phase 2: Wire koine to import empirical

### Task 7: Add empirical dependency to koine

**Objective:** Add `github.com/aj-nt/empirical` to koine's `go.mod`.

**Files:**
- Modify: `koine/go.mod`

**Step 1: Add require directive**

```bash
cd /Users/aj/Documents/repos/koine && go get github.com/aj-nt/empirical@latest
```

Or use a `replace` directive for local development:

```
replace github.com/aj-nt/empirical => ../empirical
```

**Step 2: Verify it resolves**

```bash
cd /Users/aj/Documents/repos/koine && go mod tidy
```

### Task 8: Replace `chartData` with `BaseChart` in koine's main.go

**Objective:** Rewrite `cmd/recover/main.go` to use `empirical.ComputeBaseChart` instead of the local `computePositions` + `chartData`.

**Files:**
- Modify: `koine/cmd/recover/main.go`

**Step 1: Remove `chartData` struct and `computePositions` function**

Delete lines 371-437.

**Step 2: Add import for empirical**

```go
import (
    empirical "github.com/aj-nt/empirical/internal/dignity"
)
```

**Step 3: Rewrite `computeAll` to use `ComputeBaseChart`**

```go
func computeAll(name string, year, month, day, hour, minute, second int, tzOff, lat, lng float64, cacheDir string) *dignity.FullReport {
    bc, err := empirical.ComputeBaseChart(name, year, month, day, hour, minute, second, tzOff, lat, lng)
    if err != nil {
        // handle error
    }
    // ... use bc.Tropical, bc.ASC, bc.Ayanamsa, etc.
}
```

**Step 4: Update all call sites that used `cd.planets`, `cd.ayan`, `cd.asc`, `cd.nn`, `cd.jd`**

Map old → new:
- `cd.planets` → `bc.Tropical` (map[string]float64)
- `cd.ayan` → `bc.Ayanamsa`
- `cd.asc` → `bc.ASC`
- `cd.nn` → `bc.Tropical["Node"]` (or add NN to BaseChart if not present)
- `cd.jd` → `bc.JD`

**Step 5: Update `computeTransits` to use BaseChart**

**Step 6: Update all other compute functions**

### Task 9: Delete identical files from koine

**Objective:** Remove the 8 identical files from koine. They'll be imported from empirical.

**Files to delete:**
- `koine/internal/dignity/draconic.go`
- `koine/internal/dignity/traditional.go`
- `koine/internal/dignity/synastry.go`
- `koine/internal/dignity/astrocartography.go`
- `koine/internal/dignity/directions.go`
- `koine/internal/dignity/batch.go`
- `koine/internal/dignity/lunar_mansion.go`
- `koine/internal/dignity/planet_in_house.go`

**Step 1: Delete each file**

```bash
cd /Users/aj/Documents/repos/koine && rm internal/dignity/draconic.go internal/dignity/traditional.go internal/dignity/synastry.go internal/dignity/astrocartography.go internal/dignity/directions.go internal/dignity/batch.go internal/dignity/lunar_mansion.go internal/dignity/planet_in_house.go
```

**Step 2: Update imports in koine files that reference these**

Search for any koine-internal imports of these packages and update to point to empirical.

**Step 3: Run koine tests**

```bash
cd /Users/aj/Documents/repos/koine && go test ./... -count=1
```

### Task 10: Delete trivial-diff files from koine

**Objective:** Remove files where empirical is the canonical version (trivial comment differences only).

**Files to delete:**
- `koine/internal/dignity/zodiac.go`
- `koine/internal/dignity/house.go`
- `koine/internal/dignity/aspect.go`
- `koine/internal/dignity/timing.go`

**Step 1: Delete and update imports**

**Step 2: Run tests**

### Task 11: Delete empirical-ahead files from koine

**Objective:** Remove files where empirical has more advanced versions.

**Files to delete:**
- `koine/internal/dignity/dignity.go` (empirical has per-state breakdown)
- `koine/internal/dignity/stars.go` (empirical has more)
- `koine/internal/dignity/node.go` (empirical has more)

**Step 1: Delete and update imports**

**Step 2: Run tests**

### Task 12: Delete ported files from koine

**Objective:** Remove files whose improvements were ported to empirical in Phase 1.

**Files to delete:**
- `koine/internal/dignity/pattern.go`
- `koine/internal/dignity/transit.go`
- `koine/internal/dignity/arabic_parts.go`
- `koine/internal/dignity/composite.go`
- `koine/internal/dignity/progressed.go`
- `koine/internal/dignity/api.go`

**Step 1: Delete and update imports**

**Step 2: Run tests**

### Task 13: Update koine server to use empirical's ServerConfig

**Objective:** Rewrite `koine/internal/server/server.go` to use empirical's `ServerConfig` pattern (or just import empirical's server package directly).

**Files:**
- Modify: `koine/internal/server/server.go`
- Modify: `koine/cmd/recover/main.go` (server wiring)

**Step 1: Check if koine's server can just import empirical's server**

Koine's server has 24 individual function params. Empirical's uses `ServerConfig` struct with `handleJSON` generic handler. The cleanest approach: delete koine's `server.go` and import empirical's.

**Step 2: Update koine's main.go to use empirical's `server.ServerConfig`**

**Step 3: Run server tests**

```bash
cd /Users/aj/Documents/repos/koine && go test ./internal/server/ -v -count=1
```

---

## Phase 3: Clean up and verify

### Task 14: Delete koine's swe package

**Objective:** Koine's `internal/swe/swe.go` is identical to empirical's. Delete it and import from empirical.

**Files to delete:**
- `koine/internal/swe/swe.go`
- `koine/internal/swe/swe_test.go`

**Step 1: Delete and update all imports**

```bash
cd /Users/aj/Documents/repos/koine && grep -r "koine/internal/swe" --include="*.go"
```

Update each to `github.com/aj-nt/empirical/internal/swe`.

**Step 2: Run tests**

### Task 15: Delete koine's shared packages

**Objective:** Delete packages that are identical to empirical's.

**Packages to delete:**
- `koine/internal/declination/`
- `koine/internal/divisional/`
- `koine/internal/firdaria/`
- `koine/internal/harmonic/`
- `koine/internal/parans/`
- `koine/internal/uranian/`

**Step 1: Delete each directory**

**Step 2: Update imports in koine**

**Step 3: Run tests**

### Task 16: Full test suite

**Objective:** Run all tests in both repos.

```bash
cd /Users/aj/Documents/repos/empirical && go test ./... -count=1
cd /Users/aj/Documents/repos/koine && go test ./... -count=1
```

### Task 17: Build and smoke test

**Objective:** Build both binaries and run a smoke test.

```bash
cd /Users/aj/Documents/repos/empirical && go build ./...
cd /Users/aj/Documents/repos/koine && go build ./...
```

Smoke test:
```bash
cd /Users/aj/Documents/repos/koine && go run ./cmd/recover/ AJ 1969 2 15 23 10 -8 47.038 -122.901
```

---

## Files remaining in koine after unification

| File | Reason |
|------|--------|
| `internal/dignity/synthesis.go` | Koine-specific synthesis engine |
| `internal/dignity/synthesis_test.go` | Tests for synthesis |
| `internal/dignity/planet_in_sign.go` | Koine-specific planet text |
| `internal/dignity/interpretation.go` | Koine-specific interpretation (may diverge from empirical) |
| `internal/dignity/interpretation_test.go` | Tests |
| `internal/dignity/interpretation_stars_test.go` | Tests |
| `internal/dignity/chart.go` | SVG renderer (presentation layer) |
| `internal/evolutionary/` | Koine-specific evolutionary astrology |
| `cmd/recover/main.go` | CLI (rewritten to use BaseChart) |
| `cmd/recover/main_test.go` | CLI tests |
| `cmd/bazi/main.go` | BaZi CLI |
| Other `cmd/` tools | Various diagnostic tools |

---

## Risk assessment

1. **Import cycle**: koine imports empirical, empirical must not import koine. Currently empirical doesn't import koine — safe.
2. **Go version mismatch**: koine uses go 1.21, empirical may use newer. Check and align.
3. **Test coverage gaps**: Some deleted files have tests only in koine. Ensure empirical has equivalent coverage.
4. **`chart.go` collision**: Both repos have different `chart.go`. Koine's is SVG rendering, empirical's is compute. Rename koine's to `chart_svg.go`.
5. **`interpretation.go` divergence**: These have genuinely different content. Keep koine's version in koine, empirical's in empirical.
