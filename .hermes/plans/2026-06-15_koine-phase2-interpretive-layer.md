# Koiné Astrology — Phase 2: Interpretive Layer

> **For Hermes:** Use subagent-driven-development skill to implement this plan task-by-task.

**Goal:** Transform the template-based interpretation engine from a flat list of statements into a structured narrative reading using Koiné's 9 techniques.

**Architecture:** The interpretation engine (`internal/dignity/interpretation.go`) already has planet/sign/house/aspect/pattern descriptions and pair dynamics. Phase 2 simplifies dignity to 2-state, restricts aspects to 3 universal angles, and adds a synthesis engine that weights indicators, resolves conflicts, and produces a coherent narrative with opening, body, and closing sections.

**Tech Stack:** Go 1.21, existing dignity package, no new dependencies.

---

## Current State

The interpretation engine has:
- `InterpretPlanetInSign(planet, sign)` — uses 4-state dignity (domicile, detriment, exaltation, fall)
- `InterpretPlanetInHouse(planet, house)` — flat statement
- `InterpretAspect(p1, p2, aspect, orb)` — with pair dynamics
- `InterpretPattern(name, planets)` — pattern descriptions
- `InterpretChart(name, planets, houses, aspects, patterns, dignities)` — assembles flat lists
- `ChartInterpretation` struct with `PlanetSigns`, `PlanetHouses`, `Aspects`, `Patterns` string slices

What's missing:
- 2-state dignity (domicile + peregrine only)
- 3 universal aspects only (conjunction, opposition, trine)
- Whole-sign houses enforced
- Synthesis: weighting, theme detection, conflict resolution, narrative structure
- Fixed star integration in interpretation
- Draconic interpretation path
- Fortune integration

---

### Task 1: Switch to 2-state dignity in interpretation

**Objective:** `InterpretPlanetInSign` uses only domicile vs peregrine, dropping exaltation/detriment/fall.

**Files:**
- Modify: `internal/dignity/interpretation.go:235-246`

**Step 1: Update InterpretPlanetInSign dignity logic**

Replace the 4-state dignity block (lines 235-246) with 2-state:

```go
var dignity string
if strings.Contains(domicile[planet], sign) {
    dignity = "in domicile — at home, operating at full strength"
} else {
    dignity = "peregrine — wandering, operating without natural advantage"
}
```

**Step 2: Run tests**

```bash
cd /tmp/koine && go test ./internal/dignity/ -v -run TestInterpret 2>&1
```

Expected: PASS (tests check interpretation output format, not specific dignity wording)

**Step 3: Commit**

```bash
git add internal/dignity/interpretation.go
git commit -m "koine: switch to 2-state dignity (domicile + peregrine) in interpretation"
```

---

### Task 2: Restrict aspects to 3 universal angles

**Objective:** Interpretation only uses conjunction, opposition, trine. Update aspect catalog and `InterpretAspect`.

**Files:**
- Modify: `internal/dignity/interpretation.go:136-143` (aspectDescriptions)
- Modify: `internal/dignity/interpretation.go:190` (FindNatalAspects — caller controls aspects, but default should be 3)
- Modify: `cmd/recover/main.go` — computeInterpretation uses 3 aspects

**Step 1: Update aspectDescriptions map**

Keep only conjunction, opposition, trine. Remove square, sextile, quincunx:

```go
var aspectDescriptions = map[string]string{
    "conjunction": "merge and amplify — the two planets operate as one force",
    "opposition":  "polarity and tension — a seesaw between two extremes, awareness through contrast",
    "trine":       "flow and ease — natural harmony, talent that comes without effort",
}
```

**Step 2: Update computeInterpretation in main.go**

Find the aspects list used in `computeInterpretation` and change from `DefaultAspects()` to the 3 universal aspects:

```go
aspects := []dignity.AspectDef{
    {0, "conjunction"}, {120, "trine"}, {180, "opposition"},
}
```

