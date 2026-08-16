# Empirical → TimePassages Feature Parity — Project Plan

**Date**: 2026-07-29
**Status**: Draft
**Context**: Empirical has 7 weeks of development (first commit June 8, 2026). 39 CLI subcommands, 36 API endpoints, React web GUI with local storage. Bi-wheel completed today. TimePassages has been in development since ~2000.

---

## Goal

Reach feature parity with TimePassages V6 as an open-source astrology application that a practicing astrologer would choose over a $200-300 commercial product.

## Strategic Context

**Empirical is a better engine. TimePassages is a better product.** The engine gap is structural — TimePassages can't add cross-system verification without rewriting their core. The product gap is scope — we can add bi-wheels, animation, and point-and-click because it's engineering, not research.

**Open source changes the equation.** Interpretive content becomes community-contributed. House systems and ayanamsas become PRs from users who need them. The timezone database is a solved problem in Go — import it. Chart rectification is algorithmic. The things that took TimePassages 26 years to build alone can be built faster with contributors.

**The localhost advantage is real.** Server-rendered SVG on localhost means re-rendering the entire chart on every animation frame is instant. The "server vs client rendering" debate that dominates web dev is irrelevant here. The bi-wheel spike proved the architecture is sound.

---

## Phase 0: Automated UX Testing (Now)

**Why this exists**: The chart components — `ChartWheel`, `BiWheel`, `PlanetTable`, `AspectGrid` — have zero tests. These are exactly the components that will break during Phase 1-3 UI work. Without automated testing, every UI change is a manual "does it look right?" check in the browser. That's slow, error-prone, and demoralizing.

**Current state**: Vitest with 5 test files (API client, theme, export, astrology utils, ThemeSwitcher). Playwright with 1 smoke spec (6 tests — "did the app crash?"). No visual snapshots. No component tests for chart components. No API mocking infrastructure.

**Goal**: Every chart component has visual snapshot tests and interaction tests before Phase 1 begins. When you change the wheel layout and accidentally make all the text gray, the test catches it in 2 seconds instead of you finding it 20 minutes later.

### 0.1 Playwright Visual Snapshots (Highest Impact)
**Effort**: 2-3 sessions
**Approach**: Playwright's built-in `toHaveScreenshot()` compares a screenshot against a checked-in baseline. One test per component, one screenshot per state. Catches layout regressions, color issues, text overlap, missing elements, font changes.

**Components to cover**:
- `ChartWheel` — natal wheel with planets, aspects, house cusps
- `BiWheel` — inner natal + outer transits with aspect lines
- `PlanetTable` — planet positions table
- `AspectGrid` — aspect grid with orbs
- `Dashboard` — natal dashboard with patterns, dignities
- `PageDesigner` — report composer with blocks

**Architecture**: Add `e2e/visual/` directory. Each component gets a spec file. Tests create a chart via the UI (or seed IndexedDB directly), navigate to the component, take a screenshot. Baselines stored in `e2e/visual/screenshots/`. Run with `npx playwright test --update-snapshots` to generate baselines, then `npx playwright test` to compare.

### 0.2 API Mocking with MSW
**Effort**: 1-2 sessions
**Approach**: Install `msw` (Mock Service Worker). Create handlers for all API endpoints that return known, deterministic data. This lets component tests run without the Go server — faster, more reliable, and you can test edge cases (empty response, error state, missing fields).

**Why MSW over manual fetch mocking**: MSW intercepts at the network level. Components that call `fetch('/api/chart')` don't know they're being mocked. Works identically in Vitest (jsdom) and Playwright (real browser). One set of handlers for both.

**Handlers to build**:
- `POST /api/chart` → known SVG string
- `POST /api/bi-wheel` → known bi-wheel SVG
- `POST /api/interpretation` → known interpretation JSON
- `POST /api/transits` → known transit data
- `POST /api/patterns` → known pattern data
- `POST /api/recover` → known natal data

### 0.3 Component Tests for Chart Components
**Effort**: 2-3 sessions
**Approach**: Use React Testing Library + MSW to test chart components in isolation. Each component gets a spec file with tests for: loading state, data display, error state, empty state, interaction (click, hover).

**Components to cover**:
- `ChartWheel` — renders SVG, shows loading spinner, shows error message, export buttons work
- `BiWheel` — renders bi-wheel SVG, inner/outer chart selectors work, export buttons work
- `PlanetTable` — renders planet rows, shows retrograde markers, sorts correctly
- `AspectGrid` — renders aspect cells, shows orb values, highlights tight aspects
- `Dashboard` — renders pattern list, shows dignity badges, dispositor tree renders

