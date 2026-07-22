package dignity

// ── Vedic Yogas ───────────────────────────────────────────────────────────
//
// Yogas are planetary combinations that produce specific effects in a chart.
// Detection is based on whole-sign positions, dignities, and house lords.

// VedicYoga represents a detected yoga in the chart.
type VedicYoga struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"` // "raja", "dhana", "chandra", "surya", "mahapurusha", "viparita", "other"
}

// yogaContext holds all the data needed for yoga detection.
type yogaContext struct {
	planetSigns  map[string]string // planet → sidereal sign
	planetHouses map[string]int    // planet → whole-sign house
	dignities    map[string]string // planet → Vedic dignity
	houseLords   map[int]string    // house → ruling planet
	aspects      []VedicAspect
}

// ComputeVedicYogas detects all applicable yogas in the chart.
func ComputeVedicYogas(planetSigns map[string]string, planetHouses map[string]int, dignities map[string]string, houseLords map[int]string, aspects []VedicAspect) []VedicYoga {
	ctx := &yogaContext{
		planetSigns:  planetSigns,
		planetHouses: planetHouses,
		dignities:    dignities,
		houseLords:   houseLords,
		aspects:      aspects,
	}

	var yogas []VedicYoga

	// ── Pancha Mahapurusha Yogas ──────────────────────────────────────
	yogas = append(yogas, detectMahapurusha(ctx)...)

	// ── Gaja Kesari Yoga ──────────────────────────────────────────────
	if y := detectGajaKesari(ctx); y != nil {
		yogas = append(yogas, *y)
	}

	// ── Budha-Aditya Yoga ─────────────────────────────────────────────
	if y := detectBudhaAditya(ctx); y != nil {
		yogas = append(yogas, *y)
	}

	// ── Chandra-Mangala Yoga ──────────────────────────────────────────
	if y := detectChandraMangala(ctx); y != nil {
		yogas = append(yogas, *y)
	}

	// ── Adhi Yoga ─────────────────────────────────────────────────────
	if y := detectAdhiYoga(ctx); y != nil {
		yogas = append(yogas, *y)
	}

	// ── Sunapha / Anapha / Durudhara ──────────────────────────────────
	yogas = append(yogas, detectChandraYogas(ctx)...)

	// ── Kemadruma Yoga ────────────────────────────────────────────────
	if y := detectKemadruma(ctx); y != nil {
		yogas = append(yogas, *y)
	}

	// ── Viparita Raja Yogas ───────────────────────────────────────────
	yogas = append(yogas, detectViparitaRaja(ctx)...)

	// ── Dhana Yogas ───────────────────────────────────────────────────
	yogas = append(yogas, detectDhanaYogas(ctx)...)

	// ── Raja Yogas ────────────────────────────────────────────────────
	yogas = append(yogas, detectRajaYogas(ctx)...)

	// ── Amala Yoga ────────────────────────────────────────────────────
	if y := detectAmalaYoga(ctx); y != nil {
		yogas = append(yogas, *y)
	}

	// ── Parvata Yoga ──────────────────────────────────────────────────
	if y := detectParvataYoga(ctx); y != nil {
		yogas = append(yogas, *y)
	}

	// ── Lakshmi Yoga ──────────────────────────────────────────────────
	if y := detectLakshmiYoga(ctx); y != nil {
		yogas = append(yogas, *y)
	}

	// ── Vesi / Vosi Yogas ─────────────────────────────────────────────
	yogas = append(yogas, detectSuryaYogas(ctx)...)

	return yogas
}

// ── Pancha Mahapurusha Yogas ──────────────────────────────────────────────

