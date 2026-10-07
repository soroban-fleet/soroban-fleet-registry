import React from 'react';
import { getExplorerLedgerUrl, getExplorerTxUrl, formatHash } from '../../lib/formatting';

export interface LedgerLinkProps {
  ledger?: number | bigint;
  txHash?: string;
  truncateTx?: boolean;
}

export function LedgerLink({ ledger, txHash, truncateTx = true }: LedgerLinkProps) {
  if (ledger !== undefined) {
    const ledgerUrl = getExplorerLedgerUrl(ledger);
    if (ledgerUrl) {
      return (
        <a
          href={ledgerUrl}
          target="_blank"
          rel="noopener noreferrer"
          className="mono"
          style={{ textDecoration: 'none', color: 'var(--accent)' }}
          title={`View ledger #${ledger} in explorer`}
        >
          #{ledger.toString()} ↗
        </a>
      );
    }
    return <span className="mono">#{ledger.toString()}</span>;
  }

  if (txHash) {
    const txUrl = getExplorerTxUrl(txHash);
    const displayTx = truncateTx ? formatHash(txHash, 6) : txHash;

    if (txUrl) {
      return (
        <a
          href={txUrl}
          target="_blank"
          rel="noopener noreferrer"
          className="mono"
          style={{ textDecoration: 'none', color: 'var(--accent)' }}
          title={`View transaction ${txHash} in explorer`}
        >
          {displayTx} ↗
        </a>
      );
    }
    return (
      <span className="mono" title={txHash}>
        {displayTx}
      </span>
    );
  }

  return <span>—</span>;
}
