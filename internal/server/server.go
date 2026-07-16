package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"strings"

	"github.com/aj-nt/empirical/internal/dignity"
)

// Default orb values for endpoints that accept an optional orb parameter.
const (
	OrbTight    = 1.0 // directions, declination
	OrbNarrow   = 2.0 // stars, harmonic, parans
	OrbStandard = 3.0 // transits, draconic, interpretation, electional, composite, arabic-parts, progressed-cross
	OrbWide     = 5.0 // synastry, patterns
)

// BaseChartFunc computes a BaseChart and returns it as XML bytes.
type BaseChartFunc func(bd dignity.BirthData) ([]byte, error)

// ComputeFunc computes a full multi-phase recovery report for birth data
// and returns the result as JSON bytes.
type ComputeFunc func(bd dignity.BirthData) ([]byte, error)

// AspectFunc returns the aspect catalog as JSON bytes.
type AspectFunc func() ([]byte, error)

// TimingFunc computes timing layer convergence for a target date.
type TimingFunc func(bd dignity.BirthData, targetDate string) ([]byte, error)

// TransitFunc computes transits for a natal chart over a date range.
type TransitFunc func(bd dignity.BirthData, startDate, endDate string, orbDeg float64, sidereal bool) ([]byte, error)

// SynastryFunc computes inter-aspects between two natal charts.
type SynastryFunc func(name1 string, y1, mo1, d1, h1, mi1 int, tz1, la1, lo1 float64, name2 string, y2, mo2, d2, h2, mi2 int, tz2, la2, lo2 float64, orbDeg float64) ([]byte, error)

// RelocationFunc computes a cross-validated relocation comparison between two locations.
type RelocationFunc func(bd dignity.BirthData, locA LatLng, locB LatLng, targetDate string) ([]byte, error)

// ChartFunc renders a natal chart wheel as SVG.
type ChartFunc func(bd dignity.BirthData, houseSystem string, sidereal bool, showAspects bool, outerPlanets bool, highlightPatterns bool, patternOrb float64) (string, error)

// PatternFunc detects geometric patterns in a natal chart.
type PatternFunc func(bd dignity.BirthData, orbDeg float64) ([]byte, error)

// DraconicFunc computes the draconic chart, sign shifts, and bridges.
type DraconicFunc func(bd dignity.BirthData, orbDeg float64) ([]byte, error)

// DraconicSynastryFunc computes draconic synastry between two charts.
type DraconicSynastryFunc func(name1 string, y1, mo1, d1, h1, mi1 int, tz1, la1, lo1 float64, name2 string, y2, mo2, d2, h2, mi2 int, tz2, la2, lo2 float64, orbDeg float64) ([]byte, error)

// DraconicSynastryFullFunc computes the full three-layer draconic synastry.
type DraconicSynastryFullFunc func(name1 string, y1, mo1, d1, h1, mi1 int, tz1, la1, lo1 float64, name2 string, y2, mo2, d2, h2, mi2 int, tz2, la2, lo2 float64, orbDeg float64) ([]byte, error)

// DraconicTransitFunc computes draconic transits: transiting planets → draconic chart.
type DraconicTransitFunc func(bd dignity.BirthData, startDate, endDate string, orbDeg float64) ([]byte, error)

// ProgressedDraconicFunc computes the progressed draconic chart using the current transiting NN.
type ProgressedDraconicFunc func(bd dignity.BirthData, targetDate string) ([]byte, error)

// DraconicSolarReturnFunc computes the draconic solar return for a target year.
type DraconicSolarReturnFunc func(bd dignity.BirthData, targetYear int) ([]byte, error)

// StarsFunc computes fixed star conjunctions for a natal chart.
type StarsFunc func(bd dignity.BirthData, orbDeg float64) ([]byte, error)

// DraconicTransitsCrossFunc compares draconic transits in tropical vs sidereal.
type DraconicTransitsCrossFunc func(bd dignity.BirthData, startDate, endDate string, orbDeg float64) ([]byte, error)

// ProgressedCrossFunc compares progressed-to-natal aspects in tropical vs sidereal.
type ProgressedCrossFunc func(bd dignity.BirthData, targetDate string, orbDeg float64) ([]byte, error)

// DirectionsFunc computes primary directions (Ptolemy) for a given age.
type DirectionsFunc func(bd dignity.BirthData, age float64, orbDeg float64) ([]byte, error)

// InterpretationFunc produces natural-language chart interpretation.
// system: SystemKoiné (Hellenistic, default) or SystemWestern (modern).
type InterpretationFunc func(bd dignity.BirthData, houseSystem string, orbDeg float64, system string) ([]byte, error)

// AstroCartographyFunc computes planetary lines for a world map.
// frame: FrameTropical, FrameDraconic, or FrameCross.
type AstroCartographyFunc func(bd dignity.BirthData, latStep float64, frame string) ([]byte, error)

// AstroCartographyCompareFunc returns three-way comparison at a target location.
type AstroCartographyCompareFunc func(bd dignity.BirthData, latStep, targetLat, targetLng, orb float64) ([]byte, error)

// AstroCartographyParansFunc finds MC/IC × ASC/DSC line intersections.
type AstroCartographyParansFunc func(bd dignity.BirthData, latStep float64, frame string) ([]byte, error)

