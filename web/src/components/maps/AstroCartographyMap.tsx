import { useEffect, useState, useMemo } from 'react';
import { MapContainer, TileLayer, Polyline } from 'react-leaflet';
import type { BirthData } from '../../lib/types';
import { api } from '../../lib/api';
import { planetColor } from '../../lib/astrology';
import type { AstroCartographyResponse } from '../../lib/types';
import 'leaflet/dist/leaflet.css';

export function AstroCartographyMap({ data }: { data: BirthData }) {
  const [response, setResponse] = useState<AstroCartographyResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [visiblePlanets, setVisiblePlanets] = useState<Set<string>>(new Set());

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    setError(null);
    api
      .astrocartography(data)
      .then((res) => {
        if (cancelled) return;
        setResponse(res);
        // Default: all planets visible
        const planets = [...new Set(res.lines.map((l) => l.planet))];
        setVisiblePlanets(new Set(planets));
      })
      .catch((err) => {
        if (cancelled) return;
        setError(err instanceof Error ? err.message : 'Failed to load astrocartography data');
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [data]);

  const uniquePlanets = useMemo(
    () => (response ? [...new Set(response.lines.map((l) => l.planet))] : []),
    [response],
  );

  const togglePlanet = (planet: string) => {
    setVisiblePlanets((prev) => {
      const next = new Set(prev);
      if (next.has(planet)) {
        next.delete(planet);
      } else {
        next.add(planet);
      }
      return next;
    });
  };

  const filteredLines = useMemo(
    () => (response ? response.lines.filter((l) => visiblePlanets.has(l.planet)) : []),
    [response, visiblePlanets],
  );

  if (loading) {
    return (
      <div style={{ minHeight: 400, display: 'flex', alignItems: 'center', justifyContent: 'center', color: '#8b949e' }}>
        Loading astrocartography…
      </div>
    );
  }

  if (error) {
    return (
      <div style={{ minHeight: 400, display: 'flex', alignItems: 'center', justifyContent: 'center', color: '#f85149' }}>
        Error: {error}
      </div>
    );
  }

  if (!response) return null;

  return (
    <div style={{ height: '100%', minHeight: 400, display: 'flex', flexDirection: 'column' }}>
      {/* Planet filter checkboxes */}
      <div style={{ display: 'flex', flexWrap: 'wrap', gap: 8, padding: '8px 0', fontSize: 13 }}>
        {uniquePlanets.map((planet) => (
          <label key={planet} style={{ display: 'flex', alignItems: 'center', gap: 4, cursor: 'pointer', color: '#c9d1d9' }}>
            <input
              type="checkbox"
              checked={visiblePlanets.has(planet)}
              onChange={() => togglePlanet(planet)}
              style={{ accentColor: planetColor(planet) }}
            />
            <span style={{ color: planetColor(planet), fontWeight: 600 }}>{planet}</span>
          </label>
        ))}
      </div>

      {/* Leaflet map */}
      <div style={{ flex: 1, minHeight: 360 }}>
        <MapContainer
          center={[20, 0]}
          zoom={2}
          worldCopyJump
          style={{ height: '100%', width: '100%', minHeight: 360 }}
        >
          <TileLayer
            url="https://{s}.basemaps.cartocdn.com/dark_all/{z}/{x}/{y}{r}.png"
            attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> &copy; <a href="https://carto.com/">CARTO</a>'
          />
          {filteredLines.map((line) => (
            <Polyline
              key={`${line.planet}-${line.angle}`}
              positions={line.points.map((p) => [p.lat, p.lon] as [number, number])}
              pathOptions={{
                color: planetColor(line.planet),
                opacity: 0.7,
                weight: 2,
              }}
            />
          ))}
        </MapContainer>
      </div>
    </div>
  );
}