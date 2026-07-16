package dignity

import (
	"encoding/xml"
	"fmt"
)

// ── XML serialization types ────────────────────────────────────────────────
// These mirror BaseChart but use XML-tagged structs that encoding/xml can
// marshal/unmarshal directly. Maps are converted to ordered slices.

type xmlBaseChart struct {
	XMLName     xml.Name        `xml:"BaseChart"`
	Version     string          `xml:"version,attr"`
	Identity    xmlIdentity     `xml:"Identity"`
	Time        xmlTime         `xml:"Time"`
	Location    xmlLocation     `xml:"Location"`
	Positions   xmlPositions    `xml:"Positions"`
	Angles      xmlAngles       `xml:"Angles"`
	Nodes       xmlNodes        `xml:"Nodes"`
	Houses      xmlHouses       `xml:"Houses"`
	FixedStars  xmlFixedStars   `xml:"FixedStars,omitempty"`
	ArabicParts xmlArabicParts  `xml:"ArabicParts,omitempty"`
	Aspects     xmlAspects      `xml:"Aspects,omitempty"`
}

type xmlIdentity struct {
	Name string `xml:"Name"`
}

type xmlTime struct {
	Year     int     `xml:"Year"`
	Month    int     `xml:"Month"`
	Day      int     `xml:"Day"`
	Hour     int     `xml:"Hour"`
	Minute   int     `xml:"Minute"`
	Second   int     `xml:"Second"`
	TZOffset float64 `xml:"TZOffset"`
	JD       float64 `xml:"JD"`
	DayJD    int     `xml:"DayJD"` // Julian Day at midnight UTC (for Ba Zi day pillar)
	GMST     float64 `xml:"GMST"`
}

type xmlLocation struct {
	Lat float64 `xml:"Latitude"`
	Lng float64 `xml:"Longitude"`
}

type xmlPositions struct {
	Ayanamsa float64     `xml:"Ayanamsa"`
	Planets  []xmlPlanet `xml:"Planet"`
}

type xmlPlanet struct {
	Name     string    `xml:"name,attr"`
	ID       int       `xml:"id,attr"`
	Tropical xmlCoord  `xml:"Tropical"`
	Sidereal xmlCoord  `xml:"Sidereal"`
}

type xmlCoord struct {
	Lon   float64 `xml:"Lon"`
	Lat   float64 `xml:"Lat"`
	Speed float64 `xml:"Speed"`
}

type xmlAngles struct {
	ASC float64 `xml:"ASC"`
	MC  float64 `xml:"MC"`
	DSC float64 `xml:"DSC"`
	IC  float64 `xml:"IC"`
}

type xmlNodes struct {
	NorthNode float64 `xml:"NorthNode"`
	SouthNode float64 `xml:"SouthNode"`
}

type xmlHouses struct {
	Systems []xmlHouseSystem `xml:"System"`
}

type xmlHouseSystem struct {
	Name  string    `xml:"name,attr"`
	Cusps []float64 `xml:"Cusp"`
}

type xmlFixedStars struct {
	Stars []xmlStarConjunction `xml:"Star"`
}

type xmlStarConjunction struct {
	Name   string  `xml:"name,attr"`
	Planet string  `xml:"Planet"`
	Orb    float64 `xml:"Orb"`
}

type xmlArabicParts struct {
	Parts []xmlPart `xml:"Part"`
}

type xmlPart struct {
	Name string  `xml:"name,attr"`
	Lon  float64 `xml:"Lon"`
}

type xmlAspects struct {
	Hits []xmlAspectHit `xml:"Aspect"`
}

type xmlAspectHit struct {
	Planet1 string  `xml:"Planet1"`
	Planet2 string  `xml:"Planet2"`
	Type    string  `xml:"Type"`
	Orb     float64 `xml:"Orb"`
}

// ── ToXML serializes a BaseChart to XML ────────────────────────────────────

