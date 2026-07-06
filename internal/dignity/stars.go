package dignity

import "math"

// ── Star Catalog ────────────────────────────────────────────────────────
// 116 visible stars (mag generally < 5), names from sefstars.txt.
// These are the same stars used by the Python reference implementation.

var StarNames = []string{
	"Achernar", "Acrux", "Acubens", "Adhafera", "Agena", "Ain",
	"Albireo", "Alcyone", "Aldebaran", "Algieba", "Algorab", "Alhena",
	"Alioth", "Alkaid", "Alkes", "Alnilam", "Alnitak", "Alpha Centauri",
	"Alphard", "Alphecca", "Alsephina", "Altair", "Antares", "Arcturus",
	"Ascella", "Aspidiske", "Atik", "Atlas", "Atria", "Auva",
	"Avior", "Baten Kaitos", "Bellatrix", "Betelgeuse", "Canopus", "Capella",
	"Caph", "Castor", "Chertan", "Dabih", "Deneb", "Denebola",
	"Diphda", "Dubhe", "Electra", "Elnath", "Enif", "Etamin",
	"Fomalhaut", "Gacrux", "Gienah", "Hamal", "Heze", "Izar",
	"Kaus Australis", "Kaus Borealis", "Kaus Media", "Kochab", "Labrum", "Maia",
	"Markab", "Markeb", "Menkar", "Menkent", "Menkib", "Merak",
	"Merope", "Miaplacidus", "Mimosa", "Mintaka", "Mirach", "Mirfak",
	"Mizar", "Muphrid", "Nashira", "Nekkar", "Nunki", "Peacock",
	"Pollux", "Porrima", "Procyon", "Ras Elased Aust", "Rasalhague", "Rastaban",
	"Regulus", "Rigel", "Ruchbah", "Ruchba", "Sabik", "Sadachbia",
	"Sadalmelik", "Sadalsuud", "Saiph", "Scheat", "Schedar", "Seginus",
	"Shaula", "Sirius", "Skat", "Spica", "Subra", "Suhail",
	"Suhail al Muhlif", "Syrma", "Taygeta", "Tegmine", "Thuban", "Turais",
	"Unukalhai", "Vega", "Vindemiatrix", "Zaniah", "Zavijava", "Zosma",
	"Zubenelgenubi", "Zubeneschamali",
}

