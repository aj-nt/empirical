package dignity

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// ═══════════════════════════════════════════════════════════════════════════
// Chart Import — Standard format chart data interchange
// ═══════════════════════════════════════════════════════════════════════════
//
// Supported formats:
//   - AAF (Astrological Association Format) — CSV-based
//   - AstroDatabank — JSON format from astro.com
//   - Simple CSV — name,year,month,day,hour,minute,tz,lat,lng
//   - JSON — {name, year, month, day, hour, minute, tz_offset, lat, lng}

// ImportedChart is a chart imported from an external format.
type ImportedChart struct {
	Name     string  `json:"name"`
	Year     int     `json:"year"`
	Month    int     `json:"month"`
	Day      int     `json:"day"`
	Hour     int     `json:"hour"`
	Minute   int     `json:"minute"`
	Second   int     `json:"second"`
	TZOffset float64 `json:"tz_offset"`
	Lat      float64 `json:"lat"`
	Lng      float64 `json:"lng"`
	Source   string  `json:"source"`   // original source/format
	RawData  string  `json:"raw_data"` // original raw data
}

// ImportResult holds the result of an import operation.
type ImportResult struct {
	Charts []ImportedChart `json:"charts"`
	Errors []string        `json:"errors,omitempty"`
	Format string          `json:"format"`
	Count  int             `json:"count"`
}

// ── Format Detection ───────────────────────────────────────────────────────

// DetectFormat identifies the import format from raw data.
func DetectFormat(data string) string {
	data = strings.TrimSpace(data)

	// JSON
	if strings.HasPrefix(data, "{") || strings.HasPrefix(data, "[") {
		return "json"
	}

	// AAF: starts with "NAME,DATE,TIME,ZONE,LAT,LNG" or similar
	lines := strings.Split(data, "\n")
	if len(lines) > 0 {
		header := strings.ToUpper(strings.TrimSpace(lines[0]))
		if strings.Contains(header, "NAME") && strings.Contains(header, "DATE") {
			return "aaf"
		}
	}

	// Simple CSV
	return "csv"
}

// ── Import ─────────────────────────────────────────────────────────────────

// ImportCharts imports charts from raw data, auto-detecting the format.
func ImportCharts(data string) (*ImportResult, error) {
	format := DetectFormat(data)

	switch format {
	case "json":
		return importJSON(data)
	case "aaf":
		return importAAF(data)
	case "csv":
		return importCSV(data)
	default:
		return nil, fmt.Errorf("unknown format: %s", format)
	}
}

// ── JSON Import ────────────────────────────────────────────────────────────

func importJSON(data string) (*ImportResult, error) {
	result := &ImportResult{Format: "json"}

	// Try array first
	var charts []ImportedChart
	if err := json.Unmarshal([]byte(data), &charts); err == nil {
		for i := range charts {
			charts[i].Source = "json"
			charts[i].RawData = data
		}
		result.Charts = charts
		result.Count = len(charts)
		return result, nil
	}

	// Try single object
	var chart ImportedChart
	if err := json.Unmarshal([]byte(data), &chart); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("JSON parse error: %v", err))
		return result, nil
	}
	chart.Source = "json"
	chart.RawData = data
	result.Charts = []ImportedChart{chart}
	result.Count = 1
	return result, nil
}

// ── AAF Import ─────────────────────────────────────────────────────────────

func importAAF(data string) (*ImportResult, error) {
	result := &ImportResult{Format: "aaf"}

	reader := csv.NewReader(strings.NewReader(data))
	reader.TrimLeadingSpace = true

	// Read header
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("AAF header read error: %v", err)
	}

	// Map column indices
	colMap := make(map[string]int)
	for i, col := range header {
		colMap[strings.ToUpper(strings.TrimSpace(col))] = i
	}

	// Required columns
	required := []string{"NAME", "DATE", "TIME", "ZONE", "LAT", "LNG"}
	for _, req := range required {
		if _, ok := colMap[req]; !ok {
			result.Errors = append(result.Errors, fmt.Sprintf("AAF missing required column: %s", req))
			return result, nil
		}
	}

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("AAF row read error: %v", err))
			continue
		}

		chart, err := parseAAFRecord(record, colMap)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("AAF parse error: %v", err))
			continue
		}
		chart.Source = "aaf"
		result.Charts = append(result.Charts, chart)
	}

	result.Count = len(result.Charts)
	return result, nil
}

func parseAAFRecord(record []string, colMap map[string]int) (ImportedChart, error) {
	get := func(col string) string {
		if idx, ok := colMap[col]; ok && idx < len(record) {
			return strings.TrimSpace(record[idx])
		}
		return ""
	}

	name := get("NAME")
	dateStr := get("DATE")   // YYYY-MM-DD or MM/DD/YYYY
	timeStr := get("TIME")   // HH:MM or HH:MM:SS
	zoneStr := get("ZONE")   // -8 or -8:00
	latStr := get("LAT")     // 47.038 or 47N02
	lngStr := get("LNG")     // -122.901 or 122W54

	// Parse date
	var year, month, day int
	if strings.Contains(dateStr, "-") {
		parts := strings.Split(dateStr, "-")
		if len(parts) == 3 {
			year, _ = strconv.Atoi(parts[0])
			month, _ = strconv.Atoi(parts[1])
			day, _ = strconv.Atoi(parts[2])
		}
	} else if strings.Contains(dateStr, "/") {
		parts := strings.Split(dateStr, "/")
		if len(parts) == 3 {
			month, _ = strconv.Atoi(parts[0])
			day, _ = strconv.Atoi(parts[1])
			year, _ = strconv.Atoi(parts[2])
		}
	}

	// Parse time
	var hour, minute, second int
	timeParts := strings.Split(timeStr, ":")
	if len(timeParts) >= 2 {
		hour, _ = strconv.Atoi(timeParts[0])
		minute, _ = strconv.Atoi(timeParts[1])
	}
	if len(timeParts) >= 3 {
		second, _ = strconv.Atoi(timeParts[2])
	}

	// Parse timezone
	tzOffset, _ := strconv.ParseFloat(strings.TrimSuffix(zoneStr, ":00"), 64)

	// Parse lat/lng
	lat := parseCoordinate(latStr)
	lng := parseCoordinate(lngStr)

	return ImportedChart{
		Name: name, Year: year, Month: month, Day: day,
		Hour: hour, Minute: minute, Second: second,
		TZOffset: tzOffset, Lat: lat, Lng: lng,
	}, nil
}

