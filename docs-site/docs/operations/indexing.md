# Operations: Indexing

How the indexer advances, how to watch it, and how to catch it up.

## Checkpoint

The single source of ingestion progress:

```text
indexer_checkpoints
  stream = "main"
  ledger = <last committed sequence>
  updated_at = <timestamp>
```

Inspect it:

```bash
psql "$SFR_DATABASE_URL" -c "SELECT * FROM indexer_checkpoints;"
```

Or infer progress from a fleet row's `last_indexed_ledger` (API: `GET /v1/fleets/{owner}/{tag}`).

## Ledger progress

```text
start = checkpoint + 1
while true:
    batch = getLedgers(start, SFR_INDEXER_BATCH_SIZE)
    for ledger in batch:            # ascending
        BEGIN
          apply changes
          checkpoint = ledger
        COMMIT
    start = last + 1
```

- Order is strictly increasing; gaps are never skipped.
- Each ledger commits data and checkpoint together.
- `ledger_processed` log lines report duration, changes, fleets, members, and releases per ledger.

## Catch-up

A fresh or long-stopped indexer replays from its checkpoint:

1. Ensure `SFR_RPC_URL` is reachable.
2. Start `sfr ingest`. It resumes at `checkpoint + 1` automatically.
3. Optional: set `SFR_INDEXER_START_LEDGER` for a brand-new database (ignored once a checkpoint
   exists).
4. `SFR_INDEXER_BATCH_SIZE` (default 50) trades request count against memory; raise it (50–200)
   if the RPC endpoint allows larger `getLedgers` ranges.

Catch-up speed depends on RPC throughput and how many ledgers are behind. During catch-up,
verification returns `INCOMPLETE` — expected, not an error.

## Replay

To rebuild from an earlier point:

1. Stop the indexer.
2. Point the stream checkpoint back (`UPDATE indexer_checkpoints SET ledger = <n> WHERE stream =
   'main';`) or clear the row to restart from `SFR_INDEXER_START_LEDGER`.
3. Start `sfr ingest`.

Replay is idempotent: upserts converge and releases are deduplicated by
`(owner, tag, ledger, tx_hash)`. No cleanup is required before replaying.

## Idempotency guarantees

| Record | Replay behavior |
|---|---|
| fleets | Upsert on `(owner_address, tag)` |
| members | Upsert on `contract_id` |
| releases | Inserted only if not already present for that ledger + transaction |
| checkpoints | Upsert on `stream` |

Verified by `TestPipeline_IdempotentReplayAndResume`.

## Monitoring signals

| Signal | Healthy value | Warning |
|---|---|---|
| `checkpoint.ledger` vs network latest | gap small and shrinking | gap growing |
| `updated_at` | recent | stale by minutes |
| `ledger_processed` logs | steady cadence | absent or error lines |
| process exit code | running | nonzero (`4` config, `5` runtime) |

## Failure handling

| Situation | Behavior |
|---|---|
| RPC fetch error | Warning logged, retry every second |
| Empty batch | Sleep, retry |
| Decode error | Process exits (stop, investigate the ledger) |
| Database error | Transaction rolls back, process exits |
| SIGINT/SIGTERM | Graceful stop; checkpoint stays at last committed ledger |

Related: [Health](./health.md) · [Recovery](./recovery.md) ·
[Indexing pipeline](../architecture/indexing-pipeline.md)