func detectMahapurusha(ctx *yogaContext) []VedicYoga {
	var yogas []VedicYoga
	kendras := map[int]bool{1: true, 4: true, 7: true, 10: true}

	checks := []struct {
		planet, yogaName, desc string
	}{
		{"Mars", "Ruchaka", "Mars in own or exalted sign in a kendra — courage, leadership, military prowess"},
		{"Mercury", "Bhadra", "Mercury in own or exalted sign in a kendra — intellect, learning, eloquence"},
		{"Jupiter", "Hamsa", "Jupiter in own or exalted sign in a kendra — wisdom, spirituality, good fortune"},
		{"Venus", "Malavya", "Venus in own or exalted sign in a kendra — beauty, luxury, artistic talent"},
		{"Saturn", "Sasa", "Saturn in own or exalted sign in a kendra — discipline, authority, endurance"},
	}

	for _, c := range checks {
		house, ok := ctx.planetHouses[c.planet]
		if !ok {
			continue
		}
		if !kendras[house] {
			continue
		}
		dignity := ctx.dignities[c.planet]
		if dignity == "swakshetra" || dignity == "uchcha" {
			yogas = append(yogas, VedicYoga{
				Name:        c.yogaName,
				Description: c.desc,
				Category:    "mahapurusha",
			})
		}
	}
	return yogas
}

// ── Gaja Kesari Yoga ──────────────────────────────────────────────────────

func detectGajaKesari(ctx *yogaContext) *VedicYoga {
	// Jupiter in a kendra (1,4,7,10) from the Moon
	moonHouse, ok := ctx.planetHouses["Moon"]
	if !ok {
		return nil
	}
	jupHouse, ok := ctx.planetHouses["Jupiter"]
	if !ok {
		return nil
	}
	// Kendra from Moon: Moon's house, 4th, 7th, 10th from Moon
	// Nth from H = ((H + N - 2) % 12) + 1
	kendraFromMoon := map[int]bool{
		moonHouse: true,                          // 1st
		((moonHouse + 2) % 12) + 1: true,         // 4th
		((moonHouse + 5) % 12) + 1: true,         // 7th
		((moonHouse + 8) % 12) + 1: true,         // 10th
	}
	if kendraFromMoon[jupHouse] {
		return &VedicYoga{
			Name:        "Gaja Kesari",
			Description: "Jupiter in a kendra from the Moon — wisdom, fame, prosperity, elephant-like strength",
			Category:    "raja",
		}
	}
	return nil
}

// ── Budha-Aditya Yoga ─────────────────────────────────────────────────────

func detectBudhaAditya(ctx *yogaContext) *VedicYoga {
	// Mercury and Sun in the same sign
	mercSign, ok := ctx.planetSigns["Mercury"]
	if !ok {
		return nil
	}
	sunSign, ok := ctx.planetSigns["Sun"]
	if !ok {
		return nil
	}
	if mercSign == sunSign {
		return &VedicYoga{
			Name:        "Budha-Aditya",
			Description: "Mercury and Sun in the same sign — sharp intellect, learning, communication skills",
			Category:    "surya",
		}
	}
	return nil
}

// ── Chandra-Mangala Yoga ───────────────────────────────────────────────────

func detectChandraMangala(ctx *yogaContext) *VedicYoga {
	// Moon and Mars in the same sign
	moonSign, ok := ctx.planetSigns["Moon"]
	if !ok {
		return nil
	}
	marsSign, ok := ctx.planetSigns["Mars"]
	if !ok {
		return nil
	}
	if moonSign == marsSign {
		return &VedicYoga{
			Name:        "Chandra-Mangala",
			Description: "Moon and Mars in the same sign — emotional intensity, courage, drive, potential for wealth through effort",
			Category:    "chandra",
		}
	}
	return nil
}

// ── Adhi Yoga ─────────────────────────────────────────────────────────────

func detectAdhiYoga(ctx *yogaContext) *VedicYoga {
	// Benefics (Mercury, Jupiter, Venus) in 6th, 7th, or 8th from Moon
	moonHouse, ok := ctx.planetHouses["Moon"]
	if !ok {
		return nil
	}
	benefics := []string{"Mercury", "Jupiter", "Venus"}
	// Nth from H = ((H + N - 2) % 12) + 1
	housesFromMoon := map[int]bool{
		((moonHouse + 4) % 12) + 1: true, // 6th
		((moonHouse + 5) % 12) + 1: true, // 7th
		((moonHouse + 6) % 12) + 1: true, // 8th
	}
	for _, b := range benefics {
		if h, ok := ctx.planetHouses[b]; ok && housesFromMoon[h] {
			return &VedicYoga{
				Name:        "Adhi Yoga",
				Description: "Benefics in 6th, 7th, or 8th from the Moon — authority, prosperity, freedom from enemies",
				Category:    "chandra",
			}
		}
	}
	return nil
}

