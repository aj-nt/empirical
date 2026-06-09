package dignity

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Signs is the ordered list of zodiac signs.
var Signs = []string{
	"Aries", "Taurus", "Gemini", "Cancer", "Leo", "Virgo",
	"Libra", "Scorpio", "Sagittarius", "Capricorn", "Aquarius", "Pisces",
}

// ClassicalPlanets are the seven classical planets used in dignity comparisons.
var ClassicalPlanets = []string{
	"Sun", "Moon", "Mercury", "Venus", "Mars", "Jupiter", "Saturn",
}

// ── Dignity Tables ──────────────────────────────────────────────────────

type westernRules struct {
	Domicile   []string
	Exaltation []string
	Detriment  []string
	Fall       []string
}

var westernDignityTable = map[string]westernRules{
	"Sun":     {Domicile: []string{"Leo"}, Exaltation: []string{"Aries"}, Detriment: []string{"Aquarius"}, Fall: []string{"Libra"}},
	"Moon":    {Domicile: []string{"Cancer"}, Exaltation: []string{"Taurus"}, Detriment: []string{"Capricorn"}, Fall: []string{"Scorpio"}},
	"Mercury": {Domicile: []string{"Gemini", "Virgo"}, Exaltation: []string{"Virgo"}, Detriment: []string{"Sagittarius", "Pisces"}, Fall: []string{"Pisces"}},
	"Venus":   {Domicile: []string{"Taurus", "Libra"}, Exaltation: []string{"Pisces"}, Detriment: []string{"Aries", "Scorpio"}, Fall: []string{"Virgo"}},
	"Mars":    {Domicile: []string{"Aries", "Scorpio"}, Exaltation: []string{"Capricorn"}, Detriment: []string{"Libra", "Taurus"}, Fall: []string{"Cancer"}},
	"Jupiter": {Domicile: []string{"Sagittarius", "Pisces"}, Exaltation: []string{"Cancer"}, Detriment: []string{"Gemini", "Virgo"}, Fall: []string{"Capricorn"}},
	"Saturn":  {Domicile: []string{"Capricorn", "Aquarius"}, Exaltation: []string{"Libra"}, Detriment: []string{"Cancer", "Leo"}, Fall: []string{"Aries"}},
}

type vedicRules struct {
	Swakshetra []string
	Uchcha     []string
	Neecha     []string
}

var vedicDignityTable = map[string]vedicRules{
	"Sun":     {Swakshetra: []string{"Leo"}, Uchcha: []string{"Aries"}, Neecha: []string{"Libra"}},
	"Moon":    {Swakshetra: []string{"Cancer"}, Uchcha: []string{"Taurus"}, Neecha: []string{"Scorpio"}},
	"Mercury": {Swakshetra: []string{"Gemini", "Virgo"}, Uchcha: []string{"Virgo"}, Neecha: []string{"Pisces"}},
	"Venus":   {Swakshetra: []string{"Taurus", "Libra"}, Uchcha: []string{"Pisces"}, Neecha: []string{"Virgo"}},
	"Mars":    {Swakshetra: []string{"Aries", "Scorpio"}, Uchcha: []string{"Capricorn"}, Neecha: []string{"Cancer"}},
	"Jupiter": {Swakshetra: []string{"Sagittarius", "Pisces"}, Uchcha: []string{"Cancer"}, Neecha: []string{"Capricorn"}},
	"Saturn":  {Swakshetra: []string{"Capricorn", "Aquarius"}, Uchcha: []string{"Libra"}, Neecha: []string{"Aries"}},
}

var westernCategoryNames = map[string]string{
	"domicile":   "domicile",
	"exaltation": "exalted",
	"detriment":  "detriment",
	"fall":       "fall",
}

var vedicCategoryNames = map[string]string{
	"swakshetra": "own sign (domicile)",
	"uchcha":     "exalted",
	"neecha":     "debilitated (fall)",
}

// ── Sign Lookup ─────────────────────────────────────────────────────────

// SignForLongitude returns the zodiac sign for an ecliptic longitude in degrees.
func SignForLongitude(lonDeg float64) string {
	idx := (int(lonDeg)%360 + 360) % 360 / 30
	return Signs[idx]
}

// ── Dignity Computation ─────────────────────────────────────────────────

// WesternDignity returns the Western essential dignity category for a planet
// in a tropical sign, or "peregrine" if no special dignity applies.
func WesternDignity(planet, tropicalSign string) string {
	r, ok := westernDignityTable[planet]
	if !ok {
		return "peregrine"
	}
	for _, s := range r.Domicile {
		if s == tropicalSign {
			return "domicile"
		}
	}
	for _, s := range r.Exaltation {
		if s == tropicalSign {
			return "exaltation"
		}
	}
	for _, s := range r.Detriment {
		if s == tropicalSign {
			return "detriment"
		}
	}
	for _, s := range r.Fall {
		if s == tropicalSign {
			return "fall"
		}
	}
	return "peregrine"
}

