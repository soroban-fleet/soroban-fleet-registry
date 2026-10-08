# What is Soroban Fleet Registry?

Soroban Fleet Registry (SFR) is an open-source indexing and verification service for Soroban
contracts that use CAP-85 externally managed executables.

It observes the Stellar network, records which contract instances point at the same externally
managed executable, resolves which WASM is currently active, and reports whether the members of a
group still match that executable.

```text
contract instances
       ↓
external executable reference
       ↓
(owner, tag)
       ↓
fleet
       ↓
WASM resolution
       ↓
verification
```

## What it does

- **Indexes**: reads ledger changes and records contract instances that declare an external
  executable reference.
- **Groups**: collects those instances into a *fleet*, identified by the pair `(owner, tag)`.
- **Resolves**: reads the owner contract's entry keyed by the tag to find the active WASM hash.
- **Verifies**: compares every active member's live on-chain executable against the expected hash
  and classifies the result.
- **Serves**: exposes the results through a read-only REST API, a CLI (`sfr`), and a web interface.

## What it does not do

SFR observes state. It does not control the fleet.

> Soroban Fleet Registry V1 is a read-only indexing and verification service. It does not deploy,
> own, upgrade, or execute a custom Soroban contract.

Concretely, V1 does not:

- sign transactions or hold private keys;
- connect to wallets;
- change any on-chain state, including executable upgrades;
- run a custom Soroban contract of its own;
- issue a token, collect fees, or implement staking;
- act as a general-purpose block explorer.

The service never becomes a party to the fleet it observes. Anyone can independently check its
claims against the Stellar network.

## Who uses it

| Audience | Uses |
|---|---|
| Fleet maintainers | member lists, release history, drift checks |
| Ecosystem reviewers | plain-language introduction, CAP-85 explanation, verification results |
| Developers | REST API, CLI, local development workflow |
| Operators | indexing status, health checks, recovery procedures |

## Where to go next

- Reviewers: [The problem](./problem.md), [Why CAP-85?](./why-cap85.md),
  [Fleet](../concepts/fleet.md), [Verification states](../concepts/verification-states.md).
- Developers: [Quick start](./quick-start.md), [CLI reference](../reference/cli.md),
  [API overview](../api/overview.md).
- Operators: [System architecture](../architecture/system-architecture.md),
  [Operations](../operations/indexing.md).
