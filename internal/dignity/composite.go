package dignity

import "math"

// ── Composite Charts ──────────────────────────────────────────────────────
//
// A composite chart is the midpoint of two natal charts.
// For each planet pair, composite = (lonA + lonB) / 2, handling wraparound.
// Pure math — no SWE calls.

// CompositeReport holds a full composite chart analysis.
type CompositeReport struct {
	Name1    string           `json:"name1"`
	Name2    string           `json:"name2"`
	Planets  map[string]float64 `json:"planets"`
	Aspects  []SynastryHit    `json:"aspects"`
	Patterns []Pattern        `json:"patterns"`
}

// CompositeSynastryReport holds composite-to-natal aspects for both people.
type CompositeSynastryReport struct {
	Name1     string           `json:"name1"`
	Name2     string           `json:"name2"`
	Composite map[string]float64 `json:"composite"`
	ToPerson1 []SynastryHit    `json:"to_person1"`
	ToPerson2 []SynastryHit    `json:"to_person2"`
}

// ComputeComposite computes the midpoint composite of two charts.
// Both charts must have the same planet keys.
func ComputeComposite(chart1, chart2 map[string]float64) map[string]float64 {
	composite := make(map[string]float64)
	for name, lon1 := range chart1 {
		lon2, ok := chart2[name]
		if !ok {
			continue
		}
		composite[name] = midpoint(lon1, lon2)
	}
	return composite
}

// midpoint returns the midpoint of two longitudes on the shorter arc.
func midpoint(a, b float64) float64 {
	// Normalize both to [0, 360)
	a = normalizeLon(a)
	b = normalizeLon(b)

	// Find the shortest directed distance from a to b
	dist := b - a
	if dist < 0 {
		dist += 360
	}
	// If the forward distance > 180, go the other way (from b to a)
	if dist > 180 {
		dist = a - b
		if dist < 0 {
			dist += 360
		}
		// Midpoint going from b toward a
		mp := b + dist/2.0
		if mp >= 360 {
			mp -= 360
		}
		return mp
	}

	// Midpoint going from a toward b
	mp := a + dist/2.0
	if mp >= 360 {
		mp -= 360
	}
	return mp
}

// ComputeCompositeReport computes a full composite chart with aspects and patterns.
func ComputeCompositeReport(name1, name2 string, chart1, chart2 map[string]float64, orb float64) *CompositeReport {
	composite := ComputeComposite(chart1, chart2)

	// Internal aspects within the composite
	aspects := DefaultAspects()
	var hits []SynastryHit
	names := sortedKeys(composite)
	for i := 0; i < len(names); i++ {
		for j := i + 1; j < len(names); j++ {
			dist := angleDist(composite[names[i]], composite[names[j]])
			for _, a := range aspects {
				diff := math.Abs(dist - a.Angle)
				if diff <= orb {
					hits = append(hits, SynastryHit{
						Planet1: names[i],
						Planet2: names[j],
						Aspect:  a.Name,
						Orb:     math.Round(diff*100) / 100,
					})
				}
			}
		}
	}

	// Sort by orb
	for i := 0; i < len(hits); i++ {
		for j := i + 1; j < len(hits); j++ {
			if hits[j].Orb < hits[i].Orb {
				hits[i], hits[j] = hits[j], hits[i]
			}
		}
	}

	// Pattern detection (non-TNP bodies only)
	nonTNP := []string{"Sun", "Moon", "Mercury", "Venus", "Mars", "Jupiter", "Saturn", "Uranus", "Neptune", "Pluto", "Node", "Ceres", "Pallas", "Juno", "Vesta", "Lilith", "Chiron"}
	compNonTNP := make(map[string]float64)
	for _, name := range nonTNP {
		if lon, ok := composite[name]; ok {
			compNonTNP[name] = lon
		}
	}
	patternReport := DetectPatterns(compNonTNP, 5.0)

	return &CompositeReport{
		Name1:    name1,
		Name2:    name2,
		Planets:  composite,
		Aspects:  hits,
		Patterns: patternReport.Patterns,
	}
}

// ComputeCompositeSynastry computes composite-to-natal aspects for both people.
func ComputeCompositeSynastry(name1, name2 string, chart1, chart2 map[string]float64, orb float64) *CompositeSynastryReport {
	composite := ComputeComposite(chart1, chart2)

	aspects := DefaultAspects()

	// Composite to person 1
	var toP1 []SynastryHit
	for compName, compLon := range composite {
		natLon, ok := chart1[compName]
		if !ok {
			continue
		}
		dist := angleDist(compLon, natLon)
		for _, a := range aspects {
			diff := math.Abs(dist - a.Angle)
			if diff <= orb {
				toP1 = append(toP1, SynastryHit{
					Planet1: "Comp_" + compName,
					Planet2: "Natal_" + compName,
					Aspect:  a.Name,
					Orb:     math.Round(diff*100) / 100,
				})
			}
		}
	}

	// Composite to person 2
	var toP2 []SynastryHit
	for compName, compLon := range composite {
		natLon, ok := chart2[compName]
		if !ok {
			continue
		}
		dist := angleDist(compLon, natLon)
		for _, a := range aspects {
			diff := math.Abs(dist - a.Angle)
			if diff <= orb {
				toP2 = append(toP2, SynastryHit{
					Planet1: "Comp_" + compName,
					Planet2: "Natal_" + compName,
					Aspect:  a.Name,
					Orb:     math.Round(diff*100) / 100,
				})
			}
		}
	}

	return &CompositeSynastryReport{
		Name1:     name1,
		Name2:     name2,
		Composite: composite,
		ToPerson1: toP1,
		ToPerson2: toP2,
	}
}

// sortedKeys returns sorted keys from a map.
func sortedKeys(m map[string]float64) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	// Simple insertion sort for small N
	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[j] < keys[i] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}
	return keys
}
