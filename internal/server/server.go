package server

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
)

// ComputeFunc computes a full multi-phase recovery report for birth data
// and returns the result as JSON bytes.
type ComputeFunc func(name string, year, month, day, hour, minute int, tzOffset, lat, lng float64) ([]byte, error)

// AspectFunc returns the aspect catalog as JSON bytes.
type AspectFunc func() ([]byte, error)

// TimingFunc computes timing layer convergence for a target date.
type TimingFunc func(name string, year, month, day, hour, minute int, tzOff, lat, lng float64, targetDate string) ([]byte, error)

// TransitFunc computes transits for a natal chart over a date range.
type TransitFunc func(name string, year, month, day, hour, minute int, tzOff, lat, lng float64, startDate, endDate string, orbDeg float64, sidereal bool) ([]byte, error)

// SynastryFunc computes inter-aspects between two natal charts.
type SynastryFunc func(name1 string, y1, mo1, d1, h1, mi1 int, tz1, la1, lo1 float64, name2 string, y2, mo2, d2, h2, mi2 int, tz2, la2, lo2 float64, orbDeg float64) ([]byte, error)

// RelocationFunc computes a cross-validated relocation comparison between two locations.
type RelocationFunc func(name string, year, month, day, hour, minute int, tzOff, lat, lng float64, locA LatLng, locB LatLng, targetDate string) ([]byte, error)

// ChartFunc renders a natal chart wheel as SVG.
type ChartFunc func(name string, year, month, day, hour, minute int, tzOff, lat, lng float64, houseSystem string, sidereal bool, showAspects bool, outerPlanets bool, highlightPatterns bool, patternOrb float64) (string, error)

// PatternFunc detects geometric patterns in a natal chart.
type PatternFunc func(name string, year, month, day, hour, minute int, tzOff, lat, lng float64, orbDeg float64) ([]byte, error)

// DraconicFunc computes the draconic chart, sign shifts, and bridges.
type DraconicFunc func(name string, year, month, day, hour, minute int, tzOff, lat, lng float64, orbDeg float64) ([]byte, error)

// DraconicSynastryFunc computes draconic synastry between two charts.
type DraconicSynastryFunc func(name1 string, y1, mo1, d1, h1, mi1 int, tz1, la1, lo1 float64, name2 string, y2, mo2, d2, h2, mi2 int, tz2, la2, lo2 float64, orbDeg float64) ([]byte, error)

// DraconicSynastryFullFunc computes the full three-layer draconic synastry.
type DraconicSynastryFullFunc func(name1 string, y1, mo1, d1, h1, mi1 int, tz1, la1, lo1 float64, name2 string, y2, mo2, d2, h2, mi2 int, tz2, la2, lo2 float64, orbDeg float64) ([]byte, error)

// DraconicTransitFunc computes draconic transits: transiting planets → draconic chart.
type DraconicTransitFunc func(name string, year, month, day, hour, minute int, tzOff, lat, lng float64, startDate, endDate string, orbDeg float64) ([]byte, error)

// ProgressedDraconicFunc computes the progressed draconic chart using the current transiting NN.
type ProgressedDraconicFunc func(name string, year, month, day, hour, minute int, tzOff, lat, lng float64, targetDate string) ([]byte, error)

// DraconicSolarReturnFunc computes the draconic solar return for a target year.
type DraconicSolarReturnFunc func(name string, year, month, day, hour, minute int, tzOff, lat, lng float64, targetYear int) ([]byte, error)

// StarsFunc computes fixed star conjunctions for a natal chart.
type StarsFunc func(name string, year, month, day, hour, minute int, tzOff, lat, lng float64, orbDeg float64) ([]byte, error)

// DraconicTransitsCrossFunc compares draconic transits in tropical vs sidereal.
type DraconicTransitsCrossFunc func(name string, year, month, day, hour, minute int, tzOff, lat, lng float64, startDate, endDate string, orbDeg float64) ([]byte, error)

// ProgressedCrossFunc compares progressed-to-natal aspects in tropical vs sidereal.
type ProgressedCrossFunc func(name string, year, month, day, hour, minute int, tzOff, lat, lng float64, targetDate string, orbDeg float64) ([]byte, error)

