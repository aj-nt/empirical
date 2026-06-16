package dignity

import (
	"math"
	"strings"
	"testing"
)

// ── ComputeLunarPhase ────────────────────────────────────────────────────

func TestComputeLunarPhase(t *testing.T) {
	tests := []struct {
		sun, moon float64
		wantName  string
		wantIdx   int
	}{
		{0, 0, "New Moon", 0},
		{0, 44, "New Moon", 0},
		{0, 45, "Waxing Crescent", 1},
		{0, 89, "Waxing Crescent", 1},
		{0, 90, "First Quarter", 2},
		{0, 134, "First Quarter", 2},
		{0, 135, "Waxing Gibbous", 3},
		{0, 179, "Waxing Gibbous", 3},
		{0, 180, "Full Moon", 4},
		{0, 224, "Full Moon", 4},
		{0, 225, "Waning Gibbous", 5},
		{0, 269, "Waning Gibbous", 5},
		{0, 270, "Last Quarter", 6},
		{0, 314, "Last Quarter", 6},
		{0, 315, "Waning Crescent", 7},
		{0, 359, "Waning Crescent", 7},
		// Wraparound: Moon behind Sun by 20° → New Moon
		{350, 10, "New Moon", 0},
	}
	for _, tc := range tests {
		lp := ComputeLunarPhase(tc.sun, tc.moon)
		if lp.Name != tc.wantName {
			t.Errorf("ComputeLunarPhase(%.0f, %.0f) name = %s, want %s", tc.sun, tc.moon, lp.Name, tc.wantName)
		}
		if lp.PhaseIndex != tc.wantIdx {
			t.Errorf("ComputeLunarPhase(%.0f, %.0f) idx = %d, want %d", tc.sun, tc.moon, lp.PhaseIndex, tc.wantIdx)
		}
	}
}

// ── DetectRetrogrades ────────────────────────────────────────────────────

func TestDetectRetrogrades(t *testing.T) {
	speeds := map[string]float64{
		"Sun":     0.98,
		"Moon":    13.2,
		"Mercury": -0.5,
		"Venus":   1.2,
		"Mars":    -0.3,
		"Jupiter": 0.1,
		"Saturn":  -0.05,
	}
	rx := DetectRetrogrades(speeds)
	if len(rx) != 7 {
		t.Fatalf("expected 7 planets, got %d", len(rx))
	}
	// Mercury, Mars, Saturn should be retrograde
	rxMap := make(map[string]bool)
	for _, r := range rx {
		rxMap[r.Planet] = r.Retrograde
	}
	if !rxMap["Mercury"] {
		t.Error("Mercury should be retrograde")
	}
	if !rxMap["Mars"] {
		t.Error("Mars should be retrograde")
	}
	if !rxMap["Saturn"] {
		t.Error("Saturn should be retrograde")
	}
	if rxMap["Sun"] {
		t.Error("Sun should NOT be retrograde")
	}
	if rxMap["Venus"] {
		t.Error("Venus should NOT be retrograde")
	}
}

func TestDetectRetrogradesEmpty(t *testing.T) {
	rx := DetectRetrogrades(map[string]float64{})
	if len(rx) != 0 {
		t.Errorf("empty speeds should produce 0 results, got %d", len(rx))
	}
}

// ── ComputeAntiscia ──────────────────────────────────────────────────────

func TestComputeAntiscia(t *testing.T) {
	positions := map[string]float64{
		"Sun": 0,
		"Moon": 90,
	}
	anti := ComputeAntiscia(positions)
	if len(anti) != 2 {
		t.Fatalf("expected 2 antiscia, got %d", len(anti))
	}
	// Sun at 0° → antiscion at 0° (360-0=360→0), contra at 180°
	if math.Abs(anti[0].Antiscion) > 0.01 {
		t.Errorf("Sun antiscion = %.2f, want 0", anti[0].Antiscion)
	}
	if math.Abs(anti[0].ContraAntiscion-180) > 0.01 {
		t.Errorf("Sun contra = %.2f, want 180", anti[0].ContraAntiscion)
	}
	// Moon at 90° → antiscion at 270°, contra at 90°
	if math.Abs(anti[1].Antiscion-270) > 0.01 {
		t.Errorf("Moon antiscion = %.2f, want 270", anti[1].Antiscion)
	}
	if math.Abs(anti[1].ContraAntiscion-90) > 0.01 {
		t.Errorf("Moon contra = %.2f, want 90", anti[1].ContraAntiscion)
	}
}

