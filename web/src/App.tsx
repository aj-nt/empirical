import { useState, useCallback, useEffect, Component } from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { Sidebar } from './components/layout/Sidebar';
import { ChartForm } from './components/shared/ChartForm';
import { ChartWheel } from './components/chart/ChartWheel';
import { DispositorTree } from './components/chart/DispositorTree';
import { SynastryView } from './components/synastry/SynastryView';
import { AstroCartographyMap } from './components/maps/AstroCartographyMap';
import { NatalReport } from './components/reports/NatalReport';
import { TransitReport } from './components/reports/TransitReport';
import { PageDesigner } from './components/reports/PageDesigner';
import { GraphicEphemeris } from './components/ephemeris/GraphicEphemeris';
import { TransitCalendar } from './components/calendar/TransitCalendar';
import { CrossSystemComparison } from './components/research/CrossSystemComparison';
import { DraconicView } from './components/research/DraconicView';
import { VedicView } from './components/research/VedicView';
import { ResearchTools } from './components/research/ResearchTools';
import { EmpiricalBaselines } from './components/research/EmpiricalBaselines';
import { BatchAnalysis } from './components/research/BatchAnalysis';
import { ChartSearch } from './components/research/ChartSearch';
import type { SavedChart, BirthData, InterpretationResponse, TransitResponse, TraditionalResponse } from './lib/types';
import { chartDB } from './lib/db';
import { api } from './lib/api';

class ErrorBoundary extends Component<{ children: React.ReactNode }, { error: Error | null }> {
  state = { error: null as Error | null };
  static getDerivedStateFromError(error: Error) { return { error }; }
  render() {
    if (this.state.error) {
      return (
        <div className="p-4 bg-red-900/20 border border-red-500 rounded-lg">
          <h3 className="text-red text-sm font-semibold mb-2">Research Tab Error</h3>
          <pre className="text-red text-xs whitespace-pre-wrap">{this.state.error.message}</pre>
          <pre className="text-muted text-xs mt-2 whitespace-pre-wrap">{this.state.error.stack}</pre>
        </div>
      );
    }
    return this.props.children;
  }
}

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 5 * 60 * 1000,
      retry: 1,
    },
  },
});

type View = 'wheel' | 'natal' | 'transits' | 'synastry' | 'maps' | 'reports' | 'ephemeris' | 'calendar' | 'research' | 'form';

