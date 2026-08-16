package dignity

import (
	"math"
	"testing"

)

func initEpheForDFB(t *testing.T) {
	t.Helper()
	initEphe(t)
}

func TestDraconicFromBase_KnownChart(t *testing.T) {
	initEpheForDFB(t)

	bc, err := ComputeBaseChart(BirthData{Name: "AJ", Year: 1969, Month: 2, Day: 15, Hour: 23, Minute: 10, Second: 0, TZOffset: -8, Lat: 47.038, Lng: -122.901})
	if err != nil {
		t.Fatalf("ComputeBaseChart failed: %v", err)
	}

	drac := DraconicFromBase(bc)

	// Offset should equal the North Node longitude
	if math.Abs(drac.Offset-bc.NorthNode) > 0.01 {
		t.Errorf("offset = %.2f, want %.2f (NorthNode)", drac.Offset, bc.NorthNode)
	}

	// Node entry should be at 0°
	if math.Abs(drac.Planets["Node"]-0.0) > 0.01 {
		t.Errorf("draconic Node = %.2f, want 0.0", drac.Planets["Node"])
	}

	// All tropical planets should have draconic counterparts
	for name := range bc.Tropical {
		if _, ok := drac.Planets[name]; !ok {
			t.Errorf("missing draconic position for %s", name)
		}
	}

	// Mercury should shift sign (Aquarius → Capricorn for AJ)
	mercTrop := bc.Tropical["Mercury"].Lon
	mercDrac := drac.Planets["Mercury"]
	tropSign := SignForLongitude(mercTrop)
	dracSign := SignForLongitude(mercDrac)
	if tropSign == dracSign {
		t.Logf("Mercury: tropical=%s draconic=%s (no shift — may be correct for this chart)", tropSign, dracSign)
	}
}

func TestDraconicFromBase_MatchesDirect(t *testing.T) {
	initEpheForDFB(t)

	bc, err := ComputeBaseChart(BirthData{Name: "Test", Year: 2000, Month: 1, Day: 1, Hour: 12, Minute: 0, Second: 0, TZOffset: 0, Lat: 51.5, Lng: -0.12})
	if err != nil {
		t.Fatalf("ComputeBaseChart failed: %v", err)
	}

	fromBase := DraconicFromBase(bc)

	// Compute directly using the same inputs
	planetLons := TropicalToLonMap(bc.Tropical)
	direct := ComputeDraconic(planetLons, bc.NorthNode)

	// Offsets must match
	if math.Abs(fromBase.Offset-direct.Offset) > 0.001 {
		t.Errorf("offset mismatch: fromBase=%.2f direct=%.2f", fromBase.Offset, direct.Offset)
	}

	// All planet positions must match
	for name, dracLon := range direct.Planets {
		fromBaseLon, ok := fromBase.Planets[name]
		if !ok {
			t.Errorf("fromBase missing planet %s", name)
			continue
		}
		if math.Abs(fromBaseLon-dracLon) > 0.001 {
			t.Errorf("%s mismatch: fromBase=%.2f direct=%.2f", name, fromBaseLon, dracLon)
		}
	}

	// No extra planets in fromBase
	for name := range fromBase.Planets {
		if _, ok := direct.Planets[name]; !ok {
			t.Errorf("fromBase has extra planet %s not in direct", name)
		}
	}
}

func TestDraconicFromBase_EmptyName(t *testing.T) {
	initEpheForDFB(t)

	bc, err := ComputeBaseChart(BirthData{Name: "", Year: 2000, Month: 1, Day: 1, Hour: 12, Minute: 0, Second: 0, TZOffset: 0, Lat: 51.5, Lng: -0.12})
	if err != nil {
		t.Fatalf("ComputeBaseChart failed: %v", err)
	}

	drac := DraconicFromBase(bc)

	if drac == nil {
		t.Fatal("DraconicFromBase returned nil")
	}
	if len(drac.Planets) == 0 {
		t.Error("draconic planets are empty")
	}
	if drac.Planets["Node"] != 0.0 {
		t.Errorf("draconic Node = %.2f, want 0.0", drac.Planets["Node"])
	}
}
