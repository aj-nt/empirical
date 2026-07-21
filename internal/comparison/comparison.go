// Package comparison provides cross-system comparison of astrological outputs.
// This is the empirical paper's method as code: ComputeBaseChart → [KoinéFromBase,
// WesternFromBase, VedicFromBase] → diff.
package comparison

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/aj-nt/empirical/internal/dignity"
)

// ── ComparisonReport ─────────────────────────────────────────────────────

// ComparisonReport holds the cross-system comparison results for a single chart.
type ComparisonReport struct {
	Name              string              `json:"name"`
	Systems           []string            `json:"systems"`
	PlanetSigns       []PlanetSignEntry   `json:"planet_signs"`
	PlanetHouses      []PlanetHouseEntry  `json:"planet_houses"`
	DignityComparison []DignityEntry      `json:"dignity_comparison"`
	Summary           ComparisonSummary   `json:"summary"`
}

// PlanetSignEntry shows which sign each system places a planet in.
type PlanetSignEntry struct {
	Planet  string            `json:"planet"`
	Systems map[string]string `json:"systems"` // system → sign name
	Agrees  bool              `json:"agrees"`  // all systems agree?
}

// PlanetHouseEntry shows which house each system places a planet in.
type PlanetHouseEntry struct {
	Planet  string         `json:"planet"`
	Systems map[string]int `json:"systems"` // system → house number
	Agrees  bool           `json:"agrees"`  // all systems agree?
}

// DignityEntry shows dignity state per planet per system.
type DignityEntry struct {
	Planet  string            `json:"planet"`
	Systems map[string]string `json:"systems"` // system → dignity state
	Agrees  bool              `json:"agrees"`  // all systems agree?
}

// ComparisonSummary holds aggregate agreement statistics.
type ComparisonSummary struct {
	SignAgreement    float64 `json:"sign_agreement"`    // fraction of planets where all systems agree on sign
	HouseAgreement   float64 `json:"house_agreement"`   // fraction of planets where all systems agree on house
	DignityAgreement float64 `json:"dignity_agreement"` // fraction of planets where all systems agree on dignity
	TotalPlanets     int     `json:"total_planets"`
}

// ── CompareSystems ────────────────────────────────────────────────────────

