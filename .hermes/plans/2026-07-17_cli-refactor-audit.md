# GoF + Fowler Audit: CLI Expansion (2026-07-17)

**Scope:** `cmd/recover/main.go` (1,886 lines) — CLI dispatch + all compute functions + server wiring.
**Context:** Plan to add 18 CLI subcommands for API parity. Each subcommand is ~30 lines of arg parsing + a call to an existing `compute*` function.
**Green baseline:** All 11 test packages pass. `go test ./...` green.

## Architecture Scan

```
cmd/recover/main.go — 1,886 lines, three concerns:
  Lines 1-17:    imports
  Lines 19-197:  serve subcommand (closures + server.Run)
  Lines 199-373: transit, synastry, event, recover subcommands
  Lines 375-1886: 30+ compute* functions + helpers + printReport
```

The file is a **God File** (Fowler): CLI dispatch, compute logic, and server wiring in one file. The compute functions are package-level (not closures), so they're already callable from anywhere — they just happen to live in the wrong file.

## Good Patterns Already Present

- `computePositions()` — single ephemeris call point, used by all compute functions
- `chartData` struct — pre-computed positions passed to compute functions, avoids redundant SWE calls
- `compute*` functions are package-level, not closures — already callable from CLI subcommands with `cacheDir=""` (triggers lazy init)
- `--json` flag pattern is consistent across all 4 existing subcommands
- `flag.NewFlagSet` per subcommand — clean flag isolation

## Smell Inventory

### P1: Duplicated Birth Data Parsing (Extract Method)

**Every subcommand** parses the same 9 birth data fields from `args[0:9]`:

```go
name := args[0]
year, _ := strconv.Atoi(args[1])
month, _ := strconv.Atoi(args[2])
day, _ := strconv.Atoi(args[3])
hour, _ := strconv.Atoi(args[4])
minute, _ := strconv.Atoi(args[5])
tzOff, _ := strconv.ParseFloat(args[6], 64)
lat, _ := strconv.ParseFloat(args[7], 64)
lng, _ := strconv.ParseFloat(args[8], 64)
```

This 9-line block appears in `transit` (lines 213-221), `synastry` (lines 252-269, ×2), `event` (lines 298-306), and `recover` (lines 351-359). Adding 18 subcommands means 18 more copies.

**Fix:** Extract `parseBirthData(args []string) (name string, year, month, day, hour, minute int, tzOff, lat, lng float64, err error)`. Also extract the two-person variant for synastry subcommands.

**Net:** ~-150 lines (removes 9×~22 copies, adds ~15 for the helper). **Risk:** SAFE — pure extraction, same semantics.

### P2: Duplicated Output Pattern (Extract Method)

Every subcommand has the same output dispatch:

```go
if *jsonOut {
    fmt.Println(string(result))
} else {
    // pretty-print or raw
}
```

The `event` subcommand adds a `json.Unmarshal` + `SynthesisReport` pretty-print step. The `transit`/`synastry` subcommands just print raw. This pattern will be repeated 18 more times.

**Fix:** Extract `outputResult(result []byte, jsonOut bool, prettyPrint func([]byte) string)`. For data endpoints (no pretty-print), pass `nil`. For synthesis endpoints, pass an unmarshal+print function.

**Net:** ~-50 lines. **Risk:** SAFE.

### P3: Magic Number Arg Counts (Replace Magic Number)

```go
if len(args) < 9   // birth data
if len(args) < 11  // birth data + start/end dates
if len(args) < 18  // two birth data sets
```

These encode the birth data tuple size. If a field is added (e.g., `location string`), every magic number must be updated.

**Fix:** Define `const birthDataArgCount = 9`. Use `birthDataArgCount`, `birthDataArgCount+2`, `birthDataArgCount*2`.

**Risk:** SAFE.

### P4: Linear Subcommand Dispatch (Replace Conditional with Polymorphism / Strategy)

18 `if len(os.Args) >= 2 && os.Args[1] == "X"` blocks in sequence. Each block is ~30 lines. The dispatch is O(n) linear scan. Adding a subcommand means inserting another `if` block in the right position.

**Fix:** Define a `Subcommand` struct and registry:

```go
type Subcommand struct {
    Name     string
    Usage    string
    MinArgs  int
    Run      func(args []string, jsonOut bool, orbDeg float64) error
}

var subcommands = map[string]Subcommand{...}
```

Main dispatch becomes:
```go
if len(os.Args) >= 2 {
    if cmd, ok := subcommands[os.Args[1]]; ok {
        // parse flags, validate args, call cmd.Run
    }
}
```

