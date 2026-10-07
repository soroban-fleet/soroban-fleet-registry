'use client';

import React from 'react';
import Link from 'next/link';
import { usePathname } from 'next/navigation';

export function Navigation() {
  const pathname = usePathname();

  const navItems = [
    { label: 'Overview', href: '/' },
    { label: 'Fleets', href: '/fleets' },
  ];

  return (
    <nav style={{ display: 'flex', gap: '1.25rem', alignItems: 'center' }} aria-label="Main Navigation">
      {navItems.map((item) => {
        const isActive = item.href === '/' ? pathname === '/' : pathname?.startsWith(item.href);
        return (
          <Link
            key={item.href}
            href={item.href}
            style={{
              fontSize: '0.875rem',
              fontWeight: isActive ? 600 : 400,
              color: isActive ? 'var(--text-primary)' : 'var(--text-secondary)',
              borderBottom: isActive ? '2px solid var(--accent)' : '2px solid transparent',
              padding: '0.5rem 0',
              textDecoration: 'none',
              transition: 'color 0.15s ease-in-out',
            }}
          >
            {item.label}
          </Link>
        );
      })}
      <a
        href="https://github.com/soroban-fleet/soroban-fleet-registry/tree/main/docs"
        target="_blank"
        rel="noopener noreferrer"
        style={{
          fontSize: '0.875rem',
          color: 'var(--text-secondary)',
          textDecoration: 'none',
          padding: '0.5rem 0',
        }}
      >
        Documentation ↗
      </a>
    </nav>
  );
}
