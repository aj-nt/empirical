# `/api/compare` — Multi-System Computation Endpoint

**Goal:** One endpoint that takes birth data + system list, returns full computation for each system side by side. No interpretation — pure structured data.

**Default systems:** Koiné + Modern Western (when `systems` omitted).

## Architecture

```
POST /api/compare
  → server.go: CompareFunc handler
    → cmd/recover/main.go: computeCompare()
      → internal/dignity/compare.go: SystemComparison struct + ComputeComparison()
        → reuses: WesternDignity, VedicDignity, ComputeDignityConvergence,
          ComputeZodiacComparison, ComputeNodeConvergence, ComputeTimingConvergence,
          DetectPatterns, ComputeDecans, ComputeTerms, TriplicityRuler,
          AnnualProfection, GetNakshatra, ComputeVimshottariDasha, CurrentDasha,
          ComputeDraconic, ComputeDraconicSignShifts, ComputeDraconicBridges,
          PartCatalog, ComputeFortune, etc.
```

## Response structure

```json
{
  "name": "AJ",
  "birth": { "year": 1969, "month": 2, "day": 15, "hour": 23, "minute": 10, "tz_offset": -8, "lat": 47.038, "lng": -122.901 },
  "orb": 5.0,
  "age": 57,
  "sect": "night",
  "systems": {
    "koine": { ... },
    "hellenistic": { ... },
    "modern_western": { ... },
    "vedic": { ... },
    "arabic": { ... },
    "draconic": { ... }
  },
  "convergences": {
    "dignity": { "signal_count": 5, "noise_count": 2, "convergence_rate": 0.714 },
    "node_axis": { "axis_preserved": true, "sign_match": false },
    "timing": { "planet_convergences": ["Saturn"], "convergence_count": 1 }
  }
}
```

### Per-system block

| System | Zodiac | Houses | Planets | Dignity | Aspects | Extras |
|---|---|---|---|---|---|---|
| koine | Tropical | Whole-sign | 10 + Node | 3-state (dom/exalt/fall) | 5 Ptolemaic | Patterns, stars |
| hellenistic | Tropical | Whole-sign | 7 classical + Node | 3-state + triplicity + terms + decans | 5 Ptolemaic | Sect, profection |
| modern_western | Tropical | Placidus | 10 + Node + Chiron + Lilith | 5-state (+detriment) | 5 Ptolemaic + quincunx + semi-sextile | Patterns |
| vedic | Sidereal (Lahiri) | Whole-sign | 9 grahas + Node | Own/exalted/debilitated | Graha drishti | Nakshatras, dasha |
| arabic | Tropical | Whole-sign | 7 classical + Node | 5-state + triplicity + terms + decans | 5 Ptolemaic | Arabic parts, almuten |
| draconic | Draconic (trop - NN) | Whole-sign | 10 + Node | 3-state (applied to draconic) | 5 Ptolemaic | Bridges, sign shifts |

## Files to create

### 1. `internal/dignity/compare.go` (new, ~500 lines)

Core types and computation:

```go
type SystemComparison struct {
    Name     string
    Birth    BirthData
    Orb      float64
    Age      int
    Sect     string
    Systems  map[string]SystemBlock
    Convergences ConvergenceBlock
}

type SystemBlock struct {
    Zodiac   string              // "tropical" | "sidereal" | "draconic"
    Houses   string              // "whole_sign" | "placidus"
    Planets  []PlanetEntry       // planet positions + dignity + house
    Aspects  []AspectEntry       // aspect list
    Patterns []PatternEntry      // geometric patterns
    Extras   map[string]any      // system-specific extras
}

type PlanetEntry struct {
    Planet   string  `json:"planet"`
    Lon      float64 `json:"lon"`
    Sign     string  `json:"sign"`
    Degree   float64 `json:"degree"`
    House    int     `json:"house"`
    Dignity  string  `json:"dignity"`
    Nakshatra string `json:"nakshatra,omitempty"` // vedic only
    Speed    float64 `json:"speed,omitempty"`      // retrograde detection
}

type AspectEntry struct {
    Planet1 string  `json:"planet1"`
    Planet2 string  `json:"planet2"`
    Aspect  string  `json:"aspect"`
    Orb     float64 `json:"orb"`
}

type PatternEntry struct {
    Name    string   `json:"name"`
    Planets []string `json:"planets"`
}

type ConvergenceBlock struct {
    Dignity   DignityConvergence   `json:"dignity"`
    NodeAxis  NodeConvergence      `json:"node_axis"`
    Timing    TimingConvergence    `json:"timing"`
}
```

