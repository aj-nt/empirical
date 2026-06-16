package dignity

import (
	"math"
	"sort"
)

// ═══════════════════════════════════════════════════════════════════════
// Astrocartography
// ═══════════════════════════════════════════════════════════════════════
//
// Planetary lines on a world map showing where each planet sits on the
// ASC, DSC, MC, or IC at a given moment.
//
// MC/IC lines are vertical (constant longitude). ASC/DSC lines are curved
// and require SWE houses for accurate computation — those live in
// cmd/recover/main.go as computeAstroCartography.

// GeoPoint is a geographic coordinate.
type GeoPoint struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

// AstroLine is a planetary line on the world map.
type AstroLine struct {
	Planet string     `json:"planet"`
	Angle  string     `json:"angle"`
	Points []GeoPoint `json:"points"`
}

// LineHit records a nearby astrocartography line for a location.
type LineHit struct {
	Planet string  `json:"planet"`
	Angle  string  `json:"angle"`
	Orb    float64 `json:"orb"`
}

// ── Core astronomy ─────────────────────────────────────────────────────

// ComputeGMST returns Greenwich Mean Sidereal Time in degrees.
func ComputeGMST(jd float64) float64 {
	d := jd - 2451545.0
	gmst := 280.46061837 + 360.98564736629*d
	return normalizeLon(gmst)
}

// ComputeMCLine returns the MC line for a planet (vertical, constant longitude).
func ComputeMCLine(planetRA, gmst float64, latStep float64) []GeoPoint {
	lon := normalizeGeo(planetRA - gmst)
	var points []GeoPoint
	for lat := -80.0; lat <= 80.0; lat += latStep {
		points = append(points, GeoPoint{Lat: lat, Lon: lon})
	}
	return points
}

// ComputeICLine returns the IC line (MC + 180).
func ComputeICLine(planetRA, gmst float64, latStep float64) []GeoPoint {
	lon := normalizeGeo(planetRA - gmst + 180)
	var points []GeoPoint
	for lat := -80.0; lat <= 80.0; lat += latStep {
		points = append(points, GeoPoint{Lat: lat, Lon: lon})
	}
	return points
}

// ── Location queries ───────────────────────────────────────────────────

// LinesNear returns astrocartography lines within orb degrees of a location.
func LinesNear(lat, lon float64, lines []AstroLine, orb float64) []LineHit {
	var hits []LineHit
	for _, line := range lines {
		closest := findClosestByLat(line.Points, lat)
		if closest == nil {
			continue
		}
		if math.Abs(closest.Lat-lat) > 1.5 {
			continue
		}
		lonDiff := math.Abs(normalizeGeo(closest.Lon - lon))
		if lonDiff < orb {
			hits = append(hits, LineHit{
				Planet: line.Planet,
				Angle:  line.Angle,
				Orb:    math.Round(lonDiff*100) / 100,
			})
		}
	}
	sort.Slice(hits, func(i, j int) bool { return hits[i].Orb < hits[j].Orb })
	return hits
}

// ── Line computation (pure math — MC/IC only) ─────────────────────────

// ComputeAstroCartographyLines computes MC and IC lines for all planets.
// ASC/DSC lines require SWE houses and are computed in cmd/recover.
func ComputeAstroCartographyLines(planets map[string]float64, jd float64, latStep float64) []AstroLine {
	gmst := ComputeGMST(jd)
	var lines []AstroLine
	for planet, lon := range planets {
		ra := LonToRA(lon, ObliquityDeg)
		lines = append(lines, AstroLine{Planet: planet, Angle: "MC", Points: ComputeMCLine(ra, gmst, latStep)})
		lines = append(lines, AstroLine{Planet: planet, Angle: "IC", Points: ComputeICLine(ra, gmst, latStep)})
	}
	return lines
}

// ComputeDraconicAstroCartographyLines computes MC/IC lines in the draconic frame.
// Planet longitudes are shifted by -northNodeLon. MC/IC lines are RA-based and
// therefore identical to tropical — this function exists for API symmetry.
func ComputeDraconicAstroCartographyLines(planets map[string]float64, jd float64, northNodeLon float64, latStep float64) []AstroLine {
	dracPlanets := make(map[string]float64)
	for p, lon := range planets {
		dracPlanets[p] = normalizeLon(lon - northNodeLon)
	}
	return ComputeAstroCartographyLines(dracPlanets, jd, latStep)
}

