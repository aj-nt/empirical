# Phase 2: Core Views — Implementation Plan

> **For Hermes:** Use subagent-driven-development skill to implement this plan task-by-task.

**Goal:** Upgrade Natal, Transits, and Synastry views from placeholder text to professional-grade tables, charts, and interactive diagrams.

**Architecture:** Each view is a self-contained React component fetching from existing API endpoints. D3.js for the dispositor tree and time map. No new backend endpoints — all data comes from `/api/interpretation`, `/api/traditional`, `/api/transits`, `/api/synastry`, `/api/composite`, and `/api/chart`.

**Tech Stack:** React 19, TypeScript, Tailwind CSS v4, D3.js, TanStack Query

**Repo:** `~/Documents/repos/empirical`
**Frontend:** `web/`
**Server:** `./empirical serve 5000`

---

## Pre-Flight Checks

Before any task, verify:
1. Server is running: `curl -s -o /dev/null -w "%{http_code}" http://localhost:5000/` → 200
2. TypeScript compiles: `cd web && npx tsc -b` → no errors
3. Vite builds: `cd web && npm run build` → success

---

## Task 1: Fix `dispositor_tree` TypeScript type

**Objective:** The `TraditionalResponse.dispositor_tree` is typed as `Record<string, string>` but the API returns a structured object with `nodes`, `final_dispositors`, and `mutual_receptions`.

**Files:**
- Modify: `web/src/lib/types.ts`

**Step 1: Add correct types**

Add after the `TraditionalResponse` interface:

```typescript
export interface DispositorNode {
  planet: string;
  sign: string;
  dispositor: string;
  is_final: boolean;
  in_loop: boolean;
}

export interface DispositorTree {
  nodes: DispositorNode[];
  final_dispositors: string[];
  mutual_receptions: string[][] | null;
}
```

**Step 2: Update `TraditionalResponse`**

Change `dispositor_tree: Record<string, string>` to `dispositor_tree: DispositorTree`.

**Step 3: Verify**

```bash
cd web && npx tsc -b
```

Expected: no errors.

---

## Task 2: Create `DispositorTree` D3 component

**Objective:** Interactive D3 tree diagram showing planetary dispositor chains. Mars → Venus → Jupiter → etc. Final dispositors highlighted. Mutual reception loops shown as bidirectional edges.

**Files:**
- Create: `web/src/components/chart/DispositorTree.tsx`

**Step 1: Create component skeleton**

```typescript
import { useEffect, useRef } from 'react';
import * as d3 from 'd3';
import type { DispositorTree as DispositorTreeData } from '../../lib/types';

interface Props {
  data: DispositorTreeData;
}

export function DispositorTree({ data }: Props) {
  const svgRef = useRef<SVGSVGElement>(null);

  useEffect(() => {
    if (!svgRef.current || !data.nodes.length) return;
    // D3 rendering here
  }, [data]);

  return (
    <svg ref={svgRef} style={{ width: '100%', height: '100%', minHeight: 400 }} />
  );
}
```

**Step 2: Build D3 tree layout**

- Convert `nodes` into a hierarchy: find roots (planets that are their own dispositor = final dispositors, or planets not disposed by any other planet)
- Use `d3.tree()` with `nodeSize([60, 120])` for horizontal layout
- Draw edges as curved paths
- Draw nodes as circles with planet glyphs
- Color final dispositors gold, mutual reception nodes blue
- Add tooltips on hover showing planet → sign → dispositor chain

**Step 3: Verify**

```bash
cd web && npx tsc -b && npm run build
```

Expected: no errors, builds.

---

## Task 3: Upgrade `NatalView` — tables with glyphs

**Objective:** Replace flat text lists with sortable tables showing planet glyphs, signs, houses, dignities, and retrograde markers.

**Files:**
- Modify: `web/src/App.tsx` (NatalView function)
- Create: `web/src/components/natal/PlanetTable.tsx`

**Step 1: Create `PlanetTable` component**

Props: `planets: PlanetPosition[]` (from `/api/base-chart`), `dignity?: Record<string, string>` (from `/api/compare` or `/api/traditional`)

Table columns: Glyph | Planet | Sign | House | Degree | Speed | RX | Dignity

