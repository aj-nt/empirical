package dignity

// ── Batch Operations ────────────────────────────────────────────────────

// BatchPerson holds a person's name and natal planet positions.
type BatchPerson struct {
	Name        string
	PlanetLongs map[string]float64
}

// BatchTransitResult wraps transit hits for one person.
type BatchTransitResult struct {
	Name string
	Hits []TransitHit
}

// BatchTransits runs transit scans for multiple people against the same
// date range, transit planets, aspects, and orb.
func BatchTransits(
	people []BatchPerson,
	natalPlanets []string,
	startDate, endDate string,
	aspects []AspectDef,
	orbDeg float64,
	compute ComputeFunc,
) []BatchTransitResult {
	var results []BatchTransitResult
	for _, p := range people {
		hits, err := ScanTransits(p.PlanetLongs, natalPlanets, startDate, endDate, aspects, orbDeg, compute)
		if err != nil {
			// Skip people with errors — caller can inspect results
			continue
		}
		results = append(results, BatchTransitResult{
			Name: p.Name,
			Hits: hits,
		})
	}
	return results
}

// BatchSynastryResult wraps synastry hits for one pair.
type BatchSynastryResult struct {
	Name1 string
	Name2 string
	Hits  []SynastryHit
}

// BatchSynastry computes inter-aspects for every pair in the list.
// Pair order is stable: for indices i < j, result has Name1=people[i].Name,
// Name2=people[j].Name.
func BatchSynastry(
	people []BatchPerson,
	planets []string,
	aspects []AspectDef,
	orbDeg float64,
) []BatchSynastryResult {
	var results []BatchSynastryResult
	for i := 0; i < len(people); i++ {
		for j := i + 1; j < len(people); j++ {
			hits := ComputeSynastry(people[i].PlanetLongs, people[j].PlanetLongs, planets, aspects, orbDeg)
			results = append(results, BatchSynastryResult{
				Name1: people[i].Name,
				Name2: people[j].Name,
				Hits:  hits,
			})
		}
	}
	return results
}
