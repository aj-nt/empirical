package dignity

import (
	"encoding/json"
	"fmt"

	"github.com/aj-nt/empirical/internal/swe"
)

// ── Phase 3: House Division Convergence ─────────────────────────────────────
//
// Five house systems compared: whole_sign, equal, placidus, porphyry, koch.
// All systems use the tropical zodiac. A planet is "unambiguous" if 4 out of
// 5 systems agree on its house (signal). "Disputed" means 3 or fewer agree
// (noise).
//
// Whole-sign houses are the only system common to Western and Vedic
// traditions. Quadrant systems (Placidus, Porphyry, Koch) are exclusively
// Western. If the original system used a house method, whole-sign is the
// strongest candidate for the invariant. But house placement is the weakest
// invariance layer — treat with higher uncertainty than aspect geometry
// or dignity.

// CompareHouseSystems lists all house systems used in the convergence measurement.
// The original five (whole_sign, equal, placidus, porphyry, koch) are extended
// with three Medieval systems: regiomontanus, alcabitius, campanus.
var CompareHouseSystems = []string{"whole_sign", "equal", "placidus", "porphyry", "koch", "regiomontanus", "alcabitius", "campanus"}

// sweph house codes
var swephCode = map[string]byte{
	"placidus":      'P',
	"porphyry":      'O',
	"koch":          'K',
	"equal":         'E',
	"whole_sign":    'W',
	"regiomontanus": 'R',
	"alcabitius":    'B',
	"campanus":      'C',
}

// PlanetHouse records a single planet's house placement across all systems.
type PlanetHouse struct {
	Planet       string
	TropicalSign string
	Placements   map[string]int // system_name → house_number (1-12)
}

// AgreementCount returns how many systems agree on the most common house.
func (ph *PlanetHouse) AgreementCount() int {
	house := ph.ConsensusHouse()
	count := 0
	for _, h := range ph.Placements {
		if h == house {
			count++
		}
	}
	return count
}

// ConsensusHouse returns the most common house placement across all systems.
func (ph *PlanetHouse) ConsensusHouse() int {
	counts := make(map[int]int)
	for _, h := range ph.Placements {
		counts[h]++
	}
	best := 0
	bestH := 0
	for h, c := range counts {
		if c > best || (c == best && h < bestH) {
			best = c
			bestH = h
		}
	}
	return bestH
}

// IsUnambiguous returns true if at least 75% of systems agree on the house.
// For 5 systems: 4+ (80%). For 8 systems: 6+ (75%).
func (ph *PlanetHouse) IsUnambiguous() bool {
	n := len(ph.Placements)
	threshold := int(float64(n) * 0.75)
	if threshold < 2 {
		threshold = 2
	}
	return ph.AgreementCount() >= threshold
}

// IsDisputed returns true if fewer than 75% of systems agree.
func (ph *PlanetHouse) IsDisputed() bool {
	return !ph.IsUnambiguous()
}

// AgreementRatio returns fraction of systems agreeing (0.0-1.0).
func (ph *PlanetHouse) AgreementRatio() float64 {
	return float64(ph.AgreementCount()) / float64(len(ph.Placements))
}

// HouseConvergence holds the full house convergence report for a chart.
type HouseConvergence struct {
	Name    string
	Planets []PlanetHouse
}

// UnambiguousCount returns how many planets have >=4/5 agreement.
func (hc *HouseConvergence) UnambiguousCount() int {
	count := 0
	for _, p := range hc.Planets {
		if p.IsUnambiguous() {
			count++
		}
	}
	return count
}

// DisputedCount returns how many planets have <=3/5 agreement.
func (hc *HouseConvergence) DisputedCount() int {
	count := 0
	for _, p := range hc.Planets {
		if p.IsDisputed() {
			count++
		}
	}
	return count
}

// ConvergenceRate returns fraction of unambiguous planets.
func (hc *HouseConvergence) ConvergenceRate() float64 {
	if len(hc.Planets) == 0 {
		return 0
	}
	return float64(hc.UnambiguousCount()) / float64(len(hc.Planets))
}

// UnambiguousPlanets returns planet names with >=4/5 agreement.
func (hc *HouseConvergence) UnambiguousPlanets() []string {
	var out []string
	for _, p := range hc.Planets {
		if p.IsUnambiguous() {
			out = append(out, p.Planet)
		}
	}
	return out
}

// DisputedPlanets returns planet names with <=3/5 agreement.
func (hc *HouseConvergence) DisputedPlanets() []string {
	var out []string
	for _, p := range hc.Planets {
		if p.IsDisputed() {
			out = append(out, p.Planet)
		}
	}
	return out
}

