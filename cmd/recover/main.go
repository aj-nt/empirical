package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/aj-nt/empirical"
	"github.com/aj-nt/empirical/internal/dignity"
	"github.com/aj-nt/empirical/internal/swe"
)

func main() {
	jsonOut := flag.Bool("json", false, "output as JSON instead of text")
	flag.Parse()

	args := flag.Args()
	if len(args) < 9 {
		fmt.Fprintf(os.Stderr, "Usage: recover [--json] NAME Y M D H MIN TZ_OFFSET LAT LON\n")
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
	lon, _ := strconv.ParseFloat(args[8], 64)
	_ = lat
	_ = lon

	// Extract embedded ephemeris to temp dir
	cacheDir, err := empirical.EnsureEpheCache()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize ephemeris: %v\n", err)
		os.Exit(1)
	}

	// Set ephemeris path
	swe.SetEphePath(cacheDir)

	// Set Lahiri ayanamsa
	swe.SetSidMode(swe.SIDM_LAHIRI, 0, 0)

	// Compute Julian Day in UT
	utHour := float64(hour) + float64(minute)/60.0 - tzOff
	jd := swe.Julday(year, month, day, utHour, true)

	// Get ayanamsa
	ayan := swe.GetAyanamsaUT(jd)

	// Compute planet positions (tropical)
	planetSpecs := []struct {
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

	tropicalLongitudes := make(map[string]float64)
	for _, p := range planetSpecs {
		lon, _, _, _ := swe.CalcUT(jd, p.id)
		tropicalLongitudes[p.name] = lon
	}

	// Compute dignity convergence
	result := dignity.ComputeDignityConvergence(tropicalLongitudes, ayan, name)

	// Output
	if *jsonOut {
		js, err := result.ToJSON()
		if err != nil {
			fmt.Fprintf(os.Stderr, "JSON error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(js)
	} else {
		fmt.Print(dignity.FormatConvergence(result))
	}
}
