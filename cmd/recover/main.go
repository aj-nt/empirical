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

		transits := func(name string, year, month, day, hour, minute int, tzOff, lat, lng float64, startDate, endDate string, orbDeg float64) ([]byte, error) {
			return computeTransits(name, year, month, day, hour, minute, tzOff, lat, lng, startDate, endDate, orbDeg, cacheDir)
		}

		// Use embedded web files, stripping the "web/" prefix
		staticFS, err := fs.Sub(empirical.WebFiles, "web")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to load web files: %v\n", err)
			os.Exit(1)
		}

		if err := server.Run(port, staticFS, compute, aspects, timing, transits); err != nil {
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

		result, err := computeTransits(name, year, month, day, hour, minute, tzOff, lat, lng, startDate, endDate, *orbDeg, "")
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
	specs := []struct {
		name string
		id   int
	}{
		{"Sun", swe.SUN},
		{"Moon", swe.MOON},
		{"Mercury", swe.MERCURY},
		{"Venus", swe.VENUS},
		{"Mars", swe.MARS},
		{"Jupiter", swe.JUPITER},
		{"Saturn", swe.SATURN},
	}
	for _, p := range specs {
		lon, _, _, _ := swe.CalcUT(jd, p.id)
		pls[p.name] = lon
	}

	// Get ASC and NN
	nnLon, _, _, _ := swe.CalcUT(jd, swe.MEAN_NODE)
	_, ascmc := swe.Houses(jd, lat, lng, 'P')
	asc := ascmc[0]

	return &chartData{
		planets: pls,
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
func computeTransits(name string, year, month, day, hour, minute int, tzOff, lat, lng float64, startDate, endDate string, orbDeg float64, cacheDir string) ([]byte, error) {
	cd := computePositions(year, month, day, hour, minute, tzOff, lat, lng, cacheDir)

	// Build planet positions including outer planets
	natalLongs := make(map[string]float64)
	for k, v := range cd.planets {
		natalLongs[k] = v
	}
	outerIDs := map[string]int{
		"Uranus": swe.URANUS, "Neptune": swe.NEPTUNE, "Pluto": swe.PLUTO,
	}
	for name, id := range outerIDs {
		lon, _, _, _ := swe.CalcUT(cd.jd, id)
		natalLongs[name] = lon
	}

	natalPlanets := []string{"Sun", "Moon", "Mercury", "Venus", "Mars", "Jupiter", "Saturn", "Uranus", "Neptune", "Pluto"}

	// Build real-SWE compute function
	compute := func(year, month, day int, hour float64, planetID int) (float64, float64, float64, float64) {
		utHour := hour - tzOff
		jd := swe.Julday(year, month, day, utHour, true)
		return swe.CalcUT(jd, planetID)
	}

	hits, err := dignity.ScanTransits(natalLongs, natalPlanets, startDate, endDate, dignity.HardAspectsOnly(), orbDeg, compute)
	if err != nil {
		return nil, err
	}

	compact := dignity.CompactTransitsWithRange(hits)

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
		Name    string    `json:"name"`
		Transits []hitJSON `json:"transits"`
	}{
		Name: name,
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

// printReport prints a human-readable multi-phase report.
func printReport(fr *dignity.FullReport) {
	fmt.Printf("╔══════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║  Empirical Astrology — Signal Recovery Report           ║\n")
	fmt.Printf("╚══════════════════════════════════════════════════════════╝\n")
	fmt.Printf("\n  Name: %s\n", fr.Name)
	fmt.Printf("  Ayanamsa: %.2f deg Lahiri\n\n", fr.AyanamsaDegrees)

	fmt.Print(dignity.FormatConvergence(fr.Phase1Dignity))
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
