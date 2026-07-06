# BaseChart Refactor — Implementation Plan

> **For Hermes:** Use subagent-driven-development skill to implement this plan task-by-task.

**Goal:** Extract a system-agnostic `BaseChart` struct that computes all astrological positions once, then refactor every system endpoint to use `FromBase()` pure functions instead of re-computing positions independently.

**Architecture:** `BaseChart` lives in `internal/dignity/` (or a new `internal/chart/` package). It holds tropical positions, sidereal positions, house cusps (all systems), aspects, fixed stars, Arabic Parts, latitudes, speeds, and Julian Day — everything calculable from birth data. Each system (Koiné, Vedic, BaZi, Draconic, etc.) gets a `FromBase(b BaseChart) SystemChart` pure function that extracts only what that system needs. Interpretation is a separate pass on the system chart.

**Tech Stack:** Go 1.21+, CGo to Swiss Ephemeris C library, existing `internal/dignity/` package.

---

## Current State (GoF Audit)

### Architecture Smells

1. **No shared computation base** — `chartData` (in `cmd/recover/compute.go`) is a thin wrapper: `planets map[string]float64`, `speeds`, `ayan`, `asc`, `nn`, `jd`. It doesn't store latitudes, house cusps, fixed stars, or sidereal positions. Every compute function that needs those re-calls SWE.

2. **Re-computation everywhere** — `computeDeclination` re-calls `swe.CalcUT` for all 18 planets to get latitudes. `computeParans` re-calls `swe.Fixstar` for all 25 stars. `computeFirdaria` re-calls `swe.Houses`. `computeInterpretation` re-calls `swe.Houses`. `computeAstroCartography` re-calls `swe.CalcUT` for ASC line computation. Each of these is a separate SWE call chain.

3. **No-op map copies** — `for k, v := range cd.planets { tropical[k] = v }` in ~15 functions. `cd.planets` is already a map. These are vestigial from when `chartData` didn't exist and each function called SWE directly.

4. **Two diverging repos** — `koine` and `empirical` share ~90% of `internal/dignity/` code but have diverged. `koine` adds `synthesis.go`, `planet_in_sign.go`, `evolutionary/` but also has modified copies of `api.go`, `dignity.go`, `interpretation.go`, etc. The `BaseChart` should live in `empirical`; `koine` should import it.

5. **God file** — `cmd/recover/compute.go` is 1,715 lines with 33 `compute*` functions. After `BaseChart`, most collapse to thin wrappers: `computeX(name, cd, ...) → system.FromBase(base).ToJSON()`.

6. **Divergent Change** — Adding a new astrological system currently touches: (a) new file in `dignity/`, (b) function type in `server.go`, (c) field in `ServerConfig`, (d) handler in `NewMux`, (e) closure in `main.go`, (f) compute function in `compute.go`. After `BaseChart` + `ServerConfig` (already done), this drops to: (a) `FromBase()` function, (b) function type + config field, (c) handler registration. Three touch points, all mechanical.

### Good Patterns Already Present

