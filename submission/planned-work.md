# Planned Backlog and Future Work

The repository maintains an active, scoped issue backlog representing concrete engineering tasks for future releases:

### Indexer Pipeline
- **#2**: `perf(indexer): optimize ledger change filtering for CAP-85 instances` — Streamline predicate checks to skip non-contract ledger entries early during high transaction throughput.
- **#16**: `feat(indexer): support historical backfill from Stellar ledger history archives` — Accelerate genesis-to-current catch-up by ingesting parallel history archive buckets rather than relying strictly on live JSON-RPC ledger batches.

### Resolution & Caching
- **#5**: `feat(cap85): implement TTL cache for immutable WASM hashes in resolver` — Reduce redundant RPC queries by caching verified immutable bytecode entries.
- **#15**: `test(cap85): add automated protocol migration test harness` — Automated test harness simulating upstream protocol transitions and ledger format changes.

### Verification Engine
- **#10**: `feat(verification): generate markdown compliance audit report` — Export verifiable audit summaries suitable for security reviews and compliance documentation.

### REST API
- **#3**: `feat(api): expose indexer stream health and ledger lag endpoint` — Expose precise sequence lag and stream lag metrics for upstream orchestrators.
- **#6**: `feat(api): add status and tag filtering to fleet listings` — Support query parameter filtering by verification status (`HEALTHY`, `DRIFT`, etc.) and tag patterns.
- **#11**: `feat(api): add configurable IP rate limiting middleware` — Protect public API nodes against abusive request bursts.

### CLI Usability & Automation
- **#4**: `feat(cli): add sfr ingest replay subcommand for bounded ledger ranges` — Allow operators to replay and test specific historical ledger ranges locally.
- **#7**: `docs(cli): publish and validate JSON output schemas for sfr commands` — Formal JSON Schema definitions for CI pipeline integration.
- **#9**: `feat(cli): add sfr fleet export command supporting JSON and CSV` — Export full member rosters and verification audits into CSV/JSON formats.

### Web Interface
- **#8**: `feat(web): add member drift drill-down modal on verification view` — Visual side-by-side diffing showing expected vs actual member WASM hashes on drift events.

### Operations & Monitoring
- **#12**: `feat(ops): add Prometheus metrics exporter for indexer and verification` — Native OpenMetrics endpoint for Grafana dashboards and incident monitoring.
- **#13**: `docs(ops): create operational alerting and runbook guide` — Step-by-step incident response playbooks for operators.
- **#14**: `chore(fixtures): add CLI tool to record and serialize live testnet scenarios` — Automated tooling to capture and bundle reproducible on-chain testnet scenarios as unit fixtures.
