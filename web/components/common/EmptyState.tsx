import React from 'react';

export interface EmptyStateProps {
  title?: string;
  message: string;
  actionLabel?: string;
  actionHref?: string;
}

export function EmptyState({
  title = 'No Data Available',
  message,
  actionLabel,
  actionHref,
}: EmptyStateProps) {
  return (
    <div
      className="card"
      style={{
        padding: '3rem 1.5rem',
        textAlign: 'center',
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        gap: '0.75rem',
      }}
    >
      <div style={{ fontSize: '1.125rem', fontWeight: 600, color: 'var(--text-primary)' }}>
        {title}
      </div>
      <p style={{ color: 'var(--text-secondary)', maxWidth: '450px', fontSize: '0.875rem' }}>
        {message}
      </p>
      {actionLabel && actionHref && (
        <a href={actionHref} className="btn" style={{ marginTop: '0.5rem' }}>
          {actionLabel}
        </a>
      )}
    </div>
  );
}
