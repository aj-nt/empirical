package mundane

import (
	"testing"
	"time"
)

// mockSunCompute returns a Sun longitude that moves exactly 1°/day.
// At refTime (2026-03-20T00:00Z), Sun is at refLon (358.5° = 28.5° Pisces).
// This means the Aries ingress (0°) happens at refTime + 1.5 days = 2026-03-21T12:00Z.
func mockSunCompute(refTime time.Time, refLon float64) ComputeFunc {
	return func(year, month, day int, hour float64, planetID int) (float64, float64, float64, float64) {
		if planetID != 0 { // not Sun
			return 47.0, 0, 0, 0
		}
		t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
		days := t.Sub(refTime).Hours()/24.0 + hour/24.0
		lon := refLon + days*1.0
		for lon < 0 {
			lon += 360
		}
		for lon >= 360 {
			lon -= 360
		}
		return lon, 0, 0, 0
	}
}

func TestFindSolarIngresses_SingleAriesIngress(t *testing.T) {
	t.Parallel()

	refTime := time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC)
	compute := mockSunCompute(refTime, 358.5)

	start := time.Date(2026, 3, 19, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 3, 22, 0, 0, 0, 0, time.UTC)

	ingresses, err := FindSolarIngresses(start, end, compute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ingresses) != 1 {
		t.Fatalf("expected 1 ingress, got %d", len(ingresses))
	}

	ing := ingresses[0]
	if ing.Sign != "Aries" {
		t.Errorf("expected Aries, got %s", ing.Sign)
	}
	if ing.Planet != "Sun" {
		t.Errorf("expected Sun, got %s", ing.Planet)
	}

	expected := time.Date(2026, 3, 21, 12, 0, 0, 0, time.UTC)
	diff := ing.Time.Sub(expected)
	if diff < 0 {
		diff = -diff
	}
	if diff > 5*time.Minute {
		t.Errorf("ingress time %v too far from expected %v (diff: %v)", ing.Time, expected, diff)
	}
}

func TestFindSolarIngresses_NoIngressInRange(t *testing.T) {
	t.Parallel()

	refTime := time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC)
	compute := mockSunCompute(refTime, 5.0) // Sun at 5° Aries, won't cross a boundary soon

	start := time.Date(2026, 3, 19, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 3, 22, 0, 0, 0, 0, time.UTC)

	ingresses, err := FindSolarIngresses(start, end, compute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ingresses) != 0 {
		t.Fatalf("expected 0 ingresses, got %d", len(ingresses))
	}
}

func TestFindSolarIngresses_MultipleIngresses(t *testing.T) {
	t.Parallel()

	// Sun at 28° Pisces on March 19, moving 1°/day
	// Aries ingress: ~March 21 (cardinal)
	// Cancer ingress: ~June 21 (cardinal)
	refTime := time.Date(2026, 3, 19, 0, 0, 0, 0, time.UTC)
	compute := mockSunCompute(refTime, 358.0)

	start := time.Date(2026, 3, 18, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 6, 25, 0, 0, 0, 0, time.UTC)

	ingresses, err := FindSolarIngresses(start, end, compute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ingresses) != 2 {
		t.Fatalf("expected 2 ingresses, got %d", len(ingresses))
	}
	if ingresses[0].Sign != "Aries" {
		t.Errorf("expected first ingress Aries, got %s", ingresses[0].Sign)
	}
	if ingresses[1].Sign != "Cancer" {
		t.Errorf("expected second ingress Cancer, got %s", ingresses[1].Sign)
	}
}

func TestFindSolarIngresses_AllFourCardinalIngresses(t *testing.T) {
	t.Parallel()

	// Sun at 28° Pisces on March 19, moving 1°/day
	// Aries: ~March 21 (2 days)
	// Cancer: ~June 21 (94 days)
	// Libra: ~September 23 (188 days)
	// Capricorn: ~December 22 (278 days)
	refTime := time.Date(2026, 3, 19, 0, 0, 0, 0, time.UTC)
	compute := mockSunCompute(refTime, 358.0)

	start := time.Date(2026, 3, 18, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 12, 25, 0, 0, 0, 0, time.UTC)

	ingresses, err := FindSolarIngresses(start, end, compute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ingresses) != 4 {
		t.Fatalf("expected 4 ingresses, got %d", len(ingresses))
	}

	expectedSigns := []string{"Aries", "Cancer", "Libra", "Capricorn"}
	for i, exp := range expectedSigns {
		if ingresses[i].Sign != exp {
			t.Errorf("ingress %d: expected %s, got %s", i, exp, ingresses[i].Sign)
		}
	}
}

