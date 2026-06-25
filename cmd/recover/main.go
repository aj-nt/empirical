package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"strconv"

	"github.com/aj-nt/empirical"
	"github.com/aj-nt/empirical/internal/dignity"
	"github.com/aj-nt/empirical/internal/server"
	"github.com/aj-nt/empirical/internal/swe"
)

// mustMarshal marshals v to JSON bytes, returning the error if any.
func mustMarshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// marshalResult marshals a (T, error) pair, propagating the error.
func marshalResult[T any](result T, err error) ([]byte, error) {
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func main() {
	// ── serve subcommand ────────────────────────────────────────────
	if len(os.Args) >= 2 && os.Args[1] == "serve" {
		port := 5000
		if len(os.Args) >= 3 {
			if p, err := strconv.Atoi(os.Args[2]); err == nil && p > 0 && p < 65536 {
				port = p
			}
		}

		// Load ephemeris path
		cacheDir, err := empirical.EnsureEpheCache()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to initialize ephemeris: %v\n", err)
			os.Exit(1)
		}
		swe.SetEphePath(cacheDir)
		swe.SetSidMode(swe.SIDM_LAHIRI, 0, 0)

		// Build compute function: all phases
		compute := func(name string, year, month, day, hour, minute int, tzOff, lat, lng float64) ([]byte, error) {
			result := computeAll(name, year, month, day, hour, minute, 0, tzOff, lat, lng, cacheDir)
			return result.FullReportJSON()
		}

		aspects := func() ([]byte, error) {
			catalog := dignity.AspectCatalog()
			return dignity.FormatAspectJSON(catalog)
		}

		timing := func(name string, year, month, day, hour, minute int, tzOff, lat, lng float64, targetDate string) ([]byte, error) {
			chartData := computePositions(year, month, day, hour, minute, tzOff, lat, lng, cacheDir)
			report := dignity.ComputeTimingReport(
				name, year, month, day, hour, minute, tzOff, lat, lng,
				targetDate, chartData.planets, chartData.ayan, chartData.asc,
			)
			return report.TimingReportJSON()
		}

		transits := func(name string, year, month, day, hour, minute int, tzOff, lat, lng float64, startDate, endDate string, orbDeg float64, sidereal bool) ([]byte, error) {
			return marshalResult(computeTransits(name, year, month, day, hour, minute, tzOff, lat, lng, startDate, endDate, orbDeg, sidereal, cacheDir))
		}

		synastry := func(name1 string, y1, mo1, d1, h1, mi1 int, tz1, la1, lo1 float64, name2 string, y2, mo2, d2, h2, mi2 int, tz2, la2, lo2 float64, orbDeg float64) ([]byte, error) {
			return marshalResult(computeSynastry(name1, y1, mo1, d1, h1, mi1, tz1, la1, lo1, name2, y2, mo2, d2, h2, mi2, tz2, la2, lo2, orbDeg, cacheDir))
		}

		relocation := func(name string, year, month, day, hour, minute int, tzOff, lat, lng float64, locA server.LatLng, locB server.LatLng, targetDate string) ([]byte, error) {
			return marshalResult(computeRelocation(name, year, month, day, hour, minute, tzOff, lat, lng, locA, locB, targetDate, cacheDir))
		}

		chart := func(name string, year, month, day, hour, minute int, tzOff, lat, lng float64, houseSystem string, sidereal bool, showAspects bool, outerPlanets bool, highlightPatterns bool, patternOrb float64) (string, error) {
			opts := dignity.DefaultChartOptions()
			opts.HouseSystem = houseSystem
			opts.Sidereal = sidereal
			opts.ShowAspects = showAspects
			opts.OuterPlanets = outerPlanets
			opts.HighlightPatterns = highlightPatterns
			opts.PatternOrb = patternOrb
			return dignity.RenderChartSVG(name, year, month, day, hour, minute, tzOff, lat, lng, opts), nil
		}

		patterns := func(name string, year, month, day, hour, minute int, tzOff, lat, lng float64, orbDeg float64) ([]byte, error) {
			cd := computePositions(year, month, day, hour, minute, tzOff, lat, lng, cacheDir)
			// All planets (including outer, asteroids, Lilith) are already in cd.planets
			planetMap := make(map[string]float64)
			for k, v := range cd.planets {
				planetMap[k] = v
			}
			// Add North Node
			planetMap["Node"] = cd.nn
			report := dignity.DetectPatterns(planetMap, orbDeg)
			report.Name = name
			return report.PatternReportJSON()
		}

		draconic := func(name string, year, month, day, hour, minute int, tzOff, lat, lng float64, orbDeg float64) ([]byte, error) {
			cd := computePositions(year, month, day, hour, minute, tzOff, lat, lng, cacheDir)
			return marshalResult(computeDraconic(name, cd, orbDeg))
		}

		draconicSynastry := func(name1 string, y1, mo1, d1, h1, mi1 int, tz1, la1, lo1 float64, name2 string, y2, mo2, d2, h2, mi2 int, tz2, la2, lo2 float64, orbDeg float64) ([]byte, error) {
			cd1 := computePositions(y1, mo1, d1, h1, mi1, tz1, la1, lo1, cacheDir)
			cd2 := computePositions(y2, mo2, d2, h2, mi2, tz2, la2, lo2, cacheDir)
			return marshalResult(computeDraconicSynastry(name1, cd1, name2, cd2, orbDeg))
		}

		draconicSynastryFull := func(name1 string, y1, mo1, d1, h1, mi1 int, tz1, la1, lo1 float64, name2 string, y2, mo2, d2, h2, mi2 int, tz2, la2, lo2 float64, orbDeg float64) ([]byte, error) {
			cd1 := computePositions(y1, mo1, d1, h1, mi1, tz1, la1, lo1, cacheDir)
			cd2 := computePositions(y2, mo2, d2, h2, mi2, tz2, la2, lo2, cacheDir)
			return marshalResult(computeDraconicSynastryFull(name1, cd1, name2, cd2, orbDeg))
		}

		draconicTransits := func(name string, year, month, day, hour, minute int, tzOff, lat, lng float64, startDate, endDate string, orbDeg float64) ([]byte, error) {
			cd := computePositions(year, month, day, hour, minute, tzOff, lat, lng, cacheDir)
			return marshalResult(computeDraconicTransits(name, cd, startDate, endDate, orbDeg, cacheDir))
		}

		progressedDraconic := func(name string, year, month, day, hour, minute int, tzOff, lat, lng float64, targetDate string) ([]byte, error) {
			cd := computePositions(year, month, day, hour, minute, tzOff, lat, lng, cacheDir)
			return marshalResult(computeProgressedDraconic(name, cd, targetDate, cacheDir))
		}

		draconicSolarReturn := func(name string, year, month, day, hour, minute int, tzOff, lat, lng float64, targetYear int) ([]byte, error) {
			cd := computePositions(year, month, day, hour, minute, tzOff, lat, lng, cacheDir)
			return marshalResult(computeDraconicSolarReturn(name, cd, targetYear, cacheDir))
		}

		stars := func(name string, year, month, day, hour, minute int, tzOff, lat, lng float64, orbDeg float64) ([]byte, error) {
			cd := computePositions(year, month, day, hour, minute, tzOff, lat, lng, cacheDir)
			return marshalResult(computeStars(name, cd, orbDeg, cacheDir))
		}

		draconicTransitsCross := func(name string, year, month, day, hour, minute int, tzOff, lat, lng float64, startDate, endDate string, orbDeg float64) ([]byte, error) {
			cd := computePositions(year, month, day, hour, minute, tzOff, lat, lng, cacheDir)
			return marshalResult(computeDraconicTransitsCross(name, cd, startDate, endDate, orbDeg, cacheDir))
		}

		progressedCross := func(name string, year, month, day, hour, minute int, tzOff, lat, lng float64, targetDate string, orbDeg float64) ([]byte, error) {
			cd := computePositions(year, month, day, hour, minute, tzOff, lat, lng, cacheDir)
			return marshalResult(computeProgressedCross(name, cd, targetDate, orbDeg, cacheDir))
		}

		directions := func(name string, year, month, day, hour, minute int, tzOff, lat, lng float64, age float64, orbDeg float64) ([]byte, error) {
			cd := computePositions(year, month, day, hour, minute, tzOff, lat, lng, cacheDir)
			return marshalResult(computeDirections(name, cd, lat, lng, age, orbDeg, cacheDir))
		}

		interpretation := func(name string, year, month, day, hour, minute int, tzOff, lat, lng float64, houseSystem string, orbDeg float64) ([]byte, error) {
			cd := computePositions(year, month, day, hour, minute, tzOff, lat, lng, cacheDir)
			return computeInterpretation(name, cd, lat, lng, houseSystem, orbDeg, cacheDir)
		}

		astroCartography := func(name string, year, month, day, hour, minute int, tzOff, lat, lng float64, latStep float64, frame string) ([]byte, error) {
			cd := computePositions(year, month, day, hour, minute, tzOff, lat, lng, cacheDir)
			return marshalResult(computeAstroCartography(name, cd, latStep, frame, cacheDir))
		}

		astroCartographyCompare := func(name string, year, month, day, hour, minute int, tzOff, lat, lng float64, latStep, targetLat, targetLng, orb float64) ([]byte, error) {
			cd := computePositions(year, month, day, hour, minute, tzOff, lat, lng, cacheDir)
			return marshalResult(computeAstroCartographyCompare(name, cd, latStep, targetLat, targetLng, orb, cacheDir))
		}

		electional := func(name string, year, month, day, hour, minute int, tzOff, lat, lng float64, startDate, endDate string, orbDeg float64) ([]byte, error) {
			cd := computePositions(year, month, day, hour, minute, tzOff, lat, lng, cacheDir)
			return marshalResult(computeElectional(name, cd, lat, lng, startDate, endDate, orbDeg, cacheDir))
		}

		mansionConvergence := func(name string, year, month, day, hour, minute int, tzOff, lat, lng float64) ([]byte, error) {
			cd := computePositions(year, month, day, hour, minute, tzOff, lat, lng, cacheDir)
			return marshalResult(computeMansionConvergence(name, cd, cacheDir))
		}

		arabicParts := func(name string, year, month, day, hour, minute int, tzOff, lat, lng float64, orbDeg float64) ([]byte, error) {
			cd := computePositions(year, month, day, hour, minute, tzOff, lat, lng, cacheDir)
			return marshalResult(computeArabicParts(name, cd, orbDeg, cacheDir))
		}

		solarReturn := func(name string, year, month, day, hour, minute int, tzOff, lat, lng float64, targetYear int) ([]byte, error) {
			cd := computePositions(year, month, day, hour, minute, tzOff, lat, lng, cacheDir)
			return marshalResult(computeTropicalSolarReturn(name, cd, targetYear, cacheDir))
		}

		composite := func(name1 string, y1, m1, d1, h1, min1 int, tz1, lat1, lng1 float64, name2 string, y2, m2, d2, h2, min2 int, tz2, lat2, lng2 float64, orbDeg float64) ([]byte, error) {
			cd1 := computePositions(y1, m1, d1, h1, min1, tz1, lat1, lng1, cacheDir)
			cd2 := computePositions(y2, m2, d2, h2, min2, tz2, lat2, lng2, cacheDir)
			return marshalResult(computeComposite(name1, name2, cd1, cd2, orbDeg, cacheDir))
		}

		starsCross := func(name string, year, month, day, hour, minute int, tzOff, lat, lng float64, orbDeg float64) ([]byte, error) {
			cd := computePositions(year, month, day, hour, minute, tzOff, lat, lng, cacheDir)
			return marshalResult(computeStarsCross(name, cd, orbDeg, cacheDir))
		}

		traditional := func(name string, year, month, day, hour, minute int, tzOff, lat, lng float64) ([]byte, error) {
			cd := computePositions(year, month, day, hour, minute, tzOff, lat, lng, cacheDir)
			return marshalResult(computeTraditional(name, cd))
		}

		uranian := func(name string, year, month, day, hour, minute int, tzOff, lat, lng float64) ([]byte, error) {
			cd := computePositions(year, month, day, hour, minute, tzOff, lat, lng, cacheDir)
			return marshalResult(computeUranian(name, cd))
		}

		harmonic := func(name string, year, month, day, hour, minute int, tzOff, lat, lng float64, harmonics []int, orb float64) ([]byte, error) {
			cd := computePositions(year, month, day, hour, minute, tzOff, lat, lng, cacheDir)
			return marshalResult(computeHarmonic(name, cd, harmonics, orb))
		}

		divisional := func(name string, year, month, day, hour, minute int, tzOff, lat, lng float64) ([]byte, error) {
			cd := computePositions(year, month, day, hour, minute, tzOff, lat, lng, cacheDir)
			return marshalResult(computeDivisional(name, cd, year, month, day))
		}

		parans := func(name string, year, month, day, hour, minute int, tzOff, lat, lng float64, orb float64) ([]byte, error) {
			cd := computePositions(year, month, day, hour, minute, tzOff, lat, lng, cacheDir)
			return marshalResult(computeParans(name, cd, orb, cacheDir))
		}

		declination := func(name string, year, month, day, hour, minute int, tzOff, lat, lng float64, orb float64) ([]byte, error) {
			cd := computePositions(year, month, day, hour, minute, tzOff, lat, lng, cacheDir)
			return marshalResult(computeDeclination(name, cd, orb))
		}

		firdaria := func(name string, year, month, day, hour, minute int, tzOff, lat, lng float64) ([]byte, error) {
			cd := computePositions(year, month, day, hour, minute, tzOff, lat, lng, cacheDir)
			return marshalResult(computeFirdaria(name, cd, year, month, day))
		}

		// Use embedded web files, stripping the "web/" prefix
		staticFS, err := fs.Sub(empirical.WebFiles, "web")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to load web files: %v\n", err)
			os.Exit(1)
		}

		if err := server.Run(port, server.ServerConfig{
			StaticFS:              staticFS,
			Compute:               compute,
			Aspects:               aspects,
			Timing:                timing,
			Transits:              transits,
			Synastry:              synastry,
			Relocation:            relocation,
			Chart:                 chart,
			Patterns:              patterns,
			Draconic:              draconic,
			DraconicSynastry:      draconicSynastry,
			DraconicSynastryFull:  draconicSynastryFull,
			DraconicTransits:      draconicTransits,
			ProgressedDraconic:    progressedDraconic,
			DraconicSolarReturn:   draconicSolarReturn,
			Stars:                 stars,
			DraconicTransitsCross: draconicTransitsCross,
			ProgressedCross:       progressedCross,
			Directions:            directions,
			Interpretation:        interpretation,
			AstroCartography:      astroCartography,
			AstroCartographyCompare: astroCartographyCompare,
			Electional:            electional,
			MansionConvergence:    mansionConvergence,
			ArabicParts:           arabicParts,
			SolarReturn:           solarReturn,
			Composite:             composite,
			StarsCross:            starsCross,
			Traditional:           traditional,
			Uranian:               uranian,
			Harmonic:              harmonic,
			Divisional:            divisional,
			Parans:                parans,
			Declination:           declination,
			Firdaria:              firdaria,
		}); err != nil {
			fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// ── transit subcommand ──────────────────────────────────────────
	if len(os.Args) >= 2 && os.Args[1] == "transit" {
		fs := flag.NewFlagSet("transit", flag.ExitOnError)
		jsonOut := fs.Bool("json", false, "output as JSON")
		orbDeg := fs.Float64("orb", 3.0, "max orb in degrees")
		fs.Parse(os.Args[2:])
		args := fs.Args()

		if len(args) < 11 {
			fmt.Fprintf(os.Stderr, "Usage: empirical transit [--json] [--orb 3] NAME Y M D H MIN TZ LAT LNG START_DATE END_DATE\n")
			fmt.Fprintf(os.Stderr, "Example: empirical transit \"AJ\" 1969 2 15 23 10 -8 47.038 -122.901 2026-06-09 2026-06-23\n")
			os.Exit(1)
		}

		name := args[0]
		year, _ := strconv.Atoi(args[1])
		month, _ := strconv.Atoi(args[2])
		day, _ := strconv.Atoi(args[3])
		hour, _ := strconv.Atoi(args[4])
		minute, _ := strconv.Atoi(args[5])
		tzOff, _ := strconv.ParseFloat(args[6], 64)
		lat, _ := strconv.ParseFloat(args[7], 64)
		lng, _ := strconv.ParseFloat(args[8], 64)
		startDate := args[9]
		endDate := args[10]

		result, err := computeTransits(name, year, month, day, hour, minute, tzOff, lat, lng, startDate, endDate, *orbDeg, false, "")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Transit error: %v\n", err)
			os.Exit(1)
		}
		js, _ := json.Marshal(result)
		if *jsonOut {
			fmt.Println(string(js))
		} else {
			fmt.Print(string(js))
		}
		return
	}

	// ── synastry subcommand ─────────────────────────────────────────
	if len(os.Args) >= 2 && os.Args[1] == "synastry" {
		fs := flag.NewFlagSet("synastry", flag.ExitOnError)
		jsonOut := fs.Bool("json", false, "output as JSON")
		orbDeg := fs.Float64("orb", 5.0, "max orb in degrees")
		fs.Parse(os.Args[2:])
		args := fs.Args()

		if len(args) < 18 {
			fmt.Fprintf(os.Stderr, "Usage: empirical synastry [--json] [--orb 5] NAME1 Y1 M1 D1 H1 MIN1 TZ1 LAT1 LNG1 NAME2 Y2 M2 D2 H2 MIN2 TZ2 LAT2 LNG2\n")
			fmt.Fprintf(os.Stderr, "Example: empirical synastry --orb 5 \"AJ\" 1969 2 15 23 10 -8 47.038 -122.901 \"Cait\" 1986 4 29 3 0 -4 41.034 -73.763\n")
			os.Exit(1)
		}

		name1 := args[0]
		y1, _ := strconv.Atoi(args[1])
		mo1, _ := strconv.Atoi(args[2])
		d1, _ := strconv.Atoi(args[3])
		h1, _ := strconv.Atoi(args[4])
		mi1, _ := strconv.Atoi(args[5])
		tz1, _ := strconv.ParseFloat(args[6], 64)
		la1, _ := strconv.ParseFloat(args[7], 64)
		lo1, _ := strconv.ParseFloat(args[8], 64)
		name2 := args[9]
		y2, _ := strconv.Atoi(args[10])
		mo2, _ := strconv.Atoi(args[11])
		d2, _ := strconv.Atoi(args[12])
		h2, _ := strconv.Atoi(args[13])
		mi2, _ := strconv.Atoi(args[14])
		tz2, _ := strconv.ParseFloat(args[15], 64)
		la2, _ := strconv.ParseFloat(args[16], 64)
		lo2, _ := strconv.ParseFloat(args[17], 64)

		result, err := computeSynastry(name1, y1, mo1, d1, h1, mi1, tz1, la1, lo1, name2, y2, mo2, d2, h2, mi2, tz2, la2, lo2, *orbDeg, "")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Synastry error: %v\n", err)
			os.Exit(1)
		}
		js, _ := json.Marshal(result)
		if *jsonOut {
			fmt.Println(string(js))
		} else {
			fmt.Print(string(js))
		}
		return
	}

	// ── batch subcommand ───────────────────────────────────────────
	if len(os.Args) >= 2 && os.Args[1] == "batch" {
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "Usage: empirical batch <transits|synastry> [args...]\n")
			fmt.Fprintf(os.Stderr, "  transits: --people a,b,c --start YYYY-MM-DD --end YYYY-MM-DD [--orb 3]\n")
			fmt.Fprintf(os.Stderr, "  synastry: --people a,b,c [--orb 5]\n")
			os.Exit(1)
		}

		subCmd := os.Args[2]
		fs := flag.NewFlagSet("batch "+subCmd, flag.ExitOnError)
		peopleStr := fs.String("people", "", "comma-separated names (aj,cait,pete,pat)")
		orbDeg := fs.Float64("orb", 0, "max orb in degrees")
		var startDate, endDate string
		if subCmd == "transits" {
			fs.StringVar(&startDate, "start", "", "start date (YYYY-MM-DD)")
			fs.StringVar(&endDate, "end", "", "end date (YYYY-MM-DD)")
		}
		fs.Parse(os.Args[3:])

		if *peopleStr == "" {
			fmt.Fprintf(os.Stderr, "--people is required\n")
			os.Exit(1)
		}
		peopleNames := splitComma(*peopleStr)
		if len(peopleNames) == 0 {
			fmt.Fprintf(os.Stderr, "no people listed\n")
			os.Exit(1)
		}

		// Resolve birth data
		type birth struct {
			y, mo, d, h, mi int
			tz, la, lo     float64
		}
		known := map[string]birth{
			"aj":   {1969, 2, 15, 23, 10, -8, 47.038, -122.901},
			"cait": {1986, 4, 29, 3, 0, -4, 41.034, -73.763},
			"pete": {1952, 3, 1, 12, 0, -5, 40.9312, -73.8988},
			"pat":  {1950, 12, 31, 12, 0, -5, 40.9312, -73.8988},
		}

		cacheDir := ""
		natalPlanets := dignity.AllPlanetNames
		synastryAspects := []dignity.AspectDef{
			{0, "conjunction"}, {60, "sextile"}, {90, "square"}, {120, "trine"}, {180, "opposition"},
		}

		switch subCmd {
		case "transits":
			if *orbDeg <= 0 { *orbDeg = 3.0 }
			if startDate == "" || endDate == "" {
				fmt.Fprintf(os.Stderr, "--start and --end required for batch transits\n")
				os.Exit(1)
			}

			var people []dignity.BatchPerson
			for _, name := range peopleNames {
				b, ok := known[name]
				if !ok {
					fmt.Fprintf(os.Stderr, "unknown person: %s\n", name)
					os.Exit(1)
				}
				cd := computePositions(b.y, b.mo, b.d, b.h, b.mi, b.tz, b.la, b.lo, cacheDir)
				longs := make(map[string]float64)
				for k, v := range cd.planets { longs[k] = v }
				people = append(people, dignity.BatchPerson{Name: name, PlanetLongs: longs})
			}

			// Use first person's timezone for transit JD computation
			first := known[peopleNames[0]]
			compute := func(year, month, day int, hour float64, planetID int) (float64, float64, float64, float64) {
				utHour := hour - first.tz
				jd := swe.Julday(year, month, day, utHour, true)
				return swe.CalcUT(jd, planetID)
			}

			results := dignity.BatchTransits(people, natalPlanets, startDate, endDate, dignity.HardAspectsOnly(), *orbDeg, compute)

			// Build JSON
			type hitJSON struct {
				Transit string  `json:"transit"`
				Natal   string  `json:"natal"`
				Aspect  string  `json:"aspect"`
				Orb     float64 `json:"orb"`
				Start   string  `json:"start"`
				End     string  `json:"end"`
			}
			type personJSON struct {
				Name  string    `json:"name"`
				Hits  []hitJSON `json:"hits"`
			}
			var out []personJSON
			for _, r := range results {
				compact := dignity.CompactTransitsWithRange(r.Hits)
				pj := personJSON{Name: r.Name}
				for _, c := range compact {
					pj.Hits = append(pj.Hits, hitJSON{
						Transit: c.TransitPlanet, Natal: c.NatalPlanet, Aspect: c.Aspect,
						Orb: c.MinOrb, Start: c.DateStart, End: c.DateEnd,
					})
				}
				out = append(out, pj)
			}
			js, _ := json.Marshal(out)
			fmt.Println(string(js))

		case "synastry":
			if *orbDeg <= 0 { *orbDeg = 5.0 }

			var people []dignity.BatchPerson
			for _, name := range peopleNames {
				b, ok := known[name]
				if !ok {
					fmt.Fprintf(os.Stderr, "unknown person: %s\n", name)
					os.Exit(1)
				}
				cd := computePositions(b.y, b.mo, b.d, b.h, b.mi, b.tz, b.la, b.lo, cacheDir)
				longs := make(map[string]float64)
				for k, v := range cd.planets { longs[k] = v }
				people = append(people, dignity.BatchPerson{Name: name, PlanetLongs: longs})
			}

			results := dignity.BatchSynastry(people, natalPlanets, synastryAspects, *orbDeg)

			type hitJSON struct {
				Planet1 string  `json:"planet1"`
				Planet2 string  `json:"planet2"`
				Aspect  string  `json:"aspect"`
				Orb     float64 `json:"orb"`
			}
			type pairJSON struct {
				Name1 string    `json:"name1"`
				Name2 string    `json:"name2"`
				Hits  []hitJSON `json:"hits"`
			}
			var out []pairJSON
			for _, r := range results {
				pj := pairJSON{Name1: r.Name1, Name2: r.Name2}
				for _, h := range r.Hits {
					pj.Hits = append(pj.Hits, hitJSON{Planet1: h.Planet1, Planet2: h.Planet2, Aspect: h.Aspect, Orb: h.Orb})
				}
				out = append(out, pj)
			}
			js, _ := json.Marshal(out)
			fmt.Println(string(js))

		default:
			fmt.Fprintf(os.Stderr, "unknown batch subcommand: %s (use transits or synastry)\n", subCmd)
			os.Exit(1)
		}
		return
	}

	// ── lunar-mansion subcommand ────────────────────────────────────
	if len(os.Args) >= 2 && os.Args[1] == "lunar-mansion" {
		fs := flag.NewFlagSet("lunar-mansion", flag.ExitOnError)
		jsonOut := fs.Bool("json", false, "output as JSON")
		seed := fs.Int64("seed", 42, "random seed")
		fs.Parse(os.Args[2:])

		report := dignity.ComputeLunarMansionReport(*seed)

		if *jsonOut {
			js, err := report.LunarMansionReportJSON()
			if err != nil {
				fmt.Fprintf(os.Stderr, "JSON error: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(string(js))
		} else {
			fmt.Print(dignity.FormatSharedStars())
			fmt.Println()
			fmt.Print(dignity.FormatNullModelResult(report.NullModelUniform))
			fmt.Println()
			fmt.Print(dignity.FormatNullModelResult(report.NullModelBrightness))
			fmt.Println()
			fmt.Print(dignity.FormatNullModelResult(report.NullModelEcliptic))
			fmt.Println()
			fmt.Print(dignity.FormatEclipticConfoundResult(report.EclipticConfound))
		}
		return
	}

	// ── recover command (default) ───────────────────────────────────
	jsonOut := flag.Bool("json", false, "output as JSON instead of text")
	flag.Parse()

	args := flag.Args()
	if len(args) < 9 {
		fmt.Fprintf(os.Stderr, "Usage: recover [--json] NAME Y M D H MIN TZ_OFFSET LAT LON\n")
		fmt.Fprintf(os.Stderr, "       empirical serve [port]\n")
		fmt.Fprintf(os.Stderr, "Example: recover AJ 1969 2 15 23 10 -8 47.038 -122.901\n")
		os.Exit(1)
	}

	name := args[0]
	year, _ := strconv.Atoi(args[1])
	month, _ := strconv.Atoi(args[2])
	day, _ := strconv.Atoi(args[3])
	hour, _ := strconv.Atoi(args[4])
	minute, _ := strconv.Atoi(args[5])
	tzOff, _ := strconv.ParseFloat(args[6], 64)
	lat, _ := strconv.ParseFloat(args[7], 64)
	lng, _ := strconv.ParseFloat(args[8], 64)

	result := computeAll(name, year, month, day, hour, minute, 0, tzOff, lat, lng, "")
	
	if *jsonOut {
		js, err := result.FullReportJSON()
		if err != nil {
			fmt.Fprintf(os.Stderr, "JSON error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(js))
	} else {
		printReport(result)
	}
}