// WholeSignQuadrantMatch counts planets where the whole-sign house matches
// at least one quadrant system (placidus/porphyry/koch).
func (hc *HouseConvergence) WholeSignQuadrantMatch() int {
	count := 0
	for _, p := range hc.Planets {
		ws := p.Placements["whole_sign"]
		qh := map[int]bool{}
		for _, sys := range []string{"placidus", "porphyry", "koch"} {
			if h, ok := p.Placements[sys]; ok {
				qh[h] = true
			}
		}
		if qh[ws] {
			count++
		}
	}
	return count
}

// ComputeHouseConvergence computes house placement convergence across the
// five house systems. tropicalLons maps planet name → tropical ecliptic
// longitude. Birth data is used to compute Swiss Ephemeris house cusps.
func ComputeHouseConvergence(
	tropicalLons map[string]float64,
	year, month, day, hour, minute, second int,
	tzOffset, lat, lng float64,
	name string,
) *HouseConvergence {
	result := &HouseConvergence{Name: name}

	// Compute Julian Day in UT
	utHour := float64(hour) + float64(minute)/60.0 + float64(second)/3600.0 - tzOffset
	jd := swe.Julday(year, month, day, utHour, true)

	// First: get Placidus cusps and ASC/MC (needed for whole_sign and equal)
	placCusps, placAscmc := swe.Houses(jd, lat, lng, 'P')
	ascendant := placAscmc[0]

	// Build cusps for each system
	cusps := make(map[string][13]float64)

	// Placidus: from SWE directly
	cusps["placidus"] = placCusps

	// Porphyry: from SWE directly
	porphCusps, _ := swe.Houses(jd, lat, lng, swephCode["porphyry"])
	cusps["porphyry"] = porphCusps

	// Koch: from SWE directly
	kochCusps, _ := swe.Houses(jd, lat, lng, swephCode["koch"])
	cusps["koch"] = kochCusps

	// Whole sign: each cusp = 0° of successive signs starting from ASC sign
	var ws [13]float64
	ascSign := int(ascendant) / 30
	for i := 0; i < 12; i++ {
		ws[i+1] = float64(((ascSign+i)%12)*30)
	}
	cusps["whole_sign"] = ws

	// Equal: each cusp = ASC + n*30°
	var eq [13]float64
	for i := 0; i < 12; i++ {
		eq[i+1] = normalizeLon(ascendant + float64(i)*30.0)
	}
	cusps["equal"] = eq

	// Regiomontanus: from SWE directly
	regioCusps, _ := swe.Houses(jd, lat, lng, swephCode["regiomontanus"])
	cusps["regiomontanus"] = regioCusps

	// Alcabitius: from SWE directly
	alcabCusps, _ := swe.Houses(jd, lat, lng, swephCode["alcabitius"])
	cusps["alcabitius"] = alcabCusps

	// Campanus: from SWE directly
	campCusps, _ := swe.Houses(jd, lat, lng, swephCode["campanus"])
	cusps["campanus"] = campCusps

	// Compute placements for each classical planet (empirical verification uses
	// only the 7 classical planets — outer planets, asteroids, and Uranian
	// points are Western-only and would dilute the cross-system signal)
	for _, planet := range ClassicalPlanets {
		lon, ok := tropicalLons[planet]
		if !ok {
			continue
		}

		placements := make(map[string]int)
		for _, system := range CompareHouseSystems {
			placements[system] = planetInHouse(lon, cusps[system])
		}

		result.Planets = append(result.Planets, PlanetHouse{
			Planet:       planet,
			TropicalSign: SignForLongitude(lon),
			Placements:   placements,
		})
	}

	return result
}

// planetInHouse determines which house (1-12) a planet at the given
// ecliptic longitude falls in.
func planetInHouse(longitude float64, cusps [13]float64) int {
	lon := normalizeLon(longitude)

	for h := 1; h <= 12; h++ {
		next := h + 1
		if next > 12 {
			next = 1
		}
		start := cusps[h]
		end := cusps[next]

		if start <= end {
			if start <= lon && lon < end {
				return h
			}
		} else {
			// Wraps across 0°
			if lon >= start || lon < end {
				return h
			}
		}
	}
	return 12 // fallback
}