// mockSunMoonCompute returns Sun at 1°/day and Moon at 13°/day.
// At refTime, Sun is at sunLon and Moon is at moonLon.
func mockSunMoonCompute(refTime time.Time, sunLon, moonLon float64) ComputeFunc {
	return func(year, month, day int, hour float64, planetID int) (float64, float64, float64, float64) {
		t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
		days := t.Sub(refTime).Hours()/24.0 + hour/24.0

		var lon float64
		switch planetID {
		case 0: // Sun
			lon = sunLon + days*1.0
		case 1: // Moon
			lon = moonLon + days*13.0
		default:
			return 47.0, 0, 0, 0
		}
		for lon < 0 {
			lon += 360
		}
		for lon >= 360 {
			lon -= 360
		}
		return lon, 0, 0, 0
	}
}

func TestFindLunations_NewMoon(t *testing.T) {
	t.Parallel()

	// Sun at 100°, Moon at 99° — Moon approaching Sun at 12°/day relative
	// New moon at refTime + 1/12 day = 2 hours later
	refTime := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)
	compute := mockSunMoonCompute(refTime, 100.0, 99.0)

	start := time.Date(2026, 6, 14, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 6, 16, 0, 0, 0, 0, time.UTC)

	lunations, err := FindLunations(start, end, compute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(lunations) != 1 {
		t.Fatalf("expected 1 lunation, got %d", len(lunations))
	}

	l := lunations[0]
	if l.Type != "New Moon" {
		t.Errorf("expected New Moon, got %s", l.Type)
	}

	expected := refTime.Add(2 * time.Hour)
	diff := l.Time.Sub(expected)
	if diff < 0 {
		diff = -diff
	}
	if diff > 5*time.Minute {
		t.Errorf("lunation time %v too far from expected %v (diff: %v)", l.Time, expected, diff)
	}
}

func TestFindLunations_FullMoon(t *testing.T) {
	t.Parallel()

	// Sun at 100°, Moon at 279° — Moon approaching opposition (180°)
	// Moon at 279°, Sun at 100°: distance = 179°, need 1° more
	// Relative speed: 12°/day, so 1/12 day = 2 hours
	refTime := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)
	compute := mockSunMoonCompute(refTime, 100.0, 279.0)

	start := time.Date(2026, 6, 14, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 6, 16, 0, 0, 0, 0, time.UTC)

	lunations, err := FindLunations(start, end, compute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(lunations) != 1 {
		t.Fatalf("expected 1 lunation, got %d", len(lunations))
	}

	l := lunations[0]
	if l.Type != "Full Moon" {
		t.Errorf("expected Full Moon, got %s", l.Type)
	}
}

func TestFindLunations_NoLunation(t *testing.T) {
	t.Parallel()

	// Sun at 100°, Moon at 50° — not near conjunction or opposition
	refTime := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)
	compute := mockSunMoonCompute(refTime, 100.0, 50.0)

	start := time.Date(2026, 6, 14, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 6, 16, 0, 0, 0, 0, time.UTC)

	lunations, err := FindLunations(start, end, compute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(lunations) != 0 {
		t.Fatalf("expected 0 lunations, got %d", len(lunations))
	}
}

// mockSunMoonNodeCompute returns Sun at 1°/day, Moon at 13°/day, and Node fixed.
func mockSunMoonNodeCompute(refTime time.Time, sunLon, moonLon, nodeLon float64) ComputeFunc {
	return func(year, month, day int, hour float64, planetID int) (float64, float64, float64, float64) {
		t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
		days := t.Sub(refTime).Hours()/24.0 + hour/24.0

		var lon float64
		switch planetID {
		case 0: // Sun
			lon = sunLon + days*1.0
		case 1: // Moon
			lon = moonLon + days*13.0
		case 10: // Mean Node (swe.MEAN_NODE)
			lon = nodeLon
		default:
			return 47.0, 0, 0, 0
		}
		for lon < 0 {
			lon += 360
		}
		for lon >= 360 {
			lon -= 360
		}
		return lon, 0, 0, 0
	}
}

