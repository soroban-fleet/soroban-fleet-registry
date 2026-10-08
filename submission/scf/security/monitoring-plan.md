# Production Monitoring & Incident Response Plan

## 1. Monitoring Signals & Metrics Contract

The operational health of Soroban Fleet Registry is governed by five key operational telemetry signals:

| Metric Identifier | Signal Description | Unit | Metric Type | Severity Thresholds |
|---|---|---|---|---|
| `INDEXER_LAG_LEDGERS` | Difference between network latest ledger and committed database checkpoint (`latest_ledger - indexed_through`). | Ledgers | Gauge | **Warning**: > 20 ledgers<br/>**Critical**: > 100 ledgers |
| `RPC_ERROR_RATE` | Percentage of failed Soroban RPC queries over a rolling 5-minute window. | Percent | Gauge | **Warning**: > 5%<br/>**Critical**: > 25% |
| `API_5XX_RATE` | Percentage of HTTP 5xx responses returned by the REST API over 5 minutes. | Percent | Gauge | **Warning**: > 1%<br/>**Critical**: > 5% |
| `DB_CONNECTION_FAILURES` | Unhandled connection pool timeouts or connection drop events. | Count / min | Counter | **Warning**: > 1 / min<br/>**Critical**: > 5 / min |
| `CHECKPOINT_STALL_SECONDS` | Time elapsed since the indexer last successfully advanced its ledger checkpoint. | Seconds | Gauge | **Warning**: > 60s<br/>**Critical**: > 300s |

*Note on Thresholds: Initial proposed thresholds will be empirically calibrated against live Stellar Testnet and Mainnet ledger ingestion baselines.*

---

## 2. Telemetry Architecture

1. **Health Check Endpoints** (`internal/api`):
   - `GET /healthz`: Basic liveness probe confirming HTTP process is alive.
   - `GET /v1/health`: Detailed readiness probe reporting database ping and current committed checkpoint height.
2. **Prometheus Metrics Exporter** (Planned in Tranche 1 / Issue #12):
   - Exposes `/metrics` conforming to OpenMetrics format for consumption by Prometheus and Grafana.
3. **Structured JSON Logs**:
   - Ingestion and API logs output structured JSON with fields `level`, `timestamp`, `ledger`, `duration_ms`, and `error`.

---

## 3. Incident Response Playbooks

### Playbook A: Indexer Lagging or Checkpoint Stalled (`INDEXER_LAG_LEDGERS > 100`)
- **Trigger**: Checkpoint lag exceeds 100 ledgers or `CHECKPOINT_STALL_SECONDS > 300s`.
- **Diagnosis**:
  1. Inspect indexer logs for upstream RPC timeouts (`connection refused`, `rate limit exceeded`, or HTTP 504).
  2. Verify network ledger height directly: `curl -s -X POST $SFR_RPC_URL -d '{"jsonrpc":"2.0","id":1,"method":"getLatestLedger"}'`.
  3. Check database locks or uncommitted transactions in PostgreSQL.
- **Remediation**:
  - If RPC node is unresponsive, fail over to backup RPC endpoint (`SFR_RPC_URL`).
  - Restart ingestion daemon (`sfr ingest`). Pipeline will automatically resume from the last committed checkpoint idempotently.

### Playbook B: Elevated API Error Rate (`API_5XX_RATE > 5%`)
- **Trigger**: HTTP 5xx responses exceed 5% over 5 minutes.
- **Diagnosis**:
  1. Check PostgreSQL connection pool exhaustion in API logs (`too many connections` or `connection pool timeout`).
  2. Confirm database CPU and memory metrics.
- **Remediation**:
  - Recycle API process or increase connection pool limits (`max_open_conns`).
  - Verify PostgreSQL health and connectivity via `pg_isready`.
