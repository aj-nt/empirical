# CLI Expansion: Full API Parity

**Goal:** Every API endpoint has a corresponding CLI subcommand. The compute functions already exist — they're defined in the `serve` block of `main.go` and wired through `server.go`. The CLI just needs subcommand wrappers.

**Date:** 2026-07-17

## Current State

### CLI has (5 subcommands):
| Subcommand | API equivalent |
|------------|---------------|
| `recover` (default) | `/api/recover` |
| `transit` | `/api/transits` |
| `synastry` | `/api/synastry` |
| `event` | `/api/event-chart` |
| `serve` | (starts server) |

### CLI missing (18 endpoints):
| API endpoint | Compute function exists? |
|-------------|-------------------------|
| `/api/patterns` | Yes — `patterns` closure in serve block |
| `/api/draconic` | Yes — `draconic` closure |
| `/api/draconic-synastry` | Yes — `draconicSynastry` closure |
| `/api/draconic-synastry-full` | Yes — `draconicSynastryFull` closure |
| `/api/draconic-interpretation` | Yes — `draconicInterpretation` closure |
| `/api/synastry-interpretation` | Yes — `synastryInterpretation` closure |
| `/api/transit-interpretation` | Yes — `transitInterpretation` closure |
| `/api/transit-patterns` | Yes — `transitPatterns` closure |
| `/api/stars` | Yes — `stars` closure |
| `/api/directions` | Yes — `directions` closure |
| `/api/interpretation` | Yes — `interpretation` closure |
| `/api/arabic-parts` | Yes — `arabicParts` closure |
| `/api/traditional` | Yes — `traditional` closure |
| `/api/horoscope` | Yes — `horoscope` closure |
| `/api/evolutionary` | Yes — `evolutionaryFn` closure |
| `/api/compare` | Yes — `compare` closure |
| `/api/relocation-compare` | Yes — `relocation` closure |
| `/api/chart` | Yes — `chart` closure |

## Approach

The compute functions are currently defined as closures inside the `serve` block (lines 39-177 of `main.go`). They capture `cacheDir` from the serve block's ephemeris init. The CLI subcommands need the same ephemeris init, then call the same logic.

**Strategy: Extract compute functions from closures into package-level functions.**

The closures in the serve block are thin wrappers that call package-level `compute*` functions (which already exist — `computeTransits`, `computeSynastry`, `computePatterns` via the patterns closure, etc.). The closures just add the `cacheDir` parameter. The package-level functions already handle `cacheDir` being empty (they call `EnsureEpheCache()` themselves).

So the pattern for each new subcommand is:
1. Parse args/flags
2. Call `EnsureEpheCache()` + `swe.SetEphePath()`
3. Call the existing package-level `compute*` function
4. Output JSON (or pretty-print for synthesis endpoints)

## Step-by-Step Plan

### Step 1: Extract ephemeris init into a helper
Create `initEphe()` that calls `EnsureEpheCache()` + `swe.SetEphePath()` + `swe.SetSidMode()`. Used by every subcommand.

### Step 2: Add subcommands (one per missing endpoint)

Each follows the same pattern as the existing `transit`/`synastry`/`event` subcommands:

```go
if len(os.Args) >= 2 && os.Args[1] == "patterns" {
    fs := flag.NewFlagSet("patterns", flag.ExitOnError)
    jsonOut := fs.Bool("json", false, "output as JSON")
    orbDeg := fs.Float64("orb", 5.0, "max orb in degrees")
    fs.Parse(os.Args[2:])
    args := fs.Args()
    // ... parse birth data ...
    initEphe()
    cd := computePositions(year, month, day, hour, minute, tzOff, lat, lng, "")
    // ... call compute logic, output ...
    return
}
```

**Subcommands to add (with arg signatures):**

