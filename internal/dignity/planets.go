package dignity

import "github.com/aj-nt/empirical/internal/swe"

// PlanetID pairs a planet name with its Swiss Ephemeris ID.
type PlanetID struct {
	Name string
	ID   int
}

// BasicPlanets is Sun through Pluto plus the Mean Node (12 bodies).
var BasicPlanets = []PlanetID{
	{"Sun", swe.SUN},
	{"Moon", swe.MOON},
	{"Mercury", swe.MERCURY},
	{"Venus", swe.VENUS},
	{"Mars", swe.MARS},
	{"Jupiter", swe.JUPITER},
	{"Saturn", swe.SATURN},
	{"Uranus", swe.URANUS},
	{"Neptune", swe.NEPTUNE},
	{"Pluto", swe.PLUTO},
	{"Node", swe.MEAN_NODE},
}

// AllPlanets is BasicPlanets plus asteroids and TNPs (24 bodies).
var AllPlanets []PlanetID

// ElectionalPlanets is BasicPlanets plus Chiron and Lilith (14 bodies).
var ElectionalPlanets []PlanetID

// AllPlanetNames is the name list for AllPlanets (24 bodies including TNPs).
var AllPlanetNames []string

// NonTNPPlanetNames is AllPlanets minus TNPs (18 bodies: Sun-Pluto+Node+asteroids+Chiron+Lilith).
var NonTNPPlanetNames []string

// NonTNPNoNodePlanetNames is NonTNPPlanetNames minus Node (17 bodies).
var NonTNPNoNodePlanetNames []string

func init() {
	AllPlanets = make([]PlanetID, 0, 24)
	AllPlanets = append(AllPlanets, BasicPlanets...)
	AllPlanets = append(AllPlanets,
		PlanetID{"Ceres", swe.CERES},
		PlanetID{"Pallas", swe.PALLAS},
		PlanetID{"Juno", swe.JUNO},
		PlanetID{"Vesta", swe.VESTA},
		PlanetID{"Lilith", swe.MEAN_APOG},
		PlanetID{"Chiron", swe.CHIRON},
		PlanetID{"Cupido", swe.CUPIDO},
		PlanetID{"Hades", swe.HADES},
		PlanetID{"Zeus", swe.ZEUS},
		PlanetID{"Kronos", swe.KRONOS},
		PlanetID{"Apollon", swe.APOLLON},
		PlanetID{"Admetos", swe.ADMETOS},
		PlanetID{"Poseidon", swe.POSEIDON},
		PlanetID{"Vulkanus", swe.VULKANUS},
	)

	ElectionalPlanets = make([]PlanetID, 0, 14)
	ElectionalPlanets = append(ElectionalPlanets, BasicPlanets...)
	ElectionalPlanets = append(ElectionalPlanets,
		PlanetID{"Chiron", swe.CHIRON},
		PlanetID{"Lilith", swe.MEAN_APOG},
	)

	AllPlanetNames = planetNames(AllPlanets)
	NonTNPPlanetNames = planetNames(AllPlanets[:18])

	NonTNPNoNodePlanetNames = make([]string, 0, 17)
	for _, p := range AllPlanets[:18] {
		if p.Name != "Node" {
			NonTNPNoNodePlanetNames = append(NonTNPNoNodePlanetNames, p.Name)
		}
	}
}

// planetNames extracts the Name field from a PlanetID slice.
func planetNames(planets []PlanetID) []string {
	names := make([]string, len(planets))
	for i, p := range planets {
		names[i] = p.Name
	}
	return names
}
