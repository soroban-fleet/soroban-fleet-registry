import React from 'react';
import { getFleet, verifyFleet } from '../../../../../lib/api/fleets';
import { FleetHeader } from '../../../../../components/fleet/FleetHeader';
import { VerificationPanel } from '../../../../../components/fleet/VerificationPanel';
import { ErrorState } from '../../../../../components/common/ErrorState';
import type { Fleet, VerificationResult } from '../../../../../lib/types';

interface FleetVerifyPageProps {
  params: {
    owner: string;
    tag: string;
  };
  searchParams?: {
    expected_wasm?: string;
  };
}

export default async function FleetVerifyPage({
  params,
  searchParams,
}: FleetVerifyPageProps) {
  const { owner, tag } = params;
  const expectedWasm = searchParams?.expected_wasm;

  let fleet: Fleet | null = null;
  let verification: VerificationResult | null = null;
  let fetchError: unknown = null;

  try {
    const fleetRes = await getFleet(owner, tag);
    fleet = fleetRes.data;

    const verifyRes = await verifyFleet(owner, tag, {
      expected_wasm: expectedWasm,
    });
    verification = verifyRes.data;
  } catch (err: unknown) {
    fetchError = err;
  }

  if (fetchError || !fleet || !verification) {
    return (
      <div className="container">
        <ErrorState error={fetchError} title="Error Verifying Fleet" />
      </div>
    );
  }

  return (
    <div className="container" style={{ display: 'flex', flexDirection: 'column', gap: '2rem' }}>
      <FleetHeader fleet={fleet} status={verification.status} />

      <VerificationPanel fleetId={fleet.id} verification={verification} />
    </div>
  );
}
