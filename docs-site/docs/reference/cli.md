# CLI reference

`sfr` — the command-line interface and server launcher. Source of truth: `cmd/sfr/main.go`.

## Usage

```text
sfr <command> [subcommand] [flags]
```

```bash
sfr --help      # or: sfr help, sfr -h
```

## Commands

### `sfr fleet list`

List indexed fleets.

| Flag | Default | Notes |
|---|---|---|
| `--limit` | 20 | Page size |
| `--offset` | 0 | Page offset |
| `--json` | off | JSON output `{"data":[…],"total":N}` |

```bash
sfr fleet list --limit 50 --offset 0 --json
```

Text output prints `Total Fleets: N` followed by one line per fleet:
`- <owner>:<tag> | Members: n | WASM: <hex> | Ledgers: a..b`.

### `sfr fleet inspect`

Fleet details.

| Flag | Required | Notes |
|---|---|---|
| `--owner` | yes | Owner contract StrKey address |
| `--tag` | yes | Exact tag |
| `--json` | no | Full fleet object |

Prints fleet identity, current WASM hash (hex), member count, first/last seen ledger, and last
indexed ledger.

### `sfr fleet members`

Member list.

| Flag | Default | Notes |
|---|---|---|
| `--owner` | required | — |
| `--tag` | required | — |
| `--active-only` | `true` | `--active-only=false` includes former members |
| `--limit` | 20 | — |
| `--offset` | 0 | — |
| `--json` | off | `{"data":[…],"total":N}` |

### `sfr fleet history`

Past verification results (newest first).

| Flag | Default |
|---|---|
| `--owner` / `--tag` | required |
| `--limit` / `--offset` | 20 / 0 |
| `--json` | off |

### `sfr fleet releases`

Observed executable changes for the fleet.

| Flag | Default |
|---|---|
| `--owner` / `--tag` | required |
| `--limit` / `--offset` | 20 / 0 |
| `--json` | off |

Prints `- Ledger n | Tx <hash> | Old: <hex> -> New: <hex> | <timestamp>` per release.

### `sfr fleet verify`

Live integrity check; **exit code is the verdict**.

| Flag | Default | Notes |
|---|---|---|
| `--owner` | required | — |
| `--tag` | required | — |
| `--expected-wasm` | fleet's current hash | 64 hex chars (32 bytes) |
| `--json` | off | Full `VerificationResult` |

```bash
sfr fleet verify --owner <OWNER> --tag <TAG>
echo $?
sfr fleet verify --owner <OWNER> --tag <TAG> \
  --expected-wasm 91aa618b74668f4d960fca5dfc37f0775d1797e8ff4d34f0bf26efbd99238e83 \
  --json
```

See [statuses](./statuses.md) for the exit-code table.

### `sfr contract inspect`

Contract executable and membership.

| Flag | Required |
|---|---|
| `--contract` | yes |
| `--json` | no |

Prints contract ID, `Is Fleet Member`, fleet identity and member WASM (when a membership row
exists), live executable kind, and live WASM hash.

### `sfr migrate [up|down]`

Apply or roll back database migrations. Requires `SFR_DATABASE_URL`.

```bash
sfr migrate up
sfr migrate down
```

Unknown actions exit `4`.

### `sfr api`

Start the REST API. Requires `SFR_DATABASE_URL`. Listens on `SFR_HTTP_ADDR` (default `:8080`).
Shuts down gracefully on SIGINT/SIGTERM.

### `sfr ingest`

Run the ledger ingestion pipeline. Requires `SFR_DATABASE_URL` and `SFR_RPC_URL`. Runs until
interrupted.

## Exit codes

| Code | Meaning |
|---:|---|
| 0 | Success / `HEALTHY` |
| 1 | `DRIFT` |
| 2 | `BROKEN_REFERENCE` |
| 3 | `INCOMPLETE` |
| 4 | Invalid input (flags, hex, unknown command, not found, config validation) |
| 5 | Internal error (database, RPC, config load) / `UNKNOWN` |

## Global conventions

- `--json` works on every read command; collection results are `{"data":…,"total":…}`.
- Byte fields are base64 in JSON, hexadecimal in text.
- Missing `SFR_DATABASE_URL` fails immediately with `SFR_DATABASE_URL is not configured`.
- Unknown command or missing subcommand prints usage to stderr and exits `4`.

Guides: [Inspect a fleet](../guides/inspect-fleet.md) ·
[Verify a fleet](../guides/verify-fleet.md) ·
[Inspect a contract](../guides/inspect-contract.md) ·
[Use the CLI](../guides/use-cli.md)