// FormatHouseConvergence formats a human-readable house convergence report.
func FormatHouseConvergence(hc *HouseConvergence) string {
	var b []byte
	b = append(b, fmt.Sprintf("House Division Convergence Report — %s\n", hc.Name)...)
	b = append(b, "(All systems using tropical zodiac; 5 systems compared)\n\n"...)

	// Header
	headerSys := ""
	for _, s := range CompareHouseSystems {
		headerSys += fmt.Sprintf("%-12s ", s)
	}
	b = append(b, fmt.Sprintf("%-10s %-10s %s%-8s Verdict\n",
		"Planet", "Sign", headerSys, "Agree")...)
	b = append(b, "—————————————————————————————————————————————————————————————————————\n"...)

	for _, p := range hc.Planets {
		place := ""
		for _, s := range CompareHouseSystems {
			place += fmt.Sprintf("H%-11d ", p.Placements[s])
		}
		agree := fmt.Sprintf("%d/%d", p.AgreementCount(), len(CompareHouseSystems))
		verdict := "SIGNAL"
		if p.IsDisputed() {
			verdict = "NOISE"
		}
		b = append(b, fmt.Sprintf("%-10s %-10s %s%-8s %s\n",
			p.Planet, p.TropicalSign, place, agree, verdict)...)
	}

	b = append(b, "\n"...)

	unamb := hc.UnambiguousPlanets()
	disp := hc.DisputedPlanets()
	unambStr := "none"
	if len(unamb) > 0 {
		unambStr = join(unamb, ", ")
	}
	dispStr := "none"
	if len(disp) > 0 {
		dispStr = join(disp, ", ")
	}

	b = append(b, fmt.Sprintf("Unambiguous (>=4/5 agree): %d/%d (%.0f%%) — %s\n",
		hc.UnambiguousCount(), len(hc.Planets),
		hc.ConvergenceRate()*100, unambStr)...)
	b = append(b, fmt.Sprintf("Disputed   (<=3/5 agree): %d/%d — %s\n",
		hc.DisputedCount(), len(hc.Planets), dispStr)...)
	b = append(b, "\n"...)

	// Assessment
	rate := hc.ConvergenceRate()
	if rate >= 0.8 {
		b = append(b, "HIGH convergence: Most placements stable across systems.\n"...)
	} else if rate >= 0.6 {
		b = append(b, "MODERATE convergence: Some placements shift with method.\n"...)
	} else {
		b = append(b, "LOW convergence: House division highly sensitive to method.\n"...)
	}

	wsMatch := hc.WholeSignQuadrantMatch()
	b = append(b, "\n"...)
	b = append(b, fmt.Sprintf("Whole-sign vs quadrant agreement: %d/%d planets have "+
		"whole-sign house matching at least one quadrant system.\n\n",
		wsMatch, len(hc.Planets))...)
	b = append(b, "RECOVERY IMPLICATION: Whole-sign houses are the only system "+
		"common to Western and Vedic traditions. Quadrant systems "+
		"(Placidus, Porphyry, Koch) are exclusively Western. If the "+
		"original used a house system, whole-sign is the strongest "+
		"candidate for the invariant. But house placement is the "+
		"weakest invariance layer.\n"...)

	return string(b)
}

// HouseConvergenceJSON serializes the house report for the API.
func (hc *HouseConvergence) HouseConvergenceJSON() ([]byte, error) {
	type outPlanet struct {
		Planet       string         `json:"planet"`
		TropicalSign string         `json:"tropical_sign"`
		Placements   map[string]int `json:"placements"`
		Consensus    int            `json:"consensus_house"`
		Agreement    int            `json:"agreement_count"`
		IsSignal     bool           `json:"is_signal"`
	}
	type out struct {
		Name                  string      `json:"name"`
		Planets               []outPlanet `json:"planets"`
		UnambiguousCount      int         `json:"unambiguous_count"`
		DisputedCount         int         `json:"disputed_count"`
		ConvergenceRate       float64     `json:"convergence_rate"`
		WholeSignQuadrantMatch int        `json:"ws_quadrant_match"`
	}
	var planets []outPlanet
	for _, p := range hc.Planets {
		planets = append(planets, outPlanet{
			Planet:       p.Planet,
			TropicalSign: p.TropicalSign,
			Placements:   p.Placements,
			Consensus:    p.ConsensusHouse(),
			Agreement:    p.AgreementCount(),
			IsSignal:     p.IsUnambiguous(),
		})
	}
	o := out{
		Name:                   hc.Name,
		Planets:                planets,
		UnambiguousCount:       hc.UnambiguousCount(),
		DisputedCount:          hc.DisputedCount(),
		ConvergenceRate:        hc.ConvergenceRate(),
		WholeSignQuadrantMatch: hc.WholeSignQuadrantMatch(),
	}
	return json.MarshalIndent(o, "", "  ")
}

// join joins strings with a separator (minimal import-free version).
func join(parts []string, sep string) string {
	if len(parts) == 0 {
		return ""
	}
	s := parts[0]
	for _, p := range parts[1:] {
		s += sep + p
	}
	return s
}
