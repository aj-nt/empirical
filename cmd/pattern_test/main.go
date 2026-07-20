package main

import (
	"fmt"
	"os"

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

	specs := []struct {
		name string
		id   int
	}{
		{"Sun", swe.SUN}, {"Moon", swe.MOON}, {"Mercury", swe.MERCURY},
		{"Venus", swe.VENUS}, {"Mars", swe.MARS}, {"Jupiter", swe.JUPITER},
		{"Saturn", swe.SATURN}, {"Uranus", swe.URANUS}, {"Neptune", swe.NEPTUNE},
		{"Pluto", swe.PLUTO},
	}
	fmt.Println("=== PLANETS ===")
	for _, p := range specs {
		lon, _, _, _ := swe.CalcUT(jd, p.id)
		sign := dignity.SignForLongitude(lon)
		fmt.Printf("%-10s %8.2f°  %s\n", p.name, lon, sign)
	}

	nnLon, _, _, _ := swe.CalcUT(jd, swe.MEAN_NODE)
	fmt.Printf("%-10s %8.2f°  %s\n", "Node", nnLon, dignity.SignForLongitude(nnLon))

	fmt.Println("\n=== FIXED STARS (first 15) ===")
	count := 0
	for _, starName := range dignity.StarNames {
		lon, _, _, _ := swe.Fixstar(starName, jd)
		if lon != 0 && count < 15 {
			sign := dignity.SignForLongitude(lon)
			fmt.Printf("%-20s %8.2f°  %s\n", starName, lon, sign)
			count++
		}
	}
}
