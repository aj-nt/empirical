package dignity

import (
	"encoding/xml"
	"fmt"

	"github.com/aj-nt/empirical/internal/swe"
)

// ── TransitChart XML types ─────────────────────────────────────────────────

type xmlTransitChart struct {
	XMLName  xml.Name         `xml:"TransitChart"`
	Version  string           `xml:"version,attr"`
	Identity xmlIdentity      `xml:"Identity"`
	Moment   xmlTime          `xml:"Moment"`
	Location xmlLocation      `xml:"Location"`
	Natal    xmlBaseChartBody `xml:"Natal"`    // embedded BaseChart (no XMLName)
	Transits xmlPositions     `xml:"Transits"` // transit planet positions
	Angles   xmlAngles        `xml:"Angles"`   // transit moment angles
	Nodes    xmlNodes         `xml:"Nodes"`    // transit moment nodes
	Houses   xmlHouses        `xml:"Houses"`   // transit moment houses
}

// ── TransitChart (Go-side) ─────────────────────────────────────────────────

// TransitChart holds a natal chart + transiting positions for one moment.
type TransitChart struct {
	// Identity
	Name string

	// Transit moment
	Year, Month, Day, Hour, Minute, Second int
	TZOffset                               float64
	Lat, Lng                               float64

	// Natal chart (embedded)
	Natal *BaseChart

	// Transit positions
	TransitTropical map[string]Position
	TransitSidereal map[string]Position
	TransitAyanamsa float64

	// Transit angles
	TransitASC, TransitMC, TransitDSC, TransitIC float64

	// Transit nodes
	TransitNorthNode, TransitSouthNode float64

	// Transit houses
	TransitHouses map[string][]float64

	// Transit Julian Day
	TransitJD    float64
	TransitDayJD int
}

// ToXML serializes a TransitChart to XML.
func (tc *TransitChart) ToXML() ([]byte, error) {
	// Serialize natal BaseChart
	natalXML := xmlBaseChartBody{
		Version: "1.0",
		Identity: xmlIdentity{
			Name: tc.Natal.Name,
		},
		Time: xmlTime{
			Year:     tc.Natal.Year,
			Month:    tc.Natal.Month,
			Day:      tc.Natal.Day,
			Hour:     tc.Natal.Hour,
			Minute:   tc.Natal.Minute,
			Second:   tc.Natal.Second,
			TZOffset: tc.Natal.TZOffset,
			JD:       tc.Natal.JD,
			DayJD:    tc.Natal.DayJD,
			GMST:     tc.Natal.GMST,
		},
		Location: xmlLocation{
			Lat: tc.Natal.Lat,
			Lng: tc.Natal.Lng,
		},
		Positions: xmlPositions{
			Ayanamsa: tc.Natal.Ayanamsa,
			Planets:  positionsToXML(tc.Natal.Tropical, tc.Natal.Sidereal),
		},
		Angles: xmlAngles{
			ASC: tc.Natal.ASC,
			MC:  tc.Natal.MC,
			DSC: tc.Natal.DSC,
			IC:  tc.Natal.IC,
		},
		Nodes: xmlNodes{
			NorthNode: tc.Natal.NorthNode,
			SouthNode: tc.Natal.SouthNode,
		},
		Houses:      housesToXML(tc.Natal.Houses),
		FixedStars:  starsToXML(tc.Natal.FixedStars),
		ArabicParts: partsToXML(tc.Natal.ArabicParts),
		Aspects:     aspectsToXML(tc.Natal.Aspects),
	}

	x := xmlTransitChart{
		Version: "1.0",
		Identity: xmlIdentity{
			Name: tc.Name,
		},
		Moment: xmlTime{
			Year:     tc.Year,
			Month:    tc.Month,
			Day:      tc.Day,
			Hour:     tc.Hour,
			Minute:   tc.Minute,
			Second:   tc.Second,
			TZOffset: tc.TZOffset,
			JD:       tc.TransitJD,
			DayJD:    tc.TransitDayJD,
		},
		Location: xmlLocation{
			Lat: tc.Lat,
			Lng: tc.Lng,
		},
		Natal: natalXML,
		Transits: xmlPositions{
			Ayanamsa: tc.TransitAyanamsa,
			Planets:  positionsToXML(tc.TransitTropical, tc.TransitSidereal),
		},
		Angles: xmlAngles{
			ASC: tc.TransitASC,
			MC:  tc.TransitMC,
			DSC: tc.TransitDSC,
			IC:  tc.TransitIC,
		},
		Nodes: xmlNodes{
			NorthNode: tc.TransitNorthNode,
			SouthNode: tc.TransitSouthNode,
		},
		Houses: housesToXML(tc.TransitHouses),
	}

	output, err := xml.MarshalIndent(x, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("xml marshal: %w", err)
	}

	result := []byte(xml.Header)
	result = append(result, output...)
	return result, nil
}

