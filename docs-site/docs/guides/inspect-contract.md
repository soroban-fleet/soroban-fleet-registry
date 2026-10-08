# Guide: Inspect a contract

Check what a single contract is executing and whether it belongs to a fleet.

## Command

```bash
sfr contract inspect --contract <CONTRACT_ID>
```

`<CONTRACT_ID>` must be a valid StrKey contract address. Requires `SFR_DATABASE_URL`; add
`SFR_RPC_URL` for live executable resolution.

## Example output

```text
Contract: <CONTRACT_ID>
Is Fleet Member: true
Fleet: <OWNER>:vault-v1
Member WASM: 91aa618b74668f4d960fca5dfc37f0775d1797e8ff4d34f0bf26efbd99238e83
Live Executable Kind: EXTERNAL_REF
Live WASM Hash:       91aa618b74668f4d960fca5dfc37f0775d1797e8ff4d34f0bf26efbd99238e83
```

## Fields

| Field | Source | Meaning |
|---|---|---|
| `Contract` | input | The contract you asked about |
| `Is Fleet Member` | database | `true` only if the contract has an **active** membership row |
| `Fleet` | database | Fleet identity, present when a membership row exists |
| `Member WASM` | database | Hash recorded at indexing time |
| `Live Executable Kind` | Stellar RPC | `EXTERNAL_REF`, `WASM`, or `STELLAR_ASSET` |
| `Live WASM Hash` | Stellar RPC | Hash the contract resolves to right now |

Comparing `Member WASM` with `Live WASM Hash` shows whether indexed state still matches live
state for this contract.

## Not every contract is a fleet member

Direct-WASM contracts return:

```text
Contract: <CONTRACT_ID>
Is Fleet Member: false
Live Executable Kind: WASM
Live WASM Hash:       <hash>
```

That is a normal result, not an error. Such contracts have no `(owner, tag)` reference, so SFR
does not group them anywhere.

Former members are a third case: if a contract left a fleet, `Is Fleet Member` is `false`, but the
membership row may still be present in the JSON output with `"active": false`. The contract's
current executable kind shows what it points at now.

## JSON output

```bash
sfr contract inspect --contract <CONTRACT_ID> --json
```

```json
{
  "contract_id": "<CONTRACT_ID>",
  "is_member": true,
  "member": {
    "contract_id": "<CONTRACT_ID>",
    "fleet_id": { "Owner": "<OWNER>", "Tag": "vault-v1" },
    "wasm_hash": "<base64, 32 bytes>",
    "first_seen_ledger": 100000,
    "last_seen_ledger": 105432,
    "active": true
  },
  "resolved": {
    "kind": "EXTERNAL_REF",
    "fleet": { "Owner": "<OWNER>", "Tag": "vault-v1" },
    "wasm_hash": "<base64, 32 bytes>"
  }
}
```

## Exit codes

| Code | Meaning |
|---:|---|
| 0 | Inspection returned (including `is_member: false`) |
| 4 | Missing `--contract`, invalid StrKey address, or inspection failure |
| 5 | Database error |

## When the live side is empty

If `SFR_RPC_URL` is unset, or the RPC call fails, `Live Executable Kind` is absent and the
`resolved` object carries no data. The command still prints the indexed membership. Treat an
empty live section as "not resolved", not as "no executable".

## HTTP equivalent

```bash
curl -s "http://localhost:8080/v1/contracts/<CONTRACT_ID>"
```

See [API → Contracts](../api/contracts.md) and the
[web app route](./use-web-app.md#contract-inspector).
