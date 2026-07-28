package dignity

import (
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/aj-nt/empirical/internal/swe"
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

// ── Transit Midpoints ─────────────────────────────────────────────────────

// TransitMidpointHit records a transiting planet conjunct a natal midpoint.
type TransitMidpointHit struct {
	Date          string  `json:"date"`
	TransitPlanet string  `json:"transit_planet"`
	NatalPairA    string  `json:"natal_pair_a"`
	NatalPairB    string  `json:"natal_pair_b"`
	Orb           float64 `json:"orb"`
}

// FindTransitMidpoints finds transiting planets conjunct natal midpoints
// over a date range. For each day, computes transiting planet positions
// and checks if any transiting planet sits at the midpoint of two natal
// planets within maxOrb degrees.
//
// natalLongs: natal planet name → ecliptic longitude
// transitPlanets: which transiting planets to check (name → SWE ID)
// startDate, endDate: range as "YYYY-MM-DD"
// maxOrb: max orb in degrees
// compute: function to get daily planet positions
func FindTransitMidpoints(
	natalLongs map[string]float64,
	transitPlanets []planetSpec,
	startDate, endDate string,
	maxOrb float64,
	compute ComputeFunc,
) ([]TransitMidpointHit, error) {
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start date: %w", err)
	}
	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return nil, fmt.Errorf("invalid end date: %w", err)
	}

	// Precompute all natal midpoints
	type natalMidpoint struct {
		pairA string
		pairB string
		deg   float64
	}
	natalNames := make([]string, 0, len(natalLongs))
	for k := range natalLongs {
		natalNames = append(natalNames, k)
	}
	sort.Strings(natalNames)

	var natalMPs []natalMidpoint
	for i, a := range natalNames {
		for _, b := range natalNames[i+1:] {
			mp := Midpoint(natalLongs[a], natalLongs[b])
			natalMPs = append(natalMPs, natalMidpoint{a, b, mp})
		}
	}

	var hits []TransitMidpointHit
	current := start
	for !current.After(end) {
		y, m, d := current.Year(), int(current.Month()), current.Day()

		for _, tp := range transitPlanets {
			var tLon float64
			if tp.ID == sentinelSouthNode {
				nnLon, _, _, _ := compute(y, m, d, 12.0, swe.MEAN_NODE)
				tLon = nnLon + 180
				if tLon >= 360 {
					tLon -= 360
				}
			} else {
				tLon, _, _, _ = compute(y, m, d, 12.0, tp.ID)
			}

			for _, nmp := range natalMPs {
				orb := angularDistance(tLon, nmp.deg)
				if orb <= maxOrb {
					hits = append(hits, TransitMidpointHit{
						Date:          current.Format("2006-01-02"),
						TransitPlanet: tp.Name,
						NatalPairA:    nmp.pairA,
						NatalPairB:    nmp.pairB,
						Orb:           math.Round(orb*100) / 100,
					})
				}
			}
		}
		current = current.AddDate(0, 0, 1)
	}

	return hits, nil
}
