# System architecture

SFR is composed of five parts: a web frontend, a REST API, a PostgreSQL database, an indexer, and
Stellar RPC as the data source.

```text
                    USER
                      |
                      v
              Next.js Web App
                      |
                      v
                  Go API
                  /     \
                 /       \
        PostgreSQL       Indexer
                           |
                      Stellar data
```

## Components

### User / browser

Reads pages and issues GET requests. Never talks to PostgreSQL or Stellar RPC directly.

### Next.js web app (`web/`)

Stateless presentation layer built with Next.js 14, React 18, and TypeScript.

- Renders the fleet directory, fleet detail, members, history, verification panel, and contract
  inspector.
- Calls only the REST API defined in [`api/openapi.yaml`](https://github.com/soroban-fleet/soroban-fleet-registry/blob/main/api/openapi.yaml).
- Contains no transaction signing, no wallet integration, and no private keys.
- Does not re-implement verification logic in TypeScript; it displays what the API returns.

### Go REST API (`internal/api`)

Read-only HTTP service (`sfr api`).

- Routes: `/health`, `/v1/health`, `/v1/fleets…`, `/v1/contracts/{contract_id}`.
- Bounded pagination (`limit` default 20, max 100; `offset` default 0).
- Structured JSON errors: `{"error":{"code","message"}}`.
- CORS controlled by `SFR_CORS_ALLOWED_ORIGINS`.
- Panic recovery middleware; a handler panic becomes `500 INTERNAL_ERROR`, not a crashed process.
- Reads PostgreSQL; uses Stellar RPC only for live resolution during verification and contract
  inspection.

### PostgreSQL 16+

Authoritative store for indexed state:

| Table | Purpose |
|---|---|
| `fleets` | fleet identity, current WASM, member count, ledger bounds |
| `fleet_members` | per-contract membership and active flag |
| `fleet_releases` | observed executable changes |
| `fleet_verifications` | persisted verification results |
| `indexer_checkpoints` | per-stream ingestion checkpoint |

Database access is parameterized SQL; the API's production role can be restricted to `SELECT`.

### Indexer (`internal/ingest`, `sfr ingest`)

Background pipeline.

- Fetches ledgers in batches (`SFR_INDEXER_BATCH_SIZE`, default 50) via `getLedgers`.
- Decodes `LedgerCloseMeta`, extracts contract executable and instance changes.
- Applies fleet, membership, and release writes plus checkpoint advance in **one** database
  transaction per ledger.
- Resolves owner entries over RPC when needed.
- Resumes from `checkpoint + 1` on restart; replay is idempotent.

### Stellar RPC

External dependency, not owned by this project.

- `getLedgers` for historical ingestion.
- `getLedgerEntries` for live resolution of instances and owner tag entries.
- `getLatestLedger` for coverage comparison.

## Process topology

The same binary exposes three modes:

```bash
sfr migrate up   # schema migrations
sfr api          # HTTP API
sfr ingest       # indexing pipeline
```

API and indexer run as separate processes against the same database. The API works even when the
indexer has not caught up; it just reports incomplete data honestly.

## Boundaries

- No service holds private keys or signs transactions.
- No service mutates on-chain state.
- The indexer is the only component that writes to PostgreSQL (plus `migrate`).
- The frontend is the only component that runs in the browser.

See also: [Data flow](./data-flow.md) · [Indexing pipeline](./indexing-pipeline.md) ·
[Soroban boundary](./soroban-boundary.md)
