# Koiné Astrology — Design Document

## Relationship to the Empirical Paper

Koiné is a synthesis system built from the findings of "Computational Comparison
of Western, Vedic, and Arabic Astrological Frameworks" (Flinton, 2026). The
paper measured cross-traditional agreement across 17 computational phases. Koiné
uses the elements with the strongest transmission signal and supplements them
with techniques documented in Hellenistic source texts that the paper's method
could not recover.

The design philosophy is **paper-informed, not paper-constrained**. The paper
measured what survived transmission in both descendant traditions. That method
is conservative by design: zero false positives, 40% false negatives. Techniques
that survived in only one descendant (and were therefore invisible to the paper's
two-descendant requirement) are included when they are documented in surviving
Hellenistic texts (Ptolemy's Tetrabiblos, Valens' Anthology). We have the
receipts. The ancestor had them.

This document maps each design decision to its evidence basis. It does not claim
the paper *proves* these elements are objectively correct — only that they are
the least culturally contingent, or are independently documented in the
historical record.

## What the Paper Missed

The paper compared Western and Vedic computational outputs. It required survival
in BOTH descendant traditions. This produced 9 correct recoveries and 6 false
negatives — techniques that existed in the Hellenistic original but were dropped
by one tradition and preserved by the other.

| Technique | Hellenistic source | Why the paper missed it |
|---|---|---|
| Sextile (60°) | Ptolemy I.13 | Vedic largely dropped it. Only survived in one descendant. |
| Triplicity rulers | Ptolemy I.18, Valens III | Vedic has a different triplicity system. No cross-tradition agreement. |
| Term/bound rulers | Ptolemy I.20-21, Valens I | Egyptian vs. Ptolemaic terms differ. No cross-tradition agreement. |
| Decan rulers | Ptolemy I.22 | Vedic decans (drekkana) use different rulers. No cross-tradition agreement. |
| Annual profections | Valens IV | Survived in Western, not in Vedic. Paper's Phase 4 found timing convergence is random (2.14%). |
| Zodiacal releasing | Valens V | Western-only. No Vedic equivalent. |

These are not speculative additions. They are documented features of the
original Hellenistic system. Koiné includes them as core techniques. The paper's
method was measuring agreement at the endpoints of transmission, not
reconstructing the full original. The false negatives are the cost of that
method. Koiné pays that cost by going to the source texts directly.

## Design Decisions

| Decision | Evidence Basis | Rationale |
|---|---|---|
| **10 real planets only** (Sun–Pluto) | All 17 paper phases use classical planets. Asteroids and Uranian/TNP points are 20th-century additions with zero cross-traditional signal. | If a body wasn't known to Hellenistic Alexandria, it cannot have survived transmission. Outer planets (Uranus, Neptune, Pluto) are included because they are physically real and were discovered, not invented. |
| **3-state dignity** (domicile, exaltation, fall) | Phase 1 (revised June 2026): per-state analysis shows domicile/swakshetra, exaltation/uchcha, and fall/neecha assignments are identical across Western and Vedic tables (100% agreement for all three states). | Domicile, exaltation, and fall are the three states with cross-traditional signal. Detriment is excluded — it is a Western-only innovation with zero Vedic equivalent. Outer planets (Uranus, Neptune, Pluto) have domicile assignments only; their exaltation and fall signs are modern inventions with no cross-traditional signal. |
| **Whole-sign houses** | Phase 3: house convergence across 8 systems is 88.7% at a 75% agreement threshold. Houses are the most stable structural element across traditions. | Whole-sign houses are the simplest and oldest system. The high convergence rate across quadrant systems confirms houses are not arbitrary — but the specific quadrant system you choose is. Whole-sign avoids that choice. |
| **5 aspects** (conjunction, sextile, square, trine, opposition) | Phase 2: conjunction, opposition, and trine are universal across Western, Vedic, and Chinese traditions. Square is universal in Western and Vedic, implicit in Chinese. Sextile appears in Western and Vedic only, and is documented in Ptolemy I.13. | The three universal aspects are non-negotiable. Square is included as the engine of dynamic tension. Sextile is included as a documented Hellenistic technique that survived in Western and Vedic. It is essential for pattern detection (Yod, Kite, Mystic Rectangle, Cradle). |
| **Triplicity rulers** (Ptolemaic, day/night) | Documented in Ptolemy I.18 and Valens III. The paper could not recover them because Vedic uses a different triplicity system. | Included as a documented Hellenistic technique. Planets in their own triplicity gain elemental support. Wired into planet-in-sign interpretation text. |
| **Egyptian terms (bounds)** | Documented in Ptolemy I.20-21 and Valens I. The paper could not recover them because Egyptian and Ptolemaic terms differ. | Included as a documented Hellenistic technique. Planets in their own terms gain authority over the degree they occupy. Wired into planet-in-sign interpretation text. |
| **Chaldean decans (faces)** | Documented in Ptolemy I.22. The paper could not recover them because Vedic drekkana use different rulers. | Included as a documented Hellenistic technique. Planets in their own decan gain reinforced presence. Wired into planet-in-sign interpretation text. |
| **Annual profections** | Documented in Valens IV. The paper's Phase 4 found timing convergence is random (2.14%) — timing systems are independent inventions per tradition. | Included as a documented Hellenistic technique. The profected house and its lord appear in the synthesis report as a "Timing" section. Profections are simple arithmetic (age % 12 + 1) and are the most transparent timing technique in the Hellenistic corpus. |
| **Pattern detection includes Node** | Phase 5: node sign survival is 18.7% — primarily a function of ayanamsa choice. The nodes are real astronomical points (lunar orbit intersections with the ecliptic). | Nodes are physically real, not culturally contingent. Including them in pattern detection captures real geometric structures that would otherwise be invisible. |
| **Fixed star conjunctions** | Phase 11: fixed star agreement is 100% — a geometry check. Angular distances between stars and planets are preserved under uniform coordinate shift. | Stars are physically real. Their inclusion is not an interpretive choice — it's a measurement. |
| **No lunar mansion integration** | Phase 8/8b: two-way mansion convergence is 2.2%, three-way convergence is 0%. Shared anchor stars do not produce shared mansion placements. | Mansion systems are independent — they share some anchor stars but define sector boundaries differently. There is no recoverable cross-system signal for mansion placement. |
| **Deterministic template-based interpretation** | Not a paper finding — a design philosophy. | Every planet-in-sign, planet-in-house, aspect, and pattern has a single authored interpretation. No randomization, no LLM generation, no subjective judgment at render time. The system is reproducible: same input → same output, every time. This makes it falsifiable — if an interpretation is wrong, it can be fixed at the source. |

