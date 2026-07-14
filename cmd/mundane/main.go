package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/aj-nt/empirical/internal/dignity"
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
	nation := flag.String("nation", "", "Nation name for transits and chart analysis")
	startStr := flag.String("start", "", "Start date (YYYY-MM-DD)")
	endStr := flag.String("end", "", "End date (YYYY-MM-DD)")
	orb := flag.Float64("orb", 3.0, "Orb in degrees for aspects")
	lat := flag.Float64("lat", 0, "Latitude for chart casting")
	lon := flag.Float64("lon", 0, "Longitude for chart casting")
	listNations := flag.Bool("list-nations", false, "List available national charts")
	flag.Parse()

	if *listNations {
		charts := mundane.NationalCharts()
		fmt.Printf("Available national charts (%d):\n", len(charts))
		for _, c := range charts {
			fmt.Printf("  %-20s %d-%02d-%02d  %s\n", c.Name, c.Year, c.Month, c.Day, c.Note)
		}
		return
	}

	if *startStr == "" || *endStr == "" {
		fmt.Fprintf(os.Stderr, "Usage: mundane --start YYYY-MM-DD --end YYYY-MM-DD [--nation NAME] [--lat LAT --lon LON] [--orb DEG]\n")
		fmt.Fprintf(os.Stderr, "       mundane --list-nations\n")
		os.Exit(1)
	}

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

	fmt.Printf("=== MUNDANE EVENTS: %s to %s ===\n", start.Format("2006-01-02"), end.Format("2006-01-02"))
	if *nation != "" {
		fmt.Printf("Nation: %s\n", *nation)
	}
	if *lat != 0 || *lon != 0 {
		fmt.Printf("Location: %.2f, %.2f\n", *lat, *lon)
	}
	fmt.Printf("Orb: %.1f°\n\n", *orb)

	// ── Solar Ingresses ──
	ingresses, err := mundane.FindSolarIngresses(start, end, realCompute)
	if err != nil {
		fmt.Printf("Solar ingresses error: %v\n", err)
	} else {
		fmt.Printf("── Solar Ingresses (%d) ──\n", len(ingresses))
		for _, ing := range ingresses {
			fmt.Printf("  Sun → %-8s  %s UTC\n", ing.Sign, ing.Time.Format("2006-01-02 15:04:05"))
		}
		if len(ingresses) == 0 {
			fmt.Println("  (none)")
		}
	}
	fmt.Println()

	// ── Lunations ──
	lunations, err := mundane.FindLunations(start, end, realCompute)
	if err != nil {
		fmt.Printf("Lunations error: %v\n", err)
	} else {
		fmt.Printf("── Lunations (%d) ──\n", len(lunations))
		for _, l := range lunations {
			fmt.Printf("  %-10s  %s UTC\n", l.Type, l.Time.Format("2006-01-02 15:04:05"))
		}
	}
	fmt.Println()

	// ── Eclipses ──
	eclipses, err := mundane.FindEclipses(start, end, realCompute)
	if err != nil {
		fmt.Printf("Eclipses error: %v\n", err)
	} else {
		fmt.Printf("── Eclipses (%d) ──\n", len(eclipses))
		for _, e := range eclipses {
			fmt.Printf("  %-14s  %s UTC\n", e.Type, e.Time.Format("2006-01-02 15:04:05"))
		}
		if len(eclipses) == 0 {
			fmt.Println("  (none)")
		}
	}
	fmt.Println()

	// ── Planetary Ingresses ──
	outerPlanets := []struct {
		id   int
		name string
	}{
		{4, "Mars"}, {5, "Jupiter"}, {6, "Saturn"},
		{7, "Uranus"}, {8, "Neptune"}, {9, "Pluto"},
	}
	fmt.Printf("── Planetary Ingresses ──\n")
	found := false
	for _, p := range outerPlanets {
		ingresses, _ := mundane.FindPlanetaryIngresses(start, end, p.id, p.name, realCompute)
		for _, ing := range ingresses {
			fmt.Printf("  %-8s → %-8s  %s UTC\n", ing.Planet, ing.Sign, ing.Time.Format("2006-01-02 15:04:05"))
			found = true
		}
	}
	if !found {
		fmt.Println("  (none)")
	}
	fmt.Println()

	// ── Major Conjunctions ──
	pairs := []struct {
		id1, id2 int
		n1, n2   string
	}{
		{5, 6, "Jupiter", "Saturn"},
		{6, 7, "Saturn", "Uranus"},
		{5, 7, "Jupiter", "Uranus"},
		{4, 5, "Mars", "Jupiter"},
		{4, 6, "Mars", "Saturn"},
	}
	fmt.Printf("── Major Conjunctions ──\n")
	found = false
	for _, pair := range pairs {
		conj, _ := mundane.FindConjunctions(start, end, pair.id1, pair.n1, pair.id2, pair.n2, realCompute)
		for _, c := range conj {
			fmt.Printf("  %s ☌ %s  %s UTC\n", c.Planet1, c.Planet2, c.Time.Format("2006-01-02 15:04:05"))
			found = true
		}
	}
	if !found {
		fmt.Println("  (none)")
	}
	fmt.Println()

	// ── Ingress/Lunation Charts (if location provided) ──
	if *lat != 0 || *lon != 0 {
		fmt.Printf("── Chart Analysis (%.2f, %.2f) ──\n", *lat, *lon)

		// Solar ingress charts
		for _, ing := range ingresses {
			chart, err := mundane.CastIngressChart(ing, *lat, *lon, realCompute, realHouses)
			if err != nil {
				fmt.Printf("  %s ingress chart error: %v\n", ing.Sign, err)
				continue
			}
			printChartSummary(fmt.Sprintf("Sun → %s Ingress", ing.Sign), chart, *orb)
		}

		// Lunation charts
		for _, l := range lunations {
			chart, err := mundane.CastLunationChart(l, *lat, *lon, realCompute, realHouses)
			if err != nil {
				fmt.Printf("  %s chart error: %v\n", l.Type, err)
				continue
			}
			printChartSummary(l.Type, chart, *orb)
		}

		// Eclipse charts
		for _, e := range eclipses {
			chart, err := mundane.CastChart(e.Time, *lat, *lon, realCompute, realHouses, 'W')
			if err != nil {
				fmt.Printf("  %s chart error: %v\n", e.Type, err)
				continue
			}
			printChartSummary(e.Type, chart, *orb)
		}
		fmt.Println()
	}

	// ── Nation Transits ──
	if *nation != "" {
		entry, ok := mundane.NationalChart(*nation)
		if !ok {
			fmt.Printf("Unknown nation: %s\n", *nation)
			fmt.Println()
		} else {
			fmt.Printf("── %s Natal Chart ──\n", entry.Name)
			fmt.Printf("  Date: %d-%02d-%02d %.2f UT\n", entry.Year, entry.Month, entry.Day, entry.Hour)
			fmt.Printf("  Location: %.2f, %.2f\n", entry.Lat, entry.Lon)
			fmt.Printf("  Note: %s\n", entry.Note)

			// Cast natal chart
			natalTime := time.Date(entry.Year, time.Month(entry.Month), entry.Day,
				int(entry.Hour), int((entry.Hour-float64(int(entry.Hour)))*60), 0, 0, time.UTC)
			natalChart, err := mundane.CastChart(natalTime, entry.Lat, entry.Lon, realCompute, realHouses, 'W')
			if err == nil {
				fmt.Printf("  ASC: %.2f°  MC: %.2f°\n", natalChart.ASC, natalChart.MC)
				fmt.Printf("  Sun: %.2f° (%s)  Moon: %.2f° (%s)\n",
					natalChart.Planets["Sun"], dignity.SignForLongitude(natalChart.Planets["Sun"]),
					natalChart.Planets["Moon"], dignity.SignForLongitude(natalChart.Planets["Moon"]))
			}
			fmt.Println()

			// Transits
			fmt.Printf("── Transits to %s ──\n", entry.Name)
			hits, err := mundane.NationTransits(*nation, *startStr, *endStr, *orb, realCompute, realHouses)
			if err != nil {
				fmt.Printf("  Error: %v\n", err)
			} else if len(hits) == 0 {
				fmt.Println("  (none within orb)")
			} else {
				// Sort by date
				sort.Slice(hits, func(i, j int) bool {
					return hits[i].DateStart < hits[j].DateStart
				})
				for _, h := range hits {
					dateRange := h.DateStart
					if h.DateEnd != h.DateStart {
						dateRange = fmt.Sprintf("%s → %s", h.DateStart, h.DateEnd)
					}
					fmt.Printf("  %-8s %-7s %-8s  %-20s  orb: %.2f°\n",
						h.TransitPlanet, h.Aspect, h.NatalPlanet, dateRange, h.MinOrb)
				}
			}
			fmt.Println()

			// Ingress charts for the nation's capital
			if len(ingresses) > 0 {
				fmt.Printf("── Ingress Charts for %s ──\n", entry.Name)
				for _, ing := range ingresses {
					chart, err := mundane.CastIngressChart(ing, entry.Lat, entry.Lon, realCompute, realHouses)
					if err != nil {
						fmt.Printf("  %s ingress chart error: %v\n", ing.Sign, err)
						continue
					}
					printChartSummary(fmt.Sprintf("Sun → %s Ingress", ing.Sign), chart, *orb)
				}
				fmt.Println()
			}
		}
	}
}

