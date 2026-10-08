# API: Verification

## GET `/v1/fleets/{owner}/{tag}/verify`

| | |
|---|---|
| **Method** | `GET` |
| **Path** | `/v1/fleets/{owner}/{tag}/verify` |
| **Purpose** | Run verification against live state and return the result |

Despite using `GET`, this endpoint performs work: it resolves every active member over Stellar
RPC, classifies the fleet, and inserts a row into `fleet_verifications`. It does not modify
on-chain state or fleet data.

### Parameters

| Name | In | Type | Default | Notes |
|---|---|---|---|---|
| `owner` | path | string | required | StrKey contract address |
| `tag` | path | string | required | Exact tag |
| `expected_wasm` | query | string | fleet's `current_wasm_hash` | Hex, 64 characters (32 bytes) |

### Example request

```bash
curl -s "http://localhost:8080/v1/fleets/CA3D5KRYMCMCZVACQY3CGINPVPUM2OAL6YHT3ZVM5C2MUMAA252AAAAB/vault-v1/verify"
```

With an explicit expectation:

```bash
curl -s "http://localhost:8080/v1/fleets/CA3D5KRYMCMCZVACQY3CGINPVPUM2OAL6YHT3ZVM5C2MUMAA252AAAAB/vault-v1/verify?expected_wasm=91aa618b74668f4d960fca5dfc37f0775d1797e8ff4d34f0bf26efbd99238e83"
```

### Example response

```json
{
  "data": {
    "id": 17,
    "fleet_id": { "Owner": "CA3D5KRYMCMCZVACQY3CGINPVPUM2OAL6YHT3ZVM5C2MUMAA252AAAAB", "Tag": "vault-v1" },
    "expected_wasm_hash": "kaphi3Rmj02WD8pd/Dfwd10Xl+j/TTTwvybvvZkjjoM=",
    "total_members": 12,
    "matching_members": 12,
    "mismatching_members": 0,
    "missing_members": 0,
    "indexed_through": 105432,
    "status": "HEALTHY",
    "verified_at": "2026-10-07T12:00:00Z"
  }
}
```

### Response fields

| Field | Type | Meaning |
|---|---|---|
| `expected_wasm_hash` | byte (base64) | Hash the members were compared against |
| `total_members` | integer | Active members considered |
| `matching_members` | integer | Resolved to the expected hash |
| `mismatching_members` | integer | Resolved to a different hash |
| `missing_members` | integer | Could not be resolved |
| `indexed_through` | integer | Indexer checkpoint at verification time |
| `status` | enum | `HEALTHY` \| `DRIFT` \| `BROKEN_REFERENCE` \| `INCOMPLETE` \| `UNKNOWN` |
| `verified_at` | date-time | UTC |

### Errors

| Status | Code | Cause |
|---|---|---|
| 400 | `INVALID_ARGUMENT` | Invalid owner/tag, malformed `expected_wasm`, or fleet has no indexed hash and no `expected_wasm` was supplied |
| 404 | `FLEET_NOT_FOUND` | Fleet not indexed |
| 500 | `INTERNAL_ERROR` | Database or RPC failure during verification |

### Status vs. HTTP status

The HTTP status is `200` whenever verification completed — including `DRIFT`, `BROKEN_REFERENCE`,
and `INCOMPLETE`. Those are **results**, not transport failures. Consume `data.status`, not the
HTTP code, to judge fleet state.

| `status` | CLI exit code |
|---|---:|
| `HEALTHY` | 0 |
| `DRIFT` | 1 |
| `BROKEN_REFERENCE` | 2 |
| `INCOMPLETE` | 3 |
| `UNKNOWN` | 5 |

Full semantics: [Verification states](../concepts/verification-states.md) ·
[Status reference](../reference/statuses.md)

### Notes

- Each call appends an audit row; repeated calls produce repeated history entries.
- `indexed_through` lets you judge whether the verdict was made with complete coverage.
- There is no pagination here — one result object per request.

Related: [History](./history.md) (read past results) ·
[Guide: Verify a fleet](../guides/verify-fleet.md) (CLI equivalent)
