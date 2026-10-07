import React from 'react';
import Link from 'next/link';
import { getContract } from '../../../lib/api/contracts';
import { AddressDisplay } from '../../../components/common/AddressDisplay';
import { HashDisplay } from '../../../components/common/HashDisplay';
import { LedgerLink } from '../../../components/common/LedgerLink';
import { ErrorState } from '../../../components/common/ErrorState';
import type { ContractInspection } from '../../../lib/types';

interface ContractPageProps {
  params: {
    contractId: string;
  };
}

export default async function ContractPage({ params }: ContractPageProps) {
  const { contractId } = params;

  let inspection: ContractInspection | null = null;
  let fetchError: unknown = null;

  try {
    const res = await getContract(contractId);
    inspection = res.data;
  } catch (err: unknown) {
    fetchError = err;
  }

  if (fetchError || !inspection) {
    return (
      <div className="container">
        <ErrorState error={fetchError} title="Contract Inspection Error" />
      </div>
    );
  }

  const isExternalRef = inspection.resolved?.kind === 'EXTERNAL_REF';
  const isDirectWasm = inspection.resolved?.kind === 'WASM';
  const fleet = inspection.resolved?.fleet || inspection.member?.fleet_id;
  const member = inspection.member;

  return (
    <div className="container" style={{ display: 'flex', flexDirection: 'column', gap: '2rem' }}>
      {/* Header */}
      <div>
        <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }} className="mono">
          CONTRACT INSPECTION
        </div>
        <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', marginTop: '0.25rem' }}>
          <h1 style={{ fontSize: '1.75rem', fontWeight: 800 }}>Contract</h1>
          <AddressDisplay address={inspection.contract_id} chars={8} />
        </div>
      </div>

      {/* Grid of Inspection Details */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(320px, 1fr))', gap: '1.5rem' }}>
        {/* Executable Resolution Card */}
        <div className="card">
          <h2 style={{ fontSize: '1rem', fontWeight: 600, marginBottom: '1rem', borderBottom: '1px solid var(--border-color)', paddingBottom: '0.5rem' }}>
            Resolved Executable State
          </h2>

          <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem', fontSize: '0.875rem' }}>
            <div>
              <div style={{ color: 'var(--text-muted)', fontSize: '0.75rem' }}>EXECUTABLE KIND</div>
              <div style={{ marginTop: '0.375rem' }}>
                {isExternalRef ? (
                  <span
                    style={{
                      padding: '0.2rem 0.5rem',
                      borderRadius: '4px',
                      backgroundColor: 'var(--status-healthy-bg)',
                      color: 'var(--status-healthy-text)',
                      border: '1px solid var(--status-healthy-border)',
                      fontSize: '0.75rem',
                      fontWeight: 700,
                    }}
                    className="mono"
                  >
                    CAP-85 EXTERNAL_REF
                  </span>
                ) : isDirectWasm ? (
                  <span
                    style={{
                      padding: '0.2rem 0.5rem',
                      borderRadius: '4px',
                      backgroundColor: 'var(--bg-tertiary)',
                      color: 'var(--text-secondary)',
                      border: '1px solid var(--border-color)',
                      fontSize: '0.75rem',
                      fontWeight: 600,
                    }}
                    className="mono"
                  >
                    DIRECT WASM
                  </span>
                ) : (
                  <span className="mono">{inspection.resolved?.kind || 'UNKNOWN'}</span>
                )}
              </div>
            </div>

            <div>
              <div style={{ color: 'var(--text-muted)', fontSize: '0.75rem' }}>RESOLVED WASM HASH</div>
              <div style={{ marginTop: '0.25rem' }}>
                <HashDisplay hash={inspection.resolved?.wasm_hash || member?.wasm_hash} chars={8} />
              </div>
            </div>
          </div>
        </div>

        {/* Fleet Association Card */}
        <div className="card">
          <h2 style={{ fontSize: '1rem', fontWeight: 600, marginBottom: '1rem', borderBottom: '1px solid var(--border-color)', paddingBottom: '0.5rem' }}>
            Fleet Membership
          </h2>

          <div style={{ display: 'flex', flexDirection: 'column', gap: '0.75rem', fontSize: '0.875rem' }}>
            {fleet ? (
              <>
                <div>
                  <div style={{ color: 'var(--text-muted)', fontSize: '0.75rem' }}>FLEET OWNER</div>
                  <div style={{ marginTop: '0.25rem' }}>
                    <AddressDisplay address={fleet.Owner} chars={8} linkToContract />
                  </div>
                </div>

                <div>
                  <div style={{ color: 'var(--text-muted)', fontSize: '0.75rem' }}>FLEET TAG</div>
                  <div style={{ marginTop: '0.25rem', fontWeight: 600 }} className="mono">
                    <Link href={`/fleets/${encodeURIComponent(fleet.Owner)}/${encodeURIComponent(fleet.Tag)}`}>
                      {fleet.Tag}
                    </Link>
                  </div>
                </div>

                <div>
                  <div style={{ color: 'var(--text-muted)', fontSize: '0.75rem' }}>ACTIVE STATE</div>
                  <div style={{ marginTop: '0.25rem' }}>
                    {member?.active ? (
                      <span style={{ color: 'var(--status-healthy-text)', fontWeight: 600 }} className="mono">
                        ● Active Member
                      </span>
                    ) : (
                      <span style={{ color: 'var(--text-muted)', fontWeight: 600 }} className="mono">
                        ○ Inactive Member
                      </span>
                    )}
                  </div>
                </div>

                {member && (
                  <>
                    <div>
                      <div style={{ color: 'var(--text-muted)', fontSize: '0.75rem' }}>FIRST SEEN LEDGER</div>
                      <div style={{ marginTop: '0.25rem' }}>
                        <LedgerLink ledger={member.first_seen_ledger} />
                      </div>
                    </div>

                    <div>
                      <div style={{ color: 'var(--text-muted)', fontSize: '0.75rem' }}>LAST SEEN LEDGER</div>
                      <div style={{ marginTop: '0.25rem' }}>
                        <LedgerLink ledger={member.last_seen_ledger} />
                      </div>
                    </div>

                    <div>
                      <div style={{ color: 'var(--text-muted)', fontSize: '0.75rem' }}>LAST VERIFIED LEDGER</div>
                      <div style={{ marginTop: '0.25rem' }}>
                        {member.last_verified_ledger ? (
                          <LedgerLink ledger={member.last_verified_ledger} />
                        ) : (
                          <span style={{ color: 'var(--text-muted)' }}>—</span>
                        )}
                      </div>
                    </div>
                  </>
                )}

                <div style={{ marginTop: '0.5rem', paddingTop: '0.75rem', borderTop: '1px solid var(--border-color)' }}>
                  <Link
                    href={`/fleets/${encodeURIComponent(fleet.Owner)}/${encodeURIComponent(fleet.Tag)}`}
                    className="btn btn-primary"
                    style={{ width: '100%', textAlign: 'center', boxSizing: 'border-box' }}
                  >
                    View Fleet Overview →
                  </Link>
                </div>
              </>
            ) : (
              <div>
                <p style={{ color: 'var(--text-secondary)', marginBottom: '0.75rem' }}>
                  Fleet: <strong>None</strong>
                </p>
                <p style={{ color: 'var(--text-muted)', fontSize: '0.8125rem', lineHeight: 1.5 }}>
                  This contract is executing a standalone executable and is not managed as part of a CAP-85 fleet.
                </p>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
