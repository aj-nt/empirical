package dignity

// ── Element & Modality Balance ─────────────────────────────────────────────
//
// Pure math — no SWE calls. Counts planets by element (Fire/Earth/Air/Water)
// and modality (Cardinal/Fixed/Mutable) based on sign position.

// SignElement returns the element of a zodiac sign.
func SignElement(sign string) string {
	switch sign {
	case "Aries", "Leo", "Sagittarius":
		return "Fire"
	case "Taurus", "Virgo", "Capricorn":
		return "Earth"
	case "Gemini", "Libra", "Aquarius":
		return "Air"
	case "Cancer", "Scorpio", "Pisces":
		return "Water"
	}
	return ""
}

// SignModality returns the modality of a zodiac sign.
func SignModality(sign string) string {
	switch sign {
	case "Aries", "Cancer", "Libra", "Capricorn":
		return "Cardinal"
	case "Taurus", "Leo", "Scorpio", "Aquarius":
		return "Fixed"
	case "Gemini", "Virgo", "Sagittarius", "Pisces":
		return "Mutable"
	}
	return ""
}

// ComputeElementBalance counts planets in each element.
func ComputeElementBalance(planets map[string]float64) map[string]int {
	bal := map[string]int{"Fire": 0, "Earth": 0, "Air": 0, "Water": 0}
	for _, lon := range planets {
		sign := SignForLongitude(lon)
		bal[SignElement(sign)]++
	}
	return bal
}

// ComputeModalityBalance counts planets in each modality.
func ComputeModalityBalance(planets map[string]float64) map[string]int {
	bal := map[string]int{"Cardinal": 0, "Fixed": 0, "Mutable": 0}
	for _, lon := range planets {
		sign := SignForLongitude(lon)
		bal[SignModality(sign)]++
	}
	return bal
}
