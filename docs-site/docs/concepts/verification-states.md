# Verification states

The verification engine classifies every fleet into exactly one of five states.

```text
HEALTHY
DRIFT
BROKEN_REFERENCE
INCOMPLETE
UNKNOWN
```

The definitions below match `internal/verification/verifier.go` in v1.0.0.

## The safety rule

> The verifier must not return `HEALTHY` simply because no mismatch was found.

`HEALTHY` additionally requires that indexing coverage is complete for the verification scope and
that at least one active member exists. An empty member set, a lagging indexer, or an unresolvable
reference all block `HEALTHY`.

## States

### HEALTHY — verified and complete

| | |
|---|---|
| **Meaning** | Every active member resolves on-chain to the expected WASM hash, and the index has reached the network ledger. |
| **Cause** | All members match; `total_members > 0`; `indexed_through >= latest_ledger`; no broken references. |
| **What the maintainer should do** | Nothing. Record the result as evidence of consistency at that ledger. |

Exit code `0`.

### DRIFT — at least one mismatch

| | |
|---|---|
| **Meaning** | At least one active member resolves to a different WASM hash than expected. |
| **Cause** | A member points at another fleet, a direct WASM, or a stale owner entry; or the expected hash passed to the verifier is not the one members resolve to. |
| **What the maintainer should do** | List members (`sfr fleet members`), identify the mismatching contracts, and decide whether to re-align them or investigate the divergence. |

Exit code `1`.

### BROKEN_REFERENCE — the reference does not resolve

| | |
|---|---|
| **Meaning** | The fleet's external reference cannot be resolved. |
| **Cause** | Owner contract missing, owner has no entry for the tag, or the entry is malformed. |
| **What the maintainer should do** | Inspect the owner contract on-chain. Check whether the tag entry exists and holds a 32-byte hash. Fix the reference at the source; SFR cannot repair it. |

Exit code `2`.

### INCOMPLETE — coverage insufficient

| | |
|---|---|
| **Meaning** | There is not enough indexed information to assert anything. |
| **Cause** | Indexer checkpoint behind the latest network ledger; members missing from live resolution (RPC gaps); or the fleet has zero indexed active members. |
| **What the maintainer should do** | Let the indexer catch up, then re-run verification. Do not interpret `INCOMPLETE` as healthy or as drift. |

Exit code `3`.

### UNKNOWN — insufficient information

| | |
|---|---|
| **Meaning** | The result could not be classified into any state above with confidence. |
| **Cause** | Fallback classification when inputs are present but no rule applies cleanly. |
| **What the maintainer should do** | Re-run verification; if it persists, capture the JSON output and open an issue. |

Exit code `5`.

## Classification order

The engine evaluates rules in this order (`internal/verification/verifier.go`):

```text
broken reference            → BROKEN_REFERENCE
mismatching members > 0     → DRIFT
missing members > 0         → INCOMPLETE
indexed_through < latest    → INCOMPLETE
total members == 0          → INCOMPLETE
all match + coverage        → HEALTHY
otherwise                   → UNKNOWN
```

Note the precedence: a broken reference is reported even if some members also mismatch, because
the reference problem invalidates the resolution itself.

## Counts in every result

Every verification result carries the numbers behind the verdict:

```json
{
  "fleet_id": { "Owner": "C...", "Tag": "vault-v1" },
  "expected_wasm_hash": "<base64, 32 bytes>",
  "total_members": 12,
  "matching_members": 11,
  "mismatching_members": 1,
  "missing_members": 0,
  "indexed_through": 105432,
  "status": "DRIFT",
  "verified_at": "2026-10-07T12:00:00Z"
}
```

Related: [Indexing coverage](./indexing-coverage.md) ·
[Status reference](../reference/statuses.md)
