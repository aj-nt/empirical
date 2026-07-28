import { useState, useEffect } from 'react';
import type {
  BirthData,
  HarmonicResponse,
  ParansResponse,
  ArabicPartsResponse,
  MansionConvergenceResponse,
  StarsCrossResponse,
  ProgressedCrossResponse,
  DivisionalResponse,
} from '../../lib/types';
import { api } from '../../lib/api';

type SubTab =
  | 'harmonic'
  | 'parans'
  | 'declination'
  | 'uranian'
  | 'arabic'
  | 'mansions'
  | 'stars'
  | 'progressed'
  | 'divisional';

const SUB_TABS: { id: SubTab; label: string }[] = [
  { id: 'harmonic', label: 'Harmonic' },
  { id: 'parans', label: 'Parans' },
  { id: 'declination', label: 'Declination' },
  { id: 'uranian', label: 'Uranian' },
  { id: 'arabic', label: 'Arabic Parts' },
  { id: 'mansions', label: 'Mansions' },
  { id: 'stars', label: 'Stars Cross' },
  { id: 'progressed', label: 'Progressed Cross' },
  { id: 'divisional', label: 'Divisional' },
];

export function ResearchTools({ data }: { data: BirthData }) {
  const [subTab, setSubTab] = useState<SubTab>('harmonic');

  return (
    <div className="space-y-4">
      {/* Sub-tab bar */}
      <div className="flex gap-1 flex-wrap">
        {SUB_TABS.map((tab) => (
          <button
            key={tab.id}
            onClick={() => setSubTab(tab.id)}
            className={`px-3 py-1 text-sm rounded ${
              subTab === tab.id
                ? 'bg-accent text-white'
                : 'bg-surface text-muted border border-border hover:text-text'
            }`}
          >
            {tab.label}
          </button>
        ))}
      </div>

      {/* Sub-tab content */}
      {subTab === 'harmonic' && <HarmonicTab data={data} />}
      {subTab === 'parans' && <ParansTab data={data} />}
      {subTab === 'declination' && <DeclinationTab data={data} />}
      {subTab === 'uranian' && <UranianTab data={data} />}
      {subTab === 'arabic' && <ArabicTab data={data} />}
      {subTab === 'mansions' && <MansionsTab data={data} />}
      {subTab === 'stars' && <StarsCrossTab data={data} />}
      {subTab === 'progressed' && <ProgressedCrossTab data={data} />}
      {subTab === 'divisional' && <DivisionalTab data={data} />}
    </div>
  );
}

// ── Shared helpers ──

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div className="bg-surface border border-border rounded-lg p-4">
      <h3 className="text-sm font-semibold text-muted mb-2">{title}</h3>
      {children}
    </div>
  );
}

function useApi<T>(fetcher: () => Promise<T>, deps: unknown[]) {
  const [result, setResult] = useState<T | null>(null);
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    setError('');
    setResult(null);

    fetcher()
      .then((data) => {
        if (!cancelled) {
          setResult(data);
          setLoading(false);
        }
      })
      .catch((err) => {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : String(err));
          setLoading(false);
        }
      });

    return () => {
      cancelled = true;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, deps);

  return { result, error, loading };
}

// ── 1. Harmonic ──

