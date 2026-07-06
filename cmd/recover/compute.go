package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strings"

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

// ── Response types ──────────────────────────────────────────────────────────

// TransitHitJSON is a single transit hit in JSON output.
type TransitHitJSON struct {
	TransitPlanet string  `json:"transit_planet"`
	NatalPlanet   string  `json:"natal_planet"`
	Aspect        string  `json:"aspect"`
	Orb           float64 `json:"orb"`
	StartDate     string  `json:"start_date"`
	EndDate       string  `json:"end_date"`
}

// TransitsResponse is the JSON response for /api/transits.
type TransitsResponse struct {
	Name       string          `json:"name"`
	Sidereal   bool            `json:"sidereal"`
	Transits   []TransitHitJSON `json:"transits"`
	SkyWeather []TransitHitJSON `json:"sky_weather"`
}

// SynastryResponse is the JSON response for /api/synastry.
type SynastryResponse struct {
	Name1   string                `json:"name1"`
	Name2   string                `json:"name2"`
	Aspects []dignity.SynastryHit `json:"aspects"`
}

// RelocationResponse is the JSON response for /api/relocation-compare.
type RelocationResponse struct {
	Name               string              `json:"name"`
	LocationA          server.LatLng       `json:"location_a"`
	LocationB          server.LatLng       `json:"location_b"`
	TargetDate         string              `json:"target_date"`
	HouseConvergenceA  RelocConvergence    `json:"house_convergence_a"`
	HouseConvergenceB  RelocConvergence    `json:"house_convergence_b"`
	ASCA               []RelocASCEntry     `json:"asc_a"`
	ASCB               []RelocASCEntry     `json:"asc_b"`
	TimingConvergenceA RelocTimingSummary  `json:"timing_convergence_a"`
	TimingConvergenceB RelocTimingSummary  `json:"timing_convergence_b"`
	Shifts             []RelocShiftEntry   `json:"shifts"`
}

// RelocConvergence holds house convergence stats for one location.
type RelocConvergence struct {
	UnambiguousCount int     `json:"unambiguous_count"`
	ConvergenceRate  float64 `json:"convergence_rate"`
}

// RelocASCEntry holds ASC data for one house system.
type RelocASCEntry struct {
	System string  `json:"system"`
	Sign   string  `json:"sign"`
	Degree float64 `json:"degree"`
}

// RelocTimingSummary holds timing convergence summary.
type RelocTimingSummary struct {
	HasConvergence bool     `json:"has_convergence"`
	Planets        []string `json:"planets"`
}

// RelocShiftEntry holds a house shift comparison between two locations.
type RelocShiftEntry struct {
	Planet        string `json:"planet"`
	TropicalSign  string `json:"tropical_sign"`
	HouseA        int    `json:"house_a"`
	HouseB        int    `json:"house_b"`
	AgreementA    int    `json:"agreement_a"`
	AgreementB    int    `json:"agreement_b"`
	StableA       bool   `json:"stable_a"`
	StableB       bool   `json:"stable_b"`
	ShiftReliable bool   `json:"shift_reliable"`
}

// DraconicShiftJSON is a single draconic sign shift entry.
type DraconicShiftJSON struct {
	Planet   string `json:"planet"`
	TropSign string `json:"tropical_sign"`
	DracSign string `json:"draconic_sign"`
}

// DraconicResponse is the JSON response for /api/draconic.
type DraconicResponse struct {
	Name    string                `json:"name"`
	Offset  float64               `json:"offset"`
	Planets map[string]float64    `json:"planets"`
	Shifts  []DraconicShiftJSON   `json:"sign_shifts"`
	Bridges []dignity.SynastryHit `json:"bridges"`
}

// DraconicSynastryResponse is the JSON response for /api/draconic-synastry.
type DraconicSynastryResponse struct {
	Name1 string                `json:"name1"`
	Name2 string                `json:"name2"`
	Hits  []dignity.SynastryHit `json:"hits"`
}

