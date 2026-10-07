import React from 'react';
import Link from 'next/link';
import { listFleets } from '../../lib/api/fleets';
import { FleetCard } from '../../components/fleet/FleetCard';
import { AddressDisplay } from '../../components/common/AddressDisplay';
import { HashDisplay } from '../../components/common/HashDisplay';
import { LedgerLink } from '../../components/common/LedgerLink';
import { ErrorState } from '../../components/common/ErrorState';
import { EmptyState } from '../../components/common/EmptyState';
import type { FleetListResponse } from '../../lib/types';

interface FleetsPageProps {
  searchParams?: {
    limit?: string;
    offset?: string;
  };
}

export default async function FleetsPage({ searchParams }: FleetsPageProps) {
  const limit = Math.min(Math.max(parseInt(searchParams?.limit || '20', 10), 1), 100);
  const offset = Math.max(parseInt(searchParams?.offset || '0', 10), 0);

  let response: FleetListResponse | null = null;
  let fetchError: unknown = null;

  try {
    response = await listFleets({ limit, offset });
  } catch (err: unknown) {
    fetchError = err;
  }

  if (fetchError) {
    return (
      <div className="container">
        <ErrorState error={fetchError} title="Error Loading Fleets" />
      </div>
    );
  }

  const fleets = response?.data || [];
  const total = response?.meta?.total || 0;
  const hasMore = offset + limit < total;
  const hasPrev = offset > 0;

  return (
    <div className="container" style={{ display: 'flex', flexDirection: 'column', gap: '2rem' }}>
      <div>
        <div style={{ fontSize: '0.8125rem', color: 'var(--text-muted)' }} className="mono">
          DISCOVERY
        </div>
        <h1 style={{ fontSize: '2rem', fontWeight: 700, margin: '0.25rem 0' }}>
          CAP-85 Contract Fleets
        </h1>
        <p style={{ color: 'var(--text-secondary)', fontSize: '0.9375rem' }}>
          All indexed Soroban contract groups managed under shared external executable references.
        </p>
      </div>

      {fleets.length === 0 ? (
        <EmptyState
          title="No Fleets Found"
          message="No CAP-85 fleets have been indexed yet."
        />
      ) : (
        <>
          <div className="table-container">
            <table>
              <thead>
                <tr>
                  <th>Tag</th>
                  <th>Owner</th>
                  <th style={{ textAlign: 'right' }}>Members</th>
                  <th>Current WASM</th>
                  <th>Last Indexed</th>
                  <th>Action</th>
                </tr>
              </thead>
              <tbody>
                {fleets.map((fleet) => {
                  const detailHref = `/fleets/${encodeURIComponent(fleet.id.Owner)}/${encodeURIComponent(fleet.id.Tag)}`;
                  return (
                    <tr key={`${fleet.id.Owner}-${fleet.id.Tag}`}>
                      <td style={{ fontWeight: 600 }}>
                        <Link href={detailHref} style={{ color: 'var(--text-primary)' }}>
                          {fleet.id.Tag}
                        </Link>
                      </td>
                      <td>
                        <AddressDisplay address={fleet.id.Owner} chars={4} />
                      </td>
                      <td style={{ textAlign: 'right' }} className="mono">
                        {fleet.member_count.toLocaleString()}
                      </td>
                      <td>
                        <HashDisplay hash={fleet.current_wasm_hash} chars={4} />
                      </td>
                      <td>
                        <LedgerLink ledger={fleet.last_indexed_ledger} />
                      </td>
                      <td>
                        <Link href={detailHref} className="btn" style={{ fontSize: '0.75rem', padding: '0.25rem 0.5rem' }}>
                          Inspect →
                        </Link>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>

          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginTop: '0.5rem' }}>
            <span style={{ fontSize: '0.8125rem', color: 'var(--text-secondary)' }} className="mono">
              Showing {fleets.length > 0 ? offset + 1 : 0}–{Math.min(offset + fleets.length, total)} of {total.toLocaleString()} fleets
            </span>

            <div style={{ display: 'flex', gap: '0.5rem' }}>
              {hasPrev && (
                <Link
                  href={`/fleets?limit=${limit}&offset=${Math.max(offset - limit, 0)}`}
                  className="btn"
                >
                  ← Previous
                </Link>
              )}
              {hasMore && (
                <Link
                  href={`/fleets?limit=${limit}&offset=${offset + limit}`}
                  className="btn"
                >
                  Next →
                </Link>
              )}
            </div>
          </div>
        </>
      )}
    </div>
  );
}
