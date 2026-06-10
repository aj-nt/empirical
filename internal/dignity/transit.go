package dignity

import (
	"fmt"
	"math"
	"time"
)

// ── Transit Engine ───────────────────────────────────────────────────────

// ComputeFunc computes a planet's ecliptic longitude for a given date and
// planet ID. Returns (longitude, latitude, distance, speed). The caller
// handles ephemeris lookups — this is injectable for testing.
type ComputeFunc func(year, month, day int, hour float64, planetID int) (lon, lat, dist, speed float64)

// AspectDef holds an aspect angle and its name.
type AspectDef struct {
	Angle float64
	Name  string
}

// HardAspectsOnly returns conjunction, square, and opposition.
func HardAspectsOnly() []AspectDef {
	return []AspectDef{
		{0, "conjunction"},
		{90, "square"},
		{180, "opposition"},
	}
}

// TransitHit records one transit-to-natal aspect at a specific date.
type TransitHit struct {
	Date          string  `json:"date"`
	TransitPlanet string  `json:"transit_planet"`
	NatalPlanet   string  `json:"natal_planet"`
	Aspect        string  `json:"aspect"`
	Orb           float64 `json:"orb"`
}

// planetSpec maps a planet name to its SWE ID for transit scanning.
type planetSpec struct {
	Name string
	ID   int
}

// DefaultTransitPlanets returns the standard planet set for transit scanning.
func DefaultTransitPlanets() []planetSpec {
	return []planetSpec{
		{"Sun", 0},
		{"Moon", 1},
		{"Mercury", 2},
		{"Venus", 3},
		{"Mars", 4},
		{"Jupiter", 5},
		{"Saturn", 6},
		{"Uranus", 7},
		{"Neptune", 8},
		{"Pluto", 9},
	}
}

// ScanTransits computes transits over a date range.
//
// Parameters:
//   - natalLongs: natal planet name → tropical ecliptic longitude
//   - natalPlanets: which natal planets to check
//   - startDate, endDate: range as "YYYY-MM-DD"
//   - aspects: which aspects to detect
//   - orbDeg: max orb in degrees
//   - compute: function to get daily planet positions (injectable for testing)
//
// Returns all transit hits.
func ScanTransits(
	natalLongs map[string]float64,
	natalPlanets []string,
	startDate, endDate string,
	aspects []AspectDef,
	orbDeg float64,
	compute ComputeFunc,
) ([]TransitHit, error) {
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start date: %w", err)
	}
	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return nil, fmt.Errorf("invalid end date: %w", err)
	}

	transitPlanets := DefaultTransitPlanets()
	var hits []TransitHit

	current := start
	for !current.After(end) {
		y, m, d := current.Year(), int(current.Month()), current.Day()

		for _, tp := range transitPlanets {
			tLon, _, _, _ := compute(y, m, d, 12.0, tp.ID)
			for _, np := range natalPlanets {
				nLon, ok := natalLongs[np]
				if !ok {
					continue
				}
				dist := angleDist(tLon, nLon)
				for _, asp := range aspects {
					diff := math.Abs(dist - asp.Angle)
					if diff <= orbDeg {
						hits = append(hits, TransitHit{
							Date:          current.Format("2006-01-02"),
							TransitPlanet: tp.Name,
							NatalPlanet:   np,
							Aspect:        asp.Name,
							Orb:           math.Round(diff*100) / 100,
						})
					}
				}
			}
		}
		current = current.AddDate(0, 0, 1)
	}

	return hits, nil
}

// CompactTransitsWithRange collapses sequential days of the same transit
// into a date range and keeps the closest orb.
func CompactTransitsWithRange(hits []TransitHit) []struct {
	TransitPlanet string
	NatalPlanet   string
	Aspect        string
	MinOrb        float64
	DateStart     string
	DateEnd       string
} {
	type key struct {
		TransitPlanet string
		NatalPlanet   string
		Aspect        string
	}
	groups := make(map[key][]TransitHit)
	for _, h := range hits {
		k := key{h.TransitPlanet, h.NatalPlanet, h.Aspect}
		groups[k] = append(groups[k], h)
	}

	var result []struct {
		TransitPlanet string
		NatalPlanet   string
		Aspect        string
		MinOrb        float64
		DateStart     string
		DateEnd       string
	}
	for k, group := range groups {
		best := group[0]
		for _, h := range group {
			if h.Orb < best.Orb {
				best = h
			}
		}
		result = append(result, struct {
			TransitPlanet string
			NatalPlanet   string
			Aspect        string
			MinOrb        float64
			DateStart     string
			DateEnd       string
		}{
			TransitPlanet: k.TransitPlanet,
			NatalPlanet:   k.NatalPlanet,
			Aspect:        k.Aspect,
			MinOrb:        best.Orb,
			DateStart:     group[0].Date,
			DateEnd:       group[len(group)-1].Date,
		})
	}
	return result
}
