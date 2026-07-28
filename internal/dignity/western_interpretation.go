package dignity

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ═══════════════════════════════════════════════════════════════════════
// Western Interpretation Engine
// ═══════════════════════════════════════════════════════════════════════
//
// Natural-language descriptions of chart features. Template-based,
// deterministic, no LLM dependency. Each function maps astrological
// data to a concise English description.

// AspectHit is a single aspect between two planets.
type AspectHit struct {
	Planet1 string  `json:"planet1"`
	Planet2 string  `json:"planet2"`
	Aspect  string  `json:"aspect"`
	Orb     float64 `json:"orb"`
}

// PatternHit is a detected aspect pattern.
type PatternHit struct {
	Name    string   `json:"name"`
	Planets []string `json:"planets"`
}

// WeightedAspect is an aspect ranked by importance (orb × aspect type × planet weight).
type WeightedAspect struct {
	Planet1 string  `json:"planet1"`
	Planet2 string  `json:"planet2"`
	Aspect  string  `json:"aspect"`
	Orb     float64 `json:"orb"`
	Weight  float64 `json:"weight"`
}

// ChartInterpretation holds the full interpretation of a chart.
type ChartInterpretation struct {
	Name            string         `json:"name"`
	PlanetSigns     []string       `json:"planet_signs"`
	PlanetHouses    []string       `json:"planet_houses"`
	Aspects         []string       `json:"aspects"`
	Patterns        []string       `json:"patterns"`
	Stars           []string       `json:"stars,omitempty"`
	Midpoints       []string       `json:"midpoints,omitempty"`
	Declinations    []string       `json:"declinations,omitempty"`
	Contraparallels []string       `json:"contraparallels,omitempty"`
	ElementBalance  map[string]int     `json:"element_balance,omitempty"`
	ModalityBalance map[string]int     `json:"modality_balance,omitempty"`
	Hemisphere       *HemisphereEmphasis `json:"hemisphere,omitempty"`
	RulershipChains  map[int][]string    `json:"rulership_chains,omitempty"`
	DispositorTrees  map[string][]string `json:"dispositor_trees,omitempty"`
	IsDay           bool           `json:"is_day"`

	// ── Traditional Western fields ──
	LunarPhase        string   `json:"lunar_phase,omitempty"`
	LunarPhaseAngle   float64  `json:"lunar_phase_angle,omitempty"`
	Retrogrades       []string `json:"retrogrades,omitempty"`
	Antiscia          []string `json:"antiscia,omitempty"`
	AntisciaContacts  []string `json:"antiscia_contacts,omitempty"`
	MutualReceptions  []string `json:"mutual_receptions,omitempty"`
	Decans            []string `json:"decans,omitempty"`
	Terms             []string `json:"terms,omitempty"`
	VOCMoon           string   `json:"voc_moon,omitempty"`
	Sect              string   `json:"sect,omitempty"`

	// ── Reading-optimized fields (populated when reading=true) ──
	ChartRuler             string          `json:"chart_ruler,omitempty"`
	ChartRulerTraditional  string          `json:"chart_ruler_traditional,omitempty"`
	ChartRulerHouse        int             `json:"chart_ruler_house,omitempty"`
	ChartRulerSign         string          `json:"chart_ruler_sign,omitempty"`
	ChartRulerDignity      string          `json:"chart_ruler_dignity,omitempty"`
	FinalDispositor            string      `json:"final_dispositor,omitempty"`
	FinalDispositorTraditional string      `json:"final_dispositor_traditional,omitempty"`
	WeightedAspects            []WeightedAspect `json:"weighted_aspects,omitempty"`
	KeyMidpoints               []string     `json:"key_midpoints,omitempty"`
	KeyStarAspects             []string     `json:"key_star_aspects,omitempty"`
	AngularPlanets             []string     `json:"angular_planets,omitempty"`
}

// JSON returns the interpretation as JSON bytes.
func (c *ChartInterpretation) JSON() ([]byte, error) {
	return json.Marshal(c)
}

// ── Planet dignity tables ─────────────────────────────────────────────