// VedicDignity returns the Vedic essential dignity category for a planet
// in a sidereal sign, or "peregrine" if no special dignity applies.
func VedicDignity(planet, siderealSign string) string {
	r, ok := vedicDignityTable[planet]
	if !ok {
		return "peregrine"
	}
	for _, s := range r.Swakshetra {
		if s == siderealSign {
			return "swakshetra"
		}
	}
	for _, s := range r.Uchcha {
		if s == siderealSign {
			return "uchcha"
		}
	}
	for _, s := range r.Neecha {
		if s == siderealSign {
			return "neecha"
		}
	}
	return "peregrine"
}

// ── Convergence Classification ─────────────────────────────────────────

// ClassifyConvergence determines whether Western and Vedic dignities agree,
// diverge, or one system assigns a dignity the other lacks.
func ClassifyConvergence(western, vedic string) string {
	if western == "domicile" && vedic == "swakshetra" {
		return "agree"
	}
	if western == "exaltation" && vedic == "uchcha" {
		return "agree"
	}
	if western == "fall" && vedic == "neecha" {
		return "agree"
	}
	if western == "peregrine" && vedic == "peregrine" {
		return "agree"
	}
	// Detriment has no Vedic parallel
	if western == "detriment" {
		if vedic == "peregrine" {
			return "western_only"
		}
		return "diverge"
	}
	// Western sees peregrine, Vedic sees something
	if western == "peregrine" {
		return "vedic_only"
	}
	// Both assign dignities but different ones
	return "diverge"
}

// ── Data Types ──────────────────────────────────────────────────────────

// PlanetDignity is the dignity assessment for one planet across both systems.
type PlanetDignity struct {
	Planet      string
	TropSign    string
	SidSign     string
	Western     string
	Vedic       string
	Convergence string
}

// IsSignal returns true if both systems agree.
func (p PlanetDignity) IsSignal() bool {
	return p.Convergence == "agree"
}

// IsNoise returns true if systems diverge or only one assigns a dignity.
func (p PlanetDignity) IsNoise() bool {
	return p.Convergence != "agree"
}

// DignityConvergence is the complete dignity convergence report for a chart.
type DignityConvergence struct {
	Name            string
	AyanamsaDegrees float64
	Planets         []PlanetDignity
}

// SignalCount returns the number of planets where both systems agree.
func (dc *DignityConvergence) SignalCount() int {
	n := 0
	for _, p := range dc.Planets {
		if p.IsSignal() {
			n++
		}
	}
	return n
}

// NoiseCount returns the number of planets where systems diverge.
func (dc *DignityConvergence) NoiseCount() int {
	return len(dc.Planets) - dc.SignalCount()
}

// ConvergenceRate returns the fraction of planets with agreement.
func (dc *DignityConvergence) ConvergenceRate() float64 {
	if len(dc.Planets) == 0 {
		return 0
	}
	return float64(dc.SignalCount()) / float64(len(dc.Planets))
}

// SignalPlanets returns the names of planets where both systems agree.
func (dc *DignityConvergence) SignalPlanets() []string {
	var planets []string
	for _, p := range dc.Planets {
		if p.IsSignal() {
			planets = append(planets, p.Planet)
		}
	}
	return planets
}

// NoisePlanets returns the names of planets where systems diverge.
func (dc *DignityConvergence) NoisePlanets() []string {
	var planets []string
	for _, p := range dc.Planets {
		if p.IsNoise() {
			planets = append(planets, p.Planet)
		}
	}
	return planets
}

// ── Compute ─────────────────────────────────────────────────────────────

// ComputeDignityConvergence runs the full Phase 1 dignity convergence analysis.
// tropicalLongitudes maps planet names to tropical ecliptic longitudes (0-360).
// ayanamsa is the Lahiri ayanamsa in degrees for the birth date.
func ComputeDignityConvergence(tropicalLongitudes map[string]float64, ayanamsa float64, name string) *DignityConvergence {
	result := &DignityConvergence{
		Name:            name,
		AyanamsaDegrees: ayanamsa,
	}
	for _, planet := range ClassicalPlanets {
		tropLon, ok := tropicalLongitudes[planet]
		if !ok {
			continue
		}
		sidLon := tropLon - ayanamsa
		if sidLon < 0 {
			sidLon += 360
		}
		tropSign := SignForLongitude(tropLon)
		sidSign := SignForLongitude(sidLon)

		wDig := WesternDignity(planet, tropSign)
		vDig := VedicDignity(planet, sidSign)
		conv := ClassifyConvergence(wDig, vDig)

		result.Planets = append(result.Planets, PlanetDignity{
			Planet:      planet,
			TropSign:    tropSign,
			SidSign:     sidSign,
			Western:     wDig,
			Vedic:       vDig,
			Convergence: conv,
		})
	}
	return result
}

