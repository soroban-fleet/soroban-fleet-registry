'use client';

import React from 'react';
import Link from 'next/link';
import { usePathname } from 'next/navigation';
import type { Fleet, VerificationStatus } from '../../lib/types';
import { AddressDisplay } from '../common/AddressDisplay';
import { FleetStatus } from './FleetStatus';

export interface FleetHeaderProps {
  fleet: Fleet;
  status?: VerificationStatus;
}

export function FleetHeader({ fleet, status }: FleetHeaderProps) {
  const pathname = usePathname();
  const basePath = `/fleets/${encodeURIComponent(fleet.id.Owner)}/${encodeURIComponent(fleet.id.Tag)}`;

  const tabs = [
    { label: 'Overview', href: basePath },
    { label: 'Members', href: `${basePath}/members` },
    { label: 'History & Releases', href: `${basePath}/history` },
    { label: 'Verification', href: `${basePath}/verify` },
  ];

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '1.5rem' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', flexWrap: 'wrap', gap: '1rem' }}>
        <div>
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', marginBottom: '0.25rem' }}>
            <span style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }} className="mono">
              FLEET TAG:
            </span>
            <span style={{ fontSize: '1.25rem', fontWeight: 800, color: 'var(--text-primary)' }}>
              {fleet.id.Tag}
            </span>
          </div>

          <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', fontSize: '0.875rem' }}>
            <span style={{ color: 'var(--text-secondary)' }}>Owner Contract:</span>
            <AddressDisplay address={fleet.id.Owner} chars={6} linkToContract />
          </div>
        </div>

        {status && (
          <div>
            <FleetStatus status={status} size="lg" />
          </div>
        )}
      </div>

      <nav
        style={{
          display: 'flex',
          gap: '1.5rem',
          borderBottom: '1px solid var(--border-color)',
          paddingBottom: '0.25rem',
        }}
        aria-label="Fleet Sub Navigation"
      >
        {tabs.map((tab) => {
          const isActive = tab.href === basePath ? pathname === basePath : pathname === tab.href;
          return (
            <Link
              key={tab.href}
              href={tab.href}
              style={{
                fontSize: '0.875rem',
                fontWeight: isActive ? 600 : 400,
                color: isActive ? 'var(--text-primary)' : 'var(--text-secondary)',
                borderBottom: isActive ? '2px solid var(--accent)' : '2px solid transparent',
                paddingBottom: '0.5rem',
                textDecoration: 'none',
              }}
            >
              {tab.label}
            </Link>
          );
        })}
      </nav>
    </div>
  );
}
