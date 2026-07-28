import { useState, useEffect, useCallback } from 'react';
import type { SavedChart, BirthData, BatchAnalysisResponse } from '../../lib/types';
import { api } from '../../lib/api';
import { chartDB } from '../../lib/db';

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div className="bg-surface border border-border rounded-lg p-4">
      <h3 className="text-sm font-semibold text-muted mb-2">{title}</h3>
      {children}
    </div>
  );
}

const METRICS: { key: string; label: string; unit: string }[] = [
  { key: 'aspect_pattern_count', label: 'Aspect Patterns', unit: 'patterns' },
  { key: 'draconic_bridge_count', label: 'Draconic Bridges', unit: 'bridges' },
  { key: 'paran_count', label: 'Parans', unit: 'contacts' },
  { key: 'declination_parallel_count', label: 'Declination Parallels', unit: 'pairs' },
  { key: 'mansion_convergence_count', label: 'Mansion Convergence', unit: 'planets' },
  { key: 'cross_system_sign_agreement', label: 'Cross-System Sign Agreement', unit: '%' },
  { key: 'arabic_parts_survivor_pct', label: 'Arabic Parts Survivors', unit: '%' },
  { key: 'stars_cross_survivor_pct', label: 'Stars Cross Survivors', unit: '%' },
  { key: 'harmonic_4', label: 'Harmonic 4 Conjunctions', unit: 'pairs' },
  { key: 'harmonic_5', label: 'Harmonic 5 Conjunctions', unit: 'pairs' },
  { key: 'harmonic_7', label: 'Harmonic 7 Conjunctions', unit: 'pairs' },
  { key: 'harmonic_9', label: 'Harmonic 9 Conjunctions', unit: 'pairs' },
];

function extractValue(metrics: Record<string, unknown>, key: string): number {
  if (key.startsWith('harmonic_')) {
    const h = parseInt(key.split('_')[1], 10);
    const hc = metrics.harmonic_conjunctions as Record<string, number> | undefined;
    return hc?.[String(h)] ?? 0;
  }
  return (metrics[key] as number) ?? 0;
}

function fmt(n: number): string {
  return n % 1 !== 0 ? n.toFixed(1) : String(n);
}

function exportCSV(result: BatchAnalysisResponse): void {
  const headers = ['Chart', ...METRICS.map((m) => m.label)];
  const rows = result.charts.map((c) => {
    const vals = METRICS.map(({ key }) => {
      const v = c.metrics
        ? extractValue(c.metrics as unknown as Record<string, unknown>, key)
        : null;
      return v != null ? fmt(v) : '';
    });
    return [c.name, ...vals].join(',');
  });
  const csv = [headers.join(','), ...rows].join('\n');
  const blob = new Blob([csv], { type: 'text/csv' });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = `batch-analysis-${new Date().toISOString().slice(0, 10)}.csv`;
  a.click();
  URL.revokeObjectURL(url);
}

