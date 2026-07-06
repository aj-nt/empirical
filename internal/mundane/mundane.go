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

// Signs is the ordered list of zodiac signs.
var Signs = []string{
	"Aries", "Taurus", "Gemini", "Cancer", "Leo", "Virgo",
	"Libra", "Scorpio", "Sagittarius", "Capricorn", "Aquarius", "Pisces",
}

// CardinalSigns are the four cardinal signs (equinoxes and solstices).
var CardinalSigns = map[string]bool{
	"Aries": true, "Cancer": true, "Libra": true, "Capricorn": true,
}

// signIndex returns the zodiac sign index (0-11) for a given ecliptic longitude.
func signIndex(lon float64) int {
	lon = math.Mod(lon, 360)
	if lon < 0 {
		lon += 360
	}
	return int(lon / 30)
}

// signName returns the zodiac sign name for a given ecliptic longitude.
func signName(lon float64) string {
	return Signs[signIndex(lon)]
}

// cardinalBoundary returns the longitude of the cardinal sign boundary at the
// given sign index. Only returns true for Aries(0), Cancer(90), Libra(180),
// Capricorn(270).
func cardinalBoundary(idx int) (float64, bool) {
	switch idx {
	case 0:
		return 0, true
	case 3:
		return 90, true
	case 6:
		return 180, true
	case 9:
		return 270, true
	}
	return 0, false
}

