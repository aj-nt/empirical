package dignity

import (
	"fmt"
	"math"
	"time"
)

// ── Solar Arc Directions ──────────────────────────────────────────────────
//
// Solar Arc Directions advance all planets and house cusps by the same arc
// as the secondary progressed Sun. The formula:
//
//   solarArc = progressedSunLon - natalSunLon
//   directedLon = natalLon + solarArc
//
// This is one of the oldest predictive techniques, used by Ptolemy.
// Unlike secondary progressions (which use actual planetary motion),
// solar arc directions move everything uniformly.

// SolarArcReport holds a full solar arc direction analysis.
type SolarArcReport struct {
	Name            string             `json:"name"`
	BirthDate       string             `json:"birth_date"`
	TargetDate      string             `json:"target_date"`
	Age             float64            `json:"age_years"`
	SolarArc        float64            `json:"solar_arc_deg"`
	ProgressedSunLon float64           `json:"progressed_sun_lon"`
	NatalSunLon     float64            `json:"natal_sun_lon"`
	DirectedPositions map[string]float64 `json:"directed_positions"`
	NatalPositions   map[string]float64  `json:"natal_positions"`
	Aspects         []SynastryHit      `json:"aspects"`
	TotalAspects    int                `json:"total_aspects"`
}

// ComputeSolarArc computes solar arc directions for a given birth data and target date.
// progressedSunLon is the secondary progressed Sun longitude at the target date
// (computed by the caller using day-for-a-year).
func ComputeSolarArc(
	name string,
	birthDate time.Time,
	targetDate time.Time,
	natalSunLon float64,
	progressedSunLon float64,
	natalPositions map[string]float64,
	orbDeg float64,
) *SolarArcReport {
	age := targetDate.Sub(birthDate).Hours() / (365.2425 * 24)
	solarArc := normalizeLon(progressedSunLon - natalSunLon)

	// Apply solar arc to all natal positions
	directed := make(map[string]float64)
	for p, lon := range natalPositions {
		directed[p] = normalizeLon(lon + solarArc)
	}

	// Directed-to-natal aspects
	aspects := DefaultAspects()
	planets := make([]string, 0, len(directed))
	for p := range directed {
		planets = append(planets, p)
	}

	hits := ComputeSynastry(directed, natalPositions, planets, aspects, orbDeg)

	return &SolarArcReport{
		Name:              name,
		BirthDate:         birthDate.Format("2006-01-02"),
		TargetDate:        targetDate.Format("2006-01-02"),
		Age:               math.Round(age*100) / 100,
		SolarArc:          math.Round(solarArc*100) / 100,
		ProgressedSunLon:  math.Round(progressedSunLon*100) / 100,
		NatalSunLon:       math.Round(natalSunLon*100) / 100,
		DirectedPositions: directed,
		NatalPositions:    natalPositions,
		Aspects:           hits,
		TotalAspects:      len(hits),
	}
}

// SolarArcYear computes the solar arc for a given age in years.
// This is a convenience function that doesn't require a progressed Sun calculation.
// solarArc ≈ age * (mean solar motion per day ≈ 0.9856°)
func SolarArcYear(age float64) float64 {
	return age * 0.985647
}

// FormatSolarArc returns a human-readable description of the solar arc.
func FormatSolarArc(arcDeg float64) string {
	deg := int(arcDeg)
	min := int((arcDeg - float64(deg)) * 60)
	return fmt.Sprintf("%d°%02d'", deg, min)
}
