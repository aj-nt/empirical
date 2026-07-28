package dignity

import (
	"fmt"
	"math"
	"time"
)

// ── Annual Profections ────────────────────────────────────────────────────
//
// Annual profections are a Hellenistic timing technique. The ASC advances
// by one whole sign (30°) per year of life, starting from the birth ASC.
//
//   profectedASC = natalASC + (age × 30°)
//
// The profected sign becomes the "activated" house for that year.
// The Time Lord (chronocrator) is the planet ruling the profected sign.
//
// This is one of the oldest predictive techniques, used by Valens and
// other Hellenistic astrologers.

// ProfectionReport holds a full annual profection analysis.
type ProfectionReport struct {
	Name            string  `json:"name"`
	BirthDate       string  `json:"birth_date"`
	TargetDate      string  `json:"target_date"`
	Age             float64 `json:"age_years"`
	ProfectionYear  int     `json:"profection_year"`  // 1-indexed year of life
	NatalASC        float64 `json:"natal_asc"`
	ProfectedASC    float64 `json:"profected_asc"`
	ProfectedSign   string  `json:"profected_sign"`
	ProfectedHouse  int     `json:"profected_house"`  // 1-12, relative to natal ASC
	TimeLord        string  `json:"time_lord"`        // planet ruling the profected sign
	TimeLordHouse   int     `json:"time_lord_house"`  // house of the time lord in natal
	TimeLordSign    string  `json:"time_lord_sign"`   // sign of the time lord in natal
	// Planets in the profected sign (natal)
	PlanetsInSign   []string `json:"planets_in_sign"`
	// Transits to the profected ASC
	TransitAspects  []SynastryHit `json:"transit_aspects,omitempty"`
}

// Sign rulers using traditional domicile rulership.
var signRulers = map[string]string{
	"Aries":       "Mars",
	"Taurus":      "Venus",
	"Gemini":      "Mercury",
	"Cancer":      "Moon",
	"Leo":         "Sun",
	"Virgo":       "Mercury",
	"Libra":       "Venus",
	"Scorpio":     "Mars",
	"Sagittarius": "Jupiter",
	"Capricorn":   "Saturn",
	"Aquarius":    "Saturn",
	"Pisces":      "Jupiter",
}

// ComputeProfectionReport computes the annual profection for a given birth data and target date.
func ComputeProfectionReport(
	name string,
	birthDate time.Time,
	targetDate time.Time,
	natalASC float64,
	natalPositions map[string]float64,
) *ProfectionReport {
	age := targetDate.Sub(birthDate).Hours() / (365.2425 * 24)
	profYear := int(math.Floor(age)) + 1 // 1-indexed: birth to age 1 = 1st year

	// Profected ASC: advance by one sign (30°) per year of life
	profASC := normalizeLon(natalASC + age*30.0)

	// Profected sign
	profSign := SignForLongitude(profASC)

	// Profected house (relative to natal ASC)
	profHouse := ((int(profASC/30) - int(natalASC/30) + 12) % 12) + 1

	// Time Lord: traditional ruler of the profected sign
	timeLord := signRulers[profSign]

	// Find time lord's natal position
	var tlSign string
	var tlHouse int
	if tlLon, ok := natalPositions[timeLord]; ok {
		tlSign = SignForLongitude(tlLon)
		tlHouse = ((int(tlLon/30) - int(natalASC/30) + 12) % 12) + 1
	}

	// Find planets in the profected sign (natal)
	var planetsInSign []string
	for p, lon := range natalPositions {
		if SignForLongitude(lon) == profSign {
			planetsInSign = append(planetsInSign, p)
		}
	}

	return &ProfectionReport{
		Name:           name,
		BirthDate:      birthDate.Format("2006-01-02"),
		TargetDate:     targetDate.Format("2006-01-02"),
		Age:            math.Round(age*100) / 100,
		ProfectionYear: profYear,
		NatalASC:       math.Round(natalASC*100) / 100,
		ProfectedASC:   math.Round(profASC*100) / 100,
		ProfectedSign:  profSign,
		ProfectedHouse: profHouse,
		TimeLord:       timeLord,
		TimeLordHouse:  tlHouse,
		TimeLordSign:   tlSign,
		PlanetsInSign:  planetsInSign,
	}
}

// FormatProfection returns a human-readable summary of the profection.
func FormatProfection(r *ProfectionReport) string {
	return fmt.Sprintf(
		"Year %d of life | Profected Sign: %s (House %d) | Time Lord: %s in %s (House %d)",
		r.ProfectionYear, r.ProfectedSign, r.ProfectedHouse,
		r.TimeLord, r.TimeLordSign, r.TimeLordHouse,
	)
}
