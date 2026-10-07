import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { apiClient, ApiError } from '../lib/api/client';
import {
  listFleets,
  getFleet,
  getFleetMembers,
  getFleetHistory,
  getFleetReleases,
  verifyFleet,
} from '../lib/api/fleets';
import { getContract } from '../lib/api/contracts';

describe('API Client', () => {
  const originalFetch = global.fetch;

  beforeEach(() => {
    vi.restoreAllMocks();
  });

  afterEach(() => {
    global.fetch = originalFetch;
  });

  it('serializes queries and returns successful data', async () => {
    const mockData = { data: [], meta: { total: 0, limit: 20, offset: 0 } };
    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => mockData,
    } as unknown as Response);

    const res = await listFleets({ limit: 10, offset: 20 });
    expect(res).toEqual(mockData);
    expect(global.fetch).toHaveBeenCalledWith(
      'http://localhost:8080/v1/fleets?limit=10&offset=20',
      expect.objectContaining({ cache: 'no-store' })
    );
  });

  it('encodes path parameters safely without altering tags', async () => {
    const mockData = { data: { id: { Owner: 'C123', Tag: 'vault/v1 special' } } };
    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => mockData,
    } as unknown as Response);

    await getFleet('C123', 'vault/v1 special');
    expect(global.fetch).toHaveBeenCalledWith(
      'http://localhost:8080/v1/fleets/C123/vault%2Fv1%20special',
      expect.anything()
    );
  });

  it('throws ApiError with structured error payload on 404', async () => {
    global.fetch = vi.fn().mockResolvedValue({
      ok: false,
      status: 404,
      json: async () => ({
        error: { code: 'FLEET_NOT_FOUND', message: 'Fleet does not exist' },
      }),
    } as unknown as Response);

    await expect(getFleet('C123', 'tag')).rejects.toThrow('Fleet does not exist');
    await expect(getFleet('C123', 'tag')).rejects.toMatchObject({
      status: 404,
      code: 'FLEET_NOT_FOUND',
    });
  });

  it('distinguishes network errors', async () => {
    global.fetch = vi.fn().mockRejectedValue(new Error('Connection refused'));

    await expect(listFleets()).rejects.toThrow('The Fleet Registry API could not be reached');
    await expect(listFleets()).rejects.toMatchObject({
      status: 0,
      code: 'NETWORK_ERROR',
    });
  });

  it('validates empty contractId or fleet parameters', async () => {
    await expect(getContract('')).rejects.toThrow('Contract ID is required');
    await expect(getFleet('', 'tag')).rejects.toThrow('Owner address is required');
    await expect(getFleet('C123', '')).rejects.toThrow('Fleet tag is required');
  });

  it('fetches members, releases, history, verify, and contracts', async () => {
    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: async () => ({ data: {} }),
    } as unknown as Response);

    await getFleetMembers('C123', 'tag', { active_only: true, limit: 10 });
    expect(global.fetch).toHaveBeenCalledWith(
      'http://localhost:8080/v1/fleets/C123/tag/members?active_only=true&limit=10',
      expect.anything()
    );

    await getFleetHistory('C123', 'tag');
    expect(global.fetch).toHaveBeenCalledWith(
      'http://localhost:8080/v1/fleets/C123/tag/history',
      expect.anything()
    );

    await getFleetReleases('C123', 'tag');
    expect(global.fetch).toHaveBeenCalledWith(
      'http://localhost:8080/v1/fleets/C123/tag/releases',
      expect.anything()
    );

    await verifyFleet('C123', 'tag', { expected_wasm: 'abcdef' });
    expect(global.fetch).toHaveBeenCalledWith(
      'http://localhost:8080/v1/fleets/C123/tag/verify?expected_wasm=abcdef',
      expect.anything()
    );

    await getContract('C123');
    expect(global.fetch).toHaveBeenCalledWith(
      'http://localhost:8080/v1/contracts/C123',
      expect.anything()
    );
  });
});
