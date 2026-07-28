package dignity

import (
	"math"
)

// ═══════════════════════════════════════════════════════════════════════════
// Shadbala — Six-Fold Planetary Strength (Brihat Parashara Hora Shastra)
// ═══════════════════════════════════════════════════════════════════════════
//
// Shadbala quantifies a planet's strength across six dimensions:
//   1. Sthana Bala  — positional strength (sign, house, exaltation)
//   2. Dig Bala     — directional strength (angular house placement)
//   3. Kala Bala    — temporal strength (time of day/year)
//   4. Chesta Bala  — motional strength (speed, retrogradation)
//   5. Naisargika Bala — natural strength (inherent luminosity)
//   6. Drik Bala    — aspectual strength (aspects received)
//
// Total strength is measured in Rupas (1 Rupa = 60 Shashtiamsas).
// A planet with > 390 Shashtiamsas (6.5 Rupas) is considered strong.

// ShadbalaResult holds the six-fold strength breakdown for one planet.
type ShadbalaResult struct {
	Planet       string  `json:"planet"`
	SthanaBala   float64 `json:"sthana_bala"`
	DigBala      float64 `json:"dig_bala"`
	KalaBala     float64 `json:"kala_bala"`
	ChestaBala   float64 `json:"chesta_bala"`
	NaisargikaBala float64 `json:"naisargika_bala"`
	DrikBala     float64 `json:"drik_bala"`
	Total        float64 `json:"total"`
	Rupas        float64 `json:"rupas"`       // total / 60
	IsStrong     bool    `json:"is_strong"`   // > 390 shashtiamsas
}

// ShadbalaReport holds the full shadbala analysis.
type ShadbalaReport struct {
	Planets []ShadbalaResult `json:"planets"`
}

// ComputeShadbala computes shadbala for all classical planets.
// planetLons: sidereal longitudes
// planetHouses: whole-sign houses from sidereal ASC
// planetSpeeds: speed in deg/day (positive = direct, negative = retrograde)
// asc: sidereal ascendant longitude
// sunLon: sidereal Sun longitude
// isDay: true if Sun above horizon
func ComputeShadbala(
	planetLons map[string]float64,
	planetHouses map[string]int,
	planetSpeeds map[string]float64,
	asc float64,
	sunLon float64,
	isDay bool,
) ShadbalaReport {
	order := []string{"Sun", "Moon", "Mercury", "Venus", "Mars", "Jupiter", "Saturn"}
	var results []ShadbalaResult

	for _, planet := range order {
		lon, ok := planetLons[planet]
		if !ok {
			continue
		}
		house, _ := planetHouses[planet]
		speed, _ := planetSpeeds[planet]

		sb := computeSthanaBala(planet, lon, asc)
		db := computeDigBala(planet, house)
		kb := computeKalaBala(planet, lon, sunLon, isDay)
		cb := computeChestaBala(planet, speed)
		nb := computeNaisargikaBala(planet)
		drb := computeDrikBala(planet, planetLons, planetHouses)

		total := sb + db + kb + cb + nb + drb
		rupas := total / 60.0

		results = append(results, ShadbalaResult{
			Planet:         planet,
			SthanaBala:     math.Round(sb*100) / 100,
			DigBala:        math.Round(db*100) / 100,
			KalaBala:       math.Round(kb*100) / 100,
			ChestaBala:     math.Round(cb*100) / 100,
			NaisargikaBala: math.Round(nb*100) / 100,
			DrikBala:       math.Round(drb*100) / 100,
			Total:          math.Round(total*100) / 100,
			Rupas:          math.Round(rupas*100) / 100,
			IsStrong:       total > 390,
		})
	}

	return ShadbalaReport{Planets: results}
}

// ── Sthana Bala (Positional Strength) ──────────────────────────────────────