// FindSolarIngresses finds all times the Sun enters a cardinal sign within the
// given date range. Uses binary search to pinpoint the exact ingress time.
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
	prevSign := signIndex(prevLon)

	current = current.AddDate(0, 0, 1)
	for !current.After(endDay) {
		lon, _, _, _ := compute(current.Year(), int(current.Month()), current.Day(), 12.0, 0)
		sign := signIndex(lon)

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
					return nil, fmt.Errorf("binary search failed for %s ingress: %w", Signs[sign], err)
				}
				ingresses = append(ingresses, IngressEvent{
					Planet: "Sun",
					Sign:   Signs[sign],
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
	prevSign := signIndex(prevLon)

	current = current.AddDate(0, 0, 1)
	for !current.After(endDay) {
		lon, _, _, _ := compute(current.Year(), int(current.Month()), current.Day(), 12.0, planetID)
		sign := signIndex(lon)

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
				enteredSign = Signs[sign]
			} else {
				boundary = float64(prevSign * 30)
				enteredSign = Signs[sign] // moving backward into the new (lower-indexed) sign
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
type HousesFunc func(jd, lat, lon float64, hsys byte) (cusps [13]float64, ascmc [10]float64)

// MundaneChart is a full astrological chart cast for a specific time and location.
type MundaneChart struct {
	Time    time.Time         `json:"time"`
	Lat     float64           `json:"lat"`
	Lon     float64           `json:"lon"`
	Planets map[string]float64 `json:"planets"`
	Houses  [12]float64       `json:"houses"` // cusps 1-12
	ASC     float64           `json:"asc"`
	MC      float64           `json:"mc"`
}

// DefaultMundanePlanets returns the planet set used for mundane charts.
// 10 bodies: Sun through Pluto plus North Node.
func DefaultMundanePlanets() []struct {
	Name string
	ID   int
} {
	return []struct {
		Name string
		ID   int
	}{
		{"Sun", 0},
		{"Moon", 1},
		{"Mercury", 2},
		{"Venus", 3},
		{"Mars", 4},
		{"Jupiter", 5},
		{"Saturn", 6},
		{"Uranus", 7},
		{"Neptune", 8},
		{"Pluto", 9},
		{"Node", 10},
	}
}

// julianDay computes the Julian Day from a time.Time using the standard
// astronomical formula. Pure Go, no CGo dependency.
func julianDay(t time.Time) float64 {
	y, m, day := t.Year(), int(t.Month()), t.Day()
	h := float64(t.Hour()) + float64(t.Minute())/60.0 + float64(t.Second())/3600.0

	if m <= 2 {
		y--
		m += 12
	}
	a := y / 100
	b := 2 - a + a/4
	jd := float64(int(365.25*float64(y+4716))) +
		float64(int(30.6001*float64(m+1))) +
		float64(day) + h/24.0 + float64(b) - 1524.5
	return jd
}

// CastChart computes a full mundane chart for the given time and location.
// compute provides planet positions. houses provides house cusps and angles.
// hsys is the house system (e.g., 'W' for Whole Sign, 'P' for Placidus).
func CastChart(tm time.Time, lat, lon float64, compute ComputeFunc, houses HousesFunc, hsys byte) (*MundaneChart, error) {
	jd := julianDay(tm)

	// Compute planet positions
	planets := make(map[string]float64)
	for _, p := range DefaultMundanePlanets() {
		plon, _, _, _ := compute(tm.Year(), int(tm.Month()), tm.Day(),
			float64(tm.Hour())+float64(tm.Minute())/60.0+float64(tm.Second())/3600.0,
			p.ID)
		planets[p.Name] = plon
	}

	// Compute houses
	cusps, ascmc := houses(jd, lat, lon, hsys)

	// Copy cusps 1-12
	var houseCusps [12]float64
	for i := 0; i < 12; i++ {
		houseCusps[i] = cusps[i+1]
	}

	return &MundaneChart{
		Time:    tm,
		Lat:     lat,
		Lon:     lon,
		Planets: planets,
		Houses:  houseCusps,
		ASC:     ascmc[0],
		MC:      ascmc[1],
	}, nil
}

// CastIngressChart casts a chart for an ingress event at the given location.
// Uses Whole Sign houses by default.
func CastIngressChart(event IngressEvent, lat, lon float64, compute ComputeFunc, houses HousesFunc) (*MundaneChart, error) {
	return CastChart(event.Time, lat, lon, compute, houses, 'W')
}

// CastLunationChart casts a chart for a lunation event at the given location.
// Uses Whole Sign houses by default.
func CastLunationChart(event LunationEvent, lat, lon float64, compute ComputeFunc, houses HousesFunc) (*MundaneChart, error) {
	return CastChart(event.Time, lat, lon, compute, houses, 'W')
}

// NationalChartEntry holds the birth data for a nation's chart.
type NationalChartEntry struct {
	Name      string  `json:"name"`
	Year      int     `json:"year"`
	Month     int     `json:"month"`
	Day       int     `json:"day"`
	Hour      float64 `json:"hour"`      // UT
	Lat       float64 `json:"lat"`
	Lon       float64 `json:"lon"`
	Timezone  string  `json:"timezone"`  // for display only
	Note      string  `json:"note"`      // source or variant info
}

// nationalCharts is the database of national charts.
// Times are in UT. Sources are noted where multiple charts exist.
var nationalCharts = []NationalChartEntry{
	{
		Name: "United States",
		Year: 1776, Month: 7, Day: 4,
		Hour: 22.2167, // ~5:13 PM local Philadelphia = 22:13 UT (Sibly chart)
		Lat: 39.95, Lon: -75.15,
		Timezone: "LMT",
		Note: "Sibly chart (Declaration of Independence). Alternative: Gemini rising chart (2:21 AM).",
	},
	{
		Name: "United Kingdom",
		Year: 1801, Month: 1, Day: 1,
		Hour: 0.0, // midnight, Act of Union
		Lat: 51.5, Lon: -0.1,
		Timezone: "LMT",
		Note: "Act of Union 1801. Alternative: 1066 Norman Conquest chart.",
	},
	{
		Name: "China",
		Year: 1949, Month: 10, Day: 1,
		Hour: 7.0, // ~3 PM Beijing = 7:00 UT
		Lat: 39.9, Lon: 116.4,
		Timezone: "CST",
		Note: "PRC founding. Time approximate (ceremony ~3 PM).",
	},
	{
		Name: "Russia",
		Year: 1991, Month: 12, Day: 25,
		Hour: 16.0, // ~7 PM Moscow = 16:00 UT
		Lat: 55.75, Lon: 37.6,
		Timezone: "MSK",
		Note: "Dissolution of USSR / Russian Federation. Alternative: 1990 declaration of sovereignty.",
	},
	{
		Name: "India",
		Year: 1947, Month: 8, Day: 15,
		Hour: 0.0, // midnight
		Lat: 28.6, Lon: 77.2,
		Timezone: "IST",
		Note: "Independence. Midnight chart is standard.",
	},
	{
		Name: "Japan",
		Year: 1947, Month: 5, Day: 3,
		Hour: 0.0, // midnight, constitution effective
		Lat: 35.7, Lon: 139.7,
		Timezone: "JST",
		Note: "Post-war constitution effective date.",
	},
	{
		Name: "Germany",
		Year: 1990, Month: 10, Day: 3,
		Hour: 0.0, // midnight, reunification
		Lat: 52.5, Lon: 13.4,
		Timezone: "CET",
		Note: "Reunification. Alternative: 1949 Federal Republic chart.",
	},
	{
		Name: "France",
		Year: 1958, Month: 10, Day: 5,
		Hour: 0.0, // midnight, Fifth Republic
		Lat: 48.85, Lon: 2.35,
		Timezone: "CET",
		Note: "Fifth Republic. Alternative: 1789 Revolution chart.",
	},
	{
		Name: "European Union",
		Year: 1993, Month: 11, Day: 1,
		Hour: 0.0, // midnight, Maastricht Treaty
		Lat: 50.85, Lon: 4.35,
		Timezone: "CET",
		Note: "Maastricht Treaty effective. Chart cast for Brussels.",
	},
	{
		Name: "Israel",
		Year: 1948, Month: 5, Day: 14,
		Hour: 14.0, // ~4 PM Tel Aviv = 14:00 UT
		Lat: 32.1, Lon: 34.8,
		Timezone: "IST",
		Note: "Declaration of Independence. Time approximate (~4 PM).",
	},
	{
		Name: "Brazil",
		Year: 1822, Month: 9, Day: 7,
		Hour: 16.0, // ~1 PM local = ~16:00 UT
		Lat: -23.55, Lon: -46.63,
		Timezone: "LMT",
		Note: "Independence declared. Time approximate.",
	},
	{
		Name: "United Nations",
		Year: 1945, Month: 10, Day: 24,
		Hour: 16.0, // charter ratified ~4 PM UT
		Lat: 40.75, Lon: -73.97,
		Timezone: "EST",
		Note: "UN Charter entered into force. Chart for NYC.",
	},
	{
		Name: "Australia",
		Year: 1901, Month: 1, Day: 1,
		Hour: 0.0, // midnight, Federation
		Lat: -35.3, Lon: 149.1,
		Timezone: "AEST",
		Note: "Federation. Chart for Canberra (Sydney at federation, Canberra for modern).",
	},
	{
		Name: "Canada",
		Year: 1867, Month: 7, Day: 1,
		Hour: 0.0, // midnight, Confederation
		Lat: 45.4, Lon: -75.7,
		Timezone: "LMT",
		Note: "Confederation / Dominion of Canada. Chart for Ottawa.",
	},
	{
		Name: "Iran",
		Year: 1979, Month: 4, Day: 1,
		Hour: 0.0, // Islamic Republic referendum result
		Lat: 35.7, Lon: 51.4,
		Timezone: "IRST",
		Note: "Islamic Republic established. Alternative: 1979 Feb 11 revolution victory.",
	},
	{
		Name: "Saudi Arabia",
		Year: 1932, Month: 9, Day: 23,
		Hour: 0.0, // unification
		Lat: 24.7, Lon: 46.7,
		Timezone: "AST",
		Note: "Kingdom unified. Chart for Riyadh.",
	},
	{
		Name: "Turkey",
		Year: 1923, Month: 10, Day: 29,
		Hour: 0.0, // Republic proclaimed
		Lat: 39.9, Lon: 32.9,
		Timezone: "EET",
		Note: "Republic of Turkey proclaimed. Chart for Ankara.",
	},
	{
		Name: "South Korea",
		Year: 1948, Month: 8, Day: 15,
		Hour: 0.0, // Republic established
		Lat: 37.6, Lon: 127.0,
		Timezone: "KST",
		Note: "Republic of Korea established. Chart for Seoul.",
	},
	{
		Name: "North Korea",
		Year: 1948, Month: 9, Day: 9,
		Hour: 0.0, // DPRK founded
		Lat: 39.0, Lon: 125.8,
		Timezone: "KST",
		Note: "DPRK founded. Chart for Pyongyang.",
	},
	{
		Name: "Mexico",
		Year: 1821, Month: 9, Day: 27,
		Hour: 0.0, // independence recognized
		Lat: 19.4, Lon: -99.1,
		Timezone: "LMT",
		Note: "Independence recognized. Chart for Mexico City.",
	},
	{
		Name: "Italy",
		Year: 1946, Month: 6, Day: 2,
		Hour: 0.0, // Republic referendum
		Lat: 41.9, Lon: 12.5,
		Timezone: "CET",
		Note: "Republic established. Chart for Rome.",
	},
	{
		Name: "Spain",
		Year: 1978, Month: 12, Day: 29,
		Hour: 0.0, // constitution effective
		Lat: 40.4, Lon: -3.7,
		Timezone: "CET",
		Note: "Constitution of 1978 effective. Chart for Madrid.",
	},
	{
		Name: "Ukraine",
		Year: 1991, Month: 8, Day: 24,
		Hour: 0.0, // independence declared
		Lat: 50.5, Lon: 30.5,
		Timezone: "EET",
		Note: "Independence declared. Chart for Kyiv.",
	},
	{
		Name: "Pakistan",
		Year: 1947, Month: 8, Day: 14,
		Hour: 0.0, // independence
		Lat: 33.7, Lon: 73.0,
		Timezone: "PKT",
		Note: "Independence. Chart for Islamabad (Karachi at independence).",
	},
	{
		Name: "Indonesia",
		Year: 1945, Month: 8, Day: 17,
		Hour: 0.0, // independence proclaimed
		Lat: -6.2, Lon: 106.8,
		Timezone: "WIB",
		Note: "Independence proclaimed. Chart for Jakarta.",
	},
	{
		Name: "Nigeria",
		Year: 1960, Month: 10, Day: 1,
		Hour: 0.0, // independence
		Lat: 9.1, Lon: 7.5,
		Timezone: "WAT",
		Note: "Independence. Chart for Abuja (Lagos at independence).",
	},
	{
		Name: "South Africa",
		Year: 1994, Month: 4, Day: 27,
		Hour: 0.0, // first democratic election
		Lat: -25.7, Lon: 28.2,
		Timezone: "SAST",
		Note: "Post-apartheid democracy. Chart for Pretoria.",
	},
	{
		Name: "Egypt",
		Year: 1952, Month: 7, Day: 23,
		Hour: 0.0, // revolution
		Lat: 30.0, Lon: 31.2,
		Timezone: "EET",
		Note: "1952 Revolution / Republic. Chart for Cairo.",
	},
	{
		Name: "Argentina",
		Year: 1816, Month: 7, Day: 9,
		Hour: 0.0, // independence declared
		Lat: -34.6, Lon: -58.4,
		Timezone: "LMT",
		Note: "Independence declared. Chart for Buenos Aires.",
	},
	{
		Name: "Switzerland",
		Year: 1848, Month: 9, Day: 12,
		Hour: 0.0, // federal constitution
		Lat: 46.9, Lon: 7.4,
		Timezone: "LMT",
		Note: "Federal constitution. Chart for Bern.",
	},
	{
		Name: "Sweden",
		Year: 1523, Month: 6, Day: 6,
		Hour: 0.0, // Gustav Vasa elected king
		Lat: 59.3, Lon: 18.1,
		Timezone: "LMT",
		Note: "Gustav Vasa elected king — foundation of modern Sweden. Chart for Stockholm.",
	},
	{
		Name: "Vatican City",
		Year: 1929, Month: 2, Day: 11,
		Hour: 0.0, // Lateran Treaty
		Lat: 41.9, Lon: 12.5,
		Timezone: "CET",
		Note: "Lateran Treaty — Vatican City State established. Chart for Vatican City.",
	},
	{
		Name: "Greece",
		Year: 1822, Month: 1, Day: 1,
		Hour: 0.0, // independence declared
		Lat: 38.0, Lon: 23.7,
		Timezone: "LMT",
		Note: "Independence declared. Chart for Athens.",
	},
	{
		Name: "Poland",
		Year: 1989, Month: 6, Day: 4,
		Hour: 0.0, // first free elections
		Lat: 52.2, Lon: 21.0,
		Timezone: "CET",
		Note: "First free elections — end of communist rule. Chart for Warsaw.",
	},
	{
		Name: "Taiwan",
		Year: 1912, Month: 1, Day: 1,
		Hour: 0.0, // Republic of China founding
		Lat: 25.0, Lon: 121.5,
		Timezone: "CST",
		Note: "Republic of China founding (same date as PRC claims). Chart for Taipei.",
	},
	{
		Name: "Venezuela",
		Year: 1811, Month: 7, Day: 5,
		Hour: 0.0, // independence declared
		Lat: 10.5, Lon: -66.9,
		Timezone: "LMT",
		Note: "Independence declared. Chart for Caracas.",
	},
	{
		Name: "Philippines",
		Year: 1946, Month: 7, Day: 4,
		Hour: 0.0, // independence from US
		Lat: 14.6, Lon: 121.0,
		Timezone: "PHT",
		Note: "Independence from US. Chart for Manila.",
	},
	{
		Name: "Thailand",
		Year: 1932, Month: 6, Day: 24,
		Hour: 0.0, // constitutional monarchy
		Lat: 13.8, Lon: 100.5,
		Timezone: "ICT",
		Note: "Constitutional monarchy established. Chart for Bangkok.",
	},
	{
		Name: "Vietnam",
		Year: 1976, Month: 7, Day: 2,
		Hour: 0.0, // reunification
		Lat: 21.0, Lon: 105.8,
		Timezone: "ICT",
		Note: "Reunification. Chart for Hanoi.",
	},
	{
		Name: "Ireland",
		Year: 1922, Month: 12, Day: 6,
		Hour: 0.0, // Irish Free State
		Lat: 53.3, Lon: -6.3,
		Timezone: "GMT",
		Note: "Irish Free State established. Chart for Dublin.",
	},
	{
		Name: "Netherlands",
		Year: 1815, Month: 3, Day: 16,
		Hour: 0.0, // Kingdom established
		Lat: 52.1, Lon: 4.3,
		Timezone: "LMT",
		Note: "Kingdom of the Netherlands. Chart for The Hague.",
	},
	{
		Name: "Belgium",
		Year: 1830, Month: 10, Day: 4,
		Hour: 0.0, // independence declared
		Lat: 50.9, Lon: 4.4,
		Timezone: "LMT",
		Note: "Independence declared. Chart for Brussels.",
	},
	{
		Name: "Norway",
		Year: 1905, Month: 6, Day: 7,
		Hour: 0.0, // independence from Sweden
		Lat: 59.9, Lon: 10.8,
		Timezone: "CET",
		Note: "Independence from Sweden. Chart for Oslo.",
	},
	{
		Name: "Denmark",
		Year: 1849, Month: 6, Day: 5,
		Hour: 0.0, // constitution signed
		Lat: 55.7, Lon: 12.6,
		Timezone: "LMT",
		Note: "Constitutional monarchy. Chart for Copenhagen.",
	},
	{
		Name: "Finland",
		Year: 1917, Month: 12, Day: 6,
		Hour: 0.0, // independence declared
		Lat: 60.2, Lon: 24.9,
		Timezone: "EET",
		Note: "Independence declared. Chart for Helsinki.",
	},
	{
		Name: "Austria",
		Year: 1955, Month: 5, Day: 15,
		Hour: 0.0, // State Treaty
		Lat: 48.2, Lon: 16.4,
		Timezone: "CET",
		Note: "State Treaty — restored sovereignty. Chart for Vienna.",
	},
	{
		Name: "Portugal",
		Year: 1974, Month: 4, Day: 25,
		Hour: 0.0, // Carnation Revolution
		Lat: 38.7, Lon: -9.1,
		Timezone: "WET",
		Note: "Carnation Revolution — end of Estado Novo. Chart for Lisbon.",
	},
	{
		Name: "Colombia",
		Year: 1810, Month: 7, Day: 20,
		Hour: 0.0, // independence declared
		Lat: 4.7, Lon: -74.1,
		Timezone: "LMT",
		Note: "Independence declared. Chart for Bogotá.",
	},
	{
		Name: "Chile",
		Year: 1818, Month: 2, Day: 12,
		Hour: 0.0, // independence declared
		Lat: -33.4, Lon: -70.7,
		Timezone: "LMT",
		Note: "Independence declared. Chart for Santiago.",
	},
	{
		Name: "Peru",
		Year: 1821, Month: 7, Day: 28,
		Hour: 0.0, // independence declared
		Lat: -12.0, Lon: -77.0,
		Timezone: "LMT",
		Note: "Independence declared. Chart for Lima.",
	},
	{
		Name: "New Zealand",
		Year: 1907, Month: 9, Day: 26,
		Hour: 0.0, // Dominion status
		Lat: -41.3, Lon: 174.8,
		Timezone: "NZST",
		Note: "Dominion status. Chart for Wellington.",
	},
	{
		Name: "Singapore",
		Year: 1965, Month: 8, Day: 9,
		Hour: 0.0, // independence
		Lat: 1.3, Lon: 103.8,
		Timezone: "SGT",
		Note: "Independence from Malaysia. Chart for Singapore.",
	},
	{
		Name: "Malaysia",
		Year: 1957, Month: 8, Day: 31,
		Hour: 0.0, // independence
		Lat: 3.1, Lon: 101.7,
		Timezone: "MYT",
		Note: "Independence from UK. Chart for Kuala Lumpur.",
	},
	{
		Name: "Bangladesh",
		Year: 1971, Month: 3, Day: 26,
		Hour: 0.0, // independence declared
		Lat: 23.8, Lon: 90.4,
		Timezone: "BST",
		Note: "Independence declared. Chart for Dhaka.",
	},
	{
		Name: "Ethiopia",
		Year: 1991, Month: 5, Day: 28,
		Hour: 0.0, // Derg regime falls
		Lat: 9.0, Lon: 38.7,
		Timezone: "EAT",
		Note: "Derg regime falls — modern Ethiopia. Chart for Addis Ababa.",
	},
	{
		Name: "Kenya",
		Year: 1963, Month: 12, Day: 12,
		Hour: 0.0, // independence
		Lat: -1.3, Lon: 36.8,
		Timezone: "EAT",
		Note: "Independence. Chart for Nairobi.",
	},
	{
		Name: "Cuba",
		Year: 1959, Month: 1, Day: 1,
		Hour: 0.0, // revolution victory
		Lat: 23.1, Lon: -82.4,
		Timezone: "CST",
		Note: "Revolution victory. Chart for Havana.",
	},
	{
		Name: "Iraq",
		Year: 1932, Month: 10, Day: 3,
		Hour: 0.0, // independence
		Lat: 33.3, Lon: 44.4,
		Timezone: "AST",
		Note: "Independence from British mandate. Chart for Baghdad.",
	},
	{
		Name: "Afghanistan",
		Year: 1919, Month: 8, Day: 19,
		Hour: 0.0, // independence
		Lat: 34.5, Lon: 69.2,
		Timezone: "AFT",
		Note: "Independence from British control. Chart for Kabul.",
	},
	{
		Name: "Syria",
		Year: 1946, Month: 4, Day: 17,
		Hour: 0.0, // independence
		Lat: 33.5, Lon: 36.3,
		Timezone: "EET",
		Note: "Independence from French mandate. Chart for Damascus.",
	},
	{
		Name: "Myanmar",
		Year: 1948, Month: 1, Day: 4,
		Hour: 0.0, // independence
		Lat: 16.8, Lon: 96.2,
		Timezone: "MMT",
		Note: "Independence from UK. Chart for Yangon.",
	},
	{
		Name: "Sudan",
		Year: 1956, Month: 1, Day: 1,
		Hour: 0.0, // independence
		Lat: 15.5, Lon: 32.6,
		Timezone: "CAT",
		Note: "Independence. Chart for Khartoum.",
	},
	{
		Name: "Ghana",
		Year: 1957, Month: 3, Day: 6,
		Hour: 0.0, // independence
		Lat: 5.6, Lon: -0.2,
		Timezone: "GMT",
		Note: "First sub-Saharan African independence. Chart for Accra.",
	},
	{
		Name: "Algeria",
		Year: 1962, Month: 7, Day: 5,
		Hour: 0.0, // independence
		Lat: 36.8, Lon: 3.0,
		Timezone: "CET",
		Note: "Independence from France. Chart for Algiers.",
	},
	{
		Name: "Morocco",
		Year: 1956, Month: 3, Day: 2,
		Hour: 0.0, // independence
		Lat: 34.0, Lon: -6.8,
		Timezone: "WET",
		Note: "Independence from France. Chart for Rabat.",
	},
	{
		Name: "Kazakhstan",
		Year: 1991, Month: 12, Day: 16,
		Hour: 0.0, // independence
		Lat: 51.2, Lon: 71.4,
		Timezone: "ALMT",
		Note: "Independence from USSR. Chart for Astana.",
	},
	{
		Name: "Uzbekistan",
		Year: 1991, Month: 9, Day: 1,
		Hour: 0.0, // independence
		Lat: 41.3, Lon: 69.3,
		Timezone: "UZT",
		Note: "Independence from USSR. Chart for Tashkent.",
	},
}

// NationalChart returns the chart entry for a given nation by name.
// Returns false if not found.
func NationalChart(name string) (NationalChartEntry, bool) {
	for _, c := range nationalCharts {
		if c.Name == name {
			return c, true
		}
	}
	return NationalChartEntry{}, false
}

// NationalCharts returns all national chart entries.
func NationalCharts() []NationalChartEntry {
	result := make([]NationalChartEntry, len(nationalCharts))
	copy(result, nationalCharts)
	return result
}

// ChartAspect records an aspect between two planets in a chart.
type ChartAspect struct {
	Planet1 string  `json:"planet1"`
	Planet2 string  `json:"planet2"`
	Aspect  string  `json:"aspect"`
	Orb     float64 `json:"orb"`
}

// defaultChartAspects returns the standard Ptolemaic aspects for chart analysis.
func defaultChartAspects() []struct {
	Angle float64
	Name  string
} {
	return []struct {
		Angle float64
		Name  string
	}{
		{0, "conjunction"},
		{60, "sextile"},
		{90, "square"},
		{120, "trine"},
		{180, "opposition"},
	}
}

// ChartAspects computes all aspects between planets in a MundaneChart.
// orbDeg is the maximum orb in degrees.
func ChartAspects(chart *MundaneChart, orbDeg float64) []ChartAspect {
	aspects := defaultChartAspects()
	planetNames := make([]string, 0, len(chart.Planets))
	for name := range chart.Planets {
		planetNames = append(planetNames, name)
	}

	var result []ChartAspect
	for i := 0; i < len(planetNames); i++ {
		for j := i + 1; j < len(planetNames); j++ {
			p1, p2 := planetNames[i], planetNames[j]
			dist := angleDist(chart.Planets[p1], chart.Planets[p2])
			for _, asp := range aspects {
				diff := dist - asp.Angle
				if diff < 0 {
					diff = -diff
				}
				if diff <= orbDeg {
					result = append(result, ChartAspect{
						Planet1: p1,
						Planet2: p2,
						Aspect:  asp.Name,
						Orb:     math.Round(diff*100) / 100,
					})
				}
			}
		}
	}
	return result
}

// angleDist returns the shortest angular distance between two longitudes.
func angleDist(a, b float64) float64 {
	d := math.Mod(math.Abs(a-b), 360)
	if d > 180 {
		d = 360 - d
	}
	return d
}

// NationTransits computes transits from current planets to a nation's natal
// chart over the given date range. Returns compacted transit hits (date ranges
// with closest orb). Uses the standard Ptolemaic aspects.
func NationTransits(nationName, startDate, endDate string, orbDeg float64, compute ComputeFunc, houses HousesFunc) ([]struct {
	TransitPlanet string
	NatalPlanet   string
	Aspect        string
	MinOrb        float64
	DateStart     string
	DateEnd       string
}, error) {
	entry, ok := NationalChart(nationName)
	if !ok {
		return nil, fmt.Errorf("unknown nation: %s", nationName)
	}

	// Cast the nation's natal chart
	natalTime := time.Date(entry.Year, time.Month(entry.Month), entry.Day,
		int(entry.Hour), int((entry.Hour-float64(int(entry.Hour)))*60), 0, 0, time.UTC)
	natalChart, err := CastChart(natalTime, entry.Lat, entry.Lon, compute, houses, 'W')
	if err != nil {
		return nil, fmt.Errorf("casting natal chart for %s: %w", nationName, err)
	}

	// Build natal planet map (name -> longitude)
	natalLongs := make(map[string]float64)
	natalPlanets := make([]string, 0, len(natalChart.Planets))
	for name, lon := range natalChart.Planets {
		natalLongs[name] = lon
		natalPlanets = append(natalPlanets, name)
	}

	// Parse dates
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start date: %w", err)
	}
	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return nil, fmt.Errorf("invalid end date: %w", err)
	}

	// Standard Ptolemaic aspects
	aspects := []struct {
		Angle float64
		Name  string
	}{
		{0, "conjunction"},
		{60, "sextile"},
		{90, "square"},
		{120, "trine"},
		{180, "opposition"},
	}

	// Transit planets: Sun through Pluto + Node
	transitPlanets := DefaultMundanePlanets()

	var hits []struct {
		Date          string
		TransitPlanet string
		NatalPlanet   string
		Aspect        string
		Orb           float64
	}

	current := start
	for !current.After(end) {
		y, m, d := current.Year(), int(current.Month()), current.Day()

		for _, tp := range transitPlanets {
			tLon, _, _, _ := compute(y, m, d, 12.0, tp.ID)
			for _, np := range natalPlanets {
				nLon, ok := natalLongs[np]
				if !ok {
					continue
				}
				dist := angleDist(tLon, nLon)
				for _, asp := range aspects {
					diff := dist - asp.Angle
					if diff < 0 {
						diff = -diff
					}
					if diff <= orbDeg {
						hits = append(hits, struct {
							Date          string
							TransitPlanet string
							NatalPlanet   string
							Aspect        string
							Orb           float64
						}{
							Date:          current.Format("2006-01-02"),
							TransitPlanet: tp.Name,
							NatalPlanet:   np,
							Aspect:        asp.Name,
							Orb:           math.Round(diff*100) / 100,
						})
					}
				}
			}
		}
		current = current.AddDate(0, 0, 1)
	}

	// Compact: group sequential days of the same transit
	type key struct {
		TransitPlanet string
		NatalPlanet   string
		Aspect        string
	}
	groups := make(map[key][]struct {
		Date          string
		TransitPlanet string
		NatalPlanet   string
		Aspect        string
		Orb           float64
	})
	for _, h := range hits {
		k := key{h.TransitPlanet, h.NatalPlanet, h.Aspect}
		groups[k] = append(groups[k], h)
	}

	var result []struct {
		TransitPlanet string
		NatalPlanet   string
		Aspect        string
		MinOrb        float64
		DateStart     string
		DateEnd       string
	}
	for k, group := range groups {
		best := group[0]
		for _, h := range group {
			if h.Orb < best.Orb {
				best = h
			}
		}
		result = append(result, struct {
			TransitPlanet string
			NatalPlanet   string
			Aspect        string
			MinOrb        float64
			DateStart     string
			DateEnd       string
		}{
			TransitPlanet: k.TransitPlanet,
			NatalPlanet:   k.NatalPlanet,
			Aspect:        k.Aspect,
			MinOrb:        best.Orb,
			DateStart:     group[0].Date,
			DateEnd:       group[len(group)-1].Date,
		})
	}

	return result, nil
}

// ChartPatterns delegates to the dignity engine's pattern detection on a
// MundaneChart's planet positions. Returns the full PatternReport.
func ChartPatterns(chart *MundaneChart, orbDeg float64) *dignity.PatternReport {
	return dignity.DetectPatterns(chart.Planets, orbDeg)
}

// PlanetHouses computes which house (1-12) each planet falls in.
// For Whole Sign houses, this is determined by the sign offset from ASC.
// For other house systems, planets are placed between cusps.
func PlanetHouses(chart *MundaneChart) map[string]int {
	result := make(map[string]int, len(chart.Planets))

	// Whole Sign: house = ((planet_sign_index - asc_sign_index + 12) % 12) + 1
	ascSign := int(chart.ASC / 30)

	for planet, lon := range chart.Planets {
		planetSign := int(lon / 30)
		house := ((planetSign - ascSign + 12) % 12) + 1
		result[planet] = house
	}

	return result
}

// InterpretMundaneChart produces a full chart interpretation from a MundaneChart,
// bridging to the dignity engine's InterpretChart.
func InterpretMundaneChart(name string, chart *MundaneChart, orbDeg float64) *dignity.ChartInterpretation {
	// Planet → house
	houses := PlanetHouses(chart)

	// Aspects: convert ChartAspect → dignity.AspectHit
	chartAspects := ChartAspects(chart, orbDeg)
	aspectHits := make([]dignity.AspectHit, len(chartAspects))
	for i, a := range chartAspects {
		aspectHits[i] = dignity.AspectHit{
			Planet1: a.Planet1,
			Planet2: a.Planet2,
			Aspect:  a.Aspect,
			Orb:     a.Orb,
		}
	}

	// Patterns: convert Pattern → dignity.PatternHit
	report := ChartPatterns(chart, orbDeg)
	patternHits := make([]dignity.PatternHit, len(report.Patterns))
	for i, p := range report.Patterns {
		patternHits[i] = dignity.PatternHit{
			Name:    p.Name,
			Planets: p.Planets,
		}
	}

	return dignity.InterpretChart(name, chart.Planets, houses, aspectHits, patternHits, nil)
}
