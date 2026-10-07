import { apiClient, ApiError } from './client';
import type {
  FleetListResponse,
  FleetResponse,
  FleetMembersResponse,
  FleetHistoryResponse,
  FleetReleasesResponse,
  VerificationResponse,
  PaginationParams,
  ListMembersParams,
  VerifyFleetParams,
} from '../types';

function validateFleetParams(owner: string, tag: string): void {
  if (!owner || typeof owner !== 'string') {
    throw new ApiError(400, 'Owner address is required and must be a valid string', 'INVALID_PARAM');
  }
  if (!tag || typeof tag !== 'string') {
    throw new ApiError(400, 'Fleet tag is required and must be a valid string', 'INVALID_PARAM');
  }
}

function fleetPath(owner: string, tag: string, subPath: string = ''): string {
  validateFleetParams(owner, tag);
  const encodedOwner = encodeURIComponent(owner);
  const encodedTag = encodeURIComponent(tag);
  return `/v1/fleets/${encodedOwner}/${encodedTag}${subPath}`;
}

export async function listFleets(params?: PaginationParams): Promise<FleetListResponse> {
  return apiClient<FleetListResponse>('/v1/fleets', {
    params: {
      limit: params?.limit,
      offset: params?.offset,
    },
  });
}

export async function getFleet(owner: string, tag: string): Promise<FleetResponse> {
  const path = fleetPath(owner, tag);
  return apiClient<FleetResponse>(path);
}

export async function getFleetMembers(
  owner: string,
  tag: string,
  params?: ListMembersParams
): Promise<FleetMembersResponse> {
  const path = fleetPath(owner, tag, '/members');
  return apiClient<FleetMembersResponse>(path, {
    params: {
      active_only: params?.active_only,
      limit: params?.limit,
      offset: params?.offset,
    },
  });
}

export async function getFleetHistory(
  owner: string,
  tag: string,
  params?: PaginationParams
): Promise<FleetHistoryResponse> {
  const path = fleetPath(owner, tag, '/history');
  return apiClient<FleetHistoryResponse>(path, {
    params: {
      limit: params?.limit,
      offset: params?.offset,
    },
  });
}

export async function getFleetReleases(
  owner: string,
  tag: string,
  params?: PaginationParams
): Promise<FleetReleasesResponse> {
  const path = fleetPath(owner, tag, '/releases');
  return apiClient<FleetReleasesResponse>(path, {
    params: {
      limit: params?.limit,
      offset: params?.offset,
    },
  });
}

export async function verifyFleet(
  owner: string,
  tag: string,
  params?: VerifyFleetParams
): Promise<VerificationResponse> {
  const path = fleetPath(owner, tag, '/verify');
  return apiClient<VerificationResponse>(path, {
    params: {
      expected_wasm: params?.expected_wasm,
    },
  });
}