func computeSthanaBala(planet string, lon, asc float64) float64 {
	signIdx := int(lon/30) % 12
	sign := Signs[signIdx]
	degInSign := math.Mod(lon, 30)

	var bala float64

	// 1. Uccha Bala (exaltation strength) — max 60
	exaltDeg := exaltationDegree(planet)
	if exaltDeg >= 0 {
		dist := math.Abs(lon - exaltDeg)
		if dist > 180 {
			dist = 360 - dist
		}
		// Strength decreases linearly from exaltation point
		ucchaBala := 60.0 * (1.0 - dist/180.0)
		bala += ucchaBala
	}

	// 2. Saptavargaja Bala (varga strength) — simplified: max 45
	// Own sign: 30, friend's sign: 22.5, neutral: 15, enemy: 7.5, debilitation: 0
	owner := signRuler(sign)
	if owner == planet {
		bala += 30
	} else if isFriend(planet, owner) {
		bala += 22.5
	} else if isEnemy(planet, owner) {
		bala += 7.5
	} else {
		bala += 15
	}

	// 3. Ojayugma Bala (odd/even sign) — max 15
	if planet == "Moon" || planet == "Venus" {
		// Even signs (Taurus, Cancer, Virgo, Scorpio, Capricorn, Pisces) = 0,2,4,5,7,9,11
		if signIdx%2 == 0 {
			bala += 15
		}
	} else {
		// Odd signs for other planets
		if signIdx%2 == 1 {
			bala += 15
		}
	}

	// 4. Kendradi Bala (angular/cadent) — max 60
	houseFromAsc := ((signIdx - int(asc/30) + 12) % 12) + 1
	switch {
	case houseFromAsc == 1 || houseFromAsc == 4 || houseFromAsc == 7 || houseFromAsc == 10:
		bala += 60 // Kendra (angular)
	case houseFromAsc == 2 || houseFromAsc == 5 || houseFromAsc == 8 || houseFromAsc == 11:
		bala += 30 // Panaphara (succedent)
	default:
		bala += 15 // Apoklima (cadent)
	}

	// 5. Drekkena Bala (decanate) — max 15
	dec := int(degInSign / 10)
	if (planet == "Sun" || planet == "Mars" || planet == "Jupiter") && dec == 0 {
		bala += 15
	} else if (planet == "Moon" || planet == "Venus" || planet == "Saturn") && dec == 1 {
		bala += 15
	} else if planet == "Mercury" && dec == 2 {
		bala += 15
	}

	return bala
}

// ── Dig Bala (Directional Strength) ────────────────────────────────────────

func computeDigBala(planet string, house int) float64 {
	// Each planet has a directional strength in a specific house:
	// Sun/Mars: 10th (south), Moon/Venus: 4th (north), Mercury/Jupiter: 1st (east), Saturn: 7th (west)
	digHouse := map[string]int{
		"Sun": 10, "Mars": 10,
		"Moon": 4, "Venus": 4,
		"Mercury": 1, "Jupiter": 1,
		"Saturn": 7,
	}

	ideal, ok := digHouse[planet]
	if !ok {
		return 0
	}

	// Distance from ideal house (in houses)
	dist := math.Abs(float64(house - ideal))
	if dist > 6 {
		dist = 12 - dist
	}

	// Max 60 at ideal, decreases by 10 per house away
	return math.Max(0, 60-dist*10)
}

// ── Kala Bala (Temporal Strength) ──────────────────────────────────────────

func computeKalaBala(planet string, lon, sunLon float64, isDay bool) float64 {
	var bala float64

	// 1. Nathonnata Bala (day/night strength) — max 60
	// Sun, Jupiter, Saturn are strong during day
	// Moon, Mars, Venus are strong during night
	// Mercury is always strong
	dayPlanets := map[string]bool{"Sun": true, "Jupiter": true, "Saturn": true}
	nightPlanets := map[string]bool{"Moon": true, "Mars": true, "Venus": true}

	if planet == "Mercury" {
		bala += 60
	} else if (isDay && dayPlanets[planet]) || (!isDay && nightPlanets[planet]) {
		bala += 60
	} else {
		bala += 0
	}

	// 2. Paksha Bala (lunar phase strength) — Moon only, max 60
	if planet == "Moon" {
		separation := math.Abs(lon - sunLon)
		if separation > 180 {
			separation = 360 - separation
		}
		// Waxing = strong, waning = weak
		// Moon ahead of Sun = waxing
		moonAhead := (lon - sunLon + 360)
		for moonAhead >= 360 {
			moonAhead -= 360
		}
		if moonAhead < 180 {
			bala += 60 * (separation / 180)
		} else {
			bala += 60 * (1 - separation/180)
		}
	}

	// 3. Tribhaga Bala (third-of-day strength) — max 60
	// Simplified: day = Sun strong, night = Moon strong, twilight = Mercury strong
	if isDay && planet == "Sun" {
		bala += 60
	} else if !isDay && planet == "Moon" {
		bala += 60
	} else if planet == "Mercury" {
		bala += 30
	}

	// 4. Abda Bala (yearly strength) — simplified, max 15
	// Lord of the year gets 15
	// Skip for simplicity — add 7.5 to all as baseline
	bala += 7.5

	// 5. Masa Bala (monthly strength) — simplified, max 30
	bala += 15

	// 6. Vara Bala (weekday strength) — simplified, max 45
	bala += 22.5

	// 7. Hora Bala (hour strength) — simplified, max 60
	if isDay && dayPlanets[planet] {
		bala += 30
	} else if !isDay && nightPlanets[planet] {
		bala += 30
	}

	// 8. Ayana Bala (declination strength) — max 30
	// Simplified: planets near 0° Cancer/Capricorn get strength
	distFromSolstice := math.Abs(lon - 90) // Cancer
	if distFromSolstice > 90 {
		distFromSolstice = math.Abs(lon - 270) // Capricorn
	}
	if distFromSolstice > 90 {
		distFromSolstice = 180 - distFromSolstice
	}
	bala += 30 * (1 - distFromSolstice/90)

	return bala
}

