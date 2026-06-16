package dignity

import "math"

// ── Synastry Engine ──────────────────────────────────────────────────────

// SynastryHit records an inter-aspect between two charts.
type SynastryHit struct {
	Planet1 string  `json:"planet1"`
	Planet2 string  `json:"planet2"`
	Aspect  string  `json:"aspect"`
	Orb     float64 `json:"orb"`
}

// ComputeSynastry computes inter-aspects between two natal charts.
//
// chart1: person 1's planet → tropical longitude (Planet1 in output)
// chart2: person 2's planet → tropical longitude (Planet2 in output)
// planets: which planets to check (must exist in both charts)
// aspects: which aspects to detect
// orbDeg: max orb in degrees
func ComputeSynastry(
	chart1, chart2 map[string]float64,
	planets []string,
	aspects []AspectDef,
	orbDeg float64,
) []SynastryHit {
	var hits []SynastryHit

	for _, p1 := range planets {
		lon1, ok1 := chart1[p1]
		if !ok1 {
			continue
		}
		for _, p2 := range planets {
			lon2, ok2 := chart2[p2]
			if !ok2 {
				continue
			}
			dist := angleDist(lon1, lon2)
			for _, asp := range aspects {
				diff := math.Abs(dist - asp.Angle)
				if diff <= orbDeg {
					hits = append(hits, SynastryHit{
						Planet1: p1,
						Planet2: p2,
						Aspect:  asp.Name,
						Orb:     math.Round(diff*100) / 100,
					})
				}
			}
		}
	}
	return hits
}

// ── Synastry Metrics ─────────────────────────────────────────────────────

// SynastryMetricResult holds the result of a single synastry metric.
type SynastryMetricResult struct {
	Metric      string  `json:"metric"`
	FamilyMean  float64 `json:"family_mean"`
	RandomMean  float64 `json:"random_mean"`
	RandomSD    float64 `json:"random_sd"`
	ZScore      float64 `json:"z_score"`
	PValue      float64 `json:"p_value_approx"`
	Significant bool    `json:"significant"`
}

// ComputeSynastryMetrics runs multiple synastry metrics on a set of pairs
// and compares each against a random baseline.
// pairs: list of (name1, chart1, name2, chart2) tuples
// randomPairs: list of random pair charts for baseline
func ComputeSynastryMetrics(
	pairs []struct {
		Name1  string
		Chart1 map[string]float64
		Name2  string
		Chart2 map[string]float64
	},
	randomPairs []struct {
		Name1  string
		Chart1 map[string]float64
		Name2  string
		Chart2 map[string]float64
	},
	planets []string,
) []SynastryMetricResult {
	aspects := DefaultAspects()
	conjOnly := []AspectDef{{Name: "conjunction", Angle: 0}}

	var results []SynastryMetricResult

	// Metric 1: Total aspects at 3° orb
	results = append(results, runMetric("total_aspects_3deg", pairs, randomPairs, planets, aspects, 3.0))

	// Metric 2: Conjunctions only at 3° orb
	results = append(results, runMetric("conjunctions_3deg", pairs, randomPairs, planets, conjOnly, 3.0))

	// Metric 3: Saturn contacts at 3° orb
	results = append(results, runMetric("saturn_contacts_3deg", pairs, randomPairs, planets, aspects, 3.0,
		func(h SynastryHit) bool { return h.Planet1 == "Saturn" || h.Planet2 == "Saturn" }))

	// Metric 4: Node contacts at 3° orb
	results = append(results, runMetric("node_contacts_3deg", pairs, randomPairs, planets, aspects, 3.0,
		func(h SynastryHit) bool { return h.Planet1 == "Node" || h.Planet2 == "Node" }))

	// Metric 5: Sun-Moon contacts at 3° orb
	results = append(results, runMetric("sun_moon_contacts_3deg", pairs, randomPairs, planets, aspects, 3.0,
		func(h SynastryHit) bool {
			return (h.Planet1 == "Sun" && h.Planet2 == "Moon") || (h.Planet1 == "Moon" && h.Planet2 == "Sun")
		}))

	return results
}

func runMetric(
	name string,
	pairs, randomPairs []struct {
		Name1  string
		Chart1 map[string]float64
		Name2  string
		Chart2 map[string]float64
	},
	planets []string,
	aspects []AspectDef,
	orb float64,
	filters ...func(SynastryHit) bool,
) SynastryMetricResult {
	// Family mean
	var familyTotal float64
	for _, p := range pairs {
		hits := ComputeSynastry(p.Chart1, p.Chart2, planets, aspects, orb)
		if len(filters) > 0 {
			filter := filters[0]
			count := 0
			for _, h := range hits {
				if filter(h) {
					count++
				}
			}
			familyTotal += float64(count)
		} else {
			familyTotal += float64(len(hits))
		}
	}
	familyMean := familyTotal / float64(len(pairs))

	// Random baseline
	var randomTotal float64
	var randomSamples []float64
	for _, rp := range randomPairs {
		hits := ComputeSynastry(rp.Chart1, rp.Chart2, planets, aspects, orb)
		var count float64
		if len(filters) > 0 {
			filter := filters[0]
			for _, h := range hits {
				if filter(h) {
					count++
				}
			}
		} else {
			count = float64(len(hits))
		}
		randomTotal += count
		randomSamples = append(randomSamples, count)
	}
	randomMean := randomTotal / float64(len(randomPairs))

	// SD
	var sumSq float64
	for _, s := range randomSamples {
		diff := s - randomMean
		sumSq += diff * diff
	}
	randomSD := math.Sqrt(sumSq / float64(len(randomPairs)))

	// Z-score
	var zScore float64
	if randomSD > 0 {
		zScore = (familyMean - randomMean) / randomSD
	}

	// Approximate p-value (two-tailed, normal approximation)
	pValue := 1.0
	if math.Abs(zScore) < 4.0 {
		pValue = 2.0 * (1.0 - normalCDF(math.Abs(zScore)))
	} else {
		pValue = 0.0
	}

	significant := math.Abs(zScore) >= 2.0

	return SynastryMetricResult{
		Metric:      name,
		FamilyMean:  familyMean,
		RandomMean:  randomMean,
		RandomSD:    randomSD,
		ZScore:      zScore,
		PValue:      pValue,
		Significant: significant,
	}
}

// normalCDF approximates the standard normal cumulative distribution.
func normalCDF(x float64) float64 {
	// Abramowitz and Stegun approximation
	a1 := 0.254829592
	a2 := -0.284496736
	a3 := 1.421413741
	a4 := -1.453152027
	a5 := 1.061405429
	p := 0.3275911

	sign := 1.0
	if x < 0 {
		sign = -1.0
	}
	x = math.Abs(x) / math.Sqrt(2.0)

	t := 1.0 / (1.0 + p*x)
	y := 1.0 - (((((a5*t+a4)*t)+a3)*t+a2)*t+a1)*t*math.Exp(-x*x)

	return 0.5 * (1.0 + sign*y)
}
