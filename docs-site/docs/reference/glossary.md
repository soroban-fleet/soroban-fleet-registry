# Glossary

Concise definitions of the terms used on this site.

| Term | Definition |
|---|---|
| **CAP-85** | Stellar CAP introduced in Protocol 28 and supported on Protocol 29, "Externally managed contract executables". Adds the `CONTRACT_EXECUTABLE_EXTERNAL_REF` executable variant so many contracts can share one updatable WASM reference. |
| **Soroban** | The smart contract platform on the Stellar network. |
| **ExternalRef** | `CONTRACT_EXECUTABLE_EXTERNAL_REF` — a contract executable that names an owner contract and a tag instead of storing a WASM hash directly. |
| **Executable reference** | The `(executable_owner, tag)` pair stored in a contract's executable; the data SFR decodes. |
| **Fleet** | The set of contract instances sharing one executable reference. Identity: `(owner_address, tag)`. |
| **Owner** | The contract that stores the canonical WASM hash under a tag in its persistent storage. |
| **Tag** | Exact, case-sensitive string key for the owner's entry (for example `vault-v1`). Preserved byte-for-byte; never trimmed or normalized. |
| **WASM** | WebAssembly — the compiled Soroban contract code. A WASM hash is the 32-byte hash identifying one build. |
| **WASM hash** | 32-byte hash of a WASM build; what an executable ultimately resolves to. |
| **Ledger** | A Stellar ledger — one closed, ordered unit of network state. Ledgers are numbered by sequence. |
| **LedgerCloseMeta** | XDR structure describing a closed ledger's changes; the indexer's input. |
| **Checkpoint** | The last ledger sequence the indexer has committed, stored in `indexer_checkpoints`. Progress marker and recovery point. |
| **Indexer** | The `sfr ingest` process that reads ledgers and writes fleets, members, releases, and checkpoints. |
| **Indexing coverage** | How far the index has advanced relative to the network (`indexed_through` vs. `latest_ledger`). Distinct from service health and from fleet verification. |
| **Drift** | Verification status `DRIFT`: at least one active member resolves to a WASM hash different from the expected one. |
| **Release** | An observed change of the owner's hash under a tag, recorded with ledger, transaction, and old/new hashes. |
| **Member** | A contract instance whose executable references a fleet's `(owner, tag)`. Active while the reference holds. |
| **Verification** | A recorded comparison of every active member's live executable against an expected hash, classified as `HEALTHY`, `DRIFT`, `BROKEN_REFERENCE`, `INCOMPLETE`, or `UNKNOWN`. |
| **Broken reference** | A reference whose owner entry cannot be resolved (missing, invalid, or malformed). |
| **Stellar RPC** | The JSON-RPC endpoint SFR reads ledgers, ledger entries, and the latest ledger from. |
| **StrKey** | Stellar's text encoding for addresses; contract addresses start with `C`. |
| **Read-only observer** | This project's design boundary: it reads and reports state, and never signs, submits, or mutates anything on-chain. |

Related: [Statuses](./statuses.md) · [Verification states](../concepts/verification-states.md) ·
[Why CAP-85](../introduction/why-cap85.md)
