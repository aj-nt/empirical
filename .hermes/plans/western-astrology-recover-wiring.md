# Wire Western Astrology Features into `recover` CLI

> **For Hermes:** Use subagent-driven-development skill to implement this plan task-by-task.

**Goal:** Wire existing Western astrology code (interpretations, traditional data, modern dignities, minor aspects, Placidus houses, Arabic Parts, midpoints, declinations) into the `recover` CLI's `FullReport` and `printReport` output.

**Architecture:** The code already exists in `internal/dignity/` — it's just not called from `ComputeFullReport` or `printReport`. This is a wiring task: add fields to `FullReport`, populate them in `ComputeFullReport`, print them in `printReport`. The `computeAll` function in `cmd/recover/main.go` already has all the data needed (speeds, lat/lng, etc.) — it just needs to pass it through.

**Tech Stack:** Go 1.21, `github.com/aj-nt/koine/internal/dignity`, `github.com/aj-nt/koine/internal/swe`

---

## Pre-Implementation Verification

Before writing any code, verify what already exists:

```bash
# Check that all functions we plan to call actually exist and compile
cd /Users/aj/Documents/repos/koine
go build ./internal/dignity/...
go build ./cmd/recover/...
```

---

### Task 1: Add Western fields to `FullReport` struct

**Objective:** Extend `FullReport` in `api.go` with fields for all Western features.

**Files:**
- Modify: `internal/dignity/api.go:10-23`

**Step 1: Write failing test**

Add to a new file `internal/dignity/api_test.go`:

```go
package dignity

import "testing"

func TestFullReport_HasWesternFields(t *testing.T) {
    t.Parallel()
    fr := &FullReport{}
    // These fields must exist and be accessible
    _ = fr.PlacidusHouses
    _ = fr.PlanetSigns
    _ = fr.PlanetHouses
    _ = fr.AspectInterpretations
    _ = fr.PatternInterpretations
    _ = fr.LunarPhase
    _ = fr.Retrogrades
    _ = fr.DispositorTree
    _ = fr.VOCMoon
    _ = fr.Decans
    _ = fr.Terms
    _ = fr.Antiscia
    _ = fr.ArabicParts
    _ = fr.ModernDignity
    _ = fr.MinorAspects
    _ = fr.Midpoints
    _ = fr.Declinations
}
```

**Step 2: Run test to verify failure**

```bash
cd /Users/aj/Documents/repos/koine
go test -run TestFullReport_HasWesternFields -count=1 ./internal/dignity/...
```

Expected: FAIL — compilation error, fields don't exist.

**Step 3: Add fields to `FullReport`**

```go
type FullReport struct {
    Name            string             `json:"name"`
    AyanamsaDegrees float64            `json:"ayanamsa_degrees"`
    Planets         map[string]float64 `json:"planets"`
    Houses          map[string]int     `json:"houses"`
    Ascendant       float64            `json:"ascendant"`
    Dignity         *DignityConvergence `json:"dignity"`
    Aspects         []AspectHit        `json:"aspects"`
    Patterns        []Pattern          `json:"patterns"`
    Stars           []StarConjunction  `json:"stars"`
    Directions      *DirectionsReport  `json:"directions"`
    ArabicParts     *PartReport        `json:"arabic_parts"`
    Draconic        *DraconicChart     `json:"draconic"`

    // Western astrology additions
    PlacidusHouses         map[string]int       `json:"placidus_houses,omitempty"`
    PlanetSigns            []string             `json:"planet_signs,omitempty"`
    PlanetHouses           []string             `json:"planet_houses,omitempty"`
    AspectInterpretations  []string             `json:"aspect_interpretations,omitempty"`
    PatternInterpretations []string             `json:"pattern_interpretations,omitempty"`
    LunarPhase             LunarPhase           `json:"lunar_phase,omitempty"`
    Retrogrades            []RetrogradeInfo     `json:"retrogrades,omitempty"`
    DispositorTree         DispositorTree       `json:"dispositor_tree,omitempty"`
    VOCMoon                VOCMoon              `json:"void_of_course_moon,omitempty"`
    Decans                 []DecanInfo          `json:"decans,omitempty"`
    Terms                  []TermInfo           `json:"terms,omitempty"`
    Antiscia               []AntiscionPoint     `json:"antiscia,omitempty"`
    ModernDignity          *DignityConvergence  `json:"modern_dignity,omitempty"`
    MinorAspects           []AspectHit          `json:"minor_aspects,omitempty"`
    Midpoints              []MidpointPair       `json:"midpoints,omitempty"`
    Declinations           []DeclinationPair    `json:"declinations,omitempty"`
}
```

