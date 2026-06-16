package dignity

import (
	"math"
	"testing"
)

func TestMidpoint_Direct(t *testing.T) {
	// Two points 40° apart — direct midpoint is shorter
	a := 10.0
	b := 50.0
	mp := midpoint(a, b)
	expected := 30.0
	if math.Abs(mp-expected) > 0.01 {
		t.Errorf("midpoint(10, 50) = %.2f, want 30.00", mp)
	}
}

func TestMidpoint_Wraparound(t *testing.T) {
	// Two points 40° apart across 0° — wraparound is shorter
	a := 350.0
	b := 30.0
	mp := midpoint(a, b)
	// Direct: (350+30)/2 = 190, arc = 320°
	// Wrap: (350+360+30)/2 = 370 → 10, arc = 40°
	expected := 10.0
	if math.Abs(mp-expected) > 0.01 {
		t.Errorf("midpoint(350, 30) = %.2f, want 10.00", mp)
	}
}

func TestMidpoint_Equal(t *testing.T) {
	a := 100.0
	b := 100.0
	mp := midpoint(a, b)
	if math.Abs(mp-100.0) > 0.01 {
		t.Errorf("midpoint(100, 100) = %.2f, want 100.00", mp)
	}
}

func TestMidpoint_Opposition(t *testing.T) {
	// Two points exactly opposite — both arcs are 180°, direct wins
	a := 0.0
	b := 180.0
	mp := midpoint(a, b)
	// Direct: 90, arc=180. Wrap: (0+360+180)/2=270, arc=180. Direct wins.
	expected := 90.0
	if math.Abs(mp-expected) > 0.01 {
		t.Errorf("midpoint(0, 180) = %.2f, want 90.00", mp)
	}
}

func TestComputeComposite(t *testing.T) {
	chart1 := map[string]float64{
		"Sun": 10.0, "Moon": 50.0, "Mars": 350.0,
	}
	chart2 := map[string]float64{
		"Sun": 50.0, "Moon": 10.0, "Mars": 30.0,
	}
	comp := ComputeComposite(chart1, chart2)

	if math.Abs(comp["Sun"]-30.0) > 0.01 {
		t.Errorf("Composite Sun = %.2f, want 30.00", comp["Sun"])
	}
	if math.Abs(comp["Moon"]-30.0) > 0.01 {
		t.Errorf("Composite Moon = %.2f, want 30.00", comp["Moon"])
	}
	// Mars: 350 and 30 → wraparound midpoint = 10
	if math.Abs(comp["Mars"]-10.0) > 0.01 {
		t.Errorf("Composite Mars = %.2f, want 10.00", comp["Mars"])
	}
}

func TestComputeCompositeReport(t *testing.T) {
	chart1 := map[string]float64{
		"Sun": 10.0, "Moon": 50.0, "Mercury": 70.0,
		"Venus": 80.0, "Mars": 90.0, "Jupiter": 100.0,
		"Saturn": 110.0,
	}
	chart2 := map[string]float64{
		"Sun": 20.0, "Moon": 60.0, "Mercury": 80.0,
		"Venus": 90.0, "Mars": 100.0, "Jupiter": 110.0,
		"Saturn": 120.0,
	}
	report := ComputeCompositeReport("A", "B", chart1, chart2, 3.0)

	if report.Name1 != "A" || report.Name2 != "B" {
		t.Error("Names not set correctly")
	}
	if len(report.Planets) != 7 {
		t.Errorf("Expected 7 composite planets, got %d", len(report.Planets))
	}
	// All planets 10° apart → composite is midpoint, aspects should exist
	if len(report.Aspects) == 0 {
		t.Error("Expected some composite aspects")
	}
}

func TestComputeCompositeSynastry(t *testing.T) {
	chart1 := map[string]float64{
		"Sun": 10.0, "Moon": 50.0,
	}
	chart2 := map[string]float64{
		"Sun": 20.0, "Moon": 60.0,
	}
	report := ComputeCompositeSynastry("A", "B", chart1, chart2, 3.0)

	// Composite Sun = 15, chart1 Sun = 10 → 5° apart, no aspect at 3° orb
	// Composite Moon = 55, chart1 Moon = 50 → 5° apart, no aspect
	if len(report.ToPerson1) != 0 {
		t.Errorf("Expected 0 composite-to-P1 aspects at 3° orb, got %d", len(report.ToPerson1))
	}
}