// ── JSON Output ─────────────────────────────────────────────────────────

// dignityJSONOutput is the serializable form of DignityConvergence.
type dignityJSONOutput struct {
	Name            string          `json:"name"`
	AyanamsaDegrees float64         `json:"ayanamsa_degrees"`
	Planets         []PlanetDignity `json:"planets"`
	SignalCount     int             `json:"signal_count"`
	NoiseCount      int             `json:"noise_count"`
	SignalPlanets   []string        `json:"signal_planets"`
	NoisePlanets    []string        `json:"noise_planets"`
	ConvergenceRate float64         `json:"convergence_rate"`
}

// ToJSON encodes the dignity convergence report as pretty-printed JSON.
func (dc *DignityConvergence) ToJSON() (string, error) {
	out := dignityJSONOutput{
		Name:            dc.Name,
		AyanamsaDegrees: dc.AyanamsaDegrees,
		Planets:         dc.Planets,
		SignalCount:     dc.SignalCount(),
		NoiseCount:      dc.NoiseCount(),
		SignalPlanets:   dc.SignalPlanets(),
		NoisePlanets:    dc.NoisePlanets(),
		ConvergenceRate: round(dc.ConvergenceRate(), 3),
	}
	b, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func round(v float64, decimals int) float64 {
	pow := 1.0
	for i := 0; i < decimals; i++ {
		pow *= 10
	}
	return float64(int(v*pow+0.5)) / pow
}

// FormatConvergence formats a dignity convergence report as human-readable text.
func FormatConvergence(dc *DignityConvergence) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Dignity Convergence Report \u2014 %s\n", dc.Name)
	fmt.Fprintf(&b, "Ayanamsa: %.4f deg (Lahiri)\n\n", dc.AyanamsaDegrees)

	// Header
	fmt.Fprintf(&b, "%-10s %-14s %-14s %-14s %-14s %s\n",
		"Planet", "Trop Sign", "Sid Sign", "Western", "Vedic", "Verdict")
	fmt.Fprintln(&b, strings.Repeat("-", 78))

	for _, p := range dc.Planets {
		wLabel := westernCategoryNames[p.Western]
		if wLabel == "" {
			wLabel = p.Western
		}
		vLabel := vedicCategoryNames[p.Vedic]
		if vLabel == "" {
			vLabel = p.Vedic
		}
		verdict := map[string]string{
			"agree":        "SIGNAL (agree)",
			"diverge":      "NOISE (diverge)",
			"western_only": "NOISE (Western only)",
			"vedic_only":   "NOISE (Vedic only)",
		}[p.Convergence]

		fmt.Fprintf(&b, "%-10s %-14s %-14s %-14s %-14s %s\n",
			p.Planet, p.TropSign, p.SidSign, wLabel, vLabel, verdict)
	}

	fmt.Fprintln(&b)
	sigPl := dc.SignalPlanets()
	noiPl := dc.NoisePlanets()
	sigStr := "none"
	if len(sigPl) > 0 {
		sigStr = strings.Join(sigPl, ", ")
	}
	noiStr := "none"
	if len(noiPl) > 0 {
		noiStr = strings.Join(noiPl, ", ")
	}
	fmt.Fprintf(&b, "Signal: %d/%d (%.0f%%) \u2014 %s\n",
		dc.SignalCount(), len(dc.Planets), dc.ConvergenceRate()*100, sigStr)
	fmt.Fprintf(&b, "Noise:  %d/%d \u2014 %s\n",
		dc.NoiseCount(), len(dc.Planets), noiStr)
	fmt.Fprintln(&b)

	// Analysis
	if dc.ConvergenceRate() >= 0.8 {
		fmt.Fprintln(&b, "HIGH convergence: dignity is largely invariant. The coordinate shift")
		fmt.Fprintln(&b, "(tropical \u2192 sidereal) preserves the dignity structure for most planets.")
	} else if dc.ConvergenceRate() >= 0.5 {
		fmt.Fprintln(&b, "MODERATE convergence: dignity partially invariant. The coordinate")
		fmt.Fprintln(&b, "shift breaks some placements but preserves others.")
	} else {
		fmt.Fprintln(&b, "LOW convergence: dignity is coordinate-sensitive. Much of the signal")
		fmt.Fprintln(&b, "depends on which zodiac is used.")
	}

	return b.String()
}

// ── Shared helpers ─────────────────────────────────────────────────────────

// normalizeLon normalizes a longitude to [0, 360).
func normalizeLon(lon float64) float64 {
	for lon < 0 {
		lon += 360
	}
	for lon >= 360 {
		lon -= 360
	}
	return lon
}

// angleDist returns the angular distance between two longitudes (0-180).
func angleDist(a, b float64) float64 {
	d := a - b
	if d < 0 {
		d = -d
	}
	if d > 180 {
		d = 360 - d
	}
	return d
}