// DirectionsFunc computes primary directions (Ptolemy) for a given age.
type DirectionsFunc func(name string, year, month, day, hour, minute int, tzOff, lat, lng float64, age float64, orbDeg float64) ([]byte, error)

// InterpretationFunc produces natural-language chart interpretation.
type InterpretationFunc func(name string, year, month, day, hour, minute int, tzOff, lat, lng float64, houseSystem string, orbDeg float64) ([]byte, error)

// AstroCartographyFunc computes planetary lines for a world map.
// frame: "tropical", "draconic", or "cross".
type AstroCartographyFunc func(name string, year, month, day, hour, minute int, tzOff, lat, lng float64, latStep float64, frame string) ([]byte, error)

// AstroCartographyCompareFunc returns three-way comparison at a target location.
type AstroCartographyCompareFunc func(name string, year, month, day, hour, minute int, tzOff, lat, lng float64, latStep, targetLat, targetLng, orb float64) ([]byte, error)

// ElectionalFunc scores dates in a range for launch/event timing.
type ElectionalFunc func(name string, year, month, day, hour, minute int, tzOff, lat, lng float64, startDate, endDate string, orbDeg float64) ([]byte, error)

// MansionConvergenceFunc computes nakshatra/xiu mansion placements per chart.
type MansionConvergenceFunc func(name string, year, month, day, hour, minute int, tzOff, lat, lng float64) ([]byte, error)

// ArabicPartsFunc computes Arabic Parts and cross-system comparison.
type ArabicPartsFunc func(name string, year, month, day, hour, minute int, tzOff, lat, lng float64, orbDeg float64) ([]byte, error)

// SolarReturnFunc computes a tropical solar return chart.
type SolarReturnFunc func(name string, year, month, day, hour, minute int, tzOff, lat, lng float64, targetYear int) ([]byte, error)

// CompositeFunc computes a midpoint composite chart for two people.
type CompositeFunc func(name1 string, y1, m1, d1, h1, min1 int, tz1, lat1, lng1 float64, name2 string, y2, m2, d2, h2, min2 int, tz2, lat2, lng2 float64, orbDeg float64) ([]byte, error)

// StarsCrossFunc compares star conjunctions in tropical vs sidereal frames.
type StarsCrossFunc func(name string, year, month, day, hour, minute int, tzOff, lat, lng float64, orbDeg float64) ([]byte, error)

// TraditionalFunc computes traditional astrology interpretive data.
type TraditionalFunc func(name string, year, month, day, hour, minute int, tzOff, lat, lng float64) ([]byte, error)

// UranianFunc computes Uranian/Hamburg School midpoint analysis.
type UranianFunc func(name string, year, month, day, hour, minute int, tzOff, lat, lng float64) ([]byte, error)

// HarmonicFunc computes Addey-style harmonic charts.
type HarmonicFunc func(name string, year, month, day, hour, minute int, tzOff, lat, lng float64, harmonics []int, orb float64) ([]byte, error)

// DivisionalFunc computes Vedic divisional charts (navamsha D9, nakshatras, dasha).
type DivisionalFunc func(name string, year, month, day, hour, minute int, tzOff, lat, lng float64) ([]byte, error)

// ParansFunc computes fixed star parans (star-planet angle contacts).
type ParansFunc func(name string, year, month, day, hour, minute int, tzOff, lat, lng float64, orb float64) ([]byte, error)

// DeclinationFunc computes declination parallels and contraparallels.
type DeclinationFunc func(name string, year, month, day, hour, minute int, tzOff, lat, lng float64, orb float64) ([]byte, error)

// FirdariaFunc computes Persian firdaria planetary periods.
type FirdariaFunc func(name string, year, month, day, hour, minute int, tzOff, lat, lng float64) ([]byte, error)

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
	Interpretation        InterpretationFunc
	AstroCartography      AstroCartographyFunc
	AstroCartographyCompare AstroCartographyCompareFunc
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
}