// ElectionalFunc scores dates in a range for launch/event timing.
type ElectionalFunc func(bd dignity.BirthData, startDate, endDate string, orbDeg float64) ([]byte, error)

// MansionConvergenceFunc computes nakshatra/xiu mansion placements per chart.
type MansionConvergenceFunc func(bd dignity.BirthData) ([]byte, error)

// ArabicPartsFunc computes Arabic Parts and cross-system comparison.
type ArabicPartsFunc func(bd dignity.BirthData, orbDeg float64) ([]byte, error)

// SolarReturnFunc computes a tropical solar return chart.
type SolarReturnFunc func(bd dignity.BirthData, targetYear int) ([]byte, error)

// CompositeFunc computes a midpoint composite chart for two people.
type CompositeFunc func(name1 string, y1, m1, d1, h1, min1 int, tz1, lat1, lng1 float64, name2 string, y2, m2, d2, h2, min2 int, tz2, lat2, lng2 float64, orbDeg float64) ([]byte, error)

// StarsCrossFunc compares star conjunctions in tropical vs sidereal frames.
type StarsCrossFunc func(bd dignity.BirthData, orbDeg float64) ([]byte, error)

// TraditionalFunc computes traditional astrology interpretive data.
type TraditionalFunc func(bd dignity.BirthData) ([]byte, error)

// UranianFunc computes Uranian/Hamburg School midpoint analysis.
type UranianFunc func(bd dignity.BirthData) ([]byte, error)

// HarmonicFunc computes Addey-style harmonic charts.
type HarmonicFunc func(bd dignity.BirthData, harmonics []int, orb float64) ([]byte, error)

// DivisionalFunc computes Vedic divisional charts (navamsha D9, nakshatras, dasha).
type DivisionalFunc func(bd dignity.BirthData) ([]byte, error)

// ParansFunc computes fixed star parans (star-planet angle contacts).
type ParansFunc func(bd dignity.BirthData, orb float64) ([]byte, error)

// DeclinationFunc computes declination parallels and contraparallels.
type DeclinationFunc func(bd dignity.BirthData, orb float64) ([]byte, error)

// FirdariaFunc computes Persian firdaria planetary periods.
type FirdariaFunc func(bd dignity.BirthData) ([]byte, error)

// TransitChartFunc computes a TransitChart (natal + transit positions) and
// returns it as XML bytes. Each system's XSLT does the interpretation.
type TransitChartFunc func(bd dignity.BirthData, year, month, day, hour, minute int, tzOff, lat, lng float64) ([]byte, error)

// LatLng holds a named geographic location.
type LatLng struct {
	Name string  `json:"name"`
	Lat  float64 `json:"lat"`
	Lng  float64 `json:"lng"`
}

// ServerConfig holds all function dependencies for the HTTP server.
// Use this struct instead of passing 35 individual parameters to NewMux and Run.
type ServerConfig struct {
	StaticFS              fs.FS
	BaseChart             BaseChartFunc
	Compute               ComputeFunc
	Aspects               AspectFunc
	Timing                TimingFunc
	Transits              TransitFunc
	Synastry              SynastryFunc
	Relocation            RelocationFunc
	Chart                 ChartFunc
	Patterns              PatternFunc
	Draconic              DraconicFunc
	DraconicSynastry      DraconicSynastryFunc
	DraconicSynastryFull  DraconicSynastryFullFunc
	DraconicTransits      DraconicTransitFunc
	ProgressedDraconic    ProgressedDraconicFunc
	DraconicSolarReturn   DraconicSolarReturnFunc
	Stars                 StarsFunc
	DraconicTransitsCross DraconicTransitsCrossFunc
	ProgressedCross       ProgressedCrossFunc
	Directions            DirectionsFunc
	Interpretation        InterpretationFunc
	AstroCartography      AstroCartographyFunc
	AstroCartographyCompare AstroCartographyCompareFunc
	AstroCartographyParans AstroCartographyParansFunc
	Electional            ElectionalFunc
	MansionConvergence    MansionConvergenceFunc
	ArabicParts           ArabicPartsFunc
	SolarReturn           SolarReturnFunc
	Composite             CompositeFunc
	StarsCross            StarsCrossFunc
	Traditional           TraditionalFunc
	Uranian               UranianFunc
	Harmonic              HarmonicFunc
	Divisional            DivisionalFunc
	Parans                ParansFunc
	Declination           DeclinationFunc
	Firdaria              FirdariaFunc
	TransitChart          TransitChartFunc
}

// ErrNotAvailable is returned by handler functions when an endpoint is not configured.
var ErrNotAvailable = errors.New("not available")

