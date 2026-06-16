# Empirical Astrology System — Manual

## Overview

The Empirical Astrology system is a computational engine for falsifiable
astrology. It measures what survived 2,000 years of independent cultural
evolution across Western, Vedic, and Chinese traditions — and provides
tools for actual astrological work.

**Thesis:** Hellenistic astrology (~150 BCE Alexandria) fused Babylonian
planet tracking, Egyptian decanic timekeeping, and Greek geometry into the
first horoscopic system. This spread east and west, evolving independently
in India (Vedic) and China (Ba Zi / Chinese astrology). The engine measures
transmission fidelity: what survived intact across all three branches.

**License:** Engine MIT, paper CC-BY 4.0.
**Repository:** github.com/aj-nt/empirical

---

## Quick Start

```bash
cd ~/Documents/repos/empirical
go build -buildvcs=false -o empirical ./cmd/recover/
./empirical serve 5432
```

Server runs at `http://localhost:5432`. All endpoints accept POST with JSON.

### First Query

```bash
curl -s -X POST http://localhost:5432/api/chart \
  -H 'Content-Type: application/json' \
  -d '{"name":"Test","year":1969,"month":2,"day":15,"hour":23,"minute":10,
       "tz_offset":-8,"lat":47.038,"lng":-122.901,
       "house_system":"P","sidereal":false,"show_aspects":true,
       "outer_planets":true,"highlight_patterns":true,"pattern_orb":3}'
```

Returns SVG chart with aspects and patterns highlighted.

---

## Architecture

```
cmd/recover/main.go     — Server entry point, SWE-dependent computation
cmd/baseline/main.go    — Baseline measurement tool (13 phases)
internal/dignity/       — Pure-math library (no CGo)
internal/server/        — HTTP server, endpoint wiring
internal/swe/           — Swiss Ephemeris CGo bindings
```

The `dignity` package is the core: all astrological logic lives there.
SWE-dependent code (house computation, ASC lines, transit scanning) lives
in `cmd/recover`.

---

## API Reference (33 endpoints)

All endpoints accept POST with JSON. Common fields in `ChartRequest`:

| Field | Type | Description |
|-------|------|-------------|
| name | string | Chart name |
| year | int | Birth year |
| month | int | Birth month (1-12) |
| day | int | Birth day |
| hour | int | Birth hour (local) |
| minute | int | Birth minute |
| tz_offset | float | Timezone offset from UTC (e.g. -8 for PST) |
| lat | float | Geographic latitude |
| lng | float | Geographic longitude |

### Core Chart

**`POST /api/chart`** — SVG chart with aspects and patterns.
Additional fields: `house_system` (P/W/K/etc), `sidereal` (bool),
`show_aspects` (bool), `outer_planets` (bool), `highlight_patterns` (bool),
`pattern_orb` (float).

**`POST /api/compute`** — Full chart data as JSON (positions, houses, aspects, patterns).

**`POST /api/aspects`** — Aspect catalog (conjunction, sextile, square, trine, opposition, quincunx) with cross-tradition classification.

**`POST /api/patterns`** — Detected aspect patterns (T-Square, Grand Trine, Yod, Grand Cross, Cradle, Kite, Stellium).

### Transits & Timing

**`POST /api/transits`** — Transit-to-natal aspects for a date range.
Additional: `start_date`, `end_date` (YYYY-MM-DD), `orb` (float), `sidereal` (bool).

**`POST /api/timing`** — Element-to-planet timing convergence.
Additional: `target_date` (YYYY-MM-DD).

### Synastry & Composite

**`POST /api/synastry`** — Aspects between two charts.
Additional: second person's birth data (`name2`, `y2`-`lo2`), `orb`.

**`POST /api/composite`** — Composite chart (midpoint method).
Additional: second person's birth data, `orb`.

### Relocation

**`POST /api/relocation`** — Compare natal chart at two locations.
Additional: `loc_a` {name, lat, lng}, `loc_b` {name, lat, lng}, `target_date`.

### Draconic (Soul Chart)

**`POST /api/draconic`** — Draconic chart (tropical positions shifted by north node).
Additional: `orb`.

**`POST /api/draconic-synastry`** — Draconic-to-draconic aspects between two charts.