## What Koiné Deliberately Excludes

| Exclusion | Reason |
|---|---|
| Asteroids (Ceres, Pallas, Juno, Vesta) | 19th-century discoveries. No cross-traditional signal. |
| Uranian/TNP points (Cupido, Hades, Zeus, Kronos, Apollon, Admetos, Poseidon, Vulkanus) | 20th-century mathematical constructs, not physical bodies. At orb=5 they produce 70+ patterns that are essentially permanent for everyone born within a decade — noise, not signal. |
| Lilith (Black Moon), Chiron, other minor bodies | Same rationale: no cross-traditional signal. |
| Detriment | Western-only innovation. Vedic astrology never developed a detriment concept. Zero cross-traditional signal. |
| Quincunx, semi-sextile, quintile, etc. | Only conjunction, opposition, trine, square, and sextile have cross-traditional signal or Hellenistic documentation. Minor aspects are modern elaborations. |
| Dasha systems, firdaria, zodiacal releasing | Timing systems show 2.10% cross-traditional agreement — random. Vedic dashas and Western zodiacal releasing are independent inventions. Annual profections are included because they are the simplest and most transparent Hellenistic timing technique. |
| Lunar mansion placement | 0% three-way convergence. |
| Degree symbols (Sabian, etc.) | Not falsifiable. Not computational. |
| LLM-generated interpretation | Not reproducible. Not falsifiable. |

## Architecture

Koiné is a fork of the Empirical engine. It shares the same Swiss Ephemeris
backend (CGo to libswe), the same Go module structure, and the same HTTP server
pattern. The fork point was the decision to build a synthesis system rather than
a measurement system.

```
cmd/recover/main.go          — CLI + HTTP server, all compute functions
internal/dignity/
  synthesis.go               — SynthesizeChart: opening, body sections, closing
  interpretation.go          — InterpretPlanetInSign, InterpretAspect, InterpretPattern
  planet_in_sign.go          — 144 authored interpretations (12 planets × 12 signs)
  planet_in_house.go         — 144 authored interpretations (12 planets × 12 houses)
  pattern.go                 — DetectPatterns (10-planet map, includes Node)
  stars.go, arabic_parts.go  — StarConjunction, Fortune
internal/server/server.go    — HTTP handlers, orb defaults
web/index.html               — Dashboard
```

## Version

This document describes Koiné as of June 2026. The Hellenistic integration
(triplicity rulers, terms, decans, annual profections as core features) was
completed in this revision.
