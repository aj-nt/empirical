package dignity

import "math"

// ── Secondary Progressions ────────────────────────────────────────────────
//
// Secondary progressions use the "day-for-a-year" formula:
// progressedJD = birthJD + (targetJD - birthJD) / 365.2425
//
// The progressed chart represents the native's evolution — the planets
// as they have moved one day for each year of life.

// ProgressedReport holds a full progressed chart analysis.
type ProgressedReport struct {
	Name               string         `json:"name"`
	TargetDate         string         `json:"target_date"`
	Age                float64        `json:"age_years"`
	ProgressedJD       float64        `json:"progressed_jd"`
	ProgressedPositions map[string]float64 `json:"progressed_positions"`
	NatalPositions     map[string]float64 `json:"natal_positions"`
	Aspects            []SynastryHit  `json:"aspects"`
	Patterns           []Pattern      `json:"patterns"`
	TotalAspects       int            `json:"total_aspects"`
}

// ── Progressed Cross-System Comparison ────────────────────────────────────
//
// Secondary progressions use day-for-a-year: progressedJD = birthJD + age.
// Both natal and progressed positions shift by the same ayanamsa when
// converting between tropical and sidereal. Angular distances are preserved.
//
// This is the same geometry as fixed star invariance (Phase 11, 100% survival)
// and Arabic Parts aspect invariance (Phase 9, 100% aspect survival).
// The prediction: near-100% survival at any orb.
//
// The only source of non-survival is ayanamsa drift over the progressed
// interval. At age 90, the ayanamsa has drifted ~0.003° — negligible at 3° orb.

// ProgressedCrossHit records a progressed-to-natal aspect in one zodiac system.
type ProgressedCrossHit struct {
	ProgressedPlanet string  `json:"progressed_planet"`
	NatalPlanet      string  `json:"natal_planet"`
	Aspect           string  `json:"aspect"`
	Orb              float64 `json:"orb"`
}

// ProgressedCrossResult holds the comparison of progressed-to-natal aspects
// between tropical and sidereal zodiacs. Survivors are aspects that appear
// in both systems — cross-system signal. TropicalOnly and SiderealOnly
// are zodiac-dependent (expected near-zero due to geometry).
type ProgressedCrossResult struct {
	Survivors    []ProgressedCrossHit `json:"survivors"`
	TropicalOnly []ProgressedCrossHit `json:"tropical_only"`
	SiderealOnly []ProgressedCrossHit `json:"sidereal_only"`
}

// CompareCrossSystemProgressed compares progressed-to-natal aspects in
// tropical vs sidereal zodiacs. Both natal and progressed positions are
// shifted by the same ayanamsa, so angular distances are preserved.
// Returns which aspects survive the zodiac shift.
func CompareCrossSystemProgressed(
	natal map[string]float64,
	progressed map[string]float64,
	ayan float64,
	aspects []AspectDef,
	orb float64,
) *ProgressedCrossResult {
	// Build sidereal natal and progressed by subtracting ayanamsa
	sidNatal := make(map[string]float64)
	sidProg := make(map[string]float64)
	for k, v := range natal {
		sidNatal[k] = normalizeLon(v - ayan)
	}
	for k, v := range progressed {
		sidProg[k] = normalizeLon(v - ayan)
	}

	// Find tropical hits
	tropHits := findProgressedAspects(natal, progressed, aspects, orb)

	// Find sidereal hits
	sidHits := findProgressedAspects(sidNatal, sidProg, aspects, orb)

	// Classify: survivors appear in both
	result := &ProgressedCrossResult{}

	sidKey := make(map[string]bool)
	for _, h := range sidHits {
		key := h.ProgressedPlanet + "|" + h.NatalPlanet + "|" + h.Aspect
		sidKey[key] = true
	}

	for _, h := range tropHits {
		key := h.ProgressedPlanet + "|" + h.NatalPlanet + "|" + h.Aspect
		if sidKey[key] {
			result.Survivors = append(result.Survivors, h)
		} else {
			result.TropicalOnly = append(result.TropicalOnly, h)
		}
	}

	// Find sidereal-only
	tropKey := make(map[string]bool)
	for _, h := range tropHits {
		key := h.ProgressedPlanet + "|" + h.NatalPlanet + "|" + h.Aspect
		tropKey[key] = true
	}
	for _, h := range sidHits {
		key := h.ProgressedPlanet + "|" + h.NatalPlanet + "|" + h.Aspect
		if !tropKey[key] {
			result.SiderealOnly = append(result.SiderealOnly, h)
		}
	}

	return result
}

// findProgressedAspects finds all aspects between progressed and natal positions.
func findProgressedAspects(
	natal map[string]float64,
	progressed map[string]float64,
	aspects []AspectDef,
	orb float64,
) []ProgressedCrossHit {
	var hits []ProgressedCrossHit
	for pp, plon := range progressed {
		for np, nlon := range natal {
			dist := angleDist(plon, nlon)
			for _, a := range aspects {
				diff := math.Abs(dist - a.Angle)
				if diff <= orb {
					hits = append(hits, ProgressedCrossHit{
						ProgressedPlanet: pp,
						NatalPlanet:      np,
						Aspect:           a.Name,
						Orb:              math.Round(diff*100) / 100,
					})
				}
			}
		}
	}
	// Sort by orb
	for i := 0; i < len(hits); i++ {
		for j := i + 1; j < len(hits); j++ {
			if hits[j].Orb < hits[i].Orb {
				hits[i], hits[j] = hits[j], hits[i]
			}
		}
	}
	return hits
}

// ComputeProgressedReport computes a secondary progressed chart and
// progressed-to-natal aspects for a given target date.
// progressedPositions must be pre-computed by the caller (requires SWE).
// natalPositions are the birth chart positions.
func ComputeProgressedReport(
	name string,
	targetDate string,
	age float64,
	progressedPositions map[string]float64,
	natalPositions map[string]float64,
	orbDeg float64,
) *ProgressedReport {
	// Progressed-to-natal aspects
	aspects := DefaultAspects()
	planets := make([]string, 0, len(progressedPositions))
	for p := range progressedPositions {
		planets = append(planets, p)
	}

	hits := ComputeSynastry(progressedPositions, natalPositions, planets, aspects, orbDeg)

	// Pattern detection on progressed chart (non-TNP bodies only)
	nonTNP := make(map[string]float64)
	for p, lon := range progressedPositions {
		if !isTNP(p) {
			nonTNP[p] = lon
		}
	}
	patternReport := DetectPatterns(nonTNP, 5.0)

	return &ProgressedReport{
		Name:                name,
		TargetDate:          targetDate,
		Age:                 age,
		ProgressedPositions: progressedPositions,
		NatalPositions:      natalPositions,
		Aspects:             hits,
		Patterns:            patternReport.Patterns,
		TotalAspects:        len(hits),
	}
}
