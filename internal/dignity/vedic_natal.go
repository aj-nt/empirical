package dignity

import (
	"encoding/json"
	"math"
	"time"
)

// ── VedicNatalReport ──────────────────────────────────────────────────────
//
// VedicNatalReport assembles a complete Vedic (Jyotish) natal horoscope.
// Unlike ChartInterpretation (used by Koiné/Western), this is Vedic-specific:
// dignity convergence, nakshatras, dasha periods, whole-sign houses,
// navamsha positions, and the ascendant nakshatra.

// VedicNatalPlanet holds one planet's Vedic natal data.
type VedicNatalPlanet struct {
	Planet            string `json:"planet"`
	SiderealLon       float64 `json:"sidereal_lon"`
	SiderealSign      string `json:"sidereal_sign"`
	Nakshatra         string `json:"nakshatra"`
	NakshatraPada     int    `json:"nakshatra_pada"`
	NakshatraRuler    string `json:"nakshatra_ruler"`
	NakshatraLordHouse int   `json:"nakshatra_lord_house"` // where the nakshatra ruler sits
	NavamshaSign      string `json:"navamsha_sign"`
	House             int    `json:"house"`           // whole-sign from sidereal ASC
	Retrograde        bool   `json:"retrograde"`
	Combust           bool   `json:"combust"`
	Dignity           string `json:"dignity"`         // Vedic dignity state
	WesternDignity    string `json:"western_dignity"`
	Convergence       string `json:"convergence"`     // agree / diverge / western_only
}

// VedicNatalDasha holds one dasha period (mahadasha or antardasha).
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
	HouseLords   map[int]string      `json:"house_lords"` // house number → ruling planet
	Planets      []VedicNatalPlanet  `json:"planets"`
	Dasha        []VedicNatalDasha   `json:"dasha"`
	Antardasha   []VedicNatalDasha   `json:"antardasha"`  // sub-periods of current mahadasha
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

	// ── House lords ────────────────────────────────────────────────────
	report.HouseLords = make(map[int]string, 12)
	for h := 1; h <= 12; h++ {
		houseSignIdx := (ascSignIdx + h - 1) % 12
		houseSign := Signs[houseSignIdx]
		report.HouseLords[h] = signRuler(houseSign)
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
	classicalOrder := []string{"Sun", "Moon", "Mercury", "Venus", "Mars", "Jupiter", "Saturn", "Node"}

	// Build a lookup: planet → house (for nakshatra lord placement)
	planetHouse := make(map[string]int, len(classicalOrder))

	var moonNak string
	var moonDegInNak float64

	// First pass: compute all planet data, build planetHouse map
	type planetData struct {
		sidLon   float64
		sidSign  string
		nak      NakshatraPosition
		navSign  string
		house    int
		retro    bool
		combust  bool
		dignity  string
		westDign string
		conv     string
	}
	planetDataMap := make(map[string]planetData, len(classicalOrder))

	// Get Sun longitude for combust check
	var sunSidLon float64
	if sunPos, ok := bc.Tropical["Sun"]; ok {
		sunSidLon = NormalizeLon(sunPos.Lon - bc.Ayanamsa)
	}

	for _, planetName := range classicalOrder {
		pos, ok := bc.Tropical[planetName]
		if !ok {
			continue
		}

		sidLon := NormalizeLon(pos.Lon - bc.Ayanamsa)
		sidSignIdx := int(sidLon / 30) % 12
		sidSign := Signs[sidSignIdx]
		nak := GetNakshatra(sidLon)
		_, navSign := navamshaPosition(sidLon)
		house := ((int(sidLon/30) - int(sidASC/30) + 12) % 12) + 1

		// Retrograde
		retro := pos.Speed < 0

		// Combust: within 8° of Sun (not Sun, Moon, or Node)
		combust := false
		if planetName != "Sun" && planetName != "Moon" && planetName != "Node" {
			dist := math.Abs(sidLon - sunSidLon)
			if dist > 180 {
				dist = 360 - dist
			}
			combust = dist < 8.0
		}

		dignity := ""
		westDign := ""
		conv := ""
		if pd, ok := dignityMap[planetName]; ok {
			dignity = pd.Vedic
			westDign = pd.Western
			conv = pd.Convergence
		}

		planetDataMap[planetName] = planetData{
			sidLon: sidLon, sidSign: sidSign, nak: nak, navSign: navSign,
			house: house, retro: retro, combust: combust,
			dignity: dignity, westDign: westDign, conv: conv,
		}
		planetHouse[planetName] = house

		if planetName == "Moon" {
			moonNak = nak.Nakshatra
			moonDegInNak = nak.DegreeInNakshatra
		}
	}

	// Second pass: build planet entries with nakshatra lord house
	for _, planetName := range classicalOrder {
		pd, ok := planetDataMap[planetName]
		if !ok {
			continue
		}

		// Nakshatra lord house
		nlHouse := 0
		if pd.nak.Ruler != "" {
			if h, ok := planetHouse[pd.nak.Ruler]; ok {
				nlHouse = h
			}
		}

		report.Planets = append(report.Planets, VedicNatalPlanet{
			Planet:            planetName,
			SiderealLon:       pd.sidLon,
			SiderealSign:      pd.sidSign,
			Nakshatra:         pd.nak.Nakshatra,
			NakshatraPada:     pd.nak.Pada,
			NakshatraRuler:    pd.nak.Ruler,
			NakshatraLordHouse: nlHouse,
			NavamshaSign:      pd.navSign,
			House:             pd.house,
			Retrograde:        pd.retro,
			Combust:           pd.combust,
			Dignity:           pd.dignity,
			WesternDignity:    pd.westDign,
			Convergence:       pd.conv,
		})
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

		// ── Antardasha (current mahadasha sub-periods) ──────────────────
		report.Antardasha = computeAntardasha(dashaEntries)
	}

	// ── Summary ────────────────────────────────────────────────────────
	report.TotalPlanets = len(report.Planets)
	report.SignalCount = dc.SignalCount()

	return report
}

