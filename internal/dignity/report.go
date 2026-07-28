package dignity

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"text/template"
	"time"
)

// ═══════════════════════════════════════════════════════════════════════════
// Report Writer — Template-based astrological report generation
// ═══════════════════════════════════════════════════════════════════════════
//
// Uses Go text/template with a rich set of chart data exposed as template
// variables. Supports custom templates and pre-built report types.

// ReportTemplate is a named template for generating reports.
type ReportTemplate struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Content     string `json:"content"` // Go template text
}

// ReportRequest is a request to generate a report.
type ReportRequest struct {
	TemplateName string `json:"template_name"` // pre-built: natal, transit, synastry, vedic
	CustomTemplate string `json:"custom_template,omitempty"` // custom Go template
	Format       string `json:"format"` // "text" or "markdown" or "html"
}

// ReportResult is the generated report.
type ReportResult struct {
	Content  string `json:"content"`
	Format   string `json:"format"`
	Template string `json:"template"`
}

// ── Template Data ──────────────────────────────────────────────────────────

// ReportData is the data passed to report templates.
type ReportData struct {
	Name       string
	Date       string
	Time       string
	Location   string
	Lat        float64
	Lng        float64
	ASC        float64
	MC         float64
	SunSign    string
	MoonSign   string
	RisingSign string
	Planets    []ReportPlanet
	Houses     []ReportHouse
	Aspects    []ReportAspect
	Patterns   []string
	Elements   map[string]int
	Modes      map[string]int
	// Vedic
	Vedic *ReportVedicData
	// Transits
	Transits []ReportTransit
	// Synastry
	Partner *ReportPartnerData
}

// ReportPlanet holds planet data for templates.
type ReportPlanet struct {
	Name      string
	Sign      string
	House     int
	Degree    float64
	Retrograde bool
	Dignity   string
	HouseRuler string // which house this planet rules
}

// ReportHouse holds house data for templates.
type ReportHouse struct {
	Number int
	Sign   string
	Cusp   float64
	Ruler  string
	PlanetsIn []string
}

// ReportAspect holds aspect data for templates.
type ReportAspect struct {
	Planet1 string
	Planet2 string
	Aspect  string
	Orb     float64
}

// ReportVedicData holds Vedic-specific data.
type ReportVedicData struct {
	Nakshatra     string
	Pada          int
	LagnaSign     string
	CurrentDasha  string
	DashaEnd      string
	Yogas         []string
}

// ReportTransit holds transit data.
type ReportTransit struct {
	Planet    string
	Sign      string
	House     int
	AspectTo  string // natal planet being aspected
	Aspect    string
	Orb       float64
}

// ReportPartnerData holds synastry partner data.
type ReportPartnerData struct {
	Name       string
	SunSign    string
	MoonSign   string
	RisingSign string
	SynastryAspects []ReportAspect
	CompositeHighlights []string
}

// ── Pre-built Templates ────────────────────────────────────────────────────

var prebuiltTemplates = map[string]ReportTemplate{
	"natal": {
		Name:        "natal",
		Description: "Standard natal chart report",
		Content:     natalTemplate,
	},
	"transit": {
		Name:        "transit",
		Description: "Transit report for a target date",
		Content:     transitTemplate,
	},
	"synastry": {
		Name:        "synastry",
		Description: "Synastry/compatibility report",
		Content:     synastryTemplate,
	},
	"vedic": {
		Name:        "vedic",
		Description: "Vedic (Jyotish) natal report",
		Content:     vedicTemplate,
	},
}

// ── Report Generation ──────────────────────────────────────────────────────