`ComputeComparison(name, year, month, day, hour, minute int, tzOff, lat, lng float64, systems []string, orbDeg float64, age int, sect string, cacheDir string) (*SystemComparison, error)`:

1. Call `computePositions()` once for tropical longitudes + ayanamsa
2. For each requested system, call the system-specific builder
3. Compute convergences (dignity, node, timing) — these are cross-system
4. Return the full comparison

### System builders (all in compare.go)

**buildKoinéBlock(cd, orb, age, sect):**
- Tropical zodiac, whole-sign houses
- 10 planets + Node, 3-state dignity (dom/exalt/fall)
- 5 Ptolemaic aspects via `DefaultAspects()`
- Patterns via `DetectPatterns`
- Stars via `FindStarConjunctions` (top 5)
- Profection via `AnnualProfection`

**buildHellenisticBlock(cd, orb, age, sect):**
- Tropical zodiac, whole-sign houses
- 7 classical planets + Node, 3-state dignity + triplicity + terms + decans
- 5 Ptolemaic aspects
- Sect, profection

**buildModernWesternBlock(cd, orb, age, sect):**
- Tropical zodiac, Placidus houses
- 10 planets + Node + Chiron (SWE 15) + Lilith (mean, SWE 12)
- 5-state dignity (dom/exalt/detriment/fall/peregrine)
- 7 aspects (5 Ptolemaic + quincunx + semi-sextile)
- Patterns via `DetectPatterns` (12-planet map)

**buildVedicBlock(cd, orb, age, sect):**
- Sidereal zodiac (tropical - ayanamsa), whole-sign houses
- 9 grahas (Sun-Pluto minus Uranus/Neptune/Pluto, plus Node)
- Vedic dignity (swakshetra/uchcha/neecha/peregrine)
- Graha drishti aspects (NEW — see below)
- Nakshatra per planet via `GetNakshatra`
- Current dasha via `ComputeVimshottariDasha` + `CurrentDasha`

**buildArabicBlock(cd, orb, age, sect):**
- Tropical zodiac, whole-sign houses
- 7 classical planets + Node, 5-state dignity + triplicity + terms + decans
- 5 Ptolemaic aspects
- Arabic parts via `PartCatalog` + `ComputeFortune`
- Almuten computation (NEW — see below)

**buildDraconicBlock(cd, orb, age, sect):**
- Draconic zodiac (tropical - NN), whole-sign houses
- 10 planets + Node, 3-state dignity applied to draconic signs
- 5 Ptolemaic aspects
- Bridges via `ComputeDraconicBridges`
- Sign shifts via `ComputeDraconicSignShifts`

## New code needed (not reusing existing)

### A. Vedic graha drishti (~40 lines, in compare.go or new file)

Planet-specific aspect rules:
- Sun: 7th house (opposition)
- Moon: 7th house
- Mars: 4th, 7th, 8th houses
- Mercury: 7th house
- Jupiter: 5th, 7th, 9th houses
- Venus: 7th house
- Saturn: 3rd, 7th, 10th houses

Implementation: for each planet pair, compute house distance (1-12), check if distance matches any of the planet's drishti houses. Return aspect name + orb (from exact house cusp distance).

### B. Vedic nakshatra placement per planet (~30 lines, in compare.go)

Loop over 9 grahas, call `GetNakshatra(siderealLon)` for each. Store nakshatra name + pada in PlanetEntry.

### C. Vedic dasha period for target date (~60 lines, in compare.go)

Given birth data + target date (today or provided date):
1. Compute Moon's nakshatra at birth via `GetNakshatra`
2. Call `ComputeVimshottariDasha` for full sequence
3. Call `CurrentDasha` for the target date
4. Return current mahadasha planet + start/end dates

### D. Modern Western Chiron + Lilith (~5 lines, in compare.go)

Add Chiron (SWE ID 15) and Lilith mean (SWE ID 12) to the planet set for modern_western. Call `swe.CalcUT(jd, planetID, ...)` to get longitudes. These are NOT in `computePositions` (which only does 10 planets) — compute them separately in the modern_western builder.

### E. Arabic almuten (~50 lines, in compare.go)

Weighted scoring of planets by dignity:
- Domicile: 5 points
- Exaltation: 4 points
- Triplicity: 3 points
- Term: 2 points
- Decan: 1 point
- Fall: -4 points
- Detriment: -5 points

