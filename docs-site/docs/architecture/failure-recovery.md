# Failure and recovery

What each failure looks like, what the system does automatically, and what an operator has to do.

## Principles

1. **No false conclusions.** When information is missing, the verdict is `INCOMPLETE` or
   `UNKNOWN`, never `HEALTHY`.
2. **Checkpoint never runs ahead of data.** The checkpoint advances only inside the transaction
   that commits the ledger's data.
3. **Replay is safe.** Reprocessing a ledger cannot duplicate records.
4. **API stays up.** The API serves whatever is indexed, even while the indexer is behind.

## Database restart / outage

| | |
|---|---|
| **Expected behavior** | In-flight ledger transactions roll back. The indexer's next commit fails; the process returns an error and exits. API handlers return `500` with `{"error":{"code":"INTERNAL_ERROR","message":"..."}}`. |
| **Data impact** | None. Nothing was committed for the interrupted ledger. |
| **Recovery** | Start PostgreSQL (`docker compose up -d postgres` or your orchestrator), then restart `sfr ingest` and `sfr api`. Ingestion resumes at `checkpoint + 1`. |
| **Verification while down** | Verification requires the database (member list, checkpoint, result persistence) and will fail with an error. |

Database credentials and stack traces are never included in API responses.

## Backend (API) restart

| | |
|---|---|
| **Expected behavior** | In-flight HTTP requests drop; the process exits on SIGINT/SIGTERM with a 5-second graceful shutdown window. |
| **Data impact** | Minimal. The API never writes fleet or member state; its only write is the verification audit row inserted when `…/verify` is called. |
| **Recovery** | Restart the process. Health returns `200` as soon as the server is listening. No migration or replay required. |

## Indexer restart

| | |
|---|---|
| **Expected behavior** | On startup the pipeline reads the checkpoint and continues at `checkpoint + 1`. Ledgers after the last committed one are reprocessed. |
| **Data impact** | None. Idempotency rules prevent duplicate fleets, members, releases, or checkpoints. |
| **Recovery** | Just restart it. Optionally set `SFR_INDEXER_START_LEDGER` — it is ignored when a checkpoint already exists. |
| **Until caught up** | Verification returns `INCOMPLETE` because `indexed_through < latest_ledger`. |

## RPC outage

```text
RPC unavailable
→ checkpoint must not advance
→ verification can become INCOMPLETE
```

- **Indexer**: `getLedgers` failures are logged as warnings and retried every second
  (`PollDelay`). The checkpoint does not move, because no ledger was processed.
- **Verification**: members that cannot be resolved count as `missing` → `INCOMPLETE`. A
  resolution error classified as `ErrBrokenReference` → `BROKEN_REFERENCE`. The latest network
  ledger is unavailable, so coverage comparison uses the checkpoint only.
- **Contract inspection**: returns an error rather than stale guesses.
- **Recovery**: when RPC returns, the indexer catches up from the stored checkpoint; re-run
  verification afterwards.

## Ledger replay

Replaying from an earlier point is safe and requires no cleanup:

1. Optionally truncate or keep existing rows — upserts converge to the same state.
2. Set the stream checkpoint (or start with an empty checkpoint table) at the desired ledger.
3. Run `sfr ingest`.

Because every record type is upsert- or uniqueness-protected, replay converges instead of
duplicating. `TestPipeline_IdempotentReplayAndResume` covers crash-and-resume behavior.

## Migration failure

```bash
sfr migrate up     # apply
sfr migrate down   # roll back the last migration
```

CI runs `migrate down` followed by `migrate up` against PostgreSQL 16 as a smoke check. If `up`
fails, inspect the error output, fix the migration, and re-run; the migrator applies migrations
transactionally.

## Decision table

| Symptom | First check | Fix |
|---|---|---|
| API returns 500 | database reachability | restart/repair PostgreSQL, restart API |
| `sfr ingest` exits immediately | `SFR_DATABASE_URL`, `SFR_RPC_URL` set? | export required variables |
| verification stuck `INCOMPLETE` | `indexed_through` vs `latest_ledger` | let indexer catch up, re-verify |
| verification `BROKEN_REFERENCE` | owner entry exists on-chain? | fix reference at source |
| members missing after restart | checkpoint value | reprocess from `checkpoint + 1` |

More: [Operations → Recovery](../operations/recovery.md) ·
[Troubleshooting](../operations/troubleshooting.md)
