import { describe, it, expect } from 'vitest';
import {
  formatAddress,
  formatHash,
  normalizeWasmHash,
  getStatusDescription,
  getStatusBadgeClass,
} from '../lib/formatting';

describe('Formatting Utilities', () => {
  it('truncates StrKey address correctly', () => {
    const addr = 'CA3D5KRYMCMCZVACQY3CGINPVPUM2OAL6YHT3ZVM5C2MUMAA252AAAAB';
    const formatted = formatAddress(addr, 4);
    expect(formatted).toBe('CA3D5K...AAAB');
  });

  it('handles short addresses without modification', () => {
    expect(formatAddress('C123')).toBe('C123');
    expect(formatAddress('')).toBe('');
  });

  it('normalizes and truncates WASM hashes', () => {
    // 32 zero bytes in base64: 'AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA='
    const b64 = 'AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=';
    const hex = normalizeWasmHash(b64);
    expect(hex).toHaveLength(64);
    expect(hex).toBe('0000000000000000000000000000000000000000000000000000000000000000');

    const formatted = formatHash(hex, 4);
    expect(formatted).toBe('000000...0000');
  });

  it('returns exact protocol status descriptions', () => {
    expect(getStatusDescription('HEALTHY')).toContain('All indexed active members resolve to the expected executable');
    expect(getStatusDescription('DRIFT')).toContain('One or more active members resolve to a different executable');
    expect(getStatusDescription('BROKEN_REFERENCE')).toContain('The fleet executable reference cannot be resolved correctly');
    expect(getStatusDescription('INCOMPLETE')).toContain('The indexer has not processed enough ledger history');
    expect(getStatusDescription('UNKNOWN')).toContain('The registry does not currently have enough information');
  });

  it('maps status badges', () => {
    expect(getStatusBadgeClass('HEALTHY')).toBe('badge-healthy');
    expect(getStatusBadgeClass('DRIFT')).toBe('badge-drift');
    expect(getStatusBadgeClass('BROKEN_REFERENCE')).toBe('badge-broken');
    expect(getStatusBadgeClass('INCOMPLETE')).toBe('badge-incomplete');
    expect(getStatusBadgeClass('UNKNOWN')).toBe('badge-unknown');
  });
});
