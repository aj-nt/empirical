import { useEffect, useState } from 'react';
import type { BirthData, InterpretationResponse, TraditionalResponse } from '../../lib/types';
import { api } from '../../lib/api';
import { PLANET_GLYPHS, SIGN_GLYPHS } from '../../lib/astrology';

interface PlanetInfo {
  planet: string;
  glyph: string;
  sign: string;
  signGlyph: string;
  house: number;
  retrograde: boolean;
}

const PLANET_ORDER = [
  'Sun', 'Moon', 'Mercury', 'Venus', 'Mars', 'Jupiter', 'Saturn',
  'Uranus', 'Neptune', 'Pluto', 'Node', 'Chiron', 'Lilith',
  'Ceres', 'Pallas', 'Juno', 'Vesta', 'Eris', 'Makemake', 'Gonggong',
];

function normalizePlanet(name: string): string {
  if (name === 'TrueNode' || name === 'NorthNode') return 'Node';
  return name;
}

function parsePlanets(interp: InterpretationResponse, trad: TraditionalResponse | null): PlanetInfo[] {
  const retroSet = new Set((trad?.retrogrades ?? []).filter(r => r.retrograde).map(r => r.planet));

  const signMap = new Map<string, string>();
  for (const s of interp.planet_signs ?? []) {
    const m = s.match(/^(\w+)\s+in\s+(\w+):/);
    if (!m) continue;
    const planet = normalizePlanet(m[1]);
    if (!signMap.has(planet) || m[1] === 'Node') {
      signMap.set(planet, m[2]);
    }
  }

  const houseMap = new Map<string, number>();
  for (const s of interp.planet_houses ?? []) {
    const m = s.match(/^(\w+)\s+in\s+(?:the\s+)?(\d+)(?:st|nd|rd|th)?\s+house/) || s.match(/^(\w+)\s+in\s+house\s+(\d+)/);
    if (!m) continue;
    const planet = normalizePlanet(m[1]);
    if (!houseMap.has(planet) || m[1] === 'Node') {
      houseMap.set(planet, parseInt(m[2]));
    }
  }

  const planets: PlanetInfo[] = [];
  const seen = new Set<string>();
  for (const planet of PLANET_ORDER) {
    const normalized = normalizePlanet(planet);
    if (seen.has(normalized)) continue;
    const sign = signMap.get(normalized);
    if (!sign) continue;
    seen.add(normalized);
    planets.push({
      planet: normalized,
      glyph: PLANET_GLYPHS[normalized] ?? normalized.slice(0, 2),
      sign,
      signGlyph: SIGN_GLYPHS[sign] ?? sign,
      house: houseMap.get(normalized) ?? 0,
      retrograde: retroSet.has(normalized),
    });
  }
  return planets;
}

interface WheelSidebarProps {
  data: BirthData;
}

export function WheelSidebar({ data }: WheelSidebarProps) {
  const [planets, setPlanets] = useState<PlanetInfo[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    setLoading(true);
    Promise.all([
      api.interpretation(data, 'western', 3),
      api.traditional(data),
    ])
      .then(([interp, trad]) => {
        setPlanets(parsePlanets(interp, trad));
        setLoading(false);
      })
      .catch(() => setLoading(false));
  }, [data.name, data.year, data.month, data.day, data.hour, data.minute, data.tz_offset, data.lat, data.lng]);

  if (loading) return null;

  return (
    <div className="w-72 shrink-0 border-l border-border bg-surface/50 overflow-y-auto p-4">
      <h3 className="text-sm font-semibold text-muted mb-3">Planets</h3>
      <table className="w-full text-sm">
        <thead>
          <tr className="text-muted text-left border-b border-border/50">
            <th className="py-1.5 pr-2 w-6"></th>
            <th className="py-1.5 pr-2">Planet</th>
            <th className="py-1.5 pr-2">Sign</th>
            <th className="py-1.5 w-10 text-right">H</th>
          </tr>
        </thead>
        <tbody>
          {planets.map((p) => (
            <tr key={p.planet} className="border-b border-border/20 last:border-0 hover:bg-bg/30">
              <td className="py-1.5 pr-2 text-base">{p.glyph}</td>
              <td className="py-1.5 pr-2 text-text">{p.planet}</td>
              <td className="py-1.5 pr-2 text-text">
                <span className="mr-1">{p.signGlyph}</span>
                <span className="text-muted">{p.sign}</span>
              </td>
              <td className="py-1.5 text-muted text-right">{p.house}</td>
              {p.retrograde && <td className="py-1.5 pl-1 text-purple text-xs">℞</td>}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
