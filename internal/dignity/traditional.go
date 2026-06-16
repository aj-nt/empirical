package dignity

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

// ── Traditional Astrology Interpretive Layer ──────────────────────────────
//
// Lunar phase, retrograde detection, antiscia, decans, terms,
// dispositor tree, mutual reception, void of course Moon.

// ── Lunar Phase ──────────────────────────────────────────────────────────

type LunarPhase struct {
	Name       string  `json:"name"`
	Angle      float64 `json:"angle_deg"`
	PhaseIndex int     `json:"phase_index"`
}

func ComputeLunarPhase(sunLon, moonLon float64) LunarPhase {
	angle := normalizeLon(moonLon - sunLon)
	var name string
	var idx int
	switch {
	case angle < 45:
		name = "New Moon"
		idx = 0
	case angle < 90:
		name = "Waxing Crescent"
		idx = 1
	case angle < 135:
		name = "First Quarter"
		idx = 2
	case angle < 180:
		name = "Waxing Gibbous"
		idx = 3
	case angle < 225:
		name = "Full Moon"
		idx = 4
	case angle < 270:
		name = "Waning Gibbous"
		idx = 5
	case angle < 315:
		name = "Last Quarter"
		idx = 6
	default:
		name = "Waning Crescent"
		idx = 7
	}
	return LunarPhase{Name: name, Angle: math.Round(angle*100) / 100, PhaseIndex: idx}
}

// ── Retrograde Detection ─────────────────────────────────────────────────

type RetrogradeInfo struct {
	Planet     string  `json:"planet"`
	Retrograde bool    `json:"retrograde"`
	Speed      float64 `json:"speed_deg_per_day"`
}

func DetectRetrogrades(speeds map[string]float64) []RetrogradeInfo {
	order := []string{"Sun", "Moon", "Mercury", "Venus", "Mars", "Jupiter", "Saturn",
		"Uranus", "Neptune", "Pluto", "Chiron", "Ceres", "Pallas", "Juno", "Vesta",
		"Cupido", "Hades", "Zeus", "Kronos", "Apollon", "Admetos", "Vulcanus", "Poseidon"}
	var result []RetrogradeInfo
	for _, p := range order {
		if speed, ok := speeds[p]; ok {
			result = append(result, RetrogradeInfo{
				Planet:     p,
				Retrograde: speed < 0,
				Speed:      math.Round(speed*10000) / 10000,
			})
		}
	}
	return result
}

// ── Antiscia ─────────────────────────────────────────────────────────────

type AntiscionPoint struct {
	Planet          string  `json:"planet"`
	Longitude       float64 `json:"longitude"`
	Antiscion       float64 `json:"antiscion"`
	AntiscionSign   string  `json:"antiscion_sign"`
	ContraAntiscion float64 `json:"contra_antiscion"`
	ContraSign      string  `json:"contra_antiscion_sign"`
}

func ComputeAntiscia(positions map[string]float64) []AntiscionPoint {
	order := []string{"Sun", "Moon", "Mercury", "Venus", "Mars", "Jupiter", "Saturn",
		"Uranus", "Neptune", "Pluto", "Node", "Chiron", "Lilith",
		"Ceres", "Pallas", "Juno", "Vesta",
		"Cupido", "Hades", "Zeus", "Kronos", "Apollon", "Admetos", "Vulcanus", "Poseidon"}
	var result []AntiscionPoint
	for _, p := range order {
		if lon, ok := positions[p]; ok {
			anti := normalizeLon(360 - lon)
			contra := normalizeLon(anti + 180)
			result = append(result, AntiscionPoint{
				Planet:          p,
				Longitude:       math.Round(lon*100) / 100,
				Antiscion:       math.Round(anti*100) / 100,
				AntiscionSign:   SignForLongitude(anti),
				ContraAntiscion: math.Round(contra*100) / 100,
				ContraSign:      SignForLongitude(contra),
			})
		}
	}
	return result
}

