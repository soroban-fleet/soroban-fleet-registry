import React from 'react';
import { render, screen } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { FleetHeader } from '../components/fleet/FleetHeader';
import { FleetStats } from '../components/fleet/FleetStats';
import FleetDetailPage from '../app/fleets/[owner]/[tag]/page';
import * as fleetsApi from '../lib/api/fleets';
import { ApiError } from '../lib/api/client';
import type { Fleet, VerificationResult } from '../lib/types';

describe('Fleet Detail Components', () => {
  const sampleFleet: Fleet = {
    id: {
      Owner: 'CA3D5KRYMCMCZVACQY3CGINPVPUM2OAL6YHT3ZVM5C2MUMAA252AAAAB',
      Tag: 'vault-v1',
    },
    member_count: 50,
    first_seen_ledger: 1000,
    last_seen_ledger: 2000,
    last_indexed_ledger: 2000,
    current_wasm_hash: 'AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=',
  };

  const sampleVerification: VerificationResult = {
    fleet_id: sampleFleet.id,
    expected_wasm_hash: 'AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=',
    total_members: 50,
    matching_members: 50,
    mismatching_members: 0,
    missing_members: 0,
    indexed_through: 2000,
    status: 'HEALTHY',
    verified_at: '2026-10-07T12:00:00Z',
  };

  it('renders FleetHeader with sub-navigation links', () => {
    render(<FleetHeader fleet={sampleFleet} status="HEALTHY" />);

    expect(screen.getByText('vault-v1')).toBeInTheDocument();
    expect(screen.getByText('HEALTHY')).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'Overview' })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'Members' })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'History & Releases' })).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'Verification' })).toBeInTheDocument();
  });

  it('renders FleetStats with metrics', () => {
    render(<FleetStats fleet={sampleFleet} verification={sampleVerification} />);

    expect(screen.getAllByText('50')[0]).toBeInTheDocument();
    expect(screen.getByText(/All active members match/)).toBeInTheDocument();
  });

  it('renders FleetDetailPage with HEALTHY verification', async () => {
    vi.spyOn(fleetsApi, 'getFleet').mockResolvedValueOnce({ data: sampleFleet });
    vi.spyOn(fleetsApi, 'verifyFleet').mockResolvedValueOnce({ data: sampleVerification });

    const page = await FleetDetailPage({
      params: { owner: sampleFleet.id.Owner, tag: sampleFleet.id.Tag },
    });
    render(page);

    expect(screen.getByText('Fleet Identity & Reference')).toBeInTheDocument();
    expect(screen.getByText('Current Verification Summary')).toBeInTheDocument();
    expect(screen.getAllByText('HEALTHY')[0]).toBeInTheDocument();
  });

  it('renders FleetDetailPage with DRIFT verification', async () => {
    const driftVerification: VerificationResult = {
      ...sampleVerification,
      status: 'DRIFT',
      matching_members: 49,
      mismatching_members: 1,
    };

    vi.spyOn(fleetsApi, 'getFleet').mockResolvedValueOnce({ data: sampleFleet });
    vi.spyOn(fleetsApi, 'verifyFleet').mockResolvedValueOnce({ data: driftVerification });

    const page = await FleetDetailPage({
      params: { owner: sampleFleet.id.Owner, tag: sampleFleet.id.Tag },
    });
    render(page);

    expect(screen.getAllByText('DRIFT')[0]).toBeInTheDocument();
    expect(screen.getByText(/1 mismatching members/)).toBeInTheDocument();
  });

  it('renders ErrorState when fleet does not exist', async () => {
    vi.spyOn(fleetsApi, 'getFleet').mockRejectedValueOnce(
      new ApiError(404, 'Fleet not found', 'FLEET_NOT_FOUND')
    );

    const page = await FleetDetailPage({
      params: { owner: 'C999', tag: 'missing' },
    });
    render(page);

    expect(screen.getByText('Fleet Not Found')).toBeInTheDocument();
    expect(screen.getByText('Fleet not found')).toBeInTheDocument();
  });
});
