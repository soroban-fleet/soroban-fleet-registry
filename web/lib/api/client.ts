import { config } from '../config';
import type { ApiErrorPayload } from '../types';

export class ApiError extends Error {
  public readonly status: number;
  public readonly code: string;

  constructor(status: number, message: string, code: string = 'UNKNOWN_ERROR') {
    super(message);
    this.name = 'ApiError';
    this.status = status;
    this.code = code;
  }

  static isNotFound(err: unknown): boolean {
    return err instanceof ApiError && (err.status === 404 || err.code === 'FLEET_NOT_FOUND' || err.code === 'NOT_FOUND');
  }
}

export interface RequestOptions extends RequestInit {
  params?: Record<string, string | number | boolean | undefined>;
}

export async function apiClient<T>(
  endpoint: string,
  options: RequestOptions = {}
): Promise<T> {
  const { params, headers, ...rest } = options;

  let url = `${config.apiBaseUrl.replace(/\/$/, '')}${endpoint.startsWith('/') ? '' : '/'}${endpoint}`;

  if (params) {
    const searchParams = new URLSearchParams();
    for (const [key, value] of Object.entries(params)) {
      if (value !== undefined) {
        searchParams.append(key, String(value));
      }
    }
    const query = searchParams.toString();
    if (query) {
      url += `?${query}`;
    }
  }

  let response: Response;
  try {
    response = await fetch(url, {
      ...rest,
      headers: {
        Accept: 'application/json',
        ...headers,
      },
      // Ensure dynamic freshness without stale cache for verification/fleet data
      cache: 'no-store',
    });
  } catch (networkErr: unknown) {
    const message = networkErr instanceof Error ? networkErr.message : 'Network failure';
    throw new ApiError(0, `The Fleet Registry API could not be reached: ${message}`, 'NETWORK_ERROR');
  }

  if (!response.ok) {
    let code = 'HTTP_' + response.status;
    let message = `API request failed with status ${response.status}`;

    try {
      const errorJson = (await response.json()) as ApiErrorPayload;
      if (errorJson && errorJson.error) {
        code = errorJson.error.code || code;
        message = errorJson.error.message || message;
      }
    } catch {
      // Fallback to HTTP status message if body is not JSON
      if (response.status === 404) {
        message = 'The requested resource was not found.';
      } else if (response.status === 400) {
        message = 'Invalid request parameters.';
      } else if (response.status === 429) {
        message = 'Rate limit exceeded. Please try again later.';
      } else if (response.status >= 500) {
        message = 'The registry backend encountered an error. Please try again later.';
      }
    }

    throw new ApiError(response.status, message, code);
  }

  return response.json() as Promise<T>;
}
