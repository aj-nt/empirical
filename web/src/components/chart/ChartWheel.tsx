import { useEffect, useRef, useState, useCallback } from 'react';
import type { BirthData } from '../../lib/types';
import { api } from '../../lib/api';

interface ChartWheelProps {
  data: BirthData;
}

function downloadBlob(blob: Blob, filename: string) {
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = filename;
  a.click();
  URL.revokeObjectURL(url);
}

export function ChartWheel({ data }: ChartWheelProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const [svg, setSvg] = useState('');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    setLoading(true);
    setError('');
    api.chart(data, { house_system: 'P', show_aspects: true, outer_planets: true, highlight_patterns: false, pattern_orb: 3 })
      .then((s) => {
        setSvg(s);
        setLoading(false);
      })
      .catch((e) => {
        setError(e instanceof Error ? e.message : 'Failed to load chart');
        setLoading(false);
      });
  }, [data.name, data.year, data.month, data.day, data.hour, data.minute, data.tz_offset, data.lat, data.lng]);

  useEffect(() => {
    if (!svg || !containerRef.current) return;
    containerRef.current.innerHTML = svg;
    const svgEl = containerRef.current.querySelector('svg');
    if (svgEl) {
      svgEl.setAttribute('width', '100%');
      svgEl.setAttribute('height', '100%');
      svgEl.style.display = 'block';
    }
  }, [svg]);

  const exportSVG = useCallback(() => {
    const svgEl = containerRef.current?.querySelector('svg');
    if (!svgEl) return;
    const clone = svgEl.cloneNode(true) as SVGElement;
    clone.setAttribute('xmlns', 'http://www.w3.org/2000/svg');
    const serializer = new XMLSerializer();
    const svgStr = serializer.serializeToString(clone);
    const blob = new Blob([svgStr], { type: 'image/svg+xml' });
    downloadBlob(blob, `${data.name}-chart.svg`);
  }, [data.name]);

  const exportPNG = useCallback(() => {
    const svgEl = containerRef.current?.querySelector('svg');
    if (!svgEl) return;
    const clone = svgEl.cloneNode(true) as SVGElement;
    clone.setAttribute('xmlns', 'http://www.w3.org/2000/svg');
    const serializer = new XMLSerializer();
    const svgStr = serializer.serializeToString(clone);
    const canvas = document.createElement('canvas');
    const exportSize = 1600;
    canvas.width = exportSize;
    canvas.height = exportSize;
    const ctx = canvas.getContext('2d');
    if (!ctx) return;
    const img = new Image();
    const svgBlob = new Blob([svgStr], { type: 'image/svg+xml' });
    const url = URL.createObjectURL(svgBlob);
    img.onload = () => {
      ctx.drawImage(img, 0, 0, canvas.width, canvas.height);
      URL.revokeObjectURL(url);
      canvas.toBlob((blob) => {
        if (blob) downloadBlob(blob, `${data.name}-chart.png`);
      }, 'image/png');
    };
    img.src = url;
  }, [data.name]);

  if (loading) {
    return <p className="text-yellow text-sm">Loading chart...</p>;
  }

  if (error) {
    return <p className="text-red text-sm">{error}</p>;
  }

  return (
    <div style={{ width: '100%', height: '100%', display: 'flex', flexDirection: 'column' }}>
      {/* Export buttons */}
      <div className="flex gap-2 mb-2 shrink-0">
        <button
          onClick={exportSVG}
          className="px-3 py-1 text-xs rounded bg-surface text-muted border border-border hover:text-text"
        >
          Export SVG
        </button>
        <button
          onClick={exportPNG}
          className="px-3 py-1 text-xs rounded bg-surface text-muted border border-border hover:text-text"
        >
          Export PNG
        </button>
      </div>
      <div style={{ flex: 1, display: 'flex', alignItems: 'center', justifyContent: 'center', overflow: 'hidden', minHeight: 0 }}>
        <div ref={containerRef} className="chart-svg" style={{ height: '100%', aspectRatio: '1' }} />
      </div>
    </div>
  );
}
