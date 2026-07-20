package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/aj-nt/empirical/internal/dignity"
	"github.com/aj-nt/empirical/internal/swe"
)

func main() {
	ephePath := os.Getenv("SWE_EPHE_PATH")
	if ephePath == "" {
		ephePath = "/Users/aj/Documents/repos/koine/ephe"
	}
	swe.SetEphePath(ephePath)

	utHour := 23.0 + 10.0/60.0 + 8.0
	jd := swe.Julday(1969, 2, 15, utHour, true)

	// Compute planets
	planetSpecs := []struct {
		name string
		id   int
	}{
		{"Sun", swe.SUN}, {"Moon", swe.MOON}, {"Mercury", swe.MERCURY},
		{"Venus", swe.VENUS}, {"Mars", swe.MARS}, {"Jupiter", swe.JUPITER},
		{"Saturn", swe.SATURN}, {"Uranus", swe.URANUS}, {"Neptune", swe.NEPTUNE},
		{"Pluto", swe.PLUTO},
	}

	planets := make(map[string]float64)
	for _, p := range planetSpecs {
		lon, _, _, _ := swe.CalcUT(jd, p.id)
		planets[p.name] = lon
	}
	nnLon, _, _, _ := swe.CalcUT(jd, swe.MEAN_NODE)
	planets["Node"] = nnLon

	// Compute fixed stars
	stars := make(map[string]float64)
	for _, starName := range dignity.StarNames {
		lon, _, _, _ := swe.Fixstar(starName, jd)
		if lon != 0 {
			stars[starName] = lon
		}
	}

	// Detect patterns with stars
	report := dignity.DetectPatternsWithStars(planets, stars, dignity.DefaultPatternOrb)

	// Sort by kind
	sort.Slice(report.Patterns, func(i, j int) bool {
		return report.Patterns[i].Kind < report.Patterns[j].Kind
	})

	fmt.Println("=== DETECTED PATTERNS (with stars) ===")
	for _, p := range report.Patterns {
		fmt.Printf("Kind: %s | Name: %s | Planets: %v\n", p.Kind, p.Name, p.Planets)
		if p.Description != "" {
			fmt.Printf("  %s\n", p.Description)
		}
		for _, a := range p.Aspects {
			fmt.Printf("  %s %s %s (orb: %.2f°)\n", a.Planet1, a.Aspect, a.Planet2, a.Orb)
		}
		fmt.Println()
	}
}
