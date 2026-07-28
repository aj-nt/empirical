import { useState, useEffect } from 'react';
import type { BirthData, VedicNatalReport } from '../../lib/types';
import { api } from '../../lib/api';

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div className="bg-surface border border-border rounded-lg p-4">
      <h3 className="text-sm font-semibold text-muted mb-2">{title}</h3>
      {children}
    </div>
  );
}

export function VedicView({ data }: { data: BirthData }) {
  const [result, setResult] = useState<VedicNatalReport | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    setError(null);

    api
      .vedic(data)
      .then((res) => {
        if (!cancelled) setResult(res);
      })
      .catch((err: Error) => {
        if (!cancelled) setError(err instanceof Error ? err.message : String(err));
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });

    return () => { cancelled = true; };
  }, [data.name, data.year, data.month, data.day, data.hour, data.minute, data.tz_offset, data.lat, data.lng]);

  if (loading) return <p className="text-yellow text-sm">Loading…</p>;
  if (error) return <p className="text-red text-sm">{error}</p>;
  if (!result) return null;

  return (
    <div className="space-y-4">
      <Section title="Ascendant">
        <div className="text-sm space-y-1">
          <p><span className="text-muted">Sign:</span> {result.ascendant.sidereal_sign}</p>
          <p><span className="text-muted">Nakshatra:</span> {result.ascendant.nakshatra} Pada {result.ascendant.nakshatra_pada}</p>
          <p><span className="text-muted">Ruler:</span> {result.ascendant.nakshatra_ruler}</p>
          <p><span className="text-muted">Ayanamsa:</span> {result.ayanamsa.toFixed(4)}°</p>
        </div>
      </Section>

      <Section title="Planets">
        <table className="w-full text-sm">
          <thead>
            <tr className="text-muted text-left">
              <th className="py-1 pr-2">Planet</th>
              <th className="py-1 pr-2">Sign</th>
              <th className="py-1 pr-2">House</th>
              <th className="py-1 pr-2">Nakshatra</th>
              <th className="py-1 pr-2">Pada</th>
              <th className="py-1 pr-2">Navamsha</th>
              <th className="py-1 pr-2">Dignity</th>
              <th className="py-1">R/C</th>
            </tr>
          </thead>
          <tbody>
            {result.planets.map((p, i) => (
              <tr key={i} className="border-t border-border">
                <td className="py-1 pr-2">{p.planet}</td>
                <td className="py-1 pr-2">{p.sidereal_sign}</td>
                <td className="py-1 pr-2">{p.house}</td>
                <td className="py-1 pr-2">{p.nakshatra}</td>
                <td className="py-1 pr-2">{p.nakshatra_pada}</td>
                <td className="py-1 pr-2">{p.navamsha_sign}</td>
                <td className="py-1 pr-2">{p.dignity}</td>
                <td className="py-1">
                  {p.retrograde && <span className="text-yellow">R</span>}
                  {p.combust && <span className="text-red">C</span>}
                  {!p.retrograde && !p.combust && '—'}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </Section>

      <Section title="Dasha">
        <table className="w-full text-sm">
          <thead>
            <tr className="text-muted text-left">
              <th className="py-1 pr-4">Planet</th>
              <th className="py-1 pr-4">Start</th>
              <th className="py-1">End</th>
            </tr>
          </thead>
          <tbody>
            {result.dasha.map((d, i) => (
              <tr key={i} className="border-t border-border">
                <td className="py-1 pr-4">{d.planet}</td>
                <td className="py-1 pr-4">{d.start}</td>
                <td className="py-1">{d.end}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </Section>

      <p className="text-xs text-muted">Signal count: {result.signal_count}</p>
    </div>
  );
}
