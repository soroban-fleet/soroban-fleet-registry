# Indexing pipeline

The pipeline turns a stream of ledgers into consistent fleet state. Its design priorities are
ordered processing, atomicity, and idempotent replay.

Code: `internal/ingest/pipeline.go`, `internal/ingest/ledger_processor.go`,
`internal/ingest/change_processor.go`.

## One ledger at a time

```text
read checkpoint (stream "main")
        ↓
fetch batch: getLedgers(start, batchSize)
        ↓
for each ledger, in ascending order:
        ↓
decode LedgerCloseMeta → extract changes
        ↓
BEGIN TRANSACTION
   ├─ apply tag changes (fleet upsert + release row)
   ├─ apply contract changes (member upsert / deactivate)
   ├─ advance checkpoint to this ledger
COMMIT
        ↓
start = ledger + 1
```

Sequences are processed strictly in order. A gap is never skipped.

## CAP-85 detection

Two detectors run on each ledger:

1. **Owner tag change** — an owner contract's persistent entry keyed by `SCV_EXECUTABLE_TAG`
   changed. The processor upserts the fleet with the new `current_wasm_hash` and records a release
   when the hash actually differs from the stored one.
2. **Contract instance change** — the instance's executable is now an external reference. The
   processor upserts the fleet, registers the member as active, and preserves `first_seen_ledger`.
   If the old executable was an external reference and the new one is not, the member is marked
   inactive rather than deleted.

Malformed references are counted in `errors` for that ledger and skipped; they do not abort the
ledger.

## Resolution during ingestion

If the fleet's current WASM is unknown (first sighting, and no tag change in this ledger), the
processor attempts live resolution over RPC. If RPC is unavailable, the fleet row is still written;
the hash is filled in later when a tag change or verification resolves it.

## Database transaction

Everything for one ledger commits or rolls back together:

- fleet rows;
- member rows and active flags;
- release rows;
- the checkpoint row.

If any statement fails, the transaction rolls back and `Run` returns an error — the process stops
rather than continuing with half-applied state.

## Why the checkpoint advances last

The checkpoint is written **inside** the same transaction, after the data writes.

Consequences:

- The checkpoint can never point at a ledger whose data was not committed.
- A crash between data write and checkpoint write rolls both back together — no "data missing
  but marked done" state.
- On restart, ingestion resumes at `checkpoint + 1` and reprocesses the interrupted ledger.

This is the invariant behind [Failure and recovery](./failure-recovery.md).

## Idempotent replay

Reprocessing a ledger cannot create duplicates:

| Record | Protection |
|---|---|
| fleets | `ON CONFLICT (owner_address, tag) DO UPDATE` |
| members | `ON CONFLICT (contract_id) DO UPDATE` |
| releases | uniqueness check on `(owner_address, tag, ledger, tx_hash)` before insert |
| checkpoints | `ON CONFLICT (stream) DO UPDATE` |

Verified by `TestPipeline_IdempotentReplayAndResume` in the integration suite.

## Restart behavior

```text
startLedger = SFR_INDEXER_START_LEDGER        (initial run only)
startLedger = checkpoint + 1                  (any run where a checkpoint exists)
```

A configured start ledger is therefore only used for the very first run against an empty
checkpoint table.

## Failure behavior

| Failure | Behavior |
|---|---|
| RPC fetch error | warning logged, retry after `PollDelay` (1s), loop continues |
| empty ledger batch | sleep and retry |
| ledger decode error | process returns error (stops) |
| database error | transaction rolls back, process returns error (stops) |
| context cancelled (SIGINT/SIGTERM) | graceful stop, checkpoint remains at last committed ledger |

Related: [Operations → Indexing](../operations/indexing.md)
