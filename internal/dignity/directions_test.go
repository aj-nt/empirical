package dignity

import (
	"math"
	"testing"
)

// ============================================================================
// Primary Directions — Tests
// ============================================================================

func TestPrimaryDirections_Equator_ASC(t *testing.T) {
	// At latitude 0°, OA = RA. The math is verifiable by hand.
	// ASC at 0° Aries, age 30.
	// RA_directed = 30°
	// tan(λ) = tan(RA) / cos(ε) = tan(30°) / cos(23.44°)
	//        = 0.57735 / 0.91741 = 0.62932
	// λ = atan(0.62932) = 32.18° = 2°11' Taurus

	natal := map[string]float64{
		"Sun": 32.0, // near directed ASC → conjunction
	}
	ascLon := 0.0
	mcLon := 90.0
	lat := 0.0
	age := 30.0

	aspects := DefaultAspects()
	orb := 3.0

	result := ComputePrimaryDirections(natal, ascLon, mcLon, lat, age, aspects, orb)

	expected := 32.18
	if math.Abs(result.DirectedASC-expected) > 0.05 {
		t.Errorf("Directed ASC: want ~%.2f, got %.2f", expected, result.DirectedASC)
	}

	// Should find Sun conjunction
	if len(result.ASCAspects) < 1 {
		t.Fatal("Expected at least 1 ASC aspect (Sun conjunction)")
	}
	found := false
	for _, h := range result.ASCAspects {
		if h.NatalPlanet == "Sun" && h.Aspect == "conjunction" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected Sun conjunction with directed ASC, got: %+v", result.ASCAspects)
	}
}

func TestPrimaryDirections_Equator_MC(t *testing.T) {
	// MC direction uses RA, which is latitude-independent.
	// MC at 90° (Cancer 0°), age 30.
	// RA(90°) = 90°
	// RA_directed = 120°
	// tan(λ) = tan(120°) / cos(ε) = -1.73205 / 0.91741 = -1.88798
	// λ = atan(-1.88798) = -62.09° → QII: 180-62.09 = 117.91° = 27°55' Cancer

	natal := map[string]float64{
		"Jupiter": 118.0, // near directed MC → conjunction
	}
	ascLon := 0.0
	mcLon := 90.0
	lat := 0.0
	age := 30.0

	aspects := DefaultAspects()
	orb := 3.0

	result := ComputePrimaryDirections(natal, ascLon, mcLon, lat, age, aspects, orb)

	expected := 117.91
	if math.Abs(result.DirectedMC-expected) > 0.05 {
		t.Errorf("Directed MC: want ~%.2f, got %.2f", expected, result.DirectedMC)
	}

	// Should find Jupiter conjunction
	if len(result.MCAspects) < 1 {
		t.Fatal("Expected at least 1 MC aspect (Jupiter conjunction)")
	}
	found := false
	for _, h := range result.MCAspects {
		if h.NatalPlanet == "Jupiter" && h.Aspect == "conjunction" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected Jupiter conjunction with directed MC, got: %+v", result.MCAspects)
	}
}

func TestPrimaryDirections_NonEquator_ASC(t *testing.T) {
	// At latitude 51°N (London), OA ≠ RA.
	// ASC at 0° Aries: RA=0°, Dec=0°, OA=0° (since tan(0)=0)
	// Age 30: OA_directed = 30°
	// At higher latitudes, the same OA corresponds to a larger ecliptic arc.

	natal := map[string]float64{"Sun": 0.0}
	ascLon := 0.0
	mcLon := 90.0
	lat := 51.0
	age := 30.0

	aspects := DefaultAspects()
	orb := 3.0

	result := ComputePrimaryDirections(natal, ascLon, mcLon, lat, age, aspects, orb)

	// At lat 51°, directed ASC should be > equator value (32.18°)
	if result.DirectedASC <= 32.18 {
		t.Errorf("At lat 51°, directed ASC (%.2f) should be > equator value (32.18)", result.DirectedASC)
	}

	// Verify round-trip: compute OA of directed ASC, should equal OA_ASC + age
	oaASC := raToOA(lonToRA(ascLon, obliquityDeg), lonToDec(ascLon, obliquityDeg), lat)
	oaDirected := raToOA(lonToRA(result.DirectedASC, obliquityDeg), lonToDec(result.DirectedASC, obliquityDeg), lat)
	expectedOA := oaASC + age
	if math.Abs(oaDirected-expectedOA) > 0.02 {
		t.Errorf("OA round-trip: OA(ASC)=%.4f, OA(directed)=%.4f, want %.4f (diff=%.4f)",
			oaASC, oaDirected, expectedOA, math.Abs(oaDirected-expectedOA))
	}
}

func TestPrimaryDirections_MC_RoundTrip(t *testing.T) {
	// MC direction uses RA directly. Verify round-trip for any MC position.
	natal := map[string]float64{"Moon": 200.0}
	ascLon := 237.5 // AJ's approximate ASC
	mcLon := 185.0  // some MC
	lat := 47.0     // AJ's latitude
	age := 57.33    // AJ's current age

	aspects := DefaultAspects()
	orb := 3.0

	result := ComputePrimaryDirections(natal, ascLon, mcLon, lat, age, aspects, orb)

	// Round-trip: RA(MC) + age → λ → RA(λ) should equal RA(MC) + age
	raMC := lonToRA(mcLon, obliquityDeg)
	raDirected := lonToRA(result.DirectedMC, obliquityDeg)
	expectedRA := normalizeLon(raMC + age)
	if math.Abs(angleDist(raDirected, expectedRA)) > 0.01 {
		t.Errorf("MC RA round-trip: RA(MC)=%.4f, RA(directed)=%.4f, want %.4f",
			raMC, raDirected, expectedRA)
	}
}

func TestPrimaryDirections_NoAspects(t *testing.T) {
	// No natal planets near directed points
	natal := map[string]float64{
		"Sun": 150.0,
	}
	ascLon := 0.0
	mcLon := 90.0
	lat := 0.0
	age := 30.0

	aspects := DefaultAspects()
	orb := 1.0 // tight orb

	result := ComputePrimaryDirections(natal, ascLon, mcLon, lat, age, aspects, orb)

	if len(result.ASCAspects) != 0 {
		t.Errorf("Expected 0 ASC aspects, got %d", len(result.ASCAspects))
	}
	if len(result.MCAspects) != 0 {
		t.Errorf("Expected 0 MC aspects, got %d", len(result.MCAspects))
	}
}

func TestPrimaryDirections_Opposition(t *testing.T) {
	// Natal planet at opposition to directed ASC
	// ASC at 0°, age 30, equator → directed ASC ~32.18°
	// Opposition = 32.18 + 180 = 212.18°

	natal := map[string]float64{
		"Mars": 212.0,
	}
	ascLon := 0.0
	mcLon := 90.0
	lat := 0.0
	age := 30.0

	aspects := DefaultAspects()
	orb := 3.0

	result := ComputePrimaryDirections(natal, ascLon, mcLon, lat, age, aspects, orb)

	found := false
	for _, h := range result.ASCAspects {
		if h.NatalPlanet == "Mars" && h.Aspect == "opposition" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected Mars opposition with directed ASC, got: %+v", result.ASCAspects)
	}
}