| Subcommand | Args | Flags | Compute call |
|-----------|------|-------|-------------|
| `patterns` | NAME Y M D H MIN TZ LAT LNG | `--json`, `--orb 5` | `computePositions` + `DetectPatterns` |
| `draconic` | NAME Y M D H MIN TZ LAT LNG | `--json`, `--orb 5` | `computeDraconic` |
| `draconic-synastry` | NAME1 Y1 M1 D1 H1 MIN1 TZ1 LAT1 LNG1 NAME2 Y2 M2 D2 H2 MIN2 TZ2 LAT2 LNG2 | `--json`, `--orb 5` | `computeDraconicSynastry` |
| `draconic-synastry-full` | (same as draconic-synastry) | `--json`, `--orb 5` | `computeDraconicSynastryFull` |
| `draconic-interp` | NAME Y M D H MIN TZ LAT LNG | `--json`, `--orb 5` | `computeDraconicInterpretation` |
| `synastry-interp` | NAME1 Y1 M1 D1 H1 MIN1 TZ1 LAT1 LNG1 NAME2 Y2 M2 D2 H2 MIN2 TZ2 LAT2 LNG2 | `--json`, `--orb 5` | `computeSynastryInterpretation` |
| `transit-interp` | NAME Y M D H MIN TZ LAT LNG START END | `--json`, `--orb 5` | `computeTransitInterpretation` |
| `transit-patterns` | NAME Y M D H MIN TZ LAT LNG START END | `--json`, `--orb 5` | `computeTransitPatterns` |
| `stars` | NAME Y M D H MIN TZ LAT LNG | `--json`, `--orb 2` | `computeStars` |
| `directions` | NAME Y M D H MIN TZ LAT LNG AGE | `--json`, `--orb 5` | `computeDirections` |
| `interpretation` | NAME Y M D H MIN TZ LAT LNG | `--json`, `--orb 5`, `--age 0`, `--sect day` | `computeInterpretation` |
| `arabic-parts` | NAME Y M D H MIN TZ LAT LNG | `--json`, `--orb 5` | `computeArabicParts` |
| `traditional` | NAME Y M D H MIN TZ LAT LNG | `--json` | `computeTraditional` |
| `horoscope` | NAME Y M D H MIN TZ LAT LNG START END | `--json`, `--orb 5`, `--age 0`, `--sect day` | `computeHoroscope` |
| `evolutionary` | NAME Y M D H MIN TZ LAT LNG | `--json`, `--orb 5` | `computeEvolutionary` |
| `compare` | NAME Y M D H MIN TZ LAT LNG | `--json`, `--orb 5`, `--systems koine,western,vedic,bazi`, `--age 0`, `--sect day` | `computeCompare` |
| `relocation` | NAME Y M D H MIN TZ LAT LNG LAT_A LNG_A LAT_B LNG_B | `--json`, `--name-a`, `--name-b` | `computeRelocation` |
| `chart` | NAME Y M D H MIN TZ LAT LNG | `--output chart.svg`, `--house-system W`, `--sidereal`, `--aspects`, `--outer`, `--highlight`, `--pattern-orb 5` | `dignity.RenderChartSVG` |

### Step 3: Pretty-print for synthesis endpoints

For synthesis endpoints (`interpretation`, `event`, `horoscope`, `draconic-interp`, `synastry-interp`, `transit-interp`), when `--json` is NOT set, pretty-print the synthesis report (opening, body sections, closing) — same pattern as the existing `event` subcommand (lines 317-330).

For data endpoints (`patterns`, `draconic`, `stars`, `directions`, `arabic-parts`, `traditional`, `transit-patterns`), default to JSON output since there's no natural text format.

### Step 4: Update usage text

Update the default `recover` usage to list all subcommands.

### Step 5: Tests

Add a CLI integration test that runs each new subcommand with `--json` and verifies:
- Exit code 0
- Valid JSON output
- Expected top-level keys present

### Step 6: Update skills

Update `empirical-go-api` skill's CLI reference table and `koine-development` skill's `references/cli-and-api.md`.

## Files Changed

- `cmd/recover/main.go` — add 18 subcommand blocks + `initEphe()` helper
- `cmd/recover/main_test.go` — add CLI integration tests
- Skills: `empirical-go-api`, `koine-development/references/cli-and-api.md`

## Risks

- **main.go is already 1886 lines.** Adding 18 subcommands will push it past 2500. Consider extracting subcommands into `cmd/recover/cli.go` or similar — but that's a refactor, not strictly necessary for this task.
- **Some compute functions take `cacheDir` from the closure.** The package-level functions handle empty `cacheDir` by calling `EnsureEpheCache()` themselves, so this works. But verify each one.
- **`computeRelocation` takes `server.LatLng` structs.** The CLI needs to parse these from positional args or flags.

## Open Questions

1. **Subcommand naming:** `draconic-interp` vs `draconic-interpretation`? Prefer shorter names for CLI ergonomics.
2. **Chart output:** SVG to stdout or to file? `--output` flag for file, default stdout.
3. **Relocation args:** Two locations with names — positional or `--name-a`/`--name-b` flags? Flags are cleaner since locations have lat/lng pairs.
