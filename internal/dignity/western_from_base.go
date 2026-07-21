package dignity

import (
	"fmt"
	"sort"
	"strings"
)

// WesternFromBase produces a full modern Western chart interpretation from a BaseChart.
// When reading is true, additional reading-optimized fields are computed
// (chart ruler, final dispositor, weighted aspects, key midpoints, key star aspects,
// angular planets).
func WesternFromBase(bc *BaseChart, orbDeg float64, reading bool) *ChartInterpretation {
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

	// Star conjunctions (2° orb — system-specific, now explicit)
	starConjunctions := FindStarConjunctions(bc.StarPositions, planetLons, 2.0)
	for _, sc := range starConjunctions {
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
	wsCusps := make([]float64, 13)
	for i := 1; i <= 12; i++ {
		wsCusps[i] = float64((int(bc.ASC/30)+i-1)%12) * 30.0
	}
	report.RulershipChains = ComputeRulershipChains(wsCusps, planetLons, bc.ASC)

	// Dispositor trees
	report.DispositorTrees = ComputeDispositorTrees(planetLons)

	// ── Reading-optimized fields ──────────────────────────────────────
	if reading {
		report.ChartRuler, report.ChartRulerTraditional,
			report.ChartRulerHouse, report.ChartRulerSign,
			report.ChartRulerDignity = computeChartRuler(bc, planetLons, houses)

		report.FinalDispositor, report.FinalDispositorTraditional = computeFinalDispositor(planetLons)

		report.WeightedAspects = computeWeightedAspects(aspects, orbDeg)

		report.KeyMidpoints = filterKeyMidpoints(report.Midpoints, planetLons)

		// Compute star aspects inline (2° orb — system-specific)
		var starAspects []StarAspectHit
		for starName, starLon := range bc.StarPositions {
			hits := FindStarAspects(starLon, starName, planetLons, DefaultAspects(), 2.0)
			starAspects = append(starAspects, hits...)
		}
		report.KeyStarAspects = filterKeyStarAspects(starAspects, planetLons)

		report.AngularPlanets = extractAngularPlanets(houses)
	}

	return report
}

// ── Chart ruler ───────────────────────────────────────────────────────

func computeChartRuler(bc *BaseChart, planetLons map[string]float64, houses map[string]int) (modern, traditional string, house int, sign, dignity string) {
	ascSign := SignForLongitude(bc.ASC)
	modern = SignRuler(ascSign)
	traditional = SignRulerTraditional(ascSign)

	if _, ok := planetLons[modern]; ok {
		sign = SignForLongitude(planetLons[modern])
		house = houses[modern]
		dignity = planetDignity(modern, sign)
	}
	return
}

// ── Final dispositor ──────────────────────────────────────────────────

func computeFinalDispositor(planetLons map[string]float64) (modern, traditional string) {
	modern = findFinalDispositor(planetLons, SignRuler)
	traditional = findFinalDispositor(planetLons, SignRulerTraditional)
	return
}

func findFinalDispositor(planetLons map[string]float64, rulerFn func(string) string) string {
	terminals := make(map[string]bool)
	for planet, lon := range planetLons {
		visited := make(map[string]bool)
		current := planet
		currentLon := lon
		for !visited[current] {
			visited[current] = true
			sign := SignForLongitude(currentLon)
			ruler := rulerFn(sign)
			if ruler == "" {
				break
			}
			if ruler == current {
				terminals[current] = true
				break
			}
			rulerLon, ok := planetLons[ruler]
			if !ok {
				break
			}
			current = ruler
			currentLon = rulerLon
		}
	}
	if len(terminals) == 1 {
		for p := range terminals {
			return p
		}
	}
	return ""
}

// ── Weighted aspects ─────────────────────────────────────────────────

const (
	// Aspect weights
	weightConjunction    = 10.0
	weightOpposition     = 9.0
	weightSquare         = 7.0
	weightTrine          = 5.0
	weightSextile        = 4.0
	weightQuincunx       = 3.0
	weightSemiSquare     = 2.0
	weightSesquiquadrate = 2.0
	weightSemiSextile    = 1.0

	// Planet weights
	planetWeightSun     = 10.0
	planetWeightMoon    = 10.0
	planetWeightMercury = 8.0
	planetWeightVenus   = 8.0
	planetWeightMars    = 8.0
	planetWeightJupiter = 7.0
	planetWeightSaturn  = 7.0
	planetWeightUranus  = 5.0
	planetWeightNeptune = 5.0
	planetWeightPluto   = 5.0
	planetWeightNode    = 4.0
	planetWeightChiron  = 4.0
	planetWeightLilith  = 4.0
	planetWeightAsteroid = 3.0
	planetWeightDwarf    = 3.0
	planetWeightTNP      = 1.0
)

func aspectWeight(aspect string) float64 {
	switch aspect {
	case "conjunction":
		return weightConjunction
	case "opposition":
		return weightOpposition
	case "square":
		return weightSquare
	case "trine":
		return weightTrine
	case "sextile":
		return weightSextile
	case "quincunx":
		return weightQuincunx
	case "semi-square":
		return weightSemiSquare
	case "sesquiquadrate":
		return weightSesquiquadrate
	case "semi-sextile":
		return weightSemiSextile
	}
	return 1.0
}

func planetImportance(name string) float64 {
	switch name {
	case "Sun":
		return planetWeightSun
	case "Moon":
		return planetWeightMoon
	case "Mercury":
		return planetWeightMercury
	case "Venus":
		return planetWeightVenus
	case "Mars":
		return planetWeightMars
	case "Jupiter":
		return planetWeightJupiter
	case "Saturn":
		return planetWeightSaturn
	case "Uranus":
		return planetWeightUranus
	case "Neptune":
		return planetWeightNeptune
	case "Pluto":
		return planetWeightPluto
	case "Node", "North Node", "TrueNode":
		return planetWeightNode
	case "Chiron":
		return planetWeightChiron
	case "Lilith":
		return planetWeightLilith
	case "Ceres", "Pallas", "Juno", "Vesta":
		return planetWeightAsteroid
	case "Eris", "Makemake", "Gonggong":
		return planetWeightDwarf
	}
	return planetWeightTNP
}

func computeWeightedAspects(aspects []AspectHit, maxOrb float64) []WeightedAspect {
	if maxOrb <= 0 {
		maxOrb = 5.0
	}
	result := make([]WeightedAspect, 0, len(aspects))
	for _, a := range aspects {
		aw := aspectWeight(a.Aspect)
		pw := (planetImportance(a.Planet1) + planetImportance(a.Planet2)) / 2.0
		orbFactor := 1.0 - a.Orb/maxOrb
		if orbFactor < 0 {
			orbFactor = 0
		}
		result = append(result, WeightedAspect{
			Planet1: a.Planet1,
			Planet2: a.Planet2,
			Aspect:  a.Aspect,
			Orb:     a.Orb,
			Weight:  orbFactor * aw * pw,
		})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Weight > result[j].Weight
	})
	return result
}