// handleJSON returns an http.HandlerFunc that decodes a JSON body of type T,
// calls fn with the decoded request, and writes the JSON result with CORS headers.
// If fn returns ErrNotAvailable, the handler responds with 501 Not Implemented.
func handleJSON[T any](fn func(T) ([]byte, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}
		var req T
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		result, err := fn(req)
		if err != nil {
			if errors.Is(err, ErrNotAvailable) {
				http.Error(w, "not available", http.StatusNotImplemented)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write(result)
	}
}

// NewMux builds the HTTP mux with all handlers wired to the provided functions.
// Exported so tests can exercise the real handlers with mock functions.
func NewMux(cfg ServerConfig) *http.ServeMux {
	mux := http.NewServeMux()

	if cfg.StaticFS != nil {
		mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(cfg.StaticFS))))
	}

	mux.HandleFunc("/api/recover", handleJSON(func(req ChartRequest) ([]byte, error) {
		return cfg.Compute(dignity.BirthData{
			Name:     req.Name,
			Year:     req.Year,
			Month:    req.Month,
			Day:      req.Day,
			Hour:     req.Hour,
			Minute:   req.Minute,
			TZOffset: req.TzOffset,
			Lat:      req.Lat,
			Lng:      req.Lng,
		})
	}))

	mux.HandleFunc("/api/base-chart", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}
		if cfg.BaseChart == nil {
			http.Error(w, "not available", http.StatusNotImplemented)
			return
		}
		var req ChartRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		result, err := cfg.BaseChart(dignity.BirthData{
			Name:     req.Name,
			Year:     req.Year,
			Month:    req.Month,
			Day:      req.Day,
			Hour:     req.Hour,
			Minute:   req.Minute,
			TZOffset: req.TzOffset,
			Lat:      req.Lat,
			Lng:      req.Lng,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/xml")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write(result)
	})

	mux.HandleFunc("/api/transit-chart", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}
		if cfg.TransitChart == nil {
			http.Error(w, "not available", http.StatusNotImplemented)
			return
		}
		var req TransitChartRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		result, err := cfg.TransitChart(dignity.BirthData{
			Name:     req.Name,
			Year:     req.Year,
			Month:    req.Month,
			Day:      req.Day,
			Hour:     req.Hour,
			Minute:   req.Minute,
			TZOffset: req.TzOffset,
			Lat:      req.Lat,
			Lng:      req.Lng,
		}, req.TransitYear, req.TransitMonth, req.TransitDay, req.TransitHour, req.TransitMinute, req.TransitTZ, req.TransitLat, req.TransitLng)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/xml")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write(result)
	})

	mux.HandleFunc("/api/aspect-catalog", func(w http.ResponseWriter, r *http.Request) {
		if cfg.Aspects == nil {
			http.Error(w, "not available", http.StatusNotImplemented)
			return
		}
		result, err := cfg.Aspects()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write(result)
	})

	mux.HandleFunc("/api/timing-convergence", handleJSON(func(req TimingRequest) ([]byte, error) {
		if cfg.Timing == nil {
			return nil, ErrNotAvailable
		}
		return cfg.Timing(dignity.BirthData{
		Name:     req.Name,
		Year:     req.Year,
		Month:    req.Month,
		Day:      req.Day,
		Hour:     req.Hour,
		Minute:   req.Minute,
		TZOffset: req.TzOffset,
		Lat:      req.Lat,
		Lng:      req.Lng,
	}, req.TargetDate)
	}))

	mux.HandleFunc("/api/transits", handleJSON(func(req TransitRequest) ([]byte, error) {
		if cfg.Transits == nil {
			return nil, ErrNotAvailable
		}
		return cfg.Transits(dignity.BirthData{
		Name:     req.Name,
		Year:     req.Year,
		Month:    req.Month,
		Day:      req.Day,
		Hour:     req.Hour,
		Minute:   req.Minute,
		TZOffset: req.TzOffset,
		Lat:      req.Lat,
		Lng:      req.Lng,
	}, req.StartDate, req.EndDate, defaultOrb(req.Orb, OrbStandard), req.Sidereal)
	}))

	mux.HandleFunc("/api/synastry", handleJSON(func(req SynastryRequest) ([]byte, error) {
		if cfg.Synastry == nil {
			return nil, ErrNotAvailable
		}
		return cfg.Synastry(req.Name1, req.Year1, req.Month1, req.Day1, req.Hour1, req.Min1, req.Tz1, req.Lat1, req.Lng1,
			req.Name2, req.Year2, req.Month2, req.Day2, req.Hour2, req.Min2, req.Tz2, req.Lat2, req.Lng2, defaultOrb(req.Orb, OrbWide))
	}))

	mux.HandleFunc("/api/relocation-compare", handleJSON(func(req RelocationRequest) ([]byte, error) {
		if cfg.Relocation == nil {
			return nil, ErrNotAvailable
		}
		return cfg.Relocation(dignity.BirthData{
		Name:     req.Name,
		Year:     req.Year,
		Month:    req.Month,
		Day:      req.Day,
		Hour:     req.Hour,
		Minute:   req.Minute,
		TZOffset: req.TzOffset,
		Lat:      req.Lat,
		Lng:      req.Lng,
	},
			req.LocationA, req.LocationB, req.TargetDate)
	}))

	mux.HandleFunc("/api/chart", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}
		var req ChartRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		if cfg.Chart == nil {
			http.Error(w, "not available", http.StatusNotImplemented)
			return
		}
		hs := req.HouseSystem
		if hs == "" {
			hs = "placidus"
		}
		result, err := cfg.Chart(dignity.BirthData{
		Name:     req.Name,
		Year:     req.Year,
		Month:    req.Month,
		Day:      req.Day,
		Hour:     req.Hour,
		Minute:   req.Minute,
		TZOffset: req.TzOffset,
		Lat:      req.Lat,
		Lng:      req.Lng,
	}, hs, req.Sidereal, req.ShowAspects, req.OuterPlanets, req.HighlightPatterns, req.PatternOrb)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "image/svg+xml")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write([]byte(result))
	})

	mux.HandleFunc("/api/patterns", handleJSON(func(req ChartRequest) ([]byte, error) {
		if cfg.Patterns == nil {
			return nil, ErrNotAvailable
		}
		return cfg.Patterns(dignity.BirthData{
		Name:     req.Name,
		Year:     req.Year,
		Month:    req.Month,
		Day:      req.Day,
		Hour:     req.Hour,
		Minute:   req.Minute,
		TZOffset: req.TzOffset,
		Lat:      req.Lat,
		Lng:      req.Lng,
	}, defaultOrb(req.Orb, OrbWide))
	}))

	mux.HandleFunc("/api/draconic", handleJSON(func(req ChartRequest) ([]byte, error) {
		if cfg.Draconic == nil {
			return nil, ErrNotAvailable
		}
		return cfg.Draconic(dignity.BirthData{
		Name:     req.Name,
		Year:     req.Year,
		Month:    req.Month,
		Day:      req.Day,
		Hour:     req.Hour,
		Minute:   req.Minute,
		TZOffset: req.TzOffset,
		Lat:      req.Lat,
		Lng:      req.Lng,
	}, defaultOrb(req.Orb, OrbStandard))
	}))

	mux.HandleFunc("/api/draconic-synastry", handleJSON(func(req SynastryRequest) ([]byte, error) {
		if cfg.DraconicSynastry == nil {
			return nil, ErrNotAvailable
		}
		return cfg.DraconicSynastry(req.Name1, req.Year1, req.Month1, req.Day1, req.Hour1, req.Min1, req.Tz1, req.Lat1, req.Lng1,
			req.Name2, req.Year2, req.Month2, req.Day2, req.Hour2, req.Min2, req.Tz2, req.Lat2, req.Lng2, defaultOrb(req.Orb, OrbStandard))
	}))

	mux.HandleFunc("/api/draconic-synastry-full", handleJSON(func(req SynastryRequest) ([]byte, error) {
		if cfg.DraconicSynastryFull == nil {
			return nil, ErrNotAvailable
		}
		return cfg.DraconicSynastryFull(req.Name1, req.Year1, req.Month1, req.Day1, req.Hour1, req.Min1, req.Tz1, req.Lat1, req.Lng1,
			req.Name2, req.Year2, req.Month2, req.Day2, req.Hour2, req.Min2, req.Tz2, req.Lat2, req.Lng2, defaultOrb(req.Orb, OrbStandard))
	}))

	mux.HandleFunc("/api/draconic-transits", handleJSON(func(req TransitRequest) ([]byte, error) {
		if cfg.DraconicTransits == nil {
			return nil, ErrNotAvailable
		}
		return cfg.DraconicTransits(dignity.BirthData{
		Name:     req.Name,
		Year:     req.Year,
		Month:    req.Month,
		Day:      req.Day,
		Hour:     req.Hour,
		Minute:   req.Minute,
		TZOffset: req.TzOffset,
		Lat:      req.Lat,
		Lng:      req.Lng,
	}, req.StartDate, req.EndDate, defaultOrb(req.Orb, OrbStandard))
	}))

	mux.HandleFunc("/api/progressed-draconic", handleJSON(func(req TimingRequest) ([]byte, error) {
		if cfg.ProgressedDraconic == nil {
			return nil, ErrNotAvailable
		}
		return cfg.ProgressedDraconic(dignity.BirthData{
		Name:     req.Name,
		Year:     req.Year,
		Month:    req.Month,
		Day:      req.Day,
		Hour:     req.Hour,
		Minute:   req.Minute,
		TZOffset: req.TzOffset,
		Lat:      req.Lat,
		Lng:      req.Lng,
	}, req.TargetDate)
	}))

	mux.HandleFunc("/api/draconic-solar-return", handleJSON(func(req DraconicSolarReturnRequest) ([]byte, error) {
		if cfg.DraconicSolarReturn == nil {
			return nil, ErrNotAvailable
		}
		return cfg.DraconicSolarReturn(dignity.BirthData{
		Name:     req.Name,
		Year:     req.Year,
		Month:    req.Month,
		Day:      req.Day,
		Hour:     req.Hour,
		Minute:   req.Minute,
		TZOffset: req.TzOffset,
		Lat:      req.Lat,
		Lng:      req.Lng,
	}, req.TargetYear)
	}))

	mux.HandleFunc("/api/stars", handleJSON(func(req ChartRequest) ([]byte, error) {
		if cfg.Stars == nil {
			return nil, ErrNotAvailable
		}
		return cfg.Stars(dignity.BirthData{
		Name:     req.Name,
		Year:     req.Year,
		Month:    req.Month,
		Day:      req.Day,
		Hour:     req.Hour,
		Minute:   req.Minute,
		TZOffset: req.TzOffset,
		Lat:      req.Lat,
		Lng:      req.Lng,
	}, defaultOrb(req.Orb, OrbNarrow))
	}))

	mux.HandleFunc("/api/draconic-transits-cross", handleJSON(func(req TransitRequest) ([]byte, error) {
		if cfg.DraconicTransitsCross == nil {
			return nil, ErrNotAvailable
		}
		return cfg.DraconicTransitsCross(dignity.BirthData{
		Name:     req.Name,
		Year:     req.Year,
		Month:    req.Month,
		Day:      req.Day,
		Hour:     req.Hour,
		Minute:   req.Minute,
		TZOffset: req.TzOffset,
		Lat:      req.Lat,
		Lng:      req.Lng,
	}, req.StartDate, req.EndDate, defaultOrb(req.Orb, OrbStandard))
	}))

	mux.HandleFunc("/api/progressed-cross", handleJSON(func(req TimingRequest) ([]byte, error) {
		if cfg.ProgressedCross == nil {
			return nil, ErrNotAvailable
		}
		return cfg.ProgressedCross(dignity.BirthData{
		Name:     req.Name,
		Year:     req.Year,
		Month:    req.Month,
		Day:      req.Day,
		Hour:     req.Hour,
		Minute:   req.Minute,
		TZOffset: req.TzOffset,
		Lat:      req.Lat,
		Lng:      req.Lng,
	}, req.TargetDate, defaultOrb(req.Orb, OrbStandard))
	}))

	mux.HandleFunc("/api/directions", handleJSON(func(req DirectionsRequest) ([]byte, error) {
		if cfg.Directions == nil {
			return nil, ErrNotAvailable
		}
		return cfg.Directions(dignity.BirthData{
		Name:     req.Name,
		Year:     req.Year,
		Month:    req.Month,
		Day:      req.Day,
		Hour:     req.Hour,
		Minute:   req.Minute,
		TZOffset: req.TzOffset,
		Lat:      req.Lat,
		Lng:      req.Lng,
	}, req.Age, defaultOrb(req.Orb, OrbTight))
	}))

	mux.HandleFunc("/api/interpretation", handleJSON(func(req ChartRequest) ([]byte, error) {
		if cfg.Interpretation == nil {
			return nil, ErrNotAvailable
		}
		system := req.System
		if system == "" {
			system = "koiné"
		}
		return cfg.Interpretation(dignity.BirthData{
		Name:     req.Name,
		Year:     req.Year,
		Month:    req.Month,
		Day:      req.Day,
		Hour:     req.Hour,
		Minute:   req.Minute,
		TZOffset: req.TzOffset,
		Lat:      req.Lat,
		Lng:      req.Lng,
	}, req.HouseSystem, defaultOrb(req.Orb, OrbStandard), system)
	}))

	mux.HandleFunc("/api/astrocartography", handleJSON(func(req AstroCartographyRequest) ([]byte, error) {
		if cfg.AstroCartography == nil {
			return nil, ErrNotAvailable
		}
		return cfg.AstroCartography(dignity.BirthData{
		Name:     req.Name,
		Year:     req.Year,
		Month:    req.Month,
		Day:      req.Day,
		Hour:     req.Hour,
		Minute:   req.Minute,
		TZOffset: req.TzOffset,
		Lat:      req.Lat,
		Lng:      req.Lng,
	}, defaultOrb(req.LatStep, 2.0), req.Frame)
	}))

	mux.HandleFunc("/api/astrocartography-compare", handleJSON(func(req AstroCartographyCompareRequest) ([]byte, error) {
		if cfg.AstroCartographyCompare == nil {
			return nil, ErrNotAvailable
		}
		return cfg.AstroCartographyCompare(dignity.BirthData{
		Name:     req.Name,
		Year:     req.Year,
		Month:    req.Month,
		Day:      req.Day,
		Hour:     req.Hour,
		Minute:   req.Minute,
		TZOffset: req.TzOffset,
		Lat:      req.Lat,
		Lng:      req.Lng,
	}, defaultOrb(req.LatStep, 2.0), req.TargetLat, req.TargetLng, defaultOrb(req.Orb, OrbStandard))
	}))

	mux.HandleFunc("/api/astrocartography-parans", handleJSON(func(req AstroCartographyRequest) ([]byte, error) {
		if cfg.AstroCartographyParans == nil {
			return nil, ErrNotAvailable
		}
		raw, err := cfg.AstroCartographyParans(dignity.BirthData{
		Name:     req.Name,
		Year:     req.Year,
		Month:    req.Month,
		Day:      req.Day,
		Hour:     req.Hour,
		Minute:   req.Minute,
		TZOffset: req.TzOffset,
		Lat:      req.Lat,
		Lng:      req.Lng,
	}, defaultOrb(req.LatStep, 2.0), req.Frame)
		if err != nil {
			return nil, err
		}

		// Apply optional filters
		if req.Planets != "" || req.MinLat != 0 || req.MaxLat != 0 || req.MinLng != 0 || req.MaxLng != 0 {
			var resp AstroCartographyParansResponse
			if err := json.Unmarshal(raw, &resp); err != nil {
				return nil, err
			}
			resp.Intersections = filterParans(resp.Intersections, req.Planets, req.MinLat, req.MaxLat, req.MinLng, req.MaxLng)
			return json.Marshal(resp)
		}
		return raw, nil
	}))

	mux.HandleFunc("/api/electional", handleJSON(func(req ElectionalRequest) ([]byte, error) {
		if cfg.Electional == nil {
			return nil, ErrNotAvailable
		}
		return cfg.Electional(dignity.BirthData{
		Name:     req.Name,
		Year:     req.Year,
		Month:    req.Month,
		Day:      req.Day,
		Hour:     req.Hour,
		Minute:   req.Minute,
		TZOffset: req.TzOffset,
		Lat:      req.Lat,
		Lng:      req.Lng,
	}, req.StartDate, req.EndDate, defaultOrb(req.Orb, OrbStandard))
	}))

	mux.HandleFunc("/api/mansion-convergence", handleJSON(func(req ChartRequest) ([]byte, error) {
		if cfg.MansionConvergence == nil {
			return nil, ErrNotAvailable
		}
		return cfg.MansionConvergence(dignity.BirthData{
		Name:     req.Name,
		Year:     req.Year,
		Month:    req.Month,
		Day:      req.Day,
		Hour:     req.Hour,
		Minute:   req.Minute,
		TZOffset: req.TzOffset,
		Lat:      req.Lat,
		Lng:      req.Lng,
	})
	}))

	mux.HandleFunc("/api/arabic-parts", handleJSON(func(req ArabicPartsRequest) ([]byte, error) {
		if cfg.ArabicParts == nil {
			return nil, ErrNotAvailable
		}
		return cfg.ArabicParts(dignity.BirthData{
		Name:     req.Name,
		Year:     req.Year,
		Month:    req.Month,
		Day:      req.Day,
		Hour:     req.Hour,
		Minute:   req.Minute,
		TZOffset: req.TzOffset,
		Lat:      req.Lat,
		Lng:      req.Lng,
	}, defaultOrb(req.Orb, OrbStandard))
	}))

	mux.HandleFunc("/api/solar-return", handleJSON(func(req struct {
		ChartRequest
		TargetYear int `json:"target_year"`
	}) ([]byte, error) {
		if cfg.SolarReturn == nil {
			return nil, ErrNotAvailable
		}
		if req.TargetYear == 0 {
			req.TargetYear = req.Year + 1
		}
		return cfg.SolarReturn(dignity.BirthData{
		Name:     req.Name,
		Year:     req.Year,
		Month:    req.Month,
		Day:      req.Day,
		Hour:     req.Hour,
		Minute:   req.Minute,
		TZOffset: req.TzOffset,
		Lat:      req.Lat,
		Lng:      req.Lng,
	}, req.TargetYear)
	}))

	mux.HandleFunc("/api/composite", handleJSON(func(req struct {
		Name1  string  `json:"name1"`
		Year1  int     `json:"year1"`
		Month1 int     `json:"month1"`
		Day1   int     `json:"day1"`
		Hour1  int     `json:"hour1"`
		Min1   int     `json:"min1"`
		Tz1    float64 `json:"tz1"`
		Lat1   float64 `json:"lat1"`
		Lng1   float64 `json:"lng1"`
		Name2  string  `json:"name2"`
		Year2  int     `json:"year2"`
		Month2 int     `json:"month2"`
		Day2   int     `json:"day2"`
		Hour2  int     `json:"hour2"`
		Min2   int     `json:"min2"`
		Tz2    float64 `json:"tz2"`
		Lat2   float64 `json:"lat2"`
		Lng2   float64 `json:"lng2"`
		Orb    float64 `json:"orb"`
	}) ([]byte, error) {
		if cfg.Composite == nil {
			return nil, ErrNotAvailable
		}
		return cfg.Composite(req.Name1, req.Year1, req.Month1, req.Day1, req.Hour1, req.Min1, req.Tz1, req.Lat1, req.Lng1, req.Name2, req.Year2, req.Month2, req.Day2, req.Hour2, req.Min2, req.Tz2, req.Lat2, req.Lng2, defaultOrb(req.Orb, OrbStandard))
	}))

	mux.HandleFunc("/api/stars-cross", handleJSON(func(req struct {
		ChartRequest
		Orb float64 `json:"orb"`
	}) ([]byte, error) {
		if cfg.StarsCross == nil {
			return nil, ErrNotAvailable
		}
		return cfg.StarsCross(dignity.BirthData{
		Name:     req.Name,
		Year:     req.Year,
		Month:    req.Month,
		Day:      req.Day,
		Hour:     req.Hour,
		Minute:   req.Minute,
		TZOffset: req.TzOffset,
		Lat:      req.Lat,
		Lng:      req.Lng,
	}, defaultOrb(req.Orb, OrbNarrow))
	}))

	mux.HandleFunc("/api/traditional", handleJSON(func(req ChartRequest) ([]byte, error) {
		if cfg.Traditional == nil {
			return nil, ErrNotAvailable
		}
		return cfg.Traditional(dignity.BirthData{
		Name:     req.Name,
		Year:     req.Year,
		Month:    req.Month,
		Day:      req.Day,
		Hour:     req.Hour,
		Minute:   req.Minute,
		TZOffset: req.TzOffset,
		Lat:      req.Lat,
		Lng:      req.Lng,
	})
	}))

	mux.HandleFunc("/api/uranian", handleJSON(func(req ChartRequest) ([]byte, error) {
		if cfg.Uranian == nil {
			return nil, ErrNotAvailable
		}
		return cfg.Uranian(dignity.BirthData{
		Name:     req.Name,
		Year:     req.Year,
		Month:    req.Month,
		Day:      req.Day,
		Hour:     req.Hour,
		Minute:   req.Minute,
		TZOffset: req.TzOffset,
		Lat:      req.Lat,
		Lng:      req.Lng,
	})
	}))

	mux.HandleFunc("/api/harmonic", handleJSON(func(req struct {
		ChartRequest
		Harmonics []int   `json:"harmonics"`
		Orb       float64 `json:"orb"`
	}) ([]byte, error) {
		if cfg.Harmonic == nil {
			return nil, ErrNotAvailable
		}
		if len(req.Harmonics) == 0 {
			req.Harmonics = []int{4, 5, 7, 9}
		}
		return cfg.Harmonic(dignity.BirthData{
		Name:     req.Name,
		Year:     req.Year,
		Month:    req.Month,
		Day:      req.Day,
		Hour:     req.Hour,
		Minute:   req.Minute,
		TZOffset: req.TzOffset,
		Lat:      req.Lat,
		Lng:      req.Lng,
	}, req.Harmonics, defaultOrb(req.Orb, OrbNarrow))
	}))

	mux.HandleFunc("/api/divisional", handleJSON(func(req ChartRequest) ([]byte, error) {
		if cfg.Divisional == nil {
			return nil, ErrNotAvailable
		}
		return cfg.Divisional(dignity.BirthData{
		Name:     req.Name,
		Year:     req.Year,
		Month:    req.Month,
		Day:      req.Day,
		Hour:     req.Hour,
		Minute:   req.Minute,
		TZOffset: req.TzOffset,
		Lat:      req.Lat,
		Lng:      req.Lng,
	})
	}))

	mux.HandleFunc("/api/parans", handleJSON(func(req ChartRequest) ([]byte, error) {
		if cfg.Parans == nil {
			return nil, ErrNotAvailable
		}
		return cfg.Parans(dignity.BirthData{
		Name:     req.Name,
		Year:     req.Year,
		Month:    req.Month,
		Day:      req.Day,
		Hour:     req.Hour,
		Minute:   req.Minute,
		TZOffset: req.TzOffset,
		Lat:      req.Lat,
		Lng:      req.Lng,
	}, defaultOrb(req.Orb, OrbNarrow))
	}))

	mux.HandleFunc("/api/declination", handleJSON(func(req ChartRequest) ([]byte, error) {
		if cfg.Declination == nil {
			return nil, ErrNotAvailable
		}
		return cfg.Declination(dignity.BirthData{
		Name:     req.Name,
		Year:     req.Year,
		Month:    req.Month,
		Day:      req.Day,
		Hour:     req.Hour,
		Minute:   req.Minute,
		TZOffset: req.TzOffset,
		Lat:      req.Lat,
		Lng:      req.Lng,
	}, defaultOrb(req.Orb, OrbTight))
	}))

	mux.HandleFunc("/api/firdaria", handleJSON(func(req ChartRequest) ([]byte, error) {
		if cfg.Firdaria == nil {
			return nil, ErrNotAvailable
		}
		return cfg.Firdaria(dignity.BirthData{
		Name:     req.Name,
		Year:     req.Year,
		Month:    req.Month,
		Day:      req.Day,
		Hour:     req.Hour,
		Minute:   req.Minute,
		TZOffset: req.TzOffset,
		Lat:      req.Lat,
		Lng:      req.Lng,
	})
	}))

	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.WriteHeader(http.StatusNoContent)
			return
		}
		http.NotFound(w, r)
	})

	return mux
}

