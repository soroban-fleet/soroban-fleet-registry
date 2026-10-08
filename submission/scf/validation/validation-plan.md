# Ecosystem Validation Plan

## 1. Objectives

The purpose of this validation plan is to establish whether the operational challenge addressed by Soroban Fleet Registry—namely, verifying consistency, tracking membership, and monitoring releases across contract instances referencing a shared CAP-85 executable—is a concrete problem experienced by active protocol developers in the Stellar and Soroban ecosystem.

This plan focuses on **technical workflow verification**, not speculative market sizing or endorsements.

---

## 2. Target Developer & Maintainer Personas

### Persona A: Factory / Multi-Instance Protocol Maintainer
- **Profile**: Engineers maintaining protocols that deploy numerous identical smart contract instances (e.g., decentralized liquidity vaults, smart account wallets, isolated lending markets, tokenized yield accounts).
- **Core Pain Point**: Lacks an automated index to know which instances exist, which tag they point to, and whether an owner contract upgrade propagated to every active instance without code execution drift.

### Persona B: Protocol Operations & Release Engineer
- **Profile**: Maintainers responsible for executing protocol upgrades, migrations, or contract parameter updates across multiple environments (Testnet, Mainnet).
- **Core Pain Point**: Upgrading an owner contract tag is atomic on-chain, but verifying that every dependent instance resolves correctly without RPC errors or malformed references currently requires custom ad-hoc scripts.

### Persona C: Ecosystem Security Auditor & Reviewer
- **Profile**: Security researchers and protocol reviewers verifying the integrity, code provenance, and operational risk of a deployed Soroban protocol.
- **Core Pain Point**: Discerning whether factory instances genuinely share a verified WASM bytecode or if specific contracts have been modified or left stranded on legacy implementations.

---

## 3. Concrete Operational Problems to Validate

1. **Member Discovery Blindspot**: The Stellar ledger stores contract executables per instance, but provides no native index to list all instances referencing `(owner, tag)`. Maintainers must either scan historical blocks manually or maintain a private database.
2. **Post-Upgrade Verification Latency**: When an owner updates an executable tag, verifying that 100% of member instances continue resolving without broken references is manual and error-prone.
3. **Incomplete Coverage Risk**: Traditional ad-hoc scripts check a subset of known addresses, creating false confidence when recent deployments or network reorgs have not been indexed.

---

## 4. Objective Validation Questions

These questions are structured to avoid leading respondents toward a pre-determined conclusion:

1. *How do you currently track the full list of contract instances deployed by your factories or referencing your shared code?*
2. *When you update a shared contract implementation or executable reference, what tooling or workflow do you use to verify that every deployed contract resolves to the expected WASM?*
3. *Have you encountered situations where contract instances were deployed with an outdated or drifted executable reference without your team noticing?*
4. *What specific information (e.g., active WASM hash, release history, drift flags, coverage indicators) would be necessary for your team to trust an external verification tool in CI/CD?*
5. *Under what operational conditions would an open-source, read-only indexing API be preferable to running your own in-house event indexing scripts?*
