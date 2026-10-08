# State model

Five domain objects, five tables. The schema is created by
[`migrations/000001_init_schema.up.sql`](https://github.com/soroban-fleet/soroban-fleet-registry/blob/main/migrations/000001_init_schema.up.sql);
no other domain objects exist in V1.

```text
Fleet  ──1:N──  FleetMember
   │
   ├──1:N──  FleetRelease
   ├──1:N──  Verification
   │
Checkpoint (global, per stream)
```

## Fleet

```sql
CREATE TABLE fleets (
    owner_address TEXT NOT NULL,
    tag TEXT NOT NULL,
    current_wasm_hash BYTEA,
    member_count BIGINT NOT NULL DEFAULT 0,
    first_seen_ledger BIGINT NOT NULL,
    last_seen_ledger BIGINT NOT NULL,
    last_indexed_ledger BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (owner_address, tag)
);
```

- Identity: `(owner_address, tag)`.
- `current_wasm_hash` may be `NULL` until first resolved.
- `member_count` counts **active** members only.
- `last_indexed_ledger` is the ledger at which this row was last written by ingestion.

## FleetMember

```sql
CREATE TABLE fleet_members (
    contract_id TEXT PRIMARY KEY,
    owner_address TEXT NOT NULL,
    tag TEXT NOT NULL,
    wasm_hash BYTEA NOT NULL,
    first_seen_ledger BIGINT NOT NULL,
    last_seen_ledger BIGINT NOT NULL,
    last_verified_ledger BIGINT,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    FOREIGN KEY (owner_address, tag) REFERENCES fleets(owner_address, tag)
);
```

- One row per contract (`contract_id` is the primary key) — a contract that changes fleets is
  re-parented, not duplicated.
- `active = FALSE` marks former members; rows are never deleted.
- `wasm_hash` is the hash observed for that member at indexing time.

## FleetRelease

```sql
CREATE TABLE fleet_releases (
    id BIGSERIAL PRIMARY KEY,
    owner_address TEXT NOT NULL,
    tag TEXT NOT NULL,
    old_wasm_hash BYTEA,
    new_wasm_hash BYTEA NOT NULL,
    ledger BIGINT NOT NULL,
    tx_hash TEXT NOT NULL,
    observed_at TIMESTAMPTZ NOT NULL
);
```

- One row per observed change of the owner's hash under a tag.
- `old_wasm_hash` is `NULL` for the first observed state of a fleet.
- Indexed on `(owner_address, tag, ledger)`.

## Verification

```sql
CREATE TABLE fleet_verifications (
    id BIGSERIAL PRIMARY KEY,
    owner_address TEXT NOT NULL,
    tag TEXT NOT NULL,
    expected_wasm_hash BYTEA NOT NULL,
    total_members BIGINT NOT NULL,
    matching_members BIGINT NOT NULL,
    mismatching_members BIGINT NOT NULL,
    missing_members BIGINT NOT NULL,
    status TEXT NOT NULL,
    indexed_through BIGINT NOT NULL,
    verified_at TIMESTAMPTZ NOT NULL
);
```

- Append-only audit log; every verification run inserts a row.
- `status` holds one of `HEALTHY`, `DRIFT`, `BROKEN_REFERENCE`, `INCOMPLETE`, `UNKNOWN`.
- `indexed_through` records the checkpoint at the time of verification — this is what makes
  "verified" reproducible.
- Indexed on `(owner_address, tag, verified_at)`.

## Checkpoint

```sql
CREATE TABLE indexer_checkpoints (
    stream TEXT PRIMARY KEY,
    ledger BIGINT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);
```

- One row per stream; the main pipeline uses stream name `main`.
- Stores the last **committed** ledger sequence.
- Written inside the same transaction as the ledger's data writes.

## Relationships in practice

```text
fleets (C..., vault-v1)
   ├── fleet_members: Contract A (active), Contract B (active), Contract C (inactive)
   ├── fleet_releases: #1 WASM A→B at ledger 102000
   └── fleet_verifications: #7 status DRIFT at 10:15Z

indexer_checkpoints: main → 105432
```

## Not part of the model

There is no table for tokens, balances, fees, staking, users, sessions, or contracts deployed by
this project. V1 indexes observed protocol state only.
