import { useState, useEffect } from 'react';
import type { BirthData, DraconicResponse } from '../../lib/types';
import { api } from '../../lib/api';

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div className="bg-surface border border-border rounded-lg p-4">
      <h3 className="text-sm font-semibold text-muted mb-2">{title}</h3>
      {children}
    </div>
  );
}

const SIGN_NAMES = [
  'Aries', 'Taurus', 'Gemini', 'Cancer', 'Leo', 'Virgo',
  'Libra', 'Scorpio', 'Sagittarius', 'Capricorn', 'Aquarius', 'Pisces',
];

function lonToSign(lon: number): string {
  return SIGN_NAMES[Math.floor(lon / 30) % 12];
}

export function DraconicView({ data }: { data: BirthData }) {
  const [result, setResult] = useState<DraconicResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    setError(null);

    api
      .draconic(data)
      .then((res) => {
        if (!cancelled) setResult(res);
      })
      .catch((err) => {
        if (!cancelled) setError(err instanceof Error ? err.message : String(err));
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });

    return () => { cancelled = true; };
  }, [data.name, data.year, data.month, data.day, data.hour, data.minute, data.tz_offset, data.lat, data.lng]);

  if (loading) return <p className="text-yellow text-sm">Loading draconic chart…</p>;
  if (error) return <p className="text-red text-sm">Error: {error}</p>;
  if (!result) return null;

  const shiftedPlanets = new Set(result.sign_shifts.map(s => s.planet));

  return (
    <div className="space-y-4">
      <Section title={`Draconic Chart — ${result.name}`}>
        <p className="text-sm text-muted mb-4">NN offset: {result.offset.toFixed(2)}°</p>

        <table className="w-full text-sm mb-4">
          <thead>
            <tr className="text-muted text-left">
              <th className="py-1 pr-4">Planet</th>
              <th className="py-1 pr-4">Tropical Sign</th>
              <th className="py-1 pr-4">Draconic Sign</th>
              <th className="py-1">Shifted?</th>
            </tr>
          </thead>
          <tbody>
            {Object.entries(result.planets).map(([planet, lon]) => {
              const tropicalSign = lonToSign(lon);
              const shift = result.sign_shifts.find(s => s.planet === planet);
              const draconicSign = shift ? shift.draconic_sign : tropicalSign;
              const shifted = shiftedPlanets.has(planet);
              return (
                <tr key={planet} className={`border-t border-border ${shifted ? 'bg-accent/5' : ''}`}>
                  <td className="py-1 pr-4">{planet}</td>
                  <td className="py-1 pr-4">{tropicalSign}</td>
                  <td className="py-1 pr-4">{draconicSign}</td>
                  <td className={`py-1 ${shifted ? 'text-accent font-semibold' : ''}`}>
                    {shifted ? '✦ Yes' : 'No'}
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </Section>

      {/* Sign Shifts */}
      {result.sign_shifts.length > 0 && (
        <Section title="Sign Shifts">
          <div className="space-y-1">
            {result.sign_shifts.map((s) => (
              <div key={s.planet} className="text-sm py-0.5">
                <span className="text-accent font-semibold">{s.planet}</span>
                : <span className="text-muted">{s.tropical_sign}</span>
                {' → '}
                <span className="text-accent">{s.draconic_sign}</span>
              </div>
            ))}
          </div>
        </Section>
      )}

      {/* Bridges */}
      {result.bridges.length > 0 && (
        <Section title="Bridge Aspects">
          <div className="space-y-1">
            {result.bridges.map((b, i) => (
              <div key={i} className="text-sm py-0.5">
                {b.planet1} {b.aspect} {b.planet2} ({b.orb.toFixed(1)}°)
              </div>
            ))}
          </div>
        </Section>
      )}
    </div>
  );
}
