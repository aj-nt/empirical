package dignity

import "math"

// ═══════════════════════════════════════════════════════════════════════
// Primary Directions
// ═══════════════════════════════════════════════════════════════════════
//
// Primary directions are the oldest predictive technique in Western
// astrology (Ptolemy, Tetrabiblos II.10). The core principle:
// 1 degree of right ascension = 1 year of life.
//
// The ASC is directed by oblique ascension (OA), which depends on
// geographic latitude. The MC is directed by right ascension (RA),
// which is latitude-independent.
//
// Formulas (for points on the ecliptic, β=0):
//   RA(λ)  = atan2(sin(λ)·cos(ε), cos(λ))
//   Dec(λ) = arcsin(sin(λ)·sin(ε))
//   OA(λ)  = RA(λ) - arcsin(tan(Dec) · tan(φ))
//
// Where ε = obliquity of the ecliptic (~23.44°), φ = geographic latitude.
//
// Direction:
//   OA_directed = OA(ASC) + age  →  λ_ASC via iterative search
//   RA_directed = RA(MC) + age   →  λ_MC via exact formula
//
// The OA→λ conversion is transcendental and uses binary search.
// This is safe for latitudes < 66° (all inhabited locations).
// Above 66°, OA can become non-monotonic and the search may fail.

const obliquityDeg = 23.439291 // Mean obliquity, J2000.0

// DirectionsHit records a directed-point-to-natal-planet aspect.
type DirectionsHit struct {
	DirectedPoint string  `json:"directed_point"`
	NatalPlanet   string  `json:"natal_planet"`
	Aspect        string  `json:"aspect"`
	Orb           float64 `json:"orb"`
}

// DirectionsReport holds the result of primary directions for a given age.
type DirectionsReport struct {
	Age         float64         `json:"age_years"`
	DirectedASC float64         `json:"directed_asc"`
	DirectedMC  float64         `json:"directed_mc"`
	ASCAspects  []DirectionsHit `json:"asc_aspects"`
	MCAspects   []DirectionsHit `json:"mc_aspects"`
}

// ObliquityDeg is the mean obliquity of the ecliptic at J2000.0.
const ObliquityDeg = obliquityDeg

// LonToRA converts ecliptic longitude to right ascension (degrees).
func LonToRA(lon, obliquity float64) float64 {
	return lonToRA(lon, obliquity)
}

// LonToDec converts ecliptic longitude to declination (degrees).
func LonToDec(lon, obliquity float64) float64 {
	return lonToDec(lon, obliquity)
}

// AngleDist returns the minimum angular distance between two longitudes (0-180).
func AngleDist(a, b float64) float64 {
	return angleDist(a, b)
}
func lonToRA(lon, obliquity float64) float64 {
	eps := obliquity * math.Pi / 180
	l := lon * math.Pi / 180
	ra := math.Atan2(math.Sin(l)*math.Cos(eps), math.Cos(l))
	return normalizeLon(ra * 180 / math.Pi)
}

// lonToDec converts ecliptic longitude to declination (degrees).
func lonToDec(lon, obliquity float64) float64 {
	eps := obliquity * math.Pi / 180
	l := lon * math.Pi / 180
	dec := math.Asin(math.Sin(l) * math.Sin(eps))
	return dec * 180 / math.Pi
}

// raToOA converts right ascension to oblique ascension (degrees).
func raToOA(ra, dec, lat float64) float64 {
	decRad := dec * math.Pi / 180
	latRad := lat * math.Pi / 180
	correction := math.Asin(math.Tan(decRad) * math.Tan(latRad))
	return ra - correction*180/math.Pi
}

// raToLon converts right ascension to ecliptic longitude (degrees).
// Exact formula: tan(λ) = tan(RA) / cos(ε), with quadrant matching.
func raToLon(ra, obliquity float64) float64 {
	eps := obliquity * math.Pi / 180
	raRad := ra * math.Pi / 180
	tanLambda := math.Tan(raRad) / math.Cos(eps)
	lambda := math.Atan(tanLambda) * 180 / math.Pi

	// Quadrant adjustment: RA and λ share the same quadrant.
	if ra >= 90 && ra < 270 {
		lambda += 180
	} else if ra >= 270 {
		lambda += 360
	}

	return normalizeLon(lambda)
}

// oaToLon converts oblique ascension to ecliptic longitude (degrees)
// via binary search. Requires latitude < 66° for monotonic OA.
func oaToLon(targetOA, lat, obliquity float64) float64 {
	target := normalizeLon(targetOA)
	lo := 0.0
	hi := 360.0

	for i := 0; i < 50; i++ {
		mid := (lo + hi) / 2
		ra := lonToRA(mid, obliquity)
		dec := lonToDec(mid, obliquity)
		oa := normalizeLon(raToOA(ra, dec, lat))

		if oa < target {
			lo = mid
		} else {
			hi = mid
		}
	}

	return (lo + hi) / 2
}

// ── Primary Directions computation ────────────────────────────────────

// ComputePrimaryDirections computes directed ASC and MC positions for a
// given age, and finds aspects to natal planets.
func ComputePrimaryDirections(
	natal map[string]float64,
	ascLon, mcLon, lat, age float64,
	aspects []AspectDef,
	orb float64,
) *DirectionsReport {
	// OA of natal ASC
	raASC := lonToRA(ascLon, obliquityDeg)
	decASC := lonToDec(ascLon, obliquityDeg)
	oaASC := raToOA(raASC, decASC, lat)

	// RA of natal MC
	raMC := lonToRA(mcLon, obliquityDeg)

	// Directed positions
	oaDirected := normalizeLon(oaASC + age)
	raDirected := normalizeLon(raMC + age)

	directedASC := oaToLon(oaDirected, lat, obliquityDeg)
	directedMC := raToLon(raDirected, obliquityDeg)

	// Find aspects: directed ASC vs natal planets
	ascAspects := findDirectionsAspects(directedASC, natal, "ASC", aspects, orb)

	// Find aspects: directed MC vs natal planets
	mcAspects := findDirectionsAspects(directedMC, natal, "MC", aspects, orb)

	return &DirectionsReport{
		Age:         age,
		DirectedASC: directedASC,
		DirectedMC:  directedMC,
		ASCAspects:  ascAspects,
		MCAspects:   mcAspects,
	}
}

// findDirectionsAspects finds aspects between a directed point and natal planets.
func findDirectionsAspects(
	directedLon float64,
	natal map[string]float64,
	pointName string,
	aspects []AspectDef,
	orb float64,
) []DirectionsHit {
	var hits []DirectionsHit
	for np, nlon := range natal {
		dist := angleDist(directedLon, nlon)
		for _, a := range aspects {
			diff := math.Abs(dist - a.Angle)
			if diff <= orb {
				hits = append(hits, DirectionsHit{
					DirectedPoint: pointName,
					NatalPlanet:   np,
					Aspect:        a.Name,
					Orb:           math.Round(diff*100) / 100,
				})
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
	return hits
}
