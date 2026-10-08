# API: Fleets

## GET /v1/fleets

List indexed fleets.

| | |
|---|---|
| **Method** | `GET` |
| **Path** | `/v1/fleets` |
| **Purpose** | Paginated list of all CAP-85 fleets in the index |

### Parameters

| Name | In | Type | Default | Notes |
|---|---|---|---:|---|
| `limit` | query | integer | 20 | Max 100 |
| `offset` | query | integer | 0 | — |

### Example request

```bash
curl -s "http://localhost:8080/v1/fleets?limit=2&offset=0"
```

### Example response

```json
{
  "data": [
    {
      "id": { "Owner": "CA3D5KRYMCMCZVACQY3CGINPVPUM2OAL6YHT3ZVM5C2MUMAA252AAAAB", "Tag": "vault-v1" },
      "current_wasm_hash": "kaphi3Rmj02WD8pd/Dfwd10Xl+j/TTTwvybvvZkjjoM=",
      "member_count": 12,
      "first_seen_ledger": 100000,
      "last_seen_ledger": 105432,
      "last_indexed_ledger": 105432,
      "created_at": "2026-10-07T12:00:00Z",
      "updated_at": "2026-10-07T12:05:00Z"
    }
  ],
  "meta": { "total": 1, "limit": 2, "offset": 0 }
}
```

Note the capitalized `Owner`/`Tag` keys inside `id` — that is the actual wire format.

### Errors

| Status | Code |
|---|---|
| 500 | `INTERNAL_ERROR` |

### Pagination

`data` holds one page; `meta.total` is the full count. `current_wasm_hash` is `null` when the
owner entry has not been resolved yet.

---

## GET `/v1/fleets/{owner}/{tag}`

Fetch one fleet by identity.

| | |
|---|---|
| **Method** | `GET` |
| **Path** | `/v1/fleets/{owner}/{tag}` |
| **Purpose** | Fleet details and current WASM hash |

### Parameters

| Name | In | Type | Required | Notes |
|---|---|---|---|---|
| `owner` | path | string | yes | StrKey contract address |
| `tag` | path | string | yes | Exact, case-sensitive |

### Example request

```bash
curl -s "http://localhost:8080/v1/fleets/CA3D5KRYMCMCZVACQY3CGINPVPUM2OAL6YHT3ZVM5C2MUMAA252AAAAB/vault-v1"
```

### Example response

```json
{
  "data": {
    "id": { "Owner": "CA3D5KRYMCMCZVACQY3CGINPVPUM2OAL6YHT3ZVM5C2MUMAA252AAAAB", "Tag": "vault-v1" },
    "current_wasm_hash": "kaphi3Rmj02WD8pd/Dfwd10Xl+j/TTTwvybvvZkjjoM=",
    "member_count": 12,
    "first_seen_ledger": 100000,
    "last_seen_ledger": 105432,
    "last_indexed_ledger": 105432,
    "created_at": "2026-10-07T12:00:00Z",
    "updated_at": "2026-10-07T12:05:00Z"
  }
}
```

### Errors

| Status | Code | Cause |
|---|---|---|
| 400 | `INVALID_ARGUMENT` | Invalid owner address or tag |
| 404 | `FLEET_NOT_FOUND` | Not indexed |
| 500 | `INTERNAL_ERROR` | Database failure |

---

## Schema: Fleet

| Field | Type | Notes |
|---|---|---|
| `id` | object `{Owner, Tag}` | Fleet identity |
| `current_wasm_hash` | byte (base64) | May be `null` |
| `member_count` | integer | Active members only |
| `first_seen_ledger` | integer | First observation |
| `last_seen_ledger` | integer | Last observed change |
| `last_indexed_ledger` | integer | Last ingestion write |
| `created_at`, `updated_at` | date-time | UTC |

See [Members](./members.md) · [History](./history.md) · [Verification](./verification.md)
