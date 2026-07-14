package main

import (
	"fmt"
	"math"

	"github.com/aj-nt/empirical/internal/dignity"
	"github.com/aj-nt/empirical/internal/swe"
)

func computeDraconic(name string, bc *dignity.BaseChart, orbDeg float64) (*DraconicResponse, error) {
	// Build tropical planet map from all computed positions
	tropical := dignity.TropicalToLonMap(bc.Tropical)

	// Compute draconic chart
	drac := dignity.ComputeDraconic(tropical, bc.NorthNode)

	// Compute sign shifts
	shifts := dignity.ComputeDraconicSignShifts(tropical, bc.NorthNode)

	// Compute bridges (all planets except TNPs)
	allPlanets := dignity.NonTNPNoNodePlanetNames
	bridges := dignity.ComputeDraconicBridges(tropical, bc.NorthNode, allPlanets, dignity.DefaultAspects(), orbDeg)

	// Build shift list
	var shiftList []DraconicShiftJSON
	for _, s := range shifts {
		shiftList = append(shiftList, DraconicShiftJSON{s.Planet, s.TropSign, s.DracSign})
	}

	return &DraconicResponse{
		Name:    name,
		Offset:  drac.Offset,
		Planets: drac.Planets,
		Shifts:  shiftList,
		Bridges: bridges,
	}, nil
}

func computeDraconicSynastry(name1 string, bc1 *dignity.BaseChart, name2 string, bc2 *dignity.BaseChart, orbDeg float64) (*DraconicSynastryResponse, error) {
	tropical1 := dignity.TropicalToLonMap(bc1.Tropical)
	tropical2 := dignity.TropicalToLonMap(bc2.Tropical)

	allPlanets := dignity.NonTNPNoNodePlanetNames
	hits := dignity.ComputeDraconicSynastry(tropical1, bc1.NorthNode, tropical2, bc2.NorthNode, allPlanets, dignity.DefaultAspects(), orbDeg)

	return &DraconicSynastryResponse{
		Name1: name1,
		Name2: name2,
		Hits:  hits,
	}, nil
}

func computeDraconicSynastryFull(name1 string, bc1 *dignity.BaseChart, name2 string, bc2 *dignity.BaseChart, orbDeg float64) (*DraconicSynastryFullResponse, error) {
	tropical1 := dignity.TropicalToLonMap(bc1.Tropical)
	tropical2 := dignity.TropicalToLonMap(bc2.Tropical)

	allPlanets := dignity.NonTNPNoNodePlanetNames
	result := dignity.ComputeDraconicSynastryFull(tropical1, bc1.NorthNode, tropical2, bc2.NorthNode, allPlanets, dignity.DefaultAspects(), orbDeg)

	return &DraconicSynastryFullResponse{
		Name1:        name1,
		Name2:        name2,
		DracToDrac:   result.DracToDrac,
		TropAToDracB: result.TropAToDracB,
		TropBToDracA: result.TropBToDracA,
	}, nil
}

func computeDraconicTransits(name string, bc *dignity.BaseChart, startDate, endDate string, orbDeg float64, cacheDir string) (*DraconicTransitsResponse, error) {
	// Build tropical planet map
	tropical := dignity.TropicalToLonMap(bc.Tropical)

	// Compute draconic chart (soul-level natal positions)
	drac := dignity.ComputeDraconic(tropical, bc.NorthNode)

	// Build compute function for transiting positions
	tzOff := 0.0 // not stored in BaseChart; use 0 (positions are UT-based)
	compute := func(year, month, day int, hour float64, planetID int) (float64, float64, float64, float64) {
		utHour := hour - tzOff
		jd := swe.Julday(year, month, day, utHour, true)
		lon, lat, dist, speed := swe.CalcUT(jd, planetID)
		return lon, lat, dist, speed
	}

	// Scan transits against draconic positions
	transitPlanets := dignity.AllPlanetNames
	hits, err := dignity.ScanTransits(drac.Planets, transitPlanets, startDate, endDate, dignity.DefaultAspects(), orbDeg, compute)
	if err != nil {
		return nil, err
	}

	compact := dignity.CompactTransitsWithRange(hits)

	response := &DraconicTransitsResponse{
		Name:   name,
		Offset: drac.Offset,
	}
	for _, c := range compact {
		response.Transits = append(response.Transits, TransitHitJSON{
			TransitPlanet: c.TransitPlanet,
			NatalPlanet:   c.NatalPlanet,
			Aspect:        c.Aspect,
			Orb:           c.MinOrb,
			StartDate:     c.DateStart,
			EndDate:       c.DateEnd,
		})
	}

	return response, nil
}

