# API: History

Two histories exist: **verification history** (audit results) and **release history** (observed
executable changes). Both are paginated collections.

---

## GET `/v1/fleets/{owner}/{tag}/history`

| | |
|---|---|
| **Method** | `GET` |
| **Path** | `/v1/fleets/{owner}/{tag}/history` |
| **Purpose** | Past verification results, newest first |

### Parameters

| Name | In | Type | Default |
|---|---|---|---|
| `owner` | path | string | required |
| `tag` | path | string | required |
| `limit` | query | integer | 20 (max 100) |
| `offset` | query | integer | 0 |

### Example request

```bash
curl -s "http://localhost:8080/v1/fleets/CA3D5KRYMCMCZVACQY3CGINPVPUM2OAL6YHT3ZVM5C2MUMAA252AAAAB/vault-v1/history?limit=2"
```

### Example response

```json
{
  "data": [
    {
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
    },
    {
      "id": 16,
      "fleet_id": { "Owner": "CA3D5KRYMCMCZVACQY3CGINPVPUM2OAL6YHT3ZVM5C2MUMAA252AAAAB", "Tag": "vault-v1" },
      "expected_wasm_hash": "kaphi3Rmj02WD8pd/Dfwd10Xl+j/TTTwvybvvZkjjoM=",
      "total_members": 12,
      "matching_members": 11,
      "mismatching_members": 1,
      "missing_members": 0,
      "indexed_through": 105431,
      "status": "DRIFT",
      "verified_at": "2026-10-07T11:00:00Z"
    }
  ],
  "meta": { "total": 17, "limit": 2, "offset": 0 }
}
```

Ordering: `verified_at DESC, id DESC`. Every row is an append-only audit record.

### Errors

| Status | Code |
|---|---|
| 500 | `INTERNAL_ERROR` |

---

## GET `/v1/fleets/{owner}/{tag}/releases`

| | |
|---|---|
| **Method** | `GET` |
| **Path** | `/v1/fleets/{owner}/{tag}/releases` |
| **Purpose** | Observed changes of the owner's WASM hash under this tag |

### Parameters

Same as history: `owner`, `tag` (path), `limit`, `offset` (query).

### Example request

```bash
curl -s "http://localhost:8080/v1/fleets/CA3D5KRYMCMCZVACQY3CGINPVPUM2OAL6YHT3ZVM5C2MUMAA252AAAAB/vault-v1/releases"
```

### Example response

```json
{
  "data": [
    {
      "id": 2,
      "fleet_id": { "Owner": "CA3D5KRYMCMCZVACQY3CGINPVPUM2OAL6YHT3ZVM5C2MUMAA252AAAAB", "Tag": "vault-v1" },
      "old_wasm_hash": "kaphi3Rmj02WD8pd/Dfwd10Xl+j/TTTwvybvvZkjjoM=",
      "new_wasm_hash": "kqjH6NqYh3OxVrEcTjP4vBQ1mA8Zn0sLwYdKi2pUvC0=",
      "ledger": 102000,
      "tx_hash": "3b940e53a2ef9f8...",
      "observed_at": "2026-10-06T09:30:00Z"
    }
  ],
  "meta": { "total": 2, "limit": 20, "offset": 0 }
}
```

| Field | Meaning |
|---|---|
| `old_wasm_hash` | Hash before the change; `null`/omitted for the first observed state |
| `new_wasm_hash` | Hash after the change |
| `ledger` | Ledger where the change was observed |
| `tx_hash` | Transaction that carried the change |
| `observed_at` | Ledger close time (UTC) |

### Errors

| Status | Code |
|---|---|
| 500 | `INTERNAL_ERROR` |

---

## Observed vs. verified

A release row proves the owner changed its hash. It does not prove members match it. For
consistency, call [verification](./verification.md).

Next: [Verification](./verification.md) · [Contracts](./contracts.md)
