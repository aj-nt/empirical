package mundane

import (
	"testing"
	"time"

	"github.com/aj-nt/empirical"
	"github.com/aj-nt/empirical/internal/dignity"
	"github.com/aj-nt/empirical/internal/swe"
)

func init() {
	cacheDir, err := empirical.EnsureEpheCache()
	if err != nil {
		panic(err)
	}
	swe.SetEphePath(cacheDir)
}

// realSWECompute wraps the actual Swiss Ephemeris for integration testing.
func realSWECompute(year, month, day int, hour float64, planetID int) (float64, float64, float64, float64) {
	jd := swe.Julday(year, month, day, hour, true)
	return swe.CalcUT(jd, planetID)
}

// realSWEHouses wraps the actual Swiss Ephemeris houses call.
func realSWEHouses(jd, lat, lon float64, hsys byte) ([13]float64, [10]float64) {
	return swe.Houses(jd, lat, lon, hsys)
}

func TestRealSWE_SolarIngresses2026(t *testing.T) {
	start := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 3, 31, 0, 0, 0, 0, time.UTC)

	ingresses, err := FindSolarIngresses(start, end, realSWECompute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ingresses) != 1 {
		t.Fatalf("expected 1 Aries ingress in March 2026, got %d", len(ingresses))
	}

	ing := ingresses[0]
	if ing.Sign != "Aries" {
		t.Errorf("expected Aries, got %s", ing.Sign)
	}
	if ing.Time.Year() != 2026 || ing.Time.Month() != 3 || ing.Time.Day() != 20 {
		t.Errorf("expected March 20 2026, got %s", ing.Time.Format("2006-01-02 15:04"))
	}

	t.Logf("Aries ingress 2026: %s UTC, lon=%.4f", ing.Time.Format("2006-01-02 15:04:05"), ing.Lon)
}

func TestRealSWE_Lunations2026(t *testing.T) {
	start := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC)

	lunations, err := FindLunations(start, end, realSWECompute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(lunations) == 0 {
		t.Fatal("expected at least 1 lunation in June 2026")
	}

	for _, l := range lunations {
		t.Logf("%s: %s UTC", l.Type, l.Time.Format("2006-01-02 15:04:05"))
	}
}

func TestRealSWE_Eclipses2026(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)

	eclipses, err := FindEclipses(start, end, realSWECompute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	t.Logf("Eclipses in 2026: %d", len(eclipses))
	for _, e := range eclipses {
		t.Logf("  %s: %s UTC", e.Type, e.Time.Format("2006-01-02 15:04:05"))
	}

	if len(eclipses) < 2 {
		t.Errorf("expected at least 2 eclipses in 2026, got %d", len(eclipses))
	}
}

func TestRealSWE_SaturnJupiterConjunction(t *testing.T) {
	start := time.Date(2020, 12, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2020, 12, 31, 0, 0, 0, 0, time.UTC)

	conjunctions, err := FindConjunctions(start, end, 6, "Saturn", 5, "Jupiter", realSWECompute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(conjunctions) != 1 {
		t.Fatalf("expected 1 Saturn-Jupiter conjunction in Dec 2020, got %d", len(conjunctions))
	}

	c := conjunctions[0]
	t.Logf("Saturn-Jupiter conjunction: %s UTC", c.Time.Format("2006-01-02 15:04:05"))

	if c.Time.Year() != 2020 || c.Time.Month() != 12 || c.Time.Day() != 21 {
		t.Errorf("expected Dec 21 2020, got %s", c.Time.Format("2006-01-02"))
	}
}

func TestRealSWE_CastAriesIngress2026(t *testing.T) {
	start := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 3, 31, 0, 0, 0, 0, time.UTC)
	ingresses, err := FindSolarIngresses(start, end, realSWECompute)
	if err != nil {
		t.Fatalf("FindSolarIngresses: %v", err)
	}
	if len(ingresses) != 1 {
		t.Fatalf("expected 1 ingress, got %d", len(ingresses))
	}

	chart, err := CastIngressChart(ingresses[0], 38.9, -77.0, realSWECompute, realSWEHouses)
	if err != nil {
		t.Fatalf("CastIngressChart: %v", err)
	}

	t.Logf("Aries Ingress 2026 — Washington DC")
	t.Logf("  Time: %s UTC", chart.Time.Format("2006-01-02 15:04:05"))
	t.Logf("  ASC: %.2f", chart.ASC)
	t.Logf("  MC:  %.2f", chart.MC)

	if len(chart.Planets) < 10 {
		t.Errorf("expected at least 10 planets, got %d", len(chart.Planets))
	}
	if chart.ASC == 0 {
		t.Error("ASC should not be zero")
	}
	if chart.MC == 0 {
		t.Error("MC should not be zero")
	}
	// Sun should be at 0° Aries (or very close; 360.0 == 0.0)
	sunLon := chart.Planets["Sun"]
	if sunLon > 0.5 && sunLon < 359.5 {
		t.Errorf("Sun should be near 0° Aries, got %.4f", sunLon)
	}
}

func TestRealSWE_CastNewMoonChart(t *testing.T) {
	start := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 7, 31, 0, 0, 0, 0, time.UTC)
	lunations, err := FindLunations(start, end, realSWECompute)
	if err != nil {
		t.Fatalf("FindLunations: %v", err)
	}

	var newMoon LunationEvent
	for _, l := range lunations {
		if l.Type == "New Moon" {
			newMoon = l
			break
		}
	}
	if newMoon.Time.IsZero() {
		t.Fatal("no new moon found in July 2026")
	}

	chart, err := CastLunationChart(newMoon, 51.5, -0.1, realSWECompute, realSWEHouses)
	if err != nil {
		t.Fatalf("CastLunationChart: %v", err)
	}

	t.Logf("New Moon July 2026 — London")
	t.Logf("  Time: %s UTC", chart.Time.Format("2006-01-02 15:04:05"))
	t.Logf("  ASC: %.2f", chart.ASC)
	t.Logf("  MC:  %.2f", chart.MC)

	sunLon := chart.Planets["Sun"]
	moonLon := chart.Planets["Moon"]
	diff := moonLon - sunLon
	if diff < 0 {
		diff = -diff
	}
	if diff > 360 {
		diff -= 360
	}
	if diff > 1.0 {
		t.Errorf("Sun-Moon separation should be < 1° at new moon, got %.4f", diff)
	}
}

func TestRealSWE_CastUSChart(t *testing.T) {
	us, ok := NationalChart("United States")
	if !ok {
		t.Fatal("US chart not found")
	}

	chart, err := CastChart(
		time.Date(us.Year, time.Month(us.Month), us.Day,
			int(us.Hour), int((us.Hour-float64(int(us.Hour)))*60), 0, 0, time.UTC),
		us.Lat, us.Lon, realSWECompute, realSWEHouses, 'W',
	)
	if err != nil {
		t.Fatalf("CastChart: %v", err)
	}

	t.Logf("United States natal chart (Sibly)")
	t.Logf("  Date: %d-%02d-%02d %.2f UT", us.Year, us.Month, us.Day, us.Hour)
	t.Logf("  ASC: %.2f", chart.ASC)
	t.Logf("  MC:  %.2f", chart.MC)
	t.Logf("  Sun: %.2f (%s)", chart.Planets["Sun"], dignity.SignForLongitude(chart.Planets["Sun"]))

	// US Sibly chart: ASC should be ~7° Sagittarius (~255°)
	if chart.ASC < 245 || chart.ASC > 265 {
		t.Errorf("US Sibly ASC should be ~7° Sagittarius (~255°), got %.2f", chart.ASC)
	}
}
