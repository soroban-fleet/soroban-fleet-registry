# Technical Overview

## Architecture

Soroban Fleet Registry is organized into modular subsystems:

```text
Stellar RPC / LedgerCloseMeta
              |
              v
       Indexer Pipeline
              |
              v
      PostgreSQL Storage <------- REST API (Go / chi)
              ^                         |
              |                         v
      Verification Engine <----- Next.js Web App / CLI (`sfr`)
```

- **Indexer Pipeline** (`internal/ingest`): Ingests Stellar ledger closing metadata via JSON-RPC, decodes contract executable entries, extracts `(owner, tag)` pairings, detects WASM bytecode hash updates, and updates membership state in atomic PostgreSQL transactions. Checkpoints are recorded strictly after successful commit, ensuring restart resilience and idempotent replay.
- **CAP-85 Parser & Resolver** (`internal/cap85`): Inspects XDR contract executables (`CONTRACT_EXECUTABLE_EXTERNAL_REF`), preserves canonical tag strings byte-for-byte without case normalization, and resolves the underlying 32-byte WASM hash from the owner contract's persistent data entries.
- **PostgreSQL Database** (`migrations/`): Normalized relational persistence for fleets, active/historical fleet members, release transitions, verification audits, and ledger checkpoint progress.
- **Deterministic Verification Engine** (`internal/verification`): Validates fleet member executables against live on-chain state, detects drift and broken references, verifies indexer freshness against the network's latest ledger, and applies safety precedence rules.
- **REST API** (`internal/api`): OpenAPI 3.1 compliant read-only HTTP service providing paginated fleet listings, member details, release history, contract inspection, and on-demand verification audits.
- **Command-Line Interface** (`cmd/sfr`): Compiled binary providing administrative inspection, membership auditing, ingestion triggers, and CI-ready verification with semantic process exit codes (0 for `HEALTHY`, 1 for `DRIFT`, 2 for `BROKEN_REFERENCE`, 3 for `INCOMPLETE`, 4 for invalid arguments, 5 for internal error/`UNKNOWN`).
- **Web Application** (`web/`): Server-rendered and client-hydrated Next.js 14 application with accessibility-compliant audit views, live search, drift visualization, and contract inspection.