**Step 3: Run tests**

```bash
cd /tmp/koine && go test ./... 2>&1
```

Expected: All PASS

**Step 4: Commit**

```bash
git add internal/dignity/interpretation.go cmd/recover/main.go
git commit -m "koine: restrict aspects to 3 universal angles (conj, opp, trine)"
```

---

### Task 3: Enforce whole-sign houses in interpretation endpoint

**Objective:** The `/api/interpretation` endpoint always uses whole-sign houses regardless of the `house_system` parameter.

**Files:**
- Modify: `cmd/recover/main.go` — `computeInterpretation` function

**Step 1: Hardcode whole-sign houses**

In `computeInterpretation`, replace the house system parameter usage with whole-sign:

```go
func computeInterpretation(name string, cd *chartData, lat, lng float64, houseSystem string, orbDeg float64, cacheDir string) ([]byte, error) {
    // Koiné uses whole-sign houses exclusively
    ascLon := cd.asc
    
    // Planet-to-house mapping (whole-sign from ASC)
    houses := make(map[string]int)
    for planet, lon := range cd.planets {
        house := ((int(lon/30) - int(ascLon/30) + 12) % 12) + 1
        houses[planet] = house
    }
    
    // 3 universal aspects only
    aspects := []AspectDef{{0, "conjunction"}, {120, "trine"}, {180, "opposition"}}
    aspectHits := FindNatalAspects(cd.planets, aspects, orbDeg)
    var hits []AspectHit
    for _, a := range aspectHits {
        hits = append(hits, AspectHit{
            Planet1: a.Planet1, Planet2: a.Planet2,
            Aspect: a.Aspect, Orb: a.Orb,
        })
    }
    
    // Patterns
    patternReport := DetectPatterns(cd.planets, orbDeg)
    var patternHits []PatternHit
    for _, p := range patternReport.Patterns {
        patternHits = append(patternHits, PatternHit{Name: p.Name, Planets: p.Planets})
    }
    
    report := InterpretChart(name, cd.planets, houses, hits, patternHits, nil)
    return report.JSON()
}
```

**Step 2: Run tests**

```bash
cd /tmp/koine && go test ./... 2>&1
```

Expected: All PASS

**Step 3: Commit**

```bash
git add cmd/recover/main.go
git commit -m "koine: enforce whole-sign houses in interpretation endpoint"
```

---

### Task 4: Add synthesis engine — theme detection and weighting

**Objective:** Build `SynthesizeChart` that takes the flat interpretation and produces a structured narrative with dominant theme detection, angular planet emphasis, and a closing synthesis.

**Files:**
- Create: `internal/dignity/synthesis.go`
- Create: `internal/dignity/synthesis_test.go`

**Step 1: Write failing test**

Create `internal/dignity/synthesis_test.go`:

```go
package dignity

import (
    "strings"
    "testing"
)

func TestSynthesizeChart_ReturnsStructuredNarrative(t *testing.T) {
    planets := map[string]float64{
        "Sun": 125.0, "Moon": 45.0, "Mercury": 130.0,
        "Venus": 100.0, "Mars": 200.0, "Jupiter": 300.0,
        "Saturn": 15.0,
    }
    houses := map[string]int{
        "Sun": 5, "Moon": 2, "Mercury": 5,
        "Venus": 4, "Mars": 7, "Jupiter": 10,
        "Saturn": 1,
    }
    aspects := []AspectHit{
        {Planet1: "Sun", Planet2: "Mercury", Aspect: "conjunction", Orb: 0.5},
        {Planet1: "Moon", Planet2: "Saturn", Aspect: "opposition", Orb: 1.2},
        {Planet1: "Venus", Planet2: "Mars", Aspect: "trine", Orb: 2.0},
    }
    patterns := []PatternHit{
        {Name: "Stellium", Planets: []string{"Sun", "Mercury", "Venus"}},
    }

    result := SynthesizeChart("Test", planets, houses, aspects, patterns)

    if result.Name != "Test" {
        t.Errorf("expected Name 'Test', got %q", result.Name)
    }
    if result.Opening == "" {
        t.Error("Opening should not be empty")
    }
    if len(result.Body) == 0 {
        t.Error("Body should have at least one section")
    }
    if result.Closing == "" {
        t.Error("Closing should not be empty")
    }
    if !strings.Contains(result.Opening, "Test") {
        t.Error("Opening should mention the name")
    }
}

func TestSynthesizeChart_DetectsDominantTheme(t *testing.T) {
    // Stellium in Leo (5th house) — creativity theme
    planets := map[string]float64{
        "Sun": 125.0, "Moon": 127.0, "Mercury": 128.0, "Venus": 126.0,
        "Mars": 200.0, "Jupiter": 300.0, "Saturn": 15.0,
    }
    houses := map[string]int{
        "Sun": 5, "Moon": 5, "Mercury": 5, "Venus": 5,
        "Mars": 7, "Jupiter": 10, "Saturn": 1,
    }
    aspects := []AspectHit{}
    patterns := []PatternHit{
        {Name: "Stellium", Planets: []string{"Sun", "Moon", "Mercury", "Venus"}},
    }

    result := SynthesizeChart("Test", planets, houses, aspects, patterns)

    // Should detect the stellium as dominant
    if !strings.Contains(strings.ToLower(result.Opening), "stellium") &&
       !strings.Contains(strings.ToLower(result.Opening), "concentration") {
        t.Errorf("Opening should mention the stellium: %q", result.Opening)
    }
}

func TestSynthesizeChart_AngularPlanetsEmphasized(t *testing.T) {
    // Saturn in 1st house (angular) should be highlighted
    planets := map[string]float64{
        "Sun": 125.0, "Moon": 45.0, "Saturn": 10.0,
    }
    houses := map[string]int{
        "Sun": 5, "Moon": 2, "Saturn": 1,
    }
    aspects := []AspectHit{}
    patterns := []PatternHit{}

    result := SynthesizeChart("Test", planets, houses, aspects, patterns)

    // Saturn in H1 should appear prominently
    found := false
    for _, section := range result.Body {
        if strings.Contains(section, "Saturn") && strings.Contains(section, "1") {
            found = true
            break
        }
    }
    if !found {
        t.Error("Saturn in 1st house should appear in body sections")
    }
}
```

**Step 2: Run test to verify failure**

```bash
cd /tmp/koine && go test ./internal/dignity/ -v -run TestSynthesize 2>&1
```

Expected: FAIL — "undefined: SynthesizeChart"

**Step 3: Write synthesis.go**

Create `internal/dignity/synthesis.go`:

