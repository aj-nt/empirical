# Computational Recovery of Astrological Invariants

A.J. Flinton

June 2026

## Abstract

Around 150 BCE, three pre-existing astronomical traditions were combined in Alexandria into the first horoscopic astrology. The resulting system spread to Rome and India, where it evolved largely in isolation for two millennia. Elements reached China by the seventh century CE. This paper measures what structural features of the original synthesis survived the transmission.

A single Go binary with a statically linked Swiss Ephemeris computes convergence across Western and Vedic traditions for dignity rules, house division, timing systems, node axis preservation, and zodiac comparison. A static catalog comparison adds Chinese aspect data. Each measurement includes a Monte Carlo random baseline.

Three aspect angles (conjunction, opposition, trine) are universal across all three traditions. The twelve-house division holds with 79.4% agreement across five competing methods (95% CI: 78.8-80.1). The node axis geometry is mathematically invariant. Dignity convergence averages 46.7% per chart (95% CI: 46.3-47.1) with a smooth distribution and no family patterning.

The surface layer diverged. Three timing systems produce the same answer 4.5% of the time (95% CI: 3.7-5.2). The node sign flips for 78% of people under tropical-to-sidereal conversion. Family members and strangers produce indistinguishable synastry aspect counts.

Beneath the Hellenistic layer, comparison of the 27 Indian nakshatras and 28 Chinese xiu reveals 9 shared anchor stars (33% overlap). Bootstrap resampling yields a null expectation of 10.0 matches (95% CI: 22-52%), placing the observed overlap within random expectation (p = 0.76). However, six of the nine shared stars are too faint for independent discovery. A formal null model incorporating stellar magnitude is needed to quantify this qualitative signal.

A weekly Uranian transit signal on SPY volatility (262 weeks, 2021-2026) shows a statistically significant effect with a priori planet classification (p = 0.017). This analysis uses a different astrological framework and is reported in Appendix A.

The engine, data, baselines, and manuscript are open source. All six phases are cross-validated between Go and Python implementations against three reference charts.

## 1. Introduction

Astrology begins with data. Planetary positions computed from ephemerides precise enough to navigate spacecraft are the same regardless of culture. A conjunction is a conjunction. The geometry does not care what meaning a tradition assigns to it.

Meaning is where the traditions diverge. Western astrology assigns planetary dignity by domicile and exaltation. Vedic astrology uses swakshetra and uchcha. Similar concepts, different assignments. Chinese Ba Zi maps planets to elements, five to seven, with a mapping generous enough to make overlap likely by chance [1, 2]. Three systems. One shared origin. Roughly 2,000 years of largely separate development. The traditions were not fully independent: secondary contact occurred via the Silk Road after the Indian branch had already diverged, a confound addressed in the limitations [1, 6].

The question this paper asks is not whether astrology is true. It asks what structural features of the original Hellenistic synthesis survived the transmission. The question is computational. It can be answered with a measurement tool.

The tool is a Go binary using the Swiss Ephemeris C library, which provides the same JPL DE planetary ephemeris data used for spacecraft navigation [3]. It computes six convergence measurements: dignity rules, aspect angles, house division, timing systems, node axis preservation, and zodiac comparison. Each measurement includes a Monte Carlo random baseline.

Two traditions are compared computationally: Western (the Roman elaboration of the Hellenistic synthesis) and Vedic (the Indian branch, merged with the pre-existing nakshatra lunar mansion system) [4, 5]. The Chinese tradition, which preserved some structural elements through its own indigenous framework, is included in the static aspect catalog comparison (Phase 2) but is not computed per chart for dignity or houses because those categories do not map directly to Chinese equivalents [1, 6].

The results are uneven. Some features survived. Most did not. Beneath both systems is an older layer: the lunar mansions, predating horoscopic astrology by at least a millennium, whose shared anchor stars between Indian nakshatras and Chinese xiu suggest a common origin [6, 7].

The engine does not say whether astrology works. It says what survived. The distinction matters. This paper measures the structural integrity of a 2,000-year-old transmission and reports the result.

## 2. Methods

### 2.1 Computational Phases

The engine computes six independent measurements. Each addresses a single structural question.

**Phase 1: Dignity convergence.** Compares planetary dignity assignments across Western and Vedic traditions. Seven classical planets (Sun, Moon, Mercury, Venus, Mars, Jupiter, Saturn) are classified under Western rules (domicile, exaltation, detriment, fall) and Vedic rules (swakshetra, uchcha, neecha) [4, 5]. The two tables are compared. Agreement counts as signal. The convergence rate is the proportion of planets where both systems agree on the dignity category.

