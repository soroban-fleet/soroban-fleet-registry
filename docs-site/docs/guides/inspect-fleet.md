# Guide: Inspect a fleet

Look up one fleet by owner address and tag.

## Command

```bash
sfr fleet inspect \
  --owner <OWNER> \
  --tag <TAG>
```

`<OWNER>` is the StrKey address of the contract that stores the executable (for example
`CA3D5KRY…AAAAB`). `<TAG>` is the exact tag string (`vault-v1`). Both flags are required.

The command reads from PostgreSQL (`SFR_DATABASE_URL`); it does not contact Stellar RPC.

## Example output

```text
Fleet: <OWNER>:vault-v1
Current WASM Hash: 91aa618b74668f4d960fca5dfc37f0775d1797e8ff4d34f0bf26efbd99238e83
Member Count:      12
First Seen Ledger: 100000
Last Seen Ledger:  105432
Last Indexed:      105432
```

## Field reference

| Field | Meaning |
|---|---|
| `Fleet` | Fleet identity as `owner:tag` |
| `Current WASM Hash` | 32-byte hash the owner currently publishes under the tag (hex) |
| `Member Count` | Number of **active** members |
| `First Seen Ledger` | Ledger where this fleet was first observed |
| `Last Seen Ledger` | Ledger where fleet state was last observed changing |
| `Last Indexed` | Last ledger the indexer wrote for this fleet |

Interpret `Last Indexed` against the network ledger: if it lags, conclusions drawn from this data
are provisional — see [Indexing coverage](../concepts/indexing-coverage.md).

## JSON output

```bash
sfr fleet inspect --owner <OWNER> --tag <TAG> --json
```

```json
{
  "id": { "Owner": "<OWNER>", "Tag": "vault-v1" },
  "current_wasm_hash": "<base64, 32 bytes>",
  "member_count": 12,
  "first_seen_ledger": 100000,
  "last_seen_ledger": 105432,
  "last_indexed_ledger": 105432,
  "created_at": "2026-10-07T12:00:00Z",
  "updated_at": "2026-10-07T12:05:00Z"
}
```

`current_wasm_hash` is base64 in JSON (Go marshals `[]byte` that way). The hexadecimal form in
human output is the same 32 bytes.

## Exit codes

| Code | Meaning |
|---:|---|
| 0 | Fleet found and printed |
| 4 | Missing `--owner`/`--tag`, or fleet not found |
| 5 | Database error |

## Next steps

- List who belongs to it: `sfr fleet members --owner <OWNER> --tag <TAG>`
  ([members output](#related-commands)).
- Check consistency: [Verify a fleet](./verify-fleet.md).
- See it in the UI: `/fleets/<owner>/<tag>` in the [web app](./use-web-app.md).

## Related commands

```bash
# All fleets (paginated)
sfr fleet list --limit 20 --offset 0

# Members, active only by default
sfr fleet members --owner <OWNER> --tag <TAG>
sfr fleet members --owner <OWNER> --tag <TAG> --active-only=false

# Observed executable changes
sfr fleet releases --owner <OWNER> --tag <TAG>

# Past verification results
sfr fleet history --owner <OWNER> --tag <TAG>
```

Full flag list: [CLI reference](../reference/cli.md).
