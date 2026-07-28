package dignity

import (
	"fmt"
	"math"
	"time"
)

// ═══════════════════════════════════════════════════════════════════════════
// Zodiacal Releasing (ἀπόλυσις ζῳδιακή) — Valens, Anthology IV
// ═══════════════════════════════════════════════════════════════════════════
//
// Zodiacal Releasing is a Hellenistic timing technique that releases periods
// from the Lot of Fortune (body/circumstance) or Lot of Spirit (career/action).
//
// Mechanics:
//   L1 (years): Starting from the Lot's sign, each sign releases a period
//               equal to the minor years of its domicile ruler.
//   L2 (months): Within each L1 sign, the same sequence repeats in months.
//   L3 (days): Within each L2 sign, the same sequence repeats in days.
//
// Peak periods: when the releasing sign is angular (1,4,7,10) from the Lot.
// LB (Loosing of the Bond): when the releasing sign is in a difficult house
//   (6,8,12) from the Lot, especially under malefic rulership.
//
// Minor years (Valens, Anthology III.3):
//   Sun 19, Moon 25, Mercury 20, Venus 8, Mars 15, Jupiter 12, Saturn 30

// MinorYears maps planet names to their minor periods in years.
var MinorYears = map[string]float64{
	"Sun":     19,
	"Moon":    25,
	"Mercury": 20,
	"Venus":   8,
	"Mars":    15,
	"Jupiter": 12,
	"Saturn":  30,
}

// ZRPeriod represents one zodiacal releasing period.
type ZRPeriod struct {
	Sign       string  `json:"sign"`        // Sign name (Aries, Taurus, ...)
	SignIndex  int     `json:"sign_index"`   // 0-11
	Ruler      string  `json:"ruler"`        // Domicile ruler of the sign
	MinorYears float64 `json:"minor_years"`  // Minor years of the ruler
	StartDate  string  `json:"start_date"`   // ISO 8601 date
	EndDate    string  `json:"end_date"`     // ISO 8601 date
	IsPeak     bool    `json:"is_peak"`      // Angular from Lot (1,4,7,10)
	IsLB       bool    `json:"is_lb"`        // Loosing of the Bond (6,8,12 from Lot)
	Level      int     `json:"level"`        // 1, 2, or 3
	SubPeriods []ZRPeriod `json:"sub_periods,omitempty"` // L2 for L1, L3 for L2
}

// ZRReport is the full Zodiacal Releasing report.
type ZRReport struct {
	Name      string    `json:"name"`
	Lot       string    `json:"lot"`        // "Fortune" or "Spirit"
	LotSign   string    `json:"lot_sign"`
	LotDegree float64   `json:"lot_degree"`
	LotLon    float64   `json:"lot_lon"`    // Ecliptic longitude
	BirthDate string    `json:"birth_date"`
	L1Periods []ZRPeriod `json:"l1_periods"`
	// Current period at target date (if provided)
	CurrentL1 *ZRPeriod `json:"current_l1,omitempty"`
	CurrentL2 *ZRPeriod `json:"current_l2,omitempty"`
	CurrentL3 *ZRPeriod `json:"current_l3,omitempty"`
}

// signRulers is defined in profection.go — maps sign names to domicile rulers.

// ComputeZodiacalReleasing computes a full ZR report.
// lotType: "fortune" or "spirit"
// targetDate: optional date to find current period (empty string = skip)
func ComputeZodiacalReleasing(
	name string,
	birthDate time.Time,
	asc, sun, moon float64,
	isDayChart bool,
	lotType string,
	targetDate string,
) ZRReport {
	// ── Compute the Lot ────────────────────────────────────────────────
	var lotLon float64
	if lotType == "spirit" {
		if isDayChart {
			lotLon = asc + sun - moon
		} else {
			lotLon = asc + moon - sun
		}
	} else {
		// Fortune (default)
		if isDayChart {
			lotLon = asc + moon - sun
		} else {
			lotLon = asc + sun - moon
		}
	}
	for lotLon < 0 {
		lotLon += 360
	}
	for lotLon >= 360 {
		lotLon -= 360
	}

	lotSign := SignForLongitude(lotLon)
	lotSignIdx := signIndex(lotSign)
	lotDeg := math.Mod(lotLon, 30)

	report := ZRReport{
		Name:      name,
		Lot:       lotType,
		LotSign:   lotSign,
		LotDegree: lotDeg,
		LotLon:    lotLon,
		BirthDate: birthDate.Format("2006-01-02"),
	}

	// ── Generate L1 periods (with L2 and L3 sub-periods) ───────────────
	report.L1Periods = generateZRLevels(birthDate, lotSignIdx, 1, 2, nil)

	// ── Find current period if target date provided ────────────────────
	if targetDate != "" {
		target, err := time.Parse("2006-01-02", targetDate)
		if err == nil {
			report.CurrentL1, report.CurrentL2, report.CurrentL3 = findCurrentZR(
				report.L1Periods, target,
			)
		}
	}

	return report
}

