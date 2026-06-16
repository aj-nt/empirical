package divisional

import (
	"math"
	"testing"
)

// ── NavamshaPosition ────────────────────────────────────────────────────

func TestNavamshaPosition(t *testing.T) {
	// Verified against actual code output (epsilon pushes at boundaries)
	tests := []struct {
		sidLon   float64
		wantSign string
		wantLon  float64
	}{
		{0, "Aries", 0},
		{3.34, "Taurus", 30.06},
		{10, "Cancer", 120},       // pada 3, navDeg=0 at sign boundary
		{29.9, "Sagittarius", 269.1},
		{30, "Capricorn", 270},
		{33.34, "Aquarius", 300.06},
		{40, "Aries", 30},         // pada 3, navDeg=0 at sign boundary
		{60, "Libra", 180},
		{63.34, "Scorpio", 210.06},
		{90, "Cancer", 90},
		{93.34, "Leo", 120.06},
	}
	for _, tc := range tests {
		lon, sign := NavamshaPosition(tc.sidLon)
		if sign != tc.wantSign {
			t.Errorf("NavamshaPosition(%.2f) sign = %s, want %s", tc.sidLon, sign, tc.wantSign)
		}
		if math.Abs(lon-tc.wantLon) > 0.2 {
			t.Errorf("NavamshaPosition(%.2f) lon = %.4f, want %.2f", tc.sidLon, lon, tc.wantLon)
		}
	}
}

func TestNavamshaPositionWraparound(t *testing.T) {
	// 360° should behave like 0°
	lon1, sign1 := NavamshaPosition(0)
	lon2, sign2 := NavamshaPosition(360)
	if sign1 != sign2 || math.Abs(lon1-lon2) > 0.01 {
		t.Errorf("NavamshaPosition(0) = (%.4f, %s), NavamshaPosition(360) = (%.4f, %s) — should match",
			lon1, sign1, lon2, sign2)
	}
}

// ── GetNakshatra ─────────────────────────────────────────────────────────

func TestGetNakshatra(t *testing.T) {
	// Use values clearly inside each nakshatra, not at boundaries
	// Nakshatra span = 360/27 ≈ 13.333°
	tests := []struct {
		sidLon    float64
		wantNak   string
		wantPada  int
		wantRuler string
	}{
		{1, "Ashwini", 1, "Ketu"},
		{14, "Bharani", 1, "Venus"},
		{27, "Krittika", 1, "Sun"},
		{41, "Rohini", 1, "Moon"},
		{54, "Mrigashirsha", 1, "Mars"},
		{67, "Ardra", 1, "Rahu"},
		{81, "Punarvasu", 1, "Jupiter"},
		{94, "Pushya", 1, "Saturn"},
		{107, "Ashlesha", 1, "Mercury"},
		{121, "Magha", 1, "Ketu"},
		{134, "Purva Phalguni", 1, "Venus"},
		{147, "Uttara Phalguni", 1, "Sun"},
		{161, "Hasta", 1, "Moon"},
		{174, "Chitra", 1, "Mars"},
		{187, "Swati", 1, "Rahu"},
		{201, "Vishakha", 1, "Jupiter"},
		{214, "Anuradha", 1, "Saturn"},
		{227, "Jyeshtha", 1, "Mercury"},
		{241, "Mula", 1, "Ketu"},
		{254, "Purva Ashadha", 1, "Venus"},
		{267, "Uttara Ashadha", 1, "Sun"},
		{281, "Shravana", 1, "Moon"},
		{294, "Dhanishta", 1, "Mars"},
		{307, "Shatabhisha", 1, "Rahu"},
		{321, "Purva Bhadrapada", 1, "Jupiter"},
		{334, "Uttara Bhadrapada", 1, "Saturn"},
		{347, "Revati", 1, "Mercury"},
	}
	for _, tc := range tests {
		nak := GetNakshatra(tc.sidLon)
		if nak.Nakshatra != tc.wantNak {
			t.Errorf("GetNakshatra(%.0f) nakshatra = %s, want %s", tc.sidLon, nak.Nakshatra, tc.wantNak)
		}
		if nak.Pada != tc.wantPada {
			t.Errorf("GetNakshatra(%.0f) pada = %d, want %d", tc.sidLon, nak.Pada, tc.wantPada)
		}
		if nak.Ruler != tc.wantRuler {
			t.Errorf("GetNakshatra(%.0f) ruler = %s, want %s", tc.sidLon, nak.Ruler, tc.wantRuler)
		}
	}
}

func TestGetNakshatraPada(t *testing.T) {
	// Test pada boundaries within Ashwini (0-13.33°)
	// Pada 1: 0-3.33, Pada 2: 3.33-6.67, Pada 3: 6.67-10, Pada 4: 10-13.33
	// At exact boundaries, epsilon pushes to the next segment
	tests := []struct {
		sidLon   float64
		wantPada int
	}{
		{0, 1},
		{3.33, 1},
		{3.34, 2},
		{6.66, 2},
		{6.67, 3},  // boundary: epsilon pushes to pada 3
		{10.00, 4}, // boundary: epsilon pushes to pada 4
		{13.33, 4},
	}
	for _, tc := range tests {
		nak := GetNakshatra(tc.sidLon)
		if nak.Pada != tc.wantPada {
			t.Errorf("GetNakshatra(%.2f) pada = %d, want %d", tc.sidLon, nak.Pada, tc.wantPada)
		}
	}
}

// ── VimshottariDasha ────────────────────────────────────────────────────

