import React from 'react';
import Link from 'next/link';
import { getFleet, verifyFleet } from '../../../../lib/api/fleets';
import { FleetHeader } from '../../../../components/fleet/FleetHeader';
import { FleetStats } from '../../../../components/fleet/FleetStats';
import { FleetStatus } from '../../../../components/fleet/FleetStatus';
import { AddressDisplay } from '../../../../components/common/AddressDisplay';
import { HashDisplay } from '../../../../components/common/HashDisplay';
import { LedgerLink } from '../../../../components/common/LedgerLink';
import { ErrorState } from '../../../../components/common/ErrorState';
import type { Fleet, VerificationResult } from '../../../../lib/types';

interface FleetDetailPageProps {
  params: {
    owner: string;
    tag: string;
  };
}

export default async function FleetDetailPage({ params }: FleetDetailPageProps) {
  const { owner, tag } = params;

  let fleet: Fleet | null = null;
  let verification: VerificationResult | undefined = undefined;
  let fetchError: unknown = null;

  try {
    const fleetRes = await getFleet(owner, tag);
    fleet = fleetRes.data;

    try {
      const verifyRes = await verifyFleet(owner, tag);
      verification = verifyRes.data;
    } catch {
      // If verification is unavailable or fails, fleet view still renders with status UNKNOWN
    }
  } catch (err: unknown) {
    fetchError = err;
  }

  if (fetchError || !fleet) {
    return (
      <div className="container">
        <ErrorState error={fetchError} title="Fleet Not Found" />
      </div>
    );
  }

  const status = verification?.status || 'UNKNOWN';
  const basePath = `/fleets/${encodeURIComponent(fleet.id.Owner)}/${encodeURIComponent(fleet.id.Tag)}`;

  return (
    <div className="container" style={{ display: 'flex', flexDirection: 'column', gap: '2rem' }}>
      <FleetHeader fleet={fleet} status={status} />

      <FleetStats fleet={fleet} verification={verification} />

      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(340px, 1fr))', gap: '1.5rem' }}>
        <div className="card">
          <h2 style={{ fontSize: '1rem', fontWeight: 600, marginBottom: '1rem', borderBottom: '1px solid var(--border-color)', paddingBottom: '0.5rem' }}>
            Fleet Identity & Reference
          </h2>

          <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem', fontSize: '0.875rem' }}>
            <div>
              <div style={{ color: 'var(--text-muted)', fontSize: '0.75rem' }}>OWNER CONTRACT ADDRESS</div>
              <div style={{ marginTop: '0.25rem' }}>
                <AddressDisplay address={fleet.id.Owner} chars={8} linkToContract />
              </div>
            </div>

            <div>
              <div style={{ color: 'var(--text-muted)', fontSize: '0.75rem' }}>EXECUTABLE TAG</div>
              <div style={{ marginTop: '0.25rem', fontWeight: 600 }} className="mono">
                {fleet.id.Tag}
              </div>
            </div>

            <div>
              <div style={{ color: 'var(--text-muted)', fontSize: '0.75rem' }}>FIRST SEEN LEDGER</div>
              <div style={{ marginTop: '0.25rem' }}>
                <LedgerLink ledger={fleet.first_seen_ledger} />
              </div>
            </div>

            <div>
              <div style={{ color: 'var(--text-muted)', fontSize: '0.75rem' }}>LAST SEEN LEDGER</div>
              <div style={{ marginTop: '0.25rem' }}>
                <LedgerLink ledger={fleet.last_seen_ledger} />
              </div>
            </div>
          </div>
        </div>

        <div className="card">
          <h2 style={{ fontSize: '1rem', fontWeight: 600, marginBottom: '1rem', borderBottom: '1px solid var(--border-color)', paddingBottom: '0.5rem' }}>
            Current Verification Summary
          </h2>

          <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem', fontSize: '0.875rem' }}>
            <div>
              <div style={{ color: 'var(--text-muted)', fontSize: '0.75rem' }}>STATUS</div>
              <div style={{ marginTop: '0.375rem' }}>
                <FleetStatus status={status} showDescription size="md" />
              </div>
            </div>

            {verification && (
              <>
                <div>
                  <div style={{ color: 'var(--text-muted)', fontSize: '0.75rem' }}>EXPECTED WASM HASH</div>
                  <div style={{ marginTop: '0.25rem' }}>
                    <HashDisplay hash={verification.expected_wasm_hash} chars={8} />
                  </div>
                </div>

                <div>
                  <div style={{ color: 'var(--text-muted)', fontSize: '0.75rem' }}>INDEXED THROUGH LEDGER</div>
                  <div style={{ marginTop: '0.25rem' }}>
                    <LedgerLink ledger={verification.indexed_through} />
                  </div>
                </div>
              </>
            )}

            <div style={{ marginTop: '0.5rem', paddingTop: '0.75rem', borderTop: '1px solid var(--border-color)' }}>
              <Link href={`${basePath}/verify`} className="btn btn-primary" style={{ width: '100%', textAlign: 'center', boxSizing: 'border-box' }}>
                Inspect Full Verification Details →
              </Link>
            </div>
          </div>
        </div>
      </div>

      <div className="card" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: '1rem' }}>
        <div>
          <h3 style={{ fontSize: '1rem', fontWeight: 600 }}>Fleet Membership Explorer</h3>
          <p style={{ fontSize: '0.875rem', color: 'var(--text-secondary)' }}>
            Inspect all {fleet.member_count} deployed contract instances belonging to this fleet.
          </p>
        </div>
        <Link href={`${basePath}/members`} className="btn">
          View All Members →
        </Link>
      </div>
    </div>
  );
}