// ComputeCrossAstroCartographyLines computes MC/IC lines in the cross frame.
// Uses tropical planet positions (MC/IC are RA-based, invariant under shift).
// ASC/DSC would use draconic positions — those are computed in cmd/recover.
func ComputeCrossAstroCartographyLines(planets map[string]float64, jd float64, northNodeLon float64, latStep float64) []AstroLine {
	// Cross MC/IC = tropical MC/IC (RA-based, no shift)
	return ComputeAstroCartographyLines(planets, jd, latStep)
}

// ── Three-way comparison ───────────────────────────────────────────────

// ThreeWayHit records lines near a location across all three frames.
type ThreeWayHit struct {
	Planet   string   `json:"planet"`
	Angle    string   `json:"angle"`
	Tropical *LineHit `json:"tropical,omitempty"`
	Draconic *LineHit `json:"draconic,omitempty"`
	Cross    *LineHit `json:"cross,omitempty"`
}

// CompareLinesNear finds lines near a location in all three frames.
func CompareLinesNear(lat, lon float64, tropical, draconic, cross []AstroLine, orb float64) []ThreeWayHit {
	tHits := LinesNear(lat, lon, tropical, orb)
	dHits := LinesNear(lat, lon, draconic, orb)
	cHits := LinesNear(lat, lon, cross, orb)

	// Index by planet+angle
	type key struct{ planet, angle string }
	merged := make(map[key]*ThreeWayHit)

	for _, h := range tHits {
		k := key{h.Planet, h.Angle}
		merged[k] = &ThreeWayHit{Planet: h.Planet, Angle: h.Angle, Tropical: &h}
	}
	for _, h := range dHits {
		k := key{h.Planet, h.Angle}
		if m, ok := merged[k]; ok {
			m.Draconic = &h
		} else {
			merged[k] = &ThreeWayHit{Planet: h.Planet, Angle: h.Angle, Draconic: &h}
		}
	}
	for _, h := range cHits {
		k := key{h.Planet, h.Angle}
		if m, ok := merged[k]; ok {
			m.Cross = &h
		} else {
			merged[k] = &ThreeWayHit{Planet: h.Planet, Angle: h.Angle, Cross: &h}
		}
	}

	var result []ThreeWayHit
	for _, v := range merged {
		result = append(result, *v)
	}
	sort.Slice(result, func(i, j int) bool {
		// Sort by best orb across all frames
		bestI := minOrb(result[i])
		bestJ := minOrb(result[j])
		return bestI < bestJ
	})
	return result
}

func minOrb(h ThreeWayHit) float64 {
	best := 999.0
	if h.Tropical != nil && h.Tropical.Orb < best {
		best = h.Tropical.Orb
	}
	if h.Draconic != nil && h.Draconic.Orb < best {
		best = h.Draconic.Orb
	}
	if h.Cross != nil && h.Cross.Orb < best {
		best = h.Cross.Orb
	}
	return best
}

func findClosestByLat(points []GeoPoint, targetLat float64) *GeoPoint {
	if len(points) == 0 {
		return nil
	}
	best := &points[0]
	bestDist := math.Abs(points[0].Lat - targetLat)
	for i := 1; i < len(points); i++ {
		d := math.Abs(points[i].Lat - targetLat)
		if d < bestDist {
			bestDist = d
			best = &points[i]
		}
	}
	return best
}

// ── Helpers ────────────────────────────────────────────────────────────

// NormalizeGeo normalizes a geographic longitude to [-180, 180].
func NormalizeGeo(lon float64) float64 {
	lon = math.Mod(lon, 360)
	if lon > 180 {
		lon -= 360
	}
	if lon < -180 {
		lon += 360
	}
	return lon
}

// normalizeGeo is the unexported version for internal use.
func normalizeGeo(lon float64) float64 {
	return NormalizeGeo(lon)
}
