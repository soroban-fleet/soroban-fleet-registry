# Guide: Verify a fleet

Run a live consistency check of every active member against an expected WASM hash.

## Command

```bash
sfr fleet verify \
  --owner <OWNER> \
  --tag <TAG>
```

Requires `SFR_DATABASE_URL` (member list, checkpoint) and `SFR_RPC_URL` (live resolution).

## Optional arguments

| Flag | Default | Purpose |
|---|---|---|
| `--expected-wasm <HASH>` | fleet's `current_wasm_hash` | Hex-encoded 32-byte hash to compare against |
| `--json` | off | Machine-readable result |
| `--owner <OWNER>` | required | Owner contract StrKey address |
| `--tag <TAG>` | required | Exact fleet tag |

`--expected-wasm` must be 64 hex characters (32 bytes). If it is omitted, the CLI reads the
fleet's indexed `current_wasm_hash`; if that is missing or not 32 bytes, it refuses to guess:

```text
Could not infer expected WASM from fleet; specify --expected-wasm explicitly
```

That refusal exits with code `4` rather than producing a verdict from unknown inputs.

## What the engine does

1. Loads active members from the database.
2. Reads the indexer checkpoint (`indexed_through`).
3. Reads the latest ledger from Stellar RPC (`latest_ledger`).
4. Resolves every member's current executable live over RPC.
5. Compares each resolved hash with the expected hash.
6. Classifies the result and inserts an audit row into `fleet_verifications`.

## Example output

```text
Verification Result: HEALTHY
Fleet:               <OWNER>:vault-v1
Expected WASM:       91aa618b74668f4d960fca5dfc37f0775d1797e8ff4d34f0bf26efbd99238e83
Total Members:       12
Matching Members:    12
Mismatching Members: 0
Missing Members:     0
Indexed Through:     105432
```

## JSON output

```bash
sfr fleet verify --owner <OWNER> --tag <TAG> --json; echo "exit=$?"
```

```json
{
  "id": 17,
  "fleet_id": { "Owner": "<OWNER>", "Tag": "vault-v1" },
  "expected_wasm_hash": "<base64, 32 bytes>",
  "total_members": 12,
  "matching_members": 12,
  "mismatching_members": 0,
  "missing_members": 0,
  "indexed_through": 105432,
  "status": "HEALTHY",
  "verified_at": "2026-10-07T12:00:00Z"
}
```

## Exit codes

The exit code is the primary machine interface. Use it in scripts and CI.

| Exit | Status | Meaning |
|---:|---|---|
| 0 | `HEALTHY` | All active members match and coverage is complete |
| 1 | `DRIFT` | At least one member resolves to a different hash |
| 2 | `BROKEN_REFERENCE` | Owner/tag reference cannot be resolved |
| 3 | `INCOMPLETE` | Indexing coverage insufficient or members missing |
| 4 | — | Invalid input: missing flags, bad `--expected-wasm`, fleet not found |
| 5 | `UNKNOWN` / error | Internal error (database, RPC) or unclassifiable result |

Full definitions: [Verification states](../concepts/verification-states.md) and
[Status reference](../reference/statuses.md).

## Scripting example

```bash
if sfr fleet verify --owner <OWNER> --tag <TAG> --json > result.json; then
  echo "fleet is verified healthy"
else
  case $? in
    1) echo "drift detected" ;;
    2) echo "broken reference" ;;
    3) echo "indexer behind — retry after catch-up" ;;
    *) echo "invalid input or internal error" ;;
  esac
fi
```

## Interpreting common results

| Result | What it means | Action |
|---|---|---|
| `INCOMPLETE`, `indexed_through` below network | Indexer lag | Wait for catch-up, re-run |
| `DRIFT` with `mismatching_members: 1` | One contract diverged | `sfr fleet members` to locate it |
| `BROKEN_REFERENCE` | Owner entry missing/malformed | Inspect the owner on-chain |
| `HEALTHY` with `total_members: 0` | Not possible — zero members yields `INCOMPLETE` | — |

## HTTP equivalent

```bash
curl -s "http://localhost:8080/v1/fleets/<OWNER>/vault-v1/verify"
```

Same engine, same statuses, returns HTTP 200 with a JSON `data` field. See
[API → Verification](../api/verification.md).
