package dignity

import "math"

// ── Arabic Parts (Lots / Sahams) ──────────────────────────────────────────
//
// Hellenistic mathematical points computed from three chart factors.
// Western tradition calls them Lots or Arabic Parts.
// Vedic tradition calls them Sahams — often identical formulas.
// Phase 9 measures cross-system transmission fidelity.

// PartDef holds the definition of one Arabic Part.
type PartDef struct {
	Name           string // e.g., "Fortune", "Spirit"
	DayBase        string // e.g., "Asc"
	DayAdd         string // e.g., "Moon"
	DaySub         string // e.g., "Sun"
	NightBase      string // e.g., "Asc" (empty string = same as day)
	NightAdd       string // e.g., "Sun"
	NightSub       string // e.g., "Moon"
	InBoth         bool   // documented in both Western and Vedic traditions
	VedicName      string // Sanskrit name if different
}

// PartPlacement holds a computed Part's position and aspects.
type PartPlacement struct {
	Part     string  `json:"part"`
	Lon      float64 `json:"lon"`
	Sign     string  `json:"sign"`
	SignNum  int     `json:"sign_num"`
}

// PartAspectHit is a Part-to-planet aspect.
type PartAspectHit struct {
	Part   string  `json:"part"`
	Planet string  `json:"planet"`
	Aspect string  `json:"aspect"`
	Orb    float64 `json:"orb"`
}

// PartReport holds the full Arabic Parts analysis for a chart.
type PartReport struct {
	Name       string           `json:"name"`
	IsDay      bool             `json:"is_day"`
	Parts      []PartPlacement  `json:"parts"`
	Aspects    []PartAspectHit  `json:"aspects"`
}

// PartCrossSystem holds cross-system Part comparison.
type PartCrossSystem struct {
	Name          string           `json:"name"`
	Ayanamsa      float64          `json:"ayanamsa"`
	IsDay         bool             `json:"is_day"`
	Tropical      []PartPlacement  `json:"tropical"`
	Sidereal      []PartPlacement  `json:"sidereal"`
	SignSurvivors int              `json:"sign_survivors"` // Parts in same sign both zodiacs
	Total         int              `json:"total"`
	Aspects       []PartAspectHit  `json:"aspects"`        // Part-to-planet aspects (zodiac-invariant)
}

// PartCatalog returns the standard Arabic Parts used in both traditions.
func PartCatalog() []PartDef {
	return []PartDef{
		{Name: "Fortune", DayBase: "Asc", DayAdd: "Moon", DaySub: "Sun",
			NightBase: "Asc", NightAdd: "Sun", NightSub: "Moon",
			InBoth: true, VedicName: "Punya Saham"},
		{Name: "Spirit", DayBase: "Asc", DayAdd: "Sun", DaySub: "Moon",
			NightBase: "Asc", NightAdd: "Moon", NightSub: "Sun",
			InBoth: true, VedicName: "Vidya Saham"},
		{Name: "Eros", DayBase: "Asc", DayAdd: "Venus", DaySub: "Sun",
			InBoth: false},
		{Name: "Necessity", DayBase: "Asc", DayAdd: "Mercury", DaySub: "Fortune",
			InBoth: false},
		{Name: "Courage", DayBase: "Asc", DayAdd: "Mars", DaySub: "Fortune",
			InBoth: false},
		{Name: "Victory", DayBase: "Asc", DayAdd: "Jupiter", DaySub: "Fortune",
			InBoth: true, VedicName: "Jaya Saham"},
		{Name: "Nemesis", DayBase: "Asc", DayAdd: "Saturn", DaySub: "Fortune",
			InBoth: false},
		{Name: "Basis", DayBase: "Asc", DayAdd: "Fortune", DaySub: "Spirit",
			InBoth: false},
		{Name: "Father", DayBase: "Asc", DayAdd: "Sun", DaySub: "Saturn",
			NightBase: "Asc", NightAdd: "Saturn", NightSub: "Sun",
			InBoth: true, VedicName: "Pitru Saham"},
		{Name: "Mother", DayBase: "Asc", DayAdd: "Moon", DaySub: "Venus",
			NightBase: "Asc", NightAdd: "Venus", NightSub: "Moon",
			InBoth: true, VedicName: "Matru Saham"},
		{Name: "Children", DayBase: "Asc", DayAdd: "Jupiter", DaySub: "Saturn",
			InBoth: true, VedicName: "Putra Saham"},
		{Name: "Marriage", DayBase: "Asc", DayAdd: "Venus", DaySub: "Saturn",
			InBoth: true, VedicName: "Kalyana Saham"},
		{Name: "Death", DayBase: "Asc", DayAdd: "Saturn", DaySub: "Moon",
			NightBase: "Asc", NightAdd: "Moon", NightSub: "Saturn",
			InBoth: true, VedicName: "Mrityu Saham"},
	}
}

// resolvePartLongitude resolves a factor name to its longitude.
// "Asc" returns the ascendant. "Fortune", "Spirit", etc. look up
// previously computed Parts in the parts map.
// Planet names look up in the planets map.
func resolvePartLongitude(factor string, asc float64, planets map[string]float64, parts map[string]float64) (float64, bool) {
	switch factor {
	case "Asc":
		return asc, true
	case "Fortune", "Spirit", "Eros", "Necessity", "Courage", "Victory",
		"Nemesis", "Basis", "Father", "Mother", "Children", "Marriage", "Death":
		lon, ok := parts[factor]
		return lon, ok
	default:
		lon, ok := planets[factor]
		return lon, ok
	}
}

