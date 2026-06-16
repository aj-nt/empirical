# Computational Comparison of Western, Vedic, and Arabic Astrological Frameworks

A.J. Flinton

June 2026

## Abstract

Western and Vedic astrology share a common origin in Hellenistic Alexandria (~150 BCE) but diverged over two millennia of largely separate development. This paper measures the degree of structural agreement between the two traditions using a single Go binary with a statically linked Swiss Ephemeris. Seventeen measurements compare Western, Vedic, and Arabic computational outputs — dignity rules, aspect angles, house division (extended to 8 systems including three Medieval methods), timing systems, node axis geometry, zodiac comparison, draconic transit behavior, two-way and three-way lunar mansion convergence, Arabic Parts, relocation house shift, fixed star conjunctions, secondary progressions, primary directions, relocation-dignity interaction, and electional date selection — each set against a Monte Carlo random baseline. All phases are reproducible from the current baseline tool. Results are reported with multi-seed ranges and ayanamsa sensitivity where applicable. The default ayanamsa is Lahiri, the standard adopted by the Indian government and the Swiss Ephemeris default [10].

Of the seventeen measurements, six are geometry checks — they confirm that angular distances are preserved under uniform coordinate shift, a property of subtraction rather than of astrology. Two measure the author's computational models rather than traditional practice. Two measure location effects. One is underpowered (N=26). Six remain as measurements of cross-tradition agreement.

Among those six, the dignity result is the most informative — but not for the reason the aggregate 47.0% figure suggests. Per-state analysis reveals that domicile/swakshetra, exaltation/uchcha, and fall/neecha assignments are identical across Western and Vedic tables (100% agreement). The 47.0% aggregate rate is an artifact of mapping a four-state Western system onto a three-state Vedic system: Western detriment has no Vedic equivalent, and the ayanamsa shift moves planets across sign boundaries, converting agreements into disagreements. The dignity tables did not diverge — three of four states survived transmission intact. Detriment is a Western-only innovation. Primary directions show partial cross-system survival at 51.0% mean (range 49.8-52.5%; varies 50.6-53.3% across ayanamsas). Node sign survival is 21.8% (range 21.2-22.5%; varies 18.7-26.2% across ayanamsas) and is primarily a function of ayanamsa choice. Timing convergence under the author's element-to-planet mappings is consistent with random chance (2.14% mean vs. 2.31% null expectation). House convergence across 8 systems (including Regiomontanus, Alcabitius, and Campanus) is 88.7% at a 75% agreement threshold.

Beneath the planetary layers, comparison of 27 Indian nakshatras, 28 Chinese xiu, and 28 Arabic manazil al-qamar reveals a three-tier structure. The two-way nakshatra-xiu overlap is 9 shared stars — consistent with independent brightness-weighted selection (p = 0.38), with a threshold-dependent faint-star excess (6 faint at ≥ 2.5, p = 0.0003). The nakshatra-manazil overlap is 13 shared stars — well above the brightness-weighted null expectation of 8.2, confirming the historical derivation of Arabic mansions from Indian nakshatras. The three-way overlap is 6 shared stars — consistent with independent selection (null expects 5.7, p = 0.54), with a significant faint-star excess (3 faint, null expects 0.24, p = 0.0013). However, three-way mansion convergence per chart is zero: shared anchor stars do not produce shared mansion placements when the three systems define their sector boundaries differently. The faint-star signal in the two-way comparison is threshold-dependent and sensitive to the shared-pool assumption in the null models.

The engine, data, baselines, and manuscript are open source. All numbers verified against live baseline output (June 2026).

## 1. Introduction

Around 150 BCE, at the intersection of Babylonian astronomy, Egyptian timekeeping, and Greek geometry, Hellenistic Alexandria produced the first system that cast planetary positions into a framework of signs, houses, and aspects. The earliest surviving horoscope is Babylonian (410 BCE), but the systematic integration of these elements into a coherent astrological framework is generally placed in Alexandria [4, 5]. The system spread west into Rome and east into India, where it merged with the pre-existing nakshatra system around the first or second century CE. Elements reached Tang dynasty China by the seventh century, where Ba Zi integrated the planetary framework with indigenous five-element cosmology [1, 6]. The Indian nakshatra system was transmitted to the Islamic world by the eighth century, becoming the Arabic manazil al-qamar [4].

The traditions diverged. Western astrology assigns dignity by domicile, exaltation, detriment, and fall. Vedic astrology uses swakshetra, uchcha, and neecha. Similar concepts with different assignments. Chinese Ba Zi maps planets to elements through a framework that does not map cleanly to Western or Vedic equivalents [1, 6]. Three systems. One shared origin. Roughly 2,000 years of largely separate development, with some secondary contact via the Silk Road.

This paper measures the degree of structural agreement between Western and Vedic computational outputs. It does not measure "transmission fidelity" — there is no surviving Alexandrian horoscope to establish what the original actually was. Agreement between the two traditions could reflect preservation of a shared original, or independent convergence on the same solution, or later contact reintroducing shared elements. The engine cannot distinguish these. What it can do is quantify how much the two traditions agree, and compare that agreement to what chance would produce.

The question is computational and the answer is a set of numbers. The engine does not say whether astrology works.

## 2. Related Work

No prior work measures cross-system astrological agreement computationally.

Pingree documented the historical transmission of Hellenistic astrology to India and the synthesis with the nakshatra system [4, 5, 6]. Needham catalogued the arrival of planetary astrology in China [1]. Subbarayappa and Sarma documented the nakshatra system's determinative stars and sector divisions [7]. All are historical, not computational.

The lunar mansion comparison — cross-referencing nakshatra, xiu, and manazil determinants and testing the overlap against null models — does not appear in the prior literature.

Methodologically, this paper sits within the broader tradition of computational measurement of cultural artifacts. Moretti's "distant reading" applied quantitative methods to literary corpora, treating cultural features as measurable quantities rather than purely interpretive objects [12]. Michel et al.'s culturomics work demonstrated that large-scale computational analysis of cultural corpora could reveal patterns invisible to close reading [13]. In textual criticism, phylogenetic methods adapted from evolutionary biology have been applied to manuscript traditions to reconstruct transmission histories from patterns of agreement and divergence [14, 15]. This paper applies a similar logic to astrological techniques: treat the techniques as cultural artifacts, measure agreement between descendant traditions, and compare observed agreement to null models. The difference is that astrological techniques lack the manuscript variants that stemmatics relies on — there is no chain of dated horoscopes showing when each dignity assignment changed. The measurement is therefore limited to quantifying agreement at the endpoints of transmission, without reconstructing the path between them.

## 3. Methods

### 3.1 What This Paper Measures

This paper measures agreement between Western and Vedic computational outputs. It does not measure "what survived transmission." Agreement may arise from:

