package swe

import (
	"math"
	"os"
	"path/filepath"
	"testing"
)

// initEphe sets up the ephemeris path for tests.
// Uses the embedded cache at ~/.cache/empirical/ephe/.
func initEphe(t *testing.T) {
	t.Helper()
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("cannot find home dir: %v", err)
	}
	cacheDir := filepath.Join(home, ".cache", "empirical", "ephe")
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		t.Skipf("ephemeris cache not found at %s — run the server once first", cacheDir)
	}
	SetEphePath(cacheDir)
	SetSidMode(SIDM_LAHIRI, 0, 0)
}

// ── Julday ───────────────────────────────────────────────────────────────

func TestJulday(t *testing.T) {
	// J2000.0 = 2000-01-01 12:00 UT = JD 2451545.0
	jd := Julday(2000, 1, 1, 12.0, true)
	if math.Abs(jd-2451545.0) > 0.01 {
		t.Errorf("J2000.0: got %.4f, want 2451545.0", jd)
	}

	// 1969-02-15 07:10 UT (AJ's birth, PST+8 = 23:10 PST → 07:10 UT next day? No — 23:10 PST Feb 15 = 07:10 UT Feb 16)
	// Actually: 23:10 PST = 23:10 + 8:00 = 07:10 UT on Feb 16
	jd = Julday(1969, 2, 16, 7.1667, true)
	if jd < 2440260 || jd > 2440270 {
		t.Errorf("AJ birth JD out of range: %.4f", jd)
	}
}

func TestJuldayJulian(t *testing.T) {
	// Julian calendar: 1900-01-01 12:00
	jd := Julday(1900, 1, 1, 12.0, false)
	if jd < 2415000 || jd > 2415100 {
		t.Errorf("1900-01-01 Julian JD out of range: %.4f", jd)
	}
}

// ── Revjul ───────────────────────────────────────────────────────────────

func TestRevjulRoundtrip(t *testing.T) {
	tests := []struct{ y, m, d int; h float64 }{
		{2000, 1, 1, 12.0},
		{1969, 2, 16, 7.1667},
		{2026, 6, 15, 0.0},
	}
	for _, tc := range tests {
		jd := Julday(tc.y, tc.m, tc.d, tc.h, true)
		year, month, day, hour := Revjul(jd)
		if year != tc.y || month != tc.m || day != tc.d {
			t.Errorf("Revjul roundtrip: got %d-%02d-%02d, want %d-%02d-%02d",
				year, month, day, tc.y, tc.m, tc.d)
		}
		if math.Abs(hour-tc.h) > 0.02 {
			t.Errorf("Revjul hour: got %.4f, want %.4f", hour, tc.h)
		}
	}
}

// ── CalcUT ───────────────────────────────────────────────────────────────

func TestCalcUTSun(t *testing.T) {
	initEphe(t)
	// J2000.0: Sun at ~280.46° ecliptic longitude
	jd := Julday(2000, 1, 1, 12.0, true)
	lon, lat, dist, speed := CalcUT(jd, SUN)
	if math.Abs(lon-280.46) > 0.5 {
		t.Errorf("Sun at J2000.0: lon=%.4f, want ~280.46", lon)
	}
	if math.Abs(lat) > 0.01 {
		t.Errorf("Sun latitude should be near 0: got %.4f", lat)
	}
	if dist < 0.98 || dist > 0.99 {
		t.Errorf("Sun distance at J2000.0: %.6f AU, want ~0.983", dist)
	}
	if speed < 1.0 || speed > 1.03 {
		t.Errorf("Sun speed: %.4f deg/day, want ~1.02", speed)
	}
}

func TestCalcUTMoon(t *testing.T) {
	initEphe(t)
	jd := Julday(2000, 1, 1, 12.0, true)
	lon, _, _, speed := CalcUT(jd, MOON)
	if lon < 0 || lon > 360 {
		t.Errorf("Moon longitude out of range: %.4f", lon)
	}
	if speed < 11 || speed > 15 {
		t.Errorf("Moon speed: %.4f deg/day, want ~13.2", speed)
	}
}

func TestCalcUTAllPlanets(t *testing.T) {
	initEphe(t)
	jd := Julday(2000, 1, 1, 12.0, true)
	planets := []int{SUN, MOON, MERCURY, VENUS, MARS, JUPITER, SATURN, URANUS, NEPTUNE, PLUTO}
	for _, p := range planets {
		lon, _, _, _ := CalcUT(jd, p)
		if lon < 0 || lon > 360 {
			t.Errorf("planet %d longitude out of range: %.4f", p, lon)
		}
	}
}

func TestCalcUTChiron(t *testing.T) {
	initEphe(t)
	jd := Julday(2000, 1, 1, 12.0, true)
	lon, _, _, _ := CalcUT(jd, CHIRON)
	if lon < 0 || lon > 360 {
		t.Errorf("Chiron longitude out of range: %.4f", lon)
	}
}