**Net:** ~-200 lines (replaces 18×~30 line blocks with 18×~5 line registrations + one ~20 line dispatcher). **Risk:** SAFE — mechanical transformation, same behavior.

### P5: God File — Compute Functions in CLI Entry Point (Extract File)

Lines 375-1886 (~1,500 lines) are `compute*` functions and helpers. They're package-level, not closures — they don't need to be in `main.go`. They're already callable from anywhere in the `main` package.

**Fix:** Move to `cmd/recover/compute.go` (or split by domain: `compute_natal.go`, `compute_transit.go`, `compute_draconic.go`, `compute_synthesis.go`).

**Risk:** SAFE — same package, no import changes. **Caveat:** Check `.gitignore` for `recover` pattern (known pitfall from gof-refactoring skill).

### P6: Long Parameter List — Birth Data Tuple (Introduce Parameter Object)

P11 from the 2026-07-13 audit. The 9-arg birth data tuple `(name, year, month, day, hour, minute, tzOff, lat, lng)` appears in:
- Every `compute*` function signature
- Every subcommand's arg parsing
- The `chartData` struct (partial — has positions but not name/tzOff)

**Fix:** Define `BirthData` struct and use it in compute function signatures. This is a larger refactor (touches all compute functions + server.go function types) — defer to a separate pass.

**Risk:** CAREFUL — large API surface. Defer to P2 pass.

### P7: Redundant Closures in Serve Block

Lines 39-177: 18 closures that are thin wrappers around package-level `compute*` functions. They exist only to thread `cacheDir` through. The package-level functions already handle `cacheDir=""` by calling `EnsureEpheCache()` themselves.

**Observation:** The closures ARE the server's function types. `server.Run()` takes 25 function parameters typed as `server.ComputeFunc`, `server.TransitFunc`, etc. The closures adapt the package-level functions to those types. This is the Adapter pattern — not a smell, just verbose.

**Verdict:** NOT a smell to fix now. The closures are the Adapter pattern between `server.Run()`'s function types and the package-level compute functions. Fixing this requires changing `server.Run()`'s signature (P11 from previous audit — `BirthData` struct), which is a separate pass.

## Prioritized Implementation Plan

### Pass 1 (this session): P1 + P2 + P3 + P4 — Structural cleanup before expansion

1. **P1**: Extract `parseBirthData()` and `parseTwoBirthData()` helpers
2. **P2**: Extract `outputResult()` helper
3. **P3**: Define `birthDataArgCount` constant
4. **P4**: Build `Subcommand` registry
5. **Add 18 subcommands** using the registry (now ~5 lines each instead of ~30)
6. **P5**: Move compute functions to `compute.go`

**Net line count:** ~1,886 → ~800 (main.go) + ~1,500 (compute.go). main.go becomes the CLI dispatcher; compute.go holds the engine.

### Pass 2 (separate session): P6 — BirthData struct

Introduce `BirthData` struct, update all compute function signatures, update `server.go` function types. This is P11 from the previous audit — deferred then, still deferred now, but the CLI expansion makes it more valuable.

### Pass 3 (defer): P7 — Eliminate serve-block closures

Only worth doing after P6 (BirthData struct simplifies the function types).

## What the Plan Got Right

The original plan correctly identified:
- All 18 compute functions already exist
- Each subcommand is ~30 lines of arg parsing + a function call
- `cacheDir=""` triggers lazy ephemeris init in the package-level functions

## What the Plan Missed

- **The duplication IS the problem.** Adding 18 subcommands inline means 18 copies of the same 9-line birth data parse, 18 copies of the same output dispatch, 18 copies of the same flag setup. That's ~500 lines of duplicated boilerplate.
- **The God File gets worse.** 1,886 + (18 × 30) = 2,426 lines in one file. The compute functions are already 1,500 lines of that — they should be extracted regardless.
- **No extraction before expansion.** The Fowler principle: extract the duplication first, then add new cases using the extracted abstraction. Adding 18 subcommands to the current structure is like adding 18 more copy-pasted blocks to an already-bloated file.

## Risks

- **`.gitignore` trap**: Pattern `recover` in `.gitignore` (intended to ignore the binary) will also match `cmd/recover/compute.go`. Fix: use `/recover` (root-only). Always `git status` after creating files in `cmd/recover/`.
- **patch tool + Go braces**: Known pitfall. Use Python scripts for multi-line Go edits.
- **Subcommand registry changes dispatch order**: Currently `serve` is checked first, then `transit`, `synastry`, `event`, then `recover` (default). Registry makes dispatch O(1) but changes the order. `serve` must still short-circuit before the registry lookup since it has different arg handling (no flags, takes port).
