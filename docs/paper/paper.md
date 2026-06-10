# Computational Recovery of Astrological Invariants

A.J. Flinton

June 2026

## Abstract

Around 150 BCE, Babylonian planetary astronomy, Egyptian decanic timekeeping, and Greek geometry were combined in Alexandria into the first horoscopic astrology. The system spread to Rome and India, where it evolved largely in isolation for two millennia. Elements reached China by the seventh century CE. This paper measures what structural features of the original synthesis survived.

A single Go binary with a statically linked Swiss Ephemeris computes convergence across Western and Vedic traditions for dignity rules, house division, node axis preservation, and zodiac comparison. A static catalog comparison adds Chinese aspect data. A cross-system timing measurement includes Chinese Ba Zi alongside Vedic and Hellenistic timing. Each measurement includes a Monte Carlo random baseline.

Results form a spectrum. Three aspect angles are universal. House division holds at 79.4% across five methods (95% CI: 78.8-80.1). Node axis geometry is invariant. Dignity rules average 46.7% per chart (95% CI: 46.3-47.1) with a smooth distribution and no family patterning. Timing systems show 64.5% partial overlap, with full three-way agreement at only 4.5% (95% CI: 3.7-5.2). The node sign flips for 78% of people under ayanamsa shift. Family synastry is null.

Comparison of 27 Indian nakshatras and 23 Chinese xiu with documented single-star determinants reveals 9 shared anchor stars. Under brightness-weighted random selection from the combined pool of 41 stars, the null expectation is 8.5 matches (p = 0.75 for observed). However, the faint-star composition is extreme: 6 of 9 shared stars are second magnitude or fainter, where the null expects 1.0 (p = 0.0002). The total overlap is consistent with common origin. The faint-star concentration is quantitative evidence for it.

A weekly Uranian transit signal on SPY (262 weeks, 2021-2026) shows a significant effect (p = 0.017). This analysis is reported in Supplementary Material.

The engine, data, baselines, and manuscript are open source. All phases are cross-validated between Go and Python implementations.

## 1. Introduction

Astrology begins with data. Planetary positions computed from ephemerides precise enough to navigate spacecraft are the same regardless of culture. A conjunction is a conjunction. The geometry does not care what meaning a tradition assigns to it.

Meaning is where the traditions diverge. Western astrology assigns planetary dignity by domicile and exaltation. Vedic astrology uses swakshetra and uchcha. Similar concepts, different assignments. Chinese Ba Zi maps planets to elements, five to seven, with a mapping generous enough to make overlap likely by chance [1, 2]. Three systems. One shared origin. Roughly 2,000 years of largely separate development. The traditions were not fully independent: secondary contact occurred via the Silk Road after the Indian branch had already diverged [1, 6].

The question is not whether astrology is true. It asks what structural features of the original Hellenistic synthesis survived. The question is computational.

The tool is a Go binary using the Swiss Ephemeris C library, which provides the same JPL DE ephemeris data used for spacecraft navigation [3]. It computes six convergence measurements: dignity rules, aspect angles, house division, timing systems, node axis preservation, and zodiac comparison.

Two traditions are compared computationally for dignity, houses, nodes, and zodiac: Western and Vedic [4, 5]. The Chinese tradition is included in the aspect catalog (Phase 2) and timing measurement (Phase 4), where Ba Zi luck pillars are compared alongside Vedic dasha and Hellenistic profections. Per-chart computation for Chinese dignity or houses is not possible because those categories do not map directly to the five-element framework [1, 6].

Beneath the planetary layers is an older system: lunar mansions, predating horoscopic astrology by at least a millennium. Comparison of the 27 Indian nakshatras and 23 documented Chinese xiu reveals 9 shared anchor stars. Under brightness-weighted null models, the total overlap count is unremarkable. But six of the nine are too faint for independent discovery, a concentration that exceeds random expectation by two orders of magnitude (p = 0.0002).

The engine does not say whether astrology works. It says what survived.

## 2. Methods

### 2.1 Computational Phases

