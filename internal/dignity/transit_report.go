package dignity

import (
	"fmt"
	"math"
)

// ── TransitReport: system-agnostic transit data for HTML templates ─────────

// TransitReport holds all computed transit data for a single moment.
// Each system's template renders this differently.
type TransitReport struct {
	Name        string
	TransitDate string // "YYYY-MM-DD"
	IsDay       bool   // true = day sect at transit moment

	// Natal positions (tropical longitudes, classical planets only for Koiné)
	NatalLons map[string]float64

	// Transit positions (tropical longitudes)
	TransitLons map[string]float64

	// Transit-to-natal aspects
	Aspects []TransitHit

	// Transit house overlays: planet → house number (whole-sign from transit ASC)
	HouseOverlays map[string]int
}

// ComputeTransitReport computes a TransitReport from a TransitChart.
// It computes transit-to-natal aspects and house overlays for the given
// planet set and aspect definitions.
func ComputeTransitReport(tc *TransitChart, planets []string, aspects []AspectDef, orbDeg float64) *TransitReport {
	// Natal longitudes (tropical)
	natalLons := make(map[string]float64)
	for _, p := range planets {
		if pos, ok := tc.Natal.Tropical[p]; ok {
			natalLons[p] = pos.Lon
		}
	}

	// Transit longitudes (tropical)
	transitLons := make(map[string]float64)
	for _, p := range planets {
		if pos, ok := tc.TransitTropical[p]; ok {
			transitLons[p] = pos.Lon
		}
	}

	// Compute transit-to-natal aspects
	var hits []TransitHit
	for _, tp := range planets {
		tLon, ok := transitLons[tp]
		if !ok {
			continue
		}
		for _, np := range planets {
			nLon, ok := natalLons[np]
			if !ok {
				continue
			}
			dist := angleDist(tLon, nLon)
			for _, asp := range aspects {
				diff := math.Abs(dist - asp.Angle)
				if diff <= orbDeg {
					hits = append(hits, TransitHit{
						TransitPlanet: tp,
						NatalPlanet:   np,
						Aspect:        asp.Name,
						Orb:           diff,
					})
				}
			}
		}
	}

	// Transit house overlays (whole-sign from transit ASC)
	houseOverlays := make(map[string]int)
	for _, p := range planets {
		if lon, ok := transitLons[p]; ok {
			house := ((int(lon/30) - int(tc.TransitASC/30) + 12) % 12) + 1
			houseOverlays[p] = house
		}
	}

	// Determine sect at transit moment
	sunLon := tc.TransitTropical["Sun"].Lon
	diff := sunLon - tc.TransitASC
	if diff < 0 {
		diff += 360
	}
	isDay := diff < 180

	dateStr := fmt.Sprintf("%04d-%02d-%02d", tc.Year, tc.Month, tc.Day)

	return &TransitReport{
		Name:          tc.Name,
		TransitDate:   dateStr,
		IsDay:         isDay,
		NatalLons:     natalLons,
		TransitLons:   transitLons,
		Aspects:       hits,
		HouseOverlays: houseOverlays,
	}
}

// ComputeTransitReportForDate computes a TransitReport for a specific date
// without requiring a pre-built TransitChart. It builds the TransitChart internally.
func ComputeTransitReportForDate(bc *BaseChart, year, month, day, hour, minute int, tzOff, lat, lng float64, planets []string, aspects []AspectDef, orbDeg float64) (*TransitReport, error) {
	tc, err := ComputeTransitChart(bc, year, month, day, hour, minute, 0, tzOff, lat, lng)
	if err != nil {
		return nil, fmt.Errorf("compute transit chart: %w", err)
	}
	return ComputeTransitReport(tc, planets, aspects, orbDeg), nil
}
