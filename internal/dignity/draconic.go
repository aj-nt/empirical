package dignity

import "math"

// ── Draconic Chart ────────────────────────────────────────────────────────

// DraconicChart holds the draconic (soul-level) planetary positions.
// The draconic zodiac is the tropical zodiac rotated so the North Node
// sits at 0° Aries. Every planet is shifted by the same offset.
type DraconicChart struct {
	Planets map[string]float64 `json:"planets"`
	Offset  float64            `json:"offset"`
}

// SignShift records a planet that changes zodiac sign under the draconic rotation.
type SignShift struct {
	Planet   string `json:"planet"`
	TropSign string `json:"tropical_sign"`
	DracSign string `json:"draconic_sign"`
}

// ComputeDraconic rotates a tropical chart into the draconic zodiac.
// nnLong is the tropical longitude of the Mean North Node in degrees.
// The draconic chart includes a "Node" entry at 0°.
func ComputeDraconic(tropical map[string]float64, nnLong float64) *DraconicChart {
	offset := normalizeLon(nnLong)
	drac := &DraconicChart{
		Planets: make(map[string]float64),
		Offset:  offset,
	}

	for name, lon := range tropical {
		drac.Planets[name] = normalizeLon(lon - offset)
	}
	drac.Planets["Node"] = 0.0

	return drac
}

// ComputeDraconicSignShifts returns planets that change zodiac sign
// under the draconic rotation. Planets that stay in the same sign
// are omitted from the result.
func ComputeDraconicSignShifts(tropical map[string]float64, nnLong float64) map[string]SignShift {
	offset := normalizeLon(nnLong)
	shifts := make(map[string]SignShift)

	for name, tropLon := range tropical {
		tropSign := SignForLongitude(tropLon)
		dracLon := normalizeLon(tropLon - offset)
		dracSign := SignForLongitude(dracLon)
		if tropSign != dracSign {
			shifts[name] = SignShift{
				Planet:   name,
				TropSign: tropSign,
				DracSign: dracSign,
			}
		}
	}

	return shifts
}

// ComputeDraconicBridges finds aspects between the draconic chart and the
// tropical chart — soul-to-personality links. Same-name pairs (draconic Sun
// to tropical Sun) are excluded because they are always ~offset degrees apart
// and are tautological. TNP names (Cupido, Hades, etc.) are also excluded.
func ComputeDraconicBridges(
	tropical map[string]float64,
	nnLong float64,
	planets []string,
	aspects []AspectDef,
	orbDeg float64,
) []SynastryHit {
	drac := ComputeDraconic(tropical, nnLong)

	// Build filtered planet list: exclude TNPs
	var filtered []string
	for _, p := range planets {
		if !isTNP(p) {
			filtered = append(filtered, p)
		}
	}

	// Compute synastry between draconic (chart1) and tropical (chart2)
	allHits := ComputeSynastry(drac.Planets, tropical, filtered, aspects, orbDeg)

	// Filter out same-name pairs
	var bridges []SynastryHit
	for _, h := range allHits {
		if h.Planet1 != h.Planet2 {
			bridges = append(bridges, h)
		}
	}

	return bridges
}

// ComputeDraconicSynastry computes inter-aspects between two draconic charts.
// Each chart is independently rotated by its own North Node offset.
func ComputeDraconicSynastry(
	tropicalA map[string]float64, nnA float64,
	tropicalB map[string]float64, nnB float64,
	planets []string,
	aspects []AspectDef,
	orbDeg float64,
) []SynastryHit {
	dracA := ComputeDraconic(tropicalA, nnA)
	dracB := ComputeDraconic(tropicalB, nnB)
	return ComputeSynastry(dracA.Planets, dracB.Planets, planets, aspects, orbDeg)
}

// isTNP returns true if the planet name is a Trans-Neptunian Point
// (Hamburg School hypothetical). These are excluded from bridges
// because they are single-culture constructs with no soul-level meaning.
func isTNP(name string) bool {
	switch name {
	case "Cupido", "Hades", "Zeus", "Kronos", "Apollon", "Admetos", "Poseidon", "Vulkanus":
		return true
	}
	return false
}

// DraconicToTropical converts a draconic longitude back to tropical.
// dracLon is the draconic ecliptic longitude in degrees (0-360).
// nnLong is the North Node's tropical longitude in degrees.
func DraconicToTropical(dracLon, nnLong float64) float64 {
	return normalizeLon(dracLon + normalizeLon(nnLong))
}

// ComputeProgressedDraconic computes the draconic chart using a different
// North Node offset — typically the current transiting NN instead of the
// natal NN. This reveals how the soul's orientation has evolved over time.
// The math is identical to ComputeDraconic; only the offset changes.
func ComputeProgressedDraconic(tropical map[string]float64, progressedNN float64) *DraconicChart {
	return ComputeDraconic(tropical, progressedNN)
}

// DraconicSynastryFull holds the three-layer draconic synastry result.
type DraconicSynastryFull struct {
	DracToDrac   []SynastryHit `json:"drac_to_drac"`
	TropAToDracB []SynastryHit `json:"trop_a_to_drac_b"`
	TropBToDracA []SynastryHit `json:"trop_b_to_drac_a"`
}