// ── ComputeDecans ────────────────────────────────────────────────────────

func TestComputeDecans(t *testing.T) {
	positions := map[string]float64{
		"Sun":  5,   // Aries, decan 1 → Mars
		"Moon": 15,  // Aries, decan 2 → Sun
		"Mars": 25,  // Aries, decan 3 → Venus
	}
	decans := ComputeDecans(positions)
	if len(decans) != 3 {
		t.Fatalf("expected 3 decans, got %d", len(decans))
	}
	if decans[0].Decan != 1 || decans[0].Ruler != "Mars" {
		t.Errorf("Sun decan: %d/%s, want 1/Mars", decans[0].Decan, decans[0].Ruler)
	}
	if decans[1].Decan != 2 || decans[1].Ruler != "Sun" {
		t.Errorf("Moon decan: %d/%s, want 2/Sun", decans[1].Decan, decans[1].Ruler)
	}
	if decans[2].Decan != 3 || decans[2].Ruler != "Venus" {
		t.Errorf("Mars decan: %d/%s, want 3/Venus", decans[2].Decan, decans[2].Ruler)
	}
}

func TestComputeDecansTaurus(t *testing.T) {
	// Taurus starts at 30°. Chaldean order cycles: Mars, Sun, Venus, Mercury, Moon, Saturn, Jupiter
	// Aries decans: Mars(1), Sun(2), Venus(3)
	// Taurus decans: Mercury(1), Moon(2), Saturn(3)
	positions := map[string]float64{"Sun": 35} // Taurus, decan 1
	decans := ComputeDecans(positions)
	if len(decans) != 1 {
		t.Fatalf("expected 1 decan, got %d", len(decans))
	}
	if decans[0].Sign != "Taurus" {
		t.Errorf("sign = %s, want Taurus", decans[0].Sign)
	}
	if decans[0].Ruler != "Mercury" {
		t.Errorf("Taurus decan 1 ruler = %s, want Mercury", decans[0].Ruler)
	}
}

// ── ComputeTerms ─────────────────────────────────────────────────────────

func TestComputeTerms(t *testing.T) {
	// Aries terms: Jupiter 0-6, Venus 6-12, Mercury 12-20, Mars 20-25, Saturn 25-30
	positions := map[string]float64{
		"Sun":  3,   // Aries, term 1 → Jupiter
		"Moon": 8,   // Aries, term 2 → Venus
		"Mars": 22,  // Aries, term 4 → Mars
	}
	terms := ComputeTerms(positions)
	if len(terms) != 3 {
		t.Fatalf("expected 3 terms, got %d", len(terms))
	}
	if terms[0].Term != 1 || terms[0].Ruler != "Jupiter" {
		t.Errorf("Sun term: %d/%s, want 1/Jupiter", terms[0].Term, terms[0].Ruler)
	}
	if terms[1].Term != 2 || terms[1].Ruler != "Venus" {
		t.Errorf("Moon term: %d/%s, want 2/Venus", terms[1].Term, terms[1].Ruler)
	}
	if terms[2].Term != 4 || terms[2].Ruler != "Mars" {
		t.Errorf("Mars term: %d/%s, want 4/Mars", terms[2].Term, terms[2].Ruler)
	}
}

func TestComputeTermsEdge(t *testing.T) {
	// At exact boundary (6°), should be in the next term
	positions := map[string]float64{"Sun": 6} // Aries, exactly at Venus boundary
	terms := ComputeTerms(positions)
	if len(terms) != 1 {
		t.Fatalf("expected 1 term, got %d", len(terms))
	}
	if terms[0].Ruler != "Venus" {
		t.Errorf("at 6° Aries, ruler = %s, want Venus", terms[0].Ruler)
	}
}

