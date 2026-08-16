import { describe, it, expect } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import { BiWheel } from '../components/chart/BiWheel';
import type { BirthData } from '../lib/types';

const AJ: BirthData = {
  name: 'AJ',
  year: 1969, month: 2, day: 15, hour: 23, minute: 10,
  tz_offset: -8, lat: 47.038, lng: -122.901,
};

const TRANSITS: BirthData = {
  name: 'Transits',
  year: 2026, month: 7, day: 30, hour: 0, minute: 0,
  tz_offset: 0, lat: 47.038, lng: -122.901,
};

describe('BiWheel', () => {
  it('renders loading state initially', () => {
    render(<BiWheel inner={AJ} outer={TRANSITS} />);
    expect(screen.getByText('Loading bi-wheel...')).toBeInTheDocument();
  });

  it('renders SVG after loading', async () => {
    render(<BiWheel inner={AJ} outer={TRANSITS} />);

    await waitFor(() => {
      const container = document.querySelector('.chart-svg');
      expect(container).toBeInTheDocument();
      const svg = container?.querySelector('svg');
      expect(svg).toBeInTheDocument();
    }, { timeout: 5000 });
  });

  it('shows export buttons after loading', async () => {
    render(<BiWheel inner={AJ} outer={TRANSITS} />);

    await waitFor(() => {
      expect(screen.getByText('Export SVG')).toBeInTheDocument();
      expect(screen.getByText('Export PNG')).toBeInTheDocument();
    }, { timeout: 5000 });
  });
});