func TestCalcUTMeanNode(t *testing.T) {
	initEphe(t)
	jd := Julday(2000, 1, 1, 12.0, true)
	lon, _, _, _ := CalcUT(jd, MEAN_NODE)
	if lon < 0 || lon > 360 {
		t.Errorf("Mean Node longitude out of range: %.4f", lon)
	}
}

// ── Houses ───────────────────────────────────────────────────────────────

func TestHouses(t *testing.T) {
	initEphe(t)
	// AJ's birth: 1969-02-15 23:10 PST = 1969-02-16 07:10 UT
	// Olympia WA: 47.04N, 122.90W
	jd := Julday(1969, 2, 16, 7.1667, true)
	cusps, ascmc := Houses(jd, 47.04, -122.90, 'P') // Placidus
	if len(cusps) != 13 {
		t.Errorf("expected 13 cusps, got %d", len(cusps))
	}
	if len(ascmc) != 10 {
		t.Errorf("expected 10 ascmc values, got %d", len(ascmc))
	}
	// ASC should be in early Scorpio (~210-240°)
	asc := ascmc[0]
	if asc < 200 || asc > 250 {
		t.Errorf("ASC for AJ: %.4f, expected ~210-240 (Scorpio)", asc)
	}
	// MC should be in Leo (~120-150°)
	mc := ascmc[1]
	if mc < 110 || mc > 160 {
		t.Errorf("MC for AJ: %.4f, expected ~120-150 (Leo)", mc)
	}
	// Cusps should be in ascending order (mod 360)
	for i := 1; i <= 12; i++ {
		if cusps[i] < 0 || cusps[i] > 360 {
			t.Errorf("cusp %d out of range: %.4f", i, cusps[i])
		}
	}
}

func TestHousesWholeSign(t *testing.T) {
	initEphe(t)
	jd := Julday(2000, 1, 1, 12.0, true)
	cusps, ascmc := Houses(jd, 51.5, -0.1, 'W') // London, Whole Sign
	if len(cusps) != 13 {
		t.Errorf("expected 13 cusps, got %d", len(cusps))
	}
	// Whole sign: ASC determines house 1 sign, each house = 30° of that sign
	asc := ascmc[0]
	if asc < 0 || asc > 360 {
		t.Errorf("ASC out of range: %.4f", asc)
	}
}

// ── GetAyanamsaUT ────────────────────────────────────────────────────────

func TestGetAyanamsaUT(t *testing.T) {
	initEphe(t)
	jd := Julday(2000, 1, 1, 12.0, true)
	ayan := GetAyanamsaUT(jd)
	// Lahiri ayanamsa at J2000.0 ≈ 23.86°
	if math.Abs(ayan-23.86) > 0.5 {
		t.Errorf("Lahiri ayanamsa at J2000.0: %.4f, want ~23.86", ayan)
	}
}

func TestGetAyanamsaUT1969(t *testing.T) {
	initEphe(t)
	jd := Julday(1969, 2, 16, 7.1667, true)
	ayan := GetAyanamsaUT(jd)
	// Lahiri ayanamsa at 1969 ≈ 23.43°
	if math.Abs(ayan-23.43) > 0.5 {
		t.Errorf("Lahiri ayanamsa at 1969: %.4f, want ~23.43", ayan)
	}
}

// ── Fixstar ──────────────────────────────────────────────────────────────

func TestFixstarSirius(t *testing.T) {
	initEphe(t)
	jd := Julday(2000, 1, 1, 12.0, true)
	lon, lat, _, _ := Fixstar("Sirius", jd)
	// Sirius at J2000.0 ≈ 104° (14° Cancer)
	if lon < 100 || lon > 110 {
		t.Errorf("Sirius longitude: %.4f, want ~104", lon)
	}
	// Sirius has significant southern latitude (~-39°)
	if lat > -30 {
		t.Errorf("Sirius latitude: %.4f, should be < -30 (southern)", lat)
	}
}

func TestFixstarAldebaran(t *testing.T) {
	initEphe(t)
	jd := Julday(2000, 1, 1, 12.0, true)
	lon, _, _, _ := Fixstar("Aldebaran", jd)
	// Aldebaran at J2000.0 ≈ 69.8° (9° Gemini)
	if lon < 65 || lon > 75 {
		t.Errorf("Aldebaran longitude: %.4f, want ~70", lon)
	}
}

func TestFixstarNotFound(t *testing.T) {
	initEphe(t)
	jd := Julday(2000, 1, 1, 12.0, true)
	lon, _, _, _ := Fixstar("NotARealStar", jd)
	if lon != 0 {
		t.Errorf("unknown star should return 0 lon, got %.4f", lon)
	}
}

// ── Close ────────────────────────────────────────────────────────────────

func TestClose(t *testing.T) {
	initEphe(t)
	// Close should not panic
	Close()
}
