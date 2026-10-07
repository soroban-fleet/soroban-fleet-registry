import React from 'react';
import { render, screen } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { FleetHistory } from '../components/fleet/FleetHistory';
import FleetHistoryPage from '../app/fleets/[owner]/[tag]/history/page';
import * as fleetsApi from '../lib/api/fleets';
import { ApiError } from '../lib/api/client';
import type { Fleet, FleetRelease, VerificationResult } from '../lib/types';

describe('Fleet History Components', () => {
  const sampleFleet: Fleet = {
    id: {
      Owner: 'CA3D5KRYMCMCZVACQY3CGINPVPUM2OAL6YHT3ZVM5C2MUMAA252AAAAB',
      Tag: 'vault-v1',
    },
    member_count: 10,
    first_seen_ledger: 1000,
    last_seen_ledger: 3000,
    last_indexed_ledger: 3000,
    current_wasm_hash: 'AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=',
  };

  const sampleReleases: FleetRelease[] = [
    {
      id: 1,
      fleet_id: sampleFleet.id,
      old_wasm_hash: null,
      new_wasm_hash: 'AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=',
      ledger: 1000,
      tx_hash: 'txhash11111111111111111111111111111111111111111111111111111111111',
      observed_at: '2026-10-01T10:00:00Z',
    },
  ];

  const sampleVerifications: VerificationResult[] = [
    {
      id: 1,
      fleet_id: sampleFleet.id,
      expected_wasm_hash: 'AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=',
      total_members: 10,
      matching_members: 10,
      mismatching_members: 0,
      missing_members: 0,
      indexed_through: 3000,
      status: 'HEALTHY',
      verified_at: '2026-10-07T10:00:00Z',
    },
  ];

  it('renders FleetHistory with releases, verifications, and explanatory note', () => {
    render(<FleetHistory releases={sampleReleases} verifications={sampleVerifications} />);

    expect(screen.getByText('Observed Executable Releases')).toBeInTheDocument();
    expect(screen.getByText(/A release observation records when the owner contract tag changed WASM/)).toBeInTheDocument();
    expect(screen.getByText('#1')).toBeInTheDocument();

    expect(screen.getByText('Verification Audit History')).toBeInTheDocument();
    expect(screen.getByText('HEALTHY')).toBeInTheDocument();
    expect(screen.getByText('10 / 10')).toBeInTheDocument();
  });

  it('renders EmptyStates when releases and verifications are empty', () => {
    render(<FleetHistory releases={[]} verifications={[]} />);

    expect(screen.getByText('No Releases Observed')).toBeInTheDocument();
    expect(screen.getByText('No Verification History')).toBeInTheDocument();
  });

  it('renders FleetHistoryPage successfully', async () => {
    vi.spyOn(fleetsApi, 'getFleet').mockResolvedValueOnce({ data: sampleFleet });
    vi.spyOn(fleetsApi, 'getFleetReleases').mockResolvedValueOnce({
      data: sampleReleases,
      meta: { total: 1, limit: 20, offset: 0 },
    });
    vi.spyOn(fleetsApi, 'getFleetHistory').mockResolvedValueOnce({
      data: sampleVerifications,
      meta: { total: 1, limit: 20, offset: 0 },
    });

    const page = await FleetHistoryPage({
      params: { owner: sampleFleet.id.Owner, tag: sampleFleet.id.Tag },
    });
    render(page);

    expect(screen.getByText('Observed Executable Releases')).toBeInTheDocument();
    expect(screen.getByText('#1')).toBeInTheDocument();
  });

  it('renders ErrorState when fleet not found', async () => {
    vi.spyOn(fleetsApi, 'getFleet').mockRejectedValueOnce(
      new ApiError(404, 'Fleet not found', 'FLEET_NOT_FOUND')
    );

    const page = await FleetHistoryPage({
      params: { owner: 'C999', tag: 'missing' },
    });
    render(page);

    expect(screen.getByText('Error Loading Fleet History')).toBeInTheDocument();
    expect(screen.getByText('Fleet not found')).toBeInTheDocument();
  });
});
