package parans

import (
	"math"
	"sort"
)

// ── Fixed Star Parans ───────────────────────────────────────────────────
//
// A paran (paranatellonta) occurs when a star and a planet are
// simultaneously on the angles (ASC, DSC, MC, IC) at the birth moment.
// This is different from a conjunction — it's about shared angular contact.
//
// Traditional parans computation finds the exact time when a star rises
// while a planet culminates, etc. This module does the practical version:
// at the birth moment, find stars and planets within orb of the angles,
// then report star-planet pairs that share an angle.

// Angle position constants
const (
	ASC = 0 // Ascendant
	MC  = 1 // Midheaven
	DSC = 2 // Descendant (ASC + 180)
	IC  = 3 // Imum Coeli (MC + 180)
)

var angleNames = []string{"ASC", "MC", "DSC", "IC"}

// ParanContact represents a star and planet sharing an angle.
type ParanContact struct {
	Star       string  `json:"star"`
	StarLon    float64 `json:"star_lon"`
	Planet     string  `json:"planet"`
	PlanetLon  float64 `json:"planet_lon"`
	Angle      string  `json:"angle"`
	AngleLon   float64 `json:"angle_lon"`
	StarOrb    float64 `json:"star_orb"`
	PlanetOrb  float64 `json:"planet_orb"`
}

// AngularContact records a body (star or planet) on an angle.
type AngularContact struct {
	Body    string  `json:"body"`
	BodyLon float64 `json:"body_lon"`
	Angle   string  `json:"angle"`
	AngleLon float64 `json:"angle_lon"`
	Orb     float64 `json:"orb"`
}

// FindParans finds star-planet pairs that share an angle at the birth moment.
// stars: map of star name → ecliptic longitude
// planets: map of planet name → ecliptic longitude
// asc: ascendant longitude
// mc: midheaven longitude
// orb: maximum orb in degrees for angle contact
func FindParans(stars, planets map[string]float64, asc, mc, orb float64) []ParanContact {
	// Compute the four angles
	angles := map[string]float64{
		"ASC": asc,
		"MC":  mc,
		"DSC": math.Mod(asc+180, 360),
		"IC":  math.Mod(mc+180, 360),
	}

	// Find stars on angles
	starsOnAngles := findAngularContacts(stars, angles, orb)

	// Find planets on angles
	planetsOnAngles := findAngularContacts(planets, angles, orb)

	// Cross-reference: star-planet pairs sharing the same angle
	var parans []ParanContact
	for _, sc := range starsOnAngles {
		for _, pc := range planetsOnAngles {
			if sc.Angle == pc.Angle {
				parans = append(parans, ParanContact{
					Star:      sc.Body,
					StarLon:   sc.BodyLon,
					Planet:    pc.Body,
					PlanetLon: pc.BodyLon,
					Angle:     sc.Angle,
					AngleLon:  sc.AngleLon,
					StarOrb:   sc.Orb,
					PlanetOrb: pc.Orb,
				})
			}
		}
	}

	// Sort by angle then star
	sort.Slice(parans, func(i, j int) bool {
		if parans[i].Angle != parans[j].Angle {
			return parans[i].Angle < parans[j].Angle
		}
		return parans[i].Star < parans[j].Star
	})

	return parans
}

// findAngularContacts finds bodies within orb of any angle.
func findAngularContacts(bodies, angles map[string]float64, orb float64) []AngularContact {
	var contacts []AngularContact
	for name, lon := range bodies {
		for angleName, angleLon := range angles {
			diff := math.Abs(math.Mod(lon-angleLon+540, 360) - 180)
			if diff <= orb {
				contacts = append(contacts, AngularContact{
					Body:     name,
					BodyLon:  lon,
					Angle:    angleName,
					AngleLon: angleLon,
					Orb:      math.Round(diff*1e4) / 1e4,
				})
			}
		}
	}
	return contacts
}

// ── Parans Report ───────────────────────────────────────────────────────

// ParansReport is the full fixed star parans analysis.
type ParansReport struct {
	Name              string          `json:"name"`
	Angles            map[string]float64 `json:"angles"`
	StarsOnAngles     []AngularContact  `json:"stars_on_angles"`
	PlanetsOnAngles   []AngularContact  `json:"planets_on_angles"`
	Parans            []ParanContact    `json:"parans"`
}

// ComputeParansReport computes the full parans analysis.
func ComputeParansReport(name string, starPositions, planetPositions map[string]float64, asc, mc, orb float64) ParansReport {
	angles := map[string]float64{
		"ASC": asc,
		"MC":  mc,
		"DSC": math.Mod(asc+180, 360),
		"IC":  math.Mod(mc+180, 360),
	}

	return ParansReport{
		Name:            name,
		Angles:          angles,
		StarsOnAngles:   findAngularContacts(starPositions, angles, orb),
		PlanetsOnAngles: findAngularContacts(planetPositions, angles, orb),
		Parans:          FindParans(starPositions, planetPositions, asc, mc, orb),
	}
}
