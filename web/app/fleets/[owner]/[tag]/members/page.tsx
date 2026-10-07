import React from 'react';
import Link from 'next/link';
import { getFleet, getFleetMembers } from '../../../../../lib/api/fleets';
import { FleetHeader } from '../../../../../components/fleet/FleetHeader';
import { FleetMembersTable } from '../../../../../components/fleet/FleetMembersTable';
import { ErrorState } from '../../../../../components/common/ErrorState';
import type { Fleet, FleetMembersResponse } from '../../../../../lib/types';

interface FleetMembersPageProps {
  params: {
    owner: string;
    tag: string;
  };
  searchParams?: {
    limit?: string;
    offset?: string;
    active_only?: string;
  };
}

export default async function FleetMembersPage({
  params,
  searchParams,
}: FleetMembersPageProps) {
  const { owner, tag } = params;
  const limit = Math.min(Math.max(parseInt(searchParams?.limit || '20', 10), 1), 100);
  const offset = Math.max(parseInt(searchParams?.offset || '0', 10), 0);
  const activeOnly = searchParams?.active_only !== 'false';

  let fleet: Fleet | null = null;
  let membersRes: FleetMembersResponse | null = null;
  let fetchError: unknown = null;

  try {
    const fleetRes = await getFleet(owner, tag);
    fleet = fleetRes.data;

    membersRes = await getFleetMembers(owner, tag, {
      active_only: activeOnly,
      limit,
      offset,
    });
  } catch (err: unknown) {
    fetchError = err;
  }

  if (fetchError || !fleet) {
    return (
      <div className="container">
        <ErrorState error={fetchError} title="Error Loading Fleet Members" />
      </div>
    );
  }

  const members = membersRes?.data || [];
  const total = membersRes?.meta?.total || 0;
  const hasMore = offset + limit < total;
  const hasPrev = offset > 0;
  const basePath = `/fleets/${encodeURIComponent(fleet.id.Owner)}/${encodeURIComponent(fleet.id.Tag)}/members`;

  return (
    <div className="container" style={{ display: 'flex', flexDirection: 'column', gap: '2rem' }}>
      <FleetHeader fleet={fleet} />

      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: '1rem' }}>
        <div>
          <h2 style={{ fontSize: '1.25rem', fontWeight: 700 }}>Fleet Member Contracts</h2>
          <p style={{ fontSize: '0.875rem', color: 'var(--text-secondary)' }}>
            Total active instances indexing status and live on-chain executables.
          </p>
        </div>

        <div style={{ display: 'flex', gap: '0.5rem', alignItems: 'center' }}>
          <Link
            href={`${basePath}?active_only=true&limit=${limit}`}
            className={`btn ${activeOnly ? 'btn-primary' : ''}`}
            style={{ fontSize: '0.75rem' }}
          >
            Active Only
          </Link>
          <Link
            href={`${basePath}?active_only=false&limit=${limit}`}
            className={`btn ${!activeOnly ? 'btn-primary' : ''}`}
            style={{ fontSize: '0.75rem' }}
          >
            All Members
          </Link>
        </div>
      </div>

      <FleetMembersTable members={members} />

      {members.length > 0 && (
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginTop: '0.5rem' }}>
          <span style={{ fontSize: '0.8125rem', color: 'var(--text-secondary)' }} className="mono">
            Showing {offset + 1}–{Math.min(offset + members.length, total)} of {total.toLocaleString()} members
          </span>

          <div style={{ display: 'flex', gap: '0.5rem' }}>
            {hasPrev && (
              <Link
                href={`${basePath}?active_only=${activeOnly}&limit=${limit}&offset=${Math.max(offset - limit, 0)}`}
                className="btn"
              >
                ← Previous
              </Link>
            )}
            {hasMore && (
              <Link
                href={`${basePath}?active_only=${activeOnly}&limit=${limit}&offset=${offset + limit}`}
                className="btn"
              >
                Next →
              </Link>
            )}
          </div>
        </div>
      )}
    </div>
  );
}
