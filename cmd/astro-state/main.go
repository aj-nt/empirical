package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"
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

func realHouses(jd, lat, lon float64, hsys byte) ([13]float64, [10]float64) {
	return swe.Houses(jd, lat, lon, hsys)
}

type AstroState struct {
	Date              string              `json:"date"`
	Event             string              `json:"event,omitempty"`
	Category          string              `json:"category,omitempty"`
	Transits          []TransitHit        `json:"transits"`
	RecentIngresses   []string            `json:"recent_ingresses"`
	RecentEclipses    []string            `json:"recent_eclipses"`
	RecentConjunctions []string           `json:"recent_conjunctions"`
	LunationPhase     string              `json:"lunation_phase"`
	Patterns          []string            `json:"patterns"`
	Notable           []string            `json:"notable"`
}

type TransitHit struct {
	Planet     string  `json:"planet"`
	Aspect     string  `json:"aspect"`
	NatalPoint string  `json:"natal_point"`
	Orb        float64 `json:"orb"`
}


func aspectName(orb float64) string {
	abs := orb
	if abs < 0 {
		abs = -abs
	}
	if abs <= 1.0 {
		return "conjunction"
	}
	if abs >= 179 && abs <= 181 {
		return "opposition"
	}
	if abs >= 89 && abs <= 91 {
		return "square"
	}
	if abs >= 119 && abs <= 121 {
		return "trine"
	}
	if abs >= 59 && abs <= 61 {
		return "sextile"
	}
	return fmt.Sprintf("%.0f°", orb)
}