Score each of the 7 classical planets across all placements. Return the planet with the highest score as almuten, plus the full score table.

### F. Draconic sign shifts (~10 lines, in compare.go)

Already computed by `ComputeDraconicSignShifts`. Just extract into the SystemBlock.

## Files to modify

### 2. `internal/server/server.go` (+~50 lines)

- New `CompareFunc` type (~line 95, after HoroscopeFunc):

```go
type CompareFunc func(name string, year, month, day, hour, minute int, tzOff, lat, lng float64, systems []string, orbDeg float64, age int, sect string) ([]byte, error)
```

- New `/api/compare` handler (~line 726, after horoscope handler):

```go
mux.HandleFunc("/api/compare", func(w http.ResponseWriter, r *http.Request) {
    // POST only, decode CompareRequest, call compare(), return JSON
})
```

- `CompareRequest` struct:

```go
type CompareRequest struct {
    ChartRequest
    Systems []string `json:"systems"`
    Age     int      `json:"age"`
    Sect    string   `json:"sect"`
}
```

- Add `compare CompareFunc` parameter to `NewMux` and `Run` signatures.

### 3. `cmd/recover/main.go` (+~30 lines)

- New `computeCompare` function (~line 945, after computeHoroscope):

```go
func computeCompare(name string, year, month, day, hour, minute int, tzOff, lat, lng float64, systems []string, orbDeg float64, age int, sect string, cacheDir string) ([]byte, error) {
    comp, err := dignity.ComputeComparison(name, year, month, day, hour, minute, tzOff, lat, lng, systems, orbDeg, age, sect, cacheDir)
    if err != nil { return nil, err }
    return json.Marshal(comp)
}
```

- Wire `compare` closure in `main()` (~line 170, after horoscope):

```go
compare := func(name string, year, month, day, hour, minute int, tzOff, lat, lng float64, systems []string, orbDeg float64, age int, sect string) ([]byte, error) {
    return computeCompare(name, year, month, day, hour, minute, tzOff, lat, lng, systems, orbDeg, age, sect, cacheDir)
}
```

- Add `compare` to `server.Run()` call and `server.NewMux()` call.

## Default systems

When `systems` is empty or omitted, default to `["koine", "modern_western"]`.

## Tests

### `internal/dignity/compare_test.go` (new)

- `TestComputeComparison_AllSystems` — full comparison with all 6 systems, verify structure
- `TestComputeComparison_DefaultSystems` — empty systems list → koine + modern_western
- `TestComputeComparison_KoinéOnly` — single system
- `TestVedicGrahaDrishti` — verify Mars aspects 4/7/8, Jupiter 5/7/9, Saturn 3/7/10
- `TestArabicAlmuten` — verify scoring produces correct almuten for known chart
- `TestModernWestern_ChironLilith` — verify Chiron and Lilith present in modern_western block

### `internal/server/server_test.go` (modify)

- `TestCompareEndpoint` — POST to /api/compare, verify 200 + valid JSON structure

### `cmd/recover/main_test.go` (modify)

- `TestComputeCompare_Integration` — full integration test with AJ's birth data

## Verification

1. `go build -o /tmp/koine ./cmd/recover/`
2. `strings /tmp/koine | grep compare` — verify new code in binary
3. `go test ./...` — all tests green
4. Start server, curl test:
```bash
curl -s -X POST http://localhost:5433/api/compare \
  -H 'Content-Type: application/json' \
  -d '{"name":"AJ","year":1969,"month":2,"day":15,"hour":23,"minute":10,"tz_offset":-8,"lat":47.038,"lng":-122.901,"systems":["koine","vedic"],"orb":5,"age":57,"sect":"night"}' | jq '.systems | keys'
# Should show ["koine", "vedic"]
```

## Risks

- **Chiron/Lilith SWE IDs**: Chiron is 15, Lilith mean is 12. Verify these return valid longitudes via `swe.CalcUT`. If Lilith mean fails, fall back to Lilith true (SWE 13) or skip.
- **Graha drishti orb**: Vedic aspects are house-based, not degree-based. The "orb" is the house cusp distance. This is a different model than Ptolemaic degree orbs — document the difference in the response.
- **Almuten scoring**: The weights are standard Medieval but there are variants. Use the most common weighting (Bonatti). Document in code comments.
- **Performance**: 6 systems × compute is ~6× the work of a single interpretation call. Should still be <1s for a single chart. If slow, add caching for `computePositions` result.
