package dignity

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/aj-nt/empirical/internal/swe"
)

// ── Transit Engine ───────────────────────────────────────────────────────

// ComputeFunc computes a planet's ecliptic longitude for a given date and
// planet ID. Returns (longitude, latitude, distance, speed). The caller
// handles ephemeris lookups — this is injectable for testing.
type ComputeFunc func(year, month, day int, hour float64, planetID int) (lon, lat, dist, speed float64)

// AspectDef holds an aspect angle and its name.
type AspectDef struct {
	Angle float64
	Name  string
}

// HardAspectsOnly returns conjunction, square, and opposition.
func HardAspectsOnly() []AspectDef {
	return []AspectDef{
		{0, "conjunction"},
		{90, "square"},
		{180, "opposition"},
	}
}

var (
	defaultAspectsOnce sync.Once
	defaultAspectsVal  []AspectDef
)

// DefaultAspects returns the five standard Ptolemaic aspects.
// The returned slice is cached; callers must not mutate it.
func DefaultAspects() []AspectDef {
	defaultAspectsOnce.Do(func() {
		defaultAspectsVal = []AspectDef{
			{0, "conjunction"},
			{60, "sextile"},
			{90, "square"},
			{120, "trine"},
			{180, "opposition"},
		}
	})
	return defaultAspectsVal
}

var (
	westernAspectsOnce sync.Once
	westernAspectsVal  []AspectDef
)

// WesternAspects returns the nine aspects used in modern Western astrology:
// the five Ptolemaic aspects plus semi-sextile (30°), semi-square (45°),
// sesquiquadrate (135°), and quincunx (150°).
// The returned slice is cached; callers must not mutate it.
func WesternAspects() []AspectDef {
	westernAspectsOnce.Do(func() {
		westernAspectsVal = []AspectDef{
			{0, "conjunction"},
			{30, "semi-sextile"},
			{45, "semi-square"},
			{60, "sextile"},
			{90, "square"},
			{120, "trine"},
			{135, "sesquiquadrate"},
			{150, "quincunx"},
			{180, "opposition"},
		}
	})
	return westernAspectsVal
}

// TransitHit records one transit-to-natal aspect at a specific date.
type TransitHit struct {
	Date          string  `json:"date"`
	TransitPlanet string  `json:"transit_planet"`
	NatalPlanet   string  `json:"natal_planet"`
	Aspect        string  `json:"aspect"`
	Orb           float64 `json:"orb"`
}

// planetSpec maps a planet name to its SWE ID for transit scanning.
type planetSpec struct {
	Name string
	ID   int
}

var (
	defaultTransitPlanetsOnce sync.Once
	defaultTransitPlanetsVal  []planetSpec
)

// DefaultTransitPlanets returns the standard planet set for transit scanning.
// The returned slice is cached; callers must not mutate it.
func DefaultTransitPlanets() []planetSpec {
	defaultTransitPlanetsOnce.Do(func() {
		defaultTransitPlanetsVal = []planetSpec{
			{"Sun", 0},
			{"Moon", 1},
			{"Mercury", 2},
			{"Venus", 3},
			{"Mars", 4},
			{"Jupiter", 5},
			{"Saturn", 6},
			{"Uranus", 7},
			{"Neptune", 8},
			{"Pluto", 9},
			{"Ceres", swe.CERES},
			{"Pallas", swe.PALLAS},
			{"Juno", swe.JUNO},
			{"Vesta", swe.VESTA},
			{"Lilith", swe.MEAN_APOG},
			{"Chiron", swe.CHIRON},
			{"Cupido", swe.CUPIDO},
			{"Hades", swe.HADES},
			{"Zeus", swe.ZEUS},
			{"Kronos", swe.KRONOS},
			{"Apollon", swe.APOLLON},
			{"Admetos", swe.ADMETOS},
			{"Poseidon", swe.POSEIDON},
			{"Vulkanus", swe.VULKANUS},
		}
	})
	return defaultTransitPlanetsVal
}

