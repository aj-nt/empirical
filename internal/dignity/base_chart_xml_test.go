package dignity

import (
	"encoding/xml"
	"strings"
	"testing"
)

// TestBaseChart_ToXML_Roundtrip verifies that a BaseChart can be serialized
// to XML and deserialized back, producing an equivalent chart.
func TestBaseChart_ToXML_Roundtrip(t *testing.T) {
	t.Parallel()

	// Build a minimal but complete BaseChart without SWE
	bc := &BaseChart{
		Name:     "AJ",
		Year:     1969,
		Month:    2,
		Day:      15,
		Hour:     23,
		Minute:   10,
		Second:   0,
		TZOffset: -8.0,
		Lat:      47.038,
		Lng:      -122.901,
		Tropical: map[string]Position{
			"Sun":   {Lon: 326.75, Lat: 0.0, Speed: 1.01},
			"Moon":  {Lon: 298.12, Lat: -4.2, Speed: 13.5},
			"Mars":  {Lon: 240.5, Lat: 1.2, Speed: 0.75},
		},
		Sidereal: map[string]Position{
			"Sun":   {Lon: 303.30, Lat: 0.0, Speed: 1.01},
			"Moon":  {Lon: 274.67, Lat: -4.2, Speed: 13.5},
			"Mars":  {Lon: 217.05, Lat: 1.2, Speed: 0.75},
		},
		Ayanamsa:  23.45,
		ASC:       187.5,
		MC:        97.3,
		DSC:       7.5,
		IC:        277.3,
		NorthNode: 2.16,
		SouthNode: 182.16,
		Houses: map[string][]float64{
			"placidus":    {0, 187.5, 217.3, 247.1, 277.3, 307.5, 337.3, 7.5, 37.3, 67.1, 97.3, 127.5, 157.3},
			"whole_sign":  {0, 180.0, 210.0, 240.0, 270.0, 300.0, 330.0, 0.0, 30.0, 60.0, 90.0, 120.0, 150.0},
		},
		JD: 2440268.798611,
		Aspects: []AspectHit{
			{Planet1: "Sun", Planet2: "Mars", Aspect: "square", Orb: 0.52},
			{Planet1: "Moon", Planet2: "Jupiter", Aspect: "square", Orb: 0.08},
		},
		FixedStars: []StarConjunction{
			{Star: "Regulus", Planet: "Mars", Orb: 1.5},
		},
		ArabicParts: map[string]float64{
			"Fortune": 45.2,
			"Spirit":  312.8,
		},
		GMST: 123.456,
	}

	// Serialize to XML
	xmlBytes, err := bc.ToXML()
	if err != nil {
		t.Fatalf("ToXML failed: %v", err)
	}

	// Basic XML structure checks
	xmlStr := string(xmlBytes)
	if !strings.Contains(xmlStr, "<?xml") {
		t.Error("XML output missing XML declaration")
	}
	if !strings.Contains(xmlStr, "BaseChart") {
		t.Error("XML output missing BaseChart root element")
	}
	if !strings.Contains(xmlStr, "AJ") {
		t.Error("XML output missing chart name")
	}

	// Deserialize back
	bc2, err := FromXML(xmlBytes)
	if err != nil {
		t.Fatalf("FromXML failed: %v", err)
	}

	// Verify roundtrip fidelity
	if bc2.Name != bc.Name {
		t.Errorf("Name = %q, want %q", bc2.Name, bc.Name)
	}
	if bc2.Year != bc.Year {
		t.Errorf("Year = %d, want %d", bc2.Year, bc.Year)
	}
	if bc2.Lat != bc.Lat {
		t.Errorf("Lat = %f, want %f", bc2.Lat, bc.Lat)
	}
	if bc2.ASC != bc.ASC {
		t.Errorf("ASC = %f, want %f", bc2.ASC, bc.ASC)
	}
	if bc2.Ayanamsa != bc.Ayanamsa {
		t.Errorf("Ayanamsa = %f, want %f", bc2.Ayanamsa, bc.Ayanamsa)
	}
	if bc2.NorthNode != bc.NorthNode {
		t.Errorf("NorthNode = %f, want %f", bc2.NorthNode, bc.NorthNode)
	}
	if bc2.JD != bc.JD {
		t.Errorf("JD = %f, want %f", bc2.JD, bc.JD)
	}
	if bc2.GMST != bc.GMST {
		t.Errorf("GMST = %f, want %f", bc2.GMST, bc.GMST)
	}

	// Check tropical positions
	if len(bc2.Tropical) != len(bc.Tropical) {
		t.Errorf("Tropical count = %d, want %d", len(bc2.Tropical), len(bc.Tropical))
	}
	for name, pos := range bc.Tropical {
		pos2, ok := bc2.Tropical[name]
		if !ok {
			t.Errorf("missing tropical planet %q", name)
			continue
		}
		if pos2.Lon != pos.Lon || pos2.Lat != pos.Lat || pos2.Speed != pos.Speed {
			t.Errorf("Tropical[%q] = %+v, want %+v", name, pos2, pos)
		}
	}

	// Check sidereal positions
	if len(bc2.Sidereal) != len(bc.Sidereal) {
		t.Errorf("Sidereal count = %d, want %d", len(bc2.Sidereal), len(bc.Sidereal))
	}

	// Check houses
	if len(bc2.Houses) != len(bc.Houses) {
		t.Errorf("Houses count = %d, want %d", len(bc2.Houses), len(bc.Houses))
	}
	for sys, cusps := range bc.Houses {
		cusps2, ok := bc2.Houses[sys]
		if !ok {
			t.Errorf("missing house system %q", sys)
			continue
		}
		if len(cusps2) != len(cusps) {
			t.Errorf("Houses[%q] length = %d, want %d", sys, len(cusps2), len(cusps))
		}
	}

	// Check aspects
	if len(bc2.Aspects) != len(bc.Aspects) {
		t.Errorf("Aspects count = %d, want %d", len(bc2.Aspects), len(bc.Aspects))
	}

	// Check fixed stars
	if len(bc2.FixedStars) != len(bc.FixedStars) {
		t.Errorf("FixedStars count = %d, want %d", len(bc2.FixedStars), len(bc.FixedStars))
	}

	// Check Arabic Parts
	if len(bc2.ArabicParts) != len(bc.ArabicParts) {
		t.Errorf("ArabicParts count = %d, want %d", len(bc2.ArabicParts), len(bc.ArabicParts))
	}
}