// ToXML serializes the BaseChart to an XML byte slice.
func (bc *BaseChart) ToXML() ([]byte, error) {
	x := xmlBaseChart{
		Version: "1.0",
		Identity: xmlIdentity{
			Name: bc.Name,
		},
		Time: xmlTime{
			Year:     bc.Year,
			Month:    bc.Month,
			Day:      bc.Day,
			Hour:     bc.Hour,
			Minute:   bc.Minute,
			Second:   bc.Second,
			TZOffset: bc.TZOffset,
			JD:       bc.JD,
			DayJD:    bc.DayJD,
			GMST:     bc.GMST,
		},
		Location: xmlLocation{
			Lat: bc.Lat,
			Lng: bc.Lng,
		},
		Positions: xmlPositions{
			Ayanamsa: bc.Ayanamsa,
			Planets:  positionsToXML(bc.Tropical, bc.Sidereal),
		},
		Angles: xmlAngles{
			ASC: bc.ASC,
			MC:  bc.MC,
			DSC: bc.DSC,
			IC:  bc.IC,
		},
		Nodes: xmlNodes{
			NorthNode: bc.NorthNode,
			SouthNode: bc.SouthNode,
		},
		Houses:     housesToXML(bc.Houses),
		FixedStars: starsToXML(bc.FixedStars),
		ArabicParts: partsToXML(bc.ArabicParts),
		Aspects:     aspectsToXML(bc.Aspects),
	}

	output, err := xml.MarshalIndent(x, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("xml marshal: %w", err)
	}

	// Prepend XML declaration
	result := []byte(xml.Header)
	result = append(result, output...)
	return result, nil
}

// ── FromXML deserializes XML back to a BaseChart ──────────────────────────

// FromXML deserializes an XML byte slice into a BaseChart.
func FromXML(data []byte) (*BaseChart, error) {
	var x xmlBaseChart
	if err := xml.Unmarshal(data, &x); err != nil {
		return nil, fmt.Errorf("xml unmarshal: %w", err)
	}

	bc := &BaseChart{
		Name:     x.Identity.Name,
		Year:     x.Time.Year,
		Month:    x.Time.Month,
		Day:      x.Time.Day,
		Hour:     x.Time.Hour,
		Minute:   x.Time.Minute,
		Second:   x.Time.Second,
		TZOffset: x.Time.TZOffset,
		Lat:      x.Location.Lat,
		Lng:      x.Location.Lng,
		Ayanamsa: x.Positions.Ayanamsa,
		ASC:      x.Angles.ASC,
		MC:       x.Angles.MC,
		DSC:      x.Angles.DSC,
		IC:       x.Angles.IC,
		NorthNode: x.Nodes.NorthNode,
		SouthNode: x.Nodes.SouthNode,
		JD:        x.Time.JD,
		DayJD:     x.Time.DayJD,
		GMST:      x.Time.GMST,
		Tropical:  positionsFromXML(x.Positions.Planets, false),
		Sidereal:  positionsFromXML(x.Positions.Planets, true),
		Houses:    housesFromXML(x.Houses.Systems),
		FixedStars: starsFromXML(x.FixedStars.Stars),
		ArabicParts: partsFromXML(x.ArabicParts.Parts),
		Aspects:     aspectsFromXML(x.Aspects.Hits),
	}

	return bc, nil
}

// ── Conversion helpers ─────────────────────────────────────────────────────

func positionsToXML(tropical, sidereal map[string]Position) []xmlPlanet {
	// Collect all planet names from both maps
	seen := make(map[string]bool)
	var planets []xmlPlanet

	// Use tropical as the canonical set
	for name, tp := range tropical {
		sp := sidereal[name]
		planets = append(planets, xmlPlanet{
			Name: name,
			ID:   planetID(name),
			Tropical: xmlCoord{
				Lon:   tp.Lon,
				Lat:   tp.Lat,
				Speed: tp.Speed,
			},
			Sidereal: xmlCoord{
				Lon:   sp.Lon,
				Lat:   sp.Lat,
				Speed: sp.Speed,
			},
		})
		seen[name] = true
	}
	// Add any sidereal-only planets
	for name, sp := range sidereal {
		if seen[name] {
			continue
		}
		planets = append(planets, xmlPlanet{
			Name: name,
			ID:   planetID(name),
			Sidereal: xmlCoord{
				Lon:   sp.Lon,
				Lat:   sp.Lat,
				Speed: sp.Speed,
			},
		})
	}
	return planets
}