// ScanTransits computes transits over a date range.
//
// Parameters:
//   - natalLongs: natal planet name → tropical ecliptic longitude
//   - natalPlanets: which natal planets to check
//   - startDate, endDate: range as "YYYY-MM-DD"
//   - aspects: which aspects to detect
//   - orbDeg: max orb in degrees
//   - compute: function to get daily planet positions (injectable for testing)
//
// Returns all transit hits.
func ScanTransits(
	natalLongs map[string]float64,
	natalPlanets []string,
	startDate, endDate string,
	aspects []AspectDef,
	orbDeg float64,
	compute ComputeFunc,
) ([]TransitHit, error) {
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start date: %w", err)
	}
	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return nil, fmt.Errorf("invalid end date: %w", err)
	}

	transitPlanets := DefaultTransitPlanets()
	var hits []TransitHit

	current := start
	for !current.After(end) {
		y, m, d := current.Year(), int(current.Month()), current.Day()

		for _, tp := range transitPlanets {
			tLon, _, _, _ := compute(y, m, d, 12.0, tp.ID)
			for _, np := range natalPlanets {
				nLon, ok := natalLongs[np]
				if !ok {
					continue
				}
				dist := angleDist(tLon, nLon)
				for _, asp := range aspects {
					diff := math.Abs(dist - asp.Angle)
					if diff <= orbDeg {
						hits = append(hits, TransitHit{
							Date:          current.Format("2006-01-02"),
							TransitPlanet: tp.Name,
							NatalPlanet:   np,
							Aspect:        asp.Name,
							Orb:           math.Round(diff*100) / 100,
						})
					}
				}
			}
		}
		current = current.AddDate(0, 0, 1)
	}

	return hits, nil
}

// CompactTransitsWithRange collapses sequential days of the same transit
// into a date range and keeps the closest orb.
func CompactTransitsWithRange(hits []TransitHit) []struct {
	TransitPlanet string
	NatalPlanet   string
	Aspect        string
	MinOrb        float64
	DateStart     string
	DateEnd       string
} {
	type key struct {
		TransitPlanet string
		NatalPlanet   string
		Aspect        string
	}
	groups := make(map[key][]TransitHit)
	for _, h := range hits {
		k := key{h.TransitPlanet, h.NatalPlanet, h.Aspect}
		groups[k] = append(groups[k], h)
	}

	var result []struct {
		TransitPlanet string
		NatalPlanet   string
		Aspect        string
		MinOrb        float64
		DateStart     string
		DateEnd       string
	}
	for k, group := range groups {
		best := group[0]
		for _, h := range group {
			if h.Orb < best.Orb {
				best = h
			}
		}
		result = append(result, struct {
			TransitPlanet string
			NatalPlanet   string
			Aspect        string
			MinOrb        float64
			DateStart     string
			DateEnd       string
		}{
			TransitPlanet: k.TransitPlanet,
			NatalPlanet:   k.NatalPlanet,
			Aspect:        k.Aspect,
			MinOrb:        best.Orb,
			DateStart:     group[0].Date,
			DateEnd:       group[len(group)-1].Date,
		})
	}
	return result
}

// ScanTransitToTransit computes aspects between transiting planets over a date
// range. Returns compacted hits (date ranges with closest orb). This captures
// the actual sky weather — temporary geometric structures between moving
// planets, independent of any natal chart.
func ScanTransitToTransit(
	startDate, endDate string,
	aspects []AspectDef,
	orbDeg float64,
	compute ComputeFunc,
) ([]TransitHit, error) {
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start date: %w", err)
	}
	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return nil, fmt.Errorf("invalid end date: %w", err)
	}

	transitPlanets := DefaultTransitPlanets()
	var hits []TransitHit

	current := start
	for !current.After(end) {
		y, m, d := current.Year(), int(current.Month()), current.Day()

		// Compute all transit positions for this day
		positions := make(map[string]float64)
		for _, tp := range transitPlanets {
			lon, _, _, _ := compute(y, m, d, 12.0, tp.ID)
			positions[tp.Name] = lon
		}

		// Check all pairs
		names := make([]string, 0, len(positions))
		for n := range positions {
			names = append(names, n)
		}
		for i := 0; i < len(names); i++ {
			for j := i + 1; j < len(names); j++ {
				p1, p2 := names[i], names[j]
				dist := angleDist(positions[p1], positions[p2])
				for _, asp := range aspects {
					diff := math.Abs(dist - asp.Angle)
					if diff <= orbDeg {
						hits = append(hits, TransitHit{
							Date:          current.Format("2006-01-02"),
							TransitPlanet: p1,
							NatalPlanet:   p2,
							Aspect:        asp.Name,
							Orb:           math.Round(diff*100) / 100,
						})
					}
				}
			}
		}
		current = current.AddDate(0, 0, 1)
	}

	return hits, nil
}

