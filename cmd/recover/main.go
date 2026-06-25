package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"math"
	"os"
	"strconv"
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
			return computeTransits(name, year, month, day, hour, minute, tzOff, lat, lng, startDate, endDate, orbDeg, sidereal, cacheDir)
		}

		synastry := func(name1 string, y1, mo1, d1, h1, mi1 int, tz1, la1, lo1 float64, name2 string, y2, mo2, d2, h2, mi2 int, tz2, la2, lo2 float64, orbDeg float64) ([]byte, error) {
			return computeSynastry(name1, y1, mo1, d1, h1, mi1, tz1, la1, lo1, name2, y2, mo2, d2, h2, mi2, tz2, la2, lo2, orbDeg, cacheDir)
		}

		relocation := func(name string, year, month, day, hour, minute int, tzOff, lat, lng float64, locA server.LatLng, locB server.LatLng, targetDate string) ([]byte, error) {
			return computeRelocation(name, year, month, day, hour, minute, tzOff, lat, lng, locA, locB, targetDate, cacheDir)
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
			return computeDraconic(name, cd, orbDeg)
		}

		draconicSynastry := func(name1 string, y1, mo1, d1, h1, mi1 int, tz1, la1, lo1 float64, name2 string, y2, mo2, d2, h2, mi2 int, tz2, la2, lo2 float64, orbDeg float64) ([]byte, error) {
			cd1 := computePositions(y1, mo1, d1, h1, mi1, tz1, la1, lo1, cacheDir)
			cd2 := computePositions(y2, mo2, d2, h2, mi2, tz2, la2, lo2, cacheDir)
			return computeDraconicSynastry(name1, cd1, name2, cd2, orbDeg)
		}

		draconicSynastryFull := func(name1 string, y1, mo1, d1, h1, mi1 int, tz1, la1, lo1 float64, name2 string, y2, mo2, d2, h2, mi2 int, tz2, la2, lo2 float64, orbDeg float64) ([]byte, error) {
			cd1 := computePositions(y1, mo1, d1, h1, mi1, tz1, la1, lo1, cacheDir)
			cd2 := computePositions(y2, mo2, d2, h2, mi2, tz2, la2, lo2, cacheDir)
			return computeDraconicSynastryFull(name1, cd1, name2, cd2, orbDeg)
		}

		draconicTransits := func(name string, year, month, day, hour, minute int, tzOff, lat, lng float64, startDate, endDate string, orbDeg float64) ([]byte, error) {
			cd := computePositions(year, month, day, hour, minute, tzOff, lat, lng, cacheDir)
			return computeDraconicTransits(name, cd, startDate, endDate, orbDeg, cacheDir)
		}

		progressedDraconic := func(name string, year, month, day, hour, minute int, tzOff, lat, lng float64, targetDate string) ([]byte, error) {
			cd := computePositions(year, month, day, hour, minute, tzOff, lat, lng, cacheDir)
			return computeProgressedDraconic(name, cd, targetDate, cacheDir)
		}

		draconicSolarReturn := func(name string, year, month, day, hour, minute int, tzOff, lat, lng float64, targetYear int) ([]byte, error) {
			cd := computePositions(year, month, day, hour, minute, tzOff, lat, lng, cacheDir)
			return computeDraconicSolarReturn(name, cd, targetYear, cacheDir)
		}

		stars := func(name string, year, month, day, hour, minute int, tzOff, lat, lng float64, orbDeg float64) ([]byte, error) {
			cd := computePositions(year, month, day, hour, minute, tzOff, lat, lng, cacheDir)
			return computeStars(name, cd, orbDeg, cacheDir)
		}

		draconicTransitsCross := func(name string, year, month, day, hour, minute int, tzOff, lat, lng float64, startDate, endDate string, orbDeg float64) ([]byte, error) {
			cd := computePositions(year, month, day, hour, minute, tzOff, lat, lng, cacheDir)
			return computeDraconicTransitsCross(name, cd, startDate, endDate, orbDeg, cacheDir)
		}

		progressedCross := func(name string, year, month, day, hour, minute int, tzOff, lat, lng float64, targetDate string, orbDeg float64) ([]byte, error) {
			cd := computePositions(year, month, day, hour, minute, tzOff, lat, lng, cacheDir)
			return computeProgressedCross(name, cd, targetDate, orbDeg, cacheDir)
		}

		directions := func(name string, year, month, day, hour, minute int, tzOff, lat, lng float64, age float64, orbDeg float64) ([]byte, error) {
			cd := computePositions(year, month, day, hour, minute, tzOff, lat, lng, cacheDir)
			return computeDirections(name, cd, lat, lng, age, orbDeg, cacheDir)
		}

		interpretation := func(name string, year, month, day, hour, minute int, tzOff, lat, lng float64, houseSystem string, orbDeg float64) ([]byte, error) {
			cd := computePositions(year, month, day, hour, minute, tzOff, lat, lng, cacheDir)
			return computeInterpretation(name, cd, lat, lng, houseSystem, orbDeg, cacheDir)
		}

		astroCartography := func(name string, year, month, day, hour, minute int, tzOff, lat, lng float64, latStep float64, frame string) ([]byte, error) {
			cd := computePositions(year, month, day, hour, minute, tzOff, lat, lng, cacheDir)
			return computeAstroCartography(name, cd, latStep, frame, cacheDir)
		}

		astroCartographyCompare := func(name string, year, month, day, hour, minute int, tzOff, lat, lng float64, latStep, targetLat, targetLng, orb float64) ([]byte, error) {
			cd := computePositions(year, month, day, hour, minute, tzOff, lat, lng, cacheDir)
			return computeAstroCartographyCompare(name, cd, latStep, targetLat, targetLng, orb, cacheDir)
		}

		electional := func(name string, year, month, day, hour, minute int, tzOff, lat, lng float64, startDate, endDate string, orbDeg float64) ([]byte, error) {
			cd := computePositions(year, month, day, hour, minute, tzOff, lat, lng, cacheDir)
			return computeElectional(name, cd, lat, lng, startDate, endDate, orbDeg, cacheDir)
		}

		mansionConvergence := func(name string, year, month, day, hour, minute int, tzOff, lat, lng float64) ([]byte, error) {
			cd := computePositions(year, month, day, hour, minute, tzOff, lat, lng, cacheDir)
			return computeMansionConvergence(name, cd, cacheDir)
		}

		arabicParts := func(name string, year, month, day, hour, minute int, tzOff, lat, lng float64, orbDeg float64) ([]byte, error) {
			cd := computePositions(year, month, day, hour, minute, tzOff, lat, lng, cacheDir)
			return computeArabicParts(name, cd, orbDeg, cacheDir)
		}

		solarReturn := func(name string, year, month, day, hour, minute int, tzOff, lat, lng float64, targetYear int) ([]byte, error) {
			cd := computePositions(year, month, day, hour, minute, tzOff, lat, lng, cacheDir)
			return computeTropicalSolarReturn(name, cd, targetYear, cacheDir)
		}

		composite := func(name1 string, y1, m1, d1, h1, min1 int, tz1, lat1, lng1 float64, name2 string, y2, m2, d2, h2, min2 int, tz2, lat2, lng2 float64, orbDeg float64) ([]byte, error) {
			cd1 := computePositions(y1, m1, d1, h1, min1, tz1, lat1, lng1, cacheDir)
			cd2 := computePositions(y2, m2, d2, h2, min2, tz2, lat2, lng2, cacheDir)
			return computeComposite(name1, name2, cd1, cd2, orbDeg, cacheDir)
		}

		starsCross := func(name string, year, month, day, hour, minute int, tzOff, lat, lng float64, orbDeg float64) ([]byte, error) {
			cd := computePositions(year, month, day, hour, minute, tzOff, lat, lng, cacheDir)
			return computeStarsCross(name, cd, orbDeg, cacheDir)
		}

		traditional := func(name string, year, month, day, hour, minute int, tzOff, lat, lng float64) ([]byte, error) {
			cd := computePositions(year, month, day, hour, minute, tzOff, lat, lng, cacheDir)
			return computeTraditional(name, cd)
		}

		uranian := func(name string, year, month, day, hour, minute int, tzOff, lat, lng float64) ([]byte, error) {
			cd := computePositions(year, month, day, hour, minute, tzOff, lat, lng, cacheDir)
			return computeUranian(name, cd)
		}

		harmonic := func(name string, year, month, day, hour, minute int, tzOff, lat, lng float64, harmonics []int, orb float64) ([]byte, error) {
			cd := computePositions(year, month, day, hour, minute, tzOff, lat, lng, cacheDir)
			return computeHarmonic(name, cd, harmonics, orb)
		}

		divisional := func(name string, year, month, day, hour, minute int, tzOff, lat, lng float64) ([]byte, error) {
			cd := computePositions(year, month, day, hour, minute, tzOff, lat, lng, cacheDir)
			return computeDivisional(name, cd, year, month, day)
		}

		parans := func(name string, year, month, day, hour, minute int, tzOff, lat, lng float64, orb float64) ([]byte, error) {
			cd := computePositions(year, month, day, hour, minute, tzOff, lat, lng, cacheDir)
			return computeParans(name, cd, orb, cacheDir)
		}

		declination := func(name string, year, month, day, hour, minute int, tzOff, lat, lng float64, orb float64) ([]byte, error) {
			cd := computePositions(year, month, day, hour, minute, tzOff, lat, lng, cacheDir)
			return computeDeclination(name, cd, orb)
		}

		firdaria := func(name string, year, month, day, hour, minute int, tzOff, lat, lng float64) ([]byte, error) {
			cd := computePositions(year, month, day, hour, minute, tzOff, lat, lng, cacheDir)
			return computeFirdaria(name, cd, year, month, day)
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
		if *jsonOut {
			fmt.Println(string(result))
		} else {
			fmt.Print(string(result))
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
		if *jsonOut {
			fmt.Println(string(result))
		} else {
			fmt.Print(string(result))
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

// chartData holds pre-computed chart positions.
type chartData struct {
	planets map[string]float64
	speeds  map[string]float64
	ayan    float64
	asc     float64
	nn      float64
	jd      float64
}

// computePositions calculates planet longitudes, ayanamsa, ASC, and NN.
func computePositions(year, month, day, hour, minute int, tzOff, lat, lng float64, cacheDir string) *chartData {
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

	utHour := float64(hour) + float64(minute)/60.0 - tzOff
	jd := swe.Julday(year, month, day, utHour, true)
	ayan := swe.GetAyanamsaUT(jd)

	pls := map[string]float64{}
	spds := map[string]float64{}
	specs := dignity.AllPlanets
	for _, p := range specs {
		lon, speed, _, _ := swe.CalcUT(jd, p.ID)
		pls[p.Name] = lon
		spds[p.Name] = speed
	}

	// Get ASC and NN
	nnLon, _, _, _ := swe.CalcUT(jd, swe.MEAN_NODE)
	_, ascmc := swe.Houses(jd, lat, lng, 'P')
	asc := ascmc[0]

	return &chartData{
		planets: pls,
		speeds:  spds,
		ayan:    ayan,
		asc:     asc,
		nn:      nnLon,
		jd:      jd,
	}
}

// computeAll returns a full multi-phase report.
func computeAll(name string, year, month, day, hour, minute, second int, tzOff, lat, lng float64, cacheDir string) *dignity.FullReport {
	cd := computePositions(year, month, day, hour, minute, tzOff, lat, lng, cacheDir)
	return dignity.ComputeFullReport(cd.planets, cd.ayan, cd.nn, cd.asc, name, year, month, day, hour, minute, second, tzOff, lat, lng)
}

// computeTransits runs the transit engine and returns compact JSON results.
// When sidereal is true, uses sidereal (Lahiri) positions for both natal and transiting planets.
func computeTransits(name string, year, month, day, hour, minute int, tzOff, lat, lng float64, startDate, endDate string, orbDeg float64, sidereal bool, cacheDir string) ([]byte, error) {
	cd := computePositions(year, month, day, hour, minute, tzOff, lat, lng, cacheDir)

	// Build planet positions — all bodies already in cd.planets
	natalLongs := make(map[string]float64)
	for k, v := range cd.planets {
		natalLongs[k] = v
	}

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
			v -= cd.ayan
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
	type hitJSON struct {
		TransitPlanet string  `json:"transit_planet"`
		NatalPlanet   string  `json:"natal_planet"`
		Aspect        string  `json:"aspect"`
		Orb           float64 `json:"orb"`
		StartDate     string  `json:"start_date"`
		EndDate       string  `json:"end_date"`
	}
	response := struct {
		Name       string    `json:"name"`
		Sidereal   bool      `json:"sidereal"`
		Transits   []hitJSON `json:"transits"`
		SkyWeather []hitJSON `json:"sky_weather"`
	}{
		Name:     name,
		Sidereal: sidereal,
	}
	for _, c := range compact {
		response.Transits = append(response.Transits, hitJSON{
			TransitPlanet: c.TransitPlanet,
			NatalPlanet:   c.NatalPlanet,
			Aspect:        c.Aspect,
			Orb:           c.MinOrb,
			StartDate:     c.DateStart,
			EndDate:       c.DateEnd,
		})
	}
	for _, c := range ttCompact {
		response.SkyWeather = append(response.SkyWeather, hitJSON{
			TransitPlanet: c.TransitPlanet,
			NatalPlanet:   c.NatalPlanet,
			Aspect:        c.Aspect,
			Orb:           c.MinOrb,
			StartDate:     c.DateStart,
			EndDate:       c.DateEnd,
		})
	}

	return json.Marshal(response)
}

// computeSynastry computes inter-aspects between two natal charts.
func computeSynastry(name1 string, y1, mo1, d1, h1, mi1 int, tz1, la1, lo1 float64, name2 string, y2, mo2, d2, h2, mi2 int, tz2, la2, lo2 float64, orbDeg float64, cacheDir string) ([]byte, error) {
	cd1 := computePositions(y1, mo1, d1, h1, mi1, tz1, la1, lo1, cacheDir)
	cd2 := computePositions(y2, mo2, d2, h2, mi2, tz2, la2, lo2, cacheDir)

	// Build planet maps — all bodies already in cd.planets
	chart1 := make(map[string]float64)
	for k, v := range cd1.planets {
		chart1[k] = v
	}
	chart2 := make(map[string]float64)
	for k, v := range cd2.planets {
		chart2[k] = v
	}

	planets := dignity.AllPlanetNames
	aspects := []dignity.AspectDef{
		{0, "conjunction"}, {60, "sextile"}, {90, "square"}, {120, "trine"}, {180, "opposition"},
	}

	hits := dignity.ComputeSynastry(chart1, chart2, planets, aspects, orbDeg)

	response := struct {
		Name1 string             `json:"name1"`
		Name2 string             `json:"name2"`
		Aspects []dignity.SynastryHit `json:"aspects"`
	}{
		Name1: name1,
		Name2: name2,
		Aspects: hits,
	}

	return json.Marshal(response)
}

// computeRelocation compares two locations for a single person using cross-validated
// house convergence and timing convergence. Returns which house shifts are reliable
// (unanimous at both locations) vs system-dependent.
func computeRelocation(name string, year, month, day, hour, minute int, tzOff, lat, lng float64, locA server.LatLng, locB server.LatLng, targetDate string, cacheDir string) ([]byte, error) {
	// Compute positions at both locations (natal planets are location-invariant, houses differ)
	cdA := computePositions(year, month, day, hour, minute, tzOff, locA.Lat, locA.Lng, cacheDir)
	cdB := computePositions(year, month, day, hour, minute, tzOff, locB.Lat, locB.Lng, cacheDir)

	// Phase 3: House convergence at both locations
	hcA := dignity.ComputeHouseConvergence(cdA.planets, year, month, day, hour, minute, 0, tzOff, locA.Lat, locA.Lng, name)
	hcB := dignity.ComputeHouseConvergence(cdB.planets, year, month, day, hour, minute, 0, tzOff, locB.Lat, locB.Lng, name)

	// Phase 4: Timing convergence (location-invariant, but compute for completeness)
	trA := dignity.ComputeTimingReport(name, year, month, day, hour, minute, tzOff, locA.Lat, locA.Lng, targetDate, cdA.planets, cdA.ayan, cdA.asc)
	trB := dignity.ComputeTimingReport(name, year, month, day, hour, minute, tzOff, locB.Lat, locB.Lng, targetDate, cdB.planets, cdB.ayan, cdB.asc)

	// Build ASC comparison across 5 systems
	type ascEntry struct {
		System string `json:"system"`
		Sign   string `json:"sign"`
		Degree float64 `json:"degree"`
	}
	var ascA, ascB []ascEntry
	for _, sys := range dignity.CompareHouseSystems {
		code, ok := swephCode[sys]
		if !ok {
			continue
		}
		_, ascmcA := swe.Houses(cdA.jd, locA.Lat, locA.Lng, code)
		_, ascmcB := swe.Houses(cdB.jd, locB.Lat, locB.Lng, code)
		ascA = append(ascA, ascEntry{System: sys, Sign: dignity.SignForLongitude(ascmcA[0]), Degree: ascmcA[0]})
		ascB = append(ascB, ascEntry{System: sys, Sign: dignity.SignForLongitude(ascmcB[0]), Degree: ascmcB[0]})
	}

	// Build house shift comparison
	type shiftEntry struct {
		Planet       string `json:"planet"`
		TropicalSign string `json:"tropical_sign"`
		HouseA       int    `json:"house_a"`
		HouseB       int    `json:"house_b"`
		AgreementA   int    `json:"agreement_a"` // how many systems agree at location A
		AgreementB   int    `json:"agreement_b"` // how many systems agree at location B
		StableA      bool   `json:"stable_a"`    // >=4/5 agree at A
		StableB      bool   `json:"stable_b"`    // >=4/5 agree at B
		ShiftReliable bool  `json:"shift_reliable"` // stable at BOTH locations
	}

	var shifts []shiftEntry
	for _, pA := range hcA.Planets {
		// Find matching planet in hcB
		for _, pB := range hcB.Planets {
			if pA.Planet == pB.Planet {
				shifts = append(shifts, shiftEntry{
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
	response := struct {
		Name              string       `json:"name"`
		LocationA         server.LatLng `json:"location_a"`
		LocationB         server.LatLng `json:"location_b"`
		TargetDate        string       `json:"target_date"`
		HouseConvergenceA struct {
			UnambiguousCount int     `json:"unambiguous_count"`
			ConvergenceRate  float64 `json:"convergence_rate"`
		} `json:"house_convergence_a"`
		HouseConvergenceB struct {
			UnambiguousCount int     `json:"unambiguous_count"`
			ConvergenceRate  float64 `json:"convergence_rate"`
		} `json:"house_convergence_b"`
		ASCA []ascEntry `json:"asc_a"`
		ASCB []ascEntry `json:"asc_b"`
		TimingConvergenceA struct {
			HasConvergence bool     `json:"has_convergence"`
			Planets        []string `json:"planets"`
		} `json:"timing_convergence_a"`
		TimingConvergenceB struct {
			HasConvergence bool     `json:"has_convergence"`
			Planets        []string `json:"planets"`
		} `json:"timing_convergence_b"`
		Shifts []shiftEntry `json:"shifts"`
	}{
		Name:       name,
		LocationA:  locA,
		LocationB:  locB,
		TargetDate: targetDate,
		HouseConvergenceA: struct {
			UnambiguousCount int     `json:"unambiguous_count"`
			ConvergenceRate  float64 `json:"convergence_rate"`
		}{hcA.UnambiguousCount(), hcA.ConvergenceRate()},
		HouseConvergenceB: struct {
			UnambiguousCount int     `json:"unambiguous_count"`
			ConvergenceRate  float64 `json:"convergence_rate"`
		}{hcB.UnambiguousCount(), hcB.ConvergenceRate()},
		ASCA: ascA,
		ASCB: ascB,
		TimingConvergenceA: struct {
			HasConvergence bool     `json:"has_convergence"`
			Planets        []string `json:"planets"`
		}{trA.TimingConvergence.HasConvergence, trA.TimingConvergence.PlanetConvergences},
		TimingConvergenceB: struct {
			HasConvergence bool     `json:"has_convergence"`
			Planets        []string `json:"planets"`
		}{trB.TimingConvergence.HasConvergence, trB.TimingConvergence.PlanetConvergences},
		Shifts: shifts,
	}

	return json.Marshal(response)
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
func computeDraconic(name string, cd *chartData, orbDeg float64) ([]byte, error) {
	// Build tropical planet map from all computed positions
	tropical := make(map[string]float64)
	for k, v := range cd.planets {
		tropical[k] = v
	}

	// Compute draconic chart
	drac := dignity.ComputeDraconic(tropical, cd.nn)

	// Compute sign shifts
	shifts := dignity.ComputeDraconicSignShifts(tropical, cd.nn)

	// Compute bridges (all planets except TNPs)
	allPlanets := dignity.NonTNPNoNodePlanetNames
	bridges := dignity.ComputeDraconicBridges(tropical, cd.nn, allPlanets, dignity.DefaultAspects(), orbDeg)

	// Build shift list
	type shiftJSON struct {
		Planet   string `json:"planet"`
		TropSign string `json:"tropical_sign"`
		DracSign string `json:"draconic_sign"`
	}
	var shiftList []shiftJSON
	for _, s := range shifts {
		shiftList = append(shiftList, shiftJSON{s.Planet, s.TropSign, s.DracSign})
	}

	response := struct {
		Name    string              `json:"name"`
		Offset  float64             `json:"offset"`
		Planets map[string]float64  `json:"planets"`
		Shifts  []shiftJSON         `json:"sign_shifts"`
		Bridges []dignity.SynastryHit `json:"bridges"`
	}{
		Name:    name,
		Offset:  drac.Offset,
		Planets: drac.Planets,
		Shifts:  shiftList,
		Bridges: bridges,
	}

	return json.Marshal(response)
}

// computeDraconicSynastry builds the draconic synastry JSON response.
func computeDraconicSynastry(name1 string, cd1 *chartData, name2 string, cd2 *chartData, orbDeg float64) ([]byte, error) {
	tropical1 := make(map[string]float64)
	for k, v := range cd1.planets {
		tropical1[k] = v
	}
	tropical2 := make(map[string]float64)
	for k, v := range cd2.planets {
		tropical2[k] = v
	}

	allPlanets := dignity.NonTNPNoNodePlanetNames
	hits := dignity.ComputeDraconicSynastry(tropical1, cd1.nn, tropical2, cd2.nn, allPlanets, dignity.DefaultAspects(), orbDeg)

	response := struct {
		Name1 string              `json:"name1"`
		Name2 string              `json:"name2"`
		Hits  []dignity.SynastryHit `json:"hits"`
	}{
		Name1: name1,
		Name2: name2,
		Hits:  hits,
	}

	return json.Marshal(response)
}

// computeDraconicSynastryFull builds the full three-layer draconic synastry JSON.
func computeDraconicSynastryFull(name1 string, cd1 *chartData, name2 string, cd2 *chartData, orbDeg float64) ([]byte, error) {
	tropical1 := make(map[string]float64)
	for k, v := range cd1.planets {
		tropical1[k] = v
	}
	tropical2 := make(map[string]float64)
	for k, v := range cd2.planets {
		tropical2[k] = v
	}

	allPlanets := dignity.NonTNPNoNodePlanetNames
	result := dignity.ComputeDraconicSynastryFull(tropical1, cd1.nn, tropical2, cd2.nn, allPlanets, dignity.DefaultAspects(), orbDeg)

	response := struct {
		Name1        string              `json:"name1"`
		Name2        string              `json:"name2"`
		DracToDrac   []dignity.SynastryHit `json:"drac_to_drac"`
		TropAToDracB []dignity.SynastryHit `json:"trop_a_to_drac_b"`
		TropBToDracA []dignity.SynastryHit `json:"trop_b_to_drac_a"`
	}{
		Name1:        name1,
		Name2:        name2,
		DracToDrac:   result.DracToDrac,
		TropAToDracB: result.TropAToDracB,
		TropBToDracA: result.TropBToDracA,
	}

	return json.Marshal(response)
}

// computeDraconicTransits computes transiting planets hitting the draconic chart.
func computeDraconicTransits(name string, cd *chartData, startDate, endDate string, orbDeg float64, cacheDir string) ([]byte, error) {
	// Build tropical planet map
	tropical := make(map[string]float64)
	for k, v := range cd.planets {
		tropical[k] = v
	}

	// Compute draconic chart (soul-level natal positions)
	drac := dignity.ComputeDraconic(tropical, cd.nn)

	// Build compute function for transiting positions
	tzOff := 0.0 // not stored in chartData; use 0 (positions are UT-based)
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

	type hitJSON struct {
		TransitPlanet string  `json:"transit_planet"`
		NatalPlanet   string  `json:"natal_planet"`
		Aspect        string  `json:"aspect"`
		Orb           float64 `json:"orb"`
		StartDate     string  `json:"start_date"`
		EndDate       string  `json:"end_date"`
	}
	response := struct {
		Name     string    `json:"name"`
		Offset   float64   `json:"offset"`
		Transits []hitJSON `json:"transits"`
	}{
		Name:   name,
		Offset: drac.Offset,
	}
	for _, c := range compact {
		response.Transits = append(response.Transits, hitJSON{
			TransitPlanet: c.TransitPlanet,
			NatalPlanet:   c.NatalPlanet,
			Aspect:        c.Aspect,
			Orb:           c.MinOrb,
			StartDate:     c.DateStart,
			EndDate:       c.DateEnd,
		})
	}

	return json.Marshal(response)
}

// computeProgressedDraconic computes the progressed draconic chart using the
// current transiting North Node as the zero-point.
func computeProgressedDraconic(name string, cd *chartData, targetDate string, cacheDir string) ([]byte, error) {
	// Parse target date
	var y, m, d int
	fmt.Sscanf(targetDate, "%d-%d-%d", &y, &m, &d)

	// Compute current transiting NN
	utHour := 12.0 // noon UT
	jd := swe.Julday(y, m, d, utHour, true)
	lon, _, _, _ := swe.CalcUT(jd, swe.MEAN_NODE)
	currentNN := lon

	// Build tropical planet map
	tropical := make(map[string]float64)
	for k, v := range cd.planets {
		tropical[k] = v
	}

	// Compute both draconic charts
	natalDrac := dignity.ComputeDraconic(tropical, cd.nn)
	progDrac := dignity.ComputeProgressedDraconic(tropical, currentNN)

	// Compute sign shifts between classic and progressed draconic
	shifts := dignity.ComputeDraconicSignShifts(tropical, currentNN)

	// Format datetime
	yr, mo, dy, hr := swe.Revjul(jd)
	dtStr := fmt.Sprintf("%d-%02d-%02d %02d:%02d UT", yr, mo, dy, int(hr), int((hr-float64(int(hr)))*60))

	type shiftJSON struct {
		Planet   string `json:"planet"`
		TropSign string `json:"tropical_sign"`
		DracSign string `json:"draconic_sign"`
	}
	var shiftList []shiftJSON
	for _, s := range shifts {
		shiftList = append(shiftList, shiftJSON{s.Planet, s.TropSign, s.DracSign})
	}

	response := struct {
		Name          string             `json:"name"`
		Date          string             `json:"date"`
		NatalNN       float64            `json:"natal_nn"`
		CurrentNN     float64            `json:"current_nn"`
		NNShift       float64            `json:"nn_shift"`
		NatalDraconic map[string]float64 `json:"natal_draconic"`
		ProgDraconic  map[string]float64 `json:"progressed_draconic"`
		SignShifts    []shiftJSON        `json:"sign_shifts"`
	}{
		Name:          name,
		Date:          dtStr,
		NatalNN:       cd.nn,
		CurrentNN:     currentNN,
		NNShift:       currentNN - cd.nn,
		NatalDraconic: natalDrac.Planets,
		ProgDraconic:  progDrac.Planets,
		SignShifts:    shiftList,
	}

	return json.Marshal(response)
}

// computeDraconicSolarReturn computes the draconic solar return for a target year.
func computeDraconicSolarReturn(name string, cd *chartData, targetYear int, cacheDir string) ([]byte, error) {
	// Get natal Sun longitude
	natalSun := cd.planets["Sun"]

	// Find exact solar return moment
	jdSR := findSolarReturnJD(natalSun, targetYear, cd.jd)

	// Calculate positions at solar return
	planetIDs := dignity.BasicPlanets

	tropical := make(map[string]float64)
	for _, p := range planetIDs {
		lon, _, _, _ := swe.CalcUT(jdSR, p.ID)
		tropical[p.Name] = normalizeLon(lon)
	}

	// Calculate ASC and MC at solar return
	_, ascmc := swe.Houses(jdSR, 0, 0, 'P') // lat/lng not stored in chartData; use 0,0
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
		draconicByNatal[name] = normalizeLon(lon - cd.nn)
	}

	// Format datetime
	yr, mo, dy, hr := swe.Revjul(jdSR)
	dtStr := fmt.Sprintf("%d-%02d-%02d %02d:%02d UT", yr, mo, dy, int(hr), int((hr-float64(int(hr)))*60))

	response := struct {
		Name             string             `json:"name"`
		TargetYear       int                `json:"target_year"`
		JD               float64            `json:"jd"`
		DateTime         string             `json:"datetime"`
		Tropical         map[string]float64 `json:"tropical"`
		Draconic         map[string]float64 `json:"draconic"`
		DraconicByNatalNN map[string]float64 `json:"draconic_by_natal_nn"`
	}{
		Name:             name,
		TargetYear:       targetYear,
		JD:               jdSR,
		DateTime:         dtStr,
		Tropical:         tropical,
		Draconic:         draconic,
		DraconicByNatalNN: draconicByNatal,
	}

	return json.Marshal(response)
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
func computeDraconicTransitsCross(name string, cd *chartData, startDate, endDate string, orbDeg float64, cacheDir string) ([]byte, error) {
	// Build tropical planet map
	tropical := make(map[string]float64)
	for k, v := range cd.planets {
		tropical[k] = v
	}

	// Compute draconic chart (zodiac-invariant)
	drac := dignity.ComputeDraconic(tropical, cd.nn)

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
	type hitJSON struct {
		TransitPlanet string  `json:"transit_planet"`
		NatalPlanet   string  `json:"natal_planet"`
		Aspect        string  `json:"aspect"`
		Orb           float64 `json:"orb"`
	}
	response := struct {
		Name          string    `json:"name"`
		Offset        float64   `json:"offset"`
		Ayanamsa      float64   `json:"ayanamsa"`
		Orb           float64   `json:"orb"`
		MidDate       string    `json:"mid_date"`
		Survivors     []hitJSON `json:"survivors"`
		TropicalOnly  []hitJSON `json:"tropical_only"`
		SiderealOnly  []hitJSON `json:"sidereal_only"`
	}{
		Name:         name,
		Offset:       drac.Offset,
		Ayanamsa:     ayan,
		Orb:          orbDeg,
		Survivors:    make([]hitJSON, 0),
		TropicalOnly: make([]hitJSON, 0),
		SiderealOnly: make([]hitJSON, 0),
	}
	yr, mo, dy, hr := swe.Revjul(midJD)
	response.MidDate = fmt.Sprintf("%d-%02d-%02d %02d:%02d UT", yr, mo, dy, int(hr), int((hr-float64(int(hr)))*60))

	for _, h := range result.Survivors {
		response.Survivors = append(response.Survivors, hitJSON{h.TransitPlanet, h.NatalPlanet, h.Aspect, h.Orb})
	}
	for _, h := range result.TropicalOnly {
		response.TropicalOnly = append(response.TropicalOnly, hitJSON{h.TransitPlanet, h.NatalPlanet, h.Aspect, h.Orb})
	}
	for _, h := range result.SiderealOnly {
		response.SiderealOnly = append(response.SiderealOnly, hitJSON{h.TransitPlanet, h.NatalPlanet, h.Aspect, h.Orb})
	}

	return json.Marshal(response)
}

// computeProgressedCross compares progressed-to-natal aspects in tropical vs sidereal.
// Both natal and progressed positions shift by the same ayanamsa, so angular
// distances are preserved. Near-100% survival expected (Phase 13).
func computeProgressedCross(name string, cd *chartData, targetDate string, orbDeg float64, cacheDir string) ([]byte, error) {
	// Parse target date
	var y, m, d int
	fmt.Sscanf(targetDate, "%d-%d-%d", &y, &m, &d)
	utHour := 12.0
	targetJD := swe.Julday(y, m, d, utHour, true)

	// Age in years
	age := (targetJD - cd.jd) / 365.2425

	// Progressed JD: birthJD + age (day-for-a-year)
	progJD := cd.jd + age

	// Natal positions (tropical)
	natal := make(map[string]float64)
	for k, v := range cd.planets {
		natal[k] = v
	}

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
	ayan := swe.GetAyanamsaUT(cd.jd)

	aspects := dignity.DefaultAspects()
	result := dignity.CompareCrossSystemProgressed(natal, prog, ayan, aspects, orbDeg)

	// Build JSON response
	type hitJSON struct {
		ProgressedPlanet string  `json:"progressed_planet"`
		NatalPlanet      string  `json:"natal_planet"`
		Aspect           string  `json:"aspect"`
		Orb              float64 `json:"orb"`
	}
	response := struct {
		Name          string    `json:"name"`
		TargetDate    string    `json:"target_date"`
		Age           float64   `json:"age_years"`
		Ayanamsa      float64   `json:"ayanamsa"`
		Orb           float64   `json:"orb"`
		Survivors     []hitJSON `json:"survivors"`
		TropicalOnly  []hitJSON `json:"tropical_only"`
		SiderealOnly  []hitJSON `json:"sidereal_only"`
	}{
		Name:         name,
		TargetDate:   targetDate,
		Age:          math.Round(age*100) / 100,
		Ayanamsa:     ayan,
		Orb:          orbDeg,
		Survivors:    make([]hitJSON, 0),
		TropicalOnly: make([]hitJSON, 0),
		SiderealOnly: make([]hitJSON, 0),
	}

	for _, h := range result.Survivors {
		response.Survivors = append(response.Survivors, hitJSON{h.ProgressedPlanet, h.NatalPlanet, h.Aspect, h.Orb})
	}
	for _, h := range result.TropicalOnly {
		response.TropicalOnly = append(response.TropicalOnly, hitJSON{h.ProgressedPlanet, h.NatalPlanet, h.Aspect, h.Orb})
	}
	for _, h := range result.SiderealOnly {
		response.SiderealOnly = append(response.SiderealOnly, hitJSON{h.ProgressedPlanet, h.NatalPlanet, h.Aspect, h.Orb})
	}

	return json.Marshal(response)
}

// computeDirections computes primary directions (Ptolemy) for a given age.
// Directs ASC by oblique ascension and MC by right ascension.
func computeDirections(name string, cd *chartData, lat, lng, age float64, orbDeg float64, cacheDir string) ([]byte, error) {
	// Natal positions
	natal := make(map[string]float64)
	for k, v := range cd.planets {
		natal[k] = v
	}

	// ASC from chart data, MC computed from JD + lat/lng
	ascLon := cd.asc
	_, ascmc := swe.Houses(cd.jd, lat, lng, 'P')
	mcLon := ascmc[1]

	aspects := dignity.DefaultAspects()
	result := dignity.ComputePrimaryDirections(natal, ascLon, mcLon, lat, age, aspects, orbDeg)

	// Build JSON response
	type hitJSON struct {
		DirectedPoint string  `json:"directed_point"`
		NatalPlanet   string  `json:"natal_planet"`
		Aspect        string  `json:"aspect"`
		Orb           float64 `json:"orb"`
	}
	response := struct {
		Name         string    `json:"name"`
		Age          float64   `json:"age_years"`
		DirectedASC  float64   `json:"directed_asc"`
		DirectedMC   float64   `json:"directed_mc"`
		Orb          float64   `json:"orb"`
		ASCAspects   []hitJSON `json:"asc_aspects"`
		MCAspects    []hitJSON `json:"mc_aspects"`
	}{
		Name:        name,
		Age:         age,
		DirectedASC: result.DirectedASC,
		DirectedMC:  result.DirectedMC,
		Orb:         orbDeg,
		ASCAspects:  make([]hitJSON, 0),
		MCAspects:   make([]hitJSON, 0),
	}

	for _, h := range result.ASCAspects {
		response.ASCAspects = append(response.ASCAspects, hitJSON{h.DirectedPoint, h.NatalPlanet, h.Aspect, h.Orb})
	}
	for _, h := range result.MCAspects {
		response.MCAspects = append(response.MCAspects, hitJSON{h.DirectedPoint, h.NatalPlanet, h.Aspect, h.Orb})
	}

	return json.Marshal(response)
}

// computeInterpretation produces a natural-language chart interpretation.
func computeInterpretation(name string, cd *chartData, lat, lng float64, houseSystem string, orbDeg float64, cacheDir string) ([]byte, error) {
	// Compute houses
	hs := houseSystem
	if hs == "" {
		hs = "P"
	}
	_, ascmc := swe.Houses(cd.jd, lat, lng, hs[0])
	ascLon := ascmc[0]

	// Planet-to-house mapping (whole-sign from ASC)
	houses := make(map[string]int)
	for planet, lon := range cd.planets {
		house := ((int(lon/30) - int(ascLon/30) + 12) % 12) + 1
		houses[planet] = house
	}

	// Compute aspects
	aspects := dignity.DefaultAspects()
	aspectHits := dignity.FindNatalAspects(cd.planets, aspects, orbDeg)
	var hits []dignity.AspectHit
	for _, a := range aspectHits {
		hits = append(hits, dignity.AspectHit{
			Planet1: a.Planet1,
			Planet2: a.Planet2,
			Aspect:  a.Aspect,
			Orb:     a.Orb,
		})
	}

	// Compute patterns
	patternReport := dignity.DetectPatterns(cd.planets, orbDeg)
	var patternHits []dignity.PatternHit
	for _, p := range patternReport.Patterns {
		patternHits = append(patternHits, dignity.PatternHit{
			Name:    p.Name,
			Planets: p.Planets,
		})
	}

	// Build interpretation
	report := dignity.InterpretChart(name, cd.planets, houses, hits, patternHits, nil)
	return report.JSON()
}

// computeAstroCartography computes planetary lines for a world map.
// frame: "tropical", "draconic", or "cross".
func computeAstroCartography(name string, cd *chartData, latStep float64, frame string, cacheDir string) ([]byte, error) {
	gmst := dignity.ComputeGMST(cd.jd)

	type lineJSON struct {
		Planet string            `json:"planet"`
		Angle  string            `json:"angle"`
		Points []dignity.GeoPoint `json:"points"`
	}

	// Determine planet positions based on frame
	positions := cd.planets
	nnLon := cd.nn
	switch frame {
	case "draconic":
		positions = make(map[string]float64)
		for p, lon := range cd.planets {
			positions[p] = normalizeLon(lon - nnLon)
		}
	case "cross":
		// Cross: tropical positions for MC/IC, draconic for ASC/DSC
		// We handle this per-planet below
	}

	var lines []lineJSON
	for planet, lon := range cd.planets {
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

		lines = append(lines, lineJSON{
			Planet: planet,
			Angle:  "MC",
			Points: dignity.ComputeMCLine(ra, gmst, latStep),
		})
		lines = append(lines, lineJSON{
			Planet: planet,
			Angle:  "IC",
			Points: dignity.ComputeICLine(ra, gmst, latStep),
		})

		// ASC and DSC lines
		ascPoints := computeASCLineSWE(ascLon, cd.jd, latStep)
		lines = append(lines, lineJSON{
			Planet: planet,
			Angle:  "ASC",
			Points: ascPoints,
		})
		dscPoints := make([]dignity.GeoPoint, len(ascPoints))
		for i, p := range ascPoints {
			dscPoints[i] = dignity.GeoPoint{Lat: p.Lat, Lon: dignity.NormalizeGeo(p.Lon + 180)}
		}
		lines = append(lines, lineJSON{
			Planet: planet,
			Angle:  "DSC",
			Points: dscPoints,
		})
	}

	response := struct {
		Name  string     `json:"name"`
		JD    float64    `json:"jd"`
		GMST  float64    `json:"gmst"`
		Frame string     `json:"frame"`
		Lines []lineJSON `json:"lines"`
	}{
		Name:  name,
		JD:    cd.jd,
		GMST:  gmst,
		Frame: frame,
		Lines: lines,
	}

	return json.Marshal(response)
}

// computeAstroCartographyCompare returns all three frames plus LinesNear at a target location.
func computeAstroCartographyCompare(name string, cd *chartData, latStep float64, targetLat, targetLng, orb float64, cacheDir string) ([]byte, error) {
	// Compute all three frames
	tropJSON, _ := computeAstroCartography(name, cd, latStep, "tropical", cacheDir)
	dracJSON, _ := computeAstroCartography(name, cd, latStep, "draconic", cacheDir)
	crossJSON, _ := computeAstroCartography(name, cd, latStep, "cross", cacheDir)

	// Parse back to get lines
	var tropResp, dracResp, crossResp struct {
		Lines []struct {
			Planet string            `json:"planet"`
			Angle  string            `json:"angle"`
			Points []dignity.GeoPoint `json:"points"`
		} `json:"lines"`
	}
	json.Unmarshal(tropJSON, &tropResp)
	json.Unmarshal(dracJSON, &dracResp)
	json.Unmarshal(crossJSON, &crossResp)

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

	response := struct {
		Name       string                 `json:"name"`
		TargetLat  float64                `json:"target_lat"`
		TargetLng  float64                `json:"target_lng"`
		Orb        float64                `json:"orb"`
		Hits       []dignity.ThreeWayHit  `json:"hits"`
	}{
		Name:      name,
		TargetLat: targetLat,
		TargetLng: targetLng,
		Orb:       orb,
		Hits:      hits,
	}

	return json.Marshal(response)
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
func computeElectional(name string, cd *chartData, lat, lng float64, startDate, endDate string, orbDeg float64, cacheDir string) ([]byte, error) {
	// Parse dates
	var sy, sm, sd, ey, em, ed int
	fmt.Sscanf(startDate, "%d-%d-%d", &sy, &sm, &sd)
	fmt.Sscanf(endDate, "%d-%d-%d", &ey, &em, &ed)

	startJD := swe.Julday(sy, sm, sd, 12.0, true)
	endJD := swe.Julday(ey, em, ed, 12.0, true)

	// Natal ASC for house computation
	_, ascmc := swe.Houses(cd.jd, lat, lng, 'P')
	ascSign := int(ascmc[0] / 30)

	// Natal planet positions
	natal := cd.planets

	type dayScore struct {
		Date      string   `json:"date"`
		Day       string   `json:"day"`
		Score     int      `json:"score"`
		MoonHouse int      `json:"moon_house"`
		MoonSign  string   `json:"moon_sign"`
		MercSign  string   `json:"merc_sign"`
		Good      []string `json:"good"`
		Bad       []string `json:"bad"`
	}

	var results []dayScore

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

		results = append(results, dayScore{
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

	response := struct {
		Name    string     `json:"name"`
		Start   string     `json:"start_date"`
		End     string     `json:"end_date"`
		Orb     float64    `json:"orb"`
		Results []dayScore `json:"results"`
	}{
		Name:    name,
		Start:   startDate,
		End:     endDate,
		Orb:     orbDeg,
		Results: results,
	}

	return json.Marshal(response)
}

// computeStars computes fixed star conjunctions for a natal chart.
func computeStars(name string, cd *chartData, orbDeg float64, cacheDir string) ([]byte, error) {
	// Compute star positions at birth JD
	starPositions := make(map[string]float64)
	for _, starName := range dignity.StarNames {
		lon, _, _, _ := swe.Fixstar(starName, cd.jd)
		if lon != 0 {
			starPositions[starName] = normalizeLon(lon)
		}
	}

	// Build planet position map
	planetPositions := make(map[string]float64)
	for k, v := range cd.planets {
		planetPositions[k] = v
	}

	conjunctions := dignity.FindStarConjunctions(starPositions, planetPositions, orbDeg)

	// Build JSON response
	type conjJSON struct {
		Star      string  `json:"star"`
		StarLon   float64 `json:"star_lon"`
		Planet    string  `json:"planet"`
		PlanetLon float64 `json:"planet_lon"`
		Orb       float64 `json:"orb"`
		Meaning   string  `json:"meaning"`
	}
	response := struct {
		Name         string     `json:"name"`
		Orb          float64    `json:"orb"`
		Conjunctions []conjJSON `json:"conjunctions"`
	}{
		Name: name,
		Orb:  orbDeg,
	}
	for _, c := range conjunctions {
		response.Conjunctions = append(response.Conjunctions, conjJSON{
			Star:      c.Star,
			StarLon:   c.StarLon,
			Planet:    c.Planet,
			PlanetLon: c.PlanetLon,
			Orb:       c.Orb,
			Meaning:   c.Meaning,
		})
	}

	return json.Marshal(response)
}

// computeStarsCross compares star conjunctions in tropical vs sidereal frames.
func computeStarsCross(name string, cd *chartData, orbDeg float64, cacheDir string) ([]byte, error) {
	// Compute star positions at birth JD (tropical)
	starPositions := make(map[string]float64)
	for _, starName := range dignity.StarNames {
		lon, _, _, _ := swe.Fixstar(starName, cd.jd)
		if lon != 0 {
			starPositions[starName] = normalizeLon(lon)
		}
	}

	// Build planet position map
	planetPositions := make(map[string]float64)
	for k, v := range cd.planets {
		planetPositions[k] = v
	}

	result := dignity.CompareStarConjunctionsCrossSystem(name, starPositions, planetPositions, cd.ayan, orbDeg)
	return json.Marshal(result)
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
func computeMansionConvergence(name string, cd *chartData, cacheDir string) ([]byte, error) {
	// Build tropical planet map
	tropical := make(map[string]float64)
	for k, v := range cd.planets {
		tropical[k] = v
	}

	// Ayanamsa already computed in chartData
	result := dignity.ComputeMansionConvergence(name, tropical, cd.ayan)
	return json.Marshal(result)
}

// computeArabicParts computes Arabic Parts with cross-system comparison.
func computeArabicParts(name string, cd *chartData, orbDeg float64, cacheDir string) ([]byte, error) {
	// Build tropical planet map
	tropical := make(map[string]float64)
	for k, v := range cd.planets {
		tropical[k] = v
	}

	// Determine day/night: Sun above horizon = day
	// Sun is above horizon if its longitude is between ASC and DSC (asc+180)
	sunLon := tropical["Sun"]
	asc := cd.asc
	dsc := normalizeLon(asc + 180)
	var isDay bool
	if asc < dsc {
		isDay = sunLon >= asc && sunLon < dsc
	} else {
		isDay = sunLon >= asc || sunLon < dsc
	}

	result := dignity.ComputePartCrossSystem(name, asc, tropical, cd.ayan, isDay, orbDeg)
	return json.Marshal(result)
}

// computeTropicalSolarReturn computes the tropical solar return for a target year.
// Returns positions, ASC/MC, aspects to natal, and patterns.
func computeTropicalSolarReturn(name string, cd *chartData, targetYear int, cacheDir string) ([]byte, error) {
	natalSun := cd.planets["Sun"]
	jdSR := findSolarReturnJD(natalSun, targetYear, cd.jd)

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
	natal := make(map[string]float64)
	for k, v := range cd.planets {
		natal[k] = v
	}
	natal["Ascendant"] = cd.asc

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

	response := struct {
		Name       string                `json:"name"`
		TargetYear int                   `json:"target_year"`
		JD         float64               `json:"jd"`
		DateTime   string                `json:"datetime"`
		Positions  map[string]float64    `json:"positions"`
		Aspects    []dignity.SynastryHit `json:"aspects"`
		Patterns   []dignity.Pattern     `json:"patterns"`
	}{
		Name:       name,
		TargetYear: targetYear,
		JD:         jdSR,
		DateTime:   dtStr,
		Positions:  srPositions,
		Aspects:    srAspects,
		Patterns:   patternReport.Patterns,
	}

	return json.Marshal(response)
}

// computeProgressed computes a secondary progressed chart and progressed-to-natal aspects.
func computeProgressed(name string, cd *chartData, targetDate string, orbDeg float64, cacheDir string) ([]byte, error) {
	// Parse target date
	var y, m, d int
	fmt.Sscanf(targetDate, "%d-%d-%d", &y, &m, &d)
	utHour := 12.0
	targetJD := swe.Julday(y, m, d, utHour, true)

	// Age in years
	age := (targetJD - cd.jd) / 365.2425

	// Progressed JD: birthJD + age in days (day-for-a-year)
	progJD := cd.jd + age

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
	natalPositions := make(map[string]float64)
	for k, v := range cd.planets {
		natalPositions[k] = v
	}

	report := dignity.ComputeProgressedReport(name, targetDate, age, progPositions, natalPositions, orbDeg)
	return json.Marshal(report)
}

// computeComposite computes a midpoint composite chart for two people.
func computeComposite(name1, name2 string, cd1, cd2 *chartData, orbDeg float64, cacheDir string) ([]byte, error) {
	// Build tropical planet maps
	chart1 := make(map[string]float64)
	for k, v := range cd1.planets {
		chart1[k] = v
	}
	chart1["Ascendant"] = cd1.asc

	chart2 := make(map[string]float64)
	for k, v := range cd2.planets {
		chart2[k] = v
	}
	chart2["Ascendant"] = cd2.asc

	report := dignity.ComputeCompositeReport(name1, name2, chart1, chart2, orbDeg)
	return json.Marshal(report)
}

// computeTraditional computes traditional astrology interpretive data.
func computeTraditional(name string, cd *chartData) ([]byte, error) {
	report := dignity.ComputeTraditionalReport(name, cd.planets, cd.speeds)
	return json.Marshal(report)
}

// computeUranian computes Uranian/Hamburg School midpoint analysis.
func computeUranian(name string, cd *chartData) ([]byte, error) {
	// Compute houses for the chart
	cusps, _ := swe.Houses(cd.jd, 0, 0, 'P')
	houses := make(map[string]float64)
	for i := 1; i <= 12; i++ {
		houses[fmt.Sprintf("H%d", i)] = cusps[i]
	}
	report := uranian.ComputeUranianReport(name, cd.planets, houses)
	return json.Marshal(report)
}

// computeHarmonic computes Addey-style harmonic charts.
func computeHarmonic(name string, cd *chartData, harmonics []int, orb float64) ([]byte, error) {
	report := harmonic.ComputeHarmonicReport(name, cd.planets, harmonics, orb)
	return json.Marshal(report)
}

// computeDivisional computes Vedic divisional charts.
func computeDivisional(name string, cd *chartData, year, month, day int) ([]byte, error) {
	report := divisional.ComputeDivisionalReport(name, cd.planets, cd.ayan, year, month, day)
	return json.Marshal(report)
}

// computeParans computes fixed star parans.
func computeParans(name string, cd *chartData, orb float64, cacheDir string) ([]byte, error) {
	swe.SetEphePath(cacheDir)
	// Compute star positions
	starPositions := make(map[string]float64)
	for _, starName := range dignity.StarNames {
		lon, _, _, _ := swe.Fixstar(starName, cd.jd)
		if lon > 0 {
			starPositions[starName] = lon
		}
	}
	// Compute MC for angles
	_, ascmc := swe.Houses(cd.jd, 0, 0, 'P')
	mc := ascmc[1]
	report := parans.ComputeParansReport(name, starPositions, cd.planets, cd.asc, mc, orb)
	return json.Marshal(report)
}

// computeDeclination computes declination parallels.
func computeDeclination(name string, cd *chartData, orb float64) ([]byte, error) {
	// Build positions with lon/lat pairs
	// We need latitudes — re-compute with CalcUT which returns lat
	// For now, use the stored speeds and re-derive from SWE
	positions := make(map[string][2]float64)
	// We need to re-compute with latitude. Use the JD from cd.
	// Actually, we stored speeds but not latitudes. Re-compute.
	specs := dignity.AllPlanets[:18] // Sun-Pluto+Node+asteroids+Chiron+Lilith
	for _, p := range specs {
		lon, lat, _, _ := swe.CalcUT(cd.jd, p.ID)
		positions[p.Name] = [2]float64{lon, lat}
	}
	report := declination.ComputeDeclinationReport(name, positions, orb)
	return json.Marshal(report)
}

// computeFirdaria computes Persian firdaria planetary periods.
func computeFirdaria(name string, cd *chartData, year, month, day int) ([]byte, error) {
	// Determine if Sun is above horizon (diurnal chart)
	// Sun above ASC and below DSC means above horizon
	sunLon := cd.planets["Sun"]
	// Sun is above horizon if it's between ASC and DSC (through MC)
	// Simplified: Sun in houses 7-12 = above horizon
	_, ascmc := swe.Houses(cd.jd, 0, 0, 'P')
	asc := ascmc[0]
	// Sun above horizon if its longitude is between ASC and ASC+180 (going forward)
	sunAbove := false
	diff := math.Mod(sunLon-asc+360, 360)
	if diff < 180 {
		sunAbove = true
	}
	report := firdaria.ComputeFirdaria(name, sunAbove, year, month, day)
	return json.Marshal(report)
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