**`POST /api/draconic-synastry-full`** — Full draconic synastry: draconic-draconic, tropical-draconic, draconic-tropical, plus bridges.

**`POST /api/draconic-transits`** — Draconic transit-to-natal aspects.
Additional: `start_date`, `end_date`, `orb`.

**`POST /api/draconic-transits-cross`** — Draconic transits compared in tropical vs sidereal frames.
Additional: `start_date`, `end_date`, `orb`.

**`POST /api/progressed-draconic`** — Progressed draconic chart.
Additional: `target_date`.

**`POST /api/draconic-solar-return`** — Draconic solar return chart.
Additional: `target_year`.

### Progressed

**`POST /api/progressed-cross`** — Progressed-to-natal aspects compared in tropical vs sidereal.
Additional: `target_date` (YYYY-MM-DD), `orb`.

### Primary Directions

**`POST /api/directions`** — Ptolemy's primary directions (1° RA = 1 year).
Additional: `age` (float, years), `orb`.

### Fixed Stars, Arabic Parts, Mansions

**`POST /api/stars`** — Fixed star conjunctions to natal planets.
Additional: `orb`.

**`POST /api/stars-cross`** — Fixed star conjunctions in tropical vs sidereal.

**`POST /api/arabic-parts`** — Arabic Parts (Lot of Fortune, Spirit, etc.) and their aspects.
Additional: `orb`.

**`POST /api/mansion-convergence`** — Nakshatra/xiu lunar mansion convergence.

### Advanced Techniques

**`POST /api/traditional`** — Traditional dignity scores (ruler, exaltation, triplicity, term, face).

**`POST /api/uranian`** — Uranian astrology (hard aspects only, 8 transneptunian planets).

**`POST /api/harmonic`** — Harmonic chart analysis.

**`POST /api/divisional`** — Vedic divisional charts (D-9 Navamsa, etc.).

**`POST /api/parans`** — Paran detection (star-to-star crossings on angles).

**`POST /api/declination`** — Declination aspects (parallel, contraparallel).

**`POST /api/firdaria`** — Firdaria periods (Persian planetary periods).

### Product Features

**`POST /api/interpretation`** — Natural-language chart interpretation.
Additional: `house_system`, `orb`. Returns planet-in-sign, planet-in-house,
aspect, and pattern descriptions.

**`POST /api/astrocartography`** — Planetary lines for world map rendering.
Additional: `lat_step` (float, default 2.0), `frame` ("tropical", "draconic", "cross").
Returns MC, IC, ASC, DSC lines for all planets.

**`POST /api/astrocartography-compare`** — Three-way comparison at a target location.
Additional: `lat_step`, `target_lat`, `target_lng`, `orb`.
Returns tropical, draconic, and cross hits merged.

**`POST /api/electional`** — Date scoring for launch/event timing.
Additional: `start_date`, `end_date` (YYYY-MM-DD), `orb`.
Scores each day on Moon house, Mercury sign, bad aspects, good aspects.
Returns ranked list.

---

## Baseline Phases (13 total)

Run with: `./baseline -phase <name> -n <count> -seed <seed> -orb <degrees>`

| Phase | Description | Key Result |
|-------|-------------|------------|
| draconic | Draconic-to-tropical aspect survival | 0% (geometrically guaranteed) |
| mansion | Nakshatra/xiu mansion convergence | 2.1% (mean 0.15/7) |
| arabic | Arabic Parts cross-system | 100% (ayanamsa-invariant) |
| relocation | Relocation house changes | Location-dependent |
| stars | Fixed star cross-system | 100% (ayanamsa-invariant) |
| timing | Element-to-planet timing | 1.78% tight, 2.18% generous |
| family | Family synastry patterns | Varies |
| progressed | Progressed-to-natal aspects | Standard |
| progressed-cross | Progressed cross-system | 100% (ayanamsa-invariant) |
| directions | Primary directions | Mean 3.24 ASC, 3.38 MC aspects |
| relocation-dignity | Dignity across locations | 100% identical (location-independent) |
| directions-cross | Primary directions cross-system | 50.6% survival |
| electional-cross | Electional best-day agreement | 57.1% same rank |

---

## Cross-System Measurement Table

What survives when you shift from tropical to sidereal (Lahiri ayanamsa)?

