# Operations: Recovery

Standard recovery procedures for each failure mode. Principles:
[data and checkpoint move together](../architecture/indexing-pipeline.md), and replay is
idempotent — so recovery is restart, not repair.

## Backend (API) restart

**Symptom**: connection refused; `/health` failing.

```bash
sfr api            # or: docker compose restart sfr
curl -s http://localhost:8080/health
```

- No data migration, no replay.
- In-flight requests drop; the process shuts down gracefully on SIGINT/SIGTERM (5s window).
- Previous verification history is untouched.

## Indexer restart

**Symptom**: process exited; checkpoint stale.

```bash
sfr ingest
```

- Resumes at `checkpoint + 1`. `SFR_INDEXER_START_LEDGER` is ignored when a checkpoint exists.
- Ledgers after the last commit are reprocessed; upserts make this safe.
- Until caught up, verification returns `INCOMPLETE`. Re-run verification after catch-up.

Verify progress:

```sql
SELECT ledger, updated_at FROM indexer_checkpoints WHERE stream = 'main';
```

## PostgreSQL restart

**Symptom**: API returns `500 INTERNAL_ERROR`; indexer exits with a database error.

```bash
docker compose up -d postgres     # or your orchestrator's restart
docker compose ps                 # wait for healthy
```

1. Confirm the database is accepting connections (`pg_isready`).
2. Restart the API (stateless).
3. Restart the indexer — it resumes at the last committed checkpoint; the interrupted ledger
   rolled back.
4. No data repair is needed: a failed ledger transaction commits nothing.

## Stellar RPC outage

**Symptoms**:

```text
fetch ledgers error, will retry        (indexer logs)
verification → INCOMPLETE / errors     (API)
```

What the system does by itself:

- indexer retries every second; **checkpoints do not advance** while fetches fail;
- verification counts unresolved members as missing → `INCOMPLETE`;
- no indexed data is corrupted.

What you do:

1. Confirm the endpoint: `curl -s "$SFR_RPC_URL"` (or your provider's status page).
2. Point `SFR_RPC_URL` at a healthy endpoint if needed; restart the indexer.
3. After RPC recovery and catch-up, re-run verification.

## Ledger replay (planned reindex)

```bash
# stop indexer
psql "$SFR_DATABASE_URL" -c \
  "UPDATE indexer_checkpoints SET ledger = 0 WHERE stream = 'main';"
# optional: clear indexed state if you want a clean rebuild
# psql "$SFR_DATABASE_URL" -c "TRUNCATE fleet_members, fleet_releases, fleet_verifications, fleets;"

sfr ingest
```

Replay from an earlier checkpoint with existing data is also safe — releases are deduplicated by
`(owner, tag, ledger, tx_hash)` and members/fleets upsert to the same values.

## Migration rollback

```bash
sfr migrate down     # roll back the most recent migration
sfr migrate up       # re-apply
```

CI exercises this cycle on every run. If `up` fails, fix the migration before restarting
services; do not hand-edit production schema.

## Verification stuck at INCOMPLETE

Not a failure of the verifier. In order of likelihood:

1. Indexer behind — check checkpoint vs. network ledger, let it catch up.
2. RPC unreachable during member resolution — fix RPC, re-verify.
3. Fleet has zero indexed members — the fleet may be new; wait for ingestion.

## Recovery decision table

| Symptom | Action | Data repair needed? |
|---|---|---|
| API 500 / down | restart API; check DB | No |
| Indexer exited | restart `sfr ingest` | No |
| PostgreSQL down | start DB, restart API + indexer | No |
| RPC outage | wait/switch endpoint, restart indexer | No |
| Verification `INCOMPLETE` | catch up, re-verify | No |
| Verification `BROKEN_REFERENCE` | fix owner/tag on-chain (outside this system) | No |
| Verification `DRIFT` | inspect members, resolve divergence | No |
| Migration failure | `migrate down`, fix, `migrate up` | Schema only |

See [Troubleshooting](./troubleshooting.md) for inspection commands.
