package dignity

import (
	"fmt"
	"strings"
)

// ═══════════════════════════════════════════════════════════════════════
// Koiné (Hellenistic) Interpretation Engine
// ═══════════════════════════════════════════════════════════════════════
//
// Natural-language descriptions of chart features using Hellenistic
// source texts (Valens, Ptolemy, Dorotheus). Classical planets only,
// Ptolemaic aspects, no modern patterns, sect-based dignity.

// ── Hellenistic planet descriptions ────────────────────────────────────

var koinePlanetDescriptions = map[string]string{
	"Sun":     "the sect light of day — vitality, authority, the father, the soul's rational faculty",
	"Moon":    "the sect light of night — the body, fortune, the mother, change and flux",
	"Mercury": "the common star — reason, speech, commerce, youth, education",
	"Venus":   "the benefic of the night sect — love, beauty, pleasure, harmony, desire",
	"Mars":    "the malefic of the night sect — war, severing, conflict, courage, inflammation",
	"Jupiter": "the benefic of the day sect — fortune, abundance, law, children, generosity",
	"Saturn":  "the malefic of the day sect — constraint, time, death, discipline, foundations",
}

// ── Hellenistic sign descriptions ──────────────────────────────────────

var koineSignDescriptions = map[string]string{
	"Aries":       "the house of Mars — equinoctial, cardinal, masculine, fiery, commanding",
	"Taurus":      "the house of Venus — solid, fixed, feminine, earthy, mute",
	"Gemini":      "the house of Mercury — bicorporeal, mutable, masculine, airy, human",
	"Cancer":      "the house of the Moon — solstitial, cardinal, feminine, watery, fertile",
	"Leo":         "the house of the Sun — solid, fixed, masculine, fiery, commanding",
	"Virgo":       "the house of Mercury — bicorporeal, mutable, feminine, earthy, human",
	"Libra":       "the house of Venus — equinoctial, cardinal, masculine, airy, human",
	"Scorpio":     "the house of Mars — solid, fixed, feminine, watery, mute",
	"Sagittarius": "the house of Jupiter — bicorporeal, mutable, masculine, fiery, human",
	"Capricorn":   "the house of Saturn — solstitial, cardinal, feminine, earthy, mute",
	"Aquarius":    "the house of Saturn — solid, fixed, masculine, airy, human",
	"Pisces":      "the house of Jupiter — bicorporeal, mutable, feminine, watery, mute",
}

// ── Hellenistic house descriptions (from Valens, Ptolemy) ──────────────

var koineHouseDescriptions = map[int]string{
	1:  "the helm — life, breath, the body, character, the native's steering of their own vessel",
	2:  "the gate of Hades — livelihood, possessions, what sustains the body, profit and loss",
	3:  "the goddess — siblings, travel, friends abroad, the Moon's joy, short journeys",
	4:  "the subterranean — parents, foundations, patrimony, hidden treasure, the end of life",
	5:  "good fortune — children, creativity, beneficence, Venus's joy, gifts given and received",
	6:  "bad fortune — illness, injury, slaves, toil, enemies, Mars's joy, the body's afflictions",
	7:  "the setting — marriage, partnership, open enemies, the spouse, death's approach",
	8:  "idle — death, inheritance, the partner's resources, fear, loss, the idle place",
	9:  "the god — travel abroad, divination, philosophy, the Sun's joy, foreign lands",
	10: "the midheaven — action, occupation, reputation, authority, what one does in the world",
	11: "good spirit — friends, hopes, benefactors, Jupiter's joy, gifts from fortune",
	12: "bad spirit — enemies, suffering, exile, imprisonment, Saturn's joy, hidden afflictions",
}

// ── Hellenistic aspect descriptions ────────────────────────────────────

var koineAspectDescriptions = map[string]string{
	"conjunction": "co-presence — the two stars are in the same place, their natures blend",
	"opposition":  "diametrical — the stars face each other across the circle, a contest of natures",
	"trine":       "trigon — the stars behold each other from signs of the same element, harmony",
	"square":      "tetragon — the stars are at right angles, friction that produces action",
	"sextile":     "hexagon — the stars behold each other from signs of like sect, opportunity",
}

