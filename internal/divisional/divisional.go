package divisional

import (
	"fmt"
	"math"
)

// ── Vedic Divisional Charts ──────────────────────────────────────────────
//
// Divisional charts (vargas) divide each sign into segments and reassign
// planets to new signs based on which segment they fall in.
//
// Navamsha (D9) is the most important — each sign divided into 9 parts
// of 3°20' each. Used for marriage, dharma, and inner capacity.
//
// Element-based starting sign rule:
//   Fire signs (Aries, Leo, Sag) → D9 starts at Aries
//   Earth signs (Taurus, Virgo, Cap) → D9 starts at Capricorn
//   Air signs (Gemini, Libra, Aqu) → D9 starts at Libra
//   Water signs (Cancer, Scorpio, Pisces) → D9 starts at Cancer

var zodiacSigns = []string{
	"Aries", "Taurus", "Gemini", "Cancer",
	"Leo", "Virgo", "Libra", "Scorpio",
	"Sagittarius", "Capricorn", "Aquarius", "Pisces",
}

// NavamshaPosition computes the D9 position from a sidereal longitude.
// Returns navamsha longitude and sign name.
func NavamshaPosition(siderealLon float64) (float64, string) {
	sidLon := math.Mod(siderealLon, 360)
	signIdx := int(sidLon / 30.0) % 12
	degInSign := math.Mod(sidLon, 30.0)
	padaNum := int(degInSign / (30.0 / 9.0)) // 0-8
	if padaNum > 8 {
		padaNum = 8
	}

	// Element-based starting sign for navamsha sequence
	var navStart int
	switch signIdx {
	case 0, 4, 8: // Fire: Aries, Leo, Sagittarius
		navStart = 0 // Aries
	case 1, 5, 9: // Earth: Taurus, Virgo, Capricorn
		navStart = 9 // Capricorn
	case 2, 6, 10: // Air: Gemini, Libra, Aquarius
		navStart = 6 // Libra
	default: // Water: Cancer, Scorpio, Pisces
		navStart = 3 // Cancer
	}

	navSign := (navStart + padaNum) % 12
	navDeg := math.Mod(degInSign, 30.0/9.0) * 9.0 // Scale degree within navamsha
	navLongitude := float64(navSign)*30.0 + navDeg

	return navLongitude, zodiacSigns[navSign]
}

// ── Nakshatras ───────────────────────────────────────────────────────────

// 27 Nakshatras, each spanning 13.333... degrees
var Nakshatras = []string{
	"Ashwini", "Bharani", "Krittika", "Rohini", "Mrigashirsha", "Ardra",
	"Punarvasu", "Pushya", "Ashlesha", "Magha", "Purva Phalguni", "Uttara Phalguni",
	"Hasta", "Chitra", "Swati", "Vishakha", "Anuradha", "Jyeshtha",
	"Mula", "Purva Ashadha", "Uttara Ashadha", "Shravana", "Dhanishta", "Shatabhisha",
	"Purva Bhadrapada", "Uttara Bhadrapada", "Revati",
}

const NakshatraSpan = 360.0 / 27.0 // 13.333... degrees

// Vimshottari order: 9-planet cycle repeats 3x across 27 nakshatras
var VimshottariOrder = []string{
	"Ketu", "Venus", "Sun", "Moon", "Mars",
	"Rahu", "Jupiter", "Saturn", "Mercury",
}

// NakshatraRulers is the full 27-nakshatra ruler list (9-planet cycle × 3).
var NakshatraRulers = func() []string {
	rulers := make([]string, 27)
	for i := 0; i < 27; i++ {
		rulers[i] = VimshottariOrder[i%9]
	}
	return rulers
}()

// NakshatraPosition computes nakshatra and pada from sidereal longitude.
type NakshatraInfo struct {
	Nakshatra         string  `json:"nakshatra"`
	Pada              int     `json:"pada"`
	DegreeInNakshatra float64 `json:"degree_in_nakshatra"`
	Ruler             string  `json:"ruler"`
}

func GetNakshatra(siderealLon float64) NakshatraInfo {
	sidLon := math.Mod(siderealLon, 360)
	nakshatraIdx := int(sidLon/NakshatraSpan) % 27
	degInNak := math.Mod(sidLon, NakshatraSpan)
	pada := int(degInNak/(NakshatraSpan/4.0)) + 1
	if pada > 4 {
		pada = 4
	}

	return NakshatraInfo{
		Nakshatra:         Nakshatras[nakshatraIdx],
		Pada:              pada,
		DegreeInNakshatra: degInNak,
		Ruler:             NakshatraRulers[nakshatraIdx],
	}
}

// ── Vimshottari Dasha ────────────────────────────────────────────────────

// VimshottariPeriods maps planet to dasha period in years.
var VimshottariPeriods = map[string]int{
	"Ketu": 7, "Venus": 20, "Sun": 6, "Moon": 10, "Mars": 7,
	"Rahu": 18, "Jupiter": 16, "Saturn": 19, "Mercury": 17,
}