// parseCoordinate parses a coordinate string like "47N02" or "-122.901" or "122W54".
func parseCoordinate(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}

	// Try direct float
	if val, err := strconv.ParseFloat(s, 64); err == nil {
		return val
	}

	// Try DMS format: 47N02 or 122W54
	s = strings.ToUpper(s)
	neg := 1.0
	if strings.Contains(s, "S") || strings.Contains(s, "W") {
		neg = -1.0
	}

	// Remove direction letter
	s = strings.ReplaceAll(s, "N", "")
	s = strings.ReplaceAll(s, "S", "")
	s = strings.ReplaceAll(s, "E", "")
	s = strings.ReplaceAll(s, "W", "")

	// Parse degrees and minutes
	if len(s) >= 4 {
		deg, _ := strconv.ParseFloat(s[:len(s)-2], 64)
		min, _ := strconv.ParseFloat(s[len(s)-2:], 64)
		return neg * (deg + min/60.0)
	}

	return 0
}

// ── CSV Import ─────────────────────────────────────────────────────────────

func importCSV(data string) (*ImportResult, error) {
	result := &ImportResult{Format: "csv"}

	reader := csv.NewReader(strings.NewReader(data))
	reader.TrimLeadingSpace = true

	// Skip header if present
	firstLine := strings.Split(strings.TrimSpace(data), "\n")[0]
	hasHeader := strings.Contains(strings.ToUpper(firstLine), "NAME") ||
		strings.Contains(strings.ToUpper(firstLine), "YEAR")

	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("CSV read error: %v", err)
	}

	start := 0
	if hasHeader {
		start = 1
	}

	for i := start; i < len(records); i++ {
		record := records[i]
		if len(record) < 9 {
			result.Errors = append(result.Errors, fmt.Sprintf("CSV row %d: expected 9+ fields, got %d", i+1, len(record)))
			continue
		}

		year, _ := strconv.Atoi(strings.TrimSpace(record[1]))
		month, _ := strconv.Atoi(strings.TrimSpace(record[2]))
		day, _ := strconv.Atoi(strings.TrimSpace(record[3]))
		hour, _ := strconv.Atoi(strings.TrimSpace(record[4]))
		minute, _ := strconv.Atoi(strings.TrimSpace(record[5]))
		tz, _ := strconv.ParseFloat(strings.TrimSpace(record[6]), 64)
		lat, _ := strconv.ParseFloat(strings.TrimSpace(record[7]), 64)
		lng, _ := strconv.ParseFloat(strings.TrimSpace(record[8]), 64)

		chart := ImportedChart{
			Name: strings.TrimSpace(record[0]),
			Year: year, Month: month, Day: day,
			Hour: hour, Minute: minute,
			TZOffset: tz, Lat: lat, Lng: lng,
			Source: "csv",
		}
		result.Charts = append(result.Charts, chart)
	}

	result.Count = len(result.Charts)
	return result, nil
}

// ── Export ─────────────────────────────────────────────────────────────────

// ExportChart exports a chart to the specified format.
func ExportChart(chart ImportedChart, format string) (string, error) {
	switch format {
	case "json":
		b, err := json.MarshalIndent(chart, "", "  ")
		return string(b), err
	case "csv":
		return fmt.Sprintf("%s,%d,%d,%d,%d,%d,%.1f,%.3f,%.3f",
			chart.Name, chart.Year, chart.Month, chart.Day,
			chart.Hour, chart.Minute, chart.TZOffset, chart.Lat, chart.Lng), nil
	case "aaf":
		return fmt.Sprintf("NAME,DATE,TIME,ZONE,LAT,LNG\n%s,%04d-%02d-%02d,%02d:%02d,%.0f,%.3f,%.3f",
			chart.Name, chart.Year, chart.Month, chart.Day,
			chart.Hour, chart.Minute, chart.TZOffset, chart.Lat, chart.Lng), nil
	default:
		return "", fmt.Errorf("unsupported export format: %s", format)
	}
}

// ── BirthData Conversion ───────────────────────────────────────────────────

// ToBirthData converts an ImportedChart to a BirthData for use with the engine.
func (c ImportedChart) ToBirthData() BirthData {
	return BirthData{
		Name: c.Name, Year: c.Year, Month: c.Month, Day: c.Day,
		Hour: c.Hour, Minute: c.Minute, TZOffset: c.TZOffset,
		Lat: c.Lat, Lng: c.Lng,
	}
}
