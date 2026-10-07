import React from 'react';

export default function NotFound() {
  return (
    <div className="container" style={{ textAlign: 'center', padding: '4rem 1rem' }}>
      <h1 style={{ fontSize: '2rem', marginBottom: '1rem', fontWeight: 700 }}>404 · Not Found</h1>
      <p style={{ color: 'var(--text-secondary)', marginBottom: '1.5rem' }}>
        The requested fleet, contract, or page could not be found.
      </p>
      <a href="/" className="btn btn-primary">
        Return Home
      </a>
    </div>
  );
}
