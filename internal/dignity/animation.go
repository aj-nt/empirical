package dignity

import (
	"fmt"
	"math"
	"strings"
)

// ═══════════════════════════════════════════════════════════════════════════
// Chart Animation — Animated SVG transit/progression wheels
// ═══════════════════════════════════════════════════════════════════════════
//
// Generates self-contained animated SVGs using SMIL animations.
// Planets animate along their orbital paths over a configurable time range.

// AnimationConfig configures the animated chart.
type AnimationConfig struct {
	Width       int     // SVG width (default 800)
	Height      int     // SVG height (default 800)
	StartDate   string  // YYYY-MM-DD
	EndDate     string  // YYYY-MM-DD
	Duration    float64 // animation duration in seconds (default 10)
	ShowAspects bool    // animate aspect lines
	ShowHouses  bool    // show house cusps
	DarkMode    bool    // dark background
}

// AnimatedChartData holds the planetary positions at each frame.
type AnimatedChartData struct {
	StartPositions map[string]float64 // planet → longitude at start
	EndPositions   map[string]float64 // planet → longitude at end
	Houses         []float64          // house cusps
	ASC            float64
	MC             float64
}

// classicalPlanets is the standard order of the seven classical planets.
var classicalPlanets = []string{"Sun", "Moon", "Mercury", "Venus", "Mars", "Jupiter", "Saturn"}