// DraconicSynastryFullResponse is the JSON response for /api/draconic-synastry-full.
type DraconicSynastryFullResponse struct {
	Name1        string                `json:"name1"`
	Name2        string                `json:"name2"`
	DracToDrac   []dignity.SynastryHit `json:"drac_to_drac"`
	TropAToDracB []dignity.SynastryHit `json:"trop_a_to_drac_b"`
	TropBToDracA []dignity.SynastryHit `json:"trop_b_to_drac_a"`
}

// DraconicTransitsResponse is the JSON response for /api/draconic-transits.
type DraconicTransitsResponse struct {
	Name     string          `json:"name"`
	Offset   float64         `json:"offset"`
	Transits []TransitHitJSON `json:"transits"`
}

// ProgressedDraconicResponse is the JSON response for /api/progressed-draconic.
type ProgressedDraconicResponse struct {
	Name          string             `json:"name"`
	Date          string             `json:"date"`
	NatalNN       float64            `json:"natal_nn"`
	CurrentNN     float64            `json:"current_nn"`
	NNShift       float64            `json:"nn_shift"`
	NatalDraconic map[string]float64 `json:"natal_draconic"`
	ProgDraconic  map[string]float64 `json:"progressed_draconic"`
	SignShifts    []DraconicShiftJSON `json:"sign_shifts"`
}

// DraconicSolarReturnResponse is the JSON response for /api/draconic-solar-return.
type DraconicSolarReturnResponse struct {
	Name              string             `json:"name"`
	TargetYear        int                `json:"target_year"`
	JD                float64            `json:"jd"`
	DateTime          string             `json:"datetime"`
	Tropical          map[string]float64 `json:"tropical"`
	Draconic          map[string]float64 `json:"draconic"`
	DraconicByNatalNN map[string]float64 `json:"draconic_by_natal_nn"`
}

// CrossHitJSON is a single cross-system comparison hit.
type CrossHitJSON struct {
	TransitPlanet string  `json:"transit_planet"`
	NatalPlanet   string  `json:"natal_planet"`
	Aspect        string  `json:"aspect"`
	Orb           float64 `json:"orb"`
}

// DraconicTransitsCrossResponse is the JSON response for /api/draconic-transits-cross.
type DraconicTransitsCrossResponse struct {
	Name         string        `json:"name"`
	Offset       float64       `json:"offset"`
	Ayanamsa     float64       `json:"ayanamsa"`
	Orb          float64       `json:"orb"`
	MidDate      string        `json:"mid_date"`
	Survivors    []CrossHitJSON `json:"survivors"`
	TropicalOnly []CrossHitJSON `json:"tropical_only"`
	SiderealOnly []CrossHitJSON `json:"sidereal_only"`
}

// ProgressedCrossHitJSON is a single progressed cross-system hit.
type ProgressedCrossHitJSON struct {
	ProgressedPlanet string  `json:"progressed_planet"`
	NatalPlanet      string  `json:"natal_planet"`
	Aspect           string  `json:"aspect"`
	Orb              float64 `json:"orb"`
}

// ProgressedCrossResponse is the JSON response for /api/progressed-cross.
type ProgressedCrossResponse struct {
	Name         string                   `json:"name"`
	TargetDate   string                   `json:"target_date"`
	Age          float64                  `json:"age_years"`
	Ayanamsa     float64                  `json:"ayanamsa"`
	Orb          float64                  `json:"orb"`
	Survivors    []ProgressedCrossHitJSON `json:"survivors"`
	TropicalOnly []ProgressedCrossHitJSON `json:"tropical_only"`
	SiderealOnly []ProgressedCrossHitJSON `json:"sidereal_only"`
}

// DirectionHitJSON is a single primary direction hit.
type DirectionHitJSON struct {
	DirectedPoint string  `json:"directed_point"`
	NatalPlanet   string  `json:"natal_planet"`
	Aspect        string  `json:"aspect"`
	Orb           float64 `json:"orb"`
}

