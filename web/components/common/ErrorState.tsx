import React from 'react';
import { ApiError } from '../../lib/api/client';

export interface ErrorStateProps {
  error: unknown;
  onRetry?: () => void;
  title?: string;
}

export function ErrorState({ error, onRetry, title }: ErrorStateProps) {
  let displayTitle = title || 'Error Loading Data';
  let message = 'An unexpected error occurred.';
  let code = 'UNKNOWN_ERROR';
  let isNotFound = false;

  if (error instanceof ApiError) {
    code = error.code;
    message = error.message;

    if (error.status === 404) {
      displayTitle = title || 'Not Found';
      isNotFound = true;
    } else if (error.status === 400) {
      displayTitle = title || 'Invalid Request';
    } else if (error.status === 429) {
      displayTitle = title || 'Rate Limit Exceeded';
    } else if (error.status >= 500) {
      displayTitle = title || 'Backend Service Error';
    } else if (error.status === 0 || error.code === 'NETWORK_ERROR') {
      displayTitle = title || 'Network Error';
    }
  } else if (error instanceof Error) {
    message = error.message;
  }

  return (
    <div
      className="card"
      style={{
        borderColor: isNotFound ? 'var(--border-color)' : 'var(--status-broken-border)',
        padding: '2rem',
        textAlign: 'center',
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        gap: '0.75rem',
      }}
      role="alert"
    >
      <div
        style={{
          fontSize: '1.25rem',
          fontWeight: 700,
          color: isNotFound ? 'var(--text-primary)' : 'var(--status-broken-text)',
        }}
      >
        {displayTitle}
      </div>

      <p style={{ color: 'var(--text-secondary)', maxWidth: '500px', fontSize: '0.875rem' }}>
        {message}
      </p>

      <div
        style={{
          fontSize: '0.75rem',
          color: 'var(--text-muted)',
          padding: '0.25rem 0.5rem',
          backgroundColor: 'var(--bg-tertiary)',
          borderRadius: '4px',
        }}
        className="mono"
      >
        Code: {code}
      </div>

      <div style={{ display: 'flex', gap: '0.75rem', marginTop: '0.5rem' }}>
        {onRetry && (
          <button type="button" onClick={onRetry} className="btn btn-primary">
            Retry
          </button>
        )}
        <a href="/fleets" className="btn">
          View All Fleets
        </a>
      </div>
    </div>
  );
}
