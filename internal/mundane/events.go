package mundane

import (
	"fmt"
	"math"
	"time"

	"github.com/aj-nt/empirical/internal/dignity"
)

// ComputeFunc computes a planet's ecliptic longitude for a given date and
// planet ID. Returns (longitude, latitude, distance, speed).
type ComputeFunc func(year, month, day int, hour float64, planetID int) (lon, lat, dist, speed float64)

// IngressEvent records a planet entering a zodiac sign at a specific time.
type IngressEvent struct {
	Planet string    `json:"planet"`
	Sign   string    `json:"sign"`
	Time   time.Time `json:"time"`
	Lon    float64   `json:"lon"` // exact longitude at ingress
}



// cardinalBoundary returns the longitude of the cardinal sign boundary at the
// given sign index. Only returns true for Aries(0), Cancer(90), Libra(180),
// Capricorn(270).
func cardinalBoundary(idx int) (float64, bool) {
	if idx < 0 || idx >= len(dignity.Signs) {
		return 0, false
	}
	if dignity.CardinalSigns[dignity.Signs[idx]] {
		return float64(idx) * 30.0, true
	}
	return 0, false
}

// FindSolarIngresses finds all times the Sun enters a cardinal sign within the
// given date range. Uses binary search to pinpoint the exact ingress time.
//
// Temporal sampling assumption: this function samples the Sun's position at
// noon each day. The Sun moves ~1°/day, so a 24-hour step is sufficient —
// the Sun cannot cross a 30° sign boundary and return within one sample.
// For faster planets (Mercury ~1.3°/day, Moon ~13°/day), a 24-hour step
// could miss a crossing. Use a smaller step (e.g., 6h) for inner planets.
func FindSolarIngresses(start, end time.Time, compute ComputeFunc) ([]IngressEvent, error) {
	if start.After(end) {
		return nil, fmt.Errorf("start date after end date")
	}

	var ingresses []IngressEvent

	// Scan day by day
	current := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location())
	endDay := time.Date(end.Year(), end.Month(), end.Day(), 0, 0, 0, 0, end.Location())

	// Get initial sign
	prevLon, _, _, _ := compute(current.Year(), int(current.Month()), current.Day(), 12.0, 0)
	prevSign := dignity.SignIndex(prevLon)

	current = current.AddDate(0, 0, 1)
	for !current.After(endDay) {
		lon, _, _, _ := compute(current.Year(), int(current.Month()), current.Day(), 12.0, 0)
		sign := dignity.SignIndex(lon)

		if sign != prevSign {
			// Sign boundary crossed. Check if it's a cardinal ingress.
			boundary, isCardinal := cardinalBoundary(sign)
			if isCardinal {
				// Binary search within a 48-hour window centered on the crossing.
				// The sign change was detected between noon D-1 and noon D, so
				// use [D-1 00:00, D+1 00:00] to guarantee the target is bracketed.
				prevDay := current.AddDate(0, 0, -1)
				nextDay := current.AddDate(0, 0, 1)
				exactTime, exactLon, err := findCrossingTime(prevDay, nextDay, boundary, 0, compute)
				if err != nil {
					return nil, fmt.Errorf("binary search failed for %s ingress: %w", dignity.Signs[sign], err)
				}
				ingresses = append(ingresses, IngressEvent{
					Planet: "Sun",
					Sign:   dignity.Signs[sign],
					Time:   exactTime,
					Lon:    exactLon,
				})
			}
		}

		prevSign = sign
		current = current.AddDate(0, 0, 1)
	}

	return ingresses, nil
}

