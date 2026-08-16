import { useState, useEffect } from 'react';
import type { BirthData, ZodiacalReleasingResponse, ZRPeriod } from '../../lib/types';
import { api } from '../../lib/api';

interface Props {
  data: BirthData;
}

function ZRPeriodRow({ period, depth = 0 }: { period: ZRPeriod; depth?: number }) {
  const start = new Date(period.start_date);
  const end = new Date(period.end_date);
  const now = new Date();
  const isCurrent = now >= start && now <= end;

  return (
    <>
      <div className={`px-3 py-1.5 text-sm flex justify-between ${isCurrent ? 'bg-accent/10' : ''}`}
           style={{ paddingLeft: `${12 + depth * 16}px` }}>
        <span className={isCurrent ? 'text-accent font-medium' : 'text-text'}>
          {isCurrent && '▶ '}L{period.level} {period.sign}
        </span>
        <span className="text-muted text-xs">
          {start.toLocaleDateString()} — {end.toLocaleDateString()}
        </span>
      </div>
      {period.sub_periods?.map((sp, i) => (
        <ZRPeriodRow key={i} period={sp} depth={depth + 1} />
      ))}
    </>
  );
}

export function ZodiacalReleasingView({ data }: Props) {
  const [lot, setLot] = useState<'fortune' | 'spirit' | 'eros' | 'necessity'>('fortune');
  const [result, setResult] = useState<ZodiacalReleasingResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    setLoading(true);
    setError('');
    api.zodiacalReleasing(data, lot)
      .then(setResult)
      .catch(e => setError(e.message))
      .finally(() => setLoading(false));
  }, [data.name, lot]);

  return (
    <div className="space-y-3 p-4 overflow-y-auto" style={{ maxHeight: 'calc(100vh - 250px)' }}>
      <div className="flex gap-1">
        {(['fortune', 'spirit', 'eros', 'necessity'] as const).map(l => (
          <button
            key={l}
            onClick={() => setLot(l)}
            className={`px-2 py-0.5 text-xs rounded capitalize ${
              lot === l ? 'bg-accent text-white' : 'bg-surface text-muted border border-border'
            }`}
          >
            {l}
          </button>
        ))}
      </div>

      {loading && <p className="text-yellow text-sm">Loading...</p>}
      {error && <p className="text-red text-sm">{error}</p>}

      {result && (
        <>
          <div className="text-sm text-muted">
            Lot of {result.lot}: {result.lot_sign} {result.lot_degree?.toFixed(1)}°
          </div>

          <div className="border border-border rounded-lg overflow-hidden">
            <div className="px-3 py-1.5 bg-surface text-sm font-medium">L1 Periods</div>
            <div className="divide-y divide-border">
              {result.l1_periods?.map((p, i) => (
                <ZRPeriodRow key={i} period={p} />
              ))}
            </div>
          </div>
        </>
      )}
    </div>
  );
}
