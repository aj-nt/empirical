# Astrology Systems

Empirical implements multiple astrological systems as independent computational frameworks. The core insight: these are degraded transmissions of a common Hellenistic original (~150 BCE Alexandria). By running them side by side, we can measure what survived and what diverged.

## Western (Tropical)

The modern Western system uses the tropical zodiac (tied to the vernal equinox) and Placidus houses by default. It includes:

- Classical planets (Sun through Saturn) plus Uranus, Neptune, Pluto
- Asteroids: Chiron, Ceres, Pallas, Juno, Vesta
- Dwarf planets: Eris, Makemake, Gonggong
- Points: Lilith (mean), North Node, Part of Fortune
- Trans-Neptunian Points (TNPs): Cupido, Hades, Zeus, Kronos, Apollon, Admetos, Vulcanus, Poseidon
- Essential dignities: domicile, exaltation, triplicity, term, face
- Accidental dignities: house placement, aspects, speed, orientation
- Aspect patterns: T-square, Grand Trine, Yod, Cradle, Stellium, Grand Cross, Kite, Mystic Rectangle

## Vedic (Sidereal)

The Vedic (Jyotish) system uses the sidereal zodiac (Lahiri ayanamsa by default) and Whole Sign houses. It includes:

- Nine grahas (includes Rahu/Ketu as the lunar nodes)
- 27 nakshatras (lunar mansions) with determinative stars
- Shadbala (six-fold strength) computation
- Vimshottari dasha system (120-year planetary period cycle)
- Essential dignities: own sign, exaltation, moolatrikona, friend/neutral/enemy

## Koiné (Hellenistic)

Koiné is a synthesis system built from the findings of the empirical paper. It uses elements with the strongest transmission signal across traditions, supplemented by techniques documented in Hellenistic source texts (Ptolemy's *Tetrabiblos*, Valens' *Anthology*).

Design philosophy: **paper-informed, not paper-constrained**. The paper required survival in both Western and Vedic descendants — a conservative method with zero false positives but ~40% false negatives. Koiné includes techniques that survived in only one descendant when they are independently documented in Hellenistic sources.

Features:
- Whole Sign houses
- Traditional seven planets
- Sect (day/night chart) as primary interpretive axis
- Triplicity rulers of the sect light
- Bounds (terms) — Egyptian bounds
- Decans — Chaldean decans
- Lots: Fortune, Spirit, Eros, Necessity, Courage, Victory, Nemesis, Basis
- Time lord systems: distributions through the bounds, zodiacal releasing
- Caput/Cauda Draconis (North/South Node) as amplification/diminishment — not Rahu/Ketu

## Ba Zi (Four Pillars)

Chinese astrology using the Four Pillars (year, month, day, hour) with Heavenly Stems and Earthly Branches. Includes:

- Day Master analysis
- Ten Gods (resource, parallel, output, wealth, power)
- Five Elements balance
- Clash, harm, combination, punishment relationships
- 12 growth phases (chang sheng)

## The Recover Protocol

The `recover` command runs all systems on the same birth data and produces a convergence report:

```
Planet     Trop Sign      Sid Sign       Western        Vedic          Verdict
------------------------------------------------------------------------------
Sun        Aquarius       Aquarius       detriment      peregrine      NOISE
Moon       Aquarius       Capricorn      peregrine      peregrine      SIGNAL
```

- **SIGNAL** — both systems agree on dignity classification (the placement survived transmission)
- **NOISE** — systems disagree (the placement diverged during independent evolution)

This is not about which system is "right." It's about measuring transmission fidelity. A high-signal chart has placements that survived 2,000 years of independent cultural evolution. A low-signal chart has placements where the traditions diverged — and that divergence itself is information.

## Further Reading

- [Design Document (DESIGN.md)](../DESIGN.md) — Koiné design philosophy and evidence basis
- [Manual (MANUAL.md)](../MANUAL.md) — comprehensive system reference
- [Research Paper (docs/paper/paper.md)](paper/paper.md) — "Computational Comparison of Western, Vedic, and Arabic Astrological Frameworks" (Flinton, 2026)
