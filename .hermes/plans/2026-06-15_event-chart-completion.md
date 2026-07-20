# Event Chart Mode — Completion Plan

**Goal:** Make the event chart mode fully accessible to users via CLI, web UI, and documentation.

**Date:** 2026-06-15

## Current State

The synthesis engine already supports event chart mode. What's done:

- `ChartMode` enum (`ModeNatal`, `ModeEvent`) in `internal/dignity/synthesis.go`
- `SynthesizeChart` accepts `mode ChartMode` — controls opening/closing framing, Moon section, angular display, noon warning suppression
- `buildOpening`: "The moment of X is characterized by" (event) vs "X's chart shows" (natal)
- `buildClosing`: "The moment is predominantly" (event) vs "Overall, X's chart is" (natal)
- `buildMoonOfMomentSection`: dedicated Moon section for event charts, always shown
- Angular emphasis always shown in event mode (events have known times)
- No noon placeholder warning in event mode
- `computeEventChart` in `cmd/recover/main.go` — full compute: positions, whole-sign houses, 3 aspects, patterns (with Node), stars, Fortune, synthesis
- `EventChartFunc` type and `/api/event-chart` POST endpoint in `internal/server/server.go`
- `eventChart` closure wired in `main()` serve subcommand
- 4 passing tests: opening framing, closing framing, no noon warning, Moon elevated
- `go build ./...` passes, all tests green

## Gaps

### 1. CLI subcommand (missing)

There's no `event` subcommand. Users can only reach event charts via the HTTP API.

**Add:** `event` subcommand to `cmd/recover/main.go`:
```
Usage: recover event [--json] [--orb 5] NAME Y M D H MIN TZ LAT LNG
Example: recover event "Product Launch" 2026 6 15 14 30 -4 40.7128 -74.006
```

Follow the pattern of the `transit` subcommand (lines 180-217). Call `computeEventChart` directly.

### 2. Web UI (missing)

`web/index.html` has no event chart form or display. Need a simple form like the natal interpretation form.

**Add:** Event chart section to the dashboard with:
- Name, date/time, timezone, location fields
- Orb slider (default 5.0)
- Submit → calls `/api/event-chart`
- Display synthesis report (opening, body sections, closing)

### 3. Orb default inconsistency

`/api/event-chart` defaults to `orb=3.0` (server.go line 468). `/api/interpretation` defaults to `orb=5.0`. The skill says "Pattern detection needs orb ≥5 to find Cradles." Event charts at orb=3.0 will miss multi-point patterns.

**Fix:** Change event-chart default orb to 5.0 to match interpretation endpoint.

### 4. Documentation (missing)

`MANUAL.md` has no entry for `/api/event-chart` or the event subcommand.

**Add:** Document the endpoint and CLI subcommand in MANUAL.md.

### 5. Integration test (missing)

No test for `computeEventChart` itself — only synthesis unit tests.

**Add:** Test in `cmd/recover/main_test.go` (or a new test file) that calls `computeEventChart` with known data and verifies the JSON output structure.

## Step-by-step

1. **Fix orb default** — `server.go` line 468: `orb = 3.0` → `orb = 5.0`
2. **Add CLI subcommand** — `cmd/recover/main.go`: new `event` case after `synastry` block
3. **Add integration test** — test `computeEventChart` with AJ's chart data, verify JSON structure
4. **Update MANUAL.md** — document `/api/event-chart` endpoint and `event` CLI subcommand
5. **Update web UI** — add event chart form to `web/index.html`
6. **Verify** — `go build ./...`, `go test ./...`, manual curl test

## Files to change

- `internal/server/server.go` — orb default (1 line)
- `cmd/recover/main.go` — CLI subcommand (~30 lines)
- `cmd/recover/main_test.go` — integration test (~40 lines)
- `MANUAL.md` — documentation (~20 lines)
- `web/index.html` — UI form (~80 lines)

## Risks

- Web UI changes are the most work and most likely to introduce bugs. The dashboard already has multiple forms; follow the existing pattern closely.
- The `event` CLI subcommand needs ephemeris access — it must call `empirical.EnsureEpheCache()` and `swe.SetEphePath()` like the other subcommands. Currently only `serve` does this; `transit` and `synastry` pass `""` for cacheDir. Need to add cache init to the event subcommand (and ideally backport to transit/synastry, but that's out of scope).
