import React from 'react';
import { render, screen } from '@testing-library/react';
import { describe, it, expect, vi } from 'vitest';
import ContractPage from '../app/contracts/[contractId]/page';
import * as contractsApi from '../lib/api/contracts';
import { ApiError } from '../lib/api/client';
import type { ContractInspection } from '../lib/types';

describe('Contract Detail Component', () => {
  const memberContractId = 'CB222222222222222222222222222222222222222222222222222222';
  const ownerAddress = 'CA3D5KRYMCMCZVACQY3CGINPVPUM2OAL6YHT3ZVM5C2MUMAA252AAAAB';

  const memberInspection: ContractInspection = {
    contract_id: memberContractId,
    is_member: true,
    member: {
      contract_id: memberContractId,
      fleet_id: { Owner: ownerAddress, Tag: 'vault-v1' },
      wasm_hash: 'AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=',
      first_seen_ledger: 1000,
      last_seen_ledger: 2000,
      last_verified_ledger: 2000,
      active: true,
    },
    resolved: {
      kind: 'EXTERNAL_REF',
      fleet: { Owner: ownerAddress, Tag: 'vault-v1' },
      wasm_hash: 'AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=',
    },
  };

  const standaloneInspection: ContractInspection = {
    contract_id: memberContractId,
    is_member: false,
    resolved: {
      kind: 'WASM',
      wasm_hash: 'BBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBBB=',
    },
  };

  it('renders CAP-85 fleet member contract details', async () => {
    vi.spyOn(contractsApi, 'getContract').mockResolvedValueOnce({ data: memberInspection });

    const page = await ContractPage({ params: { contractId: memberContractId } });
    render(page);

    expect(screen.getByText('CAP-85 EXTERNAL_REF')).toBeInTheDocument();
    expect(screen.getByText('vault-v1')).toBeInTheDocument();
    expect(screen.getByText('● Active Member')).toBeInTheDocument();
    expect(screen.getByRole('link', { name: 'View Fleet Overview →' })).toBeInTheDocument();
  });

  it('renders standalone direct WASM contract without fleet', async () => {
    vi.spyOn(contractsApi, 'getContract').mockResolvedValueOnce({ data: standaloneInspection });

    const page = await ContractPage({ params: { contractId: memberContractId } });
    render(page);

    expect(screen.getByText('DIRECT WASM')).toBeInTheDocument();
    expect(screen.getByText(/Fleet:/)).toBeInTheDocument();
    expect(screen.getByText('None')).toBeInTheDocument();
  });

  it('renders ErrorState when API fails', async () => {
    vi.spyOn(contractsApi, 'getContract').mockRejectedValueOnce(
      new ApiError(404, 'Contract not found', 'NOT_FOUND')
    );

    const page = await ContractPage({ params: { contractId: 'C_UNKNOWN' } });
    render(page);

    expect(screen.getByText('Contract Inspection Error')).toBeInTheDocument();
    expect(screen.getByText('Contract not found')).toBeInTheDocument();
  });
});
