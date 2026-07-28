package dignity

import (
	"github.com/aj-nt/empirical/internal/swe"
)

// ── TransitChart ────────────────────────────────────────────────────────────

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

// ComputeTransitChart computes a TransitChart for a given natal chart and
// transit datetime.
func ComputeTransitChart(natal *BaseChart, year, month, day, hour, minute, second int, tzOff, lat, lng float64) (*TransitChart, error) {
	utHour := float64(hour) + float64(minute)/60.0 + float64(second)/3600.0 - tzOff
	jd := swe.Julday(year, month, day, utHour, true)
	ayan := swe.GetAyanamsaUT(jd)

	// Transit planet positions
	tropical := make(map[string]Position)
	sidereal := make(map[string]Position)
	for _, p := range AllPlanets {
		// Skip synthetic bodies (SouthNode is NN+180°, computed below)
		if p.ID < 0 {
			continue
		}
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

	// SouthNode as synthetic body in transit maps
	snSidLon := snLon - ayan
	if snSidLon < 0 {
		snSidLon += 360
	}
	tropical["SouthNode"] = Position{Lon: snLon, Lat: 0, Speed: 0, Dist: 0}
	sidereal["SouthNode"] = Position{Lon: snSidLon, Lat: 0, Speed: 0, Dist: 0}

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
