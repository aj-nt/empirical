package dignity

import (
	"math"
	"testing"

	"github.com/aj-nt/empirical/internal/swe"
)

// mockHouses returns a simple ASC that increases with longitude.
// ASC(lon) = (lon + 180) % 360 — a 1:1 mapping for testing.
func mockHouses(jd, lat, lon float64, hsys byte) ([13]float64, [10]float64) {
	asc := math.Mod(lon+180, 360)
	if asc < 0 {
		asc += 360
	}
	var cusps [13]float64
	var ascmc [10]float64
	ascmc[0] = asc
	return cusps, ascmc
}

func TestFindASCLon_Simple(t *testing.T) {
	// With mockHouses, ASC(lon) = (lon + 180) % 360
	// So ASC = 0 at lon = 180, ASC = 90 at lon = -90, etc.
	lon := findASCLon(0, 0, 0, mockHouses)
	if lon == nil {
		t.Fatal("expected non-nil result")
	}
	// ASC = 0 when lon = 180 (or -180)
	if math.Abs(*lon-180) > 0.01 && math.Abs(*lon+180) > 0.01 {
		t.Errorf("expected lon ≈ ±180, got %.4f", *lon)
	}
}

func TestFindASCLon_Target90(t *testing.T) {
	lon := findASCLon(90, 0, 0, mockHouses)
	if lon == nil {
		t.Fatal("expected non-nil result")
	}
	// ASC = 90 when lon = -90
	if math.Abs(*lon+90) > 0.01 {
		t.Errorf("expected lon ≈ -90, got %.4f", *lon)
	}
}

func TestFindASCLon_Target180(t *testing.T) {
	lon := findASCLon(180, 0, 0, mockHouses)
	if lon == nil {
		t.Fatal("expected non-nil result")
	}
	// ASC = 180 when lon = 0
	if math.Abs(*lon) > 0.01 {
		t.Errorf("expected lon ≈ 0, got %.4f", *lon)
	}
}

func TestFindASCLon_Target270(t *testing.T) {
	lon := findASCLon(270, 0, 0, mockHouses)
	if lon == nil {
		t.Fatal("expected non-nil result")
	}
	// ASC = 270 when lon = 90
	if math.Abs(*lon-90) > 0.01 {
		t.Errorf("expected lon ≈ 90, got %.4f", *lon)
	}
}

func TestComputeASCLine_ProducesPoints(t *testing.T) {
	points := ComputeASCLine(0, 0, 10.0, mockHouses)
	if len(points) == 0 {
		t.Fatal("expected non-empty points")
	}
	// With latStep=10, should have ~17 points (-80 to 80)
	if len(points) < 15 {
		t.Errorf("expected ~17 points, got %d", len(points))
	}
	// All longitudes should be near ±180 (where ASC=0 with mockHouses)
	for _, p := range points {
		if math.Abs(p.Lon-180) > 0.01 && math.Abs(p.Lon+180) > 0.01 {
			t.Errorf("point at lat=%.1f has lon=%.4f, expected ±180", p.Lat, p.Lon)
		}
	}
}

func TestComputeDSCLine_OppositeOfASC(t *testing.T) {
	// DSC line for planetLon=0: where DSC=0, i.e. where ASC=180.
	// With mockHouses, ASC=180 at lon=0, so DSC line should be at lon=0+180=180.
	// ASC line for planetLon=0: where ASC=0, i.e. lon=180 (or -180).
	// These are different curves — DSC is not ASC+180 in longitude.
	ascPoints := ComputeASCLine(0, 0, 10.0, mockHouses)
	dscPoints := ComputeDSCLine(0, 0, 10.0, mockHouses)

	if len(ascPoints) != len(dscPoints) {
		t.Fatalf("ASC and DSC should have same length: %d vs %d", len(ascPoints), len(dscPoints))
	}

	// Verify each DSC point: at that lon, DSC should equal the planet (0)
	for _, p := range dscPoints {
		_, ascmc := mockHouses(0, p.Lat, p.Lon, 'P')
		asc := ascmc[0]
		dsc := NormalizeLon(asc + 180)
		orb := AngleDist(dsc, 0)
		if orb > 1.0 {
			t.Errorf("lat=%.1f lon=%.4f: DSC=%.4f, expected 0, orb=%.4f", p.Lat, p.Lon, dsc, orb)
		}
	}

	// Verify each ASC point: at that lon, ASC should equal the planet (0)
	for _, p := range ascPoints {
		_, ascmc := mockHouses(0, p.Lat, p.Lon, 'P')
		asc := ascmc[0]
		orb := AngleDist(asc, 0)
		if orb > 1.0 {
			t.Errorf("lat=%.1f lon=%.4f: ASC=%.4f, expected 0, orb=%.4f", p.Lat, p.Lon, asc, orb)
		}
	}
}