function App() {
  const [activeChart, setActiveChart] = useState<SavedChart | null>(null);
  const [view, setView] = useState<View>('wheel');
  const [showForm, setShowForm] = useState(false);
  const [reportType, setReportType] = useState<'natal' | 'transit' | 'designer'>('natal');
  const [researchTab, setResearchTab] = useState<'compare' | 'draconic' | 'vedic' | 'tools' | 'baselines' | 'batch' | 'search'>('compare');

  const handleSelectChart = useCallback((chart: SavedChart) => {
    setActiveChart(chart);
    setShowForm(false);
    setView('wheel');
  }, []);

  const handleNewChart = useCallback(() => {
    setShowForm(true);
    setActiveChart(null);
  }, []);

  const handleSaveChart = useCallback((chart: SavedChart) => {
    setActiveChart(chart);
    setShowForm(false);
    setView('wheel');
  }, []);

  const handleDeleteChart = useCallback(async (id: number) => {
    await chartDB.remove(id);
    if (activeChart?.id === id) {
      setActiveChart(null);
      setView('wheel');
    }
  }, [activeChart]);

  const tabs: { id: View; label: string }[] = [
    { id: 'wheel', label: 'Wheel' },
    { id: 'natal', label: 'Natal' },
    { id: 'transits', label: 'Transits' },
    { id: 'synastry', label: 'Synastry' },
    { id: 'maps', label: 'Maps' },
    { id: 'reports', label: 'Reports' },
    { id: 'ephemeris', label: 'Ephemeris' },
    { id: 'calendar', label: 'Calendar' },
    { id: 'research', label: 'Research' },
  ];

  return (
    <QueryClientProvider client={queryClient}>
      <div className="flex-1 flex h-screen bg-bg overflow-hidden">
        <Sidebar
          activeChart={activeChart}
          onSelectChart={handleSelectChart}
          onNewChart={handleNewChart}
          onDeleteChart={handleDeleteChart}
        />

        <div className="flex-1 flex flex-col min-w-0" style={{ height: '100vh' }}>
          {showForm ? (
            <ChartForm
              onSave={handleSaveChart}
              onCancel={() => setShowForm(false)}
            />
          ) : activeChart ? (
            <>
              {/* Tab Bar */}
              <div className="flex gap-1 px-4 pt-3 pb-0 border-b border-border shrink-0">
                {tabs.map((tab) => (
                  <button
                    key={tab.id}
                    onClick={() => setView(tab.id)}
                    className={`px-4 py-2 text-sm rounded-t ${
                      view === tab.id
                        ? 'bg-surface text-accent border border-border border-b-0'
                        : 'text-muted hover:text-text'
                    }`}
                  >
                    {tab.label}
                  </button>
                ))}
              </div>

              {/* Content — flex column so children get computed height */}
              <div className="flex-1 flex flex-col overflow-hidden">
                {view === 'wheel' && (
                  <div className="flex-1 overflow-hidden">
                    <ChartWheel data={activeChart.birthData} />
                  </div>
                )}
                {view === 'natal' && (
                  <div className="flex-1 overflow-y-auto p-4">
                    <NatalView data={activeChart.birthData} />
                  </div>
                )}
                {view === 'transits' && (
                  <div className="flex-1 overflow-y-auto p-4">
                    <TransitsView data={activeChart.birthData} />
                  </div>
                )}
                {view === 'synastry' && (
                  <div className="flex-1 overflow-y-auto p-4">
                    <SynastryView chartA={activeChart} />
                  </div>
                )}
                {view === 'maps' && (
                  <div className="flex-1 overflow-hidden p-4">
                    <AstroCartographyMap data={activeChart.birthData} />
                  </div>
                )}
                {view === 'reports' && (
                  <div className="flex-1 overflow-hidden p-4 flex flex-col">
                    <div className="flex gap-2 mb-4 shrink-0">
                      <button onClick={() => setReportType('natal')}
                        className={`px-3 py-1 text-sm rounded ${reportType === 'natal' ? 'bg-accent text-white' : 'bg-surface text-muted border border-border'}`}>
                        Natal Report
                      </button>
                      <button onClick={() => setReportType('transit')}
                        className={`px-3 py-1 text-sm rounded ${reportType === 'transit' ? 'bg-accent text-white' : 'bg-surface text-muted border border-border'}`}>
                        Transit Report
                      </button>
                      <button onClick={() => setReportType('designer')}
                        className={`px-3 py-1 text-sm rounded ${reportType === 'designer' ? 'bg-accent text-white' : 'bg-surface text-muted border border-border'}`}>
                        Page Designer
                      </button>
                    </div>
                    <div className="flex-1 overflow-hidden">
                      {reportType === 'natal'
                        ? <NatalReport data={activeChart.birthData} />
                        : reportType === 'transit'
                        ? <TransitReport data={activeChart.birthData} />
                        : <PageDesigner data={activeChart.birthData} />
                      }
                    </div>
                  </div>
                )}
                {view === 'ephemeris' && (
                  <div className="flex-1 overflow-hidden p-4">
                    <GraphicEphemeris
                      data={activeChart.birthData}
                      startDate={new Date(Date.now() - 15 * 86400000).toISOString().slice(0, 10)}
                      endDate={new Date(Date.now() + 15 * 86400000).toISOString().slice(0, 10)}
                    />
                  </div>
                )}
                {view === 'calendar' && (
                  <div className="flex-1 overflow-y-auto p-4">
                    <TransitCalendar data={activeChart.birthData} />
                  </div>
                )}
                {view === 'research' && (
                  <ErrorBoundary>
                    <div className="flex-1 overflow-hidden p-4 flex flex-col">
                      <div className="flex gap-2 mb-4 shrink-0">
                        {(['compare', 'draconic', 'vedic', 'tools', 'baselines', 'batch', 'search'] as const).map((tab) => (
                          <button
                            key={tab}
                            onClick={() => setResearchTab(tab)}
                            className={`px-3 py-1 text-sm rounded ${
                              researchTab === tab
                                ? 'bg-accent text-white'
                                : 'bg-surface text-muted border border-border'
                            }`}
                          >
                            {tab === 'compare' ? 'Compare' : tab === 'draconic' ? 'Draconic' : tab === 'vedic' ? 'Vedic' : tab === 'tools' ? 'Tools' : tab === 'baselines' ? 'Baselines' : tab === 'batch' ? 'Batch' : 'Search'}
                          </button>
                        ))}
                      </div>
                      <div className="flex-1 overflow-y-auto">
                        {researchTab === 'compare' && <CrossSystemComparison data={activeChart.birthData} />}
                        {researchTab === 'draconic' && <DraconicView data={activeChart.birthData} />}
                        {researchTab === 'vedic' && <VedicView data={activeChart.birthData} />}
                        {researchTab === 'tools' && <ResearchTools data={activeChart.birthData} />}
                        {researchTab === 'baselines' && <EmpiricalBaselines data={activeChart.birthData} />}
                        {researchTab === 'batch' && <BatchAnalysis />}
                        {researchTab === 'search' && <ChartSearch />}
                      </div>
                    </div>
                  </ErrorBoundary>
                )}
              </div>
            </>
          ) : (
            <div className="flex-1 flex items-center justify-center">
              <div className="text-center">
                <p className="text-4xl mb-4">☉</p>
                <h2 className="text-xl text-text mb-2">Empirical Astrology</h2>
                <p className="text-muted text-sm mb-6">
                  Select a chart from the sidebar or create a new one.
                </p>
                <button
                  onClick={handleNewChart}
                  className="bg-accent text-white rounded px-6 py-2 text-sm font-semibold hover:opacity-90"
                >
                  + New Chart
                </button>
              </div>
            </div>
          )}
        </div>
      </div>
    </QueryClientProvider>
  );
}

