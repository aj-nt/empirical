package dignity

// ── Vedic Aspects (Drishti) ───────────────────────────────────────────────
//
// Vedic aspects are based on whole-sign house positions, not angular degrees.
// Each planet casts a full aspect on specific houses counted from its position.
//
// Standard Parashari drishti:
//   All planets: 7th house (opposition)
//   Mars:        4th, 7th, 8th
//   Jupiter:     5th, 7th, 9th
//   Saturn:      3rd, 7th, 10th
//   Rahu/Ketu:   5th, 7th, 9th (like Jupiter)

// VedicAspect represents one planet aspecting another.
type VedicAspect struct {
	FromPlanet string `json:"from_planet"`
	ToPlanet   string `json:"to_planet"`
	FromHouse  int    `json:"from_house"`
	ToHouse    int    `json:"to_house"`
	Type       string `json:"type"` // "7th", "4th", "8th", "5th", "9th", "3rd", "10th"
}

// vedicAspectHouses returns the houses a planet aspects from its position.
// Houses are counted inclusively: planet's own house is 1, the next is 2, etc.
func vedicAspectHouses(planet string) []int {
	switch planet {
	case "Mars":
		return []int{4, 7, 8}
	case "Jupiter":
		return []int{5, 7, 9}
	case "Saturn":
		return []int{3, 7, 10}
	case "Node", "Rahu", "Ketu":
		return []int{5, 7, 9}
	default:
		return []int{7}
	}
}

// ComputeVedicAspects computes all Vedic aspects between the given planets
// using their whole-sign house positions (1-12).
// planetHouses maps planet name → house number.
func ComputeVedicAspects(planetHouses map[string]int) []VedicAspect {
	planetNames := []string{"Sun", "Moon", "Mercury", "Venus", "Mars", "Jupiter", "Saturn", "Node"}

	var aspects []VedicAspect
	for _, from := range planetNames {
		fromHouse, ok := planetHouses[from]
		if !ok {
			continue
		}
		aspectHouses := vedicAspectHouses(from)
		for _, to := range planetNames {
			if from == to {
				continue
			}
			toHouse, ok := planetHouses[to]
			if !ok {
				continue
			}
			// Check if toHouse is one of the aspected houses from fromHouse
			for _, ah := range aspectHouses {
				// House counting: fromHouse + (ah - 1), wrap at 12
				targetHouse := ((fromHouse + ah - 2) % 12) + 1
				if targetHouse == toHouse {
					aspects = append(aspects, VedicAspect{
						FromPlanet: from,
						ToPlanet:   to,
						FromHouse:  fromHouse,
						ToHouse:    toHouse,
						Type:       ordinalAspectName(ah),
					})
					break
				}
			}
		}
	}
	return aspects
}

// ordinalAspectName converts a house offset to a readable name.
func ordinalAspectName(offset int) string {
	switch offset {
	case 3:
		return "3rd"
	case 4:
		return "4th"
	case 5:
		return "5th"
	case 7:
		return "7th"
	case 8:
		return "8th"
	case 9:
		return "9th"
	case 10:
		return "10th"
	default:
		return "aspect"
	}
}