```go
package dignity

import (
    "fmt"
    "sort"
    "strings"
)

// SynthesisReport holds a structured narrative chart reading.
type SynthesisReport struct {
    Name    string   `json:"name"`
    Opening string   `json:"opening"`
    Body    []string `json:"body"`
    Closing string   `json:"closing"`
}

// JSON returns the synthesis as JSON bytes.
func (s *SynthesisReport) JSON() ([]byte, error) {
    return json.Marshal(s)
}

// angularHouses are the houses that give planets extra prominence.
var angularHouses = map[int]bool{1: true, 4: true, 7: true, 10: true}

// succedentHouses are secondary in strength.
var succedentHouses = map[int]bool{2: true, 5: true, 8: true, 11: true}

// SynthesizeChart produces a structured narrative reading from chart data.
func SynthesizeChart(
    name string,
    planets map[string]float64,
    houses map[string]int,
    aspects []AspectHit,
    patterns []PatternHit,
) *SynthesisReport {
    report := &SynthesisReport{
        Name: name,
        Body: make([]string, 0),
    }

    // ── Opening: dominant signature ──────────────────────────────
    report.Opening = buildOpening(name, planets, houses, aspects, patterns)

    // ── Body sections ────────────────────────────────────────────
    
    // Section 1: Angular planets (most prominent)
    angularSection := buildAngularSection(planets, houses)
    if angularSection != "" {
        report.Body = append(report.Body, angularSection)
    }

    // Section 2: Luminaries (Sun and Moon)
    lumSection := buildLuminarySection(planets, houses)
    if lumSection != "" {
        report.Body = append(report.Body, lumSection)
    }

    // Section 3: Personal planets in signs and houses
    personalSection := buildPersonalSection(planets, houses)
    if personalSection != "" {
        report.Body = append(report.Body, personalSection)
    }

    // Section 4: Key aspects
    aspectSection := buildAspectSection(aspects)
    if aspectSection != "" {
        report.Body = append(report.Body, aspectSection)
    }

    // Section 5: Patterns
    patternSection := buildPatternSection(patterns)
    if patternSection != "" {
        report.Body = append(report.Body, patternSection)
    }

    // ── Closing: synthesis ───────────────────────────────────────
    report.Closing = buildClosing(name, planets, houses, aspects, patterns)

    return report
}

// buildOpening identifies the chart's dominant signature.
func buildOpening(name string, planets map[string]float64, houses map[string]int, aspects []AspectHit, patterns []PatternHit) string {
    var parts []string
    parts = append(parts, fmt.Sprintf("%s's chart shows", name))

    // Check for stellium (3+ planets in one sign or house)
    signCounts := make(map[string]int)
    houseCounts := make(map[int]int)
    for planet, lon := range planets {
        sign := SignForLongitude(lon)
        signCounts[sign]++
        if h, ok := houses[planet]; ok {
            houseCounts[h]++
        }
    }

    var dominantSign string
    var dominantSignCount int
    for sign, count := range signCounts {
        if count >= 3 && count > dominantSignCount {
            dominantSign = sign
            dominantSignCount = count
        }
    }

    var dominantHouse int
    var dominantHouseCount int
    for house, count := range houseCounts {
        if count >= 3 && count > dominantHouseCount {
            dominantHouse = house
            dominantHouseCount = count
        }
    }

    if dominantSign != "" {
        sd := signDescriptions[dominantSign]
        parts = append(parts, fmt.Sprintf("a concentration in %s (%s)", dominantSign, sd))
    }

    if dominantHouse != 0 {
        hd := houseDescriptions[dominantHouse]
        parts = append(parts, fmt.Sprintf("with emphasis on %s", hd))
    }

    // Mention patterns
    if len(patterns) > 0 {
        patternNames := make([]string, len(patterns))
        for i, p := range patterns {
            patternNames[i] = p.Name
        }
        parts = append(parts, fmt.Sprintf("forming a %s configuration", strings.Join(patternNames, " and ")))
    }

    // Angular planet count
    angularCount := 0
    for _, h := range houses {
        if angularHouses[h] {
            angularCount++
        }
    }
    if angularCount >= 3 {
        parts = append(parts, "with strong angular emphasis — the chart is outwardly active")
    }

    if len(parts) == 1 {
        parts = append(parts, "a balanced distribution of planetary energies")
    }

    return strings.Join(parts, ", ") + "."
}

// buildAngularSection highlights planets in angular houses.
func buildAngularSection(planets map[string]float64, houses map[string]int) string {
    var lines []string
    houseNames := map[int]string{1: "1st house (self)", 4: "4th house (foundation)", 7: "7th house (partnership)", 10: "10th house (vocation)"}

    for planet, house := range houses {
        if angularHouses[house] {
            sign := SignForLongitude(planets[planet])
            desc := InterpretPlanetInSign(planet, sign)
            hn := houseNames[house]
            lines = append(lines, fmt.Sprintf("%s in the %s: %s", planet, hn, desc))
        }
    }

    if len(lines) == 0 {
        return ""
    }

    return "Angular Planets (most visible):\n" + strings.Join(lines, "\n")
}

// buildLuminarySection describes Sun and Moon placements.
func buildLuminarySection(planets map[string]float64, houses map[string]int) string {
    var lines []string

    if sunLon, ok := planets["Sun"]; ok {
        sign := SignForLongitude(sunLon)
        house := houses["Sun"]
        hd := houseDescriptions[house]
        lines = append(lines, fmt.Sprintf("Sun in %s in the %s: %s",
            sign, hd, InterpretPlanetInSign("Sun", sign)))
    }

    if moonLon, ok := planets["Moon"]; ok {
        sign := SignForLongitude(moonLon)
        house := houses["Moon"]
        hd := houseDescriptions[house]
        lines = append(lines, fmt.Sprintf("Moon in %s in the %s: %s",
            sign, hd, InterpretPlanetInSign("Moon", sign)))
    }

    if len(lines) == 0 {
        return ""
    }

    return "Luminaries (core identity and emotional nature):\n" + strings.Join(lines, "\n")
}

// buildPersonalSection covers Mercury, Venus, Mars.
func buildPersonalSection(planets map[string]float64, houses map[string]int) string {
    personalPlanets := []string{"Mercury", "Venus", "Mars"}
    var lines []string

    for _, planet := range personalPlanets {
        if lon, ok := planets[planet]; ok {
            sign := SignForLongitude(lon)
            house := houses[planet]
            hd := houseDescriptions[house]
            lines = append(lines, fmt.Sprintf("%s in %s in the %s: %s",
                planet, sign, hd, InterpretPlanetInSign(planet, sign)))
        }
    }

    if len(lines) == 0 {
        return ""
    }

    return "Personal Planets (communication, values, drive):\n" + strings.Join(lines, "\n")
}

// buildAspectSection describes the most significant aspects.
func buildAspectSection(aspects []AspectHit) string {
    if len(aspects) == 0 {
        return ""
    }

    // Sort by orb (tightest first)
    sorted := make([]AspectHit, len(aspects))
    copy(sorted, aspects)
    sort.Slice(sorted, func(i, j int) bool { return sorted[i].Orb < sorted[j].Orb })

    var lines []string
    for _, a := range sorted {
        lines = append(lines, InterpretAspect(a.Planet1, a.Planet2, a.Aspect, a.Orb))
    }

    return "Key Aspects:\n" + strings.Join(lines, "\n")
}

// buildPatternSection describes detected aspect patterns.
func buildPatternSection(patterns []PatternHit) string {
    if len(patterns) == 0 {
        return ""
    }

    var lines []string
    for _, p := range patterns {
        lines = append(lines, InterpretPattern(p.Name, p.Planets))
    }

    return "Patterns:\n" + strings.Join(lines, "\n")
}

// buildClosing synthesizes the chart's overall character.
func buildClosing(name string, planets map[string]float64, houses map[string]int, aspects []AspectHit, patterns []PatternHit) string {
    // Count elements
    elementCounts := map[string]int{"fire": 0, "earth": 0, "air": 0, "water": 0}
    elementSigns := map[string]string{
        "Aries": "fire", "Leo": "fire", "Sagittarius": "fire",
        "Taurus": "earth", "Virgo": "earth", "Capricorn": "earth",
        "Gemini": "air", "Libra": "air", "Aquarius": "air",
        "Cancer": "water", "Scorpio": "water", "Pisces": "water",
    }

    for _, lon := range planets {
        sign := SignForLongitude(lon)
        if elem, ok := elementSigns[sign]; ok {
            elementCounts[elem]++
        }
    }

    // Find dominant element
    dominantElem := ""
    dominantCount := 0
    for elem, count := range elementCounts {
        if count > dominantCount {
            dominantElem = elem
            dominantCount = count
        }
    }

    // Count modalities
    modalCounts := map[string]int{"cardinal": 0, "fixed": 0, "mutable": 0}
    modalSigns := map[string]string{
        "Aries": "cardinal", "Cancer": "cardinal", "Libra": "cardinal", "Capricorn": "cardinal",
        "Taurus": "fixed", "Leo": "fixed", "Scorpio": "fixed", "Aquarius": "fixed",
        "Gemini": "mutable", "Virgo": "mutable", "Sagittarius": "mutable", "Pisces": "mutable",
    }

    for _, lon := range planets {
        sign := SignForLongitude(lon)
        if mod, ok := modalSigns[sign]; ok {
            modalCounts[mod]++
        }
    }

    dominantMod := ""
    dominantModCount := 0
    for mod, count := range modalCounts {
        if count > dominantModCount {
            dominantMod = mod
            dominantModCount = count
        }
    }

    elementDescriptions := map[string]string{
        "fire":   "initiatory and driven by inspiration",
        "earth":  "grounded and driven by tangible results",
        "air":    "intellectual and driven by connection",
        "water":  "emotional and driven by depth of experience",
    }

    modalDescriptions := map[string]string{
        "cardinal": "initiates and pushes forward",
        "fixed":    "sustains and resists change",
        "mutable":  "adapts and flows between states",
    }

    var parts []string
    parts = append(parts, fmt.Sprintf("Overall, %s's chart is predominantly %s (%s)",
        name, dominantElem, elementDescriptions[dominantElem]))
    parts = append(parts, fmt.Sprintf("with a %s modality (%s)",
        dominantMod, modalDescriptions[dominantMod]))

    // Angular emphasis summary
    angularCount := 0
    for _, h := range houses {
        if angularHouses[h] {
            angularCount++
        }
    }
    if angularCount >= 3 {
        parts = append(parts, "The strong angular presence suggests a life lived visibly — actions have public consequence")
    } else if angularCount <= 1 {
        parts = append(parts, "The chart turns inward — most activity happens in private or intermediate houses")
    }

    // Aspect summary
    conjCount, oppCount, trineCount := 0, 0, 0
    for _, a := range aspects {
        switch a.Aspect {
        case "conjunction":
            conjCount++
        case "opposition":
            oppCount++
        case "trine":
            trineCount++
        }
    }
    if oppCount >= 2 {
        parts = append(parts, "Multiple oppositions create productive tension — growth through navigating polarities")
    }
    if trineCount >= 3 {
        parts = append(parts, "The abundance of trines suggests natural talents that may need conscious activation to avoid inertia")
    }

    return strings.Join(parts, ". ") + "."
}
```