**Phase 2: Aspect catalog.** Compares recognized aspect angles across Western, Vedic, and Chinese traditions [1, 2, 8]. This is a static catalog, not a per-chart computation: it maps the seven major angles (0, 30, 60, 90, 120, 150, 180 degrees) against their treatment in each tradition. An angle is universal if explicitly recognized in all three. It is partial if recognized in two. It is single-source if recognized in one. Source texts are the standard dignity and aspect tables as documented in the Brihat Parashara Hora Shastra, the Hellenistic Tetrabiblos, and the Ming dynasty San Ming Tong Hui [2, 8, 9].

**Phase 3: House convergence.** Measures agreement on house placement across five methods: whole sign, equal, Placidus, Porphyry, and Koch. For each of the seven classical planets, house numbers are computed under all five systems. If four or more systems assign the same house number, the planet is unambiguous. The convergence rate is the proportion of unambiguous planets.

**Phase 4: Timing convergence.** Compares active planets across three timing systems for a given target date: Vimshottari dasha (Vedic, Moon-nakshatra based, 120-year cycle), Ba Zi luck pillars (Chinese, stem-branch based, 10-year cycle), and Hellenistic annual profections (sign-house cycle, 1 year). Note the asymmetry: the Hellenistic profection is a direct descendant of the original synthesis, while Vimshottari and Ba Zi are divergent elaborations. Each system maps its active period to a set of planets. A planet appearing in two or more systems on the same target date is a convergence. The Ba Zi element-to-planet mapping follows standard assignments: Metal to Saturn and Venus, Water to Moon and Mercury, Wood to Jupiter, Fire to Mars and Sun [1].

**Phase 5: Node convergence.** Measures whether the lunar node axis preserves its sign assignment under ayanamsa shift. North Node and South Node positions are computed in tropical coordinates, then converted to sidereal by subtracting the Lahiri ayanamsa [10]. The Lahiri value was selected because it is the standard ayanamsa adopted by the Indian government for national ephemerides and is the default in the Swiss Ephemeris library. The 180 degree opposition angle is also measured.

**Phase 6: Zodiac comparison.** Computes dignity density under tropical and sidereal coordinate systems. For each of the seven classical planets, Western dignity rules are applied to the tropical longitude and to the sidereal longitude. Detriment is counted as peregrine per the Phase 1 finding. This phase confirms the analytical expectation that the dignity table, structured as domicile pairs in opposite signs, is symmetric under uniform sign shift. It serves as a sanity check rather than a novel measurement.

### 2.2 Random Baseline Generation

Each phase has a null hypothesis: the measured convergence under random birth data.

Baselines are generated by Monte Carlo simulation. Random dates are drawn uniformly from 1900 to 2030. Random geographic coordinates from latitude -60 to 60 and longitude -180 to 180. Random timezone offsets from -12 to +12 hours in 0.5 hour increments. Random local times from 00:00 to 23:59. Fixed seed 42 for reproducibility.

Sample sizes vary by phase: Phase 1 used 10,000 random charts. Phase 3 used 5,000. Phase 4 used 3,000 (fewer due to the cost of three timing systems per chart). Phase 5 used 5,000. Phase 2 is static. Phase 6 used 9,738 synthetic charts. The Phase 4 baseline additionally randomizes the target date, drawing from birth date to birth date plus 90 years.

95% confidence intervals for mean convergence rates are computed as mean plus or minus 1.96 standard errors. Population SD values are estimated from the baseline runs and reported in the Results section alongside each measurement.

### 2.3 Cross-Validation

A Python reference implementation using pyswisseph (version 2.08) validates the Go engine. Three reference charts with known birth data were computed independently in both codebases. The Go and Python implementations produce identical dignity convergence scores, aspect catalogs, house agreements, timing periods, node signs, and zodiac comparisons for all three charts.

Transit and synastry computations were validated similarly. The transit engine was confirmed against Python output for a two-week window. The synastry engine was validated for one known pair.

### 2.4 Family Dataset

A 17-person, three-generation family sample was scored for individual chart analysis and synastry comparison. The sample includes three couples, 14 parent-child pairs, 24 grandparent-grandchild pairs, and one sibling pair. Birth data (date, time, location) for all 17 individuals is recorded in the source repository. For the synastry analysis, two categories with single observations (step-parent, sibling) are excluded from aggregate comparison due to sample size. Individual scores from this sample are reported in the Results for illustrative purposes only; the sample is not large enough for population inference.