1. **Shared inheritance**: both traditions preserved a feature of the original Hellenistic synthesis.
2. **Independent discovery**: both traditions arrived at the same solution because it is encoded in astronomy or mathematics (e.g., conjunction and opposition are physically real alignments; twelve equal divisions of a circle are geometrically natural).
3. **Later contact**: Silk Road exchange reintroduced shared elements after the Indian branch had already diverged.

The engine quantifies agreement. Distinguishing between these three mechanisms is beyond its scope. The paper reports convergence rates and, where possible, identifies which mechanisms are ruled out by the direction of the result.

### 3.2 Computational Phases

**Phase 1: Dignity agreement.** Seven classical planets classified under Western rules (domicile, exaltation, detriment, fall) and Vedic rules (swakshetra, uchcha, neecha) [4, 5]. Western astrology recognizes four dignity states, Vedic three. The original analysis assessed agreement as: both domicile/swakshetra, both exaltation/uchcha, or both peregrine (non-dignified). Detriment and neecha were grouped with peregrine as a single non-dignified state. This mapping choice produced the 47.0% aggregate rate. A per-state analysis (section 4.1) reveals that domicile/swakshetra, exaltation/uchcha, and fall/neecha assignments are identical across the two tables (100% agreement). The aggregate rate is an artifact of two factors: Western detriment has no Vedic equivalent (12 of 84 planet-sign pairs are Western-only), and the ayanamsa shift moves planets across sign boundaries, converting per-state agreements into aggregate disagreements. The structural bias from mapping a 4-state system onto a 3-state system is quantified in section 4.1.

**Phase 2: Aspect catalog.** Static comparison of seven major angles (0, 30, 60, 90, 120, 150, 180 degrees) across Western, Vedic, and Chinese traditions using the Brihat Parashara Hora Shastra, Ptolemy's Tetrabiblos, and the San Ming Tong Hui [2, 8, 9]. The three universal angles (conjunction, opposition, trine) are physically real alignments or geometrically natural divisions — a geometry check.

**Phase 3: House convergence.** Eight methods: whole sign, equal, Placidus, Porphyry, Koch, Regiomontanus, Alcabitius, Campanus. Applied to each of seven classical planets. A planet is unambiguous if at least 75% of systems (6 of 8) assign the same house number. This measures agreement between house systems within a single computational framework, not cross-tradition agreement. It does not involve ayanamsa.

**Phase 4: Timing convergence.** Vimshottari dasha (Vedic), Ba Zi luck pillars (Chinese), and Hellenistic annual profections compared for a given target date. Each system maps its active period to a set of planets. A planet appearing in two or more systems is a convergence. The element-to-planet mapping is the author's construction, not a standard Ba Zi mapping. Two variants are tested: a "generous" mapping (Metal→Saturn+Venus, Water→Moon+Mercury, Wood→Jupiter, Fire→Mars+Sun) and a "tight" mapping (Metal=Venus, Water=Mercury, Wood=Jupiter, Fire=Mars, Earth=Saturn). Neither is the standard Ba Zi framework, which maps elements to heavenly stems and earthly branches, not directly to planets. The null expectation for three-system agreement is computed analytically from the known probability distributions of each system (section 4.4). Note: Venus is absent from Vimshottari dasha (the 120-year cycle assigns periods to Sun, Moon, Mars, Rahu, Jupiter, Saturn, Mercury, and Ketu; Rahu and Ketu are nodes, not classical planets, and Venus is not included).

**Phase 5: Node sign survival.** Tropical and sidereal node positions compared using the Lahiri ayanamsa [10]. Sensitivity to other ayanamsas (Raman, Krishnamurti, Fagan-Bradley) is reported.

**Phase 6: Zodiac comparison.** Dignity density under tropical and sidereal coordinates. A sanity check: the dignity table assigns domicile pairs to opposite signs, making it invariant under uniform sign shift.

**Phase 7: Draconic transit behavior.** The draconic chart rotates the tropical zodiac so the North Node sits at 0° Aries. Natal draconic positions are zodiac-invariant (the ayanamsa cancels out in the rotation). Transiting positions are not. At a 3° orb with a ~24° ayanamsa, overlap is geometrically impossible — a geometry check.

