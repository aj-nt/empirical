package dignity

import (
	"math"
	"sort"
)

// ResearchMetrics holds all research-grade metrics for a single chart.
// These are the values we compare against random baselines to answer
// "how unusual is this chart?"
type ResearchMetrics struct {
	CrossSystemSignAgreement float64        `json:"cross_system_sign_agreement"`
	DraconicBridgeCount      int            `json:"draconic_bridge_count"`
	HarmonicConjunctions     map[int]int    `json:"harmonic_conjunctions"`
	ParanCount               int            `json:"paran_count"`
	DeclinationParallelCount int            `json:"declination_parallel_count"`
	ArabicPartsSurvivorPct   float64        `json:"arabic_parts_survivor_pct"`
	MansionConvergenceCount  int            `json:"mansion_convergence_count"`
	StarsCrossSurvivorPct    float64        `json:"stars_cross_survivor_pct"`
	AspectPatternCount       int            `json:"aspect_pattern_count"`
}

// ResearchBaseline holds the distribution of a single metric across N random charts.
type ResearchBaseline struct {
	Metric string  `json:"metric"`
	N      int     `json:"n"`
	Mean   float64 `json:"mean"`
	Median float64 `json:"median"`
	StdDev float64 `json:"std_dev"`
	Min    float64 `json:"min"`
	Max    float64 `json:"max"`
	P5     float64 `json:"p5"`
	P25    float64 `json:"p25"`
	P50    float64 `json:"p50"`
	P75    float64 `json:"p75"`
	P95    float64 `json:"p95"`
}

// PercentileRank returns the percentile rank of a value within a sorted baseline.
func PercentileRank(sorted []float64, value float64) float64 {
	if len(sorted) == 0 {
		return 50
	}
	count := 0
	for _, v := range sorted {
		if v <= value {
			count++
		}
	}
	return float64(count) / float64(len(sorted)) * 100
}

// ComputeBaseline computes percentile distribution from a slice of values.
func ComputeBaseline(metric string, values []float64) ResearchBaseline {
	if len(values) == 0 {
		return ResearchBaseline{Metric: metric}
	}

	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)

	n := len(sorted)
	sum := 0.0
	for _, v := range sorted {
		sum += v
	}
	mean := sum / float64(n)

	median := sorted[n/2]
	if n%2 == 0 {
		median = (sorted[n/2-1] + sorted[n/2]) / 2
	}

	sumSq := 0.0
	for _, v := range sorted {
		d := v - mean
		sumSq += d * d
	}
	stdDev := math.Sqrt(sumSq / float64(n))

	p := func(pct float64) float64 {
		idx := int(math.Round(pct/100*float64(n-1)))
		if idx < 0 {
			idx = 0
		}
		if idx >= n {
			idx = n - 1
		}
		return sorted[idx]
	}

	return ResearchBaseline{
		Metric: metric,
		N:      n,
		Mean:   mean,
		Median: median,
		StdDev: stdDev,
		Min:    sorted[0],
		Max:    sorted[n-1],
		P5:     p(5),
		P25:    p(25),
		P50:    median,
		P75:    p(75),
		P95:    p(95),
	}
}

// ComputeResearchMetrics computes all research metrics for a BaseChart.
func ComputeResearchMetrics(bc *BaseChart) *ResearchMetrics {
	m := &ResearchMetrics{
		HarmonicConjunctions: make(map[int]int),
	}

	m.CrossSystemSignAgreement = computeCrossSystemSignAgreement(bc)
	m.DraconicBridgeCount = computeDraconicBridgeCount(bc)
	m.HarmonicConjunctions = computeHarmonicConjunctions(bc)
	m.ParanCount = computeParanCount(bc)
	m.DeclinationParallelCount = computeDeclinationParallelCount(bc)
	m.ArabicPartsSurvivorPct = computeArabicPartsSurvivorPct(bc)
	m.MansionConvergenceCount = computeMansionConvergenceCount(bc)
	m.StarsCrossSurvivorPct = computeStarsCrossSurvivorPct(bc)
	m.AspectPatternCount = computeAspectPatternCount(bc)

	return m
}

// ── Individual metric computations ──

func computeCrossSystemSignAgreement(bc *BaseChart) float64 {
	koine := ComputeDignityConvergence(TropicalToLonMap(bc.Tropical), bc.Ayanamsa, bc.Name)
	// Western uses same tropical signs as Koiné for sign agreement
	// Vedic uses sidereal signs
	vedic := VedicFromBase(bc)

	// Count planets with same sign across all three
	// Koiné and Western share tropical signs, so we compare trop vs sid
	agree := 0
	total := 0
	for _, p := range koine.Planets {
		// Find matching Vedic planet
		for _, vp := range vedic.Planets {
			if p.Planet == vp.Planet {
				total++
				if p.TropSign == vp.SidSign {
					agree++
				}
				break
			}
		}
	}
	if total == 0 {
		return 0
	}
	return float64(agree) / float64(total) * 100
}

