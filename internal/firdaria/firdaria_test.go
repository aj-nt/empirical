package firdaria

import (
	"math"
	"testing"
)

// ── ComputeFirdaria ──────────────────────────────────────────────────────

func TestComputeFirdariaDiurnal(t *testing.T) {
	report := ComputeFirdaria("test", true, 2000, 1, 1)
	if !report.Diurnal {
		t.Error("diurnal chart should have Diurnal=true")
	}
	if len(report.Order) != 9 {
		t.Errorf("expected 9 planets in order, got %d", len(report.Order))
	}
	if report.Order[0] != "Sun" {
		t.Errorf("diurnal first planet = %s, want Sun", report.Order[0])
	}
	if len(report.MajorPeriods) != 9 {
		t.Errorf("expected 9 major periods, got %d", len(report.MajorPeriods))
	}
	if report.MajorPeriods[0].Planet != "Sun" {
		t.Errorf("first major period = %s, want Sun", report.MajorPeriods[0].Planet)
	}
	if math.Abs(report.MajorPeriods[0].Years-10.0) > 0.01 {
		t.Errorf("Sun major years = %.2f, want 10", report.MajorPeriods[0].Years)
	}
}

func TestComputeFirdariaNocturnal(t *testing.T) {
	report := ComputeFirdaria("test", false, 2000, 1, 1)
	if report.Diurnal {
		t.Error("nocturnal chart should have Diurnal=false")
	}
	if report.Order[0] != "Moon" {
		t.Errorf("nocturnal first planet = %s, want Moon", report.Order[0])
	}
	if report.MajorPeriods[0].Planet != "Moon" {
		t.Errorf("first major period = %s, want Moon", report.MajorPeriods[0].Planet)
	}
	if math.Abs(report.MajorPeriods[0].Years-9.0) > 0.01 {
		t.Errorf("Moon major years = %.2f, want 9", report.MajorPeriods[0].Years)
	}
}

func TestFirdariaTotalYears(t *testing.T) {
	report := ComputeFirdaria("test", true, 2000, 1, 1)
	total := 0.0
	for _, p := range report.MajorPeriods {
		total += p.Years
	}
	if math.Abs(total-75.0) > 0.5 {
		t.Errorf("total major years = %.2f, want 75", total)
	}
}

func TestFirdariaSubPeriods(t *testing.T) {
	report := ComputeFirdaria("test", true, 2000, 1, 1)
	// 9 major × 9 sub = 81 sub-periods
	if len(report.SubPeriods) != 81 {
		t.Errorf("expected 81 sub-periods, got %d", len(report.SubPeriods))
	}
	// First sub-period should match first major period planet
	if report.SubPeriods[0].Planet != report.MajorPeriods[0].Planet {
		t.Errorf("first sub-period planet = %s, want %s", report.SubPeriods[0].Planet, report.MajorPeriods[0].Planet)
	}
	// All sub-periods should have Level="sub"
	for _, sp := range report.SubPeriods {
		if sp.Level != "sub" {
			t.Errorf("sub-period %s has level=%s, want sub", sp.Planet, sp.Level)
		}
	}
}

func TestFirdariaDateSequence(t *testing.T) {
	report := ComputeFirdaria("test", true, 2000, 1, 1)
	// Major periods should be sequential (end of one = start of next)
	for i := 1; i < len(report.MajorPeriods); i++ {
		if report.MajorPeriods[i].Start != report.MajorPeriods[i-1].End {
			t.Errorf("major period %d start=%s != previous end=%s",
				i, report.MajorPeriods[i].Start, report.MajorPeriods[i-1].End)
		}
	}
	// Sub-periods should be sequential within each major
	// (end of one = start of next)
	for i := 1; i < len(report.SubPeriods); i++ {
		if report.SubPeriods[i].Start != report.SubPeriods[i-1].End {
			// Only check within same major period (sub-periods don't cross major boundaries)
			// Actually they should — let's just check they're all sequential
			if report.SubPeriods[i].Start != report.SubPeriods[i-1].End {
				t.Errorf("sub-period %d start=%s != previous end=%s",
					i, report.SubPeriods[i].Start, report.SubPeriods[i-1].End)
			}
		}
	}
}

func TestFirdariaName(t *testing.T) {
	report := ComputeFirdaria("AJ", false, 1969, 2, 15)
	if report.Name != "AJ" {
		t.Errorf("name = %s, want AJ", report.Name)
	}
}

// ── Julian Day helpers ───────────────────────────────────────────────────

func TestFirdariaJulianDayRoundtrip(t *testing.T) {
	tests := []struct{ y, m, d int; want string }{
		{1969, 2, 15, "1969-02-15"},
		{2000, 1, 1, "2000-01-01"},
		{2026, 6, 15, "2026-06-15"},
	}
	for _, tc := range tests {
		jd := julianDay(tc.y, tc.m, tc.d)
		dateStr := jdToDate(jd)
		if dateStr != tc.want {
			t.Errorf("jd roundtrip (%d-%02d-%02d): got %s, want %s", tc.y, tc.m, tc.d, dateStr, tc.want)
		}
	}
}

func TestFirdariaDateToJD(t *testing.T) {
	jd := dateToJD("2000-01-01")
	expected := julianDay(2000, 1, 1)
	if math.Abs(jd-expected) > 0.01 {
		t.Errorf("dateToJD(2000-01-01) = %.4f, want %.4f", jd, expected)
	}
}
