package main

import (
	"fmt"
	"math"
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
	// Donald Trump: June 14, 1946, 10:54 AM EDT = 14:54 UT, Queens NY
	natalTime := time.Date(1946, 6, 14, 14, 54, 0, 0, time.UTC)
	natalLat, natalLon := 40.73, -73.79

	natalChart, err := mundane.CastChart(natalTime, natalLat, natalLon, realCompute, realHouses, 'W')
	if err != nil {
		fmt.Printf("Error casting natal: %v\n", err)
		return
	}

	fmt.Println("=== DONALD TRUMP NATAL CHART ===")
	fmt.Printf("Date: June 14, 1946 14:54 UT  |  Queens, NY\n")
	fmt.Printf("ASC: %.2f° (%s)  MC: %.2f° (%s)\n",
		natalChart.ASC, dignity.SignForLongitude(natalChart.ASC), natalChart.MC, dignity.SignForLongitude(natalChart.MC))

	var planetNames []string
	for n := range natalChart.Planets {
		planetNames = append(planetNames, n)
	}
	sort.Strings(planetNames)

	fmt.Println("\nPlanets:")
	for _, n := range planetNames {
		fmt.Printf("  %-8s %7.2f° %-12s\n", n, natalChart.Planets[n], dignity.SignForLongitude(natalChart.Planets[n]))
	}

	houses := mundane.PlanetHouses(natalChart)
	fmt.Println("\nHouses (Whole Sign):")
	for _, n := range planetNames {
		fmt.Printf("  %-8s House %d\n", n, houses[n])
	}

	report := mundane.ChartPatterns(natalChart, 5.0)
	if len(report.Patterns) > 0 {
		fmt.Println("\nPatterns:")
		for _, p := range report.Patterns {
			fmt.Printf("  %s: %s\n", p.Name, strings.Join(p.Planets, ", "))
		}
	}

	for _, month := range []struct{ start, end, label string }{
		{"2026-07-01", "2026-07-31", "JULY 2026"},
		{"2026-08-01", "2026-08-31", "AUGUST 2026"},
	} {
		fmt.Printf("\n\n=== TRANSITS: %s ===\n", month.label)

		start, _ := time.Parse("2006-01-02", month.start)
		end, _ := time.Parse("2006-01-02", month.end)

		type hit struct {
			date          string
			transitPlanet string
			natalPlanet   string
			aspect        string
			orb           float64
		}

		aspects := []struct {
			angle float64
			name  string
		}{
			{0, "conjunction"}, {60, "sextile"}, {90, "square"}, {120, "trine"}, {180, "opposition"},
		}

		transitPlanets := []struct {
			id   int
			name string
		}{
			{0, "Sun"}, {1, "Moon"}, {2, "Mercury"}, {3, "Venus"}, {4, "Mars"},
			{5, "Jupiter"}, {6, "Saturn"}, {7, "Uranus"}, {8, "Neptune"}, {9, "Pluto"},
			{10, "Node"},
		}

		var hits []hit
		current := start
		for !current.After(end) {
			y, m, d := current.Year(), int(current.Month()), current.Day()
			for _, tp := range transitPlanets {
				tLon, _, _, _ := realCompute(y, m, d, 12.0, tp.id)
				for _, np := range planetNames {
					nLon := natalChart.Planets[np]
					dist := dignity.AngleDist(tLon, nLon)
					for _, asp := range aspects {
						diff := dist - asp.angle
						if diff < 0 {
							diff = -diff
						}
						if diff <= 3.0 {
							hits = append(hits, hit{
								date:          current.Format("2006-01-02"),
								transitPlanet: tp.name,
								natalPlanet:   np,
								aspect:        asp.name,
								orb:           math.Round(diff*100) / 100,
							})
						}
					}
				}
			}
			current = current.AddDate(0, 0, 1)
		}

		type key struct {
			transitPlanet, natalPlanet, aspect string
		}
		groups := make(map[key][]hit)
		for _, h := range hits {
			k := key{h.transitPlanet, h.natalPlanet, h.aspect}
			groups[k] = append(groups[k], h)
		}

		type compacted struct {
			transitPlanet, natalPlanet, aspect, dateStart, dateEnd string
			minOrb                                                  float64
		}
		var result []compacted
		for k, group := range groups {
			if k.transitPlanet == "Moon" {
				continue
			}
			if len(group) < 2 && group[0].orb > 0.5 {
				continue
			}
			best := group[0]
			for _, h := range group {
				if h.orb < best.orb {
					best = h
				}
			}
			result = append(result, compacted{
				transitPlanet: k.transitPlanet,
				natalPlanet:   k.natalPlanet,
				aspect:        k.aspect,
				dateStart:     group[0].date,
				dateEnd:       group[len(group)-1].date,
				minOrb:        best.orb,
			})
		}

		sort.Slice(result, func(i, j int) bool { return result[i].dateStart < result[j].dateStart })

		allMonth := []compacted{}
		week1, week2, week3, week4 := []compacted{}, []compacted{}, []compacted{}, []compacted{}
		for _, r := range result {
			if r.dateStart <= month.start && r.dateEnd >= month.end {
				allMonth = append(allMonth, r)
			} else if r.dateStart <= month.start[:8]+"07" {
				week1 = append(week1, r)
			} else if r.dateStart <= month.start[:8]+"14" {
				week2 = append(week2, r)
			} else if r.dateStart <= month.start[:8]+"21" {
				week3 = append(week3, r)
			} else {
				week4 = append(week4, r)
			}
		}

		if len(allMonth) > 0 {
			fmt.Println("\nALL-MONTH:")
			for _, r := range allMonth {
				fmt.Printf("  %-8s %-11s natal %-8s  orb %.2f°\n", r.transitPlanet, r.aspect, r.natalPlanet, r.minOrb)
			}
		}

		for label, group := range map[string][]compacted{"WEEK 1": week1, "WEEK 2": week2, "WEEK 3": week3, "WEEK 4": week4} {
			if len(group) > 0 {
				fmt.Printf("\n%s:\n", label)
				for _, r := range group {
					dateRange := r.dateStart
					if r.dateEnd != r.dateStart {
						dateRange = fmt.Sprintf("%s → %s", r.dateStart, r.dateEnd)
					}
					fmt.Printf("  %-8s %-11s natal %-8s  %-22s orb %.2f°\n", r.transitPlanet, r.aspect, r.natalPlanet, dateRange, r.minOrb)
				}
			}
		}
	}
}
