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
	ShowAsteroids      bool // Ceres, Pallas, Juno, Vesta, Chiron
	ShowTNPs           bool // Cupido, Hades, Zeus, Kronos, Apollon, Admetos, Poseidon, Vulkanus
	ShowLilith         bool
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
	// Major asteroids (0-999)
	"Astraea": "AS", "Hebe": "HB", "Iris": "IR", "Flora": "FL",
	"Metis": "ME", "Hygiea": "HY", "Psyche": "PS", "Fortuna": "FO",
	"Proserpina": "PR", "Amphitrite": "AM", "Pandora": "PA",
	"Mnemosyne": "MN", "Cybele": "CY", "Diana": "DI", "Sappho": "SA",
	"Eros": "ER",
	// Distant objects
	"Orcus": "OR", "Sedna": "SE", "Haumea": "HU",
	"Eris": "\u26b1", "Makemake": "MK", "Gonggong": "GO",
}

var signGlyphs = map[string]string{
	"Aries": "\u2648", "Taurus": "\u2649", "Gemini": "\u264a",
	"Cancer": "\u264b", "Leo": "\u264c", "Virgo": "\u264d",
	"Libra": "\u264e", "Scorpio": "\u264f", "Sagittarius": "\u2650",
	"Capricorn": "\u2651", "Aquarius": "\u2652", "Pisces": "\u2653",
}

