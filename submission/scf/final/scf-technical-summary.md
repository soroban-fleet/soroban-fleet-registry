# SCF Technical Summary

## 1. Current Architecture (V1.0.0 Foundation)

Soroban Fleet Registry is implemented as a unified, modular Go and TypeScript system:

- **Ingestion & Pipeline** (`internal/ingest`): Connects to Soroban RPC, streams closed ledger metadata (`LedgerCloseMeta`), extracts contract executable state modifications, and executes atomic relational persistence in PostgreSQL. Checkpoints advance strictly after successful transaction commit, ensuring crash resilience and idempotent replay.
- **CAP-85 Parser & Resolver** (`internal/cap85`): Deserializes binary XDR contract executable structures, validates StrKey address encoding (`C...`), preserves case-sensitive tag strings byte-for-byte, and resolves active WASM bytecode hashes via owner contract storage keys (`SCV_EXECUTABLE_TAG`).
- **Deterministic Verifier Engine** (`internal/verification`): Validates live on-chain instance executable hashes against expected builds and evaluates indexer coverage (`indexed_through >= latest_ledger`). The engine follows strict precedence: broken references, mismatches, missing instances, and coverage lag are classified before asserting `HEALTHY`.
- **REST API** (`internal/api`): Clean OpenAPI 3.1 compliant HTTP interface written in Go (chi router), delivering paginated listings, detailed member rosters, release timelines, contract inspection, and audit results.
- **CLI Tool** (`cmd/sfr`): Compiled binary providing administrative inspection and CI verification with deterministic exit codes (0 for `HEALTHY`, 1 for `DRIFT`, 2 for `BROKEN_REFERENCE`, 3 for `INCOMPLETE`).
- **Web Interface** (`web/`): Next.js 14 frontend providing responsive fleet dashboards, release timelines, and accessibility-compliant audit views.

---

## 2. Protocol Boundaries & Security Scope

- **Read-Only Invariant**: The service holds no private keys, signs no transactions, connects to no wallets, and deploys no custom Soroban smart contracts.
- **Source of Truth**: The Stellar ledger is the canonical truth; the PostgreSQL database is an ephemeral, fully resynchronizable relational index.

---

## 3. Future Architecture (Funded Milestones)

- **Tranche 1**: Token-bucket IP rate limiting middleware (Issue #11), Prometheus OpenMetrics exporter (Issue #12), WASM TTL caching (Issue #5), and bounded ledger replay CLI (Issue #4).
- **Tranche 2**: Live scenario recording CLI (Issue #14), automated migration test harness (Issue #15), web member drift diff modal (Issue #8), and CSV/JSON export tooling (Issue #9).
- **Tranche 3**: Dedicated high-availability Mainnet cluster deployment, history archive backfill (Issue #16), and compliance audit report generator (Issue #10).
