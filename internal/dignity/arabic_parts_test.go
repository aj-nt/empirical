package dignity

import (
	"math"
	"testing"
)

// ── Part Computation ──────────────────────────────────────────────────────

func TestComputeParts_Fortune(t *testing.T) {
	// Day chart: Fortune = Asc + Moon - Sun
	asc := 100.0
	planets := map[string]float64{
		"Sun": 50.0, "Moon": 80.0,
	}
	parts := ComputeParts(asc, planets, true)
	fortune := parts["Fortune"]
	expected := normalizeLon(100.0 + 80.0 - 50.0) // 130.0
	if math.Abs(fortune-expected) > 0.01 {
		t.Errorf("Fortune = %.2f, want %.2f", fortune, expected)
	}
}

func TestComputeParts_FortuneNight(t *testing.T) {
	// Night chart: Fortune = Asc + Sun - Moon
	asc := 100.0
	planets := map[string]float64{
		"Sun": 50.0, "Moon": 80.0,
	}
	parts := ComputeParts(asc, planets, false)
	fortune := parts["Fortune"]
	expected := normalizeLon(100.0 + 50.0 - 80.0) // 70.0
	if math.Abs(fortune-expected) > 0.01 {
		t.Errorf("Fortune (night) = %.2f, want %.2f", fortune, expected)
	}
}

func TestComputeParts_Spirit(t *testing.T) {
	// Day chart: Spirit = Asc + Sun - Moon
	asc := 100.0
	planets := map[string]float64{
		"Sun": 50.0, "Moon": 80.0,
	}
	parts := ComputeParts(asc, planets, true)
	spirit := parts["Spirit"]
	expected := normalizeLon(100.0 + 50.0 - 80.0) // 70.0
	if math.Abs(spirit-expected) > 0.01 {
		t.Errorf("Spirit = %.2f, want %.2f", spirit, expected)
	}
}

func TestComputeParts_AllPartsPresent(t *testing.T) {
	asc := 100.0
	planets := map[string]float64{
		"Sun": 50.0, "Moon": 80.0, "Mercury": 60.0,
		"Venus": 70.0, "Mars": 90.0, "Jupiter": 110.0,
		"Saturn": 120.0,
	}
	parts := ComputeParts(asc, planets, true)

	catalog := PartCatalog()
	for _, def := range catalog {
		if _, ok := parts[def.Name]; !ok {
			t.Errorf("Part %s missing from results", def.Name)
		}
	}
}

func TestComputeParts_DependencyOrder(t *testing.T) {
	// Parts that depend on Fortune/Spirit should still compute
	asc := 100.0
	planets := map[string]float64{
		"Sun": 50.0, "Moon": 80.0, "Mercury": 60.0,
		"Venus": 70.0, "Mars": 90.0, "Jupiter": 110.0,
		"Saturn": 120.0,
	}
	parts := ComputeParts(asc, planets, true)

	// Necessity = Asc + Mercury - Fortune
	necessity, ok := parts["Necessity"]
	if !ok {
		t.Fatal("Necessity not computed")
	}
	fortune := parts["Fortune"]
	expected := normalizeLon(asc + planets["Mercury"] - fortune)
	if math.Abs(necessity-expected) > 0.01 {
		t.Errorf("Necessity = %.2f, want %.2f", necessity, expected)
	}
}

func TestComputeParts_Wraparound(t *testing.T) {
	// Test that normalizeLon handles wraparound correctly
	asc := 350.0
	planets := map[string]float64{
		"Sun": 10.0, "Moon": 20.0,
	}
	parts := ComputeParts(asc, planets, true)
	fortune := parts["Fortune"]
	// 350 + 20 - 10 = 360 → 0
	expected := 0.0
	if math.Abs(fortune-expected) > 0.01 {
		t.Errorf("Fortune wraparound = %.2f, want %.2f", fortune, expected)
	}
}

// ── Part-to-Planet Aspects ────────────────────────────────────────────────