// CompareSystems runs multiple system transforms against a BaseChart and diffs
// the outputs. It compares Koiné, Western, and Vedic systems.
func CompareSystems(bc *dignity.BaseChart, orbDeg float64) *ComparisonReport {
	if orbDeg <= 0 {
		orbDeg = 5.0
	}

	report := &ComparisonReport{
		Name:    bc.Name,
		Systems: []string{"koine", "western", "vedic"},
	}

	// Run all three system transforms
	koineCI := dignity.KoinéFromBase(bc, orbDeg)
	westernCI := dignity.WesternFromBase(bc, orbDeg, false)
	vedicDC := dignity.VedicFromBase(bc)

	// ── Planet sign comparison ─────────────────────────────────────────
	planetSigns := make(map[string]map[string]string)

	// Koiné: extract from ChartInterpretation.PlanetSigns
	for _, ps := range koineCI.PlanetSigns {
		planet, sign := parsePlanetSign(ps)
		if planet == "" {
			continue
		}
		if planetSigns[planet] == nil {
			planetSigns[planet] = make(map[string]string)
		}
		planetSigns[planet]["koine"] = sign
	}

	// Western: extract from ChartInterpretation.PlanetSigns
	for _, ps := range westernCI.PlanetSigns {
		planet, sign := parsePlanetSign(ps)
		if planet == "" {
			continue
		}
		if planetSigns[planet] == nil {
			planetSigns[planet] = make(map[string]string)
		}
		planetSigns[planet]["western"] = sign
	}

	// Vedic: use PlanetDignity.SidSign from DignityConvergence
	for _, p := range vedicDC.Planets {
		if planetSigns[p.Planet] == nil {
			planetSigns[p.Planet] = make(map[string]string)
		}
		planetSigns[p.Planet]["vedic"] = p.SidSign
	}

	// Build sorted planet sign entries
	var planetNames []string
	for p := range planetSigns {
		planetNames = append(planetNames, p)
	}
	sort.Strings(planetNames)

	for _, planet := range planetNames {
		systems := planetSigns[planet]
		agrees := allSameString(systems)
		report.PlanetSigns = append(report.PlanetSigns, PlanetSignEntry{
			Planet:  planet,
			Systems: systems,
			Agrees:  agrees,
		})
	}

	// ── Planet house comparison ────────────────────────────────────────
	planetHouses := make(map[string]map[string]int)

	for _, ph := range koineCI.PlanetHouses {
		planet, house := parsePlanetHouse(ph)
		if planet == "" {
			continue
		}
		if planetHouses[planet] == nil {
			planetHouses[planet] = make(map[string]int)
		}
		planetHouses[planet]["koine"] = house
	}

	for _, ph := range westernCI.PlanetHouses {
		planet, house := parsePlanetHouse(ph)
		if planet == "" {
			continue
		}
		if planetHouses[planet] == nil {
			planetHouses[planet] = make(map[string]int)
		}
		planetHouses[planet]["western"] = house
	}

	// Vedic: compute whole-sign houses from sidereal positions
	vedicASC := dignity.NormalizeLon(bc.ASC - bc.Ayanamsa)
	for _, p := range vedicDC.Planets {
		// Get sidereal longitude from the tropical position minus ayanamsa
		if pos, ok := bc.Tropical[p.Planet]; ok {
			sidLon := dignity.NormalizeLon(pos.Lon - bc.Ayanamsa)
			house := ((int(sidLon/30) - int(vedicASC/30) + 12) % 12) + 1
			if planetHouses[p.Planet] == nil {
				planetHouses[p.Planet] = make(map[string]int)
			}
			planetHouses[p.Planet]["vedic"] = house
		}
	}

	for _, planet := range planetNames {
		houses, ok := planetHouses[planet]
		if !ok {
			continue
		}
		agrees := allSameInt(houses)
		report.PlanetHouses = append(report.PlanetHouses, PlanetHouseEntry{
			Planet:  planet,
			Systems: houses,
			Agrees:  agrees,
		})
	}

	// ── Dignity comparison ─────────────────────────────────────────────
	dignityMap := make(map[string]map[string]string)

	// Koiné: compute 2-state dignity (domicile/peregrine) from tropical positions
	for _, planet := range planetNames {
		if pos, ok := bc.Tropical[planet]; ok {
			state := computeKoinéDignity(planet, pos.Lon)
			if dignityMap[planet] == nil {
				dignityMap[planet] = make(map[string]string)
			}
			dignityMap[planet]["koine"] = state
		}
	}

	// Western: compute modern dignity from tropical positions
	for _, planet := range planetNames {
		if pos, ok := bc.Tropical[planet]; ok {
			state := computeWesternDignity(planet, pos.Lon)
			if dignityMap[planet] == nil {
				dignityMap[planet] = make(map[string]string)
			}
			dignityMap[planet]["western"] = state
		}
	}

	// Vedic: use PlanetDignity.Vedic from DignityConvergence
	for _, p := range vedicDC.Planets {
		if dignityMap[p.Planet] == nil {
			dignityMap[p.Planet] = make(map[string]string)
		}
		dignityMap[p.Planet]["vedic"] = p.Vedic
	}

	for _, planet := range planetNames {
		dignities, ok := dignityMap[planet]
		if !ok {
			continue
		}
		agrees := allSameString(dignities)
		report.DignityComparison = append(report.DignityComparison, DignityEntry{
			Planet:  planet,
			Systems: dignities,
			Agrees:  agrees,
		})
	}

	// ── Summary ────────────────────────────────────────────────────────
	totalPlanets := len(report.PlanetSigns)
	if totalPlanets > 0 {
		signAgree := 0
		for _, ps := range report.PlanetSigns {
			if ps.Agrees {
				signAgree++
			}
		}
		houseAgree := 0
		for _, ph := range report.PlanetHouses {
			if ph.Agrees {
				houseAgree++
			}
		}
		dignityAgree := 0
		for _, de := range report.DignityComparison {
			if de.Agrees {
				dignityAgree++
			}
		}
		report.Summary = ComparisonSummary{
			SignAgreement:    safeDiv(float64(signAgree), float64(totalPlanets)),
			HouseAgreement:   safeDiv(float64(houseAgree), float64(len(report.PlanetHouses))),
			DignityAgreement: safeDiv(float64(dignityAgree), float64(len(report.DignityComparison))),
			TotalPlanets:     totalPlanets,
		}
	}

	return report
}

// JSON serializes the ComparisonReport to JSON.
func (cr *ComparisonReport) JSON() ([]byte, error) {
	return json.MarshalIndent(cr, "", "  ")
}

// ── Dignity computation ───────────────────────────────────────────────────

// computeKoinéDignity returns the 2-state Hellenistic dignity for a planet.
func computeKoinéDignity(planet string, lon float64) string {
	sign := dignity.SignForLongitude(lon)
	if isDomicile(planet, sign) {
		return "domicile"
	}
	return "peregrine"
}

