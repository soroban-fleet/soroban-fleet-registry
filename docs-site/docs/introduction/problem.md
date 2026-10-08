# The problem

Factory-style Soroban deployments create many contract instances that share one WASM build —
vaults, market accounts, liquidity pools, or any repeated component. CAP-85 lets those instances
point at a single externally managed executable. That solves the *upgrade* problem. It creates an
*operational visibility* problem.

## What maintainers need to know

A fleet is only manageable if a maintainer can answer six questions:

1. **Who is in the fleet?** Which contract instances currently reference a given `(owner, tag)`?
2. **Which implementation is active?** What WASM hash does the owner publish under that tag right
   now?
3. **What changed, and when?** Every change of the active hash is a release; the change history
   should be auditable.
4. **Did everyone pick it up?** A member can be mid-transition, or pointed elsewhere.
5. **Is my data current?** An indexer that lags behind the network can produce conclusions that
   are simply premature.
6. **Is "no problem found" the same as "verified"?** It is not.

The Stellar ledger stores the state of each contract instance. It does not store an index of which
instances share a reference, a release log for the tag, or a consistency report across the group.
Before an indexer exists, answering these questions means manually enumerating instances and
reading ledger entries one at a time.

## What SFR adds

SFR maintains that missing operational layer:

```text
ledger state
   ↓
fleet membership index        who belongs to (owner, tag)
   ↓
release history               what the active WASM was, and when it changed
   ↓
coverage tracking             how far the index has advanced
   ↓
verification report           do members match the expected WASM right now
```

## The distinction that matters most

Two states look similar from the outside and are not the same:

- **Verified consistent**: every active member resolves on-chain to the expected WASM, and the
  index has caught up with the network.
- **Incomplete**: the index has not reached the required ledger height, so no conclusion can be
  drawn yet.

Reporting the second as the first would be the most dangerous error this service could make. The
verification engine is built to refuse it — see
[Verification states](../concepts/verification-states.md) and
[Indexing coverage](../concepts/indexing-coverage.md).

## What is out of scope

This document does not claim monetary impact, fleet counts, or market statistics. The problem above
is a maintenance and audit problem: without an index, fleet state is expensive to determine and
easy to get wrong.

Next: [Why CAP-85?](./why-cap85.md)