// ComputeDraconicSynastryFull computes the full three-layer draconic synastry:
//   1. drac-to-drac: soul-to-soul (both shifted to their own NN)
//   2. tropA-to-dracB: person A's life vs person B's soul agenda
//   3. tropB-to-dracA: person B's life vs person A's soul agenda
// Unlike bridges (same-person tropical-to-draconic), same-name pairs
// are NOT filtered here — Sun-to-Sun between two people is a real contact.
func ComputeDraconicSynastryFull(
	tropicalA map[string]float64, nnA float64,
	tropicalB map[string]float64, nnB float64,
	planets []string,
	aspects []AspectDef,
	orbDeg float64,
) DraconicSynastryFull {
	dracA := ComputeDraconic(tropicalA, nnA)
	dracB := ComputeDraconic(tropicalB, nnB)

	// Filter TNPs
	var filtered []string
	for _, p := range planets {
		if !isTNP(p) {
			filtered = append(filtered, p)
		}
	}

	return DraconicSynastryFull{
		DracToDrac:   ComputeSynastry(dracA.Planets, dracB.Planets, filtered, aspects, orbDeg),
		TropAToDracB: ComputeSynastry(tropicalA, dracB.Planets, filtered, aspects, orbDeg),
		TropBToDracA: ComputeSynastry(tropicalB, dracA.Planets, filtered, aspects, orbDeg),
	}
}

// ── Cross-System Draconic Transit Comparison ─────────────────────────────

// CrossSystemHit records a transit-to-draconic aspect in one zodiac system.
type CrossSystemHit struct {
	TransitPlanet string  `json:"transit_planet"`
	NatalPlanet   string  `json:"natal_planet"`
	Aspect        string  `json:"aspect"`
	Orb           float64 `json:"orb"`
}

// CrossSystemResult holds the comparison of draconic transit contacts
// between tropical and sidereal zodiacs. Survivors are aspects that
// appear in both systems — cross-system signal. TropicalOnly and
// SiderealOnly are zodiac-dependent.
type CrossSystemResult struct {
	Survivors    []CrossSystemHit `json:"survivors"`
	TropicalOnly []CrossSystemHit `json:"tropical_only"`
	SiderealOnly []CrossSystemHit `json:"sidereal_only"`
}

// CompareCrossSystemTransits compares transiting planet contacts against
// a natal draconic chart in both tropical and sidereal zodiacs.
// Natal draconic positions are zodiac-invariant (the ayanamsa cancels out).
// Transiting positions differ by ~24° (Lahiri ayanamsa).
// Returns which aspects survive the zodiac shift and which are zodiac-dependent.
func CompareCrossSystemTransits(
	natalDrac map[string]float64,
	tropTransits map[string]float64,
	sidTransits map[string]float64,
	aspects []AspectDef,
	orb float64,
) *CrossSystemResult {
	// Find tropical hits
	tropHits := findTransitAspects(natalDrac, tropTransits, aspects, orb)

	// Find sidereal hits
	sidHits := findTransitAspects(natalDrac, sidTransits, aspects, orb)

	// Classify: survivors appear in both, others are zodiac-dependent
	result := &CrossSystemResult{}

	// Build a lookup key for sidereal hits
	sidKey := make(map[string]bool)
	for _, h := range sidHits {
		key := h.TransitPlanet + "|" + h.NatalPlanet + "|" + h.Aspect
		sidKey[key] = true
	}

	for _, h := range tropHits {
		key := h.TransitPlanet + "|" + h.NatalPlanet + "|" + h.Aspect
		if sidKey[key] {
			result.Survivors = append(result.Survivors, h)
		} else {
			result.TropicalOnly = append(result.TropicalOnly, h)
		}
	}

	// Find sidereal-only (in sidereal but not tropical)
	tropKey := make(map[string]bool)
	for _, h := range tropHits {
		key := h.TransitPlanet + "|" + h.NatalPlanet + "|" + h.Aspect
		tropKey[key] = true
	}
	for _, h := range sidHits {
		key := h.TransitPlanet + "|" + h.NatalPlanet + "|" + h.Aspect
		if !tropKey[key] {
			result.SiderealOnly = append(result.SiderealOnly, h)
		}
	}

	return result
}

// findTransitAspects finds all aspects between transiting planets and
// natal draconic positions.
func findTransitAspects(
	natalDrac map[string]float64,
	transits map[string]float64,
	aspects []AspectDef,
	orb float64,
) []CrossSystemHit {
	var hits []CrossSystemHit
	for tp, tlon := range transits {
		for np, nlon := range natalDrac {
			dist := angleDist(tlon, nlon)
			for _, a := range aspects {
				diff := math.Abs(dist - a.Angle)
				if diff <= orb {
					hits = append(hits, CrossSystemHit{
						TransitPlanet: tp,
						NatalPlanet:   np,
						Aspect:        a.Name,
						Orb:           math.Round(diff*100) / 100,
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


