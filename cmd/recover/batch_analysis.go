package main

import (
	"encoding/json"

	"github.com/aj-nt/empirical/internal/dignity"
)

// BatchAnalysisResult holds metrics for multiple charts plus aggregate stats.
type BatchAnalysisResult struct {
	Charts   []BatchChartResult `json:"charts"`
	Aggregates map[string]dignity.ResearchBaseline `json:"aggregates"`
}

// BatchChartResult holds metrics for a single chart in a batch.
type BatchChartResult struct {
	Name    string                  `json:"name"`
	Metrics *dignity.ResearchMetrics `json:"metrics"`
}

// computeBatchAnalysisJSON computes research metrics for multiple charts
// and returns aggregate statistics across the batch.
func computeBatchAnalysisJSON(charts []dignity.BirthData, cacheDir string) ([]byte, error) {
	results := make([]BatchChartResult, 0, len(charts))

	// Collect all metric values for aggregate computation
	aggValues := map[string][]float64{
		"cross_system_sign_agreement": {},
		"draconic_bridge_count":       {},
		"paran_count":                 {},
		"declination_parallel_count":  {},
		"arabic_parts_survivor_pct":   {},
		"mansion_convergence_count":   {},
		"stars_cross_survivor_pct":    {},
		"aspect_pattern_count":        {},
		"harmonic_4":                  {},
		"harmonic_5":                  {},
		"harmonic_7":                  {},
		"harmonic_9":                  {},
	}

	for _, bd := range charts {
		cd := computePositions(bd.Year, bd.Month, bd.Day, bd.Hour, bd.Minute, bd.TZOffset, bd.Lat, bd.Lng, cacheDir)
		rm := dignity.ComputeResearchMetrics(cd)

		results = append(results, BatchChartResult{
			Name:    bd.Name,
			Metrics: rm,
		})

		// Collect values for aggregates
		aggValues["cross_system_sign_agreement"] = append(aggValues["cross_system_sign_agreement"], rm.CrossSystemSignAgreement)
		aggValues["draconic_bridge_count"] = append(aggValues["draconic_bridge_count"], float64(rm.DraconicBridgeCount))
		aggValues["paran_count"] = append(aggValues["paran_count"], float64(rm.ParanCount))
		aggValues["declination_parallel_count"] = append(aggValues["declination_parallel_count"], float64(rm.DeclinationParallelCount))
		aggValues["arabic_parts_survivor_pct"] = append(aggValues["arabic_parts_survivor_pct"], rm.ArabicPartsSurvivorPct)
		aggValues["mansion_convergence_count"] = append(aggValues["mansion_convergence_count"], float64(rm.MansionConvergenceCount))
		aggValues["stars_cross_survivor_pct"] = append(aggValues["stars_cross_survivor_pct"], rm.StarsCrossSurvivorPct)
		aggValues["aspect_pattern_count"] = append(aggValues["aspect_pattern_count"], float64(rm.AspectPatternCount))
		aggValues["harmonic_4"] = append(aggValues["harmonic_4"], float64(rm.HarmonicConjunctions[4]))
		aggValues["harmonic_5"] = append(aggValues["harmonic_5"], float64(rm.HarmonicConjunctions[5]))
		aggValues["harmonic_7"] = append(aggValues["harmonic_7"], float64(rm.HarmonicConjunctions[7]))
		aggValues["harmonic_9"] = append(aggValues["harmonic_9"], float64(rm.HarmonicConjunctions[9]))
	}

	// Compute aggregate baselines
	aggregates := make(map[string]dignity.ResearchBaseline)
	for metric, values := range aggValues {
		if len(values) > 0 {
			aggregates[metric] = dignity.ComputeBaseline(metric, values)
		}
	}

	return json.Marshal(BatchAnalysisResult{
		Charts:     results,
		Aggregates: aggregates,
	})
}