**Phase 1: Dignity convergence.** Seven classical planets are classified under Western rules (domicile, exaltation, detriment, fall) and Vedic rules (swakshetra, uchcha, neecha) [4, 5]. Agreement on shared categories constitutes signal. Western astrology recognizes four dignity states, Vedic three. Convergence is assessed as agreement on domicile/swakshetra and exaltation/uchcha, with peregrine and detriment/neecha treated as a single non-dignified state. Different alignment choices would produce different numbers.

**Phase 2: Aspect catalog.** Static comparison of seven major angles (0, 30, 60, 90, 120, 150, 180 degrees) across Western, Vedic, and Chinese traditions using the Brihat Parashara Hora Shastra, Ptolemy's Tetrabiblos, and the San Ming Tong Hui [2, 8, 9].

**Phase 3: House convergence.** Five methods (whole sign, equal, Placidus, Porphyry, Koch) applied to each of seven planets. A planet is unambiguous if four or more systems assign the same house number.

**Phase 4: Timing convergence.** Vimshottari dasha (Vedic), Ba Zi luck pillars (Chinese), and Hellenistic annual profections are compared for a given target date. Each system maps its active period to a set of planets. A planet appearing in two or more systems is a convergence. To enable cross-system comparison, element-to-planet mapping follows a deliberately broad assignment: Metal to Saturn and Venus, Water to Moon and Mercury, Wood to Jupiter, Fire to Mars and Sun [1]. This is not the standard Ba Zi mapping (which assigns one planet per element), but a generous interpretation to avoid false negatives. The Hellenistic profection is the root system; Vimshottari and Ba Zi are divergent elaborations.

**Phase 5: Node convergence.** Tropical and sidereal node positions compared using the Lahiri ayanamsa, the standard adopted by the Indian government and the Swiss Ephemeris default [10].

**Phase 6: Zodiac comparison.** Dignity density under tropical and sidereal coordinates. Serves as a sanity check confirming the analytical expectation of symmetry.

### 2.2 Random Baseline Generation

Monte Carlo simulation with fixed seed 42. Random dates 1900-2030, latitudes -60 to 60, longitudes -180 to 180, timezone offsets -12 to +12 hours in 0.5 hour increments, local times 00:00-23:59. Sample sizes: Phase 1: 10,000, Phase 3: 5,000, Phase 4: 3,000, Phase 5: 5,000, Phase 6: 9,738 synthetic charts. Phase 4 additionally randomizes target dates. 95% CIs computed as mean plus or minus 1.96 standard errors.

### 2.3 Cross-Validation

A Python reference implementation (pyswisseph 2.08) validates the Go engine against three reference charts. Go and Python produce identical scores for all phases. Transit and synastry computations validated similarly.

### 2.4 Family Dataset

A 17-person, three-generation sample scored for individual chart analysis and synastry. Includes three couples, 14 parent-child pairs, 24 grandparent-grandchild pairs. Categories with single observations excluded. Age-matched random baseline of 5,000 pairs per category. Scores reported for illustration only; sample insufficient for population inference.

### 2.5 Lunar Mansion Null Models

Three null models test whether the 9 observed shared stars (33% overlap) exceed random expectation. All models use 10,000 bootstrap iterations.

**Null 1 (Uniform):** Each culture selects 27 stars uniformly from the combined pool of 41 unique stars from both systems. Expected overlap: 27 * 23 / 41 = 15.1. Bootstrap mean: 18.4 (95% CI: 16-21). Observed 9 is well below expectation (p = 0.0000).

**Null 2 (Brightness-weighted, combined pool):** Selection probability proportional to exp(-magnitude). Brighter stars are proportionally more likely to be selected by either culture. Bootstrap mean: 8.5 (95% CI: 5-12). Observed 9 is consistent with this model (p = 0.75). Faint stars (magnitude greater than or equal to 2.5): null mean 1.0 (95% CI: 0-3). Observed 6 faint stars yields p = 0.0002.

**Null 3 (Brightness-weighted, own pools):** Each culture selects from their own documented pool (27 nakshatra stars, 23 xiu stars), weighted by brightness. Bootstrap mean: 3.3 (95% CI: 1-5). Observed 9 exceeds expectation (p = 0.0000). Faint stars: null mean 0.9 (95% CI: 0-3). Observed 6 yields p = 0.0000.

