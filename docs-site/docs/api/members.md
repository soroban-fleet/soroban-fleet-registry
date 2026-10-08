# API: Members

## GET `/v1/fleets/{owner}/{tag}/members`

| | |
|---|---|
| **Method** | `GET` |
| **Path** | `/v1/fleets/{owner}/{tag}/members` |
| **Purpose** | List the contract instances that belong to a fleet |

### Parameters

| Name | In | Type | Default | Notes |
|---|---|---|---|---|
| `owner` | path | string | required | StrKey contract address |
| `tag` | path | string | required | Exact tag |
| `active_only` | query | boolean | `true` | `false` or `0` includes former members |
| `limit` | query | integer | 20 | Max 100 |
| `offset` | query | integer | 0 | — |

### Example request

```bash
curl -s "http://localhost:8080/v1/fleets/CA3D5KRYMCMCZVACQY3CGINPVPUM2OAL6YHT3ZVM5C2MUMAA252AAAAB/vault-v1/members?active_only=true&limit=2"
```

### Example response

```json
{
  "data": [
    {
      "contract_id": "CB222222222222222222222222222222222222222222222222222222",
      "fleet_id": { "Owner": "CA3D5KRYMCMCZVACQY3CGINPVPUM2OAL6YHT3ZVM5C2MUMAA252AAAAB", "Tag": "vault-v1" },
      "wasm_hash": "kaphi3Rmj02WD8pd/Dfwd10Xl+j/TTTwvybvvZkjjoM=",
      "first_seen_ledger": 100000,
      "last_seen_ledger": 105432,
      "last_verified_ledger": 105400,
      "active": true
    }
  ],
  "meta": { "total": 12, "limit": 2, "offset": 0 }
}
```

### Response fields

| Field | Type | Meaning |
|---|---|---|
| `contract_id` | string | Member contract StrKey address (primary key) |
| `fleet_id` | object | Fleet identity |
| `wasm_hash` | byte (base64) | Hash recorded at indexing time |
| `first_seen_ledger` | integer | Ledger where membership began |
| `last_seen_ledger` | integer | Ledger where membership was last observed |
| `last_verified_ledger` | integer or omitted | Last ledger at which verification saw this member |
| `active` | boolean | `false` for members that left the fleet |

### Notes

- `total` in `meta` respects the `active_only` filter.
- Former members are never deleted; request `active_only=false` to include them.
- `wasm_hash` reflects indexed state. Compare it with live state via
  [GET `/v1/contracts/{contract_id}`](./contracts.md).

### Errors

| Status | Code |
|---|---|
| 400 | `INVALID_ARGUMENT` (bad owner or tag) |
| 500 | `INTERNAL_ERROR` |

Pagination follows the [overview](./overview.md#pagination).