// ── Natal View ──
function NatalView({ data }: { data: BirthData }) {
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [interp, setInterp] = useState<InterpretationResponse | null>(null);
  const [trad, setTrad] = useState<TraditionalResponse | null>(null);

  useEffect(() => {
    setLoading(true);
    setError('');
    Promise.all([
      api.interpretation(data, 'western', 3),
      api.traditional(data),
    ])
      .then(([i, t]) => {
        setInterp(i);
        setTrad(t);
        setLoading(false);
      })
      .catch((e) => {
        setError(e instanceof Error ? e.message : 'Failed to load');
        setLoading(false);
      });
  }, [data.name, data.year, data.month, data.day, data.hour, data.minute, data.tz_offset, data.lat, data.lng]);

  if (loading) return <p className="text-yellow text-sm">Loading...</p>;
  if (error) return <p className="text-red text-sm">{error}</p>;
  if (!interp) return null;

  return (
    <div className="space-y-4">
      <Section title="Planets in Signs">
        {interp.planet_signs?.map((s, i) => (
          <div key={i} className="text-sm py-0.5">{s}</div>
        ))}
      </Section>
      <Section title="Planets in Houses">
        {interp.planet_houses?.map((s, i) => (
          <div key={i} className="text-sm py-0.5">{s}</div>
        ))}
      </Section>
      <Section title="Aspects">
        {interp.aspects?.map((s, i) => (
          <div key={i} className="text-sm py-0.5">{s}</div>
        ))}
      </Section>
      <Section title="Patterns">
        {interp.patterns?.map((s, i) => (
          <div key={i} className="text-sm py-0.5">{s}</div>
        ))}
      </Section>

      {/* Hidden Contacts */}
      {interp.chart_ruler && (
        <Section title="Chart Ruler">
          <p className="text-sm">{interp.chart_ruler}</p>
        </Section>
      )}
      {interp.final_dispositor && (
        <Section title="Final Dispositor">
          <p className="text-sm">{interp.final_dispositor}</p>
        </Section>
      )}
      {interp.angular_planets && interp.angular_planets.length > 0 && (
        <Section title="Angular Planets">
          {interp.angular_planets.map((s, i) => (
            <div key={i} className="text-sm py-0.5">{s}</div>
          ))}
        </Section>
      )}
      {interp.antiscia && interp.antiscia.length > 0 && (
        <Section title="Antiscia">
          {interp.antiscia.map((s, i) => (
            <div key={i} className="text-sm py-0.5">{s}</div>
          ))}
        </Section>
      )}
      {interp.antiscia_contacts && interp.antiscia_contacts.length > 0 && (
        <Section title="Antiscia Contacts">
          {interp.antiscia_contacts.map((s, i) => (
            <div key={i} className="text-sm py-0.5">{s}</div>
          ))}
        </Section>
      )}
      {interp.declinations && interp.declinations.length > 0 && (
        <Section title="Declination Parallels">
          {interp.declinations.map((s, i) => (
            <div key={i} className="text-sm py-0.5">{s}</div>
          ))}
        </Section>
      )}
      {interp.contraparallels && interp.contraparallels.length > 0 && (
        <Section title="Contraparallels">
          {interp.contraparallels.map((s, i) => (
            <div key={i} className="text-sm py-0.5">{s}</div>
          ))}
        </Section>
      )}
      {interp.key_midpoints && interp.key_midpoints.length > 0 && (
        <Section title="Key Midpoints">
          {interp.key_midpoints.map((s, i) => (
            <div key={i} className="text-sm py-0.5">{s}</div>
          ))}
        </Section>
      )}
      {interp.key_star_aspects && interp.key_star_aspects.length > 0 && (
        <Section title="Star Aspects">
          {interp.key_star_aspects.map((s, i) => (
            <div key={i} className="text-sm py-0.5">{s}</div>
          ))}
        </Section>
      )}
      {interp.weighted_aspects && interp.weighted_aspects.length > 0 && (
        <Section title="Weighted Aspects">
          {interp.weighted_aspects.map((s, i) => (
            <div key={i} className="text-sm py-0.5">{s}</div>
          ))}
        </Section>
      )}

      {/* Lunar Phase & Sect */}
      {trad && (
        <Section title="Traditional">
          <div className="text-sm space-y-1">
            <p>Lunar Phase: {trad.lunar_phase.name} ({trad.lunar_phase.angle_deg.toFixed(1)}° — phase {trad.lunar_phase.phase_index}/8)</p>
            <p>Void of Course Moon: {trad.void_of_course_moon.void_of_course ? 'Yes' : 'No'}</p>
            {trad.retrogrades.filter(r => r.retrograde).length > 0 && (
              <p>Retrograde: {trad.retrogrades.filter(r => r.retrograde).map(r => r.planet).join(', ')}</p>
            )}
          </div>
        </Section>
      )}

      {/* Dispositor Tree */}
      {trad?.dispositor_tree?.nodes && trad.dispositor_tree.nodes.length > 0 && (
        <Section title="Dispositor Tree">
          <div style={{ height: 400 }}>
            <DispositorTree data={trad!.dispositor_tree} />
          </div>
        </Section>
      )}
    </div>
  );
}

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div className="bg-surface border border-border rounded-lg p-4">
      <h3 className="text-sm font-semibold text-muted mb-2">{title}</h3>
      {children}
    </div>
  );
}