// DirectionsResponse is the JSON response for /api/directions.
type DirectionsResponse struct {
	Name        string             `json:"name"`
	Age         float64            `json:"age_years"`
	DirectedASC float64            `json:"directed_asc"`
	DirectedMC  float64            `json:"directed_mc"`
	Orb         float64            `json:"orb"`
	ASCAspects  []DirectionHitJSON `json:"asc_aspects"`
	MCAspects   []DirectionHitJSON `json:"mc_aspects"`
}

// AstroCartographyLineJSON is a single planetary line in astrocartography output.
type AstroCartographyLineJSON struct {
	Planet string            `json:"planet"`
	Angle  string            `json:"angle"`
	Points []dignity.GeoPoint `json:"points"`
}

// AstroCartographyResponse is the JSON response for /api/astrocartography.
type AstroCartographyResponse struct {
	Name  string                     `json:"name"`
	JD    float64                    `json:"jd"`
	GMST  float64                    `json:"gmst"`
	Frame string                     `json:"frame"`
	Lines []AstroCartographyLineJSON `json:"lines"`
}

// AstroCartographyCompareResponse is the JSON response for /api/astrocartography-compare.
type AstroCartographyCompareResponse struct {
	Name      string                `json:"name"`
	TargetLat float64               `json:"target_lat"`
	TargetLng float64               `json:"target_lng"`
	Orb       float64               `json:"orb"`
	Hits      []dignity.ThreeWayHit `json:"hits"`
}

// ElectionalDayScore is a single day's score in electional output.
type ElectionalDayScore struct {
	Date      string   `json:"date"`
	Day       string   `json:"day"`
	Score     int      `json:"score"`
	MoonHouse int      `json:"moon_house"`
	MoonSign  string   `json:"moon_sign"`
	MercSign  string   `json:"merc_sign"`
	Good      []string `json:"good"`
	Bad       []string `json:"bad"`
}

// ElectionalResponse is the JSON response for /api/electional.
type ElectionalResponse struct {
	Name    string               `json:"name"`
	Start   string               `json:"start_date"`
	End     string               `json:"end_date"`
	Orb     float64              `json:"orb"`
	Results []ElectionalDayScore `json:"results"`
}

// StarConjJSON is a single star conjunction in JSON output.
type StarConjJSON struct {
	Star      string  `json:"star"`
	StarLon   float64 `json:"star_lon"`
	Planet    string  `json:"planet"`
	PlanetLon float64 `json:"planet_lon"`
	Orb       float64 `json:"orb"`
	Meaning   string  `json:"meaning"`
}

// StarsResponse is the JSON response for /api/stars.
type StarsResponse struct {
	Name         string        `json:"name"`
	Orb          float64       `json:"orb"`
	Conjunctions []StarConjJSON `json:"conjunctions"`
}

// SolarReturnResponse is the JSON response for /api/solar-return.
type SolarReturnResponse struct {
	Name       string                `json:"name"`
	TargetYear int                   `json:"target_year"`
	JD         float64               `json:"jd"`
	DateTime   string                `json:"datetime"`
	Positions  map[string]float64    `json:"positions"`
	Aspects    []dignity.SynastryHit `json:"aspects"`
	Patterns   []dignity.Pattern     `json:"patterns"`
}

// computePositions calculates planet longitudes, ayanamsa, ASC, and NN.
// computePositions is a thin wrapper around dignity.ComputeBaseChart.
// It handles ephemeris initialization and returns a *dignity.BaseChart.
func computePositions(year, month, day, hour, minute int, tzOff, lat, lng float64, cacheDir string) *dignity.BaseChart {
	if cacheDir == "" {
		var err error
		cacheDir, err = empirical.EnsureEpheCache()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to initialize ephemeris: %v\n", err)
			os.Exit(1)
		}
	}

	swe.SetEphePath(cacheDir)
	swe.SetSidMode(swe.SIDM_LAHIRI, 0, 0)

	bc, err := dignity.ComputeBaseChart("", year, month, day, hour, minute, 0, tzOff, lat, lng)
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
	allAspects := []dignity.AspectDef{
		{0, "conjunction"}, {60, "sextile"}, {90, "square"}, {120, "trine"}, {180, "opposition"},
	}
	ttHits, err := dignity.ScanTransitToTransit(startDate, endDate, allAspects, orbDeg, compute)
	if err != nil {
		return nil, err
	}
	ttCompact := dignity.CompactTransitsWithRange(ttHits)

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
	aspects := []dignity.AspectDef{
		{0, "conjunction"}, {60, "sextile"}, {90, "square"}, {120, "trine"}, {180, "opposition"},
	}

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