| Phase | Technique | Survival Rate | Frame-Dependent? |
|-------|-----------|--------------|-----------------|
| 1 | Dignity (sign-based) | 46.7% | Yes |
| 3 | Houses (ASC-based) | 83.0% | Yes |
| 4 | Timing (element-based) | 1.78% | Yes |
| 5 | Nodes | 21.4% | Yes |
| 7 | Draconic transits | 0% | No (geometric) |
| 8 | Mansion convergence | 2.1% | Yes |
| 9 | Arabic Parts | 100% | No (invariant) |
| 11 | Fixed stars | 100% | No (invariant) |
| 12 | Progressed cross | 100% | No (invariant) |
| 13 | Primary directions | 50.6% | Partial |
| 14 | Electional (best day) | 57.1% | Partial |

**Interpretation:** Techniques fall into three categories:
- **Invariant** (100%): Angular distances preserved under uniform shift.
  Fixed stars, Arabic Parts, progressed aspects.
- **Frame-dependent** (<100%): Sign-based, house-based, or element-based
  techniques that depend on the zodiac zero point.
- **Partial** (~50%): Techniques where the shift is nonlinear
  (transcendental OA→λ conversion in primary directions) or where
  multiple factors interact (electional scoring).

---

## Interpretation Engine

The interpretation engine produces deterministic natural-language
descriptions. No LLM dependency.

**Coverage:**
- 17 planets (Sun through Lilith, plus 8 Uranian)
- 12 signs with element/mode descriptions
- 12 houses with domain descriptions
- 6 aspect types (conjunction, opposition, trine, square, sextile, quincunx)
- 7 pattern types (T-Square, Grand Trine, Yod, Grand Cross, Cradle, Kite, Stellium)
- 24 planet-pair dynamics (Sun-Moon, Venus-Mars, Saturn-Uranus, etc.)
- Dignity states: domicile, detriment, exalted, fall, neutral

**Example output:**
```
Sun in Aquarius: core identity, vitality, conscious self,
  fixed air — innovative, collective, detached. in detriment —
  out of element, operating against its nature.

Sun in house 4: core identity, vitality, conscious self
  expressed through home, roots, family, private self, foundation.

Sun conjunction Mercury (orb 0.5°): conscious will meets
  emotional need — the fundamental axis of personality —
  merge and amplify — the two planets operate as one force.

T-Square involving Mars, Saturn, Uranus: dynamic tension
  between three planets — a pressure cooker that produces results.
```

---

## Astrocartography

Planetary lines on a world map showing where each planet sits on the
ASC, DSC, MC, or IC at a given moment.

**Three frames:**
- **Tropical** (`frame: "tropical"`): Standard Western positions.
  Personality lines — where your conscious self is angular.
- **Draconic** (`frame: "draconic"`): Positions shifted by north node.
  Soul lines — where your evolutionary intent is angular.
- **Cross** (`frame: "cross"`): Tropical MC/IC + draconic ASC/DSC.
  Personality-on-soul — where your conscious self activates your
  soul's angular points.

**Line types:**
- MC/IC: Vertical lines (constant longitude). RA-based.
- ASC/DSC: Curved lines. Computed via binary search using SWE houses.

**Three-way comparison** (`/api/astrocartography-compare`):
Given a target location, returns which planetary lines are nearby in
all three frames, merged and sorted by closest orb.

**Key finding:** MC/IC lines are NOT identical across frames. RA is
nonlinear under uniform longitude shift (atan2 nonlinearity). The
Python skill's claim was incorrect — confirmed by Go tests.

---

## Electional Astrology

Date scoring for launches, announcements, and timed events.

**Scoring model:**
- Moon house placement (-3 to +3): H10/H11 best, H12 worst
- Mercury sign condition (-1 to +2): Gemini best, Sagittarius/Pisces worst
- Bad aspects (-2 each): Mars square/opp Mercury, Neptune opp Uranus,
  Saturn conj Venus
- Good aspects (+1 each): Uranus trine Mercury, Pluto trine Jupiter,
  Sun sextile Chiron, Venus in H9/H10

**Output:** Ranked list of dates with scores, Moon/Mercury details,
and good/bad aspect descriptions.

**Cross-system finding:** Tropical and sidereal electional rankings
agree only 57.1% of the time. Moon house depends on ASC sign, which
shifts by ~24° with ayanamsa. This is a frame-dependent technique.