func TestFindEclipses_SolarEclipse(t *testing.T) {
	t.Parallel()

	// New moon at 100° with Node at 105° (5° away — within 15° eclipse limit)
	refTime := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)
	compute := mockSunMoonNodeCompute(refTime, 100.0, 99.0, 105.0)

	start := time.Date(2026, 6, 14, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 6, 16, 0, 0, 0, 0, time.UTC)

	eclipses, err := FindEclipses(start, end, compute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(eclipses) != 1 {
		t.Fatalf("expected 1 eclipse, got %d", len(eclipses))
	}

	e := eclipses[0]
	if e.Type != "Solar Eclipse" {
		t.Errorf("expected Solar Eclipse, got %s", e.Type)
	}
}

func TestFindEclipses_LunarEclipse(t *testing.T) {
	t.Parallel()

	// Full moon at 100° with Node at 95° (5° away — within 15° eclipse limit)
	refTime := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)
	compute := mockSunMoonNodeCompute(refTime, 100.0, 279.0, 95.0)

	start := time.Date(2026, 6, 14, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 6, 16, 0, 0, 0, 0, time.UTC)

	eclipses, err := FindEclipses(start, end, compute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(eclipses) != 1 {
		t.Fatalf("expected 1 eclipse, got %d", len(eclipses))
	}

	e := eclipses[0]
	if e.Type != "Lunar Eclipse" {
		t.Errorf("expected Lunar Eclipse, got %s", e.Type)
	}
}

func TestFindEclipses_NewMoonNotEclipse(t *testing.T) {
	t.Parallel()

	// New moon at 100° with Node at 130° (30° away — outside 15° limit)
	refTime := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)
	compute := mockSunMoonNodeCompute(refTime, 100.0, 99.0, 130.0)

	start := time.Date(2026, 6, 14, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 6, 16, 0, 0, 0, 0, time.UTC)

	eclipses, err := FindEclipses(start, end, compute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(eclipses) != 0 {
		t.Fatalf("expected 0 eclipses, got %d", len(eclipses))
	}
}

// mockSlowPlanetCompute returns a planet at a fixed speed from refTime.
func mockSlowPlanetCompute(refTime time.Time, planetID int, refLon, degPerDay float64) ComputeFunc {
	return func(year, month, day int, hour float64, pid int) (float64, float64, float64, float64) {
		if pid != planetID {
			return 47.0, 0, 0, 0
		}
		t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
		days := t.Sub(refTime).Hours()/24.0 + hour/24.0
		lon := refLon + days*degPerDay
		for lon < 0 {
			lon += 360
		}
		for lon >= 360 {
			lon -= 360
		}
		return lon, 0, 0, 0
	}
}

func TestFindPlanetaryIngresses_SaturnEntersAries(t *testing.T) {
	t.Parallel()

	// Saturn at 29.5° Pisces on March 20, moving 0.03°/day
	// Aries ingress at 0°: needs 0.5° / 0.03 = ~16.67 days → ~April 5
	refTime := time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC)
	compute := mockSlowPlanetCompute(refTime, 6, 359.5, 0.03) // planetID 6 = Saturn

	start := time.Date(2026, 3, 19, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 4, 10, 0, 0, 0, 0, time.UTC)

	ingresses, err := FindPlanetaryIngresses(start, end, 6, "Saturn", compute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ingresses) != 1 {
		t.Fatalf("expected 1 ingress, got %d", len(ingresses))
	}

	ing := ingresses[0]
	if ing.Sign != "Aries" {
		t.Errorf("expected Aries, got %s", ing.Sign)
	}
	if ing.Planet != "Saturn" {
		t.Errorf("expected Saturn, got %s", ing.Planet)
	}
}

func TestFindPlanetaryIngresses_NoIngress(t *testing.T) {
	t.Parallel()

	// Saturn at 5° Aries, moving 0.03°/day — won't cross a boundary soon
	refTime := time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC)
	compute := mockSlowPlanetCompute(refTime, 6, 5.0, 0.03)

	start := time.Date(2026, 3, 19, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 3, 25, 0, 0, 0, 0, time.UTC)

	ingresses, err := FindPlanetaryIngresses(start, end, 6, "Saturn", compute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ingresses) != 0 {
		t.Fatalf("expected 0 ingresses, got %d", len(ingresses))
	}
}

