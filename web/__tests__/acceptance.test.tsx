import React from 'react';
import { render, screen } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import FleetsPage from '../app/fleets/page';
import FleetDetailPage from '../app/fleets/[owner]/[tag]/page';
import FleetVerifyPage from '../app/fleets/[owner]/[tag]/verify/page';
import FleetMembersPage from '../app/fleets/[owner]/[tag]/members/page';
import ContractPage from '../app/contracts/[contractId]/page';
import * as fleetsApi from '../lib/api/fleets';
import * as contractsApi from '../lib/api/contracts';
import type { Fleet, VerificationResult, FleetMember, ContractInspection } from '../lib/types';

vi.mock('next/navigation', () => ({
  useRouter: () => ({ push: vi.fn() }),
  usePathname: () => '/fleets',
}));

describe('Application Acceptance Flows', () => {
  const owner = 'CA3D5KRYMCMCZVACQY3CGINPVPUM2OAL6YHT3ZVM5C2MUMAA252AAAAB';
  const tag = 'vault-v1';
  const memberContractId = 'CB111111111111111111111111111111111111111111111111111111';

  const mockFleet: Fleet = {
    id: { Owner: owner, Tag: tag },
    member_count: 2418,
    first_seen_ledger: 1000,
    last_seen_ledger: 5000,
    last_indexed_ledger: 5000,
    current_wasm_hash: 'AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=',
  };

  const healthyVerification: VerificationResult = {
    fleet_id: mockFleet.id,
    expected_wasm_hash: 'AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=',
    total_members: 2418,
    matching_members: 2418,
    mismatching_members: 0,
    missing_members: 0,
    indexed_through: 5000,
    status: 'HEALTHY',
    verified_at: '2026-10-07T12:00:00Z',
  };

  const driftVerification: VerificationResult = {
    fleet_id: mockFleet.id,
    expected_wasm_hash: 'AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=',
    total_members: 2418,
    matching_members: 2417,
    mismatching_members: 1,
    missing_members: 0,
    indexed_through: 5000,
    status: 'DRIFT',
    verified_at: '2026-10-07T12:00:00Z',
  };

  const incompleteVerification: VerificationResult = {
    fleet_id: mockFleet.id,
    expected_wasm_hash: 'AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=',
    total_members: 2418,
    matching_members: 2000,
    mismatching_members: 0,
    missing_members: 418,
    indexed_through: 4000,
    status: 'INCOMPLETE',
    verified_at: '2026-10-07T12:00:00Z',
  };

  beforeEach(() => {
    vi.restoreAllMocks();
  });

  // Acceptance Flow 1: Healthy Fleet Progression
  it('Acceptance Flow 1: loads fleet list -> inspects detail -> verifies HEALTHY state', async () => {
    vi.spyOn(fleetsApi, 'listFleets').mockResolvedValueOnce({
      data: [mockFleet],
      meta: { total: 1, limit: 20, offset: 0 },
    });
    vi.spyOn(fleetsApi, 'getFleet').mockResolvedValue({ data: mockFleet });
    vi.spyOn(fleetsApi, 'verifyFleet').mockResolvedValue({ data: healthyVerification });

    // 1. Browse Fleets list
    const listPage = await FleetsPage({});
    const { unmount: unmountList } = render(listPage);
    expect(screen.getByText('CAP-85 Contract Fleets')).toBeInTheDocument();
    expect(screen.getByText('vault-v1')).toBeInTheDocument();
    expect(screen.getByText('2,418')).toBeInTheDocument();
    unmountList();

    // 2. Load Fleet Detail
    const detailPage = await FleetDetailPage({ params: { owner, tag } });
    const { unmount: unmountDetail } = render(detailPage);
    expect(screen.getByText('Fleet Identity & Reference')).toBeInTheDocument();
    expect(screen.getAllByText('HEALTHY')[0]).toBeInTheDocument();
    unmountDetail();

    // 3. Load Verification View
    const verifyPage = await FleetVerifyPage({ params: { owner, tag } });
    render(verifyPage);
    expect(screen.getAllByText('HEALTHY')[0]).toBeInTheDocument();
    expect(screen.getAllByText('2,418')[0]).toBeInTheDocument();
  });

  // Acceptance Flow 2: Drift Fleet Progression & Divergence Investigation
  it('Acceptance Flow 2: DRIFT response -> shows DRIFT badge -> shows mismatching count -> links to members', async () => {
    vi.spyOn(fleetsApi, 'getFleet').mockResolvedValue({ data: mockFleet });
    vi.spyOn(fleetsApi, 'verifyFleet').mockResolvedValue({ data: driftVerification });

    // 1. Load Verification View with DRIFT
    const verifyPage = await FleetVerifyPage({ params: { owner, tag } });
    const { unmount: unmountVerify } = render(verifyPage);

    expect(screen.getAllByText('DRIFT')[0]).toBeInTheDocument();
    expect(screen.getByText(/Executable Drift Detected/)).toBeInTheDocument();
    expect(screen.getByText(/1 of 2,418 active member contracts resolve to an executable different/)).toBeInTheDocument();

    const investigateLink = screen.getByRole('link', {
      name: 'Investigate Divergent Members in Member List →',
    });
    expect(investigateLink).toHaveAttribute(
      'href',
      `/fleets/${encodeURIComponent(owner)}/${encodeURIComponent(tag)}/members`
    );
    unmountVerify();

    // 2. Load Members View
    const mockMembers: FleetMember[] = [
      {
        contract_id: memberContractId,
        fleet_id: mockFleet.id,
        wasm_hash: 'BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB=',
        first_seen_ledger: 1000,
        last_seen_ledger: 5000,
        last_verified_ledger: 5000,
        active: true,
      },
    ];
    vi.spyOn(fleetsApi, 'getFleetMembers').mockResolvedValueOnce({
      data: mockMembers,
      meta: { total: 1, limit: 20, offset: 0 },
    });

    const membersPage = await FleetMembersPage({ params: { owner, tag } });
    render(membersPage);
    expect(screen.getByText('Fleet Member Contracts')).toBeInTheDocument();
    expect(screen.getByText('● ACTIVE')).toBeInTheDocument();
  });

  // Acceptance Flow 3: Incomplete Indexing Never Shows Healthy
  it('Acceptance Flow 3: INCOMPLETE response -> displays INCOMPLETE and NEVER displays HEALTHY', async () => {
    vi.spyOn(fleetsApi, 'getFleet').mockResolvedValue({ data: mockFleet });
    vi.spyOn(fleetsApi, 'verifyFleet').mockResolvedValue({ data: incompleteVerification });

    const verifyPage = await FleetVerifyPage({ params: { owner, tag } });
    render(verifyPage);

    expect(screen.getAllByText('INCOMPLETE')[0]).toBeInTheDocument();
    expect(screen.getByText(/Indexer History Incomplete/)).toBeInTheDocument();
    expect(screen.queryByText('HEALTHY')).not.toBeInTheDocument();
  });

  // Acceptance Flow 4: Contract Inspection Flow
  it('Acceptance Flow 4: contract inspection distinguishes CAP-85 member vs standalone WASM', async () => {
    const memberInspection: ContractInspection = {
      contract_id: memberContractId,
      is_member: true,
      member: {
        contract_id: memberContractId,
        fleet_id: mockFleet.id,
        wasm_hash: 'AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=',
        first_seen_ledger: 1000,
        last_seen_ledger: 5000,
        last_verified_ledger: 5000,
        active: true,
      },
      resolved: {
        kind: 'EXTERNAL_REF',
        fleet: mockFleet.id,
        wasm_hash: 'AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=',
      },
    };

    vi.spyOn(contractsApi, 'getContract').mockResolvedValueOnce({ data: memberInspection });

    const memberPage = await ContractPage({ params: { contractId: memberContractId } });
    const { unmount } = render(memberPage);

    expect(screen.getByText('CAP-85 EXTERNAL_REF')).toBeInTheDocument();
    expect(screen.getByText('vault-v1')).toBeInTheDocument();
    expect(screen.getByText('● Active Member')).toBeInTheDocument();
    unmount();

    // Standalone contract
    const standaloneInspection: ContractInspection = {
      contract_id: memberContractId,
      is_member: false,
      resolved: { kind: 'WASM', wasm_hash: 'AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=' },
    };
    vi.spyOn(contractsApi, 'getContract').mockResolvedValueOnce({ data: standaloneInspection });

    const standalonePage = await ContractPage({ params: { contractId: memberContractId } });
    render(standalonePage);

    expect(screen.getByText('DIRECT WASM')).toBeInTheDocument();
    expect(screen.getByText(/Fleet:/)).toBeInTheDocument();
    expect(screen.getByText('None')).toBeInTheDocument();
  });
});