// TransitChartFromXML deserializes an XML byte slice into a TransitChart.
func TransitChartFromXML(data []byte) (*TransitChart, error) {
	var x xmlTransitChart
	if err := xml.Unmarshal(data, &x); err != nil {
		return nil, fmt.Errorf("xml unmarshal: %w", err)
	}

	natal := &BaseChart{
		Name:        x.Natal.Identity.Name,
		Year:        x.Natal.Time.Year,
		Month:       x.Natal.Time.Month,
		Day:         x.Natal.Time.Day,
		Hour:        x.Natal.Time.Hour,
		Minute:      x.Natal.Time.Minute,
		Second:      x.Natal.Time.Second,
		TZOffset:    x.Natal.Time.TZOffset,
		Lat:         x.Natal.Location.Lat,
		Lng:         x.Natal.Location.Lng,
		Ayanamsa:    x.Natal.Positions.Ayanamsa,
		ASC:         x.Natal.Angles.ASC,
		MC:          x.Natal.Angles.MC,
		DSC:         x.Natal.Angles.DSC,
		IC:          x.Natal.Angles.IC,
		NorthNode:   x.Natal.Nodes.NorthNode,
		SouthNode:   x.Natal.Nodes.SouthNode,
		JD:          x.Natal.Time.JD,
		DayJD:       x.Natal.Time.DayJD,
		GMST:        x.Natal.Time.GMST,
		Tropical:    positionsFromXML(x.Natal.Positions.Planets, false),
		Sidereal:    positionsFromXML(x.Natal.Positions.Planets, true),
		Houses:      housesFromXML(x.Natal.Houses.Systems),
		FixedStars:  starsFromXML(x.Natal.FixedStars.Stars),
		ArabicParts: partsFromXML(x.Natal.ArabicParts.Parts),
		Aspects:     aspectsFromXML(x.Natal.Aspects.Hits),
	}

	return &TransitChart{
		Name:             x.Identity.Name,
		Year:             x.Moment.Year,
		Month:            x.Moment.Month,
		Day:              x.Moment.Day,
		Hour:             x.Moment.Hour,
		Minute:           x.Moment.Minute,
		Second:           x.Moment.Second,
		TZOffset:         x.Moment.TZOffset,
		Lat:              x.Location.Lat,
		Lng:              x.Location.Lng,
		Natal:            natal,
		TransitTropical:  positionsFromXML(x.Transits.Planets, false),
		TransitSidereal:  positionsFromXML(x.Transits.Planets, true),
		TransitAyanamsa:  x.Transits.Ayanamsa,
		TransitASC:       x.Angles.ASC,
		TransitMC:        x.Angles.MC,
		TransitDSC:       x.Angles.DSC,
		TransitIC:        x.Angles.IC,
		TransitNorthNode: x.Nodes.NorthNode,
		TransitSouthNode: x.Nodes.SouthNode,
		TransitHouses:    housesFromXML(x.Houses.Systems),
		TransitJD:        x.Moment.JD,
		TransitDayJD:     x.Moment.DayJD,
	}, nil
}

// ComputeTransitChart computes a TransitChart for a given natal chart and
// transit datetime. The caller gets XML and each system's XSLT does the rest.
func ComputeTransitChart(natal *BaseChart, year, month, day, hour, minute, second int, tzOff, lat, lng float64) (*TransitChart, error) {
	utHour := float64(hour) + float64(minute)/60.0 + float64(second)/3600.0 - tzOff
	jd := swe.Julday(year, month, day, utHour, true)
	ayan := swe.GetAyanamsaUT(jd)

	// Transit planet positions
	tropical := make(map[string]Position)
	sidereal := make(map[string]Position)
	for _, p := range AllPlanets {
		lon, lat, _, speed := swe.CalcUT(jd, p.ID)
		tropical[p.Name] = Position{Lon: lon, Lat: lat, Speed: speed}
		sidLon := lon - ayan
		if sidLon < 0 {
			sidLon += 360
		}
		sidereal[p.Name] = Position{Lon: sidLon, Lat: lat, Speed: speed}
	}

	// Transit nodes
	nnLon, _, _, _ := swe.CalcUT(jd, swe.MEAN_NODE)
	snLon := nnLon + 180
	if snLon >= 360 {
		snLon -= 360
	}

	// Transit houses
	houses := make(map[string][]float64)
	var asc, mc float64
	for _, hs := range []struct {
		name string
		code byte
	}{
		{"placidus", 'P'},
		{"whole_sign", 'W'},
		{"equal", 'E'},
		{"porphyry", 'O'},
		{"koch", 'K'},
	} {
		cusps, ascmc := swe.Houses(jd, lat, lng, hs.code)
		houseCusps := make([]float64, 13)
		copy(houseCusps[1:], cusps[1:13])
		houses[hs.name] = houseCusps
		if hs.name == "placidus" {
			asc = ascmc[0]
			mc = ascmc[1]
		}
	}

	dsc := asc + 180
	if dsc >= 360 {
		dsc -= 360
	}
	ic := mc + 180
	if ic >= 360 {
		ic -= 360
	}

	dayJD := int(swe.Julday(year, month, day, 0, true))

	return &TransitChart{
		Name:             natal.Name,
		Year:             year,
		Month:            month,
		Day:              day,
		Hour:             hour,
		Minute:           minute,
		Second:           second,
		TZOffset:         tzOff,
		Lat:              lat,
		Lng:              lng,
		Natal:            natal,
		TransitTropical:  tropical,
		TransitSidereal:  sidereal,
		TransitAyanamsa:  ayan,
		TransitASC:       asc,
		TransitMC:        mc,
		TransitDSC:       dsc,
		TransitIC:        ic,
		TransitNorthNode: nnLon,
		TransitSouthNode: snLon,
		TransitHouses:    houses,
		TransitJD:        jd,
		TransitDayJD:     dayJD,
	}, nil
}
