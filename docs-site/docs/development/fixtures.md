# Fixtures

Deterministic test data lives in `fixtures/`. Fixtures let tests verify classification and
ingestion logic without contacting a Stellar network.

```text
fixtures/
├── healthy/          fleet_healthy.json, external_ref.xdr
├── drift/            fleet_drift.json
├── broken-reference/ broken_owner_missing.json, broken_invalid_hash.json
├── releases/         release_upgrade.json
├── malformed/        malformed_tag.json, malformed_xdr.bin
├── fixtures.go       embed + loaders
└── fixtures_test.go  integrity tests
```

Files are embedded at compile time via `go:embed`, so tests never depend on the working
directory.

## Sets and what they verify

### `healthy/`

A fleet where every member resolves to the expected WASM hash.

| File | Purpose |
|---|---|
| `fleet_healthy.json` | Full scenario: owner, tag, expected hash, active members that all match |
| `external_ref.xdr` | Encoded external reference used to test XDR decoding |

**Verifies**: the happy path — membership decoding, resolution, and a `HEALTHY` classification
with complete coverage.

### `drift/`

`fleet_drift.json` — members that resolve to hashes different from the expected one.

**Verifies**: `DRIFT` classification, correct mismatch counting, and that one divergent member is
enough to block `HEALTHY`.

### `broken-reference/`

| File | Purpose |
|---|---|
| `broken_owner_missing.json` | Owner contract has no entry for the tag |
| `broken_invalid_hash.json` | Owner entry exists but holds an invalid value |

**Verifies**: `BROKEN_REFERENCE` classification and the resolver's `ErrBrokenReference` path —
distinct from drift and from missing data.

### `releases/`

`release_upgrade.json` — an owner changing the hash under a tag from `old_wasm` to `new_wasm` at
a given ledger and transaction.

**Verifies**: release detection, old/new hash persistence, ordering, and deduplication of
replays.

### `malformed/`

| File | Purpose |
|---|---|
| `malformed_tag.json` | Invalid tag input (empty / invalid UTF-8) |
| `malformed_xdr.bin` | Corrupt XDR bytes |

**Verifies**: the decoder rejects bad input with `ErrInvalidTag` / decode errors instead of
panicking or silently normalizing tags.

## Fixture schema

```json
{
  "owner_address": "C...",
  "tag": "vault-v1",
  "expected_wasm": "91aa...8e83",
  "members": [
    { "contract_id": "C...", "resolved_wasm": "91aa...8e83", "active": true }
  ],
  "ledger": 100000,
  "metadata": {}
}
```

Release fixtures use `{ owner_address, tag, old_wasm, new_wasm, ledger, tx_hash }`.

Loaders: `fixtures.LoadHealthyFixture()`, `LoadDriftFixture()`, `LoadReleaseFixture()`, plus
exported byte slices (`BrokenOwnerMissingJSON`, `MalformedTagJSON`, `MalformedXDRBin`, …).

## Running fixture tests

```bash
go test -v ./fixtures/...
go test -v ./integration/...
```

Integration tests combine fixtures with a local PostgreSQL database and a controllable resolver,
so scenario tests (healthy, drift, broken reference, replay) run hermetically.

Related: [Testing](./testing.md) · [Verification states](../concepts/verification-states.md)