func computeProgressedDraconic(name string, bc *dignity.BaseChart, targetDate string, cacheDir string) (*ProgressedDraconicResponse, error) {
	// Parse target date
	var y, m, d int
	fmt.Sscanf(targetDate, "%d-%d-%d", &y, &m, &d)

	// Compute current transiting NN
	utHour := 12.0 // noon UT
	jd := swe.Julday(y, m, d, utHour, true)
	lon, _, _, _ := swe.CalcUT(jd, swe.MEAN_NODE)
	currentNN := lon

	// Build tropical planet map
	tropical := dignity.TropicalToLonMap(bc.Tropical)

	// Compute both draconic charts
	natalDrac := dignity.ComputeDraconic(tropical, bc.NorthNode)
	progDrac := dignity.ComputeProgressedDraconic(tropical, currentNN)

	// Compute sign shifts between classic and progressed draconic
	shifts := dignity.ComputeDraconicSignShifts(tropical, currentNN)

	// Format datetime
	yr, mo, dy, hr := swe.Revjul(jd)
	dtStr := fmt.Sprintf("%d-%02d-%02d %02d:%02d UT", yr, mo, dy, int(hr), int((hr-float64(int(hr)))*60))

	var shiftList []DraconicShiftJSON
	for _, s := range shifts {
		shiftList = append(shiftList, DraconicShiftJSON{s.Planet, s.TropSign, s.DracSign})
	}

	return &ProgressedDraconicResponse{
		Name:          name,
		Date:          dtStr,
		NatalNN:       bc.NorthNode,
		CurrentNN:     currentNN,
		NNShift:       currentNN - bc.NorthNode,
		NatalDraconic: natalDrac.Planets,
		ProgDraconic:  progDrac.Planets,
		SignShifts:    shiftList,
	}, nil
}

func computeDraconicSolarReturn(name string, bc *dignity.BaseChart, targetYear int, cacheDir string) (*DraconicSolarReturnResponse, error) {
	// Get natal Sun longitude
	natalSun := dignity.TropicalToLonMap(bc.Tropical)["Sun"]

	// Find exact solar return moment
	jdSR := findSolarReturnJD(natalSun, targetYear, bc.JD)

	// Calculate positions at solar return
	planetIDs := dignity.BasicPlanets

	tropical := make(map[string]float64)
	for _, p := range planetIDs {
		lon, _, _, _ := swe.CalcUT(jdSR, p.ID)
		tropical[p.Name] = dignity.NormalizeLon(lon)
	}

	// Calculate ASC and MC at solar return
	_, ascmc := swe.Houses(jdSR, 0, 0, 'P') // lat/lng not stored in BaseChart; use 0,0
	tropical["Ascendant"] = ascmc[0]
	tropical["Midheaven"] = ascmc[1]

	// Draconic using solar return's own NN
	srNode := tropical["Node"]
	draconic := make(map[string]float64)
	for name, lon := range tropical {
		if name == "Node" {
			draconic[name] = 0.0
		} else {
			draconic[name] = dignity.NormalizeLon(lon - srNode)
		}
	}

	// Draconic using natal NN (soul chart relative to natal soul frame)
	draconicByNatal := make(map[string]float64)
	for name, lon := range tropical {
		draconicByNatal[name] = dignity.NormalizeLon(lon - bc.NorthNode)
	}

	// Format datetime
	yr, mo, dy, hr := swe.Revjul(jdSR)
	dtStr := fmt.Sprintf("%d-%02d-%02d %02d:%02d UT", yr, mo, dy, int(hr), int((hr-float64(int(hr)))*60))

	return &DraconicSolarReturnResponse{
		Name:              name,
		TargetYear:        targetYear,
		JD:                jdSR,
		DateTime:          dtStr,
		Tropical:          tropical,
		Draconic:          draconic,
		DraconicByNatalNN: draconicByNatal,
	}, nil
}