// computeDraconicSynastry builds the draconic synastry JSON response.
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

// computeDraconicSynastryFull builds the full three-layer draconic synastry JSON.
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

// computeDraconicTransits computes transiting planets hitting the draconic chart.
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

// computeProgressedDraconic computes the progressed draconic chart using the
// current transiting North Node as the zero-point.
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

// computeDraconicSolarReturn computes the draconic solar return for a target year.
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
		tropical[p.Name] = normalizeLon(lon)
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
			draconic[name] = normalizeLon(lon - srNode)
		}
	}

	// Draconic using natal NN (soul chart relative to natal soul frame)
	draconicByNatal := make(map[string]float64)
	for name, lon := range tropical {
		draconicByNatal[name] = normalizeLon(lon - bc.NorthNode)
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

// findSolarReturnJD finds the exact Julian Day when the Sun returns to natalSun longitude
// in the given targetYear. Uses binary search with 40 iterations.
func findSolarReturnJD(natalSun float64, targetYear int, natalJD float64) float64 {
	_, birthMo, birthDay, _ := swe.Revjul(natalJD)
	jdMid := swe.Julday(targetYear, birthMo, birthDay, 12.0, true)

	lo := jdMid - 1.0
	hi := jdMid + 1.0

	var jdSR float64
	for i := 0; i < 40; i++ {
		mid := (lo + hi) / 2.0
		sunLon, _, _, _ := swe.CalcUT(mid, swe.SUN)
		sunLon = normalizeLon(sunLon)

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

// computeDraconicTransitsCross compares draconic transits in tropical vs sidereal.
// Natal draconic positions are zodiac-invariant. Transiting positions differ by
// the Lahiri ayanamsa (~24°). Returns which aspects survive the zodiac shift.
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
		tropLon := normalizeLon(lon)
		tropTransits[p.Name] = tropLon
		sidTransits[p.Name] = normalizeLon(tropLon - ayan)
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

// computeProgressedCross compares progressed-to-natal aspects in tropical vs sidereal.
// Both natal and progressed positions shift by the same ayanamsa, so angular
// distances are preserved. Near-100% survival expected (Phase 13).
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

// computeInterpretation produces a natural-language chart interpretation.
func computeInterpretation(name string, bc *dignity.BaseChart, lat, lng float64, houseSystem string, orbDeg float64, cacheDir string) ([]byte, error) {
	bc.Name = name
	report := dignity.KoinéFromBase(bc, orbDeg)
	return report.JSON()
}

// computeAstroCartography computes planetary lines for a world map.
// frame: "tropical", "draconic", or "cross".
func computeAstroCartography(name string, bc *dignity.BaseChart, latStep float64, frame string, cacheDir string) (*AstroCartographyResponse, error) {
	gmst := dignity.ComputeGMST(bc.JD)

	// Determine planet positions based on frame
	positions := dignity.TropicalToLonMap(bc.Tropical)
	nnLon := bc.NorthNode
	switch frame {
	case "draconic":
		positions = make(map[string]float64)
		for p, lon := range dignity.TropicalToLonMap(bc.Tropical) {
			positions[p] = normalizeLon(lon - nnLon)
		}
	case "cross":
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
		if frame == "cross" {
			ra = tropRA // cross MC/IC = tropical
			ascLon = normalizeLon(lon - nnLon) // cross ASC/DSC = draconic
		} else if frame == "draconic" {
			dracLon := normalizeLon(lon - nnLon)
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

		// ASC and DSC lines
		ascPoints := computeASCLineSWE(ascLon, bc.JD, latStep)
		lines = append(lines, AstroCartographyLineJSON{
			Planet: planet,
			Angle:  "ASC",
			Points: ascPoints,
		})
		dscPoints := make([]dignity.GeoPoint, len(ascPoints))
		for i, p := range ascPoints {
			dscPoints[i] = dignity.GeoPoint{Lat: p.Lat, Lon: dignity.NormalizeGeo(p.Lon + 180)}
		}
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
	tropResp, _ := computeAstroCartography(name, bc, latStep, "tropical", cacheDir)
	dracResp, _ := computeAstroCartography(name, bc, latStep, "draconic", cacheDir)
	crossResp, _ := computeAstroCartography(name, bc, latStep, "cross", cacheDir)

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

// computeASCLineSWE finds the ASC line using SWE houses for accurate ASC computation.
func computeASCLineSWE(planetLon, jd, latStep float64) []dignity.GeoPoint {
	var points []dignity.GeoPoint
	for lat := -80.0; lat <= 80.0; lat += latStep {
		lon := findASCLonSWE(planetLon, jd, lat)
		if lon != nil {
			points = append(points, dignity.GeoPoint{Lat: lat, Lon: *lon})
		}
	}
	return points
}

// findASCLonSWE binary-searches geographic longitude where ASC = planetLon.
func findASCLonSWE(planetLon, jd, lat float64) *float64 {
	lo := -180.0
	hi := 180.0

	for i := 0; i < 60; i++ {
		mid := (lo + hi) / 2
		_, ascmc := swe.Houses(jd, lat, mid, 'P')
		asc := ascmc[0]

		diff := dignity.AngleDist(asc, planetLon)
		if diff < 1e-8 {
			return &mid
		}

		testLon := mid + 0.01
		if testLon > 180 {
			testLon = 180
		}
		_, testAscmc := swe.Houses(jd, lat, testLon, 'P')
		testASC := testAscmc[0]

		ascMovesToward := dignity.AngleDist(testASC, planetLon) < diff

		if ascMovesToward {
			if asc < planetLon {
				lo = mid
			} else {
				hi = mid
			}
		} else {
			if asc > planetLon {
				lo = mid
			} else {
				hi = mid
			}
		}
	}

	mid := (lo + hi) / 2
	return &mid
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
	for i := 0; i < len(results); i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].Score > results[i].Score || (results[j].Score == results[i].Score && results[j].Date < results[i].Date) {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

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
			starPositions[starName] = normalizeLon(lon)
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
			starPositions[starName] = normalizeLon(lon)
		}
	}

	// Build planet position map
	planetPositions := dignity.TropicalToLonMap(bc.Tropical)

	result := dignity.CompareStarConjunctionsCrossSystem(name, starPositions, planetPositions, bc.Ayanamsa, orbDeg)
	return result, nil
}

// normalizeLon normalizes a longitude to [0, 360).
func normalizeLon(lon float64) float64 {
	for lon < 0 {
		lon += 360
	}
	for lon >= 360 {
		lon -= 360
	}
	return lon
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
	dsc := normalizeLon(asc + 180)
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

	// Full planet set for solar return
	planetIDs := dignity.AllPlanets[:18] // Sun-Pluto+Node+asteroids+Chiron+Lilith

	srPositions := make(map[string]float64)
	for _, p := range planetIDs {
		lon, _, _, _ := swe.CalcUT(jdSR, p.ID)
		srPositions[p.Name] = normalizeLon(lon)
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
	planetIDs := dignity.AllPlanets

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
func computeUranian(name string, bc *dignity.BaseChart) (uranian.UranianReport, error) {
	// Compute houses for the chart
	cusps, _ := swe.Houses(bc.JD, 0, 0, 'P')
	houses := make(map[string]float64)
	for i := 1; i <= 12; i++ {
		houses[fmt.Sprintf("H%d", i)] = cusps[i]
	}
	report := uranian.ComputeUranianReport(name, dignity.TropicalToLonMap(bc.Tropical), houses)
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
	swe.SetEphePath(cacheDir)
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