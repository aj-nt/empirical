package dignity

import "fmt"

// WesternFromBase produces a full modern Western chart interpretation from a BaseChart.
// It extracts tropical positions for the standard Western planet set (classical +
// modern + asteroids + points, excluding Trans-Neptunian Points), computes whole-sign
// houses from the ASC, finds natal aspects and patterns using the nine modern Western
// aspects (Ptolemaic + semi-sextile, semi-square, sesquiquadrate, quincunx),
// interprets star conjunctions, and runs the modern Western interpretation engine.
// orbDeg controls the aspect orb tolerance (default 5.0).
func WesternFromBase(bc *BaseChart, orbDeg float64) *ChartInterpretation {
	if orbDeg <= 0 {
		orbDeg = 5.0
	}

	// Extract tropical longitudes — filter to Western planet set (no TNPs)
	allLons := TropicalToLonMap(bc.Tropical)
	planetLons := make(map[string]float64, len(allLons))
	for name, lon := range allLons {
		if !isTNP(name) {
			planetLons[name] = lon
		}
	}

	// Whole-sign houses from ASC
	houses := make(map[string]int)
	for planet, lon := range planetLons {
		house := ((int(lon/30) - int(bc.ASC/30) + 12) % 12) + 1
		houses[planet] = house
	}

	// Natal aspects — nine modern Western aspects
	aspects := FindNatalAspects(planetLons, WesternAspects(), orbDeg)

	// Patterns — Western planet set only
	patternReport := DetectPatterns(planetLons, orbDeg)
	var patternHits []PatternHit
	for _, p := range patternReport.Patterns {
		patternHits = append(patternHits, PatternHit{
			Name:    p.Name,
			Planets: p.Planets,
		})
	}

	// Run the modern Western interpretation engine
	report := InterpretChart(bc.Name, planetLons, houses, aspects, patternHits, nil)

	// Star conjunctions
	for _, sc := range bc.FixedStars {
		report.Stars = append(report.Stars, InterpretStarConjunction(sc))
	}

	// Direct midpoints (1° orb)
	midpointHits := FindDirectMidpoints(planetLons, 1.0)
	for _, mh := range midpointHits {
		report.Midpoints = append(report.Midpoints,
			fmt.Sprintf("%s/%s = %s (orb %.2f°)", mh.PairA, mh.PairB, mh.Planet, mh.Orb))
	}

	// Element & modality balance
	report.ElementBalance = ComputeElementBalance(planetLons)
	report.ModalityBalance = ComputeModalityBalance(planetLons)

	// Hemisphere emphasis
	report.Hemisphere = ComputeHemisphereEmphasis(planetLons, bc.ASC)

	// House rulership chains (whole-sign cusps from ASC)
	wsCusps := make([]float64, 13) // 1-indexed
	for i := 1; i <= 12; i++ {
		wsCusps[i] = float64((int(bc.ASC/30)+i-1)%12) * 30.0
	}
	report.RulershipChains = ComputeRulershipChains(wsCusps, planetLons, bc.ASC)

	// Dispositor trees
	report.DispositorTrees = ComputeDispositorTrees(planetLons)

	return report
}
