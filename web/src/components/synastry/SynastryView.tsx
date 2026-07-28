import { useState, useEffect } from 'react';
import type { SavedChart, SynastryResponse, CompositeResponse } from '../../lib/types';
import { api } from '../../lib/api';
import { chartDB } from '../../lib/db';
import { ChartWheel } from '../chart/ChartWheel';
import AspectGrid from './AspectGrid';

interface SynastryViewProps {
  chartA: SavedChart;
}

const PLANET_ORDER = [
  'Sun', 'Moon', 'Mercury', 'Venus', 'Mars',
  'Jupiter', 'Saturn', 'Uranus', 'Neptune', 'Pluto',
  'Chiron', 'Ceres', 'Pallas', 'Juno', 'Vesta',
  'Eris', 'TrueNode', 'SouthNode', 'Lilith',
];

export function SynastryView({ chartA }: SynastryViewProps) {
  const [savedCharts, setSavedCharts] = useState<SavedChart[]>([]);
  const [selectedChartId, setSelectedChartId] = useState<number | null>(null);
  const [chartB, setChartB] = useState<SavedChart | null>(null);
  const [synastry, setSynastry] = useState<SynastryResponse | null>(null);
  const [composite, setComposite] = useState<CompositeResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    chartDB.getAll().then((charts) => {
      setSavedCharts(charts.filter((c) => c.id !== chartA.id));
    });
  }, [chartA.id]);

  const handleSelectChart = async (id: number) => {
    setSelectedChartId(id);
    const chart = savedCharts.find((c) => c.id === id);
    if (!chart) return;
    setChartB(chart);

    setLoading(true);
    setError('');
    try {
      const [s, c] = await Promise.all([
        api.synastry(chartA.birthData, chart.birthData, 5),
        api.composite(chartA.birthData, chart.birthData, 3),
      ]);
      setSynastry(s);
      setComposite(c);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to load synastry');
    }
    setLoading(false);
  };

  // Extract unique planet names from aspects for the grid
  const planetSet1 = new Set<string>();
  const planetSet2 = new Set<string>();
  if (synastry) {
    for (const a of synastry.aspects) {
      planetSet1.add(a.planet1);
      planetSet2.add(a.planet2);
    }
  }
  const planets1 = PLANET_ORDER.filter((p) => planetSet1.has(p));
  const planets2 = PLANET_ORDER.filter((p) => planetSet2.has(p));

  return (
    <div className="space-y-4">
      {/* Chart B Selector */}
      <div className="bg-surface border border-border rounded-lg p-4">
        <h3 className="text-sm font-semibold text-muted mb-2">Select Second Chart</h3>
        {savedCharts.length === 0 ? (
          <p className="text-sm text-muted">
            No other saved charts. Create another chart first.
          </p>
        ) : (
          <select
            value={selectedChartId ?? ''}
            onChange={(e) => handleSelectChart(Number(e.target.value))}
            className="bg-surface border border-border rounded px-3 py-2 text-sm text-text w-full max-w-xs"
          >
            <option value="" disabled>
              Choose a chart...
            </option>
            {savedCharts.map((c) => (
              <option key={c.id} value={c.id!}>
                {c.name} ({c.birthData.year}-{String(c.birthData.month).padStart(2, '0')}-{String(c.birthData.day).padStart(2, '0')})
              </option>
            ))}
          </select>
        )}
      </div>

      {loading && <p className="text-yellow text-sm">Loading synastry...</p>}
      {error && <p className="text-red text-sm">{error}</p>}

      {chartB && synastry && (
        <>
          {/* Side-by-side Wheels */}
          <div className="grid grid-cols-2 gap-4">
            <div className="bg-surface border border-border rounded-lg p-4">
              <h3 className="text-sm font-semibold text-muted mb-2 text-center">
                {chartA.name}
              </h3>
              <div style={{ height: 300 }}>
                <ChartWheel data={chartA.birthData} />
              </div>
            </div>
            <div className="bg-surface border border-border rounded-lg p-4">
              <h3 className="text-sm font-semibold text-muted mb-2 text-center">
                {chartB.name}
              </h3>
              <div style={{ height: 300 }}>
                <ChartWheel data={chartB.birthData} />
              </div>
            </div>
          </div>

          {/* Aspect Grid */}
          <div className="bg-surface border border-border rounded-lg p-4">
            <h3 className="text-sm font-semibold text-muted mb-2">
              Aspect Grid ({synastry.aspects.length} aspects)
            </h3>
            <AspectGrid
              aspects={synastry.aspects}
              planets1={planets1}
              planets2={planets2}
            />
          </div>

          {/* Composite Chart */}
          {composite && (
            <div className="bg-surface border border-border rounded-lg p-4">
              <h3 className="text-sm font-semibold text-muted mb-2">
                Composite Chart
              </h3>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div style={{ height: 300 }}>
                  <ChartWheel data={{
                    name: `${chartA.name} + ${chartB.name}`,
                    year: chartA.birthData.year,
                    month: chartA.birthData.month,
                    day: chartA.birthData.day,
                    hour: chartA.birthData.hour,
                    minute: chartA.birthData.minute,
                    tz_offset: chartA.birthData.tz_offset,
                    lat: chartA.birthData.lat,
                    lng: chartA.birthData.lng,
                  }} />
                </div>
                <div className="text-sm space-y-2">
                  <p className="text-muted">
                    {composite.patterns?.length ?? 0} patterns detected
                  </p>
                  {composite.patterns?.map((p, i) => (
                    <div key={i} className="text-sm py-0.5">
                      <span className="text-accent">{p.name}:</span> {p.planets.join(', ')}
                    </div>
                  ))}
                  {composite.aspects?.length > 0 && (
                    <p className="text-muted pt-2">
                      {composite.aspects.length} aspects in composite
                    </p>
                  )}
                </div>
              </div>
            </div>
          )}
        </>
      )}
    </div>
  );
}
