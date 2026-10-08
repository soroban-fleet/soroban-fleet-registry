# Data flow

One direction: from Stellar ledger state to the user.

```text
Stellar ledger/state
       ↓
ingestion
       ↓
CAP-85 decoding
       ↓
owner + tag
       ↓
WASM resolution
       ↓
fleet/member persistence
       ↓
verification
       ↓
API
       ↓
CLI / web
```

## Stage by stage

### 1. Stellar ledger/state

The indexer calls `getLedgers` over Stellar RPC and receives raw `LedgerCloseMeta` for each
sequence. Verification and contract inspection separately call `getLedgerEntries` and
`getLatestLedger` for current state.

### 2. Ingestion

`ExtractChangesFromLedgerCloseMeta` flattens transaction metadata into two change sets:

- **TagChanges** — owner contract updated the WASM hash stored under a tag;
- **ContractChanges** — a contract instance's executable was created, changed, or removed.

### 3. CAP-85 decoding

For each contract change, the decoder inspects `ContractExecutable`. Only
`CONTRACT_EXECUTABLE_EXTERNAL_REF` is relevant; direct WASM and Stellar Asset executables are
skipped. The external reference yields `(executable_owner, tag)` with the tag preserved exactly.

### 4. Owner + tag

The pair becomes the fleet identity `(owner, tag)`. Fleet rows are upserted with that key.

### 5. WASM resolution

The active hash comes from the first source available, in order:

1. the tag change in the same ledger (`knownWasm`);
2. the fleet row already in the database;
3. live resolution over RPC (`ResolveContractExecutable`).

During verification, resolution is always live: every active member is resolved through
`getLedgerEntries`, and the owner entry is read to obtain the expected hash source.

### 6. Fleet/member persistence

All writes for one ledger — fleet upserts, member upserts or deactivations, release rows,
checkpoint advance — commit in a single database transaction.

### 7. Verification

On request, the engine loads active members, resolves each live, compares hashes, checks coverage,
classifies, and inserts a row into `fleet_verifications`.

### 8. API

Read-only endpoints return the stored data as JSON with pagination metadata.

### 9. CLI / web

Both are API clients. `sfr` formats output (or emits JSON) and returns a status exit code; the web
app renders the same payloads.

## Read paths

| Consumer | Read path |
|---|---|
| `sfr fleet list` / web fleet directory | `fleets` via API |
| `sfr fleet members` / members table | `fleet_members` via API |
| `sfr fleet releases` / history timeline | `fleet_releases` via API |
| `sfr fleet history` / verification list | `fleet_verifications` via API |
| `sfr fleet verify` / verify view | live RPC + `fleet_members` + `indexer_checkpoints` → `fleet_verifications` |
| `sfr contract inspect` / contract page | `fleet_members` + live RPC |

## What never flows

- private keys, signatures, or wallet data — none exist in this system;
- writes from API or web to the database — read-only;
- writes from any component to the Stellar network — none.
