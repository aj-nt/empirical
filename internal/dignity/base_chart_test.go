package dignity

import (
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/aj-nt/empirical/internal/swe"
)

// initEphe sets up the ephemeris path for tests that need SWE.
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
	swe.SetEphePath(cacheDir)
	swe.SetSidMode(swe.SIDM_LAHIRI, 0, 0)
}

func TestComputeBaseChart_KnownChart(t *testing.T) {
	initEphe(t)

	// AJ's chart: 1969-02-15 23:10 -8:00 47.038 -122.901
	bc, err := ComputeBaseChart(BirthData{Name: "AJ", Year: 1969, Month: 2, Day: 15, Hour: 23, Minute: 10, Second: 0, TZOffset: -8, Lat: 47.038, Lng: -122.901})
	if err != nil {
		t.Fatalf("ComputeBaseChart failed: %v", err)
	}

	// Basic sanity checks
	if bc.Name != "AJ" {
		t.Errorf("Name = %q, want %q", bc.Name, "AJ")
	}
	if len(bc.Tropical) < 12 {
		t.Errorf("got %d tropical positions, want at least 12", len(bc.Tropical))
	}
	if len(bc.Sidereal) != len(bc.Tropical) {
		t.Errorf("sidereal count %d != tropical count %d", len(bc.Sidereal), len(bc.Tropical))
	}

	// Sun should be in Aquarius (~326° tropical)
	sunTrop := bc.Tropical["Sun"].Lon
	if sunTrop < 300 || sunTrop > 330 {
		t.Errorf("Sun tropical = %.2f, expected ~326° (Aquarius)", sunTrop)
	}

	// Sun sidereal should be ~302° (Capricorn)
	sunSid := bc.Sidereal["Sun"].Lon
	if sunSid < 290 || sunSid > 320 {
		t.Errorf("Sun sidereal = %.2f, expected ~302° (Capricorn)", sunSid)
	}

	// Ayanamsa should be ~23.5° for 1969
	if math.Abs(bc.Ayanamsa-23.5) > 1.0 {
		t.Errorf("Ayanamsa = %.2f, expected ~23.5°", bc.Ayanamsa)
	}

	// ASC should be in Scorpio (~210-240°)
	if bc.ASC < 200 || bc.ASC > 250 {
		t.Errorf("ASC = %.2f, expected ~210-240° (Scorpio)", bc.ASC)
	}

	// Houses should have all 5 systems
	expectedSystems := []string{"placidus", "whole_sign", "equal", "porphyry", "koch"}
	for _, hs := range expectedSystems {
		if _, ok := bc.Houses[hs]; !ok {
			t.Errorf("missing house system: %s", hs)
		}
	}

	// Star positions should be non-empty
	if len(bc.StarPositions) == 0 {
		t.Error("expected non-empty star positions")
	}
}

func TestComputeBaseChart_Angles(t *testing.T) {
	initEphe(t)

	bc, err := ComputeBaseChart(BirthData{Name: "Test", Year: 2000, Month: 1, Day: 1, Hour: 12, Minute: 0, Second: 0, TZOffset: 0, Lat: 51.5, Lng: -0.12})
	if err != nil {
		t.Fatalf("ComputeBaseChart failed: %v", err)
	}

	// DSC should be ASC + 180
	dscExpected := bc.ASC + 180
	if dscExpected >= 360 {
		dscExpected -= 360
	}
	if math.Abs(bc.DSC-dscExpected) > 0.01 {
		t.Errorf("DSC = %.2f, want %.2f (ASC+180)", bc.DSC, dscExpected)
	}

	// IC should be MC + 180
	icExpected := bc.MC + 180
	if icExpected >= 360 {
		icExpected -= 360
	}
	if math.Abs(bc.IC-icExpected) > 0.01 {
		t.Errorf("IC = %.2f, want %.2f (MC+180)", bc.IC, icExpected)
	}
}
