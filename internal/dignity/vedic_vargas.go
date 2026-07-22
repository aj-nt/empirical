package dignity

import "math"

// ── Varga Charts ──────────────────────────────────────────────────────────
//
// Varga (divisional) charts divide each sign into segments and reassign
// planets to new signs based on which segment they fall in.
//
// Supported vargas:
//   D3  (Drekkana)   — siblings, courage, co-born
//   D7  (Saptamsha)  — children, creativity, progeny
//   D9  (Navamsha)   — marriage, dharma, inner capacity (already in vedic_natal.go)
//   D10 (Dashamsha)  — career, profession, status

// VargaPosition holds a planet's position in a varga chart.
type VargaPosition struct {
	Planet string `json:"planet"`
	Sign   string `json:"sign"`
	Lon    float64 `json:"lon"`
}

// ComputeVargaChart computes a varga chart for the given planets.
// varga: "D3", "D7", "D9", "D10"
// planetLons: planet name → sidereal longitude
func ComputeVargaChart(varga string, planetLons map[string]float64) []VargaPosition {
	divisions := vargaDivisions(varga)
	if divisions == 0 {
		return nil
	}

	order := []string{"Sun", "Moon", "Mercury", "Venus", "Mars", "Jupiter", "Saturn", "Node"}
	var result []VargaPosition

	for _, planet := range order {
		lon, ok := planetLons[planet]
		if !ok {
			continue
		}
		sign, vargaLon := vargaSign(lon, divisions, varga)
		result = append(result, VargaPosition{
			Planet: planet,
			Sign:   sign,
			Lon:    vargaLon,
		})
	}
	return result
}

// vargaDivisions returns the number of divisions for a varga.
func vargaDivisions(varga string) int {
	switch varga {
	case "D3":
		return 3
	case "D7":
		return 7
	case "D9":
		return 9
	case "D10":
		return 10
	}
	return 0
}

// vargaSign computes the varga sign for a sidereal longitude.
// Returns the sign name and the varga longitude.
func vargaSign(siderealLon float64, divisions int, varga string) (string, float64) {
	sidLon := math.Mod(siderealLon, 360)
	signIdx := int(sidLon / 30.0) % 12
	degInSign := math.Mod(sidLon, 30.0)
	segmentSize := 30.0 / float64(divisions)
	segment := int(degInSign / segmentSize)

	startSign := vargaStartSign(signIdx, segment, divisions, varga)
	vargaSignIdx := (startSign + segment) % 12

	// Varga longitude: proportional position within the segment
	segFrac := math.Mod(degInSign, segmentSize) / segmentSize
	vargaLon := float64(vargaSignIdx)*30.0 + segFrac*30.0

	return Signs[vargaSignIdx], vargaLon
}

// vargaStartSign returns the starting sign for a varga segment.
// Different vargas have different counting rules.
func vargaStartSign(signIdx, segment, divisions int, varga string) int {
	switch varga {
	case "D3":
		// Drekkana: each 10° segment. Fire signs start at Aries,
		// Earth at Capricorn, Air at Libra, Water at Cancer.
		// But D3 has a special rule: the first drekkana is the sign itself,
		// second is 5th from it, third is 9th from it.
		switch segment {
		case 0:
			return signIdx
		case 1:
			return (signIdx + 4) % 12
		default:
			return (signIdx + 8) % 12
		}

	case "D7":
		// Saptamsha: 7 divisions of ~4°17' each.
		// Odd signs: count from the sign itself.
		// Even signs: count from the 7th from the sign.
		if signIdx%2 == 0 { // odd sign (0-indexed: Aries=0, Gemini=2, etc.)
			return signIdx
		}
		return (signIdx + 6) % 12

	case "D9":
		// Navamsha: element-based starting sign
		element := signIdx % 4
		switch element {
		case 0: // Fire
			return 0 // Aries
		case 1: // Earth
			return 9 // Capricorn
		case 2: // Air
			return 6 // Libra
		default: // Water
			return 3 // Cancer
		}

	case "D10":
		// Dashamsha: 10 divisions of 3° each.
		// Odd signs: count from the sign itself.
		// Even signs: count from the 9th from the sign.
		if signIdx%2 == 0 { // odd sign
			return signIdx
		}
		return (signIdx + 8) % 12
	}
	return 0
}
