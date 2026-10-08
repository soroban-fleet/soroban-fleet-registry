# External executable

An *external executable* (CAP-85 `ExternalRef`) is a contract executable that points at state
managed by another contract instead of storing a WASM hash directly.

This page describes exactly how Soroban Fleet Registry handles an `ExternalRef`.

## What SFR does with an ExternalRef

> Contracts reference externally managed executable state, and SFR resolves and indexes that
> relationship.

Concretely, four operations:

| Step | Operation | Implementation |
|---|---|---|
| 1. Detect | Read ledger changes; find contract instances whose executable is `CONTRACT_EXECUTABLE_EXTERNAL_REF` | `internal/ingest`, `internal/cap85` |
| 2. Decode | Extract `executable_owner` and `tag` from the reference, preserving the tag byte-for-byte | `internal/cap85` decoder |
| 3. Resolve | Read the owner's persistent data entry keyed by `SCV_EXECUTABLE_TAG` to obtain the 32-byte WASM hash | `internal/cap85` resolver, Stellar RPC `getLedgerEntries` |
| 4. Index | Upsert the fleet and its membership, record a release if the hash changed | `internal/fleet`, `internal/history` |

## What SFR does not do with an ExternalRef

An `ExternalRef` is **not** an automatic upgrade of all contracts. Nothing in this service upgrades
anything.

- SFR does not write ledger entries.
- SFR does not call host functions or contract functions.
- SFR does not decide which contracts belong to a fleet — it reports what the ledger says.
- SFR does not verify code correctness. It compares hashes.

Upgrades happen because the owner contract updates its own storage, which the Stellar protocol
then applies to every referencing instance on next invocation. SFR merely notices.

## Resolution detail

```text
member contract (live state from RPC)
   │
   ▼
executable = CONTRACT_EXECUTABLE_EXTERNAL_REF
   ├── executable_owner = C...
   └── tag = "vault-v1"
            │
            ▼
owner contract persistent data entry
   ├── key = SCV_EXECUTABLE_TAG("vault-v1")
   └── val = <32-byte hash>
            │
            ▼
ResolvedExecutable { kind: "EXTERNAL_REF", fleet, wasm_hash }
```

Resolved states SFR can produce:

| Result | Condition |
|---|---|
| `EXTERNAL_REF` + WASM hash | reference and owner entry both resolve |
| `WASM` | instance uses a direct WASM executable (not a fleet member) |
| broken reference | owner entry missing, owner absent, or entry malformed → `ErrBrokenReference` |

## Errors

| Error | Meaning |
|---|---|
| `ErrNotExternalRef` | executable is not an external reference |
| `ErrInvalidExternalRef` | reference is malformed (empty owner or tag) |
| `ErrBrokenReference` | owner data or referenced entry cannot be resolved |
| `ErrInvalidTag` | tag empty or not valid UTF-8 |
| `ErrInvalidContractID` | address is not a valid StrKey contract address |

## Not every contract is a member

Plenty of contracts use direct WASM executables. `sfr contract inspect` reports
`Is Fleet Member: false` for them — see [Inspect a contract](../guides/inspect-contract.md).

Next: [Fleet lifecycle](./fleet-lifecycle.md)
