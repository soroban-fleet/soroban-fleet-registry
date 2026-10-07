import type { Metadata } from 'next';
import './globals.css';

export const metadata: Metadata = {
  title: 'Soroban Fleet Registry',
  description: 'Discover and verify Soroban CAP-85 externally managed contract executable fleets.',
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <body>
        <header style={{ borderBottom: '1px solid var(--border-color)', backgroundColor: 'var(--bg-secondary)' }}>
          <div className="container" style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', height: '64px' }}>
            <div style={{ display: 'flex', alignItems: 'center', gap: '1.5rem' }}>
              <a href="/" style={{ fontSize: '1.125rem', fontWeight: 700, color: 'var(--text-primary)', textDecoration: 'none', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                <span style={{ color: 'var(--accent)' }}>SFR</span>
                <span>Fleet Registry</span>
              </a>
              <nav style={{ display: 'flex', gap: '1rem', marginLeft: '1rem' }} aria-label="Main Navigation">
                <a href="/" style={{ color: 'var(--text-secondary)', fontSize: '0.875rem' }}>Overview</a>
                <a href="/fleets" style={{ color: 'var(--text-secondary)', fontSize: '0.875rem' }}>Fleets</a>
                <a
                  href="https://github.com/soroban-fleet/soroban-fleet-registry/tree/main/docs"
                  target="_blank"
                  rel="noopener noreferrer"
                  style={{ color: 'var(--text-secondary)', fontSize: '0.875rem' }}
                >
                  Documentation
                </a>
              </nav>
            </div>
            <div style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }} className="mono">
              CAP-85 · Protocol 28
            </div>
          </div>
        </header>

        <main style={{ flex: 1, padding: '2rem 0' }}>
          {children}
        </main>

        <footer style={{ borderTop: '1px solid var(--border-color)', backgroundColor: 'var(--bg-secondary)', padding: '1.5rem 0', marginTop: 'auto' }}>
          <div className="container" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', fontSize: '0.75rem', color: 'var(--text-muted)' }}>
            <div>Soroban Fleet Registry · Operational CAP-85 Visibility & Verification</div>
            <div style={{ display: 'flex', gap: '1rem' }}>
              <a href="https://github.com/soroban-fleet/soroban-fleet-registry" target="_blank" rel="noopener noreferrer" style={{ color: 'var(--text-secondary)' }}>
                GitHub
              </a>
            </div>
          </div>
        </footer>
      </body>
    </html>
  );
}