### 0.4 Test Data Fixtures
**Effort**: 1 session
**Approach**: Create a `test/fixtures/` directory with known birth data and expected API responses. AJ's chart is the primary fixture — we know exactly what it should look like. Add 2-3 additional charts with different characteristics (night chart, southern hemisphere, extreme latitude).

**Fixtures**:
- `aj.json` — AJ's birth data + expected natal positions
- `cait.json` — Cait's birth data (for synastry/composite tests)
- `night-chart.json` — a night chart (for sect-dependent tests)
- `southern.json` — southern hemisphere (for house system tests)

### 0.5 CI Integration
**Effort**: 1 session
**Approach**: Add a GitHub Actions workflow that runs `npm test` (Vitest) and `npx playwright test` (Playwright) on every PR. Playwright needs the Go server running — the existing `playwright.config.ts` already handles this with `webServer`. Upload screenshots on failure for debugging.

### 0.6 Test Coverage Baseline
**Effort**: 0.5 sessions
**Approach**: Run `npm run test:coverage` to get current coverage numbers. Set a coverage threshold in `vite.config.ts` that prevents coverage from dropping. Start low (current coverage) and ratchet up as tests are added.

---

## Phase 1: Critical UX Gaps (After Phase 0)

These are the features a practicing astrologer reaches for daily. Without them, no one switches.

### 1.1 Tri-Wheel (natal + progressed + transits)
**Effort**: 1-2 days
**Approach**: Same code path as bi-wheel. Add a third ring at r=8. `RenderTriWheelSVG` in `chart.go`, `/api/tri-wheel` endpoint, `TriWheel.tsx` component. The bi-wheel already handles two rings with aspect lines between them — a tri-wheel is the same thing with three.

### 1.2 Point-and-Click Wheel Interpretations
**Effort**: 1-2 weeks
**Approach**: The wheel is an SVG. Every planet circle and house cusp has known coordinates. On click/hover, map the (x,y) to the nearest planet or house cusp, look up the interpretation from the existing interpretation engine, display in a tooltip or side panel. The hard part is the interpretation text — but the engine already generates `planet_signs`, `planet_houses`, and `aspects` strings. This is a UI layer on existing data.

**Architecture**: Add `onClick` handlers to the SVG elements in `ChartWheel.tsx`. On click, determine what was clicked (planet at angle X, house cusp Y), fetch the relevant interpretation string from the already-loaded interpretation data, display in a positioned popover. No new backend endpoints needed — the interpretation data is already fetched for the Natal tab.

### 1.3 Chart Animation (Transits Over Time)
**Effort**: 1-2 weeks
**Approach**: The server is localhost. Re-render the entire SVG on every frame. Add a date slider to the bi-wheel/tri-wheel view. On slider change, POST new transit date to `/api/bi-wheel` or `/api/tri-wheel`, inject the new SVG. At 60fps this is 60 POSTs/second — on localhost, that's instant. The SVG is ~19KB. No WebGL, no Canvas, no D3 rewrite. Just re-render.

**Architecture**: Add `animate` mode to `BiWheel.tsx` / `TriWheel.tsx`. Date range slider (start/end). Play/pause button. On each frame, compute the transit date, call the API, swap the SVG. Use `requestAnimationFrame` for smooth playback. Cache the last few frames if needed, but localhost latency is sub-millisecond — caching may not be necessary.

### 1.4 Multiple House Systems
**Effort**: 1 week
**Approach**: Swiss Ephemeris already supports all major house systems via `swe_houses()`. The `house_system` parameter is already threaded through the API. Add the missing systems to the dropdown: Koch, Campanus, Regiomontanus, Equal (ASC), Whole Sign. Each is a single character code passed to SWE. The math is done. This is a UI change + validation.

**Systems to add**: Koch (`K`), Campanus (`C`), Regiomontanus (`R`), Equal/ASC (`E`), Whole Sign (`W`). Already have: Placidus (`P`).

### 1.5 Real Timezone Database
**Effort**: 1-2 weeks
**Approach**: Import `github.com/evanoberholster/timezoneLookup` or similar Go timezone library. Replace the current longitude-based estimation (`lon/15` rounded to nearest 0.5h) with actual IANA timezone lookups. This matters for historical charts — DST rules change over time, and political boundaries don't follow longitude lines.

**Architecture**: Add a timezone lookup to the geocode pipeline. When a city is selected, look up its IANA timezone. When a birth date is entered, compute the correct UTC offset for that date using the timezone's historical DST rules. Store the IANA timezone alongside the birth data. The current `tz_offset` field becomes a fallback for locations not in the database.