**Phase 8: Lunar mansion convergence (two-way).** Each of the seven classical planets is assigned to its nakshatra (27 equal 13°20' sectors) and its Chinese xiu (28 unequal sectors). A planet converges if its nakshatra and xiu share the same determinative star — one of the 9 shared anchors identified in the static catalog comparison.

**Phase 8b: Lunar mansion convergence (three-way).** Same as Phase 8, extended to include the 28 Arabic manazil al-qamar. A planet converges if its nakshatra, xiu, and manazil all share the same determinative star — one of the 6 three-way shared anchors. Sector boundaries are defined by midpoints between consecutive determinative stars for each system independently.

**Phase 9: Arabic Parts.** Thirteen Hellenistic Arabic Parts computed from the ascendant and planetary positions. Sign placement survival under ayanamsa shift is frame-dependent. Part-to-planet aspects are angular distances and therefore preserved under uniform coordinate shift — a geometry check. Eight of the 13 Parts are documented in both Western and Vedic traditions under identical formulas; this formula comparison is a legitimate historical finding.

**Phase 10: Relocation house shift.** Seven classical planets placed into Placidus houses at two random geographic locations. Measures how many planets change houses under relocation.

**Phase 11: Fixed star conjunctions.** 116-star catalog computed via swe.Fixstar. Star-to-planet conjunctions detected in both tropical and sidereal frames at 2° orb. Angular distances between stars and planets are preserved under uniform coordinate shift — a geometry check.

**Phase 12: Progressed cross-system.** Secondary progressions (Naibod rate: one day per year). Progressed-to-natal aspects compared in tropical and sidereal frames. Angular distances are preserved under uniform coordinate shift — a geometry check.

**Phase 13: Primary directions.** Ptolemy's method: 1° of right ascension equals 1 year of life. ASC directed by oblique ascension (transcendental equation), MC directed by right ascension. Cross-system comparison at 3° orb. Unlike progressed cross-system, the OA-to-longitude conversion is nonlinear under ayanamsa shift, so this is not a geometry check — it is a genuine measurement.

**Phase 14: Relocation-dignity interaction.** Dignity scores computed at two random geographic locations and compared. Dignity is a function of ecliptic longitude, which is location-independent — a geometry check.

**Phase 15: Electional cross-system.** The author's electional scoring model (Moon house placement, Mercury sign condition, good/bad aspects) run in tropical and sidereal frames over 7-day windows. This measures the author's model, not electional astrology as practiced in either tradition. Moon house depends on ASC sign, which shifts with ayanamsa.

### 3.3 Random Baseline Generation

Monte Carlo simulation. Random dates 1900-2030, latitudes -60 to 60, longitudes -180 to 180, timezone offsets -12 to +12 hours in 0.5 hour increments, local times 00:00-23:59. The baseline tool uses 25 planets: the 7 classical planets plus Uranus, Neptune, Pluto, Node, Chiron, Lilith, Ceres, Pallas, Juno, Vesta, and 8 Uranian planets.

Sample sizes: Phase 1: 10,000, Phase 3: 5,000, Phase 4: 5,000, Phase 5: 5,000, Phase 6: 10,000, all others: 1,000. Phase 4 randomizes target dates from birth to birth plus 90 years. Phase 7 randomizes transit dates within ±5 years of birth. Phase 12 randomizes target dates within 0-90 years of birth. Phase 13 randomizes ages from 0-90 years. Phase 15 randomizes 7-day windows within 0-80 years of birth.

All results reported with multi-seed ranges (seeds 42, 123, 456, 789) and ayanamsa sensitivity (Lahiri, Raman, Krishnamurti, Fagan-Bradley) where applicable. The Lahiri ayanamsa is the default; it is the standard adopted by the Indian government and the Swiss Ephemeris default [10]. The baseline tool explicitly sets `SetSidMode(SIDM_LAHIRI)` before all computations.

### 3.4 Code Validation

The Go engine's internal logic packages are tested with 281 test functions, all passing (`go test ./...` exit 0). The dignity, directions, astrocartography, progressed, and interpretation modules each have dedicated test suites verifying correctness against hand-computed reference values. The Swiss Ephemeris bindings are integration-tested against known planetary positions.

No independent implementation exists to cross-validate these results. The numbers should be treated as single-implementation measurements until reproduced by an independent codebase.

### 3.5 Family Dataset

Seventeen people across three generations scored for individual chart analysis and synastry. Three couples, 14 parent-child pairs, 24 grandparent-grandchild pairs. Age-matched random baseline of 5,000 pairs per category. N=26 pairs total. The sample is too small for population inference. Results reported for illustration only.

### 3.6 Lunar Mansion Null Models

Three null models test whether the observed shared stars between mansion systems exceed random expectation. All use 10,000 bootstrap iterations. Star pools are from documented single-star anchors per Wikipedia. Magnitudes are from the Swiss Ephemeris catalog.

**Two-way (nakshatra-xiu):** 27 nakshatra determinants, 28 xiu determinants. Combined pool: 46 unique stars.

**Two-way (nakshatra-manazil):** 27 nakshatra determinants, 28 manazil determinants. Combined pool: 48 unique stars.

**Three-way (nakshatra-xiu-manazil):** All three systems. Combined pool: 57 unique stars.

**Null 1 (Uniform):** Each system selects stars uniformly from the combined pool, without replacement. Bootstrap mean for two-way: 16.4. Observed 9 is well below expectation (p = 0.0000). Both cultures were selective, not random.

**Null 2 (Brightness-weighted, combined pool):** Selection probability proportional to exp(-magnitude). For nakshatra-xiu: bootstrap mean total 8.1 (95% CI: 8.0-8.1). Observed 9 is consistent (p = 0.38). For faint stars (magnitude ≥ 2.5): null mean 1.2 (95% CI: 1.2-1.2). Observed 6 yields p = 0.0003. For nakshatra-manazil: null mean 8.2. Observed 13 yields p = 0.0000 — well above expectation, confirming historical derivation. For three-way: null mean 5.7. Observed 6 is consistent (p = 0.54). Three-way faint: null mean 0.24, observed 3 yields p = 0.0013.

**Null 3 (Brightness + ecliptic proximity weighted):** Selection probability proportional to exp(-magnitude) * exp(-|latitude|/10 deg). For nakshatra-xiu: bootstrap mean total 7.6 (95% CI: 7.6-7.7). Observed 9, p = 0.28. Faint stars: null mean 1.9 (95% CI: 1.9-1.9). Observed 6 yields p = 0.0031.

**Threshold sensitivity:** The faint-star count depends on the magnitude cutoff. At ≥ 2.5: 6 faint stars. At ≥ 2.75: 5 faint stars (Sheratan at 2.65 and Zubenelgenubi at 2.75 flip to bright). At ≥ 3.0: 3 faint stars (Algenib at 2.84 also flips). At ≥ 3.5: 3 faint stars. At ≥ 4.0: 2 faint stars. The p = 0.0003 result is specific to the 2.5 threshold. The paper uses 2.5 as a round number near the median of the magnitude distribution; no optimization was performed, but the sensitivity is acknowledged.

**Geographic visibility caveat:** The null models assume all cultures selected from the same combined pool. Nakshatra determinants were selected from stars visible from Indian latitudes (~8-35°N). Xiu determinants were selected from stars visible from Chinese latitudes (~18-54°N). Manazil determinants were selected from stars visible from Middle Eastern latitudes (~15-40°N). These are overlapping but not identical pools. The shared-pool assumption may inflate the expected overlap.

### 3.7 Limitations

The engine measures agreement between Western and Vedic traditions for Phases 1, 5, 6, 9, 10, 11, 12, 13, 14, and 15. Phase 3 measures within-tradition house system agreement, not cross-tradition. Phase 4 compares all three traditions using the author's element-to-planet mappings. Phase 2 is a static catalog. Chinese dignity and houses cannot be computed per chart.

Six phases are geometry checks (Phases 2, 7, 9-aspects, 11, 12, 14): they confirm that angular distances are preserved under uniform coordinate shift. This is a property of subtraction (`(X - a) - (Y - a) = X - Y`), not a property of astrological techniques. They are included for completeness and as code validation, not as findings about transmission.

The "invariant" techniques (Phases 9 aspects, 11, 12) are invariant only under the assumption that both traditions use identical computational pipelines — same star catalog, same Part formulas, same progression rate. In practice, Vedic astrologers may use different ayanamsas for different purposes, different Part formulas for some Lots, and different progression rates. The paper measures the computational pipeline, not the traditions as actually practiced.

All cross-system numbers depend on the choice of ayanamsa. Different ayanamsas (Raman ~22.5°, Krishnamurti ~23.8°, Fagan-Bradley ~24.5°) produce different results for every frame-dependent phase. The paper reports Lahiri as the default and provides sensitivity ranges.

Phase 1 maps a four-state Western system onto a three-state Vedic system. The original mapping grouped Western detriment and Vedic neecha with peregrine as a single non-dignified state. This mapping choice produced the 47.0% aggregate rate. Per-state analysis (section 4.1) shows that domicile/swakshetra, exaltation/uchcha, and fall/neecha assignments are identical across the two tables. The aggregate rate reflects the mapping asymmetry (Western detriment has no Vedic equivalent) and ayanamsa-driven sign-boundary crossings, not table divergence.

Phase 4 uses element-to-planet mappings constructed by the author. Neither the "generous" nor "tight" mapping is the standard Ba Zi framework. The results describe the author's mappings, not the traditions.

Phase 15 uses an electional scoring model constructed by the author. The results describe the author's model, not electional astrology as practiced.

The lunar mansion null models assume a shared star pool and do not control for geographic visibility constraints. The faint-star result is threshold-dependent.

No independent implementation exists. All results are single-implementation until reproduced.

### 3.8 Code and Data Availability

Open source under MIT license at github.com/aj-nt/empirical. Go source: 281 test functions, all passing. Swiss Ephemeris via github.com/aloistr/swisseph (GPL). This paper is licensed under CC-BY 4.0. All baseline results verified against live output (June 2026).

## 4. Results

### 4.1 Phase 1: Dignity Agreement

**Per-state analysis.** The Western and Vedic dignity tables are compared state-by-state for all 84 planet-sign pairs (7 planets × 12 signs):

| Western State | Vedic Equivalent | Agreement | Pairs |
|---|---|---|---|
| Domicile | Swakshetra | **100%** (12/12) | Identical assignments |
| Exaltation | Uchcha | **100%** (6/6) | Identical assignments |
| Fall | Neecha | **100%** (6/6) | Identical assignments |
| Detriment | — (no Vedic equivalent) | **0%** | 1 maps to neecha, 11 to peregrine |
| Peregrine | Peregrine | **100%** (48/48) | Identical assignments |

Domicile, exaltation, and fall assignments are identical across the two traditions. The dignity tables did not diverge — three of four states survived transmission intact. Detriment is a Western-only innovation: Vedic astrology never developed a detriment concept. Of the 12 Western detriment pairs, 11 map to Vedic peregrine and 1 (Sun in Aquarius) maps to neecha.

**Aggregate rate.** The original analysis assessed agreement as: both domicile/swakshetra, both exaltation/uchcha, or both peregrine. Detriment and neecha were grouped with peregrine as a single non-dignified state. Under this mapping, the static table agreement is 72/84 = 85.7% (the 12 detriment pairs are counted as disagreements because Vedic has no equivalent category).

Random baseline (N=10,000, seed 42, Lahiri): mean 3.29 of 7 planets agreeing (47.0%). Distribution:

```
0:  1.6%  ▏
1:  8.2%  ▍
2: 20.7%  ▋
3: 29.5%  ▊
4: 23.7%  ▋
5: 12.6%  ▍
6:  3.4%  ▏
7:  0.3%  ▏
```

Multi-seed range (N=10,000): 46.7-47.0% (seeds 42, 123, 456, 789).

Ayanamsa sensitivity (N=10,000, seed 42): Lahiri 47.0%, Raman 49.5%, Krishnamurti 47.2%, Fagan-Bradley 45.5%. The 4.0pp range across ayanamsas is larger than the 0.3pp range across seeds.

**Why 47.0% when the tables are 85.7% identical?** The gap between the static table agreement (85.7%) and the observed per-chart rate (47.0%) is driven by the ayanamsa shift. The Lahiri ayanamsa (~24°) moves planets across sign boundaries in ~53% of cases. When a planet crosses a sign boundary, its Western and Vedic sign assignments differ, and the dignity comparison becomes a cross-sign comparison. A planet that would be domicile in both systems if it stayed in the same sign becomes domicile in one and peregrine in the other when the ayanamsa places it in different signs. The 47.0% rate measures the combined effect of the mapping asymmetry (detriment has no Vedic equivalent) and the ayanamsa-driven sign-boundary crossings. It does not measure table divergence — the tables are identical for the three states both traditions share.

**Null expectation.** Under random sign assignment, the probability that a given planet's Western and Vedic dignity classifications agree (both domicile/swakshetra, both exaltation/uchcha, or both peregrine) can be computed analytically from the dignity tables. The mean per-planet probability is 78.6%. The expected mean converging planets per chart is 5.50 out of 7 (78.6%). The observed 3.29 (47.0%) is 31.6 percentage points below the null expectation.

The null expectation of 78.6% reflects the structure of the tables: domicile pairs are opposite signs, exaltation signs are specific, and most signs are peregrine for most planets. Two independently constructed dignity tables would agree most of the time by chance. The observed rate is below chance not because the tables disagree but because the ayanamsa shift breaks the sign correspondence that the tables require. The tables agree on what dignity a planet has in a given sign. The ayanamsa determines whether the planet is in the same sign in both systems.

The 17-person family sample scored from 14.3% to 71.4%. Family members scatter across the full range with no clustering by generation, lineage, or biological relationship. The reference male subject scores 4 of 7 (57.1%), above the population mean but within the normal range.

### 4.2 Phase 2: Aspect Catalog

Three angles are universal across all three traditions: conjunction (0°), opposition (180°), trine (120°). Square (90°) is explicit in Western and Vedic, implicit in Chinese through the punishment relationship. Sextile (60°) appears in Western and Vedic only. Semi-sextile (30°) and quincunx (150°) are Western/Vedic with partial Chinese equivalents.

These same three universal angles appear in the Babylonian Mul.Apin catalog, suggesting they predate the Hellenistic synthesis [5]. Conjunction and opposition are physically real alignments independent of culture. The trine divides the circle in three, a geometrically natural division. Agreement on these angles reflects astronomy and geometry, not cultural transmission. This phase is a geometry check.

### 4.3 Phase 3: House Convergence

Random baseline (N=5,000, seed 42, 8 systems, 75% threshold): mean 6.21 of 7 unambiguous (88.7%). Distribution: 0: 0.0%, 1: 0.3%, 2: 1.2%, 3: 2.5%, 4: 4.5%, 5: 10.6%, 6: 26.2%, 7: 54.6%.

With the original 5 systems at 80% threshold: 83.0% (range 82.5-83.0% across seeds). Adding Regiomontanus, Alcabitius, and Campanus raises convergence to 88.7% — the three Medieval quadrant systems largely agree with Placidus, Porphyry, and Koch on house placement.

This measures agreement between house systems within a single computational framework. It does not measure cross-tradition agreement and does not involve ayanamsa. The high convergence rate reflects the fact that all eight systems divide the same 360° circle into twelve sectors; planets far from cusp boundaries will land in the same house under any reasonable system.

### 4.4 Phase 4: Timing Convergence

Random baseline (N=5,000, seed 42, Lahiri): 33.1% of birth/target pairs produce at least one converging planet under the generous mapping. Distribution: 66.9% zero, 28.6% one, 4.4% two. None at three or more.

Full three-system agreement (generous mapping): 2.18% (seed 42), multi-seed range 2.04-2.22% (mean 2.14%). Tight mapping: 1.76% (seed 42), multi-seed range 1.56-1.76% (mean 1.66%). Timing convergence is effectively ayanamsa-invariant (all four ayanamsas produce 2.16-2.18% generous, 1.76-1.78% tight at seed 42).

**Null expectation.** The three systems do not select planets uniformly. Vimshottari dasha assigns periods proportional to planetary periods: Sun 8.0%, Moon 13.3%, Mars 9.3%, Jupiter 21.3%, Saturn 25.3%, Mercury 22.7%. Venus is absent from Vimshottari dasha. Ba Zi luck pillars (generous mapping, renormalized to exclude the 20% Earth-element cases that map to no planet): Sun 12.5%, Moon 12.5%, Mercury 12.5%, Venus 12.5%, Mars 12.5%, Jupiter 25.0%, Saturn 12.5%. Hellenistic annual profections: Sun 8.3%, Moon 8.3%, Mercury 16.7%, Venus 16.7%, Mars 16.7%, Jupiter 16.7%, Saturn 16.7%. The probability that all three systems activate the same planet by chance is the sum over planets of the product of these three probabilities: 2.31% for the generous mapping, 2.62% for the tight mapping.

The observed mean of 2.14% (generous) is within 0.17pp of the 2.31% null expectation. The observed mean of 1.66% (tight) is 0.96pp below the 2.62% null expectation. Neither is strong evidence for or against transmission. The generous mapping result is consistent with random chance. The tight mapping result is directionally below chance but the effect is small relative to sampling variability.

These results describe the author's element-to-planet mappings, not the Ba Zi tradition. The standard Ba Zi framework maps elements to heavenly stems and earthly branches, not directly to planets. Different mapping choices would produce different numbers.

### 4.5 Phase 5: Node Sign Survival

Random baseline (N=5,000, seed 42, Lahiri): 21.8% sign survival (1,089 of 5,000). Multi-seed range: 21.2-22.5%.

Ayanamsa sensitivity (N=5,000, seed 42): Lahiri 21.8%, Raman 26.2%, Krishnamurti 22.1%, Fagan-Bradley 18.7%. The 7.5pp range across ayanamsas is the largest of any phase. The node sign survival rate is primarily a function of the ayanamsa value, not of the technique.

The 180° opposition axis is preserved in every case regardless of ayanamsa. The node axis is orbital mechanics, not a cultural artifact.

### 4.6 Phase 6: Zodiac Comparison

Random baseline (N=10,000, seed 42, Lahiri): tropical produces more dignified placements in 37.2% of charts, sidereal in 38.0%, and 24.8% are ties. Multi-seed range: trop 37.2-38.0%, sid 38.0-38.9%, tie 23.6-24.8%.

Neither zodiac produces systematically more dignified placements. The dignity table assigns domicile pairs to opposite signs, making it invariant under uniform sign shift. This is an analytical expectation confirmed computationally.

### 4.7 Phase 8: Lunar Mansion Convergence (Two-Way)

Random baseline (N=1,000, seed 42, Lahiri): mean 0.15 converging planets per chart (2.2% of 7). Distribution: 86.8% zero, 11.6% one, 1.6% two. No chart at three or more. Multi-seed range: 0.14-0.17 (2.0-2.4%).

The convergence rate is the lowest of any phase. Nine shared anchor stars out of 55 mansion pairs means 16.4% of mansion pairs share a star, but planets must actually land in those pairs. The baseline confirms the rate is barely above zero.

### 4.8 Phase 8b: Lunar Mansion Convergence (Three-Way)

Random baseline (N=1,000, seed 42, Lahiri): mean 0.00 converging planets per chart (0.0%). All 1,000 charts have zero three-way convergences.

This is a structural finding, not a bug. The three systems share 6 anchor stars (Sheratan, Meissa, Spica, Zubenelgenubi, Antares, Markab) but define their sector boundaries differently — midpoints between different neighboring stars. A planet in Sheratan's nakshatra sector can be outside Sheratan's xiu sector because the xiu boundaries are different. Shared stars do not produce shared mansion placements when the boundary definitions differ. The three-way convergence rate is zero because the probability of a planet landing in the intersection of three independently-defined sectors around the same star is negligible.

### 4.9 Phase 9: Arabic Parts

Random baseline (N=1,000, seed 42, Lahiri): mean 2.85 of 13 Parts retain the same sign under ayanamsa shift (21.9%). Multi-seed range: 2.75-2.89 (21.1-22.2%).

Part-to-planet aspects are angular distances and therefore preserved under uniform coordinate shift — a geometry check. Eight of the 13 Parts (Fortune, Spirit, Victory, Father, Mother, Children, Marriage, Death) are documented in both Western and Vedic traditions under identical formulas. This formula comparison is a legitimate historical finding: the formulas themselves survived transmission intact. The sign placements they produce are coordinate-dependent. The aspects they form to planets are invariant under coordinate shift, like all angular distances.

### 4.10 Phase 10: Relocation House Shift

Random baseline (N=1,000, seed 42): mean 6.36 of 7 planets shift houses (90.9%). Distribution: 82.7% of charts shift all seven planets. Multi-seed range: 6.35-6.42 (90.7-91.7%).

House placement is overwhelmingly location-dependent. The ascendant shifts with latitude and longitude, dragging all twelve cusps with it. Planet longitudes are invariant. The result is that most planets change houses under relocation.

### 4.11 Phase 13: Primary Directions

Random baseline (N=1,000, seed 42, 3° orb): mean 3.24 directed ASC aspects, 3.38 directed MC aspects per chart. 96.9% of charts have at least one ASC hit, 98.3% at least one MC hit. Multi-seed ranges: ASC 3.24-3.36, MC 3.25-3.38.

Cross-system survival (N=1,000, seed 42, Lahiri, 3° orb): 51.6%. Multi-seed range: 49.8-52.5% (mean 51.0%).

Ayanamsa sensitivity (seed 42): Lahiri 51.6%, Raman 53.3%, Krishnamurti 51.7%, Fagan-Bradley 50.6%. Range: 2.7pp.

Unlike progressed cross-system (Phase 12), primary directions are not a geometry check. The OA-to-longitude conversion is transcendental (binary search to invert the OA formula) and does not preserve the uniform shift. The MC direction is RA-based but the RA-to-longitude conversion depends on obliquity, and RA itself shifts nonlinearly with longitude. The ~51% survival rate is a genuine measurement: the technique partially survives the zodiac shift, but the transcendental geometry prevents complete invariance.

### 4.12 Phase 15: Electional Cross-System

Random baseline (N=1,000, seed 42, Lahiri, 3° orb, 7-day windows): tropical and sidereal rankings agree on the best day 51.7% of the time. Multi-seed range: 47.2-51.7% (mean 48.8%).

Ayanamsa sensitivity (seed 42): Lahiri 51.7%, Raman 42.5%, Krishnamurti 51.4%, Fagan-Bradley 57.1%. Range: 14.6pp — the largest ayanamsa sensitivity of any phase.

This measures the author's electional scoring model, not electional astrology as practiced in either tradition. The model scores dates on Moon house placement, Mercury sign condition, and good/bad aspects. Moon house depends on ASC sign, which shifts with ayanamsa, making the result frame-dependent. Different scoring models would produce different numbers.

### 4.13 Geometry Checks: Phases 7, 11, 12, 14

Four phases confirm that angular distances are preserved under uniform coordinate shift. They are code validation, not findings about transmission.

| Phase | Technique | Result |
|-------|-----------|--------|
| 7 | Draconic transit behavior | 0% cross-system survivors (geometrically impossible at 3° orb with ~24° ayanamsa) |
| 11 | Fixed star conjunctions | 100% survival (mean 33.37 conjunctions per chart at 2° orb) |
| 12 | Progressed cross-system | 100% survival (mean 96.0 aspects per chart at 3° orb) |
| 14 | Relocation-dignity interaction | 100% identical dignity scores across random locations |

In each case, both the target and reference positions shift by the same ayanamsa, so angular distances are unchanged. Phase 7 is the limit case: the ayanamsa (~24°) exceeds the orb (3°) by a factor of eight, making overlap impossible. The zero result confirms the orb filter works correctly. Phases 11, 12, and 14 confirm that the code correctly computes star positions, secondary progressions, and the separation of location-dependent from location-independent computation.

### 4.14 Lunar Mansions: Star-Level Comparison

**Two-way (nakshatra-xiu):** 9 shared anchor stars. The nine: Sheratan (Beta Arietis, 2.65), 35 Arietis (4.60), Spica (Alpha Virginis, 0.97), Antares (Alpha Scorpii, 0.91), Markab (Alpha Pegasi, 2.48), Algenib (Gamma Pegasi, 2.84), Meissa (Lambda Orionis, 3.66), Zubenelgenubi (Alpha Librae, 2.75), Delta Hydrae (4.14). Magnitudes from the Swiss Ephemeris catalog.

Krittika and Mao both point to the Pleiades cluster but to different specific stars (Alcyone vs. Electra). Under strict star-by-star matching they are not shared.

Three are first magnitude or brighter (Spica, Antares, Markab). The remaining six are magnitude ≥ 2.5. Under brightness-weighted null selection (Null 2), the expected faint-star overlap is 1.2. Observed 6 yields p = 0.0003.

**Two-way (nakshatra-manazil):** 13 shared stars. The additional four beyond the nakshatra-xiu set: Aldebaran (Alpha Tauri, 0.86), Alcyone (Eta Tauri, 2.87), Castor (Alpha Geminorum, 1.58), Regulus (Alpha Leonis, 1.40). Under brightness-weighted null selection, the expected overlap is 8.2. Observed 13 yields p = 0.0000 — well above expectation. This confirms the historical derivation of Arabic manazil from Indian nakshatras: the overlap is substantially higher than independent selection would produce.

**Three-way (nakshatra-xiu-manazil):** 6 shared stars: Sheratan, Meissa, Spica, Zubenelgenubi, Antares, Markab. Under brightness-weighted null selection, the expected three-way overlap is 5.7. Observed 6 is consistent (p = 0.54). Three-way faint stars: 3 (Sheratan 2.65, Meissa 3.66, Zubenelgenubi 2.75). Null expects 0.24. Observed 3 yields p = 0.0013 — the faint-star excess persists in the three-way comparison.

**Threshold sensitivity.** The faint-star count depends on the magnitude cutoff:

| Threshold | Faint stars (nak-xiu) | Stars that flip |
|-----------|----------------------|-----------------|
| ≥ 2.5 | 6 | — |
| ≥ 2.75 | 5 | Sheratan (2.65), Zubenelgenubi (2.75) |
| ≥ 3.0 | 3 | + Algenib (2.84) |
| ≥ 3.5 | 3 | — |
| ≥ 4.0 | 2 | + Meissa (3.66) |

The p = 0.0003 result is specific to the 2.5 threshold. The threshold was chosen as a round number near the median of the magnitude distribution; no optimization was performed. The sensitivity is reported in full.

Under Null 3 (brightness + ecliptic proximity), the expected faint overlap rises to 1.9 and observed 6 yields p = 0.0031. The signal weakens but survives at the 2.5 threshold.

An ecliptic position confound test (100,000 permutations): shared faint stars have mean |lat| 14.8° vs. 22.0° for non-shared faint stars. The difference does not reach significance (p = 0.19).

**Geographic visibility caveat.** The null models assume all cultures selected from the same combined pool. Nakshatra determinants were selected from stars visible from Indian latitudes (~8-35°N). Xiu determinants were selected from stars visible from Chinese latitudes (~18-54°N). Manazil determinants were selected from stars visible from Middle Eastern latitudes (~15-40°N). These are overlapping but not identical pools. The shared-pool assumption may inflate the expected overlap.

The total nakshatra-xiu overlap of 9 is consistent with independent brightness-weighted selection (p = 0.38). The nakshatra-manazil overlap of 13 is well above expectation (p = 0.0000), confirming historical derivation. The three-way overlap of 6 is consistent with independent selection (p = 0.54). The faint-star composition is suggestive but threshold-dependent and sensitive to the shared-pool assumption.

### 4.15 Family Synastry

Five metrics at 3° orb, all null. Family sample of 26 pairs versus random baseline of 5,000 pairs (seed 42):

| Metric | Family Mean | Random Mean | Random SD | z-score | p (approx) |
|--------|------------|-------------|-----------|---------|------------|
| Total aspects | 8.73 | 8.55 | 2.72 | 0.07 | 0.95 |
| Conjunctions only | 1.00 | 1.07 | 1.11 | -0.07 | 0.95 |
| Saturn contacts | 1.92 | 1.96 | 1.32 | -0.03 | 0.98 |
| Node contacts | 2.04 | 1.98 | 1.31 | 0.04 | 0.97 |
| Sun-Moon contacts | 0.42 | 0.26 | 0.49 | 0.33 | 0.74 |

No metric deviates significantly from random expectation. N=26 pairs is underpowered for small effects. These results do not rule out synastry; they only fail to find a signal in aspect density with this sample size.

## 5. Discussion

### 5.1 What Was Actually Measured

The seventeen phases fall into five categories:

**Geometry checks (6 phases):** Phases 2, 7, 9 (aspects), 11, 12, 14. These confirm that angular distances are preserved under uniform coordinate shift. This is a property of subtraction, not of astrology. They are included for completeness and as code validation. They do not constitute findings about cultural transmission.

**Cross-tradition measurements (6 phases):** Phases 1, 5, 8, 8b, 9 (signs), 13. These measure agreement between Western and Vedic computational outputs, or between mansion systems. The numbers are: dignity 47.0% (range 46.7-47.0% across seeds, 45.5-49.5% across ayanamsas), node sign 21.8% (range 21.2-22.5% across seeds, 18.7-26.2% across ayanamsas), two-way mansion convergence 2.2% (range 2.0-2.4%), three-way mansion convergence 0.0%, Arabic Part signs 21.9% (range 21.1-22.2%), primary directions 51.0% mean cross-system survival (range 49.8-52.5% across seeds, 50.6-53.3% across ayanamsas).

**Author's models (2 phases):** Phases 4, 15. These measure the author's element-to-planet mappings and electional scoring model, not traditional practice. Timing: 2.14% mean full agreement (consistent with 2.31% null). Electional: 48.8% mean best-day agreement (range 47.2-51.7% across seeds, 42.5-57.1% across ayanamsas).

**Within-tradition measurement (1 phase):** Phase 3 (house convergence, 88.7% with 8 systems at 75% threshold; 83.0% with original 5 systems at 80% threshold). Measures agreement between house systems within a single computational framework. Does not involve ayanamsa or cross-tradition comparison.

**Location effects (2 phases):** Phase 10 (90.9% house shift) and Phase 14 (0% dignity shift).

### 5.2 Ayanamsa Dependence

Every frame-dependent number in this paper is a function of the chosen ayanamsa. The Lahiri ayanamsa (~24°) is the default because it is the Indian government standard and the Swiss Ephemeris default. But different ayanamsas exist and are used in practice.

| Phase | Lahiri | Raman | Krishnamurti | Fagan-Bradley | Range |
|-------|--------|-------|-------------|---------------|-------|
| Dignity | 47.0% | 49.5% | 47.2% | 45.5% | 4.0pp |
| Node sign | 21.8% | 26.2% | 22.1% | 18.7% | 7.5pp |
| Primary directions | 51.6% | 53.3% | 51.7% | 50.6% | 2.7pp |
| Electional | 51.7% | 42.5% | 51.4% | 57.1% | 14.6pp |

These ranges are not error bars — they are the actual values different ayanamsas produce. A paper reporting a single ayanamsa's numbers as "the" convergence rate is reporting a conditional measurement. The condition is the choice of ayanamsa.

### 5.3 The Dignity Result

The dignity finding is the most informative in the paper — but the original interpretation was wrong.

The original analysis reported 47.0% aggregate agreement and concluded that the dignity tables had diverged, ruling out simple preservation of a shared original. Per-state analysis reveals the opposite: domicile/swakshetra, exaltation/uchcha, and fall/neecha assignments are identical across the two traditions (100% agreement for all three states). The dignity tables did not diverge — three of four states survived transmission intact.

The 47.0% aggregate rate is an artifact of two factors, neither of which reflects table divergence:

1. **Mapping asymmetry.** Western astrology has four dignity states; Vedic has three. Western detriment has no Vedic equivalent. The original analysis grouped detriment with peregrine as "non-dignified," but Vedic peregrine is not a detriment equivalent — it is the absence of any special dignity. The 12 Western detriment pairs (14.3% of all planet-sign pairs) are counted as disagreements because Vedic has no matching category. This is a category mismatch, not a disagreement about what dignity a planet has in a given sign.

2. **Ayanamsa-driven sign-boundary crossings.** The Lahiri ayanamsa (~24°) moves planets across sign boundaries in approximately half of all cases. When a planet crosses a sign boundary, its Western and Vedic sign assignments differ, and the dignity comparison becomes a cross-sign comparison. A planet that would be domicile in both systems if it stayed in the same sign becomes domicile in one and peregrine in the other when the ayanamsa places it in different signs. The tables agree on what dignity a planet has in a given sign. The ayanamsa determines whether the planet is in the same sign in both systems.

The static table agreement is 85.7% (72/84 pairs). The observed per-chart rate of 47.0% reflects the combined effect of the mapping asymmetry and the ayanamsa shift. Neither factor indicates that the traditions modified the dignity assignments. The assignments are identical for the three states both traditions share.

This finding has implications beyond the paper. A synthesis system built from the paper's findings should use domicile, exaltation, and fall — the three states that survived transmission intact — and should exclude detriment, which is a Western-only innovation with zero cross-traditional signal. The Koiné astrology system (github.com/aj-nt/koine) implements this three-state model.

### 5.4 The Timing Result

The observed 2.14% mean full three-system agreement is within 0.17pp of the 2.31% null expectation — consistent with random chance under the author's element-to-planet mappings. This does not mean the timing systems are random. It means the author's mappings, combined with the known probability distributions of Vimshottari dasha and annual profections, produce a three-system agreement rate indistinguishable from chance. Different mappings would produce different numbers. The standard Ba Zi framework does not map elements directly to planets, so these results describe the author's mapping choices, not the tradition.

### 5.5 The Lunar Mansion Results

The lunar mansion analysis is the most methodologically novel part of the paper and the most caveated.

**Star-level overlap.** The two-way nakshatra-xiu total overlap of 9 is consistent with independent brightness-weighted selection (p = 0.38). The nakshatra-manazil overlap of 13 is well above expectation (p = 0.0000), confirming the historical derivation of Arabic mansions from Indian nakshatras. The three-way overlap of 6 is consistent with independent selection (p = 0.54).

**Faint-star composition.** At the ≥ 2.5 threshold, 6 of 9 nakshatra-xiu shared stars are faint, where the null expects 1.2 (p = 0.0003). The three-way comparison shows 3 faint stars where the null expects 0.24 (p = 0.0013). The faint-star excess persists across both comparisons. But the result is threshold-dependent: at ≥ 2.75 the count drops to 5, at ≥ 3.0 to 3. The threshold was chosen as a round number, not optimized, but the sensitivity is real.

**Chart-level convergence.** The two-way mansion convergence rate is 2.2% — barely above zero. The three-way convergence rate is 0.0%. Shared anchor stars do not produce shared mansion placements when the three systems define their sector boundaries differently. The star-level overlap is a catalog comparison; the chart-level convergence is a geometric consequence of boundary definitions.

**Caveats.** The shared-pool assumption is significant. If the three cultures selected from different visibility-constrained pools, the expected overlap under independence would be lower. The faint-star result is threshold-dependent. The most conservative reading: the total overlap is unremarkable, the faint-star composition is suggestive but threshold-dependent, and the shared-pool assumption may inflate the null. The least conservative reading: the faint-star excess survives the most obvious confounds (ecliptic proximity, combined null model, three-way comparison) and is unlikely under independent selection. The truth is somewhere between, and the paper provides the data for the reader to decide.

### 5.6 Interpreting Agreement

Not all measurements bear equally on the question of transmission. The phases can be sorted by what they tell us:

**Results that constrain interpretation:**

The dignity result, correctly analyzed, is the paper's strongest finding — but in the opposite direction from the original interpretation. Per-state analysis shows that domicile/swakshetra, exaltation/uchcha, and fall/neecha assignments are identical across the two traditions (100% agreement). Three of four dignity states survived transmission intact. The 47.0% aggregate rate is an artifact of mapping asymmetry (Western detriment has no Vedic equivalent) and ayanamsa-driven sign-boundary crossings. The dignity tables did not diverge. This is the paper's most important correction.

The nakshatra-manazil overlap (13 vs. 8.2 null, p = 0.0000) confirms the historical derivation of Arabic mansions from Indian nakshatras. This is not a discovery — historians already knew this — but it validates the null model framework: when transmission is known to have occurred, the engine detects it.

**Results that are genuinely ambiguous:**

Primary directions (~51% cross-system survival) occupy a middle ground. The transcendental OA-to-longitude conversion prevents the uniform shift from canceling out, so the result is not a geometry check. But the ~51% rate is not clearly above or below any meaningful baseline. Partial transmission is a plausible interpretation; independent convergence on similar mathematical methods (Ptolemy's algorithm is the natural solution to the problem of directing the ascendant) is equally plausible. The measurement cannot distinguish them.

**Results that are primarily functions of the coordinate system:**

Node sign survival (21.8%) varies by 7.5pp across ayanamsas — more than any other phase except electional. The node axis is orbital mechanics. The sign the node occupies is a function of where you draw the sign boundaries, which is what the ayanamsa controls. This result tells us about coordinate systems, not about cultural transmission.

**Results that are consistent with independent selection:**

The nakshatra-xiu total overlap (9 stars, p = 0.38) is what you would expect if two cultures independently picked bright stars near the ecliptic as calendar anchors. The faint-star excess is the only part of the mansion result that resists the null, and it is threshold-dependent. The most parsimonious reading is that the total overlap reflects independent selection and the faint-star composition is suggestive but not conclusive.

**Results that do not constrain anything:**

The timing result is consistent with chance under the author's mappings. The family synastry result is null in an underpowered sample. The geometry checks confirm properties of subtraction. The author's electional model describes the author's model. The house convergence and relocation results describe geometry, not transmission.

The paper's contribution is not seventeen findings about transmission. It is one finding that constrains interpretation (dignity: three of four states survived transmission intact, the 47.0% aggregate rate is a mapping artifact), one that validates the framework against known history (nakshatra-manazil: above chance), one that is genuinely ambiguous (primary directions), one that is suggestive but caveated (lunar mansions: faint-star excess), and a measurement framework that can be applied to any pair of astrological techniques with a shared origin. The framework is the contribution. The numbers are the first application of it.

### 5.7 What This Paper Does Not Show

This paper does not show that astrology "works." It does not show that any astrological technique has predictive validity. It does not show that the Hellenistic synthesis was historically real (though the historical consensus supports it). It does not show that agreement between traditions implies preservation of an original — the agreement could reflect independent discovery or later contact.

What it does show: the degree of structural agreement between Western and Vedic computational outputs, quantified against random baselines, with multi-seed ranges and ayanamsa sensitivity. Per-state dignity analysis reveals that domicile/swakshetra, exaltation/uchcha, and fall/neecha assignments are identical across the two traditions (100% agreement). Three of four dignity states survived transmission intact. The 47.0% aggregate rate is an artifact of mapping asymmetry and ayanamsa-driven sign-boundary crossings. The nakshatra-manazil overlap confirms known historical transmission. The timing systems agree at chance levels under the author's mappings. The node sign is overwhelmingly ayanamsa-dependent. Houses are overwhelmingly location-dependent. The nakshatra-xiu overlap is consistent with independent selection, with a threshold-dependent faint-star excess. The three-way mansion convergence is zero — shared stars don't produce shared placements when boundaries differ. Primary directions show partial cross-system survival that is genuinely ambiguous. Family synastry shows no signal in aspect density (N=26, underpowered).

## 6. Conclusion

A single Go binary ran seventeen measurements comparing Western, Vedic, and Arabic astrological frameworks.

Six of the seventeen are geometry checks — they confirm that subtraction works. Two measure the author's models, not traditional practice. Two measure location effects. One is underpowered. Six remain as measurements of cross-tradition agreement.

Of those six, the dignity result is the most informative — but in the opposite direction from the original interpretation. Per-state analysis reveals that domicile/swakshetra, exaltation/uchcha, and fall/neecha assignments are identical across the two traditions (100% agreement). Three of four dignity states survived transmission intact. The 47.0% aggregate rate is an artifact of mapping asymmetry (Western detriment has no Vedic equivalent) and ayanamsa-driven sign-boundary crossings. The dignity tables did not diverge. The nakshatra-manazil overlap (13 stars, p = 0.0000) confirms known historical transmission and validates the null model framework. Primary directions show partial cross-system survival (~51%) — a genuinely ambiguous result. Node sign survival (21.8%) is primarily a function of ayanamsa choice. Two-way mansion convergence (2.2%) is consistent with independent selection, with a threshold-dependent faint-star excess (p = 0.0003 at ≥ 2.5, weakening at higher thresholds). Three-way mansion convergence is zero — shared stars don't produce shared placements when the three systems define their sector boundaries differently. The shared-pool assumption in the null models is a significant caveat.

The measurement framework — cross-tradition comparison against computed null expectations, with multi-seed ranges and ayanamsa sensitivity — is the paper's primary contribution. The numbers are the first application of it. All phases are reproducible from the current baseline tool. No independent implementation exists.

The code is at github.com/aj-nt/empirical. The baselines are reproducible, the measurements falsifiable. The full system includes 35 API endpoints, a web dashboard, and a comprehensive manual at MANUAL.md.

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

[12] Moretti, F. (2005). Graphs, Maps, Trees: Abstract Models for a Literary History. Verso.

[13] Michel, J.-B. et al. (2011). Quantitative Analysis of Culture Using Millions of Digitized Books. Science, 331(6014), 176-182.

[14] Spencer, M. et al. (2004). Phylogenetics and the Cohesion of the Canterbury Tales Manuscript Tradition. Literary and Linguistic Computing, 19(3), 331-348.

[15] Roos, T. and Heikkilä, T. (2009). Evaluating Methods for Computer-Assisted Stemmatology Using Artificial Benchmark Data Sets. Literary and Linguistic Computing, 24(4), 417-433.
