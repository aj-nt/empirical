import type { InterpretationResponse } from '../../lib/types';
import { PLANET_GLYPHS } from '../../lib/astrology';

interface AspectEntry {
  planet1: string;
  planet2: string;
  aspect: string;
  orb: number;
}

const ASPECT_COLORS: Record<string, string> = {
  conjunction: '#f0c040',
  opposition: '#a371f7',
  trine: '#58a6ff',
  square: '#f85149',
  sextile: '#3fb950',
  quincunx: '#8b949e',
  'semi-sextile': '#8b949e',
  'semi-square': '#8b949e',
  sesquiquadrate: '#8b949e',
};

const ASPECT_ABBREV: Record<string, string> = {
  conjunction: '☌',
  opposition: '☍',
  trine: '△',
  square: '□',
  sextile: '⚹',
  quincunx: '⚻',
  'semi-sextile': '⚺',
  'semi-square': '∠',
  sesquiquadrate: '⚼',
};

const PLANET_ORDER = [
  'Sun', 'Moon', 'Mercury', 'Venus', 'Mars', 'Jupiter', 'Saturn',
  'Uranus', 'Neptune', 'Pluto', 'Node', 'Chiron', 'Lilith',
  'Ceres', 'Pallas', 'Juno', 'Vesta', 'Eris', 'Makemake', 'Gonggong',
];

function normalizePlanet(name: string): string {
  if (name === 'TrueNode' || name === 'NorthNode') return 'Node';
  return name;
}

function parseAspects(interp: InterpretationResponse): AspectEntry[] {
  const entries: AspectEntry[] = [];
  for (const s of interp.aspects ?? []) {
    // "Sun square Neptune (orb 1.2°): ..."
    const m = s.match(/^(\w+)\s+(conjunction|opposition|trine|square|sextile|quincunx|semi-sextile|semi-square|sesquiquadrate)\s+(\w+)\s+\(orb\s+([\d.]+)°\)/);
    if (!m) continue;
    entries.push({
      planet1: normalizePlanet(m[1]),
      planet2: normalizePlanet(m[3]),
      aspect: m[2],
      orb: parseFloat(m[4]),
    });
  }
  return entries;
}

interface AspectGridProps {
  interp: InterpretationResponse;
}

export function AspectGrid({ interp }: AspectGridProps) {
  const aspects = parseAspects(interp);

  // Build planet list from aspects
  const planetSet = new Set<string>();
  for (const a of aspects) {
    planetSet.add(a.planet1);
    planetSet.add(a.planet2);
  }
  const planets = PLANET_ORDER.filter(p => planetSet.has(p));

  // Build lookup: "planet1|planet2" -> aspect
  const lookup = new Map<string, AspectEntry>();
  for (const a of aspects) {
    const key = [a.planet1, a.planet2].sort().join('|');
    // Keep the tighter orb
    const existing = lookup.get(key);
    if (!existing || a.orb < existing.orb) {
      lookup.set(key, a);
    }
  }

  if (planets.length === 0) return null;

  return (
    <div className="bg-surface border border-border rounded-lg overflow-hidden">
      <h3 className="text-sm font-semibold text-muted px-4 pt-3 pb-2">Aspects</h3>
      <div className="overflow-x-auto">
        <table className="text-xs" style={{ borderCollapse: 'collapse' }}>
          <thead>
            <tr>
              <th className="w-8"></th>
              {planets.map(p => (
                <th key={p} className="w-8 text-center text-muted font-normal py-1" title={p}>
                  {PLANET_GLYPHS[p] ?? p.slice(0, 2)}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {planets.map((p1, i) => (
              <tr key={p1}>
                <td className="text-center text-muted py-0.5" title={p1}>
                  {PLANET_GLYPHS[p1] ?? p1.slice(0, 2)}
                </td>
                {planets.map((p2, j) => {
                  if (j >= i) {
                    return <td key={p2} className="w-8" />;
                  }
                  const key = [p1, p2].sort().join('|');
                  const entry = lookup.get(key);
                  if (!entry) {
                    return <td key={p2} className="w-8 text-center text-muted/30">—</td>;
                  }
                  const color = ASPECT_COLORS[entry.aspect] ?? '#8b949e';
                  const abbrev = ASPECT_ABBREV[entry.aspect] ?? entry.aspect.slice(0, 2);
                  return (
                    <td
                      key={p2}
                      className="w-8 text-center font-bold py-0.5"
                      style={{ color }}
                      title={`${entry.planet1} ${entry.aspect} ${entry.planet2} (${entry.orb}°)`}
                    >
                      {abbrev}
                    </td>
                  );
                })}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
