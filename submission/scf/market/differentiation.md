# Product Differentiation

## Core Architectural Distinction

Soroban Fleet Registry does not compete with or replace general blockchain indexers or explorers. Instead, it occupies a specialized operational tier above raw ledger data:

```text
Layer 1: Stellar Network & Soroban Runtime
         - LedgerCloseMeta, persistent contract entries, SCVal XDR

Layer 2: Upstream Ingestion & Raw RPC
         - Soroban RPC, Horizon, raw ledger streaming
         [Provides low-level contract instances and bytecode lookups]

Layer 3: Relationship & Semantic Indexing (Soroban Fleet Registry)
         - Decodes CAP-85 CONTRACT_EXECUTABLE_EXTERNAL_REF
         - Groups instances into Fleet Identity: (owner_address, tag)
         - Resolves active WASM through owner contract data
         - Tracks release versioning over time

Layer 4: Verification & Operational Safety
         - Compares live on-chain instance code against expected hash
         - Coverage-aware safety classification (HEALTHY, DRIFT, INCOMPLETE)
         - Semantic CLI exit codes for CI/CD gating
```

---

## What Existing Tooling Provides vs What SFR Provides

| Capability | Generic Explorer / Indexer | Custom In-House Script | Soroban Fleet Registry |
|---|---|---|---|
| Contract Bytecode Lookup | ✅ Yes (raw WASM hash) | ✅ Yes | ✅ Yes (normalized hex format) |
| Multi-Contract Fleet Grouping | ❌ No | ⚠️ Partial (hardcoded list) | ✅ Automatic (via ledger discovery) |
| Exact Tag Case Preservation | ❌ No | ⚠️ Varies | ✅ Guaranteed (RFC-compliant XDR parser) |
| Release History Tracking | ❌ No | ❌ No | ✅ Yes (immutable transition audit log) |
| Drift & Divergence Detection | ❌ No | ⚠️ Manual comparison | ✅ Automated deterministic classification |
| Coverage-Aware Incomplete Status | ❌ No (silent omission) | ❌ No | ✅ Enforced (refuses HEALTHY if coverage lags) |
| CI/CD CLI Exit Codes | ❌ No | ⚠️ Custom glue code | ✅ Built-in POSIX exit codes (0, 1, 2, 3, 4, 5) |
| Contract Upgrade Mutations | ❌ No | ⚠️ Custom admin scripts | ❌ No (strictly read-only observer) |
