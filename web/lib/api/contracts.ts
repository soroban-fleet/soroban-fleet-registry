import { apiClient, ApiError } from './client';
import type { ContractResponse } from '../types';

export async function getContract(contractId: string): Promise<ContractResponse> {
  if (!contractId || typeof contractId !== 'string') {
    throw new ApiError(400, 'Contract ID is required and must be a valid string', 'INVALID_PARAM');
  }

  const encodedId = encodeURIComponent(contractId);
  return apiClient<ContractResponse>(`/v1/contracts/${encodedId}`);
}
