import { useState, useEffect } from 'react';
import type { SavedChart, BirthData, SynastryResponse, CompositeResponse, DraconicSynastryFullResponse, Aspect } from '../../lib/types';
import { api } from '../../lib/api';
import { chartDB } from '../../lib/db';
import { ChartWheel } from '../chart/ChartWheel';
import AspectGrid from './AspectGrid';
import { SynastryReport } from '../reports/SynastryReport';

interface SynastryViewProps {
  chartA: SavedChart;
}

const PLANET_ORDER = [
  'Sun', 'Moon', 'Mercury', 'Venus', 'Mars',
  'Jupiter', 'Saturn', 'Uranus', 'Neptune', 'Pluto',
  'Chiron', 'Ceres', 'Pallas', 'Juno', 'Vesta',
  'Eris', 'TrueNode', 'SouthNode', 'Lilith',
];

function aspectSummary(aspects: Aspect[]): Record<string, number> {
  const counts: Record<string, number> = {};
  for (const a of aspects) {
    counts[a.aspect] = (counts[a.aspect] || 0) + 1;
  }
  return counts;
}

function strongestAspects(aspects: Aspect[], n = 5): Aspect[] {
  return [...aspects].sort((a, b) => a.orb - b.orb).slice(0, n);
}

export function SynastryView({ chartA }: SynastryViewProps) {
  const [savedCharts, setSavedCharts] = useState<SavedChart[]>([]);
  const [selectedChartId, setSelectedChartId] = useState<number | null>(null);
  const [chartB, setChartB] = useState<SavedChart | null>(null);
  const [synastry, setSynastry] = useState<SynastryResponse | null>(null);
  const [composite, setComposite] = useState<CompositeResponse | null>(null);
  const [draconicFull, setDraconicFull] = useState<DraconicSynastryFullResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [activeTab, setActiveTab] = useState<'synastry' | 'composite' | 'draconic' | 'synthesis' | 'report'>('synastry');

  // Partner form state
  const [showPartnerForm, setShowPartnerForm] = useState(false);
  const [partnerName, setPartnerName] = useState('');
  const [partnerYear, setPartnerYear] = useState(1970);
  const [partnerMonth, setPartnerMonth] = useState(1);
  const [partnerDay, setPartnerDay] = useState(1);
  const [partnerHour, setPartnerHour] = useState(12);
  const [partnerMinute, setPartnerMinute] = useState(0);
  const [partnerTz, setPartnerTz] = useState(-5);
  const [partnerLat, setPartnerLat] = useState(40.7128);
  const [partnerLng, setPartnerLng] = useState(-74.006);

  useEffect(() => {
    chartDB.getAll().then((charts) => {
      setSavedCharts(charts.filter((c) => c.id !== chartA.id));
    });
  }, [chartA.id]);

  const loadSynastry = async (bd: BirthData) => {
    setLoading(true);
    setError('');
    try {
      const [s, c, df] = await Promise.all([
        api.synastry(chartA.birthData, bd, 5),
        api.composite(chartA.birthData, bd, 3),
        api.draconicSynastryFull(chartA.birthData, bd, 5),
      ]);
      setSynastry(s);
      setComposite(c);
      setDraconicFull(df);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to load synastry');
    }
    setLoading(false);
  };

  const handleSelectChart = async (id: number) => {
    setSelectedChartId(id);
    const chart = savedCharts.find((c) => c.id === id);
    if (!chart) return;
    setChartB(chart);
    await loadSynastry(chart.birthData);
  };

  const handleSavePartner = async () => {
    const bd: BirthData = {
      name: partnerName || 'Partner',
      year: partnerYear, month: partnerMonth, day: partnerDay,
      hour: partnerHour, minute: partnerMinute,
      tz_offset: partnerTz, lat: partnerLat, lng: partnerLng,
    };
    const id = await chartDB.createFromBirthData(bd.name, bd);
    const chart: SavedChart = {
      id, name: bd.name, birthData: bd, houseSystem: 'placidus',
      tags: [], notes: '', createdAt: new Date().toISOString(), updatedAt: new Date().toISOString(),
    };
    setChartB(chart);
    setSelectedChartId(id);
    setShowPartnerForm(false);
    setSavedCharts(prev => [...prev, chart]);
    await loadSynastry(bd);
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

  // House overlays: partner's planets in chartA's houses (derived from composite)
  const houseOverlays: { planet: string; sign: string; house: number }[] = [];
  if (composite) {
    for (const [planet, lon] of Object.entries(composite.planets)) {
      const signIdx = Math.floor((lon % 360) / 30);
      const signs = ['Aries', 'Taurus', 'Gemini', 'Cancer', 'Leo', 'Virgo', 'Libra', 'Scorpio', 'Sagittarius', 'Capricorn', 'Aquarius', 'Pisces'];
      houseOverlays.push({ planet, sign: signs[signIdx], house: 0 });
    }
  }

  // Bridges: aspects that appear in all three draconic layers
  const bridges: Aspect[] = [];
  if (draconicFull) {
    const abSet = new Set(draconicFull.trop_a_to_drac_b.map(a => `${a.planet1}|${a.planet2}|${a.aspect}`));
    const baSet = new Set(draconicFull.trop_b_to_drac_a.map(a => `${a.planet1}|${a.planet2}|${a.aspect}`));
    for (const a of draconicFull.drac_to_drac) {
      const key = `${a.planet1}|${a.planet2}|${a.aspect}`;
      if (abSet.has(key) && baSet.has(key)) {
        bridges.push(a);
      }
    }
  }

  return (
    <div className="space-y-4">
      {/* Chart B Selector + Partner Form */}
      <div className="bg-surface border border-border rounded-lg p-4">
        <div className="flex items-center gap-4 flex-wrap">
          <h3 className="text-sm font-semibold text-muted">Second Chart</h3>
          {savedCharts.length > 0 && (
            <select
              value={selectedChartId ?? ''}
              onChange={(e) => handleSelectChart(Number(e.target.value))}
              className="bg-surface border border-border rounded px-3 py-1.5 text-sm text-text"
            >
              <option value="" disabled>Choose a chart...</option>
              {savedCharts.map((c) => (
                <option key={c.id} value={c.id!}>
                  {c.name} ({c.birthData.year}-{String(c.birthData.month).padStart(2, '0')}-{String(c.birthData.day).padStart(2, '0')})
                </option>
              ))}
            </select>
          )}
          <button
            onClick={() => setShowPartnerForm(!showPartnerForm)}
            className="px-3 py-1.5 text-sm rounded bg-accent/20 text-accent border border-accent/30"
          >
            {showPartnerForm ? 'Cancel' : '+ New Partner'}
          </button>
        </div>

        {showPartnerForm && (
          <div className="mt-3 grid grid-cols-2 sm:grid-cols-4 gap-2">
            <input placeholder="Name" value={partnerName} onChange={e => setPartnerName(e.target.value)}
              className="bg-bg border border-border rounded px-2 py-1 text-sm text-text" />
            <input type="number" placeholder="Year" value={partnerYear} onChange={e => setPartnerYear(Number(e.target.value))}
              className="bg-bg border border-border rounded px-2 py-1 text-sm text-text" />
            <input type="number" placeholder="Month" value={partnerMonth} onChange={e => setPartnerMonth(Number(e.target.value))}
              className="bg-bg border border-border rounded px-2 py-1 text-sm text-text" min={1} max={12} />
            <input type="number" placeholder="Day" value={partnerDay} onChange={e => setPartnerDay(Number(e.target.value))}
              className="bg-bg border border-border rounded px-2 py-1 text-sm text-text" min={1} max={31} />
            <input type="number" placeholder="Hour" value={partnerHour} onChange={e => setPartnerHour(Number(e.target.value))}
              className="bg-bg border border-border rounded px-2 py-1 text-sm text-text" min={0} max={23} />
            <input type="number" placeholder="Minute" value={partnerMinute} onChange={e => setPartnerMinute(Number(e.target.value))}
              className="bg-bg border border-border rounded px-2 py-1 text-sm text-text" min={0} max={59} />
            <input type="number" placeholder="TZ offset" value={partnerTz} onChange={e => setPartnerTz(Number(e.target.value))}
              className="bg-bg border border-border rounded px-2 py-1 text-sm text-text" step={0.5} />
            <input type="number" placeholder="Lat" value={partnerLat} onChange={e => setPartnerLat(Number(e.target.value))}
              className="bg-bg border border-border rounded px-2 py-1 text-sm text-text" step={0.01} />
            <input type="number" placeholder="Lng" value={partnerLng} onChange={e => setPartnerLng(Number(e.target.value))}
              className="bg-bg border border-border rounded px-2 py-1 text-sm text-text" step={0.01} />
            <button onClick={handleSavePartner}
              className="px-3 py-1 text-sm rounded bg-accent text-white col-span-2 sm:col-span-1">
              Save & Load
            </button>
          </div>
        )}
      </div>

      {loading && <p className="text-yellow text-sm">Loading synastry...</p>}
      {error && <p className="text-red text-sm">{error}</p>}

      {chartB && synastry && (
        <>
          {/* Side-by-side Wheels */}
          <div className="grid grid-cols-2 gap-4">
            <div className="bg-surface border border-border rounded-lg p-4">
              <h3 className="text-sm font-semibold text-muted mb-2 text-center">{chartA.name}</h3>
              <div style={{ height: 300 }}>
                <ChartWheel data={chartA.birthData} />
              </div>
            </div>
            <div className="bg-surface border border-border rounded-lg p-4">
              <h3 className="text-sm font-semibold text-muted mb-2 text-center">{chartB.name}</h3>
              <div style={{ height: 300 }}>
                <ChartWheel data={chartB.birthData} />
              </div>
            </div>
          </div>

          {/* Sub-tabs */}
          <div className="flex gap-2">
            {(['synastry', 'composite', 'draconic', 'synthesis', 'report'] as const).map(tab => (
              <button key={tab}
                onClick={() => setActiveTab(tab)}
                className={`px-3 py-1 text-sm rounded ${activeTab === tab ? 'bg-accent text-white' : 'bg-surface text-muted border border-border'}`}>
                {tab === 'synastry' ? 'Aspects' : tab === 'composite' ? 'Composite' : tab === 'draconic' ? 'Draconic' : tab === 'synthesis' ? 'Synthesis' : 'Report'}
              </button>
            ))}
          </div>

          {/* Synastry Aspects */}
          {activeTab === 'synastry' && (
            <div className="space-y-4">
              <div className="bg-surface border border-border rounded-lg p-4">
                <h3 className="text-sm font-semibold text-muted mb-2">
                  Aspect Grid ({synastry.aspects.length} aspects)
                </h3>
                <AspectGrid aspects={synastry.aspects} planets1={planets1} planets2={planets2} />
              </div>

              {/* House Overlays */}
              {houseOverlays.length > 0 && (
                <div className="bg-surface border border-border rounded-lg p-4">
                  <h3 className="text-sm font-semibold text-muted mb-2">
                    House Overlays — {chartB.name}'s planets in {chartA.name}'s houses
                  </h3>
                  <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 gap-1">
                    {houseOverlays.map((ho, i) => (
                      <div key={i} className="text-sm px-2 py-1 bg-bg rounded">
                        <span className="text-accent">{ho.planet}</span>
                        <span className="text-muted"> → </span>
                        <span className="text-text">H{ho.house}</span>
                        <span className="text-muted text-xs ml-1">({ho.sign})</span>
                      </div>
                    ))}
                  </div>
                </div>
              )}
            </div>
          )}

          {/* Composite */}
          {activeTab === 'composite' && composite && (
            <div className="bg-surface border border-border rounded-lg p-4">
              <h3 className="text-sm font-semibold text-muted mb-2">Composite Chart</h3>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div className="text-sm space-y-1">
                  <p className="text-muted text-xs mb-2">Midpoint positions</p>
                  {Object.entries(composite.planets).sort(([a], [b]) => a.localeCompare(b)).map(([planet, lon]) => {
                    const signIdx = Math.floor((lon % 360) / 30);
                    const signs = ['Aries', 'Taurus', 'Gemini', 'Cancer', 'Leo', 'Virgo', 'Libra', 'Scorpio', 'Sagittarius', 'Capricorn', 'Aquarius', 'Pisces'];
                    const deg = (lon % 30).toFixed(1);
                    return (
                      <div key={planet} className="text-sm px-2 py-0.5 bg-bg rounded flex justify-between">
                        <span className="text-accent">{planet}</span>
                        <span className="text-muted">{deg}° {signs[signIdx]}</span>
                      </div>
                    );
                  })}
                </div>
                <div className="text-sm space-y-2">
                  <p className="text-muted">{Object.keys(composite.planets).length} planets</p>
                  <p className="text-muted">{composite.aspects?.length ?? 0} aspects</p>
                  <p className="text-muted">{composite.patterns?.length ?? 0} patterns</p>
                  {composite.patterns?.map((p, i) => (
                    <div key={i} className="text-sm py-0.5">
                      <span className="text-accent">{p.name}:</span> {p.planets?.join(', ') || ''}
                    </div>
                  ))}
                </div>
              </div>
            </div>
          )}

          {/* Draconic Synastry */}
          {activeTab === 'draconic' && draconicFull && (
            <div className="space-y-4">
              <div className="bg-surface border border-border rounded-lg p-4">
                <h3 className="text-sm font-semibold text-muted mb-2">
                  Draconic-to-Draconic ({draconicFull.drac_to_drac.length} aspects)
                </h3>
                <p className="text-xs text-muted mb-2">Soul-level connection — both charts rotated to NN=0° Aries</p>
                <AspectGrid aspects={draconicFull.drac_to_drac} planets1={planets1} planets2={planets2} />
              </div>

              <div className="bg-surface border border-border rounded-lg p-4">
                <h3 className="text-sm font-semibold text-muted mb-2">
                  Tropical A → Draconic B ({draconicFull.trop_a_to_drac_b.length} aspects)
                </h3>
                <p className="text-xs text-muted mb-2">How {chartA.name}'s natal self connects to {chartB.name}'s soul</p>
                <AspectGrid aspects={draconicFull.trop_a_to_drac_b} planets1={planets1} planets2={planets2} />
              </div>

              <div className="bg-surface border border-border rounded-lg p-4">
                <h3 className="text-sm font-semibold text-muted mb-2">
                  Tropical B → Draconic A ({draconicFull.trop_b_to_drac_a.length} aspects)
                </h3>
                <p className="text-xs text-muted mb-2">How {chartB.name}'s natal self connects to {chartA.name}'s soul</p>
                <AspectGrid aspects={draconicFull.trop_b_to_drac_a} planets1={planets1} planets2={planets2} />
              </div>

              {/* Bridges */}
              {bridges.length > 0 && (
                <div className="bg-surface border border-accent/30 rounded-lg p-4">
                  <h3 className="text-sm font-semibold text-accent mb-2">
                    🌉 Bridges — {bridges.length} aspects in all 3 layers
                  </h3>
                  <p className="text-xs text-muted mb-2">These connections persist across tropical and draconic — the strongest relationship signatures</p>
                  <div className="space-y-1">
                    {bridges.map((b, i) => (
                      <div key={i} className="text-sm px-2 py-0.5 bg-bg rounded">
                        <span className="text-accent">{b.planet1}</span>
                        <span className="text-muted"> {b.aspect} </span>
                        <span className="text-accent">{b.planet2}</span>
                        <span className="text-muted text-xs ml-1">({b.orb}°)</span>
                      </div>
                    ))}
                  </div>
                </div>
              )}
            </div>
          )}

          {/* Synthesis */}
          {activeTab === 'synthesis' && synastry && (
            <div className="space-y-4">
              {/* Summary Card */}
              <div className="bg-surface border border-border rounded-lg p-4">
                <h3 className="text-sm font-semibold text-muted mb-3">Relationship Summary</h3>
                <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
                  <div className="text-center">
                    <div className="text-2xl font-bold text-accent">{synastry.aspects.length}</div>
                    <div className="text-xs text-muted">Total Aspects</div>
                  </div>
                  {Object.entries(aspectSummary(synastry.aspects)).sort(([,a], [,b]) => b - a).slice(0, 3).map(([aspect, count]) => (
                    <div key={aspect} className="text-center">
                      <div className="text-2xl font-bold text-text">{count}</div>
                      <div className="text-xs text-muted capitalize">{aspect}s</div>
                    </div>
                  ))}
                </div>
              </div>

              {/* Strongest Connections */}
              <div className="bg-surface border border-border rounded-lg p-4">
                <h3 className="text-sm font-semibold text-muted mb-2">Strongest Connections</h3>
                <div className="space-y-1">
                  {strongestAspects(synastry.aspects, 8).map((a, i) => (
                    <div key={i} className="text-sm px-2 py-0.5 bg-bg rounded flex justify-between">
                      <span>
                        <span className="text-accent">{a.planet1}</span>
                        <span className="text-muted"> {a.aspect} </span>
                        <span className="text-accent">{a.planet2}</span>
                      </span>
                      <span className="text-muted text-xs">{a.orb}°</span>
                    </div>
                  ))}
                </div>
              </div>

              {/* House Emphasis */}
              {houseOverlays.length > 0 && (
                <div className="bg-surface border border-border rounded-lg p-4">
                  <h3 className="text-sm font-semibold text-muted mb-2">House Emphasis</h3>
                  <div className="space-y-1">
                    {(() => {
                      const counts: Record<number, string[]> = {};
                      for (const ho of houseOverlays) {
                        if (!counts[ho.house]) counts[ho.house] = [];
                        counts[ho.house].push(ho.planet);
                      }
                      return Object.entries(counts)
                        .sort(([, a], [, b]) => b.length - a.length)
                        .map(([house, planets]) => (
                          <div key={house} className="text-sm px-2 py-0.5 bg-bg rounded flex justify-between">
                            <span className="text-accent">House {house}</span>
                            <span className="text-muted">{planets.join(', ')}</span>
                          </div>
                        ));
                    })()}
                  </div>
                </div>
              )}

              {/* Draconic Bridge Summary */}
              {bridges.length > 0 && (
                <div className="bg-surface border border-accent/30 rounded-lg p-4">
                  <h3 className="text-sm font-semibold text-accent mb-2">
                    🌉 {bridges.length} Draconic Bridges
                  </h3>
                  <p className="text-xs text-muted">
                    These aspects survive all three draconic layers — the most structurally significant connections in this relationship.
                  </p>
                </div>
              )}
            </div>
          )}

          {/* Report Tab */}
          {activeTab === 'report' && chartB && (
            <SynastryReport
              chartA={chartA.birthData}
              chartB={chartB.birthData}
              nameA={chartA.name}
              nameB={chartB.name}
            />
          )}
        </>
      )}
    </div>
  );
}
