import { useState, useEffect } from 'react';
import type { BirthData, FirdariaResponse, FirdariaPeriod } from '../../lib/types';
import { api } from '../../lib/api';

interface Props {
  data: BirthData;
}

function PeriodRow({ period, depth = 0 }: { period: FirdariaPeriod; depth?: number }) {
  const start = new Date(period.start);
  const end = new Date(period.end);
  const now = new Date();
  const isCurrent = now >= start && now <= end;

  return (
    <div className={`px-3 py-1.5 text-sm flex justify-between ${isCurrent ? 'bg-accent/10' : ''}`}
         style={{ paddingLeft: `${12 + depth * 16}px` }}>
      <span className={isCurrent ? 'text-accent font-medium' : 'text-text'}>
        {isCurrent && '▶ '}{period.planet}
        <span className="text-muted text-xs ml-1">({period.years}y)</span>
      </span>
      <span className="text-muted text-xs">
        {start.toLocaleDateString()} — {end.toLocaleDateString()}
      </span>
    </div>
  );
}

export function FirdariaView({ data }: Props) {
  const [result, setResult] = useState<FirdariaResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    setLoading(true);
    setError('');
    api.firdaria(data)
      .then(setResult)
      .catch(e => setError(e.message))
      .finally(() => setLoading(false));
  }, [data.name]);

  if (loading) return <p className="text-yellow text-sm p-4">Loading...</p>;
  if (error) return <p className="text-red text-sm p-4">{error}</p>;
  if (!result) return <p className="text-muted text-sm p-4">No data.</p>;

  return (
    <div className="space-y-3 p-4 overflow-y-auto" style={{ maxHeight: 'calc(100vh - 250px)' }}>
      <div className="text-sm text-muted">
        {result.diurnal ? '☀️ Diurnal' : '🌙 Nocturnal'} · Order: {result.order?.join(' → ')}
      </div>

      <div className="border border-border rounded-lg overflow-hidden">
        <div className="px-3 py-1.5 bg-surface text-sm font-medium">Major Periods ({result.major_periods?.length || 0})</div>
        <div className="divide-y divide-border">
          {result.major_periods?.map((p, i) => (
            <PeriodRow key={i} period={p} />
          ))}
        </div>
      </div>

      <div className="border border-border rounded-lg overflow-hidden">
        <div className="px-3 py-1.5 bg-surface text-sm font-medium">Sub-Periods ({result.sub_periods?.length || 0})</div>
        <div className="divide-y divide-border">
          {result.sub_periods?.map((p, i) => (
            <PeriodRow key={i} period={p} />
          ))}
        </div>
      </div>
    </div>
  );
}
