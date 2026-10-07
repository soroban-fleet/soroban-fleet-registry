# soroban-fleet-registry (`sfr`)

Production-grade indexing, discovery, resolution, and verification system for **Soroban CAP-85 externally managed contract executables**.

---

## 1. What is this?

`soroban-fleet-registry` (`sfr`) is an open-source indexing and verification service for Soroban smart contracts built on Stellar Protocol 28. It discovers, tracks, and verifies contracts that use CAP-85 externally managed contract executables.

---

## 2. What problem does it solve?

When a protocol uses a factory pattern to deploy dozens or thousands of contract instances sharing the same WASM code (such as vaults, accounts, or liquidity pools), upgrading every contract individually requires submitting individual upgrade transactions for each instance. This process is prone to partial rollouts, temporary version drift, and manual tracking errors.

CAP-85 introduces externally managed executables to enable simultaneous upgrades. However, the blockchain does not natively maintain an index of which contract instances point to a shared external reference, what the release history has been, or whether all active instances actually resolve to the expected executable. `soroban-fleet-registry` provides this missing operational layer.

---

## 3. Why CAP-85?

Soroban CAP-85 introduces the `CONTRACT_EXECUTABLE_EXTERNAL_REF` executable variant. Instead of binding a contract instance directly to a WASM bytecode hash, an external reference binds the instance to an `(owner, tag)` reference. The owner contract holds the canonical WASM hash in a persistent contract data entry keyed by that tag string (`SCV_EXECUTABLE_TAG`). When the owner updates that entry, all referencing contract instances immediately execute the updated code on their next invocation.

---

## 4. How does a fleet work?

A **fleet** is identified by:

$$\text{Fleet Identity} = (\text{owner\_address}, \text{tag})$$

- **Owner Address**: The StrKey address of the contract that manages the executable.
- **Tag**: An exact, canonical string identifier (e.g. `vault-v1`).
- **Fleet Members**: All contract instances holding an `ExternalRef` pointing to `(owner, tag)`.
- **Active WASM**: The 32-byte bytecode hash currently stored by the owner contract under the tag key.

When the owner updates the tag entry from WASM A to WASM B:
- The fleet identity does not change.
- A new release record is created.
- All members inherit the upgrade simultaneously.

---

## 5. How does verification work?

The verification engine verifies fleet integrity by:
1. Loading the active member set from the indexed database.
2. Querying live Stellar RPC to resolve each member's current on-chain executable state.
3. Comparing resolved WASM hashes against the expected hash.
4. Checking indexer coverage against the latest network ledger.
5. Classifying status according to strict protocol safety rules and persisting the audit record.

---

## 6. How do I run it?

### Quickstart

```bash
# Download dependencies
go mod download

# Run all tests
go test ./...

# Run code checks
go vet ./...
gofmt -w .

# Start local PostgreSQL
docker compose up -d postgres

# Apply schema migrations
export SFR_DATABASE_URL="postgres://sfr:sfrpassword@localhost:5432/sfr?sslmode=disable"
sfr migrate up

# Run ingestion pipeline against Stellar RPC
export SFR_RPC_URL="https://soroban-testnet.stellar.org"
sfr ingest

# Run REST API server
export SFR_HTTP_ADDR=":8080"
sfr api
```

---

## 7. How do I query a fleet?

### Using the CLI (`sfr`)

```bash
# List all indexed fleets
sfr fleet list

# Inspect fleet details
sfr fleet inspect \
  --owner CA3D5KRYMCMCZVACQY3CGINPVPUM2OAL6YHT3ZVM5C2MUMAA252AAAAB \
  --tag vault-v1

# List fleet members
sfr fleet members \
  --owner CA3D5KRYMCMCZVACQY3CGINPVPUM2OAL6YHT3ZVM5C2MUMAA252AAAAB \
  --tag vault-v1

# View release history
sfr fleet releases \
  --owner CA3D5KRYMCMCZVACQY3CGINPVPUM2OAL6YHT3ZVM5C2MUMAA252AAAAB \
  --tag vault-v1

# Run live verification
sfr fleet verify \
  --owner CA3D5KRYMCMCZVACQY3CGINPVPUM2OAL6YHT3ZVM5C2MUMAA252AAAAB \
  --tag vault-v1 \
  --expected-wasm 91aa618b74668f4d960fca5dfc37f0775d1797e8ff4d34f0bf26efbd99238e83

# Inspect an individual contract
sfr contract inspect \
  --contract CB222222222222222222222222222222222222222222222222222222
```

Add `--json` to any command for machine-readable JSON output.

### Using the REST API

- `GET /v1/fleets`
- `GET /v1/fleets/{owner}/{tag}`
- `GET /v1/fleets/{owner}/{tag}/members`
- `GET /v1/fleets/{owner}/{tag}/history`
- `GET /v1/fleets/{owner}/{tag}/releases`
- `GET /v1/fleets/{owner}/{tag}/verify`
- `GET /v1/contracts/{contract_id}`

See [`api/openapi.yaml`](api/openapi.yaml) for the full OpenAPI 3.1 specification.

---

## 8. What does each status mean?

| Status | Exit Code | Description |
|---|---|---|
| **`HEALTHY`** | `0` | All active members resolve on-chain, all match the expected WASM hash, total members > 0, and indexer coverage is complete. |
| **`DRIFT`** | `1` | One or more active member contracts resolve to a different WASM hash. |
| **`BROKEN_REFERENCE`** | `2` | The fleet's external reference cannot resolve (owner data missing, invalid value, or malformed reference). |
| **`INCOMPLETE`** | `3` | The indexer has not reached the ledger height required to make an assertion (`checkpoint < latest_ledger`), or members are missing. |
| **`UNKNOWN`** | `5` | Insufficient information exists to safely assert correctness. |

> **Safety Rule**: Never report `HEALTHY` simply because no mismatch was found.

---

## 9. What is currently supported?

- CAP-85 `CONTRACT_EXECUTABLE_EXTERNAL_REF` decoding and exact tag preservation.
- Live on-chain executable resolution using batched Stellar RPC calls.
- Historical ledger ingestion from `LedgerCloseMeta` with atomic PostgreSQL transactions.
- Idempotent checkpointing and ledger replay.
- Member lifecycle tracking (active/inactive states retained).
- Automatic release detection and audit trail logging.
- Deterministic verification engine with CLI exit codes.
- OpenAPI 3.1 compliant HTTP API.

---

## 10. What is intentionally out of scope?

V1 is strictly a **read-only observer and indexer**.

V1 does **NOT**:
- Sign transactions;
- Manage private keys or wallets;
- Execute contract upgrades or mutations;
- Deploy Soroban contracts;
- Act as a general-purpose blockchain explorer.

---

## Documentation

- [Architecture Guide](docs/architecture.md)
- [CAP-85 Specification](docs/cap85.md)
- [Indexing Model](docs/indexing.md)
- [Verification Engine](docs/verification.md)
- [Development & Operating Modes](docs/development.md)