func printChartSummary(title string, chart *mundane.MundaneChart, orb float64) {
	fmt.Printf("\n  ▸ %s\n", title)
	fmt.Printf("    Time: %s UTC\n", chart.Time.Format("2006-01-02 15:04:05"))
	fmt.Printf("    ASC: %.2f° (%s)  MC: %.2f° (%s)\n",
		chart.ASC, dignity.SignForLongitude(chart.ASC), chart.MC, dignity.SignForLongitude(chart.MC))

	// Planet positions (condensed)
	var planetNames []string
	for n := range chart.Planets {
		planetNames = append(planetNames, n)
	}
	sort.Strings(planetNames)

	fmt.Print("    Planets: ")
	for i, n := range planetNames {
		if i > 0 {
			fmt.Print(", ")
		}
		fmt.Printf("%s %.1f° %s", n, chart.Planets[n], dignity.SignForLongitude(chart.Planets[n]))
	}
	fmt.Println()

	// Patterns
	report := mundane.ChartPatterns(chart, orb)
	if len(report.Patterns) > 0 {
		fmt.Printf("    Patterns (%d):\n", len(report.Patterns))
		for _, p := range report.Patterns {
			fmt.Printf("      %s: %s\n", p.Name, strings.Join(p.Planets, ", "))
		}
	}
}