// ── Decans (Faces) ───────────────────────────────────────────────────────

var chaldeanOrder = []string{"Mars", "Sun", "Venus", "Mercury", "Moon", "Saturn", "Jupiter"}
var signOrder = []string{
	"Aries", "Taurus", "Gemini", "Cancer", "Leo", "Virgo",
	"Libra", "Scorpio", "Sagittarius", "Capricorn", "Aquarius", "Pisces",
}

type DecanInfo struct {
	Planet string `json:"planet"`
	Sign   string `json:"sign"`
	Decan  int    `json:"decan"`
	Ruler  string `json:"ruler"`
}

func ComputeDecans(positions map[string]float64) []DecanInfo {
	order := []string{"Sun", "Moon", "Mercury", "Venus", "Mars", "Jupiter", "Saturn",
		"Uranus", "Neptune", "Pluto", "Node", "Chiron", "Lilith"}
	var result []DecanInfo
	for _, p := range order {
		if lon, ok := positions[p]; ok {
			sign := SignForLongitude(lon)
			degInSign := lon - signStart(sign)
			if degInSign < 0 {
				degInSign += 30
			}
			decan := int(degInSign/10) + 1
			signIdx := -1
			for i, s := range signOrder {
				if s == sign {
					signIdx = i
					break
				}
			}
			rulerIdx := (signIdx*3 + (decan - 1)) % 7
			ruler := chaldeanOrder[rulerIdx]
			result = append(result, DecanInfo{Planet: p, Sign: sign, Decan: decan, Ruler: ruler})
		}
	}
	return result
}

func signStart(sign string) float64 {
	for i, s := range signOrder {
		if s == sign {
			return float64(i * 30)
		}
	}
	return 0
}

// ── Egyptian Terms ───────────────────────────────────────────────────────

type TermBound struct {
	Ruler  string  `json:"ruler"`
	EndDeg float64 `json:"end_degree"`
}

var egyptianTerms = map[string][]TermBound{
	"Aries":       {{"Jupiter", 6}, {"Venus", 12}, {"Mercury", 20}, {"Mars", 25}, {"Saturn", 30}},
	"Taurus":      {{"Venus", 8}, {"Mercury", 14}, {"Jupiter", 22}, {"Saturn", 27}, {"Mars", 30}},
	"Gemini":      {{"Mercury", 6}, {"Jupiter", 12}, {"Venus", 17}, {"Mars", 24}, {"Saturn", 30}},
	"Cancer":      {{"Mars", 7}, {"Venus", 13}, {"Mercury", 19}, {"Jupiter", 26}, {"Saturn", 30}},
	"Leo":         {{"Jupiter", 6}, {"Venus", 11}, {"Saturn", 18}, {"Mercury", 24}, {"Mars", 30}},
	"Virgo":       {{"Mercury", 7}, {"Venus", 17}, {"Jupiter", 21}, {"Mars", 28}, {"Saturn", 30}},
	"Libra":       {{"Saturn", 6}, {"Mercury", 14}, {"Jupiter", 21}, {"Venus", 28}, {"Mars", 30}},
	"Scorpio":     {{"Mars", 7}, {"Venus", 11}, {"Mercury", 19}, {"Jupiter", 24}, {"Saturn", 30}},
	"Sagittarius": {{"Jupiter", 12}, {"Venus", 17}, {"Mercury", 21}, {"Saturn", 26}, {"Mars", 30}},
	"Capricorn":   {{"Mercury", 7}, {"Jupiter", 14}, {"Venus", 22}, {"Saturn", 26}, {"Mars", 30}},
	"Aquarius":    {{"Mercury", 7}, {"Venus", 13}, {"Jupiter", 20}, {"Mars", 25}, {"Saturn", 30}},
	"Pisces":      {{"Venus", 12}, {"Jupiter", 16}, {"Mercury", 19}, {"Mars", 28}, {"Saturn", 30}},
}

