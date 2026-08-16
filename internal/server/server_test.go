package server

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aj-nt/empirical/internal/dignity"
)

// ── Mock functions ────────────────────────────────────────────────────────

func mockCompute(bd dignity.BirthData) ([]byte, error) {
	return json.Marshal(map[string]interface{}{"name": bd.Name, "year": bd.Year})
}

func mockAspects() ([]byte, error) {
	return json.Marshal([]map[string]interface{}{{"name": "Conjunction", "angle": 0}})
}

func mockTiming(bd dignity.BirthData, targetDate string) ([]byte, error) {
	return json.Marshal(map[string]interface{}{"name": bd.Name, "target": targetDate})
}

func mockTransits(bd dignity.BirthData, startDate, endDate string, orb float64, sidereal bool, ayanamsa string) ([]byte, error) {
	return json.Marshal(map[string]interface{}{"name": bd.Name, "start": startDate, "end": endDate})
}

func mockSynastry(name1 string, y1, mo1, d1, h1, mi1 int, tz1, la1, lo1 float64, name2 string, y2, mo2, d2, h2, mi2 int, tz2, la2, lo2 float64, orb float64) ([]byte, error) {
	return json.Marshal(map[string]interface{}{"name1": name1, "name2": name2})
}

func mockRelocation(bd dignity.BirthData, locA LatLng, locB LatLng, targetDate string) ([]byte, error) {
	return json.Marshal(map[string]interface{}{"name": bd.Name, "locA": locA.Name, "locB": locB.Name})
}

func mockChart(bd dignity.BirthData, houseSystem string, sidereal bool, ayanamsa string, showAspects bool, outerPlanets bool, highlightPatterns bool, patternOrb float64) (string, error) {
	return "<svg>mock chart</svg>", nil
}

func mockPatterns(bd dignity.BirthData, orb float64) ([]byte, error) {
	return json.Marshal(map[string]interface{}{"name": bd.Name})
}

func mockDraconic(bd dignity.BirthData, orb float64) ([]byte, error) {
	return json.Marshal(map[string]interface{}{"name": bd.Name})
}

func mockDraconicSynastry(name1 string, y1, mo1, d1, h1, mi1 int, tz1, la1, lo1 float64, name2 string, y2, mo2, d2, h2, mi2 int, tz2, la2, lo2 float64, orb float64) ([]byte, error) {
	return json.Marshal(map[string]interface{}{"name1": name1, "name2": name2})
}

func mockDraconicSynastryFull(name1 string, y1, mo1, d1, h1, mi1 int, tz1, la1, lo1 float64, name2 string, y2, mo2, d2, h2, mi2 int, tz2, la2, lo2 float64, orb float64) ([]byte, error) {
	return json.Marshal(map[string]interface{}{"name1": name1, "name2": name2})
}

func mockDraconicTransits(bd dignity.BirthData, startDate, endDate string, orb float64) ([]byte, error) {
	return json.Marshal(map[string]interface{}{"name": bd.Name, "start": startDate, "end": endDate})
}

func mockProgressedDraconic(bd dignity.BirthData, targetDate string) ([]byte, error) {
	return json.Marshal(map[string]interface{}{"name": bd.Name, "target": targetDate})
}

func mockDraconicSolarReturn(bd dignity.BirthData, targetYear int) ([]byte, error) {
	return json.Marshal(map[string]interface{}{"name": bd.Name, "targetYear": targetYear})
}

func mockStars(bd dignity.BirthData, orb float64) ([]byte, error) {
	return json.Marshal(map[string]interface{}{"name": bd.Name})
}

func mockDraconicTransitsCross(bd dignity.BirthData, startDate, endDate string, orb float64) ([]byte, error) {
	return json.Marshal(map[string]interface{}{"name": bd.Name, "start": startDate, "end": endDate})
}

func mockProgressedCross(bd dignity.BirthData, targetDate string, orb float64) ([]byte, error) {
	return json.Marshal(map[string]interface{}{"name": bd.Name, "target": targetDate})
}

func mockDirections(bd dignity.BirthData, age float64, orb float64) ([]byte, error) {
	return json.Marshal(map[string]interface{}{"name": bd.Name, "age": age})
}

func mockInterpretation(bd dignity.BirthData, houseSystem string, orb float64, system string) ([]byte, error) {
	return json.Marshal(map[string]interface{}{"name": bd.Name, "system": system})
}

