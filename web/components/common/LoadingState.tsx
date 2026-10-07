import React from 'react';

export interface LoadingStateProps {
  label?: string;
  rows?: number;
}

export function LoadingState({ label = 'Loading registry data...', rows = 3 }: LoadingStateProps) {
  return (
    <div
      style={{
        padding: '2.5rem 1.5rem',
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
        gap: '1rem',
        borderRadius: '6px',
        border: '1px solid var(--border-color)',
        backgroundColor: 'var(--bg-secondary)',
      }}
      role="status"
      aria-live="polite"
    >
      <div
        style={{
          width: '28px',
          height: '28px',
          border: '2px solid var(--border-color)',
          borderTopColor: 'var(--accent)',
          borderRadius: '50%',
          animation: 'spin 0.8s linear infinite',
        }}
      />
      <div style={{ color: 'var(--text-secondary)', fontSize: '0.875rem' }}>{label}</div>

      <style>{`
        @keyframes spin {
          to { transform: rotate(360deg); }
        }
      `}</style>
    </div>
  );
}

export function SkeletonRow() {
  return (
    <div
      style={{
        height: '2rem',
        backgroundColor: 'var(--bg-tertiary)',
        borderRadius: '4px',
        animation: 'pulse 1.5s ease-in-out infinite',
        marginBottom: '0.5rem',
      }}
    >
      <style>{`
        @keyframes pulse {
          0%, 100% { opacity: 0.6; }
          50% { opacity: 0.3; }
        }
      `}</style>
    </div>
  );
}
