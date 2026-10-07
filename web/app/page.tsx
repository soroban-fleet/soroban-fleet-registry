import React from 'react';

export default function HomePage() {
  return (
    <div className="container" style={{ display: 'flex', flexDirection: 'column', gap: '2.5rem' }}>
      <section style={{ maxWidth: '800px', display: 'flex', flexDirection: 'column', gap: '1.25rem' }}>
        <div style={{ display: 'inline-flex', alignItems: 'center', gap: '0.5rem', width: 'fit-content', padding: '0.25rem 0.75rem', borderRadius: '9999px', backgroundColor: 'var(--bg-tertiary)', border: '1px solid var(--border-color)', fontSize: '0.75rem', color: 'var(--accent)' }} className="mono">
          CAP-85 Observer & Verification Engine
        </div>
        <h1 style={{ fontSize: '2.5rem', fontWeight: 800, letterSpacing: '-0.025em', lineHeight: 1.15 }}>
          Soroban Fleet Registry
        </h1>
        <p style={{ fontSize: '1.25rem', color: 'var(--text-secondary)', lineHeight: 1.5 }}>
          Discover and verify CAP-85 executable fleets.
        </p>
        <div style={{ display: 'flex', gap: '1rem', marginTop: '0.5rem' }}>
          <a href="/fleets" className="btn btn-primary" style={{ padding: '0.75rem 1.5rem', fontSize: '1rem' }}>
            Browse fleets
          </a>
          <a
            href="https://github.com/soroban-fleet/soroban-fleet-registry/blob/main/docs/cap85.md"
            target="_blank"
            rel="noopener noreferrer"
            className="btn"
            style={{ padding: '0.75rem 1.5rem', fontSize: '1rem' }}
          >
            Protocol Spec
          </a>
        </div>
      </section>

      <section className="card" style={{ maxWidth: '800px' }}>
        <h2 style={{ fontSize: '1.125rem', marginBottom: '0.75rem', fontWeight: 600 }}>
          What is a Fleet?
        </h2>
        <p style={{ color: 'var(--text-secondary)', marginBottom: '1rem', lineHeight: 1.6 }}>
          A fleet is a group of Soroban contracts sharing an externally managed executable reference (<code>owner</code>, <code>tag</code>). The owner contract stores the WASM hash associated with that tag, enabling coordinated upgrades across hundreds or thousands of instances.
        </p>
        <p style={{ color: 'var(--text-secondary)', lineHeight: 1.6 }}>
          This application provides complete operational visibility: fleet membership, active WASM hashes, release history, and deterministic verification against live on-chain ledger state.
        </p>
      </section>

      <section style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(260px, 1fr))', gap: '1.5rem' }}>
        <div className="card">
          <h3 style={{ fontSize: '1rem', fontWeight: 600, marginBottom: '0.5rem', color: 'var(--accent)' }}>
            Discovery & Membership
          </h3>
          <p style={{ fontSize: '0.875rem', color: 'var(--text-secondary)', lineHeight: 1.5 }}>
            Index and track contract instances deployed with CAP-85 external references across all historical ledgers.
          </p>
        </div>

        <div className="card">
          <h3 style={{ fontSize: '1rem', fontWeight: 600, marginBottom: '0.5rem', color: 'var(--status-healthy-text)' }}>
            Deterministic Verification
          </h3>
          <p style={{ fontSize: '0.875rem', color: 'var(--text-secondary)', lineHeight: 1.5 }}>
            Continuously verify whether all active members resolve to the expected executable or if executable drift has occurred.
          </p>
        </div>

        <div className="card">
          <h3 style={{ fontSize: '1rem', fontWeight: 600, marginBottom: '0.5rem', color: 'var(--status-drift-text)' }}>
            Release History
          </h3>
          <p style={{ fontSize: '0.875rem', color: 'var(--text-secondary)', lineHeight: 1.5 }}>
            Audit coordinated upgrades, transaction hashes, ledger numbers, and previous WASM hashes over time.
          </p>
        </div>
      </section>
    </div>
  );
}
