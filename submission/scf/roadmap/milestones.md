# Roadmap Milestones & Verifiable Deliverables

Each milestone defines an engineering scope, measurable acceptance criteria, and explicit evidence for reviewer verification.

---

## Tranche 1 Milestones

### Milestone 1.1: Security & Rate-Limiting Hardening
- **Scope**: Implement configurable token-bucket IP rate-limiting middleware (Issue #11) and input sanitization to mitigate DoS threats identified in the STRIDE matrix.
- **Acceptance Criteria**:
  - [ ] Unit and benchmark tests verifying request throttling at configured limits.
  - [ ] 429 Too Many Requests response with standard `Retry-After` headers.
  - [ ] Zero measurable latency regression on legitimate unthrottled requests.
- **Verifiable Evidence**: Merged PR with unit tests and benchmark report.

### Milestone 1.2: Production Telemetry & Prometheus Metrics
- **Scope**: Build Prometheus `/metrics` exporter (Issue #12) capturing `INDEXER_LAG_LEDGERS`, `RPC_ERROR_RATE`, `API_5XX_RATE`, and verification counts.
- **Acceptance Criteria**:
  - [ ] Prometheus OpenMetrics output available on authenticated metrics port.
  - [ ] Grafana dashboard template committed to `deploy/dashboards/`.
- **Verifiable Evidence**: OpenMetrics endpoint functional in Docker Compose topology.

### Milestone 1.3: Ingestion Performance & Replay Tooling
- **Scope**: Implement WASM TTL caching (Issue #5), optimized ledger change filtering (Issue #2), and bounded replay CLI subcommand `sfr ingest replay` (Issue #4).
- **Acceptance Criteria**:
  - [ ] Benchmark showing > 2x reduction in redundant RPC queries for immutable WASM hashes.
  - [ ] CLI command `sfr ingest replay --from <LEDGER> --to <LEDGER>` successfully verifies bounded history.
- **Verifiable Evidence**: Benchmark results and CLI test suite in `cmd/sfr`.

---

## Tranche 2 Milestones

### Milestone 2.1: Pilot Protocol Onboarding & Validation Report
- **Scope**: Conduct structured onboarding with at least two Stellar protocol teams managing multi-instance contracts on Testnet.
- **Acceptance Criteria**:
  - [ ] Documented integration of at least 2 distinct protocol fleets into the testnet registry.
  - [ ] Completed validation log entries and published pilot case studies.
- **Verifiable Evidence**: Published validation report in `docs/reports/pilot-validation.md`.

### Milestone 2.2: Live Scenario Recording & Protocol Migration Test Harness
- **Scope**: Build scenario recording CLI (Issue #14) and automated migration test harness (Issue #15) to validate upstream protocol behavior.
- **Acceptance Criteria**:
  - [ ] Tooling to serialize live testnet ledger sequences into regression fixtures.
  - [ ] Automated regression tests simulating network upgrades.
- **Verifiable Evidence**: Test fixtures committed under `fixtures/` and executed in CI.

### Milestone 2.3: Web Drift Drill-Down & Export Tooling
- **Scope**: Implement UI diff modal for drifted members (Issue #8) and CLI export command `sfr fleet export` (Issue #9) supporting CSV and JSON.
- **Acceptance Criteria**:
  - [ ] Visual diff component rendering expected vs actual WASM hash.
  - [ ] Automated Vitest component tests passing.
- **Verifiable Evidence**: Passing frontend test suite and interactive component in web UI.

---

## Tranche 3 Milestones

### Milestone 3.1: Public Mainnet Service Deployment
- **Scope**: Deploy dedicated, highly available indexing service and public read-only API on Stellar Mainnet with TLS, automated database backups, and reverse proxy caching.
- **Acceptance Criteria**:
  - [ ] Stable public HTTPS endpoint with 99.5% uptime SLA.
  - [ ] Continuous indexer keeping pace with Stellar Mainnet (`lag < 5 ledgers`).
- **Verifiable Evidence**: Live, monitorable public endpoint and uptime status page.

### Milestone 3.2: Operational Alerting & Compliance Reporting
- **Scope**: Publish comprehensive operational runbooks (Issue #13) and implement markdown compliance audit report generation (Issue #10).
- **Acceptance Criteria**:
  - [ ] Step-by-step incident runbooks covering RPC failover and database recovery.
  - [ ] CLI command `sfr fleet verify --report-format markdown` generating exportable audit logs.
- **Verifiable Evidence**: Documentation pages published on Docusaurus site and CLI test verifying markdown report generation.