func TestSignedAngularDist(t *testing.T) {
	tests := []struct {
		a, b     float64
		expected float64
	}{
		{10, 0, 10},
		{0, 10, -10},
		{350, 10, -20},  // 350 is 20 behind 10
		{10, 350, 20},   // 10 is 20 ahead of 350
		{180, 0, 180},   // boundary: positive
		{0, 180, -180},  // boundary: negative
		{0, 0, 0},
		{359, 1, -2},
		{1, 359, 2},
	}
	for _, tt := range tests {
		got := signedAngularDist(tt.a, tt.b)
		if math.Abs(got-tt.expected) > 0.01 {
			t.Errorf("signedAngularDist(%.0f, %.0f) = %.4f, want %.4f", tt.a, tt.b, got, tt.expected)
		}
	}
}

func TestFindASCLon_DoesNotConfuseDSC(t *testing.T) {
	// The key test: with mockHouses, ASC(lon) = (lon+180)%360.
	// For target=0, ASC=0 at lon=180. But ASC=180 at lon=0.
	// The old code would confuse these. The new code must not.
	lon := findASCLon(0, 0, 0, mockHouses)
	if lon == nil {
		t.Fatal("expected non-nil")
	}
	// Verify: ASC at this lon should be ~0, not ~180
	_, ascmc := mockHouses(0, 0, *lon, 'P')
	asc := ascmc[0]
	orb := AngleDist(asc, 0)
	if orb > 1.0 {
		t.Errorf("ASC=%.4f at lon=%.4f, orb=%.4f — found DSC instead of ASC!", asc, *lon, orb)
	}
}

// mockHousesWithLat returns ASC = (lon + lat*0.5) % 360.
// This gives ASC lines a latitude-dependent slope so they can intersect MC lines.
func mockHousesWithLat(jd, lat, lon float64, hsys byte) ([13]float64, [10]float64) {
	asc := math.Mod(lon+lat*0.5, 360)
	if asc < 0 {
		asc += 360
	}
	var cusps [13]float64
	var ascmc [10]float64
	ascmc[0] = asc
	return cusps, ascmc
}

func TestFindMCxASCIntersection_Simple(t *testing.T) {
	// ASC(lat, lon) = (lon + lat*0.5) % 360
	// ASC line for target 0: lon = -lat*0.5 (normalized to [-180,180])
	// MC line at lon=0: vertical
	// Intersection: 0 = -lat*0.5 → lat=0, lon=0
	result := findMCxASCIntersection(0, 0, 0, mockHousesWithLat)
	if result == nil {
		t.Fatal("expected intersection at (0, 0)")
	}
	if math.Abs(result.Lat) > 0.01 || math.Abs(result.Lon) > 0.01 {
		t.Errorf("expected (0, 0), got (%.4f, %.4f)", result.Lat, result.Lon)
	}
}

