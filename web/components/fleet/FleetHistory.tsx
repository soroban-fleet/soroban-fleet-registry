import React from 'react';
import type { FleetRelease, VerificationResult } from '../../lib/types';
import { HashDisplay } from '../common/HashDisplay';
import { LedgerLink } from '../common/LedgerLink';
import { FleetStatus } from './FleetStatus';
import { EmptyState } from '../common/EmptyState';

export interface FleetHistoryProps {
  releases: FleetRelease[];
  verifications: VerificationResult[];
}

export function FleetHistory({ releases, verifications }: FleetHistoryProps) {
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '2.5rem' }}>
      {/* 1. Coordinated Releases Section */}
      <div>
        <div style={{ marginBottom: '1rem' }}>
          <h3 style={{ fontSize: '1.125rem', fontWeight: 600 }}>Observed Executable Releases</h3>
          <p style={{ fontSize: '0.875rem', color: 'var(--text-secondary)' }}>
            On-chain upgrades observed where the fleet owner updated the WASM hash associated with this tag.
          </p>
          <div
            style={{
              marginTop: '0.5rem',
              padding: '0.5rem 0.75rem',
              backgroundColor: 'var(--bg-tertiary)',
              borderRadius: '4px',
              fontSize: '0.75rem',
              color: 'var(--text-muted)',
              border: '1px solid var(--border-color)',
            }}
          >
            <strong>Note:</strong> A release observation records when the owner contract tag changed WASM. It does not automatically imply that all active fleet members have successfully verified against the new WASM.
          </div>
        </div>

        {releases.length === 0 ? (
          <EmptyState
            title="No Releases Observed"
            message="No historical contract executable changes have been recorded for this fleet."
          />
        ) : (
          <div className="table-container">
            <table>
              <thead>
                <tr>
                  <th>Release</th>
                  <th>Ledger</th>
                  <th>Old WASM</th>
                  <th>New WASM</th>
                  <th>Transaction</th>
                  <th>Observed At</th>
                </tr>
              </thead>
              <tbody>
                {releases.map((release) => (
                  <tr key={release.id}>
                    <td className="mono" style={{ fontWeight: 600 }}>
                      #{release.id}
                    </td>
                    <td>
                      <LedgerLink ledger={release.ledger} />
                    </td>
                    <td>
                      <HashDisplay hash={release.old_wasm_hash} chars={6} />
                    </td>
                    <td>
                      <HashDisplay hash={release.new_wasm_hash} chars={6} />
                    </td>
                    <td>
                      <LedgerLink txHash={release.tx_hash} />
                    </td>
                    <td style={{ fontSize: '0.75rem', color: 'var(--text-secondary)' }}>
                      {new Date(release.observed_at).toLocaleString()}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* 2. Verification History Section */}
      <div>
        <div style={{ marginBottom: '1rem' }}>
          <h3 style={{ fontSize: '1.125rem', fontWeight: 600 }}>Verification Audit History</h3>
          <p style={{ fontSize: '0.875rem', color: 'var(--text-secondary)' }}>
            Historical verification snapshots evaluating consistency between active fleet members and expected WASM.
          </p>
        </div>

        {verifications.length === 0 ? (
          <EmptyState
            title="No Verification History"
            message="No historical fleet verification snapshots are available."
          />
        ) : (
          <div className="table-container">
            <table>
              <thead>
                <tr>
                  <th>Status</th>
                  <th>Scope</th>
                  <th>Mismatching</th>
                  <th>Missing</th>
                  <th>Indexed Through</th>
                  <th>Verified At</th>
                </tr>
              </thead>
              <tbody>
                {verifications.map((v, idx) => (
                  <tr key={v.id || idx}>
                    <td>
                      <FleetStatus status={v.status} size="sm" />
                    </td>
                    <td className="mono">
                      {v.matching_members} / {v.total_members}
                    </td>
                    <td className="mono" style={{ color: v.mismatching_members > 0 ? 'var(--status-drift-text)' : 'inherit' }}>
                      {v.mismatching_members}
                    </td>
                    <td className="mono">
                      {v.missing_members}
                    </td>
                    <td>
                      <LedgerLink ledger={v.indexed_through} />
                    </td>
                    <td style={{ fontSize: '0.75rem', color: 'var(--text-secondary)' }}>
                      {new Date(v.verified_at).toLocaleString()}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
}