- **ServerConfig struct** (P1 #1 from 2026-06-24 audit) — clean DI, 35 params collapsed to one struct
- **Generic `handleJSON[T]`** (P1 #2) — 30 duplicate handlers collapsed to one-liners
- **Named response types** (P2 #7) — every compute function returns a typed struct, not raw `[]byte`
- **`marshalResult[T]`** — generic JSON serialization helper
- **`dignity.AllPlanets`** (P1 #3) — single source of truth for planet ID list
- **Orb constants** (P2 #8) — `OrbTight`, `OrbNarrow`, `OrbStandard`, `OrbWide`

---

## Target Architecture

```
┌─────────────────────────────────────────────┐
│              BaseChart                       │
│  (computed once from birth data)             │
│                                              │
│  Tropical positions (all 24 bodies)          │
│  Sidereal positions (Lahiri)                │
│  Latitudes + speeds (all bodies)             │
│  House cusps (Placidus, Whole Sign, etc.)   │
│  ASC, MC, DSC, IC                           │
│  North Node, Ayanamsa                       │
│  Julian Day                                  │
│  Fixed star positions (25 stars)            │
│  Arabic Parts (13 parts)                    │
│  Natal aspects (all bodies, all aspects)    │
│  GMST                                        │
└──────────────┬──────────────────────────────┘
               │
    ┌──────────┼──────────┬──────────┬──────────┐
    ▼          ▼          ▼          ▼          ▼
┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐
│Koiné   │ │Vedic   │ │BaZi    │ │Draconic│ │Uranian │
│.From   │ │.From   │ │.From   │ │.From   │ │.From   │
│Base()  │ │Base()  │ │Base()  │ │Base()  │ │Base()  │
└───┬────┘ └───┬────┘ └───┬────┘ └───┬────┘ └───┬────┘
    ▼          ▼          ▼          ▼          ▼
┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐
│Interp  │ │Interp  │ │Interp  │ │Interp  │ │Interp  │
│(skills)│ │(skills)│ │(skills)│ │(skills)│ │(skills)│
└────────┘ └────────┘ └────────┘ └────────┘ └────────┘
```

**Key principle:** `BaseChart` is pure computation. `FromBase()` is pure extraction/filtering. Interpretation is a separate pass. No system re-derives positions.

---

## Tasks

### Phase 1: BaseChart struct + constructor

### Task 1: Define `BaseChart` struct

**Objective:** Create the `BaseChart` struct in `internal/dignity/base_chart.go` with all fields needed by every system.

**Files:**
- Create: `internal/dignity/base_chart.go`

**Step 1: Write the struct**

```go
package dignity

import "github.com/aj-nt/empirical/internal/swe"

// BaseChart holds all computed astrological positions for a single chart.
// Computed once from birth data; every system extracts what it needs via FromBase().
type BaseChart struct {
    // Identity
    Name string

    // Birth data
    Year, Month, Day, Hour, Minute, Second int
    TZOffset                               float64
    Lat, Lng                               float64

    // Core positions
    Tropical map[string]Position  // planet name → lon+lat+speed
    Sidereal map[string]Position  // same, shifted by ayanamsa
    Ayanamsa float64

    // Angles
    ASC, MC, DSC, IC float64

    // Nodes
    NorthNode  float64
    SouthNode  float64

    // Houses (all systems)
    Houses map[string][]float64 // "placidus" → [12]cusps, "whole_sign" → [12]cusps, etc.

    // Julian Day
    JD float64

    // Pre-computed derivatives
    Aspects     []AspectHit       // all natal aspects
    FixedStars []StarConjunction  // all star-planet conjunctions
    ArabicParts map[string]float64 // all 13 Arabic Parts
    GMST       float64
}

// Position holds a planet's longitude, latitude, and daily speed.
type Position struct {
    Lon   float64
    Lat   float64
    Speed float64
}
```

**Step 2: Verify it compiles**

```bash
cd ~/Documents/repos/empirical && go build ./internal/dignity/
```

Expected: compiles (struct definition only, no logic yet).

**Step 3: Commit**

```bash
git add internal/dignity/base_chart.go
git commit -m "feat: add BaseChart struct definition"
```

---

### Task 2: Write `ComputeBaseChart` constructor

**Objective:** Implement `func ComputeBaseChart(name string, year, month, day, hour, minute, second int, tzOff, lat, lng float64) (*BaseChart, error)` that computes everything once.

**Files:**
- Modify: `internal/dignity/base_chart.go`

**Step 1: Write the constructor**

```go
// ComputeBaseChart computes all astrological positions for a birth chart.
// This is the single entry point — every system extracts from the returned BaseChart.
func ComputeBaseChart(name string, year, month, day, hour, minute, second int, tzOff, lat, lng float64) (*BaseChart, error) {
    utHour := float64(hour) + float64(minute)/60.0 + float64(second)/3600.0 - tzOff
    jd := swe.Julday(year, month, day, utHour, true)
    ayan := swe.GetAyanamsaUT(jd)

    // Compute all planet positions (tropical + sidereal)
    tropical := make(map[string]Position)
    sidereal := make(map[string]Position)
    for _, p := range AllPlanets {
        lon, lat, _, speed := swe.CalcUT(jd, p.ID)
        tropical[p.Name] = Position{Lon: lon, Lat: lat, Speed: speed}
        sidLon := lon - ayan
        if sidLon < 0 {
            sidLon += 360
        }
        sidereal[p.Name] = Position{Lon: sidLon, Lat: lat, Speed: speed}
    }

    // Nodes
    nnLon, _, _, _ := swe.CalcUT(jd, swe.MEAN_NODE)
    snLon := nnLon + 180
    if snLon >= 360 {
        snLon -= 360
    }

    // Houses for all systems
    houses := make(map[string][]float64)
    for _, hs := range []struct{ name string; code int }{
        {"placidus", 'P'},
        {"whole_sign", 'W'},
        {"equal", 'E'},
        {"porphyry", 'O'},
        {"koch", 'K'},
    } {
        cusps, ascmc := swe.Houses(jd, lat, lng, hs.code)
        houseCusps := make([]float64, 13) // 1-indexed
        copy(houseCusps[1:], cusps[1:13])
        houses[hs.name] = houseCusps
        if hs.name == "placidus" {
            asc = ascmc[0]
            mc = ascmc[1]
        }
    }

    // Angles
    asc := houses["placidus"][1]
    mc := houses["placidus"][10]
    dsc := asc + 180
    if dsc >= 360 { dsc -= 360 }
    ic := mc + 180
    if ic >= 360 { ic -= 360 }

    // Fixed stars
    stars := ComputeStarConjunctions(tropical, jd, OrbNarrow)

    // Arabic Parts
    parts := ComputeAllArabicParts(tropical, asc)

    // Natal aspects
    aspects := FindNatalAspects(tropicalToLonMap(tropical), DefaultAspects(), OrbStandard)

    // GMST
    gmst := ComputeGMST(jd)

    return &BaseChart{
        Name:       name,
        Year:       year, Month: month, Day: day,
        Hour:       hour, Minute: minute, Second: second,
        TZOffset:   tzOff,
        Lat:        lat, Lng: lng,
        Tropical:   tropical,
        Sidereal:   sidereal,
        Ayanamsa:   ayan,
        ASC:        asc, MC: mc, DSC: dsc, IC: ic,
        NorthNode:  nnLon, SouthNode: snLon,
        Houses:     houses,
        JD:         jd,
        Aspects:    aspects,
        FixedStars: stars,
        ArabicParts: parts,
        GMST:       gmst,
    }, nil
}

// tropicalToLonMap extracts longitude-only map for functions that expect map[string]float64.
func tropicalToLonMap(tropical map[string]Position) map[string]float64 {
    m := make(map[string]float64, len(tropical))
    for k, v := range tropical {
        m[k] = v.Lon
    }
    return m
}
```

**Step 2: Verify it compiles**

```bash
cd ~/Documents/repos/empirical && go build ./internal/dignity/
```

Expected: compiles. May need to add helper functions (`ComputeStarConjunctions`, `ComputeAllArabicParts`, `tropicalToLonMap`).

**Step 3: Commit**

```bash
git add internal/dignity/base_chart.go
git commit -m "feat: add ComputeBaseChart constructor"
```

---

### Task 3: Write tests for `ComputeBaseChart`

**Objective:** Verify the constructor produces correct positions against known values.

**Files:**
- Create: `internal/dignity/base_chart_test.go`

**Step 1: Write the test**

```go
package dignity

import (
    "math"
    "testing"
)

func TestComputeBaseChart_KnownChart(t *testing.T) {
    // AJ's chart: 1969-02-15 23:10 -8:00 47.038 -122.901
    bc, err := ComputeBaseChart("AJ", 1969, 2, 15, 23, 10, 0, -8, 47.038, -122.901)
    if err != nil {
        t.Fatalf("ComputeBaseChart failed: %v", err)
    }

    // Basic sanity checks
    if bc.Name != "AJ" {
        t.Errorf("Name = %q, want %q", bc.Name, "AJ")
    }
    if len(bc.Tropical) < 12 {
        t.Errorf("got %d tropical positions, want at least 12", len(bc.Tropical))
    }
    if len(bc.Sidereal) != len(bc.Tropical) {
        t.Errorf("sidereal count %d != tropical count %d", len(bc.Sidereal), len(bc.Tropical))
    }

    // Sun should be in Aquarius (~326° tropical)
    sunTrop := bc.Tropical["Sun"].Lon
    if sunTrop < 300 || sunTrop > 330 {
        t.Errorf("Sun tropical = %.2f, expected ~326° (Aquarius)", sunTrop)
    }

    // Sun sidereal should be ~302° (Capricorn)
    sunSid := bc.Sidereal["Sun"].Lon
    if sunSid < 290 || sunSid > 320 {
        t.Errorf("Sun sidereal = %.2f, expected ~302° (Capricorn)", sunSid)
    }

    // Ayanamsa should be ~23.5° for 1969
    if math.Abs(bc.Ayanamsa-23.5) > 1.0 {
        t.Errorf("Ayanamsa = %.2f, expected ~23.5°", bc.Ayanamsa)
    }

    // ASC should be in Scorpio (~210-240°)
    if bc.ASC < 200 || bc.ASC > 250 {
        t.Errorf("ASC = %.2f, expected ~210-240° (Scorpio)", bc.ASC)
    }

    // Houses should have all 5 systems
    expectedSystems := []string{"placidus", "whole_sign", "equal", "porphyry", "koch"}
    for _, hs := range expectedSystems {
        if _, ok := bc.Houses[hs]; !ok {
            t.Errorf("missing house system: %s", hs)
        }
    }

    // Aspects should be non-empty
    if len(bc.Aspects) == 0 {
        t.Error("expected non-empty aspects")
    }

    // Fixed stars should be non-empty
    if len(bc.FixedStars) == 0 {
        t.Error("expected non-empty fixed stars")
    }

    // Arabic Parts should have all 13
    if len(bc.ArabicParts) < 13 {
        t.Errorf("got %d Arabic Parts, want at least 13", len(bc.ArabicParts))
    }
}

func TestComputeBaseChart_Angles(t *testing.T) {
    bc, err := ComputeBaseChart("Test", 2000, 1, 1, 12, 0, 0, 0, 51.5, -0.12)
    if err != nil {
        t.Fatalf("ComputeBaseChart failed: %v", err)
    }

    // DSC should be ASC + 180
    dscExpected := bc.ASC + 180
    if dscExpected >= 360 {
        dscExpected -= 360
    }
    if math.Abs(bc.DSC-dscExpected) > 0.01 {
        t.Errorf("DSC = %.2f, want %.2f (ASC+180)", bc.DSC, dscExpected)
    }

    // IC should be MC + 180
    icExpected := bc.MC + 180
    if icExpected >= 360 {
        icExpected -= 360
    }
    if math.Abs(bc.IC-icExpected) > 0.01 {
        t.Errorf("IC = %.2f, want %.2f (MC+180)", bc.IC, icExpected)
    }
}
```

**Step 2: Run tests**

```bash
cd ~/Documents/repos/empirical && go test ./internal/dignity/ -run TestComputeBaseChart -v -count=1
```

Expected: PASS.

**Step 3: Commit**

```bash
git add internal/dignity/base_chart_test.go
git commit -m "test: add BaseChart constructor tests"
```

---

### Phase 2: System FromBase() functions

### Task 4: Write `KoinéFromBase` function

**Objective:** Extract Koiné-specific data from `BaseChart` — whole-sign houses, 7 Hellenistic factors, triplicity rulers, terms, decans.

**Files:**
- Create: `internal/dignity/koine_from_base.go`

**Step 1: Write the function**

```go
package dignity

// KoinéChart holds the Koiné (Hellenistic) view of a chart.
type KoinéChart struct {
    Name             string
    WholeSignHouses  map[string]int          // planet → house number (1-12)
    ASC              float64
    MC               float64
    Dignities        map[string]KoinéDignity // planet → dignities
    TriplicityRulers map[string][]string      // planet → triplicity rulers
    Terms            map[string]TermInfo      // planet → term info
    Decans           map[string]DecanInfo     // planet → decan info
    ArabicParts      map[string]float64
    NorthNode        float64
    SouthNode        float64
    Sect             string // "diurnal" or "nocturnal"
}

type KoinéDignity struct {
    Domicile   bool
    Exaltation bool
    Triplicity bool
    Term       bool
    Decan      bool
    Peregrine  bool
}

type TermInfo struct {
    Ruler    string
    Degree   int
}

type DecanInfo struct {
    Ruler    string
    Number   int // 1, 2, or 3
}

// KoinéFromBase extracts the Koiné (Hellenistic) view from a BaseChart.
func KoinéFromBase(bc *BaseChart) *KoinéChart {
    kc := &KoinéChart{
        Name:            bc.Name,
        WholeSignHouses: make(map[string]int),
        Dignities:       make(map[string]KoinéDignity),
        TriplicityRulers: make(map[string][]string),
        Terms:           make(map[string]TermInfo),
        Decans:          make(map[string]DecanInfo),
        ArabicParts:     bc.ArabicParts,
        NorthNode:       bc.NorthNode,
        SouthNode:       bc.SouthNode,
        ASC:             bc.ASC,
        MC:              bc.MC,
    }

    // Whole-sign houses from ASC
    ascSign := int(bc.ASC / 30)
    for name, pos := range bc.Tropical {
        planetSign := int(pos.Lon / 30)
        house := ((planetSign - ascSign + 12) % 12) + 1
        kc.WholeSignHouses[name] = house
    }

    // Sect: Sun above horizon = diurnal
    sunLon := bc.Tropical["Sun"].Lon
    sunAbove := false
    diff := sunLon - bc.ASC
    if diff < 0 { diff += 360 }
    if diff < 180 { sunAbove = true }
    if sunAbove {
        kc.Sect = "diurnal"
    } else {
        kc.Sect = "nocturnal"
    }

    // Dignities for classical planets
    for _, planet := range ClassicalPlanets {
        pos, ok := bc.Tropical[planet]
        if !ok { continue }
        sign := SignForLongitude(pos.Lon)
        kd := KoinéDignity{}

        // Domicile
        if rules, ok := westernDignityTable[planet]; ok {
            for _, d := range rules.Domicile {
                if d == sign { kd.Domicile = true; break }
            }
            for _, d := range rules.Exaltation {
                if d == sign { kd.Exaltation = true; break }
            }
        }

        // Triplicity (Hellenistic: Dorothean triplicities)
        kd.Triplicity = isTriplicityRuler(planet, sign, kc.Sect)

        // Terms (Egyptian bounds)
        termDeg := pos.Lon - float64(int(pos.Lon/30)*30)
        termRuler := egyptianTermRuler(sign, int(termDeg))
        if termRuler == planet {
            kd.Term = true
        }
        kc.Terms[planet] = TermInfo{Ruler: termRuler, Degree: int(termDeg)}

        // Decans (Chaldean)
        decanNum := int(termDeg/10) + 1
        decanRuler := chaldeanDecanRuler(sign, decanNum)
        if decanRuler == planet {
            kd.Decan = true
        }
        kc.Decans[planet] = DecanInfo{Ruler: decanRuler, Number: decanNum}

        // Peregrine: no essential dignity at all
        kd.Peregrine = !kd.Domicile && !kd.Exaltation && !kd.Triplicity && !kd.Term && !kd.Decan

        kc.Dignities[planet] = kd
    }

    // Triplicity rulers
    for _, planet := range ClassicalPlanets {
        pos, ok := bc.Tropical[planet]
        if !ok { continue }
        sign := SignForLongitude(pos.Lon)
        kc.TriplicityRulers[planet] = triplicityRulersForSign(sign, kc.Sect)
    }

    return kc
}
```

**Step 2: Add helper functions**

Add to the same file: `isTriplicityRuler`, `egyptianTermRuler`, `chaldeanDecanRuler`, `triplicityRulersForSign`. These reference the existing dignity tables in `dignity.go` and `traditional.go`.

**Step 3: Verify it compiles**

```bash
cd ~/Documents/repos/empirical && go build ./internal/dignity/
```

**Step 4: Commit**

```bash
git add internal/dignity/koine_from_base.go
git commit -m "feat: add KoinéFromBase extraction function"
```

---

### Task 5: Write `VedicFromBase` function

**Objective:** Extract Vedic-specific data — sidereal signs, nakshatras, Vimshottari dasha, yogas.

**Files:**
- Create: `internal/dignity/vedic_from_base.go`

**Step 1: Write the function**

```go
package dignity

// VedicChart holds the Vedic (Jyotish) view of a chart.
type VedicChart struct {
    Name       string
    Planets    map[string]VedicPlanet
    ASC        float64
    Nakshatras map[string]NakshatraPlacement
    Dasha      []DashaPeriod
    Yogas      []Yoga
}

type VedicPlanet struct {
    Sign      string
    Degree    float64
    Nakshatra string
    Pada      int
    House     int
    Dignity   string // swakshetra, uchcha, neecha, neutral
}

// VedicFromBase extracts the Vedic (Jyotish) view from a BaseChart.
func VedicFromBase(bc *BaseChart) *VedicChart {
    vc := &VedicChart{
        Name:       bc.Name,
        Planets:    make(map[string]VedicPlanet),
        Nakshatras: make(map[string]NakshatraPlacement),
    }

    // Use sidereal positions
    for name, pos := range bc.Sidereal {
        sign := SignForLongitude(pos.Lon)
        deg := pos.Lon - float64(int(pos.Lon/30)*30)
        nak := GetNakshatra(pos.Lon)
        pada := GetNakshatraPada(pos.Lon)

        // House (whole-sign from sidereal ASC)
        ascSign := int(bc.Sidereal["Sun"].Lon/30) // approximate — use actual sidereal ASC
        planetSign := int(pos.Lon / 30)
        house := ((planetSign - ascSign + 12) % 12) + 1

        // Dignity
        dignity := "neutral"
        if rules, ok := vedicDignityTable[name]; ok {
            for _, s := range rules.Swakshetra {
                if s == sign { dignity = "swakshetra"; break }
            }
            for _, s := range rules.Uchcha {
                if s == sign { dignity = "uchcha"; break }
            }
            for _, s := range rules.Neecha {
                if s == sign { dignity = "neecha"; break }
            }
        }

        vc.Planets[name] = VedicPlanet{
            Sign:      sign,
            Degree:    deg,
            Nakshatra: nak.Name,
            Pada:      pada,
            House:     house,
            Dignity:   dignity,
        }
        vc.Nakshatras[name] = nak
    }

    // Vimshottari Dasha
    moonNak := vc.Nakshatras["Moon"]
    vc.Dasha = ComputeVimshottariDasha(moonNak, bc.Year, bc.Month, bc.Day)

    // Yogas
    vc.Yogas = ComputeYogas(bc.Sidereal)

    return vc
}
```

**Step 2: Verify it compiles**

```bash
cd ~/Documents/repos/empirical && go build ./internal/dignity/
```

**Step 3: Commit**

```bash
git add internal/dignity/vedic_from_base.go
git commit -m "feat: add VedicFromBase extraction function"
```

---

### Task 6: Write `BaZiFromBase` function

**Objective:** Extract BaZi-specific data — Four Pillars, Day Master, Ten Gods, hidden stems.

**Files:**
- Create: `internal/dignity/bazi_from_base.go`

**Step 1: Write the function**

```go
package dignity

// BaZiChart holds the BaZi (Four Pillars) view of a chart.
type BaZiChart struct {
    Name       string
    Pillars    BaZiPillars
    DayMaster  string
    TenGods    map[string]string // pillar → ten god
    HiddenStems map[string][]string
    LuckPillars []LuckPillar
}

// BaZiFromBase extracts the BaZi view from a BaseChart.
func BaZiFromBase(bc *BaseChart) *BaZiChart {
    pillars := ComputeBaZiPillars(bc.Year, bc.Month, bc.Day, bc.Hour)
    dayMaster := pillars.Day.Stem
    tenGods := ComputeTenGods(pillars, dayMaster)
    hidden := ComputeHiddenStems(pillars)
    luck := ComputeLuckPillars(pillars, bc.Year, bc.Month, bc.Day, bc.Hour)

    return &BaZiChart{
        Name:        bc.Name,
        Pillars:     pillars,
        DayMaster:   dayMaster,
        TenGods:     tenGods,
        HiddenStems: hidden,
        LuckPillars: luck,
    }
}
```

**Step 2: Verify it compiles**

```bash
cd ~/Documents/repos/empirical && go build ./internal/dignity/
```

**Step 3: Commit**

```bash
git add internal/dignity/bazi_from_base.go
git commit -m "feat: add BaZiFromBase extraction function"
```

---

### Task 7: Write `DraconicFromBase` function

**Objective:** Extract Draconic-specific data — draconic positions, sign shifts, bridges.

**Files:**
- Create: `internal/dignity/draconic_from_base.go`

**Step 1: Write the function**

```go
package dignity

// DraconicChart holds the draconic view of a chart.
type DraconicChart struct {
    Name    string
    Offset  float64
    Planets map[string]float64 // draconic longitudes
    Shifts  []DraconicShift
    Bridges []DraconicBridge
}

// DraconicFromBase extracts the draconic view from a BaseChart.
func DraconicFromBase(bc *BaseChart, orbDeg float64) *DraconicChart {
    drac := ComputeDraconic(tropicalToLonMap(bc.Tropical), bc.NorthNode)
    shifts := ComputeDraconicSignShifts(tropicalToLonMap(bc.Tropical), bc.NorthNode)
    bridges := ComputeDraconicBridges(
        tropicalToLonMap(bc.Tropical), bc.NorthNode,
        NonTNPNoNodePlanetNames, DefaultAspects(), orbDeg,
    )

    return &DraconicChart{
        Name:    bc.Name,
        Offset:  drac.Offset,
        Planets: drac.Planets,
        Shifts:  shifts,
        Bridges: bridges,
    }
}
```

**Step 2: Verify it compiles**

```bash
cd ~/Documents/repos/empirical && go build ./internal/dignity/
```

**Step 3: Commit**

```bash
git add internal/dignity/draconic_from_base.go
git commit -m "feat: add DraconicFromBase extraction function"
```

---

### Phase 3: Refactor compute.go to use BaseChart

### Task 8: Replace `chartData` with `BaseChart` in compute.go

**Objective:** Replace the `chartData` struct and `computePositions()` with `BaseChart` and `ComputeBaseChart()`. Update all 33 compute functions to accept `*BaseChart` instead of `*chartData`.

**Files:**
- Modify: `cmd/recover/compute.go`
- Modify: `cmd/recover/main.go`

**Step 1: Replace `chartData` usage**

The pattern change for every compute function:

```go
// BEFORE:
func computeTransits(name string, ..., cacheDir string) (*TransitsResponse, error) {
    cd := computePositions(year, month, day, hour, minute, tzOff, lat, lng, cacheDir)
    // ... use cd.planets, cd.ayan, cd.asc, cd.nn, cd.jd
}

// AFTER:
func computeTransits(name string, ..., cacheDir string) (*TransitsResponse, error) {
    bc, err := dignity.ComputeBaseChart(name, year, month, day, hour, minute, 0, tzOff, lat, lng)
    if err != nil {
        return nil, err
    }
    // ... use bc.Tropical, bc.Ayanamsa, bc.ASC, bc.NorthNode, bc.JD
}
```

**Step 2: Remove no-op map copies**

Every `for k, v := range cd.planets { tropical[k] = v }` becomes unnecessary — `bc.Tropical` is already the source of truth. Use `tropicalToLonMap(bc.Tropical)` where a `map[string]float64` is needed.

**Step 3: Remove re-computation**

- `computeDeclination`: use `bc.Tropical[planet].Lat` instead of re-calling `swe.CalcUT`
- `computeParans`: use `bc.FixedStars` instead of re-calling `swe.Fixstar`
- `computeFirdaria`: use `bc.ASC` and `bc.Tropical["Sun"].Lon` instead of re-calling `swe.Houses`
- `computeInterpretation`: use `bc.Houses` and `bc.Aspects` instead of re-computing

**Step 4: Update main.go closures**

The closures in `main.go` that call compute functions need to pass `*BaseChart` instead of `*chartData`. The `compute` closure (for `/api/recover`) changes from:

```go
compute := func(name string, ...) ([]byte, error) {
    result := computeAll(name, ...)
    return result.FullReportJSON()
}
```

To:

```go
compute := func(name string, ...) ([]byte, error) {
    bc, err := dignity.ComputeBaseChart(name, year, month, day, hour, minute, 0, tzOff, lat, lng)
    if err != nil {
        return nil, err
    }
    kc := dignity.KoinéFromBase(bc)
    return json.Marshal(kc)
}
```

**Step 5: Run all tests**

```bash
cd ~/Documents/repos/empirical && go test ./... -count=1 -race
```

Expected: all green.

**Step 6: Commit**

```bash
git add cmd/recover/compute.go cmd/recover/main.go
git commit -m "refactor: replace chartData with BaseChart in compute.go"
```

---

### Task 9: Remove `chartData` struct and `computePositions()`

**Objective:** Delete the now-unused `chartData` struct and `computePositions()` function.

**Files:**
- Modify: `cmd/recover/compute.go`

**Step 1: Delete**

Remove the `chartData` struct definition and `computePositions()` function. Verify no references remain:

```bash
cd ~/Documents/repos/empirical && grep -rn "chartData\|computePositions" --include='*.go'
```

Expected: no output.

**Step 2: Run tests**

```bash
cd ~/Documents/repos/empirical && go test ./... -count=1 -race
```

**Step 3: Commit**

```bash
git add cmd/recover/compute.go
git commit -m "refactor: remove chartData struct and computePositions"
```

---

### Phase 4: Unify koine and empirical

### Task 10: Make koine import empirical's BaseChart

**Objective:** Replace koine's duplicated `internal/dignity/` with an import of `github.com/aj-nt/empirical`. Keep only koine-specific files (`synthesis.go`, `planet_in_sign.go`, `evolutionary/`).

**Files:**
- Modify: `~/Documents/repos/koine/go.mod` (add `require github.com/aj-nt/empirical`)
- Modify: `~/Documents/repos/koine/cmd/recover/main.go` (import empirical)
- Delete: `~/Documents/repos/koine/internal/dignity/*.go` (except koine-specific)
- Keep: `synthesis.go`, `planet_in_sign.go`, `interpretation.go` (koine-specific)

**Step 1: Add dependency**

```bash
cd ~/Documents/repos/koine
go mod edit -require github.com/aj-nt/empirical@latest
go mod tidy
```

**Step 2: Update imports**

Replace `github.com/aj-nt/koine/internal/dignity` with `github.com/aj-nt/empirical/internal/dignity` in all koine files.

**Step 3: Remove duplicated files**

Delete all files in `koine/internal/dignity/` that are identical to or superseded by `empirical/internal/dignity/`. Keep only `synthesis.go`, `planet_in_sign.go`, and any koine-specific modifications to `interpretation.go`.

**Step 4: Run koine tests**

```bash
cd ~/Documents/repos/koine && go test ./... -count=1 -race
```

Expected: all green.

**Step 5: Commit**

```bash
git add -A
git commit -m "refactor: import empirical BaseChart, remove duplicated dignity package"
```

---

## Verification

After all tasks:

```bash
# Empirical tests
cd ~/Documents/repos/empirical && go test ./... -count=1 -race

# Koiné tests
cd ~/Documents/repos/koine && go test ./... -count=1 -race

# Build both
cd ~/Documents/repos/empirical && go build ./...
cd ~/Documents/repos/koine && go build ./...

# Start server and hit an endpoint
cd ~/Documents/repos/empirical && go run ./cmd/recover serve 5432 &
sleep 2
curl -s -X POST http://localhost:5432/api/recover \
  -H 'Content-Type: application/json' \
  -d '{"name":"AJ","year":1969,"month":2,"day":15,"hour":23,"minute":10,"tz_offset":-8,"lat":47.038,"lng":-122.901}' | jq '.name'
```

Expected: `"AJ"` and valid JSON response.

---

## Net Line Change Estimate

| Phase | Files | Est. Δ |
|-------|-------|--------|
| BaseChart struct + constructor | +1 file | +150 lines |
| BaseChart tests | +1 file | +80 lines |
| KoinéFromBase | +1 file | +120 lines |
| VedicFromBase | +1 file | +100 lines |
| BaZiFromBase | +1 file | +60 lines |
| DraconicFromBase | +1 file | +40 lines |
| Refactor compute.go | ~33 functions | -200 lines (remove no-op copies, re-computation) |
| Remove chartData | -1 struct, -1 func | -50 lines |
| Unify koine | delete ~25 files | -5,000 lines (duplicated code) |
| **Total** | | **~-4,700 lines net** |

---

## Risks

1. **SWE initialization** — `ComputeBaseChart` must call `swe.SetEphePath()` before any SWE calls. Currently `computePositions` handles this. The constructor should accept an optional `ephePath` parameter or use `empirical.EnsureEpheCache()`.

2. **Performance** — Computing all 5 house systems + 25 fixed stars + 13 Arabic Parts + aspects for every request is more work than the current lazy approach. But: (a) it's still sub-second, (b) it eliminates 5-10 redundant SWE calls per request, (c) the cacheability is worth it.

3. **API compatibility** — The `/api/recover` endpoint currently returns `FullReport` with `phase1_dignity`, `phase3_houses`, etc. Changing it to return `KoinéChart` is a breaking change. Mitigation: keep the old endpoint working during transition, add new `/api/koine` endpoint.

4. **koine repo divergence** — The two repos have diverged in `interpretation.go`, `dignity.go`, `api.go`, etc. Merging them requires careful diff review. Some koine changes may need to be ported back to empirical.

5. **Test coverage** — `cmd/recover/` has no tests. The refactored compute functions need at least smoke tests to catch regressions.