func TestFindMCxASCIntersection_Offset(t *testing.T) {
	// ASC line for target 40: lon = (40 - lat*0.5) normalized
	// MC line at lon=0: vertical
	// Intersection: 0 = 40 - lat*0.5 → lat=80, lon=0
	result := findMCxASCIntersection(0, 40, 0, mockHousesWithLat)
	if result == nil {
		t.Fatal("expected intersection at (80, 0)")
	}
	if math.Abs(result.Lat-80) > 0.5 || math.Abs(result.Lon) > 0.5 {
		t.Errorf("expected (80, 0), got (%.4f, %.4f)", result.Lat, result.Lon)
	}
}

func TestFindMCxASCIntersection_NoIntersection(t *testing.T) {
	// ASC line for target 200: lon = (200 - lat*0.5) normalized
	// At lat=-80: lon = (200 + 40) = 240 → -120
	// At lat=80: lon = (200 - 40) = 160
	// MC line at lon=0: neither -120 nor 160 is near 0, and the line
	// doesn't cross 0 in between (it goes from -120 to 160, crossing 180/-180 but not 0)
	// Actually: -120 → 160 crosses 0? No, -120 to 160 goes through 0 if it wraps.
	// Let me pick values that definitely don't cross.
	// ASC target 90: lon = (90 - lat*0.5). At lat=-80: 130. At lat=80: 50.
	// MC at lon=-90: 130 to 50 doesn't cross -90. ✓
	result := findMCxASCIntersection(-90, 90, 0, mockHousesWithLat)
	if result != nil {
		t.Errorf("expected no intersection, got (%.4f, %.4f)", result.Lat, result.Lon)
	}
}

func TestIsInShortArc(t *testing.T) {
	tests := []struct {
		val, a, b float64
		expected  bool
	}{
		// Simple: val between a and b, short arc ≤ 180
		{10, 0, 20, true},
		{30, 0, 20, false},
		{0, 0, 20, true},
		{20, 0, 20, true},

		// Short arc wraps through 0: [350, 360) ∪ [0, 10]
		{5, 350, 10, true},
		{355, 350, 10, true},
		{180, 350, 10, false},
		{20, 350, 10, false},

		// Short arc exactly 180
		{90, 0, 180, true},
		{270, 0, 180, false},

		// Short arc > 180: the complement is the short arc
		// a=0, b=200: short arc is 200→360∪0→0? No, dist=200>180,
		// short arc wraps: [200, 360) ∪ [0, 0] = [200, 360)
		{300, 0, 200, true},
		{100, 0, 200, false},
	}
	for _, tt := range tests {
		got := isInShortArc(tt.val, tt.a, tt.b)
		if got != tt.expected {
			t.Errorf("isInShortArc(%.0f, %.0f, %.0f) = %v, want %v",
				tt.val, tt.a, tt.b, got, tt.expected)
		}
	}
}

