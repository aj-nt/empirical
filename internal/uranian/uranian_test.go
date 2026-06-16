package uranian

import (
	"math"
	"testing"
)

// ── Midpoint ────────────────────────────────────────────────────────────

func TestMidpoint(t *testing.T) {
	tests := []struct {
		a, b, want float64
	}{
		{0, 180, 90},       // opposition → midpoint at 90
		{0, 90, 45},        // square → midpoint at 45
		{350, 10, 0},       // wraparound: shorter arc through 0
		{10, 350, 0},       // symmetric
		{0, 0, 0},          // same point
		{120, 240, 180},    // trine → midpoint at 180
		{45, 315, 0},       // wraparound: 45→315 shorter through 0
		{100, 200, 150},    // normal case
	}
	for _, tc := range tests {
		got := Midpoint(tc.a, tc.b)
		if math.Abs(got-tc.want) > 0.01 {
			t.Errorf("Midpoint(%.0f, %.0f) = %.4f, want %.0f", tc.a, tc.b, got, tc.want)
		}
	}
}

func TestMidpointSymmetry(t *testing.T) {
	// Midpoint(a,b) and Midpoint(b,a) should be on the same axis
	// (differ by 0 or 180 mod 180, since both are valid midpoints)
	for a := 0.0; a < 360; a += 30 {
		for b := 0.0; b < 360; b += 30 {
			m1 := Midpoint(a, b)
			m2 := Midpoint(b, a)
			diff := math.Mod(math.Abs(m1-m2), 180)
			if diff > 0.01 && math.Abs(diff-180) > 0.01 {
				t.Errorf("Midpoint(%.0f, %.0f)=%.4f and Midpoint(%.0f, %.0f)=%.4f not on same axis (diff=%.4f)", a, b, m1, b, a, m2, diff)
			}
		}
	}
}

// ── IndirectHalfSums ────────────────────────────────────────────────────

func TestIndirectHalfSums(t *testing.T) {
	pts := IndirectHalfSums(0, 180)
	if len(pts) != 4 {
		t.Fatalf("expected 4 points, got %d", len(pts))
	}
	// Direct = 90, indirects = 180, 270, 0
	expected := []float64{0, 90, 180, 270}
	for i, want := range expected {
		if math.Abs(pts[i]-want) > 0.01 {
			t.Errorf("IndirectHalfSums[%d] = %.4f, want %.0f", i, pts[i], want)
		}
	}
}

func TestIndirectHalfSumsSorted(t *testing.T) {
	// Must be sorted ascending
	for a := 0.0; a < 360; a += 45 {
		for b := 0.0; b < 360; b += 45 {
			pts := IndirectHalfSums(a, b)
			for i := 1; i < len(pts); i++ {
				if pts[i] < pts[i-1] {
					t.Errorf("IndirectHalfSums(%.0f, %.0f) not sorted: %v", a, b, pts)
				}
			}
		}
	}
}

// ── PlanetaryPicture ─────────────────────────────────────────────────────

func TestPlanetaryPicture(t *testing.T) {
	tests := []struct {
		a, b, c, want float64
	}{
		{100, 200, 50, 250},   // 100+200-50 = 250
		{350, 20, 10, 0},      // wraparound: 350+20-10 = 360 → 0
		{0, 0, 0, 0},
		{180, 180, 180, 180},  // 180+180-180 = 180
		{300, 100, 50, 350},   // 300+100-50 = 350
	}
	for _, tc := range tests {
		got := PlanetaryPicture(tc.a, tc.b, tc.c)
		if math.Abs(got-tc.want) > 0.01 {
			t.Errorf("PlanetaryPicture(%.0f, %.0f, %.0f) = %.4f, want %.0f", tc.a, tc.b, tc.c, got, tc.want)
		}
	}
}

// ── ToDial ───────────────────────────────────────────────────────────────

func TestToDial(t *testing.T) {
	tests := []struct {
		lon, want float64
	}{
		{0, 0},
		{45, 45},
		{90, 0},
		{135, 45},
		{180, 0},
		{225, 45},
		{270, 0},
		{315, 45},
		{360, 0},
		{359.5, 89.5},
	}
	for _, tc := range tests {
		got := ToDial(tc.lon)
		if math.Abs(got-tc.want) > 0.01 {
			t.Errorf("ToDial(%.1f) = %.4f, want %.1f", tc.lon, got, tc.want)
		}
	}
}

// ── GetDialPosition ──────────────────────────────────────────────────────

