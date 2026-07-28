import { useState, useEffect } from 'react';
import type { BirthData } from '../../lib/types';
import { api } from '../../lib/api';

interface TransitReportProps {
  data: BirthData;
}

export function TransitReport({ data }: TransitReportProps) {
  const [html, setHtml] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    setHtml(null);
    setError(null);

    const now = new Date();

    api
      .transitHTML(
        data,
        {
          year: now.getFullYear(),
          month: now.getMonth() + 1,
          day: now.getDate(),
          hour: now.getHours(),
          minute: now.getMinutes(),
          tz: data.tz_offset,
          lat: data.lat,
          lng: data.lng,
        },
        'western',
        3,
      )
      .then((result) => {
        if (!cancelled) {
          setHtml(result);
        }
      })
      .catch((err) => {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : String(err));
        }
      });

    return () => {
      cancelled = true;
    };
  }, [data]);

  if (error) {
    return <div className="text-red-500">{error}</div>;
  }

  if (html === null) {
    return <div className="text-yellow-500">Generating report...</div>;
  }

  return (
    <iframe
      srcDoc={html}
      className="w-full h-full border-0"
      sandbox="allow-scripts"
      title="Transit Report"
    />
  );
}