// findCrossingTime uses binary search to find the exact time a planet crosses
// a target longitude between t1 and t2. The planet must be on opposite sides
// of the target at t1 and t2.
func findCrossingTime(t1, t2 time.Time, targetLon float64, planetID int, compute ComputeFunc) (time.Time, float64, error) {
	// Get longitudes at endpoints
	lon1, _, _, _ := compute(t1.Year(), int(t1.Month()), t1.Day(), float64(t1.Hour())+float64(t1.Minute())/60.0+float64(t1.Second())/3600.0, planetID)
	lon2, _, _, _ := compute(t2.Year(), int(t2.Month()), t2.Day(), float64(t2.Hour())+float64(t2.Minute())/60.0+float64(t2.Second())/3600.0, planetID)

	// Normalize to [0, 360)
	lon1 = math.Mod(lon1, 360)
	if lon1 < 0 {
		lon1 += 360
	}
	lon2 = math.Mod(lon2, 360)
	if lon2 < 0 {
		lon2 += 360
	}

	// Determine direction: forward if lon2 is ahead of lon1 by < 180°
	diff := lon2 - lon1
	if diff < 0 {
		diff += 360
	}
	forward := diff < 180

	// Check if the target is between lon1 and lon2 in the direction of motion
	var crosses bool
	if forward {
		crosses = crossesTarget(lon1, lon2, targetLon)
	} else {
		crosses = crossesTarget(lon2, lon1, targetLon)
	}
	if !crosses {
		return time.Time{}, 0, fmt.Errorf("planet does not cross target %.1f between %v and %v (lon1=%.4f, lon2=%.4f, forward=%v)", targetLon, t1, t2, lon1, lon2, forward)
	}

	// Binary search for 20 iterations (sub-second precision over 24h)
	lo, hi := t1, t2
	for i := 0; i < 20; i++ {
		mid := lo.Add(hi.Sub(lo) / 2)
		midHour := float64(mid.Hour()) + float64(mid.Minute())/60.0 + float64(mid.Second())/3600.0
		midLon, _, _, _ := compute(mid.Year(), int(mid.Month()), mid.Day(), midHour, planetID)
		midLon = math.Mod(midLon, 360)
		if midLon < 0 {
			midLon += 360
		}

		var midCrosses bool
		if forward {
			midCrosses = crossesTarget(lon1, midLon, targetLon)
		} else {
			midCrosses = crossesTarget(midLon, lon1, targetLon)
		}

		if midCrosses {
			hi = mid
			lon2 = midLon
		} else {
			lo = mid
			lon1 = midLon
		}
	}

	// Return the midpoint as the best estimate
	result := lo.Add(hi.Sub(lo) / 2)
	resultHour := float64(result.Hour()) + float64(result.Minute())/60.0 + float64(result.Second())/3600.0
	resultLon, _, _, _ := compute(result.Year(), int(result.Month()), result.Day(), resultHour, planetID)
	resultLon = math.Mod(resultLon, 360)
	if resultLon < 0 {
		resultLon += 360
	}

	return result, resultLon, nil
}

// crossesTarget returns true if targetLon lies between lon1 and lon2 in the
// forward (increasing longitude) direction, accounting for 0° wrap.
func crossesTarget(lon1, lon2, target float64) bool {
	// Normalize all to [0, 360)
	lon1 = math.Mod(lon1, 360)
	if lon1 < 0 {
		lon1 += 360
	}
	lon2 = math.Mod(lon2, 360)
	if lon2 < 0 {
		lon2 += 360
	}
	target = math.Mod(target, 360)
	if target < 0 {
		target += 360
	}

	if lon1 <= lon2 {
		return lon1 <= target && target <= lon2
	}
	// Wraps around 0°
	return lon1 <= target || target <= lon2
}

// LunationEvent records a new or full moon at a specific time.
type LunationEvent struct {
	Type string    `json:"type"` // "New Moon" or "Full Moon"
	Time time.Time `json:"time"`
}

