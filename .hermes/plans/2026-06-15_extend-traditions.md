# Extend Measurement Framework: Tibetan, Persian, Medieval Traditions

## Goal

Add Tibetan (Tsi), Persian (Medieval Islamic), and Medieval European astrological traditions to the computational comparison framework.

## Honest Assessment

Most of what these traditions add is thin. They are variations on themes already measured:

- **Tibetan** = Vedic astrology with a different ayanamsa and some unique timing systems. The dignity table, houses, aspects, and lunar mansions are the same as Indian. Adding it would mostly produce ayanamsa sensitivity runs under a new label.

- **Persian** = Hellenistic astrology with firdaria (already in the engine), Arabic Parts (already measured), and the Arabic lunar mansions (manazil al-qamar). The dignity, aspects, and houses are Hellenistic — same as modern Western.

- **Medieval** = Hellenistic astrology with different house systems (Regiomontanus, Alcabitius, Campanus) and the same techniques already measured (primary directions, profections, Parts).

The one genuinely new measurement is the **Arabic lunar mansions** (manazil al-qamar): 28 mansions with different determinative stars from both nakshatras and xiu. A three-way mansion comparison (Indian, Chinese, Arabic) is novel and would strengthen or weaken the faint-star finding.

## What's Actually Measurable

### 1. Arabic Lunar Mansions (manazil al-qamar) — NEW
- 28 mansions, each with a determinative star
- Derived from Indian nakshatras but with different star assignments
- Compare against nakshatras (27) and xiu (28) in a three-way null model
- This is the one genuinely new measurement

### 2. Medieval House Systems — EXTENSION of Phase 3
- Add Regiomontanus, Alcabitius, Campanus to the existing 5 systems
- Swiss Ephemeris supports all three via swe.Houses
- Measures whether these systems converge with the existing 5

### 3. Tibetan Dasha System — EXTENSION of Phase 4
- Tibetan astrology has its own planetary period system (similar to Vimshottari but with different period lengths)
- Research needed: what are the actual period assignments?
- If they differ from Vimshottari, add as a fourth timing system

### 4. Persian/Medieval Dignity — POSSIBLE EXTENSION of Phase 1
- Research needed: do Persian or Medieval sources document dignity assignments that differ from standard Hellenistic?
- If yes, add as additional dignity systems
- If no (likely), skip — same as Western

### 5. Tibetan Ayanamsa — MINOR
- Tibetan astrology uses a different ayanamsa value
- This is just another ayanamsa sensitivity run, not a new measurement
- Worth documenting but not a new phase

## What's NOT Worth Adding

- **Aspect catalogs**: All three use the same Ptolemaic aspects. Geometry is geometry.
- **Fixed stars**: Same geometry check.
- **Progressions**: Same geometry check.
- **Node sign**: Same geometry.
- **Zodiac comparison**: Tibetan = sidereal (different ayanamsa), Persian/Medieval = tropical. Already covered.
- **Electional**: Each tradition has complex electional rules. Modeling them computationally is a separate research project, not a measurement extension.
- **Relocation**: Same geometry.

## Proposed Plan

### Step 1: Research Arabic Lunar Mansions
- Compile the 28 manazil determinative stars from primary sources (al-Biruni, Ibn Qutayba, or modern references)
- Verify magnitudes from Swiss Ephemeris catalog
- Identify which stars are shared with nakshatras and xiu
- This is the critical data-gathering step

### Step 2: Three-Way Mansion Null Models
- Extend the null model framework to three systems
- Null 1 (uniform): each selects 27-28 stars from combined pool
- Null 2 (brightness-weighted): selection proportional to exp(-magnitude)
- Null 3 (brightness + ecliptic): add ecliptic proximity weight
- Compute expected pairwise and three-way overlaps
- Compare observed to null

### Step 3: Add Medieval House Systems to Phase 3
- Add Regiomontanus ('R'), Alcabitius ('B' or 'A'), Campanus ('C') to house convergence baseline
- Swiss Ephemeris house system codes: Regiomontanus='R', Alcabitius='B', Campanus='C'
- Re-run Phase 3 with 8 systems instead of 5
- Report new convergence rates

### Step 4: Research Tibetan Dasha
- Find Tibetan planetary period assignments
- If they differ from Vimshottari, implement in timing module
- Add to Phase 4 as a fourth system

### Step 5: Research Persian/Medieval Dignity
- Check primary sources for dignity table differences
- Likely outcome: same as Hellenistic → skip

### Step 6: Update Paper
- Add Arabic mansions as Phase 16 (or new section in Phase 8)
- Add Medieval houses to Phase 3 results
- Add Tibetan dasha to Phase 4 (if data found)
- Update abstract, methods, results, discussion

## Files Likely to Change

- `cmd/baseline/main.go` — new phases or extended existing phases
- `internal/dignity/mansion.go` — three-way comparison logic
- `internal/dignity/house.go` — new house systems (just constants)
- `docs/paper/paper.md` — new results

## Risks

- **Arabic mansion star data quality**: Wikipedia has a list but primary source verification is needed. Different sources may list different determinative stars for the same mansion.
- **Tibetan dasha data availability**: Tibetan astrological texts are less translated than Indian. The period assignments may not be readily available in English.
- **Diminishing returns**: After Arabic mansions, the remaining additions are minor. The paper may not justify a full rewrite for "we added three house systems and the convergence rate went from 83.0% to 82.7%."

## Open Questions

1. Do we have reliable determinative star data for the 28 Arabic manazil?
2. Are Tibetan dasha period assignments documented in accessible English sources?
3. Should this be a paper extension or a separate short paper on the three-way mansion comparison?
