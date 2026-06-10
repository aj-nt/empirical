package dignity

import "math"

// ── Synastry Engine ──────────────────────────────────────────────────────

// SynastryHit records an inter-aspect between two charts.
type SynastryHit struct {
	Planet1 string  `json:"planet1"`
	Planet2 string  `json:"planet2"`
	Aspect  string  `json:"aspect"`
	Orb     float64 `json:"orb"`
}

// ComputeSynastry computes inter-aspects between two natal charts.
//
// chart1: person 1's planet → tropical longitude (Planet1 in output)
// chart2: person 2's planet → tropical longitude (Planet2 in output)
// planets: which planets to check (must exist in both charts)
// aspects: which aspects to detect
// orbDeg: max orb in degrees
func ComputeSynastry(
	chart1, chart2 map[string]float64,
	planets []string,
	aspects []AspectDef,
	orbDeg float64,
) []SynastryHit {
	var hits []SynastryHit

	for _, p1 := range planets {
		lon1, ok1 := chart1[p1]
		if !ok1 {
			continue
		}
		for _, p2 := range planets {
			lon2, ok2 := chart2[p2]
			if !ok2 {
				continue
			}
			dist := angleDist(lon1, lon2)
			for _, asp := range aspects {
				diff := math.Abs(dist - asp.Angle)
				if diff <= orbDeg {
					hits = append(hits, SynastryHit{
						Planet1: p1,
						Planet2: p2,
						Aspect:  asp.Name,
						Orb:     math.Round(diff*100) / 100,
					})
				}
			}
		}
	}
	return hits
}