// FindLunations finds all new and full moons within the given date range.
// Uses daily scanning + binary search for exact times.
func FindLunations(start, end time.Time, compute ComputeFunc) ([]LunationEvent, error) {
	if start.After(end) {
		return nil, fmt.Errorf("start date after end date")
	}

	var lunations []LunationEvent

	current := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location())
	endDay := time.Date(end.Year(), end.Month(), end.Day(), 0, 0, 0, 0, end.Location())

	// Get initial Sun-Moon angle
	prevAngle := sunMoonAngle(current, 12.0, compute)

	current = current.AddDate(0, 0, 1)
	for !current.After(endDay) {
		angle := sunMoonAngle(current, 12.0, compute)

		// Check for new moon (crosses 0°/360°)
		if crossesAngle(prevAngle, angle, 0) {
			prevDay := current.AddDate(0, 0, -1)
			nextDay := current.AddDate(0, 0, 1)
			exactTime, err := findLunationTime(prevDay, nextDay, 0, compute)
			if err != nil {
				return nil, fmt.Errorf("new moon search failed: %w", err)
			}
			lunations = append(lunations, LunationEvent{
				Type: "New Moon",
				Time: exactTime,
			})
		}

		// Check for full moon (crosses 180°)
		if crossesAngle(prevAngle, angle, 180) {
			prevDay := current.AddDate(0, 0, -1)
			nextDay := current.AddDate(0, 0, 1)
			exactTime, err := findLunationTime(prevDay, nextDay, 180, compute)
			if err != nil {
				return nil, fmt.Errorf("full moon search failed: %w", err)
			}
			lunations = append(lunations, LunationEvent{
				Type: "Full Moon",
				Time: exactTime,
			})
		}

		prevAngle = angle
		current = current.AddDate(0, 0, 1)
	}

	return lunations, nil
}

// sunMoonAngle returns the angular difference (Moon - Sun) normalized to [0, 360).
func sunMoonAngle(t time.Time, hour float64, compute ComputeFunc) float64 {
	sunLon, _, _, _ := compute(t.Year(), int(t.Month()), t.Day(), hour, 0)
	moonLon, _, _, _ := compute(t.Year(), int(t.Month()), t.Day(), hour, 1)
	diff := math.Mod(moonLon-sunLon, 360)
	if diff < 0 {
		diff += 360
	}
	return diff
}

// crossesAngle returns true if target (0 or 180) lies between a and b in the
// forward (increasing) direction, accounting for 0° wrap.
func crossesAngle(a, b, target float64) bool {
	a = math.Mod(a, 360)
	if a < 0 {
		a += 360
	}
	b = math.Mod(b, 360)
	if b < 0 {
		b += 360
	}
	target = math.Mod(target, 360)
	if target < 0 {
		target += 360
	}

	if a <= b {
		return a <= target && target <= b
	}
	return a <= target || target <= b
}

// findLunationTime uses binary search to find when the Sun-Moon angle equals
// the target (0 for new moon, 180 for full moon).
func findLunationTime(t1, t2 time.Time, target float64, compute ComputeFunc) (time.Time, error) {
	angle1 := sunMoonAngle(t1, float64(t1.Hour())+float64(t1.Minute())/60.0+float64(t1.Second())/3600.0, compute)
	angle2 := sunMoonAngle(t2, float64(t2.Hour())+float64(t2.Minute())/60.0+float64(t2.Second())/3600.0, compute)

	if !crossesAngle(angle1, angle2, target) {
		return time.Time{}, fmt.Errorf("Sun-Moon angle does not cross %.0f between %v and %v (a1=%.4f, a2=%.4f)", target, t1, t2, angle1, angle2)
	}

	lo, hi := t1, t2
	for i := 0; i < 20; i++ {
		mid := lo.Add(hi.Sub(lo) / 2)
		midHour := float64(mid.Hour()) + float64(mid.Minute())/60.0 + float64(mid.Second())/3600.0
		midAngle := sunMoonAngle(mid, midHour, compute)

		if crossesAngle(angle1, midAngle, target) {
			hi = mid
			angle2 = midAngle
		} else {
			lo = mid
			angle1 = midAngle
		}
	}

	return lo.Add(hi.Sub(lo) / 2), nil
}

