# Indexing coverage

Three different questions get confused with each other in practice. SFR keeps them separate and
answers each with its own signal.

```text
service health      ≠      indexer coverage      ≠      fleet verification
```

| Signal | Question | Source |
|---|---|---|
| **Service health** | Is the API process alive and serving? | `GET /health` |
| **Indexer coverage** | How far has the index advanced relative to the network? | `indexer_checkpoints` vs. latest network ledger |
| **Fleet verification** | Do the members match the expected executable? | `GET /v1/fleets/{owner}/{tag}/verify` |

## The three are independent

A perfectly healthy API says nothing about the index, and a complete index says nothing about
whether members match.

```text
API: healthy
Indexer: incomplete
Fleet verification: INCOMPLETE
```

This combination is normal right after a deployment, a restart, or an RPC outage: the HTTP server
is up, the checkpoint has not yet caught up, and therefore verification refuses to conclude.

The reverse also happens. The index can be fully caught up while verification reports `DRIFT`
because a real mismatch exists. Coverage and correctness are different axes.

## How coverage is measured

- The indexer writes a checkpoint row (`stream = 'main'`) inside the same database transaction as
  the ledger's data writes.
- Verification reads that checkpoint (`indexed_through`) and compares it with the latest ledger
  reported by Stellar RPC (`latest_ledger`).
- If `indexed_through < latest_ledger`, the verdict is `INCOMPLETE` — regardless of how many
  members matched.

```text
indexed_through = 105432
latest_ledger   = 105440
gap             = 8 ledgers
result          → INCOMPLETE (coverage incomplete)
```

## What each signal is good for

**Load balancers and orchestrators** should watch `GET /health`. It answers only liveness.

**Operators** should watch the checkpoint versus the network ledger to know whether the index is
current. See [Operations → Indexing](../operations/indexing.md).

**Fleet maintainers and auditors** should read the verification status together with
`indexed_through`, and only treat `HEALTHY` as an assertion. See
[Verification states](./verification-states.md).

## Rules that follow from this

1. Never infer fleet health from API uptime.
2. Never infer fleet health from a release record.
3. Only `HEALTHY` is an assertion — and only when coverage is complete for that verification.
4. `INCOMPLETE` is an honest "not yet", not a soft failure and not a pass.
