import { describe, it, expect } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import { http, HttpResponse } from 'msw';
import { ChartWheel } from '../components/chart/ChartWheel';
import { server } from './mocks/server';
import type { BirthData } from '../lib/types';

const AJ: BirthData = {
  name: 'AJ',
  year: 1969,
  month: 2,
  day: 15,
  hour: 23,
  minute: 10,
  tz_offset: -8,
  lat: 47.038,
  lng: -122.901,
};

describe('ChartWheel', () => {
  it('renders loading state initially', () => {
    render(<ChartWheel data={AJ} />);
    expect(screen.getByText('Loading chart...')).toBeInTheDocument();
  });

  it('renders SVG after loading', async () => {
    render(<ChartWheel data={AJ} />);

    await waitFor(() => {
      const container = document.querySelector('.chart-svg');
      expect(container).toBeInTheDocument();
      const svg = container?.querySelector('svg');
      expect(svg).toBeInTheDocument();
    }, { timeout: 5000 });
  });

  it('shows export buttons after loading', async () => {
    render(<ChartWheel data={AJ} />);

    await waitFor(() => {
      expect(screen.getByText('Export SVG')).toBeInTheDocument();
      expect(screen.getByText('Export PNG')).toBeInTheDocument();
    }, { timeout: 5000 });
  });

  it('shows error state on API failure', async () => {
    // Override handler BEFORE rendering
    server.use(
      http.post('/api/chart', () => {
        return new HttpResponse('Server error', { status: 500 });
      })
    );

    render(<ChartWheel data={AJ} />);

    await waitFor(() => {
      expect(screen.getByText(/500/)).toBeInTheDocument();
    }, { timeout: 5000 });
  });
});
