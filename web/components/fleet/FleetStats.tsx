import React from 'react';
import type { Fleet, VerificationResult } from '../../lib/types';
import { HashDisplay } from '../common/HashDisplay';
import { LedgerLink } from '../common/LedgerLink';

export interface FleetStatsProps {
  fleet: Fleet;
  verification?: VerificationResult;
}

export function FleetStats({ fleet, verification }: FleetStatsProps) {
  return (
    <div
      style={{
        display: 'grid',
        gridTemplateColumns: 'repeat(auto-fit, minmax(220px, 1fr))',
        gap: '1.25rem',
      }}
    >
      <div className="card" style={{ padding: '1.25rem' }}>
        <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)', marginBottom: '0.375rem' }}>
          MEMBER COUNT
        </div>
        <div style={{ fontSize: '1.75rem', fontWeight: 800 }} className="mono">
          {fleet.member_count.toLocaleString()}
        </div>
        <div style={{ fontSize: '0.75rem', color: 'var(--text-secondary)', marginTop: '0.25rem' }}>
          Active contracts in fleet
        </div>
      </div>

      <div className="card" style={{ padding: '1.25rem' }}>
        <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)', marginBottom: '0.375rem' }}>
          ACTIVE WASM HASH
        </div>
        <div style={{ marginTop: '0.5rem' }}>
          <HashDisplay hash={fleet.current_wasm_hash} chars={6} />
        </div>
        <div style={{ fontSize: '0.75rem', color: 'var(--text-secondary)', marginTop: '0.5rem' }}>
          Current target executable
        </div>
      </div>

      <div className="card" style={{ padding: '1.25rem' }}>
        <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)', marginBottom: '0.375rem' }}>
          INDEXER COVERAGE
        </div>
        <div style={{ fontSize: '1.25rem', fontWeight: 700 }} className="mono">
          <LedgerLink ledger={fleet.last_indexed_ledger} />
        </div>
        <div style={{ fontSize: '0.75rem', color: 'var(--text-secondary)', marginTop: '0.25rem' }}>
          First seen at ledger <LedgerLink ledger={fleet.first_seen_ledger} />
        </div>
      </div>

      <div className="card" style={{ padding: '1.25rem' }}>
        <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)', marginBottom: '0.375rem' }}>
          VERIFICATION SCOPE
        </div>
        {verification ? (
          <div>
            <div style={{ fontSize: '1.25rem', fontWeight: 700 }} className="mono">
              <span style={{ color: 'var(--status-healthy-text)' }}>
                {verification.matching_members}
              </span>
              <span style={{ color: 'var(--text-muted)' }}> / </span>
              <span>{verification.total_members}</span>
            </div>
            <div style={{ fontSize: '0.75rem', color: 'var(--text-secondary)', marginTop: '0.25rem' }}>
              {verification.mismatching_members > 0 ? (
                <span style={{ color: 'var(--status-drift-text)' }}>
                  {verification.mismatching_members} mismatching members
                </span>
              ) : (
                'All active members match'
              )}
            </div>
          </div>
        ) : (
          <div style={{ color: 'var(--text-muted)', fontSize: '0.875rem' }}>
            No verification recorded
          </div>
        )}
      </div>
    </div>
  );
}
