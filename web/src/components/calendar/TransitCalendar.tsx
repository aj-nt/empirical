import { useState, useEffect, useMemo, useCallback } from 'react';
import type { BirthData, TransitResponse, TransitHit } from '../../lib/types';
import { api } from '../../lib/api';
import { planetGlyph, planetColor } from '../../lib/astrology';
import { generateICal, downloadICS } from '../../lib/export';

interface TransitCalendarProps {
  data: BirthData;
}

export function TransitCalendar({ data }: TransitCalendarProps) {
  const [currentMonth, setCurrentMonth] = useState(() => {
    const now = new Date();
    return new Date(now.getFullYear(), now.getMonth(), 1);
  });
  const [transitData, setTransitData] = useState<TransitResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Memoize month boundaries so they're stable between renders
  const monthStart = useMemo(
    () => new Date(currentMonth.getFullYear(), currentMonth.getMonth(), 1),
    [currentMonth],
  );
  const monthEnd = useMemo(
    () => new Date(currentMonth.getFullYear(), currentMonth.getMonth() + 1, 0),
    [currentMonth],
  );
  const calendarStart = useMemo(() => {
    const d = new Date(monthStart);
    d.setDate(d.getDate() - monthStart.getDay());
    return d;
  }, [monthStart]);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    setError(null);

    const startDate = monthStart.toISOString().split('T')[0];
    const endDate = monthEnd.toISOString().split('T')[0];

    api
      .transits(data, startDate, endDate, 3)
      .then((result) => {
        if (!cancelled) {
          setTransitData(result);
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
  }, [data.name, data.year, data.month, data.day, data.hour, data.minute, data.tz_offset, data.lat, data.lng, monthStart.getTime()]);

  // Group transit hits by date — expand each hit across its start_date..end_date range
  const hitsByDate = new Map<string, TransitHit[]>();
  if (transitData?.transits) {
    for (const hit of transitData.transits) {
      const start = hit.start_date || hit.date;
      const end = hit.end_date || hit.start_date || hit.date;
      if (!start || !end) continue;
      // Walk each day in the range
      const s = new Date(start);
      const e = new Date(end);
      for (let d = new Date(s); d <= e; d.setDate(d.getDate() + 1)) {
        const dateKey = d.toISOString().split('T')[0];
        const existing = hitsByDate.get(dateKey) ?? [];
        existing.push(hit);
        hitsByDate.set(dateKey, existing);
      }
    }
  }

  const navigateMonth = (delta: number) => {
    setCurrentMonth(
      new Date(currentMonth.getFullYear(), currentMonth.getMonth() + delta, 1),
    );
  };

  const exportICal = useCallback(() => {
    if (!transitData?.transits) return;
    const ics = generateICal(transitData.transits, data.name);
    const monthLabel = currentMonth.toISOString().slice(0, 7);
    downloadICS(ics, `${data.name}-transits-${monthLabel}.ics`);
  }, [transitData, data.name, currentMonth]);

  const today = new Date();
  const todayStr = `${today.getFullYear()}-${String(today.getMonth() + 1).padStart(2, '0')}-${String(today.getDate()).padStart(2, '0')}`;

  const dayHeaders = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];

  // Build calendar rows
  const rows: Date[][] = [];
  let rowStart = new Date(calendarStart);
  while (rowStart <= monthEnd || rowStart.getMonth() === currentMonth.getMonth()) {
    const row: Date[] = [];
    for (let i = 0; i < 7; i++) {
      const d = new Date(rowStart);
      row.push(d);
      rowStart.setDate(rowStart.getDate() + 1);
    }
    rows.push(row);
    // Stop if we've passed the end of the month and the row starts in the next month
    if (rowStart.getMonth() !== currentMonth.getMonth() && row[0].getMonth() !== currentMonth.getMonth()) {
      break;
    }
  }

  if (error) {
    return <div className="text-red-500 p-4">Error: {error}</div>;
  }

  return (
    <div className="transit-calendar">
      {/* Month Navigation */}
      <div className="flex items-center justify-between mb-4">
        <button
          onClick={() => navigateMonth(-1)}
          className="px-3 py-1 rounded bg-gray-700 hover:bg-gray-600 text-gray-200 transition-colors"
          aria-label="Previous month"
        >
          &lt;
        </button>
        <h2 className="text-lg font-semibold text-gray-100">
          {currentMonth.toLocaleString('default', { month: 'long', year: 'numeric' })}
        </h2>
        <div className="flex gap-2">
          <button
            onClick={exportICal}
            disabled={!transitData}
            className="px-3 py-1 text-xs rounded bg-surface text-muted border border-border hover:text-text disabled:opacity-30"
            title="Export as iCal"
          >
            iCal
          </button>
          <button
            onClick={() => navigateMonth(1)}
            className="px-3 py-1 rounded bg-gray-700 hover:bg-gray-600 text-gray-200 transition-colors"
            aria-label="Next month"
          >
            &gt;
          </button>
        </div>
      </div>

      {/* Day Headers */}
      <div className="grid grid-cols-7 gap-1 mb-1">
        {dayHeaders.map((day) => (
          <div key={day} className="text-center text-xs font-medium text-gray-400 py-1">
            {day}
          </div>
        ))}
      </div>

      {/* Calendar Grid */}
      {loading ? (
        <div className="text-yellow-500 p-4">Loading transits...</div>
      ) : (
        <div className="grid grid-cols-7 gap-1">
          {rows.map((row, rowIdx) =>
            row.map((day, colIdx) => {
              const dateStr = `${day.getFullYear()}-${String(day.getMonth() + 1).padStart(2, '0')}-${String(day.getDate()).padStart(2, '0')}`;
              const isCurrentMonth = day.getMonth() === currentMonth.getMonth();
              const isToday = dateStr === todayStr;
              const hits = hitsByDate.get(dateStr) ?? [];
              const maxDisplay = 5;
              const visibleHits = hits.slice(0, maxDisplay);
              const extraCount = hits.length - maxDisplay;

              return (
                <div
                  key={`${rowIdx}-${colIdx}`}
                  className={`
                    min-h-[60px] p-1 rounded text-sm
                    bg-gray-800/50
                    ${isToday ? 'ring-2 ring-blue-400' : ''}
                    ${isCurrentMonth ? 'text-gray-200' : 'text-gray-600'}
                  `}
                >
                  <div className="text-xs font-medium mb-0.5">{day.getDate()}</div>
                  <div className="flex flex-wrap gap-0.5">
                    {visibleHits.map((hit, i) => (
                      <span
                        key={i}
                        style={{
                          color: planetColor(hit.transit_planet),
                          fontSize: '10px',
                          lineHeight: '1',
                        }}
                        title={`${hit.transit_planet} ${hit.aspect} ${hit.natal_planet}`}
                      >
                        {planetGlyph(hit.transit_planet)}
                      </span>
                    ))}
                    {extraCount > 0 && (
                      <span className="text-gray-400" style={{ fontSize: '10px', lineHeight: '1' }}>
                        +{extraCount} more
                      </span>
                    )}
                  </div>
                </div>
              );
            }),
          )}
        </div>
      )}
    </div>
  );
}