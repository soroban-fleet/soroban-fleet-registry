import React from 'react';
import { getFleet, getFleetReleases, getFleetHistory } from '../../../../../lib/api/fleets';
import { FleetHeader } from '../../../../../components/fleet/FleetHeader';
import { FleetHistory } from '../../../../../components/fleet/FleetHistory';
import { ErrorState } from '../../../../../components/common/ErrorState';
import type { Fleet, FleetRelease, VerificationResult } from '../../../../../lib/types';

interface FleetHistoryPageProps {
  params: {
    owner: string;
    tag: string;
  };
  searchParams?: {
    limit?: string;
    offset?: string;
  };
}

export default async function FleetHistoryPage({
  params,
  searchParams,
}: FleetHistoryPageProps) {
  const { owner, tag } = params;
  const limit = Math.min(Math.max(parseInt(searchParams?.limit || '20', 10), 1), 100);
  const offset = Math.max(parseInt(searchParams?.offset || '0', 10), 0);

  let fleet: Fleet | null = null;
  let releases: FleetRelease[] = [];
  let verifications: VerificationResult[] = [];
  let fetchError: unknown = null;

  try {
    const fleetRes = await getFleet(owner, tag);
    fleet = fleetRes.data;

    try {
      const releasesRes = await getFleetReleases(owner, tag, { limit, offset });
      releases = releasesRes.data || [];
    } catch {
      releases = [];
    }

    try {
      const historyRes = await getFleetHistory(owner, tag, { limit, offset });
      verifications = historyRes.data || [];
    } catch {
      verifications = [];
    }
  } catch (err: unknown) {
    fetchError = err;
  }

  if (fetchError || !fleet) {
    return (
      <div className="container">
        <ErrorState error={fetchError} title="Error Loading Fleet History" />
      </div>
    );
  }

  return (
    <div className="container" style={{ display: 'flex', flexDirection: 'column', gap: '2rem' }}>
      <FleetHeader fleet={fleet} />

      <FleetHistory releases={releases} verifications={verifications} />
    </div>
  );
}
