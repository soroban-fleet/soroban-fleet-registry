import React from 'react';
import { render, screen } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import { FleetMembersTable } from '../components/fleet/FleetMembersTable';
import FleetMembersPage from '../app/fleets/[owner]/[tag]/members/page';
import * as fleetsApi from '../lib/api/fleets';
import { ApiError } from '../lib/api/client';
import type { Fleet, FleetMember } from '../lib/types';

describe('Fleet Members Components', () => {
  const sampleFleet: Fleet = {
    id: {
      Owner: 'CA3D5KRYMCMCZVACQY3CGINPVPUM2OAL6YHT3ZVM5C2MUMAA252AAAAB',
      Tag: 'vault-v1',
    },
    member_count: 2,
    first_seen_ledger: 1000,
    last_seen_ledger: 2000,
    last_indexed_ledger: 2000,
    current_wasm_hash: 'AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=',
  };

  const sampleMembers: FleetMember[] = [
    {
      contract_id: 'CB222222222222222222222222222222222222222222222222222222',
      fleet_id: sampleFleet.id,
      wasm_hash: 'AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=',
      first_seen_ledger: 1000,
      last_seen_ledger: 2000,
      last_verified_ledger: 2000,
      active: true,
    },
    {
      contract_id: 'CB333333333333333333333333333333333333333333333333333333',
      fleet_id: sampleFleet.id,
      wasm_hash: 'BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB=',
      first_seen_ledger: 1100,
      last_seen_ledger: 1800,
      last_verified_ledger: null,
      active: false,
    },
  ];

  it('renders FleetMembersTable with active and inactive member states', () => {
    render(<FleetMembersTable members={sampleMembers} />);

    expect(screen.getByText('● ACTIVE')).toBeInTheDocument();
    expect(screen.getByText('○ INACTIVE')).toBeInTheDocument();
    expect(screen.getByText('CB222222...222222')).toBeInTheDocument();
    expect(screen.getByText('CB333333...333333')).toBeInTheDocument();
  });

  it('renders EmptyState when members list is empty', () => {
    render(<FleetMembersTable members={[]} />);

    expect(screen.getByText('No Members')).toBeInTheDocument();
    expect(screen.getByText('No active members were returned for this fleet.')).toBeInTheDocument();
  });

  it('renders FleetMembersPage with data and pagination indicator', async () => {
    vi.spyOn(fleetsApi, 'getFleet').mockResolvedValueOnce({ data: sampleFleet });
    vi.spyOn(fleetsApi, 'getFleetMembers').mockResolvedValueOnce({
      data: sampleMembers,
      meta: { total: 2, limit: 20, offset: 0 },
    });

    const page = await FleetMembersPage({
      params: { owner: sampleFleet.id.Owner, tag: sampleFleet.id.Tag },
    });
    render(page);

    expect(screen.getByText('Fleet Member Contracts')).toBeInTheDocument();
    expect(screen.getByText('Showing 1–2 of 2 members')).toBeInTheDocument();
  });

  it('renders ErrorState when API fails', async () => {
    vi.spyOn(fleetsApi, 'getFleet').mockRejectedValueOnce(
      new ApiError(500, 'Database query error', 'DB_ERROR')
    );

    const page = await FleetMembersPage({
      params: { owner: sampleFleet.id.Owner, tag: sampleFleet.id.Tag },
    });
    render(page);

    expect(screen.getByText('Error Loading Fleet Members')).toBeInTheDocument();
    expect(screen.getByText('Database query error')).toBeInTheDocument();
  });
});
