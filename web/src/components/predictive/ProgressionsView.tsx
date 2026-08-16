import { useState, useEffect } from 'react';
import type { BirthData, ProgressedResponse } from '../../lib/types';
import { api } from '../../lib/api';

interface Props {
  data: BirthData;
}

export function ProgressionsView({ data }: Props) {
  const today = new Date().toISOString().split('T')[0];
  const [targetDate, setTargetDate] = useState(today);
  const [result, setResult] = useState<ProgressedResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    setLoading(true);
    setError('');
    api.progressed(data, targetDate)
      .then(setResult)
      .catch(e => setError(e.message))
      .finally(() => setLoading(false));
  }, [data.name, targetDate]);

  return (
    <div className="space-y-3 p-4 overflow-y-auto" style={{ maxHeight: 'calc(100vh - 250px)' }}>
      <div className="flex items-center gap-2">
        <input
          type="date"
          value={targetDate}
          onChange={e => setTargetDate(e.target.value)}
          className="bg-bg border border-border rounded px-2 py-0.5 text-sm text-text"
        />
      </div>

      {loading && <p className="text-yellow text-sm">Loading...</p>}
      {error && <p className="text-red text-sm">{error}</p>}

      {result && (
        <>
          <div className="border border-border rounded-lg overflow-hidden">
            <div className="px-3 py-1.5 bg-surface text-sm font-medium">Progressed Planets ({result.planets.length})</div>
            <div className="divide-y divide-border">
              {result.planets.map((p, i) => (
                <div key={i} className="px-3 py-1.5 flex justify-between text-sm">
                  <span className="text-text">{p.planet}</span>
                  <span className="text-muted">{p.sign} {p.house ? `H${p.house}` : ''} {p.lon?.toFixed(1)}°</span>
                </div>
              ))}
            </div>
          </div>

          <div className="border border-border rounded-lg overflow-hidden">
            <div className="px-3 py-1.5 bg-surface text-sm font-medium">Progressed-to-Natal Aspects ({result.aspects.length})</div>
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
