package dignity

// WesternFromBase produces a full modern Western chart interpretation from a BaseChart.
// It extracts tropical positions, computes whole-sign houses from the ASC,
// finds natal aspects and patterns (all planets), interprets star conjunctions,
// and runs the modern Western interpretation engine.
// orbDeg controls the aspect orb tolerance (default 5.0).
func WesternFromBase(bc *BaseChart, orbDeg float64) *ChartInterpretation {
	if orbDeg <= 0 {
		orbDeg = 5.0
	}

	// Extract tropical longitudes — all planets, not just classical
	planetLons := TropicalToLonMap(bc.Tropical)

	// Whole-sign houses from ASC
	houses := make(map[string]int)
	for planet, lon := range planetLons {
		house := ((int(lon/30) - int(bc.ASC/30) + 12) % 12) + 1
		houses[planet] = house
	}

	// Natal aspects — all planets
	aspects := FindNatalAspects(planetLons, DefaultAspects(), orbDeg)

	// Patterns — all planets (not just classical)
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

	return report
}
