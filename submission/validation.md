# Validation and Test Evidence

The Soroban Fleet Registry repository contains comprehensive, automated test suites verifying all core behaviors under strict execution conditions.

## Automated Backend Test Suite

Executed via `go test -v -race -p 1 ./...`:
- **Race Detection**: Zero data races detected across concurrent operations.
- **Go Vet**: `go vet ./...` completed with zero warnings.
- **Packages Verified**:
  - `cmd/sfr`: CLI argument parsing, subcommands, and exit codes (0, 1, 2, 3, 4, 5).
  - `fixtures`: Serialization and schema validation of controlled protocol fixtures.
  - `integration`: Flagship end-to-end multi-instance lifecycle, factory patterns, and Section 26 verification scenarios.
  - `internal/api`: Full HTTP routing, pagination headers, CORS negotiation, and error envelopes.
  - `internal/cap85`: Binary XDR parsing of `CONTRACT_EXECUTABLE_EXTERNAL_REF`, StrKey conversion, exact case-sensitive tag preservation, and batch resolution.
  - `internal/config`: Environment variable ingestion, defaulting, and configuration boundaries.
  - `internal/fleet`: Fleet CRUD, contract inspection, active membership transitions, and foreign key integrity.
  - `internal/history`: Release transition persistence, chronological ledger queries, and hash transitions.
  - `internal/ingest`: Ledger meta parsing, atomic batch persistence, checkpoint progress, and idempotent replay resume (`TestPipeline_IdempotentReplayAndResume`).
  - `internal/rpc`: Soroban JSON-RPC ledger and entry querying with automatic chunking and batching.
  - `internal/verification`: Deterministic classification across all 5 verification states (`HEALTHY`, `DRIFT`, `BROKEN_REFERENCE`, `INCOMPLETE`, `UNKNOWN`).
  - `migrations`: Two-way database schema migrations (`up`, idempotent re-run, and clean `down`).

## Frontend Test Suite

Executed via `npm run test` in `web/` (Vitest + Testing Library):
- **11 Test Files Passed (11/11)**
- **47 Tests Passed (47/47)**
- **Coverage**:
  - API client error normalization and response typing
  - Accessible layout, navigation, empty states, and loading indicators
  - Status badge formatting across all 5 verification states
  - Fleet listing and pagination
  - Fleet detail, active WASM presentation, and coverage metrics
  - Active member rosters
  - Historical release timelines
  - Verification audit execution and drift indicators
  - Single contract inspection view

## Documentation Build

Executed via `npm ci && npm run build` in `docs-site/`:
- **Docusaurus 3**: Production static build generated successfully.
- **Link Integrity**: Verified zero broken internal links or anchors (`onBrokenLinks: 'throw'`, `onBrokenAnchors: 'throw'`).
- **Search**: Local search index built and packaged cleanly.