var aspectColors = map[string]string{
	"conjunction": "#000000",
	"sextile":     "#000000",
	"square":      "#000000",
	"trine":       "#000000",
	"opposition":  "#000000",
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
			// Dwarf planets — always included with outer planets
			struct {
				name string
				id   int
			}{"Eris", swe.ERIS},
			struct {
				name string
				id   int
			}{"Makemake", swe.MAKEMAKE},
			struct {
				name string
				id   int
			}{"Gonggong", swe.GONGGONG},
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
	var extraIDs []struct {
		name string
		id   int
	}
	if opts.ShowAsteroids {
		extraIDs = append(extraIDs,
			struct {
				name string
				id   int
			}{"Ceres", swe.CERES},
			struct {
				name string
				id   int
			}{"Pallas", swe.PALLAS},
			struct {
				name string
				id   int
			}{"Juno", swe.JUNO},
			struct {
				name string
				id   int
			}{"Vesta", swe.VESTA},
			struct {
				name string
				id   int
			}{"Chiron", swe.CHIRON},
			// Major asteroids (0-999)
			struct {
				name string
				id   int
			}{"Astraea", swe.ASTRAEA},
			struct {
				name string
				id   int
			}{"Hebe", swe.HEBE},
			struct {
				name string
				id   int
			}{"Iris", swe.IRIS},
			struct {
				name string
				id   int
			}{"Flora", swe.FLORA},
			struct {
				name string
				id   int
			}{"Metis", swe.METIS},
			struct {
				name string
				id   int
			}{"Hygiea", swe.HYGIEA},
			struct {
				name string
				id   int
			}{"Psyche", swe.PSYCHE},
			struct {
				name string
				id   int
			}{"Fortuna", swe.FORTUNA},
			struct {
				name string
				id   int
			}{"Proserpina", swe.PROSERPINA},
			struct {
				name string
				id   int
			}{"Amphitrite", swe.AMPHITRITE},
			struct {
				name string
				id   int
			}{"Pandora", swe.PANDORA},
			struct {
				name string
				id   int
			}{"Mnemosyne", swe.MNEMOSYNE},
			struct {
				name string
				id   int
			}{"Cybele", swe.CYBELE},
			struct {
				name string
				id   int
			}{"Diana", swe.DIANA},
			struct {
				name string
				id   int
			}{"Sappho", swe.SAPPHO},
			struct {
				name string
				id   int
			}{"Eros", swe.EROS},
			struct {
				name string
				id   int
			}{"Orcus", swe.ORCUS},
			struct {
				name string
				id   int
			}{"Sedna", swe.SEDNA},
			struct {
				name string
				id   int
			}{"Haumea", swe.HAUMEA},
		)
	}
	if opts.ShowLilith {
		extraIDs = append(extraIDs,
			struct {
				name string
				id   int
			}{"Lilith", swe.MEAN_APOG},
		)
	}
	if opts.ShowTNPs {
		extraIDs = append(extraIDs,
			struct {
				name string
				id   int
			}{"Cupido", swe.CUPIDO},
			struct {
				name string
				id   int
			}{"Hades", swe.HADES},
			struct {
				name string
				id   int
			}{"Zeus", swe.ZEUS},
			struct {
				name string
				id   int
			}{"Kronos", swe.KRONOS},
			struct {
				name string
				id   int
			}{"Apollon", swe.APOLLON},
			struct {
				name string
				id   int
			}{"Admetos", swe.ADMETOS},
			struct {
				name string
				id   int
			}{"Poseidon", swe.POSEIDON},
			struct {
				name string
				id   int
			}{"Vulkanus", swe.VULKANUS},
		)
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
	sb.WriteString(`<rect width="800" height="800" fill="#ffffff"/>`)

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
	innerR := 305.0
	for i, signName := range Signs {
		signStart := float64(i * 30)
		signEnd := signStart + 30

		a1 := toAngle(signStart)
		a2 := toAngle(signEnd)

		x1o, y1o := toXY(a1, outerR)
		x2o, y2o := toXY(a2, outerR)
		x2i, y2i := toXY(a2, innerR)
		x1i, y1i := toXY(a1, innerR)

		fill := "#ffffff"
		if i%2 == 0 {
			fill = "#f5f5f5"
		}

		// Path: outer arc clockwise → line inward → inner arc counter-clockwise → close
		sb.WriteString(fmt.Sprintf(
			`<path d="M%.1f,%.1f A%.0f,%.0f 0 0,1 %.1f,%.1f L%.1f,%.1f A%.0f,%.0f 0 0,0 %.1f,%.1f Z" fill="%s" stroke="#000000" stroke-width="1"/>`,
			x1o, y1o, outerR, outerR, x2o, y2o,
			x2i, y2i, innerR, innerR, x1i, y1i,
			fill,
		))

		// Sign glyph at midpoint of segment
		midAngle := toAngle(signStart + 15)
		mx, my := toXY(midAngle, 352)
		sb.WriteString(fmt.Sprintf(
			`<text x="%.1f" y="%.1f" fill="#000000" font-size="20" text-anchor="middle" dominant-baseline="central" font-weight="bold">%s</text>`,
			mx, my, signGlyphs[signName],
		))
	}

	// Outer and inner border circles
	sb.WriteString(fmt.Sprintf(
		`<circle cx="%.1f" cy="%.1f" r="%.0f" fill="none" stroke="#000000" stroke-width="2"/>`,
		cx, cy, outerR,
	))
	sb.WriteString(fmt.Sprintf(
		`<circle cx="%.1f" cy="%.1f" r="%.0f" fill="none" stroke="#000000" stroke-width="1"/>`,
		cx, cy, innerR,
	))

	// ── Degree tick ring ───────────────────────────────────────────────
	tickOuter := 303.0
	tickInner := 290.0
	for deg := 0.0; deg < 360; deg += 5 {
		angle := toAngle(deg)
		xo, yo := toXY(angle, tickOuter)
		xi, yi := toXY(angle, tickInner)
		sw := "0.5"
		if math.Mod(deg, 10) == 0 {
			sw = "1"
		}
		sb.WriteString(fmt.Sprintf(
			`<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="#000000" stroke-width="%s"/>`,
			xo, yo, xi, yi, sw,
		))
	}

	// ── House cusp lines ──────────────────────────────────────────────
	cuspR := 280.0
	for h := 1; h <= 12; h++ {
		cuspLon := cusps[h]
		angle := toAngle(cuspLon)
		x, y := toXY(angle, cuspR)

		stroke := "#000000"
		sw := "1"
		if h == 1 {
			sw = "2"
		}

		sb.WriteString(fmt.Sprintf(
			`<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="%s" stroke-width="%s"/>`,
			cx, cy, x, y, stroke, sw,
		))

		// House number
		lx, ly := toXY(angle, 265)
		sb.WriteString(fmt.Sprintf(
			`<text x="%.1f" y="%.1f" fill="#000000" font-size="16" text-anchor="middle" dominant-baseline="central" font-weight="bold">%d</text>`,
			lx, ly, h,
		))
	}

	// MC line
	mcAngle := toAngle(mc)
	mcx, mcy := toXY(mcAngle, cuspR)
	sb.WriteString(fmt.Sprintf(
		`<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="#000000" stroke-width="2"/>`,
		cx, cy, mcx, mcy,
	))
	mlx, mly := toXY(mcAngle, 258)
	sb.WriteString(fmt.Sprintf(
		`<text x="%.1f" y="%.1f" fill="#000000" font-size="14" text-anchor="middle" dominant-baseline="central" font-weight="bold">MC</text>`,
		mlx, mly,
	))

	// ── Planet markers ─────────────────────────────────────────────────
	planetR := 195.0
	// Track placed positions for overlap avoidance
	type placed struct {
		angle float64
		count int // how many planets at this angle
	}
	var placedList []placed

	for _, p := range planets {
		angle := toAngle(p.Lon)
		px, py := toXY(angle, planetR)

		// Check for nearby planets (within 10°) and offset if needed
		offset := 0.0
		labelStagger := 0 // stagger index for degree labels
		for pi := range placedList {
			prev := &placedList[pi]
			diff := math.Abs(angle - prev.angle)
			if diff > 180 {
				diff = 360 - diff
			}
			if diff < 10 {
				offset += 22
				labelStagger = prev.count
				prev.count++
			}
		}
		if offset > 0 {
			px, py = toXY(angle, planetR+offset)
		}
		placedList = append(placedList, placed{angle: angle, count: 1})

		// Planet circle — white fill, black border
		sb.WriteString(fmt.Sprintf(
			`<circle cx="%.1f" cy="%.1f" r="16" fill="#ffffff" stroke="#000000" stroke-width="2"/>`,
			px, py,
		))

		// Planet glyph
		sb.WriteString(fmt.Sprintf(
			`<text x="%.1f" y="%.1f" fill="#000000" font-size="16" text-anchor="middle" dominant-baseline="central" font-weight="bold">%s</text>`,
			px, py, p.Glyph,
		))

		// Degree label — staggered inward to avoid overlap
		degInSign := math.Mod(p.Lon, 30)
		deg := int(degInSign)
		min := int(math.Round((degInSign - float64(deg)) * 60))
		signName := SignForLongitude(p.Lon)
		// Stagger: first label at -30, second at -50, third at -70
		labelR := planetR - 30 - float64(labelStagger)*20
		dlx, dly := toXY(angle, labelR)
		sb.WriteString(fmt.Sprintf(
			`<text x="%.1f" y="%.1f" fill="#000000" font-size="12" text-anchor="middle" dominant-baseline="central" font-weight="bold">%d°%02d′%s</text>`,
			dlx, dly, deg, min, signGlyphs[signName],
		))

		// Retrograde marker (skip nodes — they always move retrograde)
		if p.Speed < 0 && p.Name != "Node" && p.Name != "SouthNode" {
			rxR := labelR - 12
			rx, ry := toXY(angle, rxR)
			sb.WriteString(fmt.Sprintf(
				`<text x="%.1f" y="%.1f" fill="#000000" font-size="10" text-anchor="middle" dominant-baseline="central" font-weight="bold">℞</text>`,
				rx, ry,
			))
		}
	}

	// ── Aspect lines (straight chords) ─────────────────────────────────
	if opts.ShowAspects {
		aspectDefs := []struct {
			angle float64
			name  string
		}{
			{0, "conjunction"}, {90, "square"}, {120, "trine"}, {180, "opposition"},
		}
		orb := 3.0

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
						x1, y1 := toXY(a1, planetR)
						x2, y2 := toXY(a2, planetR)

						color := aspectColors[ad.name]
						sb.WriteString(fmt.Sprintf(
							`<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="%s" stroke-width="0.5" opacity="0.3"/>`,
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
				x, y := toXY(angle, planetR)
				polyPoints = append(polyPoints, fmt.Sprintf("%.1f,%.1f", x, y))
			}

			if len(polyPoints) >= 3 {
				pointsStr := strings.Join(polyPoints, " ")
				sb.WriteString(fmt.Sprintf(
					`<polygon points="%s" fill="%s" opacity="0.15" stroke="%s" stroke-width="2" stroke-dasharray="6,4"/>`,
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
				x, y := toXY(angle, planetR+40)
				sumX += x
				sumY += y
				count++
			}
			if count > 0 {
				lx, ly := sumX/float64(count), sumY/float64(count)
				sb.WriteString(fmt.Sprintf(
					`<text x="%.1f" y="%.1f" fill="%s" font-size="10" text-anchor="middle" dominant-baseline="central" font-weight="bold">%s</text>`,
					lx, ly, color, pat.Name,
				))
			}

			// Pattern edge lines (straight chords)
			for _, a := range pat.Aspects {
				lon1, ok1 := planetMap[a.Planet1]
				lon2, ok2 := planetMap[a.Planet2]
				if !ok1 || !ok2 {
					continue
				}
				a1 := toAngle(lon1)
				a2 := toAngle(lon2)
				x1, y1 := toXY(a1, planetR)
				x2, y2 := toXY(a2, planetR)
				edgeColor := aspectColors[a.Aspect]
				if edgeColor == "" {
					edgeColor = color
				}
				sb.WriteString(fmt.Sprintf(
					`<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="%s" stroke-width="2" opacity="0.7"/>`,
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
				`<text x="%.1f" y="%.1f" fill="#a0aec0" font-size="11" font-weight="bold">PATTERNS</text>`,
				legendX, legendY,
			))
			for i, pat := range report.Patterns {
				color := patternColors[pat.Kind]
				if color == "" {
					color = "#8b949e"
				}
				ly := legendY + 16 + float64(i)*16
				sb.WriteString(fmt.Sprintf(
					`<rect x="%.1f" y="%.1f" width="10" height="10" fill="%s" rx="2"/>`,
					legendX, ly-5, color,
				))
				sb.WriteString(fmt.Sprintf(
					`<text x="%.1f" y="%.1f" fill="#c9d1d9" font-size="10">%s: %s</text>`,
					legendX+14, ly, pat.Name, strings.Join(pat.Planets, ", "),
				))
			}
		}
	}

	// ── Center label ───────────────────────────────────────────────────
	sb.WriteString(fmt.Sprintf(
		`<text x="%.1f" y="%.1f" fill="#000000" font-size="18" text-anchor="middle" dominant-baseline="central" font-weight="bold">%s</text>`,
		cx, cy-14, name,
	))
	sb.WriteString(fmt.Sprintf(
		`<text x="%.1f" y="%.1f" fill="#000000" font-size="16" text-anchor="middle" dominant-baseline="central">%d-%02d-%02d %02d:%02d</text>`,
		cx, cy+8, year, month, day, hour, minute,
	))
	coordLabel := string(FrameTropical)
	if opts.Sidereal {
		coordLabel = string(FrameSidereal)
	}
	sb.WriteString(fmt.Sprintf(
		`<text x="%.1f" y="%.1f" fill="#000000" font-size="14" text-anchor="middle" dominant-baseline="central">%s · %s</text>`,
		cx, cy+24, opts.HouseSystem, coordLabel,
	))

	sb.WriteString(`</svg>`)
	return sb.String()
}

// ── Bi-Wheel SVG Renderer ─────────────────────────────────────────────────
//
// A bi-wheel shows two charts on the same zodiac ring:
//   - Inner ring: primary (natal) chart planets
//   - Outer ring: secondary (transit/progressed) chart planets
//   - Sign ring: shared zodiac
//   - House cusps: from the primary chart
//   - Aspect lines: between inner and outer planets

// BiWheelOptions configures the bi-wheel rendering.
type BiWheelOptions struct {
	HouseSystem       string  // house system for primary chart
	ShowAspects       bool
	ShowOuterAspects  bool    // aspects between inner and outer planets
	ShowInnerAspects  bool    // aspects within inner chart
	ShowOuterPlanets  bool
	ShowAsteroids     bool
	ShowTNPs          bool
	Sidereal          bool
	Orb               float64 // aspect orb (default 3°)
}

// DefaultBiWheelOptions returns sensible defaults.
func DefaultBiWheelOptions() BiWheelOptions {
	return BiWheelOptions{
		HouseSystem:      "placidus",
		ShowAspects:      true,
		ShowOuterAspects: true,
		ShowInnerAspects: false,
		ShowOuterPlanets: true,
		Sidereal:         false,
		Orb:              3.0,
	}
}

// RenderBiWheelSVG generates a bi-wheel SVG with two charts.
func RenderBiWheelSVG(
	innerName string, innerYear, innerMonth, innerDay, innerHour, innerMinute int,
	innerTZOffset, innerLat, innerLng float64,
	outerName string, outerYear, outerMonth, outerDay, outerHour, outerMinute int,
	outerTZOffset, outerLat, outerLng float64,
	opts BiWheelOptions,
) string {
	// ── Compute inner (primary) chart ──────────────────────────────────
	innerUT := float64(innerHour) + float64(innerMinute)/60.0 - innerTZOffset
	innerJD := swe.Julday(innerYear, innerMonth, innerDay, innerUT, true)

	hcode, ok := swephCode[opts.HouseSystem]
	if !ok {
		hcode = 'P'
	}
	cusps, ascmc := swe.Houses(innerJD, innerLat, innerLng, hcode)
	innerASC := ascmc[0]
	innerMC := ascmc[1]

	// ── Compute outer (secondary) chart ────────────────────────────────
	outerUT := float64(outerHour) + float64(outerMinute)/60.0 - outerTZOffset
	outerJD := swe.Julday(outerYear, outerMonth, outerDay, outerUT, true)

	// ── Planet positions ───────────────────────────────────────────────
	type planetPos struct {
		Name  string
		Lon   float64
		Speed float64
		Glyph string
	}

	// Build planet ID list
	planetIDs := []struct {
		name string
		id   int
	}{
		{"Sun", swe.SUN}, {"Moon", swe.MOON}, {"Mercury", swe.MERCURY},
		{"Venus", swe.VENUS}, {"Mars", swe.MARS}, {"Jupiter", swe.JUPITER},
		{"Saturn", swe.SATURN},
	}
	if opts.ShowOuterPlanets {
		planetIDs = append(planetIDs,
			struct{ name string; id int }{"Uranus", swe.URANUS},
			struct{ name string; id int }{"Neptune", swe.NEPTUNE},
			struct{ name string; id int }{"Pluto", swe.PLUTO},
			struct{ name string; id int }{"Eris", swe.ERIS},
			struct{ name string; id int }{"Makemake", swe.MAKEMAKE},
			struct{ name string; id int }{"Gonggong", swe.GONGGONG},
		)
	}

	var innerPlanets, outerPlanets []planetPos

	innerAyan := 0.0
	outerAyan := 0.0
	if opts.Sidereal {
		innerAyan = swe.GetAyanamsaUT(innerJD)
		outerAyan = swe.GetAyanamsaUT(outerJD)
	}

	for _, p := range planetIDs {
		// Inner
		lon, _, _, speed := swe.CalcUT(innerJD, p.id)
		if opts.Sidereal {
			lon -= innerAyan
			if lon < 0 { lon += 360 }
		}
		innerPlanets = append(innerPlanets, planetPos{Name: p.name, Lon: lon, Speed: speed, Glyph: planetGlyphs[p.name]})

		// Outer
		lon, _, _, speed = swe.CalcUT(outerJD, p.id)
		if opts.Sidereal {
			lon -= outerAyan
			if lon < 0 { lon += 360 }
		}
		outerPlanets = append(outerPlanets, planetPos{Name: p.name, Lon: lon, Speed: speed, Glyph: planetGlyphs[p.name]})
	}

	// Nodes
	nnLon, _, _, nnSpeed := swe.CalcUT(innerJD, swe.MEAN_NODE)
	if opts.Sidereal { nnLon -= innerAyan; if nnLon < 0 { nnLon += 360 } }
	innerPlanets = append(innerPlanets, planetPos{Name: "Node", Lon: nnLon, Speed: nnSpeed, Glyph: planetGlyphs["Node"]})
	snLon := nnLon + 180; if snLon >= 360 { snLon -= 360 }
	innerPlanets = append(innerPlanets, planetPos{Name: "SouthNode", Lon: snLon, Speed: nnSpeed, Glyph: planetGlyphs["SouthNode"]})

	nnLon, _, _, nnSpeed = swe.CalcUT(outerJD, swe.MEAN_NODE)
	if opts.Sidereal { nnLon -= outerAyan; if nnLon < 0 { nnLon += 360 } }
	outerPlanets = append(outerPlanets, planetPos{Name: "Node", Lon: nnLon, Speed: nnSpeed, Glyph: planetGlyphs["Node"]})
	snLon = nnLon + 180; if snLon >= 360 { snLon -= 360 }
	outerPlanets = append(outerPlanets, planetPos{Name: "SouthNode", Lon: snLon, Speed: nnSpeed, Glyph: planetGlyphs["SouthNode"]})

	// Asteroids
	if opts.ShowAsteroids {
		extraIDs := []struct{ name string; id int }{
			{"Ceres", swe.CERES}, {"Pallas", swe.PALLAS}, {"Juno", swe.JUNO},
			{"Vesta", swe.VESTA}, {"Chiron", swe.CHIRON},
			{"Astraea", swe.ASTRAEA}, {"Hebe", swe.HEBE}, {"Iris", swe.IRIS},
			{"Flora", swe.FLORA}, {"Metis", swe.METIS}, {"Hygiea", swe.HYGIEA},
			{"Psyche", swe.PSYCHE}, {"Fortuna", swe.FORTUNA}, {"Proserpina", swe.PROSERPINA},
			{"Amphitrite", swe.AMPHITRITE}, {"Pandora", swe.PANDORA},
			{"Mnemosyne", swe.MNEMOSYNE}, {"Cybele", swe.CYBELE}, {"Diana", swe.DIANA},
			{"Sappho", swe.SAPPHO}, {"Eros", swe.EROS},
			{"Orcus", swe.ORCUS}, {"Sedna", swe.SEDNA}, {"Haumea", swe.HAUMEA},
		}
		for _, p := range extraIDs {
			lon, _, _, speed := swe.CalcUT(innerJD, p.id)
			if opts.Sidereal { lon -= innerAyan; if lon < 0 { lon += 360 } }
			innerPlanets = append(innerPlanets, planetPos{Name: p.name, Lon: lon, Speed: speed, Glyph: planetGlyphs[p.name]})

			lon, _, _, speed = swe.CalcUT(outerJD, p.id)
			if opts.Sidereal { lon -= outerAyan; if lon < 0 { lon += 360 } }
			outerPlanets = append(outerPlanets, planetPos{Name: p.name, Lon: lon, Speed: speed, Glyph: planetGlyphs[p.name]})
		}
	}

	// TNPs
	if opts.ShowTNPs {
		tnpIDs := []struct{ name string; id int }{
			{"Cupido", swe.CUPIDO}, {"Hades", swe.HADES}, {"Zeus", swe.ZEUS},
			{"Kronos", swe.KRONOS}, {"Apollon", swe.APOLLON}, {"Admetos", swe.ADMETOS},
			{"Poseidon", swe.POSEIDON}, {"Vulkanus", swe.VULKANUS},
		}
		for _, p := range tnpIDs {
			lon, _, _, speed := swe.CalcUT(innerJD, p.id)
			if opts.Sidereal { lon -= innerAyan; if lon < 0 { lon += 360 } }
			innerPlanets = append(innerPlanets, planetPos{Name: p.name, Lon: lon, Speed: speed, Glyph: planetGlyphs[p.name]})

			lon, _, _, speed = swe.CalcUT(outerJD, p.id)
			if opts.Sidereal { lon -= outerAyan; if lon < 0 { lon += 360 } }
			outerPlanets = append(outerPlanets, planetPos{Name: p.name, Lon: lon, Speed: speed, Glyph: planetGlyphs[p.name]})
		}
	}

	// ── Build SVG ──────────────────────────────────────────────────────
	var sb strings.Builder
	sb.WriteString(`<svg viewBox="0 0 800 800" xmlns="http://www.w3.org/2000/svg">`)
	sb.WriteString(`<rect width="800" height="800" fill="#ffffff"/>`)

	cx, cy := 400.0, 400.0

	// toAngle converts ecliptic longitude to SVG angle (degrees clockwise from 3-o'clock).
	// ASC is pinned at 9-o'clock (left edge of wheel).
	toAngle := func(lon float64) float64 {
		a := innerASC - lon + 180
		for a < 0 { a += 360 }
		for a >= 360 { a -= 360 }
		return a
	}

	toXY := func(angle, r float64) (float64, float64) {
		rad := angle * math.Pi / 180.0
		return cx + r*math.Cos(rad), cy + r*math.Sin(rad)
	}

	// ── Sign ring (outer band) ─────────────────────────────────────────
	outerR := 340.0
	innerR := 305.0
	for i, signName := range Signs {
		signStart := float64(i * 30)
		signEnd := signStart + 30

		a1 := toAngle(signStart)
		a2 := toAngle(signEnd)

		x1o, y1o := toXY(a1, outerR)
		x2o, y2o := toXY(a2, outerR)
		x2i, y2i := toXY(a2, innerR)
		x1i, y1i := toXY(a1, innerR)

		fill := "#ffffff"
		if i%2 == 0 { fill = "#f5f5f5" }

		sb.WriteString(fmt.Sprintf(
			`<path d="M%.1f,%.1f A%.0f,%.0f 0 0,1 %.1f,%.1f L%.1f,%.1f A%.0f,%.0f 0 0,0 %.1f,%.1f Z" fill="%s" stroke="#000000" stroke-width="1"/>`,
			x1o, y1o, outerR, outerR, x2o, y2o,
			x2i, y2i, innerR, innerR, x1i, y1i,
			fill,
		))

		midAngle := toAngle(signStart + 15)
		mx, my := toXY(midAngle, 352)
		sb.WriteString(fmt.Sprintf(
			`<text x="%.1f" y="%.1f" fill="#000000" font-size="20" text-anchor="middle" dominant-baseline="central" font-weight="bold">%s</text>`,
			mx, my, signGlyphs[signName],
		))
	}

	// Border circles
	sb.WriteString(fmt.Sprintf(`<circle cx="%.1f" cy="%.1f" r="%.0f" fill="none" stroke="#000000" stroke-width="2"/>`, cx, cy, outerR))
	sb.WriteString(fmt.Sprintf(`<circle cx="%.1f" cy="%.1f" r="%.0f" fill="none" stroke="#000000" stroke-width="1"/>`, cx, cy, innerR))

	// ── Degree tick ring ───────────────────────────────────────────────
	tickOuter := 303.0
	tickInner := 290.0
	for deg := 0.0; deg < 360; deg += 5 {
		angle := toAngle(deg)
		xo, yo := toXY(angle, tickOuter)
		xi, yi := toXY(angle, tickInner)
		sw := "0.5"
		if math.Mod(deg, 10) == 0 { sw = "1" }
		sb.WriteString(fmt.Sprintf(
			`<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="#000000" stroke-width="%s"/>`,
			xo, yo, xi, yi, sw,
		))
	}

	// ── House cusp lines (from inner chart) ────────────────────────────
	cuspR := 280.0
	for h := 1; h <= 12; h++ {
		cuspLon := cusps[h]
		angle := toAngle(cuspLon)
		x, y := toXY(angle, cuspR)

		sw := "1"
		if h == 1 { sw = "2" }
		sb.WriteString(fmt.Sprintf(
			`<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="#000000" stroke-width="%s"/>`,
			cx, cy, x, y, sw,
		))

		lx, ly := toXY(angle, 265)
		sb.WriteString(fmt.Sprintf(
			`<text x="%.1f" y="%.1f" fill="#000000" font-size="16" text-anchor="middle" dominant-baseline="central" font-weight="bold">%d</text>`,
			lx, ly, h,
		))
	}

	// MC line
	mcAngle := toAngle(innerMC)
	mcx, mcy := toXY(mcAngle, cuspR)
	sb.WriteString(fmt.Sprintf(
		`<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="#000000" stroke-width="2"/>`,
		cx, cy, mcx, mcy,
	))
	mlx, mly := toXY(mcAngle, 258)
	sb.WriteString(fmt.Sprintf(
		`<text x="%.1f" y="%.1f" fill="#000000" font-size="14" text-anchor="middle" dominant-baseline="central" font-weight="bold">MC</text>`,
		mlx, mly,
	))

	// ── Helper: draw planet set at given radius ────────────────────────
	drawPlanets := func(planets []planetPos, planetR float64, color string, small bool) {
		type placed struct {
			angle float64
			count int
		}
		var placedList []placed

		circleR := 16.0
		glyphSize := 16
		labelSize := 12
		if small {
			circleR = 12.0
			glyphSize = 12
			labelSize = 9
		}

		for _, p := range planets {
			angle := toAngle(p.Lon)
			px, py := toXY(angle, planetR)

			// Overlap avoidance
			offset := 0.0
			labelStagger := 0
			for pi := range placedList {
				prev := &placedList[pi]
				diff := math.Abs(angle - prev.angle)
				if diff > 180 { diff = 360 - diff }
				if diff < 10 {
					offset += 18
					labelStagger = prev.count
					prev.count++
				}
			}
			if offset > 0 {
				px, py = toXY(angle, planetR+offset)
			}
			placedList = append(placedList, placed{angle: angle, count: 1})

			// Planet circle
			sb.WriteString(fmt.Sprintf(
				`<circle cx="%.1f" cy="%.1f" r="%.0f" fill="#ffffff" stroke="%s" stroke-width="2"/>`,
				px, py, circleR, color,
			))

			// Glyph
			sb.WriteString(fmt.Sprintf(
				`<text x="%.1f" y="%.1f" fill="%s" font-size="%d" text-anchor="middle" dominant-baseline="central" font-weight="bold">%s</text>`,
				px, py, color, glyphSize, p.Glyph,
			))

			// Degree label
			degInSign := math.Mod(p.Lon, 30)
			deg := int(degInSign)
			min := int(math.Round((degInSign - float64(deg)) * 60))
			signName := SignForLongitude(p.Lon)
			labelR := planetR - 25 - float64(labelStagger)*18
			dlx, dly := toXY(angle, labelR)
			sb.WriteString(fmt.Sprintf(
				`<text x="%.1f" y="%.1f" fill="%s" font-size="%d" text-anchor="middle" dominant-baseline="central" font-weight="bold">%d°%02d′%s</text>`,
				dlx, dly, color, labelSize, deg, min, signGlyphs[signName],
			))

			// Retrograde marker
			if p.Speed < 0 && p.Name != "Node" && p.Name != "SouthNode" {
				rxR := labelR - 10
				rx, ry := toXY(angle, rxR)
				sb.WriteString(fmt.Sprintf(
					`<text x="%.1f" y="%.1f" fill="%s" font-size="8" text-anchor="middle" dominant-baseline="central" font-weight="bold">℞</text>`,
					rx, ry, color,
				))
			}
		}
	}

	// ── Inner planets (natal) ──────────────────────────────────────────
	innerPlanetR := 195.0
	drawPlanets(innerPlanets, innerPlanetR, "#000000", false)

	// ── Outer planets (transit/secondary) ──────────────────────────────
	outerPlanetR := 370.0
	drawPlanets(outerPlanets, outerPlanetR, "#cc0000", true)

	// ── Aspect lines between inner and outer planets ───────────────────
	if opts.ShowAspects && opts.ShowOuterAspects {
		orb := opts.Orb
		if orb <= 0 { orb = 3.0 }

		for _, ip := range innerPlanets {
			for _, op := range outerPlanets {
				diff := math.Abs(ip.Lon - op.Lon)
				if diff > 180 { diff = 360 - diff }

				// Check all Ptolemaic aspects
				for _, aspAngle := range []float64{0, 60, 90, 120, 180} {
					if math.Abs(diff-aspAngle) <= orb {
						a1 := toAngle(ip.Lon)
						a2 := toAngle(op.Lon)
						x1, y1 := toXY(a1, innerPlanetR)
						x2, y2 := toXY(a2, outerPlanetR)

						// Color by aspect type
						aspColor := "#000000"
						opacity := "0.25"
						sw := "0.5"
						switch aspAngle {
						case 0:
							aspColor = "#cc0000"; opacity = "0.4"; sw = "1"
						case 90:
							aspColor = "#cc0000"; opacity = "0.3"
						case 180:
							aspColor = "#cc0000"; opacity = "0.35"; sw = "0.8"
						case 120:
							aspColor = "#006600"; opacity = "0.3"
						case 60:
							aspColor = "#006600"; opacity = "0.2"
						}

						sb.WriteString(fmt.Sprintf(
							`<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="%s" stroke-width="%s" opacity="%s"/>`,
							x1, y1, x2, y2, aspColor, sw, opacity,
						))
						break
					}
				}
			}
		}
	}

	// ── Inner aspect lines (within natal) ──────────────────────────────
	if opts.ShowAspects && opts.ShowInnerAspects {
		orb := opts.Orb
		if orb <= 0 { orb = 3.0 }
		for i := 0; i < len(innerPlanets); i++ {
			for j := i + 1; j < len(innerPlanets); j++ {
				diff := math.Abs(innerPlanets[i].Lon - innerPlanets[j].Lon)
				if diff > 180 { diff = 360 - diff }
				for _, aspAngle := range []float64{0, 90, 120, 180} {
					if math.Abs(diff-aspAngle) <= orb {
						a1 := toAngle(innerPlanets[i].Lon)
						a2 := toAngle(innerPlanets[j].Lon)
						x1, y1 := toXY(a1, innerPlanetR)
						x2, y2 := toXY(a2, innerPlanetR)
						sb.WriteString(fmt.Sprintf(
							`<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="#000000" stroke-width="0.5" opacity="0.2"/>`,
							x1, y1, x2, y2,
						))
						break
					}
				}
			}
		}
	}

	// ── Center label ───────────────────────────────────────────────────
	sb.WriteString(fmt.Sprintf(
		`<text x="%.1f" y="%.1f" fill="#000000" font-size="16" text-anchor="middle" dominant-baseline="central" font-weight="bold">%s</text>`,
		cx, cy-18, innerName,
	))
	sb.WriteString(fmt.Sprintf(
		`<text x="%.1f" y="%.1f" fill="#000000" font-size="14" text-anchor="middle" dominant-baseline="central">%d-%02d-%02d</text>`,
		cx, cy, innerYear, innerMonth, innerDay,
	))
	sb.WriteString(fmt.Sprintf(
		`<text x="%.1f" y="%.1f" fill="#cc0000" font-size="14" text-anchor="middle" dominant-baseline="central" font-weight="bold">%s</text>`,
		cx, cy+18, outerName,
	))
	sb.WriteString(fmt.Sprintf(
		`<text x="%.1f" y="%.1f" fill="#cc0000" font-size="12" text-anchor="middle" dominant-baseline="central">%d-%02d-%02d</text>`,
		cx, cy+34, outerYear, outerMonth, outerDay,
	))

	sb.WriteString(`</svg>`)
	return sb.String()
}
