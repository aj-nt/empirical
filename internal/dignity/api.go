package dignity

import (
	"encoding/json"
	"time"
)

// ── Combined API: Multi-phase report ──────────────────────────────────────

// FullReport collects all phase results for a single chart into one structure.
type FullReport struct {
	Name            string            `json:"name"`
	AyanamsaDegrees float64           `json:"ayanamsa_degrees"`
	Phase1Dignity   *DignityConvergence  `json:"phase1_dignity"`
	Phase3Houses    *HouseConvergence    `json:"phase3_houses"`
	Phase5Nodes     *NodeConvergence     `json:"phase5_nodes"`
	Phase6Zodiac    *ZodiacComparison    `json:"phase6_zodiac"`
}

// FullReportJSON serializes a FullReport to JSON.
func (fr *FullReport) FullReportJSON() ([]byte, error) {
	return json.Marshal(fr)
}

// ComputeFullReport computes all available phases for a chart.
// tropicalLons maps planet name → tropical longitude.
// ayan is the Lahiri ayanamsa in degrees.
// nnLong is the tropical longitude of the Mean North Node.
// ascLong is the tropical ASC longitude.
func ComputeFullReport(
	tropicalLons map[string]float64,
	ayan float64,
	nnLong float64,
	ascLong float64,
	name string,
	year, month, day, hour, minute, second int,
	tzOffset, lat, lng float64,
) *FullReport {
	fr := &FullReport{
		Name:            name,
		AyanamsaDegrees: ayan,
	}

	// Phase 1: Dignity convergence
	fr.Phase1Dignity = ComputeDignityConvergence(tropicalLons, ayan, name)

	// Phase 3: House division convergence
	fr.Phase3Houses = ComputeHouseConvergence(tropicalLons, year, month, day, hour, minute, second, tzOffset, lat, lng, name)

	// Phase 5: Node axis convergence
	fr.Phase5Nodes = ComputeNodeConvergence(nnLong, ayan, name)

	// Phase 6: Zodiac comparison
	fr.Phase6Zodiac = ComputeZodiacComparison(tropicalLons, ayan, name)

	return fr
}

// ── Timing Report for Phase 4 ─────────────────────────────────────────────

// TimingReport collects timing convergence for a target date.
type TimingReport struct {
	Name            string             `json:"name"`
	TargetDate      string             `json:"target_date"`
	TimingConvergence *TimingConvergence `json:"timing_convergence"`
}

// TimingReportJSON serializes a TimingReport to JSON.
func (tr *TimingReport) TimingReportJSON() ([]byte, error) {
	return json.Marshal(tr)
}

// ComputeTimingReport computes Phase 4 timing convergence for a target date.
func ComputeTimingReport(
	name string,
	year, month, day, hour, minute int,
	tzOffset, lat, lng float64,
	targetDate string,
	tropicalLons map[string]float64,
	ayan float64,
	ascLong float64,
) *TimingReport {
	tr := &TimingReport{
		Name:       name,
		TargetDate: targetDate,
	}

	// Get Moon sidereal longitude for nakshatra
	moonLon := 0.0
	if ml, ok := tropicalLons["Moon"]; ok {
		moonLon = normalizeLon(ml - ayan)
	}
	nak := GetNakshatra(moonLon)

	// Compute Vimshottari dasha from birth
	dashaSeq := ComputeVimshottariDasha(nak, year, month, day)
	currentDasha := CurrentDasha(dashaSeq, targetDate)

	// Compute Ba Zi pillars
	pillars := ComputeBaZiPillars(year, month, day, hour)

	// Compute profection
	td, _ := time.Parse("2006-01-02", targetDate)
	prof := ComputeProfection(year, month, day, td, ascLong)

	// Compute timing convergence
	tr.TimingConvergence = ComputeTimingConvergence(
		targetDate, name, currentDasha, pillars,
		year, month, day, prof,
	)

	return tr
}
