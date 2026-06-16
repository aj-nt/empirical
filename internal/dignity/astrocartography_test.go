package dignity

import (
	"math"
	"testing"
)

// ============================================================================
// Astrocartography — Tests
// ============================================================================

func TestGMST(t *testing.T) {
	jd := 2451545.0
	gmst := ComputeGMST(jd)
	expected := 280.46061837
	if math.Abs(gmst-expected) > 0.01 {
		t.Errorf("GMST at J2000: want %.2f, got %.2f", expected, gmst)
	}

	jd2 := 2451546.0
	gmst2 := ComputeGMST(jd2)
	advance := (gmst2 - gmst + 360) - 360
	expectedAdvance := 0.98564736629
	if math.Abs(advance-expectedAdvance) > 0.01 {
		t.Errorf("GMST advance per day (normalized): want %.4f, got %.4f", expectedAdvance, advance)
	}
}

func TestMCLine(t *testing.T) {
	planetRA := 0.0
	gmst := 0.0
	line := ComputeMCLine(planetRA, gmst, 2.0)

	if len(line) < 10 {
		t.Fatalf("MC line too short: %d points", len(line))
	}
	for _, pt := range line {
		if math.Abs(pt.Lon-line[0].Lon) > 0.01 {
			t.Errorf("MC line longitude varies: %.2f vs %.2f", pt.Lon, line[0].Lon)
		}
	}
	if math.Abs(line[0].Lon-0.0) > 0.01 {
		t.Errorf("MC line longitude: want 0, got %.2f", line[0].Lon)
	}
}

func TestMCLine_Wrap(t *testing.T) {
	planetRA := 10.0
	gmst := 350.0
	line := ComputeMCLine(planetRA, gmst, 2.0)

	expected := 20.0
	if math.Abs(line[0].Lon-expected) > 0.01 {
		t.Errorf("MC line wrap: want %.2f, got %.2f", expected, line[0].Lon)
	}
}

func TestLinesNear(t *testing.T) {
	lines := []AstroLine{
		{
			Planet: "Sun",
			Angle:  "MC",
			Points: []GeoPoint{
				{Lat: -80, Lon: 0},
				{Lat: -40, Lon: 0},
				{Lat: 0, Lon: 0},
				{Lat: 40, Lon: 0},
				{Lat: 80, Lon: 0},
			},
		},
	}

	hits := LinesNear(41.0, 0.5, lines, 2.0)
	if len(hits) < 1 {
		t.Fatal("Expected at least 1 hit near (41, 0.5)")
	}
	if hits[0].Planet != "Sun" || hits[0].Angle != "MC" {
		t.Errorf("Expected Sun/MC hit, got %s/%s", hits[0].Planet, hits[0].Angle)
	}
}

func TestLinesNear_NoHits(t *testing.T) {
	lines := []AstroLine{
		{
			Planet: "Sun",
			Angle:  "MC",
			Points: []GeoPoint{
				{Lat: -80, Lon: 0},
				{Lat: 80, Lon: 0},
			},
		},
	}

	hits := LinesNear(41.0, 120.0, lines, 2.0)
	if len(hits) != 0 {
		t.Errorf("Expected 0 hits far from line, got %d", len(hits))
	}
}

func TestNormalizeGeo(t *testing.T) {
	tests := []struct{ in, want float64 }{
		{0, 0},
		{180, 180},
		{-180, -180},
		{181, -179},
		{-181, 179},
		{359, -1},
		{-359, 1},
		{540, 180},
		{-540, -180},
	}
	for _, tc := range tests {
		got := NormalizeGeo(tc.in)
		if math.Abs(got-tc.want) > 0.01 {
			t.Errorf("NormalizeGeo(%.1f): want %.1f, got %.1f", tc.in, tc.want, got)
		}
	}
}

func TestComputeAstroCartographyLines(t *testing.T) {
	planets := map[string]float64{
		"Sun": 0.0, "Moon": 90.0, "Mars": 180.0,
	}
	jd := 2451545.0

	lines := ComputeAstroCartographyLines(planets, jd, 10.0)

	// 3 planets x 2 angles (MC, IC) = 6 lines
	if len(lines) != 6 {
		t.Errorf("Expected 6 lines (3 planets x MC+IC), got %d", len(lines))
	}

	for _, line := range lines {
		if line.Planet == "" {
			t.Error("Line has empty planet name")
		}
		if line.Angle != "MC" && line.Angle != "IC" {
			t.Errorf("Unexpected angle: %s", line.Angle)
		}
		if len(line.Points) < 5 {
			t.Errorf("Line %s/%s has too few points: %d", line.Planet, line.Angle, len(line.Points))
		}
	}
}