// ── Chesta Bala (Motional Strength) ────────────────────────────────────────

func computeChestaBala(planet string, speed float64) float64 {
	// Retrograde planets get full 60
	if speed < 0 {
		return 60
	}

	// Direct planets: strength proportional to speed relative to mean
	meanSpeeds := map[string]float64{
		"Sun": 0.9856, "Moon": 13.176, "Mercury": 1.383,
		"Venus": 1.202, "Mars": 0.524, "Jupiter": 0.083, "Saturn": 0.034,
	}

	mean, ok := meanSpeeds[planet]
	if !ok {
		return 30
	}

	// Speed ratio: faster than mean = stronger
	ratio := math.Abs(speed) / mean
	if ratio > 2 {
		ratio = 2
	}
	return math.Min(60, ratio*30)
}

// ── Naisargika Bala (Natural Strength) ─────────────────────────────────────

func computeNaisargikaBala(planet string) float64 {
	// Inherent luminosity — fixed values from Parashara
	natural := map[string]float64{
		"Sun": 60, "Moon": 51.43, "Venus": 42.85,
		"Jupiter": 34.28, "Mercury": 25.71, "Mars": 17.14, "Saturn": 8.57,
	}
	if v, ok := natural[planet]; ok {
		return v
	}
	return 0
}

// ── Drik Bala (Aspectual Strength) ─────────────────────────────────────────

func computeDrikBala(planet string, planetLons map[string]float64, planetHouses map[string]int) float64 {
	// Strength from aspects received
	// Benefic aspects add, malefic aspects subtract
	benefics := map[string]bool{"Jupiter": true, "Venus": true, "Mercury": true, "Moon": true}
	malefics := map[string]bool{"Sun": true, "Mars": true, "Saturn": true}

	myHouse, ok := planetHouses[planet]
	if !ok {
		return 0
	}

	var bala float64
	for aspPlanet, aspHouse := range planetHouses {
		if aspPlanet == planet {
			continue
		}

		// Vedic aspects: each planet aspects certain houses from its position
		aspectedHouses := vedicAspectHousesAbs(aspPlanet, aspHouse)
		for _, h := range aspectedHouses {
			if h == myHouse {
				if benefics[aspPlanet] {
					bala += 15 // Benefic aspect = +15
				} else if malefics[aspPlanet] {
					bala -= 15 // Malefic aspect = -15
				}
			}
		}
	}

	return bala
}

// vedicAspectHousesAbs returns which absolute houses a planet aspects from its position.
// Uses the canonical vedicAspectHouses from vedic_drishti.go and converts relative
// house offsets to absolute house numbers.
func vedicAspectHousesAbs(planet string, house int) []int {
	offsets := vedicAspectHouses(planet)
	abs := make([]int, len(offsets))
	for i, off := range offsets {
		abs[i] = ((house + off - 2) % 12) + 1
	}
	return abs
}

// ── Helpers ────────────────────────────────────────────────────────────────

func exaltationDegree(planet string) float64 {
	exalt := map[string]float64{
		"Sun": 10, "Moon": 33, "Mars": 298, "Mercury": 165,
		"Jupiter": 95, "Venus": 357, "Saturn": 200,
	}
	if v, ok := exalt[planet]; ok {
		return v
	}
	return -1
}

func isFriend(planet, other string) bool {
	friends := map[string]map[string]bool{
		"Sun":     {"Moon": true, "Mars": true, "Jupiter": true},
		"Moon":    {"Sun": true, "Mercury": true},
		"Mars":    {"Sun": true, "Moon": true, "Jupiter": true},
		"Mercury": {"Sun": true, "Venus": true},
		"Jupiter": {"Sun": true, "Moon": true, "Mars": true},
		"Venus":   {"Mercury": true, "Saturn": true},
		"Saturn":  {"Mercury": true, "Venus": true},
	}
	if f, ok := friends[planet]; ok {
		return f[other]
	}
	return false
}

func isEnemy(planet, other string) bool {
	enemies := map[string]map[string]bool{
		"Sun":     {"Venus": true, "Saturn": true},
		"Moon":    {}, // Moon has no enemies
		"Mars":    {"Mercury": true},
		"Mercury": {"Moon": true},
		"Jupiter": {"Mercury": true, "Venus": true},
		"Venus":   {"Sun": true, "Moon": true},
		"Saturn":  {"Sun": true, "Moon": true, "Mars": true},
	}
	if e, ok := enemies[planet]; ok {
		return e[other]
	}
	return false
}
