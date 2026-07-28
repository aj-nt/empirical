import { useState, useEffect, useCallback } from 'react';
import type { BirthData, ResearchMetrics, ResearchBaseline } from '../../lib/types';
import { api } from '../../lib/api';

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div className="bg-surface border border-border rounded-lg p-4">
      <h3 className="text-sm font-semibold text-muted mb-2">{title}</h3>
      {children}
    </div>
  );
}

const METRICS: { key: string; label: string; unit: string; higherIsRarer: boolean }[] = [
  { key: 'aspect_pattern_count', label: 'Aspect Patterns', unit: 'patterns', higherIsRarer: true },
  { key: 'draconic_bridge_count', label: 'Draconic Bridges', unit: 'bridges', higherIsRarer: true },
  { key: 'paran_count', label: 'Parans', unit: 'contacts', higherIsRarer: true },
  { key: 'declination_parallel_count', label: 'Declination Parallels', unit: 'pairs', higherIsRarer: true },
  { key: 'mansion_convergence_count', label: 'Mansion Convergence', unit: 'planets', higherIsRarer: true },
  { key: 'cross_system_sign_agreement', label: 'Cross-System Sign Agreement', unit: '%', higherIsRarer: false },
  { key: 'arabic_parts_survivor_pct', label: 'Arabic Parts Survivors', unit: '%', higherIsRarer: false },
  { key: 'stars_cross_survivor_pct', label: 'Stars Cross Survivors', unit: '%', higherIsRarer: false },
  { key: 'harmonic_4', label: 'Harmonic 4 Conjunctions', unit: 'pairs', higherIsRarer: true },
  { key: 'harmonic_5', label: 'Harmonic 5 Conjunctions', unit: 'pairs', higherIsRarer: true },
  { key: 'harmonic_7', label: 'Harmonic 7 Conjunctions', unit: 'pairs', higherIsRarer: true },
  { key: 'harmonic_9', label: 'Harmonic 9 Conjunctions', unit: 'pairs', higherIsRarer: true },
];

function extractValue(metrics: ResearchMetrics, key: string): number {
  if (key.startsWith('harmonic_')) {
    const h = parseInt(key.split('_')[1], 10);
    return metrics.harmonic_conjunctions?.[String(h)] ?? 0;
  }
  switch (key) {
    case 'cross_system_sign_agreement': return metrics.cross_system_sign_agreement;
    case 'draconic_bridge_count': return metrics.draconic_bridge_count;
    case 'paran_count': return metrics.paran_count;
    case 'declination_parallel_count': return metrics.declination_parallel_count;
    case 'arabic_parts_survivor_pct': return metrics.arabic_parts_survivor_pct;
    case 'mansion_convergence_count': return metrics.mansion_convergence_count;
    case 'stars_cross_survivor_pct': return metrics.stars_cross_survivor_pct;
    case 'aspect_pattern_count': return metrics.aspect_pattern_count;
    default: return 0;
  }
}

function percentileRank(value: number, baseline: ResearchBaseline): number {
  // Approximate: where does value fall in the distribution?
  // We don't have the raw values, so we interpolate between known percentiles.
  const percentiles = [
    { p: 5, v: baseline.p5 },
    { p: 25, v: baseline.p25 },
    { p: 50, v: baseline.p50 },
    { p: 75, v: baseline.p75 },
    { p: 95, v: baseline.p95 },
  ];

  if (value <= percentiles[0].v) return 5;
  if (value >= percentiles[4].v) return 95;

  for (let i = 0; i < percentiles.length - 1; i++) {
    const lo = percentiles[i];
    const hi = percentiles[i + 1];
    if (value >= lo.v && value <= hi.v) {
      const frac = (value - lo.v) / (hi.v - lo.v);
      return lo.p + frac * (hi.p - lo.p);
    }
  }
  return 50;
}

function rankColor(pct: number, higherIsRarer: boolean): string {
  const rarity = higherIsRarer ? pct : 100 - pct;
  if (rarity >= 90) return 'text-purple';
  if (rarity >= 75) return 'text-accent';
  if (rarity >= 50) return 'text-green';
  return 'text-muted';
}

