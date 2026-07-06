package main

import (
	"fmt"
	"github.com/aj-nt/empirical/internal/swe"
)

func main() {
	swe.SetEphePath("ephe")
	jd := swe.Julday(1969, 2, 15, 23.0+10.0/60.0+8.0, true) // UT
	lon, lat, _, mag := swe.Fixstar("Regulus", jd)
	fmt.Printf("Regulus: lon=%.6f, lat=%.6f, mag=%.2f\n", lon, lat, mag)

	for _, name := range []string{"Sirius", "Spica", "Aldebaran", "Antares"} {
		lon, _, _, _ := swe.Fixstar(name, jd)
		fmt.Printf("%s: lon=%.6f\n", name, lon)
	}
}