func mockAstroCartography(bd dignity.BirthData, latStep float64, frame string) ([]byte, error) {
	return json.Marshal(map[string]interface{}{"name": bd.Name, "frame": frame})
}

func mockAstroCartographyCompare(bd dignity.BirthData, latStep, targetLat, targetLng, orb float64) ([]byte, error) {
	return json.Marshal(map[string]interface{}{"name": bd.Name, "targetLat": targetLat, "targetLng": targetLng})
}

func mockElectional(bd dignity.BirthData, startDate, endDate string, orb float64) ([]byte, error) {
	return json.Marshal(map[string]interface{}{"name": bd.Name, "start": startDate, "end": endDate})
}

func mockMansionConvergence(bd dignity.BirthData) ([]byte, error) {
	return json.Marshal(map[string]interface{}{"name": bd.Name})
}

func mockArabicParts(bd dignity.BirthData, orb float64) ([]byte, error) {
	return json.Marshal(map[string]interface{}{"name": bd.Name})
}

func mockSolarReturn(bd dignity.BirthData, targetYear int) ([]byte, error) {
	return json.Marshal(map[string]interface{}{"name": bd.Name, "targetYear": targetYear})
}

func mockComposite(name1 string, y1, m1, d1, h1, min1 int, tz1, lat1, lng1 float64, name2 string, y2, m2, d2, h2, min2 int, tz2, lat2, lng2 float64, orb float64) ([]byte, error) {
	return json.Marshal(map[string]interface{}{"name1": name1, "name2": name2})
}

func mockStarsCross(bd dignity.BirthData, orb float64) ([]byte, error) {
	return json.Marshal(map[string]interface{}{"name": bd.Name})
}

func mockTraditional(bd dignity.BirthData) ([]byte, error) {
	return json.Marshal(map[string]interface{}{"name": bd.Name})
}

func mockUranian(bd dignity.BirthData) ([]byte, error) {
	return json.Marshal(map[string]interface{}{"name": bd.Name})
}

func mockHarmonic(bd dignity.BirthData, harmonics []int, orb float64) ([]byte, error) {
	return json.Marshal(map[string]interface{}{"name": bd.Name})
}

func mockDivisional(bd dignity.BirthData) ([]byte, error) {
	return json.Marshal(map[string]interface{}{"name": bd.Name})
}

func mockParans(bd dignity.BirthData, orb float64) ([]byte, error) {
	return json.Marshal(map[string]interface{}{"name": bd.Name})
}

func mockDeclination(bd dignity.BirthData, orb float64) ([]byte, error) {
	return json.Marshal(map[string]interface{}{"name": bd.Name})
}

func mockFirdaria(bd dignity.BirthData) ([]byte, error) {
	return json.Marshal(map[string]interface{}{"name": bd.Name})
}

// testConfig returns a ServerConfig with all mock functions populated.
func testConfig() ServerConfig {
	return ServerConfig{
		Compute:               mockCompute,
		Aspects:               mockAspects,
		Timing:                mockTiming,
		Transits:              mockTransits,
		Synastry:              mockSynastry,
		Relocation:            mockRelocation,
		Chart:                 mockChart,
		Patterns:              mockPatterns,
		Draconic:              mockDraconic,
		DraconicSynastry:      mockDraconicSynastry,
		DraconicSynastryFull:  mockDraconicSynastryFull,
		DraconicTransits:      mockDraconicTransits,
		ProgressedDraconic:    mockProgressedDraconic,
		DraconicSolarReturn:   mockDraconicSolarReturn,
		Stars:                 mockStars,
		DraconicTransitsCross: mockDraconicTransitsCross,
		ProgressedCross:       mockProgressedCross,
		Directions:            mockDirections,
		Interpretation:        mockInterpretation,
		AstroCartography:      mockAstroCartography,
		AstroCartographyCompare: mockAstroCartographyCompare,
		Electional:            mockElectional,
		MansionConvergence:    mockMansionConvergence,
		ArabicParts:           mockArabicParts,
		SolarReturn:           mockSolarReturn,
		Composite:             mockComposite,
		StarsCross:            mockStarsCross,
		Traditional:           mockTraditional,
		Uranian:               mockUranian,
		Harmonic:              mockHarmonic,
		Divisional:            mockDivisional,
		Parans:                mockParans,
		Declination:           mockDeclination,
		Firdaria:              mockFirdaria,
	}
}

// ── Test helpers ─────────────────────────────────────────────────────────

