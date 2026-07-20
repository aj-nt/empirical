package main

import (
	"fmt"
	"math"

	"github.com/aj-nt/empirical/internal/swe"
)

func main() {
	swe.SetEphePath("/Users/aj/.local/share/ephe")

	// Cait's natal Moon
	jdBirth := swe.Julday(1986, 4, 29, 7.0, true)
	moonLon, _, _, _ := swe.CalcUT(jdBirth, swe.MOON)
	fmt.Printf("Cait natal Moon: %.4f°\n", moonLon)

	orbDeg := 3.0
	targetLon := math.Mod(moonLon+90.0, 360.0)
	ingressLon := math.Mod(targetLon-orbDeg+360.0, 360.0)
	egressLon := math.Mod(targetLon+orbDeg, 360.0)

	fmt.Printf("Target (Moon+90°): %.4f°\n", targetLon)
	fmt.Printf("Ingress boundary:  %.4f°\n", ingressLon)
	fmt.Printf("Egress boundary:   %.4f°\n", egressLon)

	// ── Binary search for ingress/egress ──
	// The aspect distance function: how far is Saturn from exact square to Moon?
	aspectDist := func(jd float64) float64 {
		lon, _, _, _ := swe.CalcUT(jd, swe.SATURN)
		return angleDist(lon, targetLon)
	}

	// Scan window: 2026-07-10 to 2026-07-12
	jdScanStart := swe.Julday(2026, 7, 10, 12.0, true)
	jdScanEnd := swe.Julday(2026, 7, 12, 12.0, true)

	fmt.Println("\n── Saturn square Moon: binary search for boundaries ──")

	// Verify we're in orb at scan start
	dist := aspectDist(jdScanStart)
	fmt.Printf("At scan start (2026-07-10): dist = %.4f° (in orb: %v)\n", dist, dist <= orbDeg)

	// Find ingress: binary search backward from scan start
	// We need a point where we're OUT of orb, then binary search toward scan start
	// Saturn moves ~0.028 deg/day, so to go from 0.15° orb to 3.0° orb:
	// need ~ (3.0 - 0.15) / 0.028 ≈ 102 days backward
	jdFarBack := jdScanStart - 200 // ~6.5 months before
	distFarBack := aspectDist(jdFarBack)
	fmt.Printf("200 days before scan: dist = %.4f° (in orb: %v)\n", distFarBack, distFarBack <= orbDeg)

	// Binary search for ingress (the JD where dist crosses orbDeg from above)
	jdIngress := binarySearchBoundary(aspectDist, jdFarBack, jdScanStart, orbDeg, true)
	if jdIngress > 0 {
		y, m, d, h := swe.Revjul(jdIngress)
		dist := aspectDist(jdIngress)
		fmt.Printf("Ingress: %04d-%02d-%02d %.2fh UT, dist=%.4f°\n", y, m, d, h, dist)
	}

	// Find egress: binary search forward from scan end
	jdFarAhead := jdScanEnd + 200
	distFarAhead := aspectDist(jdFarAhead)
	fmt.Printf("200 days after scan: dist = %.4f° (in orb: %v)\n", distFarAhead, distFarAhead <= orbDeg)

	jdEgress := binarySearchBoundary(aspectDist, jdScanEnd, jdFarAhead, orbDeg, false)
	if jdEgress > 0 {
		y, m, d, h := swe.Revjul(jdEgress)
		dist := aspectDist(jdEgress)
		fmt.Printf("Egress:  %04d-%02d-%02d %.2fh UT, dist=%.4f°\n", y, m, d, h, dist)
	}

	// Find peak: binary search for minimum distance between ingress and egress
	if jdIngress > 0 && jdEgress > 0 {
		jdPeak := binarySearchPeak(aspectDist, jdIngress, jdEgress)
		y, m, d, h := swe.Revjul(jdPeak)
		dist := aspectDist(jdPeak)
		fmt.Printf("Peak:    %04d-%02d-%02d %.2fh UT, dist=%.4f°\n", y, m, d, h, dist)
		fmt.Printf("Duration: %.1f days\n", jdEgress-jdIngress)
	}

	// ── Day-by-day scan for comparison ──
	fmt.Println("\n── Day-by-day scan (2026-07-09 to 2026-07-13) ──")
	for day := 9; day <= 13; day++ {
		jd := swe.Julday(2026, 7, day, 12.0, true)
		lon, _, _, _ := swe.CalcUT(jd, swe.SATURN)
		dist := angleDist(lon, targetLon)
		inOrb := "out"
		if dist <= orbDeg {
			inOrb = "IN"
		}
		fmt.Printf("  2026-07-%02d: Saturn %.4f°, dist %.4f° [%s]\n", day, lon, dist, inOrb)
	}

	// ── Mercury retrograde: find all ingress/egress pairs ──
	fmt.Println("\n── Mercury square Moon: all ingress/egress pairs (2026) ──")
	mercTarget := math.Mod(moonLon+90.0, 360.0)
	mercDist := func(jd float64) float64 {
		lon, _, _, _ := swe.CalcUT(jd, swe.MERCURY)
		return angleDist(lon, mercTarget)
	}

	// Scan 2026 day by day at 1-day resolution to find all in-orb periods
	jdMerc := swe.Julday(2026, 1, 1, 12.0, true)
	jdMercEnd := swe.Julday(2027, 1, 1, 12.0, true)
	inOrb := false
	var ingressJDs []float64
	var egressJDs []float64

	for jd := jdMerc; jd <= jdMercEnd; jd += 1.0 {
		d := mercDist(jd)
		if d <= orbDeg && !inOrb {
			inOrb = true
			ingressJDs = append(ingressJDs, jd)
		} else if d > orbDeg && inOrb {
			inOrb = false
			egressJDs = append(egressJDs, jd)
		}
	}
	if inOrb {
		egressJDs = append(egressJDs, jdMercEnd)
	}

	for i := 0; i < len(ingressJDs) && i < len(egressJDs); i++ {
		// Refine with binary search
		jdIn := binarySearchBoundary(mercDist, ingressJDs[i]-1, ingressJDs[i]+1, orbDeg, true)
		jdOut := binarySearchBoundary(mercDist, egressJDs[i]-1, egressJDs[i]+1, orbDeg, false)
		yIn, mIn, dIn, hIn := swe.Revjul(jdIn)
		yOut, mOut, dOut, hOut := swe.Revjul(jdOut)
		lonIn, _, _, speedIn := swe.CalcUT(jdIn, swe.MERCURY)
		fmt.Printf("  Pair %d: %04d-%02d-%02d %.1fh → %04d-%02d-%02d %.1fh (%.1f days, %s)\n",
			i+1, yIn, mIn, dIn, hIn, yOut, mOut, dOut, hOut, jdOut-jdIn, dirStr(speedIn))
		fmt.Printf("          In: %.4f°, Out: %.4f°\n", lonIn, mercDist(jdOut))
	}

	// ── Key findings ──
	fmt.Println("\n── Key findings ──")
	fmt.Println("1. swe_helio_cross_ut is HELIOCENTRIC — doesn't work for geocentric transits")
	fmt.Println("2. SWE has no built-in geocentric planet crossing function")
	fmt.Println("3. Binary search with swe_calc_ut works: ~30 iterations → sub-second precision")
	fmt.Println("4. For slow outer planets: one ingress/egress pair, straightforward")
	fmt.Println("5. For Mercury: multiple ingress/egress pairs per year (retrogrades)")
	fmt.Println("6. Day-by-day scan + binary refinement is the practical approach")
	fmt.Println("7. The current flat start_date/end_date model needs to become an array of pairs")
}