func main() {
	dateStr := flag.String("date", "", "Date to analyze (YYYY-MM-DD)")
	event := flag.String("event", "", "Event description")
	category := flag.String("category", "", "Event category")
	nation := flag.String("nation", "United States", "Nation to analyze")
	orbDeg := flag.Float64("orb", 3.0, "Orb for transits")
	lookback := flag.Int("lookback", 90, "Days to look back for recent events")
	jsonOut := flag.Bool("json", false, "Output as JSON")
	flag.Parse()

	if *dateStr == "" {
		fmt.Fprintf(os.Stderr, "Usage: astro-state --date YYYY-MM-DD [--event 'description'] [--category 'war|financial|political|...']\n")
		os.Exit(1)
	}

	date, err := time.Parse("2006-01-02", *dateStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid date: %v\n", err)
		os.Exit(1)
	}

	state := AstroState{
		Date:     *dateStr,
		Event:    *event,
		Category: *category,
	}

	// Get national chart
	entry, ok := mundane.NationalChart(*nation)
	if !ok {
		fmt.Fprintf(os.Stderr, "Unknown nation: %s\n", *nation)
		os.Exit(1)
	}

	natalTime := time.Date(entry.Year, time.Month(entry.Month), entry.Day,
		int(entry.Hour), int((entry.Hour-float64(int(entry.Hour)))*60), 0, 0, time.UTC)
	natalChart, err := mundane.CastChart(natalTime, entry.Lat, entry.Lon, realCompute, realHouses, 'W')
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error casting natal chart: %v\n", err)
		os.Exit(1)
	}

	// Cast chart for the event date
	eventChart, err := mundane.CastChart(date, entry.Lat, entry.Lon, realCompute, realHouses, 'W')
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error casting event chart: %v\n", err)
		os.Exit(1)
	}

	// Transits: compare every transiting planet to every natal planet
	planetOrder := []string{"Sun", "Moon", "Mercury", "Venus", "Mars", "Jupiter", "Saturn", "Uranus", "Neptune", "Pluto", "Node", "SouthNode"}
	for _, tName := range planetOrder {
		transitPos, tok := eventChart.Planets[tName]
		if !tok {
			continue
		}
		for _, nName := range planetOrder {
			natalPos, nok := natalChart.Planets[nName]
			if !nok {
				continue
			}

			diff := transitPos - natalPos
			if diff < 0 {
				diff = -diff
			}
			if diff > 180 {
				diff = 360 - diff
			}

			if diff <= *orbDeg {
				asp := "conjunction"
				if diff > 1.0 {
					asp = fmt.Sprintf("conjunction (%.1f°)", diff)
				}
				state.Transits = append(state.Transits, TransitHit{
					Planet: tName, Aspect: asp, NatalPoint: nName, Orb: diff,
				})
			}
			if diff >= 180-*orbDeg && diff <= 180 {
				asp := "opposition"
				orb := 180 - diff
				if orb < 0 {
					orb = -orb
				}
				if orb > 1.0 {
					asp = fmt.Sprintf("opposition (%.1f°)", orb)
				}
				state.Transits = append(state.Transits, TransitHit{
					Planet: tName, Aspect: asp, NatalPoint: nName, Orb: orb,
				})
			}
			if diff >= 90-*orbDeg && diff <= 90+*orbDeg {
				orb := diff - 90
				if orb < 0 {
					orb = -orb
				}
				asp := "square"
				if orb > 1.0 {
					asp = fmt.Sprintf("square (%.1f°)", orb)
				}
				state.Transits = append(state.Transits, TransitHit{
					Planet: tName, Aspect: asp, NatalPoint: nName, Orb: orb,
				})
			}
			if diff >= 120-*orbDeg && diff <= 120+*orbDeg {
				orb := diff - 120
				if orb < 0 {
					orb = -orb
				}
				asp := "trine"
				if orb > 1.0 {
					asp = fmt.Sprintf("trine (%.1f°)", orb)
				}
				state.Transits = append(state.Transits, TransitHit{
					Planet: tName, Aspect: asp, NatalPoint: nName, Orb: orb,
				})
			}
			if diff >= 60-*orbDeg && diff <= 60+*orbDeg {
				orb := diff - 60
				if orb < 0 {
					orb = -orb
				}
				asp := "sextile"
				if orb > 1.0 {
					asp = fmt.Sprintf("sextile (%.1f°)", orb)
				}
				state.Transits = append(state.Transits, TransitHit{
					Planet: tName, Aspect: asp, NatalPoint: nName, Orb: orb,
				})
			}
		}
	}

	// Sort transits by orb
	sort.Slice(state.Transits, func(i, j int) bool {
		oi := state.Transits[i].Orb
		if oi < 0 {
			oi = -oi
		}
		oj := state.Transits[j].Orb
		if oj < 0 {
			oj = -oj
		}
		return oi < oj
	})

	// Recent ingresses (lookback days)
	lookbackStart := date.AddDate(0, 0, -*lookback)
	ingresses, _ := mundane.FindSolarIngresses(lookbackStart, date, realCompute)
	for _, ing := range ingresses {
		state.RecentIngresses = append(state.RecentIngresses,
			fmt.Sprintf("%s: Sun→%s", ing.Time.Format("2006-01-02"), ing.Sign))
	}

	// Recent eclipses
	eclipses, _ := mundane.FindEclipses(lookbackStart, date, realCompute)
	for _, e := range eclipses {
		state.RecentEclipses = append(state.RecentEclipses,
			fmt.Sprintf("%s: %s", e.Time.Format("2006-01-02"), e.Type))
	}

	// Recent major conjunctions
	pairs := []struct {
		id1, id2 int
		n1, n2   string
	}{
		{5, 6, "Jupiter", "Saturn"},
		{6, 9, "Saturn", "Pluto"},
		{5, 9, "Jupiter", "Pluto"},
		{7, 9, "Uranus", "Pluto"},
		{4, 6, "Mars", "Saturn"},
		{4, 9, "Mars", "Pluto"},
	}
	for _, pair := range pairs {
		conj, _ := mundane.FindConjunctions(lookbackStart, date, pair.id1, pair.n1, pair.id2, pair.n2, realCompute)
		for _, c := range conj {
			state.RecentConjunctions = append(state.RecentConjunctions,
				fmt.Sprintf("%s: %s ☌ %s", c.Time.Format("2006-01-02"), c.Planet1, c.Planet2))
		}
	}

	// Lunation phase
	sunPos := eventChart.Planets["Sun"]
	moonPos := eventChart.Planets["Moon"]
	diff := moonPos - sunPos
	if diff < 0 {
		diff += 360
	}
	if diff < 45 {
		state.LunationPhase = "New Moon (waxing crescent)"
	} else if diff < 90 {
		state.LunationPhase = "First Quarter"
	} else if diff < 135 {
		state.LunationPhase = "Waxing Gibbous"
	} else if diff < 180 {
		state.LunationPhase = "Full Moon approaching"
	} else if diff < 225 {
		state.LunationPhase = "Full Moon (waning)"
	} else if diff < 270 {
		state.LunationPhase = "Last Quarter"
	} else if diff < 315 {
		state.LunationPhase = "Waning Crescent"
	} else {
		state.LunationPhase = "New Moon approaching"
	}

	// Patterns in the event chart
	patterns := mundane.ChartPatterns(eventChart, *orbDeg)
	for _, p := range patterns.Patterns {
		state.Patterns = append(state.Patterns,
			fmt.Sprintf("%s: %s", p.Name, strings.Join(p.Planets, ", ")))
	}

	// Notable: flag tight transits, eclipse proximity, etc.
	for _, t := range state.Transits {
		if t.Orb < 0 {
			t.Orb = -t.Orb
		}
		if t.Orb < 0.5 {
			state.Notable = append(state.Notable,
				fmt.Sprintf("TIGHT: transiting %s %s natal %s (orb %.2f°)",
					t.Planet, t.Aspect, t.NatalPoint, t.Orb))
		}
	}
	for _, e := range state.RecentEclipses {
		if strings.Contains(e, date.Format("2006-01-02")) || 
		   strings.Contains(e, date.AddDate(0, 0, -1).Format("2006-01-02")) ||
		   strings.Contains(e, date.AddDate(0, 0, -2).Format("2006-01-02")) {
			state.Notable = append(state.Notable, "ECLIPSE: within 2 days of event")
		}
	}
	// Saturn-Pluto conjunction proximity
	for _, c := range state.RecentConjunctions {
		if strings.Contains(c, "Saturn") && strings.Contains(c, "Pluto") {
			state.Notable = append(state.Notable, "SATURN-PLUTO: conjunction in recent window")
		}
	}

	if *jsonOut {
		data, _ := json.MarshalIndent(state, "", "  ")
		fmt.Println(string(data))
		return
	}

	// Text output
	fmt.Printf("╔══════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║  ASTRO-STATE: %s", *dateStr)
	if *event != "" {
		fmt.Printf(" — %s", *event)
	}
	fmt.Printf("\n")
	fmt.Printf("║  Nation: %s\n", *nation)
	fmt.Printf("╚══════════════════════════════════════════════════════════╝\n\n")

	fmt.Printf("LUNATION: %s\n\n", state.LunationPhase)

	if len(state.Transits) > 0 {
		fmt.Printf("TRANSITS TO NATAL (orb ≤ %.1f°):\n", *orbDeg)
		for _, t := range state.Transits {
			fmt.Printf("  %-8s %-20s → natal %s (orb %.2f°)\n", t.Planet, t.Aspect, t.NatalPoint, t.Orb)
		}
		fmt.Println()
	}

	if len(state.RecentIngresses) > 0 {
		fmt.Printf("RECENT INGRESSES (last %d days):\n", *lookback)
		for _, s := range state.RecentIngresses {
			fmt.Printf("  %s\n", s)
		}
		fmt.Println()
	}

	if len(state.RecentEclipses) > 0 {
		fmt.Printf("RECENT ECLIPSES (last %d days):\n", *lookback)
		for _, s := range state.RecentEclipses {
			fmt.Printf("  %s\n", s)
		}
		fmt.Println()
	}

	if len(state.RecentConjunctions) > 0 {
		fmt.Printf("RECENT MAJOR CONJUNCTIONS (last %d days):\n", *lookback)
		for _, s := range state.RecentConjunctions {
			fmt.Printf("  %s\n", s)
		}
		fmt.Println()
	}

	if len(state.Patterns) > 0 {
		fmt.Printf("CHART PATTERNS:\n")
		for _, s := range state.Patterns {
			fmt.Printf("  %s\n", s)
		}
		fmt.Println()
	}

	if len(state.Notable) > 0 {
		fmt.Printf("NOTABLE:\n")
		for _, s := range state.Notable {
			fmt.Printf("  ⚡ %s\n", s)
		}
		fmt.Println()
	}
}