// Run starts an HTTP server on the given port. It builds the mux via NewMux
// and calls ListenAndServe.
func Run(port int, cfg ServerConfig) error {
	mux := NewMux(cfg)

	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("Empirical server running at http://localhost%s\n", addr)
	return http.ListenAndServe(addr, mux)
}

// DraconicSolarReturnRequest is the request for /api/draconic-solar-return.
type DraconicSolarReturnRequest struct {
	ChartRequest
	TargetYear int `json:"target_year"`
}

// DirectionsRequest is the request for /api/directions.
type DirectionsRequest struct {
	ChartRequest
	Age float64 `json:"age"`
	Orb float64 `json:"orb"`
}

// AstroCartographyRequest is the request for /api/astrocartography.
type AstroCartographyRequest struct {
	ChartRequest
	LatStep float64 `json:"lat_step"`
	Frame   string  `json:"frame"`
	// Optional filters for /api/astrocartography-parans
	Planets string  `json:"planets"`  // comma-separated planet names, e.g. "Sun,Moon,Venus"
	MinLat  float64 `json:"min_lat"`  // minimum latitude for geographic filter
	MaxLat  float64 `json:"max_lat"`  // maximum latitude for geographic filter
	MinLng  float64 `json:"min_lng"`  // minimum longitude for geographic filter
	MaxLng  float64 `json:"max_lng"`  // maximum longitude for geographic filter
}

