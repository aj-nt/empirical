package dignity

import (
	"encoding/xml"
	"strings"
	"testing"
)

// TestTransitChart_ToXML_Roundtrip verifies that a TransitChart can be
// serialized to XML and deserialized back, producing an equivalent chart.
func TestTransitChart_ToXML_Roundtrip(t *testing.T) {
	t.Parallel()

	natal := &BaseChart{
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
		},
		Sidereal: map[string]Position{
			"Sun":   {Lon: 303.30, Lat: 0.0, Speed: 1.01},
			"Moon":  {Lon: 274.67, Lat: -4.2, Speed: 13.5},
		},
		Ayanamsa:  23.45,
		ASC:       187.5,
		MC:        97.3,
		DSC:       7.5,
		IC:        277.3,
		NorthNode: 2.16,
		SouthNode: 182.16,
		Houses: map[string][]float64{
			"placidus":   {0, 187.5, 217.3, 247.1, 277.3, 307.5, 337.3, 7.5, 37.3, 67.1, 97.3, 127.5, 157.3},
			"whole_sign": {0, 180.0, 210.0, 240.0, 270.0, 300.0, 330.0, 0.0, 30.0, 60.0, 90.0, 120.0, 150.0},
		},
		JD: 2440268.798611,
	}

	tc := &TransitChart{
		Name:             "AJ",
		Year:             2026,
		Month:            7,
		Day:              15,
		Hour:             12,
		Minute:           0,
		Second:           0,
		TZOffset:         7.0,
		Lat:              7.8,
		Lng:              98.3,
		Natal:            natal,
		TransitTropical: map[string]Position{
			"Sun":     {Lon: 112.5, Lat: 0.0, Speed: 0.95},
			"Mercury": {Lon: 125.3, Lat: -2.1, Speed: 1.5},
		},
		TransitSidereal: map[string]Position{
			"Sun":     {Lon: 88.3, Lat: 0.0, Speed: 0.95},
			"Mercury": {Lon: 101.1, Lat: -2.1, Speed: 1.5},
		},
		TransitAyanamsa:  24.2,
		TransitASC:       195.0,
		TransitMC:        105.0,
		TransitDSC:       15.0,
		TransitIC:        285.0,
		TransitNorthNode: 5.5,
		TransitSouthNode: 185.5,
		TransitHouses: map[string][]float64{
			"whole_sign": {0, 195.0, 225.0, 255.0, 285.0, 315.0, 345.0, 15.0, 45.0, 75.0, 105.0, 135.0, 165.0},
		},
		TransitJD:    2461234.5,
		TransitDayJD: 2461234,
	}

	xmlBytes, err := tc.ToXML()
	if err != nil {
		t.Fatalf("ToXML failed: %v", err)
	}

	xmlStr := string(xmlBytes)
	if !strings.Contains(xmlStr, "<?xml") {
		t.Error("XML output missing XML declaration")
	}
	if !strings.Contains(xmlStr, "TransitChart") {
		t.Error("XML output missing TransitChart root element")
	}
	if !strings.Contains(xmlStr, "AJ") {
		t.Error("XML output missing chart name")
	}
	if !strings.Contains(xmlStr, "Natal") {
		t.Error("XML output missing Natal section")
	}
	if !strings.Contains(xmlStr, "Transits") {
		t.Error("XML output missing Transits section")
	}
	if !strings.Contains(xmlStr, "Moment") {
		t.Error("XML output missing Moment section")
	}

	// Deserialize back
	tc2, err := TransitChartFromXML(xmlBytes)
	if err != nil {
		t.Fatalf("TransitChartFromXML failed: %v", err)
	}

	if tc2.Name != tc.Name {
		t.Errorf("Name = %q, want %q", tc2.Name, tc.Name)
	}
	if tc2.Year != tc.Year {
		t.Errorf("Year = %d, want %d", tc2.Year, tc.Year)
	}
	if tc2.Lat != tc.Lat {
		t.Errorf("Lat = %f, want %f", tc2.Lat, tc.Lat)
	}
	if tc2.TransitJD != tc.TransitJD {
		t.Errorf("TransitJD = %f, want %f", tc2.TransitJD, tc.TransitJD)
	}
	if tc2.TransitASC != tc.TransitASC {
		t.Errorf("TransitASC = %f, want %f", tc2.TransitASC, tc.TransitASC)
	}
	if tc2.TransitAyanamsa != tc.TransitAyanamsa {
		t.Errorf("TransitAyanamsa = %f, want %f", tc2.TransitAyanamsa, tc.TransitAyanamsa)
	}
	if tc2.TransitNorthNode != tc.TransitNorthNode {
		t.Errorf("TransitNorthNode = %f, want %f", tc2.TransitNorthNode, tc.TransitNorthNode)
	}

	// Check transit positions
	if len(tc2.TransitTropical) != len(tc.TransitTropical) {
		t.Errorf("TransitTropical count = %d, want %d", len(tc2.TransitTropical), len(tc.TransitTropical))
	}
	for name, pos := range tc.TransitTropical {
		pos2, ok := tc2.TransitTropical[name]
		if !ok {
			t.Errorf("missing transit tropical planet %q", name)
			continue
		}
		if pos2.Lon != pos.Lon || pos2.Lat != pos.Lat || pos2.Speed != pos.Speed {
			t.Errorf("TransitTropical[%q] = %+v, want %+v", name, pos2, pos)
		}
	}

	// Check natal survived roundtrip
	if tc2.Natal == nil {
		t.Fatal("Natal is nil after roundtrip")
	}
	if tc2.Natal.Name != natal.Name {
		t.Errorf("Natal.Name = %q, want %q", tc2.Natal.Name, natal.Name)
	}
	if tc2.Natal.ASC != natal.ASC {
		t.Errorf("Natal.ASC = %f, want %f", tc2.Natal.ASC, natal.ASC)
	}
}