func computeDraconicTransitsCross(name string, bc *dignity.BaseChart, startDate, endDate string, orbDeg float64, cacheDir string) (*DraconicTransitsCrossResponse, error) {
	// Build tropical planet map
	tropical := dignity.TropicalToLonMap(bc.Tropical)

	// Compute draconic chart (zodiac-invariant)
	drac := dignity.ComputeDraconic(tropical, bc.NorthNode)

	// Parse date range
	var sy, sm, sd, ey, em, ed int
	fmt.Sscanf(startDate, "%d-%d-%d", &sy, &sm, &sd)
	fmt.Sscanf(endDate, "%d-%d-%d", &ey, &em, &ed)

	// Compute transiting positions at midpoint of range for snapshot comparison
	midJD := swe.Julday(sy+(ey-sy)/2, sm+(em-sm)/2, sd+(ed-sd)/2, 12.0, true)
	ayan := swe.GetAyanamsaUT(midJD)

	// Compute tropical transiting positions
	tropTransits := make(map[string]float64)
	sidTransits := make(map[string]float64)
	planetIDs := dignity.AllPlanets
	for _, p := range planetIDs {
		lon, _, _, _ := swe.CalcUT(midJD, p.ID)
		tropLon := dignity.NormalizeLon(lon)
		tropTransits[p.Name] = tropLon
		sidTransits[p.Name] = dignity.NormalizeLon(tropLon - ayan)
	}

	aspects := dignity.DefaultAspects()
	result := dignity.CompareCrossSystemTransits(drac.Planets, tropTransits, sidTransits, aspects, orbDeg)

	// Build JSON response
	response := &DraconicTransitsCrossResponse{
		Name:         name,
		Offset:       drac.Offset,
		Ayanamsa:     ayan,
		Orb:          orbDeg,
		Survivors:    make([]CrossHitJSON, 0),
		TropicalOnly: make([]CrossHitJSON, 0),
		SiderealOnly: make([]CrossHitJSON, 0),
	}
	yr, mo, dy, hr := swe.Revjul(midJD)
	response.MidDate = fmt.Sprintf("%d-%02d-%02d %02d:%02d UT", yr, mo, dy, int(hr), int((hr-float64(int(hr)))*60))

	for _, h := range result.Survivors {
		response.Survivors = append(response.Survivors, CrossHitJSON{h.TransitPlanet, h.NatalPlanet, h.Aspect, h.Orb})
	}
	for _, h := range result.TropicalOnly {
		response.TropicalOnly = append(response.TropicalOnly, CrossHitJSON{h.TransitPlanet, h.NatalPlanet, h.Aspect, h.Orb})
	}
	for _, h := range result.SiderealOnly {
		response.SiderealOnly = append(response.SiderealOnly, CrossHitJSON{h.TransitPlanet, h.NatalPlanet, h.Aspect, h.Orb})
	}

	return response, nil
}

func computeProgressedCross(name string, bc *dignity.BaseChart, targetDate string, orbDeg float64, cacheDir string) (*ProgressedCrossResponse, error) {
	// Parse target date
	var y, m, d int
	fmt.Sscanf(targetDate, "%d-%d-%d", &y, &m, &d)
	utHour := 12.0
	targetJD := swe.Julday(y, m, d, utHour, true)

	// Age in years
	age := (targetJD - bc.JD) / 365.2425

	// Progressed JD: birthJD + age (day-for-a-year)
	progJD := bc.JD + age

	// Natal positions (tropical)
	natal := dignity.TropicalToLonMap(bc.Tropical)

	// Progressed positions (tropical)
	planetIDs := dignity.AllPlanets
	prog := make(map[string]float64)
	for _, p := range planetIDs {
		lon, _, _, _ := swe.CalcUT(progJD, p.ID)
		for lon < 0 {
			lon += 360
		}
		for lon >= 360 {
			lon -= 360
		}
		prog[p.Name] = lon
	}

	// Ayanamsa at birth
	ayan := swe.GetAyanamsaUT(bc.JD)

	aspects := dignity.DefaultAspects()
	result := dignity.CompareCrossSystemProgressed(natal, prog, ayan, aspects, orbDeg)

	// Build JSON response
	response := &ProgressedCrossResponse{
		Name:         name,
		TargetDate:   targetDate,
		Age:          math.Round(age*100) / 100,
		Ayanamsa:     ayan,
		Orb:          orbDeg,
		Survivors:    make([]ProgressedCrossHitJSON, 0),
		TropicalOnly: make([]ProgressedCrossHitJSON, 0),
		SiderealOnly: make([]ProgressedCrossHitJSON, 0),
	}

	for _, h := range result.Survivors {
		response.Survivors = append(response.Survivors, ProgressedCrossHitJSON{h.ProgressedPlanet, h.NatalPlanet, h.Aspect, h.Orb})
	}
	for _, h := range result.TropicalOnly {
		response.TropicalOnly = append(response.TropicalOnly, ProgressedCrossHitJSON{h.ProgressedPlanet, h.NatalPlanet, h.Aspect, h.Orb})
	}
	for _, h := range result.SiderealOnly {
		response.SiderealOnly = append(response.SiderealOnly, ProgressedCrossHitJSON{h.ProgressedPlanet, h.NatalPlanet, h.Aspect, h.Orb})
	}

	return response, nil
}

func findSolarReturnJD(natalSun float64, targetYear int, natalJD float64) float64 {
	_, birthMo, birthDay, _ := swe.Revjul(natalJD)
	jdMid := swe.Julday(targetYear, birthMo, birthDay, 12.0, true)

	lo := jdMid - 1.0
	hi := jdMid + 1.0

	var jdSR float64
	for i := 0; i < 40; i++ {
		mid := (lo + hi) / 2.0
		sunLon, _, _, _ := swe.CalcUT(mid, swe.SUN)
		sunLon = dignity.NormalizeLon(sunLon)

		diff := sunLon - natalSun
		if diff > 180 {
			diff -= 360
		} else if diff < -180 {
			diff += 360
		}

		if math.Abs(diff) < 0.00001 {
			jdSR = mid
			break
		}

		if diff < 0 {
			lo = mid
		} else {
			hi = mid
		}
	}
	if jdSR == 0 {
		jdSR = (lo + hi) / 2.0
	}
	return jdSR
}