// GenerateReport generates a report from a template and chart data.
func GenerateReport(req ReportRequest, data *ReportData) (*ReportResult, error) {
	var tmplContent string
	var tmplName string

	if req.CustomTemplate != "" {
		tmplContent = req.CustomTemplate
		tmplName = "custom"
	} else if pt, ok := prebuiltTemplates[req.TemplateName]; ok {
		tmplContent = pt.Content
		tmplName = pt.Name
	} else {
		return nil, fmt.Errorf("unknown template: %s (available: natal, transit, synastry, vedic)", req.TemplateName)
	}

	tmpl, err := template.New(tmplName).Funcs(reportTemplateFuncs).Parse(tmplContent)
	if err != nil {
		return nil, fmt.Errorf("template parse error: %v", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("template execution error: %v", err)
	}

	format := req.Format
	if format == "" {
		format = "text"
	}

	return &ReportResult{
		Content:  buf.String(),
		Format:   format,
		Template: tmplName,
	}, nil
}

// ── Template Functions ─────────────────────────────────────────────────────

var reportTemplateFuncs = template.FuncMap{
	"join":    strings.Join,
	"lower":   strings.ToLower,
	"upper":   strings.ToUpper,
	"title":   strings.Title,
	"now":     time.Now,
	"formatDate": func(t time.Time) string { return t.Format("January 2, 2006") },
	"elementOf": func(sign string) string {
		elements := map[string]string{
			"Aries": "Fire", "Leo": "Fire", "Sagittarius": "Fire",
			"Taurus": "Earth", "Virgo": "Earth", "Capricorn": "Earth",
			"Gemini": "Air", "Libra": "Air", "Aquarius": "Air",
			"Cancer": "Water", "Scorpio": "Water", "Pisces": "Water",
		}
		return elements[sign]
	},
	"modeOf": func(sign string) string {
		modes := map[string]string{
			"Aries": "Cardinal", "Cancer": "Cardinal", "Libra": "Cardinal", "Capricorn": "Cardinal",
			"Taurus": "Fixed", "Leo": "Fixed", "Scorpio": "Fixed", "Aquarius": "Fixed",
			"Gemini": "Mutable", "Virgo": "Mutable", "Sagittarius": "Mutable", "Pisces": "Mutable",
		}
		return modes[sign]
	},
	"dignityLabel": func(d string) string {
		labels := map[string]string{
			"domicile": "🏠 Domicile", "exaltation": "⬆ Exaltation",
			"detriment": "⬇ Detriment", "fall": "📉 Fall",
			"peregrine": "Peregrine", "": "—",
		}
		if l, ok := labels[d]; ok {
			return l
		}
		return d
	},
	"aspectSymbol": func(a string) string {
		symbols := map[string]string{
			"conjunction": "☌", "sextile": "⚹", "square": "□",
			"trine": "△", "opposition": "☍",
		}
		if s, ok := symbols[a]; ok {
			return s
		}
		return a
	},
	"percent": func(n, total int) string {
		if total == 0 {
			return "0%"
		}
		return fmt.Sprintf("%.0f%%", float64(n)/float64(total)*100)
	},
}

// ── BuildReportData ────────────────────────────────────────────────────────

// BuildReportData builds ReportData from a BaseChart.
func BuildReportData(bc *BaseChart) *ReportData {
	data := &ReportData{
		Name:     bc.Name,
		Date:     fmt.Sprintf("%d-%02d-%02d", bc.Year, bc.Month, bc.Day),
		Time:     fmt.Sprintf("%02d:%02d", bc.Hour, bc.Minute),
		Lat:      bc.Lat,
		Lng:      bc.Lng,
		ASC:      bc.ASC,
		MC:       bc.MC,
		Elements: make(map[string]int),
		Modes:    make(map[string]int),
	}

	// Signs
	ascSignIdx := int(bc.ASC/30) % 12
	data.RisingSign = Signs[ascSignIdx]

	// Planets
	order := []string{"Sun", "Moon", "Mercury", "Venus", "Mars", "Jupiter", "Saturn", "Uranus", "Neptune", "Pluto"}
	for _, name := range order {
		pos, ok := bc.Tropical[name]
		if !ok {
			continue
		}
		signIdx := int(pos.Lon/30) % 12
		sign := Signs[signIdx]
		house := ((signIdx - ascSignIdx + 12) % 12) + 1

		if name == "Sun" {
			data.SunSign = sign
		}
		if name == "Moon" {
			data.MoonSign = sign
		}

		// Element/mode counting
		data.Elements[reportTemplateFuncs["elementOf"].(func(string) string)(sign)]++
		data.Modes[reportTemplateFuncs["modeOf"].(func(string) string)(sign)]++

		data.Planets = append(data.Planets, ReportPlanet{
			Name:       name,
			Sign:       sign,
			House:      house,
			Degree:     math.Mod(pos.Lon, 30),
			Retrograde: pos.Speed < 0,
			Dignity:    "", // filled by caller if needed
		})
	}

	// Houses
	for h := 1; h <= 12; h++ {
		houseSignIdx := (ascSignIdx + h - 1) % 12
		houseSign := Signs[houseSignIdx]
		houseRuler := signRuler(houseSign)

		var planetsIn []string
		for _, p := range data.Planets {
			if p.House == h {
				planetsIn = append(planetsIn, p.Name)
			}
		}

		data.Houses = append(data.Houses, ReportHouse{
			Number:    h,
			Sign:      houseSign,
			Ruler:     houseRuler,
			PlanetsIn: planetsIn,
		})
	}

	return data
}

// ── JSON ───────────────────────────────────────────────────────────────────

// JSON serializes the report result to JSON.
func (r *ReportResult) JSON() ([]byte, error) {
	return json.Marshal(r)
}

// ── Pre-built Template Content ─────────────────────────────────────────────

const natalTemplate = `{{.Name}} — Natal Chart Report
{{.Date}} {{.Time}} | Lat {{printf "%.2f" .Lat}} Lng {{printf "%.2f" .Lng}}

═══ The Three Lights ═══
☉ Sun in {{.SunSign}} | ☽ Moon in {{.MoonSign}} | ASC {{.RisingSign}}

═══ Planets ═══
{{range .Planets}}{{.Name | printf "%-10s"}} {{.Sign | printf "%-12s"}} House {{.House}}  {{if .Retrograde}}℞ {{end}}{{.Degree | printf "%.1f"}}°
{{end}}
═══ Houses ═══
{{range .Houses}}House {{.Number | printf "%2d"}}: {{.Sign | printf "%-12s"}} (Ruler: {{.Ruler}}){{if .PlanetsIn}} — {{join .PlanetsIn ", "}}{{end}}
{{end}}
═══ Element Balance ═══
{{range $el, $n := .Elements}}{{$el | printf "%-6s"}}: {{$n}}
{{end}}
═══ Mode Balance ═══
{{range $m, $n := .Modes}}{{$m | printf "%-9s"}}: {{$n}}
{{end}}
`

const transitTemplate = `{{.Name}} — Transit Report
{{.Date}}

═══ Current Transits ═══
{{range .Transits}}{{.Planet | printf "%-10s"}} in {{.Sign | printf "%-12s"}} (House {{.House}}) {{if .AspectTo}}{{.Aspect | aspectSymbol}} {{.AspectTo}} ({{.Orb | printf "%.1f"}}°){{end}}
{{end}}
{{if not .Transits}}No significant transits for this date.{{end}}
`

const synastryTemplate = `Synastry Report
{{.Name}} & {{.Partner.Name}}

═══ {{.Name}} ═══
☉ {{.SunSign}} | ☽ {{.MoonSign}} | ASC {{.RisingSign}}

═══ {{.Partner.Name}} ═══
☉ {{.Partner.SunSign}} | ☽ {{.Partner.MoonSign}} | ASC {{.Partner.RisingSign}}

═══ Key Aspects ═══
{{range .Partner.SynastryAspects}}{{.Planet1}} {{.Aspect | aspectSymbol}} {{.Planet2}} ({{.Orb | printf "%.1f"}}°)
{{end}}
{{if .Partner.CompositeHighlights}}
═══ Composite Highlights ═══
{{range .Partner.CompositeHighlights}}• {{.}}
{{end}}
{{end}}
`

const vedicTemplate = `{{.Name}} — Vedic (Jyotish) Natal Report

═══ Lagna ═══
{{if .Vedic}}{{.Vedic.LagnaSign}} — {{.Vedic.Nakshatra}} Pada {{.Vedic.Pada}}
{{end}}
═══ Current Dasha ═══
{{if .Vedic}}{{.Vedic.CurrentDasha}} (until {{.Vedic.DashaEnd}})
{{end}}
═══ Yogas ═══
{{if .Vedic}}{{range .Vedic.Yogas}}• {{.}}
{{end}}{{else}}No major yogas detected.
{{end}}
═══ Planets ═══
{{range .Planets}}{{.Name | printf "%-10s"}} {{.Sign | printf "%-12s"}} House {{.House}}  {{.Dignity | dignityLabel}}
{{end}}
`