// StarMeanings holds traditional star lore (Ptolemaic / Medieval).
// Natures from Ptolemy's Tetrabiblos, Al-Sufi, and Robson (1923).
var StarMeanings = map[string]string{
	"Achernar":       "Jupiter nature. Leadership, guidance. 'End of the River.' Navigational star of the southern hemisphere. Brings success in public office, philanthropy, and spiritual insight.",
	"Acrux":          "Jupiter nature. Spirituality, devotion. Southern Cross. The brightest star in Crux — faith made visible. Exalted spirituality, protection, moral authority.",
	"Acubens":        "Saturn-Mercury nature. The Crab's Claw. Mental tenacity, persistence through difficulty. Analytical mind that works slowly but deeply. Liability to slander or gossip.",
	"Adhafera":       "Saturn-Mercury nature. The Lion's mane. Pride in intellect. Scholarly achievement through sustained effort. Cold ambition cloaked in dignity.",
	"Agena":          "Venus-Jupiter nature. Ambition, power. 'The Centaur's Knee.' Strength tempered by grace. Success in law, philosophy, and positions of trust.",
	"Ain":            "Venus nature. Gentle force. 'Eye of the Bull.' The northernmost bright star of the Hyades. Creative expression, gentleness that achieves what force cannot.",
	"Albireo":        "Venus-Mercury nature. Artistic talent, beauty. Beautiful double star — gold and sapphire. The painter's star. Aesthetic refinement, grace in expression.",
	"Alcyone":        "Moon-Jupiter nature. Mysticism, sorrow, compassion. Central Pleiades star. The 'Weeping Sisters.' Profound emotional depth, occult insight, prophecy through dreams.",
	"Aldebaran":      "Mars nature. The Eye of the Bull. Ambition, leadership, courage, military honor. The Watcher of the East. One of the Four Royal Stars of Persia. Brings riches and honor but requires integrity to hold them.",
	"Algieba":        "Mars-Mercury nature. Leadership, intellect. The Lion's mane. Command through intellect rather than brute force. Strategic mind in positions of authority.",
	"Algorab":        "Mars-Saturn nature. Obstructive, destructive, scavenging. The Raven's Wing. Brings delays and obstacles. Tests character through frustration.",
	"Alhena":         "Mercury-Venus nature. Artistic talent, skill with words. 'The Wounded Heel.' Creative genius that compensates for some innate limitation. Literary and musical ability.",
	"Alioth":         "Mars nature in Ursa Major. Administrative ability, leadership. 'The Black Horse.' Independence, self-reliance. Military or executive command.",
	"Alkaid":         "Moon-Venus nature. Artistic sensitivity, emotional depth. 'The Chief of the Mourners.' The last star in the Big Dipper's handle. Profound empathy that borders on grief.",
	"Alkes":          "Venus nature. The Wine Cup. Artistic refinement, aesthetic sensibility. Love of beauty, pleasure, and sensory experience. Generosity of spirit.",
	"Alnilam":        "Jupiter-Saturn nature. Ambition, authority. 'String of Pearls.' The central star of Orion's Belt. Balanced ambition — expansion governed by structure.",
	"Alnitak":        "Jupiter-Saturn nature. Ambition. 'The Girdle.' Easternmost belt star. Drive for achievement with the patience to see it through.",
	"Alpha Centauri": "Venus-Jupiter nature. Riches, honor, travel. Nearest star system to Earth. Success in foreign lands, broad-minded philosophy. Natural leader.",
	"Alphard":        "Saturn-Venus nature. Heart of Hydra. Solitude, poison, wisdom. 'The Solitary One.' Knowledge gained in isolation. Danger from emotional attachments.",
	"Alphecca":       "Venus-Mercury nature. Artistic honors. 'The Broken Ring.' The jewel of the Northern Crown. Recognition for creative achievement. Fame that comes late but lasts.",
	"Alsephina":      "Venus nature. Speed, travel. Sails of Argo. Swiftness in action and thought. Restlessness channeled into achievement.",
	"Altair":         "Mars-Jupiter nature. Boldness, courage, sudden advancement. 'The Flying Eagle.' Rapid rise to prominence. Military or athletic distinction.",
	"Antares":        "Mars-Jupiter nature. Heart of the Scorpion. Rival of Mars. Intensity, obsession, power. The Watcher of the West. Great power that can consume if not mastered.",
	"Arcturus":       "Jupiter-Mars nature. Prosperity, honor, guardianship. 'Bear Watcher.' Guardian of the northern sky. Success through sustained effort. The protector.",
	"Ascella":        "Mercury-Mars nature. The Horse's Armpit. The Archer's bow-hand. Mental quickness applied to action. Skill in competitive pursuits.",
	"Aspidiske":      "Jupiter nature. The Ship's Keel. Leadership, navigation, exploration. Guides others through unknown waters. Philanthropy in later life.",
	"Atik":           "Mars nature. Perseus's shoulder. Martial energy focused through action rather than intellect. Physical courage, pioneer spirit.",
	"Atlas":          "Moon-Jupiter nature. Endurance, burden-bearing. The Titan who holds the sky. A Pleiad with weight. Responsibility accepted willingly. Philosophical depth.",
	"Atria":          "Jupiter nature. Leadership in exploration and expansion. The brightest star in Triangulum Australe. Success in ventures beyond familiar boundaries.",
	"Auva":           "Mercury nature. The Virgin's girdle. Analytical mind applied to practical matters. Precision in communication and commerce.",
	"Avior":          "Venus-Jupiter nature. Guidance, beauty, navigation. One of the brightest stars of Argo. Generosity and grace in positions of leadership.",
	"Baten Kaitos":   "Saturn nature. The Whale's Belly. Isolation, depth, hidden wisdom. Profound thinker who works best alone. Liability to melancholy.",
	"Bellatrix":      "Mars-Mercury nature. Female warrior. Military courage, eloquence. 'Amazon Star.' Success through bold speech and decisive action.",
	"Betelgeuse":     "Mars-Mercury nature. Wealth, military honor, fame. 'The Armpit of the Giant.' Sudden success that requires character to sustain.",
	"Canopus":        "Saturn-Jupiter nature. Travel, philosophy, leadership. 'Navigator's Star.' The second brightest star. Wisdom gained through experience. Success in old age.",
	"Capella":        "Mars-Mercury nature. Wealth, honor, eminence. 'The Little She-Goat.' Military and civic success. Sharp mind in positions of command.",
	"Caph":           "Saturn-Venus nature. The Queen's Breast. Authority tempered by grace. The brightest star in Cassiopeia. Dignity in adversity.",
	"Castor":         "Mercury nature. Intellect, sudden fame, occult skill. Twin of Pollux. Brilliance in law, writing, or scholarship. Rapid mental processing.",
	"Chertan":        "Saturn nature. The Lion's hindquarters. Authority earned through persistence. The slow climb rather than the sudden rise. Durable reputation.",
	"Dabih":          "Saturn-Venus nature. Authority, structure. 'The Slaughterer.' The Goat's horn. Disciplined ambition. Success in governance and administration.",
	"Deneb":          "Venus-Mercury nature. Creativity, idealism, leadership. 'The Tail of the Swan.' Artistic vision that inspires others. The poet-king archetype.",
	"Denebola":       "Saturn-Venus nature. Noble but unfortunate. 'The Lion's Tail.' Talent and refinement undermined by circumstance. Late recognition.",
	"Diphda":         "Saturn nature. Solitude, sorrow, transformation through loss. 'The Second Frog.' Profound emotional depth. Wisdom born of suffering.",
	"Dubhe":          "Mars nature. Destructive energy channeled into leadership. 'The Bear.' The pointer star. Raw power directed toward protection.",
	"Electra":        "Moon-Mars nature. The shining one of the Pleiades. Emotional intensity, creative fire. Passion that drives artistic expression or personal conflict.",
	"Elnath":         "Mars nature. Conflict, ambition, butting heads. 'The Butting One.' The tip of the Bull's horn. Aggressive pursuit of goals.",
	"Enif":           "Mars-Mercury nature. Intuitive action. 'The Horse's Nose.' The nose of Pegasus. Quick decisions in the moment. Athletic intuition.",
	"Etamin":         "Saturn-Mars nature. Penetrating mind, danger. 'The Dragon.' The Dragon's head. Intellectual rigor that isolates. Danger from overconfidence.",
	"Fomalhaut":      "Venus-Mercury nature. Idealism, magic, poetry. One of the Four Royal Stars. The Watcher of the South. Spiritual elevation through art and devotion.",
	"Gacrux":         "Venus-Jupiter nature. Creativity, devotion. Top of the Cross. Artistic genius in service of higher truth. The Cross's crown.",
	"Gienah":         "Mars-Venus nature. Creative tension. 'The Raven's Wing.' The balance of passion and grace. Artistic expression born of internal conflict.",
	"Hamal":          "Mars-Saturn nature. Head of the Ram. Brutal honesty. Pioneer spirit. The first bright star of the zodiac. Raw, unfiltered force.",
	"Heze":           "Mercury-Venus nature. The Virgin's waist. Mental acuity with aesthetic refinement. Precision in creative work. Success in design, writing, or scholarship.",
	"Izar":           "Venus-Mercury nature. Artistic talent. Beautiful binary — orange and blue. The Loincloth of Bootes. Refined expression, grace under pressure.",
	"Kaus Australis": "Mercury-Mars nature. Archery, skill. 'Southern Bow.' The Archer's southernmost bright star. Competitive excellence. Precision under pressure.",
	"Kaus Borealis":  "Mercury-Mars nature. Aim, direction, competitive drive. The Archer's northern bow. Clear vision of distant goals pursued with determination.",
	"Kaus Media":     "Mercury-Mars nature. Skill, precision. 'Middle Bow.' The central star of the Archer's bow. Technical mastery in any field of competition.",
	"Kochab":         "Saturn-Venus nature. Guardianship. 'The Pole Star's Companion.' The brighter of the two Guardians of the Pole. Steady, quiet authority.",
	"Labrum":         "Venus-Mercury nature. The Holy Grail. Artistic vision bordering on the sacred. The Cup. Beauty as revelation. Creative work with spiritual depth.",
	"Maia":           "Moon-Mercury nature. Growth, nurture, intellectual fertility. Eldest of the Pleiades. Maternal wisdom. Protective instinct combined with mental agility.",
	"Markab":         "Mars-Mercury nature. Danger, but also honor. 'The Saddle.' The brightest star of Pegasus. Risk-taking that leads to distinction.",
	"Markeb":         "Jupiter nature. 'The Ship's Rib.' Guidance through uncertainty. Navigational wisdom. Steady leadership in uncharted territory.",
	"Menkar":         "Saturn-Venus nature. Danger from large animals. Whale's Jaw. Primal forces. Wisdom in confronting the uncontrollable.",
	"Menkent":        "Venus-Mercury nature. Wisdom, healing. 'Centaur's Shoulder.' Knowledge that soothes. The physician's star.",
	"Menkib":         "Mars-Mercury nature. Perseus's knee. Quick martial action. Pioneer energy channeled through intellect rather than brute force.",
	"Merak":          "Mars nature. Leadership, guardianship. 'The Loin.' The southern pointer star of the Big Dipper. Protective authority.",
	"Merope":         "Moon-Mars nature. The lost Pleiad. Emotional intensity, shame, hidden brilliance. Talent obscured by circumstance or personal difficulty.",
	"Miaplacidus":    "Jupiter nature. Leadership, guidance, steady hand. The helm of the great ship Argo. Calm authority in crisis. Trusted counselor.",
	"Mimosa":         "Venus-Mercury nature. Creativity, artistic skill. Southern Cross. Beauty that inspires. The second-brightest star of Crux.",
	"Mintaka":        "Jupiter-Saturn nature. Balanced judgment, authority. The westernmost belt star of Orion. Justice tempered with wisdom. Success through fairness.",
	"Mirach":         "Venus nature. Beauty, artistic sensitivity, marital happiness. The Girdle of Andromeda. Grace, charm, and harmony in relationships.",
	"Mirfak":         "Jupiter-Saturn nature. Power, eloquence. 'The Elbow' of Perseus. Command through presence and speech. Authority that persuades rather than compels.",
	"Mizar":          "Venus-Mercury nature. Musical talent, artistic genius. 'The Girdle.' The famous double star in the Big Dipper. Creative brilliance.",
	"Muphrid":        "Venus-Mercury nature. The Solitary One. Grace and artistic refinement. The hermit artist — beauty created in contemplation.",
	"Nashira":        "Saturn-Venus nature. Authority, grace under pressure. The Goat's tail. Dignified accomplishment. Success in mature years.",
	"Nekkar":         "Mars-Mercury nature. The Herdsman's head. Leadership through intellect. Command that relies on strategic thinking rather than force.",
	"Nunki":          "Mercury-Jupiter nature. Philosophy, travel. 'Star of the Proclamation of the Sea.' Broad-minded wisdom from wide experience. The sage.",
	"Peacock":        "Venus-Mercury nature. Beauty, vanity. Peacock Star. Art for art's sake. Creative brilliance that draws admiration and envy.",
	"Pollux":         "Mars nature. Boldness, cunning, cruelty. Twin of Castor. More physical, less intellectual. Success through will and audacity.",
	"Porrima":        "Venus-Mercury nature. Prophecy. Binary that was separating, now converging again. Foresight, intuition, reconciliation.",
	"Procyon":        "Mercury-Mars nature. Activity, wealth, notoriety. 'Before the Dog.' The precursor. Quick success that arrives before the main event.",
	"Ras Elased Aust": "Saturn-Mercury nature. The Lion's southern head. Intellectual authority. Scholarly reputation earned through sustained labor.",
	"Rasalhague":     "Saturn-Venus nature. Healing, wisdom, danger from reptiles. Head of the Snake Charmer. Knowledge that cures and protects.",
	"Rastaban":       "Saturn-Mars nature. The Dragon's head. Intellectual force with destructive potential. Penetrating mind that can wound.",
	"Regulus":        "Mars-Jupiter nature. The Heart of the Lion. Power, success, revenge. Royal star. The Watcher of the North. One of the Four Royal Stars. Great success if integrity is maintained.",
	"Rigel":          "Jupiter-Saturn nature. Riches, honor, creative genius. 'The Left Foot' of Orion. The brightest star of Orion. Success in enterprise, education, and the arts.",
	"Ruchbah":        "Saturn-Venus nature. The Queen's knee. Authority and grace held in balance. Diplomacy in positions of power.",
	"Ruchba":         "Venus-Mercury nature. The Swan's tail-tip. Artistic refinement at the furthest reach. Creative vision that extends beyond the familiar.",
	"Sabik":          "Saturn-Venus nature. Waste, perversion. 'The Preceding One.' The Serpent Bearer's knee. Corruption of gifts. Talent misdirected.",
	"Sadachbia":      "Saturn-Mercury nature. Occult skill, hidden knowledge. 'Luck of the Hidden.' Wisdom gained in secret study.",
	"Sadalmelik":     "Saturn-Mercury nature. Fortune, intellect. 'Lucky One of the King.' Success through knowledge. The Aquarian ruler star.",
	"Sadalsuud":      "Saturn-Mercury nature. Fortune, occult skill. 'Luckiest of the Lucky.' The Aquarian star of prosperity. Intellectual wealth.",
	"Saiph":          "Mars-Jupiter nature. The Sword. Martial energy with scope and vision. The giant Orion's knee. Decisive action on a grand scale.",
	"Scheat":         "Mars-Mercury nature. Danger, imprisonment, but great creativity. 'The Horse's Shoulder.' Genius born of constraint.",
	"Schedar":        "Saturn-Venus nature. Astrology, mysticism. 'The Breast' of the Queen. Profound insight into hidden patterns.",
	"Seginus":        "Mercury-Saturn nature. The Herdsman's shoulder. Intellectual discipline. Structured thought applied to practical ends.",
	"Shaula":         "Mars-Mercury nature. Danger, but also ambition. 'The Sting.' The Scorpion's tail. Power that wounds. Risk-taking rewarded.",
	"Sirius":         "Jupiter-Mars nature. The brightest star. Wealth, honor, devotion, scorching intensity. 'The Scorcher.' The Dog Star. Success that burns as brightly as it illuminates.",
	"Skat":           "Saturn-Mercury nature. Fortune, but also isolation. 'The Shin.' Wealth acquired through solitary effort.",
	"Spica":          "Venus-Mars nature. The Ear of Wheat. Abundance, knowledge, justice. Giver of fortunes. The brightest star in Virgo. Success in arts, sciences, and law.",
	"Subra":          "Saturn-Mars nature. The Lion's paw. Power applied with restraint. Authority that doesn't need to announce itself.",
	"Suhail":         "Venus-Mercury nature. Wisdom, travel, eloquence. 'The Smooth Plain.' Broad experience elegantly expressed.",
	"Suhail al Muhlif": "Venus-Jupiter nature. Generosity, leadership, navigation. The brightest star in the Argo's sails. Magnanimity in command.",
	"Syrma":          "Mercury-Venus nature. Precision, refinement, service. The Virgin's train. Meticulous attention to detail in creative work.",
	"Taygeta":        "Moon nature. The shadowed Pleiad. Emotional depth, hidden creativity. Sensitivity that borders on the psychic.",
	"Tegmine":        "Saturn-Mercury nature. The Crab's shell. Mental fortification. Protected intelligence. Withdrawn but formidable.",
	"Thuban":         "Saturn-Venus nature. Ancient wisdom. Former Pole Star (circa 2800 BCE). Guardianship of timeless knowledge.",
	"Turais":         "Jupiter nature. Navigation, leadership. The false cross of Carina. Visionary guidance. Success through charting new paths.",
	"Unukalhai":      "Saturn-Mars nature. Danger from poison, but also wisdom. 'Serpent's Neck.' Knowledge of what harms and what heals.",
	"Vega":           "Venus-Mercury nature. Artistic genius, idealism, leadership. The Harp Star. The brightest star of the northern hemisphere. Creative brilliance that sets standards.",
	"Vindemiatrix":   "Saturn-Mercury nature. Harvest, endings. 'Grape-Gatherer.' The reaper. Wisdom in closure. Success in the final accounting.",
	"Zaniah":         "Mercury-Venus nature. Precision, analysis, beauty. The Virgin's back. Refinement in technical work. Craft elevated to art.",
	"Zavijava":       "Mercury-Mars nature. The Virgin's corner. Mental sharpness applied to practical problems. Incisive intelligence.",
	"Zosma":          "Saturn-Venus nature. Victimization, self-pity. 'The Lion's Hip.' The wound of pride. Talent diminished by circumstance.",
	"Zubenelgenubi":  "Saturn-Mars nature. Injustice, but also legal skill. 'Southern Claw.' The southern scale pan. Justice sought through conflict.",
	"Zubeneschamali": "Jupiter-Mercury nature. Justice, eloquence, success in law. 'Northern Claw.' The northern scale pan. Victory through reasoned argument.",
}