---

## Primary Directions

Ptolemy's oldest predictive technique: 1° of right ascension = 1 year
of life.

**Math:**
- ASC directed by oblique ascension (OA): OA_directed = OA_ASC + age
- MC directed by right ascension (RA): RA_directed = RA_MC + age
- OA→λ conversion via binary search (transcendental equation)
- RA→λ conversion exact: tan(λ) = tan(RA) / cos(ε)

**Cross-system finding:** 50.6% survival. The OA→λ conversion is
nonlinear under ayanamsa shift, unlike the simpler progressed case
where both natal and progressed positions shift equally.

---

## Relocation Analysis

**Key findings:**
- Dignity is 100% location-independent (sign-based, signs are ecliptic
  longitude, which doesn't change with location)
- Houses change with location (ASC shifts, whole-sign houses shift)
- Relocated Vertex can differ by 100+ degrees from birth Vertex
- Astrocartography lines show where each planet is angular

**Workflow:**
1. Compute relocated chart (same UTC JD, different lat/lng)
2. Check house changes for all planets
3. Check relocated Vertex conjunctions
4. Cross-reference with astrocartography lines
5. For couples: compare both relocated charts

---

## Draconic Features

The draconic chart shifts all positions by the north node longitude.
This creates a "soul chart" — the evolutionary intent layer.

**Key findings:**
- Draconic-to-tropical aspects: 0% survival (geometrically guaranteed —
  the NN offset breaks all aspects)
- Draconic transits cross-system: 0% (same reason)
- Draconic synastry bridges: tropical-draconic and draconic-tropical
  contacts between two people reveal soul-level connections

---

## Appendix A: Planet List

| Planet | SWE ID | Description |
|--------|--------|-------------|
| Sun | 0 | Core identity, vitality |
| Moon | 1 | Emotional nature, instincts |
| Mercury | 2 | Communication, intellect |
| Venus | 3 | Love, values, aesthetics |
| Mars | 4 | Drive, assertion |
| Jupiter | 5 | Expansion, faith |
| Saturn | 6 | Structure, discipline |
| Uranus | 7 | Innovation, rebellion |
| Neptune | 8 | Imagination, dissolution |
| Pluto | 9 | Power, transformation |
| Node | 10 | Mean north node |
| Chiron | 15 | Wounding and healing |
| Lilith | 12 | Mean apogee (shadow feminine) |
| Ceres | 17 | Nurturing, cycles |
| Pallas | 18 | Pattern recognition |
| Juno | 19 | Partnership contracts |
| Vesta | 20 | Devotion, sacred focus |
| Cupido | 40 | Uranian: family, community |
| Hades | 41 | Uranian: depth, ancient |
| Zeus | 42 | Uranian: directed energy |
| Kronos | 43 | Uranian: authority, height |
| Apollon | 44 | Uranian: expansion, science |
| Admetos | 45 | Uranian: endurance, stone |
| Vulcanus | 46 | Uranian: force, intensity |
| Poseidon | 47 | Uranian: illumination, spirit |

---

## Appendix B: Aspect Types

| Aspect | Angle | Classification |
|--------|-------|----------------|
| Conjunction | 0° | Universal (all three traditions) |
| Semi-sextile | 30° | Partial |
| Sextile | 60° | Universal |
| Square | 90° | Universal |
| Trine | 120° | Universal |
| Quincunx | 150° | Partial |
| Opposition | 180° | Universal |

---

## Appendix C: Sign Descriptions

| Sign | Element/Mode | Keywords |
|------|-------------|----------|
| Aries | Cardinal Fire | Initiatory, direct, competitive |
| Taurus | Fixed Earth | Steady, sensual, accumulating |
| Gemini | Mutable Air | Curious, dual, networking |
| Cancer | Cardinal Water | Protective, nurturing, cyclical |
| Leo | Fixed Fire | Radiant, creative, commanding |
| Virgo | Mutable Earth | Analytical, refining, serving |
| Libra | Cardinal Air | Balancing, relating, aesthetic |
| Scorpio | Fixed Water | Penetrating, transformative, intense |
| Sagittarius | Mutable Fire | Seeking, philosophical, expansive |
| Capricorn | Cardinal Earth | Ambitious, structuring, enduring |
| Aquarius | Fixed Air | Innovative, collective, detached |
| Pisces | Mutable Water | Dissolving, transcendent, compassionate |

---

## Appendix D: House Descriptions

| House | Domain |
|-------|--------|
| 1 | Self, persona, physical body |
| 2 | Resources, values, self-worth |
| 3 | Communication, siblings, learning |
| 4 | Home, roots, family, foundation |
| 5 | Creativity, pleasure, self-expression |
| 6 | Work, health, service, routines |
| 7 | Partnership, marriage, open enemies |
| 8 | Shared resources, transformation, intimacy |
| 9 | Expansion, travel, philosophy |
| 10 | Career, public role, reputation |
| 11 | Community, networks, collective future |
| 12 | Retreat, unconscious, dissolution |

---

## Appendix E: Pattern Types

| Pattern | Configuration | Description |
|---------|--------------|-------------|
| T-Square | Two squares + opposition | Dynamic tension, pressure cooker |
| Grand Trine | Three trines in triangle | Effortless talent, potential inertia |
| Yod | Two sextiles + two quincunxes | Finger of god, fated pressure point |
| Grand Cross | Four squares in cross | Constant tension, extraordinary resilience |
| Cradle | Sextiles + trines | Bowl of support, nurturing structure |
| Kite | Grand Trine + opposition | Talent with release valve |
| Stellium | 3+ planets in one sign/house | Intense focus, concentrated energy |

---

## Appendix F: Dignity States

| State | Meaning |
|-------|---------|
| Domicile (Ruler) | Planet in sign it rules — at home, full strength |
| Detriment | Planet opposite its domicile — out of element, weakened |
| Exaltation | Planet in sign of exceptional clarity — elevated |
| Fall | Planet opposite its exaltation — diminished, struggling |
| Neutral | Neither strengthened nor weakened by this sign |

---

## Building and Testing

```bash
# Build server
go build -buildvcs=false -o empirical ./cmd/recover/

# Build baseline tool
go build -buildvcs=false -o baseline ./cmd/baseline/

# Run all tests
go test ./... -count=1

# Run specific package tests
go test ./internal/dignity/ -v -count=1
go test ./internal/server/ -v -count=1

# Run baseline
./baseline -phase directions-cross -n 1000 -seed 42 -orb 3
```

---

## Project Structure

```
~/Documents/repos/empirical/
├── cmd/
│   ├── recover/main.go      — Server (33 endpoints)
│   └── baseline/main.go     — Baseline tool (13 phases)
├── internal/
│   ├── dignity/             — Pure-math library
│   │   ├── dignity.go       — Dignity convergence
│   │   ├── aspect.go        — Aspect catalog
│   │   ├── pattern.go       — Pattern detection
│   │   ├── transit.go       — Transit scanning
│   │   ├── draconic.go      — Draconic chart + cross-system
│   │   ├── progressed.go    — Progressed + cross-system
│   │   ├── directions.go    — Primary directions
│   │   ├── astrocartography.go — Astrocartography primitives
│   │   ├── interpretation.go   — Interpretation engine
│   │   ├── timing.go        — Element-to-planet timing
│   │   ├── arabic_parts.go  — Arabic Parts
│   │   ├── stars.go         — Fixed stars
│   │   ├── synastry.go      — Synastry
│   │   ├── relocation.go    — Relocation
│   │   ├── composite.go     — Composite charts
│   │   ├── traditional.go   — Traditional dignity
│   │   ├── uranian.go       — Uranian astrology
│   │   ├── harmonic.go      — Harmonic charts
│   │   ├── divisional.go    — Vedic divisional charts
│   │   ├── parans.go        — Paran detection
│   │   ├── declination.go   — Declination aspects
│   │   ├── firdaria.go      — Firdaria periods
│   │   ├── batch.go         — Batch transit scanning
│   │   └── *_test.go        — Tests (all passing)
│   ├── server/
│   │   ├── server.go        — HTTP server + endpoint wiring
│   │   └── server_test.go   — Server tests
│   └── swe/
│       └── swe.go           — Swiss Ephemeris CGo bindings
├── ephe/                    — Embedded ephemeris files
├── MANUAL.md                — This document
├── go.mod
└── go.sum
```