// EclipseEvent records a solar or lunar eclipse.
type EclipseEvent struct {
	Type string    `json:"type"` // "Solar Eclipse" or "Lunar Eclipse"
	Time time.Time `json:"time"`
}

// EclipseNodeOrb is the maximum angular distance from the lunar nodes for an
// eclipse to occur. Standard value is ~15° (partial eclipses can occur up to
// ~18°, but 15° is the traditional threshold for notable eclipses).
const EclipseNodeOrb = 15.0

// FindEclipses finds all solar and lunar eclipses within the given date range.
// An eclipse occurs when a new or full moon happens within EclipseNodeOrb of
// the lunar nodes.
func FindEclipses(start, end time.Time, compute ComputeFunc) ([]EclipseEvent, error) {
	if start.After(end) {
		return nil, fmt.Errorf("start date after end date")
	}

	var eclipses []EclipseEvent

	current := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location())
	endDay := time.Date(end.Year(), end.Month(), end.Day(), 0, 0, 0, 0, end.Location())

	// Get initial Sun-Moon angle and node distance
	prevAngle := sunMoonAngle(current, 12.0, compute)

	current = current.AddDate(0, 0, 1)
	for !current.After(endDay) {
		angle := sunMoonAngle(current, 12.0, compute)

		// Check for new moon (crosses 0°)
		if crossesAngle(prevAngle, angle, 0) {
			prevDay := current.AddDate(0, 0, -1)
			nextDay := current.AddDate(0, 0, 1)
			exactTime, err := findLunationTime(prevDay, nextDay, 0, compute)
			if err != nil {
				return nil, fmt.Errorf("solar eclipse search failed: %w", err)
			}
			// Check node proximity
			if isNearNode(exactTime, compute) {
				eclipses = append(eclipses, EclipseEvent{
					Type: "Solar Eclipse",
					Time: exactTime,
				})
			}
		}

		// Check for full moon (crosses 180°)
		if crossesAngle(prevAngle, angle, 180) {
			prevDay := current.AddDate(0, 0, -1)
			nextDay := current.AddDate(0, 0, 1)
			exactTime, err := findLunationTime(prevDay, nextDay, 180, compute)
			if err != nil {
				return nil, fmt.Errorf("lunar eclipse search failed: %w", err)
			}
			// Check node proximity
			if isNearNode(exactTime, compute) {
				eclipses = append(eclipses, EclipseEvent{
					Type: "Lunar Eclipse",
					Time: exactTime,
				})
			}
		}

		prevAngle = angle
		current = current.AddDate(0, 0, 1)
	}

	return eclipses, nil
}

// isNearNode returns true if the Sun (for solar eclipses) or the Sun-Moon
// opposition midpoint is within EclipseNodeOrb of the North Node.
// For new moons: check Sun-Node distance (Sun ≈ Moon at conjunction).
// For full moons: check Sun-Node distance (opposition midpoint is 180° from Sun,
// and Node is 180° from South Node, so Sun-Node distance works for both).
func isNearNode(t time.Time, compute ComputeFunc) bool {
	hour := float64(t.Hour()) + float64(t.Minute())/60.0 + float64(t.Second())/3600.0
	sunLon, _, _, _ := compute(t.Year(), int(t.Month()), t.Day(), hour, 0)
	nodeLon, _, _, _ := compute(t.Year(), int(t.Month()), t.Day(), hour, 10) // MEAN_NODE

	// Angular distance between Sun and Node
	diff := math.Mod(math.Abs(sunLon-nodeLon), 360)
	if diff > 180 {
		diff = 360 - diff
	}
	// Also check South Node (Node + 180°)
	southNode := math.Mod(nodeLon+180, 360)
	diff2 := math.Mod(math.Abs(sunLon-southNode), 360)
	if diff2 > 180 {
		diff2 = 360 - diff2
	}
	if diff2 < diff {
		diff = diff2
	}

	return diff <= EclipseNodeOrb
}

