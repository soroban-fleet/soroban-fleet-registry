import type { ReactNode } from 'react';
import clsx from 'clsx';
import Link from '@docusaurus/Link';
import Layout from '@theme/Layout';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import styles from './index.module.css';

function HomepageHeader() {
  const { siteConfig } = useDocusaurusContext();
  return (
    <header className={clsx('home-hero', 'hero hero--primary')}>
      <div className="container">
        <h1>{siteConfig.title}</h1>
        <p className="tagline">{siteConfig.tagline}</p>
        <div className="home-actions">
          <Link className="button button--secondary button--lg" to="/docs/introduction/what-is-sfr">
            Read the docs
          </Link>
          <Link className="button button--outline button--lg" to="/docs/introduction/quick-start">
            Quick start
          </Link>
        </div>
      </div>
    </header>
  );
}

function Section({ title, children }: { title: string; children: ReactNode }) {
  return (
    <section>
      <h2>{title}</h2>
      {children}
    </section>
  );
}

function Home(): ReactNode {
  const { siteConfig } = useDocusaurusContext();
  return (
    <Layout
      title="Documentation"
      description="Open-source indexing and verification service for Soroban contracts using CAP-85 externally managed executables."
    >
      <HomepageHeader />
      <main className="container">
        <div className={styles.features}>
          <div className="home-card">
            <h3>For ecosystem reviewers</h3>
            <p>
              What the project is, the maintenance problem it addresses, how CAP-85 fleets work,
              what verification means, and what this service does not do.
            </p>
            <Link to="/docs/introduction/what-is-sfr">Start with the introduction →</Link>
          </div>
          <div className="home-card">
            <h3>For developers</h3>
            <p>
              Local setup, CLI commands, the REST API, configuration variables, test commands,
              and fixtures.
            </p>
            <Link to="/docs/development/prerequisites">Development guide →</Link>
          </div>
          <div className="home-card">
            <h3>For maintainers</h3>
            <p>
              Indexing pipeline behavior, checkpoint semantics, failure recovery, health checks,
              and troubleshooting.
            </p>
            <Link to="/docs/architecture/system-architecture">Architecture →</Link>
          </div>
        </div>

        <Section title="Scope">
          <div className="boundary-note">
            <p style={{ margin: 0 }}>
              Soroban Fleet Registry V1 is a read-only indexing and verification service. It does
              not deploy, own, upgrade, or execute a custom Soroban contract. It has no token,
              staking mechanism, fee schedule, or custom financial contract. Documentation on this
              site corresponds to the{' '}
              <a href="https://github.com/soroban-fleet/soroban-fleet-registry/releases/tag/v1.0.0">
                v1.0.0
              </a>{' '}
              release of{' '}
              <a href="https://github.com/soroban-fleet/soroban-fleet-registry">
                {siteConfig.projectName}
              </a>
              .
            </p>
          </div>
        </Section>

        <Section title="What the service does">
          <div className="home-grid">
            <div className="home-card">
              <h3>Index</h3>
              <p>Reads Stellar ledgers, decodes CAP-85 external references, and stores fleet membership.</p>
            </div>
            <div className="home-card">
              <h3>Resolve</h3>
              <p>Reads the owner contract entry keyed by the tag to find the active WASM hash.</p>
            </div>
            <div className="home-card">
              <h3>Verify</h3>
              <p>Compares each member's live executable against the expected hash and records the result.</p>
            </div>
            <div className="home-card">
              <h3>Serve</h3>
              <p>Exposes the indexed data over a read-only REST API, a CLI, and a web interface.</p>
            </div>
          </div>
        </Section>
      </main>
    </Layout>
  );
}

export default Home;
