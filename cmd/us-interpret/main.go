package main

import (
	"fmt"

	"github.com/aj-nt/empirical/internal/mundane"
	"github.com/aj-nt/empirical/internal/swe"
)

func init() {
	swe.SetEphePath("/Users/aj/Documents/repos/empirical/ephe")
}

func realCompute(year, month, day int, hour float64, planetID int) (float64, float64, float64, float64) {
	jd := swe.Julday(year, month, day, hour, true)
	return swe.CalcUT(jd, planetID)
}

func realHouses(jd, lat, lon float64, hsys byte) ([13]float64, [10]float64) {
	return swe.Houses(jd, lat, lon, hsys)
}

func main() {
	// US Sibly chart
	report, err := mundane.InterpretNationalChartFull("United States", realCompute, realHouses)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("=== UNITED STATES NATAL CHART — MUNDANE INTERPRETATION ===")
	fmt.Printf("Date: %s\n", report.DateTime)
	fmt.Printf("Location: %s\n", report.Location)
	fmt.Printf("ASC: %s | MC: %s\n\n", report.ASCSign, report.MCSign)

	fmt.Println("ASC INTERPRETATION:")
	fmt.Println(report.ASCInterpretation)
	fmt.Println()
	fmt.Println("MC INTERPRETATION:")
	fmt.Println(report.MCInterpretation)
	fmt.Println()

	fmt.Println("PLANET-IN-HOUSE INTERPRETATIONS:")
	for _, ph := range report.PlanetHouses {
		fmt.Println(ph)
		fmt.Println()
	}

	if len(report.Patterns) > 0 {
		fmt.Println("PATTERNS:")
		for _, p := range report.Patterns {
			fmt.Println(p)
			fmt.Println()
		}
	}

	fmt.Println("SUMMARY:")
	fmt.Println(report.Summary)
}
