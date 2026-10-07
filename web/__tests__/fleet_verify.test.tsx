import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { VerificationPanel } from '../components/fleet/VerificationPanel';
import FleetVerifyPage from '../app/fleets/[owner]/[tag]/verify/page';
import * as fleetsApi from '../lib/api/fleets';
import type { Fleet, VerificationResult } from '../lib/types';

vi.mock('next/navigation', () => ({
  useRouter: () => ({
    push: vi.fn(),
  }),
  usePathname: () => '/fleets/C123/vault-v1/verify',
}));

describe('Fleet Verification Components', () => {
  const sampleFleet: Fleet = {
    id: {
      Owner: 'CA3D5KRYMCMCZVACQY3CGINPVPUM2OAL6YHT3ZVM5C2MUMAA252AAAAB',
      Tag: 'vault-v1',
    },
    member_count: 100,
    first_seen_ledger: 1000,
    last_seen_ledger: 5000,
    last_indexed_ledger: 5000,
    current_wasm_hash: 'AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=',
  };

  const healthyVerification: VerificationResult = {
    fleet_id: sampleFleet.id,
    expected_wasm_hash: 'AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=',
    total_members: 100,
    matching_members: 100,
    mismatching_members: 0,
    missing_members: 0,
    indexed_through: 5000,
    status: 'HEALTHY',
    verified_at: '2026-10-07T12:00:00Z',
  };

  const driftVerification: VerificationResult = {
    fleet_id: sampleFleet.id,
    expected_wasm_hash: 'AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=',
    total_members: 2418,
    matching_members: 2417,
    mismatching_members: 1,
    missing_members: 0,
    indexed_through: 5000,
    status: 'DRIFT',
    verified_at: '2026-10-07T12:00:00Z',
  };

  it('renders HEALTHY verification panel with all matching members', () => {
    render(<VerificationPanel fleetId={sampleFleet.id} verification={healthyVerification} />);

    expect(screen.getByText('HEALTHY')).toBeInTheDocument();
    expect(screen.getAllByText('100')[0]).toBeInTheDocument();
    expect(screen.getByText('Target executable hash')).toBeInTheDocument();
  });

  it('renders DRIFT verification panel with diagnostic alert and link to members', () => {
    render(<VerificationPanel fleetId={sampleFleet.id} verification={driftVerification} />);

    expect(screen.getByText('DRIFT')).toBeInTheDocument();
    expect(screen.getByText(/Executable Drift Detected/)).toBeInTheDocument();
    expect(screen.getByText(/1 of 2,418 active member contracts resolve to an executable different/)).toBeInTheDocument();
    expect(
      screen.getByRole('link', { name: 'Investigate Divergent Members in Member List →' })
    ).toBeInTheDocument();
  });

  it('renders BROKEN_REFERENCE verification panel with explicit broken state', () => {
    const brokenVerification: VerificationResult = {
      ...healthyVerification,
      status: 'BROKEN_REFERENCE',
      matching_members: 0,
      missing_members: 100,
    };

    render(<VerificationPanel fleetId={sampleFleet.id} verification={brokenVerification} />);

    expect(screen.getByText('BROKEN_REFERENCE')).toBeInTheDocument();
    expect(screen.getByText(/Broken Executable Reference/)).toBeInTheDocument();
  });

  it('renders INCOMPLETE verification panel with indexer notice', () => {
    const incompleteVerification: VerificationResult = {
      ...healthyVerification,
      status: 'INCOMPLETE',
    };

    render(<VerificationPanel fleetId={sampleFleet.id} verification={incompleteVerification} />);

    expect(screen.getByText('INCOMPLETE')).toBeInTheDocument();
    expect(screen.getByText(/Indexer History Incomplete/)).toBeInTheDocument();
  });

  it('renders FleetVerifyPage with candidate WASM evaluation form', async () => {
    vi.spyOn(fleetsApi, 'getFleet').mockResolvedValueOnce({ data: sampleFleet });
    vi.spyOn(fleetsApi, 'verifyFleet').mockResolvedValueOnce({ data: healthyVerification });

    const page = await FleetVerifyPage({
      params: { owner: sampleFleet.id.Owner, tag: sampleFleet.id.Tag },
    });
    render(page);

    expect(screen.getByText('Evaluate Candidate WASM Hash')).toBeInTheDocument();
    expect(
      screen.getByPlaceholderText('Candidate 64-char hex WASM hash...')
    ).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Verify Candidate WASM' })).toBeInTheDocument();
  });
});