// ── Event-Based Transit Detection ─────────────────────────────────────────

// ContactType identifies the kind of transit boundary contact.
type ContactType int

const (
	ContactIngress ContactType = iota
	ContactPeak
	ContactEgress
)

func (c ContactType) String() string {
	switch c {
	case ContactIngress:
		return "ingress"
	case ContactPeak:
		return "peak"
	case ContactEgress:
		return "egress"
	}
	return "unknown"
}

// Contact is one point in a transit period.
type Contact struct {
	Type       ContactType `json:"type"`
	JD         float64     `json:"jd"`
	Orb        float64     `json:"orb"`
	Retrograde bool        `json:"retrograde"`
}

// TransitPeriod represents one transit-to-natal aspect over its full duration.
// Contacts are ordered chronologically. For a simple direct transit:
//
//	[ingress, peak, egress]
//
// For a retrograde multi-pass:
//
//	[ingress(direct), peak(direct), egress(station), ingress(station), peak(retro), egress(final)]
type TransitPeriod struct {
	TransitPlanet string    `json:"transit_planet"`
	NatalPlanet   string    `json:"natal_planet"`
	Aspect        string    `json:"aspect"`
	Contacts      []Contact `json:"contacts"`
}

// ephemCache holds precomputed planet positions and speeds for fast scanning.
type ephemCache struct {
	startJD   float64
	stepDays  float64
	positions map[int][]float64
	speeds    map[int][]float64
}

func buildEphemCache(planetIDs []int, startJD, endJD, stepDays float64) *ephemCache {
	nSteps := int((endJD-startJD)/stepDays) + 1
	c := &ephemCache{
		startJD:   startJD,
		stepDays:  stepDays,
		positions: make(map[int][]float64),
		speeds:    make(map[int][]float64),
	}
	for _, pid := range planetIDs {
		c.positions[pid] = make([]float64, nSteps)
		c.speeds[pid] = make([]float64, nSteps)
		for i := 0; i < nSteps; i++ {
			jd := startJD + float64(i)*stepDays
			lon, _, _, speed := swe.CalcUT(jd, pid)
			c.positions[pid][i] = lon
			c.speeds[pid][i] = speed
		}
	}
	return c
}

func (c *ephemCache) getLon(planetID int, jd float64) float64 {
	idx := (jd - c.startJD) / c.stepDays
	i := int(idx)
	if i < 0 {
		i = 0
	}
	if i >= len(c.positions[planetID])-1 {
		i = len(c.positions[planetID]) - 2
	}
	frac := idx - float64(i)
	return c.positions[planetID][i] + frac*(c.positions[planetID][i+1]-c.positions[planetID][i])
}

func (c *ephemCache) getSpeed(planetID int, jd float64) float64 {
	idx := (jd - c.startJD) / c.stepDays
	i := int(idx)
	if i < 0 {
		i = 0
	}
	if i >= len(c.speeds[planetID])-1 {
		i = len(c.speeds[planetID]) - 2
	}
	frac := idx - float64(i)
	return c.speeds[planetID][i] + frac*(c.speeds[planetID][i+1]-c.speeds[planetID][i])
}