export function BatchAnalysis() {
  const [savedCharts, setSavedCharts] = useState<SavedChart[]>([]);
  const [selectedIds, setSelectedIds] = useState<Set<number>>(new Set());
  const [result, setResult] = useState<BatchAnalysisResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    chartDB.getAll().then((charts) => {
      setSavedCharts(charts.sort((a, b) => b.updatedAt.localeCompare(a.updatedAt)));
    });
  }, []);

  const toggleChart = useCallback((id: number) => {
    setSelectedIds((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  }, []);

  const selectAll = useCallback(() => {
    setSelectedIds(new Set(savedCharts.map((c) => c.id!)));
  }, [savedCharts]);

  const clearSelection = useCallback(() => {
    setSelectedIds(new Set());
    setResult(null);
  }, []);

  const runAnalysis = useCallback(async () => {
    if (selectedIds.size === 0) return;
    setLoading(true);
    setError('');
    try {
      const charts: BirthData[] = savedCharts
        .filter((c) => selectedIds.has(c.id!))
        .map((c) => c.birthData);
      const r = await api.batchAnalysis(charts);
      setResult(r);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Batch analysis failed');
    } finally {
      setLoading(false);
    }
  }, [selectedIds, savedCharts]);

  return (
    <div className="space-y-4">
      {/* Chart Selection */}
      <Section title="Select Charts for Batch Analysis">
        <div className="flex gap-2 mb-3">
          <button
            onClick={selectAll}
            className="px-2 py-0.5 text-xs rounded bg-surface text-muted border border-border hover:text-text"
          >
            Select All
          </button>
          <button
            onClick={clearSelection}
            className="px-2 py-0.5 text-xs rounded bg-surface text-muted border border-border hover:text-text"
          >
            Clear
          </button>
          <span className="text-xs text-muted ml-auto">
            {selectedIds.size} of {savedCharts.length} selected
          </span>
        </div>

        {savedCharts.length === 0 ? (
          <p className="text-sm text-muted">No saved charts. Create charts first.</p>
        ) : (
          <div className="max-h-48 overflow-y-auto space-y-1">
            {savedCharts.map((chart) => (
              <label
                key={chart.id}
                className={`flex items-center gap-2 px-2 py-1 rounded text-sm cursor-pointer hover:bg-surface/50 ${
                  selectedIds.has(chart.id!) ? 'bg-accent/10' : ''
                }`}
              >
                <input
                  type="checkbox"
                  checked={selectedIds.has(chart.id!)}
                  onChange={() => toggleChart(chart.id!)}
                  className="accent-accent"
                />
                <span className="text-text">{chart.name}</span>
                <span className="text-muted text-xs">
                  {chart.birthData.year}-{String(chart.birthData.month).padStart(2, '0')}-
                  {String(chart.birthData.day).padStart(2, '0')}
                </span>
              </label>
            ))}
          </div>
        )}

        <button
          onClick={runAnalysis}
          disabled={selectedIds.size === 0 || loading}
          className={`mt-3 px-4 py-1.5 text-sm rounded ${
            selectedIds.size === 0 || loading
              ? 'bg-surface text-muted border border-border cursor-not-allowed'
              : 'bg-accent text-white hover:bg-accent/80'
          }`}
        >
          {loading ? 'Running…' : `Analyze ${selectedIds.size} Chart${selectedIds.size !== 1 ? 's' : ''}`}
        </button>
      </Section>

      {error && <p className="text-red text-sm">{error}</p>}

      {/* Results */}
      {result && (
        <>
          {/* Per-Chart Metrics */}
          <Section title={`Results — ${result.charts.length} Chart${result.charts.length !== 1 ? 's' : ''}`}>
            <div className="flex justify-end mb-2">
              <button
                onClick={() => exportCSV(result)}
                className="px-3 py-1 text-xs rounded bg-surface text-muted border border-border hover:text-text"
              >
                Export CSV
              </button>
            </div>
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="text-muted text-left border-b border-border">
                    <th className="py-2 pr-4">Chart</th>
                    {METRICS.map(({ key, label }) => (
                      <th key={key} className="py-2 pr-2 text-right text-xs">{label}</th>
                    ))}
                  </tr>
                </thead>
                <tbody>
                  {result.charts.map((chart) => (
                    <tr key={chart.name} className="border-t border-border hover:bg-surface/50">
                      <td className="py-2 pr-4 font-medium">{chart.name}</td>
                      {METRICS.map(({ key, unit }) => {
                        const val = chart.metrics
                          ? extractValue(chart.metrics as unknown as Record<string, unknown>, key)
                          : null;
                        return (
                          <td key={key} className="py-2 pr-2 text-right font-mono text-xs">
                            {val != null ? (
                              <>
                                {fmt(val)}
                                <span className="text-muted ml-0.5">{unit}</span>
                              </>
                            ) : (
                              <span className="text-muted">—</span>
                            )}
                          </td>
                        );
                      })}
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </Section>

          {/* Aggregate Statistics */}
          <Section title="Aggregate Statistics">
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="text-muted text-left border-b border-border">
                    <th className="py-2 pr-4">Metric</th>
                    <th className="py-2 pr-4 text-right">Mean</th>
                    <th className="py-2 pr-4 text-right">Median</th>
                    <th className="py-2 pr-4 text-right">Std Dev</th>
                    <th className="py-2 pr-4 text-right">Min</th>
                    <th className="py-2 pr-4 text-right">Max</th>
                    <th className="py-2 pr-4 text-right">P5</th>
                    <th className="py-2 pr-4 text-right">P95</th>
                  </tr>
                </thead>
                <tbody>
                  {METRICS.map(({ key, label }) => {
                    const agg = result.aggregates[key];
                    if (!agg) return null;
                    return (
                      <tr key={key} className="border-t border-border hover:bg-surface/50">
                        <td className="py-2 pr-4">{label}</td>
                        <td className="py-2 pr-4 text-right font-mono text-xs">{fmt(agg.mean)}</td>
                        <td className="py-2 pr-4 text-right font-mono text-xs">{fmt(agg.median)}</td>
                        <td className="py-2 pr-4 text-right font-mono text-xs">{fmt(agg.std_dev)}</td>
                        <td className="py-2 pr-4 text-right font-mono text-xs">{fmt(agg.min)}</td>
                        <td className="py-2 pr-4 text-right font-mono text-xs">{fmt(agg.max)}</td>
                        <td className="py-2 pr-4 text-right font-mono text-xs">{fmt(agg.p5)}</td>
                        <td className="py-2 pr-4 text-right font-mono text-xs">{fmt(agg.p95)}</td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          </Section>

          {/* Outliers */}
          <Section title="Outliers">
            <div className="space-y-2">
              {METRICS.map(({ key, label }) => {
                const agg = result.aggregates[key];
                if (!agg) return null;
                const threshold = agg.mean + 2 * agg.std_dev;
                const outliers = result.charts.filter((c) => {
                  if (!c.metrics) return false;
                  const val = extractValue(c.metrics as unknown as Record<string, unknown>, key);
                  return val > threshold;
                });
                if (outliers.length === 0) return null;
                return (
                  <div key={key} className="text-sm">
                    <span className="text-accent font-semibold">{label}</span>
                    <span className="text-muted"> — {outliers.length} chart{outliers.length !== 1 ? 's' : ''} above 2σ: </span>
                    {outliers.map((c, i) => (
                      <span key={c.name}>
                        <span className="text-text font-medium">{c.name}</span>
                        {i < outliers.length - 1 && <span className="text-muted">, </span>}
                      </span>
                    ))}
                  </div>
                );
              })}
              {METRICS.every(({ key }) => {
                const agg = result.aggregates[key];
                if (!agg) return true;
                const threshold = agg.mean + 2 * agg.std_dev;
                return !result.charts.some((c) => {
                  if (!c.metrics) return false;
                  const val = extractValue(c.metrics as unknown as Record<string, unknown>, key);
                  return val > threshold;
                });
              }) && (
                <p className="text-sm text-muted">No outliers detected at 2σ threshold.</p>
              )}
            </div>
          </Section>
        </>
      )}
    </div>
  );
}