**Step 4: Run tests**

```bash
cd /tmp/koine && go test ./internal/dignity/ -v -run TestSynthesize 2>&1
```

Expected: PASS

**Step 5: Commit**

```bash
git add internal/dignity/synthesis.go internal/dignity/synthesis_test.go
git commit -m "koine: add synthesis engine with theme detection, weighting, narrative structure"
```

---

### Task 5: Wire synthesis into the interpretation endpoint

**Objective:** The `/api/interpretation` endpoint returns the synthesis report alongside the flat interpretation.

**Files:**
- Modify: `internal/dignity/interpretation.go` — add `Synthesis` field to `ChartInterpretation`
- Modify: `cmd/recover/main.go` — `computeInterpretation` calls `SynthesizeChart`

**Step 1: Add Synthesis field to ChartInterpretation**

```go
type ChartInterpretation struct {
    Name         string          `json:"name"`
    PlanetSigns  []string        `json:"planet_signs"`
    PlanetHouses []string        `json:"planet_houses"`
    Aspects      []string        `json:"aspects"`
    Patterns     []string        `json:"patterns"`
    Synthesis    *SynthesisReport `json:"synthesis"`
}
```

**Step 2: Update computeInterpretation to call SynthesizeChart**

After building the `ChartInterpretation`, add:

