import type { components, paths } from './schema';

export type { components, paths };

// Domain model types derived strictly from api/openapi.yaml
export type Fleet = components['schemas']['Fleet'];
export type FleetID = components['schemas']['FleetID'];
export type FleetMember = components['schemas']['FleetMember'];
export type FleetRelease = components['schemas']['FleetRelease'];
export type VerificationStatus = components['schemas']['VerificationStatus'];
export type VerificationResult = components['schemas']['VerificationResult'];
export type ContractInspection = components['schemas']['ContractInspection'];
export type PaginationMeta = components['schemas']['PaginationMeta'];

export interface PaginationParams {
  limit?: number;
  offset?: number;
}

export interface ListMembersParams extends PaginationParams {
  active_only?: boolean;
}

export interface VerifyFleetParams {
  expected_wasm?: string;
}

export interface FleetListResponse {
  data: Fleet[];
  meta: PaginationMeta;
}

export interface FleetMembersResponse {
  data: FleetMember[];
  meta: PaginationMeta;
}

export interface FleetHistoryResponse {
  data: VerificationResult[];
  meta: PaginationMeta;
}

export interface FleetReleasesResponse {
  data: FleetRelease[];
  meta: PaginationMeta;
}

export interface FleetResponse {
  data: Fleet;
}

export interface VerificationResponse {
  data: VerificationResult;
}

export interface ContractResponse {
  data: ContractInspection;
}

export interface ApiErrorPayload {
  error: {
    code: string;
    message: string;
  };
}
