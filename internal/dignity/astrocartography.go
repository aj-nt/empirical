package dignity

import (
	"math"
	"sort"

	"github.com/aj-nt/empirical/internal/geocode"
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

// ── Line computation ──────────────────────────────────────────────────

// HousesFunc computes house cusps and angles for a given JD, latitude, and longitude.
// Returns (cusps [13]float64, ascmc [10]float64) where ascmc[0] = ASC.
// This is injectable for testing — production code passes swe.Houses.
type HousesFunc func(jd, lat, lon float64, hsys byte) (cusps [13]float64, ascmc [10]float64)

// ComputeASCLine returns the ASC line for a planet: geographic longitudes where
// the Ascendant equals planetLon, sampled at each latitude step.
// housesFn is injectable for testing; pass swe.Houses in production.
func ComputeASCLine(planetLon, jd float64, latStep float64, housesFn HousesFunc) []GeoPoint {
	var points []GeoPoint
	for lat := -80.0; lat <= 80.0; lat += latStep {
		lon := findASCLon(planetLon, jd, lat, housesFn)
		if lon != nil {
			points = append(points, GeoPoint{Lat: lat, Lon: *lon})
		}
	}
	return points
}

// ComputeDSCLine returns the DSC line for a planet: geographic longitudes where
// the Descendant equals planetLon. This is the ASC line for (planetLon - 180).
func ComputeDSCLine(planetLon, jd float64, latStep float64, housesFn HousesFunc) []GeoPoint {
	dscTarget := NormalizeLon(planetLon - 180)
	return ComputeASCLine(dscTarget, jd, latStep, housesFn)
}

// findASCLon binary-searches geographic longitude where ASC = targetLon.
// Uses signed angular distance to distinguish the two zero crossings
// (ASC=target vs ASC=target+180) that a simple abs-diff search conflates.
// Returns nil if no solution exists (e.g., circumpolar target at high latitude).
func findASCLon(targetLon, jd, lat float64, housesFn HousesFunc) *float64 {
	lo := -180.0
	hi := 180.0

	for i := 0; i < 80; i++ {
		mid := (lo + hi) / 2
		_, ascmc := housesFn(jd, lat, mid, 'P')
		asc := ascmc[0]

		// Signed diff: positive = ASC ahead of target
		diff := signedAngularDist(asc, targetLon)

		if diff < 1e-8 && diff > -1e-8 {
			return &mid
		}

		if diff > 0 {
			hi = mid // ASC too far ahead, move west
		} else {
			lo = mid // ASC behind, move east
		}
	}

	mid := (lo + hi) / 2
	// Validate: at circumpolar latitudes the search converges to a boundary
	// where ASC never reaches targetLon. Reject if orb > 2°.
	// At high latitudes (>50°) ASC changes rapidly with longitude;
	// the binary search may converge to ~1-2° precision even for valid targets.
	_, ascmc := housesFn(jd, lat, mid, 'P')
	if AngleDist(ascmc[0], targetLon) > 2.0 {
		return nil
	}
	return &mid
}

// signedAngularDist returns a-b normalized to [-180, 180). Positive = a ahead of b.
func signedAngularDist(a, b float64) float64 {
	d := a - b
	for d > 180 {
		d -= 360
	}
	for d < -180 {
		d += 360
	}
	return d
}

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
		h := h
		k := key{h.Planet, h.Angle}
		merged[k] = &ThreeWayHit{Planet: h.Planet, Angle: h.Angle, Tropical: &h}
	}
	for _, h := range dHits {
		h := h
		k := key{h.Planet, h.Angle}
		if m, ok := merged[k]; ok {
			m.Draconic = &h
		} else {
			merged[k] = &ThreeWayHit{Planet: h.Planet, Angle: h.Angle, Draconic: &h}
		}
	}
	for _, h := range cHits {
		h := h
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

// ── Paran intersections ─────────────────────────────────────────────────

// ParanPoint is a geographic intersection of two astrocartography lines.
type ParanPoint struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

// ParanIntersection records where two planetary lines cross on the map.
// Only MC/IC × ASC/DSC pairs can intersect — same-type lines are parallel
// (MC ∥ MC) or disjoint level sets of the same function (ASC ∥ ASC).
type ParanIntersection struct {
	Planet1         string  `json:"planet1"`
	Angle1          string  `json:"angle1"`
	Planet2         string  `json:"planet2"`
	Angle2          string  `json:"angle2"`
	Lat             float64 `json:"lat"`
	Lon             float64 `json:"lon"`
	LocationName    string  `json:"location_name,omitempty"`
	LocationCountry string  `json:"location_country,omitempty"`
}

// isInShortArc returns true if val lies on the shortest arc from a to b.
// All values are normalized to [0, 360) before comparison.
func isInShortArc(val, a, b float64) bool {
	a = normalizeLon(a)
	b = normalizeLon(b)
	val = normalizeLon(val)

	if a > b {
		a, b = b, a
	}

	dist := b - a
	if dist <= 180 {
		return val >= a && val <= b
	}
	// Short arc wraps through 0: [b, 360) ∪ [0, a]
	return val >= b || val <= a
}

// findValidLatRange finds the maximum latitude (positive and negative) where
// findASCLon returns non-nil for the given target. Starts at 0 and searches outward.
func findValidLatRange(ascTarget, jd float64, housesFn HousesFunc) (minLat, maxLat float64, ok bool) {
	// Verify lat=0 works (it always should for non-degenerate targets)
	if findASCLon(ascTarget, jd, 0, housesFn) == nil {
		return 0, 0, false
	}

	// Search positive direction
	hi := 0.0
	step := 10.0
	for step >= 0.1 {
		for {
			next := hi + step
			if next > 85 {
				step /= 2
				break
			}
			if findASCLon(ascTarget, jd, next, housesFn) == nil {
				step /= 2
				break
			}
			hi = next
		}
	}

	// Search negative direction
	lo := 0.0
	step = 10.0
	for step >= 0.1 {
		for {
			next := lo - step
			if next < -85 {
				step /= 2
				break
			}
			if findASCLon(ascTarget, jd, next, housesFn) == nil {
				step /= 2
				break
			}
			lo = next
		}
	}

	return lo, hi, true
}

// findMCxASCIntersection finds where a vertical MC line (at mcLon)
// crosses an ASC line (where ASC = ascTarget).
// Returns nil if no intersection exists.
func findMCxASCIntersection(mcLon, ascTarget, jd float64, housesFn HousesFunc) *ParanPoint {
	lo, hi, ok := findValidLatRange(ascTarget, jd, housesFn)
	if !ok {
		return nil
	}
	return findMCxASCIntersectionCached(mcLon, ascTarget, jd, lo, hi, housesFn)
}

// findMCxASCIntersectionCached is like findMCxASCIntersection but uses a
// pre-computed latitude range (from findValidLatRange) to avoid recomputation.
func findMCxASCIntersectionCached(mcLon, ascTarget, jd, lo, hi float64, housesFn HousesFunc) *ParanPoint {

	loLon := findASCLon(ascTarget, jd, lo, housesFn)
	hiLon := findASCLon(ascTarget, jd, hi, housesFn)
	if loLon == nil || hiLon == nil {
		return nil
	}

	// Check if mcLon lies on the short arc between the ASC line's endpoints.
	if !isInShortArc(mcLon, *loLon, *hiLon) {
		return nil
	}

	for i := 0; i < 60; i++ {
		mid := (lo + hi) / 2
		midLon := findASCLon(ascTarget, jd, mid, housesFn)
		if midLon == nil {
			return nil
		}

		if math.Abs(normalizeGeo(*midLon-mcLon)) < 1e-6 {
			return &ParanPoint{Lat: mid, Lon: *midLon}
		}

		if isInShortArc(mcLon, *loLon, *midLon) {
			hi = mid
			hiLon = midLon
		} else {
			lo = mid
			loLon = midLon
		}
	}

	mid := (lo + hi) / 2
	midLon := findASCLon(ascTarget, jd, mid, housesFn)
	if midLon == nil {
		return nil
	}
	return &ParanPoint{Lat: mid, Lon: *midLon}
}

// FindParans finds all MC/IC × ASC/DSC line intersections for a set of planets.
// planets: map of planet name → ecliptic longitude (used for both MC/IC and ASC/DSC)
// jd: Julian day
// gmst: Greenwich Mean Sidereal Time in degrees
// housesFn: house calculation function (swe.Houses in production)
func FindParans(planets map[string]float64, jd, gmst float64, housesFn HousesFunc) []ParanIntersection {
	return FindParansCross(planets, planets, jd, gmst, housesFn)
}

// FindParansCross finds MC/IC × ASC/DSC line intersections using separate
// position maps for the two axis families.
//
// mcPlanets: planet → ecliptic longitude for MC/IC lines (typically tropical)
// ascPlanets: planet → ecliptic longitude for ASC/DSC targets (typically draconic in cross frame)
//
// When mcPlanets == ascPlanets this is equivalent to FindParans.
func FindParansCross(mcPlanets, ascPlanets map[string]float64, jd, gmst float64, housesFn HousesFunc) []ParanIntersection {
	var intersections []ParanIntersection

	// Precompute valid latitude ranges for each ASC/DSC target.
	// findValidLatRange is expensive (~80 swe.Houses calls each) and there are
	// only 2N unique targets (ASC + DSC per planet), not N² pairs.
	type latRange struct{ lo, hi float64 }
	ascRanges := make(map[float64]latRange)
	dscRanges := make(map[float64]latRange)

	getRange := func(target float64, cache map[float64]latRange) (lo, hi float64, ok bool) {
		if r, found := cache[target]; found {
			return r.lo, r.hi, true
		}
		l, h, ok := findValidLatRange(target, jd, housesFn)
		if ok {
			cache[target] = latRange{l, h}
		}
		return l, h, ok
	}

	for p1, lon1 := range mcPlanets {
		ra1 := LonToRA(lon1, ObliquityDeg)
		mcLon := normalizeGeo(ra1 - gmst)
		icLon := normalizeGeo(ra1 - gmst + 180)

		for p2, lon2 := range ascPlanets {
			if p1 == p2 {
				continue
			}

			// MC × ASC
			if lo, hi, ok := getRange(lon2, ascRanges); ok {
				if pt := findMCxASCIntersectionCached(mcLon, lon2, jd, lo, hi, housesFn); pt != nil {
					intersections = append(intersections, ParanIntersection{
						Planet1: p1, Angle1: "MC",
						Planet2: p2, Angle2: "ASC",
						Lat: pt.Lat, Lon: pt.Lon,
					})
				}
			}
			// MC × DSC
			dscTarget := NormalizeLon(lon2 - 180)
			if lo, hi, ok := getRange(dscTarget, dscRanges); ok {
				if pt := findMCxASCIntersectionCached(mcLon, dscTarget, jd, lo, hi, housesFn); pt != nil {
					intersections = append(intersections, ParanIntersection{
						Planet1: p1, Angle1: "MC",
						Planet2: p2, Angle2: "DSC",
						Lat: pt.Lat, Lon: pt.Lon,
					})
				}
			}
			// IC × ASC
			if lo, hi, ok := getRange(lon2, ascRanges); ok {
				if pt := findMCxASCIntersectionCached(icLon, lon2, jd, lo, hi, housesFn); pt != nil {
					intersections = append(intersections, ParanIntersection{
						Planet1: p1, Angle1: "IC",
						Planet2: p2, Angle2: "ASC",
						Lat: pt.Lat, Lon: pt.Lon,
					})
				}
			}
			// IC × DSC
			if lo, hi, ok := getRange(dscTarget, dscRanges); ok {
				if pt := findMCxASCIntersectionCached(icLon, dscTarget, jd, lo, hi, housesFn); pt != nil {
					intersections = append(intersections, ParanIntersection{
						Planet1: p1, Angle1: "IC",
						Planet2: p2, Angle2: "DSC",
						Lat: pt.Lat, Lon: pt.Lon,
					})
				}
			}
		}
	}

	// Sort by latitude then longitude
	sort.Slice(intersections, func(i, j int) bool {
		if math.Abs(intersections[i].Lat-intersections[j].Lat) > 0.01 {
			return intersections[i].Lat < intersections[j].Lat
		}
		return intersections[i].Lon < intersections[j].Lon
	})

	return intersections
}

// GeocodeParans populates LocationName and LocationCountry for each paran
// intersection by finding the nearest city in the embedded GeoNames database.
// This is a best-effort operation — if the cities database fails to load,
// the fields are left empty (no error returned).
func GeocodeParans(parans []ParanIntersection) {
	cities, err := geocode.LoadCities()
	if err != nil {
		return
	}
	for i := range parans {
		if city, ok := geocode.NearestCity(parans[i].Lat, parans[i].Lon, cities); ok {
			parans[i].LocationName = city.Name
			parans[i].LocationCountry = city.Country
		}
	}
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