func TestFindParans_Integration(t *testing.T) {
	// With mockHousesWithLat, ASC = (lon + lat*0.5) % 360.
	// ASC line for target=0: lon = -lat*0.5 (normalized to [-180,180]).
	// At lat=-80: lon=40. At lat=80: lon=-40.
	// Short arc from 40 to -40: [320, 360) ∪ [0, 40] in [0,360).
	//
	// MC line for planet at lon=0: RA=0, MC lon = normalizeGeo(-GMST).
	// For intersection, MC lon must be in [320,360) ∪ [0,40].
	// That means GMST ∈ [0,40] ∪ [320,360) mod 360.
	//
	// GMST at JD=0: let's compute it.
	gmst := ComputeGMST(0)
	t.Logf("GMST(0) = %.4f", gmst)

	// Planet Sun at 0°: MC lon = normalizeGeo(0 - gmst)
	// Planet Moon at 0°: ASC target = 0
	planets := map[string]float64{"Sun": 0, "Moon": 0}

	parans := FindParans(planets, 0, gmst, mockHousesWithLat)

	// Whether we get intersections depends on GMST(0).
	// If GMST(0) ∈ [0,40] ∪ [320,360), we get Sun-MC × Moon-ASC.
	// Let's just verify the results are geometrically consistent.
	for _, p := range parans {
		// Verify: at this (lat, lon), the ASC should equal the ASC planet's target
		_, ascmc := mockHousesWithLat(0, p.Lat, p.Lon, 'P')
		asc := ascmc[0]

		// Determine which planet is the ASC/DSC one
		var ascTarget float64
		if p.Angle2 == "ASC" {
			ascTarget = planets[p.Planet2]
		} else if p.Angle2 == "DSC" {
			ascTarget = NormalizeLon(planets[p.Planet2] - 180)
		} else {
			t.Errorf("unexpected angle2: %s", p.Angle2)
			continue
		}

		orb := AngleDist(asc, ascTarget)
		if orb > 2.0 {
			t.Errorf("at (%.4f, %.4f): ASC=%.4f, expected %s target=%.4f, orb=%.4f",
				p.Lat, p.Lon, asc, p.Planet2, ascTarget, orb)
		}

		// Verify: MC lon should be near this longitude
		mcLon := normalizeGeo(LonToRA(planets[p.Planet1], ObliquityDeg) - gmst)
		if p.Angle1 == "IC" {
			mcLon = normalizeGeo(mcLon + 180)
		}
		lonDiff := math.Abs(normalizeGeo(p.Lon - mcLon))
		if lonDiff > 2.0 {
			t.Errorf("at (%.4f, %.4f): lon=%.4f, expected %s lon=%.4f, diff=%.4f",
				p.Lat, p.Lon, p.Lon, p.Angle1, mcLon, lonDiff)
		}
	}
}

func TestFindParans_AJChart(t *testing.T) {
	// AJ's birth data: 1969-02-15 23:10 PST (UTC-8), Olympia WA
	bc, err := ComputeBaseChart(BirthData{
		Year: 1969, Month: 2, Day: 15,
		Hour: 23, Minute: 10,
		TZOffset: -8, Lat: 47.038, Lng: -122.901,
	})
	if err != nil {
		t.Skipf("ephemeris not available: %v", err)
	}

	gmst := ComputeGMST(bc.JD)
	positions := TropicalToLonMap(bc.Tropical)

	parans := FindParans(positions, bc.JD, gmst, swe.Houses)

	t.Logf("JD: %.4f, GMST: %.4f", bc.JD, gmst)
	t.Logf("Parans found: %d", len(parans))
	for _, p := range parans {
		t.Logf("  %s-%s × %s-%s at (%.2f, %.2f)",
			p.Planet1, p.Angle1, p.Planet2, p.Angle2, p.Lat, p.Lon)
	}
}


func TestGeocodeParans(t *testing.T) {
	t.Parallel()
	parans := []ParanIntersection{
		{Planet1: "Sun", Angle1: "MC", Planet2: "Moon", Angle2: "ASC", Lat: 7.88, Lon: 98.39},
		{Planet1: "Mars", Angle1: "IC", Planet2: "Jupiter", Angle2: "DSC", Lat: 18.79, Lon: 98.99},
		{Planet1: "Venus", Angle1: "MC", Planet2: "Saturn", Angle2: "ASC", Lat: -33.87, Lon: 151.21},
	}
	GeocodeParans(parans)

	// Phuket area
	if parans[0].LocationName != "Phuket" {
		t.Errorf("expected Phuket, got %q", parans[0].LocationName)
	}
	if parans[0].LocationCountry != "TH" {
		t.Errorf("expected TH, got %q", parans[0].LocationCountry)
	}

	// Chiang Mai area
	if parans[1].LocationName != "Chiang Mai" {
		t.Errorf("expected Chiang Mai, got %q", parans[1].LocationName)
	}
	if parans[1].LocationCountry != "TH" {
		t.Errorf("expected TH, got %q", parans[1].LocationCountry)
	}

	// Sydney area
	if parans[2].LocationName != "Sydney" {
		t.Errorf("expected Sydney, got %q", parans[2].LocationName)
	}
	if parans[2].LocationCountry != "AU" {
		t.Errorf("expected AU, got %q", parans[2].LocationCountry)
	}
}
