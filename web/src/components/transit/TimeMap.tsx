import { useEffect, useRef } from 'react';
import * as d3 from 'd3';
import type { TransitHit } from '../../lib/types';
import { aspectColor } from '../../lib/astrology';

interface TimeMapProps {
  hits: TransitHit[];
  startDate: string;
  endDate: string;
  onHover?: (hit: TransitHit | null) => void;
}

const MARGIN = { top: 20, right: 20, bottom: 40, left: 80 };

const parseDate = d3.timeParse('%Y-%m-%d');

export function TimeMap({ hits, startDate, endDate, onHover }: TimeMapProps) {
  const svgRef = useRef<SVGSVGElement>(null);

  useEffect(() => {
    const svg = svgRef.current;
    if (!svg) return;

    const container = svg.parentElement;
    if (!container) return;

    const width = container.clientWidth;
    const height = Math.max(container.clientHeight, 300);

    // Clear previous render
    d3.select(svg).selectAll('*').remove();

    if (hits.length === 0) return;

    // Parse range boundaries
    const rangeStart = parseDate(startDate);
    const rangeEnd = parseDate(endDate);
    if (!rangeStart || !rangeEnd) return;

    // Enrich hits with computed start/end and filter to valid range
    const processed = hits
      .map((hit) => {
        const exact = parseDate(hit.date);
        if (!exact) return null;

        const start = hit.start_date
          ? parseDate(hit.start_date)
          : d3.timeDay.offset(exact, -3);
        const end = hit.end_date
          ? parseDate(hit.end_date)
          : d3.timeDay.offset(exact, 3);

        if (!start || !end) return null;

        return { ...hit, exact, start, end };
      })
      .filter(
        (h): h is NonNullable<typeof h> =>
          h !== null && h.end >= rangeStart && h.start <= rangeEnd,
      );

    if (processed.length === 0) return;

    // Scales
    const xScale = d3
      .scaleTime()
      .domain([rangeStart, rangeEnd])
      .range([MARGIN.left, width - MARGIN.right]);

    const planets = [...new Set(processed.map((h) => h.transit_planet))].sort(
      d3.ascending,
    );

    const yScale = d3
      .scaleBand()
      .domain(planets)
      .range([MARGIN.top, height - MARGIN.bottom])
      .padding(0.25);

    const selection = d3.select(svg);

    // Grid lines (vertical, monthly)
    const months = d3.timeMonths(rangeStart, rangeEnd);
    selection
      .append('g')
      .attr('class', 'grid-lines')
      .selectAll('line')
      .data(months)
      .join('line')
      .attr('x1', (d) => xScale(d))
      .attr('x2', (d) => xScale(d))
      .attr('y1', MARGIN.top)
      .attr('y2', height - MARGIN.bottom)
      .attr('stroke', '#333')
      .attr('stroke-dasharray', '2,2');

    // Horizontal grid lines
    selection
      .append('g')
      .attr('class', 'h-grid')
      .selectAll('line')
      .data(planets)
      .join('line')
      .attr('x1', MARGIN.left)
      .attr('x2', width - MARGIN.right)
      .attr('y1', (d) => yScale(d)! + yScale.bandwidth() / 2)
      .attr('y2', (d) => yScale(d)! + yScale.bandwidth() / 2)
      .attr('stroke', '#222')
      .attr('stroke-dasharray', '1,3');

    // Bars
    const barGroup = selection
      .append('g')
      .attr('class', 'bars')
      .selectAll('rect')
      .data(processed)
      .join('rect')
      .attr('x', (d) => xScale(d.start))
      .attr('y', (d) => yScale(d.transit_planet)!)
      .attr('width', (d) =>
        Math.max(0, xScale(d.end) - xScale(d.start)),
      )
      .attr('height', yScale.bandwidth())
      .attr('fill', (d) => aspectColor(d.aspect))
      .attr('fill-opacity', 0.35)
      .attr('rx', 2);

    // Tooltips via <title>
    barGroup.append('title').text(
      (d) =>
        `${d.transit_planet} ${d.aspect} ${d.natal_planet} — ${d.date} (orb ${d.orb}°)`,
    );

    // Hover handlers
    if (onHover) {
      barGroup
        .on('mouseenter', (_event, d) => onHover(d))
        .on('mouseleave', () => onHover(null));
    }

    // Exact hit markers (thin vertical lines)
    selection
      .append('g')
      .attr('class', 'exact-markers')
      .selectAll('line')
      .data(processed)
      .join('line')
      .attr('x1', (d) => xScale(d.exact))
      .attr('x2', (d) => xScale(d.exact))
      .attr(
        'y1',
        (d) => yScale(d.transit_planet)!,
      )
      .attr(
        'y2',
        (d) => yScale(d.transit_planet)! + yScale.bandwidth(),
      )
      .attr('stroke', (d) => aspectColor(d.aspect))
      .attr('stroke-width', 2);

    // X-axis
    const xAxis = d3
      .axisBottom(xScale)
      .ticks(d3.timeMonth.every(1))
      .tickFormat((domainValue: Date | d3.NumberValue) => {
        if (domainValue instanceof Date) return d3.timeFormat('%b %Y')(domainValue);
        return '';
      });

    selection
      .append('g')
      .attr('class', 'x-axis')
      .attr('transform', `translate(0,${height - MARGIN.bottom})`)
      .call(xAxis as unknown as (s: d3.Selection<SVGGElement, unknown, null, undefined>) => void)
      .selectAll('text')
      .attr('fill', '#e0e0e0')
      .attr('font-size', '11px');

    selection
      .selectAll('.x-axis line, .x-axis path')
      .attr('stroke', '#555');

    // Y-axis
    const yAxis = d3.axisLeft(yScale);

    selection
      .append('g')
      .attr('class', 'y-axis')
      .attr('transform', `translate(${MARGIN.left},0)`)
      .call(yAxis as unknown as (s: d3.Selection<SVGGElement, unknown, null, undefined>) => void)
      .selectAll('text')
      .attr('fill', '#e0e0e0')
      .attr('font-size', '12px');

    selection
      .selectAll('.y-axis line, .y-axis path')
      .attr('stroke', '#555');
  }, [hits, startDate, endDate, onHover]);

  return (
    <svg
      ref={svgRef}
      style={{ width: '100%', height: '100%', minHeight: 300 }}
    />
  );
}