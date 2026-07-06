package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

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

// EventExport is the JSON output format.
type EventExport struct {
	Generated   string              `json:"generated"`
	StartDate   string              `json:"start_date"`
	EndDate     string              `json:"end_date"`
	Ingresses   []IngressExport     `json:"solar_ingresses"`
	Lunations   []LunationExport    `json:"lunations"`
	Eclipses    []EclipseExport     `json:"eclipses"`
	Conjunctions []ConjunctionExport `json:"conjunctions"`
	PlanetIngresses []PlanetIngressExport `json:"planetary_ingresses"`
}

type IngressExport struct {
	Sign string `json:"sign"`
	Date string `json:"date"`
	JD   float64 `json:"jd"`
}

type LunationExport struct {
	Type string `json:"type"`
	Date string `json:"date"`
	JD   float64 `json:"jd"`
}

type EclipseExport struct {
	Type string `json:"type"`
	Date string `json:"date"`
	JD   float64 `json:"jd"`
}

type ConjunctionExport struct {
	Planet1 string `json:"planet1"`
	Planet2 string `json:"planet2"`
	Date    string `json:"date"`
	JD      float64 `json:"jd"`
}

type PlanetIngressExport struct {
	Planet string `json:"planet"`
	Sign   string `json:"sign"`
	Date   string `json:"date"`
	JD     float64 `json:"jd"`
}

func main() {
	startStr := flag.String("start", "1900-01-01", "Start date (YYYY-MM-DD)")
	endStr := flag.String("end", "2030-12-31", "End date (YYYY-MM-DD)")
	output := flag.String("output", "", "Output file (default: stdout)")
	flag.Parse()

	start, err := time.Parse("2006-01-02", *startStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid start date: %v\n", err)
		os.Exit(1)
	}
	end, err := time.Parse("2006-01-02", *endStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid end date: %v\n", err)
		os.Exit(1)
	}

	export := EventExport{
		Generated: time.Now().UTC().Format(time.RFC3339),
		StartDate: *startStr,
		EndDate:   *endStr,
	}

	// Solar ingresses
	ingresses, _ := mundane.FindSolarIngresses(start, end, realCompute)
	for _, ing := range ingresses {
		export.Ingresses = append(export.Ingresses, IngressExport{
			Sign: ing.Sign,
			Date: ing.Time.Format("2006-01-02"),
			JD:   swe.Julday(ing.Time.Year(), int(ing.Time.Month()), ing.Time.Day(),
				float64(ing.Time.Hour())+float64(ing.Time.Minute())/60.0, true),
		})
	}

	// Lunations
	lunations, _ := mundane.FindLunations(start, end, realCompute)
	for _, l := range lunations {
		export.Lunations = append(export.Lunations, LunationExport{
			Type: l.Type,
			Date: l.Time.Format("2006-01-02"),
			JD:   swe.Julday(l.Time.Year(), int(l.Time.Month()), l.Time.Day(),
				float64(l.Time.Hour())+float64(l.Time.Minute())/60.0, true),
		})
	}

	// Eclipses
	eclipses, _ := mundane.FindEclipses(start, end, realCompute)
	for _, e := range eclipses {
		export.Eclipses = append(export.Eclipses, EclipseExport{
			Type: e.Type,
			Date: e.Time.Format("2006-01-02"),
			JD:   swe.Julday(e.Time.Year(), int(e.Time.Month()), e.Time.Day(),
				float64(e.Time.Hour())+float64(e.Time.Minute())/60.0, true),
		})
	}

	// Major conjunctions
	pairs := []struct {
		id1, id2 int
		n1, n2   string
	}{
		{5, 6, "Jupiter", "Saturn"},
		{6, 7, "Saturn", "Uranus"},
		{5, 7, "Jupiter", "Uranus"},
		{5, 9, "Jupiter", "Pluto"},
		{6, 9, "Saturn", "Pluto"},
		{7, 9, "Uranus", "Pluto"},
		{4, 6, "Mars", "Saturn"},
		{4, 5, "Mars", "Jupiter"},
		{4, 7, "Mars", "Uranus"},
		{4, 9, "Mars", "Pluto"},
	}
	for _, pair := range pairs {
		conj, _ := mundane.FindConjunctions(start, end, pair.id1, pair.n1, pair.id2, pair.n2, realCompute)
		for _, c := range conj {
			export.Conjunctions = append(export.Conjunctions, ConjunctionExport{
				Planet1: c.Planet1,
				Planet2: c.Planet2,
				Date:    c.Time.Format("2006-01-02"),
				JD:      swe.Julday(c.Time.Year(), int(c.Time.Month()), c.Time.Day(),
					float64(c.Time.Hour())+float64(c.Time.Minute())/60.0, true),
			})
		}
	}

	// Planetary ingresses (outer planets only)
	outerPlanets := []struct {
		id   int
		name string
	}{
		{5, "Jupiter"}, {6, "Saturn"}, {7, "Uranus"}, {8, "Neptune"}, {9, "Pluto"},
	}
	for _, p := range outerPlanets {
		ing, _ := mundane.FindPlanetaryIngresses(start, end, p.id, p.name, realCompute)
		for _, i := range ing {
			export.PlanetIngresses = append(export.PlanetIngresses, PlanetIngressExport{
				Planet: i.Planet,
				Sign:   i.Sign,
				Date:   i.Time.Format("2006-01-02"),
				JD:     swe.Julday(i.Time.Year(), int(i.Time.Month()), i.Time.Day(),
					float64(i.Time.Hour())+float64(i.Time.Minute())/60.0, true),
			})
		}
	}

	data, err := json.MarshalIndent(export, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "JSON marshal error: %v\n", err)
		os.Exit(1)
	}

	if *output != "" {
		if err := os.WriteFile(*output, data, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Write error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Wrote %d bytes to %s\n", len(data), *output)
	} else {
		fmt.Println(string(data))
	}
}
