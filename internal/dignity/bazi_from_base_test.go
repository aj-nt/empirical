package dignity

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/aj-nt/empirical/internal/swe"
)

func initEpheForBFB(t *testing.T) {
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

func TestBaZiFromBase_KnownChart(t *testing.T) {
	initEpheForBFB(t)

	bc, err := ComputeBaseChart(BirthData{Name: "AJ", Year: 1969, Month: 2, Day: 15, Hour: 23, Minute: 10, Second: 0, TZOffset: -8, Lat: 47.038, Lng: -122.901})
	if err != nil {
		t.Fatalf("ComputeBaseChart failed: %v", err)
	}

	pillars := BaZiFromBase(bc)

	// All four pillars should be non-empty
	if pillars.Year.Stem == "" || pillars.Year.Branch == "" {
		t.Error("Year pillar is empty")
	}
	if pillars.Month.Stem == "" || pillars.Month.Branch == "" {
		t.Error("Month pillar is empty")
	}
	if pillars.Day.Stem == "" || pillars.Day.Branch == "" {
		t.Error("Day pillar is empty")
	}
	if pillars.Hour.Stem == "" || pillars.Hour.Branch == "" {
		t.Error("Hour pillar is empty")
	}

	// Day master should be set
	if pillars.DayMaster.Element == "" {
		t.Error("DayMaster element is empty")
	}
	if pillars.DayMaster.YinYang == "" {
		t.Error("DayMaster YinYang is empty")
	}
}

func TestBaZiFromBase_MatchesDirect(t *testing.T) {
	initEpheForBFB(t)

	bc, err := ComputeBaseChart(BirthData{Name: "Test", Year: 2000, Month: 1, Day: 1, Hour: 12, Minute: 0, Second: 0, TZOffset: 0, Lat: 51.5, Lng: -0.12})
	if err != nil {
		t.Fatalf("ComputeBaseChart failed: %v", err)
	}

	fromBase := BaZiFromBase(bc)
	direct := ComputeBaZiPillars(2000, 1, 1, 12)

	if fromBase.Year.Stem != direct.Year.Stem || fromBase.Year.Branch != direct.Year.Branch {
		t.Errorf("Year mismatch: fromBase=%s/%s direct=%s/%s",
			fromBase.Year.Stem, fromBase.Year.Branch, direct.Year.Stem, direct.Year.Branch)
	}
	if fromBase.Month.Stem != direct.Month.Stem || fromBase.Month.Branch != direct.Month.Branch {
		t.Errorf("Month mismatch: fromBase=%s/%s direct=%s/%s",
			fromBase.Month.Stem, fromBase.Month.Branch, direct.Month.Stem, direct.Month.Branch)
	}
	if fromBase.Day.Stem != direct.Day.Stem || fromBase.Day.Branch != direct.Day.Branch {
		t.Errorf("Day mismatch: fromBase=%s/%s direct=%s/%s",
			fromBase.Day.Stem, fromBase.Day.Branch, direct.Day.Stem, direct.Day.Branch)
	}
	if fromBase.Hour.Stem != direct.Hour.Stem || fromBase.Hour.Branch != direct.Hour.Branch {
		t.Errorf("Hour mismatch: fromBase=%s/%s direct=%s/%s",
			fromBase.Hour.Stem, fromBase.Hour.Branch, direct.Hour.Stem, direct.Hour.Branch)
	}
	if fromBase.DayMaster.Element != direct.DayMaster.Element {
		t.Errorf("DayMaster element mismatch: fromBase=%s direct=%s",
			fromBase.DayMaster.Element, direct.DayMaster.Element)
	}
}

func TestBaZiFromBase_EmptyName(t *testing.T) {
	initEpheForBFB(t)

	bc, err := ComputeBaseChart(BirthData{Name: "", Year: 2000, Month: 1, Day: 1, Hour: 12, Minute: 0, Second: 0, TZOffset: 0, Lat: 51.5, Lng: -0.12})
	if err != nil {
		t.Fatalf("ComputeBaseChart failed: %v", err)
	}

	pillars := BaZiFromBase(bc)
	if pillars.Year.Stem == "" {
		t.Error("Year pillar empty for empty-name chart")
	}
}
