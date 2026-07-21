package dignity

import "fmt"

// ── Rulership Chains & Dispositor Trees ────────────────────────────────────
//
// Pure math — no SWE calls. Traces sign→ruler→sign chains for houses
// (rulership chains) and planets (dispositor trees).

// SignRuler returns the modern Western ruler of a zodiac sign.
func SignRuler(sign string) string {
	switch sign {
	case "Aries":
		return "Mars"
	case "Taurus":
		return "Venus"
	case "Gemini":
		return "Mercury"
	case "Cancer":
		return "Moon"
	case "Leo":
		return "Sun"
	case "Virgo":
		return "Mercury"
	case "Libra":
		return "Venus"
	case "Scorpio":
		return "Pluto"
	case "Sagittarius":
		return "Jupiter"
	case "Capricorn":
		return "Saturn"
	case "Aquarius":
		return "Uranus"
	case "Pisces":
		return "Neptune"
	}
	return ""
}

// SignRulerTraditional returns the traditional (pre-modern) ruler of a zodiac sign.
// Differs from SignRuler for Scorpio (Mars), Aquarius (Saturn), and Pisces (Jupiter).
func SignRulerTraditional(sign string) string {
	switch sign {
	case "Aries":
		return "Mars"
	case "Taurus":
		return "Venus"
	case "Gemini":
		return "Mercury"
	case "Cancer":
		return "Moon"
	case "Leo":
		return "Sun"
	case "Virgo":
		return "Mercury"
	case "Libra":
		return "Venus"
	case "Scorpio":
		return "Mars"
	case "Sagittarius":
		return "Jupiter"
	case "Capricorn":
		return "Saturn"
	case "Aquarius":
		return "Saturn"
	case "Pisces":
		return "Jupiter"
	}
	return ""
}

// ComputeRulershipChains traces the house rulership chain for each house.
// For each house: cusp sign → ruler planet → house ruler is in → that house's
// cusp sign → ... until a loop is detected. Returns map[houseNumber]chain.
// houseCusps is 1-indexed (index 0 unused, indices 1-12 are cusp longitudes).
func ComputeRulershipChains(houseCusps []float64, planets map[string]float64, asc float64) map[int][]string {
	chains := make(map[int][]string)
	for h := 1; h <= 12; h++ {
		chain := traceRulershipChain(h, houseCusps, planets, asc)
		if len(chain) > 0 {
			chains[h] = chain
		}
	}
	return chains
}

// traceRulershipChain follows the rulership chain for a single house.
func traceRulershipChain(startHouse int, houseCusps []float64, planets map[string]float64, asc float64) []string {
	visited := make(map[int]bool)
	var chain []string
	house := startHouse

	for !visited[house] {
		visited[house] = true
		cuspSign := SignForLongitude(houseCusps[house])
		ruler := SignRuler(cuspSign)
		if ruler == "" {
			break
		}

		// Find which house the ruler is in
		rulerLon, ok := planets[ruler]
		if !ok {
			chain = append(chain, fmt.Sprintf("H%d %s→%s (not in chart)", house, cuspSign, ruler))
			break
		}
		rulerHouse := ((int(rulerLon/30) - int(asc/30) + 12) % 12) + 1
		rulerSign := SignForLongitude(rulerLon)

		chain = append(chain, fmt.Sprintf("H%d %s→%s in H%d %s", house, cuspSign, ruler, rulerHouse, rulerSign))
		house = rulerHouse
	}

	return chain
}

// ComputeDispositorTrees traces the dispositor chain for each planet.
// For each planet: sign → ruler of that sign → sign of that ruler → ...
// until a loop or a planet in its own sign (final dispositor).
func ComputeDispositorTrees(planets map[string]float64) map[string][]string {
	trees := make(map[string][]string)
	for planet, lon := range planets {
		tree := traceDispositor(planet, lon, planets)
		if len(tree) > 0 {
			trees[planet] = tree
		}
	}
	return trees
}

// traceDispositor follows the dispositor chain for a single planet.
func traceDispositor(startPlanet string, startLon float64, planets map[string]float64) []string {
	visited := make(map[string]bool)
	var tree []string
	planet := startPlanet
	lon := startLon

	for !visited[planet] {
		visited[planet] = true
		sign := SignForLongitude(lon)
		ruler := SignRuler(sign)
		if ruler == "" {
			break
		}

		if ruler == planet {
			// Planet in its own sign — final dispositor
			tree = append(tree, fmt.Sprintf("%s in %s (final dispositor)", planet, sign))
			break
		}

		rulerLon, ok := planets[ruler]
		if !ok {
			tree = append(tree, fmt.Sprintf("%s in %s→%s (not in chart)", planet, sign, ruler))
			break
		}
		rulerSign := SignForLongitude(rulerLon)

		tree = append(tree, fmt.Sprintf("%s in %s→%s in %s", planet, sign, ruler, rulerSign))
		planet = ruler
		lon = rulerLon
	}

	return tree
}
