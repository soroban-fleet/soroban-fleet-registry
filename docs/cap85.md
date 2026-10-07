# CAP-85 Protocol Specification & Mechanics

CAP-85 (Protocol 28, "Adapter") introduces externally managed contract executables to Soroban.

## The Problem Solved by CAP-85

Traditionally, a Soroban smart contract instance holds a direct reference to a WASM bytecode hash:

```text
Contract Instance ---> wasm_hash
```

When upgrading a factory-deployed group ("fleet") of contract instances that share the same code, an administrator had to upgrade every instance individually. For large fleets, individual updates cannot fit into a single transaction due to network resource limits. This created an operational window where instances ran mismatched versions, complicating non-backwards-compatible upgrades and emergency security patches.

## CAP-85 Mechanics

CAP-85 extends the `ContractExecutable` XDR union:

```text
union ContractExecutable switch (ContractExecutableType type) {
case CONTRACT_EXECUTABLE_WASM:
    Hash wasm_hash;
case CONTRACT_EXECUTABLE_STELLAR_ASSET:
    void;
case CONTRACT_EXECUTABLE_EXTERNAL_REF:
    ContractExecutableExternalRef external_ref;
};
```

Where `ContractExecutableExternalRef` is:

```text
struct ContractExecutableExternalRef {
    SCAddress executable_owner;
    SCString tag;
};
```

### Reference Resolution Flow

```text
Contract Instance
      |
      v
ContractExecutable::CONTRACT_EXECUTABLE_EXTERNAL_REF
      |
      +---> executable_owner: SCAddress
      +---> tag:              SCString
      |
      v
Owner Contract Persistent Data Entry
      |
      +---> Key: SCValType::SCV_EXECUTABLE_TAG (tag string)
      +---> Val: SCValType::SCV_BYTES (32-byte WASM hash)
      |
      v
Active Executable WASM Hash
```

## Atomic Fleet Upgrades

When the owner contract updates the `SCV_EXECUTABLE_TAG` entry:
1. The 32-byte WASM hash stored under that tag changes in a single ledger entry update.
2. All contract instances holding an `ExternalRef` pointing to `(owner, tag)` immediately execute the new WASM bytecode on their next invocation.
3. The upgrade applies across all fleet members simultaneously without touching individual instance ledger entries.

## Tag Handling Rules

Per CAP-85:
- An executable tag is an exact string key.
- Case is preserved exactly (no lowercase conversion).
- Whitespace is preserved exactly (no trimming).
- Unicode normalization is not applied.
- The tag must be valid UTF-8 and non-empty.
