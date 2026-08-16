import { useState, useEffect } from 'react';
import type { BirthData, TransitResponse, TransitHit } from '../../lib/types';
import { api } from '../../lib/api';

interface Props {
  data: BirthData;
}

const PLANET_ORDER = ['Sun', 'Moon', 'Mercury', 'Venus', 'Mars', 'Jupiter', 'Saturn', 'Uranus', 'Neptune', 'Pluto', 'Node', 'Chiron', 'Lilith', 'Ceres', 'Pallas', 'Juno', 'Vesta', 'Eris', 'Makemake', 'Gonggong'];

function planetSort(a: string, b: string): number {
  const ai = PLANET_ORDER.indexOf(a);
  const bi = PLANET_ORDER.indexOf(b);
  if (ai >= 0 && bi >= 0) return ai - bi;
  if (ai >= 0) return -1;
  if (bi >= 0) return 1;
  return a.localeCompare(b);
}

export function TransitInterpretation({ data }: Props) {
  const [transits, setTransits] = useState<TransitResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [filter, setFilter] = useState<'all' | 'personal' | 'outer'>('personal');
  const [showSkyWeather, setShowSkyWeather] = useState(false);

  useEffect(() => {
    const now = new Date();
    const start = now.toISOString().slice(0, 10);
    const end = new Date(now.getTime() + 30 * 86400000).toISOString().slice(0, 10);

    setLoading(true);
    setError('');
    api.transits(data, start, end, 3)
      .then(setTransits)
      .catch(e => setError(e.message))
      .finally(() => setLoading(false));
  }, [data.name, data.year, data.month, data.day, data.hour, data.minute, data.tz_offset, data.lat, data.lng]);

  if (loading) return <p className="text-yellow text-sm p-4">Loading transits...</p>;
  if (error) return <p className="text-red text-sm p-4">{error}</p>;
  if (!transits) return <p className="text-muted text-sm p-4">No transit data available.</p>;

  const personalPlanets = new Set(['Sun', 'Moon', 'Mercury', 'Venus', 'Mars']);
  const outerPlanets = new Set(['Jupiter', 'Saturn', 'Uranus', 'Neptune', 'Pluto']);

  const filteredTransits = transits.transits.filter(t => {
    if (filter === 'personal') return personalPlanets.has(t.transit_planet);
    if (filter === 'outer') return outerPlanets.has(t.transit_planet);
    return true;
  });

  const filteredSkyWeather = transits.sky_weather.filter(t => {
    if (filter === 'personal') return personalPlanets.has(t.transit_planet);
    if (filter === 'outer') return outerPlanets.has(t.transit_planet);
    return true;
  });

  // Group transits by transit planet
  const grouped = new Map<string, TransitHit[]>();
  for (const t of filteredTransits) {
    const key = t.transit_planet;
    if (!grouped.has(key)) grouped.set(key, []);
    grouped.get(key)!.push(t);
  }

  const sortedPlanets = Array.from(grouped.keys()).sort(planetSort);

  return (
    <div className="space-y-3 p-4 overflow-y-auto" style={{ maxHeight: 'calc(100vh - 250px)' }}>
      <div className="flex items-center gap-2 mb-2">
        <span className="text-xs text-muted">Show:</span>
        {(['personal', 'outer', 'all'] as const).map(f => (
          <button
            key={f}
            onClick={() => setFilter(f)}
            className={`px-2 py-0.5 text-xs rounded ${
              filter === f ? 'bg-accent text-white' : 'bg-surface text-muted border border-border'
            }`}
          >
            {f === 'personal' ? 'Personal' : f === 'outer' ? 'Outer' : 'All'}
          </button>
        ))}
        <span className="text-xs text-muted ml-4">
          {transits.start_date} → {transits.end_date}
        </span>
      </div>

      {/* Transit-to-natal aspects */}
      <div className="space-y-2">
        <h3 className="text-sm font-semibold text-text">
          Transit-to-Natal Aspects ({filteredTransits.length})
        </h3>
        {sortedPlanets.map(planet => {
          const hits = grouped.get(planet)!;
          return (
            <div key={planet} className="border border-border rounded-lg overflow-hidden">
              <div className="px-3 py-1.5 bg-surface text-sm font-medium">
                {planet} ({hits.length})
              </div>
              <div className="px-3 py-1.5 space-y-0.5">
                {hits.sort((a, b) => a.start_date?.localeCompare(b.start_date || '') || 0).map((hit, i) => (
                  <div key={i} className="text-xs flex justify-between">
                    <span>
                      {hit.aspect} <span className="text-muted">{hit.natal_planet}</span>
                      {' '}orb {hit.orb.toFixed(1)}°
                    </span>
                    <span className="text-muted">
                      {hit.start_date}{hit.end_date && hit.end_date !== hit.start_date ? ` → ${hit.end_date}` : ''}
                    </span>
                  </div>
                ))}
              </div>
            </div>
          );
        })}
      </div>

      {/* Sky Weather (transit-to-transit) */}
      <div>
        <button
          onClick={() => setShowSkyWeather(!showSkyWeather)}
          className="text-sm font-semibold text-text hover:text-accent"
        >
          {showSkyWeather ? '▾' : '▸'} Sky Weather ({filteredSkyWeather.length})
        </button>
        {showSkyWeather && (
          <div className="mt-2 space-y-1">
            {filteredSkyWeather.map((hit, i) => (
              <div key={i} className="text-xs flex justify-between">
                <span>
                  {hit.transit_planet} {hit.aspect} {hit.natal_planet}
                  {' '}orb {hit.orb.toFixed(1)}°
                </span>
                <span className="text-muted">
                  {hit.start_date}{hit.end_date && hit.end_date !== hit.start_date ? ` → ${hit.end_date}` : ''}
                </span>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
