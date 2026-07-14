package mundane

import (
	"fmt"
	"math"
	"time"

	"github.com/aj-nt/empirical/internal/dignity"
)

type ChartAspect struct {
	Planet1 string  `json:"planet1"`
	Planet2 string  `json:"planet2"`
	Aspect  string  `json:"aspect"`
	Orb     float64 `json:"orb"`
}

// defaultChartAspects returns the standard Ptolemaic aspects for chart analysis.
func defaultChartAspects() []dignity.AspectDef {
	return dignity.DefaultAspects()
}

// ChartAspects computes all aspects between planets in a MundaneChart.
// orbDeg is the maximum orb in degrees.
func ChartAspects(chart *MundaneChart, orbDeg float64) []ChartAspect {
	aspects := defaultChartAspects()
	planetNames := make([]string, 0, len(chart.Planets))
	for name := range chart.Planets {
		planetNames = append(planetNames, name)
	}

	var result []ChartAspect
	for i := 0; i < len(planetNames); i++ {
		for j := i + 1; j < len(planetNames); j++ {
			p1, p2 := planetNames[i], planetNames[j]
			dist := dignity.AngleDist(chart.Planets[p1], chart.Planets[p2])
			for _, asp := range aspects {
				diff := dist - asp.Angle
				if diff < 0 {
					diff = -diff
				}
				if diff <= orbDeg {
					result = append(result, ChartAspect{
						Planet1: p1,
						Planet2: p2,
						Aspect:  asp.Name,
						Orb:     math.Round(diff*100) / 100,
					})
				}
			}
		}
	}
	return result
}



// NationTransits computes transits from current planets to a nation's natal
// chart over the given date range. Returns compacted transit hits (date ranges
// with closest orb). Uses the standard Ptolemaic aspects.
func NationTransits(nationName, startDate, endDate string, orbDeg float64, compute ComputeFunc, houses HousesFunc) ([]struct {
	TransitPlanet string
	NatalPlanet   string
	Aspect        string
	MinOrb        float64
	DateStart     string
	DateEnd       string
}, error) {
	entry, ok := NationalChart(nationName)
	if !ok {
		return nil, fmt.Errorf("unknown nation: %s", nationName)
	}

	// Cast the nation's natal chart
	natalTime := time.Date(entry.Year, time.Month(entry.Month), entry.Day,
		int(entry.Hour), int((entry.Hour-float64(int(entry.Hour)))*60), 0, 0, time.UTC)
	natalChart, err := CastChart(natalTime, entry.Lat, entry.Lon, compute, houses, 'W')
	if err != nil {
		return nil, fmt.Errorf("casting natal chart for %s: %w", nationName, err)
	}

	// Build natal planet map (name -> longitude)
	natalLongs := make(map[string]float64)
	natalPlanets := make([]string, 0, len(natalChart.Planets))
	for name, lon := range natalChart.Planets {
		natalLongs[name] = lon
		natalPlanets = append(natalPlanets, name)
	}

	// Parse dates
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start date: %w", err)
	}
	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return nil, fmt.Errorf("invalid end date: %w", err)
	}

	// Standard Ptolemaic aspects
	aspects := dignity.DefaultAspects()

	// Transit planets: Sun through Pluto + Node
	transitPlanets := DefaultMundanePlanets()

	var hits []struct {
		Date          string
		TransitPlanet string
		NatalPlanet   string
		Aspect        string
		Orb           float64
	}

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
				dist := dignity.AngleDist(tLon, nLon)
				for _, asp := range aspects {
					diff := dist - asp.Angle
					if diff < 0 {
						diff = -diff
					}
					if diff <= orbDeg {
						hits = append(hits, struct {
							Date          string
							TransitPlanet string
							NatalPlanet   string
							Aspect        string
							Orb           float64
						}{
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

	// Compact: group sequential days of the same transit
	type key struct {
		TransitPlanet string
		NatalPlanet   string
		Aspect        string
	}
	groups := make(map[key][]struct {
		Date          string
		TransitPlanet string
		NatalPlanet   string
		Aspect        string
		Orb           float64
	})
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

	return result, nil
}

// ChartPatterns delegates to the dignity engine's pattern detection on a
// MundaneChart's planet positions. Returns the full PatternReport.
func ChartPatterns(chart *MundaneChart, orbDeg float64) *dignity.PatternReport {
	return dignity.DetectPatterns(chart.Planets, orbDeg)
}

// PlanetHouses computes which house (1-12) each planet falls in.
// For Whole Sign houses, this is determined by the sign offset from ASC.
// For other house systems, planets are placed between cusps.
func PlanetHouses(chart *MundaneChart) map[string]int {
	result := make(map[string]int, len(chart.Planets))

	// Whole Sign: house = ((planet_sign_index - asc_sign_index + 12) % 12) + 1
	ascSign := int(chart.ASC / 30)

	for planet, lon := range chart.Planets {
		planetSign := int(lon / 30)
		house := ((planetSign - ascSign + 12) % 12) + 1
		result[planet] = house
	}

	return result
}

// InterpretMundaneChart produces a full chart interpretation from a MundaneChart,
// bridging to the dignity engine's InterpretChart.
func InterpretMundaneChart(name string, chart *MundaneChart, orbDeg float64) *dignity.ChartInterpretation {
	// Planet → house
	houses := PlanetHouses(chart)

	// Aspects: convert ChartAspect → dignity.AspectHit
	chartAspects := ChartAspects(chart, orbDeg)
	aspectHits := make([]dignity.AspectHit, len(chartAspects))
	for i, a := range chartAspects {
		aspectHits[i] = dignity.AspectHit{
			Planet1: a.Planet1,
			Planet2: a.Planet2,
			Aspect:  a.Aspect,
			Orb:     a.Orb,
		}
	}

	// Patterns: convert Pattern → dignity.PatternHit
	report := ChartPatterns(chart, orbDeg)
	patternHits := make([]dignity.PatternHit, len(report.Patterns))
	for i, p := range report.Patterns {
		patternHits[i] = dignity.PatternHit{
			Name:    p.Name,
			Planets: p.Planets,
		}
	}

	return dignity.InterpretChart(name, chart.Planets, houses, aspectHits, patternHits, nil)
}
