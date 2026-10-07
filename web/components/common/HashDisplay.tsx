'use client';

import React, { useState } from 'react';
import { formatHash, normalizeWasmHash } from '../../lib/formatting';

export interface HashDisplayProps {
  hash: string | undefined | null;
  chars?: number;
}

export function HashDisplay({ hash, chars = 6 }: HashDisplayProps) {
  const [copied, setCopied] = useState(false);

  if (!hash) {
    return <span style={{ color: 'var(--text-muted)' }}>—</span>;
  }

  const normalized = normalizeWasmHash(hash);
  const truncated = formatHash(normalized, chars);

  const handleCopy = async (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
    try {
      await navigator.clipboard.writeText(normalized);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {
      // fallback
    }
  };

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
      <span title={normalized}>{truncated}</span>

      <button
        type="button"
        onClick={handleCopy}
        aria-label={copied ? 'Copied WASM hash' : 'Copy WASM hash'}
        title={copied ? 'Copied!' : 'Copy full WASM hash'}
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
