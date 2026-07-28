package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"strings"

	"github.com/aj-nt/empirical/internal/comparison"
	"github.com/aj-nt/empirical/internal/dignity"
	"github.com/aj-nt/empirical/internal/geocode"
)

// Default orb values for endpoints that accept an optional orb parameter.
const (
	OrbTight    = 1.0 // directions, declination
	OrbNarrow   = 2.0 // stars, harmonic, parans
	OrbStandard = 3.0 // transits, draconic, interpretation, electional, composite, arabic-parts, progressed-cross
	OrbWide     = 5.0 // synastry, patterns
)

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

// SolarArcFunc computes solar arc directions for a target date.
type SolarArcFunc func(bd dignity.BirthData, targetDate string, orbDeg float64) ([]byte, error)

// ProfectionFunc computes annual profections for a target date.
type ProfectionFunc func(bd dignity.BirthData, targetDate string) ([]byte, error)

// BiWheelFunc generates a bi-wheel SVG comparing two charts.
type BiWheelFunc func(inner, outer dignity.BirthData, opts dignity.BiWheelOptions) ([]byte, error)

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

// NatalHTMLFunc renders a natal chart interpretation as HTML for a given system.
// system is one of: "koine", "western", "vedic", "bazi".
type NatalHTMLFunc func(bd dignity.BirthData, system string, orbDeg float64) (string, error)

// TransitHTMLFunc renders a transit chart interpretation as HTML for a given system.
// system is one of: "koine", "western", "vedic", "bazi".
type TransitHTMLFunc func(bd dignity.BirthData, year, month, day, hour, minute int, tzOff, lat, lng float64, system string, orbDeg float64) (string, error)

// ResearchMetricsFunc computes research metrics for a single chart.
type ResearchMetricsFunc func(bd dignity.BirthData) ([]byte, error)

// ResearchBaselineFunc generates a baseline distribution for a metric across N random charts.
type ResearchBaselineFunc func(metric string, n int, seed int64) ([]byte, error)

// BatchAnalysisFunc computes research metrics for multiple charts and returns aggregate stats.
type BatchAnalysisFunc func(charts []dignity.BirthData) ([]byte, error)

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
	SolarArc              SolarArcFunc
	Profection            ProfectionFunc
	BiWheel               BiWheelFunc
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
	NatalHTML             NatalHTMLFunc
	TransitHTML           TransitHTMLFunc
	ResearchMetrics       ResearchMetricsFunc
	ResearchBaseline      ResearchBaselineFunc
	BatchAnalysis         BatchAnalysisFunc
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

// handleHTML returns an http.HandlerFunc that decodes a JSON body of type T,
// calls fn with the decoded request, and writes the string result with the given
// content type and CORS headers. If fn is nil, responds with 501 Not Implemented.
func handleHTML[T any](fn func(T) (string, error), contentType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}
		if fn == nil {
			http.Error(w, "not available", http.StatusNotImplemented)
			return
		}
		var req T
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		result, err := fn(req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write([]byte(result))
	}
}

