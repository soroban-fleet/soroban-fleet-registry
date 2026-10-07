import React from 'react';
import type { VerificationStatus } from '../../lib/types';
import { getStatusDescription } from '../../lib/formatting';

export interface FleetStatusProps {
  status: VerificationStatus;
  showDescription?: boolean;
  size?: 'sm' | 'md' | 'lg';
}

export function FleetStatus({
  status,
  showDescription = false,
  size = 'md',
}: FleetStatusProps) {
  let bg = 'var(--status-unknown-bg)';
  let text = 'var(--status-unknown-text)';
  let border = 'var(--status-unknown-border)';
  let symbol = '?';

  switch (status) {
    case 'HEALTHY':
      bg = 'var(--status-healthy-bg)';
      text = 'var(--status-healthy-text)';
      border = 'var(--status-healthy-border)';
      symbol = '✓';
      break;
    case 'DRIFT':
      bg = 'var(--status-drift-bg)';
      text = 'var(--status-drift-text)';
      border = 'var(--status-drift-border)';
      symbol = '⚠';
      break;
    case 'BROKEN_REFERENCE':
      bg = 'var(--status-broken-bg)';
      text = 'var(--status-broken-text)';
      border = 'var(--status-broken-border)';
      symbol = '✕';
      break;
    case 'INCOMPLETE':
      bg = 'var(--status-incomplete-bg)';
      text = 'var(--status-incomplete-text)';
      border = 'var(--status-incomplete-border)';
      symbol = '⏳';
      break;
    case 'UNKNOWN':
    default:
      bg = 'var(--status-unknown-bg)';
      text = 'var(--status-unknown-text)';
      border = 'var(--status-unknown-border)';
      symbol = '?';
      break;
  }

  const padding = size === 'sm' ? '0.2rem 0.5rem' : size === 'lg' ? '0.5rem 1rem' : '0.35rem 0.75rem';
  const fontSize = size === 'sm' ? '0.75rem' : size === 'lg' ? '1rem' : '0.875rem';

  return (
    <div style={{ display: 'inline-flex', flexDirection: 'column', gap: '0.375rem' }}>
      <div
        style={{
          display: 'inline-flex',
          alignItems: 'center',
          gap: '0.5rem',
          backgroundColor: bg,
          color: text,
          border: `1px solid ${border}`,
          borderRadius: '4px',
          padding,
          fontSize,
          fontWeight: 700,
          letterSpacing: '0.025em',
          width: 'fit-content',
        }}
        className="mono"
        role="status"
        aria-label={`Verification status: ${status}`}
      >
        <span aria-hidden="true">{symbol}</span>
        <span>{status}</span>
      </div>

      {showDescription && (
        <p style={{ fontSize: '0.8125rem', color: 'var(--text-secondary)', lineHeight: 1.5, maxWidth: '600px' }}>
          {getStatusDescription(status)}
        </p>
      )}
    </div>
  );
}