```go
report.Synthesis = SynthesizeChart(name, cd.planets, houses, hits, patternHits)
```

**Step 3: Run tests**

```bash
cd /tmp/koine && go test ./... 2>&1
```

Expected: All PASS

**Step 4: Commit**

```bash
git add internal/dignity/interpretation.go cmd/recover/main.go
git commit -m "koine: wire synthesis engine into interpretation endpoint"
```

---

### Task 6: Add draconic interpretation endpoint

**Objective:** New `/api/draconic-interpretation` endpoint that interprets the draconic chart using the same synthesis engine.

**Files:**
- Modify: `internal/server/server.go` — add `DraconicInterpretationFunc` type and handler
- Modify: `cmd/recover/main.go` — add `computeDraconicInterpretation` and wire it

**Step 1: Add function type and handler in server.go**

Add after the existing DraconicSynastryFullFunc:

```go
// DraconicInterpretationFunc produces a synthesis reading of the draconic chart.
type DraconicInterpretationFunc func(name string, year, month, day, hour, minute int, tzOff, lat, lng float64, orbDeg float64) ([]byte, error)
```

Add handler in NewMux (and update NewMux signature + Run signature):

```go
mux.HandleFunc("/api/draconic-interpretation", func(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "POST only", http.StatusMethodNotAllowed)
        return
    }
    var req ChartRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
        return
    }
    if draconicInterpretation == nil {
        http.Error(w, "not available", http.StatusNotImplemented)
        return
    }
    result, err := draconicInterpretation(req.Name, req.Year, req.Month, req.Day, req.Hour, req.Minute, req.TzOffset, req.Lat, req.Lng, req.Orb)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.Write(result)
})
```

