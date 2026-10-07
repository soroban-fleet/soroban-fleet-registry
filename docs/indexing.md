# Indexing Model

The indexer processes Stellar ledgers sequentially and idempotently to construct an authoritative operational view of fleets, memberships, and upgrades.

## Ingestion Cycle

For each ledger:

```text
Load stream checkpoint
          |
          v
Fetch ledger metadata via RPC (getLedgers)
          |
          v
Extract changes from LedgerCloseMeta
          |
          +---> Detect contract instance modifications
          |
          +---> Detect owner tag executable changes
          |
Begin Database Transaction
          |
          +---> Upsert fleet records
          +---> Upsert member lifecycle states
          +---> Record release rows (if WASM changed)
          +---> Advance indexer_checkpoints to current ledger
          |
COMMIT Transaction
```

## Atomic Transaction Guarantee

A ledger's state writes and checkpoint advancement are atomic:
- If any write fails, the entire transaction rolls back.
- The checkpoint never advances ahead of committed ledger writes.
- Upon process restart, ingestion resumes from `last_checkpoint + 1`.

## Idempotency Rules

Reprocessing an already ingested ledger will never create:
- Duplicate fleet records (handled via `ON CONFLICT (owner_address, tag) DO UPDATE`).
- Duplicate members (handled via `ON CONFLICT (contract_id) DO UPDATE`).
- Duplicate releases (checked via `(owner_address, tag, ledger, tx_hash)` uniqueness before insert).
- Duplicate checkpoints.

## Member Lifecycle Transitions

Members are never physically deleted when they transition away from a fleet:
- **Joined**: When a contract instance is created or updated to point to `(owner, tag)`, it is marked `active = TRUE` with `first_seen_ledger` preserved.
- **Upgraded within fleet**: When the owner updates the tag's WASM hash, active members inherit the new hash without changing membership status.
- **Left fleet**: When a contract instance updates its executable to a different fleet, a direct WASM, or is destroyed, it is marked `active = FALSE` in the old fleet with `last_seen_ledger` recorded.