Null 2 is the most reasonable model: both cultures independently selected bright stars from the available celestial candidates. The total overlap is unremarkable. The faint-star composition is not.

### 2.6 Limitations

The engine measures convergence between Western and Vedic traditions for Phases 1, 3, 5, and 6. Phase 4 compares all three traditions. Phase 2 is a static catalog. Chinese dignity and houses cannot be computed per chart due to categorical mismatch with the five-element framework.

Phase 4's 64.5% baseline is inflated by the generous element-to-planet mapping. A tighter mapping would produce a lower baseline but risk false negatives.

Phase 1 baseline was computed in Go. Phases 3, 4, and 5 baselines were computed in Python. Full reproduction requires both environments.

The traditions were not fully independent: Silk Road contact occurred after the Indian branch had already diverged [1, 6]. Chinese-Vedic convergence may reflect later contact.

Convergence may arise from astronomical data rather than cultural transmission. Conjunction and opposition are physically real alignments. The paper reports convergence. The interpretation is separate.

Traditions use different numbers of categories for comparable concepts. Phase 1 maps four-state Western dignity onto three-state Vedic. Phase 4 maps five-element Chinese onto seven classical planets. The reported rates are conditioned on the mapping choices described above.

### 2.7 Code and Data Availability

Open source under MIT license at github.com/aj-nt/empirical. Go source: 89 test functions, all passing. Python reference: 1,185 test functions. Baseline scripts included. Swiss Ephemeris via github.com/aloistr/swisseph (GPL). Paper licensed under CC-BY 4.0.

## 3. Historical Background

Organized celestial interpretation emerged in Mesopotamia by approximately 1600 BCE. The Enuma Anu Enlil, 70 cuneiform tablets, tracked the seven visible planets as divine messengers [5]. The Mul.Apin star catalog (c. 1000 BCE) catalogued 66 stars [5]. By the fourth century BCE, Babylonian astronomers computed planetary positions with sufficient accuracy to produce mathematical ephemerides [3].

Other cultures developed independent frameworks. Egyptian coffin lids from approximately 2100 BCE depicted 36 decans dividing the year into ten-day periods [6]. The Vedanga Jyotisha (1400-1200 BCE) described 27 nakshatras for tracking the lunar month [7, 11]. Shang oracle bones (c. 1250 BCE) recorded 10 Heavenly Stems and 12 Earthly Branches alongside 28 xiu lunar mansions [1, 6]. None produced horoscopic astrology. They were timekeeping and calendrical systems.

Around 150 BCE, in Alexandria, Babylonian astronomy, Egyptian decans, and Greek geometry were combined into the first horoscopic system: ascendant, houses, dignity rules, and aspect angles [4, 5]. The system spread west into Rome and east into India (first-second century CE), merging with the nakshatras to produce Vedic astrology [4, 5]. Elements reached Tang dynasty China (seventh century CE), where Ba Zi integrated the planetary framework with indigenous five-element cosmology [1, 6].

The lunar mansions predate this synthesis. Both the 27 nakshatras and 28 xiu divide the ecliptic into sectors based on the sidereal month, each with a determinative star. Both are documented before 1200 BCE [6, 7, 11]. Comparing determinative stars across the 27 nakshatras and 23 xiu with documented single-star determinants reveals 9 shared anchor stars: Sheratan (Beta Arietis) anchors Ashwini and Lou (婁). 35 Arietis anchors Bharani and Wei (胃). The Pleiades anchor Krittika and Mao (昴). Spica (Alpha Virginis) anchors Chitra and Jiao (角). Antares (Alpha Scorpii) anchors Jyeshtha and Xin (心). Markab (Alpha Pegasi) anchors Purva Bhadrapada and Shi (室). Algenib (Gamma Pegasi) anchors Uttara Bhadrapada and Bi (壁). Meissa (Lambda Orionis) anchors Mrigashira and Zi (觜). Zubenelgenubi (Alpha Librae) anchors Vishakha and Di (氐). Three of these (Spica, Antares, Pleiades) are first magnitude or brighter. Six are second magnitude or fainter.

## 4. Results

