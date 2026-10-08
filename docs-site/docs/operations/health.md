# Operations: Health

Three different questions, three different signals.

```text
service health   ≠   indexing freshness   ≠   fleet verification
```

## Service health

```bash
curl -s http://localhost:8080/health
curl -s http://localhost:8080/v1/health     # versioned alias, identical payload
```

```json
{
  "status": "UP",
  "service": "soroban-fleet-registry",
  "timestamp": "2026-10-07T12:00:00Z"
}
```

- HTTP `200` means the process is listening and handler panics are being recovered.
- Both routes are registered in `internal/api/server.go`.
- Use for liveness/readiness probes and load-balancer targets.

**What it does not say**: whether the indexer has caught up, whether data is fresh, or whether any
fleet is consistent.

## Indexing freshness

Not exposed as its own endpoint. Derive it from:

```sql
SELECT ledger, updated_at FROM indexer_checkpoints WHERE stream = 'main';
```

compare with the network's latest ledger (RPC `getLatestLedger`), or read
`last_indexed_ledger` from a fleet:

```bash
curl -s "http://localhost:8080/v1/fleets/<OWNER>/<TAG>"
```

| Condition | Interpretation |
|---|---|
| checkpoint ≈ network latest | index current |
| checkpoint stale / `updated_at` old | indexer stopped, RPC down, or DB errors |
| no checkpoint row | indexer never ran against this database |

## Fleet verification

```bash
curl -s "http://localhost:8080/v1/fleets/<OWNER>/<TAG>/verify"
```

The `status` field plus `indexed_through` is the authoritative consistency signal. HTTP `200`
covers `HEALTHY`, `DRIFT`, `BROKEN_REFERENCE`, and `INCOMPLETE` — read the body.

```bash
sfr fleet verify --owner <OWNER> --tag <TAG>
echo $?    # 0 healthy, 1 drift, 2 broken, 3 incomplete, 4 input, 5 internal
```

## Combined dashboard example

```text
GET /health                → UP                      (API is alive)
checkpoint                 → 105410                  (8 ledgers behind)
GET …/verify               → INCOMPLETE              (coverage insufficient)
```

Nothing here is broken. The API is up, the indexer is briefly behind, and verification correctly
refuses to conclude. See [Indexing coverage](../concepts/indexing-coverage.md).

## What to alert on

| Alert | Condition | Priority |
|---|---|---|
| API down | `/health` failing | High |
| Indexer stalled | `updated_at` older than your SLO, or process exited | High |
| Verification degraded | fleet you depend on leaves `HEALTHY` | Per fleet policy |
| Coverage gap growing | checkpoint falling further behind | Medium |

Alerting on `HEALTHY → INCOMPLETE` during deploys is noise; alert on sustained gaps.

## Other endpoints

There are no further health endpoints in v1.0.0 — no `/metrics`, no `/ready`, no `/live`. If you
need metrics, scrape process-level signals or add instrumentation in a future release.
