import { useState, useEffect, useCallback } from 'react';
import type { SavedChart } from '../../lib/types';
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

const PLANETS = ['Sun', 'Moon', 'Mercury', 'Venus', 'Mars', 'Jupiter', 'Saturn', 'Uranus', 'Neptune', 'Pluto', 'Node', 'Chiron'];
const SIGNS = ['Aries', 'Taurus', 'Gemini', 'Cancer', 'Leo', 'Virgo', 'Libra', 'Scorpio', 'Sagittarius', 'Capricorn', 'Aquarius', 'Pisces'];
const ASPECTS = ['conjunction', 'opposition', 'trine', 'square', 'sextile', 'quincunx'];

interface QueryCondition {
  id: string;
  type: 'planet_in_sign' | 'planet_in_house' | 'aspect' | 'pattern';
  planet?: string;
  sign?: string;
  house?: number;
  planet2?: string;
  aspect?: string;
  orb?: number;
  pattern?: string;
}

let condId = 0;
function newCond(): QueryCondition {
  return { id: String(++condId), type: 'planet_in_sign' };
}

export function ChartSearch() {
  const [savedCharts, setSavedCharts] = useState<SavedChart[]>([]);
  const [conditions, setConditions] = useState<QueryCondition[]>([newCond()]);
  const [results, setResults] = useState<SavedChart[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [searched, setSearched] = useState(false);

  // Simple text search
  const [textQuery, setTextQuery] = useState('');

  useEffect(() => {
    chartDB.getAll().then((charts) => {
      setSavedCharts(charts.sort((a, b) => b.updatedAt.localeCompare(a.updatedAt)));
    });
  }, []);

  const updateCond = useCallback((id: string, updates: Partial<QueryCondition>) => {
    setConditions((prev) => prev.map((c) => (c.id === id ? { ...c, ...updates } : c)));
  }, []);

  const addCondition = useCallback(() => {
    setConditions((prev) => [...prev, newCond()]);
  }, []);

  const removeCondition = useCallback((id: string) => {
    setConditions((prev) => prev.filter((c) => c.id !== id));
  }, []);

  const checkChart = useCallback(async (chart: SavedChart, conds: QueryCondition[]): Promise<boolean> => {
    if (conds.length === 0) return true;

    try {
      // Fetch chart data once
      const [interp, patternsResp] = await Promise.all([
        api.interpretation(chart.birthData, 'western', 3),
        api.patterns(chart.birthData, 3),
      ]);

      // We need planet positions — use the base chart endpoint
      const base = await api.baseChart(chart.birthData);

      // Build lookup maps
      const planetSigns: Record<string, string> = {};
      const planetHouses: Record<string, number> = {};
      const signNames = ['Aries', 'Taurus', 'Gemini', 'Cancer', 'Leo', 'Virgo', 'Libra', 'Scorpio', 'Sagittarius', 'Capricorn', 'Aquarius', 'Pisces'];

      for (const [name, pos] of Object.entries(base.tropical)) {
        const signIdx = Math.floor(pos.lon / 30);
        planetSigns[name] = signNames[signIdx];
        // Approximate house from longitude vs ASC
        const houseLon = ((pos.lon - base.asc + 360) % 360);
        planetHouses[name] = Math.floor(houseLon / 30) + 1;
      }

      // Also check western interpretation for planet signs
      if (interp.planet_signs) {
        for (const line of interp.planet_signs) {
          for (const planet of PLANETS) {
            if (line.startsWith(planet + ' in ') || line.startsWith(planet + ': ')) {
              for (const sign of SIGNS) {
                if (line.includes(sign)) {
                  planetSigns[planet] = sign;
                  break;
                }
              }
            }
          }
        }
      }

      const patternNames = patternsResp?.patterns?.map((p) => p.name.toLowerCase()) ?? [];

      for (const cond of conds) {
        switch (cond.type) {
          case 'planet_in_sign': {
            if (!cond.planet || !cond.sign) continue;
            if (planetSigns[cond.planet] !== cond.sign) return false;
            break;
          }
          case 'planet_in_house': {
            if (!cond.planet || !cond.house) continue;
            if (planetHouses[cond.planet] !== cond.house) return false;
            break;
          }
          case 'aspect': {
            if (!cond.planet || !cond.planet2 || !cond.aspect) continue;
            const aspects = interp.aspects ?? [];
            const found = aspects.some((a) => {
              const match = (a.includes(cond.planet!) && a.includes(cond.planet2!)) ||
                (a.includes(cond.planet2!) && a.includes(cond.planet!));
              return match && a.toLowerCase().includes(cond.aspect!.toLowerCase());
            });
            if (!found) return false;
            break;
          }
          case 'pattern': {
            if (!cond.pattern) continue;
            if (!patternNames.includes(cond.pattern.toLowerCase())) return false;
            break;
          }
        }
      }
      return true;
    } catch {
      return false;
    }
  }, []);

  const runSearch = useCallback(async () => {
    setLoading(true);
    setError('');
    setSearched(true);

    try {
      // First filter by text query (fast, IndexedDB)
      let candidates = savedCharts;
      if (textQuery.trim()) {
        const q = textQuery.trim().toLowerCase();
        candidates = savedCharts.filter(
          (c) =>
            c.name.toLowerCase().includes(q) ||
            c.tags.some((t) => t.toLowerCase().includes(q)) ||
            c.notes.toLowerCase().includes(q)
        );
      }

      // If no astrological conditions, return text results immediately
      const activeConds = conditions.filter(
        (c) => {
          if (c.type === 'planet_in_sign') return c.planet && c.sign;
          if (c.type === 'planet_in_house') return c.planet && c.house;
          if (c.type === 'aspect') return c.planet && c.planet2 && c.aspect;
          if (c.type === 'pattern') return c.pattern;
          return false;
        }
      );

      if (activeConds.length === 0) {
        setResults(candidates);
        return;
      }

      // Check each candidate against conditions
      const matched: SavedChart[] = [];
      for (const chart of candidates) {
        const ok = await checkChart(chart, activeConds);
        if (ok) matched.push(chart);
      }
      setResults(matched);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Search failed');
    } finally {
      setLoading(false);
    }
  }, [savedCharts, textQuery, conditions, checkChart]);

  return (
    <div className="space-y-4">
      {/* Text Search */}
      <Section title="Quick Search">
        <div className="flex gap-2">
          <input
            type="text"
            value={textQuery}
            onChange={(e) => setTextQuery(e.target.value)}
            placeholder="Search by name, tag, or notes..."
            className="flex-1 px-3 py-1.5 text-sm bg-background border border-border rounded text-text placeholder:text-muted"
            onKeyDown={(e) => e.key === 'Enter' && runSearch()}
          />
          <button
            onClick={runSearch}
            disabled={loading}
            className="px-4 py-1.5 text-sm rounded bg-accent text-white hover:bg-accent/80 disabled:opacity-50"
          >
            {loading ? 'Searching…' : 'Search'}
          </button>
        </div>
      </Section>

      {/* Query Builder */}
      <Section title="Astrological Query Builder">
        <p className="text-xs text-muted mb-3">
          Add conditions to find charts matching specific astrological features. All conditions must match (AND logic).
        </p>

        <div className="space-y-2">
          {conditions.map((cond) => (
            <div key={cond.id} className="flex gap-2 items-center flex-wrap">
              <select
                value={cond.type}
                onChange={(e) => updateCond(cond.id, { type: e.target.value as QueryCondition['type'] })}
                className="px-2 py-1 text-xs bg-background border border-border rounded text-text"
              >
                <option value="planet_in_sign">Planet in Sign</option>
                <option value="planet_in_house">Planet in House</option>
                <option value="aspect">Aspect</option>
                <option value="pattern">Pattern</option>
              </select>

              {cond.type === 'planet_in_sign' && (
                <>
                  <select
                    value={cond.planet ?? ''}
                    onChange={(e) => updateCond(cond.id, { planet: e.target.value || undefined })}
                    className="px-2 py-1 text-xs bg-background border border-border rounded text-text"
                  >
                    <option value="">Planet…</option>
                    {PLANETS.map((p) => (
                      <option key={p} value={p}>{p}</option>
                    ))}
                  </select>
                  <span className="text-muted text-xs">in</span>
                  <select
                    value={cond.sign ?? ''}
                    onChange={(e) => updateCond(cond.id, { sign: e.target.value || undefined })}
                    className="px-2 py-1 text-xs bg-background border border-border rounded text-text"
                  >
                    <option value="">Sign…</option>
                    {SIGNS.map((s) => (
                      <option key={s} value={s}>{s}</option>
                    ))}
                  </select>
                </>
              )}

              {cond.type === 'planet_in_house' && (
                <>
                  <select
                    value={cond.planet ?? ''}
                    onChange={(e) => updateCond(cond.id, { planet: e.target.value || undefined })}
                    className="px-2 py-1 text-xs bg-background border border-border rounded text-text"
                  >
                    <option value="">Planet…</option>
                    {PLANETS.map((p) => (
                      <option key={p} value={p}>{p}</option>
                    ))}
                  </select>
                  <span className="text-muted text-xs">in house</span>
                  <select
                    value={cond.house ?? ''}
                    onChange={(e) => updateCond(cond.id, { house: e.target.value ? parseInt(e.target.value) : undefined })}
                    className="px-2 py-1 text-xs bg-background border border-border rounded text-text"
                  >
                    <option value="">House…</option>
                    {Array.from({ length: 12 }, (_, i) => i + 1).map((h) => (
                      <option key={h} value={h}>{h}</option>
                    ))}
                  </select>
                </>
              )}

              {cond.type === 'aspect' && (
                <>
                  <select
                    value={cond.planet ?? ''}
                    onChange={(e) => updateCond(cond.id, { planet: e.target.value || undefined })}
                    className="px-2 py-1 text-xs bg-background border border-border rounded text-text"
                  >
                    <option value="">Planet…</option>
                    {PLANETS.map((p) => (
                      <option key={p} value={p}>{p}</option>
                    ))}
                  </select>
                  <select
                    value={cond.aspect ?? ''}
                    onChange={(e) => updateCond(cond.id, { aspect: e.target.value || undefined })}
                    className="px-2 py-1 text-xs bg-background border border-border rounded text-text"
                  >
                    <option value="">Aspect…</option>
                    {ASPECTS.map((a) => (
                      <option key={a} value={a}>{a}</option>
                    ))}
                  </select>
                  <select
                    value={cond.planet2 ?? ''}
                    onChange={(e) => updateCond(cond.id, { planet2: e.target.value || undefined })}
                    className="px-2 py-1 text-xs bg-background border border-border rounded text-text"
                  >
                    <option value="">Planet…</option>
                    {PLANETS.map((p) => (
                      <option key={p} value={p}>{p}</option>
                    ))}
                  </select>
                </>
              )}

              {cond.type === 'pattern' && (
                <input
                  type="text"
                  value={cond.pattern ?? ''}
                  onChange={(e) => updateCond(cond.id, { pattern: e.target.value || undefined })}
                  placeholder="e.g. Grand Trine, T-Square, Yod…"
                  className="px-2 py-1 text-xs bg-background border border-border rounded text-text placeholder:text-muted flex-1 min-w-[200px]"
                />
              )}

              <button
                onClick={() => removeCondition(cond.id)}
                className="px-2 py-1 text-xs rounded text-red hover:bg-red/10"
                title="Remove condition"
              >
                ✕
              </button>
            </div>
          ))}
        </div>

        <button
          onClick={addCondition}
          className="mt-2 px-3 py-1 text-xs rounded bg-surface text-muted border border-border hover:text-text"
        >
          + Add Condition
        </button>
      </Section>

      {error && <p className="text-red text-sm">{error}</p>}

      {/* Results */}
      {searched && (
        <Section title={`Results — ${results.length} Chart${results.length !== 1 ? 's' : ''}`}>
          {results.length === 0 ? (
            <p className="text-sm text-muted">No charts match your query.</p>
          ) : (
            <div className="space-y-1">
              {results.map((chart) => (
                <div
                  key={chart.id}
                  className="flex items-center gap-3 px-3 py-2 rounded bg-background border border-border text-sm"
                >
                  <span className="text-text font-medium">{chart.name}</span>
                  <span className="text-muted text-xs">
                    {chart.birthData.year}-{String(chart.birthData.month).padStart(2, '0')}-
                    {String(chart.birthData.day).padStart(2, '0')}
                  </span>
                  {chart.tags.length > 0 && (
                    <span className="text-muted text-xs">
                      {chart.tags.join(', ')}
                    </span>
                  )}
                </div>
              ))}
            </div>
          )}
        </Section>
      )}
    </div>
  );
}
