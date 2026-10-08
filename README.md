# soroban-fleet-registry (`sfr`)

<p align="center">
  <img src="docs/assets/logo.png" alt="Soroban Fleet Registry Banner" width="100%" />
</p>

<p align="center">
  <a href="https://github.com/soroban-fleet/soroban-fleet-registry/actions/workflows/ci.yml"><img src="https://github.com/soroban-fleet/soroban-fleet-registry/actions/workflows/ci.yml/badge.svg?branch=main" alt="CI" /></a>
  <a href="go.mod"><img src="https://img.shields.io/github/go-mod/go-version/soroban-fleet/soroban-fleet-registry" alt="Go Version" /></a>
  <a href="https://github.com/soroban-fleet/soroban-fleet-registry/releases"><img src="https://img.shields.io/github/v/release/soroban-fleet/soroban-fleet-registry?color=blue" alt="Release" /></a>
  <a href="https://github.com/stellar/stellar-protocol/blob/master/core/cap-0085.md"><img src="https://img.shields.io/badge/Stellar-Protocol%2028-blue.svg" alt="Protocol" /></a>
</p>

<p align="center">
  CAP-85 fleet discovery, indexing, executable resolution, history, and verification for Soroban contracts.
</p>

> **Project status**: Early infrastructure release (v1.0.0). Read-only observer and deterministic indexer.

Documentation:
https://soroban-fleet.github.io/soroban-fleet-registry/

