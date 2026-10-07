import React from 'react';
import Link from 'next/link';
import type { Fleet } from '../../lib/types';
import { AddressDisplay } from '../common/AddressDisplay';
import { HashDisplay } from '../common/HashDisplay';
import { LedgerLink } from '../common/LedgerLink';

export interface FleetCardProps {
  fleet: Fleet;
}

export function FleetCard({ fleet }: FleetCardProps) {
  const detailHref = `/fleets/${encodeURIComponent(fleet.id.Owner)}/${encodeURIComponent(fleet.id.Tag)}`;

  return (
    <div
      className="card"
      style={{
        display: 'flex',
        flexDirection: 'column',
        gap: '1rem',
        transition: 'border-color 0.15s ease-in-out',
      }}
    >
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
        <div>
          <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }} className="mono">
            TAG
          </div>
          <Link
            href={detailHref}
            style={{
              fontSize: '1.125rem',
              fontWeight: 700,
              color: 'var(--text-primary)',
              textDecoration: 'none',
            }}
          >
            {fleet.id.Tag}
          </Link>
        </div>

        <div style={{ textAlign: 'right' }}>
          <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>MEMBERS</div>
          <div style={{ fontSize: '1.125rem', fontWeight: 700 }} className="mono">
            {fleet.member_count.toLocaleString()}
          </div>
        </div>
      </div>

      <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem', fontSize: '0.8125rem' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <span style={{ color: 'var(--text-secondary)' }}>Owner:</span>
          <AddressDisplay address={fleet.id.Owner} chars={4} />
        </div>

        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <span style={{ color: 'var(--text-secondary)' }}>Current WASM:</span>
          <HashDisplay hash={fleet.current_wasm_hash} chars={4} />
        </div>

        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <span style={{ color: 'var(--text-secondary)' }}>Last Indexed:</span>
          <LedgerLink ledger={fleet.last_indexed_ledger} />
        </div>
      </div>

      <div style={{ marginTop: 'auto', paddingTop: '0.75rem', borderTop: '1px solid var(--border-color)' }}>
        <Link
          href={detailHref}
          className="btn"
          style={{ width: '100%', textAlign: 'center', boxSizing: 'border-box' }}
        >
          View Fleet Details →
        </Link>
      </div>
    </div>
  );
}
