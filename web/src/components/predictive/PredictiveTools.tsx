import { useState } from 'react';
import type { BirthData } from '../../lib/types';
import { SolarReturnView } from './SolarReturnView';
import { ProgressionsView } from './ProgressionsView';
import { DirectionsView } from './DirectionsView';
import { ProfectionView } from './ProfectionView';
import { FirdariaView } from './FirdariaView';
import { ZodiacalReleasingView } from './ZodiacalReleasingView';
import { TimingConvergenceView } from './TimingConvergenceView';

interface Props {
  data: BirthData;
}

type PredictiveTab = 'solar-return' | 'progressions' | 'directions' | 'profection' | 'firdaria' | 'zr' | 'timing';

const TABS: { id: PredictiveTab; label: string }[] = [
  { id: 'solar-return', label: 'Solar Return' },
  { id: 'progressions', label: 'Progressions' },
  { id: 'directions', label: 'Directions' },
  { id: 'profection', label: 'Profections' },
  { id: 'firdaria', label: 'Firdaria' },
  { id: 'zr', label: 'ZR' },
  { id: 'timing', label: 'Timing' },
];

export function PredictiveTools({ data }: Props) {
  const [tab, setTab] = useState<PredictiveTab>('solar-return');

  return (
    <div className="flex flex-col h-full">
      <div className="flex gap-1 px-4 py-2 border-b border-border shrink-0 overflow-x-auto">
        {TABS.map(t => (
          <button
            key={t.id}
            onClick={() => setTab(t.id)}
            className={`px-2 py-0.5 text-xs rounded whitespace-nowrap ${
              tab === t.id ? 'bg-accent text-white' : 'bg-surface text-muted border border-border'
            }`}
          >
            {t.label}
          </button>
        ))}
      </div>

      <div className="flex-1 overflow-hidden">
        {tab === 'solar-return' && <SolarReturnView data={data} />}
        {tab === 'progressions' && <ProgressionsView data={data} />}
        {tab === 'directions' && <DirectionsView data={data} />}
        {tab === 'profection' && <ProfectionView data={data} />}
        {tab === 'firdaria' && <FirdariaView data={data} />}
        {tab === 'zr' && <ZodiacalReleasingView data={data} />}
        {tab === 'timing' && <TimingConvergenceView data={data} />}
      </div>
    </div>
  );
}
