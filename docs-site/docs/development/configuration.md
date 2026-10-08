# Configuration

All configuration is environment variables. There are no config files, no CLI config flags, and
no secrets stored in the repository.

Templates:

```bash
cp .env.example .env            # backend
cp web/.env.example web/.env.local
```

## How configuration is loaded

- **Backend (`sfr`)**: `internal/config.Load()` reads `SFR_*` variables at process start,
  applies defaults, and validates mode-specific requirements when a command runs
  (`ValidateForAPI`, `ValidateForIndexer`).
- **Frontend (`web/`)**: `NEXT_PUBLIC_*` variables are read by `web/lib/config.ts` at **build
  time** for client bundles and at runtime for server rendering. Changing them requires a rebuild.

## Mode requirements

| Command | Required variables | Validation |
|---|---|---|
| `sfr migrate up/down` | `SFR_DATABASE_URL` | fails with exit `5` if the DB is unreachable |
| `sfr api` | `SFR_DATABASE_URL` | exit `4` if missing; `SFR_RPC_URL` optional but needed for live resolution |
| `sfr ingest` | `SFR_DATABASE_URL`, `SFR_RPC_URL` | exit `4` if either is missing |
| `sfr fleet …`, `sfr contract …` | `SFR_DATABASE_URL` | `SFR_RPC_URL` needed for `verify` and live inspection |
| `web` | `NEXT_PUBLIC_API_BASE_URL` (or `NEXT_PUBLIC_SFR_API_URL`) | defaults to `http://localhost:8080` |

## Backend variables

| Variable | Purpose | Required? | Example |
|---|---|---|---|
| `SFR_NETWORK` | Network label stored in configuration; v1.0.0 has no code path that branches on it — the live network is whatever `SFR_RPC_URL` points at | No | `testnet` |
| `SFR_RPC_URL` | Stellar Soroban RPC endpoint | Yes for `ingest`; strongly recommended otherwise | `https://soroban-testnet.stellar.org` |
| `SFR_DATABASE_URL` | PostgreSQL connection string | Yes for every command that touches data | `postgres://sfr:sfrpassword@localhost:5432/sfr?sslmode=disable` |
| `SFR_HTTP_ADDR` | API listen address | No (default `:8080`) | `:8080` |
| `SFR_LOG_LEVEL` | `debug`, `info`, `warn`, `error` | No (default `info`) | `info` |
| `SFR_INDEXER_START_LEDGER` | First ledger when no checkpoint exists | No (default `0`) | `1000` |
| `SFR_INDEXER_BATCH_SIZE` | Ledgers per `getLedgers` call; must be > 0 | No (default `50`) | `50` |
| `SFR_CORS_ALLOWED_ORIGINS` | Comma-separated allowed browser origins | No (unset = allow any origin) | `http://localhost:3000` |
| `SFR_METRICS_ADDR` | Read into configuration; not consumed by any v1.0.0 code path | No | `:9090` |
| `SFR_READ_ONLY` | Parsed as boolean into configuration; not enforced by any v1.0.0 code path | No | `true` |

Invalid numeric values fail fast:

```text
invalid SFR_INDEXER_START_LEDGER "abc": ...
SFR_INDEXER_BATCH_SIZE must be greater than 0
invalid SFR_READ_ONLY "yes": ...
```

## Frontend variables

| Variable | Purpose | Required? | Example |
|---|---|---|---|
| `NEXT_PUBLIC_API_BASE_URL` | API base URL used by the client and SSR | No (default `http://localhost:8080`) | `https://api.fleet.example.com` |
| `NEXT_PUBLIC_SFR_API_URL` | Alias checked **before** `NEXT_PUBLIC_API_BASE_URL` | No | `http://localhost:8080` |
| `NEXT_PUBLIC_NETWORK` | Network name shown in the header badge | No (default `testnet`) | `testnet` |
| `NEXT_PUBLIC_EXPLORER_BASE_URL` | Base URL for address/ledger/transaction links; empty disables links | No | `https://stellar.expert/explorer/testnet` |

```ts
// web/lib/config.ts
apiBaseUrl: process.env.NEXT_PUBLIC_SFR_API_URL
         || process.env.NEXT_PUBLIC_API_BASE_URL
         || 'http://localhost:8080'
```

## Secrets

- Never put database passwords, RPC API keys, or private keys in `NEXT_PUBLIC_*` — those values
  are embedded in browser assets.
- Never commit `.env` files (`.gitignore` covers logs and build output; keep your `.env` local).
- No private key is required by this system at all.

## Docker Compose

`docker-compose.yml` passes `SFR_*` values into the `sfr` service with defaults suitable for the
bundled PostgreSQL container. Override via a root `.env` file or shell environment:

```bash
SFR_CORS_ALLOWED_ORIGINS=https://fleet.example.com docker compose up -d
```

Full variable list with deployment context: [Reference → Environment](../reference/environment.md).