**Step 2: Add computeDraconicInterpretation in main.go**

```go
draconicInterpretation := func(name string, year, month, day, hour, minute int, tzOff, lat, lng float64, orbDeg float64) ([]byte, error) {
    cd := computePositions(year, month, day, hour, minute, tzOff, lat, lng, cacheDir)
    return computeDraconicInterpretation(name, cd, orbDeg)
}
```

And the compute function:

```go
func computeDraconicInterpretation(name string, cd *chartData, orbDeg float64) ([]byte, error) {
    tropical := make(map[string]float64)
    for k, v := range cd.planets {
        tropical[k] = v
    }
    
    // Compute draconic positions
    drac := ComputeDraconic(tropical, cd.nn)
    
    // Whole-sign houses from draconic ASC (use tropical ASC as reference — 
    // draconic houses use the same ASC since houses are relative to Earth, not zodiac)
    ascLon := cd.asc
    houses := make(map[string]int)
    for planet, lon := range drac.Planets {
        houses[planet] = ((int(lon/30) - int(ascLon/30) + 12) % 12) + 1
    }
    
    // 3 universal aspects on draconic positions
    aspects := []AspectDef{{0, "conjunction"}, {120, "trine"}, {180, "opposition"}}
    aspectHits := FindNatalAspects(drac.Planets, aspects, orbDeg)
    var hits []AspectHit
    for _, a := range aspectHits {
        hits = append(hits, AspectHit{
            Planet1: a.Planet1, Planet2: a.Planet2,
            Aspect: a.Aspect, Orb: a.Orb,
        })
    }
    
    // Patterns on draconic positions
    patternReport := DetectPatterns(drac.Planets, orbDeg)
    var patternHits []PatternHit
    for _, p := range patternReport.Patterns {
        patternHits = append(patternHits, PatternHit{Name: p.Name, Planets: p.Planets})
    }
    
    // Synthesize
    synthesis := SynthesizeChart(name+" (draconic)", drac.Planets, houses, hits, patternHits)
    return synthesis.JSON()
}
```

**Step 3: Update server.Run call in main.go**

Add `draconicInterpretation` param to the server.Run call.

**Step 4: Run tests**

```bash
cd /tmp/koine && go test ./... 2>&1
```

Expected: All PASS

**Step 5: Commit**

```bash
git add internal/server/server.go cmd/recover/main.go
git commit -m "koine: add draconic interpretation endpoint"
```

