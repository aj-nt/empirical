import { useState, useEffect } from 'react';
import type { BirthData } from '../../lib/types';
import { api } from '../../lib/api';

export function NatalReport({ data }: { data: BirthData }) {
  const [html, setHtml] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    setHtml(null);
    setError(null);

    api
      .natalHTML(data, 'western', 3)
      .then((result) => {
        if (!cancelled) setHtml(result);
      })
      .catch((err) => {
        if (!cancelled) setError(err instanceof Error ? err.message : String(err));
      });

    return () => {
      cancelled = true;
    };
  }, [data.name, data.year, data.month, data.day, data.hour, data.minute, data.tz_offset, data.lat, data.lng]);

  if (error) {
    return <p className="text-red-500">{error}</p>;
  }

  if (html === null) {
    return <p className="text-yellow-500">Generating report...</p>;
  }

  return (
    <iframe
      srcDoc={html}
      className="w-full h-full border-0"
      sandbox="allow-scripts"
      title="Natal Report"
    />
  );
}