Use `planetGlyph` and `signGlyph` from `lib/astrology.ts`.

**Step 2: Fetch base-chart data in NatalView**

Add a second API call to `api.baseChart(data)` alongside the existing `api.interpretation()` call. Use `Promise.all` to fetch both in parallel.

**Step 3: Wire into NatalView**

Replace the "Planets in Signs" and "Planets in Houses" text sections with:
- `PlanetTable` showing all planets with sign, house, degree, speed, retrograde
- Keep the text interpretation sections below the table

**Step 4: Verify**

```bash
cd web && npx tsc -b && npm run build
```

Expected: no errors, builds.

---

## Task 4: Add hidden contacts section to NatalView

**Objective:** Display antiscia, declination parallels, star aspects, and midpoints that are already in the `InterpretationResponse`.

**Files:**
- Modify: `web/src/App.tsx` (NatalView function)

**Step 1: Add sections for hidden contacts**

After the existing Patterns section, add:

```tsx
{interp.chart_ruler && (
  <Section title="Chart Ruler">
    <p className="text-sm">{interp.chart_ruler}</p>
  </Section>
)}
{interp.final_dispositor && (
  <Section title="Final Dispositor">
    <p className="text-sm">{interp.final_dispositor}</p>
  </Section>
)}
{interp.antiscia?.length > 0 && (
  <Section title="Antiscia">
    {interp.antiscia.map((s, i) => <div key={i} className="text-sm py-0.5">{s}</div>)}
  </Section>
)}
{interp.antiscia_contacts?.length > 0 && (
  <Section title="Antiscia Contacts">
    {interp.antiscia_contacts.map((s, i) => <div key={i} className="text-sm py-0.5">{s}</div>)}
  </Section>
)}
{interp.declinations?.length > 0 && (
  <Section title="Declination Parallels">
    {interp.declinations.map((s, i) => <div key={i} className="text-sm py-0.5">{s}</div>)}
  </Section>
)}
{interp.contraparallels?.length > 0 && (
  <Section title="Contraparallels">
    {interp.contraparallels.map((s, i) => <div key={i} className="text-sm py-0.5">{s}</div>)}
  </Section>
)}
{interp.key_midpoints?.length > 0 && (
  <Section title="Key Midpoints">
    {interp.key_midpoints.map((s, i) => <div key={i} className="text-sm py-0.5">{s}</div>)}
  </Section>
)}
{interp.key_star_aspects?.length > 0 && (
  <Section title="Star Aspects">
    {interp.key_star_aspects.map((s, i) => <div key={i} className="text-sm py-0.5">{s}</div>)}
  </Section>
)}
{interp.angular_planets?.length > 0 && (
  <Section title="Angular Planets">
    {interp.angular_planets.map((s, i) => <div key={i} className="text-sm py-0.5">{s}</div>)}
  </Section>
)}
```

**Step 2: Add DispositorTree**

Fetch traditional data and render the tree:

```tsx
const [trad, setTrad] = useState<TraditionalResponse | null>(null);
// fetch in useEffect alongside interpretation
api.traditional(data).then(setTrad);

// render after hidden contacts
{trad?.dispositor_tree?.nodes?.length > 0 && (
  <Section title="Dispositor Tree">
    <DispositorTree data={trad.dispositor_tree} />
  </Section>
)}
```

**Step 3: Verify**

```bash
cd web && npx tsc -b && npm run build
```

Expected: no errors, builds.

---

## Task 5: Create `TimeMap` D3 component

**Objective:** Horizontal bar chart showing transit durations (enter → exact → leave) for each transit hit. X-axis = time, Y-axis = transit planet. Color-coded by aspect type.

**Files:**
- Create: `web/src/components/transit/TimeMap.tsx`

**Step 1: Create component**

```typescript
import { useEffect, useRef } from 'react';
import * as d3 from 'd3';
import type { TransitHit } from '../../lib/types';
import { aspectColor } from '../../lib/astrology';

interface Props {
  hits: TransitHit[];
  startDate: string;
  endDate: string;
  onHover?: (hit: TransitHit | null) => void;
}

export function TimeMap({ hits, startDate, endDate, onHover }: Props) {
  const svgRef = useRef<SVGSVGElement>(null);

  useEffect(() => {
    if (!svgRef.current || !hits.length) return;
    // D3 rendering
  }, [hits, startDate, endDate]);

  return (
    <svg ref={svgRef} style={{ width: '100%', height: '100%', minHeight: 300 }} />
  );
}
```