var domicile = map[string]string{
	"Sun": "Leo", "Moon": "Cancer", "Mercury": "Gemini,Virgo",
	"Venus": "Taurus,Libra", "Mars": "Aries,Scorpio",
	"Jupiter": "Sagittarius,Pisces", "Saturn": "Capricorn,Aquarius",
	"Uranus": "Aquarius", "Neptune": "Pisces", "Pluto": "Scorpio",
	"Chiron": "Sagittarius", "Ceres": "Virgo", "Pallas": "Libra",
	"Juno": "Scorpio", "Vesta": "Virgo",
}

var detriment = map[string]string{
	"Sun": "Aquarius", "Moon": "Capricorn", "Mercury": "Sagittarius,Pisces",
	"Venus": "Aries,Scorpio", "Mars": "Taurus,Libra",
	"Jupiter": "Gemini,Virgo", "Saturn": "Cancer,Leo",
	"Uranus": "Leo", "Neptune": "Virgo", "Pluto": "Taurus",
}

var exaltation = map[string]string{
	"Sun": "Aries", "Moon": "Taurus", "Mercury": "Virgo",
	"Venus": "Pisces", "Mars": "Capricorn", "Jupiter": "Cancer",
	"Saturn": "Libra", "Uranus": "Scorpio", "Neptune": "Cancer",
	"Pluto": "Leo",
}

var fall = map[string]string{
	"Sun": "Libra", "Moon": "Scorpio", "Mercury": "Pisces",
	"Venus": "Virgo", "Mars": "Cancer", "Jupiter": "Capricorn",
	"Saturn": "Aries", "Uranus": "Taurus", "Neptune": "Capricorn",
	"Pluto": "Aquarius",
}

// ── Planet descriptions ───────────────────────────────────────────────

var planetDescriptions = map[string]string{
	"Sun":     "core identity, vitality, conscious self",
	"Moon":    "emotional nature, instincts, inner world",
	"Mercury": "communication, intellect, how you think and speak",
	"Venus":   "love, values, aesthetics, what you attract",
	"Mars":    "drive, assertion, how you fight and pursue",
	"Jupiter": "expansion, faith, where you grow and find meaning",
	"Saturn":  "structure, discipline, where you face limits and build mastery",
	"Uranus":  "innovation, rebellion, where you break patterns",
	"Neptune": "imagination, dissolution, where you transcend or escape",
	"Pluto":   "power, transformation, where you destroy and regenerate",
	"Chiron":  "wounding and healing, the teacher from pain",
	"Node":    "evolutionary path, what you're moving toward",
	"SouthNode": "past mastery, what you're releasing, the diminishment point",
	"Ceres":   "nurturing, cycles of care and loss",
	"Pallas":  "pattern recognition, strategic intelligence",
	"Juno":    "partnership contracts, what you commit to",
	"Vesta":   "devotion, sacred focus, what you tend",
	"Lilith":  "shadow feminine, what was suppressed",
}

// ── Sign descriptions ─────────────────────────────────────────────────

var signDescriptions = map[string]string{
	"Aries":       "cardinal fire — initiatory, direct, competitive",
	"Taurus":      "fixed earth — steady, sensual, accumulating",
	"Gemini":      "mutable air — curious, dual, networking",
	"Cancer":      "cardinal water — protective, nurturing, cyclical",
	"Leo":         "fixed fire — radiant, creative, commanding",
	"Virgo":       "mutable earth — analytical, refining, serving",
	"Libra":       "cardinal air — balancing, relating, aesthetic",
	"Scorpio":     "fixed water — penetrating, transformative, intense",
	"Sagittarius": "mutable fire — seeking, philosophical, expansive",
	"Capricorn":   "cardinal earth — ambitious, structuring, enduring",
	"Aquarius":    "fixed air — innovative, collective, detached",
	"Pisces":      "mutable water — dissolving, transcendent, compassionate",
}

// ── House descriptions ─────────────────────────────────────────────────

var houseDescriptions = map[int]string{
	1:  "self, persona, physical body, how you show up",
	2:  "resources, values, self-worth, what you own",
	3:  "communication, siblings, local environment, learning",
	4:  "home, roots, family, private self, foundation",
	5:  "creativity, pleasure, children, self-expression",
	6:  "work, health, service, daily routines",
	7:  "partnership, marriage, open enemies, the other",
	8:  "shared resources, transformation, intimacy, death",
	9:  "expansion, travel, philosophy, higher learning",
	10: "career, public role, reputation, authority",
	11: "community, networks, hopes, collective future",
	12: "retreat, unconscious, hidden things, dissolution",
}