An age-matched random baseline of 5,000 pairs per relationship category was generated. Random pairs are drawn with birth-year gaps matching the real pairs within each category (couples: 0-10 year difference, parent-child: 20-40 years, grandparent-grandchild: 50-90 years). Aspect counts are computed at an 8 degree orb for all planet-planet pairs.

### 2.5 Limitations

The engine measures convergence between Western and Vedic traditions for Phases 1, 3, 5, and 6, and between all three traditions for Phase 4. The Chinese tradition is included only in the static Phase 2 aspect catalog because dignity and house categories do not map directly to the Chinese five-element framework. This is not a limitation of the Chinese system. It is a category mismatch: the traditions diverged enough that per-chart convergence cannot be computed for dignity or houses between Chinese and Western/Vedic.

Second, the Phase 4 timing convergence baseline is high (64.5% of random pairs show at least one converging planet) because the Ba Zi element-to-planet mapping is generous. Metal maps to both Venus and Saturn. Water maps to both Moon and Mercury. This broad mapping inflates overlap probability [1].

Third, the Phase 1 baseline was computed in Go using the statically linked Swiss Ephemeris. The Phase 3, 4, and 5 baselines were computed in Python using the reference implementation. Reproducing the full set requires both environments.

Fourth, the traditions were not fully independent after the initial transmission. Elements of Hellenistic and Vedic astrology reached China via the Silk Road centuries after the Indian branch had already diverged [1, 6]. Convergence between Chinese and Vedic systems may reflect later contact rather than preservation from the original synthesis.

Fifth, the assumption that convergence indicates preservation of an original feature, rather than independent convergence on the same astronomical data, is untestable with the current method. Conjunction and opposition are physically significant alignments regardless of cultural transmission. Twelve equal divisions of a circle are mathematically natural. Some convergence may reflect the structure of the sky rather than the structure of the tradition. The paper reports convergence. The interpretation of why it converges is a separate question.

### 2.6 Code and Data Availability

The engine is open source under the MIT license at github.com/aj-nt/empirical. The Go source has 89 test functions, all passing. The Python reference implementation has 1,185 test functions. Baseline scripts for Phases 3, 4, and 5 are included in the repository. Phase 1 and 6 baselines require the Go toolchain.

The Swiss Ephemeris C library is available from github.com/aloistr/swisseph (GPL or commercial license from Astrodienst). JPL DE ephemeris data is public domain. The Go wrapper is MIT licensed.

This paper is licensed under CC-BY 4.0.

## 3. Historical Background

### 3.1 Early Celestial Tracking

Organized celestial interpretation emerged in Mesopotamia by approximately 1600 BCE. The Enuma Anu Enlil, 70 cuneiform tablets containing roughly 7,000 celestial omens, tracked the seven visible planets as divine messengers [5]. The Mul.Apin star catalog, compiled around 1000 BCE, catalogued 66 stars and constellations with heliacal rising and setting dates [5]. By the fourth century BCE, Babylonian astronomers were computing planetary positions with sufficient accuracy to produce mathematical ephemerides [3]. The system was mundane throughout: omens for state purposes, not individual horoscopes.

Other cultures developed independent frameworks. The Egyptians aligned temples to solstices and tracked Sirius for Nile flood prediction. By approximately 2100 BCE, coffin lids depicted the decanic system: 36 star groups dividing the year into ten-day periods [6]. The Vedanga Jyotisha, dated to 1400-1200 BCE, described 27 nakshatras dividing the ecliptic into sectors of 13 degrees 20 minutes for tracking the lunar month [7, 11]. Shang dynasty oracle bones (approximately 1250 BCE) recorded 10 Heavenly Stems and 12 Earthly Branches for day counting, alongside 28 xiu lunar mansions [1, 6].

None of these pre-Hellenistic traditions produced horoscopic astrology. The systems were timekeeping and calendrical.

### 3.2 The Hellenistic Transmission

Around 150 BCE, in Alexandria, Babylonian mathematical astronomy, Egyptian decanic timekeeping, and Greek geometry were combined [4, 5]. The result was the first horoscopic astrology: an ascendant marking the eastern horizon, 12 houses dividing the local sky, essential dignity rules, and aspect angles. The seven visible planets were mapped into this new structure.