### 1.6 Multiple Ayanamsas
**Effort**: 2-3 days
**Approach**: Swiss Ephemeris supports all major ayanamsas via `swe_set_sid_mode()`. Add a dropdown to the chart form. The `sidereal` flag already exists in the API. Add the 8 ayanamsas TimePassages supports: Fagan-Bradley, Lahiri (already have), Deluce, Raman, Usha Shashi, Krishnamurti, Djwhal-Khul, Sri Yukteswar. Each is a constant in SWE.

---

## Phase 2: Predictive Techniques (Weeks 12-16)

### 2.1 Lunar and Planetary Returns
**Effort**: 2-3 weeks
**Approach**: Solar return already exists (`/api/solar-return`). Lunar return is the same algorithm with the Moon instead of the Sun — binary search for the exact moment the Moon returns to its natal position. Planetary returns are the same for each planet. Add `/api/lunar-return`, `/api/planetary-return` (parameterized by planet). Frontend: add Return tabs or a Return selector in the existing transit/progression views.

### 2.2 Solar Arc Directions
**Effort**: 1-2 weeks
**Approach**: Solar Arc is simpler than secondary progressions — all planets advance by the same arc (the Sun's progressed motion). Add `solarArcPositions` to the existing progression code. Add `/api/solar-arc` endpoint. Add Solar Arc as an option in the bi-wheel/tri-wheel outer ring selector.

### 2.3 Chart Rectification Tools
**Effort**: 2-3 weeks
**Approach**: Three methods from TimePassages:
1. **ASC animation**: Same as chart animation (Phase 1.3) but varying birth time instead of transit date. Slider for birth time, re-render the wheel on each change. User watches the ASC move and house cusps shift.
2. **Character-based**: Compare the person's known traits against the chart's ASC sign at different birth times. This is interpretive, not algorithmic — provide the ASC sign for each time and let the user judge.
3. **Uranus transit method**: Find when transiting Uranus crossed the angles (ASC/DSC/MC/IC) and match against major life events. Algorithmic — compute Uranus transits to angles for a range of birth times, compare against user-provided event dates.

---

## Phase 3: Content & Polish (Weeks 16-20)

### 3.1 Sabian Symbols
**Effort**: 2-3 days
**Approach**: 360 strings. Data entry, not code. Add a `sabian_symbols.json` file with the degree → symbol mapping. Add a Sabian tab or integrate into the point-and-click interpretation (Phase 1.2). The symbols are public domain (originally published 1925-1932).

### 3.2 Chart Shape Detection
**Effort**: 1 week
**Approach**: Marc Edmund Jones' 7 chart shapes (bowl, bucket, splash, seesaw, bundle, locomotive, splay) are geometric classifications of planet distribution. Algorithm: sort planets by longitude, find the largest gap, classify based on gap size and planet clustering. Add to the pattern detection engine. Display in the Dashboard or Natal tab.

### 3.3 Expanded Asteroid Catalog
**Effort**: 1-2 weeks (mostly data ingestion)
**Approach**: Swiss Ephemeris supports thousands of asteroids. The engine already handles optional bodies via `CalcUTErr`. Add the major asteroids astrologers use: Eris, Sedna, Haumea, Hygeia, Astraea, and the centaurs (Pholus, Nessus, etc.). Each needs an ephemeris file download and a name/ID mapping. The frontend already has show/hide toggles for asteroid categories.

### 3.4 Keyboard Shortcuts
**Effort**: 2-3 days
**Approach**: Add keyboard event handlers to App.tsx. Tab switching (1-9 for tabs), chart navigation (arrows for prev/next chart), export (Ctrl+S for SVG, Ctrl+Shift+S for PNG), new chart (Ctrl+N). Document in a help panel.

### 3.5 Sample Charts
**Effort**: 1-2 days
**Approach**: Bundle 10-20 sample charts (public figures with known birth data) as a JSON file. On first launch, offer to import them. Include a mix of chart types to demonstrate features: a stellium chart, a bowl chart, a chart with a grand cross, etc.

---

## Phase 4: Interpretive Content (Ongoing)

This is the hardest gap to close. Henry Seltzer spent 20+ years writing TimePassages' interpretations. No amount of code replaces authored text.

### 4.1 Community Contribution Pipeline
**Approach**: The interpretation engine is template-based and deterministic. Make the templates editable. Add a "contribute interpretation" flow where users can submit improved text for planet-in-sign, planet-in-house, or aspect interpretations. Review via GitHub PRs. Over time, the interpretation quality improves through community effort.

### 4.2 Interpretation Depth Tiers
**Approach**: Not every user wants a 25-page report. Offer three depth levels:
- **Quick**: Planet-in-sign + planet-in-house + major aspects (1-2 pages)
- **Standard**: Above + patterns + dispositor tree + lunar phase + dignities (5-10 pages)
- **Full**: Above + fixed stars + midpoints + declinations + Arabic parts + traditional layer (15-25 pages)

The engine already computes all of this. The tiering is about presentation, not computation.

### 4.3 Multi-Language Support
**Approach**: The interpretation templates are strings. Extract them to a locale file. Community translators can provide translations. This is a structural advantage over TimePassages — their interpretations are English-only.

---

## Phase 5: Professional Practice (Weeks 20-24)

### 5.1 Client Management
**Effort**: 1-2 weeks
**Approach**: Add a "clients" layer on top of the existing chart database. Group charts by client. Add client notes, session dates, billing status. This is a CRUD app on top of IndexedDB.

### 5.2 Branded Reports
**Effort**: 1 week
**Approach**: The Page Designer already exists. Add a "branding" section: practice name, logo, contact info, custom CSS. Apply to exported reports. This is what TimePassages' commercial license provides.

### 5.3 PDF Export
**Effort**: 1 week
**Approach**: The Page Designer exports HTML. Add a server-side PDF renderer (Go's `github.com/jung-kurt/gofpdf` or similar) that takes the same block-based layout and produces a PDF. Or use the browser's print-to-PDF with a print-optimized stylesheet (already partially done).

---

## Phase 6: Ecosystem & Distribution (Ongoing)

### 6.1 Mobile Companion App
**Approach**: The API is JSON. A React Native or PWA mobile app that talks to the same backend. Chart viewing, transit notifications, quick interpretations. Not a full replacement for the desktop app, but a companion for on-the-go use.

### 6.2 Plugin System
**Approach**: The interpretation engine is template-based. Make it pluggable — users can write custom interpretation modules in JavaScript or Python that plug into the pipeline. This is how you get community-contributed interpretive content at scale.

### 6.3 Chart Marketplace
**Approach**: A community repository of chart data (public figures, historical events) with verified birth data. Like Rodden's database but open-source and community-maintained.

---

## What We Don't Need to Build

Some TimePassages features aren't worth replicating:

- **Video tutorials** — YouTube exists. Community will create content.
- **Astro Clock (real-time sky)** — Nice to have, not critical. Low priority.
- **Multi-window support** — Browser tabs are multi-window. SPA tab-based UI is fine.
- **Custom report writer (scriptable)** — The CLI and API are the scriptable interface. `jq` is more powerful than any built-in report writer.
- **Commercial license branding** — Open source. No license codes needed.
- **Installation videos** — It's a web app. There's nothing to install.

---

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Interpretive content quality lags behind TimePassages for years | High | Medium | Community contribution pipeline. Tiered depth. Multi-language as differentiator. |
| Chart animation performance on complex wheels | Low | Low | Localhost latency is sub-ms. SVG is 19KB. 60fps is 60 POSTs/sec — trivial. |
| Timezone database edge cases (historical DST changes) | Medium | Medium | Use a well-maintained Go timezone library. Fall back to manual offset for edge cases. |
| Community doesn't materialize | Medium | High | Build the critical UX gaps first (Phase 1). The product must be usable solo before it can attract contributors. |
| Scope creep — trying to match every TimePassages feature | High | Medium | This plan is the scope. New features go through the same prioritization: does a practicing astrologer need this daily? |

---

## Immediate Next Actions (This Week)

1. **Tri-wheel** — Same code path as bi-wheel. Add third ring. 1-2 days.
2. **House systems** — Add Koch, Campanus, Regiomontanus, Equal, Whole Sign to the dropdown. SWE already supports them. 1-2 days.
3. **Ayanamsas** — Add the 7 missing ayanamsas to the dropdown. SWE constants. 1 day.
4. **Timezone database** — Research Go timezone libraries. Prototype integration with the geocode pipeline. 2-3 days.

These four items close the most glaring calculation gaps and are all low-risk, high-visibility wins.

---

## Success Metrics

- **3 months**: A practicing astrologer can cast a natal chart, view transits in a bi-wheel/tri-wheel, click planets for interpretations, and export a report. They'd still prefer TimePassages for daily use, but they'd check Empirical for research.
- **6 months**: The practicing astrologer uses Empirical for daily client work. They miss some polish (keyboard shortcuts, sample charts) but the core workflow is there. They contribute interpretations for their specialty.
- **12 months**: Empirical is the default recommendation for astrology students. It's free, it works on any device, and the community has built interpretive content that rivals commercial products. TimePassages still has the professional practice features, but Empirical has the ecosystem.