// FindPlanetaryIngresses finds all times a planet enters a new zodiac sign
// within the given date range. Works for any planet, including retrograde motion.
func FindPlanetaryIngresses(start, end time.Time, planetID int, planetName string, compute ComputeFunc) ([]IngressEvent, error) {
	if start.After(end) {
		return nil, fmt.Errorf("start date after end date")
	}

	var ingresses []IngressEvent

	current := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location())
	endDay := time.Date(end.Year(), end.Month(), end.Day(), 0, 0, 0, 0, end.Location())

	// Get initial sign
	prevLon, _, _, _ := compute(current.Year(), int(current.Month()), current.Day(), 12.0, planetID)
	prevSign := dignity.SignIndex(prevLon)

	current = current.AddDate(0, 0, 1)
	for !current.After(endDay) {
		lon, _, _, _ := compute(current.Year(), int(current.Month()), current.Day(), 12.0, planetID)
		sign := dignity.SignIndex(lon)

		if sign != prevSign {
			// Sign boundary crossed. Determine which boundary.
			// If moving forward: entered sign at sign*30.
			// If moving backward (retrograde): entered prevSign at prevSign*30.
			var boundary float64
			var enteredSign string

			// Check direction: get positions at start and end of the 24h window
			prevDay := current.AddDate(0, 0, -1)
			nextDay := current.AddDate(0, 0, 1)
			lonBefore, _, _, _ := compute(prevDay.Year(), int(prevDay.Month()), prevDay.Day(), 0.0, planetID)
			lonAfter, _, _, _ := compute(nextDay.Year(), int(nextDay.Month()), nextDay.Day(), 0.0, planetID)

			lonBefore = math.Mod(lonBefore, 360)
			if lonBefore < 0 {
				lonBefore += 360
			}
			lonAfter = math.Mod(lonAfter, 360)
			if lonAfter < 0 {
				lonAfter += 360
			}

			// Determine direction: if lonAfter > lonBefore (accounting for wrap), forward
			forward := true
			diff := lonAfter - lonBefore
			if diff < 0 {
				diff += 360
			}
			if diff > 180 {
				forward = false
			}

			if forward {
				boundary = float64(sign * 30)
				enteredSign = dignity.Signs[sign]
			} else {
				boundary = float64(prevSign * 30)
				enteredSign = dignity.Signs[sign] // moving backward into the new (lower-indexed) sign
			}

			exactTime, exactLon, err := findCrossingTime(prevDay, nextDay, boundary, planetID, compute)
			if err != nil {
				return nil, fmt.Errorf("binary search failed for %s ingress into %s: %w", planetName, enteredSign, err)
			}
			ingresses = append(ingresses, IngressEvent{
				Planet: planetName,
				Sign:   enteredSign,
				Time:   exactTime,
				Lon:    exactLon,
			})
		}

		prevSign = sign
		current = current.AddDate(0, 0, 1)
	}

	return ingresses, nil
}

// ConjunctionEvent records a conjunction between two planets at a specific time.
type ConjunctionEvent struct {
	Planet1 string    `json:"planet1"`
	Planet2 string    `json:"planet2"`
	Time    time.Time `json:"time"`
}