The system spread west into Rome (becoming Western astrology) and east along trade routes into India around the first or second century CE, merging with the nakshatra system to produce Vedic astrology [4, 5]. Elements reached China during the Tang dynasty (seventh century CE), where Ba Zi integrated the imported framework with indigenous stem-and-branch counting and five-element cosmology [1, 6].

The question is not whether the system spread. It did. The question is what structural features survived.

### 3.3 Lunar Mansions: A Deeper Layer

The 27 nakshatras and 28 xiu predate horoscopic astrology. Both are ecliptic division systems based on the sidereal month (roughly 13 degree sectors), each with a determinative star. Both are documented before 1200 BCE [6, 7, 11].

Comparing determinative stars reveals 9 shared anchor stars across the two systems (33% overlap): Sheratan (Beta Arietis) anchors both Ashwini and Lou (婁, Pinyin: Lou). 35 Arietis anchors both Bharani and Wei (胃, Pinyin: Wei). The Pleiades anchor both Krittika and Mao (昴, Pinyin: Mao). Spica (Alpha Virginis) anchors both Chitra and Jiao (角, Pinyin: Jiao). Antares (Alpha Scorpii) anchors both Jyeshtha and Xin (心, Pinyin: Xin). Markab (Alpha Pegasi) anchors both Purva Bhadrapada and Shi (室, Pinyin: Shi). Algenib (Gamma Pegasi) anchors both Uttara Bhadrapada and Bi (壁, Pinyin: Bi). Meissa (Lambda Orionis) and Zubenelgenubi (Alpha Librae) also appear in both lists.

Three of the shared stars (Spica, Antares, Pleiades) are first magnitude or brighter and could be discovered independently. Six (Sheratan at 2.6 mag, 35 Arietis at 4.6, Meissa at 3.5, Zubenelgenubi at 2.6, Markab at 2.5, Algenib at 2.8) are second magnitude or fainter. The formal overlap count does not distinguish signal from chance (see Results 4.7). A star-magnitude-aware null model is needed to test whether the faint-star concentration exceeds random expectation.

The lunar mansions predate horoscopic astrology by at least a millennium. They form a parallel system built on the sidereal month. The Hellenistic fusion grafted a planetary system onto a lunar foundation that was already ancient.

### 3.4 Measurement Framework

The engine measures convergence across the planetary and horoscopic layers that originated in the Hellenistic synthesis and spread to India. Dignity rules, aspect angles, house division, timing systems, and coordinate frames are later additions to an older lunar infrastructure. The nakshatra-xiu comparison provides a window into that deeper layer.

## 4. Results

### 4.1 Phase 1: Dignity Convergence

The random baseline, computed from 10,000 charts (1900-2030), yields a mean of 3.27 planets agreeing out of 7 (46.7%, 95% CI: 46.3-47.1, population SD 1.52). The distribution is a bell curve centered on 3 of 7: 29.0% of charts at exactly 3, 25.6% at 4, 19.8% at 2. At the extremes, 1.1% show 0 and 0.3% show all 7.

A 17-person family dataset spanning three generations was scored. Scores ranged from 1 of 7 (14.3%) to 5 of 7 (71.4%), consistent with the random distribution. Family members scatter across the full range with no clustering by generation, lineage, or biological relationship.

The four charts scoring 1 of 7 (cumulative percentile 1.1-8.1) are the wife, her mother, the paternal grandmother, and the maternal grandmother. These four span three different lineages with no common descent. The two highest at 5 of 7 (above the 82nd percentile) are from different generations and unrelated lineages. The reference male subject scores 4 of 7 (between the 57th and 82nd percentile), indistinguishable from random.

### 4.2 Phase 2: Aspect Catalog

Three angles are universal across Western, Vedic, and Chinese traditions: conjunction (0 degrees), opposition (180 degrees), and trine (120 degrees). Two are partially preserved: the square (90 degrees) appears explicitly in Western and Vedic, implicitly in Chinese through the punishment relationship. The sextile (60 degrees) appears in Western and Vedic only. The semi-sextile (30 degrees) and quincunx (150 degrees) are Western/Vedic with partial Chinese equivalents through the six harmonies subset [2, 8, 9].

### 4.3 Phase 3: House Convergence

The random baseline (N=5,000) yields a mean of 5.56 unambiguous planets out of 7 (79.4%, 95% CI: 78.8-80.1, population SD 1.71). The distribution is right-skewed: 39.1% of charts have all seven unambiguous. House convergence is a weak differentiator between individual charts. Planets near cusp boundaries are the only source of disagreement.

