package dignity

// KoinéFromBase produces a full Koiné chart interpretation from a BaseChart.
// It extracts tropical positions, computes whole-sign houses from the ASC,
// finds natal aspects and patterns, and runs the Hellenistic interpretation engine.
// orbDeg controls the aspect orb tolerance (default 5.0).
func KoinéFromBase(bc *BaseChart, orbDeg float64) *ChartInterpretation {
	if orbDeg <= 0 {
		orbDeg = 5.0
	}

	// Extract tropical longitudes
	planetLons := TropicalToLonMap(bc.Tropical)

	// Whole-sign houses from ASC
	houses := make(map[string]int)
	for planet, lon := range planetLons {
		house := ((int(lon/30) - int(bc.ASC/30) + 12) % 12) + 1
		houses[planet] = house
	}

	// Natal aspects
	aspects := FindNatalAspects(planetLons, DefaultAspects(), orbDeg)

	// Patterns — filter to classical planets only
	classicalLons := make(map[string]float64)
	for _, cp := range ClassicalPlanets {
		if pos, ok := bc.Tropical[cp]; ok {
			classicalLons[cp] = pos.Lon
		}
	}
	patternReport := DetectPatterns(classicalLons, orbDeg)
	var patternHits []PatternHit
	for _, p := range patternReport.Patterns {
		patternHits = append(patternHits, PatternHit{
			Name:    p.Name,
			Planets: p.Planets,
		})
	}

	// Determine day/night sect: Sun above horizon = day
	sunLon := bc.Tropical["Sun"].Lon
	diff := sunLon - bc.ASC
	if diff < 0 {
		diff += 360
	}
	isDay := diff < 180

	return KoineInterpretChart(bc.Name, planetLons, houses, aspects, patternHits, isDay)
}