// RenderAnimatedWheel generates an animated SVG chart wheel.
// Planets animate from start positions to end positions.
func RenderAnimatedWheel(data *AnimatedChartData, cfg AnimationConfig) string {
	if cfg.Width == 0 {
		cfg.Width = 800
	}
	if cfg.Height == 0 {
		cfg.Height = 800
	}
	if cfg.Duration == 0 {
		cfg.Duration = 10
	}

	cx := float64(cfg.Width) / 2
	cy := float64(cfg.Height) / 2
	r := float64(cfg.Width)/2 - 60

	bg := "#ffffff"
	fg := "#1a1a1a"
	if cfg.DarkMode {
		bg = "#1a1a2e"
		fg = "#e0e0e0"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d" width="%d" height="%d">
<style>
  @keyframes pulse { 0%%,100%% { opacity:1 } 50%% { opacity:0.5 } }
  .planet { transition: transform 0.1s linear; }
  .planet-label { font-family: sans-serif; font-size: 11px; text-anchor: middle; }
  .sign-label { font-family: sans-serif; font-size: 13px; text-anchor: middle; fill: %s; }
  .house-line { stroke: %s; stroke-width: 1; stroke-dasharray: 4,4; }
  .aspect-line { stroke-width: 1.5; opacity: 0.6; }
  .aspect-line-hard { stroke: #c0392b; }
  .aspect-line-soft { stroke: #27ae60; }
</style>
<rect width="%d" height="%d" fill="%s"/>
`, cfg.Width, cfg.Height, cfg.Width, cfg.Height, fg, fg, cfg.Width, cfg.Height, bg))

	// ── Outer ring ──────────────────────────────────────────────────────
	sb.WriteString(fmt.Sprintf(`<circle cx="%.1f" cy="%.1f" r="%.1f" fill="none" stroke="%s" stroke-width="2"/>
`, cx, cy, r, fg))

	// ── Sign ring ───────────────────────────────────────────────────────
	signR := r - 30
	for i := 0; i < 12; i++ {
		angle := float64(i)*30 - 90 // start at Aries (0° = right, -90 = top)
		rad := angle * math.Pi / 180
		lx := cx + signR*math.Cos(rad)
		ly := cy + signR*math.Sin(rad)
		sb.WriteString(fmt.Sprintf(`<text x="%.1f" y="%.1f" class="sign-label">%s</text>
`, lx, ly, Signs[i]))
	}

	// ── House cusps ─────────────────────────────────────────────────────
	if cfg.ShowHouses && len(data.Houses) > 0 {
		for _, cusp := range data.Houses {
			angle := cusp - 90
			rad := angle * math.Pi / 180
			x2 := cx + (r-5)*math.Cos(rad)
			y2 := cy + (r-5)*math.Sin(rad)
			sb.WriteString(fmt.Sprintf(`<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" class="house-line"/>
`, cx, cy, x2, y2))
		}
	}

	// ── Planets ─────────────────────────────────────────────────────────
	planetColors := map[string]string{
		"Sun": "#f39c12", "Moon": "#bdc3c7", "Mercury": "#e67e22",
		"Venus": "#2ecc71", "Mars": "#e74c3c", "Jupiter": "#3498db", "Saturn": "#9b59b6",
	}
	planetRadii := map[string]float64{
		"Sun": 0.55, "Moon": 0.65, "Mercury": 0.75, "Venus": 0.65,
		"Mars": 0.60, "Jupiter": 0.45, "Saturn": 0.50,
	}

	for _, planet := range classicalPlanets {
		startLon, ok1 := data.StartPositions[planet]
		endLon, ok2 := data.EndPositions[planet]
		if !ok1 || !ok2 {
			continue
		}

		color := planetColors[planet]
		pr := planetRadii[planet] * r

		// Normalize: ensure shortest path
		diff := endLon - startLon
		if diff > 180 {
			diff -= 360
		} else if diff < -180 {
			diff += 360
		}
		endLon = startLon + diff

		// Start position
		startAngle := startLon - 90
		startRad := startAngle * math.Pi / 180
		sx := cx + pr*math.Cos(startRad)
		sy := cy + pr*math.Sin(startRad)

		// End position
		endAngle := endLon - 90
		endRad := endAngle * math.Pi / 180
		ex := cx + pr*math.Cos(endRad)
		ey := cy + pr*math.Sin(endRad)

		// Animated planet with SMIL
		sb.WriteString(fmt.Sprintf(`<g>
  <circle r="6" fill="%s" stroke="%s" stroke-width="1">
    <animateMotion dur="%.1fs" repeatCount="indefinite" path="M%.1f,%.1f A%.1f,%.1f 0 0,1 %.1f,%.1f"/>
  </circle>
  <text class="planet-label" fill="%s">
    <animateMotion dur="%.1fs" repeatCount="indefinite" path="M%.1f,%.1f A%.1f,%.1f 0 0,1 %.1f,%.1f"/>
    %s
  </text>
</g>
`, color, fg, cfg.Duration, sx, sy, pr, pr, ex, ey, fg, cfg.Duration, sx, sy, pr, pr, ex, ey, planet))
	}

	// ── Aspect lines (static, between start positions) ──────────────────
	if cfg.ShowAspects {
		aspectTypes := []struct {
			name string
			deg  float64
			cls  string
		}{
			{"conjunction", 0, "aspect-line aspect-line-soft"},
			{"sextile", 60, "aspect-line aspect-line-soft"},
			{"square", 90, "aspect-line aspect-line-hard"},
			{"trine", 120, "aspect-line aspect-line-soft"},
			{"opposition", 180, "aspect-line aspect-line-hard"},
		}

		for i := 0; i < len(classicalPlanets); i++ {
			for j := i + 1; j < len(classicalPlanets); j++ {
				p1, p2 := classicalPlanets[i], classicalPlanets[j]
				lon1, ok1 := data.StartPositions[p1]
				lon2, ok2 := data.StartPositions[p2]
				if !ok1 || !ok2 {
					continue
				}

				dist := math.Abs(lon1 - lon2)
				if dist > 180 {
					dist = 360 - dist
				}

				for _, asp := range aspectTypes {
					orb := math.Abs(dist - asp.deg)
					if orb <= 8.0 {
						pr1 := planetRadii[p1] * r
						pr2 := planetRadii[p2] * r
						a1 := (lon1 - 90) * math.Pi / 180
						a2 := (lon2 - 90) * math.Pi / 180
						x1 := cx + pr1*math.Cos(a1)
						y1 := cy + pr1*math.Sin(a1)
						x2 := cx + pr2*math.Cos(a2)
						y2 := cy + pr2*math.Sin(a2)
						sb.WriteString(fmt.Sprintf(`<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" class="%s"/>
`, x1, y1, x2, y2, asp.cls))
						break
					}
				}
			}
		}
	}

	sb.WriteString("</svg>")
	return sb.String()
}

// ── Animated Bi-Wheel ──────────────────────────────────────────────────────

// RenderAnimatedBiWheel generates an animated bi-wheel SVG.
// Inner wheel = natal (static), outer wheel = transits (animated).
func RenderAnimatedBiWheel(inner, outer *AnimatedChartData, cfg AnimationConfig) string {
	if cfg.Width == 0 {
		cfg.Width = 800
	}
	if cfg.Height == 0 {
		cfg.Height = 800
	}
	if cfg.Duration == 0 {
		cfg.Duration = 10
	}

	cx := float64(cfg.Width) / 2
	cy := float64(cfg.Height) / 2
	innerR := float64(cfg.Width)/2 - 100
	outerR := innerR + 40

	bg := "#ffffff"
	fg := "#1a1a1a"
	if cfg.DarkMode {
		bg = "#1a1a2e"
		fg = "#e0e0e0"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d" width="%d" height="%d">
<style>
  .inner-planet { font-family: sans-serif; font-size: 10px; text-anchor: middle; fill: %s; }
  .outer-planet { font-family: sans-serif; font-size: 10px; text-anchor: middle; fill: #c0392b; }
  .sign-label { font-family: sans-serif; font-size: 13px; text-anchor: middle; fill: %s; }
</style>
<rect width="%d" height="%d" fill="%s"/>
`, cfg.Width, cfg.Height, cfg.Width, cfg.Height, fg, fg, cfg.Width, cfg.Height, bg))

	// Sign ring
	signR := innerR - 20
	for i := 0; i < 12; i++ {
		angle := float64(i)*30 - 90
		rad := angle * math.Pi / 180
		lx := cx + signR*math.Cos(rad)
		ly := cy + signR*math.Sin(rad)
		sb.WriteString(fmt.Sprintf(`<text x="%.1f" y="%.1f" class="sign-label">%s</text>
`, lx, ly, Signs[i]))
	}

	// Inner ring
	sb.WriteString(fmt.Sprintf(`<circle cx="%.1f" cy="%.1f" r="%.1f" fill="none" stroke="%s" stroke-width="1.5"/>
`, cx, cy, innerR, fg))

	// Outer ring
	sb.WriteString(fmt.Sprintf(`<circle cx="%.1f" cy="%.1f" r="%.1f" fill="none" stroke="#c0392b" stroke-width="1.5"/>
`, cx, cy, outerR))

	// Static inner planets
	for _, planet := range classicalPlanets {
		lon, ok := inner.StartPositions[planet]
		if !ok {
			continue
		}
		angle := lon - 90
		rad := angle * math.Pi / 180
		px := cx + innerR*math.Cos(rad)
		py := cy + innerR*math.Sin(rad)
		sb.WriteString(fmt.Sprintf(`<circle cx="%.1f" cy="%.1f" r="4" fill="%s"/>
<text x="%.1f" y="%.1f" class="inner-planet">%s</text>
`, px, py, fg, px, py-10, planet))
	}

	// Animated outer planets
	for _, planet := range classicalPlanets {
		startLon, ok1 := outer.StartPositions[planet]
		endLon, ok2 := outer.EndPositions[planet]
		if !ok1 || !ok2 {
			continue
		}

		diff := endLon - startLon
		if diff > 180 {
			diff -= 360
		} else if diff < -180 {
			diff += 360
		}
		endLon = startLon + diff

		startAngle := startLon - 90
		startRad := startAngle * math.Pi / 180
		sx := cx + outerR*math.Cos(startRad)
		sy := cy + outerR*math.Sin(startRad)

		endAngle := endLon - 90
		endRad := endAngle * math.Pi / 180
		ex := cx + outerR*math.Cos(endRad)
		ey := cy + outerR*math.Sin(endRad)

		sb.WriteString(fmt.Sprintf(`<g>
  <circle r="5" fill="#c0392b" stroke="#c0392b" stroke-width="1">
    <animateMotion dur="%.1fs" repeatCount="indefinite" path="M%.1f,%.1f A%.1f,%.1f 0 0,1 %.1f,%.1f"/>
  </circle>
  <text class="outer-planet">
    <animateMotion dur="%.1fs" repeatCount="indefinite" path="M%.1f,%.1f A%.1f,%.1f 0 0,1 %.1f,%.1f"/>
    %s
  </text>
</g>
`, cfg.Duration, sx, sy, outerR, outerR, ex, ey, cfg.Duration, sx, sy, outerR, outerR, ex, ey, planet))
	}

	sb.WriteString("</svg>")
	return sb.String()
}