The reference male and female subjects both score 6 of 7 unambiguous (65th percentile). The concept of 12 houses survives. Individual variation is noise-level.

### 4.4 Phase 4: Timing Convergence

The random baseline (N=3,000) shows 64.5% of birth/target pairs produce at least one converging planet. The distribution: 55.8% have exactly one converging planet, 35.5% have zero, 8.7% have two. One birth chart in this sample produced a computation error and was excluded, leaving 2,999 valid records. No chart in this sample produced three or more converging planets; the Poisson expectation for the observed mean of 0.73 is 114 charts with 3+ in a sample this size, making zero an extreme deviation that warrants investigation of the underlying distribution.

All three systems agree on at least one planet in 4.5% of cases (134 of 2,999, 95% CI: 3.7-5.2%). This is the only rare outcome in the timing layer.

### 4.5 Phase 5: Node Convergence

The random baseline (N=5,000) shows the node sign survives tropical-to-sidereal shift in 22.2% of charts (95% CI: 21.0-23.4). For 77.8%, the Lahiri ayanamsa of approximately 24 degrees pushes the North Node across a sign boundary. The 180 degree opposition axis is preserved in 100% of charts.

### 4.6 Phase 6: Zodiac Comparison

The 9,738 synthetic charts confirm that dignity density is symmetric under ayanamsa shift. Neither tropical nor sidereal zodiac produces systematically more dignified placements. This result is analytically expected: the dignity table assigns domicile pairs to opposite signs, making it invariant under uniform sign shift. The computation confirms the analytical expectation.

### 4.7 Lunar Mansions

Nine shared anchor stars (33% overlap) between 27 nakshatras and 28 xiu. Bootstrap resampling from the actual star pools (10,000 iterations) yields a null expectation of 10.0 matches (37%, 95% CI: 22-52%). The observed 9 matches is consistent with random assignment (p = 0.76). The raw overlap count does not distinguish signal from chance.

What distinguishes the result is the brightness distribution. Three of the nine shared stars (Spica, Antares, Pleiades) are first magnitude or brighter, plausible as independent discoveries. Six (Sheratan at 2.6 mag, 35 Arietis at 4.6, Meissa at 3.5, Zubenelgenubi at 2.6, Markab at 2.5, Algenib at 2.8) are second magnitude or fainter. A formal null model incorporating stellar magnitude would test whether the concentration of faint stars in the shared set exceeds random expectation. The raw count is noise. The faint-star pattern is a qualitative signal that requires quantitative follow-up.

### 4.8 Null Result: Family Synastry

The 17-person family synastry, measured as aspect count at 8 degree orb, shows no relationship signal. For the three categories with more than one observation: couples averaged 35.8 aspects versus a random mean of 35.8 (SD 4.5, 51st percentile). Parent-child pairs averaged 37.9 versus 35.3 (SD 4.5, 31st percentile). Grandparent-grandchild pairs averaged 37.1 versus 37.0 (SD 4.7, 45th percentile). No category differed from random by more than 0.6 aspects.

Individual pairs scatter evenly across the random distribution. Blood relationship and marriage produce indistinguishable aspect counts. The aspect-density metric does not carry relationship information.

## 5. Discussion

### 5.1 What Survived

Three categories of structural features show transmission resilience under measurement.

The conjunction, opposition, and trine aspect angles are universal. These same three angles appear in the Babylonian Mul.Apin catalog, suggesting they may predate the Hellenistic synthesis [5].

The 12-house division holds across five competing methods at 79.4% agreement. The concept is robust. The disagreement is concentrated at cusp boundaries.

Node axis geometry is invariant. The 180 degree opposition is a mathematical property of the orbit. No cultural transmission can alter it.

Dignity convergence is partial: 46.7% mean agreement, a smooth bell curve, no family patterning. Individual variation spans the full range from 14% to 71%. The dignity table survived partially. It is not random. It is not intact.

### 5.2 What Diverged

Timing systems share modest overlap. Two thirds of charts show at least one converging planet (64.5%), but the high baseline is driven by generous element-to-planet mappings in the Chinese system. The tail is instructive: only 4.5% of charts achieve three-system agreement. When Vedic dasha says Mercury, Chinese pillars say Venus-and-Saturn, and Hellenistic profection says Sun, the systems are measuring different things.

The node sign is coordinate-dependent. The node axis concept survived. The sign-label on it did not.

Family synastry at the aspect-count level is null. The metric failed. Relationships may be encoded in specific angles, orbs, or planet-point combinations that a simple count cannot capture.