// ── Chandra Yogas (Sunapha / Anapha / Durudhara) ──────────────────────────

func detectChandraYogas(ctx *yogaContext) []VedicYoga {
	moonHouse, ok := ctx.planetHouses["Moon"]
	if !ok {
		return nil
	}
	// Nth from H = ((H + N - 2) % 12) + 1
	house2 := (moonHouse % 12) + 1       // 2nd from Moon
	house12 := ((moonHouse + 10) % 12) + 1 // 12th from Moon

	planetsIn2nd := planetsInHouse(ctx.planetHouses, house2, "Moon", "Sun")
	planetsIn12th := planetsInHouse(ctx.planetHouses, house12, "Moon", "Sun")

	has2nd := len(planetsIn2nd) > 0
	has12th := len(planetsIn12th) > 0

	var yogas []VedicYoga
	if has2nd && !has12th {
		yogas = append(yogas, VedicYoga{
			Name:        "Sunapha",
			Description: "Planets in the 2nd from the Moon — wealth, intelligence, self-made success",
			Category:    "chandra",
		})
	}
	if has12th && !has2nd {
		yogas = append(yogas, VedicYoga{
			Name:        "Anapha",
			Description: "Planets in the 12th from the Moon — prosperity, good health, commanding presence",
			Category:    "chandra",
		})
	}
	if has2nd && has12th {
		yogas = append(yogas, VedicYoga{
			Name:        "Durudhara",
			Description: "Planets in both 2nd and 12th from the Moon — wealth, fame, balanced fortune",
			Category:    "chandra",
		})
	}
	return yogas
}

// ── Kemadruma Yoga ────────────────────────────────────────────────────────

func detectKemadruma(ctx *yogaContext) *VedicYoga {
	// No planets (except Sun) in 2nd or 12th from Moon
	moonHouse, ok := ctx.planetHouses["Moon"]
	if !ok {
		return nil
	}
	// Nth from H = ((H + N - 2) % 12) + 1
	house2 := (moonHouse % 12) + 1       // 2nd from Moon
	house12 := ((moonHouse + 10) % 12) + 1 // 12th from Moon

	for planet, house := range ctx.planetHouses {
		if planet == "Moon" || planet == "Sun" {
			continue
		}
		if house == house2 || house == house12 {
			return nil // planet found — no Kemadruma
		}
	}
	return &VedicYoga{
		Name:        "Kemadruma",
		Description: "No planets in 2nd or 12th from Moon — isolation, struggle, but can produce spiritual depth",
		Category:    "chandra",
	}
}

// ── Viparita Raja Yogas ───────────────────────────────────────────────────

func detectViparitaRaja(ctx *yogaContext) []VedicYoga {
	// Lord of a dusthana (6,8,12) placed in a dusthana
	dusthanas := map[int]bool{6: true, 8: true, 12: true}
	var yogas []VedicYoga

	for _, dusthana := range []int{6, 8, 12} {
		lord, ok := ctx.houseLords[dusthana]
		if !ok {
			continue
		}
		lordHouse, ok := ctx.planetHouses[lord]
		if !ok {
			continue
		}
		if dusthanas[lordHouse] {
			names := map[int]string{6: "Harsha", 8: "Sarala", 12: "Vimala"}
			yogas = append(yogas, VedicYoga{
				Name:        names[dusthana],
				Description: "Lord of a dusthana placed in a dusthana — adversity transformed into success, rise through challenges",
				Category:    "viparita",
			})
		}
	}
	return yogas
}

// ── Dhana Yogas ───────────────────────────────────────────────────────────

