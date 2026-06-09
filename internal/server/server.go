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

// Run starts an HTTP server on the given port. staticFS contains the embedded
// static files (index.html, etc.). compute handles POST /api/recover.
// aspects handles GET /api/aspect-catalog.
// timing handles POST /api/timing-convergence.
func Run(port int, staticFS fs.FS, compute ComputeFunc, aspects AspectFunc, timing TimingFunc) error {
	mux := http.NewServeMux()

	// Static files
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	// API: full recovery (all phases)
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
		result, err := compute(req.Name, req.Year, req.Month, req.Day,
			req.Hour, req.Minute, req.TzOffset, req.Lat, req.Lng)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write(result)
	})

	// API: aspect catalog (Phase 2)
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

	// API: timing convergence (Phase 4)
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
		result, err := timing(req.Name, req.Year, req.Month, req.Day,
			req.Hour, req.Minute, req.TzOffset, req.Lat, req.Lng,
			req.TargetDate)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write(result)
	})

	// CORS preflight
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

	// Root: serve index.html
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

// ChartRequest is the JSON payload for /api/recover.
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

// TimingRequest is the JSON payload for /api/timing-convergence.
type TimingRequest struct {
	ChartRequest
	TargetDate string `json:"target_date"`
}
