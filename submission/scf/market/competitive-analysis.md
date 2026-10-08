# Competitive & Ecosystem Tooling Analysis

This analysis examines existing tooling across the Stellar and Soroban ecosystem to objectively evaluate market positioning and capability boundaries.

---

## Tooling Landscape Overview

| Tool / Platform | Category | Primary Target User | Primary Focus | Overlap Level with SFR | Relationship to SFR |
|---|---|---|---|---|---|
| **Stellar Horizon** | Network API / Explorer Backend | Wallet & App Developers | Accounts, operations, payments, order books | **ADJACENT** | Complementary; Horizon does not index Soroban contract storage or executable relationships. |
| **Soroban RPC** | Protocol RPC Node | Smart Contract Developers | Direct transaction simulation, raw ledger entry lookups | **COMPLEMENTARY** | Primary upstream source; SFR queries Soroban RPC to ingest ledgers and fetch on-chain entries. |
| **Stellar Expert / Stellar.Live** | Public Blockchain Explorer | General Users, Traders | Historical transactions, asset distribution, ledger explorer | **ADJACENT** | General explorer; lacks fleet identity grouping, CAP-85 tag resolution, or drift verification. |
| **Mercury / Zephyr** | Custom Event Indexers | DApp Developers | Custom smart contract event streaming into Postgres | **ADJACENT** | Custom event ingest; requires protocol developers to write custom indexing logic; does not model CAP-85 executables. |
| **Stellar CLI (`stellar`)** | Official Developer CLI | Smart Contract Developers | Contract deployment, WASM optimization, key management | **COMPLEMENTARY** | CLI deployment tool; does not provide fleet-level multi-instance tracking or historical release auditing. |
| **Soroban Fleet Registry (SFR)** | **CAP-85 Fleet Indexer & Verifier** | Protocol Maintainers, Security Reviewers | Fleet discovery, WASM resolution, drift verification, release provenance | **N/A (Subject)** | Dedicated relationship-level indexing for CAP-85 externally managed contract executables. |

---

## Overlap Classification Definitions

- **DIRECT**: Identical scope, targeting the same users with the same capability. *(None identified).*
- **ADJACENT**: Operates in the same ecosystem (indexing/explorers) but focuses on different ledger primitives (e.g., token balances vs executable pointers).
- **COMPLEMENTARY**: Upstream infrastructure consumed by SFR (e.g., Soroban RPC, Stellar CLI).
- **NOT COMPETITIVE**: Completely distinct domain (wallets, DEX frontends, AMMs).