func TestGetDialPosition(t *testing.T) {
	dp := GetDialPosition(125.5)
	if dp.Sign != "Leo" {
		t.Errorf("125.5° sign = %s, want Leo", dp.Sign)
	}
	if math.Abs(dp.SignDeg-5.5) > 0.01 {
		t.Errorf("125.5° sign deg = %.4f, want 5.5", dp.SignDeg)
	}
	if math.Abs(dp.DialDeg-35.5) > 0.01 {
		t.Errorf("125.5° dial = %.4f, want 35.5", dp.DialDeg)
	}
}

// ── DialSort ─────────────────────────────────────────────────────────────

func TestDialSort(t *testing.T) {
	factors := map[string]float64{
		"Sun":   45,
		"Moon":  90,
		"Mars":  0,
		"Venus": 45.5,
	}
	entries := DialSort(factors)
	if len(entries) != 4 {
		t.Fatalf("expected 4 entries, got %d", len(entries))
	}
	// Mars(0) < Sun(45) < Venus(45.5) < Moon(0 on dial, but 90 mod 90 = 0)
	// Actually Moon at 90 → dial 0, so Moon(0) < Mars(0) alphabetically
	if entries[0].Name != "Mars" && entries[0].Name != "Moon" {
		t.Errorf("first should be Mars or Moon (dial 0), got %s", entries[0].Name)
	}
	// Verify sorted by dial then name
	for i := 1; i < len(entries); i++ {
		if entries[i].DialDeg < entries[i-1].DialDeg {
			t.Errorf("DialSort not sorted by dial: %v", entries)
		}
		if entries[i].DialDeg == entries[i-1].DialDeg && entries[i].Name < entries[i-1].Name {
			t.Errorf("DialSort not sorted by name on tie: %v", entries)
		}
	}
}

// ── FindMidpointPictures ─────────────────────────────────────────────────

func TestFindMidpointPictures(t *testing.T) {
	// Sun/Moon midpoint = 45°, Mars at 45° → 4th harmonic picture
	factors := map[string]float64{
		"Sun":  0,
		"Moon": 90,
		"Mars": 45,
	}
	pics := FindMidpointPictures(factors, 1.0, true)
	if len(pics) == 0 {
		t.Fatal("expected at least one midpoint picture")
	}
	// Should find Sun/Moon = Mars at 4th harmonic
	found := false
	for _, p := range pics {
		if p.FactorA == "Mars" && p.FactorB == "Moon" && p.Activator == "Sun" {
			continue // skip permutations
		}
		if (p.FactorA == "Sun" && p.FactorB == "Moon" && p.Activator == "Mars") ||
			(p.FactorA == "Moon" && p.FactorB == "Sun" && p.Activator == "Mars") {
			if p.Harmonic == "4th" && p.Orb < 0.01 {
				found = true
			}
		}
	}
	if !found {
		t.Errorf("did not find Sun/Moon = Mars at 4th harmonic: %+v", pics)
	}
}

func TestFindMidpointPicturesExcludeCuspCusp(t *testing.T) {
	factors := map[string]float64{
		"H1": 0,
		"H2": 90,
		"H3": 45,
	}
	pics := FindMidpointPictures(factors, 1.0, true)
	if len(pics) != 0 {
		t.Errorf("cusp-cusp-cusp should be excluded, got %d pictures", len(pics))
	}

	// With excludeCuspCusp=false, should find them
	picsAll := FindMidpointPictures(factors, 1.0, false)
	if len(picsAll) == 0 {
		t.Error("expected cusp-cusp pictures when excludeCuspCusp=false")
	}
}

func TestFindMidpointPictures8thHarmonic(t *testing.T) {
	// Sun/Moon midpoint = 45°, Mars at 90° → 45 offset on dial = 8th harmonic
	factors := map[string]float64{
		"Sun":  0,
		"Moon": 90,
		"Mars": 90,
	}
	pics := FindMidpointPictures(factors, 1.0, true)
	found := false
	for _, p := range pics {
		if (p.FactorA == "Sun" && p.FactorB == "Moon" && p.Activator == "Mars") ||
			(p.FactorA == "Moon" && p.FactorB == "Sun" && p.Activator == "Mars") {
			if p.Harmonic == "8th" {
				found = true
			}
		}
	}
	if !found {
		t.Errorf("did not find Sun/Moon = Mars at 8th harmonic: %+v", pics)
	}
}

// ── FindTightPictures ────────────────────────────────────────────────────

