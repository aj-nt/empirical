package main

import (
	"github.com/aj-nt/empirical/internal/dignity"
	"github.com/aj-nt/empirical/internal/server"
)

// TransitHitJSON is a single transit hit in JSON output.
type TransitHitJSON struct {
	TransitPlanet string  `json:"transit_planet"`
	NatalPlanet   string  `json:"natal_planet"`
	Aspect        string  `json:"aspect"`
	Orb           float64 `json:"orb"`
	StartDate     string  `json:"start_date"`
	EndDate       string  `json:"end_date"`
}

// TransitsResponse is the JSON response for /api/transits.
type TransitsResponse struct {
	Name       string          `json:"name"`
	Sidereal   bool            `json:"sidereal"`
	Transits   []TransitHitJSON `json:"transits"`
	SkyWeather []TransitHitJSON `json:"sky_weather"`
}

// SynastryResponse is the JSON response for /api/synastry.
type SynastryResponse struct {
	Name1   string                `json:"name1"`
	Name2   string                `json:"name2"`
	Aspects []dignity.SynastryHit `json:"aspects"`
}

// RelocationResponse is the JSON response for /api/relocation-compare.
type RelocationResponse struct {
	Name               string              `json:"name"`
	LocationA          server.LatLng       `json:"location_a"`
	LocationB          server.LatLng       `json:"location_b"`
	TargetDate         string              `json:"target_date"`
	HouseConvergenceA  RelocConvergence    `json:"house_convergence_a"`
	HouseConvergenceB  RelocConvergence    `json:"house_convergence_b"`
	ASCA               []RelocASCEntry     `json:"asc_a"`
	ASCB               []RelocASCEntry     `json:"asc_b"`
	TimingConvergenceA RelocTimingSummary  `json:"timing_convergence_a"`
	TimingConvergenceB RelocTimingSummary  `json:"timing_convergence_b"`
	Shifts             []RelocShiftEntry   `json:"shifts"`
}

// RelocConvergence holds house convergence stats for one location.
type RelocConvergence struct {
	UnambiguousCount int     `json:"unambiguous_count"`
	ConvergenceRate  float64 `json:"convergence_rate"`
}

// RelocASCEntry holds ASC data for one house system.
type RelocASCEntry struct {
	System string  `json:"system"`
	Sign   string  `json:"sign"`
	Degree float64 `json:"degree"`
}

// RelocTimingSummary holds timing convergence summary.
type RelocTimingSummary struct {
	HasConvergence bool     `json:"has_convergence"`
	Planets        []string `json:"planets"`
}

// RelocShiftEntry holds a house shift comparison between two locations.
type RelocShiftEntry struct {
	Planet        string `json:"planet"`
	TropicalSign  string `json:"tropical_sign"`
	HouseA        int    `json:"house_a"`
	HouseB        int    `json:"house_b"`
	AgreementA    int    `json:"agreement_a"`
	AgreementB    int    `json:"agreement_b"`
	StableA       bool   `json:"stable_a"`
	StableB       bool   `json:"stable_b"`
	ShiftReliable bool   `json:"shift_reliable"`
}

// DraconicShiftJSON is a single draconic sign shift entry.
type DraconicShiftJSON struct {
	Planet   string `json:"planet"`
	TropSign string `json:"tropical_sign"`
	DracSign string `json:"draconic_sign"`
}

// DraconicResponse is the JSON response for /api/draconic.
type DraconicResponse struct {
	Name    string                `json:"name"`
	Offset  float64               `json:"offset"`
	Planets map[string]float64    `json:"planets"`
	Shifts  []DraconicShiftJSON   `json:"sign_shifts"`
	Bridges []dignity.SynastryHit `json:"bridges"`
}

// DraconicSynastryResponse is the JSON response for /api/draconic-synastry.
type DraconicSynastryResponse struct {
	Name1 string                `json:"name1"`
	Name2 string                `json:"name2"`
	Hits  []dignity.SynastryHit `json:"hits"`
}

// DraconicSynastryFullResponse is the JSON response for /api/draconic-synastry-full.
type DraconicSynastryFullResponse struct {
	Name1        string                `json:"name1"`
	Name2        string                `json:"name2"`
	DracToDrac   []dignity.SynastryHit `json:"drac_to_drac"`
	TropAToDracB []dignity.SynastryHit `json:"trop_a_to_drac_b"`
	TropBToDracA []dignity.SynastryHit `json:"trop_b_to_drac_a"`
}

// DraconicTransitsResponse is the JSON response for /api/draconic-transits.
type DraconicTransitsResponse struct {
	Name     string          `json:"name"`
	Offset   float64         `json:"offset"`
	Transits []TransitHitJSON `json:"transits"`
}

// ProgressedDraconicResponse is the JSON response for /api/progressed-draconic.
type ProgressedDraconicResponse struct {
	Name          string             `json:"name"`
	Date          string             `json:"date"`
	NatalNN       float64            `json:"natal_nn"`
	CurrentNN     float64            `json:"current_nn"`
	NNShift       float64            `json:"nn_shift"`
	NatalDraconic map[string]float64 `json:"natal_draconic"`
	ProgDraconic  map[string]float64 `json:"progressed_draconic"`
	SignShifts    []DraconicShiftJSON `json:"sign_shifts"`
}

// DraconicSolarReturnResponse is the JSON response for /api/draconic-solar-return.
type DraconicSolarReturnResponse struct {
	Name              string             `json:"name"`
	TargetYear        int                `json:"target_year"`
	JD                float64            `json:"jd"`
	DateTime          string             `json:"datetime"`
	Tropical          map[string]float64 `json:"tropical"`
	Draconic          map[string]float64 `json:"draconic"`
	DraconicByNatalNN map[string]float64 `json:"draconic_by_natal_nn"`
}