// ── Key midpoints ─────────────────────────────────────────────────────

func filterKeyMidpoints(midpoints []string, planetLons map[string]float64) []string {
	personal := map[string]bool{
		"Sun": true, "Moon": true, "Mercury": true, "Venus": true, "Mars": true,
	}
	result := make([]string, 0)
	for _, m := range midpoints {
		// Parse midpoint string like "Ceres/Jupiter = Mars (orb 0.08°)"
		// Split on " = " then parse the left side
		parts := strings.SplitN(m, " = ", 2)
		if len(parts) != 2 {
			continue
		}
		leftParts := strings.SplitN(parts[0], "/", 2)
		if len(leftParts) != 2 {
			continue
		}
		pairA, pairB := leftParts[0], leftParts[1]
		// Parse planet and orb from right side: "Mars (orb 0.08°)"
		var planet string
		var orb float64
		if _, err := fmt.Sscanf(parts[1], "%s (orb %f°)", &planet, &orb); err != nil {
			continue
		}
		if orb > 0.5 {
			continue
		}
		if personal[pairA] || personal[pairB] || personal[planet] {
			result = append(result, m)
		}
	}
	return result
}

// ── Key star aspects ──────────────────────────────────────────────────

func filterKeyStarAspects(starAspects []StarAspectHit, planetLons map[string]float64) []string {
	personal := map[string]bool{
		"Sun": true, "Moon": true, "Mercury": true, "Venus": true, "Mars": true,
	}
	result := make([]string, 0)
	for _, sa := range starAspects {
		if sa.Orb > 1.0 {
			continue
		}
		if personal[sa.Planet] {
			result = append(result, fmt.Sprintf("%s %s %s (orb %.2f°)",
				sa.Star, sa.Aspect, sa.Planet, sa.Orb))
		}
	}
	return result
}

// ── Angular planets ───────────────────────────────────────────────────

func extractAngularPlanets(houses map[string]int) []string {
	angular := map[int]bool{1: true, 4: true, 7: true, 10: true}
	var result []string
	for planet, house := range houses {
		if angular[house] {
			result = append(result, planet)
		}
	}
	return result
}

// ── Planet dignity ────────────────────────────────────────────────────

func planetDignity(planet, sign string) string {
	if strings.Contains(domicile[planet], sign) {
		return "domicile"
	}
	if strings.Contains(detriment[planet], sign) {
		return "detriment"
	}
	if exaltation[planet] == sign {
		return "exaltation"
	}
	if fall[planet] == sign {
		return "fall"
	}
	return "peregrine"
}