func postJSON(t *testing.T, ts *httptest.Server, path string, body interface{}) *http.Response {
	t.Helper()
	b, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	resp, err := http.Post(ts.URL+path, "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatalf("POST %s: %v", path, err)
	}
	return resp
}

func getJSON(t *testing.T, ts *httptest.Server, path string) *http.Response {
	t.Helper()
	resp, err := http.Get(ts.URL + path)
	if err != nil {
		t.Fatalf("GET %s: %v", path, err)
	}
	return resp
}

// ── Server tests using NewMux with mock functions ────────────────────────

func TestServerRecover(t *testing.T) {
	mux := NewMux(testConfig())
	ts := httptest.NewServer(mux)
	defer ts.Close()

	req := ChartRequest{
		Name: "test", Year: 2000, Month: 1, Day: 1, Hour: 12, Minute: 0,
		TzOffset: 0, Lat: 51.5, Lng: -0.1,
	}
	resp := postJSON(t, ts, "/api/recover", req)
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if !strings.Contains(string(body), "test") {
		t.Errorf("response should contain name: %s", string(body))
	}
}

func TestServerRecoverMethodNotAllowed(t *testing.T) {
	mux := NewMux(testConfig())
	ts := httptest.NewServer(mux)
	defer ts.Close()

	resp := getJSON(t, ts, "/api/recover")
	if resp.StatusCode != 405 {
		t.Errorf("expected 405, got %d", resp.StatusCode)
	}
}

