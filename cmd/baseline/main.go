package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strings"

	"github.com/aj-nt/empirical"
	"github.com/aj-nt/empirical/internal/dignity"
	"github.com/aj-nt/empirical/internal/swe"
)

func main() {
	nCharts := flag.Int("n", 3000, "number of random charts")
	seed := flag.Int64("seed", 42, "random seed")
	flag.Parse()

	cacheDir, err := empirical.EnsureEpheCache()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize ephemeris: %v\n", err)
		os.Exit(1)
	}
	swe.SetEphePath(cacheDir)
	swe.SetSidMode(swe.SIDM_LAHIRI, 0, 0)

	rng := rand.New(rand.NewSource(*seed))

	// Phase 3: house convergence distribution
	type houseResult struct {
		unambiguous int
		disputed    int
	}
	var houseResults []houseResult

	// Phase 5: node convergence
	var nodeSignMatches int
	var nodeAxisPreserved int

	// Phase 4: timing convergence
	type timingResult struct {
		count    int
		hasConv  bool
		allAgree bool
	}
	var timingResults []timingResult

	for i := 0; i < *nCharts; i++ {
		y := 1900 + rng.Intn(131)
		mo := 1 + rng.Intn(12)
		maxD := 31
		switch mo {
		case 2:
			isLeap := (y%4 == 0 && y%100 != 0) || (y%400 == 0)
			if isLeap {
				maxD = 29
			} else {
				maxD = 28
			}
		case 4, 6, 9, 11:
			maxD = 30
		}
		d := 1 + rng.Intn(maxD)
		h := rng.Intn(24)
		mi := rng.Intn(60)
		s := rng.Intn(60)
		tzOff := -12.0 + float64(rng.Intn(49))/2.0
		lat := -60.0 + rng.Float64()*120.0
		lng := -180.0 + rng.Float64()*360.0

		// Compute positions once for all phases
		utHour := float64(h) + float64(mi)/60.0 + float64(s)/3600.0 - tzOff
		jd := swe.Julday(y, mo, d, utHour, true)
		ayan := swe.GetAyanamsaUT(jd)

		tropicalLons := make(map[string]float64)
		planetSpecs := []struct {
			name string
			id   int
		}{
			{"Sun", swe.SUN},
			{"Moon", swe.MOON},
			{"Mercury", swe.MERCURY},
			{"Venus", swe.VENUS},
			{"Mars", swe.MARS},
			{"Jupiter", swe.JUPITER},
			{"Saturn", swe.SATURN},
		}
		for _, p := range planetSpecs {
			lon, _, _, _ := swe.CalcUT(jd, p.id)
			tropicalLons[p.name] = lon
		}

		nnLon, _, _, _ := swe.CalcUT(jd, swe.MEAN_NODE)
		_, ascmc := swe.Houses(jd, lat, lng, 'P')
		asc := ascmc[0]

		// Phase 3: House convergence
		hc := dignity.ComputeHouseConvergence(
			tropicalLons, y, mo, d, h, mi, s, tzOff, lat, lng, "",
		)
		houseResults = append(houseResults, houseResult{
			unambiguous: hc.UnambiguousCount(),
			disputed:    hc.DisputedCount(),
		})

		// Phase 5: Node convergence
		nc := dignity.ComputeNodeConvergence(nnLon, ayan, "")
		if nc.SignMatch {
			nodeSignMatches++
		}
		if nc.AxisPreserved() {
			nodeAxisPreserved++
		}

		// Phase 4: Timing convergence
		ageYears := rng.Intn(91)
		targetY := y + ageYears
		targetMo := mo
		targetD := d
		if targetD > 28 {
			targetD = 28
		}
		targetDate := fmt.Sprintf("%04d-%02d-%02d", targetY, targetMo, targetD)

		report := dignity.ComputeTimingReport(
			"", y, mo, d, h, mi, tzOff, lat, lng,
			targetDate, tropicalLons, ayan, asc,
		)
		tc := report.TimingConvergence

		timingResults = append(timingResults, timingResult{
			count:    tc.ConvergenceCount,
			hasConv:  tc.HasConvergence,
			allAgree: tc.AllSystemsAgree,
		})

		if (i+1)%500 == 0 {
			fmt.Fprintf(os.Stderr, "  %d/%d...\n", i+1, *nCharts)
		}
	}

	total := len(houseResults)

	// ── Phase 3 summary ──
	var houseSum float64
	houseDist := make(map[int]int)
	for _, r := range houseResults {
		houseSum += float64(r.unambiguous)
		houseDist[r.unambiguous]++
	}
	fmt.Printf("=== PHASE 3: HOUSE CONVERGENCE (Go) ===\n")
	fmt.Printf("  %d random charts\n", total)
	fmt.Printf("  Mean unambiguous planets: %.2f/7 (%.1f%%)\n", houseSum/float64(total), houseSum/float64(total)/7*100)
	fmt.Printf("  Distribution:\n")
	for k := 0; k <= 7; k++ {
		n := houseDist[k]
		if n > 0 {
			bar := strings.Repeat("#", max(1, n*60/total))
			fmt.Printf("    %d: %s  (%d, %.1f%%)\n", k, bar, n, float64(n)/float64(total)*100)
		}
	}

	// ── Phase 5 summary ──
	fmt.Printf("\n=== PHASE 5: NODE CONVERGENCE (Go) ===\n")
	fmt.Printf("  %d random charts\n", total)
	fmt.Printf("  Sign preserved: %d/%d (%.1f%%)\n", nodeSignMatches, total, float64(nodeSignMatches)/float64(total)*100)
	fmt.Printf("  Axis preserved: %d/%d (%.1f%%)\n", nodeAxisPreserved, total, float64(nodeAxisPreserved)/float64(total)*100)

	// ── Phase 4 summary ──
	var timingSum float64
	timingCountDist := make(map[int]int)
	timingHasConv := 0
	timingAllAgree := 0
	for _, r := range timingResults {
		timingSum += float64(r.count)
		timingCountDist[r.count]++
		if r.hasConv {
			timingHasConv++
		}
		if r.allAgree {
			timingAllAgree++
		}
	}
	fmt.Printf("\n=== PHASE 4: TIMING CONVERGENCE (Go) ===\n")
	fmt.Printf("  %d random charts + target dates\n", len(timingResults))
	fmt.Printf("  Partial (1+ converging): %d/%d (%.1f%%)\n", timingHasConv, len(timingResults), float64(timingHasConv)/float64(len(timingResults))*100)
	fmt.Printf("  Full (all 3 agree): %d/%d (%.1f%%)\n", timingAllAgree, len(timingResults), float64(timingAllAgree)/float64(len(timingResults))*100)
	fmt.Printf("  Mean converging planets: %.2f\n", timingSum/float64(len(timingResults)))
	fmt.Printf("  Distribution:\n")
	for k := 0; k <= 7; k++ {
		n := timingCountDist[k]
		if n > 0 {
			bar := strings.Repeat("#", max(1, n*60/len(timingResults)))
			fmt.Printf("    %d: %s  (%d, %.1f%%)\n", k, bar, n, float64(n)/float64(len(timingResults))*100)
		}
	}
}