// CrossHitJSON is a single cross-system comparison hit.
type CrossHitJSON struct {
	TransitPlanet string  `json:"transit_planet"`
	NatalPlanet   string  `json:"natal_planet"`
	Aspect        string  `json:"aspect"`
	Orb           float64 `json:"orb"`
}

// DraconicTransitsCrossResponse is the JSON response for /api/draconic-transits-cross.
type DraconicTransitsCrossResponse struct {
	Name         string        `json:"name"`
	Offset       float64       `json:"offset"`
	Ayanamsa     float64       `json:"ayanamsa"`
	Orb          float64       `json:"orb"`
	MidDate      string        `json:"mid_date"`
	Survivors    []CrossHitJSON `json:"survivors"`
	TropicalOnly []CrossHitJSON `json:"tropical_only"`
	SiderealOnly []CrossHitJSON `json:"sidereal_only"`
}

// ProgressedCrossHitJSON is a single progressed cross-system hit.
type ProgressedCrossHitJSON struct {
	ProgressedPlanet string  `json:"progressed_planet"`
	NatalPlanet      string  `json:"natal_planet"`
	Aspect           string  `json:"aspect"`
	Orb              float64 `json:"orb"`
}

// ProgressedCrossResponse is the JSON response for /api/progressed-cross.
type ProgressedCrossResponse struct {
	Name         string                   `json:"name"`
	TargetDate   string                   `json:"target_date"`
	Age          float64                  `json:"age_years"`
	Ayanamsa     float64                  `json:"ayanamsa"`
	Orb          float64                  `json:"orb"`
	Survivors    []ProgressedCrossHitJSON `json:"survivors"`
	TropicalOnly []ProgressedCrossHitJSON `json:"tropical_only"`
	SiderealOnly []ProgressedCrossHitJSON `json:"sidereal_only"`
}

// DirectionHitJSON is a single primary direction hit.
type DirectionHitJSON struct {
	DirectedPoint string  `json:"directed_point"`
	NatalPlanet   string  `json:"natal_planet"`
	Aspect        string  `json:"aspect"`
	Orb           float64 `json:"orb"`
}

// DirectionsResponse is the JSON response for /api/directions.
type DirectionsResponse struct {
	Name        string             `json:"name"`
	Age         float64            `json:"age_years"`
	DirectedASC float64            `json:"directed_asc"`
	DirectedMC  float64            `json:"directed_mc"`
	Orb         float64            `json:"orb"`
	ASCAspects  []DirectionHitJSON `json:"asc_aspects"`
	MCAspects   []DirectionHitJSON `json:"mc_aspects"`
}

// AstroCartographyLineJSON is a single planetary line in astrocartography output.
type AstroCartographyLineJSON struct {
	Planet string            `json:"planet"`
	Angle  string            `json:"angle"`
	Points []dignity.GeoPoint `json:"points"`
}

// AstroCartographyResponse is the JSON response for /api/astrocartography.
type AstroCartographyResponse struct {
	Name  string                     `json:"name"`
	JD    float64                    `json:"jd"`
	GMST  float64                    `json:"gmst"`
	Frame string                     `json:"frame"`
	Lines []AstroCartographyLineJSON `json:"lines"`
}

// AstroCartographyCompareResponse is the JSON response for /api/astrocartography-compare.
type AstroCartographyCompareResponse struct {
	Name      string                `json:"name"`
	TargetLat float64               `json:"target_lat"`
	TargetLng float64               `json:"target_lng"`
	Orb       float64               `json:"orb"`
	Hits      []dignity.ThreeWayHit `json:"hits"`
}

// AstroCartographyParansResponse is the JSON response for /api/astrocartography-parans.
type AstroCartographyParansResponse struct {
	Name        string                      `json:"name"`
	Frame       string                      `json:"frame"`
	Intersections []dignity.ParanIntersection `json:"intersections"`
}

// ElectionalDayScore is a single day's score in electional output.
type ElectionalDayScore struct {
	Date      string   `json:"date"`
	Day       string   `json:"day"`
	Score     int      `json:"score"`
	MoonHouse int      `json:"moon_house"`
	MoonSign  string   `json:"moon_sign"`
	MercSign  string   `json:"merc_sign"`
	Good      []string `json:"good"`
	Bad       []string `json:"bad"`
}

// ElectionalResponse is the JSON response for /api/electional.
type ElectionalResponse struct {
	Name    string               `json:"name"`
	Start   string               `json:"start_date"`
	End     string               `json:"end_date"`
	Orb     float64              `json:"orb"`
	Results []ElectionalDayScore `json:"results"`
}

// StarConjJSON is a single star conjunction in JSON output.
type StarConjJSON struct {
	Star      string  `json:"star"`
	StarLon   float64 `json:"star_lon"`
	Planet    string  `json:"planet"`
	PlanetLon float64 `json:"planet_lon"`
	Orb       float64 `json:"orb"`
	Meaning   string  `json:"meaning"`
}

// StarsResponse is the JSON response for /api/stars.
type StarsResponse struct {
	Name         string        `json:"name"`
	Orb          float64       `json:"orb"`
	Conjunctions []StarConjJSON `json:"conjunctions"`
}

// SolarReturnResponse is the JSON response for /api/solar-return.
type SolarReturnResponse struct {
	Name       string                `json:"name"`
	TargetYear int                   `json:"target_year"`
	JD         float64               `json:"jd"`
	DateTime   string                `json:"datetime"`
	Positions  map[string]float64    `json:"positions"`
	Aspects    []dignity.SynastryHit `json:"aspects"`
	Patterns   []dignity.Pattern     `json:"patterns"`
}

// computePositions calculates planet longitudes, ayanamsa, ASC, and NN.