// ── Aspect descriptions ───────────────────────────────────────────────

var aspectDescriptions = map[string]string{
	"conjunction":    "merge and amplify — the two planets operate as one force",
	"opposition":     "polarity and tension — a seesaw between two extremes",
	"trine":          "flow and ease — natural harmony, talent that comes without effort",
	"square":         "friction and growth — conflict that forces development",
	"sextile":        "opportunity — a door that opens when you walk through it",
	"quincunx":       "adjustment — two energies that don't understand each other",
	"semi-sextile":   "adjacent signs — a subtle nudge, two energies that barely recognize each other",
	"semi-square":    "irritation — a low-grade friction that wears down or sharpens over time",
	"sesquiquadrate": "agitation — a square-and-a-half, restless tension that demands realignment",
}

// ── Pattern descriptions ──────────────────────────────────────────────

var patternDescriptions = map[string]string{
	"T-Square":    "dynamic tension between three planets — a pressure cooker that produces results",
	"Grand Trine": "closed loop of flowing trines — effortless talent that can become inertia",
	"Yod":         "finger of god — two planets in sextile both quincunx a focal planet, creating a fated pressure point",
	"Grand Cross": "four planets in mutual squares — constant tension, extraordinary resilience",
	"Cradle":      "a bowl of support — sextiles and trines forming a nurturing structure",
	"Kite":        "a Grand Trine with an opposition — talent with a release valve",
	"Stellium":    "concentration of three or more planets in one sign or house — intense focus",
}

// ── Planet pair dynamics ──────────────────────────────────────────────

var pairDynamics = map[string]string{
	"Sun,Moon":       "conscious will meets emotional need — the fundamental axis of personality",
	"Sun,Mercury":    "identity and expression are fused — you are what you say",
	"Sun,Venus":      "self meets values — what you love defines who you are",
	"Sun,Mars":       "will meets drive — identity expressed through action",
	"Sun,Jupiter":    "self meets expansion — confidence and reach amplify each other",
	"Sun,Saturn":     "will meets limit — identity forged through discipline and constraint",
	"Moon,Mercury":   "emotion meets thought — feelings shape your words",
	"Moon,Venus":     "need meets love — emotional security through relationship",
	"Moon,Mars":      "instinct meets action — reactive, passionate, protective",
	"Moon,Saturn":    "feeling meets structure — emotional containment, maturity through restraint",
	"Mercury,Mars":   "thought meets fire — sharp words, quick decisions, verbal combat",
	"Mercury,Saturn": "mind meets discipline — precise thinking, slow but thorough",
	"Venus,Mars":     "attraction meets pursuit — the classic love-and-desire axis",
	"Venus,Saturn":   "love meets limit — commitment, seriousness, values under audit",
	"Venus,Jupiter":  "love meets expansion — generosity, abundance, pleasure amplified",
	"Mars,Saturn":    "drive meets wall — frustration that builds strength over time",
	"Mars,Jupiter":   "action meets faith — bold moves, risk-taking, crusading energy",
	"Mars,Pluto":     "force meets power — volcanic, transformative, potentially destructive",
	"Jupiter,Saturn": "expansion meets contraction — the rhythm of growth and consolidation",
	"Saturn,Uranus":  "tradition meets revolution — old structures vs. new paradigms",
	"Saturn,Neptune": "structure meets dissolution — the tension between form and formlessness",
	"Saturn,Pluto":   "limit meets power — deep structural transformation, slow and irreversible",
	"Uranus,Neptune": "breakthrough meets transcendence — generational shifts in consciousness",
	"Uranus,Pluto":   "revolution meets power — systemic upheaval, creative destruction",
	"Neptune,Pluto":  "dissolution meets transformation — the deepest generational currents",
}

// ── Interpretation functions ──────────────────────────────────────────