// FindTransitPeriods finds all transit-to-natal aspect periods over a date range
// using event-based detection with an ephemeris cache.
//
// Parameters:
//   - natalLongs: natal planet name → tropical ecliptic longitude
//   - transitPlanets: which transiting planets to scan (e.g., []string{"Saturn"})
//   - natalPlanets: which natal planets to check
//   - startDate, endDate: range as "YYYY-MM-DD"
//   - aspects: which aspects to detect
//   - orbDeg: max orb in degrees
//   - coarseStepDays: cache resolution (1.0 for slow planets, 0.25 for Moon)
//
// Returns all transit periods with ingress/peak/egress contacts.
func FindTransitPeriods(
	natalLongs map[string]float64,
	transitPlanets []string,
	natalPlanets []string,
	startDate, endDate string,
	aspects []AspectDef,
	orbDeg float64,
	coarseStepDays float64,
) ([]TransitPeriod, error) {
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start date: %w", err)
	}
	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return nil, fmt.Errorf("invalid end date: %w", err)
	}

	startJD := swe.Julday(start.Year(), int(start.Month()), start.Day(), 0, true)
	endJD := swe.Julday(end.Year(), int(end.Month()), end.Day(), 0, true)

	// Resolve transit planet names to IDs
	allTransitPlanets := DefaultTransitPlanets()
	nameToID := make(map[string]int, len(allTransitPlanets))
	for _, tp := range allTransitPlanets {
		nameToID[tp.Name] = tp.ID
	}

	// Collect planet IDs for the requested transit planets
	planetIDs := make([]int, 0, len(transitPlanets))
	for _, name := range transitPlanets {
		id, ok := nameToID[name]
		if !ok {
			return nil, fmt.Errorf("unknown transit planet: %s", name)
		}
		planetIDs = append(planetIDs, id)
	}

	// Build cache
	cache := buildEphemCache(planetIDs, startJD, endJD, coarseStepDays)

	// Scan all combinations
	var allPeriods []TransitPeriod
	for _, name := range transitPlanets {
		tID := nameToID[name]
		for _, np := range natalPlanets {
			nLon, ok := natalLongs[np]
			if !ok {
				continue
			}
			for _, asp := range aspects {
				periods := findAspectCrossings(
					cache, tID, name,
					nLon, np,
					asp.Angle, asp.Name,
					orbDeg,
					startJD, endJD,
					coarseStepDays,
				)
				allPeriods = append(allPeriods, periods...)
			}
		}
	}
	return allPeriods, nil
}

func findAspectCrossings(
	cache *ephemCache,
	transitID int,
	transitName string,
	natalLon float64,
	natalName string,
	aspectAngle float64,
	aspectName string,
	orbThreshold float64,
	startJD, endJD float64,
	coarseStepDays float64,
) []TransitPeriod {

	type crossing struct {
		jd   float64
		kind ContactType
	}

	var crossings []crossing
	prevInOrb := false

	for jd := startJD; jd <= endJD; jd += coarseStepDays {
		tLon := cache.getLon(transitID, jd)
		dist := angleDist(tLon, natalLon)
		orb := math.Abs(dist - aspectAngle)
		inOrb := orb <= orbThreshold

		if inOrb && !prevInOrb {
			crossings = append(crossings, crossing{jd, ContactIngress})
		} else if !inOrb && prevInOrb {
			crossings = append(crossings, crossing{jd, ContactEgress})
		}
		prevInOrb = inOrb
	}
	if prevInOrb {
		crossings = append(crossings, crossing{endJD, ContactEgress})
	}
	if len(crossings) == 0 {
		return nil
	}

	// Refine each crossing
	for i := range crossings {
		crossings[i].jd = refineCrossing(
			cache, transitID, natalLon, aspectAngle, orbThreshold,
			crossings[i].jd-coarseStepDays, crossings[i].jd+coarseStepDays,
			crossings[i].kind,
		)
	}

	// Pair ingress→egress and find peaks
	var periods []TransitPeriod
	for i := 0; i < len(crossings); {
		if crossings[i].kind != ContactIngress {
			i++
			continue
		}
		egressIdx := -1
		for j := i + 1; j < len(crossings); j++ {
			if crossings[j].kind == ContactEgress {
				egressIdx = j
				break
			}
		}
		if egressIdx == -1 {
			break
		}

		peakJD := findPeak(cache, transitID, natalLon, aspectAngle,
			crossings[i].jd, crossings[egressIdx].jd)

		ingressJD := crossings[i].jd
		egressJD := crossings[egressIdx].jd

		tLonI := cache.getLon(transitID, ingressJD)
		orbI := math.Abs(angleDist(tLonI, natalLon) - aspectAngle)
		isRetroI := cache.getSpeed(transitID, ingressJD) < 0

		tLonP := cache.getLon(transitID, peakJD)
		orbP := math.Abs(angleDist(tLonP, natalLon) - aspectAngle)
		isRetroP := cache.getSpeed(transitID, peakJD) < 0

		tLonE := cache.getLon(transitID, egressJD)
		orbE := math.Abs(angleDist(tLonE, natalLon) - aspectAngle)
		isRetroE := cache.getSpeed(transitID, egressJD) < 0

		period := TransitPeriod{
			TransitPlanet: transitName,
			NatalPlanet:   natalName,
			Aspect:        aspectName,
			Contacts: []Contact{
				{ContactIngress, ingressJD, orbI, isRetroI},
				{ContactPeak, peakJD, orbP, isRetroP},
				{ContactEgress, egressJD, orbE, isRetroE},
			},
		}
		periods = append(periods, period)
		i = egressIdx + 1
	}
	return periods
}

