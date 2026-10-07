import { config } from '../config';
import type { VerificationStatus } from '../types';

/**
 * Converts a base64 encoded byte string (standard Go JSON byte slice serialization)
 * into a canonical lowercase hex string.
 * If the input is already a hex string, it is returned normalized.
 */
export function normalizeWasmHash(hash: string | undefined | null): string {
  if (!hash) return '';
  // Check if string is already 64-char hex
  if (/^[0-9a-fA-F]{64}$/.test(hash)) {
    return hash.toLowerCase();
  }

  try {
    // If base64
    if (typeof window !== 'undefined') {
      const binary = atob(hash);
      const bytes = new Uint8Array(binary.length);
      for (let i = 0; i < binary.length; i++) {
        bytes[i] = binary.charCodeAt(i);
      }
      return Array.from(bytes)
        .map((b) => b.toString(16).padStart(2, '0'))
        .join('');
    } else {
      // Node.js server side
      const buf = Buffer.from(hash, 'base64');
      return buf.toString('hex');
    }
  } catch {
    return hash;
  }
}

/**
 * Shortens an address for clean UI display while preserving start and end segments.
 */
export function formatAddress(address: string, chars: number = 4): string {
  if (!address) return '';
  if (address.length <= chars * 2 + 3) return address;
  return `${address.slice(0, chars + 2)}...${address.slice(-chars)}`;
}

/**
 * Shortens a hash for clean UI display.
 */
export function formatHash(hash: string, chars: number = 4): string {
  if (!hash) return '';
  const normalized = normalizeWasmHash(hash);
  if (normalized.length <= chars * 2 + 3) return normalized;
  return `${normalized.slice(0, chars + 2)}...${normalized.slice(-chars)}`;
}

/**
 * Exact descriptions matching protocol and backend verification semantics.
 */
export function getStatusDescription(status: VerificationStatus): string {
  switch (status) {
    case 'HEALTHY':
      return 'All indexed active members resolve to the expected executable, and indexing is complete for the verification scope.';
    case 'DRIFT':
      return 'One or more active members resolve to a different executable.';
    case 'BROKEN_REFERENCE':
      return 'The fleet executable reference cannot be resolved correctly.';
    case 'INCOMPLETE':
      return 'The indexer has not processed enough ledger history to make a complete verification claim.';
    case 'UNKNOWN':
    default:
      return 'The registry does not currently have enough information to determine the fleet state.';
  }
}

export function getStatusBadgeClass(status: VerificationStatus): string {
  switch (status) {
    case 'HEALTHY':
      return 'badge-healthy';
    case 'DRIFT':
      return 'badge-drift';
    case 'BROKEN_REFERENCE':
      return 'badge-broken';
    case 'INCOMPLETE':
      return 'badge-incomplete';
    case 'UNKNOWN':
    default:
      return 'badge-unknown';
  }
}

/**
 * Constructs an external explorer URL for a ledger only if NEXT_PUBLIC_EXPLORER_BASE_URL is configured.
 */
export function getExplorerLedgerUrl(ledger: number | bigint): string | null {
  if (!config.explorerBaseUrl) return null;
  const base = config.explorerBaseUrl.replace(/\/$/, '');
  return `${base}/ledger/${ledger}`;
}

/**
 * Constructs an external explorer URL for a transaction only if NEXT_PUBLIC_EXPLORER_BASE_URL is configured.
 */
export function getExplorerTxUrl(txHash: string): string | null {
  if (!config.explorerBaseUrl || !txHash) return null;
  const base = config.explorerBaseUrl.replace(/\/$/, '');
  return `${base}/tx/${txHash}`;
}
