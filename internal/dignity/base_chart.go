package dignity

import (
	"github.com/aj-nt/empirical/internal/swe"
)

// ── BaseChart: system-agnostic computed chart ──────────────────────────────
//
// BaseChart holds all computed astrological positions for a single chart.
// Computed once from birth data; every system extracts what it needs via
// FromBase() pure functions. No system re-derives positions.

// Position holds a planet's longitude, latitude, and daily speed.
type Position struct {
	Lon   float64
	Lat   float64
	Speed float64
}

// BaseChart holds all computed astrological positions for a single chart.
type BaseChart struct {
	// Identity
	Name string

	// Birth data
	Year, Month, Day, Hour, Minute, Second int
	TZOffset                               float64
	Lat, Lng                               float64

	// Core positions
	Tropical map[string]Position // planet name → lon+lat+speed
	Sidereal map[string]Position // same, shifted by ayanamsa
	Ayanamsa float64

	// Angles
	ASC, MC, DSC, IC float64

	// Nodes
	NorthNode float64
	SouthNode float64

	// Houses (all systems)
	Houses map[string][]float64 // "placidus" → [13]cusps (1-indexed), etc.

	// Julian Day
	JD    float64
	DayJD int // Julian Day at midnight UTC (for Ba Zi day pillar)

	// Pre-computed derivatives
	Aspects     []AspectHit       // all natal aspects
	FixedStars  []StarConjunction // all star-planet conjunctions
	ArabicParts map[string]float64 // all Arabic Parts
	GMST        float64
}

// ComputeBaseChart computes all astrological positions for a birth chart.
// This is the single entry point — every system extracts from the returned
// BaseChart. Caller must have already called swe.SetEphePath and swe.SetSidMode.
func ComputeBaseChart(bd BirthData) (*BaseChart, error) {
	validateStarCatalog() // one-time consistency check on StarNames/StarMeanings

	utHour := float64(bd.Hour) + float64(bd.Minute)/60.0 + float64(bd.Second)/3600.0 - bd.TZOffset
	jd := swe.Julday(bd.Year, bd.Month, bd.Day, utHour, true)
	ayan := swe.GetAyanamsaUT(jd)

	// ── Planet positions (tropical + sidereal) ──────────────────────────
	tropical := make(map[string]Position)
	sidereal := make(map[string]Position)
	for _, p := range AllPlanets {
		lon, lat, _, speed := swe.CalcUT(jd, p.ID)
		tropical[p.Name] = Position{Lon: lon, Lat: lat, Speed: speed}
		sidLon := lon - ayan
		if sidLon < 0 {
			sidLon += 360
		}
		sidereal[p.Name] = Position{Lon: sidLon, Lat: lat, Speed: speed}
	}

	// ── Nodes ───────────────────────────────────────────────────────────
	nnLon, _, _, _ := swe.CalcUT(jd, swe.MEAN_NODE)
	snLon := nnLon + 180
	if snLon >= 360 {
		snLon -= 360
	}

	// ── Houses (all systems) ────────────────────────────────────────────
	houses := make(map[string][]float64)
	var asc, mc float64
	for _, hs := range []struct {
		name string
		code byte
	}{
		{"placidus", 'P'},
		{"whole_sign", 'W'},
		{"equal", 'E'},
		{"porphyry", 'O'},
		{"koch", 'K'},
	} {
		cusps, ascmc := swe.Houses(jd, bd.Lat, bd.Lng, hs.code)
		houseCusps := make([]float64, 13) // 1-indexed
		copy(houseCusps[1:], cusps[1:13])
		houses[hs.name] = houseCusps
		if hs.name == "placidus" {
			asc = ascmc[0]
			mc = ascmc[1]
		}
	}

	// ── Angles ──────────────────────────────────────────────────────────
	dsc := asc + 180
	if dsc >= 360 {
		dsc -= 360
	}
	ic := mc + 180
	if ic >= 360 {
		ic -= 360
	}

	// ── Fixed stars ─────────────────────────────────────────────────────
	starPositions := make(map[string]float64)
	for _, starName := range StarNames {
		lon, _, _, _ := swe.Fixstar(starName, jd)
		if lon != 0 {
			starPositions[starName] = normalizeLon(lon)
		}
	}
	planetLons := TropicalToLonMap(tropical)
	stars := FindStarConjunctions(starPositions, planetLons, 2.0)

	// ── Arabic Parts ────────────────────────────────────────────────────
	// Determine day/night: Sun above horizon = day
	sunLon := tropical["Sun"].Lon
	diff := sunLon - asc
	if diff < 0 {
		diff += 360
	}
	isDay := diff < 180
	parts := ComputeParts(asc, planetLons, isDay)

	// ── Natal aspects ───────────────────────────────────────────────────
	aspects := FindNatalAspects(planetLons, DefaultAspects(), 5.0)

	// ── GMST ────────────────────────────────────────────────────────────
	gmst := ComputeGMST(jd)

	return &BaseChart{
		Name:        bd.Name,
		Year:        bd.Year,
		Month:       bd.Month,
		Day:         bd.Day,
		Hour:        bd.Hour,
		Minute:      bd.Minute,
		Second:      bd.Second,
		TZOffset:    bd.TZOffset,
		Lat:         bd.Lat,
		Lng:         bd.Lng,
		Tropical:    tropical,
		Sidereal:    sidereal,
		Ayanamsa:    ayan,
		ASC:         asc,
		MC:          mc,
		DSC:         dsc,
		IC:          ic,
		NorthNode:   nnLon,
		SouthNode:   snLon,
		Houses:      houses,
		JD:          jd,
		DayJD:       int(swe.Julday(bd.Year, bd.Month, bd.Day, 0, true)),
		Aspects:     aspects,
		FixedStars:  stars,
		ArabicParts: parts,
		GMST:        gmst,
	}, nil
}

// TropicalToLonMap extracts a longitude-only map from tropical positions.
func TropicalToLonMap(tropical map[string]Position) map[string]float64 {
	m := make(map[string]float64, len(tropical))
	for k, v := range tropical {
		m[k] = v.Lon
	}
	return m
}
