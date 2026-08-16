import type { InterpretationResponse, TraditionalResponse } from '../../lib/types';
import { PLANET_GLYPHS, SIGN_GLYPHS } from '../../lib/astrology';

interface PlanetRow {
  planet: string;
  glyph: string;
  sign: string;
  signGlyph: string;
  house: number;
  dignity: string;
  retrograde: boolean;
}

const PLANET_ORDER = [
  'Sun', 'Moon', 'Mercury', 'Venus', 'Mars', 'Jupiter', 'Saturn',
  'Uranus', 'Neptune', 'Pluto', 'Node', 'TrueNode', 'NorthNode', 'SouthNode',
  'Chiron', 'Lilith', 'Ceres', 'Pallas', 'Juno', 'Vesta',
  'Eris', 'Makemake', 'Gonggong',
];

function normalizePlanet(name: string): string {
  if (name === 'TrueNode' || name === 'NorthNode') return 'Node';
  return name;
}

function parsePlanetData(interp: InterpretationResponse, trad: TraditionalResponse | null): PlanetRow[] {
  const retroSet = new Set((trad?.retrogrades ?? []).filter(r => r.retrograde).map(r => r.planet));

  // Parse planet_signs: "Sun in Aquarius: ... fixed air — ... in detriment — ..."
  const signMap = new Map<string, { sign: string; dignity: string }>();
  for (const s of interp.planet_signs ?? []) {
    const m = s.match(/^(\w+)\s+in\s+(\w+):/);
    if (!m) continue;
    const planet = normalizePlanet(m[1]);
    const sign = m[2];

    let dignity = 'neutral';
    const dm = s.match(/\.\s*(in\s+(?:detriment|fall|domicile|exaltation)|neutral|peregrine)\s*[—–-]/);
    if (dm) dignity = dm[1];

    // Don't overwrite Node with TrueNode/NorthNode
    if (!signMap.has(planet) || m[1] === 'Node') {
      signMap.set(planet, { sign, dignity });
    }
  }

  // Parse planet_houses: "Sun in the 4th house: ..."
  const houseMap = new Map<string, number>();
  for (const s of interp.planet_houses ?? []) {
    const m = s.match(/^(\w+)\s+in\s+(?:the\s+)?(\d+)(?:st|nd|rd|th)?\s+house/) || s.match(/^(\w+)\s+in\s+house\s+(\d+)/);
    if (!m) continue;
    const planet = normalizePlanet(m[1]);
    if (!houseMap.has(planet) || m[1] === 'Node') {
      houseMap.set(planet, parseInt(m[2]));
    }
  }

  // Build rows in standard order
  const rows: PlanetRow[] = [];
  const seen = new Set<string>();
  for (const planet of PLANET_ORDER) {
    const normalized = normalizePlanet(planet);
    if (seen.has(normalized)) continue;
    const info = signMap.get(normalized);
    if (!info) continue;
    seen.add(normalized);
    rows.push({
      planet: normalized,
      glyph: PLANET_GLYPHS[normalized] ?? normalized.slice(0, 2),
      sign: info.sign,
      signGlyph: SIGN_GLYPHS[info.sign] ?? info.sign,
      house: houseMap.get(normalized) ?? 0,
      dignity: info.dignity,
      retrograde: retroSet.has(normalized),
    });
  }
  return rows;
}

const DIGNITY_COLORS: Record<string, string> = {
  'in domicile': '#3fb950',
  'in exaltation': '#3fb950',
  'in detriment': '#f85149',
  'in fall': '#f85149',
  'neutral': '#8b949e',
  'peregrine': '#8b949e',
};

interface PlanetTableProps {
  interp: InterpretationResponse;
  trad: TraditionalResponse | null;
}

export function PlanetTable({ interp, trad }: PlanetTableProps) {
  const rows = parsePlanetData(interp, trad);

  return (
    <div className="bg-surface border border-border rounded-lg overflow-hidden">
      <h3 className="text-sm font-semibold text-muted px-4 pt-3 pb-2">Planets</h3>
      <table className="w-full text-sm">
        <thead>
          <tr className="text-muted text-left border-b border-border">
            <th className="py-1.5 pl-4 pr-2 w-8"></th>
            <th className="py-1.5 pr-3">Planet</th>
            <th className="py-1.5 pr-3">Sign</th>
            <th className="py-1.5 pr-3 w-12">House</th>
            <th className="py-1.5 pr-4">Dignity</th>
          </tr>
        </thead>
        <tbody>
          {rows.map((r, i) => (
            <tr key={r.planet} className={i % 2 === 0 ? 'bg-bg/30' : ''}>
              <td className="py-1 pl-4 pr-2 text-base">{r.glyph}</td>
              <td className="py-1 pr-3 text-text">{r.planet}</td>
              <td className="py-1 pr-3 text-text">
                {r.signGlyph} {r.sign}
              </td>
              <td className="py-1 pr-3 text-text">{r.house || '—'}</td>
              <td className="py-1 pr-4" style={{ color: DIGNITY_COLORS[r.dignity] ?? '#8b949e' }}>
                {r.dignity}
                {r.retrograde && <span className="text-purple ml-1"> ℞</span>}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