// computeWesternDignity returns the modern Western dignity for a planet.
func computeWesternDignity(planet string, lon float64) string {
	sign := dignity.SignForLongitude(lon)
	if isDomicile(planet, sign) {
		return "domicile"
	}
	if isExaltation(planet, sign) {
		return "exaltation"
	}
	if isDetriment(planet, sign) {
		return "detriment"
	}
	if isFall(planet, sign) {
		return "fall"
	}
	return "peregrine"
}

// ── Dignity tables ────────────────────────────────────────────────────────

var domicileMap = map[string][]string{
	"Sun":     {"Leo"},
	"Moon":    {"Cancer"},
	"Mercury": {"Gemini", "Virgo"},
	"Venus":   {"Taurus", "Libra"},
	"Mars":    {"Aries", "Scorpio"},
	"Jupiter": {"Sagittarius", "Pisces"},
	"Saturn":  {"Capricorn", "Aquarius"},
}

var exaltationMap = map[string]string{
	"Sun": "Aries", "Moon": "Taurus", "Mercury": "Virgo",
	"Venus": "Pisces", "Mars": "Capricorn", "Jupiter": "Cancer",
	"Saturn": "Libra",
}

var detrimentMap = map[string][]string{
	"Sun": {"Aquarius"}, "Moon": {"Capricorn"},
	"Mercury": {"Sagittarius", "Pisces"}, "Venus": {"Aries", "Scorpio"},
	"Mars": {"Taurus", "Libra"}, "Jupiter": {"Gemini", "Virgo"},
	"Saturn": {"Cancer", "Leo"},
}

var fallMap = map[string]string{
	"Sun": "Libra", "Moon": "Scorpio", "Mercury": "Pisces",
	"Venus": "Virgo", "Mars": "Cancer", "Jupiter": "Capricorn",
	"Saturn": "Aries",
}

func isDomicile(planet, sign string) bool {
	for _, s := range domicileMap[planet] {
		if s == sign {
			return true
		}
	}
	return false
}

func isExaltation(planet, sign string) bool {
	return exaltationMap[planet] == sign
}

func isDetriment(planet, sign string) bool {
	for _, s := range detrimentMap[planet] {
		if s == sign {
			return true
		}
	}
	return false
}

func isFall(planet, sign string) bool {
	return fallMap[planet] == sign
}

// ── Helpers ───────────────────────────────────────────────────────────────

// parsePlanetSign extracts planet name and sign from a "Planet in Sign — ..." string.
func parsePlanetSign(s string) (planet, sign string) {
	return parsePlanetIn(s)
}

// parsePlanetHouse extracts planet name and house number from a "Planet in Nth house — ..." string.
func parsePlanetHouse(s string) (planet string, house int) {
	planet, rest := parsePlanetIn(s)
	if planet == "" {
		return "", 0
	}
	var nth int
	n, err := fmt.Sscanf(rest, "%d", &nth)
	if n == 1 && err == nil {
		return planet, nth
	}
	n, err = fmt.Sscanf(rest, "the %d", &nth)
	if n == 1 && err == nil {
		return planet, nth
	}
	return "", 0
}

// parsePlanetIn extracts "Planet" and the rest from "Planet in ..." format.
func parsePlanetIn(s string) (planet, rest string) {
	for i := 0; i < len(s); i++ {
		if s[i] == ' ' && i+3 < len(s) && s[i+1:i+4] == "in " {
			return s[:i], s[i+4:]
		}
		if s[i] == ':' && i+1 < len(s) && s[i+1] == ' ' {
			return s[:i], s[i+2:]
		}
	}
	return "", ""
}

// allSameString returns true if all values in the map are identical.
func allSameString(m map[string]string) bool {
	if len(m) < 2 {
		return true
	}
	var first string
	firstSet := false
	for _, v := range m {
		if !firstSet {
			first = v
			firstSet = true
			continue
		}
		if v != first {
			return false
		}
	}
	return true
}

// allSameInt returns true if all values in the map are identical.
func allSameInt(m map[string]int) bool {
	if len(m) < 2 {
		return true
	}
	var first int
	firstSet := false
	for _, v := range m {
		if !firstSet {
			first = v
			firstSet = true
			continue
		}
		if v != first {
			return false
		}
	}
	return true
}

// safeDiv returns a/b, or 0 if b is 0.
func safeDiv(a, b float64) float64 {
	if b == 0 {
		return 0
	}
	return a / b
}