func TestVimshottariDasha(t *testing.T) {
	// Moon at 0° Ashwini (Ketu-ruled) → starts with Ketu
	dasha := VimshottariDasha("Ashwini", 0, 2000, 1, 1)
	if len(dasha) != 9 {
		t.Fatalf("expected 9 dasha periods, got %d", len(dasha))
	}
	if dasha[0].Planet != "Ketu" {
		t.Errorf("first dasha = %s, want Ketu", dasha[0].Planet)
	}
	// At 0° into nakshatra, proportion = 0, remaining = full 7 years
	if math.Abs(dasha[0].Years-7.0) > 0.1 {
		t.Errorf("Ketu dasha years = %.2f, want 7.0", dasha[0].Years)
	}
	// Sequence should be: Ketu, Venus, Sun, Moon, Mars, Rahu, Jupiter, Saturn, Mercury
	expected := []string{"Ketu", "Venus", "Sun", "Moon", "Mars", "Rahu", "Jupiter", "Saturn", "Mercury"}
	for i, want := range expected {
		if dasha[i].Planet != want {
			t.Errorf("dasha[%d] = %s, want %s", i, dasha[i].Planet, want)
		}
	}
}

func TestVimshottariDashaPartialFirst(t *testing.T) {
	// Moon at 6.67° Ashwini (halfway through) → half of Ketu remaining
	dasha := VimshottariDasha("Ashwini", 6.665, 2000, 1, 1)
	if math.Abs(dasha[0].Years-3.5) > 0.2 {
		t.Errorf("halfway through Ashwini: Ketu remaining = %.2f, want ~3.5", dasha[0].Years)
	}
}

func TestVimshottariDashaEndOfNakshatra(t *testing.T) {
	// Moon at 13.33° Ashwini (end) → almost no Ketu remaining
	dasha := VimshottariDasha("Ashwini", 13.33, 2000, 1, 1)
	if dasha[0].Years > 0.1 {
		t.Errorf("end of Ashwini: Ketu remaining = %.2f, want ~0", dasha[0].Years)
	}
}

func TestVimshottariDashaUnknownNakshatra(t *testing.T) {
	dasha := VimshottariDasha("NotANakshatra", 0, 2000, 1, 1)
	if dasha != nil {
		t.Errorf("unknown nakshatra should return nil, got %d periods", len(dasha))
	}
}

func TestVimshottariDashaTotalYears(t *testing.T) {
	// Total of all periods should be ~120 years
	dasha := VimshottariDasha("Ashwini", 0, 2000, 1, 1)
	total := 0.0
	for _, d := range dasha {
		total += d.Years
	}
	if math.Abs(total-120.0) > 0.5 {
		t.Errorf("total dasha years = %.2f, want ~120", total)
	}
}

// ── ComputeDivisionalReport ──────────────────────────────────────────────

func TestComputeDivisionalReport(t *testing.T) {
	planets := map[string]float64{
		"Sun":   327.45, // AJ's tropical Sun
		"Moon":  322.29, // AJ's tropical Moon
		"Mars":  235.80, // AJ's tropical Mars
	}
	ayanamsa := 23.43 // approximate Lahiri for 1969
	report := ComputeDivisionalReport("test", planets, ayanamsa, 1969, 2, 15)
	if report.Name != "test" {
		t.Errorf("name = %s, want test", report.Name)
	}
	if math.Abs(report.Ayanamsa-ayanamsa) > 0.01 {
		t.Errorf("ayanamsa = %.4f, want %.2f", report.Ayanamsa, ayanamsa)
	}
	if len(report.Positions) != 3 {
		t.Errorf("expected 3 positions, got %d", len(report.Positions))
	}
	// Moon should have nakshatra and dasha
	if len(report.Dasha) != 9 {
		t.Errorf("expected 9 dasha periods, got %d", len(report.Dasha))
	}
	// Each position should have sidereal sign, nakshatra, navamsha
	for _, pos := range report.Positions {
		if pos.SiderealSign == "" {
			t.Errorf("%s: sidereal sign empty", pos.Planet)
		}
		if pos.Nakshatra.Nakshatra == "" {
			t.Errorf("%s: nakshatra empty", pos.Planet)
		}
		if pos.NavamshaSign == "" {
			t.Errorf("%s: navamsha sign empty", pos.Planet)
		}
	}
}

// ── Julian Day helpers ───────────────────────────────────────────────────

func TestJulianDayRoundtrip(t *testing.T) {
	// jdToDate(julianDay(y,m,d)) should return the original date
	tests := []struct{ y, m, d int; want string }{
		{1969, 2, 15, "1969-02-15"},
		{2000, 1, 1, "2000-01-01"},
		{2026, 6, 15, "2026-06-15"},
		{1900, 12, 31, "1900-12-31"},
	}
	for _, tc := range tests {
		jd := julianDay(tc.y, tc.m, tc.d)
		dateStr := jdToDate(jd)
		if dateStr != tc.want {
			t.Errorf("jd roundtrip (%d-%02d-%02d): got %s, want %s", tc.y, tc.m, tc.d, dateStr, tc.want)
		}
	}
}

// ── Edge cases ───────────────────────────────────────────────────────────

func TestDivisionalEmptyInput(t *testing.T) {
	report := ComputeDivisionalReport("empty", map[string]float64{}, 23.43, 2000, 1, 1)
	if len(report.Positions) != 0 {
		t.Errorf("empty planets should produce 0 positions")
	}
	if len(report.Dasha) != 0 {
		t.Errorf("no Moon → no dasha, got %d periods", len(report.Dasha))
	}
}