function HarmonicTab({ data }: { data: BirthData }) {
  const { result, error, loading } = useApi<HarmonicResponse>(
    () => api.harmonic(data),
    [data.name, data.year, data.month, data.day, data.hour, data.minute, data.tz_offset, data.lat, data.lng],
  );

  if (loading) return <p className="text-yellow text-sm">Loading...</p>;
  if (error) return <p className="text-red text-sm">{error}</p>;
  if (!result) return null;

  return (
    <div className="space-y-4">
      <Section title={`Harmonics: ${result.harmonics.map(h => `H${h.harmonic}`).join(', ')}`}>
        {result.harmonics.map((entry) => (
          <div key={entry.harmonic} className="mb-4">
            <h4 className="text-sm font-semibold text-accent mb-2">
              H{entry.harmonic} — {entry.aspect_name}
            </h4>
            {entry.conjunctions?.length > 0 ? (
              <table className="w-full text-sm">
                <thead>
                  <tr className="text-muted text-left">
                    <th className="py-1 pr-4">Planet A</th>
                    <th className="py-1 pr-4">Planet B</th>
                    <th className="py-1">Orb</th>
                  </tr>
                </thead>
                <tbody>
                  {entry.conjunctions.map((c, i) => (
                    <tr key={i} className="border-t border-border">
                      <td className="py-1 pr-4">{c.planet_a}</td>
                      <td className="py-1 pr-4">{c.planet_b}</td>
                      <td className="py-1">{c.orb != null ? c.orb.toFixed(2) + '°' : '—'}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            ) : (
              <p className="text-muted text-xs">No conjunctions within orb</p>
            )}
          </div>
        ))}
      </Section>
    </div>
  );
}

// ── 2. Parans ──

function ParansTab({ data }: { data: BirthData }) {
  const { result, error, loading } = useApi<ParansResponse>(
    () => api.parans(data),
    [data.name, data.year, data.month, data.day, data.hour, data.minute, data.tz_offset, data.lat, data.lng],
  );

  if (loading) return <p className="text-yellow text-sm">Loading...</p>;
  if (error) return <p className="text-red text-sm">{error}</p>;
  if (!result) return null;

  return (
    <div className="space-y-4">
      <Section title="Stars on Angles">
        {result.stars_on_angles?.length === 0 ? (
          <p className="text-sm text-muted">No stars on angles found.</p>
        ) : (
          <table className="w-full text-sm">
            <thead>
              <tr className="text-muted text-left">
                <th className="py-1 pr-4">Star</th>
                <th className="py-1 pr-4">Angle</th>
                <th className="py-1">Orb</th>
              </tr>
            </thead>
            <tbody>
              {result.stars_on_angles?.map((p, i) => (
                <tr key={i} className="border-t border-border">
                  <td className="py-1 pr-4">{p.body}</td>
                  <td className="py-1 pr-4">{p.angle}</td>
                  <td className="py-1">{p.orb?.toFixed(2)}°</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </Section>

      <Section title="Planets on Angles">
        {result.planets_on_angles?.length === 0 ? (
          <p className="text-sm text-muted">No planets on angles found.</p>
        ) : (
          <table className="w-full text-sm">
            <thead>
              <tr className="text-muted text-left">
                <th className="py-1 pr-4">Planet</th>
                <th className="py-1 pr-4">Angle</th>
                <th className="py-1">Orb</th>
              </tr>
            </thead>
            <tbody>
              {result.planets_on_angles?.map((p, i) => (
                <tr key={i} className="border-t border-border">
                  <td className="py-1 pr-4">{p.body}</td>
                  <td className="py-1 pr-4">{p.angle}</td>
                  <td className="py-1">{p.orb?.toFixed(2)}°</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </Section>

      <Section title="Parans">
        {result.parans?.length === 0 ? (
          <p className="text-sm text-muted">No parans found.</p>
        ) : (
          <table className="w-full text-sm">
            <thead>
              <tr className="text-muted text-left">
                <th className="py-1 pr-4">Star</th>
                <th className="py-1 pr-4">Planet</th>
                <th className="py-1 pr-4">Angle</th>
                <th className="py-1">Star Orb</th>
                <th className="py-1">Planet Orb</th>
              </tr>
            </thead>
            <tbody>
              {result.parans?.map((p, i) => (
                <tr key={i} className="border-t border-border">
                  <td className="py-1 pr-4">{p.star}</td>
                  <td className="py-1 pr-4">{p.planet}</td>
                  <td className="py-1 pr-4">{p.angle}</td>
                  <td className="py-1">{p.star_orb?.toFixed(2)}°</td>
                  <td className="py-1">{p.planet_orb?.toFixed(2)}°</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </Section>
    </div>
  );
}

// ── 3. Declination ──

interface DeclinationRow {
  body: string;
  longitude: number;
  latitude: number;
  declination: number;
  hemisphere: string;
}

interface DeclinationParallel {
  body_a: string;
  body_b: string;
  decl_a: number;
  decl_b: number;
  orb: number;
  type: string;
}

interface DeclinationFullResponse {
  name: string;
  declinations: DeclinationRow[];
  parallels: DeclinationParallel[];
}

function DeclinationTab({ data }: { data: BirthData }) {
  const { result, error, loading } = useApi<DeclinationFullResponse>(
    () => api.declination(data) as unknown as Promise<DeclinationFullResponse>,
    [data.name, data.year, data.month, data.day, data.hour, data.minute, data.tz_offset, data.lat, data.lng],
  );

  if (loading) return <p className="text-yellow text-sm">Loading...</p>;
  if (error) return <p className="text-red text-sm">{error}</p>;
  if (!result) return null;

  const declinations = result.declinations ?? [];
  const parallels = result.parallels ?? [];

  return (
    <div className="space-y-4">
      <Section title="Declinations">
        {declinations.length > 0 ? (
          <table className="w-full text-sm">
            <thead>
              <tr className="text-muted text-left">
                <th className="py-1 pr-4">Body</th>
                <th className="py-1 pr-4">Declination</th>
                <th className="py-1">Hemisphere</th>
              </tr>
            </thead>
            <tbody>
              {declinations.map((d, i) => (
                <tr key={i} className="border-t border-border">
                  <td className="py-1 pr-4">{d.body}</td>
                  <td className="py-1 pr-4">{d.declination.toFixed(2)}°</td>
                  <td className="py-1 text-muted">{d.hemisphere}</td>
                </tr>
              ))}
            </tbody>
          </table>
        ) : (
          <p className="text-sm text-muted">No declination data available.</p>
        )}
      </Section>

      <Section title="Parallels">
        {parallels.length > 0 ? (
          <table className="w-full text-sm">
            <thead>
              <tr className="text-muted text-left">
                <th className="py-1 pr-4">Body A</th>
                <th className="py-1 pr-4">Body B</th>
                <th className="py-1 pr-4">Type</th>
                <th className="py-1">Orb</th>
              </tr>
            </thead>
            <tbody>
              {parallels.map((p, i) => (
                <tr key={i} className="border-t border-border">
                  <td className="py-1 pr-4">{p.body_a}</td>
                  <td className="py-1 pr-4">{p.body_b}</td>
                  <td className="py-1 pr-4 text-muted">{p.type}</td>
                  <td className="py-1">{p.orb.toFixed(2)}°</td>
                </tr>
              ))}
            </tbody>
          </table>
        ) : (
          <p className="text-sm text-muted">No parallels found.</p>
        )}
      </Section>
    </div>
  );
}

// ── 4. Uranian ──

interface UranianDialPosition {
  name: string;
  dial_deg: number;
}

interface UranianDirectMidpoint {
  pair_a: string;
  pair_b: string;
  planet: string;
  orb: number;
}

interface UranianMidpointPicture {
  factor_a: string;
  factor_b: string;
  activator: string;
  harmonic: string;
  offset: number;
  orb: number;
}

interface UranianActivation {
  factor_a: string;
  factor_b: string;
  factor_c: string;
  picture_lon: number;
  target_lon: number;
  orb: number;
}

interface UranianFullResponse {
  name: string;
  dial_positions: UranianDialPosition[];
  direct_midpoints: UranianDirectMidpoint[];
  midpoint_pictures: UranianMidpointPicture[];
  tight_pictures: UranianMidpointPicture[];
  activations: UranianActivation[];
}

function UranianTab({ data }: { data: BirthData }) {
  const { result, error, loading } = useApi<UranianFullResponse>(
    () => api.uranian(data) as unknown as Promise<UranianFullResponse>,
    [data.name, data.year, data.month, data.day, data.hour, data.minute, data.tz_offset, data.lat, data.lng],
  );

  if (loading) return <p className="text-yellow text-sm">Loading...</p>;
  if (error) return <p className="text-red text-sm">{error}</p>;
  if (!result) return null;

  return (
    <div className="space-y-4">
      {result.tight_pictures.length > 0 && (
        <Section title="Tight Midpoint Pictures">
          <table className="w-full text-sm">
            <thead>
              <tr className="text-muted text-left">
                <th className="py-1 pr-4">Factor A</th>
                <th className="py-1 pr-4">Factor B</th>
                <th className="py-1 pr-4">Activator</th>
                <th className="py-1 pr-4">Harmonic</th>
                <th className="py-1">Orb</th>
              </tr>
            </thead>
            <tbody>
              {result.tight_pictures.map((tp, i) => (
                <tr key={i} className="border-t border-border">
                  <td className="py-1 pr-4">{tp.factor_a}</td>
                  <td className="py-1 pr-4">{tp.factor_b}</td>
                  <td className="py-1 pr-4">{tp.activator}</td>
                  <td className="py-1 pr-4 text-muted">{tp.harmonic}</td>
                  <td className="py-1">{tp.orb.toFixed(2)}°</td>
                </tr>
              ))}
            </tbody>
          </table>
        </Section>
      )}

      {result.midpoint_pictures.length > 0 && (
        <Section title="Midpoint Pictures">
          <table className="w-full text-sm">
            <thead>
              <tr className="text-muted text-left">
                <th className="py-1 pr-4">Factor A</th>
                <th className="py-1 pr-4">Factor B</th>
                <th className="py-1 pr-4">Activator</th>
                <th className="py-1 pr-4">Harmonic</th>
                <th className="py-1">Orb</th>
              </tr>
            </thead>
            <tbody>
              {result.midpoint_pictures.map((mp, i) => (
                <tr key={i} className="border-t border-border">
                  <td className="py-1 pr-4">{mp.factor_a}</td>
                  <td className="py-1 pr-4">{mp.factor_b}</td>
                  <td className="py-1 pr-4">{mp.activator}</td>
                  <td className="py-1 pr-4 text-muted">{mp.harmonic}</td>
                  <td className="py-1">{mp.orb.toFixed(2)}°</td>
                </tr>
              ))}
            </tbody>
          </table>
        </Section>
      )}

      {result.direct_midpoints.length > 0 && (
        <Section title="Direct Midpoints">
          <table className="w-full text-sm">
            <thead>
              <tr className="text-muted text-left">
                <th className="py-1 pr-4">Pair A</th>
                <th className="py-1 pr-4">Pair B</th>
                <th className="py-1 pr-4">Planet</th>
                <th className="py-1">Orb</th>
              </tr>
            </thead>
            <tbody>
              {result.direct_midpoints.map((dm, i) => (
                <tr key={i} className="border-t border-border">
                  <td className="py-1 pr-4">{dm.pair_a}</td>
                  <td className="py-1 pr-4">{dm.pair_b}</td>
                  <td className="py-1 pr-4">{dm.planet}</td>
                  <td className="py-1">{dm.orb.toFixed(4)}°</td>
                </tr>
              ))}
            </tbody>
          </table>
        </Section>
      )}

      {result.dial_positions.length > 0 && (
        <Section title="Dial Positions">
          <table className="w-full text-sm">
            <thead>
              <tr className="text-muted text-left">
                <th className="py-1 pr-4">Point</th>
                <th className="py-1">Degree</th>
              </tr>
            </thead>
            <tbody>
              {result.dial_positions.map((dp, i) => (
                <tr key={i} className="border-t border-border">
                  <td className="py-1 pr-4">{dp.name}</td>
                  <td className="py-1">{dp.dial_deg.toFixed(1)}°</td>
                </tr>
              ))}
            </tbody>
          </table>
        </Section>
      )}

      {result.activations.length > 0 && (
        <Section title="Activations">
          <table className="w-full text-sm">
            <thead>
              <tr className="text-muted text-left">
                <th className="py-1 pr-4">Factor A</th>
                <th className="py-1 pr-4">Factor B</th>
                <th className="py-1 pr-4">Factor C</th>
                <th className="py-1">Orb</th>
              </tr>
            </thead>
            <tbody>
              {result.activations.map((a, i) => (
                <tr key={i} className="border-t border-border">
                  <td className="py-1 pr-4">{a.factor_a}</td>
                  <td className="py-1 pr-4">{a.factor_b}</td>
                  <td className="py-1 pr-4">{a.factor_c}</td>
                  <td className="py-1">{a.orb.toFixed(4)}°</td>
                </tr>
              ))}
            </tbody>
          </table>
        </Section>
      )}
    </div>
  );
}

// ── 5. Arabic Parts ──

interface ArabicPartEntry {
  part: string;
  lon: number;
  sign: string;
  sign_num: number;
}

function ArabicTab({ data }: { data: BirthData }) {
  const { result, error, loading } = useApi<ArabicPartsResponse>(
    () => api.arabicParts(data),
    [data.name, data.year, data.month, data.day, data.hour, data.minute, data.tz_offset, data.lat, data.lng],
  );

  if (loading) return <p className="text-yellow text-sm">Loading...</p>;
  if (error) return <p className="text-red text-sm">{error}</p>;
  if (!result) return null;

  const tropical = (result as any).tropical ?? [];
  const sidereal = (result as any).sidereal ?? [];
  const survivors = (result as any).sign_survivors ?? 0;

  // Build a lookup: part name -> sidereal sign
  const siderealByPart: Record<string, string> = {};
  for (const s of sidereal) {
    siderealByPart[s.part] = s.sign;
  }

  return (
    <div className="space-y-4">
      <Section title={`Arabic Parts (${tropical.length} total, ${survivors} sign survivors)`}>
        <table className="w-full text-sm">
          <thead>
            <tr className="text-muted text-left">
              <th className="py-1 pr-4">Part Name</th>
              <th className="py-1 pr-4">Tropical</th>
              <th className="py-1">Sidereal</th>
            </tr>
          </thead>
          <tbody>
            {tropical.map((p: ArabicPartEntry, i: number) => {
              const siderealSign = siderealByPart[p.part];
              const isSurvivor = siderealSign === p.sign;
              return (
                <tr key={i} className="border-t border-border">
                  <td className={`py-1 pr-4 ${isSurvivor ? 'text-green-400 font-semibold' : ''}`}>
                    {p.part}
                  </td>
                  <td className="py-1 pr-4">{p.sign}</td>
                  <td className="py-1">{siderealSign ?? '—'}</td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </Section>

      {result.aspects && result.aspects.length > 0 && (
        <Section title="Part Aspects">
          <table className="w-full text-sm">
            <thead>
              <tr className="text-muted text-left">
                <th className="py-1 pr-4">Part</th>
                <th className="py-1 pr-4">Planet</th>
                <th className="py-1 pr-4">Aspect</th>
                <th className="py-1">Orb</th>
              </tr>
            </thead>
            <tbody>
              {result.aspects.map((a: any, i: number) => (
                <tr key={i} className="border-t border-border">
                  <td className="py-1 pr-4">{a.part}</td>
                  <td className="py-1 pr-4">{a.planet}</td>
                  <td className="py-1 pr-4">{a.aspect}</td>
                  <td className="py-1">{a.orb.toFixed(2)}°</td>
                </tr>
              ))}
            </tbody>
          </table>
        </Section>
      )}
    </div>
  );
}

// ── 6. Mansions ──

function MansionsTab({ data }: { data: BirthData }) {
  const { result, error, loading } = useApi<MansionConvergenceResponse>(
    () => api.mansionConvergence(data),
    [data.name, data.year, data.month, data.day, data.hour, data.minute, data.tz_offset, data.lat, data.lng],
  );

  if (loading) return <p className="text-yellow text-sm">Loading...</p>;
  if (error) return <p className="text-red text-sm">{error}</p>;
  if (!result) return null;

  const planets = (result as any).planets ?? [];
  const converging = planets.filter((p: any) => p.converges);

  return (
    <div className="space-y-4">
      <Section title={`Mansion Convergence (${converging.length}/${planets.length} converging)`}>
        {planets.length === 0 ? (
          <p className="text-sm text-muted">No data.</p>
        ) : (
          <table className="w-full text-sm">
            <thead>
              <tr className="text-muted text-left">
                <th className="py-1 pr-4">Planet</th>
                <th className="py-1 pr-4">Nakshatra</th>
                <th className="py-1 pr-4">Xiu</th>
                <th className="py-1">Converges?</th>
              </tr>
            </thead>
            <tbody>
              {planets.map((p: any, i: number) => (
                <tr key={i} className={`border-t border-border ${p.converges ? 'bg-accent/5' : ''}`}>
                  <td className="py-1 pr-4">{p.planet}</td>
                  <td className="py-1 pr-4">{p.nakshatra}</td>
                  <td className="py-1 pr-4">{p.xiu} ({p.xiu_pinyin})</td>
                  <td className={`py-1 ${p.converges ? 'text-green-400 font-semibold' : 'text-muted'}`}>
                    {p.converges ? 'Yes' : 'No'}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </Section>
    </div>
  );
}

// ── 7. Stars Cross ──

function StarsCrossTab({ data }: { data: BirthData }) {
  const { result, error, loading } = useApi<StarsCrossResponse>(
    () => api.starsCross(data),
    [data.name, data.year, data.month, data.day, data.hour, data.minute, data.tz_offset, data.lat, data.lng],
  );

  if (loading) return <p className="text-yellow text-sm">Loading...</p>;
  if (error) return <p className="text-red text-sm">{error}</p>;
  if (!result) return null;

  const renderConjunctions = (title: string, items: any[], highlight = false) => (
    <div>
      <h4 className="text-sm font-semibold text-text mb-2">
        {title}
        <span className="text-muted font-normal ml-2">({items.length})</span>
      </h4>
      {items.length === 0 ? (
        <p className="text-sm text-muted">None found.</p>
      ) : (
        <table className="w-full text-sm">
          <thead>
            <tr className="text-muted text-left">
              <th className="py-1 pr-4">Star</th>
              <th className="py-1 pr-4">Planet</th>
              <th className="py-1">Orb</th>
            </tr>
          </thead>
          <tbody>
            {items.map((c, i) => (
              <tr key={i} className={`border-t border-border ${highlight ? 'text-green-400' : ''}`}>
                <td className="py-1 pr-4">{c.star}</td>
                <td className="py-1 pr-4">{c.planet}</td>
                <td className="py-1">{c.orb.toFixed(2)}°</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  );

  const totalTropical = (result as any).total_tropical ?? 0;
  const totalSurvivors = (result as any).total_survivors ?? 0;
  const survivalRate = totalTropical > 0 ? (totalSurvivors / totalTropical * 100).toFixed(0) : null;

  return (
    <div className="space-y-4">
      {survivalRate !== null && (
        <Section title="Survival Rate">
          <p className="text-sm">
            <span className="font-semibold text-green-400">{survivalRate}%</span> of star conjunctions survive cross-system verification.
          </p>
        </Section>
      )}

      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <Section title="Tropical">
          {renderConjunctions('Tropical Conjunctions', (result as any).tropical ?? [])}
        </Section>

        <Section title="Survivors ✓">
          {renderConjunctions('Cross-System Survivors', (result as any).survivors ?? [], true)}
        </Section>

        <Section title="Sidereal">
          {renderConjunctions('Sidereal Conjunctions', (result as any).sidereal ?? [])}
        </Section>
      </div>
    </div>
  );
}

// ── 8. Progressed Cross ──

function ProgressedCrossTab({ data }: { data: BirthData }) {
  const today = new Date().toISOString().slice(0, 10);
  const { result, error, loading } = useApi<ProgressedCrossResponse>(
    () => api.progressedCross(data, today),
    [data.name, data.year, data.month, data.day, data.hour, data.minute, data.tz_offset, data.lat, data.lng],
  );

  if (loading) return <p className="text-yellow text-sm">Loading...</p>;
  if (error) return <p className="text-red text-sm">{error}</p>;
  if (!result) return null;

  const renderAspects = (title: string, aspects: any[], highlight = false) => (
    <div>
      <h4 className="text-sm font-semibold text-text mb-2">
        {title}
        <span className="text-muted font-normal ml-2">({aspects.length})</span>
      </h4>
      {aspects.length === 0 ? (
        <p className="text-sm text-muted">None found.</p>
      ) : (
        <table className="w-full text-sm">
          <thead>
            <tr className="text-muted text-left">
              <th className="py-1 pr-4">Progressed</th>
              <th className="py-1 pr-4">Natal</th>
              <th className="py-1 pr-4">Aspect</th>
              <th className="py-1">Orb</th>
            </tr>
          </thead>
          <tbody>
            {aspects.map((a, i) => (
              <tr key={i} className={`border-t border-border ${highlight ? 'text-green-400' : ''}`}>
                <td className="py-1 pr-4">{a.progressed_planet}</td>
                <td className="py-1 pr-4">{a.natal_planet}</td>
                <td className="py-1 pr-4">{a.aspect}</td>
                <td className="py-1">{a.orb.toFixed(2)}°</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  );

  const survivors = (result as any).survivors ?? [];
  const tropicalOnly = (result as any).tropical_only ?? [];
  const siderealOnly = (result as any).sidereal_only ?? [];

  return (
    <div className="space-y-4">
      <Section title={`Progressed Chart: ${result.target_date} (age ${result.age_years.toFixed(1)})`}>
        <p className="text-sm text-muted">
          {survivors.length} cross-system survivors
        </p>
      </Section>

      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <Section title="Survivors ✓">
          {renderAspects('Cross-System Survivors', survivors, true)}
        </Section>

        <Section title="Tropical Only">
          {renderAspects('Only in Tropical', tropicalOnly)}
        </Section>

        <Section title="Sidereal Only">
          {renderAspects('Only in Sidereal', siderealOnly)}
        </Section>
      </div>
    </div>
  );
}

// ── 9. Divisional ──

function DivisionalTab({ data }: { data: BirthData }) {
  const { result, error, loading } = useApi<DivisionalResponse>(
    () => api.divisional(data),
    [data.name, data.year, data.month, data.day, data.hour, data.minute, data.tz_offset, data.lat, data.lng],
  );

  if (loading) return <p className="text-yellow text-sm">Loading...</p>;
  if (error) return <p className="text-red text-sm">{error}</p>;
  if (!result) return null;

  const positions = (result as any).positions ?? [];
  const dasha = (result as any).dasha ?? [];

  return (
    <div className="space-y-4">
      {/* Dasha info */}
      {dasha.length > 0 && (() => {
        const today = new Date().toISOString().slice(0, 10);
        const currentIdx = dasha.findIndex((d: any) => d.start <= today && d.end > today);
        const startIdx = currentIdx >= 0 ? currentIdx : 0;
        return (
        <Section title="Current Dasha">
          <table className="w-full text-sm">
            <thead>
              <tr className="text-muted text-left">
                <th className="py-1 pr-4">Planet</th>
                <th className="py-1 pr-4">Start</th>
                <th className="py-1">End</th>
              </tr>
            </thead>
            <tbody>
              {dasha.slice(startIdx, startIdx + 3).map((d: any, i: number) => (
                <tr key={i} className="border-t border-border">
                  <td className={`py-1 pr-4 ${i === 0 ? 'font-semibold text-accent' : ''}`}>{d.planet}</td>
                  <td className="py-1 pr-4">{d.start}</td>
                  <td className="py-1">{d.end}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </Section>
        );
      })()}

      {/* Navamsha (D9) */}
      {positions.length > 0 && (
        <Section title="Navamsha (D9)">
          <table className="w-full text-sm">
            <thead>
              <tr className="text-muted text-left">
                <th className="py-1 pr-4">Planet</th>
                <th className="py-1 pr-4">Sign</th>
                <th className="py-1 pr-4">Navamsha Sign</th>
                <th className="py-1">Navamsha Lon</th>
              </tr>
            </thead>
            <tbody>
              {positions.map((p: any, i: number) => (
                <tr key={i} className="border-t border-border">
                  <td className="py-1 pr-4">{p.planet}</td>
                  <td className="py-1 pr-4">{p.sidereal_sign}</td>
                  <td className="py-1 pr-4">{p.navamsha_sign}</td>
                  <td className="py-1">{p.navamsha_lon.toFixed(1)}°</td>
                </tr>
              ))}
            </tbody>
          </table>
        </Section>
      )}

      {/* Nakshatras */}
      {positions.length > 0 && (
        <Section title="Nakshatras">
          <table className="w-full text-sm">
            <thead>
              <tr className="text-muted text-left">
                <th className="py-1 pr-4">Planet</th>
                <th className="py-1 pr-4">Nakshatra</th>
                <th className="py-1 pr-4">Pada</th>
                <th className="py-1">Ruler</th>
              </tr>
            </thead>
            <tbody>
              {positions.map((p: any, i: number) => (
                <tr key={i} className="border-t border-border">
                  <td className="py-1 pr-4">{p.planet}</td>
                  <td className="py-1 pr-4">{p.nakshatra.nakshatra}</td>
                  <td className="py-1 pr-4">{p.nakshatra.pada}</td>
                  <td className="py-1">{p.nakshatra.ruler}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </Section>
      )}
    </div>
  );
}