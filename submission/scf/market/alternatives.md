# Operational Alternatives Analysis

Before adopting a dedicated infrastructure service like Soroban Fleet Registry, protocol maintainers typically consider three primary operational alternatives. This document factually assesses their costs and limitations.

---

## 1. Ad-Hoc Manual Scripts (Python / TypeScript / Bash)

- **Description**: Maintainers run custom scripts that query `stellar-cli` or Soroban RPC against a hardcoded list of contract addresses.
- **Costs**: Low initial development effort; zero dedicated infrastructure.
- **Limitations**:
  - **Static Registry**: Only checks known addresses; new contract instances deployed by third parties or sub-factories are silently missed.
  - **No Historical Context**: Does not log historical upgrade timestamps, release transactions, or previous WASM hashes.
  - **No Coverage Guarantee**: Scripts do not check network ledger height, creating false confidence when network nodes lag.
  - **Maintenance Overhead**: Custom scripts require continuous maintenance across protocol updates.

---

## 2. General-Purpose Event Indexer (e.g., Mercury, Zephyr)

- **Description**: Maintainers configure event listeners to capture contract deployment events emitted by factory contracts.
- **Costs**: Moderate development; requires maintaining custom schema definitions and subscription endpoints.
- **Limitations**:
  - **Relies on Event Emission**: Only discovers instances that successfully emit specific standard events; misses instances configured directly through ledger transactions without custom events.
  - **No Executable Resolution**: Event indexers index logs, not ledger state; resolving the owner's active tag WASM requires a separate query layer.
  - **No Standardized Verification Semantics**: Teams must write their own drift classification and coverage verification logic from scratch.

---

## 3. General Blockchain Explorers (e.g., Stellar Expert)

- **Description**: Maintainers inspect individual contracts manually on public blockchain explorer web pages.
- **Costs**: Zero development cost.
- **Limitations**:
  - **Single-Contract Focus**: Explorers display one contract address at a time; inspecting a fleet of 500 vaults requires 500 manual page visits.
  - **No CI/CD Integration**: Web explorers cannot be integrated into automated deployment pipelines or regression suites.
  - **No Fleet Semantics**: Explorers do not recognize the `(owner, tag)` relationship as a unified entity.
