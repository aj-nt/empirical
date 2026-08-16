import { useState, useEffect, useCallback } from 'react';
import type { BirthData, InterpretationResponse } from '../../lib/types';
import { api } from '../../lib/api';

interface TooltipState {
  visible: boolean;
  x: number;
  y: number;
  title: string;
  text: string;
}

export function useWheelTooltip(data: BirthData) {
  const [tooltip, setTooltip] = useState<TooltipState>({ visible: false, x: 0, y: 0, title: '', text: '' });
  const [interp, setInterp] = useState<InterpretationResponse | null>(null);

  // Preload interpretation data when chart data changes
  useEffect(() => {
    let cancelled = false;
    setInterp(null);
    api.interpretation(data, 'western', 3)
      .then(result => { if (!cancelled) setInterp(result); })
      .catch(() => {}); // silently fail
    return () => { cancelled = true; };
  }, [data.name, data.year, data.month, data.day, data.hour, data.minute, data.tz_offset, data.lat, data.lng]);

  const handleClick = useCallback((e: React.MouseEvent) => {
    const target = e.target as Element;
    
    // Check for planet click
    const planetEl = target.closest('[data-planet]');
    if (planetEl) {
      e.stopPropagation();
      const planetName = planetEl.getAttribute('data-planet')!;
      const rect = planetEl.getBoundingClientRect();
      setTooltip({
        visible: true,
        x: rect.left + rect.width / 2,
        y: rect.top - 8,
        title: planetName,
        text: findPlanetText(interp, planetName),
      });
      return;
    }

    // Check for house click
    const houseEl = target.closest('[data-house]');
    if (houseEl) {
      e.stopPropagation();
      const houseNum = parseInt(houseEl.getAttribute('data-house')!);
      const rect = houseEl.getBoundingClientRect();
      setTooltip({
        visible: true,
        x: rect.left + rect.width / 2,
        y: rect.top - 8,
        title: `House ${houseNum}`,
        text: findHouseText(interp, houseNum),
      });
      return;
    }

    // Clicked elsewhere — close tooltip
    setTooltip(t => t.visible ? { ...t, visible: false } : t);
  }, [interp]);

  // Close on Escape
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') setTooltip(t => t.visible ? { ...t, visible: false } : t);
    };
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, []);

  return { tooltip, handleClick };
}

function findPlanetText(interp: InterpretationResponse | null, name: string): string {
  if (!interp) return 'Loading interpretation...';
  
  const planetOrder = ['Sun', 'Moon', 'Mercury', 'Venus', 'Mars', 'Jupiter', 'Saturn',
    'Uranus', 'Neptune', 'Pluto', 'Node', 'Chiron', 'Lilith', 'Ceres', 'Pallas', 'Juno',
    'Vesta', 'Eris', 'Makemake', 'Gonggong'];
  const idx = planetOrder.indexOf(name);
  
  const parts: string[] = [];
  if (idx >= 0 && idx < interp.planet_signs.length) {
    parts.push(interp.planet_signs[idx]);
  }
  if (idx >= 0 && idx < interp.planet_houses.length) {
    parts.push(interp.planet_houses[idx]);
  }
  
  if (interp.retrogrades?.some(r => r.startsWith(name))) {
    const retro = interp.retrogrades.find(r => r.startsWith(name));
    if (retro) parts.push(retro);
  }
  
  return parts.join('\n\n') || `${name} interpretation not available.`;
}

function findHouseText(interp: InterpretationResponse | null, house: number): string {
  if (!interp) return 'Loading interpretation...';
  
  const houseTexts = interp.planet_houses?.filter(h => h.includes(`house ${house}`) || h.includes(`${house}th house`)) || [];
  if (houseTexts.length > 0) return houseTexts.join('\n\n');
  
  const planetOrder = ['Sun', 'Moon', 'Mercury', 'Venus', 'Mars', 'Jupiter', 'Saturn',
    'Uranus', 'Neptune', 'Pluto', 'Node', 'Chiron', 'Lilith', 'Ceres', 'Pallas', 'Juno',
    'Vesta', 'Eris', 'Makemake', 'Gonggong'];
  const inHouse: string[] = [];
  for (let i = 0; i < interp.planet_houses.length && i < planetOrder.length; i++) {
    if (interp.planet_houses[i]?.includes(`house ${house}`) || interp.planet_houses[i]?.includes(`${house}th house`)) {
      inHouse.push(planetOrder[i]);
    }
  }
  
  if (inHouse.length > 0) {
    return `Planets in House ${house}: ${inHouse.join(', ')}`;
  }
  
  return `House ${house} interpretation not available.`;
}