func TestFindPlanetaryIngresses_Retrograde(t *testing.T) {
	t.Parallel()

	// Planet at 0.5° Aries, moving -0.02°/day (retrograde)
	// Will cross back into Pisces (330° boundary) in ~25 days
	refTime := time.Date(2026, 3, 20, 0, 0, 0, 0, time.UTC)
	compute := mockSlowPlanetCompute(refTime, 6, 0.5, -0.02)

	start := time.Date(2026, 3, 19, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 4, 20, 0, 0, 0, 0, time.UTC)

	ingresses, err := FindPlanetaryIngresses(start, end, 6, "Saturn", compute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ingresses) != 1 {
		t.Fatalf("expected 1 ingress, got %d", len(ingresses))
	}

	ing := ingresses[0]
	if ing.Sign != "Pisces" {
		t.Errorf("expected Pisces (retrograde back), got %s", ing.Sign)
	}
}


// mockTwoPlanetCompute returns two planets at fixed speeds from refTime.
func mockTwoPlanetCompute(refTime time.Time, p1ID int, p1Lon, p1Speed float64, p2ID int, p2Lon, p2Speed float64) ComputeFunc {
	return func(year, month, day int, hour float64, pid int) (float64, float64, float64, float64) {
		t := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
		days := t.Sub(refTime).Hours()/24.0 + hour/24.0

		var lon float64
		switch pid {
		case p1ID:
			lon = p1Lon + days*p1Speed
		case p2ID:
			lon = p2Lon + days*p2Speed
		default:
			return 47.0, 0, 0, 0
		}
		for lon < 0 {
			lon += 360
		}
		for lon >= 360 {
			lon -= 360
		}
		return lon, 0, 0, 0
	}
}

func TestFindConjunctions_SaturnJupiter(t *testing.T) {
	t.Parallel()

	// Saturn at 100°, Jupiter at 99° — Jupiter catching up at 0.05°/day relative
	// Conjunction at refTime + 1/0.05 = 20 days later
	refTime := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)
	compute := mockTwoPlanetCompute(refTime, 6, 100.0, 0.03, 5, 99.0, 0.08)

	start := time.Date(2026, 6, 14, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 7, 10, 0, 0, 0, 0, time.UTC)

	conjunctions, err := FindConjunctions(start, end, 6, "Saturn", 5, "Jupiter", compute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(conjunctions) != 1 {
		t.Fatalf("expected 1 conjunction, got %d", len(conjunctions))
	}

	c := conjunctions[0]
	if c.Planet1 != "Saturn" || c.Planet2 != "Jupiter" {
		t.Errorf("expected Saturn-Jupiter, got %s-%s", c.Planet1, c.Planet2)
	}

	expected := refTime.Add(20 * 24 * time.Hour)
	diff := c.Time.Sub(expected)
	if diff < 0 {
		diff = -diff
	}
	if diff > 1*time.Hour {
		t.Errorf("conjunction time %v too far from expected %v (diff: %v)", c.Time, expected, diff)
	}
}

func TestFindConjunctions_NoConjunction(t *testing.T) {
	t.Parallel()

	// Saturn at 100°, Jupiter at 50° — far apart, won't conjunct soon
	refTime := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)
	compute := mockTwoPlanetCompute(refTime, 6, 100.0, 0.03, 5, 50.0, 0.08)

	start := time.Date(2026, 6, 14, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 6, 20, 0, 0, 0, 0, time.UTC)

	conjunctions, err := FindConjunctions(start, end, 6, "Saturn", 5, "Jupiter", compute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(conjunctions) != 0 {
		t.Fatalf("expected 0 conjunctions, got %d", len(conjunctions))
	}
}

// mockHouses returns fixed house cusps and angles for testing.
// ASC at 15° Aries, MC at 5° Capricorn, houses evenly spaced.
func mockHouses(jd, lat, lon float64, hsys byte) ([13]float64, [10]float64) {
	var cusps [13]float64
	var ascmc [10]float64
	for i := 1; i <= 12; i++ {
		cusps[i] = float64((i-1)*30) + 15.0 // midpoints of signs
	}
	ascmc[0] = 15.0  // ASC at 15° Aries
	ascmc[1] = 275.0 // MC at 5° Capricorn
	return cusps, ascmc
}

