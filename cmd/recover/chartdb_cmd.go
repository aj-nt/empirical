package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/aj-nt/empirical"
)

// chartEntry mirrors the JSON structure in charts/people.json.
type chartEntry struct {
	Name     string `json:"name"`
	Birth    struct {
		Year     int     `json:"year"`
		Month    int     `json:"month"`
		Day      int     `json:"day"`
		Hour     int     `json:"hour"`
		Minute   int     `json:"minute"`
		TZOffset float64 `json:"tz_offset"`
		Lat      float64 `json:"lat"`
		Lng      float64 `json:"lng"`
	} `json:"birth"`
	Source   string `json:"source"`
	Category string `json:"category"`
	Notes    string `json:"notes"`
}

// runChartDB handles the "chartdb" subcommand.
func runChartDB(fs *flag.FlagSet, jsonOut *bool) {
	fs.Parse(os.Args[2:])

	action := "list"
	if fs.NArg() > 0 {
		action = fs.Arg(0)
	}

	var db []chartEntry
	if err := json.Unmarshal(empirical.ChartDBData, &db); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse chart database: %v\n", err)
		os.Exit(1)
	}

	switch action {
	case "list":
		chartDBList(db, *jsonOut)
	case "show":
		c := chartDBFind(db, fs, "show")
		chartDBShow(c, *jsonOut)
	case "recover":
		c := chartDBFind(db, fs, "recover")
		chartDBRecover(c, *jsonOut)
	case "search":
		chartDBSearch(db, fs, *jsonOut)
	default:
		fmt.Fprintf(os.Stderr, "Usage: empirical chartdb [list|show|recover|search] [args]\n")
		os.Exit(1)
	}
}

// chartDBFind finds a chart by name substring. Exits on missing arg or not found.
func chartDBFind(db []chartEntry, fs *flag.FlagSet, subcmd string) chartEntry {
	if fs.NArg() < 2 {
		fmt.Fprintf(os.Stderr, "Usage: empirical chartdb %s <name>\n", subcmd)
		os.Exit(1)
	}
	query := strings.ToLower(fs.Arg(1))
	for _, c := range db {
		if strings.Contains(strings.ToLower(c.Name), query) {
			return c
		}
	}
	fmt.Fprintf(os.Stderr, "Chart %q not found\n", query)
	os.Exit(1)
	panic("unreachable")
}

func chartDBList(db []chartEntry, jsonOut bool) {
	if jsonOut {
		js, _ := json.Marshal(db)
		fmt.Println(string(js))
		return
	}
	fmt.Printf("%-30s %-12s %-10s %s\n", "Name", "Category", "Rodden", "Birth Date")
	fmt.Println(strings.Repeat("-", 80))
	for _, c := range db {
		fmt.Printf("%-30s %-12s %-10s %04d-%02d-%02d\n",
			c.Name, c.Category, c.Source, c.Birth.Year, c.Birth.Month, c.Birth.Day)
	}
	fmt.Printf("\n%d charts\n", len(db))
}

func chartDBShow(c chartEntry, jsonOut bool) {
	if jsonOut {
		js, _ := json.Marshal(c)
		fmt.Println(string(js))
		return
	}
	fmt.Printf("Name:     %s\n", c.Name)
	fmt.Printf("Birth:    %04d-%02d-%02d %02d:%02d (UTC%+g)\n",
		c.Birth.Year, c.Birth.Month, c.Birth.Day, c.Birth.Hour, c.Birth.Minute, c.Birth.TZOffset)
	fmt.Printf("Location: %.3f, %.3f\n", c.Birth.Lat, c.Birth.Lng)
	fmt.Printf("Source:   %s\n", c.Source)
	fmt.Printf("Category: %s\n", c.Category)
	if c.Notes != "" {
		fmt.Printf("Notes:    %s\n", c.Notes)
	}
}

func chartDBRecover(c chartEntry, jsonOut bool) {
	result := computeAll(c.Name, c.Birth.Year, c.Birth.Month, c.Birth.Day,
		c.Birth.Hour, c.Birth.Minute, 0, c.Birth.TZOffset, c.Birth.Lat, c.Birth.Lng, "")
	if jsonOut {
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

func chartDBSearch(db []chartEntry, fs *flag.FlagSet, jsonOut bool) {
	if fs.NArg() < 2 {
		fmt.Fprintf(os.Stderr, "Usage: empirical chartdb search <query>\n")
		os.Exit(1)
	}
	query := strings.ToLower(fs.Arg(1))
	var matches []chartEntry
	for _, c := range db {
		if strings.Contains(strings.ToLower(c.Name), query) ||
			strings.Contains(strings.ToLower(c.Category), query) ||
			strings.Contains(strings.ToLower(c.Notes), query) {
			matches = append(matches, c)
		}
	}
	if jsonOut {
		js, _ := json.Marshal(matches)
		fmt.Println(string(js))
		return
	}
	if len(matches) == 0 {
		fmt.Printf("No charts match %q\n", query)
		return
	}
	for _, c := range matches {
		fmt.Printf("%-30s %-12s %s\n", c.Name, c.Category, c.Source)
	}
	fmt.Printf("\n%d match(es)\n", len(matches))
}
