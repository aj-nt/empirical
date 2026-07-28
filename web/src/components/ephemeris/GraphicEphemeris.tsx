import { useEffect, useRef, useState } from 'react';
import * as d3 from 'd3';
import type { BirthData, TransitResponse } from '../../lib/types';
import { api } from '../../lib/api';
import { planetColor, planetGlyph, sortPlanets } from '../../lib/astrology';

interface GraphicEphemerisProps {
  data: BirthData;
  startDate: string;
  endDate: string;
}

const MARGIN = { top: 30, right: 30, bottom: 50, left: 60 };
const Y_DOMAIN = 45; // 45° modulus

// Simple hash function to spread planets across Y-axis positions
function planetHash(name: string): number {
  let h = 0;
  for (let i = 0; i < name.length; i++) {
    h = ((h << 5) - h + name.charCodeAt(i)) | 0;
  }
  return Math.abs(h);
}

export function GraphicEphemeris({ data, startDate, endDate }: GraphicEphemerisProps) {
  const svgRef = useRef<SVGSVGElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  const [transitData, setTransitData] = useState<TransitResponse | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [visiblePlanets, setVisiblePlanets] = useState<Record<string, boolean>>({});

  // Fetch transit data
  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    setError(null);

    api
      .transits(data, startDate, endDate, 3)
      .then((res) => {
        if (cancelled) return;
        setTransitData(res);

        // Initialize all planets as visible
        const planets = new Set<string>();
        res.transits.forEach((t) => {
          planets.add(t.transit_planet);
          planets.add(t.natal_planet);
        });
        const initial: Record<string, boolean> = {};
        sortPlanets([...planets]).forEach((p) => {
          initial[p] = true;
        });
        setVisiblePlanets(initial);
      })
      .catch((err) => {
        if (!cancelled) setError(err.message || 'Failed to load transit data');
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });

    return () => {
      cancelled = true;
    };
  }, [data, startDate, endDate]);

  // D3 render
  useEffect(() => {
    const svg = svgRef.current;
    const container = containerRef.current;
    if (!svg || !container || !transitData) return;

    const width = container.clientWidth || 800;
    const height = Math.max(container.clientHeight, 400);
    const innerW = width - MARGIN.left - MARGIN.right;
    const innerH = height - MARGIN.top - MARGIN.bottom;

    // Clear previous render
    d3.select(svg).selectAll('*').remove();

    const parseDate = d3.timeParse('%Y-%m-%d');
    const rangeStart = parseDate(startDate);
    const rangeEnd = parseDate(endDate);
    if (!rangeStart || !rangeEnd) return;

    // X scale: time
    const x = d3.scaleTime().domain([rangeStart, rangeEnd]).range([0, innerW]);

    // Y scale: 0–45°
    const y = d3.scaleLinear().domain([0, Y_DOMAIN]).range([innerH, 0]);

    // Root group
    const root = d3
      .select(svg)
      .attr('width', width)
      .attr('height', height)
      .append('g')
      .attr('transform', `translate(${MARGIN.left},${MARGIN.top})`);

    // Grid lines (horizontal every 5°)
    const yTicks = d3.range(0, Y_DOMAIN + 1, 5);
    yTicks.forEach((deg) => {
      root
        .append('line')
        .attr('x1', 0)
        .attr('x2', innerW)
        .attr('y1', y(deg))
        .attr('y2', y(deg))
        .attr('stroke', '#333')
        .attr('stroke-dasharray', deg % 15 === 0 ? 'none' : '2,4')
        .attr('stroke-width', deg % 15 === 0 ? 0.8 : 0.4);
    });

    // Vertical grid lines
    const timeTicks: Date[] = x.ticks(d3.timeMonth.every(1) as unknown as number);
    timeTicks.forEach((tick) => {
      root
        .append('line')
        .attr('x1', x(tick))
        .attr('x2', x(tick))
        .attr('y1', 0)
        .attr('y2', innerH)
        .attr('stroke', '#333')
        .attr('stroke-dasharray', '2,4')
        .attr('stroke-width', 0.4);
    });

    // X axis
    root
      .append('g')
      .attr('transform', `translate(0,${innerH})`)
      .call(
        d3
          .axisBottom(x)
          .ticks(d3.timeMonth.every(1))
          .tickFormat(d3.timeFormat('%b %Y') as never)
      )
      .call((g) => g.select('.domain').attr('stroke', '#555'))
      .call((g) => g.selectAll('.tick line').attr('stroke', '#555'))
      .call((g) => g.selectAll('.tick text').attr('fill', '#e0e0e0').attr('font-size', '10px'));

    // Y axis
    root
      .append('g')
      .call(
        d3
          .axisLeft(y)
          .tickValues(yTicks)
          .tickFormat((d) => `${d}°`)
      )
      .call((g) => g.select('.domain').attr('stroke', '#555'))
      .call((g) => g.selectAll('.tick line').attr('stroke', '#555'))
      .call((g) => g.selectAll('.tick text').attr('fill', '#e0e0e0').attr('font-size', '10px'));

    // Collect natal planets for reference lines
    const natalPlanets = new Set<string>();
    transitData.transits.forEach((t) => natalPlanets.add(t.natal_planet));
    const sortedNatal = sortPlanets([...natalPlanets]);

    // Draw natal reference lines at evenly spaced Y positions
    const natalYSpacing = sortedNatal.length > 1 ? Y_DOMAIN / (sortedNatal.length + 1) : Y_DOMAIN / 2;
    sortedNatal.forEach((planet, i) => {
      const yPos = natalYSpacing * (i + 1);
      const color = planetColor(planet);
      root
        .append('line')
        .attr('x1', 0)
        .attr('x2', innerW)
        .attr('y1', y(yPos))
        .attr('y2', y(yPos))
        .attr('stroke', color)
        .attr('stroke-dasharray', '6,3')
        .attr('stroke-width', 0.8)
        .attr('opacity', 0.5);

      // Label on the right
      root
        .append('text')
        .attr('x', innerW + 4)
        .attr('y', y(yPos))
        .attr('dy', '0.35em')
        .attr('fill', color)
        .attr('font-size', '10px')
        .text(planet);
    });

    // Group transit hits by transit_planet
    const byTransitPlanet = d3.group(transitData.transits, (t) => t.transit_planet);

    // Draw transit planet dots
    for (const [planet, hits] of byTransitPlanet) {
      if (!visiblePlanets[planet]) continue;

      const color = planetColor(planet);
      const glyph = planetGlyph(planet);
      // Use hash-based Y position so different planets don't overlap
      const baseY = (planetHash(planet) % Y_DOMAIN);

      const validHits = hits
        .map((h) => {
          const d = parseDate(h.date);
          return d ? { ...h, dateObj: d } : null;
        })
        .filter((h): h is NonNullable<typeof h> => h !== null)
        .filter((h) => h.dateObj >= rangeStart && h.dateObj <= rangeEnd);

      // Dots
      root
        .selectAll(`.dot-${planet}`)
        .data(validHits)
        .join('circle')
        .attr('cx', (d) => x(d.dateObj))
        .attr('cy', (d) => {
          // Spread dots slightly based on natal planet to avoid overlap at same date
          const natalOffset = (planetHash(d.natal_planet) % 10) - 5;
          return y(Math.max(0, Math.min(Y_DOMAIN, baseY + natalOffset)));
        })
        .attr('r', 4)
        .attr('fill', color)
        .attr('opacity', 0.85)
        .attr('stroke', '#000')
        .attr('stroke-width', 0.5);

      // Planet label at first dot
      if (validHits.length > 0) {
        const firstHit = validHits[0];
        const natalOffset = (planetHash(firstHit.natal_planet) % 10) - 5;
        const firstY = y(Math.max(0, Math.min(Y_DOMAIN, baseY + natalOffset)));
        root
          .append('text')
          .attr('x', x(firstHit.dateObj) - 8)
          .attr('y', firstY - 8)
          .attr('fill', color)
          .attr('font-size', '12px')
          .attr('text-anchor', 'end')
          .text(glyph);
      }
    }

    // Title
    root
      .append('text')
      .attr('x', innerW / 2)
      .attr('y', -10)
      .attr('text-anchor', 'middle')
      .attr('fill', '#e0e0e0')
      .attr('font-size', '14px')
      .attr('font-weight', 'bold')
      .text(`45° Graphic Ephemeris — ${data.name}`);

    // Y axis label
    root
      .append('text')
      .attr('transform', 'rotate(-90)')
      .attr('x', -innerH / 2)
      .attr('y', -45)
      .attr('text-anchor', 'middle')
      .attr('fill', '#e0e0e0')
      .attr('font-size', '11px')
      .text('Degree (mod 45°)');
  }, [transitData, visiblePlanets, startDate, endDate, data]);

  // Toggle a planet's visibility
  const togglePlanet = (planet: string) => {
    setVisiblePlanets((prev) => ({ ...prev, [planet]: !prev[planet] }));
  };

  if (loading) {
    return (
      <div style={{ minHeight: 400, display: 'flex', alignItems: 'center', justifyContent: 'center', color: '#e0e0e0' }}>
        Loading graphic ephemeris…
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

  const sortedPlanetNames = transitData
    ? sortPlanets([...new Set(transitData.transits.map((t) => t.transit_planet))])
    : [];

  return (
    <div style={{ width: '100%', height: '100%', minHeight: 400 }}>
      {/* Planet toggles */}
      {sortedPlanetNames.length > 0 && (
        <div
          style={{
            display: 'flex',
            flexWrap: 'wrap',
            gap: '8px',
            marginBottom: '8px',
            padding: '0 4px',
          }}
        >
          {sortedPlanetNames.map((planet) => (
            <label
              key={planet}
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: '4px',
                cursor: 'pointer',
                color: visiblePlanets[planet] ? planetColor(planet) : '#555',
                fontSize: '12px',
                userSelect: 'none',
                opacity: visiblePlanets[planet] ? 1 : 0.5,
              }}
            >
              <input
                type="checkbox"
                checked={visiblePlanets[planet] ?? true}
                onChange={() => togglePlanet(planet)}
                style={{ accentColor: planetColor(planet) }}
              />
              {planetGlyph(planet)} {planet}
            </label>
          ))}
        </div>
      )}

      {/* Chart container */}
      <div ref={containerRef} style={{ width: '100%', height: '100%', minHeight: 400 }}>
        <svg ref={svgRef} style={{ width: '100%', height: '100%', minHeight: 400 }} />
      </div>
    </div>
  );
}