// AstroCartographyCompareRequest is the request for /api/astrocartography-compare.
type AstroCartographyCompareRequest struct {
	ChartRequest
	LatStep   float64 `json:"lat_step"`
	TargetLat float64 `json:"target_lat"`
	TargetLng float64 `json:"target_lng"`
	Orb       float64 `json:"orb"`
}

// ElectionalRequest is the request for /api/electional.
type ElectionalRequest struct {
	ChartRequest
	StartDate string  `json:"start_date"`
	EndDate   string  `json:"end_date"`
	Orb       float64 `json:"orb"`
}

// ArabicPartsRequest is the request for /api/arabic-parts.
type ArabicPartsRequest struct {
	ChartRequest
	Orb float64 `json:"orb"`
}

// defaultOrb returns orb if > 0, otherwise returns defaultVal.
func defaultOrb(orb, defaultVal float64) float64 {
	if orb <= 0 {
		return defaultVal
	}
	return orb
}

type ChartRequest struct {
	Name        string  `json:"name"`
	Year        int     `json:"year"`
	Month       int     `json:"month"`
	Day         int     `json:"day"`
	Hour        int     `json:"hour"`
	Minute      int     `json:"minute"`
	TzOffset    float64 `json:"tz_offset"`
	Lat         float64 `json:"lat"`
	Lng         float64 `json:"lng"`
	HouseSystem string  `json:"house_system"`
	Sidereal    bool    `json:"sidereal"`
	ShowAspects bool    `json:"show_aspects"`
	OuterPlanets bool   `json:"outer_planets"`
	Orb         float64 `json:"orb"`
	HighlightPatterns bool    `json:"highlight_patterns"`
	PatternOrb        float64 `json:"pattern_orb"`
	System      string  `json:"system"`
}