// TestBaseChart_ToXML_ValidXML verifies the output is well-formed XML
// that can be parsed by encoding/xml.
func TestBaseChart_ToXML_ValidXML(t *testing.T) {
	t.Parallel()

	bc := &BaseChart{
		Name:     "Test",
		Year:     2000,
		Month:    1,
		Day:      1,
		Hour:     12,
		Minute:   0,
		Second:   0,
		TZOffset: 0,
		Lat:      51.5,
		Lng:      -0.12,
		Tropical: map[string]Position{
			"Sun": {Lon: 280.0, Lat: 0.0, Speed: 1.02},
		},
		Sidereal: map[string]Position{
			"Sun": {Lon: 256.0, Lat: 0.0, Speed: 1.02},
		},
		Ayanamsa:  24.0,
		ASC:       100.0,
		MC:        10.0,
		DSC:       280.0,
		IC:        190.0,
		NorthNode: 120.0,
		SouthNode: 300.0,
		Houses: map[string][]float64{
			"placidus": {0, 100, 130, 160, 190, 220, 250, 280, 310, 340, 10, 40, 70},
		},
		JD:          2451545.0,
		Aspects:     []AspectHit{},
		FixedStars:  []StarConjunction{},
		ArabicParts: map[string]float64{},
		GMST:        0.0,
	}

	xmlBytes, err := bc.ToXML()
	if err != nil {
		t.Fatalf("ToXML failed: %v", err)
	}

	// Verify it's valid XML by parsing it
	var doc struct {
		XMLName xml.Name `xml:"BaseChart"`
	}
	if err := xml.Unmarshal(xmlBytes, &doc); err != nil {
		t.Fatalf("output is not valid XML: %v", err)
	}
}

// TestBaseChart_ToXML_EmptyChart verifies that an empty chart serializes
// without error (nil maps, empty slices).
func TestBaseChart_ToXML_EmptyChart(t *testing.T) {
	t.Parallel()

	bc := &BaseChart{
		Name:        "Empty",
		Tropical:    map[string]Position{},
		Sidereal:    map[string]Position{},
		Houses:      map[string][]float64{},
		Aspects:     []AspectHit{},
		FixedStars:  []StarConjunction{},
		ArabicParts: map[string]float64{},
	}

	xmlBytes, err := bc.ToXML()
	if err != nil {
		t.Fatalf("ToXML on empty chart failed: %v", err)
	}

	// Should still be valid XML
	var doc struct {
		XMLName xml.Name `xml:"BaseChart"`
	}
	if err := xml.Unmarshal(xmlBytes, &doc); err != nil {
		t.Fatalf("empty chart output is not valid XML: %v", err)
	}
}
