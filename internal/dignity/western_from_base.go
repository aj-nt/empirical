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

	// Day/night sect
	sunLon := planetLons["Sun"]
	diff := sunLon - bc.ASC
	if diff < 0 {
		diff += 360
	}
	report.IsDay = diff < 180

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

	// Declination parallels & contraparallels (1° orb)
	report.Declinations, report.Contraparallels = computeDeclinationContacts(bc.Declinations, planetLons)

	// ── Traditional Western fields ────────────────────────────────────

	// Lunar phase
	lp := ComputeLunarPhase(planetLons["Sun"], planetLons["Moon"])
	report.LunarPhase = lp.Name
	report.LunarPhaseAngle = lp.Angle

	// Retrogrades
	// Build speed map from BaseChart
	speeds := make(map[string]float64, len(bc.Tropical))
	for name, pos := range bc.Tropical {
		speeds[name] = pos.Speed
	}
	for _, r := range DetectRetrogrades(speeds) {
		if r.Retrograde {
			report.Retrogrades = append(report.Retrogrades,
				fmt.Sprintf("%s Rx (%.4f°/day)", r.Planet, r.Speed))
		}
	}

	// Antiscia
	for _, a := range ComputeAntiscia(planetLons) {
		report.Antiscia = append(report.Antiscia,
			fmt.Sprintf("%s (%.2f° %s) → antiscion %.2f° %s, contra-antiscion %.2f° %s",
				a.Planet, a.Longitude, SignForLongitude(a.Longitude),
				a.Antiscion, a.AntiscionSign,
				a.ContraAntiscion, a.ContraSign))
	}

	// Antiscia contacts (natal planet conjunct another's antiscion, ≤1°)
	report.AntisciaContacts = computeAntisciaContacts(planetLons)

	// Mutual receptions (from traditional dispositor tree)
	dt := ComputeDispositorTree(planetLons)
	report.MutualReceptions = dt.MutualReceptions

	// Decans (Faces)
	for _, d := range ComputeDecans(planetLons) {
		report.Decans = append(report.Decans,
			fmt.Sprintf("%s: %s decan %d (ruler: %s)", d.Planet, d.Sign, d.Decan, d.Ruler))
	}

	// Egyptian Terms
	for _, t := range ComputeTerms(planetLons) {
		report.Terms = append(report.Terms,
			fmt.Sprintf("%s: %s term %d (ruler: %s)", t.Planet, t.Sign, t.Term, t.Ruler))
	}

	// Void of Course Moon
	voc := ComputeVOCMoon(planetLons)
	if voc.VOC {
		report.VOCMoon = fmt.Sprintf("Moon is VOID OF COURSE in %s (%.2f°). %.2f° remaining until %s.",
			voc.MoonSign, voc.MoonLon, voc.DegreesToNext, voc.NextSign)
	} else {
		report.VOCMoon = fmt.Sprintf("Moon is NOT void of course in %s. Last applying aspect: %s to %s (orb %.2f°).",
			voc.MoonSign, voc.LastAspect, voc.LastAspectTo, voc.LastAspectOrb)
	}

	// Sect
	if report.IsDay {
		report.Sect = "Day chart (diurnal) — Sun above horizon. Sun is sect light. Jupiter is the benefic of sect; Saturn is the contrary-to-sect malefic."
	} else {
		report.Sect = "Night chart (nocturnal) — Sun below horizon. Moon is sect light. Venus is the benefic of sect; Mars is the contrary-to-sect malefic."
	}

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
	case "Node", "North Node", "TrueNode", "SouthNode":
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

// ── Declination contacts ────────────────────────────────────────────────

// computeDeclinationContacts finds parallels (same hemisphere, ≤1° orb) and
// contraparallels (opposite hemisphere, ≤1° orb) from the declination map.
// Only bodies present in planetLons are included.
func computeDeclinationContacts(decls map[string]float64, planetLons map[string]float64) (parallels, contraparallels []string) {
	// Build a list of (name, declination) for bodies in planetLons
	type entry struct {
		name string
		dec  float64
	}
	var entries []entry
	for name, dec := range decls {
		if _, ok := planetLons[name]; ok {
			entries = append(entries, entry{name, dec})
		}
	}

	// Sort by absolute declination for cleaner output
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].dec < entries[j].dec
	})

	for i := 0; i < len(entries); i++ {
		for j := i + 1; j < len(entries); j++ {
			a, b := entries[i], entries[j]
			// Same hemisphere → parallel
			if (a.dec >= 0 && b.dec >= 0) || (a.dec < 0 && b.dec < 0) {
				diff := a.dec - b.dec
				if diff < 0 {
					diff = -diff
				}
				if diff > 1.0 {
					continue
				}
				hemi := "N"
				if a.dec < 0 {
					hemi = "S"
				}
				parallels = append(parallels,
					fmt.Sprintf("%s ∥ %s (orb %.2f° %s)", a.name, b.name, diff, hemi))
			} else {
				// Opposite hemisphere → contraparallel (compare absolute values)
				absA, absB := a.dec, b.dec
				if absA < 0 {
					absA = -absA
				}
				if absB < 0 {
					absB = -absB
				}
				diff := absA - absB
				if diff < 0 {
					diff = -diff
				}
				if diff > 1.0 {
					continue
				}
				contraparallels = append(contraparallels,
					fmt.Sprintf("%s ⧄ %s (orb %.2f°)", a.name, b.name, diff))
			}
		}
	}
	return
}

// ── Antiscia contacts ──────────────────────────────────────────────────

// computeAntisciaContacts finds natal planets conjunct another planet's
// antiscion or contra-antiscion point (≤1° orb).
func computeAntisciaContacts(planetLons map[string]float64) []string {
	// Build antiscia map: planet → antiscion longitude
	type antiPoint struct {
		planet   string
		anti     float64
		contra   float64
	}
	var points []antiPoint
	for name, lon := range planetLons {
		anti := normalizeLon(360 - lon)
		contra := normalizeLon(anti + 180)
		points = append(points, antiPoint{name, anti, contra})
	}

	var contacts []string
	for _, ap := range points {
		for other, otherLon := range planetLons {
			if other == ap.planet {
				continue
			}
			// Check antiscion contact
			orb := angleDist(otherLon, ap.anti)
			if orb <= 1.0 {
				contacts = append(contacts,
					fmt.Sprintf("%s conjunct %s's antiscion (orb %.2f°)", other, ap.planet, orb))
			}
			// Check contra-antiscion contact
			orb = angleDist(otherLon, ap.contra)
			if orb <= 1.0 {
				contacts = append(contacts,
					fmt.Sprintf("%s conjunct %s's contra-antiscion (orb %.2f°)", other, ap.planet, orb))
			}
		}
	}
	return contacts
}