type TimingRequest struct {
	ChartRequest
	TargetDate string `json:"target_date"`
}

type TransitRequest struct {
	ChartRequest
	StartDate string  `json:"start_date"`
	EndDate   string  `json:"end_date"`
	Orb       float64 `json:"orb"`
	Sidereal  bool    `json:"sidereal"`
}

// TransitChartRequest is the JSON body for /api/transit-chart.
// It includes birth data + transit moment datetime + location.
type TransitChartRequest struct {
	ChartRequest
	TransitYear   int     `json:"transit_year"`
	TransitMonth  int     `json:"transit_month"`
	TransitDay    int     `json:"transit_day"`
	TransitHour   int     `json:"transit_hour"`
	TransitMinute int     `json:"transit_minute"`
	TransitTZ     float64 `json:"transit_tz"`
	TransitLat    float64 `json:"transit_lat"`
	TransitLng    float64 `json:"transit_lng"`
}

type SynastryRequest struct {
	Name1  string  `json:"name1"`
	Year1  int     `json:"year1"`
	Month1 int     `json:"month1"`
	Day1   int     `json:"day1"`
	Hour1  int     `json:"hour1"`
	Min1   int     `json:"min1"`
	Tz1    float64 `json:"tz1"`
	Lat1   float64 `json:"lat1"`
	Lng1   float64 `json:"lng1"`
	Name2  string  `json:"name2"`
	Year2  int     `json:"year2"`
	Month2 int     `json:"month2"`
	Day2   int     `json:"day2"`
	Hour2  int     `json:"hour2"`
	Min2   int     `json:"min2"`
	Tz2    float64 `json:"tz2"`
	Lat2   float64 `json:"lat2"`
	Lng2   float64 `json:"lng2"`
	Orb    float64 `json:"orb"`
}

