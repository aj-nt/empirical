package dignity

import (
	"encoding/json"
	"math"
)

// ── VedicNatalReport ──────────────────────────────────────────────────────
//
// VedicNatalReport assembles a complete Vedic (Jyotish) natal horoscope.
// Unlike ChartInterpretation (used by Koiné/Western), this is Vedic-specific:
// dignity convergence, nakshatras, dasha periods, whole-sign houses,
// navamsha positions, and the ascendant nakshatra.

// VedicNatalPlanet holds one planet's Vedic natal data.
type VedicNatalPlanet struct {
	Planet         string `json:"planet"`
	SiderealLon    float64 `json:"sidereal_lon"`
	SiderealSign   string `json:"sidereal_sign"`
	Nakshatra      string `json:"nakshatra"`
	NakshatraPada  int    `json:"nakshatra_pada"`
	NakshatraRuler string `json:"nakshatra_ruler"`
	NavamshaSign   string `json:"navamsha_sign"`
	House          int    `json:"house"`           // whole-sign from sidereal ASC
	Dignity        string `json:"dignity"`         // Vedic dignity state
	WesternDignity string `json:"western_dignity"`
	Convergence    string `json:"convergence"`     // agree / diverge / western_only
}

// VedicNatalDasha holds one mahadasha period.
type VedicNatalDasha struct {
	Planet string  `json:"planet"`
	Start  string  `json:"start"`
	End    string  `json:"end"`
	Years  float64 `json:"years"`
}

// VedicNatalAscendant holds the ascendant's Vedic data.
type VedicNatalAscendant struct {
	SiderealLon    float64 `json:"sidereal_lon"`
	SiderealSign   string  `json:"sidereal_sign"`
	Nakshatra      string  `json:"nakshatra"`
	NakshatraPada  int     `json:"nakshatra_pada"`
	NakshatraRuler string  `json:"nakshatra_ruler"`
}

// VedicNatalReport is the complete Vedic natal horoscope.
type VedicNatalReport struct {
	Name         string              `json:"name"`
	Ayanamsa     float64             `json:"ayanamsa"`
	Ascendant    VedicNatalAscendant `json:"ascendant"`
	Planets      []VedicNatalPlanet  `json:"planets"`
	Dasha        []VedicNatalDasha   `json:"dasha"`
	SignalCount  int                 `json:"signal_count"`
	TotalPlanets int                 `json:"total_planets"`
}

// JSON serializes the report to JSON.
func (r *VedicNatalReport) JSON() ([]byte, error) {
	return json.Marshal(r)
}

// ── ComputeVedicNatalReport ──────────────────────────────────────────────

// ComputeVedicNatalReport assembles a complete Vedic natal horoscope from a
// BaseChart. It combines dignity convergence, nakshatras, dasha periods,
// whole-sign houses, and navamsha positions into a single report.
func ComputeVedicNatalReport(bc *BaseChart) *VedicNatalReport {
	report := &VedicNatalReport{
		Name:     bc.Name,
		Ayanamsa: bc.Ayanamsa,
	}

	// ── Ascendant ──────────────────────────────────────────────────────
	sidASC := NormalizeLon(bc.ASC - bc.Ayanamsa)
	ascSignIdx := int(sidASC / 30) % 12
	ascSign := Signs[ascSignIdx]
	ascNak := GetNakshatra(sidASC)
	report.Ascendant = VedicNatalAscendant{
		SiderealLon:    sidASC,
		SiderealSign:   ascSign,
		Nakshatra:      ascNak.Nakshatra,
		NakshatraPada:  ascNak.Pada,
		NakshatraRuler: ascNak.Ruler,
	}

	// ── Dignity convergence ────────────────────────────────────────────
	planetLons := TropicalToLonMap(bc.Tropical)
	dc := ComputeDignityConvergence(planetLons, bc.Ayanamsa, bc.Name)

	// Build a lookup: planet → PlanetDignity
	dignityMap := make(map[string]PlanetDignity, len(dc.Planets))
	for _, p := range dc.Planets {
		dignityMap[p.Planet] = p
	}

	// ── Planets ────────────────────────────────────────────────────────
	// Use classical planets in order
	classicalOrder := []string{"Sun", "Moon", "Mercury", "Venus", "Mars", "Jupiter", "Saturn", "Node"}

	var moonNak string
	var moonDegInNak float64

	for _, planetName := range classicalOrder {
		pos, ok := bc.Tropical[planetName]
		if !ok {
			continue
		}

		sidLon := NormalizeLon(pos.Lon - bc.Ayanamsa)
		sidSignIdx := int(sidLon / 30) % 12
		sidSign := Signs[sidSignIdx]

		// Nakshatra
		nak := GetNakshatra(sidLon)

		// Navamsha
		_, navSign := navamshaPosition(sidLon)

		// Whole-sign house from sidereal ASC
		house := ((int(sidLon/30) - int(sidASC/30) + 12) % 12) + 1

		// Dignity
		dignity := ""
		westernDignity := ""
		convergence := ""
		if pd, ok := dignityMap[planetName]; ok {
			dignity = pd.Vedic
			westernDignity = pd.Western
			convergence = pd.Convergence
		}

		report.Planets = append(report.Planets, VedicNatalPlanet{
			Planet:         planetName,
			SiderealLon:    sidLon,
			SiderealSign:   sidSign,
			Nakshatra:      nak.Nakshatra,
			NakshatraPada:  nak.Pada,
			NakshatraRuler: nak.Ruler,
			NavamshaSign:   navSign,
			House:          house,
			Dignity:        dignity,
			WesternDignity: westernDignity,
			Convergence:    convergence,
		})

		if planetName == "Moon" {
			moonNak = nak.Nakshatra
			moonDegInNak = nak.DegreeInNakshatra
		}
	}

	// ── Dasha ──────────────────────────────────────────────────────────
	if moonNak != "" {
		dashaEntries := ComputeVimshottariDasha(
			NakshatraPosition{Nakshatra: moonNak, DegreeInNakshatra: moonDegInNak},
			bc.Year, bc.Month, bc.Day,
		)
		for _, d := range dashaEntries {
			report.Dasha = append(report.Dasha, VedicNatalDasha{
				Planet: d.Planet,
				Start:  d.Start,
				End:    d.End,
				Years:  d.Years,
			})
		}
	}

	// ── Summary ────────────────────────────────────────────────────────
	report.TotalPlanets = len(report.Planets)
	report.SignalCount = dc.SignalCount()

	return report
}

// ── Navamsha ──────────────────────────────────────────────────────────────

// navamshaPosition computes the D9 position from a sidereal longitude.
// Returns navamsha longitude and sign name.
func navamshaPosition(siderealLon float64) (float64, string) {
	sidLon := math.Mod(siderealLon, 360)
	signIdx := int(sidLon / 30.0) % 12
	degInSign := math.Mod(sidLon, 30.0)
	padaNum := int(degInSign / (30.0 / 9.0)) // 0-8

	// Element-based starting sign
	element := signIdx % 4
	var startSign int
	switch element {
	case 0: // Fire
		startSign = 0 // Aries
	case 1: // Earth
		startSign = 9 // Capricorn
	case 2: // Air
		startSign = 6 // Libra
	default: // Water
		startSign = 3 // Cancer
	}

	navSignIdx := (startSign + padaNum) % 12
	navSign := Signs[navSignIdx]

	// Navamsha longitude: sign start + proportional position within the 3°20' segment
	segFrac := math.Mod(degInSign, 30.0/9.0) / (30.0 / 9.0)
	navLon := float64(navSignIdx)*30.0 + segFrac*30.0

	return navLon, navSign
}