// ── ComputeDispositorTree ────────────────────────────────────────────────

func TestComputeDispositorTree(t *testing.T) {
	// Mars in Scorpio → Mars (final dispositor)
	// Venus in Aries → Mars
	// Sun in Aquarius → Saturn
	// Saturn in Aries → Mars
	positions := map[string]float64{
		"Sun":    327, // Aquarius
		"Moon":   322, // Aquarius
		"Venus":  12,  // Aries
		"Mars":   235, // Scorpio
		"Jupiter": 184, // Libra
		"Saturn": 21,  // Aries
		"Mercury": 302, // Aquarius
	}
	tree := ComputeDispositorTree(positions)
	if len(tree.Nodes) != 7 {
		t.Fatalf("expected 7 nodes, got %d", len(tree.Nodes))
	}
	if len(tree.FinalDispositors) != 1 || tree.FinalDispositors[0] != "Mars" {
		t.Errorf("final dispositors = %v, want [Mars]", tree.FinalDispositors)
	}
	if len(tree.MutualReceptions) != 0 {
		t.Errorf("expected 0 mutual receptions, got %v", tree.MutualReceptions)
	}
}

func TestComputeDispositorTreeMutualReception(t *testing.T) {
	// Venus in Aries (ruled by Mars), Mars in Taurus (ruled by Venus) → mutual reception
	positions := map[string]float64{
		"Sun":     0,
		"Moon":    0,
		"Venus":   5,   // Aries → Mars
		"Mars":    35,  // Taurus → Venus
		"Jupiter": 0,
		"Saturn":  0,
		"Mercury": 0,
	}
	tree := ComputeDispositorTree(positions)
	if len(tree.MutualReceptions) == 0 {
		t.Error("expected mutual reception between Venus and Mars")
	}
}

// ── ComputeVOCMoon ───────────────────────────────────────────────────────

func TestComputeVOCMoon(t *testing.T) {
	// Moon at 5° Aries, no planets ahead within 3° orb of an aspect
	// before the sign boundary at 30° Aries
	positions := map[string]float64{
		"Moon":    5,
		"Sun":     200, // far away
		"Mercury": 200,
		"Venus":   200,
		"Mars":    200,
		"Jupiter": 200,
		"Saturn":  200,
	}
	voc := ComputeVOCMoon(positions)
	if !voc.VOC {
		t.Error("Moon should be void of course")
	}
	if voc.MoonSign != "Aries" {
		t.Errorf("Moon sign = %s, want Aries", voc.MoonSign)
	}
	if math.Abs(voc.DegreesToNext-25) > 0.1 {
		t.Errorf("degrees to next = %.2f, want ~25", voc.DegreesToNext)
	}
}

func TestComputeVOCMoonNotVOC(t *testing.T) {
	// Moon at 5° Aries, Venus at 7° Aries → applying conjunction before sign boundary
	positions := map[string]float64{
		"Moon":    5,
		"Sun":     200,
		"Mercury": 200,
		"Venus":   7,
		"Mars":    200,
		"Jupiter": 200,
		"Saturn":  200,
	}
	voc := ComputeVOCMoon(positions)
	if voc.VOC {
		t.Error("Moon should NOT be void of course (Venus applying conjunction)")
	}
}

func TestComputeVOCMoonNoMoon(t *testing.T) {
	voc := ComputeVOCMoon(map[string]float64{})
	if voc.MoonSign != "" {
		t.Error("no Moon should produce empty VOCMoon")
	}
}

// ── ComputeTraditionalReport ─────────────────────────────────────────────

