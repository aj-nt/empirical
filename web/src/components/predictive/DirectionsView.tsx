import { useState, useEffect } from 'react';
import type { BirthData, DirectionsResponse, SolarArcResponse } from '../../lib/types';
import { api } from '../../lib/api';

interface Props {
  data: BirthData;
}

export function DirectionsView({ data }: Props) {
  const today = new Date().toISOString().split('T')[0];
  const [age, setAge] = useState(() => {
    const birth = new Date(data.year, data.month - 1, data.day);
    return Math.floor((Date.now() - birth.getTime()) / (365.25 * 24 * 60 * 60 * 1000));
  });
  const [dirResult, setDirResult] = useState<DirectionsResponse | null>(null);
  const [arcResult, setArcResult] = useState<SolarArcResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    setLoading(true);
    setError('');
    Promise.all([
      api.directions(data, age),
      api.solarArc(data, today),
    ])
      .then(([d, a]) => { setDirResult(d); setArcResult(a); })
      .catch(e => setError(e.message))
      .finally(() => setLoading(false));
  }, [data.name, age]);

  return (
    <div className="space-y-3 p-4 overflow-y-auto" style={{ maxHeight: 'calc(100vh - 250px)' }}>
      <div className="flex items-center gap-2">
        <span className="text-sm text-muted">Age:</span>
        <input
          type="range"
          min={0}
          max={100}
          value={age}
          onChange={e => setAge(Number(e.target.value))}
          className="flex-1"
        />
        <span className="text-sm font-medium">{age}</span>
      </div>

      {loading && <p className="text-yellow text-sm">Loading...</p>}
      {error && <p className="text-red text-sm">{error}</p>}

      {dirResult && (
        <div className="border border-border rounded-lg overflow-hidden">
          <div className="px-3 py-1.5 bg-surface text-sm font-medium">Primary Directions</div>
          <div className="px-3 py-1.5 text-xs text-muted">
            Directed ASC: {dirResult.directed_asc?.toFixed(1)}° | Directed MC: {dirResult.directed_mc?.toFixed(1)}°
          </div>
          <div className="divide-y divide-border">
            <div className="px-3 py-1 text-xs text-muted">ASC aspects ({dirResult.asc_aspects?.length || 0})</div>
            {dirResult.asc_aspects?.map((a, i) => (
              <div key={i} className="px-3 py-1.5 text-sm">
                <span className="text-text">ASC {a.aspect} {a.planet2}</span>
                <span className="text-muted ml-2">{a.orb?.toFixed(1)}°</span>
              </div>
            ))}
            <div className="px-3 py-1 text-xs text-muted">MC aspects ({dirResult.mc_aspects?.length || 0})</div>
            {dirResult.mc_aspects?.map((a, i) => (
              <div key={i} className="px-3 py-1.5 text-sm">
                <span className="text-text">MC {a.aspect} {a.planet2}</span>
                <span className="text-muted ml-2">{a.orb?.toFixed(1)}°</span>
              </div>
            ))}
          </div>
        </div>
      )}

      {arcResult && (
        <div className="border border-border rounded-lg overflow-hidden">
          <div className="px-3 py-1.5 bg-surface text-sm font-medium">Solar Arc ({arcResult.solar_arc_deg?.toFixed(2)}°)</div>
          <div className="divide-y divide-border">
            {arcResult.aspects?.map((a, i) => (
              <div key={i} className="px-3 py-1.5 text-sm">
                <span className="text-text">{a.planet1} {a.aspect} {a.planet2}</span>
                <span className="text-muted ml-2">{a.orb?.toFixed(1)}°</span>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