**Step 2: D3 implementation**

- Parse dates with `d3.timeParse("%Y-%m-%d")`
- X scale: `d3.scaleTime()` from startDate to endDate
- Y scale: `d3.scaleBand()` with one row per unique transit planet
- Each hit = horizontal bar from `start_date` to `end_date` (or date ± orb-based estimate if start/end not provided)
- Color bars by aspect type using `aspectColor`
- Add exact hit markers (vertical lines at exact date)
- Add hover tooltips
- Add axis labels

**Step 3: Verify**

```bash
cd web && npx tsc -b && npm run build
```

Expected: no errors, builds.

---

## Task 6: Upgrade `TransitsView` — time map + filtering

**Objective:** Add TimeMap above the hit list table, add filter controls, add sky weather section.

**Files:**
- Modify: `web/src/App.tsx` (TransitsView function)

**Step 1: Add filter state**

```tsx
const [filterPlanet, setFilterPlanet] = useState<string>('');
const [filterAspect, setFilterAspect] = useState<string>('');
const [maxOrb, setMaxOrb] = useState<number>(10);
```

**Step 2: Add filter controls**

Above the table, add a row of filter dropdowns:

```tsx
<div className="flex gap-2 mb-4 flex-wrap">
  <select value={filterPlanet} onChange={e => setFilterPlanet(e.target.value)}
    className="bg-surface border border-border rounded px-2 py-1 text-sm">
    <option value="">All Planets</option>
    {uniquePlanets.map(p => <option key={p} value={p}>{p}</option>)}
  </select>
  <select value={filterAspect} onChange={e => setFilterAspect(e.target.value)}
    className="bg-surface border border-border rounded px-2 py-1 text-sm">
    <option value="">All Aspects</option>
    {uniqueAspects.map(a => <option key={a} value={a}>{a}</option>)}
  </select>
  <label className="text-sm text-muted flex items-center gap-1">
    Max orb: <input type="number" value={maxOrb} onChange={e => setMaxOrb(Number(e.target.value))}
      className="bg-surface border border-border rounded px-2 py-1 w-16 text-sm" />
  </label>
</div>
```

**Step 3: Add TimeMap**

Above the filter controls:

```tsx
<Section title="Time Map">
  <TimeMap hits={filteredHits} startDate={transits.start_date} endDate={transits.end_date} />
</Section>
```

**Step 4: Add Sky Weather section**

After the transit table:

```tsx
{transits.sky_weather?.length > 0 && (
  <Section title="Sky Weather (Transit-to-Transit)">
    <table className="w-full text-sm">
      {/* same table structure as transits */}
    </table>
  </Section>
)}
```

**Step 5: Apply filters**

```tsx
const filteredHits = transits.transits.filter(h => {
  if (filterPlanet && h.transit_planet !== filterPlanet && h.natal_planet !== filterPlanet) return false;
  if (filterAspect && h.aspect !== filterAspect) return false;
  if (h.orb > maxOrb) return false;
  return true;
});
```

**Step 6: Verify**

```bash
cd web && npx tsc -b && npm run build
```

Expected: no errors, builds.

---

## Task 7: Create `AspectGrid` component for synastry

**Objective:** Color-coded matrix where rows = Person A's planets, columns = Person B's planets, cells show aspect symbol and orb.

**Files:**
- Create: `web/src/components/synastry/AspectGrid.tsx`

**Step 1: Create component**