func TestComputeTraditionalReport(t *testing.T) {
	positions := map[string]float64{
		"Sun": 0, "Moon": 90, "Mercury": 60, "Venus": 120,
		"Mars": 240, "Jupiter": 180, "Saturn": 300,
	}
	speeds := map[string]float64{
		"Sun": 0.98, "Moon": 13.2, "Mercury": 0.5,
		"Venus": 1.2, "Mars": 0.3, "Jupiter": 0.1, "Saturn": -0.05,
	}
	report := ComputeTraditionalReport("test", positions, speeds)
	if report.Name != "test" {
		t.Errorf("name = %s, want test", report.Name)
	}
	if report.LunarPhase.Name == "" {
		t.Error("lunar phase should be computed")
	}
	if len(report.Retrogrades) == 0 {
		t.Error("retrogrades should be computed")
	}
	if len(report.Antiscia) == 0 {
		t.Error("antiscia should be computed")
	}
	if len(report.Decans) == 0 {
		t.Error("decans should be computed")
	}
	if len(report.Terms) == 0 {
		t.Error("terms should be computed")
	}
	if len(report.DispositorTree.Nodes) == 0 {
		t.Error("dispositor tree should be computed")
	}
}

// ── FormatTraditionalReport ──────────────────────────────────────────────

func TestFormatTraditionalReport(t *testing.T) {
	positions := map[string]float64{
		"Sun": 0, "Moon": 90, "Mercury": 60, "Venus": 120,
		"Mars": 240, "Jupiter": 180, "Saturn": 300,
	}
	speeds := map[string]float64{
		"Sun": 0.98, "Moon": 13.2, "Mercury": 0.5,
		"Venus": 1.2, "Mars": 0.3, "Jupiter": 0.1, "Saturn": -0.05,
	}
	report := ComputeTraditionalReport("test", positions, speeds)
	formatted := FormatTraditionalReport(report)
	if !strings.Contains(formatted, "test") {
		t.Error("formatted report should contain name")
	}
	if !strings.Contains(formatted, "Lunar Phase") {
		t.Error("formatted report should contain Lunar Phase section")
	}
	if !strings.Contains(formatted, "Retrogrades") {
		t.Error("formatted report should contain Retrogrades section")
	}
	if !strings.Contains(formatted, "Dispositor Tree") {
		t.Error("formatted report should contain Dispositor Tree section")
	}
	if !strings.Contains(formatted, "Decans") {
		t.Error("formatted report should contain Decans section")
	}
	if !strings.Contains(formatted, "Egyptian Terms") {
		t.Error("formatted report should contain Terms section")
	}
	if !strings.Contains(formatted, "Antiscia") {
		t.Error("formatted report should contain Antiscia section")
	}
	if !strings.Contains(formatted, "Void of Course Moon") {
		t.Error("formatted report should contain VOC Moon section")
	}
}

func TestFormatTraditionalReportNoRx(t *testing.T) {
	positions := map[string]float64{
		"Sun": 0, "Moon": 90, "Mercury": 60, "Venus": 120,
		"Mars": 240, "Jupiter": 180, "Saturn": 300,
	}
	speeds := map[string]float64{
		"Sun": 0.98, "Moon": 13.2, "Mercury": 0.5,
		"Venus": 1.2, "Mars": 0.3, "Jupiter": 0.1, "Saturn": 0.05,
	}
	report := ComputeTraditionalReport("test", positions, speeds)
	formatted := FormatTraditionalReport(report)
	if !strings.Contains(formatted, "No planets retrograde") {
		t.Error("should say 'No planets retrograde' when none are Rx")
	}
}

func TestFormatTraditionalReportVOC(t *testing.T) {
	positions := map[string]float64{
		"Moon": 5, "Sun": 200, "Mercury": 200, "Venus": 200,
		"Mars": 200, "Jupiter": 200, "Saturn": 200,
	}
	speeds := map[string]float64{"Moon": 13.2}
	report := ComputeTraditionalReport("test", positions, speeds)
	formatted := FormatTraditionalReport(report)
	if !strings.Contains(formatted, "VOID OF COURSE") {
		t.Error("should say VOID OF COURSE when Moon is VOC")
	}
}