func refineCrossing(
	cache *ephemCache,
	transitID int,
	natalLon, aspectAngle, orbThreshold float64,
	loJD, hiJD float64,
	kind ContactType,
) float64 {
	for i := 0; i < 30; i++ {
		mid := (loJD + hiJD) / 2
		tLon := cache.getLon(transitID, mid)
		dist := angleDist(tLon, natalLon)
		orb := math.Abs(dist - aspectAngle)
		inOrb := orb <= orbThreshold

		if kind == ContactIngress {
			if inOrb {
				hiJD = mid
			} else {
				loJD = mid
			}
		} else {
			if inOrb {
				loJD = mid
			} else {
				hiJD = mid
			}
		}
	}
	return (loJD + hiJD) / 2
}

func findPeak(cache *ephemCache, transitID int, natalLon, aspectAngle, loJD, hiJD float64) float64 {
	phi := (1 + math.Sqrt(5)) / 2
	resphi := 2 - phi

	a, b := loJD, hiJD
	c := a + resphi*(b-a)
	d := b - resphi*(b-a)

	for i := 0; i < 30; i++ {
		orbC := math.Abs(angleDist(cache.getLon(transitID, c), natalLon) - aspectAngle)
		orbD := math.Abs(angleDist(cache.getLon(transitID, d), natalLon) - aspectAngle)

		if orbC < orbD {
			b = d
		} else {
			a = c
		}
		c = a + resphi*(b-a)
		d = b - resphi*(b-a)
	}
	return (a + b) / 2
}

// FindStations finds all retrograde station points for a planet over a date range.
func FindStations(planetName string, startDate, endDate string, stepDays float64) ([]float64, error) {
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start date: %w", err)
	}
	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return nil, fmt.Errorf("invalid end date: %w", err)
	}

	startJD := swe.Julday(start.Year(), int(start.Month()), start.Day(), 0, true)
	endJD := swe.Julday(end.Year(), int(end.Month()), end.Day(), 0, true)

	// Find planet ID
	var planetID int
	found := false
	for _, tp := range DefaultTransitPlanets() {
		if tp.Name == planetName {
			planetID = tp.ID
			found = true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("unknown planet: %s", planetName)
	}

	cache := buildEphemCache([]int{planetID}, startJD, endJD, stepDays)
	return findStations(cache, planetID, startJD, endJD, stepDays), nil
}

func findStations(cache *ephemCache, transitID int, startJD, endJD, stepDays float64) []float64 {
	var stations []float64
	prevSpeed := cache.getSpeed(transitID, startJD)

	for jd := startJD + stepDays; jd <= endJD; jd += stepDays {
		speed := cache.getSpeed(transitID, jd)
		if (prevSpeed > 0 && speed < 0) || (prevSpeed < 0 && speed > 0) {
			stationJD := refineStation(cache, transitID, jd-stepDays, jd)
			stations = append(stations, stationJD)
		}
		prevSpeed = speed
	}
	return stations
}

func refineStation(cache *ephemCache, transitID int, loJD, hiJD float64) float64 {
	for i := 0; i < 30; i++ {
		mid := (loJD + hiJD) / 2
		speed := cache.getSpeed(transitID, mid)
		if speed > 0 {
			loJD = mid
		} else {
			hiJD = mid
		}
	}
	return (loJD + hiJD) / 2
}