// FindConjunctions finds all conjunctions (0° aspect) between two planets
// within the given date range. Uses daily scanning + binary search.
// Handles both forward and backward relative motion.
func FindConjunctions(start, end time.Time, p1ID int, p1Name string, p2ID int, p2Name string, compute ComputeFunc) ([]ConjunctionEvent, error) {
	if start.After(end) {
		return nil, fmt.Errorf("start date after end date")
	}

	var conjunctions []ConjunctionEvent

	current := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location())
	endDay := time.Date(end.Year(), end.Month(), end.Day(), 0, 0, 0, 0, end.Location())

	// Get initial angular difference (p2 - p1 normalized to [0, 360))
	prevDiff := planetAngleDiff(current, 12.0, p1ID, p2ID, compute)

	current = current.AddDate(0, 0, 1)
	for !current.After(endDay) {
		diff := planetAngleDiff(current, 12.0, p1ID, p2ID, compute)

		// Determine direction of angular change
		delta := diff - prevDiff
		if delta < 0 {
			delta += 360
		}
		forward := delta < 180

		// Check if the difference crossed 0° in the direction of motion
		var crossed bool
		if forward {
			crossed = crossesAngle(prevDiff, diff, 0)
		} else {
			crossed = crossesAngle(diff, prevDiff, 0)
		}

		if crossed {
			prevDay := current.AddDate(0, 0, -1)
			nextDay := current.AddDate(0, 0, 1)
			exactTime, err := findConjunctionTime(prevDay, nextDay, p1ID, p2ID, forward, compute)
			if err != nil {
				return nil, fmt.Errorf("conjunction search failed for %s-%s: %w", p1Name, p2Name, err)
			}
			conjunctions = append(conjunctions, ConjunctionEvent{
				Planet1: p1Name,
				Planet2: p2Name,
				Time:    exactTime,
			})
		}

		prevDiff = diff
		current = current.AddDate(0, 0, 1)
	}

	return conjunctions, nil
}

// planetAngleDiff returns the angular difference (p2 - p1) normalized to [0, 360).
func planetAngleDiff(t time.Time, hour float64, p1ID, p2ID int, compute ComputeFunc) float64 {
	p1Lon, _, _, _ := compute(t.Year(), int(t.Month()), t.Day(), hour, p1ID)
	p2Lon, _, _, _ := compute(t.Year(), int(t.Month()), t.Day(), hour, p2ID)
	diff := math.Mod(p2Lon-p1Lon, 360)
	if diff < 0 {
		diff += 360
	}
	return diff
}

// findConjunctionTime uses binary search to find when two planets are exactly
// conjunct (0° angular separation). forward indicates the direction of angular
// change (true = diff increasing, false = diff decreasing).
func findConjunctionTime(t1, t2 time.Time, p1ID, p2ID int, forward bool, compute ComputeFunc) (time.Time, error) {
	diff1 := planetAngleDiff(t1, float64(t1.Hour())+float64(t1.Minute())/60.0+float64(t1.Second())/3600.0, p1ID, p2ID, compute)
	diff2 := planetAngleDiff(t2, float64(t2.Hour())+float64(t2.Minute())/60.0+float64(t2.Second())/3600.0, p1ID, p2ID, compute)

	var crosses bool
	if forward {
		crosses = crossesAngle(diff1, diff2, 0)
	} else {
		crosses = crossesAngle(diff2, diff1, 0)
	}
	if !crosses {
		return time.Time{}, fmt.Errorf("planets do not conjunct between %v and %v (d1=%.4f, d2=%.4f, forward=%v)", t1, t2, diff1, diff2, forward)
	}

	lo, hi := t1, t2
	for i := 0; i < 20; i++ {
		mid := lo.Add(hi.Sub(lo) / 2)
		midHour := float64(mid.Hour()) + float64(mid.Minute())/60.0 + float64(mid.Second())/3600.0
		midDiff := planetAngleDiff(mid, midHour, p1ID, p2ID, compute)

		var midCrosses bool
		if forward {
			midCrosses = crossesAngle(diff1, midDiff, 0)
		} else {
			midCrosses = crossesAngle(midDiff, diff1, 0)
		}

		if midCrosses {
			hi = mid
			diff2 = midDiff
		} else {
			lo = mid
			diff1 = midDiff
		}
	}

	return lo.Add(hi.Sub(lo) / 2), nil
}

// HousesFunc computes house cusps and angles for a given JD and location.
// hsys is the house system: 'W' = Whole Sign, 'P' = Placidus, etc.
// Returns 13 cusps (index 1-12) and 10 ascmc values (ASC=0, MC=1).
