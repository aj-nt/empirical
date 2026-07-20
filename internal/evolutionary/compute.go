package evolutionary

import (
	"encoding/json"
	"math"
	"sort"

	"github.com/aj-nt/empirical/internal/zodiac"
)

// ── Types ────────────────────────────────────────────────────────────────

// EvolutionaryReport holds all evolutionary astrology indicators for a chart.
type EvolutionaryReport struct {
	Name            string          `json:"name"`
	Pluto           PlanetPlacement `json:"pluto"`
	NorthNode       PlanetPlacement `json:"north_node"`
	SouthNode       PlanetPlacement `json:"south_node"`
	Saturn          PlanetPlacement `json:"saturn"`
	PlutoPolarity   PlanetPlacement `json:"pluto_polarity"`
	SouthNodeRuler  RulerInfo       `json:"south_node_ruler"`
	SkippedSteps    []SkippedStep   `json:"skipped_steps"`
	Narrative       string          `json:"narrative"`
}

// PlanetPlacement holds a planet's sign and house position.
type PlanetPlacement struct {
	Planet string `json:"planet"`
	Sign   string `json:"sign"`
	House  int    `json:"house"`
	Lon    float64 `json:"lon"`
}

// SkippedStep is a planet in hard aspect to the nodal axis (unfinished business).
type SkippedStep struct {
	Planet string  `json:"planet"`
	Aspect string  `json:"aspect"`
	Orb    float64 `json:"orb"`
}

// RulerInfo holds the ruler of a sign and its placement.
type RulerInfo struct {
	Planet string `json:"planet"`
	Sign   string `json:"sign"`
	House  int    `json:"house"`
}

// ── Sign tables ───────────────────────────────────────────────────────────

var signRulers = map[string]string{
	"Aries":       "Mars",
	"Taurus":      "Venus",
	"Gemini":      "Mercury",
	"Cancer":      "Moon",
	"Leo":         "Sun",
	"Virgo":       "Mercury",
	"Libra":       "Venus",
	"Scorpio":     "Mars",
	"Sagittarius": "Jupiter",
	"Capricorn":   "Saturn",
	"Aquarius":    "Saturn",
	"Pisces":      "Jupiter",
}

var signs = zodiac.Signs

// JSON returns the report as JSON bytes.
func (r *EvolutionaryReport) JSON() ([]byte, error) {
	return json.Marshal(r)
}

// ── Compute ───────────────────────────────────────────────────────────────

// ComputeEvolutionary assembles all evolutionary astrology indicators from
// chart data. planets is a map of planet name → ecliptic longitude (0-360).
// houses is planet name → house number (1-12). nn is the North Node longitude.
// nnHouse and snHouse are the house positions of the North and South Nodes.
// ppHouse is the house position of the Pluto polarity point.
// orb is the max orb in degrees for skipped-step detection.
func ComputeEvolutionary(
	name string,
	planets map[string]float64,
	houses map[string]int,
	nn float64,
	nnHouse int,
	snHouse int,
	ppHouse int,
	orb float64,
) *EvolutionaryReport {
	r := &EvolutionaryReport{Name: name}

	// Pluto
	plutoLon := planets["Pluto"]
	r.Pluto = PlanetPlacement{
		Planet: "Pluto",
		Sign:   signForLon(plutoLon),
		House:  houses["Pluto"],
		Lon:    plutoLon,
	}

	// North Node
	r.NorthNode = PlanetPlacement{
		Planet: "North Node",
		Sign:   signForLon(nn),
		House:  nnHouse,
		Lon:    nn,
	}

	// South Node (opposite NN)
	sn := normalizeLon(nn + 180.0)
	r.SouthNode = PlanetPlacement{
		Planet: "South Node",
		Sign:   signForLon(sn),
		House:  snHouse,
		Lon:    sn,
	}

	// Saturn
	satLon := planets["Saturn"]
	r.Saturn = PlanetPlacement{
		Planet: "Saturn",
		Sign:   signForLon(satLon),
		House:  houses["Saturn"],
		Lon:    satLon,
	}

	// Pluto polarity point (opposite Pluto)
	ppLon := normalizeLon(plutoLon + 180.0)
	r.PlutoPolarity = PlanetPlacement{
		Planet: "Pluto Polarity",
		Sign:   signForLon(ppLon),
		House:  ppHouse,
		Lon:    ppLon,
	}

	// South Node ruler
	snSign := r.SouthNode.Sign
	snRuler := signRulers[snSign]
	r.SouthNodeRuler = RulerInfo{
		Planet: snRuler,
		Sign:   signForLon(planets[snRuler]),
		House:  houses[snRuler],
	}

	// Skipped steps: planets square or opposite the nodal axis
	r.SkippedSteps = findSkippedSteps(planets, nn, sn, orb)

	// Narrative
	r.Narrative = buildNarrative(r)

	return r
}

// ── Helpers ───────────────────────────────────────────────────────────────

func signForLon(lon float64) string {
	idx := int(normalizeLon(lon) / 30.0)
	if idx >= 12 {
		idx = 11
	}
	return signs[idx]
}

func normalizeLon(lon float64) float64 {
	lon = math.Mod(lon, 360.0)
	if lon < 0 {
		lon += 360.0
	}
	return lon
}

func angleDist(a, b float64) float64 {
	d := math.Abs(normalizeLon(a) - normalizeLon(b))
	if d > 180.0 {
		d = 360.0 - d
	}
	return d
}

func findSkippedSteps(planets map[string]float64, nn, sn, orb float64) []SkippedStep {
	var steps []SkippedStep

	// Planets to check: all 10 classical + outer
	checkPlanets := []string{
		"Sun", "Moon", "Mercury", "Venus", "Mars",
		"Jupiter", "Saturn", "Uranus", "Neptune", "Pluto",
	}

	for _, p := range checkPlanets {
		lon, ok := planets[p]
		if !ok {
			continue
		}

		// Check square to NN (90°)
		distNN := angleDist(lon, nn)
		if math.Abs(distNN-90.0) <= orb {
			steps = append(steps, SkippedStep{
				Planet: p,
				Aspect: "square",
				Orb:    math.Round(math.Abs(distNN-90.0)*100) / 100,
			})
			continue
		}

		// Check opposition to NN (180°)
		if math.Abs(distNN-180.0) <= orb {
			steps = append(steps, SkippedStep{
				Planet: p,
				Aspect: "opposition",
				Orb:    math.Round(math.Abs(distNN-180.0)*100) / 100,
			})
			continue
		}

		// Check square to SN (90°)
		distSN := angleDist(lon, sn)
		if math.Abs(distSN-90.0) <= orb {
			steps = append(steps, SkippedStep{
				Planet: p,
				Aspect: "square",
				Orb:    math.Round(math.Abs(distSN-90.0)*100) / 100,
			})
			continue
		}

		// Check opposition to SN (180°) — same as conjunction to NN, skip
		// (already covered by the NN checks above)
	}

	// Sort by orb
	sort.Slice(steps, func(i, j int) bool {
		return steps[i].Orb < steps[j].Orb
	})

	return steps
}