func TestFindTightPictures(t *testing.T) {
	factors := map[string]float64{
		"Sun":  0,
		"Moon": 90,
		"Mars": 45,
		"Jupiter": 44.2, // 0.8° off → excluded at 0.5 orb
	}
	tight := FindTightPictures(factors, 0.5)
	// Should only find Sun/Moon=Mars (exact), not Jupiter
	for _, p := range tight {
		if p.Activator == "Jupiter" {
			t.Errorf("Jupiter at orb 0.8 should be excluded from tight pictures")
		}
	}
}

// ── FindActivations ──────────────────────────────────────────────────────

func TestFindActivations(t *testing.T) {
	factors := map[string]float64{
		"Sun":  0,
		"Moon": 90,
		"Mars": 45,
	}
	// Target: 45° (Mars position)
	targets := []float64{45}
	acts := FindActivations(targets, factors, 1.0)
	if len(acts) == 0 {
		t.Fatal("expected at least one activation")
	}
	// Sun+Moon-Mars = 0+90-45 = 45 → activates Mars
	// (ordering may vary since keys are sorted alphabetically)
	found := false
	for _, a := range acts {
		// Check the set {Sun, Moon} → Mars, regardless of A/B order
		ab := map[string]bool{a.FactorA: true, a.FactorB: true}
		if ab["Sun"] && ab["Moon"] && a.FactorC == "Mars" {
			if math.Abs(a.PictureLon-45) < 0.01 {
				found = true
			}
		}
	}
	if !found {
		t.Errorf("did not find Sun+Moon-Mars activating Mars: %+v", acts)
	}
}

// ── AllPicturesForTarget ─────────────────────────────────────────────────

func TestAllPicturesForTarget(t *testing.T) {
	factors := map[string]float64{
		"Sun":  0,
		"Moon": 90,
		"Mars": 45,
	}
	matches := AllPicturesForTarget(45, factors, 1.0)
	if len(matches) == 0 {
		t.Fatal("expected at least one picture match")
	}
	// Sun+Moon-Mars = 45 (ordering may vary)
	found := false
	for _, m := range matches {
		ab := map[string]bool{m.FactorA: true, m.FactorB: true}
		if ab["Sun"] && ab["Moon"] && m.FactorC == "Mars" {
			found = true
		}
	}
	if !found {
		t.Errorf("did not find Sun+Moon-Mars = 45: %+v", matches)
	}
}

// ── isHouseCusp ──────────────────────────────────────────────────────────

func TestIsHouseCusp(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"H1", true},
		{"H12", true},
		{"H0", false},   // no house 0
		{"H13", false},  // no house 13
		{"Sun", false},
		{"Ascendant", false},
		{"H", false},    // too short
		{"", false},
	}
	for _, tc := range tests {
		got := isHouseCusp(tc.name)
		if got != tc.want {
			t.Errorf("isHouseCusp(%q) = %v, want %v", tc.name, got, tc.want)
		}
	}
}

// ── ComputeUranianReport ─────────────────────────────────────────────────

func TestComputeUranianReport(t *testing.T) {
	planets := map[string]float64{
		"Sun": 0, "Moon": 90, "Mars": 45,
	}
	houses := map[string]float64{
		"H1": 0, "H2": 30, "H3": 60,
	}
	report := ComputeUranianReport("test", planets, houses)
	if report.Name != "test" {
		t.Errorf("name = %s, want test", report.Name)
	}
	if len(report.DialPositions) != 6 {
		t.Errorf("expected 6 dial positions (3 planets + 3 houses), got %d", len(report.DialPositions))
	}
	if report.MidpointPictures == nil {
		t.Error("MidpointPictures should be non-nil (empty slice, not nil)")
	}
	if report.TightPictures == nil {
		t.Error("TightPictures should be non-nil")
	}
	if report.Activations == nil {
		t.Error("Activations should be non-nil")
	}
}

// ── Edge cases ───────────────────────────────────────────────────────────

func TestEmptyInputs(t *testing.T) {
	// Empty planets should not panic
	report := ComputeUranianReport("empty", map[string]float64{}, map[string]float64{})
	if len(report.DialPositions) != 0 {
		t.Errorf("empty input should produce 0 dial positions")
	}
	if len(report.MidpointPictures) != 0 {
		t.Errorf("empty input should produce 0 midpoint pictures")
	}

	// Empty factors in FindMidpointPictures
	pics := FindMidpointPictures(map[string]float64{}, 1.0, true)
	if len(pics) != 0 {
		t.Errorf("empty factors should produce 0 pictures")
	}

	// Empty factors in FindActivations
	acts := FindActivations([]float64{}, map[string]float64{}, 1.0)
	if len(acts) != 0 {
		t.Errorf("empty factors should produce 0 activations")
	}
}