// ============================================================================
// Draconic & Cross Astrocartography — Tests
// ============================================================================

func TestDraconicAstroCartography_MCIC_Shifted(t *testing.T) {
	// MC/IC lines are RA-based. RA is NOT invariant under uniform longitude
	// shift — atan2(sin(λ)·cos(ε), cos(λ)) is nonlinear.
	// Draconic MC/IC lines should differ from tropical, but the shift should
	// be consistent (same offset for all latitudes on a given line).

	planets := map[string]float64{
		"Sun": 0.0, "Moon": 90.0, "Mars": 180.0,
	}
	jd := 2451545.0
	northNodeLon := 45.0

	tropical := ComputeAstroCartographyLines(planets, jd, 10.0)
	draconic := ComputeDraconicAstroCartographyLines(planets, jd, northNodeLon, 10.0)

	if len(draconic) != len(tropical) {
		t.Fatalf("Line count mismatch: tropical %d, draconic %d", len(tropical), len(draconic))
	}

	// Each line should have a consistent offset across all latitudes
	for i := 0; i < len(tropical); i++ {
		tl := tropical[i]
		dl := draconic[i]
		if len(tl.Points) != len(dl.Points) {
			t.Errorf("%s/%s: point count differs", tl.Planet, tl.Angle)
			continue
		}
		// Verify the offset is consistent across latitudes
		baseOffset := normalizeGeo(dl.Points[0].Lon - tl.Points[0].Lon)
		for j := 1; j < len(tl.Points); j++ {
			offset := normalizeGeo(dl.Points[j].Lon - tl.Points[j].Lon)
			if math.Abs(offset-baseOffset) > 0.01 {
				t.Errorf("%s/%s: offset varies at lat %.1f: base %.2f, got %.2f",
					tl.Planet, tl.Angle, tl.Points[j].Lat, baseOffset, offset)
			}
		}
		// Offset should be nonzero (draconic ≠ tropical)
		if math.Abs(baseOffset) < 0.01 {
			t.Errorf("%s/%s: draconic MC/IC should differ from tropical, but offset is ~0",
				tl.Planet, tl.Angle)
		}
	}
}

func TestCrossAstroCartography_MCIC_Identical(t *testing.T) {
	// Cross MC/IC = tropical MC/IC (RA-based, no shift applied).

	planets := map[string]float64{
		"Sun": 0.0, "Jupiter": 120.0,
	}
	jd := 2451545.0
	northNodeLon := 30.0

	tropical := ComputeAstroCartographyLines(planets, jd, 10.0)
	cross := ComputeCrossAstroCartographyLines(planets, jd, northNodeLon, 10.0)

	// Index by planet+angle (map iteration order is random)
	tropByKey := make(map[string]*AstroLine)
	for i := range tropical {
		tropByKey[tropical[i].Planet+"/"+tropical[i].Angle] = &tropical[i]
	}

	for i := range cross {
		cl := cross[i]
		key := cl.Planet + "/" + cl.Angle
		tl, ok := tropByKey[key]
		if !ok {
			t.Errorf("Cross line %s not found in tropical", key)
			continue
		}
		for j := range tl.Points {
			if math.Abs(tl.Points[j].Lon-cl.Points[j].Lon) > 0.01 {
				t.Errorf("Cross %s MC/IC should match tropical: %.2f vs %.2f",
					key, tl.Points[j].Lon, cl.Points[j].Lon)
				break // one failure per line is enough
			}
		}
	}
}

func TestThreeWayComparison(t *testing.T) {
	planets := map[string]float64{"Sun": 0.0, "Moon": 90.0}
	jd := 2451545.0
	nn := 30.0

	tropical := ComputeAstroCartographyLines(planets, jd, 10.0)
	draconic := ComputeDraconicAstroCartographyLines(planets, jd, nn, 10.0)
	cross := ComputeCrossAstroCartographyLines(planets, jd, nn, 10.0)

	// At J2000, GMST=280.46°. Sun RA=0° → MC lon = 0-280.46 = 79.54°.
	// Test near that line: lat=0, lon=79.5
	hits := CompareLinesNear(0.0, 79.5, tropical, draconic, cross, 5.0)

	if len(hits) == 0 {
		t.Error("Expected Sun MC hit near (0, 79.5)")
	}

	for _, h := range hits {
		if h.Planet == "" {
			t.Error("Hit has empty planet")
		}
	}
}
