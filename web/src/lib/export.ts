/**
 * Generate an iCalendar (.ics) string from transit hits.
 * Each hit becomes a VEVENT spanning its start_date..end_date.
 */
export function generateICal(
  hits: Array<{
    transit_planet: string;
    natal_planet: string;
    aspect: string;
    date: string;
    start_date?: string;
    end_date?: string;
    orb: number;
  }>,
  chartName: string,
): string {
  const lines: string[] = [
    'BEGIN:VCALENDAR',
    'VERSION:2.0',
    'PRODID:-//Empirical Astrology//Transit Calendar//EN',
    'CALSCALE:GREGORIAN',
    'METHOD:PUBLISH',
    `X-WR-CALNAME:Transits for ${chartName}`,
    'X-WR-TIMEZONE:UTC',
  ];

  for (const hit of hits) {
    const start = toICSDate(hit.start_date || hit.date);
    const end = toICSDate(hit.end_date || hit.start_date || hit.date, true);
    const summary = `${hit.transit_planet} ${hit.aspect} ${hit.natal_planet}`;
    const uid = `${hit.date}-${hit.transit_planet}-${hit.aspect}-${hit.natal_planet}@empirical`;

    lines.push(
      'BEGIN:VEVENT',
      `DTSTART;VALUE=DATE:${start}`,
      `DTEND;VALUE=DATE:${end}`,
      `SUMMARY:${escapeICS(summary)}`,
      `DESCRIPTION:${escapeICS(`${summary} (orb: ${hit.orb}°)`)}`,
      `UID:${uid}`,
      'TRANSP:TRANSPARENT',
      'END:VEVENT',
    );
  }

  lines.push('END:VCALENDAR');
  return lines.join('\r\n');
}

function toICSDate(dateStr: string, isEnd = false): string {
  // dateStr is YYYY-MM-DD
  const d = new Date(dateStr + 'T00:00:00Z');
  if (isEnd) {
    // iCal end dates are exclusive — add one day
    d.setDate(d.getDate() + 1);
  }
  return d.toISOString().split('T')[0].replace(/-/g, '');
}

function escapeICS(text: string): string {
  return text
    .replace(/\\/g, '\\\\')
    .replace(/;/g, '\\;')
    .replace(/,/g, '\\,')
    .replace(/\n/g, '\\n');
}

export function downloadICS(ics: string, filename: string) {
  const blob = new Blob([ics], { type: 'text/calendar;charset=utf-8' });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = filename;
  a.click();
  URL.revokeObjectURL(url);
}