// ── Star Conjunction ────────────────────────────────────────────────────

// StarConjunction records a fixed star conjunct a planet or point.
type StarConjunction struct {
	Star      string  `json:"star"`
	StarLon   float64 `json:"star_lon"`
	Planet    string  `json:"planet"`
	PlanetLon float64 `json:"planet_lon"`
	Orb       float64 `json:"orb"`
	Meaning   string  `json:"meaning"`
}

// FindStarConjunctions finds all conjunctions between star positions and
// planet/point positions within the given orb. Results are sorted by orb
// (tightest first). Uses the shortest angular distance across 0°/360°.
func FindStarConjunctions(starPositions, planetPositions map[string]float64, maxOrb float64) []StarConjunction {
	var conjunctions []StarConjunction

	for star, slon := range starPositions {
		for planet, plon := range planetPositions {
			orb := math.Abs(math.Mod(slon-plon+540, 360) - 180)
			orbRounded := math.Round(orb*100) / 100
			if orbRounded <= maxOrb {
				conjunctions = append(conjunctions, StarConjunction{
					Star:      star,
					StarLon:   slon,
					Planet:    planet,
					PlanetLon: plon,
					Orb:       orbRounded,
					Meaning:   StarMeanings[star],
				})
			}
		}
	}

	// Sort by orb, tightest first
	for i := 0; i < len(conjunctions); i++ {
		for j := i + 1; j < len(conjunctions); j++ {
			if conjunctions[j].Orb < conjunctions[i].Orb {
				conjunctions[i], conjunctions[j] = conjunctions[j], conjunctions[i]
			}
		}
	}

	return conjunctions
}