- [Quick Start](https://soroban-fleet.github.io/soroban-fleet-registry/docs/introduction/quick-start)
- [Architecture](https://soroban-fleet.github.io/soroban-fleet-registry/docs/architecture/system-architecture)
- [API](https://soroban-fleet.github.io/soroban-fleet-registry/docs/api/overview)
- [Contributing](https://soroban-fleet.github.io/soroban-fleet-registry/docs/contributing/workflow)
- [Security](https://soroban-fleet.github.io/soroban-fleet-registry/docs/contributing/security)

---

## 1. What is this?

`soroban-fleet-registry` (`sfr`) is an open-source indexing and verification service for Soroban smart contracts. CAP-85 was introduced in Protocol 28 and is supported on the current Protocol 29 network. SFR discovers, tracks, and verifies contracts that use CAP-85 externally managed contract executables.

---

## 2. Problem

When a protocol uses a factory pattern to deploy dozens or thousands of contract instances sharing the same WASM code (such as vaults, accounts, or liquidity pools), upgrading every contract individually requires submitting individual upgrade transactions for each instance. This process is prone to partial rollouts, temporary version drift, and manual tracking errors.

CAP-85 introduces externally managed executables to enable simultaneous upgrades. However, the blockchain does not natively maintain an index of which contract instances point to a shared external reference, what the release history has been, or whether all active instances actually resolve to the expected executable. `soroban-fleet-registry` provides this missing operational layer.

---

## 3. Why CAP-85?

Soroban CAP-85 introduces the `CONTRACT_EXECUTABLE_EXTERNAL_REF` executable variant. Instead of binding a contract instance directly to a WASM bytecode hash, an external reference binds the instance to an `(owner, tag)` reference. The owner contract holds the canonical WASM hash in a persistent contract data entry keyed by that tag string (`SCV_EXECUTABLE_TAG`). When the owner updates that entry, all referencing contract instances immediately execute the updated code on their next invocation.

---

## 4. How Fleet Identity Works

A **fleet** is identified by:

**Fleet Identity:** `(owner_address, tag)`

- **Owner Address**: The StrKey address of the contract that manages the executable.
- **Tag**: An exact, canonical string identifier (e.g. `vault-v1`).
- **Fleet Members**: All contract instances holding an `ExternalRef` pointing to `(owner, tag)`.
- **Active WASM**: The 32-byte bytecode hash currently stored by the owner contract under the tag key.

When the owner updates the tag entry from WASM A to WASM B:
- The fleet identity does not change.
- A new release record is created.
- All members inherit the upgrade simultaneously.

---

## 5. How Verification Works

The verification engine verifies fleet integrity by:
1. Loading the active member set from the indexed database.
2. Querying live Stellar RPC to resolve each member's current on-chain executable state.
3. Comparing resolved WASM hashes against the expected hash.
4. Checking indexer coverage against the latest network ledger.
5. Classifying status according to strict protocol safety rules and persisting the audit record.

```text
                                (owner, tag)
                                      |
                                      v
                           [Current Tag WASM Hash]
                                      |
            +-------------------------+-------------------------+
            |                                                   |
      Member C_1                                          Member C_n
    (Live Resolution)                                   (Live Resolution)
            |                                                   |
    Matches Tag Hash?                                   Matches Tag Hash?
      [YES]     [NO]                                      [YES]     [NO]
        |         |                                         |         |
        v         v                                         v         v
     HEALTHY    DRIFT                                    HEALTHY    DRIFT
```

---

## 6. Architecture

```text
User / Browser Client
         |
         v
Next.js 14 Web Application (web/)
         |
         v
Go REST API (internal/api)
         |
         +---- PostgreSQL 16+ (fleets, members, releases, checkpoints)
         |
         +---- Indexer Pipeline (internal/ingest)
                 |
                 +---- Stellar Soroban RPC (Live State & Ledger Stream)
```

---

## 7. Quickstart

### Prerequisites
- Go `1.26+`
- PostgreSQL `16+`
- Node.js `18+` (for web UI)

### Running Locally

```bash
# 1. Start PostgreSQL
docker compose up -d postgres

# 2. Apply migrations
export SFR_DATABASE_URL="postgres://sfr:sfrpassword@localhost:5432/sfr?sslmode=disable"
make build
./bin/sfr migrate up

# 3. Start the API server
export SFR_HTTP_ADDR=":8080"
export SFR_RPC_URL="https://soroban-testnet.stellar.org"
./bin/sfr api

# 4. Run the web frontend
cd web && npm install && npm run dev
```

---

## 8. CLI Usage

```bash
# Verify a fleet and exit with status code
sfr fleet verify \
  --owner CA3D5KRYMCMCZVACQY3CGINPVPUM2OAL6YHT3ZVM5C2MUMAA252AAAAB \
  --tag vault-v1

# Inspect a contract's live resolution
sfr contract inspect \
  --contract CB222222222222222222222222222222222222222222222222222222

# List all fleet members
sfr fleet members --owner CA3D... --tag vault-v1
```

Add `--json` to any command for machine-readable output.

---

## 9. REST API

The server exposes an OpenAPI 3.1 compliant REST API:

- `GET /health`: Service liveness check
- `GET /v1/fleets`: List indexed fleets (paginated)
- `GET /v1/fleets/{owner}/{tag}`: Fleet details and current WASM hash
- `GET /v1/fleets/{owner}/{tag}/members`: Fleet member contracts and drift flags
- `GET /v1/fleets/{owner}/{tag}/history`: Historical verification records
- `GET /v1/fleets/{owner}/{tag}/releases`: Release audit trail
- `GET /v1/fleets/{owner}/{tag}/verify`: Run on-demand verification
- `GET /v1/contracts/{contract_id}`: Contract lookup and fleet membership

See [`api/openapi.yaml`](api/openapi.yaml) for the full specification.

---

## 10. Web Application

The `web/` directory contains the read-only web interface built with Next.js 14, React 18, and TypeScript:

- **Fleet Directory**: Searchable list of indexed fleets with health badges.
- **Fleet Detail**: Member counts, current WASM hash, and subview navigation.
- **Members View**: Paginated member table highlighting drifted contracts.
- **Release Timeline**: Visual chronological history of executable upgrades.
- **Verification Panel**: Detailed breakdown of matched vs. drifted contracts and indexer lag.
- **Contract Inspector**: Dedicated contract resolution view.

---

## 11. Configuration

| Variable | Description | Default |
|---|---|---|
| `SFR_NETWORK` | Stellar network identifier (`testnet`, `pubnet`) | `testnet` |
| `SFR_RPC_URL` | Soroban RPC endpoint URL | `""` |
| `SFR_DATABASE_URL` | PostgreSQL connection string | `""` |
| `SFR_HTTP_ADDR` | HTTP API listen address | `:8080` |
| `SFR_LOG_LEVEL` | Log level (`debug`, `info`, `warn`, `error`) | `info` |
| `SFR_INDEXER_START_LEDGER` | Starting ledger sequence for indexer | `0` |
| `SFR_INDEXER_BATCH_SIZE` | RPC ledger batch size | `50` |
| `SFR_CORS_ALLOWED_ORIGINS` | Allowed CORS origins (comma-separated) | `""` |

See [`.env.example`](.env.example) and [`web/.env.example`](web/.env.example).

---

## 12. Verification Semantics

| Status | Exit Code | Description |
|---|---|---|
| **`HEALTHY`** | `0` | All active members resolve on-chain, all match the expected WASM hash, total members > 0, and indexer coverage is complete. |
| **`DRIFT`** | `1` | One or more active member contracts resolve to a different WASM hash. |
| **`BROKEN_REFERENCE`** | `2` | The fleet's external reference cannot resolve (owner data missing, invalid value, or malformed reference). |
| **`INCOMPLETE`** | `3` | The indexer has not reached the ledger height required to make an assertion (`checkpoint < latest_ledger`), or members are missing. |
| **`UNKNOWN`** | `5` | Insufficient information exists to safely assert correctness. |

> **Safety Rule**: Never report `HEALTHY` simply because no mismatch was found.

---

## 13. Validation

The implementation has been verified against:
- **Testnet Connectivity**: Verified live against Stellar Testnet RPC (`https://soroban-testnet.stellar.org`) on Protocol 29.
- **Automated Test Suite**: Unit, integration, and scenario tests with the Go race detector (`go test -v -race -p 1 ./...`).
- **Frontend Acceptance Tests**: Complete suite of Vitest and Testing Library acceptance flows covering healthy, drifted, incomplete, and contract inspection flows (`47 passed`).
- **Replay & Recovery**: Atomic checkpointing and idempotent replay verified (`TestPipeline_IdempotentReplayAndResume`).
- **Database Migrations**: Two-way migration testing (`sfr migrate down` followed by `sfr migrate up`).

---

## 14. Deployment

See [`docs/deployment.md`](docs/deployment.md) for full service topology, Docker setup, and production guidelines.

---

## 15. Security

See [`SECURITY.md`](SECURITY.md) for vulnerability disclosure procedures.

---

## 16. Contributing

See [`CONTRIBUTING.md`](CONTRIBUTING.md) for development workflows, commit conventions, and pull request expectations.

---

## 17. Maintainers & Community

`soroban-fleet-registry` is maintained as an open-source project within the `soroban-fleet` organization.

- **Issues & Discussions**: Please open an issue on [GitHub Issues](https://github.com/soroban-fleet/soroban-fleet-registry/issues).
- **Security Inquiries**: Use GitHub Private Vulnerability Reporting via [Security Advisories](https://github.com/soroban-fleet/soroban-fleet-registry/security/advisories).

---

## 18. Out of Scope

V1 is strictly a **read-only observer and indexer**.

V1 does **NOT**:
- Sign transactions;
- Manage private keys or wallets;
- Execute contract upgrades or mutations;
- Deploy Soroban contracts;
- Act as a general-purpose blockchain explorer.

---

## 19. License

This project is licensed under the [Apache License 2.0](LICENSE).