// ── Hellenistic pair dynamics ──────────────────────────────────────────

var koinePairDynamics = map[string]string{
	"Sun,Moon":       "the lights in contact — the rational soul and the body, day and night natures",
	"Sun,Mercury":    "the king and the messenger — authority expressed through speech and reason",
	"Sun,Venus":      "the day star and the night benefic — vitality joined with pleasure and beauty",
	"Sun,Mars":       "the king and the soldier — authority backed by force, courage or conflict",
	"Sun,Jupiter":    "the day star and the day benefic — authority amplified by fortune and abundance",
	"Sun,Saturn":     "the king and the old man — authority constrained by time, discipline, or limitation",
	"Moon,Mercury":   "the body and the mind — instinct shaped by reason, or reason colored by mood",
	"Moon,Venus":     "the night light and the night benefic — fortune through pleasure, the mother's love",
	"Moon,Mars":      "the body and the sword — emotional volatility, reactive courage, or inflammation",
	"Moon,Jupiter":   "the body and fortune — emotional abundance, generosity, or excess",
	"Moon,Saturn":    "the body and the limit — emotional restraint, melancholy, or endurance",
	"Mercury,Mars":   "reason and the sword — sharp speech, quick decisions, the debater's edge",
	"Mercury,Jupiter": "the messenger and fortune — expansive thinking, persuasive speech, good counsel",
	"Mercury,Saturn": "reason and the limit — precise thought, slow judgment, the scholar's discipline",
	"Venus,Mars":     "the benefic and malefic of night — desire and pursuit, love and conflict intertwined",
	"Venus,Jupiter":  "the two benefics — pleasure amplified by fortune, generosity, abundance of good things",
	"Venus,Saturn":   "pleasure and the limit — love constrained, commitment, or austerity in desire",
	"Mars,Jupiter":   "the sword and fortune — bold action, crusading energy, or reckless excess",
	"Mars,Saturn":    "the two malefics — force against the wall, frustration that builds or breaks",
	"Jupiter,Saturn": "fortune and the limit — the great cycle, expansion and contraction in rhythm",
}

// ── Hellenistic dignity tables (classical planets only) ────────────────

var koineDomicile = map[string]string{
	"Sun": "Leo", "Moon": "Cancer", "Mercury": "Gemini,Virgo",
	"Venus": "Taurus,Libra", "Mars": "Aries,Scorpio",
	"Jupiter": "Sagittarius,Pisces", "Saturn": "Capricorn,Aquarius",
}

var koineExaltation = map[string]string{
	"Sun": "Aries", "Moon": "Taurus", "Mercury": "Virgo",
	"Venus": "Pisces", "Mars": "Capricorn", "Jupiter": "Cancer",
	"Saturn": "Libra",
}

var koineFall = map[string]string{
	"Sun": "Libra", "Moon": "Scorpio", "Mercury": "Pisces",
	"Venus": "Virgo", "Mars": "Cancer", "Jupiter": "Capricorn",
	"Saturn": "Aries",
}

var koineDetriment = map[string]string{
	"Sun": "Aquarius", "Moon": "Capricorn", "Mercury": "Sagittarius,Pisces",
	"Venus": "Aries,Scorpio", "Mars": "Taurus,Libra",
	"Jupiter": "Gemini,Virgo", "Saturn": "Cancer,Leo",
}

// ── Hellenistic triplicity rulers (Dorothean) ──────────────────────────

var koineTriplicityRulers = map[string][3]string{
	"Aries":       {"Sun", "Jupiter", "Saturn"},
	"Taurus":      {"Venus", "Moon", "Mars"},
	"Gemini":      {"Saturn", "Mercury", "Jupiter"},
	"Cancer":      {"Venus", "Mars", "Moon"},
	"Leo":         {"Sun", "Jupiter", "Saturn"},
	"Virgo":       {"Venus", "Moon", "Mars"},
	"Libra":       {"Saturn", "Mercury", "Jupiter"},
	"Scorpio":     {"Venus", "Mars", "Moon"},
	"Sagittarius": {"Sun", "Jupiter", "Saturn"},
	"Capricorn":   {"Venus", "Moon", "Mars"},
	"Aquarius":    {"Saturn", "Mercury", "Jupiter"},
	"Pisces":      {"Venus", "Mars", "Moon"},
}

