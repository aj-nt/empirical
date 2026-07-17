package dignity

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"strings"
)

//go:embed templates/*.gohtml
var templateFS embed.FS

// templateFuncs provides helper functions available in all templates.
var templateFuncs = template.FuncMap{
	"ordinalSuffix": func(n int) string {
		switch n {
		case 1:
			return "st"
		case 2:
			return "nd"
		case 3:
			return "rd"
		default:
			return "th"
		}
	},
}

// ── Natal render functions ─────────────────────────────────────────────────

// RenderKoinéNatal renders a Koiné natal chart interpretation as HTML.
func RenderKoinéNatal(report *ChartInterpretation) (string, error) {
	return renderTemplate("templates/koine_natal.gohtml", report)
}

// RenderWesternNatal renders a Western natal chart interpretation as HTML.
func RenderWesternNatal(report *ChartInterpretation) (string, error) {
	return renderTemplate("templates/western_natal.gohtml", report)
}

// RenderVedicNatal renders a Vedic natal chart dignity convergence as HTML.
func RenderVedicNatal(report *DignityConvergence) (string, error) {
	return renderTemplate("templates/vedic_natal.gohtml", report)
}

// RenderBaZiNatal renders a BaZi Four Pillars chart as HTML.
func RenderBaZiNatal(report BaZiFourPillars) (string, error) {
	return renderTemplate("templates/bazi_natal.gohtml", report)
}

// ── Transit render functions ────────────────────────────────────────────────

// RenderKoinéTransit renders a Koiné transit report as HTML.
func RenderKoinéTransit(report *TransitReport) (string, error) {
	return renderTemplate("templates/koine_transit.gohtml", report)
}

// RenderWesternTransit renders a Western transit report as HTML.
func RenderWesternTransit(report *TransitReport) (string, error) {
	return renderTemplate("templates/western_transit.gohtml", report)
}

// RenderVedicTransit renders a Vedic transit report as HTML.
func RenderVedicTransit(report *TransitReport) (string, error) {
	return renderTemplate("templates/vedic_transit.gohtml", report)
}

// RenderBaZiTransit renders a BaZi transit report as HTML.
func RenderBaZiTransit(report *TransitReport) (string, error) {
	return renderTemplate("templates/bazi_transit.gohtml", report)
}

// ── Internal ────────────────────────────────────────────────────────────────

func renderTemplate(name string, data any) (string, error) {
	tmpl, err := template.New(name).Funcs(templateFuncs).ParseFS(templateFS, name)
	if err != nil {
		return "", fmt.Errorf("parse template %s: %w", name, err)
	}
	// ParseFS registers the template under the file's base name, not the full path.
	// Look it up by base name for execution.
	base := name
	if idx := strings.LastIndexByte(name, '/'); idx >= 0 {
		base = name[idx+1:]
	}
	tmpl = tmpl.Lookup(base)
	if tmpl == nil {
		return "", fmt.Errorf("template %q not found after parse (looked up %q)", name, base)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute template %s: %w", name, err)
	}
	return buf.String(), nil
}
