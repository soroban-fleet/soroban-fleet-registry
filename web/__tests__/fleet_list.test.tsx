import React from 'react';
import { render, screen } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { FleetCard } from '../components/fleet/FleetCard';
import FleetsPage from '../app/fleets/page';
import * as fleetsApi from '../lib/api/fleets';
import { ApiError } from '../lib/api/client';
import type { Fleet } from '../lib/types';

describe('Fleet List Components', () => {
  const sampleFleet: Fleet = {
    id: {
      Owner: 'CA3D5KRYMCMCZVACQY3CGINPVPUM2OAL6YHT3ZVM5C2MUMAA252AAAAB',
      Tag: 'vault-v1',
    },
    member_count: 42,
    first_seen_ledger: 1000,
    last_seen_ledger: 1500,
    last_indexed_ledger: 1500,
    current_wasm_hash: 'AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=',
  };

  it('renders FleetCard with correct tag and member count', () => {
    render(<FleetCard fleet={sampleFleet} />);

    expect(screen.getByText('vault-v1')).toBeInTheDocument();
    expect(screen.getByText('42')).toBeInTheDocument();
    expect(screen.getByText('View Fleet Details →')).toBeInTheDocument();
  });

  it('renders FleetsPage with table rows when data is returned', async () => {
    vi.spyOn(fleetsApi, 'listFleets').mockResolvedValueOnce({
      data: [sampleFleet],
      meta: { total: 1, limit: 20, offset: 0 },
    });

    const page = await FleetsPage({});
    render(page);

    expect(screen.getByText('CAP-85 Contract Fleets')).toBeInTheDocument();
    expect(screen.getByText('vault-v1')).toBeInTheDocument();
    expect(screen.getByText('42')).toBeInTheDocument();
    expect(screen.getByText(/Showing 1–1 of 1 fleets/)).toBeInTheDocument();
  });

  it('renders EmptyState when no fleets exist', async () => {
    vi.spyOn(fleetsApi, 'listFleets').mockResolvedValueOnce({
      data: [],
      meta: { total: 0, limit: 20, offset: 0 },
    });

    const page = await FleetsPage({});
    render(page);

    expect(screen.getByText('No Fleets Found')).toBeInTheDocument();
    expect(screen.getByText('No CAP-85 fleets have been indexed yet.')).toBeInTheDocument();
  });

  it('renders ErrorState when API fails', async () => {
    vi.spyOn(fleetsApi, 'listFleets').mockRejectedValueOnce(
      new ApiError(500, 'Database query failed', 'INTERNAL_ERROR')
    );

    const page = await FleetsPage({});
    render(page);

    expect(screen.getByText('Error Loading Fleets')).toBeInTheDocument();
    expect(screen.getByText('Database query failed')).toBeInTheDocument();
  });
});