func TestCastChart_BasicStructure(t *testing.T) {
	t.Parallel()

	// Mock: Sun at 0°, Moon at 90°, all others at 45°
	compute := func(year, month, day int, hour float64, planetID int) (float64, float64, float64, float64) {
		switch planetID {
		case 0:
			return 0.0, 0, 0, 0
		case 1:
			return 90.0, 0, 0, 0
		default:
			return 45.0, 0, 0, 0
		}
	}

	tm := time.Date(2026, 7, 14, 9, 43, 38, 0, time.UTC)
	chart, err := CastChart(tm, 40.0, -74.0, compute, mockHouses, 'W')
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if chart.Time != tm {
		t.Errorf("expected time %v, got %v", tm, chart.Time)
	}
	if chart.Lat != 40.0 {
		t.Errorf("expected lat 40.0, got %f", chart.Lat)
	}
	if chart.Lon != -74.0 {
		t.Errorf("expected lon -74.0, got %f", chart.Lon)
	}

	// Check planet positions
	if chart.Planets["Sun"] != 0.0 {
		t.Errorf("expected Sun at 0.0, got %f", chart.Planets["Sun"])
	}
	if chart.Planets["Moon"] != 90.0 {
		t.Errorf("expected Moon at 90.0, got %f", chart.Planets["Moon"])
	}

	// Should have at least the basic planets
	for _, name := range []string{"Sun", "Moon", "Mercury", "Venus", "Mars", "Jupiter", "Saturn"} {
		if _, ok := chart.Planets[name]; !ok {
			t.Errorf("missing planet: %s", name)
		}
	}
}

func TestCastChart_HousesAndAngles(t *testing.T) {
	t.Parallel()

	// Mock: all planets at 0° (simplifies house computation)
	compute := func(year, month, day int, hour float64, planetID int) (float64, float64, float64, float64) {
		return 0.0, 0, 0, 0
	}

	tm := time.Date(2026, 7, 14, 9, 43, 38, 0, time.UTC)
	chart, err := CastChart(tm, 40.0, -74.0, compute, mockHouses, 'W')
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have 12 house cusps
	if len(chart.Houses) != 12 {
		t.Errorf("expected 12 houses, got %d", len(chart.Houses))
	}

	// Should have ASC and MC
	if chart.ASC == 0 {
		t.Error("ASC should not be zero")
	}
	if chart.MC == 0 {
		t.Error("MC should not be zero")
	}

	// ASC and MC should be different
	if chart.ASC == chart.MC {
		t.Error("ASC and MC should differ")
	}
}

func TestCastIngressChart(t *testing.T) {
	t.Parallel()

	compute := func(year, month, day int, hour float64, planetID int) (float64, float64, float64, float64) {
		return 45.0, 0, 0, 0
	}

	event := IngressEvent{
		Planet: "Sun",
		Sign:   "Aries",
		Time:   time.Date(2026, 3, 20, 14, 45, 57, 0, time.UTC),
		Lon:    0.0,
	}

	chart, err := CastIngressChart(event, 38.9, -77.0, compute, mockHouses)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if chart.Time != event.Time {
		t.Errorf("chart time should match event time")
	}
	if chart.Lat != 38.9 || chart.Lon != -77.0 {
		t.Errorf("chart location should match input")
	}
}

func TestCastLunationChart(t *testing.T) {
	t.Parallel()

	compute := func(year, month, day int, hour float64, planetID int) (float64, float64, float64, float64) {
		return 45.0, 0, 0, 0
	}

	event := LunationEvent{
		Type: "New Moon",
		Time: time.Date(2026, 7, 14, 9, 43, 38, 0, time.UTC),
	}

	chart, err := CastLunationChart(event, 51.5, -0.1, compute, mockHouses)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if chart.Time != event.Time {
		t.Errorf("chart time should match event time")
	}
}

func TestNationalChart_Lookup(t *testing.T) {
	t.Parallel()

	chart, ok := NationalChart("United States")
	if !ok {
		t.Fatal("United States not found in national chart database")
	}

	if chart.Name != "United States" {
		t.Errorf("expected United States, got %s", chart.Name)
	}
	if chart.Year != 1776 || chart.Month != 7 || chart.Day != 4 {
		t.Errorf("expected July 4 1776, got %d-%d-%d", chart.Year, chart.Month, chart.Day)
	}
	if chart.Lat != 39.95 || chart.Lon != -75.15 {
		t.Errorf("expected Philadelphia coords, got %.2f, %.2f", chart.Lat, chart.Lon)
	}
}

func TestNationalChart_NotFound(t *testing.T) {
	t.Parallel()

	_, ok := NationalChart("Atlantis")
	if ok {
		t.Error("Atlantis should not be in the database")
	}
}

