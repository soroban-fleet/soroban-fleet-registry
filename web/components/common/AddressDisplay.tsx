'use client';

import React, { useState } from 'react';
import Link from 'next/link';
import { formatAddress } from '../../lib/formatting';

export interface AddressDisplayProps {
  address: string;
  linkToContract?: boolean;
  chars?: number;
}

export function AddressDisplay({
  address,
  linkToContract = false,
  chars = 6,
}: AddressDisplayProps) {
  const [copied, setCopied] = useState(false);

  const handleCopy = async (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
    try {
      await navigator.clipboard.writeText(address);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {
      // fallback
    }
  };

  const truncated = formatAddress(address, chars);

  return (
    <span
      style={{
        display: 'inline-flex',
        alignItems: 'center',
        gap: '0.375rem',
        fontSize: '0.875rem',
      }}
      className="mono"
    >
      {linkToContract ? (
        <Link
          href={`/contracts/${encodeURIComponent(address)}`}
          title={address}
          style={{ textDecoration: 'none' }}
        >
          {truncated}
        </Link>
      ) : (
        <span title={address}>{truncated}</span>
      )}

      <button
        type="button"
        onClick={handleCopy}
        aria-label={copied ? 'Copied address' : 'Copy address'}
        title={copied ? 'Copied!' : 'Copy full address'}
        style={{
          background: 'none',
          border: '1px solid var(--border-color)',
          borderRadius: '3px',
          padding: '0.125rem 0.375rem',
          fontSize: '0.75rem',
          color: copied ? 'var(--status-healthy-text)' : 'var(--text-muted)',
          cursor: 'pointer',
        }}
      >
        {copied ? 'Copied' : 'Copy'}
      </button>
    </span>
  );
}