// generateZRLevels generates ZR periods for a given level starting from a sign.
// level: 1 (years), 2 (months), 3 (days)
// maxDepth: how many sub-levels to generate (0 = no subs, 1 = one level down, etc.)
// parentEnd: if non-nil, stop generating when periods exceed this date
func generateZRLevels(startDate time.Time, startSignIdx int, level int, maxDepth int, parentEnd *time.Time) []ZRPeriod {
	var periods []ZRPeriod
	currentDate := startDate

	// Repeat the 12-sign cycle as needed to fill the parent period
	for {
		cycleStart := currentDate
		anyAdded := false

		for i := 0; i < 12; i++ {
			// If we have a parent end and we've passed it, stop
			if parentEnd != nil && !currentDate.Before(*parentEnd) {
				return periods
			}

			signIdx := (startSignIdx + i) % 12
			signName := Signs[signIdx]
			ruler := signRulers[signName]
			minorYears := MinorYears[ruler]

			var endDate time.Time
			switch level {
			case 1:
				days := int(math.Round(minorYears * 365.25))
				endDate = currentDate.AddDate(0, 0, days)
			case 2:
				months := int(math.Round(minorYears))
				endDate = currentDate.AddDate(0, months, 0)
			case 3:
				days := int(math.Round(minorYears))
				endDate = currentDate.AddDate(0, 0, days)
			}

			// Clamp to parent end if needed
			if parentEnd != nil && endDate.After(*parentEnd) {
				endDate = *parentEnd
			}

			// Determine peak/LB status
			houseFromLot := ((signIdx - startSignIdx + 12) % 12) + 1
			isPeak := houseFromLot == 1 || houseFromLot == 4 || houseFromLot == 7 || houseFromLot == 10
			isLB := houseFromLot == 6 || houseFromLot == 8 || houseFromLot == 12

			period := ZRPeriod{
				Sign:       signName,
				SignIndex:  signIdx,
				Ruler:      ruler,
				MinorYears: minorYears,
				StartDate:  currentDate.Format("2006-01-02"),
				EndDate:    endDate.Format("2006-01-02"),
				IsPeak:     isPeak,
				IsLB:       isLB,
				Level:      level,
			}

			// Generate sub-periods
			if maxDepth > 0 {
				period.SubPeriods = generateZRLevels(currentDate, signIdx, level+1, maxDepth-1, &endDate)
			}

			periods = append(periods, period)
			anyAdded = true
			currentDate = endDate

			// If clamped to parent end, we're done
			if parentEnd != nil && !currentDate.Before(*parentEnd) {
				return periods
			}
		}

		// If we completed a full cycle without adding anything, we're done
		if !anyAdded || currentDate.Equal(cycleStart) {
			break
		}

		// If no parent end, only do one cycle
		if parentEnd == nil {
			break
		}
	}

	return periods
}

// findCurrentZR finds the current L1, L2, and L3 periods for a target date.
func findCurrentZR(l1Periods []ZRPeriod, target time.Time) (*ZRPeriod, *ZRPeriod, *ZRPeriod) {
	for i := range l1Periods {
		p := &l1Periods[i]
		start, _ := time.Parse("2006-01-02", p.StartDate)
		end, _ := time.Parse("2006-01-02", p.EndDate)

		if (target.Equal(start) || target.After(start)) && target.Before(end) {
			// Found L1 — now find L2
			var currentL2, currentL3 *ZRPeriod
			for j := range p.SubPeriods {
				l2 := &p.SubPeriods[j]
				l2Start, _ := time.Parse("2006-01-02", l2.StartDate)
				l2End, _ := time.Parse("2006-01-02", l2.EndDate)

				if (target.Equal(l2Start) || target.After(l2Start)) && target.Before(l2End) {
					currentL2 = l2
					// Find L3
					for k := range l2.SubPeriods {
						l3 := &l2.SubPeriods[k]
						l3Start, _ := time.Parse("2006-01-02", l3.StartDate)
						l3End, _ := time.Parse("2006-01-02", l3.EndDate)

						if (target.Equal(l3Start) || target.After(l3Start)) && target.Before(l3End) {
							currentL3 = l3
							break
						}
					}
					break
				}
			}
			return p, currentL2, currentL3
		}
	}
	return nil, nil, nil
}

// signIndex returns the 0-based index of a sign name.
func signIndex(name string) int {
	for i, s := range Signs {
		if s == name {
			return i
		}
	}
	return 0
}

// FormatZRPeriod returns a human-readable description of a ZR period.
func FormatZRPeriod(p ZRPeriod) string {
	level := "L1"
	switch p.Level {
	case 2:
		level = "L2"
	case 3:
		level = "L3"
	}

	peak := ""
	if p.IsPeak {
		peak = " ★PEAK"
	}
	lb := ""
	if p.IsLB {
		lb = " ⚠LB"
	}

	return fmt.Sprintf("[%s] %s (%s, %.0fy) %s → %s%s%s",
		level, p.Sign, p.Ruler, p.MinorYears, p.StartDate, p.EndDate, peak, lb)
}
