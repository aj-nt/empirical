package dignity

import (
	"math"
	"sort"
)

// ── Midpoint Analysis ──────────────────────────────────────────────────────
//
// Pure math — no SWE calls. Midpoints are the half-sum of two ecliptic
// longitudes. A direct midpoint hit occurs when a third body sits at the
// midpoint of two others within a given orb.

// DirectMidpointHit represents a planet occupying a midpoint of two other objects.
type DirectMidpointHit struct {
	PairA  string  `json:"pair_a"`
	PairB  string  `json:"pair_b"`
	Planet string  `json:"planet"`
	Orb    float64 `json:"orb"`
}

// Midpoint returns the direct half-sum of two ecliptic positions.
// Handles 360° wraparound by taking the shorter arc.
func Midpoint(a, b float64) float64 {
	diff := math.Mod(b-a+360, 360)
	if diff > 180 {
		diff -= 360
	}
	return math.Mod(a+diff/2+360, 360)
}

// FindDirectMidpoints finds all direct midpoint hits in a set of objects.
// For every pair (A, B), checks if any third object C sits at the A/B midpoint
// within maxOrb degrees. Results are sorted by orb ascending.
// A planet is never a hit for a pair it's part of.
func FindDirectMidpoints(objects map[string]float64, maxOrb float64) []DirectMidpointHit {
	names := make([]string, 0, len(objects))
	for k := range objects {
		names = append(names, k)
	}
	sort.Strings(names)

	var hits []DirectMidpointHit
	for i, nameA := range names {
		for _, nameB := range names[i+1:] {
			mp := Midpoint(objects[nameA], objects[nameB])
			for _, nameC := range names {
				if nameC == nameA || nameC == nameB {
					continue
				}
				orb := angularDistance(mp, objects[nameC])
				if orb <= maxOrb {
					hits = append(hits, DirectMidpointHit{
						PairA:  nameA,
						PairB:  nameB,
						Planet: nameC,
						Orb:    math.Round(orb*1e4) / 1e4,
					})
				}
			}
		}
	}

	sort.Slice(hits, func(i, j int) bool { return hits[i].Orb < hits[j].Orb })
	return hits
}

// angularDistance returns the shortest angular distance between two longitudes.
func angularDistance(a, b float64) float64 {
	d := math.Abs(a - b)
	if d > 180 {
		d = 360 - d
	}
	return d
}
