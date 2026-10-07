import React from 'react';

export function Footer() {
  return (
    <footer
      style={{
        borderTop: '1px solid var(--border-color)',
        backgroundColor: 'var(--bg-secondary)',
        padding: '1.75rem 0',
        marginTop: 'auto',
      }}
    >
      <div
        className="container"
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          fontSize: '0.8125rem',
          color: 'var(--text-muted)',
          flexWrap: 'wrap',
          gap: '1rem',
        }}
      >
        <div>
          <span style={{ fontWeight: 600, color: 'var(--text-secondary)' }}>
            Soroban Fleet Registry
          </span>{' '}
          · CAP-85 Externally Managed Contract Executables Indexer & Verifier
        </div>
        <div style={{ display: 'flex', gap: '1.25rem' }}>
          <a
            href="https://github.com/soroban-fleet/soroban-fleet-registry"
            target="_blank"
            rel="noopener noreferrer"
            style={{ color: 'var(--text-secondary)' }}
          >
            GitHub
          </a>
          <a
            href="https://github.com/stellar/stellar-protocol/blob/master/core/cap-0085.md"
            target="_blank"
            rel="noopener noreferrer"
            style={{ color: 'var(--text-secondary)' }}
          >
            CAP-85 Spec
          </a>
        </div>
      </div>
    </footer>
  );
}
