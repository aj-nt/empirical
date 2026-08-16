# Phase 3: Predictive Tools — Implementation Plan

**Goal:** Build a dedicated Predictive Tools tab that surfaces all existing backend predictive endpoints in readable, interactive views.

**Backend already has:** transits, primary directions, solar arc, profections, zodiacal releasing, firdaria, solar return, secondary progressions, progressed cross, progressed draconic, draconic solar return, draconic transits, timing convergence.

**Frontend gap:** No dedicated predictive tab. Only "Progressed Cross" buried in Research. Solar return, profections, firdaria, directions, solar arc, zodiacal releasing — all have API methods but no UI.

## Tasks

### 3.1: PredictiveTools tab in App.tsx
- Add `'predictive'` to View type
- Add "Predictive" tab to tabs array
- Create `PredictiveTools` component with sub-tabs for each technique
- Wire into App.tsx rendering

### 3.2: Solar Return view
- Fetch `/api/solar-return` for a target year
- Display: ASC/MC, house placements, aspects to natal
- Year selector (current year ± 5)

### 3.3: Secondary Progressions view
- Fetch `/api/progressed` for a target date
- Display: progressed planet positions, aspects to natal
- Date selector

### 3.4: Primary Directions view
- Fetch `/api/directions` for an age
- Display: directed ASC/MC aspects
- Age slider

### 3.5: Solar Arc Directions view
- Fetch `/api/solar-arc`
- Display: solar arc directed planet aspects

### 3.6: Annual Profections view
- Fetch `/api/profection`
- Display: profected house, time lord, aspects

### 3.7: Firdaria view
- Fetch `/api/firdaria`
- Display: current period/sub-period, timeline

### 3.8: Zodiacal Releasing view
- Fetch `/api/zodiacal-releasing`
- Display: Lot of Fortune/Spirit periods

### 3.9: Timing Convergence view
- Fetch `/api/timing-convergence`
- Display: cross-system timing hits

### 3.10: End-to-end verification
- TypeScript, build, tests, browser verify