// NewMux builds the HTTP mux with all handlers wired to the provided functions.
// Exported so tests can exercise the real handlers with mock functions.
func NewMux(cfg ServerConfig) *http.ServeMux {
	mux := http.NewServeMux()

	if cfg.StaticFS != nil {
		fsys := http.FS(cfg.StaticFS)
		fileServer := http.FileServer(fsys)

		mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Don't intercept API calls
			if strings.HasPrefix(r.URL.Path, "/api/") {
				http.NotFound(w, r)
				return
			}
			// SPA fallback: if the file doesn't exist, serve index.html
			path := strings.TrimPrefix(r.URL.Path, "/")
			if path == "" {
				path = "index.html"
			}
			if _, err := fsys.Open(path); err != nil {
				r.URL.Path = "/index.html"
			}
			fileServer.ServeHTTP(w, r)
		}))
	}

	mux.HandleFunc("/api/recover", handleJSON(func(req ChartRequest) ([]byte, error) {
		return cfg.Compute(req.ToBirthData())
	}))

	mux.HandleFunc("/api/natal-html", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}
		if cfg.NatalHTML == nil {
			http.Error(w, "not available", http.StatusNotImplemented)
			return
		}
		var req ChartRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		result, err := cfg.NatalHTML(req.ToBirthData(), req.System, defaultOrb(req.Orb, OrbStandard))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write([]byte(result))
	})

	mux.HandleFunc("/api/transit-html", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}
		if cfg.TransitHTML == nil {
			http.Error(w, "not available", http.StatusNotImplemented)
			return
		}
		var req TransitChartRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		result, err := cfg.TransitHTML(req.ToBirthData(), req.TransitYear, req.TransitMonth, req.TransitDay, req.TransitHour, req.TransitMinute, req.TransitTZ, req.TransitLat, req.TransitLng, req.System, defaultOrb(req.Orb, OrbStandard))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write([]byte(result))
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
		return cfg.Timing(req.ToBirthData(), req.TargetDate)
	}))

	mux.HandleFunc("/api/transits", handleJSON(func(req TransitRequest) ([]byte, error) {
		if cfg.Transits == nil {
			return nil, ErrNotAvailable
		}
		return cfg.Transits(req.ToBirthData(), req.StartDate, req.EndDate, defaultOrb(req.Orb, OrbStandard), req.Sidereal)
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
		return cfg.Relocation(req.ToBirthData(),
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
		result, err := cfg.Chart(req.ToBirthData(), hs, req.Sidereal, req.ShowAspects, req.OuterPlanets, req.HighlightPatterns, req.PatternOrb)
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
		return cfg.Patterns(req.ToBirthData(), defaultOrb(req.Orb, OrbWide))
	}))

	mux.HandleFunc("/api/draconic", handleJSON(func(req ChartRequest) ([]byte, error) {
		if cfg.Draconic == nil {
			return nil, ErrNotAvailable
		}
		return cfg.Draconic(req.ToBirthData(), defaultOrb(req.Orb, OrbStandard))
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
		return cfg.DraconicTransits(req.ToBirthData(), req.StartDate, req.EndDate, defaultOrb(req.Orb, OrbStandard))
	}))

	mux.HandleFunc("/api/progressed-draconic", handleJSON(func(req TimingRequest) ([]byte, error) {
		if cfg.ProgressedDraconic == nil {
			return nil, ErrNotAvailable
		}
		return cfg.ProgressedDraconic(req.ToBirthData(), req.TargetDate)
	}))

	mux.HandleFunc("/api/draconic-solar-return", handleJSON(func(req DraconicSolarReturnRequest) ([]byte, error) {
		if cfg.DraconicSolarReturn == nil {
			return nil, ErrNotAvailable
		}
		return cfg.DraconicSolarReturn(req.ToBirthData(), req.TargetYear)
	}))

	mux.HandleFunc("/api/stars", handleJSON(func(req ChartRequest) ([]byte, error) {
		if cfg.Stars == nil {
			return nil, ErrNotAvailable
		}
		return cfg.Stars(req.ToBirthData(), defaultOrb(req.Orb, OrbNarrow))
	}))

	mux.HandleFunc("/api/draconic-transits-cross", handleJSON(func(req TransitRequest) ([]byte, error) {
		if cfg.DraconicTransitsCross == nil {
			return nil, ErrNotAvailable
		}
		return cfg.DraconicTransitsCross(req.ToBirthData(), req.StartDate, req.EndDate, defaultOrb(req.Orb, OrbStandard))
	}))

	mux.HandleFunc("/api/progressed-cross", handleJSON(func(req TimingRequest) ([]byte, error) {
		if cfg.ProgressedCross == nil {
			return nil, ErrNotAvailable
		}
		return cfg.ProgressedCross(req.ToBirthData(), req.TargetDate, defaultOrb(req.Orb, OrbStandard))
	}))

	mux.HandleFunc("/api/directions", handleJSON(func(req DirectionsRequest) ([]byte, error) {
		if cfg.Directions == nil {
			return nil, ErrNotAvailable
		}
		return cfg.Directions(req.ToBirthData(), req.Age, defaultOrb(req.Orb, OrbTight))
	}))

	mux.HandleFunc("/api/solar-arc", handleJSON(func(req SolarArcRequest) ([]byte, error) {
		if cfg.SolarArc == nil {
			return nil, ErrNotAvailable
		}
		return cfg.SolarArc(req.ToBirthData(), req.TargetDate, defaultOrb(req.Orb, OrbStandard))
	}))

	mux.HandleFunc("/api/profection", handleJSON(func(req ProfectionRequest) ([]byte, error) {
		if cfg.Profection == nil {
			return nil, ErrNotAvailable
		}
		return cfg.Profection(req.ToBirthData(), req.TargetDate)
	}))

	mux.HandleFunc("/api/bi-wheel", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}
		var req BiWheelRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		if cfg.BiWheel == nil {
			http.Error(w, "not available", http.StatusNotImplemented)
			return
		}
		opts := dignity.DefaultBiWheelOptions()
		if req.HouseSystem != "" {
			opts.HouseSystem = req.HouseSystem
		}
		if req.ShowAsteroids {
			opts.ShowAsteroids = true
		}
		if req.ShowTNPs {
			opts.ShowTNPs = true
		}
		if req.Sidereal {
			opts.Sidereal = true
		}
		if req.Orb > 0 {
			opts.Orb = req.Orb
		}
		result, err := cfg.BiWheel(req.Inner.ToBirthData(), req.Outer.ToBirthData(), opts)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "image/svg+xml")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write(result)
	})

	mux.HandleFunc("/api/interpretation", handleJSON(func(req ChartRequest) ([]byte, error) {
		if cfg.Interpretation == nil {
			return nil, ErrNotAvailable
		}
		system := req.System
		if system == "" {
			system = "koiné"
		}
		return cfg.Interpretation(req.ToBirthData(), req.HouseSystem, defaultOrb(req.Orb, OrbStandard), system)
	}))

	mux.HandleFunc("/api/astrocartography", handleJSON(func(req AstroCartographyRequest) ([]byte, error) {
		if cfg.AstroCartography == nil {
			return nil, ErrNotAvailable
		}
		return cfg.AstroCartography(req.ToBirthData(), defaultOrb(req.LatStep, 2.0), req.Frame)
	}))

	mux.HandleFunc("/api/astrocartography-compare", handleJSON(func(req AstroCartographyCompareRequest) ([]byte, error) {
		if cfg.AstroCartographyCompare == nil {
			return nil, ErrNotAvailable
		}
		return cfg.AstroCartographyCompare(req.ToBirthData(), defaultOrb(req.LatStep, 2.0), req.TargetLat, req.TargetLng, defaultOrb(req.Orb, OrbStandard))
	}))

	mux.HandleFunc("/api/astrocartography-parans", handleJSON(func(req AstroCartographyRequest) ([]byte, error) {
		if cfg.AstroCartographyParans == nil {
			return nil, ErrNotAvailable
		}
		raw, err := cfg.AstroCartographyParans(req.ToBirthData(), defaultOrb(req.LatStep, 2.0), req.Frame)
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
		return cfg.Electional(req.ToBirthData(), req.StartDate, req.EndDate, defaultOrb(req.Orb, OrbStandard))
	}))

	mux.HandleFunc("/api/mansion-convergence", handleJSON(func(req ChartRequest) ([]byte, error) {
		if cfg.MansionConvergence == nil {
			return nil, ErrNotAvailable
		}
		return cfg.MansionConvergence(req.ToBirthData())
	}))

	mux.HandleFunc("/api/arabic-parts", handleJSON(func(req ArabicPartsRequest) ([]byte, error) {
		if cfg.ArabicParts == nil {
			return nil, ErrNotAvailable
		}
		return cfg.ArabicParts(req.ToBirthData(), defaultOrb(req.Orb, OrbStandard))
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
		return cfg.SolarReturn(req.ToBirthData(), req.TargetYear)
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
		return cfg.StarsCross(req.ToBirthData(), defaultOrb(req.Orb, OrbNarrow))
	}))

	mux.HandleFunc("/api/traditional", handleJSON(func(req ChartRequest) ([]byte, error) {
		if cfg.Traditional == nil {
			return nil, ErrNotAvailable
		}
		return cfg.Traditional(req.ToBirthData())
	}))

	mux.HandleFunc("/api/uranian", handleJSON(func(req ChartRequest) ([]byte, error) {
		if cfg.Uranian == nil {
			return nil, ErrNotAvailable
		}
		return cfg.Uranian(req.ToBirthData())
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
		return cfg.Harmonic(req.ToBirthData(), req.Harmonics, defaultOrb(req.Orb, OrbNarrow))
	}))

	mux.HandleFunc("/api/divisional", handleJSON(func(req ChartRequest) ([]byte, error) {
		if cfg.Divisional == nil {
			return nil, ErrNotAvailable
		}
		return cfg.Divisional(req.ToBirthData())
	}))

	mux.HandleFunc("/api/parans", handleJSON(func(req ChartRequest) ([]byte, error) {
		if cfg.Parans == nil {
			return nil, ErrNotAvailable
		}
		return cfg.Parans(req.ToBirthData(), defaultOrb(req.Orb, OrbNarrow))
	}))

	mux.HandleFunc("/api/declination", handleJSON(func(req ChartRequest) ([]byte, error) {
		if cfg.Declination == nil {
			return nil, ErrNotAvailable
		}
		return cfg.Declination(req.ToBirthData(), defaultOrb(req.Orb, OrbTight))
	}))

	mux.HandleFunc("/api/firdaria", handleJSON(func(req ChartRequest) ([]byte, error) {
		if cfg.Firdaria == nil {
			return nil, ErrNotAvailable
		}
		return cfg.Firdaria(req.ToBirthData())
	}))

	// ── New parameterized endpoints (Phase C) ──────────────────────────

	// POST /api/base-chart — compute a BaseChart from birth data.
	// Returns the full physics-only chart: positions, houses, angles, nodes, stars, declinations.
	mux.HandleFunc("/api/base-chart", handleJSON(func(req ChartRequest) ([]byte, error) {
		if cfg.Compute == nil {
			return nil, ErrNotAvailable
		}
		return cfg.Compute(req.ToBirthData())
	}))

	// POST /api/system — apply a system transform to a BaseChart.
	// Body: {name, year, month, day, hour, minute, tz_offset, lat, lng, system, orb}
	// system: "koine", "western", "vedic", "bazi", "draconic"
	mux.HandleFunc("/api/system", handleJSON(func(req ChartRequest) ([]byte, error) {
		if cfg.Interpretation == nil {
			return nil, ErrNotAvailable
		}
		system := req.System
		if system == "" {
			system = "koine"
		}
		return cfg.Interpretation(req.ToBirthData(), req.HouseSystem, defaultOrb(req.Orb, OrbStandard), system)
	}))

	// POST /api/compare — cross-system comparison.
	// Body: {name, year, month, day, hour, minute, tz_offset, lat, lng, orb}
	// Runs Koiné, Western, and Vedic transforms and diffs the outputs.
	mux.HandleFunc("/api/compare", handleJSON(func(req ChartRequest) ([]byte, error) {
		if cfg.Compute == nil {
			return nil, ErrNotAvailable
		}
		// Compute the base chart, then compare systems
		bc, err := dignity.ComputeBaseChart(req.ToBirthData())
		if err != nil {
			return nil, err
		}
		report := comparison.CompareSystems(bc, defaultOrb(req.Orb, OrbWide))
		return report.JSON()
	}))

	// GET /api/geocode/search?q=... — search cities by name
	// Returns up to 20 matching cities with lat, lon, and estimated timezone offset.
	mux.HandleFunc("/api/geocode/search", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "GET only", http.StatusMethodNotAllowed)
			return
		}
		q := r.URL.Query().Get("q")
		if q == "" {
			http.Error(w, "missing ?q= parameter", http.StatusBadRequest)
			return
		}
		cities, err := geocode.SearchCities(q, 20)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// Enrich with estimated timezone offset
		type CityResult struct {
			Name    string  `json:"name"`
			Country string  `json:"country"`
			Lat     float64 `json:"lat"`
			Lon     float64 `json:"lon"`
			TZOffset float64 `json:"tz_offset"`
		}
		results := make([]CityResult, len(cities))
		for i, c := range cities {
			results[i] = CityResult{
				Name:    c.Name,
				Country: c.Country,
				Lat:     c.Lat,
				Lon:     c.Lon,
				TZOffset: geocode.EstimateTZOffset(c.Lon),
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		json.NewEncoder(w).Encode(results)
	})

	// POST /api/research-metrics — compute all research metrics for a chart.
	mux.HandleFunc("/api/research-metrics", handleJSON(func(req ChartRequest) ([]byte, error) {
		if cfg.ResearchMetrics == nil {
			return nil, ErrNotAvailable
		}
		return cfg.ResearchMetrics(req.ToBirthData())
	}))

	// POST /api/research-baseline — generate baseline distribution for a metric.
	mux.HandleFunc("/api/research-baseline", handleJSON(func(req struct {
		Metric string `json:"metric"`
		N      int    `json:"n"`
		Seed   int64  `json:"seed"`
	}) ([]byte, error) {
		if cfg.ResearchBaseline == nil {
			return nil, ErrNotAvailable
		}
		if req.N == 0 {
			req.N = 1000
		}
		return cfg.ResearchBaseline(req.Metric, req.N, req.Seed)
	}))

	// POST /api/batch-analysis — compute research metrics for multiple charts.
	mux.HandleFunc("/api/batch-analysis", handleJSON(func(req struct {
		Charts []dignity.BirthData `json:"charts"`
	}) ([]byte, error) {
		if cfg.BatchAnalysis == nil {
			return nil, ErrNotAvailable
		}
		if len(req.Charts) == 0 {
			return nil, fmt.Errorf("at least one chart required")
		}
		return cfg.BatchAnalysis(req.Charts)
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

// SolarArcRequest is the request for /api/solar-arc.
type SolarArcRequest struct {
	ChartRequest
	TargetDate string  `json:"target_date"`
	Orb        float64 `json:"orb"`
}

// ProfectionRequest is the request for /api/profection.
type ProfectionRequest struct {
	ChartRequest
	TargetDate string `json:"target_date"`
}

// BiWheelRequest is the request for /api/bi-wheel.
type BiWheelRequest struct {
	Inner         ChartRequest `json:"inner"`
	Outer         ChartRequest `json:"outer"`
	HouseSystem   string       `json:"house_system"`
	ShowAsteroids bool         `json:"show_asteroids"`
	ShowTNPs      bool         `json:"show_tnps"`
	Sidereal      bool         `json:"sidereal"`
	Orb           float64      `json:"orb"`
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

// ToBirthData converts a ChartRequest to a dignity.BirthData.
func (req ChartRequest) ToBirthData() dignity.BirthData {
	return dignity.BirthData{
		Name:     req.Name,
		Year:     req.Year,
		Month:    req.Month,
		Day:      req.Day,
		Hour:     req.Hour,
		Minute:   req.Minute,
		TZOffset: req.TzOffset,
		Lat:      req.Lat,
		Lng:      req.Lng,
	}
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
		for _, p := range strings.Split(planets, ",") {
			p = strings.TrimSpace(p)
			if p != "" {
				planetSet[p] = true
			}
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