// NewMux builds the HTTP mux with all handlers wired to the provided functions.
// Exported so tests can exercise the real handlers with mock functions.
func NewMux(cfg ServerConfig) *http.ServeMux {
	mux := http.NewServeMux()

	if cfg.StaticFS != nil {
		mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(cfg.StaticFS))))
	}

	mux.HandleFunc("/api/recover", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}
		var req ChartRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		result, err := cfg.Compute(req.Name, req.Year, req.Month, req.Day, req.Hour, req.Minute, req.TzOffset, req.Lat, req.Lng)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
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

	mux.HandleFunc("/api/timing-convergence", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}
		var req TimingRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		if cfg.Timing == nil {
			http.Error(w, "not available", http.StatusNotImplemented)
			return
		}
		result, err := cfg.Timing(req.Name, req.Year, req.Month, req.Day, req.Hour, req.Minute, req.TzOffset, req.Lat, req.Lng, req.TargetDate)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write(result)
	})

	mux.HandleFunc("/api/transits", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}
		var req TransitRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		if cfg.Transits == nil {
			http.Error(w, "not available", http.StatusNotImplemented)
			return
		}
		orb := req.Orb
		if orb <= 0 {
			orb = 3.0
		}
		result, err := cfg.Transits(req.Name, req.Year, req.Month, req.Day, req.Hour, req.Minute, req.TzOffset, req.Lat, req.Lng, req.StartDate, req.EndDate, orb, req.Sidereal)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write(result)
	})

	mux.HandleFunc("/api/synastry", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}
		var req SynastryRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		if cfg.Synastry == nil {
			http.Error(w, "not available", http.StatusNotImplemented)
			return
		}
		orb := req.Orb
		if orb <= 0 {
			orb = 5.0
		}
		result, err := cfg.Synastry(req.Name1, req.Year1, req.Month1, req.Day1, req.Hour1, req.Min1, req.Tz1, req.Lat1, req.Lng1,
			req.Name2, req.Year2, req.Month2, req.Day2, req.Hour2, req.Min2, req.Tz2, req.Lat2, req.Lng2, orb)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write(result)
	})

	mux.HandleFunc("/api/relocation-compare", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}
		var req RelocationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		if cfg.Relocation == nil {
			http.Error(w, "not available", http.StatusNotImplemented)
			return
		}
		result, err := cfg.Relocation(req.Name, req.Year, req.Month, req.Day, req.Hour, req.Minute, req.TzOffset, req.Lat, req.Lng,
			req.LocationA, req.LocationB, req.TargetDate)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write(result)
	})

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
		result, err := cfg.Chart(req.Name, req.Year, req.Month, req.Day, req.Hour, req.Minute, req.TzOffset, req.Lat, req.Lng, hs, req.Sidereal, req.ShowAspects, req.OuterPlanets, req.HighlightPatterns, req.PatternOrb)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "image/svg+xml")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write([]byte(result))
	})

	mux.HandleFunc("/api/patterns", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}
		var req ChartRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		if cfg.Patterns == nil {
			http.Error(w, "not available", http.StatusNotImplemented)
			return
		}
		orb := req.Orb
		if orb <= 0 {
			orb = 5.0
		}
		result, err := cfg.Patterns(req.Name, req.Year, req.Month, req.Day, req.Hour, req.Minute, req.TzOffset, req.Lat, req.Lng, orb)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write(result)
	})

	mux.HandleFunc("/api/draconic", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}
		var req ChartRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		if cfg.Draconic == nil {
			http.Error(w, "not available", http.StatusNotImplemented)
			return
		}
		orb := req.Orb
		if orb <= 0 {
			orb = 3.0
		}
		result, err := cfg.Draconic(req.Name, req.Year, req.Month, req.Day, req.Hour, req.Minute, req.TzOffset, req.Lat, req.Lng, orb)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write(result)
	})

	mux.HandleFunc("/api/draconic-synastry", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}
		var req SynastryRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		if cfg.DraconicSynastry == nil {
			http.Error(w, "not available", http.StatusNotImplemented)
			return
		}
		orb := req.Orb
		if orb <= 0 {
			orb = 3.0
		}
		result, err := cfg.DraconicSynastry(req.Name1, req.Year1, req.Month1, req.Day1, req.Hour1, req.Min1, req.Tz1, req.Lat1, req.Lng1,
			req.Name2, req.Year2, req.Month2, req.Day2, req.Hour2, req.Min2, req.Tz2, req.Lat2, req.Lng2, orb)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write(result)
	})

	mux.HandleFunc("/api/draconic-synastry-full", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}
		var req SynastryRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		if cfg.DraconicSynastryFull == nil {
			http.Error(w, "not available", http.StatusNotImplemented)
			return
		}
		orb := req.Orb
		if orb <= 0 {
			orb = 3.0
		}
		result, err := cfg.DraconicSynastryFull(req.Name1, req.Year1, req.Month1, req.Day1, req.Hour1, req.Min1, req.Tz1, req.Lat1, req.Lng1,
			req.Name2, req.Year2, req.Month2, req.Day2, req.Hour2, req.Min2, req.Tz2, req.Lat2, req.Lng2, orb)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write(result)
	})

	mux.HandleFunc("/api/draconic-transits", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}
		var req TransitRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		if cfg.DraconicTransits == nil {
			http.Error(w, "not available", http.StatusNotImplemented)
			return
		}
		orb := req.Orb
		if orb <= 0 {
			orb = 3.0
		}
		result, err := cfg.DraconicTransits(req.Name, req.Year, req.Month, req.Day, req.Hour, req.Minute, req.TzOffset, req.Lat, req.Lng, req.StartDate, req.EndDate, orb)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write(result)
	})

	mux.HandleFunc("/api/progressed-draconic", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}
		var req TimingRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		if cfg.ProgressedDraconic == nil {
			http.Error(w, "not available", http.StatusNotImplemented)
			return
		}
		result, err := cfg.ProgressedDraconic(req.Name, req.Year, req.Month, req.Day, req.Hour, req.Minute, req.TzOffset, req.Lat, req.Lng, req.TargetDate)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write(result)
	})

	mux.HandleFunc("/api/draconic-solar-return", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			ChartRequest
			TargetYear int `json:"target_year"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		if cfg.DraconicSolarReturn == nil {
			http.Error(w, "not available", http.StatusNotImplemented)
			return
		}
		result, err := cfg.DraconicSolarReturn(req.Name, req.Year, req.Month, req.Day, req.Hour, req.Minute, req.TzOffset, req.Lat, req.Lng, req.TargetYear)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write(result)
	})

	mux.HandleFunc("/api/stars", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}
		var req ChartRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		if cfg.Stars == nil {
			http.Error(w, "not available", http.StatusNotImplemented)
			return
		}
		orb := req.Orb
		if orb <= 0 {
			orb = 2.0
		}
		result, err := cfg.Stars(req.Name, req.Year, req.Month, req.Day, req.Hour, req.Minute, req.TzOffset, req.Lat, req.Lng, orb)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write(result)
	})

	mux.HandleFunc("/api/draconic-transits-cross", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			ChartRequest
			StartDate string  `json:"start_date"`
			EndDate   string  `json:"end_date"`
			Orb       float64 `json:"orb"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		if cfg.DraconicTransitsCross == nil {
			http.Error(w, "not available", http.StatusNotImplemented)
			return
		}
		orb := req.Orb
		if orb <= 0 {
			orb = 3.0
		}
		result, err := cfg.DraconicTransitsCross(req.Name, req.Year, req.Month, req.Day, req.Hour, req.Minute, req.TzOffset, req.Lat, req.Lng, req.StartDate, req.EndDate, orb)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write(result)
	})

	mux.HandleFunc("/api/progressed-cross", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			ChartRequest
			TargetDate string  `json:"target_date"`
			Orb        float64 `json:"orb"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		if cfg.ProgressedCross == nil {
			http.Error(w, "not available", http.StatusNotImplemented)
			return
		}
		orb := req.Orb
		if orb <= 0 {
			orb = 3.0
		}
		result, err := cfg.ProgressedCross(req.Name, req.Year, req.Month, req.Day, req.Hour, req.Minute, req.TzOffset, req.Lat, req.Lng, req.TargetDate, orb)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write(result)
	})

	mux.HandleFunc("/api/directions", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			ChartRequest
			Age float64 `json:"age"`
			Orb float64 `json:"orb"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		if cfg.Directions == nil {
			http.Error(w, "not available", http.StatusNotImplemented)
			return
		}
		orb := req.Orb
		if orb <= 0 {
			orb = 3.0
		}
		result, err := cfg.Directions(req.Name, req.Year, req.Month, req.Day, req.Hour, req.Minute, req.TzOffset, req.Lat, req.Lng, req.Age, orb)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write(result)
	})

	mux.HandleFunc("/api/interpretation", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			ChartRequest
			HouseSystem string  `json:"house_system"`
			Orb         float64 `json:"orb"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		if cfg.Interpretation == nil {
			http.Error(w, "not available", http.StatusNotImplemented)
			return
		}
		orb := req.Orb
		if orb <= 0 {
			orb = 3.0
		}
		hs := req.HouseSystem
		if hs == "" {
			hs = "P"
		}
		result, err := cfg.Interpretation(req.Name, req.Year, req.Month, req.Day, req.Hour, req.Minute, req.TzOffset, req.Lat, req.Lng, hs, orb)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write(result)
	})

	mux.HandleFunc("/api/astrocartography", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			ChartRequest
			LatStep float64 `json:"lat_step"`
			Frame   string  `json:"frame"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		if cfg.AstroCartography == nil {
			http.Error(w, "not available", http.StatusNotImplemented)
			return
		}
		ls := req.LatStep
		if ls <= 0 {
			ls = 2.0
		}
		frame := req.Frame
		if frame == "" {
			frame = "tropical"
		}
		result, err := cfg.AstroCartography(req.Name, req.Year, req.Month, req.Day, req.Hour, req.Minute, req.TzOffset, req.Lat, req.Lng, ls, frame)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write(result)
	})

	mux.HandleFunc("/api/astrocartography-compare", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			ChartRequest
			LatStep   float64 `json:"lat_step"`
			TargetLat float64 `json:"target_lat"`
			TargetLng float64 `json:"target_lng"`
			Orb       float64 `json:"orb"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		if cfg.AstroCartographyCompare == nil {
			http.Error(w, "not available", http.StatusNotImplemented)
			return
		}
		ls := req.LatStep
		if ls <= 0 {
			ls = 2.0
		}
		orb := req.Orb
		if orb <= 0 {
			orb = 3.0
		}
		result, err := cfg.AstroCartographyCompare(req.Name, req.Year, req.Month, req.Day, req.Hour, req.Minute, req.TzOffset, req.Lat, req.Lng, ls, req.TargetLat, req.TargetLng, orb)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write(result)
	})

	mux.HandleFunc("/api/electional", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			ChartRequest
			StartDate string  `json:"start_date"`
			EndDate   string  `json:"end_date"`
			Orb       float64 `json:"orb"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		if cfg.Electional == nil {
			http.Error(w, "not available", http.StatusNotImplemented)
			return
		}
		orb := req.Orb
		if orb <= 0 {
			orb = 3.0
		}
		result, err := cfg.Electional(req.Name, req.Year, req.Month, req.Day, req.Hour, req.Minute, req.TzOffset, req.Lat, req.Lng, req.StartDate, req.EndDate, orb)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write(result)
	})

	mux.HandleFunc("/api/mansion-convergence", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}
		var req ChartRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		if cfg.MansionConvergence == nil {
			http.Error(w, "not available", http.StatusNotImplemented)
			return
		}
		result, err := cfg.MansionConvergence(req.Name, req.Year, req.Month, req.Day, req.Hour, req.Minute, req.TzOffset, req.Lat, req.Lng)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write(result)
	})

	mux.HandleFunc("/api/arabic-parts", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			ChartRequest
			Orb float64 `json:"orb"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		if cfg.ArabicParts == nil {
			http.Error(w, "not available", http.StatusNotImplemented)
			return
		}
		orb := req.Orb
		if orb <= 0 {
			orb = 3.0
		}
		result, err := cfg.ArabicParts(req.Name, req.Year, req.Month, req.Day, req.Hour, req.Minute, req.TzOffset, req.Lat, req.Lng, orb)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write(result)
	})

	mux.HandleFunc("/api/solar-return", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			ChartRequest
			TargetYear int `json:"target_year"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		if cfg.SolarReturn == nil {
			http.Error(w, "not available", http.StatusNotImplemented)
			return
		}
		if req.TargetYear == 0 {
			req.TargetYear = req.Year + 1
		}
		result, err := cfg.SolarReturn(req.Name, req.Year, req.Month, req.Day, req.Hour, req.Minute, req.TzOffset, req.Lat, req.Lng, req.TargetYear)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write(result)
	})

	mux.HandleFunc("/api/composite", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}
		var req struct {
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
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		if cfg.Composite == nil {
			http.Error(w, "not available", http.StatusNotImplemented)
			return
		}
		orb := req.Orb
		if orb <= 0 {
			orb = 3.0
		}
		result, err := cfg.Composite(req.Name1, req.Year1, req.Month1, req.Day1, req.Hour1, req.Min1, req.Tz1, req.Lat1, req.Lng1, req.Name2, req.Year2, req.Month2, req.Day2, req.Hour2, req.Min2, req.Tz2, req.Lat2, req.Lng2, orb)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write(result)
	})

	mux.HandleFunc("/api/stars-cross", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			ChartRequest
			Orb float64 `json:"orb"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		if cfg.StarsCross == nil {
			http.Error(w, "not available", http.StatusNotImplemented)
			return
		}
		orb := req.Orb
		if orb <= 0 {
			orb = 2.0
		}
		result, err := cfg.StarsCross(req.Name, req.Year, req.Month, req.Day, req.Hour, req.Minute, req.TzOffset, req.Lat, req.Lng, orb)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write(result)
	})

	mux.HandleFunc("/api/traditional", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}
		if cfg.Traditional == nil {
			http.Error(w, "not available", http.StatusNotImplemented)
			return
		}
		var req ChartRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		result, err := cfg.Traditional(req.Name, req.Year, req.Month, req.Day, req.Hour, req.Minute, req.TzOffset, req.Lat, req.Lng)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write(result)
	})

	mux.HandleFunc("/api/uranian", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}
		if cfg.Uranian == nil {
			http.Error(w, "not available", http.StatusNotImplemented)
			return
		}
		var req ChartRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		result, err := cfg.Uranian(req.Name, req.Year, req.Month, req.Day, req.Hour, req.Minute, req.TzOffset, req.Lat, req.Lng)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write(result)
	})

	mux.HandleFunc("/api/harmonic", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}
		if cfg.Harmonic == nil {
			http.Error(w, "not available", http.StatusNotImplemented)
			return
		}
		var req struct {
			ChartRequest
			Harmonics []int   `json:"harmonics"`
			Orb       float64 `json:"orb"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		if len(req.Harmonics) == 0 {
			req.Harmonics = []int{4, 5, 7, 9}
		}
		orb := req.Orb
		if orb <= 0 {
			orb = 2.0
		}
		result, err := cfg.Harmonic(req.Name, req.Year, req.Month, req.Day, req.Hour, req.Minute, req.TzOffset, req.Lat, req.Lng, req.Harmonics, orb)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write(result)
	})

	mux.HandleFunc("/api/divisional", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}
		if cfg.Divisional == nil {
			http.Error(w, "not available", http.StatusNotImplemented)
			return
		}
		var req ChartRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		result, err := cfg.Divisional(req.Name, req.Year, req.Month, req.Day, req.Hour, req.Minute, req.TzOffset, req.Lat, req.Lng)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write(result)
	})

	mux.HandleFunc("/api/parans", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}
		if cfg.Parans == nil {
			http.Error(w, "not available", http.StatusNotImplemented)
			return
		}
		var req ChartRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		orb := req.Orb
		if orb <= 0 {
			orb = 2.0
		}
		result, err := cfg.Parans(req.Name, req.Year, req.Month, req.Day, req.Hour, req.Minute, req.TzOffset, req.Lat, req.Lng, orb)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write(result)
	})

	mux.HandleFunc("/api/declination", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}
		if cfg.Declination == nil {
			http.Error(w, "not available", http.StatusNotImplemented)
			return
		}
		var req ChartRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		orb := req.Orb
		if orb <= 0 {
			orb = 1.0
		}
		result, err := cfg.Declination(req.Name, req.Year, req.Month, req.Day, req.Hour, req.Minute, req.TzOffset, req.Lat, req.Lng, orb)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write(result)
	})

	mux.HandleFunc("/api/firdaria", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}
		if cfg.Firdaria == nil {
			http.Error(w, "not available", http.StatusNotImplemented)
			return
		}
		var req ChartRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		result, err := cfg.Firdaria(req.Name, req.Year, req.Month, req.Day, req.Hour, req.Minute, req.TzOffset, req.Lat, req.Lng)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write(result)
	})

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
