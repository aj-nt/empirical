package dignity

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aj-nt/empirical/internal/swe"
)

func initEpheForVFB(t *testing.T) {
	t.Helper()
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("cannot find home dir: %v", err)
	}
	cacheDir := filepath.Join(home, ".cache", "empirical", "ephe")
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		t.Skipf("ephemeris cache not found at %s", cacheDir)
	}
	swe.SetEphePath(cacheDir)
	swe.SetSidMode(swe.SIDM_LAHIRI, 0, 0)
}

func TestVedicFromBase_KnownChart(t *testing.T) {
	initEpheForVFB(t)

	bc, err := ComputeBaseChart(BirthData{Name: "AJ", Year: 1969, Month: 2, Day: 15, Hour: 23, Minute: 10, Second: 0, TZOffset: -8, Lat: 47.038, Lng: -122.901})
	if err != nil {
		t.Fatalf("ComputeBaseChart failed: %v", err)
	}

	report := VedicFromBase(bc)

	if report == nil {
		t.Fatal("VedicFromBase returned nil")
	}
	if report.Name != "AJ" {
		t.Errorf("Name = %q, want %q", report.Name, "AJ")
	}
	if len(report.Planets) != 7 {
		t.Errorf("Planets = %d, want 7 classical planets", len(report.Planets))
	}
	if report.AyanamsaDegrees <= 0 {
		t.Errorf("AyanamsaDegrees = %f, want positive", report.AyanamsaDegrees)
	}
}

func TestVedicFromBase_EmptyName(t *testing.T) {
	initEpheForVFB(t)

	bc, err := ComputeBaseChart(BirthData{Name: "", Year: 2000, Month: 1, Day: 1, Hour: 12, Minute: 0, Second: 0, TZOffset: 0, Lat: 51.5, Lng: -0.12})
	if err != nil {
		t.Fatalf("ComputeBaseChart failed: %v", err)
	}

	report := VedicFromBase(bc)
	if report == nil {
		t.Fatal("VedicFromBase returned nil for empty name")
	}
	if len(report.Planets) != 7 {
		t.Errorf("Planets = %d, want 7", len(report.Planets))
	}
}

func TestVedicFromBase_Convergence(t *testing.T) {
	initEpheForVFB(t)

	bc, err := ComputeBaseChart(BirthData{Name: "Test", Year: 2000, Month: 1, Day: 1, Hour: 12, Minute: 0, Second: 0, TZOffset: 0, Lat: 51.5, Lng: -0.12})
	if err != nil {
		t.Fatalf("ComputeBaseChart failed: %v", err)
	}

	report := VedicFromBase(bc)
	if report == nil {
		t.Fatal("VedicFromBase returned nil")
	}

	// Every planet should have a convergence classification
	for _, p := range report.Planets {
		if p.Convergence == "" {
			t.Errorf("planet %s has empty convergence", p.Planet)
		}
		if p.Western == "" {
			t.Errorf("planet %s has empty western dignity", p.Planet)
		}
		if p.Vedic == "" {
			t.Errorf("planet %s has empty vedic dignity", p.Planet)
		}
	}

	// Signal + noise should equal total
	if report.SignalCount()+report.NoiseCount() != len(report.Planets) {
		t.Errorf("signal (%d) + noise (%d) != total (%d)",
			report.SignalCount(), report.NoiseCount(), len(report.Planets))
	}
}