// DashaPeriod represents a single mahadasha period.
type DashaPeriod struct {
	Planet string  `json:"planet"`
	Start  string  `json:"start"`
	End    string  `json:"end"`
	Years  float64 `json:"years"`
}

// VimshottariDasha computes the full 120-year dasha sequence from birth.
// Moon's nakshatra determines the starting mahadasha.
// The proportion already traversed at birth determines the elapsed portion.
func VimshottariDasha(moonNakshatra string, degreeInNakshatra float64, birthYear, birthMonth, birthDay int) []DashaPeriod {
	// Find nakshatra index
	nakIdx := -1
	for i, n := range Nakshatras {
		if n == moonNakshatra {
			nakIdx = i
			break
		}
	}
	if nakIdx < 0 {
		return nil
	}

	ruler := NakshatraRulers[nakIdx]
	proportion := degreeInNakshatra / NakshatraSpan
	firstYears := float64(VimshottariPeriods[ruler])
	remainingYears := firstYears * (1.0 - proportion)

	// Julian Day for birth
	birthJD := julianDay(birthYear, birthMonth, birthDay)

	sequence := make([]DashaPeriod, 0, 9)
	currentJD := birthJD
	startIdx := -1
	for i, p := range VimshottariOrder {
		if p == ruler {
			startIdx = i
			break
		}
	}

	// First period (partial — remaining from birth)
	firstEndJD := currentJD + remainingYears*365.25
	sequence = append(sequence, DashaPeriod{
		Planet: ruler,
		Start:  jdToDate(birthJD),
		End:    jdToDate(firstEndJD),
		Years:  math.Round(remainingYears*100) / 100,
	})
	currentJD = firstEndJD

	// Subsequent periods (full)
	for i := 1; i < 9; i++ {
		planet := VimshottariOrder[(startIdx+i)%9]
		years := float64(VimshottariPeriods[planet])
		endJD := currentJD + years*365.25
		sequence = append(sequence, DashaPeriod{
			Planet: planet,
			Start:  jdToDate(currentJD),
			End:    jdToDate(endJD),
			Years:  years,
		})
		currentJD = endJD
	}

	return sequence
}

// ── Divisional Report ────────────────────────────────────────────────────

// DivisionalPosition holds a planet's divisional chart data.
type DivisionalPosition struct {
	Planet          string        `json:"planet"`
	SiderealLon     float64       `json:"sidereal_lon"`
	SiderealSign    string        `json:"sidereal_sign"`
	Nakshatra       NakshatraInfo `json:"nakshatra"`
	NavamshaSign    string        `json:"navamsha_sign"`
	NavamshaLon     float64       `json:"navamsha_lon"`
}

// DivisionalReport is the full Vedic divisional chart analysis.
type DivisionalReport struct {
	Name       string               `json:"name"`
	Ayanamsa   float64              `json:"ayanamsa"`
	Positions  []DivisionalPosition `json:"positions"`
	Dasha      []DashaPeriod        `json:"dasha"`
}

// ComputeDivisionalReport computes the full Vedic divisional analysis.
func ComputeDivisionalReport(name string, planets map[string]float64, ayanamsa float64, birthYear, birthMonth, birthDay int) DivisionalReport {
	report := DivisionalReport{
		Name:     name,
		Ayanamsa: ayanamsa,
	}

	var moonNak string
	var moonDegInNak float64

	for planetName, tropLon := range planets {
		sidLon := math.Mod(tropLon-ayanamsa+360, 360)
		sidSignIdx := int(sidLon / 30) % 12
		sidSign := zodiacSigns[sidSignIdx]
		nak := GetNakshatra(sidLon)
		navLon, navSign := NavamshaPosition(sidLon)

		report.Positions = append(report.Positions, DivisionalPosition{
			Planet:       planetName,
			SiderealLon:  sidLon,
			SiderealSign: sidSign,
			Nakshatra:    nak,
			NavamshaSign: navSign,
			NavamshaLon:  navLon,
		})

		if planetName == "Moon" {
			moonNak = nak.Nakshatra
			moonDegInNak = nak.DegreeInNakshatra
		}
	}

	if moonNak != "" {
		report.Dasha = VimshottariDasha(moonNak, moonDegInNak, birthYear, birthMonth, birthDay)
	}

	return report
}

// ── Julian Day helpers ───────────────────────────────────────────────────

func julianDay(year, month, day int) float64 {
	// Simplified Julian Day calculation (valid for CE dates)
	a := (14 - month) / 12
	y := year + 4800 - a
	m := month + 12*a - 3
	jd := float64(day) + float64(153*m+2)/5.0 + float64(365*y) + float64(y/4) - float64(y/100) + float64(y/400) - 32045.0
	return jd
}

func jdToDate(jd float64) string {
	// Convert Julian Day back to Gregorian date
	// Algorithm from Fliegel & Van Flandern
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

	return dateStr(year, month, day)
}

func dateStr(year, month, day int) string {
	return fmt.Sprintf("%04d-%02d-%02d", year, month, day)
}