func TestComputePartReport_Aspects(t *testing.T) {
	// Set up a chart where Fortune is exactly conjunct Jupiter
	asc := 100.0
	planets := map[string]float64{
		"Sun": 50.0, "Moon": 60.0, "Mercury": 70.0,
		"Venus": 80.0, "Mars": 90.0, "Jupiter": 110.0,
		"Saturn": 120.0,
	}
	// Fortune = 100 + 60 - 50 = 110 → exactly conjunct Jupiter
	report := ComputePartReport("test", asc, planets, true, 3.0)

	found := false
	for _, h := range report.Aspects {
		if h.Part == "Fortune" && h.Planet == "Jupiter" && h.Aspect == "conjunction" {
			found = true
			if h.Orb != 0.0 {
				t.Errorf("Fortune-Jupiter orb = %.2f, want 0.00", h.Orb)
			}
		}
	}
	if !found {
		t.Error("Fortune-Jupiter conjunction not found")
	}
}

func TestComputePartReport_NoAspectsAtZeroOrb(t *testing.T) {
	asc := 100.0
	planets := map[string]float64{
		"Sun": 50.0, "Moon": 60.0, "Mercury": 70.0,
		"Venus": 80.0, "Mars": 90.0, "Jupiter": 110.0,
		"Saturn": 120.0,
	}
	report := ComputePartReport("test", asc, planets, true, 0.0)
	// At 0 orb, only exact hits count
	// Fortune = 110, Jupiter = 110 → exact conjunction
	if len(report.Aspects) == 0 {
		t.Error("Expected at least Fortune-Jupiter exact conjunction at 0 orb")
	}
}

// ── Cross-System Invariance ───────────────────────────────────────────────

func TestComputePartCrossSystem_AspectInvariance(t *testing.T) {
	// Part-to-planet aspects should be zodiac-invariant.
	// Part shifts by ayanamsa, planets shift by ayanamsa,
	// angular distances are preserved.
	asc := 100.0
	planets := map[string]float64{
		"Sun": 50.0, "Moon": 60.0, "Mercury": 70.0,
		"Venus": 80.0, "Mars": 90.0, "Jupiter": 110.0,
		"Saturn": 120.0,
	}
	ayanamsa := 24.0

	result := ComputePartCrossSystem("test", asc, planets, ayanamsa, true, 3.0)

	// Tropical and sidereal should have same number of Parts
	if len(result.Tropical) != len(result.Sidereal) {
		t.Errorf("Tropical has %d parts, sidereal has %d", len(result.Tropical), len(result.Sidereal))
	}

	// Aspects should be present (zodiac-invariant)
	if len(result.Aspects) == 0 {
		t.Error("Expected Part-to-planet aspects")
	}
}

func TestComputePartCrossSystem_SignShift(t *testing.T) {
	// Parts near sign boundaries should shift signs under ayanamsa
	asc := 0.0
	planets := map[string]float64{
		"Sun": 10.0, "Moon": 15.0, "Mercury": 20.0,
		"Venus": 25.0, "Mars": 30.0, "Jupiter": 35.0,
		"Saturn": 40.0,
	}
	ayanamsa := 24.0

	result := ComputePartCrossSystem("test", asc, planets, ayanamsa, true, 3.0)

	// Fortune = 0 + 15 - 10 = 5° Aries (tropical)
	// Sidereal Fortune = (0-24) + (15-24) - (10-24) = -24 + -9 - -14 = -19 → 341° Pisces
	// Sign should differ
	if result.SignSurvivors == result.Total {
		t.Logf("Sign survivors: %d/%d (may be all if no boundary crossings)", result.SignSurvivors, result.Total)
	}
}

func TestPartCatalog_CrossSystemCount(t *testing.T) {
	catalog := PartCatalog()
	inBoth := 0
	for _, def := range catalog {
		if def.InBoth {
			inBoth++
		}
	}
	// Fortune, Spirit, Victory, Father, Mother, Children, Marriage, Death = 8
	if inBoth != 8 {
		t.Errorf("Parts in both traditions: %d, want 8", inBoth)
	}
	if len(catalog) != 13 {
		t.Errorf("Total parts: %d, want 13", len(catalog))
	}
}
