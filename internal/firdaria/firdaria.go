package firdaria

import (
	"fmt"
	"math"
)

// ── Firdaria ─────────────────────────────────────────────────────────────
//
// Persian/Medieval planetary period system. A 75-year cycle divided into
// periods ruled by the 7 classical planets + the nodes.
//
// Order differs for diurnal (Sun above horizon) vs nocturnal charts:
//   Diurnal: Sun → Venus → Mercury → Moon → Saturn → Jupiter → Mars → NN → SN
//   Nocturnal: Moon → Saturn → Jupiter → Mars → Sun → Venus → Mercury → NN → SN
//
// Each period has sub-periods in the same order.
// Total cycle: 75 years.

// Firdaria periods in years
var firdariaYears = map[string]float64{
	"Sun":    10,
	"Venus":  8,
	"Mercury": 13,
	"Moon":   9,
	"Saturn": 11,
	"Jupiter": 12,
	"Mars":   7,
	"NorthNode": 3,
	"SouthNode": 2,
}

var diurnalOrder = []string{"Sun", "Venus", "Mercury", "Moon", "Saturn", "Jupiter", "Mars", "NorthNode", "SouthNode"}
var nocturnalOrder = []string{"Moon", "Saturn", "Jupiter", "Mars", "Sun", "Venus", "Mercury", "NorthNode", "SouthNode"}

// FirdariaPeriod represents a single firdaria period or sub-period.
type FirdariaPeriod struct {
	Planet string  `json:"planet"`
	Start  string  `json:"start"`
	End    string  `json:"end"`
	Years  float64 `json:"years"`
	Level  string  `json:"level"` // "major" or "sub"
}

// FirdariaReport is the full firdaria analysis.
type FirdariaReport struct {
	Name       string           `json:"name"`
	Diurnal    bool             `json:"diurnal"`
	Order      []string         `json:"order"`
	MajorPeriods []FirdariaPeriod `json:"major_periods"`
	SubPeriods   []FirdariaPeriod `json:"sub_periods"`
}

// ComputeFirdaria computes the full firdaria sequence from birth.
// sunAboveHorizon: true if Sun is above the horizon (diurnal chart).
// birthYear, birthMonth, birthDay: birth date.
func ComputeFirdaria(name string, sunAboveHorizon bool, birthYear, birthMonth, birthDay int) FirdariaReport {
	var order []string
	if sunAboveHorizon {
		order = diurnalOrder
	} else {
		order = nocturnalOrder
	}

	birthJD := julianDay(birthYear, birthMonth, birthDay)

	// Major periods
	majorPeriods := computePeriods(order, firdariaYears, birthJD, "major")

	// Sub-periods: for each major period, divide into sub-periods
	// Each sub-period = (major_years / 75) * sub_planet_years
	var subPeriods []FirdariaPeriod
	for _, major := range majorPeriods {
		majorStartJD := dateToJD(major.Start)
		majorEndJD := dateToJD(major.End)
		majorSpan := majorEndJD - majorStartJD

		// Sub-periods follow the same order, starting from the major planet
		startIdx := -1
		for i, p := range order {
			if p == major.Planet {
				startIdx = i
				break
			}
		}

		subCurrentJD := majorStartJD
		for i := 0; i < len(order); i++ {
			subPlanet := order[(startIdx+i)%len(order)]
			subYears := firdariaYears[subPlanet]
			subSpan := majorSpan * (subYears / 75.0)
			subEndJD := subCurrentJD + subSpan

			subPeriods = append(subPeriods, FirdariaPeriod{
				Planet: subPlanet,
				Start:  jdToDate(subCurrentJD),
				End:    jdToDate(subEndJD),
				Years:  math.Round(subSpan/365.25*100) / 100,
				Level:  "sub",
			})
			subCurrentJD = subEndJD
		}
	}

	return FirdariaReport{
		Name:         name,
		Diurnal:      sunAboveHorizon,
		Order:        order,
		MajorPeriods: majorPeriods,
		SubPeriods:   subPeriods,
	}
}

func computePeriods(order []string, years map[string]float64, startJD float64, level string) []FirdariaPeriod {
	var periods []FirdariaPeriod
	currentJD := startJD
	for _, planet := range order {
		y := years[planet]
		endJD := currentJD + y*365.25
		periods = append(periods, FirdariaPeriod{
			Planet: planet,
			Start:  jdToDate(currentJD),
			End:    jdToDate(endJD),
			Years:  y,
			Level:  level,
		})
		currentJD = endJD
	}
	return periods
}

// ── Julian Day helpers ───────────────────────────────────────────────────

func julianDay(year, month, day int) float64 {
	a := (14 - month) / 12
	y := year + 4800 - a
	m := month + 12*a - 3
	jd := float64(day) + float64(153*m+2)/5.0 + float64(365*y) + float64(y/4) - float64(y/100) + float64(y/400) - 32045.0
	return jd
}

func jdToDate(jd float64) string {
	l := int(jd) + 68569
	n := 4 * l / 146097
	l = l - (146097*n+3)/4
	i := 4000 * (l + 1) / 1461001
	l = l - 1461*i/4 + 31
	j := 80 * l / 2447
	day := l - 2447*j/80
	l = j / 11
	month := j + 2 - 12*l
	year := 100*(n-49) + i + l
	return fmt.Sprintf("%04d-%02d-%02d", year, month, day)
}

func dateToJD(dateStr string) float64 {
	var y, m, d int
	fmt.Sscanf(dateStr, "%04d-%02d-%02d", &y, &m, &d)
	return julianDay(y, m, d)
}
