package declination

import (
	"math"
	"sort"
)

// ── Declination Parallels ───────────────────────────────────────────────
//
// Declination is the north-south position of a body relative to the
// celestial equator. Two bodies at the same declination are "parallel"
// — an aspect by declination rather than longitude.
//
// Contraparallel: same absolute declination but opposite hemisphere
// (one north, one south). Equivalent to an opposition in declination.
//
// These are invisible in standard longitude-based aspect analysis.
// SWE gives ecliptic latitude for free; declination is computed from
// ecliptic longitude, latitude, and the obliquity of the ecliptic.

const obliquity = 23.4392911 // Mean obliquity of the ecliptic (J2000.0)

// EclipticToDeclination converts ecliptic longitude and latitude to declination.
// lon, lat in degrees. Returns declination in degrees (positive = north).
func EclipticToDeclination(lon, lat float64) float64 {
	lonRad := lon * math.Pi / 180.0
	latRad := lat * math.Pi / 180.0
	oblRad := obliquity * math.Pi / 180.0

	sinDecl := math.Sin(latRad)*math.Cos(oblRad) + math.Cos(latRad)*math.Sin(oblRad)*math.Sin(lonRad)
	declRad := math.Asin(sinDecl)
	return declRad * 180.0 / math.Pi
}

// DeclinationData holds a body's declination info.
type DeclinationData struct {
	Body        string  `json:"body"`
	Longitude   float64 `json:"longitude"`
	Latitude    float64 `json:"latitude"`
	Declination float64 `json:"declination"`
	Hemisphere  string  `json:"hemisphere"` // "North" or "South"
}

// ParallelContact represents a declination parallel between two bodies.
type ParallelContact struct {
	BodyA       string  `json:"body_a"`
	BodyB       string  `json:"body_b"`
	DeclA       float64 `json:"decl_a"`
	DeclB       float64 `json:"decl_b"`
	Orb         float64 `json:"orb"`
	Type        string  `json:"type"` // "parallel" or "contraparallel"
}

// ComputeDeclinations computes declination for all bodies.
// positions: map of body name → {lon, lat}
func ComputeDeclinations(positions map[string][2]float64) []DeclinationData {
	var results []DeclinationData
	for name, ll := range positions {
		lon, lat := ll[0], ll[1]
		decl := EclipticToDeclination(lon, lat)
		hem := "North"
		if decl < 0 {
			hem = "South"
		}
		results = append(results, DeclinationData{
			Body:        name,
			Longitude:   lon,
			Latitude:    lat,
			Declination: math.Round(decl*1e4) / 1e4,
			Hemisphere:  hem,
		})
	}
	sort.Slice(results, func(i, j int) bool { return results[i].Body < results[j].Body })
	return results
}

// FindParallels finds all declination parallels and contraparallels.
// orb in degrees.
func FindParallels(declinations []DeclinationData, orb float64) []ParallelContact {
	var parallels []ParallelContact
	for i := 0; i < len(declinations); i++ {
		for j := i + 1; j < len(declinations); j++ {
			a, b := declinations[i], declinations[j]

			// Parallel: same hemisphere, close declination
			if a.Hemisphere == b.Hemisphere {
				diff := math.Abs(a.Declination - b.Declination)
				if diff <= orb {
					parallels = append(parallels, ParallelContact{
						BodyA: a.Body,
						BodyB: b.Body,
						DeclA: a.Declination,
						DeclB: b.Declination,
						Orb:   math.Round(diff*1e4) / 1e4,
						Type:  "parallel",
					})
				}
			}

			// Contraparallel: opposite hemisphere, close absolute declination
			if a.Hemisphere != b.Hemisphere {
				diff := math.Abs(math.Abs(a.Declination) - math.Abs(b.Declination))
				if diff <= orb {
					parallels = append(parallels, ParallelContact{
						BodyA: a.Body,
						BodyB: b.Body,
						DeclA: a.Declination,
						DeclB: b.Declination,
						Orb:   math.Round(diff*1e4) / 1e4,
						Type:  "contraparallel",
					})
				}
			}
		}
	}
	sort.Slice(parallels, func(i, j int) bool { return parallels[i].Orb < parallels[j].Orb })
	return parallels
}

// ── Declination Report ──────────────────────────────────────────────────

// DeclinationReport is the full declination analysis.
type DeclinationReport struct {
	Name        string            `json:"name"`
	Declinations []DeclinationData `json:"declinations"`
	Parallels   []ParallelContact  `json:"parallels"`
}

// ComputeDeclinationReport computes the full declination analysis.
func ComputeDeclinationReport(name string, positions map[string][2]float64, orb float64) DeclinationReport {
	decls := ComputeDeclinations(positions)
	pars := FindParallels(decls, orb)
	return DeclinationReport{
		Name:         name,
		Declinations: decls,
		Parallels:    pars,
	}
}
