package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/aj-nt/empirical"
	"github.com/aj-nt/empirical/internal/declination"
	"github.com/aj-nt/empirical/internal/dignity"
	"github.com/aj-nt/empirical/internal/divisional"
	"github.com/aj-nt/empirical/internal/firdaria"
	"github.com/aj-nt/empirical/internal/harmonic"
	"github.com/aj-nt/empirical/internal/parans"
	"github.com/aj-nt/empirical/internal/server"
	"github.com/aj-nt/empirical/internal/swe"
	"github.com/aj-nt/empirical/internal/uranian"
)

// chartData removed — use dignity.BaseChart instead.


var (
	epheOnce sync.Once
	epheDir  string
)

// initEphe ensures ephemeris is initialized exactly once.
func initEphe() {
	epheOnce.Do(func() {
		var err error
		epheDir, err = empirical.EnsureEpheCache()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to initialize ephemeris: %v\n", err)
			os.Exit(1)
		}
		swe.SetEphePath(epheDir)
		swe.SetSidMode(swe.SIDM_LAHIRI, 0, 0)
	})
}

// computePositions calculates planet longitudes, ayanamsa, ASC, and NN.
// computePositions is a thin wrapper around dignity.ComputeBaseChart.
// It handles ephemeris initialization and returns a *dignity.BaseChart.
func computePositions(year, month, day, hour, minute int, tzOff, lat, lng float64, cacheDir string) *dignity.BaseChart {
	initEphe()

	bc, err := dignity.ComputeBaseChart(dignity.BirthData{
		Name:     "",
		Year:     year,
		Month:    month,
		Day:      day,
		Hour:     hour,
		Minute:   minute,
		Second:   0,
		TZOffset: tzOff,
		Lat:      lat,
		Lng:      lng,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to compute base chart: %v\n", err)
		os.Exit(1)
	}
	return bc
}

// computeAll returns a full multi-phase report.
func computeAll(name string, year, month, day, hour, minute, second int, tzOff, lat, lng float64, cacheDir string) *dignity.FullReport {
	bc := computePositions(year, month, day, hour, minute, tzOff, lat, lng, cacheDir)
	return dignity.ComputeFullReport(dignity.TropicalToLonMap(bc.Tropical), bc.Ayanamsa, bc.NorthNode, bc.ASC, name, year, month, day, hour, minute, second, tzOff, lat, lng)
}

// computeTransits runs the transit engine and returns compact JSON results.
// When sidereal is true, uses sidereal (Lahiri) positions for both natal and transiting planets.
func computeTransits(name string, year, month, day, hour, minute int, tzOff, lat, lng float64, startDate, endDate string, orbDeg float64, sidereal bool, cacheDir string) (*TransitsResponse, error) {
	bc := computePositions(year, month, day, hour, minute, tzOff, lat, lng, cacheDir)

	// Build planet positions — all bodies already in dignity.TropicalToLonMap(bc.Tropical)
	natalLongs := dignity.TropicalToLonMap(bc.Tropical)

	natalPlanets := dignity.AllPlanetNames

	// Build compute function — sidereal subtracts ayanamsa from tropical positions
	compute := func(year, month, day int, hour float64, planetID int) (float64, float64, float64, float64) {
		utHour := hour - tzOff
		jd := swe.Julday(year, month, day, utHour, true)
		lon, lat, dist, speed := swe.CalcUT(jd, planetID)
		if sidereal {
			ayan := swe.GetAyanamsaUT(jd)
			lon -= ayan
			if lon < 0 {
				lon += 360
			}
		}
		return lon, lat, dist, speed
	}

	// Build natal positions — sidereal subtracts ayanamsa
	scanLongs := make(map[string]float64)
	for k, v := range natalLongs {
		if sidereal {
			v -= bc.Ayanamsa
			if v < 0 {
				v += 360
			}
		}
		scanLongs[k] = v
	}

	hits, err := dignity.ScanTransits(scanLongs, natalPlanets, startDate, endDate, dignity.HardAspectsOnly(), orbDeg, compute)
	if err != nil {
		return nil, err
	}

	compact := dignity.CompactTransitsWithRange(hits)

	// Transit-to-transit aspects (sky weather)
	allAspects := dignity.DefaultAspects()
	ttHits, err := dignity.ScanTransitToTransit(startDate, endDate, allAspects, orbDeg, compute)
	if err != nil {
		return nil, err
	}
	ttCompact := dignity.CompactTransitsWithRange(ttHits)

	// Transit midpoints (transiting planet conjunct natal midpoint)
	mpHits, err := dignity.FindTransitMidpoints(scanLongs, dignity.DefaultTransitPlanets(), startDate, endDate, orbDeg, compute)
	if err != nil {
		return nil, err
	}

	// Build JSON response
	response := &TransitsResponse{
		Name:     name,
		Sidereal: sidereal,
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
	for _, c := range ttCompact {
		response.SkyWeather = append(response.SkyWeather, TransitHitJSON{
			TransitPlanet: c.TransitPlanet,
			NatalPlanet:   c.NatalPlanet,
			Aspect:        c.Aspect,
			Orb:           c.MinOrb,
			StartDate:     c.DateStart,
			EndDate:       c.DateEnd,
		})
	}
	for _, m := range mpHits {
		response.Midpoints = append(response.Midpoints, TransitMidpointJSON{
			Date:          m.Date,
			TransitPlanet: m.TransitPlanet,
			NatalPairA:    m.NatalPairA,
			NatalPairB:    m.NatalPairB,
			Orb:           m.Orb,
		})
	}

	return response, nil
}

// computeSynastry computes inter-aspects between two natal charts.
func computeSynastry(name1 string, y1, mo1, d1, h1, mi1 int, tz1, la1, lo1 float64, name2 string, y2, mo2, d2, h2, mi2 int, tz2, la2, lo2 float64, orbDeg float64, cacheDir string) (*SynastryResponse, error) {
	bc1 := computePositions(y1, mo1, d1, h1, mi1, tz1, la1, lo1, cacheDir)
	bc2 := computePositions(y2, mo2, d2, h2, mi2, tz2, la2, lo2, cacheDir)

	// Build planet maps — all bodies already in dignity.TropicalToLonMap(bc.Tropical)
	chart1 := dignity.TropicalToLonMap(bc1.Tropical)
	chart2 := dignity.TropicalToLonMap(bc2.Tropical)

	planets := dignity.AllPlanetNames
	aspects := dignity.DefaultAspects()

	hits := dignity.ComputeSynastry(chart1, chart2, planets, aspects, orbDeg)

	return &SynastryResponse{
		Name1:   name1,
		Name2:   name2,
		Aspects: hits,
	}, nil
}

// computeRelocation compares two locations for a single person using cross-validated
// house convergence and timing convergence. Returns which house shifts are reliable
// (unanimous at both locations) vs system-dependent.
func computeRelocation(name string, year, month, day, hour, minute int, tzOff, lat, lng float64, locA server.LatLng, locB server.LatLng, targetDate string, cacheDir string) (*RelocationResponse, error) {
	// Compute positions at both locations (natal planets are location-invariant, houses differ)
	bcA := computePositions(year, month, day, hour, minute, tzOff, locA.Lat, locA.Lng, cacheDir)
	bcB := computePositions(year, month, day, hour, minute, tzOff, locB.Lat, locB.Lng, cacheDir)

	// Phase 3: House convergence at both locations
	hcA := dignity.ComputeHouseConvergence(dignity.TropicalToLonMap(bcA.Tropical), year, month, day, hour, minute, 0, tzOff, locA.Lat, locA.Lng, name)
	hcB := dignity.ComputeHouseConvergence(dignity.TropicalToLonMap(bcB.Tropical), year, month, day, hour, minute, 0, tzOff, locB.Lat, locB.Lng, name)

	// Phase 4: Timing convergence (location-invariant, but compute for completeness)
	trA := dignity.ComputeTimingReport(name, year, month, day, hour, minute, tzOff, locA.Lat, locA.Lng, targetDate, dignity.TropicalToLonMap(bcA.Tropical), bcA.Ayanamsa, bcA.ASC)
	trB := dignity.ComputeTimingReport(name, year, month, day, hour, minute, tzOff, locB.Lat, locB.Lng, targetDate, dignity.TropicalToLonMap(bcB.Tropical), bcB.Ayanamsa, bcB.ASC)

	// Build ASC comparison across 5 systems
	var ascA, ascB []RelocASCEntry
	for _, sys := range dignity.CompareHouseSystems {
		code, ok := swephCode[sys]
		if !ok {
			continue
		}
		_, ascmcA := swe.Houses(bcA.JD, locA.Lat, locA.Lng, code)
		_, ascmcB := swe.Houses(bcB.JD, locB.Lat, locB.Lng, code)
		ascA = append(ascA, RelocASCEntry{System: sys, Sign: dignity.SignForLongitude(ascmcA[0]), Degree: ascmcA[0]})
		ascB = append(ascB, RelocASCEntry{System: sys, Sign: dignity.SignForLongitude(ascmcB[0]), Degree: ascmcB[0]})
	}

	// Build house shift comparison
	var shifts []RelocShiftEntry
	for _, pA := range hcA.Planets {
		// Find matching planet in hcB
		for _, pB := range hcB.Planets {
			if pA.Planet == pB.Planet {
				shifts = append(shifts, RelocShiftEntry{
					Planet:        pA.Planet,
					TropicalSign:  pA.TropicalSign,
					HouseA:        pA.ConsensusHouse(),
					HouseB:        pB.ConsensusHouse(),
					AgreementA:    pA.AgreementCount(),
					AgreementB:    pB.AgreementCount(),
					StableA:       pA.IsUnambiguous(),
					StableB:       pB.IsUnambiguous(),
					ShiftReliable: pA.IsUnambiguous() && pB.IsUnambiguous(),
				})
				break
			}
		}
	}

	// Build response
	return &RelocationResponse{
		Name:       name,
		LocationA:  locA,
		LocationB:  locB,
		TargetDate: targetDate,
		HouseConvergenceA: RelocConvergence{
			UnambiguousCount: hcA.UnambiguousCount(),
			ConvergenceRate:  hcA.ConvergenceRate(),
		},
		HouseConvergenceB: RelocConvergence{
			UnambiguousCount: hcB.UnambiguousCount(),
			ConvergenceRate:  hcB.ConvergenceRate(),
		},
		ASCA: ascA,
		ASCB: ascB,
		TimingConvergenceA: RelocTimingSummary{
			HasConvergence: trA.TimingConvergence.HasConvergence,
			Planets:        trA.TimingConvergence.PlanetConvergences,
		},
		TimingConvergenceB: RelocTimingSummary{
			HasConvergence: trB.TimingConvergence.HasConvergence,
			Planets:        trB.TimingConvergence.PlanetConvergences,
		},
		Shifts: shifts,
	}, nil
}

// swephCode maps house system names to Swiss Ephemeris codes.
var swephCode = map[string]byte{
	"placidus":  'P',
	"porphyry":  'O',
	"koch":      'K',
	"equal":     'E',
	"whole_sign": 'W',
}

// computeDraconic builds the draconic chart JSON response.

// computeDraconicSynastry builds the draconic synastry JSON response.

// computeDraconicSynastryFull builds the full three-layer draconic synastry JSON.

// computeDraconicTransits computes transiting planets hitting the draconic chart.

// computeProgressedDraconic computes the progressed draconic chart using the
// current transiting North Node as the zero-point.

// computeDraconicSolarReturn computes the draconic solar return for a target year.

// findSolarReturnJD finds the exact Julian Day when the Sun returns to natalSun longitude
// in the given targetYear. Uses binary search with 40 iterations.

// computeDraconicTransitsCross compares draconic transits in tropical vs sidereal.
// Natal draconic positions are zodiac-invariant. Transiting positions differ by
// the Lahiri ayanamsa (~24°). Returns which aspects survive the zodiac shift.

// computeProgressedCross compares progressed-to-natal aspects in tropical vs sidereal.
// Both natal and progressed positions shift by the same ayanamsa, so angular
// distances are preserved. Near-100% survival expected (Phase 13).

// computeDirections computes primary directions (Ptolemy) for a given age.
// Directs ASC by oblique ascension and MC by right ascension.
func computeDirections(name string, bc *dignity.BaseChart, lat, lng, age float64, orbDeg float64, cacheDir string) (*DirectionsResponse, error) {
	// Natal positions
	natal := dignity.TropicalToLonMap(bc.Tropical)

	// ASC from chart data, MC computed from JD + lat/lng
	ascLon := bc.ASC
	_, ascmc := swe.Houses(bc.JD, lat, lng, 'P')
	mcLon := ascmc[1]

	aspects := dignity.DefaultAspects()
	result := dignity.ComputePrimaryDirections(natal, ascLon, mcLon, lat, age, aspects, orbDeg)

	// Build JSON response
	response := &DirectionsResponse{
		Name:        name,
		Age:         age,
		DirectedASC: result.DirectedASC,
		DirectedMC:  result.DirectedMC,
		Orb:         orbDeg,
		ASCAspects:  make([]DirectionHitJSON, 0),
		MCAspects:   make([]DirectionHitJSON, 0),
	}

	for _, h := range result.ASCAspects {
		response.ASCAspects = append(response.ASCAspects, DirectionHitJSON{h.DirectedPoint, h.NatalPlanet, h.Aspect, h.Orb})
	}
	for _, h := range result.MCAspects {
		response.MCAspects = append(response.MCAspects, DirectionHitJSON{h.DirectedPoint, h.NatalPlanet, h.Aspect, h.Orb})
	}

	return response, nil
}

// computeSolarArc computes solar arc directions for a target date.
func computeSolarArc(name string, bc *dignity.BaseChart, targetDate string, orbDeg float64, cacheDir string) (*SolarArcResponse, error) {
	// Parse target date
	target, err := time.Parse("2006-01-02", targetDate)
	if err != nil {
		return nil, fmt.Errorf("invalid target date: %w", err)
	}

	// Birth date
	birth := time.Date(bc.Year, time.Month(bc.Month), bc.Day, bc.Hour, bc.Minute, 0, 0, time.UTC)

	// Compute secondary progressed Sun (day-for-a-year)
	age := target.Sub(birth).Hours() / (365.2425 * 24)
	progressedJD := bc.JD + age

	// Get progressed Sun position
	progSunLon, _, _, _, err := swe.CalcUTErr(progressedJD, swe.SUN)
	if err != nil {
		return nil, fmt.Errorf("progressed Sun calculation failed: %w", err)
	}

	// Natal positions
	natal := dignity.TropicalToLonMap(bc.Tropical)
	natalSunLon := natal["Sun"]

	// Compute solar arc
	report := dignity.ComputeSolarArc(name, birth, target, natalSunLon, progSunLon, natal, orbDeg)

	return &SolarArcResponse{
		Name:              report.Name,
		BirthDate:         report.BirthDate,
		TargetDate:        report.TargetDate,
		Age:               report.Age,
		SolarArc:          report.SolarArc,
		ProgressedSunLon:  report.ProgressedSunLon,
		NatalSunLon:       report.NatalSunLon,
		DirectedPositions: report.DirectedPositions,
		NatalPositions:    report.NatalPositions,
		Aspects:           report.Aspects,
		TotalAspects:      report.TotalAspects,
	}, nil
}

// computeProfection computes annual profections for a target date.
func computeProfection(name string, bc *dignity.BaseChart, targetDate string) (*ProfectionResponse, error) {
	target, err := time.Parse("2006-01-02", targetDate)
	if err != nil {
		return nil, fmt.Errorf("invalid target date: %w", err)
	}

	birth := time.Date(bc.Year, time.Month(bc.Month), bc.Day, bc.Hour, bc.Minute, 0, 0, time.UTC)
	natal := dignity.TropicalToLonMap(bc.Tropical)

	report := dignity.ComputeProfectionReport(name, birth, target, bc.ASC, natal)

	return &ProfectionResponse{
		Name:           report.Name,
		BirthDate:      report.BirthDate,
		TargetDate:     report.TargetDate,
		Age:            report.Age,
		ProfectionYear: report.ProfectionYear,
		NatalASC:       report.NatalASC,
		ProfectedASC:   report.ProfectedASC,
		ProfectedSign:  report.ProfectedSign,
		ProfectedHouse: report.ProfectedHouse,
		TimeLord:       report.TimeLord,
		TimeLordHouse:  report.TimeLordHouse,
		TimeLordSign:   report.TimeLordSign,
		PlanetsInSign:  report.PlanetsInSign,
	}, nil
}

// computeBiWheel generates a bi-wheel SVG comparing two charts.
func computeBiWheel(inner, outer dignity.BirthData, opts dignity.BiWheelOptions) ([]byte, error) {
	svg := dignity.RenderBiWheelSVG(
		inner.Name, inner.Year, inner.Month, inner.Day, inner.Hour, inner.Minute,
		inner.TZOffset, inner.Lat, inner.Lng,
		outer.Name, outer.Year, outer.Month, outer.Day, outer.Hour, outer.Minute,
		outer.TZOffset, outer.Lat, outer.Lng,
		opts,
	)
	return []byte(svg), nil
}

// computeZodiacalReleasing computes zodiacal releasing periods.
func computeZodiacalReleasing(bd dignity.BirthData, lotType, targetDate string) ([]byte, error) {
	bc := computePositions(bd.Year, bd.Month, bd.Day, bd.Hour, bd.Minute, bd.TZOffset, bd.Lat, bd.Lng, "")
	birth := time.Date(bd.Year, time.Month(bd.Month), bd.Day, bd.Hour, bd.Minute, 0, 0, time.UTC)

	// Determine day/night
	sunLon := bc.Tropical["Sun"].Lon
	ascLon := bc.ASC
	diff := sunLon - ascLon
	if diff < 0 {
		diff += 360
	}
	isDay := diff < 180

	report := dignity.ComputeZodiacalReleasing(
		bd.Name, birth, bc.ASC,
		bc.Tropical["Sun"].Lon, bc.Tropical["Moon"].Lon,
		isDay, lotType, targetDate,
	)
	return json.Marshal(report)
}

// computeHorary judges a horary question.
func computeHorary(bd dignity.BirthData, question string) ([]byte, error) {
	bc := computePositions(bd.Year, bd.Month, bd.Day, bd.Hour, bd.Minute, bd.TZOffset, bd.Lat, bd.Lng, "")
	judgment, err := dignity.ComputeHoraryJudgment(bc, question)
	if err != nil {
		return nil, err
	}
	return judgment.JSON()
}

// computeImport imports charts from external formats.
func computeImport(data string) ([]byte, error) {
	result, err := dignity.ImportCharts(data)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

// computeInterpretation produces a natural-language chart interpretation.
// system: SystemKoiné (Hellenistic, default) or SystemWestern (modern).
func computeInterpretation(name string, bc *dignity.BaseChart, lat, lng float64, houseSystem string, orbDeg float64, system string, cacheDir string) ([]byte, error) {
	bc.Name = name
	switch system {
	case string(dignity.SystemWestern):
		report := dignity.WesternFromBase(bc, orbDeg, false)
		return report.JSON()
	case "vedic":
		report := dignity.ComputeVedicNatalReport(bc)
		return report.JSON()
	case "bazi":
		report := dignity.BaZiFromBase(bc)
		return json.Marshal(report)
	default:
		report := dignity.KoinéFromBase(bc, orbDeg)
		return report.JSON()
	}
}

// computeAstroCartography computes planetary lines for a world map.
// frame: FrameTropical, FrameDraconic, or FrameCross.
func computeAstroCartography(name string, bc *dignity.BaseChart, latStep float64, frame string, cacheDir string) (*AstroCartographyResponse, error) {
	gmst := dignity.ComputeGMST(bc.JD)

	// Determine planet positions based on frame
	positions := dignity.TropicalToLonMap(bc.Tropical)
	nnLon := bc.NorthNode
	switch frame {
	case string(dignity.FrameDraconic):
		positions = make(map[string]float64)
		for p, lon := range dignity.TropicalToLonMap(bc.Tropical) {
			positions[p] = dignity.NormalizeLon(lon - nnLon)
		}
	case string(dignity.FrameCross):
		// Cross: tropical positions for MC/IC, draconic for ASC/DSC
		// We handle this per-planet below
	}

	var lines []AstroCartographyLineJSON
	for planet, lon := range dignity.TropicalToLonMap(bc.Tropical) {
		tropRA := dignity.LonToRA(lon, dignity.ObliquityDeg)

		// MC and IC lines — use tropical RA for all frames (RA-based, but
		// draconic shift changes RA nonlinearly, so we use frame-specific RA)
		var ra float64
		var ascLon float64
		if frame == string(dignity.FrameCross) {
			ra = tropRA // cross MC/IC = tropical
			ascLon = dignity.NormalizeLon(lon - nnLon) // cross ASC/DSC = draconic
		} else if frame == string(dignity.FrameDraconic) {
			dracLon := dignity.NormalizeLon(lon - nnLon)
			ra = dignity.LonToRA(dracLon, dignity.ObliquityDeg)
			ascLon = dracLon
		} else {
			ra = tropRA
			ascLon = lon
		}

		lines = append(lines, AstroCartographyLineJSON{
			Planet: planet,
			Angle:  "MC",
			Points: dignity.ComputeMCLine(ra, gmst, latStep),
		})
		lines = append(lines, AstroCartographyLineJSON{
			Planet: planet,
			Angle:  "IC",
			Points: dignity.ComputeICLine(ra, gmst, latStep),
		})

		// ASC and DSC lines — use corrected binary search from dignity package
		ascPoints := dignity.ComputeASCLine(ascLon, bc.JD, latStep, swe.Houses)
		lines = append(lines, AstroCartographyLineJSON{
			Planet: planet,
			Angle:  "ASC",
			Points: ascPoints,
		})
		dscPoints := dignity.ComputeDSCLine(ascLon, bc.JD, latStep, swe.Houses)
		lines = append(lines, AstroCartographyLineJSON{
			Planet: planet,
			Angle:  "DSC",
			Points: dscPoints,
		})
	}

	return &AstroCartographyResponse{
		Name:  name,
		JD:    bc.JD,
		GMST:  gmst,
		Frame: frame,
		Lines: lines,
	}, nil
}

// computeAstroCartographyCompare returns all three frames plus LinesNear at a target location.
func computeAstroCartographyCompare(name string, bc *dignity.BaseChart, latStep float64, targetLat, targetLng, orb float64, cacheDir string) (*AstroCartographyCompareResponse, error) {
	// Compute all three frames
	tropResp, _ := computeAstroCartography(name, bc, latStep, string(dignity.FrameTropical), cacheDir)
	dracResp, _ := computeAstroCartography(name, bc, latStep, string(dignity.FrameDraconic), cacheDir)
	crossResp, _ := computeAstroCartography(name, bc, latStep, string(dignity.FrameCross), cacheDir)

	// Convert to AstroLine slices
	tropLines := make([]dignity.AstroLine, len(tropResp.Lines))
	for i, l := range tropResp.Lines {
		tropLines[i] = dignity.AstroLine{Planet: l.Planet, Angle: l.Angle, Points: l.Points}
	}
	dracLines := make([]dignity.AstroLine, len(dracResp.Lines))
	for i, l := range dracResp.Lines {
		dracLines[i] = dignity.AstroLine{Planet: l.Planet, Angle: l.Angle, Points: l.Points}
	}
	crossLines := make([]dignity.AstroLine, len(crossResp.Lines))
	for i, l := range crossResp.Lines {
		crossLines[i] = dignity.AstroLine{Planet: l.Planet, Angle: l.Angle, Points: l.Points}
	}

	hits := dignity.CompareLinesNear(targetLat, targetLng, tropLines, dracLines, crossLines, orb)

	return &AstroCartographyCompareResponse{
		Name:      name,
		TargetLat: targetLat,
		TargetLng: targetLng,
		Orb:       orb,
		Hits:      hits,
	}, nil
}

// computeAstroCartographyParans finds MC/IC × ASC/DSC line intersections.
func computeAstroCartographyParans(name string, bc *dignity.BaseChart, latStep float64, frame string, cacheDir string) (*AstroCartographyParansResponse, error) {
	gmst := dignity.ComputeGMST(bc.JD)

	// Determine planet positions based on frame
	tropPositions := dignity.TropicalToLonMap(bc.Tropical)
	nnLon := bc.NorthNode

	var intersections []dignity.ParanIntersection

	switch frame {
	case string(dignity.FrameDraconic):
		dracPositions := make(map[string]float64)
		for p, lon := range tropPositions {
			dracPositions[p] = dignity.NormalizeLon(lon - nnLon)
		}
		intersections = dignity.FindParans(dracPositions, bc.JD, gmst, swe.Houses)

	case string(dignity.FrameCross):
		// Cross: tropical positions for MC/IC, draconic for ASC/DSC.
		dracPositions := make(map[string]float64)
		for p, lon := range tropPositions {
			dracPositions[p] = dignity.NormalizeLon(lon - nnLon)
		}
		intersections = dignity.FindParansCross(tropPositions, dracPositions, bc.JD, gmst, swe.Houses)

	default: // tropical
		intersections = dignity.FindParans(tropPositions, bc.JD, gmst, swe.Houses)
	}

	dignity.GeocodeParans(intersections)

	return &AstroCartographyParansResponse{
		Name:         name,
		Frame:        frame,
		Intersections: intersections,
	}, nil
}

// computeElectional scores dates in a range for launch/event timing.
func computeElectional(name string, bc *dignity.BaseChart, lat, lng float64, startDate, endDate string, orbDeg float64, cacheDir string) (*ElectionalResponse, error) {
	// Parse dates
	var sy, sm, sd, ey, em, ed int
	fmt.Sscanf(startDate, "%d-%d-%d", &sy, &sm, &sd)
	fmt.Sscanf(endDate, "%d-%d-%d", &ey, &em, &ed)

	startJD := swe.Julday(sy, sm, sd, 12.0, true)
	endJD := swe.Julday(ey, em, ed, 12.0, true)

	// Natal ASC for house computation
	_, ascmc := swe.Houses(bc.JD, lat, lng, 'P')
	ascSign := int(ascmc[0] / 30)

	// Natal planet positions
	natal := dignity.TropicalToLonMap(bc.Tropical)

	var results []ElectionalDayScore

	for jd := startJD; jd <= endJD; jd++ {
		// Transit positions at noon UT
		transit := make(map[string]float64)
		planetIDs := dignity.ElectionalPlanets
		for _, p := range planetIDs {
			lon, _, _, _ := swe.CalcUT(jd, p.ID)
			for lon < 0 {
				lon += 360
			}
			for lon >= 360 {
				lon -= 360
			}
			transit[p.Name] = lon
		}

		// Moon house (whole-sign from natal ASC)
		moonLon := transit["Moon"]
		moonHouse := ((int(moonLon/30) - ascSign + 12) % 12) + 1
		moonSign := dignity.SignForLongitude(moonLon)

		// Moon score
		moonScore := map[int]int{10: 3, 11: 3, 1: 2, 5: 2, 9: 1, 3: 1, 7: 0, 2: -1, 6: -1, 8: -2, 12: -3}[moonHouse]

		// Mercury sign score
		mercLon := transit["Mercury"]
		mercSign := dignity.SignForLongitude(mercLon)
		mercScore := 0
		switch mercSign {
		case "Gemini":
			mercScore = 2
		case "Virgo":
			mercScore = 1
		case "Sagittarius", "Pisces":
			mercScore = -1
		}

		// Bad aspects
		var bad []string
		// Mars square/opposition natal Mercury
		marsMercDist := dignity.AngleDist(transit["Mars"], natal["Mercury"])
		for _, a := range []struct{ name string; angle float64 }{{"square", 90}, {"opposition", 180}} {
			diff := math.Abs(marsMercDist - a.angle)
			if diff < orbDeg {
				bad = append(bad, fmt.Sprintf("Mars %s natal Mercury (%.1f°)", a.name, diff))
			}
		}
		// Neptune opposite natal Uranus
		nepUraDist := dignity.AngleDist(transit["Neptune"], natal["Uranus"])
		if math.Abs(nepUraDist-180) < 1.5 {
			bad = append(bad, fmt.Sprintf("Neptune opposite natal Uranus (%.1f°)", math.Abs(nepUraDist-180)))
		}
		// Saturn conjunct natal Venus
		satVenDist := dignity.AngleDist(transit["Saturn"], natal["Venus"])
		if satVenDist < 1.0 {
			bad = append(bad, fmt.Sprintf("Saturn conjunct natal Venus (%.1f°)", satVenDist))
		}

		// Good aspects
		var good []string
		// Uranus trine natal Mercury
		uraMercDist := dignity.AngleDist(transit["Uranus"], natal["Mercury"])
		if math.Abs(uraMercDist-120) < orbDeg {
			good = append(good, fmt.Sprintf("Uranus trine natal Mercury (%.1f°)", math.Abs(uraMercDist-120)))
		}
		// Pluto trine natal Jupiter
		pluJupDist := dignity.AngleDist(transit["Pluto"], natal["Jupiter"])
		if math.Abs(pluJupDist-120) < orbDeg {
			good = append(good, fmt.Sprintf("Pluto trine natal Jupiter (%.1f°)", math.Abs(pluJupDist-120)))
		}
		// Sun sextile natal Chiron
		sunChiDist := dignity.AngleDist(transit["Sun"], natal["Chiron"])
		if math.Abs(sunChiDist-60) < orbDeg {
			good = append(good, fmt.Sprintf("Sun sextile natal Chiron (%.1f°)", math.Abs(sunChiDist-60)))
		}
		// Venus in H9 or H10
		venLon := transit["Venus"]
		venHouse := ((int(venLon/30) - ascSign + 12) % 12) + 1
		if venHouse == 9 || venHouse == 10 {
			good = append(good, fmt.Sprintf("Venus in H%d", venHouse))
		}

		score := moonScore + mercScore + len(good) - len(bad)*2

		// Day name
		y, m, d, _ := swe.Revjul(jd)
		dayNames := []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}
		// Compute day of week from JD
		dow := (int(jd+1.5) % 7)
		if dow < 0 {
			dow += 7
		}

		results = append(results, ElectionalDayScore{
			Date:      fmt.Sprintf("%04d-%02d-%02d", y, m, d),
			Day:       dayNames[dow],
			Score:     score,
			MoonHouse: moonHouse,
			MoonSign:  moonSign,
			MercSign:  mercSign,
			Good:      good,
			Bad:       bad,
		})
	}

	// Sort by score descending, then date ascending
	sort.Slice(results, func(i, j int) bool {
		if results[i].Score != results[j].Score {
			return results[i].Score > results[j].Score
		}
		return results[i].Date < results[j].Date
	})

	return &ElectionalResponse{
		Name:    name,
		Start:   startDate,
		End:     endDate,
		Orb:     orbDeg,
		Results: results,
	}, nil
}

// computeStars computes fixed star conjunctions for a natal chart.
func computeStars(name string, bc *dignity.BaseChart, orbDeg float64, cacheDir string) (*StarsResponse, error) {
	// Compute star positions at birth JD
	starPositions := make(map[string]float64)
	for _, starName := range dignity.StarNames {
		lon, _, _, _ := swe.Fixstar(starName, bc.JD)
		if lon != 0 {
			starPositions[starName] = dignity.NormalizeLon(lon)
		}
	}

	// Build planet position map
	planetPositions := dignity.TropicalToLonMap(bc.Tropical)

	conjunctions := dignity.FindStarConjunctions(starPositions, planetPositions, orbDeg)

	// Build JSON response
	response := &StarsResponse{
		Name: name,
		Orb:  orbDeg,
	}
	for _, c := range conjunctions {
		response.Conjunctions = append(response.Conjunctions, StarConjJSON{
			Star:      c.Star,
			StarLon:   c.StarLon,
			Planet:    c.Planet,
			PlanetLon: c.PlanetLon,
			Orb:       c.Orb,
			Meaning:   c.Meaning,
		})
	}

	return response, nil
}

// computeStarsCross compares star conjunctions in tropical vs sidereal frames.
func computeStarsCross(name string, bc *dignity.BaseChart, orbDeg float64, cacheDir string) (*dignity.StarCrossSystem, error) {
	// Compute star positions at birth JD (tropical)
	starPositions := make(map[string]float64)
	for _, starName := range dignity.StarNames {
		lon, _, _, _ := swe.Fixstar(starName, bc.JD)
		if lon != 0 {
			starPositions[starName] = dignity.NormalizeLon(lon)
		}
	}

	// Build planet position map
	planetPositions := dignity.TropicalToLonMap(bc.Tropical)

	result := dignity.CompareStarConjunctionsCrossSystem(name, starPositions, planetPositions, bc.Ayanamsa, orbDeg)
	return result, nil
}

// splitComma splits a comma-separated string into trimmed lowercase names.
func splitComma(s string) []string {
	var parts []string
	for _, p := range strings.Split(s, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			parts = append(parts, p)
		}
	}
	return parts
}

// printReport prints a human-readable multi-phase report.
func printReport(fr *dignity.FullReport) {
	fmt.Printf("╔══════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║  Empirical Astrology — Signal Recovery Report           ║\n")
	fmt.Printf("╚══════════════════════════════════════════════════════════╝\n")
	fmt.Printf("\n  Name: %s\n", fr.Name)
	fmt.Printf("  Ayanamsa: %.2f deg Lahiri\n\n", fr.AyanamsaDegrees)

	fmt.Print(dignity.FormatConvergence(fr.Phase1Dignity))
	fmt.Print(dignity.FormatNodeConvergence(fr.Phase5Nodes))
	zcJSON, _ := fr.Phase6Zodiac.ZodiacComparisonJSON()
	var zcMap map[string]interface{}
	json.Unmarshal(zcJSON, &zcMap)
	fmt.Printf("\nPhase 6: Zodiac Comparison\n")
	fmt.Printf("  Winner: %s\n", fr.Phase6Zodiac.Winner())
	fmt.Printf("  Tropical dignity: %.1f%%, Sidereal: %.1f%%\n",
		fr.Phase6Zodiac.Tropical.DignityDensity()*100,
		fr.Phase6Zodiac.Sidereal.DignityDensity()*100)
}

// computeMansionConvergence computes nakshatra/xiu mansion placements for a chart.
func computeMansionConvergence(name string, bc *dignity.BaseChart, cacheDir string) (*dignity.MansionConvergence, error) {
	// Build tropical planet map
	tropical := dignity.TropicalToLonMap(bc.Tropical)

	// Ayanamsa already computed in BaseChart
	result := dignity.ComputeMansionConvergence(name, tropical, bc.Ayanamsa)
	return result, nil
}

// computeArabicParts computes Arabic Parts with cross-system comparison.
func computeArabicParts(name string, bc *dignity.BaseChart, orbDeg float64, cacheDir string) (*dignity.PartCrossSystem, error) {
	// Build tropical planet map
	tropical := dignity.TropicalToLonMap(bc.Tropical)

	// Determine day/night: Sun above horizon = day
	// Sun is above horizon if its longitude is between ASC and DSC (asc+180)
	sunLon := tropical["Sun"]
	asc := bc.ASC
	dsc := dignity.NormalizeLon(asc + 180)
	var isDay bool
	if asc < dsc {
		isDay = sunLon >= asc && sunLon < dsc
	} else {
		isDay = sunLon >= asc || sunLon < dsc
	}

	result := dignity.ComputePartCrossSystem(name, asc, tropical, bc.Ayanamsa, isDay, orbDeg)
	return result, nil
}

// computeTropicalSolarReturn computes the tropical solar return for a target year.
// Returns positions, ASC/MC, aspects to natal, and patterns.
func computeTropicalSolarReturn(name string, bc *dignity.BaseChart, targetYear int, cacheDir string) (*SolarReturnResponse, error) {
	natalSun := dignity.TropicalToLonMap(bc.Tropical)["Sun"]
	jdSR := findSolarReturnJD(natalSun, targetYear, bc.JD)

	// Full planet set for solar return (non-TNP: Sun-Pluto+Nodes+asteroids+Chiron+Lilith+dwarfs)
	planetIDs := dignity.AllPlanets[:22]

	srPositions := make(map[string]float64)
	for _, p := range planetIDs {
		lon, _, _, _ := swe.CalcUT(jdSR, p.ID)
		srPositions[p.Name] = dignity.NormalizeLon(lon)
	}

	// ASC/MC at solar return
	_, ascmc := swe.Houses(jdSR, 0, 0, 'P')
	srPositions["Ascendant"] = ascmc[0]
	srPositions["Midheaven"] = ascmc[1]

	// SR-to-natal aspects
	natal := dignity.TropicalToLonMap(bc.Tropical)
	natal["Ascendant"] = bc.ASC

	aspects := dignity.DefaultAspects()
	orb := 3.0
	var srAspects []dignity.SynastryHit

	// Same-body aspects: SR Sun to natal Sun, etc.
	for _, p := range planetIDs {
		srLon, ok := srPositions[p.Name]
		if !ok {
			continue
		}
		natLon, ok := natal[p.Name]
		if !ok {
			continue
		}
		dist := angularDistance(srLon, natLon)
		for _, a := range aspects {
			diff := math.Abs(dist - a.Angle)
			if diff <= orb {
				srAspects = append(srAspects, dignity.SynastryHit{
					Planet1: "SR_" + p.Name,
					Planet2: "Natal_" + p.Name,
					Aspect:  a.Name,
					Orb:     math.Round(diff*100) / 100,
				})
			}
		}
	}

	// Cross-body aspects: SR planet to natal planet
	for _, p1 := range planetIDs {
		srLon, ok1 := srPositions[p1.Name]
		if !ok1 {
			continue
		}
		for _, p2 := range planetIDs {
			if p1.Name == p2.Name {
				continue
			}
			natLon, ok2 := natal[p2.Name]
			if !ok2 {
				continue
			}
			dist := angularDistance(srLon, natLon)
			for _, a := range aspects {
				diff := math.Abs(dist - a.Angle)
				if diff <= orb {
					srAspects = append(srAspects, dignity.SynastryHit{
						Planet1: "SR_" + p1.Name,
						Planet2: "Natal_" + p2.Name,
						Aspect:  a.Name,
						Orb:     math.Round(diff*100) / 100,
					})
				}
			}
		}
	}

	// Sort by orb
	for i := 0; i < len(srAspects); i++ {
		for j := i + 1; j < len(srAspects); j++ {
			if srAspects[j].Orb < srAspects[i].Orb {
				srAspects[i], srAspects[j] = srAspects[j], srAspects[i]
			}
		}
	}

	// Pattern detection on SR chart (non-TNP bodies only)
	nonTNP := dignity.NonTNPPlanetNames
	srNonTNP := make(map[string]float64)
	for _, name := range nonTNP {
		if lon, ok := srPositions[name]; ok {
			srNonTNP[name] = lon
		}
	}
	patternReport := dignity.DetectPatterns(srNonTNP, 5.0)

	// Format datetime
	yr, mo, dy, hr := swe.Revjul(jdSR)
	dtStr := fmt.Sprintf("%d-%02d-%02d %02d:%02d UT", yr, mo, dy, int(hr), int((hr-float64(int(hr)))*60))

	return &SolarReturnResponse{
		Name:       name,
		TargetYear: targetYear,
		JD:         jdSR,
		DateTime:   dtStr,
		Positions:  srPositions,
		Aspects:    srAspects,
		Patterns:   patternReport.Patterns,
	}, nil
}

// computeProgressed computes a secondary progressed chart and progressed-to-natal aspects.
func computeProgressed(name string, bc *dignity.BaseChart, targetDate string, orbDeg float64, cacheDir string) (*dignity.ProgressedReport, error) {
	// Parse target date
	var y, m, d int
	fmt.Sscanf(targetDate, "%d-%d-%d", &y, &m, &d)
	utHour := 12.0
	targetJD := swe.Julday(y, m, d, utHour, true)

	// Age in years
	age := (targetJD - bc.JD) / 365.2425

	// Progressed JD: birthJD + age in days (day-for-a-year)
	progJD := bc.JD + age

	// Compute progressed positions
	planetIDs := dignity.AllPlanets[:22]

	progPositions := make(map[string]float64)
	for _, p := range planetIDs {
		lon, _, _, _ := swe.CalcUT(progJD, p.ID)
		for lon < 0 {
			lon += 360
		}
		for lon >= 360 {
			lon -= 360
		}
		progPositions[p.Name] = lon
	}

	// Natal positions
	natalPositions := dignity.TropicalToLonMap(bc.Tropical)

	report := dignity.ComputeProgressedReport(name, targetDate, age, progPositions, natalPositions, orbDeg)
	return report, nil
}

// computeComposite computes a midpoint composite chart for two people.
func computeComposite(name1, name2 string, bc1, bc2 *dignity.BaseChart, orbDeg float64, cacheDir string) (*dignity.CompositeReport, error) {
	// Build tropical planet maps with Ascendant
	chart1 := dignity.TropicalToLonMap(bc1.Tropical)
	chart1["Ascendant"] = bc1.ASC

	chart2 := dignity.TropicalToLonMap(bc2.Tropical)
	chart2["Ascendant"] = bc2.ASC

	report := dignity.ComputeCompositeReport(name1, name2, chart1, chart2, orbDeg)
	return report, nil
}

// computeTraditional computes traditional astrology interpretive data.
func computeTraditional(name string, bc *dignity.BaseChart) (dignity.TraditionalReport, error) {
	report := dignity.ComputeTraditionalReport(name, dignity.TropicalToLonMap(bc.Tropical), buildSpeeds(bc.Tropical))
	return report, nil
}

// computeUranian computes Uranian/Hamburg School midpoint analysis.
func computeUranian(name string, bc *dignity.BaseChart) (uranian.MidpointReport, error) {
	// Compute houses for the chart
	cusps, ascmc := swe.Houses(bc.JD, bc.Lat, bc.Lng, 'P')
	houses := make(map[string]float64)
	for i := 1; i <= 12; i++ {
		houses[fmt.Sprintf("H%d", i)] = cusps[i]
	}
	// Angles: ASC, MC, DSC (ASC+180), IC (MC+180)
	angles := map[string]float64{
		"ASC": ascmc[0],
		"MC":  ascmc[1],
		"DSC": math.Mod(ascmc[0]+180, 360),
		"IC":  math.Mod(ascmc[1]+180, 360),
	}
	report := uranian.ComputeMidpointReport(name, dignity.TropicalToLonMap(bc.Tropical), houses, angles)
	return report, nil
}

// computeHarmonic computes Addey-style harmonic charts.
func computeHarmonic(name string, bc *dignity.BaseChart, harmonics []int, orb float64) (harmonic.HarmonicReport, error) {
	report := harmonic.ComputeHarmonicReport(name, dignity.TropicalToLonMap(bc.Tropical), harmonics, orb)
	return report, nil
}

// computeDivisional computes Vedic divisional charts.
func computeDivisional(name string, bc *dignity.BaseChart, year, month, day int) (divisional.DivisionalReport, error) {
	report := divisional.ComputeDivisionalReport(name, dignity.TropicalToLonMap(bc.Tropical), bc.Ayanamsa, year, month, day)
	return report, nil
}

// computeParans computes fixed star parans.
func computeParans(name string, bc *dignity.BaseChart, orb float64, cacheDir string) (parans.ParansReport, error) {
	// Compute star positions
	starPositions := make(map[string]float64)
	for _, starName := range dignity.StarNames {
		lon, _, _, _ := swe.Fixstar(starName, bc.JD)
		if lon > 0 {
			starPositions[starName] = lon
		}
	}
	// Compute MC for angles
	_, ascmc := swe.Houses(bc.JD, 0, 0, 'P')
	mc := ascmc[1]
	report := parans.ComputeParansReport(name, starPositions, dignity.TropicalToLonMap(bc.Tropical), bc.ASC, mc, orb)
	return report, nil
}

// computeDeclination computes declination parallels.
func computeDeclination(name string, bc *dignity.BaseChart, orb float64) (declination.DeclinationReport, error) {
	// Extract lon/lat from BaseChart — no SWE re-computation needed.
	positions := make(map[string][2]float64, len(bc.Tropical))
	for k, v := range bc.Tropical {
		positions[k] = [2]float64{v.Lon, v.Lat}
	}
	report := declination.ComputeDeclinationReport(name, positions, orb)
	return report, nil
}

// computeFirdaria computes Persian firdaria planetary periods.
func computeFirdaria(name string, bc *dignity.BaseChart, year, month, day int) (firdaria.FirdariaReport, error) {
	// Determine if Sun is above horizon (diurnal chart) using BaseChart data.
	sunLon := dignity.TropicalToLonMap(bc.Tropical)["Sun"]
	sunAbove := false
	diff := math.Mod(sunLon-bc.ASC+360, 360)
	if diff < 180 {
		sunAbove = true
	}
	report := firdaria.ComputeFirdaria(name, sunAbove, year, month, day)
	return report, nil
}

// angularDistance returns the shortest angular distance between two longitudes.
func angularDistance(a, b float64) float64 {
	d := a - b
	if d < 0 {
		d = -d
	}
	if d > 180 {
		d = 360 - d
	}
	return d
}


// buildSpeeds extracts a speed map from BaseChart tropical positions.
func buildSpeeds(tropical map[string]dignity.Position) map[string]float64 {
	m := make(map[string]float64, len(tropical))
	for k, v := range tropical {
		m[k] = v.Speed
	}
	return m
}

// formatVedicNatal formats a VedicNatalReport as human-readable text.
func formatVedicNatal(r *dignity.VedicNatalReport) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Vedic Natal Horoscope — %s\n", r.Name))
	b.WriteString(fmt.Sprintf("Ayanamsa: %.2f° (Lahiri)\n\n", r.Ayanamsa))

	asc := r.Ascendant
	b.WriteString(fmt.Sprintf("Lagna: %s · %s · Pada %d · Ruler: %s\n\n",
		asc.SiderealSign, asc.Nakshatra, asc.NakshatraPada, asc.NakshatraRuler))

	// House lords
	b.WriteString("House Lords:\n")
	for h := 1; h <= 12; h++ {
		if lord, ok := r.HouseLords[h]; ok {
			b.WriteString(fmt.Sprintf("  %2d: %-10s", h, lord))
			if h%4 == 0 {
				b.WriteString("\n")
			}
		}
	}
	if len(r.HouseLords)%4 != 0 {
		b.WriteString("\n")
	}
	b.WriteString("\n")

	// Planet table
	b.WriteString(fmt.Sprintf("%-10s %-12s %-22s %-5s %-8s %-3s %-3s %-4s %-16s %s\n",
		"Planet", "Sign", "Nakshatra", "Pada", "NkLord", "H", "NL", "R/C", "Dignity", "Conv"))
	b.WriteString(strings.Repeat("-", 120) + "\n")
	for _, p := range r.Planets {
		rc := ""
		if p.Retrograde {
			rc += "R"
		}
		if p.Combust {
			rc += "C"
		}
		if rc == "" {
			rc = "—"
		}
		nl := ""
		if p.NakshatraLordHouse > 0 {
			nl = fmt.Sprintf("%d", p.NakshatraLordHouse)
		} else {
			nl = "—"
		}
		b.WriteString(fmt.Sprintf("%-10s %-12s %-22s %-5d %-8s %-3d %-3s %-4s %-16s %s\n",
			p.Planet, p.SiderealSign, p.Nakshatra, p.NakshatraPada,
			p.NakshatraRuler, p.House, nl, rc, p.Dignity, p.Convergence))
	}

	b.WriteString(fmt.Sprintf("\nR = Retrograde  C = Combust  NL = Nakshatra Lord House\n"))
	b.WriteString(fmt.Sprintf("Signal: %d/%d planets agree with Western dignity\n\n",
		r.SignalCount, r.TotalPlanets))

	// Vedic Aspects (Drishti)
	if len(r.Aspects) > 0 {
		b.WriteString("Vedic Aspects (Drishti):\n")
		for _, a := range r.Aspects {
			b.WriteString(fmt.Sprintf("  %-10s (%dH) %3s → %-10s (%dH)\n",
				a.FromPlanet, a.FromHouse, a.Type, a.ToPlanet, a.ToHouse))
		}
		b.WriteString("\n")
	}

	// Yogas
	if len(r.Yogas) > 0 {
		b.WriteString("Yogas:\n")
		for _, y := range r.Yogas {
			b.WriteString(fmt.Sprintf("  %-20s [%s] %s\n", y.Name, y.Category, y.Description))
		}
		b.WriteString("\n")
	}

	// Varga Charts
	if len(r.Vargas) > 0 {
		vargaOrder := []string{"D3", "D7", "D9", "D10"}
		vargaNames := map[string]string{
			"D3": "Drekkana (Siblings)", "D7": "Saptamsha (Children)",
			"D9": "Navamsha (Marriage)", "D10": "Dashamsha (Career)",
		}
		for _, v := range vargaOrder {
			positions, ok := r.Vargas[v]
			if !ok || len(positions) == 0 {
				continue
			}
			b.WriteString(fmt.Sprintf("%s:\n", vargaNames[v]))
			for _, p := range positions {
				b.WriteString(fmt.Sprintf("  %-10s %s\n", p.Planet, p.Sign))
			}
			b.WriteString("\n")
		}
	}

	// Mahadasha
	b.WriteString("Vimshottari Mahadasha:\n")
	for _, d := range r.Dasha {
		marker := ""
		if d.Start <= "2026-07-22" && "2026-07-22" < d.End {
			marker = " ← CURRENT"
		}
		b.WriteString(fmt.Sprintf("  %-10s %s → %s  (%.1fy)%s\n",
			d.Planet, d.Start, d.End, d.Years, marker))
	}

	// Antardasha
	if len(r.Antardasha) > 0 {
		b.WriteString("\nCurrent Antardasha (Bhukti):\n")
		for _, d := range r.Antardasha {
			marker := ""
			if d.Start <= "2026-07-22" && "2026-07-22" < d.End {
				marker = " ← CURRENT"
			}
			b.WriteString(fmt.Sprintf("  %-10s %s → %s  (%.2fy)%s\n",
				d.Planet, d.Start, d.End, d.Years, marker))
		}
	}

	return b.String()
}