// ── Star Aspects (all aspect types, not just conjunctions) ────────────────

// StarAspectHit records an aspect between a planet and a fixed star.
type StarAspectHit struct {
	Star      string  `json:"star"`
	StarLon   float64 `json:"star_lon"`
	Planet    string  `json:"planet"`
	PlanetLon float64 `json:"planet_lon"`
	Aspect    string  `json:"aspect"`
	Orb       float64 `json:"orb"`
}

// FindStarAspects finds all aspects between a single fixed star and a set of
// planet positions within the given orb. Results are sorted by orb (tightest
// first). Uses the same angleDist and AspectDef types as the rest of the
// library — no duplicated math.
func FindStarAspects(starLon float64, starName string, planetPositions map[string]float64, aspects []AspectDef, orb float64) []StarAspectHit {
	var hits []StarAspectHit

	for planet, plon := range planetPositions {
		dist := angleDist(starLon, plon)
		for _, a := range aspects {
			diff := math.Abs(dist - a.Angle)
			if diff <= orb {
				hits = append(hits, StarAspectHit{
					Star:      starName,
					StarLon:   starLon,
					Planet:    planet,
					PlanetLon: plon,
					Aspect:    a.Name,
					Orb:       math.Round(diff*100) / 100,
				})
			}
		}
	}

	// Sort by orb, tightest first
	for i := 0; i < len(hits); i++ {
		for j := i + 1; j < len(hits); j++ {
			if hits[j].Orb < hits[i].Orb {
				hits[i], hits[j] = hits[j], hits[i]
			}
		}
	}

	return hits
}