type RelocationRequest struct {
	ChartRequest
	LocationA  LatLng `json:"location_a"`
	LocationB  LatLng `json:"location_b"`
	TargetDate string `json:"target_date"`
}

// AstroCartographyParansResponse is the JSON response for /api/astrocartography-parans.
type AstroCartographyParansResponse struct {
	Name          string                      `json:"name"`
	Frame         string                      `json:"frame"`
	Intersections []dignity.ParanIntersection `json:"intersections"`
}

// filterParans filters paran intersections by planet names and geographic bounds.
// planets: comma-separated list, e.g. "Sun,Moon,Venus". Empty means no filter.
// minLat/maxLat/minLng/maxLng: zero means no bound on that axis.
func filterParans(intersections []dignity.ParanIntersection, planets string, minLat, maxLat, minLng, maxLng float64) []dignity.ParanIntersection {
	// Build planet set
	planetSet := make(map[string]bool)
	if planets != "" {
		for _, p := range splitComma(planets) {
			planetSet[p] = true
		}
	}

	// Determine if geo filter is active
	hasGeo := minLat != 0 || maxLat != 0 || minLng != 0 || maxLng != 0

	out := make([]dignity.ParanIntersection, 0)
	for _, p := range intersections {
		// Planet filter: intersection must involve at least one of the specified planets
		if len(planetSet) > 0 && !planetSet[p.Planet1] && !planetSet[p.Planet2] {
			continue
		}
		// Geographic filter
		if hasGeo {
			if minLat != 0 && p.Lat < minLat {
				continue
			}
			if maxLat != 0 && p.Lat > maxLat {
				continue
			}
			if minLng != 0 && p.Lon < minLng {
				continue
			}
			if maxLng != 0 && p.Lon > maxLng {
				continue
			}
		}
		out = append(out, p)
	}
	return out
}

// splitComma splits a comma-separated string, trimming whitespace.
func splitComma(s string) []string {
	var parts []string
	for _, p := range strings.Split(s, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			parts = append(parts, p)
		}
	}
	return parts
}