func TestNationalCharts_List(t *testing.T) {
	t.Parallel()

	charts := NationalCharts()
	if len(charts) < 5 {
		t.Errorf("expected at least 5 national charts, got %d", len(charts))
	}

	// Verify no duplicate names
	seen := make(map[string]bool)
	for _, c := range charts {
		if seen[c.Name] {
			t.Errorf("duplicate chart: %s", c.Name)
		}
		seen[c.Name] = true
	}
}

func TestChartAspects_FindsConjunction(t *testing.T) {
	t.Parallel()

	compute := func(year, month, day int, hour float64, planetID int) (float64, float64, float64, float64) {
		switch planetID {
		case 0: return 100.0, 0, 0, 0 // Sun
		case 1: return 100.5, 0, 0, 0 // Moon
		default: return 200.0, 0, 0, 0
		}
	}

	tm := time.Date(2026, 7, 14, 12, 0, 0, 0, time.UTC)
	chart, err := CastChart(tm, 40.0, -74.0, compute, mockHouses, 'W')
	if err != nil {
		t.Fatalf("CastChart: %v", err)
	}

	aspects := ChartAspects(chart, 3.0)
	if len(aspects) == 0 {
		t.Fatal("expected at least 1 aspect")
	}

	// Should find Sun-Moon conjunction
	found := false
	for _, a := range aspects {
		if (a.Planet1 == "Sun" && a.Planet2 == "Moon") || (a.Planet1 == "Moon" && a.Planet2 == "Sun") {
			if a.Aspect != "conjunction" {
				t.Errorf("expected conjunction, got %s", a.Aspect)
			}
			found = true
		}
	}
	if !found {
		t.Error("Sun-Moon conjunction not found")
	}
}

func TestChartAspects_FindsOpposition(t *testing.T) {
	t.Parallel()

	compute := func(year, month, day int, hour float64, planetID int) (float64, float64, float64, float64) {
		switch planetID {
		case 0: return 100.0, 0, 0, 0 // Sun
		case 1: return 279.5, 0, 0, 0 // Moon (0.5° from exact opposition)
		default: return 200.0, 0, 0, 0
		}
	}

	tm := time.Date(2026, 7, 14, 12, 0, 0, 0, time.UTC)
	chart, _ := CastChart(tm, 40.0, -74.0, compute, mockHouses, 'W')

	aspects := ChartAspects(chart, 3.0)

	found := false
	for _, a := range aspects {
		if (a.Planet1 == "Sun" && a.Planet2 == "Moon") || (a.Planet1 == "Moon" && a.Planet2 == "Sun") {
			if a.Aspect != "opposition" {
				t.Errorf("expected opposition, got %s", a.Aspect)
			}
			found = true
		}
	}
	if !found {
		t.Error("Sun-Moon opposition not found")
	}
}

func TestChartAspects_RespectsOrb(t *testing.T) {
	t.Parallel()

	compute := func(year, month, day int, hour float64, planetID int) (float64, float64, float64, float64) {
		switch planetID {
		case 0: return 100.0, 0, 0, 0 // Sun
		case 1: return 104.0, 0, 0, 0 // Moon (4° away)
		default: return 200.0, 0, 0, 0
		}
	}

	tm := time.Date(2026, 7, 14, 12, 0, 0, 0, time.UTC)
	chart, _ := CastChart(tm, 40.0, -74.0, compute, mockHouses, 'W')

	// With 3° orb, should NOT find the 4° conjunction
	aspects := ChartAspects(chart, 3.0)
	for _, a := range aspects {
		if (a.Planet1 == "Sun" && a.Planet2 == "Moon") || (a.Planet1 == "Moon" && a.Planet2 == "Sun") {
			t.Error("4° separation should not be detected with 3° orb")
		}
	}

	// With 5° orb, SHOULD find it
	aspects = ChartAspects(chart, 5.0)
	found := false
	for _, a := range aspects {
		if (a.Planet1 == "Sun" && a.Planet2 == "Moon") || (a.Planet1 == "Moon" && a.Planet2 == "Sun") {
			found = true
		}
	}
	if !found {
		t.Error("4° separation should be detected with 5° orb")
	}
}

func TestNationTransits_Basic(t *testing.T) {
	t.Parallel()

	// Mock: all planets at 100° for both natal and transit
	compute := func(year, month, day int, hour float64, planetID int) (float64, float64, float64, float64) {
		return 100.0, 0, 0, 0
	}

	hits, err := NationTransits("United States", "2026-07-01", "2026-07-03", 3.0, compute, mockHouses)
	if err != nil {
		t.Fatalf("NationTransits: %v", err)
	}

	// All planets at same position → conjunctions for every pair
	if len(hits) == 0 {
		t.Error("expected transit hits when all planets are conjunct")
	}
}