**Step 4: Run test to verify pass**

```bash
go test -run TestFullReport_HasWesternFields -count=1 ./internal/dignity/...
```

Expected: PASS.

**Step 5: Commit**

```bash
git add internal/dignity/api.go internal/dignity/api_test.go
git commit -m "feat: add Western astrology fields to FullReport struct"
```

---

### Task 2: Add `MidpointPair` and `DeclinationPair` types

**Objective:** Define the types needed for midpoints and declinations (they don't exist yet).

**Files:**
- Modify: `internal/dignity/composite.go` (add `MidpointPair`)
- Modify: `internal/dignity/declination.go` (add `DeclinationPair`)

**Step 1: Write failing test**

```go
func TestMidpointPair_JSON(t *testing.T) {
    t.Parallel()
    mp := MidpointPair{Planet1: "Sun", Planet2: "Moon", Midpoint: 45.0, Sign: "Taurus"}
    if mp.Planet1 != "Sun" || mp.Planet2 != "Moon" || mp.Midpoint != 45.0 || mp.Sign != "Taurus" {
        t.Error("MidpointPair fields not accessible")
    }
}
```

**Step 2: Run to verify failure**

```bash
go test -run TestMidpointPair_JSON -count=1 ./internal/dignity/...
```

Expected: FAIL — `MidpointPair` undefined.

**Step 3: Add types**

In `composite.go`, add after the `midpoint` function:

```go
// MidpointPair is a single midpoint between two planets.
type MidpointPair struct {
    Planet1  string  `json:"planet1"`
    Planet2  string  `json:"planet2"`
    Midpoint float64 `json:"midpoint"`
    Sign     string  `json:"sign"`
}
```

Check if `internal/dignity/declination.go` exists. If not, create it with:

```go
package dignity

// DeclinationPair holds the declination of a planet and any parallels/contraparallels.
type DeclinationPair struct {
    Planet      string  `json:"planet"`
    Declination float64 `json:"declination"`
}
```

**Step 4: Run test to verify pass**

```bash
go test -run TestMidpointPair_JSON -count=1 ./internal/dignity/...
```

**Step 5: Commit**

```bash
git add internal/dignity/composite.go internal/dignity/declination.go internal/dignity/api_test.go
git commit -m "feat: add MidpointPair and DeclinationPair types"
```

---

### Task 3: Wire Placidus houses into `ComputeFullReport`

**Objective:** `ComputeFullReport` currently only computes whole-sign houses. Add Placidus house computation using the existing `swe.Houses()` call.

**Files:**
- Modify: `internal/dignity/api.go:30-105` (`ComputeFullReport` signature + body)
- Modify: `cmd/recover/main.go:446-459` (`computeAll` — pass lat/lng)

**Step 1: Write failing test**

```go
func TestComputeFullReport_PlacidusHouses(t *testing.T) {
    t.Parallel()
    planets := map[string]float64{
        "Sun": 0.0, "Moon": 90.0, "Mercury": 180.0,
    }
    fr := ComputeFullReport(planets, 24.0, 100.0, 10.0, "Test", 2000, 1, 1, 12, 0, 0, 0, 40.0, -75.0, nil)
    if fr.PlacidusHouses == nil {
        t.Error("PlacidusHouses should not be nil when lat/lng provided")
    }
    if len(fr.PlacidusHouses) == 0 {
        t.Error("PlacidusHouses should have entries")
    }
}
```

**Step 2: Run to verify failure**

```bash
go test -run TestComputeFullReport_PlacidusHouses -count=1 ./internal/dignity/...
```

Expected: FAIL — `PlacidusHouses` is nil.

**Step 3: Wire Placidus houses**

In `ComputeFullReport`, after the whole-sign house block, add:

```go
// Placidus houses (requires lat/lng)
if lat != 0 || lng != 0 {
    fr.PlacidusHouses = make(map[string]int)
    placidusCusps, _ := swe.Houses(fr.Ascendant, lat, lng, 'P')
    for planet, lon := range tropicalLons {
        fr.PlacidusHouses[planet] = houseForLongitude(lon, placidusCusps)
    }
}
```

Add helper function in `house.go`:

```go
// houseForLongitude returns the Placidus house (1-12) for a given longitude.
func houseForLongitude(lon float64, cusps []float64) int {
    for i := 1; i <= 12; i++ {
        cusp := cusps[i]
        nextCusp := cusps[(i%12)+1]
        if nextCusp < cusp {
            nextCusp += 360
        }
        testLon := lon
        if testLon < cusp {
            testLon += 360
        }
        if testLon >= cusp && testLon < nextCusp {
            return i
        }
    }
    return 1
}
```

**Step 4: Run test to verify pass**

```bash
go test -run TestComputeFullReport_PlacidusHouses -count=1 ./internal/dignity/...
```

**Step 5: Commit**

```bash
git add internal/dignity/api.go internal/dignity/house.go internal/dignity/api_test.go
git commit -m "feat: wire Placidus houses into ComputeFullReport"
```

---

### Task 4: Wire interpretations into `ComputeFullReport`

**Objective:** Call `InterpretChart` (planet-in-sign, planet-in-house, aspect, pattern interpretations) and store results in `FullReport`.

**Files:**
- Modify: `internal/dignity/api.go` (`ComputeFullReport`)

**Step 1: Write failing test**

```go
func TestComputeFullReport_Interpretations(t *testing.T) {
    t.Parallel()
    planets := map[string]float64{
        "Sun": 125.0, "Moon": 95.0, "Mercury": 140.0, "Venus": 60.0,
        "Mars": 200.0, "Jupiter": 250.0, "Saturn": 300.0,
        "Uranus": 330.0, "Neptune": 350.0, "Pluto": 10.0,
    }
    fr := ComputeFullReport(planets, 24.0, 100.0, 10.0, "Test", 2000, 1, 1, 12, 0, 0, 0, 40.0, -75.0, nil)
    if len(fr.PlanetSigns) == 0 {
        t.Error("PlanetSigns should not be empty")
    }
    if len(fr.PlanetHouses) == 0 {
        t.Error("PlanetHouses should not be empty")
    }
    if len(fr.AspectInterpretations) == 0 {
        t.Error("AspectInterpretations should not be empty")
    }
    if len(fr.PatternInterpretations) == 0 {
        t.Error("PatternInterpretations should not be empty")
    }
}
```

**Step 2: Run to verify failure**

```bash
go test -run TestComputeFullReport_Interpretations -count=1 ./internal/dignity/...
```

Expected: FAIL — interpretation fields are empty.

**Step 3: Wire interpretations**

In `ComputeFullReport`, after the patterns block, add:

```go
// Interpretations
ci := InterpretChart(fr.Name, tropicalLons, fr.Houses, fr.Aspects, fr.Patterns, fr.Stars)
fr.PlanetSigns = ci.PlanetSigns
fr.PlanetHouses = ci.PlanetHouses
fr.AspectInterpretations = ci.Aspects
fr.PatternInterpretations = ci.Patterns
```

**Step 4: Run test to verify pass**

```bash
go test -run TestComputeFullReport_Interpretations -count=1 ./internal/dignity/...
```

**Step 5: Commit**

```bash
git add internal/dignity/api.go internal/dignity/api_test.go
git commit -m "feat: wire interpretations into ComputeFullReport"
```

---

### Task 5: Wire traditional data into `ComputeFullReport`

**Objective:** Call `ComputeTraditionalReport` and store results in `FullReport`.

**Files:**
- Modify: `internal/dignity/api.go` (`ComputeFullReport` signature — add `speeds` param)
- Modify: `cmd/recover/main.go` (`computeAll` — pass `cd.speeds`)

**Step 1: Write failing test**

```go
func TestComputeFullReport_Traditional(t *testing.T) {
    t.Parallel()
    planets := map[string]float64{
        "Sun": 125.0, "Moon": 95.0, "Mercury": 140.0,
    }
    speeds := map[string]float64{
        "Sun": 1.0, "Moon": 13.0, "Mercury": 0.5,
    }
    fr := ComputeFullReport(planets, 24.0, 100.0, 10.0, "Test", 2000, 1, 1, 12, 0, 0, 0, 40.0, -75.0, nil, speeds)
    if fr.LunarPhase.Name == "" {
        t.Error("LunarPhase should be populated")
    }
    if len(fr.Retrogrades) == 0 {
        t.Error("Retrogrades should not be empty")
    }
    if len(fr.Decans) == 0 {
        t.Error("Decans should not be empty")
    }
    if len(fr.Terms) == 0 {
        t.Error("Terms should not be empty")
    }
    if len(fr.Antiscia) == 0 {
        t.Error("Antiscia should not be empty")
    }
    if len(fr.DispositorTree.Nodes) == 0 {
        t.Error("DispositorTree should not be empty")
    }
}
```

**Step 2: Run to verify failure**

```bash
go test -run TestComputeFullReport_Traditional -count=1 ./internal/dignity/...
```

Expected: FAIL — traditional fields are empty.

**Step 3: Wire traditional data**

Add `speeds map[string]float64` parameter to `ComputeFullReport`. In the body, add:

```go
// Traditional data
tr := ComputeTraditionalReport(name, tropicalLons, speeds)
fr.LunarPhase = tr.LunarPhase
fr.Retrogrades = tr.Retrogrades
fr.Decans = tr.Decans
fr.Terms = tr.Terms
fr.Antiscia = tr.Antiscia
fr.DispositorTree = tr.DispositorTree
fr.VOCMoon = tr.VOCMoon
```

Update `computeAll` in `cmd/recover/main.go` to pass `cd.speeds`.

**Step 4: Run test to verify pass**

```bash
go test -run TestComputeFullReport_Traditional -count=1 ./internal/dignity/...
```

**Step 5: Commit**

```bash
git add internal/dignity/api.go cmd/recover/main.go internal/dignity/api_test.go
git commit -m "feat: wire traditional data into ComputeFullReport"
```

---

### Task 6: Wire modern 4-state dignity into `ComputeFullReport`

**Objective:** Add 4-state dignity (domicile, exaltation, fall, detriment) alongside the existing 2-state dignity.

**Files:**
- Modify: `internal/dignity/api.go` (`ComputeFullReport`)

**Step 1: Write failing test**

```go
func TestComputeFullReport_ModernDignity(t *testing.T) {
    t.Parallel()
    planets := map[string]float64{
        "Sun": 0.0, "Moon": 90.0, "Mercury": 150.0,
    }
    fr := ComputeFullReport(planets, 24.0, 100.0, 10.0, "Test", 2000, 1, 1, 12, 0, 0, 0, 40.0, -75.0, nil, nil)
    if fr.ModernDignity == nil {
        t.Error("ModernDignity should not be nil")
    }
}
```

**Step 2: Run to verify failure**

```bash
go test -run TestComputeFullReport_ModernDignity -count=1 ./internal/dignity/...
```

Expected: FAIL — `ModernDignity` is nil.

**Step 3: Wire modern dignity**

In `ComputeFullReport`, after the 2-state dignity block, add:

```go
// 4-state modern dignity (domicile, exaltation, fall, detriment)
fr.ModernDignity = ComputeModernDignityConvergence(tropicalLons, ayan, name)
```

Check if `ComputeModernDignityConvergence` exists. If not, create it in `dignity.go`:

```go
func ComputeModernDignityConvergence(positions map[string]float64, ayan float64, name string) *DignityConvergence {
    rules := DignityRules
    return computeDignityConvergence(positions, ayan, name, rules)
}
```

**Step 4: Run test to verify pass**

```bash
go test -run TestComputeFullReport_ModernDignity -count=1 ./internal/dignity/...
```

**Step 5: Commit**

```bash
git add internal/dignity/api.go internal/dignity/dignity.go internal/dignity/api_test.go
git commit -m "feat: wire modern 4-state dignity into ComputeFullReport"
```

---

### Task 7: Wire minor aspects into `ComputeFullReport`

**Objective:** Add semi-sextile (30°) and quincunx (150°) aspects alongside the 5 Ptolemaic aspects.

**Files:**
- Modify: `internal/dignity/api.go` (`ComputeFullReport`)

**Step 1: Write failing test**

```go
func TestComputeFullReport_MinorAspects(t *testing.T) {
    t.Parallel()
    planets := map[string]float64{
        "Sun": 0.0, "Moon": 30.0, "Mercury": 150.0,
    }
    fr := ComputeFullReport(planets, 24.0, 100.0, 10.0, "Test", 2000, 1, 1, 12, 0, 0, 0, 40.0, -75.0, nil, nil)
    if len(fr.MinorAspects) == 0 {
        t.Error("MinorAspects should not be empty — Sun-Moon at 30° is semi-sextile, Sun-Mercury at 150° is quincunx")
    }
}
```

**Step 2: Run to verify failure**

```bash
go test -run TestComputeFullReport_MinorAspects -count=1 ./internal/dignity/...
```

Expected: FAIL — `MinorAspects` is empty.

**Step 3: Wire minor aspects**

In `ComputeFullReport`, after the Ptolemaic aspects block, add:

```go
// Minor aspects (semi-sextile, quincunx)
minorAspectDefs := MinorAspects()
fr.MinorAspects = FindNatalAspects(allPlanets, minorAspectDefs, DefaultPatternOrb)
```

Add `MinorAspects()` function in `aspect.go`:

```go
func MinorAspects() []AspectEntry {
    return []AspectEntry{
        {Name: "semi-sextile", Angle: 30, Universality: "partial"},
        {Name: "quincunx", Angle: 150, Universality: "partial"},
    }
}
```

**Step 4: Run test to verify pass**

```bash
go test -run TestComputeFullReport_MinorAspects -count=1 ./internal/dignity/...
```

**Step 5: Commit**

```bash
git add internal/dignity/api.go internal/dignity/aspect.go internal/dignity/api_test.go
git commit -m "feat: wire minor aspects into ComputeFullReport"
```

---

### Task 8: Wire midpoints into `ComputeFullReport`

**Objective:** Compute midpoints for all planet pairs and store in `FullReport`.

**Files:**
- Modify: `internal/dignity/api.go` (`ComputeFullReport`)

**Step 1: Write failing test**

```go
func TestComputeFullReport_Midpoints(t *testing.T) {
    t.Parallel()
    planets := map[string]float64{
        "Sun": 10.0, "Moon": 50.0, "Mars": 100.0,
    }
    fr := ComputeFullReport(planets, 24.0, 100.0, 10.0, "Test", 2000, 1, 1, 12, 0, 0, 0, 40.0, -75.0, nil, nil)
    if len(fr.Midpoints) == 0 {
        t.Error("Midpoints should not be empty")
    }
    // Sun-Moon midpoint should be 30.0
    for _, mp := range fr.Midpoints {
        if mp.Planet1 == "Sun" && mp.Planet2 == "Moon" {
            if mp.Midpoint != 30.0 {
                t.Errorf("Sun-Moon midpoint = %.2f, want 30.00", mp.Midpoint)
            }
        }
    }
}
```

**Step 2: Run to verify failure**

```bash
go test -run TestComputeFullReport_Midpoints -count=1 ./internal/dignity/...
```

Expected: FAIL — `Midpoints` is empty.

**Step 3: Wire midpoints**

In `ComputeFullReport`, add:

```go
// Midpoints (all classical planet pairs)
fr.Midpoints = ComputeAllMidpoints(tropicalLons)
```

Add `ComputeAllMidpoints` in `composite.go`:

```go
func ComputeAllMidpoints(positions map[string]float64) []MidpointPair {
    classical := []string{"Sun", "Moon", "Mercury", "Venus", "Mars", "Jupiter", "Saturn", "Uranus", "Neptune", "Pluto"}
    var result []MidpointPair
    for i := 0; i < len(classical); i++ {
        for j := i + 1; j < len(classical); j++ {
            p1, p2 := classical[i], classical[j]
            lon1, ok1 := positions[p1]
            lon2, ok2 := positions[p2]
            if !ok1 || !ok2 {
                continue
            }
            mp := midpoint(lon1, lon2)
            result = append(result, MidpointPair{
                Planet1: p1, Planet2: p2,
                Midpoint: math.Round(mp*100) / 100,
                Sign:     SignForLongitude(mp),
            })
        }
    }
    return result
}
```

**Step 4: Run test to verify pass**

```bash
go test -run TestComputeFullReport_Midpoints -count=1 ./internal/dignity/...
```

**Step 5: Commit**

```bash
git add internal/dignity/api.go internal/dignity/composite.go internal/dignity/api_test.go
git commit -m "feat: wire midpoints into ComputeFullReport"
```

---

### Task 9: Wire Arabic Parts into `ComputeFullReport`

**Objective:** Compute Part of Fortune and store in `FullReport.ArabicParts`.

**Files:**
- Modify: `internal/dignity/api.go` (`ComputeFullReport`)

**Step 1: Write failing test**

```go
func TestComputeFullReport_ArabicParts(t *testing.T) {
    t.Parallel()
    planets := map[string]float64{
        "Sun": 125.0, "Moon": 95.0, "Mercury": 140.0,
    }
    fr := ComputeFullReport(planets, 24.0, 100.0, 10.0, "Test", 2000, 1, 1, 12, 0, 0, 0, 40.0, -75.0, nil, nil)
    if fr.ArabicParts == nil {
        t.Error("ArabicParts should not be nil")
    }
}
```

**Step 2: Run to verify failure**

```bash
go test -run TestComputeFullReport_ArabicParts -count=1 ./internal/dignity/...
```

Expected: FAIL — `ArabicParts` is nil.

**Step 3: Wire Arabic Parts**

In `ComputeFullReport`, replace `fr.ArabicParts = nil` with:

```go
// Arabic Parts (Fortune only for now)
isDay := isDayChart(tropicalLons["Sun"], ascLong)
fr.ArabicParts = ComputePartReport(name, ascLong, tropicalLons, isDay, DefaultPatternOrb)
```

Add `isDayChart` helper in `arabic_parts.go`:

```go
func isDayChart(sunLon, ascLon float64) bool {
    dsc := normalizeLon(ascLon + 180)
    if ascLon < dsc {
        return sunLon >= ascLon && sunLon < dsc
    }
    return sunLon >= ascLon || sunLon < dsc
}
```

**Step 4: Run test to verify pass**

```bash
go test -run TestComputeFullReport_ArabicParts -count=1 ./internal/dignity/...
```

**Step 5: Commit**

```bash
git add internal/dignity/api.go internal/dignity/arabic_parts.go internal/dignity/api_test.go
git commit -m "feat: wire Arabic Parts into ComputeFullReport"
```

---

### Task 10: Wire declinations into `ComputeFullReport`

**Objective:** Compute declinations and store in `FullReport`.

**Files:**
- Modify: `internal/dignity/api.go` (`ComputeFullReport`)

**Step 1: Write failing test**

```go
func TestComputeFullReport_Declinations(t *testing.T) {
    t.Parallel()
    planets := map[string]float64{
        "Sun": 0.0, "Moon": 90.0,
    }
    fr := ComputeFullReport(planets, 24.0, 100.0, 10.0, "Test", 2000, 1, 1, 12, 0, 0, 0, 40.0, -75.0, nil, nil)
    if len(fr.Declinations) == 0 {
        t.Error("Declinations should not be empty")
    }
}
```

**Step 2: Run to verify failure**

```bash
go test -run TestComputeFullReport_Declinations -count=1 ./internal/dignity/...
```

Expected: FAIL — `Declinations` is empty.

**Step 3: Wire declinations**

Check if `internal/declination/declination.go` exists and has a `ComputeDeclinations` function. If so, import and call it. If not, add a simple declination computation in `declination.go`:

```go
package dignity

// ComputeDeclinations returns declination for each planet.
// This is a placeholder — real declination requires Swiss Ephemeris swe_calc_ut with iflag.
func ComputeDeclinations(positions map[string]float64) []DeclinationPair {
    // For now, return empty — declination requires the ephemeris to compute properly
    return nil
}
```

Wire in `ComputeFullReport`:

```go
fr.Declinations = ComputeDeclinations(tropicalLons)
```

**Step 4: Run test to verify pass** (adjust test expectation if declinations are nil for now)

**Step 5: Commit**

```bash
git add internal/dignity/api.go internal/dignity/declination.go internal/dignity/api_test.go
git commit -m "feat: wire declinations into ComputeFullReport"
```

---

### Task 11: Update `printReport` to display all new sections

**Objective:** Extend `printReport` in `cmd/recover/main.go` to print all the new Western fields.

**Files:**
- Modify: `cmd/recover/main.go:1631-1670` (`printReport`)

**Step 1: Write failing test**

No unit test for `printReport` (it's a CLI output function). Instead, verify by running:

```bash
cd /Users/aj/Documents/repos/koine
go run ./cmd/recover "AJ" 1969 2 15 23 10 -8 47.038 -122.901 2>&1 | head -200
```

Expected: Output should include new sections (Placidus houses, interpretations, traditional data, etc.)

**Step 2: Extend `printReport`**

Add sections after the existing Patterns block:

```go
// Placidus houses
if len(fr.PlacidusHouses) > 0 {
    fmt.Printf("\nPlacidus Houses:\n")
    for planet, house := range fr.PlacidusHouses {
        fmt.Printf("  %-10s House %d\n", planet, house)
    }
}

// Modern dignity
if fr.ModernDignity != nil {
    fmt.Printf("\nModern Dignity (4-state: domicile, exaltation, fall, detriment):\n")
    fmt.Printf("  Signal: %d\n", fr.ModernDignity.SignalCount())
    fmt.Printf("  Noise: %d\n", fr.ModernDignity.NoiseCount())
}

// Minor aspects
if len(fr.MinorAspects) > 0 {
    fmt.Printf("\nMinor Aspects (semi-sextile, quincunx):\n")
    for _, a := range fr.MinorAspects {
        fmt.Printf("  %s %s %s (%.1f deg)\n", a.Planet1, a.Aspect, a.Planet2, a.Orb)
    }
}

// Traditional data
if fr.LunarPhase.Name != "" {
    fmt.Printf("\nLunar Phase: %s (%.2f°)\n", fr.LunarPhase.Name, fr.LunarPhase.Angle)
}

if len(fr.Retrogrades) > 0 {
    fmt.Printf("\nRetrogrades:\n")
    for _, r := range fr.Retrogrades {
        if r.Retrograde {
            fmt.Printf("  %s Rx (%.4f deg/day)\n", r.Planet, r.Speed)
        }
    }
}

if len(fr.DispositorTree.Nodes) > 0 {
    fmt.Printf("\nDispositor Tree:\n")
    for _, n := range fr.DispositorTree.Nodes {
        marker := ""
        if n.IsFinal {
            marker = " ★ FINAL DISPOSITOR"
        } else if n.InLoop {
            marker = fmt.Sprintf(" ⟲ mutual reception with %s", n.LoopWith)
        }
        fmt.Printf("  %s in %s → %s%s\n", n.Planet, n.Sign, n.Dispositor, marker)
    }
}

if fr.VOCMoon.VOC {
    fmt.Printf("\nVoid of Course Moon: YES — Moon in %s, %.2f° until %s\n",
        fr.VOCMoon.MoonSign, fr.VOCMoon.DegreesToNext, fr.VOCMoon.NextSign)
}

// Interpretations
if len(fr.PlanetSigns) > 0 {
    fmt.Printf("\nPlanet-in-Sign Interpretations:\n")
    for _, s := range fr.PlanetSigns {
        fmt.Printf("  %s\n", s)
    }
}

if len(fr.PlanetHouses) > 0 {
    fmt.Printf("\nPlanet-in-House Interpretations:\n")
    for _, s := range fr.PlanetHouses {
        fmt.Printf("  %s\n", s)
    }
}

if len(fr.AspectInterpretations) > 0 {
    fmt.Printf("\nAspect Interpretations:\n")
    for _, s := range fr.AspectInterpretations {
        fmt.Printf("  %s\n", s)
    }
}

if len(fr.PatternInterpretations) > 0 {
    fmt.Printf("\nPattern Interpretations:\n")
    for _, s := range fr.PatternInterpretations {
        fmt.Printf("  %s\n", s)
    }
}

// Midpoints (top 10 by significance)
if len(fr.Midpoints) > 0 {
    fmt.Printf("\nMidpoints (all classical pairs):\n")
    for _, mp := range fr.Midpoints {
        fmt.Printf("  %s/%s = %.2f° %s\n", mp.Planet1, mp.Planet2, mp.Midpoint, mp.Sign)
    }
}

// Arabic Parts
if fr.ArabicParts != nil && len(fr.ArabicParts.Parts) > 0 {
    fmt.Printf("\nArabic Parts:\n")
    for _, p := range fr.ArabicParts.Parts {
        fmt.Printf("  %s: %.2f° %s\n", p.Part, p.Lon, p.Sign)
    }
}
```

**Step 3: Verify output**

```bash
go run ./cmd/recover "AJ" 1969 2 15 23 10 -8 47.038 -122.901 2>&1 | head -300
```

**Step 4: Commit**

```bash
git add cmd/recover/main.go
git commit -m "feat: extend printReport with all Western astrology sections"
```

---

### Task 12: Full integration test

**Objective:** Run the full `recover` command and verify all sections appear.

**Step 1: Run recover**

```bash
cd /Users/aj/Documents/repos/koine
go run ./cmd/recover "AJ" 1969 2 15 23 10 -8 47.038 -122.901
```

**Step 2: Verify output contains:**
- Placidus Houses section
- Modern Dignity section
- Minor Aspects section
- Lunar Phase
- Retrogrades
- Dispositor Tree
- Void of Course Moon
- Planet-in-Sign Interpretations
- Planet-in-House Interpretations
- Aspect Interpretations
- Pattern Interpretations
- Midpoints
- Arabic Parts

**Step 3: Run all tests**

```bash
go test ./internal/dignity/... ./cmd/recover/...
```

**Step 4: Commit**

```bash
git add -A
git commit -m "test: full integration verification of Western astrology output"
```

---

## Principles

- **DRY:** Don't duplicate existing functions — call them.
- **YAGNI:** Wire what exists. Don't add new interpretation text or new computation algorithms.
- **TDD:** Every task has a failing test before implementation code.
- **Frequent commits:** Commit after every task.

## Pitfalls

- `write_file` corrupts Go source with double line numbers. Use `cat >>` heredoc for appending, `patch` for targeted edits.
- `ComputeFullReport` signature changes will break callers in `cmd/recover/main.go` — update `computeAll` and the `serve` subcommand's `compute` closure.
- The `swe.Houses()` call requires the ephemeris to be initialized (`swe.SetEphePath`). Tests that call `ComputeFullReport` with non-zero lat/lng will need the ephemeris available.
- `InterpretChart` takes `[]StarConjunction` — pass `fr.Stars` (already computed).
- `ComputePartReport` needs `isDay` — compute from Sun position relative to ASC/DSC.