// TestTransitChart_ToXML_ValidXML verifies the output is well-formed XML.
func TestTransitChart_ToXML_ValidXML(t *testing.T) {
	t.Parallel()

	natal := &BaseChart{
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
		JD: 2451545.0,
	}

	tc := &TransitChart{
		Name:             "Test",
		Year:             2026,
		Month:            7,
		Day:              15,
		Hour:             12,
		Minute:           0,
		Second:           0,
		TZOffset:         0,
		Lat:              51.5,
		Lng:              -0.12,
		Natal:            natal,
		TransitTropical:  map[string]Position{},
		TransitSidereal:  map[string]Position{},
		TransitHouses:    map[string][]float64{},
		TransitJD:        2461234.5,
		TransitDayJD:     2461234,
	}

	xmlBytes, err := tc.ToXML()
	if err != nil {
		t.Fatalf("ToXML failed: %v", err)
	}

	var doc struct {
		XMLName xml.Name `xml:"TransitChart"`
	}
	if err := xml.Unmarshal(xmlBytes, &doc); err != nil {
		t.Fatalf("output is not valid XML: %v", err)
	}
}

// TestTransitChart_ToXML_EmptyChart verifies that an empty chart serializes
// without error (nil maps, empty slices).
func TestTransitChart_ToXML_EmptyChart(t *testing.T) {
	t.Parallel()

	natal := &BaseChart{
		Name:        "Empty",
		Tropical:    map[string]Position{},
		Sidereal:    map[string]Position{},
		Houses:      map[string][]float64{},
		Aspects:     []AspectHit{},
		FixedStars:  []StarConjunction{},
		ArabicParts: map[string]float64{},
	}

	tc := &TransitChart{
		Name:             "Empty",
		Natal:            natal,
		TransitTropical:  map[string]Position{},
		TransitSidereal:  map[string]Position{},
		TransitHouses:    map[string][]float64{},
	}

	xmlBytes, err := tc.ToXML()
	if err != nil {
		t.Fatalf("ToXML on empty chart failed: %v", err)
	}

	var doc struct {
		XMLName xml.Name `xml:"TransitChart"`
	}
	if err := xml.Unmarshal(xmlBytes, &doc); err != nil {
		t.Fatalf("empty chart output is not valid XML: %v", err)
	}
}