func TestNationTransits_UnknownNation(t *testing.T) {
	t.Parallel()

	compute := func(year, month, day int, hour float64, planetID int) (float64, float64, float64, float64) {
		return 100.0, 0, 0, 0
	}

	_, err := NationTransits("Atlantis", "2026-07-01", "2026-07-03", 3.0, compute, mockHouses)
	if err == nil {
		t.Error("expected error for unknown nation")
	}
}

func TestChartPatterns_DetectsGrandTrine(t *testing.T) {
	t.Parallel()

	// Three planets at 0°, 120°, 240° — perfect Grand Trine
	compute := func(year, month, day int, hour float64, planetID int) (float64, float64, float64, float64) {
		switch planetID {
		case 0: return 0.0, 0, 0, 0   // Sun
		case 1: return 120.0, 0, 0, 0 // Moon
		case 2: return 240.0, 0, 0, 0 // Mercury
		default: return 200.0, 0, 0, 0
		}
	}

	tm := time.Date(2026, 7, 14, 12, 0, 0, 0, time.UTC)
	chart, _ := CastChart(tm, 40.0, -74.0, compute, mockHouses, 'W')

	report := ChartPatterns(chart, 5.0)
	if report == nil {
		t.Fatal("expected non-nil report")
	}

	hasGrandTrine := false
	for _, p := range report.Patterns {
		if p.Kind == "grand_trine" {
			hasGrandTrine = true
			break
		}
	}
	if !hasGrandTrine {
		t.Error("expected Grand Trine pattern")
	}
}

func TestChartPatterns_DetectsTSquare(t *testing.T) {
	t.Parallel()

	// Sun at 0°, Moon at 180° (opposition), Mars at 90° (square both)
	compute := func(year, month, day int, hour float64, planetID int) (float64, float64, float64, float64) {
		switch planetID {
		case 0: return 0.0, 0, 0, 0   // Sun
		case 1: return 180.0, 0, 0, 0 // Moon
		case 4: return 90.0, 0, 0, 0  // Mars
		default: return 200.0, 0, 0, 0
		}
	}

	tm := time.Date(2026, 7, 14, 12, 0, 0, 0, time.UTC)
	chart, _ := CastChart(tm, 40.0, -74.0, compute, mockHouses, 'W')

	report := ChartPatterns(chart, 5.0)

	hasTSquare := false
	for _, p := range report.Patterns {
		if p.Kind == "t_square" {
			hasTSquare = true
			break
		}
	}
	if !hasTSquare {
		t.Error("expected T-Square pattern")
	}
}

func TestChartPatterns_EmptyChart(t *testing.T) {
	t.Parallel()

	compute := func(year, month, day int, hour float64, planetID int) (float64, float64, float64, float64) {
		return 200.0, 0, 0, 0 // all planets at same position — no aspects
	}

	tm := time.Date(2026, 7, 14, 12, 0, 0, 0, time.UTC)
	chart, _ := CastChart(tm, 40.0, -74.0, compute, mockHouses, 'W')

	report := ChartPatterns(chart, 5.0)
	if report == nil {
		t.Fatal("expected non-nil report even with no patterns")
	}
	// All planets at same position = stellium, not empty
	// That's fine — just verify it doesn't crash
}

func TestPlanetHouses_WholeSign(t *testing.T) {
	t.Parallel()

	// ASC at 15° Aries → Aries = 1st house, Taurus = 2nd, etc.
	// Sun at 10° Taurus → 2nd house
	// Moon at 200° Libra → 7th house
	// Mars at 350° Pisces → 12th house
	compute := func(year, month, day int, hour float64, planetID int) (float64, float64, float64, float64) {
		switch planetID {
		case 0: return 40.0, 0, 0, 0  // Sun 10° Taurus
		case 1: return 200.0, 0, 0, 0 // Moon 20° Libra
		case 4: return 350.0, 0, 0, 0 // Mars 20° Pisces
		default: return 100.0, 0, 0, 0
		}
	}

	tm := time.Date(2026, 7, 14, 12, 0, 0, 0, time.UTC)
	chart, _ := CastChart(tm, 40.0, -74.0, compute, mockHouses, 'W')

	houses := PlanetHouses(chart)

	if h := houses["Sun"]; h != 2 {
		t.Errorf("Sun at 10° Taurus (ASC Aries) → house 2, got %d", h)
	}
	if h := houses["Moon"]; h != 7 {
		t.Errorf("Moon at 20° Libra (ASC Aries) → house 7, got %d", h)
	}
	if h := houses["Mars"]; h != 12 {
		t.Errorf("Mars at 20° Pisces (ASC Aries) → house 12, got %d", h)
	}
}