// binarySearchBoundary finds the JD where dist crosses the threshold.
// If ingress=true: finds where dist goes from >threshold to <=threshold.
// If ingress=false: finds where dist goes from <=threshold to >threshold.
// lo and hi must bracket the crossing (one side in orb, one side out).
func binarySearchBoundary(distFn func(float64) float64, lo, hi, threshold float64, ingress bool) float64 {
	for i := 0; i < 50; i++ {
		mid := (lo + hi) / 2.0
		d := distFn(mid)
		if ingress {
			// We want the first point where d <= threshold
			if d <= threshold {
				hi = mid // mid is in orb, boundary is to the left
			} else {
				lo = mid // mid is out of orb, boundary is to the right
			}
		} else {
			// We want the first point where d > threshold
			if d > threshold {
				hi = mid // mid is out of orb, boundary is to the left
			} else {
				lo = mid // mid is in orb, boundary is to the right
			}
		}
		if hi-lo < 1.0/86400.0 { // 1 second precision
			break
		}
	}
	return (lo + hi) / 2.0
}

// binarySearchPeak finds the JD where distFn is minimized (peak orb).
func binarySearchPeak(distFn func(float64) float64, lo, hi float64) float64 {
	// Golden section search
	phi := (math.Sqrt(5) - 1) / 2
	a, b := lo, hi
	c := b - phi*(b-a)
	d := a + phi*(b-a)

	for i := 0; i < 50; i++ {
		if distFn(c) < distFn(d) {
			b = d
		} else {
			a = c
		}
		c = b - phi*(b-a)
		d = a + phi*(b-a)
		if b-a < 1.0/86400.0 {
			break
		}
	}
	return (a + b) / 2.0
}

func dirStr(speed float64) string {
	if speed < 0 {
		return "retrograde"
	}
	return "direct"
}

func angleDist(a, b float64) float64 {
	d := math.Abs(a - b)
	if d > 180 {
		d = 360 - d
	}
	return d
}
