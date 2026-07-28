package main

import (
	"encoding/json"
	"math/rand"

	"github.com/aj-nt/empirical/internal/dignity"
)

// computeResearchBaselineJSON generates N random charts, computes the specified
// metric for each, and returns the baseline distribution as JSON.
func computeResearchBaselineJSON(metric string, n int, seed int64, cacheDir string) ([]byte, error) {
	if n <= 0 {
		n = 1000
	}
	if n > 10000 {
		n = 10000
	}

	rng := rand.New(rand.NewSource(seed))
	values := make([]float64, 0, n)

	for i := 0; i < n; i++ {
		bd := randomBirthData(rng)
		cd := computePositions(bd.Year, bd.Month, bd.Day, bd.Hour, bd.Minute, bd.TZOffset, bd.Lat, bd.Lng, cacheDir)
		rm := dignity.ComputeResearchMetrics(cd)
		v := extractMetric(rm, metric)
		values = append(values, v)
	}

	baseline := dignity.ComputeBaseline(metric, values)
	return json.Marshal(baseline)
}

// randomBirthData generates a random birth data within reasonable bounds.
func randomBirthData(rng *rand.Rand) dignity.BirthData {
	return dignity.BirthData{
		Name:     "random",
		Year:     1900 + rng.Intn(200),   // 1900-2099
		Month:    1 + rng.Intn(12),
		Day:      1 + rng.Intn(28),       // safe for all months
		Hour:     rng.Intn(24),
		Minute:   rng.Intn(60),
		TZOffset: float64(-12 + rng.Intn(25)), // -12 to +12
		Lat:      -90 + rng.Float64()*180,    // -90 to 90
		Lng:      -180 + rng.Float64()*360,   // -180 to 180
	}
}

// extractMetric pulls a single metric value from ResearchMetrics by name.
func extractMetric(rm *dignity.ResearchMetrics, metric string) float64 {
	switch metric {
	case "cross_system_sign_agreement":
		return rm.CrossSystemSignAgreement
	case "draconic_bridge_count":
		return float64(rm.DraconicBridgeCount)
	case "paran_count":
		return float64(rm.ParanCount)
	case "declination_parallel_count":
		return float64(rm.DeclinationParallelCount)
	case "arabic_parts_survivor_pct":
		return rm.ArabicPartsSurvivorPct
	case "mansion_convergence_count":
		return float64(rm.MansionConvergenceCount)
	case "stars_cross_survivor_pct":
		return rm.StarsCrossSurvivorPct
	case "aspect_pattern_count":
		return float64(rm.AspectPatternCount)
	default:
		// Check harmonic conjunctions: "harmonic_4", "harmonic_5", etc.
		if len(metric) > 9 && metric[:9] == "harmonic_" {
			h := 0
			for _, c := range metric[9:] {
				if c >= '0' && c <= '9' {
					h = h*10 + int(c-'0')
				}
			}
			if h > 0 {
				return float64(rm.HarmonicConjunctions[h])
			}
		}
		return 0
	}
}