func TestServerRecoverBadJSON(t *testing.T) {
	mux := NewMux(testConfig())
	ts := httptest.NewServer(mux)
	defer ts.Close()

	resp, _ := http.Post(ts.URL+"/api/recover", "application/json", bytes.NewReader([]byte("not json")))
	if resp.StatusCode != 400 {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func TestServerAspectCatalog(t *testing.T) {
	mux := NewMux(testConfig())
	ts := httptest.NewServer(mux)
	defer ts.Close()

	resp := getJSON(t, ts, "/api/aspect-catalog")
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if !strings.Contains(string(body), "Conjunction") {
		t.Errorf("response should contain Conjunction: %s", string(body))
	}
}

func TestServerTimingConvergence(t *testing.T) {
	mux := NewMux(testConfig())
	ts := httptest.NewServer(mux)
	defer ts.Close()

	req := TimingRequest{
		ChartRequest: ChartRequest{Name: "test", Year: 2000, Month: 1, Day: 1, Hour: 12, Minute: 0, TzOffset: 0, Lat: 51.5, Lng: -0.1},
		TargetDate:   "2026-06-15",
	}
	resp := postJSON(t, ts, "/api/timing-convergence", req)
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestServerTransits(t *testing.T) {
	mux := NewMux(testConfig())
	ts := httptest.NewServer(mux)
	defer ts.Close()

	req := TransitRequest{
		ChartRequest: ChartRequest{Name: "test", Year: 2000, Month: 1, Day: 1, Hour: 12, Minute: 0, TzOffset: 0, Lat: 51.5, Lng: -0.1},
		StartDate:    "2026-01-01",
		EndDate:      "2026-12-31",
	}
	resp := postJSON(t, ts, "/api/transits", req)
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestServerTransitsDefaultOrb(t *testing.T) {
	mux := NewMux(testConfig())
	ts := httptest.NewServer(mux)
	defer ts.Close()

	req := TransitRequest{
		ChartRequest: ChartRequest{Name: "test", Year: 2000, Month: 1, Day: 1, Hour: 12, Minute: 0, TzOffset: 0, Lat: 51.5, Lng: -0.1},
		StartDate:    "2026-01-01",
		EndDate:      "2026-12-31",
		Orb:          0, // should default to 3.0
	}
	resp := postJSON(t, ts, "/api/transits", req)
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestServerSynastry(t *testing.T) {
	mux := NewMux(testConfig())
	ts := httptest.NewServer(mux)
	defer ts.Close()

	req := SynastryRequest{
		Name1: "A", Year1: 2000, Month1: 1, Day1: 1, Hour1: 12, Min1: 0, Tz1: 0, Lat1: 51.5, Lng1: -0.1,
		Name2: "B", Year2: 2000, Month2: 6, Day2: 15, Hour2: 12, Min2: 0, Tz2: 0, Lat2: 40.7, Lng2: -74.0,
	}
	resp := postJSON(t, ts, "/api/synastry", req)
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestServerChart(t *testing.T) {
	mux := NewMux(testConfig())
	ts := httptest.NewServer(mux)
	defer ts.Close()

	req := ChartRequest{
		Name: "test", Year: 2000, Month: 1, Day: 1, Hour: 12, Minute: 0,
		TzOffset: 0, Lat: 51.5, Lng: -0.1,
	}
	resp := postJSON(t, ts, "/api/chart", req)
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	if !strings.Contains(string(body), "<svg>") {
		t.Errorf("response should be SVG: %s", string(body))
	}
}

func TestServerChartDefaultHouseSystem(t *testing.T) {
	mux := NewMux(testConfig())
	ts := httptest.NewServer(mux)
	defer ts.Close()

	req := ChartRequest{
		Name: "test", Year: 2000, Month: 1, Day: 1, Hour: 12, Minute: 0,
		TzOffset: 0, Lat: 51.5, Lng: -0.1, HouseSystem: "",
	}
	resp := postJSON(t, ts, "/api/chart", req)
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestServerPatterns(t *testing.T) {
	mux := NewMux(testConfig())
	ts := httptest.NewServer(mux)
	defer ts.Close()

	req := ChartRequest{
		Name: "test", Year: 2000, Month: 1, Day: 1, Hour: 12, Minute: 0,
		TzOffset: 0, Lat: 51.5, Lng: -0.1,
	}
	resp := postJSON(t, ts, "/api/patterns", req)
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestServerDraconic(t *testing.T) {
	mux := NewMux(testConfig())
	ts := httptest.NewServer(mux)
	defer ts.Close()

	req := ChartRequest{
		Name: "test", Year: 2000, Month: 1, Day: 1, Hour: 12, Minute: 0,
		TzOffset: 0, Lat: 51.5, Lng: -0.1,
	}
	resp := postJSON(t, ts, "/api/draconic", req)
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestServerStars(t *testing.T) {
	mux := NewMux(testConfig())
	ts := httptest.NewServer(mux)
	defer ts.Close()

	req := ChartRequest{
		Name: "test", Year: 2000, Month: 1, Day: 1, Hour: 12, Minute: 0,
		TzOffset: 0, Lat: 51.5, Lng: -0.1,
	}
	resp := postJSON(t, ts, "/api/stars", req)
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestServerTraditional(t *testing.T) {
	mux := NewMux(testConfig())
	ts := httptest.NewServer(mux)
	defer ts.Close()

	req := ChartRequest{
		Name: "test", Year: 2000, Month: 1, Day: 1, Hour: 12, Minute: 0,
		TzOffset: 0, Lat: 51.5, Lng: -0.1,
	}
	resp := postJSON(t, ts, "/api/traditional", req)
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestServerUranian(t *testing.T) {
	mux := NewMux(testConfig())
	ts := httptest.NewServer(mux)
	defer ts.Close()

	req := ChartRequest{
		Name: "test", Year: 2000, Month: 1, Day: 1, Hour: 12, Minute: 0,
		TzOffset: 0, Lat: 51.5, Lng: -0.1,
	}
	resp := postJSON(t, ts, "/api/uranian", req)
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestServerHarmonic(t *testing.T) {
	mux := NewMux(testConfig())
	ts := httptest.NewServer(mux)
	defer ts.Close()

	body := `{"name":"test","year":2000,"month":1,"day":1,"hour":12,"minute":0,"tz_offset":0,"lat":51.5,"lng":-0.1,"harmonics":[5,7],"orb":1.0}`
	resp, _ := http.Post(ts.URL+"/api/harmonic", "application/json", strings.NewReader(body))
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestServerDivisional(t *testing.T) {
	mux := NewMux(testConfig())
	ts := httptest.NewServer(mux)
	defer ts.Close()

	req := ChartRequest{
		Name: "test", Year: 2000, Month: 1, Day: 1, Hour: 12, Minute: 0,
		TzOffset: 0, Lat: 51.5, Lng: -0.1,
	}
	resp := postJSON(t, ts, "/api/divisional", req)
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestServerParans(t *testing.T) {
	mux := NewMux(testConfig())
	ts := httptest.NewServer(mux)
	defer ts.Close()

	req := ChartRequest{
		Name: "test", Year: 2000, Month: 1, Day: 1, Hour: 12, Minute: 0,
		TzOffset: 0, Lat: 51.5, Lng: -0.1,
	}
	resp := postJSON(t, ts, "/api/parans", req)
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestServerDeclination(t *testing.T) {
	mux := NewMux(testConfig())
	ts := httptest.NewServer(mux)
	defer ts.Close()

	req := ChartRequest{
		Name: "test", Year: 2000, Month: 1, Day: 1, Hour: 12, Minute: 0,
		TzOffset: 0, Lat: 51.5, Lng: -0.1,
	}
	resp := postJSON(t, ts, "/api/declination", req)
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestServerFirdaria(t *testing.T) {
	mux := NewMux(testConfig())
	ts := httptest.NewServer(mux)
	defer ts.Close()

	req := ChartRequest{
		Name: "test", Year: 2000, Month: 1, Day: 1, Hour: 12, Minute: 0,
		TzOffset: 0, Lat: 51.5, Lng: -0.1,
	}
	resp := postJSON(t, ts, "/api/firdaria", req)
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestServerRelocation(t *testing.T) {
	mux := NewMux(testConfig())
	ts := httptest.NewServer(mux)
	defer ts.Close()

	req := RelocationRequest{
		ChartRequest: ChartRequest{Name: "test", Year: 2000, Month: 1, Day: 1, Hour: 12, Minute: 0, TzOffset: 0, Lat: 51.5, Lng: -0.1},
		LocationA:   LatLng{Name: "London", Lat: 51.5, Lng: -0.1},
		LocationB:   LatLng{Name: "NYC", Lat: 40.7, Lng: -74.0},
		TargetDate:  "2026-06-15",
	}
	resp := postJSON(t, ts, "/api/relocation-compare", req)
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestServerDraconicSynastry(t *testing.T) {
	mux := NewMux(testConfig())
	ts := httptest.NewServer(mux)
	defer ts.Close()

	req := SynastryRequest{
		Name1: "A", Year1: 2000, Month1: 1, Day1: 1, Hour1: 12, Min1: 0, Tz1: 0, Lat1: 51.5, Lng1: -0.1,
		Name2: "B", Year2: 2000, Month2: 6, Day2: 15, Hour2: 12, Min2: 0, Tz2: 0, Lat2: 40.7, Lng2: -74.0,
	}
	resp := postJSON(t, ts, "/api/draconic-synastry", req)
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestServerDraconicSynastryFull(t *testing.T) {
	mux := NewMux(testConfig())
	ts := httptest.NewServer(mux)
	defer ts.Close()

	req := SynastryRequest{
		Name1: "A", Year1: 2000, Month1: 1, Day1: 1, Hour1: 12, Min1: 0, Tz1: 0, Lat1: 51.5, Lng1: -0.1,
		Name2: "B", Year2: 2000, Month2: 6, Day2: 15, Hour2: 12, Min2: 0, Tz2: 0, Lat2: 40.7, Lng2: -74.0,
	}
	resp := postJSON(t, ts, "/api/draconic-synastry-full", req)
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestServerDraconicTransits(t *testing.T) {
	mux := NewMux(testConfig())
	ts := httptest.NewServer(mux)
	defer ts.Close()

	req := TransitRequest{
		ChartRequest: ChartRequest{Name: "test", Year: 2000, Month: 1, Day: 1, Hour: 12, Minute: 0, TzOffset: 0, Lat: 51.5, Lng: -0.1},
		StartDate:    "2026-01-01",
		EndDate:      "2026-12-31",
	}
	resp := postJSON(t, ts, "/api/draconic-transits", req)
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestServerProgressedDraconic(t *testing.T) {
	mux := NewMux(testConfig())
	ts := httptest.NewServer(mux)
	defer ts.Close()

	req := TimingRequest{
		ChartRequest: ChartRequest{Name: "test", Year: 2000, Month: 1, Day: 1, Hour: 12, Minute: 0, TzOffset: 0, Lat: 51.5, Lng: -0.1},
		TargetDate:   "2026-06-15",
	}
	resp := postJSON(t, ts, "/api/progressed-draconic", req)
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestServerDraconicSolarReturn(t *testing.T) {
	mux := NewMux(testConfig())
	ts := httptest.NewServer(mux)
	defer ts.Close()

	body := `{"name":"test","year":2000,"month":1,"day":1,"hour":12,"minute":0,"tz_offset":0,"lat":51.5,"lng":-0.1,"target_year":2026}`
	resp, _ := http.Post(ts.URL+"/api/draconic-solar-return", "application/json", strings.NewReader(body))
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestServerDraconicTransitsCross(t *testing.T) {
	mux := NewMux(testConfig())
	ts := httptest.NewServer(mux)
	defer ts.Close()

	body := `{"name":"test","year":2000,"month":1,"day":1,"hour":12,"minute":0,"tz_offset":0,"lat":51.5,"lng":-0.1,"start_date":"2026-01-01","end_date":"2026-12-31","orb":3.0}`
	resp, _ := http.Post(ts.URL+"/api/draconic-transits-cross", "application/json", strings.NewReader(body))
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestServerMansionConvergence(t *testing.T) {
	mux := NewMux(testConfig())
	ts := httptest.NewServer(mux)
	defer ts.Close()

	req := ChartRequest{
		Name: "test", Year: 2000, Month: 1, Day: 1, Hour: 12, Minute: 0,
		TzOffset: 0, Lat: 51.5, Lng: -0.1,
	}
	resp := postJSON(t, ts, "/api/mansion-convergence", req)
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestServerArabicParts(t *testing.T) {
	mux := NewMux(testConfig())
	ts := httptest.NewServer(mux)
	defer ts.Close()

	req := ChartRequest{
		Name: "test", Year: 2000, Month: 1, Day: 1, Hour: 12, Minute: 0,
		TzOffset: 0, Lat: 51.5, Lng: -0.1,
	}
	resp := postJSON(t, ts, "/api/arabic-parts", req)
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestServerSolarReturn(t *testing.T) {
	mux := NewMux(testConfig())
	ts := httptest.NewServer(mux)
	defer ts.Close()

	body := `{"name":"test","year":2000,"month":1,"day":1,"hour":12,"minute":0,"tz_offset":0,"lat":51.5,"lng":-0.1,"target_year":2026}`
	resp, _ := http.Post(ts.URL+"/api/solar-return", "application/json", strings.NewReader(body))
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestServerComposite(t *testing.T) {
	mux := NewMux(testConfig())
	ts := httptest.NewServer(mux)
	defer ts.Close()

	req := SynastryRequest{
		Name1: "A", Year1: 2000, Month1: 1, Day1: 1, Hour1: 12, Min1: 0, Tz1: 0, Lat1: 51.5, Lng1: -0.1,
		Name2: "B", Year2: 2000, Month2: 6, Day2: 15, Hour2: 12, Min2: 0, Tz2: 0, Lat2: 40.7, Lng2: -74.0,
	}
	resp := postJSON(t, ts, "/api/composite", req)
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestServerStarsCross(t *testing.T) {
	mux := NewMux(testConfig())
	ts := httptest.NewServer(mux)
	defer ts.Close()

	req := ChartRequest{
		Name: "test", Year: 2000, Month: 1, Day: 1, Hour: 12, Minute: 0,
		TzOffset: 0, Lat: 51.5, Lng: -0.1,
	}
	resp := postJSON(t, ts, "/api/stars-cross", req)
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestServerCORS(t *testing.T) {
	mux := NewMux(testConfig())
	ts := httptest.NewServer(mux)
	defer ts.Close()

	req := ChartRequest{
		Name: "test", Year: 2000, Month: 1, Day: 1, Hour: 12, Minute: 0,
		TzOffset: 0, Lat: 51.5, Lng: -0.1,
	}
	resp := postJSON(t, ts, "/api/recover", req)
	if resp.Header.Get("Access-Control-Allow-Origin") != "*" {
		t.Error("CORS header missing")
	}
}

func TestServerNilFunctionReturns501(t *testing.T) {
	// When a function is nil, the handler should return 501
	cfg := testConfig()
	cfg.Aspects = nil
	mux := NewMux(cfg)
	ts := httptest.NewServer(mux)
	defer ts.Close()

	resp := getJSON(t, ts, "/api/aspect-catalog")
	if resp.StatusCode != 501 {
		t.Errorf("expected 501 for nil function, got %d", resp.StatusCode)
	}
}

func TestServerOptions(t *testing.T) {
	mux := NewMux(testConfig())
	ts := httptest.NewServer(mux)
	defer ts.Close()

	// OPTIONS to a path without a specific handler hits the /api/ catch-all
	req, _ := http.NewRequest("OPTIONS", ts.URL+"/api/unknown", nil)
	resp, _ := http.DefaultClient.Do(req)
	if resp.StatusCode != 204 {
		t.Errorf("OPTIONS expected 204, got %d", resp.StatusCode)
	}
	if resp.Header.Get("Access-Control-Allow-Origin") != "*" {
		t.Error("OPTIONS CORS header missing")
	}
}