func positionsFromXML(planets []xmlPlanet, sidereal bool) map[string]Position {
	m := make(map[string]Position, len(planets))
	for _, p := range planets {
		coord := p.Tropical
		if sidereal {
			coord = p.Sidereal
		}
		m[p.Name] = Position{
			Lon:   coord.Lon,
			Lat:   coord.Lat,
			Speed: coord.Speed,
		}
	}
	return m
}

func planetID(name string) int {
	for _, p := range AllPlanets {
		if p.Name == name {
			return p.ID
		}
	}
	return -1
}

func housesToXML(houses map[string][]float64) xmlHouses {
	var systems []xmlHouseSystem
	// Deterministic order
	for _, name := range []string{"placidus", "whole_sign", "equal", "porphyry", "koch"} {
		cusps, ok := houses[name]
		if !ok {
			continue
		}
		// Skip index 0 (unused in 1-indexed arrays)
		c := make([]float64, 0, 12)
		if len(cusps) > 1 {
			c = append(c, cusps[1:]...)
		}
		systems = append(systems, xmlHouseSystem{Name: name, Cusps: c})
	}
	return xmlHouses{Systems: systems}
}

func housesFromXML(systems []xmlHouseSystem) map[string][]float64 {
	m := make(map[string][]float64, len(systems))
	for _, s := range systems {
		// Rebuild 1-indexed array (index 0 = 0)
		cusps := make([]float64, 13)
		for i, c := range s.Cusps {
			if i < 12 {
				cusps[i+1] = c
			}
		}
		m[s.Name] = cusps
	}
	return m
}

func starsToXML(stars []StarConjunction) xmlFixedStars {
	var s []xmlStarConjunction
	for _, sc := range stars {
		s = append(s, xmlStarConjunction{
			Name:   sc.Star,
			Planet: sc.Planet,
			Orb:    sc.Orb,
		})
	}
	return xmlFixedStars{Stars: s}
}

func starsFromXML(stars []xmlStarConjunction) []StarConjunction {
	result := make([]StarConjunction, len(stars))
	for i, s := range stars {
		result[i] = StarConjunction{
			Star:   s.Name,
			Planet: s.Planet,
			Orb:    s.Orb,
		}
	}
	return result
}

func partsToXML(parts map[string]float64) xmlArabicParts {
	var p []xmlPart
	// Deterministic order
	partOrder := []string{
		"Fortune", "Spirit", "Basis", "Love", "Passion",
		"Necessity", "Eros", "Courage", "Victory",
		"Nemesis", "Debt", "Father", "Mother",
	}
	for _, name := range partOrder {
		lon, ok := parts[name]
		if !ok {
			continue
		}
		p = append(p, xmlPart{Name: name, Lon: lon})
	}
	// Add any parts not in the canonical order
	seen := make(map[string]bool)
	for _, name := range partOrder {
		seen[name] = true
	}
	for name, lon := range parts {
		if seen[name] {
			continue
		}
		p = append(p, xmlPart{Name: name, Lon: lon})
	}
	return xmlArabicParts{Parts: p}
}

func partsFromXML(parts []xmlPart) map[string]float64 {
	m := make(map[string]float64, len(parts))
	for _, p := range parts {
		m[p.Name] = p.Lon
	}
	return m
}

func aspectsToXML(aspects []AspectHit) xmlAspects {
	var hits []xmlAspectHit
	for _, a := range aspects {
		hits = append(hits, xmlAspectHit{
			Planet1: a.Planet1,
			Planet2: a.Planet2,
			Type:    a.Aspect,
			Orb:     a.Orb,
		})
	}
	return xmlAspects{Hits: hits}
}

func aspectsFromXML(hits []xmlAspectHit) []AspectHit {
	result := make([]AspectHit, len(hits))
	for i, h := range hits {
		result[i] = AspectHit{
			Planet1: h.Planet1,
			Planet2: h.Planet2,
			Aspect:  h.Type,
			Orb:     h.Orb,
		}
	}
	return result
}