---

### Task 7: Update dashboard with synthesis and draconic tabs

**Objective:** Dashboard shows the synthesis narrative and has a draconic interpretation tab.

**Files:**
- Modify: `web/index.html`

**Step 1: Add synthesis display to interpretation tab**

Update `loadInterpretation` to show the synthesis opening and closing prominently at the top and bottom:

```javascript
async function loadInterpretation() {
  const body = {...bd(), house_system: $('#hsys').value, orb: +$('#orb').value};
  const r = await post('/api/interpretation', body);
  const d = await r.json();
  let html = '';
  
  // Synthesis opening
  if (d.synthesis) {
    html += `<div class="card"><h2>Synthesis</h2>`;
    html += `<p style="font-size:1.05rem;line-height:1.6">${d.synthesis.opening}</p>`;
    for (const section of d.synthesis.body) {
      html += `<div style="margin:12px 0;white-space:pre-line;font-size:0.9rem">${section}</div>`;
    }
    html += `<p style="font-size:1.05rem;line-height:1.6;margin-top:12px">${d.synthesis.closing}</p>`;
    html += '</div>';
  }
  
  // Flat interpretation below
  html += '<div class="card"><h2>Planets in Signs</h2>';
  for (const s of d.planet_signs) html += `<div style="margin:4px 0;font-size:0.85rem">${s}</div>`;
  html += '</div><div class="card"><h2>Planets in Houses</h2>';
  for (const s of d.planet_houses) html += `<div style="margin:4px 0;font-size:0.85rem">${s}</div>`;
  html += '</div><div class="card"><h2>Aspects</h2>';
  for (const s of d.aspects) html += `<div style="margin:4px 0;font-size:0.85rem">${s}</div>`;
  html += '</div><div class="card"><h2>Patterns</h2>';
  for (const s of d.patterns) html += `<div style="margin:4px 0;font-size:0.85rem">${s}</div>`;
  html += '</div>';
  $('#tab-interpretation').innerHTML = html;
}
```

**Step 2: Add draconic interpretation tab**

Add tab button: `<div class="tab" onclick="switchTab('draconic-interp')">Draconic Reading</div>`

Add tab content div: `<div id="tab-draconic-interp" class="tab-content" style="display:none"></div>`

Add to TAB_ORDER array and loadAll().

Add loadDraconicInterpretation function:

```javascript
async function loadDraconicInterpretation() {
  const body = {...bd(), orb: +$('#orb').value};
  const r = await post('/api/draconic-interpretation', body);
  const d = await r.json();
  let html = '<div class="card"><h2>Draconic Chart Reading</h2>';
  html += `<p style="font-size:1.05rem;line-height:1.6">${d.opening}</p>`;
  for (const section of d.body) {
    html += `<div style="margin:12px 0;white-space:pre-line;font-size:0.9rem">${section}</div>`;
  }
  html += `<p style="font-size:1.05rem;line-height:1.6;margin-top:12px">${d.closing}</p>`;
  html += '</div>';
  $('#tab-draconic-interp').innerHTML = html;
}
```

**Step 3: Commit**

```bash
git add web/index.html
git commit -m "koine: add synthesis display and draconic interpretation tab to dashboard"
```

---

## Verification

After all tasks:

```bash
cd /tmp/koine && go build ./... && go test ./... 2>&1
```

Expected: Build clean, all tests pass.

Start server and test:

```bash
cd /tmp/koine && go run ./cmd/recover serve 5433 &
sleep 1
curl -s -X POST http://localhost:5433/api/interpretation \
  -H 'Content-Type: application/json' \
  -d '{"name":"AJ","year":1969,"month":2,"day":15,"hour":23,"minute":10,"tz_offset":-8,"lat":47.038,"lng":-122.901,"house_system":"W","orb":3}' | python3 -m json.tool | head -40
```

Expected: JSON response with `synthesis` field containing `opening`, `body`, `closing`.
