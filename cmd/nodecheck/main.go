package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/aj-nt/empirical"
	"github.com/aj-nt/empirical/internal/swe"
)

func main() {
	if len(os.Args) < 4 {
		fmt.Fprintf(os.Stderr, "Usage: nodecheck Y M D\n")
		os.Exit(1)
	}
	year, _ := strconv.Atoi(os.Args[1])
	month, _ := strconv.Atoi(os.Args[2])
	day, _ := strconv.Atoi(os.Args[3])

	cacheDir, err := empirical.EnsureEpheCache()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ephe error: %v\n", err)
		os.Exit(1)
	}
	swe.SetEphePath(cacheDir)

	// Compute Node at midnight UT
	jd := swe.Julday(year, month, day, 0, true)
	lon, _, _, _, err := swe.CalcUTErr(jd, 11) // 11 = True Node
	if err != nil {
		fmt.Fprintf(os.Stderr, "calc error: %v\n", err)
		os.Exit(1)
	}

	sign := ""
	signs := []string{"Aries", "Taurus", "Gemini", "Cancer", "Leo", "Virgo",
		"Libra", "Scorpio", "Sagittarius", "Capricorn", "Aquarius", "Pisces"}
	si := int(lon/30) % 12
	sign = signs[si]
	signLon := lon - float64(si)*30

	// Natal Sun: 327.4456° = 27°27' Aquarius
	natalSun := 327.4456
	orb := lon - natalSun
	if orb < -180 {
		orb += 360
	}
	if orb > 180 {
		orb -= 360
	}
	absOrb := orb
	if absOrb < 0 {
		absOrb = -absOrb
	}

	fmt.Printf("Transiting True Node on %04d-%02d-%02d:\n", year, month, day)
	fmt.Printf("  Longitude: %.4f° (%d°%.0f' %s)\n", lon, int(signLon), (signLon-float64(int(signLon)))*60, sign)
	fmt.Printf("  Natal Sun: 327.4456° (27°27' Aquarius)\n")
	fmt.Printf("  Orb: %.2f°\n", orb)
	if absOrb < 1.0 {
		fmt.Printf("  VERDICT: YES — conjunction within 1° orb\n")
	} else if absOrb < 3.0 {
		fmt.Printf("  VERDICT: Approaching — within %.2f° orb, but not tight conjunction\n", absOrb)
	} else {
		fmt.Printf("  VERDICT: NO — %.2f° away, not a conjunction\n", absOrb)
	}
}