// ComputeParts computes all cataloged Arabic Parts for a chart.
// Parts that depend on other Parts (Necessity, Courage, Victory, Nemesis, Basis)
// are computed in dependency order.
func ComputeParts(asc float64, planets map[string]float64, isDay bool) map[string]float64 {
	catalog := PartCatalog()
	parts := make(map[string]float64)

	// First pass: compute Parts that only depend on Asc and planets
	for _, def := range catalog {
		if def.DayAdd == "Fortune" || def.DayAdd == "Spirit" || def.DaySub == "Fortune" || def.DaySub == "Spirit" {
			continue // depends on another Part — second pass
		}
		lon := computePart(def, asc, planets, parts, isDay)
		parts[def.Name] = lon
	}

	// Second pass: compute Parts that depend on Fortune/Spirit
	for _, def := range catalog {
		if _, ok := parts[def.Name]; ok {
			continue
		}
		lon := computePart(def, asc, planets, parts, isDay)
		parts[def.Name] = lon
	}

	return parts
}

// computePart computes a single Part's longitude.
func computePart(def PartDef, asc float64, planets map[string]float64, parts map[string]float64, isDay bool) float64 {
	base := def.DayBase
	add := def.DayAdd
	sub := def.DaySub

	if !isDay && def.NightBase != "" {
		base = def.NightBase
		add = def.NightAdd
		sub = def.NightSub
	}

	baseLon, ok := resolvePartLongitude(base, asc, planets, parts)
	if !ok {
		return 0
	}
	addLon, ok := resolvePartLongitude(add, asc, planets, parts)
	if !ok {
		return 0
	}
	subLon, ok := resolvePartLongitude(sub, asc, planets, parts)
	if !ok {
		return 0
	}

	return normalizeLon(baseLon + addLon - subLon)
}

// ComputePartReport computes Parts and their aspects to classical planets.
func ComputePartReport(name string, asc float64, planets map[string]float64, isDay bool, orb float64) *PartReport {
	parts := ComputeParts(asc, planets, isDay)

	var placements []PartPlacement
	for _, def := range PartCatalog() {
		lon, ok := parts[def.Name]
		if !ok {
			continue
		}
		sign := SignForLongitude(lon)
		signNum := int(lon/30) + 1
		if signNum > 12 {
			signNum = 12
		}
		placements = append(placements, PartPlacement{
			Part:    def.Name,
			Lon:     lon,
			Sign:    sign,
			SignNum: signNum,
		})
	}

	// Find Part-to-planet aspects (classical 7 only)
	classical := []string{"Sun", "Moon", "Mercury", "Venus", "Mars", "Jupiter", "Saturn"}
	aspects := DefaultAspects()
	var hits []PartAspectHit
	for _, def := range PartCatalog() {
		partLon, ok := parts[def.Name]
		if !ok {
			continue
		}
		for _, p := range classical {
			planetLon, ok := planets[p]
			if !ok {
				continue
			}
			dist := angleDist(partLon, planetLon)
			for _, a := range aspects {
				diff := math.Abs(dist - a.Angle)
				if diff <= orb {
					hits = append(hits, PartAspectHit{
						Part:   def.Name,
						Planet: p,
						Aspect: a.Name,
						Orb:    math.Round(diff*100) / 100,
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

	return &PartReport{
		Name:    name,
		IsDay:   isDay,
		Parts:   placements,
		Aspects: hits,
	}
}

// ComputePartCrossSystem computes Parts in both tropical and sidereal,
// comparing sign placement and verifying aspect invariance.
func ComputePartCrossSystem(name string, asc float64, planets map[string]float64, ayanamsa float64, isDay bool, orb float64) *PartCrossSystem {
	// Tropical computation
	tropReport := ComputePartReport(name, asc, planets, isDay, orb)

	// Sidereal computation
	sidAsc := normalizeLon(asc - ayanamsa)
	sidPlanets := make(map[string]float64)
	for k, v := range planets {
		sidPlanets[k] = normalizeLon(v - ayanamsa)
	}
	sidReport := ComputePartReport(name, sidAsc, sidPlanets, isDay, orb)

	// Count sign survivors
	signSurvivors := 0
	for i := range tropReport.Parts {
		if i < len(sidReport.Parts) && tropReport.Parts[i].Sign == sidReport.Parts[i].Sign {
			signSurvivors++
		}
	}

	// Aspects are zodiac-invariant (Part shifts by ayanamsa, planets shift by ayanamsa,
	// angular distances are preserved). Use tropical aspects as canonical.
	return &PartCrossSystem{
		Name:          name,
		Ayanamsa:      ayanamsa,
		IsDay:         isDay,
		Tropical:      tropReport.Parts,
		Sidereal:      sidReport.Parts,
		SignSurvivors: signSurvivors,
		Total:         len(tropReport.Parts),
		Aspects:       tropReport.Aspects,
	}
}
