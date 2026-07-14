# Modern Western Natal Chart Interpretation

**Goal:** Add a modern Western natal chart interpretation output path to the empirical engine, selectable alongside the existing Koiné (Hellenistic) path.

## Current State

- `computeInterpretation` in `cmd/recover/compute.go` hardcodes `dignity.KoinéFromBase(bc, orbDeg)` — only Hellenistic output
- `InterpretChart` in `western_interpretation.go` exists and works, but is only called by `mundane.go`
- All the interpretation pieces exist: planet-in-sign, planet-in-house, aspects, patterns, star conjunctions, modern dignities (domicile/detriment/exaltation/fall for all planets including outers)
- Missing: a `WesternFromBase` function that mirrors `KoinéFromBase` but routes through the modern Western engine

## Approach

1. **Create `WesternFromBase`** — mirrors `KoinéFromBase` but uses `InterpretChart` instead of `KoineInterpretChart`. Key differences:
   - All planets (not just classical 7)
   - Modern planet/sign/house descriptions
   - Modern dignities (domicile/detriment/exaltation/fall for outers too)
   - No sect-based dignity
   - No triplicity
   - Includes star conjunction interpretations
   - Includes pattern interpretations with modern descriptions

2. **Add `system` parameter** to the interpretation pipeline:
   - `InterpretationFunc` signature gains a `system string` parameter
   - `computeInterpretation` routes to Koiné or Western based on system
   - Server endpoint `/api/interpretation` accepts `"system": "koine"` or `"system": "western"` in request body
   - `recover` CLI gets `--system` flag (default: `"koine"` for backward compatibility)

3. **TDD**: Write tests for `WesternFromBase` first, then implement.

## Files to Change

| File | Change |
|------|--------|
| `internal/dignity/western_from_base.go` | **NEW** — `WesternFromBase(bc *BaseChart, orbDeg float64) *ChartInterpretation` |
| `internal/dignity/western_from_base_test.go` | **NEW** — tests for WesternFromBase |
| `internal/dignity/western_interpretation.go` | Add `InterpretStarConjunctions` helper (batch star interpretation) |
| `internal/server/server.go` | `InterpretationFunc` gains `system` param; endpoint parses `system` from request |
| `cmd/recover/compute.go` | `computeInterpretation` gains `system` param, routes to Koiné or Western |
| `cmd/recover/main.go` | `interpretation` closure passes `system` through; CLI `--system` flag |
| `internal/server/server_test.go` | Update mock and add system routing tests |

## Step-by-Step

### Step 1: `WesternFromBase` (TDD)
- Write `TestWesternFromBase_ProducesAllPlanets` — verifies outers (Uranus, Neptune, Pluto) appear
- Write `TestWesternFromBase_IncludesStarConjunctions` — verifies star interpretations in output
- Write `TestWesternFromBase_ModernDignities` — verifies modern dignity language (no sect, no triplicity)
- Write `TestWesternFromBase_JSON` — round-trip test
- Implement `WesternFromBase`

### Step 2: `InterpretStarConjunctions` helper
- Add batch function to `western_interpretation.go` that takes `[]StarConjunction` and returns `[]string`
- Test it

### Step 3: Wire `system` parameter through the stack
- Update `InterpretationFunc` type signature
- Update `computeInterpretation` to route based on system
- Update server endpoint to parse `system` field
- Update `recover` CLI closure and add `--system` flag
- Update server test mocks

### Step 4: Integration test
- `recover` with `--system western` produces modern output
- `/api/interpretation` with `"system": "western"` produces modern output
- Backward compat: default `"koine"` produces same output as before

## Risks / Tradeoffs

- **Backward compatibility**: Default stays `"koine"` — no existing callers break
- **`ChartInterpretation` struct**: Currently has `PlanetSigns`, `PlanetHouses`, `Aspects`, `Patterns` — Western output fits the same struct, so no API break
- **Star conjunctions**: `ChartInterpretation` doesn't have a `Stars` field. Options: (a) add one, (b) append star interpretations to `Aspects`. Prefer (a) — cleaner separation.
- **`InterpretChart` signature**: Currently takes `dignities []PlanetDignity` but never uses them (the param is always `nil` in callers). WesternFromBase won't pass it either. Consider removing the unused param later.

## Verification

```bash
# Unit tests
go test -run TestWesternFromBase -count=1 ./internal/dignity/...

# Server tests
go test -run TestInterpretation -count=1 ./internal/server/...

# CLI integration
go run ./cmd/recover/ recover "AJ" 1969 2 15 23 10 -8 47.038 -122.901 --system western | head -50
```
