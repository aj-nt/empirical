import { useState, useEffect } from 'react';
import type { BirthData, ProfectionResponse } from '../../lib/types';
import { api } from '../../lib/api';

interface Props {
  data: BirthData;
}

export function ProfectionView({ data }: Props) {
  const today = new Date().toISOString().split('T')[0];
  const [targetDate, setTargetDate] = useState(today);
  const [result, setResult] = useState<ProfectionResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    setLoading(true);
    setError('');
    api.profection(data, targetDate)
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
        <div className="space-y-3">
          <div className="border border-border rounded-lg p-3 bg-surface/50">
            <div className="text-sm text-muted">Year {result.profection_year} — Age {result.age_years?.toFixed(1)}</div>
            <div className="text-lg font-medium text-accent">House {result.profected_house} ({result.profected_sign})</div>
            <div className="text-xs text-muted">ASC: {result.natal_asc?.toFixed(1)}° → {result.profected_asc?.toFixed(1)}°</div>
          </div>

          <div className="border border-border rounded-lg p-3 bg-surface/50">
            <div className="text-sm text-muted">Time Lord</div>
            <div className="text-lg font-medium text-accent">{result.time_lord}</div>
            <div className="text-sm text-muted">in {result.time_lord_sign} (House {result.time_lord_house})</div>
          </div>

          {result.planets_in_sign && result.planets_in_sign.length > 0 && (
            <div className="border border-border rounded-lg overflow-hidden">
              <div className="px-3 py-1.5 bg-surface text-sm font-medium">Planets in Profected Sign ({result.planets_in_sign.length})</div>
              <div className="px-3 py-1.5 text-sm text-text">
                {result.planets_in_sign.join(', ')}
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