```typescript
import type { Aspect } from '../../lib/types';
import { aspectColor, planetGlyph } from '../../lib/astrology';

interface Props {
  aspects: Aspect[];
  planets1: string[];
  planets2: string[];
}

export function AspectGrid({ aspects, planets1, planets2 }: Props) {
  // Build lookup: "planet1|planet2" → aspect
  const lookup = new Map<string, Aspect>();
  for (const a of aspects) {
    lookup.set(`${a.planet1}|${a.planet2}`, a);
  }

  return (
    <div className="overflow-x-auto">
      <table className="text-sm border-collapse">
        <thead>
          <tr>
            <th className="p-1"></th>
            {planets2.map(p => (
              <th key={p} className="p-1 text-center text-muted" title={p}>
                {planetGlyph(p)}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {planets1.map(p1 => (
            <tr key={p1}>
              <td className="p-1 text-muted" title={p1}>{planetGlyph(p1)}</td>
              {planets2.map(p2 => {
                const a = lookup.get(`${p1}|${p2}`);
                return (
                  <td key={p2} className="p-1 text-center"
                    style={{ backgroundColor: a ? aspectColor(a.aspect) + '33' : 'transparent' }}>
                    {a ? `${a.aspect[0].toUpperCase()}${a.orb.toFixed(1)}°` : ''}
                  </td>
                );
              })}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
```

**Step 2: Verify**

```bash
cd web && npx tsc -b && npm run build
```

Expected: no errors, builds.

---

## Task 8: Create `SynastryView` — full implementation

**Objective:** Side-by-side chart wheels, aspect grid, composite chart.

**Files:**
- Create: `web/src/components/synastry/SynastryView.tsx`
- Modify: `web/src/App.tsx` (replace synastry placeholder)

**Step 1: Create SynastryView component**

The component needs:
- A second chart selector (pick from saved charts or enter birth data)
- Side-by-side ChartWheel components
- AspectGrid
- Composite chart section

```typescript
import { useState, useEffect } from 'react';
import type { SavedChart, BirthData, SynastryResponse, CompositeResponse, Aspect } from '../../lib/types';
import { api } from '../../lib/api';
import { ChartWheel } from '../chart/ChartWheel';
import { AspectGrid } from './AspectGrid';
import { chartDB } from '../../lib/db';

interface Props {
  chartA: SavedChart;
}

const PLANET_ORDER = ['Sun','Moon','Mercury','Venus','Mars','Jupiter','Saturn','Uranus','Neptune','Pluto','Chiron','Ceres','Pallas','Juno','Vesta','Eris','TrueNode'];

export function SynastryView({ chartA }: Props) {
  const [chartB, setChartB] = useState<SavedChart | null>(null);
  const [synastry, setSynastry] = useState<SynastryResponse | null>(null);
  const [composite, setComposite] = useState<CompositeResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [savedCharts, setSavedCharts] = useState<SavedChart[]>([]);

  useEffect(() => {
    chartDB.getAll().then(setSavedCharts);
  }, []);

  const loadSynastry = async (b: BirthData) => {
    setLoading(true);
    setError('');
    try {
      const [s, c] = await Promise.all([
        api.synastry(chartA.birthData, b, 5),
        api.composite(chartA.birthData, b, 3),
      ]);
      setSynastry(s);
      setComposite(c);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed');
    }
    setLoading(false);
  };

  // ... render
}
```

**Step 2: Build the UI**

```
┌─────────────────────────────────────────────┐
│ Select second chart: [dropdown]  or [manual] │
├──────────────────┬──────────────────────────┤
│   Chart A Wheel  │    Chart B Wheel         │
│   (AJ)           │    (Cait)                │
├──────────────────┴──────────────────────────┤
│              Aspect Grid                     │
│  rows=A's planets, cols=B's planets         │
├─────────────────────────────────────────────┤
│           Composite Chart Wheel              │
└─────────────────────────────────────────────┘
```

**Step 3: Wire into App.tsx**

Replace the synastry placeholder:

```tsx
{view === 'synastry' && (
  <div style={{ height: '100%' }} className="overflow-y-auto p-4">
    <SynastryView chartA={activeChart} />
  </div>
)}
```

**Step 4: Verify**

```bash
cd web && npx tsc -b && npm run build
```

Expected: no errors, builds.

---

## Task 9: Extract NatalView and TransitsView to separate files

**Objective:** Move NatalView and TransitsView out of App.tsx into their own component files. App.tsx should only contain layout and tab routing.

**Files:**
- Create: `web/src/components/natal/NatalView.tsx`
- Create: `web/src/components/transit/TransitsView.tsx`
- Modify: `web/src/App.tsx` (remove inline components, add imports)

**Step 1: Move NatalView**

