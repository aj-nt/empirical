import { useState, useEffect, useRef, useCallback } from 'react';
import type { BirthData } from '../../lib/types';
import { api } from '../../lib/api';

interface TransitAnimationProps {
  data: BirthData;
  houseSystem?: string;
  ayanamsa?: string;
}

interface Frame {
  date: Date;
  svg: string;
}

export function TransitAnimation({ data, houseSystem, ayanamsa }: TransitAnimationProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const [frames, setFrames] = useState<Frame[]>([]);
  const [currentFrame, setCurrentFrame] = useState(0);
  const [playing, setPlaying] = useState(false);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [speed, setSpeed] = useState(1); // frames per second
  const timerRef = useRef<ReturnType<typeof setInterval> | null>(null);

  const sidereal = !!ayanamsa;

  // Pre-compute frames: one bi-wheel per day for the next 30 days
  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    setError('');

    async function loadFrames() {
      const result: Frame[] = [];
      const now = new Date();
      const hs = houseSystem || 'placidus';

      for (let i = 0; i < 30; i++) {
        if (cancelled) break;
        const d = new Date(now);
        d.setDate(d.getDate() + i);

        try {
          const svg = await api.biWheel(
            data,
            {
              name: 'Transit',
              year: d.getFullYear(),
              month: d.getMonth() + 1,
              day: d.getDate(),
              hour: d.getHours(),
              minute: d.getMinutes(),
              tz_offset: data.tz_offset,
              lat: data.lat,
              lng: data.lng,
            },
            { houseSystem: hs, sidereal, ayanamsa }
          );
          result.push({ date: new Date(d), svg });
        } catch {
          // skip failed frames
        }
      }

      if (!cancelled) {
        setFrames(result);
        setLoading(false);
      }
    }

    loadFrames();
    return () => { cancelled = true; };
  }, [data.name, data.year, data.month, data.day, data.hour, data.minute, data.tz_offset, data.lat, data.lng, houseSystem, ayanamsa]);

  // Display current frame
  useEffect(() => {
    if (!containerRef.current || frames.length === 0) return;
    containerRef.current.innerHTML = frames[currentFrame]?.svg || '';
    const svgEl = containerRef.current.querySelector('svg');
    if (svgEl) {
      svgEl.setAttribute('width', '100%');
      svgEl.setAttribute('height', '100%');
      svgEl.style.display = 'block';
    }
  }, [frames, currentFrame]);

  // Play/pause timer
  useEffect(() => {
    if (playing && frames.length > 1) {
      timerRef.current = setInterval(() => {
        setCurrentFrame(prev => {
          if (prev >= frames.length - 1) {
            setPlaying(false);
            return prev;
          }
          return prev + 1;
        });
      }, 1000 / speed);
    } else {
      if (timerRef.current) {
        clearInterval(timerRef.current);
        timerRef.current = null;
      }
    }
    return () => {
      if (timerRef.current) clearInterval(timerRef.current);
    };
  }, [playing, speed, frames.length]);

  const togglePlay = useCallback(() => setPlaying(p => !p), []);
  const reset = useCallback(() => { setPlaying(false); setCurrentFrame(0); }, []);

  const currentDate = frames[currentFrame]?.date;
  const dateLabel = currentDate
    ? currentDate.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' })
    : '';

  if (loading) {
    return (
      <div className="flex items-center justify-center h-full text-muted">
        <div className="text-center">
          <div className="animate-spin text-2xl mb-2">⏳</div>
          <p className="text-sm">Computing transit frames...</p>
          <p className="text-xs text-muted mt-1">{frames.length}/30</p>
        </div>
      </div>
    );
  }

  if (error) {
    return <div className="flex items-center justify-center h-full text-red text-sm">{error}</div>;
  }

  if (frames.length === 0) {
    return <div className="flex items-center justify-center h-full text-muted text-sm">No frames available</div>;
  }

  return (
    <div className="flex flex-col h-full">
      {/* Controls */}
      <div className="flex items-center gap-3 px-4 py-2 border-b border-border shrink-0">
        <button
          onClick={togglePlay}
          className="px-3 py-1 rounded bg-accent text-white text-sm font-semibold hover:opacity-90 min-w-[60px]"
        >
          {playing ? '⏸ Pause' : '▶ Play'}
        </button>
        <button
          onClick={reset}
          className="px-3 py-1 rounded bg-surface border border-border text-sm text-muted hover:text-text"
        >
          ⏮ Reset
        </button>

        {/* Speed control */}
        <div className="flex items-center gap-1 text-xs text-muted">
          <span>Speed:</span>
          {[0.5, 1, 2, 5].map(s => (
            <button
              key={s}
              onClick={() => setSpeed(s)}
              className={`px-1.5 py-0.5 rounded ${speed === s ? 'bg-accent text-white' : 'hover:text-text'}`}
            >
              {s}x
            </button>
          ))}
        </div>

        {/* Frame counter */}
        <span className="text-xs text-muted ml-auto">
          {currentFrame + 1} / {frames.length}
        </span>
      </div>

      {/* Scrubber */}
      <div className="px-4 py-1 shrink-0">
        <input
          type="range"
          min={0}
          max={frames.length - 1}
          value={currentFrame}
          onChange={e => { setPlaying(false); setCurrentFrame(Number(e.target.value)); }}
          className="w-full h-2 appearance-none bg-surface rounded accent-accent cursor-pointer"
        />
        <div className="text-center text-xs text-muted mt-0.5">{dateLabel}</div>
      </div>

      {/* Chart display */}
      <div className="flex-1 flex items-center justify-center overflow-hidden min-h-0 p-2">
        <div ref={containerRef} className="chart-svg" style={{ height: '100%', aspectRatio: '1' }} />
      </div>
    </div>
  );
}