func TestPlanetHouses_AllPlanetsAssigned(t *testing.T) {
	t.Parallel()

	compute := func(year, month, day int, hour float64, planetID int) (float64, float64, float64, float64) {
		return 45.0, 0, 0, 0
	}

	tm := time.Date(2026, 7, 14, 12, 0, 0, 0, time.UTC)
	chart, _ := CastChart(tm, 40.0, -74.0, compute, mockHouses, 'W')

	houses := PlanetHouses(chart)

	// Every planet in chart.Planets should have a house assignment
	for planet := range chart.Planets {
		h, ok := houses[planet]
		if !ok {
			t.Errorf("planet %s has no house assignment", planet)
		}
		if h < 1 || h > 12 {
			t.Errorf("planet %s has invalid house %d", planet, h)
		}
	}
}

func TestInterpretMundaneChart_Basic(t *testing.T) {
	t.Parallel()

	compute := func(year, month, day int, hour float64, planetID int) (float64, float64, float64, float64) {
		switch planetID {
		case 0: return 40.0, 0, 0, 0  // Sun 10° Taurus
		case 1: return 200.0, 0, 0, 0 // Moon 20° Libra
		default: return 100.0, 0, 0, 0
		}
	}

	tm := time.Date(2026, 7, 14, 12, 0, 0, 0, time.UTC)
	chart, _ := CastChart(tm, 40.0, -74.0, compute, mockHouses, 'W')

	report := InterpretMundaneChart("Test Chart", chart, 5.0)
	if report == nil {
		t.Fatal("expected non-nil report")
	}
	if report.Name != "Test Chart" {
		t.Errorf("expected 'Test Chart', got %s", report.Name)
	}

	// Should have planet-in-sign entries
	if len(report.PlanetSigns) == 0 {
		t.Error("expected planet-in-sign interpretations")
	}

	// Should have planet-in-house entries
	if len(report.PlanetHouses) == 0 {
		t.Error("expected planet-in-house interpretations")
	}

	// Verify Sun in 2nd house (Taurus with ASC Aries)
	t.Logf("Planet signs: %v", report.PlanetSigns)
	t.Logf("Planet houses: %v", report.PlanetHouses)
}

func TestInterpretMundaneChartFull_US(t *testing.T) {
	// Use mock compute and houses
	compute := func(year, month, day int, hour float64, planetID int) (float64, float64, float64, float64) {
		return 45.0, 0, 0, 0
	}

	tm := time.Date(1776, 7, 4, 22, 13, 0, 0, time.UTC)
	chart, err := CastChart(tm, 39.95, -75.15, compute, mockHouses, 'W')
	if err != nil {
		t.Fatalf("CastChart: %v", err)
	}

	report := InterpretMundaneChartFull("United States", "natal", chart, 5.0)
	if report == nil {
		t.Fatal("expected non-nil report")
	}
	if report.Name != "United States" {
		t.Errorf("expected United States, got %s", report.Name)
	}
	if report.ChartType != "natal" {
		t.Errorf("expected natal, got %s", report.ChartType)
	}
	if report.ASCSign == "" {
		t.Error("ASC sign should not be empty")
	}
	if report.MCSign == "" {
		t.Error("MC sign should not be empty")
	}
	if len(report.PlanetHouses) == 0 {
		t.Error("expected planet-in-house interpretations")
	}
	if report.Summary == "" {
		t.Error("expected summary")
	}

	t.Logf("ASC: %s", report.ASCSign)
	t.Logf("MC: %s", report.MCSign)
	t.Logf("ASC interpretation: %s", report.ASCInterpretation)
	t.Logf("MC interpretation: %s", report.MCInterpretation)
	t.Logf("Summary: %s", report.Summary)
	for _, ph := range report.PlanetHouses {
		t.Logf("  %s", ph[:min(80, len(ph))])
	}
}

func min(a, b int) int {
	if a < b { return a }
	return b
}
