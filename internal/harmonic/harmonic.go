package harmonic

import (
	"math"
	"sort"
)

// ── Addey-Style Harmonic Charts ─────────────────────────────────────────
//
// Harmonic astrology (John Addey) maps the ecliptic into a harmonic
// circle where the nth harmonic compresses the zodiac by factor n.
// Multiply every longitude by n, then wrap at 360.
//
// In the harmonic chart, aspects in the tropical chart become conjunctions:
//   - 5th harmonic: quintiles (72°) and biquintiles (144°) → conjunctions
//   - 7th harmonic: septiles (~51.4°) → conjunctions
//   - 9th harmonic: noviles (40°) → conjunctions
//   - 4th harmonic: squares + oppositions → conjunctions
//
// This reveals hidden patterns — planets clustering together in high
// harmonics show affinities invisible in the tropical zodiac.
//
// Different from Uranian harmonics (midpoint pictures on the 90-degree dial).

// HarmonicLongitude converts a tropical longitude to a harmonic chart position.
func HarmonicLongitude(tropicalLon float64, harmonic int) float64 {
	return math.Mod(tropicalLon*float64(harmonic), 360)
}

// HarmonicChart returns all factors in a given harmonic chart.
func HarmonicChart(planets map[string]float64, harmonic int) map[string]float64 {
	result := make(map[string]float64)
	for name, lon := range planets {
		result[name] = HarmonicLongitude(lon, harmonic)
	}
	return result
}

// HarmonicConjunction represents a conjunction in a harmonic chart.
type HarmonicConjunction struct {
	PlanetA       string  `json:"planet_a"`
	PlanetB       string  `json:"planet_b"`
	HarmonicLonA  float64 `json:"harmonic_lon_a"`
	HarmonicLonB  float64 `json:"harmonic_lon_b"`
	Orb           float64 `json:"orb"`
}

// FindHarmonicConjunctions finds conjunctions (clusters) in a harmonic chart.
// Conjunctions in harmonic charts indicate underlying minor aspects in the tropical chart.
func FindHarmonicConjunctions(planets map[string]float64, harmonic int, orb float64) []HarmonicConjunction {
	hchart := HarmonicChart(planets, harmonic)
	names := sortedKeys(hchart)

	var conjunctions []HarmonicConjunction
	for i := 0; i < len(names); i++ {
		for j := i + 1; j < len(names); j++ {
			nameA, lonA := names[i], hchart[names[i]]
			nameB, lonB := names[j], hchart[names[j]]
			diff := math.Abs(math.Mod(lonA-lonB+540, 360) - 180)
			if diff <= orb {
				conjunctions = append(conjunctions, HarmonicConjunction{
					PlanetA:      nameA,
					PlanetB:      nameB,
					HarmonicLonA: lonA,
					HarmonicLonB: lonB,
					Orb:          math.Round(diff*1e4) / 1e4,
				})
			}
		}
	}

	sort.Slice(conjunctions, func(i, j int) bool { return conjunctions[i].Orb < conjunctions[j].Orb })
	return conjunctions
}

// HarmonicAspectNames maps harmonic numbers to their traditional aspect names.
var HarmonicAspectNames = map[int]string{
	1:  "conjunction",
	2:  "opposition",
	3:  "trine",
	4:  "square",
	5:  "quintile",
	6:  "sextile",
	7:  "septile",
	8:  "semisquare",
	9:  "novile",
	10: "decile",
	11: "undecile",
	12: "semisextile",
}

// HarmonicAspectName returns the traditional name for a harmonic's aspect type.
func HarmonicAspectName(h int) string {
	if name, ok := HarmonicAspectNames[h]; ok {
		return name
	}
	return "unknown"
}

// ── Harmonic Report ─────────────────────────────────────────────────────

// HarmonicChartResult holds the result for a single harmonic.
type HarmonicChartResult struct {
	Harmonic     int                   `json:"harmonic"`
	AspectName   string                `json:"aspect_name"`
	Positions    map[string]float64    `json:"positions"`
	Conjunctions []HarmonicConjunction `json:"conjunctions"`
}

// HarmonicReport is a multi-harmonic analysis.
type HarmonicReport struct {
	Name     string                `json:"name"`
	Harmonics []HarmonicChartResult `json:"harmonics"`
}

// ComputeHarmonicReport computes harmonic charts for the key harmonics.
func ComputeHarmonicReport(name string, planets map[string]float64, harmonics []int, orb float64) HarmonicReport {
	report := HarmonicReport{Name: name, Harmonics: make([]HarmonicChartResult, 0, len(harmonics))}

	for _, h := range harmonics {
		positions := HarmonicChart(planets, h)
		conjunctions := FindHarmonicConjunctions(planets, h, orb)
		if conjunctions == nil {
			conjunctions = []HarmonicConjunction{}
		}
		report.Harmonics = append(report.Harmonics, HarmonicChartResult{
			Harmonic:     h,
			AspectName:   HarmonicAspectName(h),
			Positions:    positions,
			Conjunctions: conjunctions,
		})
	}

	return report
}

// ── Helpers ─────────────────────────────────────────────────────────────

func sortedKeys(m map[string]float64) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
