# Status reference

Definitive list of verification statuses and CLI exit codes for v1.0.0. Verified against
`internal/verification/status.go` and `cmd/sfr/main.go`.

## Statuses

| Status | Meaning | CLI Exit |
|---|---|---:|
| `HEALTHY` | Verified and complete: every active member resolves to the expected WASM, coverage is complete, member set is non-empty | 0 |
| `DRIFT` | At least one active member resolves to a different WASM hash | 1 |
| `BROKEN_REFERENCE` | The external executable reference cannot be resolved (owner data missing, invalid value, or malformed reference) | 2 |
| `INCOMPLETE` | Indexing coverage insufficient (`indexed_through < latest_ledger`), members missing, or zero active members | 3 |
| `UNKNOWN` | Insufficient information to classify | 5 |

## Non-status exit codes

| Code | Meaning |
|---:|---|
| 4 | Invalid input: missing/unknown command, missing required flags, invalid `--expected-wasm`, fleet or contract not found |
| 5 | Internal error: database failure, RPC failure, configuration load failure — also returned for `UNKNOWN` |

```text
0  success / HEALTHY
1  DRIFT
2  BROKEN_REFERENCE
3  INCOMPLETE
4  invalid input
5  internal error or UNKNOWN
```

## Where each status appears

| Surface | Representation |
|---|---|
| CLI `sfr fleet verify` | `Verification Result: <STATUS>` + exit code |
| CLI `--json` | `"status": "<STATUS>"` |
| REST `GET /v1/fleets/{owner}/{tag}/verify` | `data.status` with HTTP `200` |
| Database | `fleet_verifications.status` |
| Web app | Status badge on the fleet and verify pages |

## Classification rules

Evaluated top-down; first match wins (`internal/verification/verifier.go`):

```text
1. broken reference                      → BROKEN_REFERENCE
2. mismatching_members > 0               → DRIFT
3. missing_members > 0                   → INCOMPLETE
4. indexed_through < latest_ledger       → INCOMPLETE
5. total_members == 0                    → INCOMPLETE
6. all members match + full coverage     → HEALTHY
7. otherwise                             → UNKNOWN
```

## The safety rule

> Never report `HEALTHY` merely because no mismatch was detected.

`HEALTHY` additionally requires:

- all members resolved successfully;
- every resolved hash equals the expected hash;
- `total_members > 0`;
- `indexed_through >= latest_ledger`.

## Reading results correctly

- HTTP `200` from the verify endpoint does **not** mean healthy — read `data.status`.
- A release record does **not** mean verified.
- `INCOMPLETE` is neither a pass nor a failure of the fleet; it is a statement about coverage.

Related: [Verification states](../concepts/verification-states.md) ·
[Indexing coverage](../concepts/indexing-coverage.md) ·
[API → Verification](../api/verification.md)
