package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/aj-nt/empirical"
	"github.com/aj-nt/empirical/internal/dignity"
	"github.com/aj-nt/empirical/internal/swe"
)

func main() {
	// AJ's birth data
	bd := dignity.BirthData{
		Name:     "AJ",
		Year:     1969,
		Month:    2,
		Day:      15,
		Hour:     23,
		Minute:   10,
		Second:   0,
		TZOffset: -8,
		Lat:      47.038,
		Lng:      -122.901,
	}

	// Thailand locations
	locations := []struct {
		Name string
		Lat  float64
		Lng  float64
	}{
		{"Phuket (Rawai)", 7.78, 98.33},
		{"Krabi", 8.09, 98.91},
		{"Chiang Mai", 18.79, 98.98},
		{"Bangkok", 13.75, 100.50},
		{"Koh Samui", 9.50, 100.00},
		{"Pattaya", 12.93, 100.88},
	}

	// Init ephemeris
	cacheDir, err := empirical.EnsureEpheCache()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to init ephemeris: %v\n", err)
		os.Exit(1)
	}
	swe.SetEphePath(cacheDir)
	swe.SetSidMode(swe.SIDM_LAHIRI, 0, 0)

	// Compute base chart
	bc, err := dignity.ComputeBaseChart(bd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to compute base chart: %v\n", err)
		os.Exit(1)
	}

	latStep := 2.0
	orb := 3.0 // degrees — wider orb to catch more lines

	// Compute all three frames
	tropLines := computeAllLines(bc, latStep, dignity.FrameTropical)
	dracLines := computeAllLines(bc, latStep, dignity.FrameDraconic)
	crossLines := computeAllLines(bc, latStep, dignity.FrameCross)

	// Query each location
	type LocationResult struct {
		Name  string                `json:"name"`
		Lat   float64               `json:"lat"`
		Lng   float64               `json:"lng"`
		Hits  []dignity.ThreeWayHit `json:"hits"`
	}

	var results []LocationResult
	for _, loc := range locations {
		hits := dignity.CompareLinesNear(loc.Lat, loc.Lng, tropLines, dracLines, crossLines, orb)
		results = append(results, LocationResult{
			Name: loc.Name,
			Lat:  loc.Lat,
			Lng:  loc.Lng,
			Hits: hits,
		})
	}

	// Output
	out, _ := json.MarshalIndent(results, "", "  ")
	fmt.Println(string(out))
}

// computeAllLines computes MC, IC, ASC, and DSC lines for all planets in a given frame.
func computeAllLines(bc *dignity.BaseChart, latStep float64, frame dignity.Frame) []dignity.AstroLine {
	gmst := dignity.ComputeGMST(bc.JD)
	nnLon := bc.NorthNode

	var lines []dignity.AstroLine
	for planet, lon := range dignity.TropicalToLonMap(bc.Tropical) {
		var ra, ascLon float64

		switch frame {
		case dignity.FrameDraconic:
			dracLon := dignity.NormalizeLon(lon - nnLon)
			ra = dignity.LonToRA(dracLon, dignity.ObliquityDeg)
			ascLon = dracLon
		case dignity.FrameCross:
			ra = dignity.LonToRA(lon, dignity.ObliquityDeg) // tropical MC/IC
			ascLon = dignity.NormalizeLon(lon - nnLon)      // draconic ASC/DSC
		default: // FrameTropical
			ra = dignity.LonToRA(lon, dignity.ObliquityDeg)
			ascLon = lon
		}

		lines = append(lines, dignity.AstroLine{
			Planet: planet,
			Angle:  "MC",
			Points: dignity.ComputeMCLine(ra, gmst, latStep),
		})
		lines = append(lines, dignity.AstroLine{
			Planet: planet,
			Angle:  "IC",
			Points: dignity.ComputeICLine(ra, gmst, latStep),
		})
		lines = append(lines, dignity.AstroLine{
			Planet: planet,
			Angle:  "ASC",
			Points: dignity.ComputeASCLine(ascLon, bc.JD, latStep, swe.Houses),
		})
		lines = append(lines, dignity.AstroLine{
			Planet: planet,
			Angle:  "DSC",
			Points: dignity.ComputeDSCLine(ascLon, bc.JD, latStep, swe.Houses),
		})
	}
	return lines
}