### 4.1 Phase 1: Dignity Convergence

Random baseline (N=10,000): mean 3.27/7 planets agreeing (46.7%, 95% CI: 46.3-47.1, sample SD 1.52). Distribution: 29.0% at exactly 3, 25.6% at 4, 19.8% at 2, 1.1% at 0, 0.3% at 7.

Family sample (17 people): scores ranged 14.3% to 71.4%, consistent with random. No clustering by generation, lineage, or biological relationship. The four charts at 1/7 (percentile 1.1-8.1) span three lineages: wife, her mother, paternal grandmother, maternal grandmother. No common descent.

### 4.2 Phase 2: Aspect Catalog

Three angles are universal: conjunction (0 degrees), opposition (180 degrees), trine (120 degrees). Square (90 degrees) is Western/Vedic explicit, Chinese implicit. Sextile (60 degrees) is Western/Vedic only. Semi-sextile and quincunx are Western/Vedic with partial Chinese equivalents.

### 4.3 Phase 3: House Convergence

Random baseline (N=5,000): mean 5.56/7 unambiguous (79.4%, 95% CI: 78.8-80.1, sample SD 1.71). Right-skewed: 39.1% of charts have all seven unambiguous. Weak differentiator; planets near cusps are the only disagreement source.

### 4.4 Phase 4: Timing Convergence

Random baseline (N=3,000): 64.5% of pairs produce at least one converging planet. Distribution: 55.8% exactly one, 35.5% zero, 8.7% two. None observed at three or more; the convergence count is structurally bounded by seven planets. Full three-system agreement: 4.5% (134 of 2,999 valid records after excluding one chart for computation error, 95% CI: 3.7-5.2%).

### 4.5 Phase 5: Node Convergence

Random baseline (N=5,000): node sign survives tropical-to-sidereal shift in 22.2% (95% CI: 21.0-23.4). Opposition axis preserved in 100%.

### 4.6 Phase 6: Zodiac Comparison

9,738 synthetic charts confirm symmetry. Neither tropical nor sidereal zodiac produces systematically more dignified placements. Confirms the analytical expectation that the dignity table is invariant under uniform sign shift.

### 4.7 Lunar Mansions

Nine shared anchor stars (33% overlap) between 27 nakshatras and 23 xiu with documented single-star determinants. Three null models tested at 10,000 bootstrap iterations each.

Under Null 1 (uniform selection, combined pool), the expected overlap is 18.4 (CI: 16-21). Observed 9 is below expectation.

Under Null 2 (brightness-weighted, combined pool), the expected overlap is 8.5 (CI: 5-12). Observed 9 is consistent with the null (p = 0.75). The total count is unremarkable: both cultures independently selecting bright stars from the available celestial pool would produce roughly this many matches.

However, Null 2 predicts only 1.0 faint star matches on average (CI: 0-3). The observed 6 faint matches is extreme (p = 0.0002). Six of the nine shared stars (Sheratan at 2.6 mag, Markab at 2.5, Zubenelgenubi at 2.6, Algenib at 2.8, Meissa at 3.5, 35 Arietis at 4.6) are too dim to be plausible independent anchor choices under brightness-weighted selection.

Under Null 3 (brightness-weighted, own pools), expected overlap drops to 3.3 (CI: 1-5). Observed 9 far exceeds expectation (p = 0.0000). Faint stars: expected 0.9 (CI: 0-3), observed 6 (p = 0.0000). This model is conservative but assumes each culture had access to the same 27 and 23 star pools respectively, which biases against overlap.

The conclusion is that the total overlap count does not distinguish signal from noise under the most reasonable null model. The composition of the overlap does. Two cultures independently selecting bright stars should share roughly eight stars, with perhaps one of them faint. Instead they share nine, with six faint. The faint-star concentration is the quantitative evidence for common origin.

### 4.8 Family Synastry

Null result at 8 degree orb. Couples: 35.8 vs 35.8 random (SD 4.5). Parent-child: 37.9 vs 35.3 (SD 4.5). Grandparent-grandchild: 37.1 vs 37.0 (SD 4.7). No category deviates by more than 0.6 aspects. Aspect-density does not carry relationship information.

