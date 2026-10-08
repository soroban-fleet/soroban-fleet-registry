# Why CAP-85?

[CAP-85 "Externally managed contract executables"](https://github.com/stellar/stellar-protocol/blob/master/core/cap-0085.md)
is a Final CAP introduced in Stellar Protocol 28 and supported on the current Protocol 29 network. It adds a third executable variant to Soroban
contracts.

Primary source: `stellar/stellar-protocol`, `core/cap-0085.md`.

## Ordinary executable vs. externally managed executable

### Ordinary executable (direct WASM)

A contract instance stores its own WASM hash. To change the code, you update that instance.

```text
Contract instance  ──owns──▶  wasm_hash
```

If fifty instances share one build, fifty updates are required. CAP-85's motivation section notes
that a fleet too large to fit in one transaction cannot be upgraded atomically, which creates a
window where some instances run the old build and some run the new one, and that tracking every
active instance by hand is error-prone.

### Externally managed executable (external reference)

A contract instance stores a *reference* instead of a hash. The reference names another contract
(the owner) and a string (the tag). The owner contract stores the actual WASM hash in its own
persistent storage, keyed by that tag.

```text
Contract instance
   │
   ▼
CONTRACT_EXECUTABLE_EXTERNAL_REF
   ├── executable_owner: SCAddress
   └── tag:              SCString
             │
             ▼
Owner contract persistent data entry
   ├── key: SCV_EXECUTABLE_TAG (tag string)
   └── val: 32-byte WASM hash
             │
             ▼
         active WASM
```

When the owner updates the entry, every instance that references `(owner, tag)` executes the new
build on its next invocation. The upgrade happens without touching the individual instances.

## The three fields

| Field | Type | Meaning |
|---|---|---|
| `executable_owner` | `SCAddress` | Contract that holds the canonical WASM hash |
| `tag` | `SCString` | Exact string key for the entry inside the owner's storage |
| WASM hash | 32 bytes | Value stored by the owner under that tag |

The XDR shape from CAP-85:

```text
struct ContractExecutableExternalRef {
    SCAddress executable_owner;
    SCString tag;
};
```

The tag key uses the `SCV_EXECUTABLE_TAG` SCVal variant. Per CAP-85 the owner cannot delete a tag
entry, and an update must point at a Wasm that has already been uploaded — the protocol validates
this at write time.

## Tag rules

A tag is an exact string. CAP-85 and this implementation preserve it as-is:

- case is not normalized;
- whitespace is not trimmed;
- no Unicode normalization is applied;
- the tag must be non-empty valid UTF-8.

`vault-v1` and `Vault-v1` are different tags, therefore different fleets.

## What SFR does with CAP-85

```text
SFR does not implement CAP-85.
SFR indexes and interprets CAP-85 state.
```

CAP-85 is implemented by the Stellar protocol and its host functions
(`create_executable_tag`, `create_external_ref_contract`, `update_current_contract_executable_ref`).
SFR reads the resulting ledger state: it detects external references in ledger changes, decodes
`owner` and `tag`, resolves the owner entry to a WASM hash, and reports what it finds.

## What this means for trust

Because the owner contract can update the tag entry at any time, the owner effectively controls
which code every member executes. That is the intended design of CAP-85. It is also exactly why
an independent observer that records membership, releases, and per-member consistency is useful.

Next: [Quick start](./quick-start.md) or [Fleet concept](../concepts/fleet.md)
