# Architecture Guide

The `soroban-fleet-registry` (`sfr`) is an operational indexing, resolution, and verification system for **Soroban CAP-85 externally managed contract executables**.

## System Overview

```
                      +-----------------------------+
                      |   Stellar RPC JSON-RPC API  |
                      +--------------+--------------+
                                     |
               +---------------------+---------------------+
               |                                           |
               v                                           v
    [ Ingestion Pipeline ]                      [ Executable Resolver ]
    Reads Ledgers (getLedgers)                  Current State (getLedgerEntries)
    Parses LedgerCloseMeta                      Resolves Instance & Tag Bytes
               |                                           |
               +---------------------+---------------------+
                                     |
                                     v
                        +-------------------------+
                        |  PostgreSQL Database    |
                        |  - fleets               |
                        |  - fleet_members        |
                        |  - fleet_releases       |
                        |  - fleet_verifications  |
                        |  - indexer_checkpoints  |
                        +------------+------------+
                                     |
               +---------------------+---------------------+
               |                                           |
               v                                           v
      [ Verification Engine ]                       [ REST API & CLI ]
      Evaluates Active Set                          Exposes Endpoints
      Asserts Integrity Status                      CLI Tooling
```

## Core Components

### 1. CAP-85 Decoder (`internal/cap85`)
- Decodes the `xdr.ContractExecutable` union.
- Identifies `CONTRACT_EXECUTABLE_EXTERNAL_REF` containing `(executable_owner, tag)`.
- Rejects malformed XDR, preserves tag casing, whitespace, and byte representation canonically.
- Ignores direct WASM and Stellar Asset executables cleanly.

### 2. Live Resolver (`internal/cap85`)
- Queries live Stellar RPC using batched `getLedgerEntries`.
- Resolves contract instances to their executable definitions.
- For external references, resolves the owner contract's persistent data entry keyed by the executable tag (`SCV_EXECUTABLE_TAG`) to extract the 32-byte WASM hash.
- Identifies broken references when owner data or referenced entries are missing.

### 3. Ingestion Pipeline (`internal/ingest`)
- Reads ledger metadata sequentially using `getLedgers`.
- Extracts ledger changes from `xdr.LedgerCloseMeta`.
- Detects new and updated contract instances pointing to CAP-85 external references.
- Detects tag updates on owner contracts emitting new WASM hashes.
- Enforces database write atomicity: fleet updates, member updates, release records, and indexer checkpoints commit in a single database transaction per ledger.

### 4. Fleet and Membership Services (`internal/fleet`)
- Manages fleet lifecycle and membership sets.
- Fleet identity is `(owner_address, tag)`.
- Maintains historical membership: contracts that transition away from a fleet are preserved as inactive records rather than physically deleted.

### 5. Verification Engine (`internal/verification`)
- Loads active member contracts for a fleet.
- Resolves live executable state across all members.
- Compares resolved WASM hashes against the expected hash.
- Enforces strict classification: `HEALTHY`, `DRIFT`, `BROKEN_REFERENCE`, `INCOMPLETE`, or `UNKNOWN`.
- Persists verification audit records to `fleet_verifications`.

### 6. HTTP API (`internal/api`)
- Provides RESTful JSON endpoints described by OpenAPI 3.1.
- Implements bounded pagination (`limit`, `offset`) and structured JSON error responses.
- Read-only observer interface.

### 7. CLI (`cmd/sfr`)
- Binary `sfr` supporting human-readable and machine-readable (`--json`) output.
- Emits deterministic exit codes: 0 (`HEALTHY`/success), 1 (`DRIFT`), 2 (`BROKEN_REFERENCE`), 3 (`INCOMPLETE`), 4 (invalid input), 5 (internal error).