export function EmpiricalBaselines({ data }: { data: BirthData }) {
  const [metrics, setMetrics] = useState<ResearchMetrics | null>(null);
  const [baselines, setBaselines] = useState<Record<string, ResearchBaseline>>({});
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [baselineN, setBaselineN] = useState(100);

  const fetchAll = useCallback(async () => {
    setLoading(true);
    setError('');
    try {
      // Fetch chart metrics
      const m = await api.researchMetrics(data);
      setMetrics(m);

      // Fetch baselines in parallel (small N for speed)
      const baselineKeys = [
        'aspect_pattern_count', 'draconic_bridge_count', 'paran_count',
        'declination_parallel_count', 'mansion_convergence_count',
        'cross_system_sign_agreement', 'arabic_parts_survivor_pct',
        'stars_cross_survivor_pct', 'harmonic_4', 'harmonic_5',
        'harmonic_7', 'harmonic_9',
      ];

      const results = await Promise.all(
        baselineKeys.map((key) =>
          api.researchBaseline(key, baselineN, 42).catch(() => null)
        )
      );

      const b: Record<string, ResearchBaseline> = {};
      baselineKeys.forEach((key, i) => {
        if (results[i]) b[key] = results[i]!;
      });
      setBaselines(b);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to load research data');
    } finally {
      setLoading(false);
    }
  }, [data.name, data.year, data.month, data.day, data.hour, data.minute, data.tz_offset, data.lat, data.lng, baselineN]);

  useEffect(() => {
    fetchAll();
  }, [fetchAll]);

  if (loading) {
    return (
      <div className="space-y-2">
        <p className="text-yellow text-sm">Computing research metrics and baselines…</p>
        <p className="text-muted text-xs">Generating {baselineN} random charts per metric. This may take a moment.</p>
      </div>
    );
  }

  if (error) return <p className="text-red text-sm">{error}</p>;
  if (!metrics) return null;

  return (
    <div className="space-y-4">
      <Section title="Empirical Baselines">
        <p className="text-sm text-muted mb-4">
          Each metric is compared against {baselineN} random charts to determine how unusual this chart is.
          Percentile rank shows where this chart falls — 95th percentile means only 5% of random charts score higher.
        </p>

        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="text-muted text-left border-b border-border">
                <th className="py-2 pr-4">Metric</th>
                <th className="py-2 pr-4 text-right">Your Chart</th>
                <th className="py-2 pr-4 text-right">Baseline Mean</th>
                <th className="py-2 pr-4 text-right">Median</th>
                <th className="py-2 pr-4 text-right">Std Dev</th>
                <th className="py-2 pr-4 text-right">P5–P95</th>
                <th className="py-2 text-right">Percentile</th>
              </tr>
            </thead>
            <tbody>
              {METRICS.map(({ key, label, unit, higherIsRarer }) => {
                const value = extractValue(metrics, key);
                const bl = baselines[key];
                if (!bl) return null;

                const pct = percentileRank(value, bl);
                const colorClass = rankColor(pct, higherIsRarer);

                return (
                  <tr key={key} className="border-t border-border hover:bg-surface/50">
                    <td className="py-2 pr-4">{label}</td>
                    <td className="py-2 pr-4 text-right font-mono">
                      {typeof value === 'number' && value % 1 !== 0
                        ? value.toFixed(1)
                        : value}
                      <span className="text-muted ml-1">{unit}</span>
                    </td>
                    <td className="py-2 pr-4 text-right text-muted font-mono">
                      {bl.mean % 1 !== 0 ? bl.mean.toFixed(1) : bl.mean}
                    </td>
                    <td className="py-2 pr-4 text-right text-muted font-mono">
                      {bl.median % 1 !== 0 ? bl.median.toFixed(1) : bl.median}
                    </td>
                    <td className="py-2 pr-4 text-right text-muted font-mono">
                      {bl.std_dev % 1 !== 0 ? bl.std_dev.toFixed(1) : bl.std_dev}
                    </td>
                    <td className="py-2 pr-4 text-right text-muted font-mono text-xs">
                      {bl.p5 % 1 !== 0 ? bl.p5.toFixed(1) : bl.p5}–{bl.p95 % 1 !== 0 ? bl.p95.toFixed(1) : bl.p95}
                    </td>
                    <td className={`py-2 text-right font-semibold font-mono ${colorClass}`}>
                      {pct.toFixed(0)}%
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      </Section>

      {/* Summary */}
      <Section title="What Stands Out">
        <div className="space-y-2">
          {METRICS.map(({ key, label, higherIsRarer }) => {
            const value = extractValue(metrics, key);
            const bl = baselines[key];
            if (!bl) return null;

            const pct = percentileRank(value, bl);
            const rarity = higherIsRarer ? pct : 100 - pct;

            if (rarity >= 90) {
              return (
                <div key={key} className="text-sm">
                  <span className="text-purple font-semibold">{label}</span>
                  <span className="text-muted"> — in the </span>
                  <span className="text-purple font-semibold">{rarity.toFixed(0)}th percentile</span>
                  <span className="text-muted">
                    {' '}(only {rarity >= 95 ? (100 - rarity).toFixed(0) : (100 - rarity).toFixed(0)}% of random charts score {
                      higherIsRarer ? 'higher' : 'lower'
                    })
                  </span>
                </div>
              );
            }
            if (rarity >= 75) {
              return (
                <div key={key} className="text-sm">
                  <span className="text-accent font-semibold">{label}</span>
                  <span className="text-muted"> — notably high at the </span>
                  <span className="text-accent font-semibold">{rarity.toFixed(0)}th percentile</span>
                </div>
              );
            }
            return null;
          })}
          {METRICS.every(({ key, higherIsRarer }) => {
            const value = extractValue(metrics, key);
            const bl = baselines[key];
            if (!bl) return true;
            const pct = percentileRank(value, bl);
            const rarity = higherIsRarer ? pct : 100 - pct;
            return rarity < 75;
          }) && (
            <p className="text-sm text-muted">
              No metrics stand out significantly. This chart is within normal ranges across all measured dimensions.
            </p>
          )}
        </div>
      </Section>

      {/* Refresh with larger N */}
      <div className="flex gap-2 items-center">
        <span className="text-xs text-muted">Baseline sample size:</span>
        {[100, 500, 1000].map((n) => (
          <button
            key={n}
            onClick={() => setBaselineN(n)}
            className={`px-2 py-0.5 text-xs rounded ${
              baselineN === n
                ? 'bg-accent text-white'
                : 'bg-surface text-muted border border-border hover:text-text'
            }`}
          >
            {n}
          </button>
        ))}
        <span className="text-xs text-muted ml-auto">
          Larger samples are more accurate but slower.
        </span>
      </div>
    </div>
  );
}