// InterpretPlanetInSign returns a natural-language description of a planet in a sign.
func InterpretPlanetInSign(planet, sign string) string {
	pd, ok := planetDescriptions[planet]
	if !ok {
		pd = strings.ToLower(planet)
	}
	sd, ok := signDescriptions[sign]
	if !ok {
		sd = strings.ToLower(sign)
	}

	var dignity string
	if strings.Contains(domicile[planet], sign) {
		dignity = "in domicile — at home, operating at full strength"
	} else if strings.Contains(detriment[planet], sign) {
		dignity = "in detriment — out of element, operating against its nature"
	} else if exaltation[planet] == sign {
		dignity = "exalted — elevated, operating with exceptional clarity"
	} else if fall[planet] == sign {
		dignity = "in fall — diminished, operating with difficulty"
	} else {
		dignity = "neutral — neither strengthened nor weakened by this sign"
	}

	return fmt.Sprintf("%s in %s: %s, %s. %s.", planet, sign, pd, sd, dignity)
}

// InterpretPlanetInHouse returns a natural-language description of a planet in a house.
func InterpretPlanetInHouse(planet string, house int) string {
	if planetMap, ok := planetHouseMeanings[planet]; ok {
		if meaning, ok := planetMap[house]; ok {
			return meaning
		}
	}

	pd, ok := planetDescriptions[planet]
	if !ok {
		pd = strings.ToLower(planet)
	}
	hd, ok := houseDescriptions[house]
	if !ok {
		hd = fmt.Sprintf("house %d", house)
	}

	return fmt.Sprintf("%s in house %d: %s expressed through %s.", planet, house, pd, hd)
}

// InterpretAspect returns a natural-language description of an aspect between two planets.
func InterpretAspect(planet1, planet2, aspect string, orb float64) string {
	ad, ok := aspectDescriptions[aspect]
	if !ok {
		ad = fmt.Sprintf("%s aspect", aspect)
	}

	key1 := planet1 + "," + planet2
	key2 := planet2 + "," + planet1
	dynamics, ok := pairDynamics[key1]
	if !ok {
		dynamics, ok = pairDynamics[key2]
	}
	if !ok {
		dynamics = fmt.Sprintf("%s and %s in contact", planet1, planet2)
	}

	return fmt.Sprintf("%s %s %s (orb %.1f°): %s — %s.",
		planet1, aspect, planet2, orb, dynamics, ad)
}

// InterpretPattern returns a natural-language description of an aspect pattern.
func InterpretPattern(name string, planets []string) string {
	pd, ok := patternDescriptions[name]
	if !ok {
		pd = fmt.Sprintf("a %s configuration", name)
	}

	planetList := strings.Join(planets, ", ")
	return fmt.Sprintf("%s involving %s: %s.", name, planetList, pd)
}

// InterpretStarConjunction returns a natural-language description of a fixed star conjunct a planet.
func InterpretStarConjunction(sc StarConjunction) string {
	pd, ok := planetDescriptions[sc.Planet]
	if !ok {
		pd = strings.ToLower(sc.Planet)
	}

	sign := SignForLongitude(sc.PlanetLon)

	return fmt.Sprintf(
		"%s conjunct %s at %.2f° %s (orb %.2f°): %s touches %s — %s.",
		sc.Star, sc.Planet, sc.PlanetLon, sign, sc.Orb,
		sc.Star, pd, sc.Meaning,
	)
}

// InterpretChart produces a full chart interpretation.
func InterpretChart(
	name string,
	planets map[string]float64,
	houses map[string]int,
	aspects []AspectHit,
	patterns []PatternHit,
	dignities []PlanetDignity,
) *ChartInterpretation {
	report := &ChartInterpretation{
		Name:         name,
		PlanetSigns:  make([]string, 0),
		PlanetHouses: make([]string, 0),
		Aspects:      make([]string, 0),
		Patterns:     make([]string, 0),
	}

	for planet, lon := range planets {
		sign := SignForLongitude(lon)
		report.PlanetSigns = append(report.PlanetSigns,
			InterpretPlanetInSign(planet, sign))
	}

	for planet, house := range houses {
		report.PlanetHouses = append(report.PlanetHouses,
			InterpretPlanetInHouse(planet, house))
	}

	for _, a := range aspects {
		report.Aspects = append(report.Aspects,
			InterpretAspect(a.Planet1, a.Planet2, a.Aspect, a.Orb))
	}

	for _, p := range patterns {
		report.Patterns = append(report.Patterns,
			InterpretPattern(p.Name, p.Planets))
	}

	return report
}