type TermInfo struct {
	Planet string `json:"planet"`
	Sign   string `json:"sign"`
	Term   int    `json:"term"`
	Ruler  string `json:"ruler"`
}

func ComputeTerms(positions map[string]float64) []TermInfo {
	order := []string{"Sun", "Moon", "Mercury", "Venus", "Mars", "Jupiter", "Saturn"}
	var result []TermInfo
	for _, p := range order {
		if lon, ok := positions[p]; ok {
			sign := SignForLongitude(lon)
			degInSign := lon - signStart(sign)
			if degInSign < 0 {
				degInSign += 30
			}
			terms, ok := egyptianTerms[sign]
			if !ok {
				continue
			}
			termNum := 1
			termRuler := ""
			for _, t := range terms {
				if degInSign < t.EndDeg {
					termRuler = t.Ruler
					break
				}
				termNum++
			}
			if termRuler == "" {
				termRuler = terms[len(terms)-1].Ruler
				termNum = len(terms)
			}
			result = append(result, TermInfo{Planet: p, Sign: sign, Term: termNum, Ruler: termRuler})
		}
	}
	return result
}

// ── Dispositor Tree ──────────────────────────────────────────────────────

var DomicileRulers = map[string]string{
	"Aries": "Mars", "Taurus": "Venus", "Gemini": "Mercury",
	"Cancer": "Moon", "Leo": "Sun", "Virgo": "Mercury",
	"Libra": "Venus", "Scorpio": "Mars", "Sagittarius": "Jupiter",
	"Capricorn": "Saturn", "Aquarius": "Saturn", "Pisces": "Jupiter",
}

type DispositorNode struct {
	Planet     string `json:"planet"`
	Sign       string `json:"sign"`
	Dispositor string `json:"dispositor"`
	IsFinal    bool   `json:"is_final"`
	InLoop     bool   `json:"in_loop"`
	LoopWith   string `json:"loop_with,omitempty"`
}

type DispositorTree struct {
	Nodes            []DispositorNode `json:"nodes"`
	FinalDispositors []string         `json:"final_dispositors"`
	MutualReceptions []string         `json:"mutual_receptions"`
}

func ComputeDispositorTree(positions map[string]float64) DispositorTree {
	classical := []string{"Sun", "Moon", "Mercury", "Venus", "Mars", "Jupiter", "Saturn"}
	planetSign := make(map[string]string)
	for _, p := range classical {
		if lon, ok := positions[p]; ok {
			planetSign[p] = SignForLongitude(lon)
		}
	}
	dispositor := make(map[string]string)
	for p, sign := range planetSign {
		if ruler, ok := DomicileRulers[sign]; ok {
			dispositor[p] = ruler
		}
	}
	final := make(map[string]bool)
	loops := make(map[string]string)
	for _, p := range classical {
		if _, ok := dispositor[p]; !ok {
			continue
		}
		visited := make(map[string]bool)
		current := p
		for {
			if visited[current] {
				for other := range visited {
					if other != current && dispositor[other] == current {
						loops[current] = other
						loops[other] = current
					}
				}
				break
			}
			visited[current] = true
			next, ok := dispositor[current]
			if !ok {
				break
			}
			if next == current {
				final[current] = true
				break
			}
			current = next
		}
	}
	var nodes []DispositorNode
	for _, p := range classical {
		sign, ok := planetSign[p]
		if !ok {
			continue
		}
		disp, ok := dispositor[p]
		if !ok {
			continue
		}
		node := DispositorNode{Planet: p, Sign: sign, Dispositor: disp, IsFinal: final[p], InLoop: loops[p] != ""}
		if loops[p] != "" {
			node.LoopWith = loops[p]
		}
		nodes = append(nodes, node)
	}
	var finalList []string
	for p := range final {
		finalList = append(finalList, p)
	}
	sort.Strings(finalList)
	seen := make(map[string]bool)
	var receptionList []string
	for p1, p2 := range loops {
		pair := []string{p1, p2}
		sort.Strings(pair)
		key := pair[0] + "-" + pair[1]
		if !seen[key] {
			seen[key] = true
			receptionList = append(receptionList, key)
		}
	}
	sort.Strings(receptionList)
	return DispositorTree{Nodes: nodes, FinalDispositors: finalList, MutualReceptions: receptionList}
}

