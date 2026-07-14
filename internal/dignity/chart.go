package dignity

import (
	"fmt"
	"math"
	"strings"

	"github.com/aj-nt/empirical/internal/swe"
)

// ── Natal Chart SVG Renderer ─────────────────────────────────────────────

// ChartOptions configures the chart rendering.
type ChartOptions struct {
	HouseSystem        string // "placidus", "whole_sign", "equal", "porphyry", "koch"
	ShowAspects        bool
	OuterPlanets       bool
	Sidereal           bool
	HighlightPatterns  bool
	PatternOrb         float64 // orb for pattern detection (default 5°)
}

// DefaultChartOptions returns sensible defaults.
func DefaultChartOptions() ChartOptions {
	return ChartOptions{
		HouseSystem:  "placidus",
		ShowAspects:  true,
		OuterPlanets: true,
		Sidereal:     false,
	}
}

var planetGlyphs = map[string]string{
	"Sun": "\u2609", "Moon": "\u263d", "Mercury": "\u263f",
	"Venus": "\u2640", "Mars": "\u2642", "Jupiter": "\u2643",
	"Saturn": "\u2644", "Uranus": "\u2645", "Neptune": "\u2646",
	"Pluto": "\u2647", "Node": "\u260a",
	"Ceres": "\u26b3", "Pallas": "\u26b4", "Juno": "\u26b5",
	"Vesta": "\u26b6", "Lilith": "\u26b8", "Chiron": "\u26b7",
	"SouthNode": "\u260b",
	"Cupido": "CU", "Hades": "HA", "Zeus": "ZE", "Kronos": "KR",
	"Apollon": "AP", "Admetos": "AD", "Poseidon": "PO", "Vulkanus": "VU",
}

var signGlyphs = map[string]string{
	"Aries": "\u2648", "Taurus": "\u2649", "Gemini": "\u264a",
	"Cancer": "\u264b", "Leo": "\u264c", "Virgo": "\u264d",
	"Libra": "\u264e", "Scorpio": "\u264f", "Sagittarius": "\u2650",
	"Capricorn": "\u2651", "Aquarius": "\u2652", "Pisces": "\u2653",
}

var aspectColors = map[string]string{
	"conjunction": "#d2991d",
	"sextile":     "#58a6ff",
	"square":      "#f85149",
	"trine":       "#3fb950",
	"opposition":  "#a371f7",
}

