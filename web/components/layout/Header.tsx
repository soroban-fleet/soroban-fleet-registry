import React from 'react';
import Link from 'next/link';
import { Navigation } from './Navigation';
import { config } from '../../lib/config';

export function Header() {
  return (
    <header
      style={{
        borderBottom: '1px solid var(--border-color)',
        backgroundColor: 'var(--bg-secondary)',
        position: 'sticky',
        top: 0,
        zIndex: 50,
      }}
    >
      <div
        className="container"
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          height: '64px',
        }}
      >
        <div style={{ display: 'flex', alignItems: 'center', gap: '2rem' }}>
          <Link
            href="/"
            style={{
              fontSize: '1.125rem',
              fontWeight: 700,
              color: 'var(--text-primary)',
              textDecoration: 'none',
              display: 'flex',
              alignItems: 'center',
              gap: '0.625rem',
            }}
          >
            <span
              style={{
                backgroundColor: 'var(--accent)',
                color: '#030712',
                padding: '0.125rem 0.375rem',
                borderRadius: '4px',
                fontSize: '0.75rem',
                fontWeight: 800,
              }}
              className="mono"
            >
              SFR
            </span>
            <span>Soroban Fleet Registry</span>
          </Link>

          <Navigation />
        </div>

        <div style={{ display: 'flex', alignItems: 'center', gap: '1rem' }}>
          <div
            style={{
              display: 'inline-flex',
              alignItems: 'center',
              gap: '0.375rem',
              fontSize: '0.75rem',
              padding: '0.25rem 0.5rem',
              borderRadius: '4px',
              backgroundColor: 'var(--bg-tertiary)',
              border: '1px solid var(--border-color)',
              color: 'var(--text-secondary)',
            }}
            className="mono"
          >
            <span
              style={{
                width: '6px',
                height: '6px',
                borderRadius: '50%',
                backgroundColor: 'var(--status-healthy-text)',
              }}
            />
            <span>{config.network}</span>
          </div>
          <span style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }} className="mono">
            Protocol 28
          </span>
        </div>
      </div>
    </header>
  );
}