// ── Void of Course Moon ──────────────────────────────────────────────────

type VOCMoon struct {
	VOC           bool    `json:"void_of_course"`
	MoonSign      string  `json:"moon_sign"`
	MoonLon       float64 `json:"moon_lon"`
	LastAspect    string  `json:"last_aspect,omitempty"`
	LastAspectTo  string  `json:"last_aspect_to,omitempty"`
	LastAspectOrb float64 `json:"last_aspect_orb,omitempty"`
	NextSign      string  `json:"next_sign"`
	DegreesToNext float64 `json:"degrees_to_next_sign"`
}

func ComputeVOCMoon(positions map[string]float64) VOCMoon {
	moonLon, ok := positions["Moon"]
	if !ok {
		return VOCMoon{}
	}
	moonSign := SignForLongitude(moonLon)
	nextSignStart := signStart(moonSign) + 30
	if nextSignStart >= 360 {
		nextSignStart -= 360
	}
	degRemaining := nextSignStart - moonLon
	if degRemaining < 0 {
		degRemaining += 360
	}
	signIdx := -1
	for i, s := range signOrder {
		if s == moonSign {
			signIdx = i
			break
		}
	}
	nextSign := signOrder[(signIdx+1)%12]
	moonSpeed := 13.2
	aspects := DefaultAspects()
	orb := 3.0
	checkPlanets := []string{"Sun", "Mercury", "Venus", "Mars", "Jupiter", "Saturn",
		"Uranus", "Neptune", "Pluto"}
	var lastAspect string
	var lastAspectTo string
	var lastAspectOrb float64
	hasApplying := false
	for _, planet := range checkPlanets {
		plon, ok := positions[planet]
		if !ok {
			continue
		}
		dist := angleDist(moonLon, plon)
		for _, a := range aspects {
			diff := math.Abs(dist - a.Angle)
			if diff <= orb {
				planetAhead := normalizeLon(plon - moonLon)
				if planetAhead < 180 && planetAhead > 0 {
					timeToClose := planetAhead / moonSpeed
					moonTravel := timeToClose * moonSpeed
					if moonTravel <= degRemaining {
						hasApplying = true
						lastAspect = a.Name
						lastAspectTo = planet
						lastAspectOrb = math.Round(diff*100) / 100
					}
				}
			}
		}
	}
	return VOCMoon{
		VOC:           !hasApplying,
		MoonSign:      moonSign,
		MoonLon:       math.Round(moonLon*100) / 100,
		LastAspect:    lastAspect,
		LastAspectTo:  lastAspectTo,
		LastAspectOrb: lastAspectOrb,
		NextSign:      nextSign,
		DegreesToNext: math.Round(degRemaining*100) / 100,
	}
}

// ── Traditional Report ───────────────────────────────────────────────────

type TraditionalReport struct {
	Name           string           `json:"name"`
	LunarPhase     LunarPhase       `json:"lunar_phase"`
	Retrogrades    []RetrogradeInfo `json:"retrogrades"`
	Antiscia       []AntiscionPoint `json:"antiscia"`
	Decans         []DecanInfo      `json:"decans"`
	Terms          []TermInfo       `json:"terms"`
	DispositorTree DispositorTree   `json:"dispositor_tree"`
	VOCMoon        VOCMoon          `json:"void_of_course_moon"`
}