## 5. Discussion

### 5.1 Spectrum of Preservation

The results form a continuum. Three aspect angles are universal. Houses hold at 79.4%. Node axis is geometric. Dignity is partial at 46.7% with wide individual variation. Timing shows 64.5% partial overlap with 4.5% full agreement. Node sign is coordinate-dependent. Family synastry is null.

### 5.2 Lunar Mansion Layer

The quantitative signal is in the faint-star composition, not the total count. Under brightness-weighted null models, the expected number of faint-star matches is approximately one. The observed six is a two-order-of-magnitude deviation (p = 0.0002). This constitutes quantitative evidence for a common origin: two cultures independently selecting bright stars would not produce six faint matches by chance.

Three millennia separate the Shang oracle bones from the present. Nine stars survived the drift. The Hellenistic fusion grafted a planetary system onto a lunar foundation already ancient when the first horoscope was drawn.

A computational nakshatra-xiu module could extend this to per-chart measurement of sector boundary alignment and systematic offset.

### 5.3 What the Engine Doesn't Measure

The engine measures structural convergence. It does not measure whether astrology works. A dignity rule (Sun domicile in Leo) is structural. The interpretation is not. The engine reports agreement rates. It cannot assess correspondence to external reality.

Convergence may arise from the data rather than the tradition. Conjunction and opposition are physically real. Twelve divisions of a circle are mathematically natural. The engine cannot distinguish transmission from independent discovery of the same astronomical facts.

### 5.4 Future Work

A nakshatra-xiu computational module for per-chart sector boundary comparison. A star-magnitude-aware null model incorporating ecliptic position to test whether the faint-star concentration survives controls for sector location. Timing baseline sharpened by restricting element-to-planet mapping. Additional coordinate systems (draconic, heliocentric, alternative ayanamsa values). Larger family datasets for synastry subgroup analysis.

## 6. Conclusion

A single Go binary, six convergence measurements, and a question from 150 BCE.

Three aspect angles are universal. Twelve houses hold. The node axis is geometry. Dignity is partially preserved at 46.7%. Timing systems share 64.5% partial overlap but converge fully only 4.5% of the time. Family synastry is indistinguishable from random.

Beneath it all: nine shared anchor stars between Indian nakshatras and Chinese xiu. The total overlap is unremarkable under brightness-weighted null (p = 0.75). The composition is not: six faint stars where null expects one (p = 0.0002). Whatever astronomical tradition existed before writing, before horoscopes, before Alexandria, it left trace evidence in the stars two cultures independently chose. Two thousand years after the Hellenistic fusion, the oldest structural layer of astrology is the one that survived best.

The engine, data, baselines, and manuscript are open source. The measurements are falsifiable. The engine does not say whether any of this is true, only what survived.

## References

[1] Needham, J. (1959). Science and Civilisation in China, Vol. 3. Cambridge University Press.

[2] Pingree, D. (1978). The Yavanajataka of Sphujidhvaja. Harvard Oriental Series, Vol. 48.

[3] Standish, E.M. (1998). JPL Planetary and Lunar Ephemerides, DE405/LE405. JPL IOM 312.F-98-048.

[4] Pingree, D. (1997). From Astral Omens to Astrology: From Babylon to Bikaner. Istituto Italiano per l'Africa e l'Oriente.

[5] Hunger, H. and Pingree, D. (1999). Astral Sciences in Mesopotamia. Brill.

[6] Pingree, D. (1963). Astronomy and Astrology in India and Iran. Isis, 54(2), 229-246.

[7] Subbarayappa, B.V. and Sarma, K.V. (1985). Indian Astronomy: A Source Book. Nehru Centre.

[8] Ptolemy, C. (c. 150 CE). Tetrabiblos. Translated by F.E. Robbins (1940). Loeb Classical Library.

[9] Wan, M. (1998). San Ming Tong Hui (三命通会). Ming Dynasty compilation.

[10] Lahiri, N.C. (1985). Lahiri's Indian Ephemeris of Planets' Positions. Astro-Research Bureau.

[11] Pingree, D. (1981). Jyotihsastra: Astral and Mathematical Literature. Otto Harrassowitz.