### 5.3 The Lunar Mansion Layer

The 33% overlap is consistent with a common origin modified by independent evolution. Three millennia separate the Shang oracle bones from the present. The quantitative overlap does not exceed random expectation. The qualitative signal is in the faint-star concentration: six of nine shared stars are too dim for plausible independent discovery. A proper null model incorporating stellar magnitude is the necessary next step.

The lunar mansions predate horoscopic astrology. They are a parallel system built on the sidereal month. The Hellenistic fusion grafted a planetary system onto a lunar foundation already ancient when the first horoscope was drawn.

A computational nakshatra-xiu comparison module could extend this analysis to per-chart measurement of sector boundary alignment and systematic offset.

### 5.4 What the Engine Doesn't Measure

The engine measures structural convergence between traditions. It does not measure whether astrology works.

A dignity rule (Sun domicile in Leo) is structural. The interpretation (confidence, leadership) is not. The engine can report that two systems agree on the rule 46.7% of the time. It cannot say whether the rule corresponds to anything outside itself.

Convergence may arise from the data, not the tradition. Conjunction and opposition are physically real alignments. Twelve equal divisions of a circle are mathematically natural. Some structural features may survive because they are encoded in the sky, not because they were faithfully transmitted. The engine cannot distinguish these explanations.

### 5.5 Future Work

The nakshatra-xiu comparison deserves a computational module. Sector boundary alignment, star precision, and systematic offset measurement would produce per-chart results rather than a static catalog comparison.

The timing convergence baseline could be sharpened by restricting the Ba Zi element-to-planet mapping. If Metal maps to Venus only and Water to Moon only, the overlap drops and the baseline may become discriminating.

Additional coordinate systems (draconic zodiac, heliocentric frame, ayanamsa values other than Lahiri) would extend the Phase 5 and 6 results.

Larger family datasets with confirmed biological relationships would strengthen or falsify the synastry null. Subgroup analysis by aspect type, orb tightness, and planet-point pairing requires more statistical power than 17 people can provide.

## 6. Conclusion

The transmission from Alexandria left a measurable imprint. Some of it survived. Most of it did not.

Three aspect angles are universal. Twelve houses hold. The node axis is geometry. The dignity table is partially preserved. The timing systems diverged completely.

Beneath all of it is the lunar mansion layer: nine shared stars at 33% overlap, predating horoscopic astrology by a millennium. The raw count is within random expectation. The faint-star concentration is the qualitative signal, awaiting quantitative test.

The engine is open source. The data, baselines, and paper are public. The measurements are falsifiable. The engine does not say whether any of this is true, only what survived.

## Appendix A: Stock Market Backtest

A Uranian transit backtest was conducted on SPY (S&P 500 ETF) from April 2021 to April 2026. This analysis uses the Hamburg School Uranian system, a 20th-century framework distinct from the Hellenistic-based measurements in the main paper [12]. The system employs hypothetical trans-Neptunian points (Vulkanus, Hades, Kronos, Admetos, Cupido, Poseidon), not classical planets. This appendix is included for completeness as the study's only external validation result.

At daily resolution (1,262 trading days), the continuous point score showed in-sample correlation with daily range (r = 0.34), but the signal inverted out of sample. Points predicting high volatility in 2021-2023 predicted low volatility in 2024-2026. Score autocorrelation of 0.93 at lag 1 indicated slow-moving regime measurement, not daily event detection.

At weekly resolution (262 weeks), a discrete net signal was computed from exact transit hits. Point classifications were a priori from tradition: high-volatility (Vulkanus, Hades, Kronos) and low-volatility (Admetos, Cupido, Poseidon). HV weeks (n = 60) showed 35% high-volatility rate versus 7% for LV weeks (n = 46). Quarterly direction was correct in 10 of 12 quarters (83%, p = 0.017 by permutation test). Pairwise comparison within quarters: 68.5% of HV-LV pairs had HV with higher range (p = 0.009). Year 2025 (April tariff shock during an LV week) weakened but did not eliminate the effect.

This is a single-domain result with a single methodology. Replication across indices, time periods, and point sets is required before the signal can be considered robust.

## References

[1] Needham, J. (1959). Science and Civilisation in China, Vol. 3: Mathematics and the Sciences of the Heavens and the Earth. Cambridge University Press.

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

[12] Witte, A. and Lefeldt, H. (1928). Rules for Planetary Pictures. Hamburg School.