func detectDhanaYogas(ctx *yogaContext) []VedicYoga {
	var yogas []VedicYoga

	// Lord of 2nd in a kendra (1,4,7,10)
	lord2, ok := ctx.houseLords[2]
	if ok {
		if h, ok := ctx.planetHouses[lord2]; ok {
			if h == 1 || h == 4 || h == 7 || h == 10 {
				yogas = append(yogas, VedicYoga{
					Name:        "Dhana Yoga",
					Description: "Lord of the 2nd house in a kendra — wealth through own efforts, financial stability",
					Category:    "dhana",
				})
			}
		}
	}

	// Lord of 11th in a trikona (1,5,9) or 2nd
	lord11, ok := ctx.houseLords[11]
	if ok {
		if h, ok := ctx.planetHouses[lord11]; ok {
			if h == 1 || h == 5 || h == 9 || h == 2 {
				yogas = append(yogas, VedicYoga{
					Name:        "Dhana Yoga",
					Description: "Lord of the 11th house in a trikona or 2nd — gains, income, financial prosperity",
					Category:    "dhana",
				})
			}
		}
	}

	// Lord of 9th in 11th or lord of 11th in 9th
	lord9, ok9 := ctx.houseLords[9]
	lord11b, ok11 := ctx.houseLords[11]
	if ok9 && ok11 {
		h9 := ctx.planetHouses[lord9]
		h11 := ctx.planetHouses[lord11b]
		if h9 == 11 || h11 == 9 {
			yogas = append(yogas, VedicYoga{
				Name:        "Dhana Yoga",
				Description: "Lords of 9th and 11th in mutual exchange — fortune and gains combine, significant wealth",
				Category:    "dhana",
			})
		}
	}

	return yogas
}

// ── Raja Yogas ────────────────────────────────────────────────────────────

func detectRajaYogas(ctx *yogaContext) []VedicYoga {
	var yogas []VedicYoga
	kendras := map[int]bool{1: true, 4: true, 7: true, 10: true}
	trikonas := map[int]bool{1: true, 5: true, 9: true}

	// Lord of a kendra and lord of a trikona in mutual relationship
	// (conjunction, mutual aspect, or exchange)
	for k := range kendras {
		for t := range trikonas {
			if k == t {
				continue
			}
			lordK, okK := ctx.houseLords[k]
			lordT, okT := ctx.houseLords[t]
			if !okK || !okT || lordK == lordT {
				continue
			}
			hK, ok := ctx.planetHouses[lordK]
			if !ok {
				continue
			}
			hT, ok := ctx.planetHouses[lordT]
			if !ok {
				continue
			}
			// Conjunction: same house
			if hK == hT {
				yogas = append(yogas, VedicYoga{
					Name:        "Raja Yoga",
					Description: "Lords of a kendra and trikona conjoined — power, authority, leadership",
					Category:    "raja",
				})
			}
			// Mutual aspect via drishti
			if mutuallyAspecting(ctx.aspects, lordK, lordT) {
				yogas = append(yogas, VedicYoga{
					Name:        "Raja Yoga",
					Description: "Lords of a kendra and trikona in mutual aspect — rise to prominence, respect",
					Category:    "raja",
				})
			}
		}
	}

	return yogas
}

// ── Amala Yoga ────────────────────────────────────────────────────────────

func detectAmalaYoga(ctx *yogaContext) *VedicYoga {
	// A benefic in the 10th house from the ascendant
	benefics := []string{"Jupiter", "Venus", "Mercury", "Moon"}
	for _, b := range benefics {
		if h, ok := ctx.planetHouses[b]; ok && h == 10 {
			return &VedicYoga{
				Name:        "Amala",
				Description: "A benefic in the 10th house — pure reputation, virtuous conduct, success in career",
				Category:    "raja",
			}
		}
	}
	return nil
}

// ── Parvata Yoga ──────────────────────────────────────────────────────────

