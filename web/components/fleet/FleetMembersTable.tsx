import React from 'react';
import type { FleetMember } from '../../lib/types';
import { AddressDisplay } from '../common/AddressDisplay';
import { HashDisplay } from '../common/HashDisplay';
import { LedgerLink } from '../common/LedgerLink';
import { EmptyState } from '../common/EmptyState';

export interface FleetMembersTableProps {
  members: FleetMember[];
}

export function FleetMembersTable({ members }: FleetMembersTableProps) {
  if (members.length === 0) {
    return (
      <EmptyState
        title="No Members"
        message="No active members were returned for this fleet."
      />
    );
  }

  return (
    <div className="table-container">
      <table>
        <thead>
          <tr>
            <th>Contract ID</th>
            <th>WASM Hash</th>
            <th>State</th>
            <th>First Seen</th>
            <th>Last Seen</th>
            <th>Last Verified</th>
          </tr>
        </thead>
        <tbody>
          {members.map((member) => (
            <tr key={member.contract_id}>
              <td>
                <AddressDisplay address={member.contract_id} chars={6} linkToContract />
              </td>
              <td>
                <HashDisplay hash={member.wasm_hash} chars={6} />
              </td>
              <td>
                {member.active ? (
                  <span
                    style={{
                      display: 'inline-flex',
                      alignItems: 'center',
                      gap: '0.25rem',
                      padding: '0.15rem 0.5rem',
                      borderRadius: '4px',
                      fontSize: '0.75rem',
                      fontWeight: 600,
                      backgroundColor: 'var(--status-healthy-bg)',
                      color: 'var(--status-healthy-text)',
                      border: '1px solid var(--status-healthy-border)',
                    }}
                    className="mono"
                  >
                    ● ACTIVE
                  </span>
                ) : (
                  <span
                    style={{
                      display: 'inline-flex',
                      alignItems: 'center',
                      gap: '0.25rem',
                      padding: '0.15rem 0.5rem',
                      borderRadius: '4px',
                      fontSize: '0.75rem',
                      fontWeight: 600,
                      backgroundColor: 'var(--bg-tertiary)',
                      color: 'var(--text-muted)',
                      border: '1px solid var(--border-color)',
                    }}
                    className="mono"
                  >
                    ○ INACTIVE
                  </span>
                )}
              </td>
              <td>
                <LedgerLink ledger={member.first_seen_ledger} />
              </td>
              <td>
                <LedgerLink ledger={member.last_seen_ledger} />
              </td>
              <td>
                {member.last_verified_ledger ? (
                  <LedgerLink ledger={member.last_verified_ledger} />
                ) : (
                  <span style={{ color: 'var(--text-muted)' }}>—</span>
                )}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
