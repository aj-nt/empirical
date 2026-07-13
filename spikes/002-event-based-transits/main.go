package main

import (
	"fmt"
	"math"
	"time"

	"github.com/aj-nt/empirical/internal/dignity"
	"github.com/aj-nt/empirical/internal/swe"
)

// ── New Data Model ───────────────────────────────────────────────────────

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
	Type      ContactType
	JD        float64
	Orb       float64 // degrees from exact
	Retrograde bool   // true if transit planet is retrograde at this contact
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
	TransitPlanet string
	NatalPlanet   string
	Aspect        string
	Contacts      []Contact
}

// ── Ephemeris Cache ──────────────────────────────────────────────────────

type ephemCache struct {
	startJD   float64
	stepDays  float64
	positions map[int][]float64
	speeds    map[int][]float64
}

func buildCache(planetIDs []int, startJD, endJD, stepDays float64) *ephemCache {
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

// ── Event-Based Detection ─────────────────────────────────────────────────

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

		// Build contacts with retrograde flag
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

// ── Station Detection ─────────────────────────────────────────────────────

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

// ── Utility ───────────────────────────────────────────────────────────────

func angleDist(a, b float64) float64 {
	d := math.Abs(a - b)
	if d > 180 {
		d = 360 - d
	}
	return d
}

func jdToDate(jd float64) string {
	y, m, d, h := swe.Revjul(jd)
	return fmt.Sprintf("%04d-%02d-%02d %02d:%02d",
		y, m, d, int(h), int((h-float64(int(h)))*60))
}

func retroLabel(isRetro bool) string {
	if isRetro {
		return " (R)"
	}
	return ""
}

// ── Main ──────────────────────────────────────────────────────────────────

func main() {
	swe.SetEphePath("/Users/aj/.local/share/ephe")

	// AJ's natal: 1969-02-16 07:10 UT
	natalJD := swe.Julday(1969, 2, 16, 7.1667, true)

	planetIDs := []struct {
		name string
		id   int
	}{
		{"Sun", swe.SUN}, {"Moon", swe.MOON}, {"Mercury", swe.MERCURY},
		{"Venus", swe.VENUS}, {"Mars", swe.MARS}, {"Jupiter", swe.JUPITER},
		{"Saturn", swe.SATURN}, {"Uranus", swe.URANUS}, {"Neptune", swe.NEPTUNE},
		{"Pluto", swe.PLUTO},
	}

	natalPlanets := map[string]float64{}
	for _, p := range planetIDs {
		lon, _, _, _ := swe.CalcUT(natalJD, p.id)
		natalPlanets[p.name] = lon
	}

	fmt.Println("=== Natal Positions ===")
	for _, p := range planetIDs {
		fmt.Printf("  %-8s: %7.2f°\n", p.name, natalPlanets[p.name])
	}

	planetIDInts := make([]int, len(planetIDs))
	for i, p := range planetIDs {
		planetIDInts[i] = p.id
	}

	// Build 2-year cache
	startJD := swe.Julday(2026, 1, 1, 0, true)
	endJD := swe.Julday(2028, 1, 1, 0, true)

	fmt.Println("\n=== Building Ephemeris Cache (2 years, 1-day step) ===")
	cacheStart := time.Now()
	cache := buildCache(planetIDInts, startJD, endJD, 1.0)
	fmt.Printf("  Cache built in %v (%d SWE calls)\n",
		time.Since(cacheStart), len(planetIDs)*int((endJD-startJD)/1.0+1))

	// Saturn stations
	fmt.Println("\n=== Saturn Stations (2026-2027) ===")
	stations := findStations(cache, swe.SATURN, startJD, endJD, 1.0)
	for _, s := range stations {
		speed := cache.getSpeed(swe.SATURN, s)
		dir := "direct→retro"
		if speed > 0 {
			dir = "retro→direct"
		}
		fmt.Printf("  %s  %s\n", jdToDate(s), dir)
	}

	// Saturn square Moon — verify zero
	fmt.Println("\n=== Saturn □ Moon (2026-2027, 3° orb) ===")
	satMoonPeriods := findAspectCrossings(
		cache, swe.SATURN, "Saturn",
		natalPlanets["Moon"], "Moon",
		90, "square", 3.0,
		startJD, endJD, 1.0,
	)
	fmt.Printf("  Periods: %d (correctly zero — Saturn never squares Moon here)\n",
		len(satMoonPeriods))

	// Saturn conjunction natal Saturn (retrograde multi-pass)
	fmt.Println("\n=== Saturn ☌ Saturn (2026-2027, 3° orb) ===")
	satSatPeriods := findAspectCrossings(
		cache, swe.SATURN, "Saturn",
		natalPlanets["Saturn"], "Saturn",
		0, "conjunction", 3.0,
		startJD, endJD, 1.0,
	)
	for i, p := range satSatPeriods {
		fmt.Printf("  Period %d:\n", i+1)
		for _, c := range p.Contacts {
			fmt.Printf("    %-8s: %s%s  (orb: %.4f°)\n",
				c.Type, jdToDate(c.JD), retroLabel(c.Retrograde), c.Orb)
		}
		dur := p.Contacts[2].JD - p.Contacts[0].JD
		fmt.Printf("    Duration: %.1f days\n", dur)
	}

	// Mercury square Moon (fast planet, retrograde multi-pass)
	fmt.Println("\n=== Mercury □ Moon (2026, 3° orb) ===")
	mercStart := swe.Julday(2026, 1, 1, 0, true)
	mercEnd := swe.Julday(2027, 1, 1, 0, true)
	mercPeriods := findAspectCrossings(
		cache, swe.MERCURY, "Mercury",
		natalPlanets["Moon"], "Moon",
		90, "square", 3.0,
		mercStart, mercEnd, 1.0,
	)
	fmt.Printf("  Found %d periods\n", len(mercPeriods))
	for i, p := range mercPeriods {
		if i >= 5 {
			fmt.Printf("  ... and %d more\n", len(mercPeriods)-5)
			break
		}
		fmt.Printf("  Period %d: %s → %s (peak: %s%s, orb: %.4f°)\n",
			i+1,
			jdToDate(p.Contacts[0].JD),
			jdToDate(p.Contacts[2].JD),
			jdToDate(p.Contacts[1].JD),
			retroLabel(p.Contacts[1].Retrograde),
			p.Contacts[1].Orb,
		)
	}

	// ── Performance ──
	allAspects := dignity.DefaultAspects()

	// 1-year
	fmt.Println("\n=== Performance: 1-Year Full Scan (cached) ===")
	start1Y := swe.Julday(2026, 1, 1, 0, true)
	end1Y := swe.Julday(2027, 1, 1, 0, true)
	cache1Y := buildCache(planetIDInts, start1Y, end1Y, 1.0)

	start := time.Now()
	totalPeriods := 0
	for _, tp := range planetIDs {
		for _, np := range planetIDs {
			if tp.name == np.name {
				continue
			}
			for _, asp := range allAspects {
				periods := findAspectCrossings(
					cache1Y, tp.id, tp.name,
					natalPlanets[np.name], np.name,
					asp.Angle, asp.Name,
					3.0,
					start1Y, end1Y,
					1.0,
				)
				totalPeriods += len(periods)
			}
		}
	}
	elapsed := time.Since(start)
	fmt.Printf("  Total periods: %d\n", totalPeriods)
	fmt.Printf("  Time: %v\n", elapsed)

	// 2-year
	fmt.Println("\n=== Performance: 2-Year Full Scan (cached) ===")
	start2Y := swe.Julday(2026, 1, 1, 0, true)
	end2Y := swe.Julday(2028, 1, 1, 0, true)
	cache2Y := buildCache(planetIDInts, start2Y, end2Y, 1.0)

	start = time.Now()
	totalPeriods = 0
	for _, tp := range planetIDs {
		for _, np := range planetIDs {
			if tp.name == np.name {
				continue
			}
			for _, asp := range allAspects {
				periods := findAspectCrossings(
					cache2Y, tp.id, tp.name,
					natalPlanets[np.name], np.name,
					asp.Angle, asp.Name,
					3.0,
					start2Y, end2Y,
					1.0,
				)
				totalPeriods += len(periods)
			}
		}
	}
	elapsed = time.Since(start)
	fmt.Printf("  Total periods: %d\n", totalPeriods)
	fmt.Printf("  Time: %v\n", elapsed)
}