// ── Hellenistic sect ───────────────────────────────────────────────────

// SectLuminary returns the luminary of the sect: "Sun" for day, "Moon" for night.
func SectLuminary(isDayChart bool) string {
	if isDayChart {
		return "Sun"
	}
	return "Moon"
}

// SectPreference returns whether a planet prefers day or night sect.
func SectPreference(planet string) string {
	switch planet {
	case "Sun", "Jupiter", "Saturn":
		return "day"
	case "Moon", "Venus", "Mars":
		return "night"
	case "Mercury":
		return "neutral"
	}
	return "unknown"
}

// IsInSect returns true if the planet is in its preferred sect.
func IsInSect(planet string, isDayChart bool) bool {
	pref := SectPreference(planet)
	if pref == "neutral" {
		return true // Mercury is always in sect
	}
	return (pref == "day") == isDayChart
}

// ── Hellenistic dignity assessment ─────────────────────────────────────

// KoineDignityLevel represents the Hellenistic dignity of a planet in a sign.
type KoineDignityLevel struct {
	Domicile    bool   `json:"domicile"`
	Exaltation  bool   `json:"exaltation"`
	Triplicity  bool   `json:"triplicity"`
	Detriment   bool   `json:"detriment"`
	Fall        bool   `json:"fall"`
	Peregrine   bool   `json:"peregrine"`
	InSect      bool   `json:"in_sect"`
	Description string `json:"description"`
}

// AssessKoineDignity evaluates a planet's Hellenistic dignity in a sign.
func AssessKoineDignity(planet, sign string, isDayChart bool) KoineDignityLevel {
	d := KoineDignityLevel{}

	// Domicile
	if strings.Contains(koineDomicile[planet], sign) {
		d.Domicile = true
	}

	// Exaltation
	if koineExaltation[planet] == sign {
		d.Exaltation = true
	}

	// Triplicity
	if rulers, ok := koineTriplicityRulers[sign]; ok {
		for _, r := range rulers {
			if r == planet {
				d.Triplicity = true
				break
			}
		}
	}

	// Detriment
	if strings.Contains(koineDetriment[planet], sign) {
		d.Detriment = true
	}

	// Fall
	if koineFall[planet] == sign {
		d.Fall = true
	}

	// Peregrine: no essential dignity at all
	d.Peregrine = !d.Domicile && !d.Exaltation && !d.Triplicity

	// Sect
	d.InSect = IsInSect(planet, isDayChart)

	// Build description
	var parts []string
	if d.Domicile {
		parts = append(parts, "in its own house — the planet is at home, operating with full authority")
	}
	if d.Exaltation {
		parts = append(parts, "exalted — the planet is elevated, honored as a guest of great status")
	}
	if d.Triplicity {
		parts = append(parts, "in its triplicity — the planet has support from its elemental kin")
	}
	if d.Detriment {
		parts = append(parts, "in its detriment — the planet is in a foreign house, operating against its nature")
	}
	if d.Fall {
		parts = append(parts, "in its fall — the planet is humbled, cast down from its place of honor")
	}
	if d.Peregrine {
		parts = append(parts, "peregrine — the planet is a wanderer with no essential dignity in this sign")
	}
	if d.InSect {
		parts = append(parts, "in sect — the planet is of the same party as the chart, strengthened")
	} else {
		parts = append(parts, "out of sect — the planet is of the opposite party, operating against the chart's nature")
	}

	d.Description = strings.Join(parts, "; ") + "."
	return d
}

// ── Interpretation functions ──────────────────────────────────────────

// KoineInterpretPlanetInSign returns a Hellenistic description of a planet in a sign.
func KoineInterpretPlanetInSign(planet, sign string, isDayChart bool) string {
	pd, ok := koinePlanetDescriptions[planet]
	if !ok {
		pd = strings.ToLower(planet)
	}
	sd, ok := koineSignDescriptions[sign]
	if !ok {
		sd = strings.ToLower(sign)
	}

	dignity := AssessKoineDignity(planet, sign, isDayChart)

	return fmt.Sprintf("%s in %s: %s, %s. %s",
		planet, sign, pd, sd, dignity.Description)
}

