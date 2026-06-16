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

// NewMux builds the HTTP mux with all handlers wired to the provided functions.
// Exported so tests can exercise the real handlers with mock functions.
func NewMux(staticFS fs.FS, compute ComputeFunc, aspects AspectFunc, timing TimingFunc, transits TransitFunc, synastry SynastryFunc, relocation RelocationFunc, chart ChartFunc, patterns PatternFunc, draconic DraconicFunc, draconicSynastry DraconicSynastryFunc, draconicSynastryFull DraconicSynastryFullFunc, draconicTransits DraconicTransitFunc, progressedDraconic ProgressedDraconicFunc, draconicSolarReturn DraconicSolarReturnFunc, stars StarsFunc, draconicTransitsCross DraconicTransitsCrossFunc, progressedCross ProgressedCrossFunc, directions DirectionsFunc, interpretation InterpretationFunc, astroCartography AstroCartographyFunc, astroCartographyCompare AstroCartographyCompareFunc, electional ElectionalFunc, mansionConvergence MansionConvergenceFunc, arabicParts ArabicPartsFunc, solarReturn SolarReturnFunc, composite CompositeFunc, starsCross StarsCrossFunc, traditional TraditionalFunc, uranian UranianFunc, harmonic HarmonicFunc, divisional DivisionalFunc, parans ParansFunc, declination DeclinationFunc, firdaria FirdariaFunc) *http.ServeMux {
	mux := http.NewServeMux()

	if staticFS != nil {
		mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))
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
		result, err := compute(req.Name, req.Year, req.Month, req.Day, req.Hour, req.Minute, req.TzOffset, req.Lat, req.Lng)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write(result)
	})

	mux.HandleFunc("/api/aspect-catalog", func(w http.ResponseWriter, r *http.Request) {
		if aspects == nil {
			http.Error(w, "not available", http.StatusNotImplemented)
			return
		}
		result, err := aspects()
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
		if timing == nil {
			http.Error(w, "not available", http.StatusNotImplemented)
			return
		}
		result, err := timing(req.Name, req.Year, req.Month, req.Day, req.Hour, req.Minute, req.TzOffset, req.Lat, req.Lng, req.TargetDate)
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
		if transits == nil {
			http.Error(w, "not available", http.StatusNotImplemented)
			return
		}
		orb := req.Orb
		if orb <= 0 {
			orb = 3.0
		}
		result, err := transits(req.Name, req.Year, req.Month, req.Day, req.Hour, req.Minute, req.TzOffset, req.Lat, req.Lng, req.StartDate, req.EndDate, orb, req.Sidereal)
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
		if synastry == nil {
			http.Error(w, "not available", http.StatusNotImplemented)
			return
		}
		orb := req.Orb
		if orb <= 0 {
			orb = 5.0
		}
		result, err := synastry(req.Name1, req.Year1, req.Month1, req.Day1, req.Hour1, req.Min1, req.Tz1, req.Lat1, req.Lng1,
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
		if relocation == nil {
			http.Error(w, "not available", http.StatusNotImplemented)
			return
		}
		result, err := relocation(req.Name, req.Year, req.Month, req.Day, req.Hour, req.Minute, req.TzOffset, req.Lat, req.Lng,
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
		if chart == nil {
			http.Error(w, "not available", http.StatusNotImplemented)
			return
		}
		hs := req.HouseSystem
		if hs == "" {
			hs = "placidus"
		}
		result, err := chart(req.Name, req.Year, req.Month, req.Day, req.Hour, req.Minute, req.TzOffset, req.Lat, req.Lng, hs, req.Sidereal, req.ShowAspects, req.OuterPlanets, req.HighlightPatterns, req.PatternOrb)
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
		if patterns == nil {
			http.Error(w, "not available", http.StatusNotImplemented)
			return
		}
		orb := req.Orb
		if orb <= 0 {
			orb = 5.0
		}
		result, err := patterns(req.Name, req.Year, req.Month, req.Day, req.Hour, req.Minute, req.TzOffset, req.Lat, req.Lng, orb)
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
		if draconic == nil {
			http.Error(w, "not available", http.StatusNotImplemented)
			return
		}
		orb := req.Orb
		if orb <= 0 {
			orb = 3.0
		}
		result, err := draconic(req.Name, req.Year, req.Month, req.Day, req.Hour, req.Minute, req.TzOffset, req.Lat, req.Lng, orb)
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
		if draconicSynastry == nil {
			http.Error(w, "not available", http.StatusNotImplemented)
			return
		}
		orb := req.Orb
		if orb <= 0 {
			orb = 3.0
		}
		result, err := draconicSynastry(req.Name1, req.Year1, req.Month1, req.Day1, req.Hour1, req.Min1, req.Tz1, req.Lat1, req.Lng1,
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
		if draconicSynastryFull == nil {
			http.Error(w, "not available", http.StatusNotImplemented)
			return
		}
		orb := req.Orb
		if orb <= 0 {
			orb = 3.0
		}
		result, err := draconicSynastryFull(req.Name1, req.Year1, req.Month1, req.Day1, req.Hour1, req.Min1, req.Tz1, req.Lat1, req.Lng1,
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
		if draconicTransits == nil {
			http.Error(w, "not available", http.StatusNotImplemented)
			return
		}
		orb := req.Orb
		if orb <= 0 {
			orb = 3.0
		}
		result, err := draconicTransits(req.Name, req.Year, req.Month, req.Day, req.Hour, req.Minute, req.TzOffset, req.Lat, req.Lng, req.StartDate, req.EndDate, orb)
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
		if progressedDraconic == nil {
			http.Error(w, "not available", http.StatusNotImplemented)
			return
		}
		result, err := progressedDraconic(req.Name, req.Year, req.Month, req.Day, req.Hour, req.Minute, req.TzOffset, req.Lat, req.Lng, req.TargetDate)
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
		if draconicSolarReturn == nil {
			http.Error(w, "not available", http.StatusNotImplemented)
			return
		}
		result, err := draconicSolarReturn(req.Name, req.Year, req.Month, req.Day, req.Hour, req.Minute, req.TzOffset, req.Lat, req.Lng, req.TargetYear)
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
		if stars == nil {
			http.Error(w, "not available", http.StatusNotImplemented)
			return
		}
		orb := req.Orb
		if orb <= 0 {
			orb = 2.0
		}
		result, err := stars(req.Name, req.Year, req.Month, req.Day, req.Hour, req.Minute, req.TzOffset, req.Lat, req.Lng, orb)
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
		if draconicTransitsCross == nil {
			http.Error(w, "not available", http.StatusNotImplemented)
			return
		}
		orb := req.Orb
		if orb <= 0 {
			orb = 3.0
		}
		result, err := draconicTransitsCross(req.Name, req.Year, req.Month, req.Day, req.Hour, req.Minute, req.TzOffset, req.Lat, req.Lng, req.StartDate, req.EndDate, orb)
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
		if progressedCross == nil {
			http.Error(w, "not available", http.StatusNotImplemented)
			return
		}
		orb := req.Orb
		if orb <= 0 {
			orb = 3.0
		}
		result, err := progressedCross(req.Name, req.Year, req.Month, req.Day, req.Hour, req.Minute, req.TzOffset, req.Lat, req.Lng, req.TargetDate, orb)
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
		if directions == nil {
			http.Error(w, "not available", http.StatusNotImplemented)
			return
		}
		orb := req.Orb
		if orb <= 0 {
			orb = 3.0
		}
		result, err := directions(req.Name, req.Year, req.Month, req.Day, req.Hour, req.Minute, req.TzOffset, req.Lat, req.Lng, req.Age, orb)
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
		if interpretation == nil {
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
		result, err := interpretation(req.Name, req.Year, req.Month, req.Day, req.Hour, req.Minute, req.TzOffset, req.Lat, req.Lng, hs, orb)
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
		if astroCartography == nil {
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
		result, err := astroCartography(req.Name, req.Year, req.Month, req.Day, req.Hour, req.Minute, req.TzOffset, req.Lat, req.Lng, ls, frame)
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
		if astroCartographyCompare == nil {
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
		result, err := astroCartographyCompare(req.Name, req.Year, req.Month, req.Day, req.Hour, req.Minute, req.TzOffset, req.Lat, req.Lng, ls, req.TargetLat, req.TargetLng, orb)
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
		if electional == nil {
			http.Error(w, "not available", http.StatusNotImplemented)
			return
		}
		orb := req.Orb
		if orb <= 0 {
			orb = 3.0
		}
		result, err := electional(req.Name, req.Year, req.Month, req.Day, req.Hour, req.Minute, req.TzOffset, req.Lat, req.Lng, req.StartDate, req.EndDate, orb)
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
		if mansionConvergence == nil {
			http.Error(w, "not available", http.StatusNotImplemented)
			return
		}
		result, err := mansionConvergence(req.Name, req.Year, req.Month, req.Day, req.Hour, req.Minute, req.TzOffset, req.Lat, req.Lng)
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
		if arabicParts == nil {
			http.Error(w, "not available", http.StatusNotImplemented)
			return
		}
		orb := req.Orb
		if orb <= 0 {
			orb = 3.0
		}
		result, err := arabicParts(req.Name, req.Year, req.Month, req.Day, req.Hour, req.Minute, req.TzOffset, req.Lat, req.Lng, orb)
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
		if solarReturn == nil {
			http.Error(w, "not available", http.StatusNotImplemented)
			return
		}
		if req.TargetYear == 0 {
			req.TargetYear = req.Year + 1
		}
		result, err := solarReturn(req.Name, req.Year, req.Month, req.Day, req.Hour, req.Minute, req.TzOffset, req.Lat, req.Lng, req.TargetYear)
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
		if composite == nil {
			http.Error(w, "not available", http.StatusNotImplemented)
			return
		}
		orb := req.Orb
		if orb <= 0 {
			orb = 3.0
		}
		result, err := composite(req.Name1, req.Year1, req.Month1, req.Day1, req.Hour1, req.Min1, req.Tz1, req.Lat1, req.Lng1, req.Name2, req.Year2, req.Month2, req.Day2, req.Hour2, req.Min2, req.Tz2, req.Lat2, req.Lng2, orb)
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
		if starsCross == nil {
			http.Error(w, "not available", http.StatusNotImplemented)
			return
		}
		orb := req.Orb
		if orb <= 0 {
			orb = 2.0
		}
		result, err := starsCross(req.Name, req.Year, req.Month, req.Day, req.Hour, req.Minute, req.TzOffset, req.Lat, req.Lng, orb)
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
		if traditional == nil {
			http.Error(w, "not available", http.StatusNotImplemented)
			return
		}
		var req ChartRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		result, err := traditional(req.Name, req.Year, req.Month, req.Day, req.Hour, req.Minute, req.TzOffset, req.Lat, req.Lng)
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
		if uranian == nil {
			http.Error(w, "not available", http.StatusNotImplemented)
			return
		}
		var req ChartRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		result, err := uranian(req.Name, req.Year, req.Month, req.Day, req.Hour, req.Minute, req.TzOffset, req.Lat, req.Lng)
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
		if harmonic == nil {
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
		result, err := harmonic(req.Name, req.Year, req.Month, req.Day, req.Hour, req.Minute, req.TzOffset, req.Lat, req.Lng, req.Harmonics, orb)
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
		if divisional == nil {
			http.Error(w, "not available", http.StatusNotImplemented)
			return
		}
		var req ChartRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		result, err := divisional(req.Name, req.Year, req.Month, req.Day, req.Hour, req.Minute, req.TzOffset, req.Lat, req.Lng)
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
		if parans == nil {
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
		result, err := parans(req.Name, req.Year, req.Month, req.Day, req.Hour, req.Minute, req.TzOffset, req.Lat, req.Lng, orb)
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
		if declination == nil {
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
		result, err := declination(req.Name, req.Year, req.Month, req.Day, req.Hour, req.Minute, req.TzOffset, req.Lat, req.Lng, orb)
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
		if firdaria == nil {
			http.Error(w, "not available", http.StatusNotImplemented)
			return
		}
		var req ChartRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}
		result, err := firdaria(req.Name, req.Year, req.Month, req.Day, req.Hour, req.Minute, req.TzOffset, req.Lat, req.Lng)
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
func Run(port int, staticFS fs.FS, compute ComputeFunc, aspects AspectFunc, timing TimingFunc, transits TransitFunc, synastry SynastryFunc, relocation RelocationFunc, chart ChartFunc, patterns PatternFunc, draconic DraconicFunc, draconicSynastry DraconicSynastryFunc, draconicSynastryFull DraconicSynastryFullFunc, draconicTransits DraconicTransitFunc, progressedDraconic ProgressedDraconicFunc, draconicSolarReturn DraconicSolarReturnFunc, stars StarsFunc, draconicTransitsCross DraconicTransitsCrossFunc, progressedCross ProgressedCrossFunc, directions DirectionsFunc, interpretation InterpretationFunc, astroCartography AstroCartographyFunc, astroCartographyCompare AstroCartographyCompareFunc, electional ElectionalFunc, mansionConvergence MansionConvergenceFunc, arabicParts ArabicPartsFunc, solarReturn SolarReturnFunc, composite CompositeFunc, starsCross StarsCrossFunc, traditional TraditionalFunc, uranian UranianFunc, harmonic HarmonicFunc, divisional DivisionalFunc, parans ParansFunc, declination DeclinationFunc, firdaria FirdariaFunc) error {
	mux := NewMux(staticFS, compute, aspects, timing, transits, synastry, relocation, chart, patterns, draconic, draconicSynastry, draconicSynastryFull, draconicTransits, progressedDraconic, draconicSolarReturn, stars, draconicTransitsCross, progressedCross, directions, interpretation, astroCartography, astroCartographyCompare, electional, mansionConvergence, arabicParts, solarReturn, composite, starsCross, traditional, uranian, harmonic, divisional, parans, declination, firdaria)

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