// ── Antardasha ────────────────────────────────────────────────────────────

// computeAntardasha returns the antardasha (bhukti) sub-periods for the
// current mahadasha. Each mahadasha is divided into 9 antardashas in the
// Vimshottari order, starting with the mahadasha lord.
func computeAntardasha(dashaEntries []VimshottariDashaEntry) []VedicNatalDasha {
	// Find current mahadasha
	today := time.Now().Format("2006-01-02")
	var current VimshottariDashaEntry
	found := false
	for _, d := range dashaEntries {
		if d.Start <= today && today < d.End {
			current = d
			found = true
			break
		}
	}
	if !found {
		return nil
	}

	// Find starting index in Vimshottari order
	startIdx := -1
	for i, p := range vedicVimshottariOrder {
		if p == current.Planet {
			startIdx = i
			break
		}
	}
	if startIdx < 0 {
		return nil
	}

	// Parse current mahadasha start
	start, err := time.Parse("2006-01-02", current.Start)
	if err != nil {
		return nil
	}

	var result []VedicNatalDasha
	cursor := start

	for i := 0; i < 9; i++ {
		planet := vedicVimshottariOrder[(startIdx+i)%9]
		// Antardasha duration = (mahadasha_years * planet_years) / 120
		adYears := (current.Years * vedicVimshottariPeriods[planet]) / 120.0
		end := cursor.Add(time.Duration(adYears*365.25*24) * time.Hour)

		result = append(result, VedicNatalDasha{
			Planet: planet,
			Start:  cursor.Format("2006-01-02"),
			End:    end.Format("2006-01-02"),
			Years:  math.Round(adYears*100) / 100,
		})
		cursor = end
	}

	return result
}

// ── Sign ruler ────────────────────────────────────────────────────────────

// signRuler returns the Vedic ruler of a sign.
func signRuler(sign string) string {
	switch sign {
	case "Aries", "Scorpio":
		return "Mars"
	case "Taurus", "Libra":
		return "Venus"
	case "Gemini", "Virgo":
		return "Mercury"
	case "Cancer":
		return "Moon"
	case "Leo":
		return "Sun"
	case "Sagittarius", "Pisces":
		return "Jupiter"
	case "Capricorn", "Aquarius":
		return "Saturn"
	}
	return ""
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