// ── Cross-System Star Conjunction Comparison ──────────────────────────────

// StarCrossSystem holds the result of comparing star conjunctions
// in tropical vs sidereal frames.
type StarCrossSystem struct {
	Name          string            `json:"name"`
	Ayanamsa      float64           `json:"ayanamsa"`
	Orb           float64           `json:"orb"`
	Tropical      []StarConjunction `json:"tropical"`
	Sidereal      []StarConjunction `json:"sidereal"`
	Survivors     []StarConjunction `json:"survivors"`
	TropicalOnly  []StarConjunction `json:"tropical_only"`
	SiderealOnly  []StarConjunction `json:"sidereal_only"`
	TotalTrop     int               `json:"total_tropical"`
	TotalSid      int               `json:"total_sidereal"`
	TotalSurvivors int              `json:"total_survivors"`
}

// CompareStarConjunctionsCrossSystem computes star conjunctions in both
// tropical and sidereal frames and classifies which survive the zodiac shift.
// Star positions are computed via swe.Fixstar (tropical). Sidereal star
// positions are tropical - ayanamsa. Planet positions shift by the same
// ayanamsa, so angular distances are preserved. This function verifies that
// computationally.
func CompareStarConjunctionsCrossSystem(
	name string,
	starPositionsTropical map[string]float64,
	planetPositionsTropical map[string]float64,
	ayanamsa float64,
	orb float64,
) *StarCrossSystem {
	// Tropical conjunctions
	tropConj := FindStarConjunctions(starPositionsTropical, planetPositionsTropical, orb)

	// Sidereal: shift both stars and planets by ayanamsa
	starPosSid := make(map[string]float64)
	for k, v := range starPositionsTropical {
		starPosSid[k] = normalizeLon(v - ayanamsa)
	}
	planetPosSid := make(map[string]float64)
	for k, v := range planetPositionsTropical {
		planetPosSid[k] = normalizeLon(v - ayanamsa)
	}
	sidConj := FindStarConjunctions(starPosSid, planetPosSid, orb)

	// Classify survivors: same star + same planet + same orb (within 0.01°)
	survivors, tropOnly, sidOnly := classifyStarSurvivors(tropConj, sidConj)

	return &StarCrossSystem{
		Name:           name,
		Ayanamsa:       ayanamsa,
		Orb:            orb,
		Tropical:       tropConj,
		Sidereal:       sidConj,
		Survivors:      survivors,
		TropicalOnly:   tropOnly,
		SiderealOnly:   sidOnly,
		TotalTrop:      len(tropConj),
		TotalSid:       len(sidConj),
		TotalSurvivors: len(survivors),
	}
}

// classifyStarSurvivors matches conjunctions between two lists.
// A survivor is a conjunction with the same star and planet in both lists,
// with orb difference < 0.01°.
func classifyStarSurvivors(trop, sid []StarConjunction) (survivors, tropOnly, sidOnly []StarConjunction) {
	survivors = make([]StarConjunction, 0)
	tropOnly = make([]StarConjunction, 0)
	sidOnly = make([]StarConjunction, 0)

	// Build a set of sidereal conjunctions keyed by star+planet
	sidSet := make(map[string]StarConjunction)
	for _, c := range sid {
		key := c.Star + "|" + c.Planet
		sidSet[key] = c
	}

	tropSeen := make(map[string]bool)
	for _, c := range trop {
		key := c.Star + "|" + c.Planet
		tropSeen[key] = true
		if s, ok := sidSet[key]; ok {
			if math.Abs(c.Orb-s.Orb) < 0.01 {
				survivors = append(survivors, c)
				continue
			}
		}
		tropOnly = append(tropOnly, c)
	}

	for _, c := range sid {
		key := c.Star + "|" + c.Planet
		if !tropSeen[key] {
			sidOnly = append(sidOnly, c)
		}
	}

	return
}