Copy the NatalView function + Section helper from App.tsx to `web/src/components/natal/NatalView.tsx`. Export NatalView as default. Import `api`, types, `DispositorTree`, `PlanetTable`.

**Step 2: Move TransitsView**

Copy TransitsView from App.tsx to `web/src/components/transit/TransitsView.tsx`. Export as default. Import `api`, types, `TimeMap`, `Section`.

**Step 3: Update App.tsx**

Remove the inline NatalView, TransitsView, and Section definitions. Add imports:

```typescript
import { NatalView } from './components/natal/NatalView';
import { TransitsView } from './components/transit/TransitsView';
import { SynastryView } from './components/synastry/SynastryView';
```

**Step 4: Verify**

```bash
cd web && npx tsc -b && npm run build
```

Expected: no errors, builds.

---

## Task 10: End-to-end smoke test

**Objective:** Verify all three views work end-to-end with real data.

**Step 1: Start server**

```bash
cd ~/Documents/repos/empirical
go build -buildvcs=false -o /tmp/empirical ./cmd/recover/
/tmp/empirical serve 5000 &
```

**Step 2: Test Natal view**

```bash
# Verify interpretation endpoint
curl -s http://localhost:5000/api/interpretation -X POST \
  -H 'Content-Type: application/json' \
  -d '{"name":"AJ","year":1969,"month":2,"day":15,"hour":23,"minute":10,"tz_offset":-8,"lat":47.038,"lng":-122.901,"house_system":"P","system":"western","orb":3}' \
  | python3 -c "import sys,json; d=json.load(sys.stdin); print('planet_signs:', len(d.get('planet_signs',[])), 'aspects:', len(d.get('aspects',[])), 'patterns:', len(d.get('patterns',[])))"
```

Expected: planet_signs > 10, aspects > 20, patterns > 0

**Step 3: Test Transits view**

```bash
curl -s http://localhost:5000/api/transits -X POST \
  -H 'Content-Type: application/json' \
  -d '{"name":"AJ","year":1969,"month":2,"day":15,"hour":23,"minute":10,"tz_offset":-8,"lat":47.038,"lng":-122.901,"house_system":"P","start_date":"2026-07-26","end_date":"2026-08-26","orb":3}' \
  | python3 -c "import sys,json; d=json.load(sys.stdin); print('transits:', len(d.get('transits',[])), 'sky_weather:', len(d.get('sky_weather',[])))"
```

Expected: transits > 0, sky_weather > 0

**Step 4: Test Synastry view**

```bash
curl -s http://localhost:5000/api/synastry -X POST \
  -H 'Content-Type: application/json' \
  -d '{"name1":"AJ","year1":1969,"month1":2,"day1":15,"hour1":23,"min1":10,"tz1":-8,"lat1":47.038,"lng1":-122.901,"name2":"Test","year2":1970,"month2":6,"day2":15,"hour2":12,"min2":0,"tz2":-7,"lat2":40.7128,"lng2":-74.006,"orb":5}' \
  | python3 -c "import sys,json; d=json.load(sys.stdin); print('aspects:', len(d.get('aspects',[])))"
```

Expected: aspects > 0

**Step 5: Open browser**

Navigate to http://localhost:5000, create a chart, and verify:
- Natal tab: shows tables with glyphs, dispositor tree, hidden contacts
- Transits tab: shows time map, filterable hit list, sky weather
- Synastry tab: shows chart selector, side-by-side wheels, aspect grid, composite

---

## Summary

| Task | What | Files |
|------|------|-------|
| 1 | Fix dispositor_tree type | `types.ts` |
| 2 | DispositorTree D3 component | `DispositorTree.tsx` (new) |
| 3 | NatalView tables with glyphs | `PlanetTable.tsx` (new), `App.tsx` |
| 4 | Hidden contacts + dispositor tree | `App.tsx` |
| 5 | TimeMap D3 component | `TimeMap.tsx` (new) |
| 6 | TransitsView upgrade | `App.tsx` |
| 7 | AspectGrid component | `AspectGrid.tsx` (new) |
| 8 | SynastryView full implementation | `SynastryView.tsx` (new), `App.tsx` |
| 9 | Extract to separate files | `NatalView.tsx`, `TransitsView.tsx` (new), `App.tsx` |
| 10 | End-to-end smoke test | — |