// KoineInterpretPlanetInHouse returns a Hellenistic description of a planet in a house.
func KoineInterpretPlanetInHouse(planet string, house int) string {
	pd, ok := koinePlanetDescriptions[planet]
	if !ok {
		pd = strings.ToLower(planet)
	}
	hd, ok := koineHouseDescriptions[house]
	if !ok {
		hd = fmt.Sprintf("house %d", house)
	}

	return fmt.Sprintf("%s in the %s: %s expressed through %s.",
		planet, ordinal(house), pd, hd)
}

// KoineInterpretAspect returns a Hellenistic description of an aspect between two planets.
func KoineInterpretAspect(planet1, planet2, aspect string, orb float64) string {
	ad, ok := koineAspectDescriptions[aspect]
	if !ok {
		ad = fmt.Sprintf("%s aspect", aspect)
	}

	// Look up pair dynamics (order-independent)
	key1 := planet1 + "," + planet2
	key2 := planet2 + "," + planet1
	dynamics, ok := koinePairDynamics[key1]
	if !ok {
		dynamics, ok = koinePairDynamics[key2]
	}
	if !ok {
		dynamics = fmt.Sprintf("%s and %s in contact", planet1, planet2)
	}

	return fmt.Sprintf("%s %s %s (orb %.1f°): %s — %s.",
		planet1, aspect, planet2, orb, dynamics, ad)
}

// KoineInterpretChart produces a full Koiné chart interpretation from
// planetary positions, house placements, aspects, and patterns.
// Uses Hellenistic source texts, classical planets only, Ptolemaic aspects.
func KoineInterpretChart(
	name string,
	planets map[string]float64,
	houses map[string]int,
	aspects []AspectHit,
	patterns []PatternHit,
	isDayChart bool,
) *ChartInterpretation {
	report := &ChartInterpretation{
		Name:         name,
		PlanetSigns:  make([]string, 0),
		PlanetHouses: make([]string, 0),
		Aspects:      make([]string, 0),
		Patterns:     make([]string, 0),
	}

	// Planet-in-sign interpretations (classical planets only)
	for _, planet := range ClassicalPlanets {
		lon, ok := planets[planet]
		if !ok {
			continue
		}
		sign := SignForLongitude(lon)
		report.PlanetSigns = append(report.PlanetSigns,
			KoineInterpretPlanetInSign(planet, sign, isDayChart))
	}

	// Planet-in-house interpretations
	for planet, house := range houses {
		// Only interpret classical planets
		isClassical := false
		for _, cp := range ClassicalPlanets {
			if planet == cp {
				isClassical = true
				break
			}
		}
		if !isClassical {
			continue
		}
		report.PlanetHouses = append(report.PlanetHouses,
			KoineInterpretPlanetInHouse(planet, house))
	}

	// Aspect interpretations (classical planets only)
	for _, a := range aspects {
		if !isClassicalPlanet(a.Planet1) || !isClassicalPlanet(a.Planet2) {
			continue
		}
		report.Aspects = append(report.Aspects,
			KoineInterpretAspect(a.Planet1, a.Planet2, a.Aspect, a.Orb))
	}

	// Pattern interpretations (Hellenistic doesn't use modern patterns,
	// but we include them for structural completeness)
	for _, p := range patterns {
		report.Patterns = append(report.Patterns,
			fmt.Sprintf("%s involving %s: a configuration of planets in aspect.",
				p.Name, strings.Join(p.Planets, ", ")))
	}

	return report
}

// isClassicalPlanet returns true if the planet is one of the seven classical planets.
func isClassicalPlanet(name string) bool {
	for _, cp := range ClassicalPlanets {
		if name == cp {
			return true
		}
	}
	return false
}

// ordinal returns the English ordinal for a number (1st, 2nd, etc.).
func ordinal(n int) string {
	switch n {
	case 1:
		return "1st house"
	case 2:
		return "2nd house"
	case 3:
		return "3rd house"
	default:
		return fmt.Sprintf("%dth house", n)
	}
}
