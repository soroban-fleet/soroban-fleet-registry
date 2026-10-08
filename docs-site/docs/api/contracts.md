# API: Contracts

## GET `/v1/contracts/{contract_id}`

| | |
|---|---|
| **Method** | `GET` |
| **Path** | `/v1/contracts/{contract_id}` |
| **Purpose** | Indexed fleet membership plus live resolved executable for one contract |

### Parameters

| Name | In | Type | Required | Notes |
|---|---|---|---|---|
| `contract_id` | path | string | yes | Valid StrKey contract address |

### Example request

```bash
curl -s "http://localhost:8080/v1/contracts/CB222222222222222222222222222222222222222222222222222222"
```

### Example response — fleet member

```json
{
  "data": {
    "contract_id": "CB222222222222222222222222222222222222222222222222222222",
    "is_member": true,
    "member": {
      "contract_id": "CB222222222222222222222222222222222222222222222222222222",
      "fleet_id": { "Owner": "CA3D5KRYMCMCZVACQY3CGINPVPUM2OAL6YHT3ZVM5C2MUMAA252AAAAB", "Tag": "vault-v1" },
      "wasm_hash": "kaphi3Rmj02WD8pd/Dfwd10Xl+j/TTTwvybvvZkjjoM=",
      "first_seen_ledger": 100000,
      "last_seen_ledger": 105432,
      "active": true
    },
    "resolved": {
      "kind": "EXTERNAL_REF",
      "fleet": { "Owner": "CA3D5KRYMCMCZVACQY3CGINPVPUM2OAL6YHT3ZVM5C2MUMAA252AAAAB", "Tag": "vault-v1" },
      "wasm_hash": "kaphi3Rmj02WD8pd/Dfwd10Xl+j/TTTwvybvvZkjjoM="
    }
  }
}
```

### Example response — not a member

```json
{
  "data": {
    "contract_id": "CB222222222222222222222222222222222222222222222222222222",
    "is_member": false,
    "resolved": {
      "kind": "WASM",
      "wasm_hash": "kaphi3Rmj02WD8pd/Dfwd10Xl+j/TTTwvybvvZkjjoM="
    }
  }
}
```

### Fields

| Field | Source | Meaning |
|---|---|---|
| `is_member` | database | `true` only when an **active** membership row exists |
| `member` | database | Full membership record; present if a row exists (including inactive) |
| `resolved.kind` | Stellar RPC | `EXTERNAL_REF`, `WASM`, or `STELLAR_ASSET` |
| `resolved.fleet` | Stellar RPC | Owner/tag decoded from the live reference (`EXTERNAL_REF` only) |
| `resolved.wasm_hash` | Stellar RPC | Hash the contract resolves to right now |

Comparing `member.wasm_hash` (indexed) with `resolved.wasm_hash` (live) shows local divergence for
a single contract.

### Not every contract is a fleet member

`is_member: false` with `resolved.kind: "WASM"` is a normal, successful response — the contract
simply does not use an external reference. Empty `resolved` (no `kind`) means live resolution did
not happen (missing `SFR_RPC_URL`, RPC error, or contract not found on-chain); treat it as
unresolved, not as absent.

### Errors

| Status | Code | Cause |
|---|---|---|
| 400 | `INVALID_ARGUMENT` | `contract_id` is not a valid StrKey contract address |
| 500 | `INTERNAL_ERROR` | Database failure |

### Notes

- No pagination — one object per contract.
- This endpoint does not write verification history.

Related: [Guide: Inspect a contract](../guides/inspect-contract.md) ·
[Members](./members.md)
