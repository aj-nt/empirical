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
type TransitFunc func(name string, year, month, day, hour, minute int, tzOff, lat, lng float64, startDate, endDate string, orbDeg float64) ([]byte, error)

// SynastryFunc computes inter-aspects between two natal charts.
type SynastryFunc func(name1 string, y1, mo1, d1, h1, mi1 int, tz1, la1, lo1 float64, name2 string, y2, mo2, d2, h2, mi2 int, tz2, la2, lo2 float64, orbDeg float64) ([]byte, error)

// Run starts an HTTP server on the given port.
func Run(port int, staticFS fs.FS, compute ComputeFunc, aspects AspectFunc, timing TimingFunc, transits TransitFunc, synastry SynastryFunc) error {
	mux := http.NewServeMux()

	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

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
		result, err := transits(req.Name, req.Year, req.Month, req.Day, req.Hour, req.Minute, req.TzOffset, req.Lat, req.Lng, req.StartDate, req.EndDate, orb)
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

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" || r.URL.Path == "/index.html" {
			b, err := fs.ReadFile(staticFS, "index.html")
			if err != nil {
				http.Error(w, "index.html not found", http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write(b)
			return
		}
		http.NotFound(w, r)
	})

	addr := fmt.Sprintf(":%d", port)
	fmt.Printf("Empirical server running at http://localhost%s\n", addr)
	return http.ListenAndServe(addr, mux)
}

type ChartRequest struct {
	Name     string  `json:"name"`
	Year     int     `json:"year"`
	Month    int     `json:"month"`
	Day      int     `json:"day"`
	Hour     int     `json:"hour"`
	Minute   int     `json:"minute"`
	TzOffset float64 `json:"tz_offset"`
	Lat      float64 `json:"lat"`
	Lng      float64 `json:"lng"`
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
