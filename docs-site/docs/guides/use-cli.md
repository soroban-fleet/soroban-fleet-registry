# Guide: Use the CLI

The `sfr` binary is the command-line interface to the indexed data. It reads PostgreSQL directly
and, for live checks, Stellar RPC.

## Build

```bash
make build
./bin/sfr --help
```

## Command map

```text
sfr
├── fleet
│   ├── list                 List all fleets
│   ├── inspect              Fleet details
│   ├── members              Member list
│   ├── history              Verification history
│   ├── releases             Release history
│   └── verify               Live integrity check (exit code = status)
├── contract
│   └── inspect              Contract executable and membership
├── migrate [up|down]        Database schema migrations
├── api                      Start the REST API server
└── ingest                   Run the historical ingestion pipeline
```

## Flags

| Flag | Applies to | Meaning |
|---|---|---|
| `--owner <ADDR>` | `fleet` subcommands except `list` | Owner contract address (required) |
| `--tag <TAG>` | same | Exact fleet tag (required) |
| `--contract <ADDR>` | `contract inspect` | Contract address (required) |
| `--expected-wasm <HEX>` | `fleet verify` | 64-char hex hash; default is the fleet's current hash |
| `--json` | all read commands | JSON output instead of text |
| `--limit <N>` | `list`, `members`, `history`, `releases` | Page size, default 20 |
| `--offset <N>` | same | Page offset, default 0 |
| `--active-only=<BOOL>` | `fleet members` | Default `true`; set `false` to include former members |

No other flags exist in v1.0.0. Unknown commands or missing required flags exit with code `4`.

## Common tasks

```bash
# What fleets are indexed?
sfr fleet list

# Details for one fleet
sfr fleet inspect --owner <OWNER> --tag <TAG>

# Who is in it (active members only)
sfr fleet members --owner <OWNER> --tag <TAG>

# What changed over time
sfr fleet releases --owner <OWNER> --tag <TAG>

# What has been verified before
sfr fleet history --owner <OWNER> --tag <TAG>

# Verify now — exit code is the verdict
sfr fleet verify --owner <OWNER> --tag <TAG>
echo $?

# What does this contract run?
sfr contract inspect --contract <CONTRACT_ID>

# Machine-readable output for scripts
sfr fleet verify --owner <OWNER> --tag <TAG> --json
```

## Exit codes

| Code | Meaning |
|---:|---|
| 0 | Success / `HEALTHY` |
| 1 | `DRIFT` |
| 2 | `BROKEN_REFERENCE` |
| 3 | `INCOMPLETE` |
| 4 | Invalid input (bad flags, bad hex, not found) |
| 5 | Internal error (database, RPC) / `UNKNOWN` |

See [Status reference](../reference/statuses.md).

## Environment

The CLI reads the same environment as the servers:

| Variable | Needed for |
|---|---|
| `SFR_DATABASE_URL` | every command (connects to PostgreSQL) |
| `SFR_RPC_URL` | `verify`, `contract inspect`, and RPC-backed resolution |

```bash
export SFR_DATABASE_URL="postgres://sfr:sfrpassword@localhost:5432/sfr?sslmode=disable"
export SFR_RPC_URL="https://soroban-testnet.stellar.org"
```

If `SFR_DATABASE_URL` is empty, commands fail immediately with:

```text
SFR_DATABASE_URL is not configured
```

## Server modes

`sfr api` and `sfr ingest` are long-running processes, not query commands:

```bash
sfr api      # REST API on SFR_HTTP_ADDR (default :8080)
sfr ingest   # ledger ingestion loop until interrupted
```

Both validate their configuration first and exit `4` if required variables are missing.

## JSON conventions

- Collection results: `{"data": [...], "total": N}` (CLI) or `{"data": [...], "meta": {...}}` (API).
- Byte fields (`wasm_hash`) are base64 in JSON and hexadecimal in text output.
- Timestamps are RFC 3339 UTC.

Next: [Verify a fleet](./verify-fleet.md) · [CLI reference](../reference/cli.md)
