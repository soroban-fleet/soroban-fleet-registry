'use client';

import React, { useState } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import type { VerificationResult, FleetID } from '../../lib/types';
import { FleetStatus } from './FleetStatus';
import { HashDisplay } from '../common/HashDisplay';
import { LedgerLink } from '../common/LedgerLink';

export interface VerificationPanelProps {
  fleetId: FleetID;
  verification: VerificationResult;
}

export function VerificationPanel({ fleetId, verification }: VerificationPanelProps) {
  const router = useRouter();
  const [customWasm, setCustomWasm] = useState('');
  const [isVerifying, setIsVerifying] = useState(false);

  const membersHref = `/fleets/${encodeURIComponent(fleetId.Owner)}/${encodeURIComponent(fleetId.Tag)}/members`;
  const basePath = `/fleets/${encodeURIComponent(fleetId.Owner)}/${encodeURIComponent(fleetId.Tag)}/verify`;

  const handleCustomVerify = (e: React.FormEvent) => {
    e.preventDefault();
    if (customWasm.trim()) {
      setIsVerifying(true);
      router.push(`${basePath}?expected_wasm=${encodeURIComponent(customWasm.trim())}`);
    } else {
      router.push(basePath);
    }
  };

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '2rem' }}>
      {/* 1. Status Banner */}
      <div
        className="card"
        style={{
          display: 'flex',
          flexDirection: 'column',
          gap: '1rem',
          borderLeft: '4px solid',
          borderLeftColor:
            verification.status === 'HEALTHY'
              ? 'var(--status-healthy-border)'
              : verification.status === 'DRIFT'
              ? 'var(--status-drift-border)'
              : verification.status === 'BROKEN_REFERENCE'
              ? 'var(--status-broken-border)'
              : verification.status === 'INCOMPLETE'
              ? 'var(--status-incomplete-border)'
              : 'var(--border-color)',
        }}
      >
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: '1rem' }}>
          <div>
            <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)', marginBottom: '0.25rem' }}>
              VERIFICATION STATUS
            </div>
            <FleetStatus status={verification.status} size="lg" showDescription />
          </div>

          <div style={{ textAlign: 'right' }}>
            <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>VERIFIED AT</div>
            <div style={{ fontSize: '0.875rem', color: 'var(--text-secondary)' }} className="mono">
              {new Date(verification.verified_at).toLocaleString()}
            </div>
          </div>
        </div>

        {/* Diagnostic alert for DRIFT */}
        {verification.status === 'DRIFT' && (
          <div
            style={{
              padding: '1rem',
              backgroundColor: 'var(--status-drift-bg)',
              borderRadius: '4px',
              border: '1px solid var(--status-drift-border)',
              display: 'flex',
              flexDirection: 'column',
              gap: '0.75rem',
            }}
          >
            <div style={{ fontWeight: 600, color: 'var(--status-drift-text)', fontSize: '0.9375rem' }}>
              ⚠ Executable Drift Detected
            </div>
            <p style={{ fontSize: '0.875rem', color: 'var(--text-primary)', lineHeight: 1.5 }}>
              {verification.mismatching_members.toLocaleString()} of {verification.total_members.toLocaleString()} active member contracts resolve to an executable different from the expected WASM hash.
            </p>
            <div>
              <Link href={membersHref} className="btn btn-primary" style={{ fontSize: '0.8125rem' }}>
                Investigate Divergent Members in Member List →
              </Link>
            </div>
          </div>
        )}

        {/* Diagnostic alert for BROKEN_REFERENCE */}
        {verification.status === 'BROKEN_REFERENCE' && (
          <div
            style={{
              padding: '1rem',
              backgroundColor: 'var(--status-broken-bg)',
              borderRadius: '4px',
              border: '1px solid var(--status-broken-border)',
              display: 'flex',
              flexDirection: 'column',
              gap: '0.5rem',
            }}
          >
            <div style={{ fontWeight: 600, color: 'var(--status-broken-text)', fontSize: '0.9375rem' }}>
              ✕ Broken Executable Reference
            </div>
            <p style={{ fontSize: '0.875rem', color: 'var(--text-primary)', lineHeight: 1.5 }}>
              The CAP-85 executable reference cannot be resolved on-chain. The owner contract may not exist or does not contain an entry for tag &ldquo;{fleetId.Tag}&rdquo;.
            </p>
          </div>
        )}

        {/* Diagnostic alert for INCOMPLETE */}
        {verification.status === 'INCOMPLETE' && (
          <div
            style={{
              padding: '1rem',
              backgroundColor: 'var(--status-incomplete-bg)',
              borderRadius: '4px',
              border: '1px solid var(--status-incomplete-border)',
              display: 'flex',
              flexDirection: 'column',
              gap: '0.5rem',
            }}
          >
            <div style={{ fontWeight: 600, color: 'var(--status-incomplete-text)', fontSize: '0.9375rem' }}>
              ⏳ Indexer History Incomplete
            </div>
            <p style={{ fontSize: '0.875rem', color: 'var(--text-primary)', lineHeight: 1.5 }}>
              The indexer has not caught up to the latest on-chain ledger or some contract instances could not be resolved.
            </p>
          </div>
        )}
      </div>

      {/* 2. Verification Metrics Grid */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(220px, 1fr))', gap: '1.25rem' }}>
        <div className="card">
          <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)', marginBottom: '0.25rem' }}>
            EXPECTED WASM
          </div>
          <HashDisplay hash={verification.expected_wasm_hash} chars={6} />
          <div style={{ fontSize: '0.75rem', color: 'var(--text-secondary)', marginTop: '0.5rem' }}>
            Target executable hash
          </div>
        </div>

        <div className="card">
          <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)', marginBottom: '0.25rem' }}>
            TOTAL MEMBERS
          </div>
          <div style={{ fontSize: '1.75rem', fontWeight: 800 }} className="mono">
            {verification.total_members.toLocaleString()}
          </div>
          <div style={{ fontSize: '0.75rem', color: 'var(--text-secondary)', marginTop: '0.25rem' }}>
            Scope of verification
          </div>
        </div>

        <div className="card">
          <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)', marginBottom: '0.25rem' }}>
            MATCHING MEMBERS
          </div>
          <div style={{ fontSize: '1.75rem', fontWeight: 800, color: 'var(--status-healthy-text)' }} className="mono">
            {verification.matching_members.toLocaleString()}
          </div>
          <div style={{ fontSize: '0.75rem', color: 'var(--text-secondary)', marginTop: '0.25rem' }}>
            Contracts with expected WASM
          </div>
        </div>

        <div className="card">
          <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)', marginBottom: '0.25rem' }}>
            MISMATCHING MEMBERS
          </div>
          <div
            style={{
              fontSize: '1.75rem',
              fontWeight: 800,
              color: verification.mismatching_members > 0 ? 'var(--status-drift-text)' : 'inherit',
            }}
            className="mono"
          >
            {verification.mismatching_members.toLocaleString()}
          </div>
          <div style={{ fontSize: '0.75rem', color: 'var(--text-secondary)', marginTop: '0.25rem' }}>
            Contracts with divergent WASM
          </div>
        </div>

        <div className="card">
          <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)', marginBottom: '0.25rem' }}>
            MISSING / UNRESOLVED
          </div>
          <div
            style={{
              fontSize: '1.75rem',
              fontWeight: 800,
              color: verification.missing_members > 0 ? 'var(--status-broken-text)' : 'inherit',
            }}
            className="mono"
          >
            {verification.missing_members.toLocaleString()}
          </div>
          <div style={{ fontSize: '0.75rem', color: 'var(--text-secondary)', marginTop: '0.25rem' }}>
            Contracts failing resolution
          </div>
        </div>

        <div className="card">
          <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)', marginBottom: '0.25rem' }}>
            INDEXED THROUGH LEDGER
          </div>
          <div style={{ fontSize: '1.25rem', fontWeight: 700 }} className="mono">
            <LedgerLink ledger={verification.indexed_through} />
          </div>
          <div style={{ fontSize: '0.75rem', color: 'var(--text-secondary)', marginTop: '0.25rem' }}>
            Ledger snapshot height
          </div>
        </div>
      </div>

      {/* 3. Custom Verification Audit Form */}
      <div className="card">
        <h3 style={{ fontSize: '1rem', fontWeight: 600, marginBottom: '0.5rem' }}>
          Evaluate Candidate WASM Hash
        </h3>
        <p style={{ fontSize: '0.875rem', color: 'var(--text-secondary)', marginBottom: '1rem' }}>
          Maintainers can test fleet readiness against a planned upgrade by evaluating all member instances against a candidate WASM hash without modifying state.
        </p>

        <form onSubmit={handleCustomVerify} style={{ display: 'flex', gap: '0.75rem', flexWrap: 'wrap' }}>
          <input
            type="text"
            placeholder="Candidate 64-char hex WASM hash..."
            value={customWasm}
            onChange={(e) => setCustomWasm(e.target.value)}
            style={{
              flex: 1,
              minWidth: '280px',
              padding: '0.625rem 0.875rem',
              borderRadius: '4px',
              border: '1px solid var(--border-color)',
              backgroundColor: 'var(--bg-tertiary)',
              color: 'var(--text-primary)',
              fontSize: '0.875rem',
            }}
            className="mono"
          />
          <button type="submit" className="btn btn-primary" disabled={isVerifying}>
            {isVerifying ? 'Verifying...' : 'Verify Candidate WASM'}
          </button>
          {customWasm && (
            <button
              type="button"
              className="btn"
              onClick={() => {
                setCustomWasm('');
                router.push(basePath);
              }}
            >
              Reset to Current WASM
            </button>
          )}
        </form>
      </div>
    </div>
  );
}