func detectParvataYoga(ctx *yogaContext) *VedicYoga {
	// Benefics in kendras, 6th and 8th empty
	kendras := map[int]bool{1: true, 4: true, 7: true, 10: true}
	benefics := map[string]bool{"Jupiter": true, "Venus": true, "Mercury": true, "Moon": true}

	hasBeneficInKendra := false
	for planet, house := range ctx.planetHouses {
		if benefics[planet] && kendras[house] {
			hasBeneficInKendra = true
			break
		}
	}
	if !hasBeneficInKendra {
		return nil
	}

	// Check 6th and 8th are empty
	for _, planet := range []string{"Sun", "Moon", "Mercury", "Venus", "Mars", "Jupiter", "Saturn", "Node"} {
		if h, ok := ctx.planetHouses[planet]; ok && (h == 6 || h == 8) {
			return nil
		}
	}

	return &VedicYoga{
		Name:        "Parvata",
		Description: "Benefics in kendras with 6th and 8th empty — fortune, fame, mountain-like stability",
		Category:    "raja",
	}
}

// ── Lakshmi Yoga ───────────────────────────────────────────────────────────

func detectLakshmiYoga(ctx *yogaContext) *VedicYoga {
	// Venus in own sign in a kendra, OR lord of 9th in 11th
	kendras := map[int]bool{1: true, 4: true, 7: true, 10: true}

	// Venus in own sign in kendra
	if h, ok := ctx.planetHouses["Venus"]; ok && kendras[h] {
		if ctx.dignities["Venus"] == "swakshetra" {
			return &VedicYoga{
				Name:        "Lakshmi",
				Description: "Venus in own sign in a kendra — wealth, beauty, prosperity, goddess Lakshmi's blessing",
				Category:    "dhana",
			}
		}
	}

	// Lord of 9th in 11th
	if lord9, ok := ctx.houseLords[9]; ok {
		if h, ok := ctx.planetHouses[lord9]; ok && h == 11 {
			return &VedicYoga{
				Name:        "Lakshmi",
				Description: "Lord of the 9th in the 11th — fortune flows into gains, sustained prosperity",
				Category:    "dhana",
			}
		}
	}

	return nil
}

// ── Surya Yogas (Vesi / Vosi) ─────────────────────────────────────────────

func detectSuryaYogas(ctx *yogaContext) []VedicYoga {
	sunHouse, ok := ctx.planetHouses["Sun"]
	if !ok {
		return nil
	}
	// Nth from H = ((H + N - 2) % 12) + 1
	house2 := (sunHouse % 12) + 1       // 2nd from Sun
	house12 := ((sunHouse + 10) % 12) + 1 // 12th from Sun

	planetsIn2nd := planetsInHouse(ctx.planetHouses, house2, "Sun")
	planetsIn12th := planetsInHouse(ctx.planetHouses, house12, "Sun")

	var yogas []VedicYoga
	if len(planetsIn2nd) > 0 {
		yogas = append(yogas, VedicYoga{
			Name:        "Vesi",
			Description: "Planets in the 2nd from the Sun — self-reliance, determination, independent success",
			Category:    "surya",
		})
	}
	if len(planetsIn12th) > 0 {
		yogas = append(yogas, VedicYoga{
			Name:        "Vosi",
			Description: "Planets in the 12th from the Sun — refinement, generosity, spiritual inclination",
			Category:    "surya",
		})
	}
	return yogas
}

// ── Helpers ───────────────────────────────────────────────────────────────

// planetsInHouse returns planet names in a given house, excluding the listed exceptions.
func planetsInHouse(planetHouses map[string]int, house int, exclude ...string) []string {
	excludeSet := make(map[string]bool, len(exclude))
	for _, e := range exclude {
		excludeSet[e] = true
	}
	var result []string
	for planet, h := range planetHouses {
		if h == house && !excludeSet[planet] {
			result = append(result, planet)
		}
	}
	return result
}

// mutuallyAspecting checks if two planets aspect each other via Vedic drishti.
func mutuallyAspecting(aspects []VedicAspect, a, b string) bool {
	aToB := false
	bToA := false
	for _, asp := range aspects {
		if asp.FromPlanet == a && asp.ToPlanet == b {
			aToB = true
		}
		if asp.FromPlanet == b && asp.ToPlanet == a {
			bToA = true
		}
	}
	return aToB && bToA
}