func ComputeTraditionalReport(name string, positions map[string]float64, speeds map[string]float64) TraditionalReport {
	return TraditionalReport{
		Name:           name,
		LunarPhase:     ComputeLunarPhase(positions["Sun"], positions["Moon"]),
		Retrogrades:    DetectRetrogrades(speeds),
		Antiscia:       ComputeAntiscia(positions),
		Decans:         ComputeDecans(positions),
		Terms:          ComputeTerms(positions),
		DispositorTree: ComputeDispositorTree(positions),
		VOCMoon:        ComputeVOCMoon(positions),
	}
}

func FormatTraditionalReport(tr TraditionalReport) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Traditional Astrology Report — %s\n\n", tr.Name))
	b.WriteString("── Lunar Phase ──\n")
	b.WriteString(fmt.Sprintf("  %s (%.2f° Sun-Moon angle)\n\n", tr.LunarPhase.Name, tr.LunarPhase.Angle))
	b.WriteString("── Retrogrades ──\n")
	hasRx := false
	for _, r := range tr.Retrogrades {
		if r.Retrograde {
			b.WriteString(fmt.Sprintf("  %s Rx (%.4f deg/day)\n", r.Planet, r.Speed))
			hasRx = true
		}
	}
	if !hasRx {
		b.WriteString("  No planets retrograde\n")
	}
	b.WriteString("\n── Dispositor Tree ──\n")
	for _, n := range tr.DispositorTree.Nodes {
		marker := ""
		if n.IsFinal {
			marker = " ★ FINAL DISPOSITOR"
		} else if n.InLoop {
			marker = fmt.Sprintf(" ⟲ mutual reception with %s", n.LoopWith)
		}
		b.WriteString(fmt.Sprintf("  %s in %s → %s%s\n", n.Planet, n.Sign, n.Dispositor, marker))
	}
	if len(tr.DispositorTree.MutualReceptions) > 0 {
		b.WriteString(fmt.Sprintf("  Mutual receptions: %s\n", strings.Join(tr.DispositorTree.MutualReceptions, ", ")))
	}
	b.WriteString("\n── Decans (Faces) ──\n")
	for _, d := range tr.Decans {
		b.WriteString(fmt.Sprintf("  %s: %s decan %d (ruler: %s)\n", d.Planet, d.Sign, d.Decan, d.Ruler))
	}
	b.WriteString("\n── Egyptian Terms ──\n")
	for _, t := range tr.Terms {
		b.WriteString(fmt.Sprintf("  %s: %s term %d (ruler: %s)\n", t.Planet, t.Sign, t.Term, t.Ruler))
	}
	b.WriteString("\n── Antiscia ──\n")
	for _, a := range tr.Antiscia {
		b.WriteString(fmt.Sprintf("  %s (%.2f° %s) → antiscion %.2f° %s, contra %.2f° %s\n",
			a.Planet, a.Longitude, SignForLongitude(a.Longitude),
			a.Antiscion, a.AntiscionSign,
			a.ContraAntiscion, a.ContraSign))
	}
	b.WriteString("\n── Void of Course Moon ──\n")
	if tr.VOCMoon.VOC {
		b.WriteString(fmt.Sprintf("  Moon is VOID OF COURSE in %s (%.2f°)\n", tr.VOCMoon.MoonSign, tr.VOCMoon.MoonLon))
		b.WriteString(fmt.Sprintf("  %.2f° remaining until %s\n", tr.VOCMoon.DegreesToNext, tr.VOCMoon.NextSign))
	} else {
		b.WriteString(fmt.Sprintf("  Moon is NOT void of course in %s\n", tr.VOCMoon.MoonSign))
		b.WriteString(fmt.Sprintf("  Last applying aspect: %s to %s (orb %.2f°)\n",
			tr.VOCMoon.LastAspect, tr.VOCMoon.LastAspectTo, tr.VOCMoon.LastAspectOrb))
	}
	return b.String()
}
