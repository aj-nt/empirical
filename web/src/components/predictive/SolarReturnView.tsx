import { useState, useEffect } from 'react';
import type { BirthData, SolarReturnResponse } from '../../lib/types';
import { api } from '../../lib/api';

interface Props {
  data: BirthData;
}

export function SolarReturnView({ data }: Props) {
  const currentYear = new Date().getFullYear();
  const [year, setYear] = useState(currentYear);
  const [result, setResult] = useState<SolarReturnResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    setLoading(true);
    setError('');
    api.solarReturn(data, year)
      .then(setResult)
      .catch(e => setError(e.message))
      .finally(() => setLoading(false));
  }, [data.name, year]);

  return (
    <div className="space-y-3 p-4 overflow-y-auto" style={{ maxHeight: 'calc(100vh - 250px)' }}>
      <div className="flex items-center gap-2">
        <button onClick={() => setYear(y => y - 1)} className="px-2 py-0.5 text-xs bg-surface border border-border rounded">←</button>
        <span className="text-sm font-medium">{year}</span>
        <button onClick={() => setYear(y => y + 1)} className="px-2 py-0.5 text-xs bg-surface border border-border rounded">→</button>
      </div>

      {loading && <p className="text-yellow text-sm">Loading...</p>}
      {error && <p className="text-red text-sm">{error}</p>}

      {result && (
        <>
          <div className="text-sm text-muted">Return: {new Date(result.return_date).toLocaleDateString()}</div>

          <div className="border border-border rounded-lg overflow-hidden">
            <div className="px-3 py-1.5 bg-surface text-sm font-medium">Planets ({result.planets.length})</div>
            <div className="divide-y divide-border">
              {result.planets.map((p, i) => (
                <div key={i} className="px-3 py-1.5 flex justify-between text-sm">
                  <span className="text-text">{p.planet}</span>
                  <span className="text-muted">{p.sign} {p.house ? `H${p.house}` : ''}</span>
                </div>
              ))}
            </div>
          </div>

          <div className="border border-border rounded-lg overflow-hidden">
            <div className="px-3 py-1.5 bg-surface text-sm font-medium">Aspects ({result.aspects.length})</div>
            <div className="divide-y divide-border">
              {result.aspects.map((a, i) => (
                <div key={i} className="px-3 py-1.5 text-sm">
                  <span className="text-text">{a.planet1} {a.aspect} {a.planet2}</span>
                  <span className="text-muted ml-2">{a.orb?.toFixed(1)}°</span>
                </div>
              ))}
            </div>
          </div>

          {result.patterns && result.patterns.length > 0 && (
            <div className="border border-border rounded-lg overflow-hidden">
              <div className="px-3 py-1.5 bg-surface text-sm font-medium">Patterns ({result.patterns.length})</div>
              <div className="divide-y divide-border">
                {result.patterns.map((p, i) => (
                  <div key={i} className="px-3 py-1.5 text-sm text-text">{p.name}: {p.planets?.join(', ')}</div>
                ))}
              </div>
            </div>
          )}
        </>
      )}
    </div>
  );
}
