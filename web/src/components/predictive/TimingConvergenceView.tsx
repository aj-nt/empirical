import { useState, useEffect } from 'react';
import type { BirthData, TimingConvergenceResponse } from '../../lib/types';
import { api } from '../../lib/api';

interface Props {
  data: BirthData;
}

export function TimingConvergenceView({ data }: Props) {
  const today = new Date().toISOString().split('T')[0];
  const oneYear = new Date(Date.now() + 365 * 24 * 60 * 60 * 1000).toISOString().split('T')[0];
  const [startDate, setStartDate] = useState(today);
  const [endDate, setEndDate] = useState(oneYear);
  const [result, setResult] = useState<TimingConvergenceResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    setLoading(true);
    setError('');
    api.timingConvergence(data, startDate, endDate)
      .then(setResult)
      .catch(e => setError(e.message))
      .finally(() => setLoading(false));
  }, [data.name, startDate, endDate]);

  return (
    <div className="space-y-3 p-4 overflow-y-auto" style={{ maxHeight: 'calc(100vh - 250px)' }}>
      <div className="flex items-center gap-2">
        <input type="date" value={startDate} onChange={e => setStartDate(e.target.value)}
          className="bg-bg border border-border rounded px-2 py-0.5 text-sm text-text" />
        <span className="text-muted text-xs">to</span>
        <input type="date" value={endDate} onChange={e => setEndDate(e.target.value)}
          className="bg-bg border border-border rounded px-2 py-0.5 text-sm text-text" />
      </div>

      {loading && <p className="text-yellow text-sm">Loading...</p>}
      {error && <p className="text-red text-sm">{error}</p>}

      {result && result.timing_convergence && (
        <div className="border border-border rounded-lg overflow-hidden">
          <div className="px-3 py-1.5 bg-surface text-sm font-medium">
            Timing Convergence ({Object.keys(result.timing_convergence).length} dates)
          </div>
          <div className="divide-y divide-border">
            {Object.entries(result.timing_convergence).map(([date, techniques]) => (
              <div key={date} className="px-3 py-1.5 text-sm">
                <span className="text-accent font-medium">{date}</span>
                <span className="text-muted ml-2">{Array.isArray(techniques) ? techniques.join(', ') : techniques}</span>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