// RenderChartSVG generates an SVG natal chart wheel.
// All positions are computed fresh via Swiss Ephemeris.
func RenderChartSVG(
	name string,
	year, month, day, hour, minute int,
	tzOffset, lat, lng float64,
	opts ChartOptions,
) string {
	utHour := float64(hour) + float64(minute)/60.0 - tzOffset
	jd := swe.Julday(year, month, day, utHour, true)

	// House cusps and angles
	hcode, ok := swephCode[opts.HouseSystem]
	if !ok {
		hcode = 'P'
	}
	cusps, ascmc := swe.Houses(jd, lat, lng, hcode)
	asc := ascmc[0]
	mc := ascmc[1]

	// Planet positions
	type planetPos struct {
		Name  string
		Lon   float64
		Speed float64 // degrees/day, negative = retrograde
		Glyph string
	}

	planetIDs := []struct {
		name string
		id   int
	}{
		{"Sun", swe.SUN}, {"Moon", swe.MOON}, {"Mercury", swe.MERCURY},
		{"Venus", swe.VENUS}, {"Mars", swe.MARS}, {"Jupiter", swe.JUPITER},
		{"Saturn", swe.SATURN},
	}
	if opts.OuterPlanets {
		planetIDs = append(planetIDs,
			struct {
				name string
				id   int
			}{"Uranus", swe.URANUS},
			struct {
				name string
				id   int
			}{"Neptune", swe.NEPTUNE},
			struct {
				name string
				id   int
			}{"Pluto", swe.PLUTO},
		)
	}

	var planets []planetPos
	ayan := 0.0
	if opts.Sidereal {
		ayan = swe.GetAyanamsaUT(jd)
	}
	for _, p := range planetIDs {
		lon, _, _, speed := swe.CalcUT(jd, p.id)
		if opts.Sidereal {
			lon -= ayan
			if lon < 0 {
				lon += 360
			}
		}
		planets = append(planets, planetPos{Name: p.name, Lon: lon, Speed: speed, Glyph: planetGlyphs[p.name]})
	}

	// North Node
	nnLon, _, _, nnSpeed := swe.CalcUT(jd, swe.MEAN_NODE)
	if opts.Sidereal {
		nnLon -= ayan
		if nnLon < 0 {
			nnLon += 360
		}
	}
	planets = append(planets, planetPos{Name: "Node", Lon: nnLon, Speed: nnSpeed, Glyph: planetGlyphs["Node"]})

	// Asteroids, Lilith, Chiron
	extraIDs := []struct {
		name string
		id   int
	}{
		{"Ceres", swe.CERES}, {"Pallas", swe.PALLAS}, {"Juno", swe.JUNO},
		{"Vesta", swe.VESTA}, {"Lilith", swe.MEAN_APOG}, {"Chiron", swe.CHIRON},
		{"Cupido", swe.CUPIDO}, {"Hades", swe.HADES}, {"Zeus", swe.ZEUS},
		{"Kronos", swe.KRONOS}, {"Apollon", swe.APOLLON}, {"Admetos", swe.ADMETOS},
		{"Poseidon", swe.POSEIDON}, {"Vulkanus", swe.VULKANUS},
	}
	for _, p := range extraIDs {
		lon, _, _, speed := swe.CalcUT(jd, p.id)
		if opts.Sidereal {
			lon -= ayan
			if lon < 0 {
				lon += 360
			}
		}
		planets = append(planets, planetPos{Name: p.name, Lon: lon, Speed: speed, Glyph: planetGlyphs[p.name]})
	}

	// South Node (opposite North Node)
	snLon := nnLon + 180
	if snLon >= 360 {
		snLon -= 360
	}
	planets = append(planets, planetPos{Name: "SouthNode", Lon: snLon, Speed: nnSpeed, Glyph: planetGlyphs["SouthNode"]})

	// ── Build SVG ──────────────────────────────────────────────────────
	var sb strings.Builder
	sb.WriteString(`<svg viewBox="0 0 800 800" xmlns="http://www.w3.org/2000/svg">`)
	sb.WriteString(`<rect width="800" height="800" fill="#0d1117"/>`)

	cx, cy := 400.0, 400.0

	// toAngle converts ecliptic longitude to SVG angle (degrees clockwise from 3-o'clock).
	// ASC is pinned at 9-o'clock (left edge of wheel).
	toAngle := func(lon float64) float64 {
		a := asc - lon + 180
		for a < 0 {
			a += 360
		}
		for a >= 360 {
			a -= 360
		}
		return a
	}

	// toXY converts an SVG angle + radius to (x, y) coordinates.
	toXY := func(angle, r float64) (float64, float64) {
		rad := angle * math.Pi / 180.0
		return cx + r*math.Cos(rad), cy + r*math.Sin(rad)
	}

	// ── Sign ring (outer band) ─────────────────────────────────────────
	outerR := 340.0
	innerR := 310.0
	for i, signName := range Signs {
		signStart := float64(i * 30)
		signEnd := signStart + 30

		a1 := toAngle(signStart)
		a2 := toAngle(signEnd)

		x1o, y1o := toXY(a1, outerR)
		x2o, y2o := toXY(a2, outerR)
		x2i, y2i := toXY(a2, innerR)
		x1i, y1i := toXY(a1, innerR)

		fill := "#161b22"
		if i%2 == 0 {
			fill = "#1c2128"
		}

		// Path: outer arc clockwise → line inward → inner arc counter-clockwise → close
		sb.WriteString(fmt.Sprintf(
			`<path d="M%.1f,%.1f A%.0f,%.0f 0 0,1 %.1f,%.1f L%.1f,%.1f A%.0f,%.0f 0 0,0 %.1f,%.1f Z" fill="%s" stroke="#30363d" stroke-width="0.5"/>`,
			x1o, y1o, outerR, outerR, x2o, y2o,
			x2i, y2i, innerR, innerR, x1i, y1i,
			fill,
		))

		// Sign glyph at midpoint of segment
		midAngle := toAngle(signStart + 15)
		mx, my := toXY(midAngle, 355)
		sb.WriteString(fmt.Sprintf(
			`<text x="%.1f" y="%.1f" fill="#8b949e" font-size="15" text-anchor="middle" dominant-baseline="central">%s</text>`,
			mx, my, signGlyphs[signName],
		))
	}

	// ── Degree tick ring ───────────────────────────────────────────────
	tickOuter := 308.0
	tickInner := 298.0
	for deg := 0.0; deg < 360; deg += 5 {
		angle := toAngle(deg)
		xo, yo := toXY(angle, tickOuter)
		xi, yi := toXY(angle, tickInner)
		sw := "0.3"
		if math.Mod(deg, 10) == 0 {
			sw = "0.6"
		}
		sb.WriteString(fmt.Sprintf(
			`<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="#30363d" stroke-width="%s"/>`,
			xo, yo, xi, yi, sw,
		))
	}

	// ── House cusp lines ──────────────────────────────────────────────
	cuspR := 290.0
	for h := 1; h <= 12; h++ {
		cuspLon := cusps[h]
		angle := toAngle(cuspLon)
		x, y := toXY(angle, cuspR)

		stroke := "#30363d"
		sw := "0.5"
		if h == 1 {
			stroke = "#3fb950"
			sw = "1.5"
		}

		sb.WriteString(fmt.Sprintf(
			`<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="%s" stroke-width="%s"/>`,
			cx, cy, x, y, stroke, sw,
		))

		// House number
		lx, ly := toXY(angle, 280)
		sb.WriteString(fmt.Sprintf(
			`<text x="%.1f" y="%.1f" fill="#8b949e" font-size="9" text-anchor="middle" dominant-baseline="central">%d</text>`,
			lx, ly, h,
		))
	}

	// MC line
	mcAngle := toAngle(mc)
	mcx, mcy := toXY(mcAngle, cuspR)
	sb.WriteString(fmt.Sprintf(
		`<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="#58a6ff" stroke-width="1.5"/>`,
		cx, cy, mcx, mcy,
	))
	mlx, mly := toXY(mcAngle, 275)
	sb.WriteString(fmt.Sprintf(
		`<text x="%.1f" y="%.1f" fill="#58a6ff" font-size="10" text-anchor="middle" dominant-baseline="central" font-weight="bold">MC</text>`,
		mlx, mly,
	))

	// ── Planet markers ─────────────────────────────────────────────────
	planetR := 200.0
	// Track placed positions for overlap avoidance
	type placed struct {
		angle float64
		labelY float64 // offset for degree label
	}
	var placedList []placed

	for _, p := range planets {
		angle := toAngle(p.Lon)
		px, py := toXY(angle, planetR)

		// Check for nearby planets (within 8°) and offset if needed
		offset := 0.0
		for _, prev := range placedList {
			diff := math.Abs(angle - prev.angle)
			if diff > 180 {
				diff = 360 - diff
			}
			if diff < 8 {
				offset += 18
			}
		}
		if offset > 0 {
			px, py = toXY(angle, planetR+offset)
		}
		placedList = append(placedList, placed{angle: angle})

		// Planet circle
		sb.WriteString(fmt.Sprintf(
			`<circle cx="%.1f" cy="%.1f" r="13" fill="#161b22" stroke="#58a6ff" stroke-width="1"/>`,
			px, py,
		))

		// Planet glyph
		sb.WriteString(fmt.Sprintf(
			`<text x="%.1f" y="%.1f" fill="#c9d1d9" font-size="13" text-anchor="middle" dominant-baseline="central">%s</text>`,
			px, py, p.Glyph,
		))

		// Degree label with minutes
		degInSign := math.Mod(p.Lon, 30)
		deg := int(degInSign)
		min := int(math.Round((degInSign - float64(deg)) * 60))
		signName := SignForLongitude(p.Lon)
		dlx, dly := toXY(angle, planetR-20)
		sb.WriteString(fmt.Sprintf(
			`<text x="%.1f" y="%.1f" fill="#8b949e" font-size="7" text-anchor="middle" dominant-baseline="central">%d°%02d′%s</text>`,
			dlx, dly, deg, min, signGlyphs[signName],
		))

		// Retrograde marker (skip nodes — they always move retrograde)
		if p.Speed < 0 && p.Name != "Node" && p.Name != "SouthNode" {
			// Use offset-aware position, not original angle
			rx, ry := toXY(angle, planetR-30)
			if offset > 0 {
				rx, ry = toXY(angle, planetR+offset-30)
			}
			sb.WriteString(fmt.Sprintf(
				`<text x="%.1f" y="%.1f" fill="#f85149" font-size="6" text-anchor="middle" dominant-baseline="central">RX</text>`,
				rx, ry,
			))
		}
	}

	// ── Aspect lines ───────────────────────────────────────────────────
	if opts.ShowAspects {
		aspectDefs := []struct {
			angle float64
			name  string
		}{
			{0, "conjunction"}, {60, "sextile"}, {90, "square"}, {120, "trine"}, {180, "opposition"},
		}
		orb := 5.0

		for i := 0; i < len(planets); i++ {
			for j := i + 1; j < len(planets); j++ {
				diff := math.Abs(planets[i].Lon - planets[j].Lon)
				if diff > 180 {
					diff = 360 - diff
				}

				for _, ad := range aspectDefs {
					if math.Abs(diff-ad.angle) <= orb {
						a1 := toAngle(planets[i].Lon)
						a2 := toAngle(planets[j].Lon)
						x1, y1 := toXY(a1, planetR-15)
						x2, y2 := toXY(a2, planetR-15)

						color := aspectColors[ad.name]
						sb.WriteString(fmt.Sprintf(
							`<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="%s" stroke-width="0.8" opacity="0.4"/>`,
							x1, y1, x2, y2, color,
						))
						break // only draw closest aspect per pair
					}
				}
			}
		}
	}

	// ── Pattern highlighting ──────────────────────────────────────────
	if opts.HighlightPatterns {
		orb := opts.PatternOrb
		if orb <= 0 {
			orb = 5.0
		}
		// Build planet map for pattern detection
		planetMap := make(map[string]float64)
		for _, p := range planets {
			planetMap[p.Name] = p.Lon
		}
		report := DetectPatterns(planetMap, orb)

		patternColors := map[PatternKind]string{
			Stellium:        "#d2991d",
			GrandTrine:      "#3fb950",
			Kite:            "#58a6ff",
			TSquare:         "#f85149",
			GrandCross:      "#a371f7",
			Yod:             "#d2991d",
			MysticRectangle: "#58a6ff",
			Cradle:          "#3fb950",
			Wedge:           "#f85149",
		}

		// Draw pattern polygons
		for pi, pat := range report.Patterns {
			color := patternColors[pat.Kind]
			if color == "" {
				color = "#8b949e"
			}

			// Build polygon points from pattern planets
			var polyPoints []string
			for _, pn := range pat.Planets {
				lon, ok := planetMap[pn]
				if !ok {
					continue
				}
				angle := toAngle(lon)
				x, y := toXY(angle, planetR-10)
				polyPoints = append(polyPoints, fmt.Sprintf("%.1f,%.1f", x, y))
			}

			if len(polyPoints) >= 3 {
				pointsStr := strings.Join(polyPoints, " ")
				sb.WriteString(fmt.Sprintf(
					`<polygon points="%s" fill="%s" opacity="0.08" stroke="%s" stroke-width="1.5" stroke-dasharray="4,3"/>`,
					pointsStr, color, color,
				))
			}

			// Pattern label at midpoint of polygon
			var sumX, sumY float64
			count := 0
			for _, pn := range pat.Planets {
				lon, ok := planetMap[pn]
				if !ok {
					continue
				}
				angle := toAngle(lon)
				x, y := toXY(angle, planetR+35)
				sumX += x
				sumY += y
				count++
			}
			if count > 0 {
				lx, ly := sumX/float64(count), sumY/float64(count)
				sb.WriteString(fmt.Sprintf(
					`<text x="%.1f" y="%.1f" fill="%s" font-size="8" text-anchor="middle" dominant-baseline="central" font-weight="bold">%s</text>`,
					lx, ly, color, pat.Name,
				))
			}

			// Thicker highlight lines for pattern edges
			for _, a := range pat.Aspects {
				lon1, ok1 := planetMap[a.Planet1]
				lon2, ok2 := planetMap[a.Planet2]
				if !ok1 || !ok2 {
					continue
				}
				a1 := toAngle(lon1)
				a2 := toAngle(lon2)
				x1, y1 := toXY(a1, planetR-15)
				x2, y2 := toXY(a2, planetR-15)
				edgeColor := aspectColors[a.Aspect]
				if edgeColor == "" {
					edgeColor = color
				}
				sb.WriteString(fmt.Sprintf(
					`<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="%s" stroke-width="1.5" opacity="0.6"/>`,
					x1, y1, x2, y2, edgeColor,
				))
			}

			_ = pi // suppress unused
		}

		// Pattern legend in bottom-left corner
		if len(report.Patterns) > 0 {
			legendX := 30.0
			legendY := 720.0
			sb.WriteString(fmt.Sprintf(
				`<text x="%.1f" y="%.1f" fill="#8b949e" font-size="9" font-weight="bold">PATTERNS</text>`,
				legendX, legendY,
			))
			for i, pat := range report.Patterns {
				color := patternColors[pat.Kind]
				if color == "" {
					color = "#8b949e"
				}
				ly := legendY + 14 + float64(i)*14
				sb.WriteString(fmt.Sprintf(
					`<rect x="%.1f" y="%.1f" width="8" height="8" fill="%s" rx="1"/>`,
					legendX, ly-4, color,
				))
				sb.WriteString(fmt.Sprintf(
					`<text x="%.1f" y="%.1f" fill="#c9d1d9" font-size="8">%s: %s</text>`,
					legendX+12, ly, pat.Name, strings.Join(pat.Planets, ", "),
				))
			}
		}
	}

	// ── Center label ───────────────────────────────────────────────────
	sb.WriteString(fmt.Sprintf(
		`<text x="%.1f" y="%.1f" fill="#c9d1d9" font-size="15" text-anchor="middle" dominant-baseline="central" font-weight="bold">%s</text>`,
		cx, cy-12, name,
	))
	sb.WriteString(fmt.Sprintf(
		`<text x="%.1f" y="%.1f" fill="#8b949e" font-size="9" text-anchor="middle" dominant-baseline="central">%d-%02d-%02d %02d:%02d</text>`,
		cx, cy+6, year, month, day, hour, minute,
	))
	coordLabel := string(FrameTropical)
	if opts.Sidereal {
		coordLabel = string(FrameSidereal)
	}
	sb.WriteString(fmt.Sprintf(
		`<text x="%.1f" y="%.1f" fill="#8b949e" font-size="8" text-anchor="middle" dominant-baseline="central">%s · %s</text>`,
		cx, cy+20, opts.HouseSystem, coordLabel,
	))

	sb.WriteString(`</svg>`)
	return sb.String()
}