// ── Transits View ──
function TransitsView({ data }: { data: BirthData }) {
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [transits, setTransits] = useState<TransitResponse | null>(null);
  const [filterPlanet, setFilterPlanet] = useState('');
  const [filterAspect, setFilterAspect] = useState('');
  const [maxOrb, setMaxOrb] = useState(10);

  useEffect(() => {
    setLoading(true);
    setError('');
    const today = new Date().toISOString().slice(0, 10);
    const end = new Date(Date.now() + 30 * 86400000).toISOString().slice(0, 10);
    api.transits(data, today, end, 3)
      .then((d) => {
        setTransits(d);
        setLoading(false);
      })
      .catch((e) => {
        setError(e instanceof Error ? e.message : 'Failed to load');
        setLoading(false);
      });
  }, [data.name, data.year, data.month, data.day, data.hour, data.minute, data.tz_offset, data.lat, data.lng]);

  if (loading) return <p className="text-yellow text-sm">Loading transits...</p>;
  if (error) return <p className="text-red text-sm">{error}</p>;
  if (!transits) return null;

  const uniquePlanets = [...new Set(transits.transits.flatMap(h => [h.transit_planet, h.natal_planet]))].sort();
  const uniqueAspects = [...new Set(transits.transits.map(h => h.aspect))].sort();

  const filteredHits = transits.transits.filter(h => {
    if (filterPlanet && h.transit_planet !== filterPlanet && h.natal_planet !== filterPlanet) return false;
    if (filterAspect && h.aspect !== filterAspect) return false;
    if (h.orb > maxOrb) return false;
    return true;
  });

  return (
    <div className="space-y-4">
      <Section title={`Transits: ${transits.start_date} to ${transits.end_date}`}>
        {/* Filters */}
        <div className="flex gap-2 mb-4 flex-wrap">
          <select value={filterPlanet} onChange={e => setFilterPlanet(e.target.value)}
            className="bg-surface border border-border rounded px-2 py-1 text-sm text-text">
            <option value="">All Planets</option>
            {uniquePlanets.map(p => <option key={p} value={p}>{p}</option>)}
          </select>
          <select value={filterAspect} onChange={e => setFilterAspect(e.target.value)}
            className="bg-surface border border-border rounded px-2 py-1 text-sm text-text">
            <option value="">All Aspects</option>
            {uniqueAspects.map(a => <option key={a} value={a}>{a}</option>)}
          </select>
          <label className="text-sm text-muted flex items-center gap-1">
            Max orb:
            <input type="number" value={maxOrb} onChange={e => setMaxOrb(Number(e.target.value))}
              className="bg-surface border border-border rounded px-2 py-1 w-16 text-sm text-text" />
            °
          </label>
          <span className="text-xs text-muted self-center ml-auto">
            {filteredHits.length} of {transits.transits.length} hits
          </span>
        </div>

        {filteredHits.length === 0 ? (
          <p className="text-sm text-muted">No transits match the current filters.</p>
        ) : (
          <table className="w-full text-sm">
            <thead>
              <tr className="text-muted text-left">
                <th className="py-1 pr-4">Date</th>
                <th className="py-1 pr-4">Transit</th>
                <th className="py-1 pr-4">Natal</th>
                <th className="py-1 pr-4">Aspect</th>
                <th className="py-1">Orb</th>
              </tr>
            </thead>
            <tbody>
              {filteredHits.map((h, i) => (
                <tr key={i} className="border-t border-border">
                  <td className="py-1 pr-4">{h.date}</td>
                  <td className="py-1 pr-4">{h.transit_planet}</td>
                  <td className="py-1 pr-4">{h.natal_planet}</td>
                  <td className="py-1 pr-4">{h.aspect}</td>
                  <td className="py-1">{h.orb}°</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </Section>

      {/* Sky Weather */}
      {transits.sky_weather && transits.sky_weather.length > 0 && (
        <Section title="Sky Weather (Transit-to-Transit)">
          <table className="w-full text-sm">
            <thead>
              <tr className="text-muted text-left">
                <th className="py-1 pr-4">Date</th>
                <th className="py-1 pr-4">Planet 1</th>
                <th className="py-1 pr-4">Planet 2</th>
                <th className="py-1 pr-4">Aspect</th>
                <th className="py-1">Orb</th>
              </tr>
            </thead>
            <tbody>
              {transits.sky_weather.slice(0, 30).map((h, i) => (
                <tr key={i} className="border-t border-border">
                  <td className="py-1 pr-4">{h.date}</td>
                  <td className="py-1 pr-4">{h.transit_planet}</td>
                  <td className="py-1 pr-4">{h.natal_planet}</td>
                  <td className="py-1 pr-4">{h.aspect}</td>
                  <td className="py-1">{h.orb}°</td>
                </tr>
              ))}
            </tbody>
          </table>
        </Section>
      )}
    </div>
  );
}

export default App;
