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
	{"TrueNode", swe.TRUE_NODE},
}

// AllPlanets is BasicPlanets plus asteroids, dwarf planets, and TNPs (28 bodies).
var AllPlanets []PlanetID

// ElectionalPlanets is BasicPlanets plus Chiron and Lilith (14 bodies).
var ElectionalPlanets []PlanetID

// AllPlanetNames is the name list for AllPlanets (29 bodies including TNPs and dwarf planets).
var AllPlanetNames []string

// NonTNPPlanetNames is AllPlanets minus TNPs (23 bodies: Sun-Pluto+Node+TrueNode+asteroids+Chiron+Lilith+dwarfs+SouthNode).
var NonTNPPlanetNames []string

// NonTNPNoNodePlanetNames is NonTNPPlanetNames minus Node, TrueNode, and SouthNode (20 bodies).
var NonTNPNoNodePlanetNames []string

func init() {
	AllPlanets = make([]PlanetID, 0, 49)
	AllPlanets = append(AllPlanets, BasicPlanets...)
	AllPlanets = append(AllPlanets,
		PlanetID{"Ceres", swe.CERES},
		PlanetID{"Pallas", swe.PALLAS},
		PlanetID{"Juno", swe.JUNO},
		PlanetID{"Vesta", swe.VESTA},
		PlanetID{"Lilith", swe.MEAN_APOG},
		PlanetID{"Chiron", swe.CHIRON},
		// Major asteroids (0-999)
		PlanetID{"Astraea", swe.ASTRAEA},
		PlanetID{"Hebe", swe.HEBE},
		PlanetID{"Iris", swe.IRIS},
		PlanetID{"Flora", swe.FLORA},
		PlanetID{"Metis", swe.METIS},
		PlanetID{"Hygiea", swe.HYGIEA},
		PlanetID{"Psyche", swe.PSYCHE},
		PlanetID{"Fortuna", swe.FORTUNA},
		PlanetID{"Proserpina", swe.PROSERPINA},
		PlanetID{"Amphitrite", swe.AMPHITRITE},
		PlanetID{"Pandora", swe.PANDORA},
		PlanetID{"Mnemosyne", swe.MNEMOSYNE},
		PlanetID{"Cybele", swe.CYBELE},
		PlanetID{"Diana", swe.DIANA},
		PlanetID{"Sappho", swe.SAPPHO},
		PlanetID{"Eros", swe.EROS},
		// Dwarf planets
		PlanetID{"Eris", swe.ERIS},
		PlanetID{"Makemake", swe.MAKEMAKE},
		PlanetID{"Gonggong", swe.GONGGONG},
		// Distant objects
		PlanetID{"Orcus", swe.ORCUS},
		PlanetID{"Sedna", swe.SEDNA},
		PlanetID{"Haumea", swe.HAUMEA},
		PlanetID{"SouthNode", -1}, // synthetic: NN + 180°
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
	NonTNPPlanetNames = planetNames(AllPlanets[:41])

	NonTNPNoNodePlanetNames = make([]string, 0, 38)
	for _, p := range AllPlanets[:41] {
		if p.Name != "Node" && p.Name != "TrueNode" && p.Name != "SouthNode" {
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