func computeDraconicBridgeCount(bc *BaseChart) int {
	tropical := TropicalToLonMap(bc.Tropical)
	bridges := ComputeDraconicBridges(tropical, bc.NorthNode, NonTNPNoNodePlanetNames, DefaultAspects(), 1.0)
	return len(bridges)
}

func computeHarmonicConjunctions(bc *BaseChart) map[int]int {
	result := make(map[int]int)
	tropical := TropicalToLonMap(bc.Tropical)
	planets := NonTNPNoNodePlanetNames

	harmonics := []int{4, 5, 7, 9}
	for _, h := range harmonics {
		count := 0
		for i := 0; i < len(planets); i++ {
			for j := i + 1; j < len(planets); j++ {
				p1, p2 := planets[i], planets[j]
				diff := math.Abs(tropical[p1] - tropical[p2])
				if diff > 180 {
					diff = 360 - diff
				}
				target := 360.0 / float64(h)
				orb := 3.0
				mod := math.Mod(diff, target)
				if mod < orb || mod > target-orb {
					count++
				}
			}
		}
		result[h] = count
	}
	return result
}

func computeParanCount(bc *BaseChart) int {
	stars := bc.StarPositions
	if stars == nil {
		return 0
	}

	count := 0
	angles := map[string]float64{
		"ASC": bc.ASC,
		"MC":  bc.MC,
		"DSC": bc.DSC,
		"IC":  bc.IC,
	}

	orb := 2.0
	for _, starLon := range stars {
		for _, angleLon := range angles {
			diff := math.Abs(starLon - angleLon)
			if diff > 180 {
				diff = 360 - diff
			}
			if diff < orb {
				count++
			}
		}
	}

	for _, p := range NonTNPNoNodePlanetNames {
		pos, ok := bc.Tropical[p]
		if !ok {
			continue
		}
		for _, angleLon := range angles {
			diff := math.Abs(pos.Lon - angleLon)
			if diff > 180 {
				diff = 360 - diff
			}
			if diff < orb {
				count++
			}
		}
	}

	return count
}

func computeDeclinationParallelCount(bc *BaseChart) int {
	if bc.Declinations == nil {
		return 0
	}

	count := 0
	planets := NonTNPNoNodePlanetNames
	orb := 1.0

	for i := 0; i < len(planets); i++ {
		for j := i + 1; j < len(planets); j++ {
			d1, ok1 := bc.Declinations[planets[i]]
			d2, ok2 := bc.Declinations[planets[j]]
			if !ok1 || !ok2 {
				continue
			}
			if math.Abs(d1-d2) < orb {
				count++
			}
			if math.Abs(d1+d2) < orb {
				count++
			}
		}
	}
	return count
}

func computeArabicPartsSurvivorPct(bc *BaseChart) float64 {
	planets := TropicalToLonMap(bc.Tropical)
	isDay := bc.Tropical["Sun"].Lon > bc.ASC && bc.Tropical["Sun"].Lon < bc.DSC
	if bc.DSC < bc.ASC {
		isDay = bc.Tropical["Sun"].Lon > bc.ASC || bc.Tropical["Sun"].Lon < bc.DSC
	}

	report := ComputePartCrossSystem(bc.Name, bc.ASC, planets, bc.Ayanamsa, isDay, 3)
	if report.Total == 0 {
		return 0
	}
	return float64(report.SignSurvivors) / float64(report.Total) * 100
}

func computeMansionConvergenceCount(bc *BaseChart) int {
	tropical := TropicalToLonMap(bc.Tropical)
	conv := ComputeMansionConvergence(bc.Name, tropical, bc.Ayanamsa)
	return conv.Converging
}

func computeStarsCrossSurvivorPct(bc *BaseChart) float64 {
	if bc.StarPositions == nil {
		return 0
	}

	tropical := TropicalToLonMap(bc.Tropical)
	result := CompareStarConjunctionsCrossSystem(bc.Name, bc.StarPositions, tropical, bc.Ayanamsa, 3)

	total := result.TotalTrop + result.TotalSid
	if total == 0 {
		return 0
	}
	return float64(result.TotalSurvivors) / float64(total) * 100
}

func computeAspectPatternCount(bc *BaseChart) int {
	western := WesternFromBase(bc, 3, true)
	return len(western.Patterns)
}
