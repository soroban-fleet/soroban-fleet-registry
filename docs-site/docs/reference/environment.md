# Environment reference

Definitive list of environment variables recognized by v1.0.0. Sources: `internal/config/config.go`,
`web/lib/config.ts`, `.env.example`, `web/.env.example`.

No secret values are listed here. Never commit real credentials.

## Backend (`SFR_*`)

Read by the `sfr` binary (API, indexer, CLI, migrations) at process start. Applies to the
**server/CLI process**, not the browser.

| Variable | Purpose | Required? | Scope | Example |
|---|---|---|---|---|
| `SFR_NETWORK` | Network label stored in configuration. No v1.0.0 code branches on it; the effective network is determined by `SFR_RPC_URL`. | No | Server | `testnet` |
| `SFR_RPC_URL` | Stellar Soroban RPC endpoint used for ledger ingestion and live resolution | Required for `ingest`; optional elsewhere (live features degrade without it) | Server | `https://soroban-testnet.stellar.org` |
| `SFR_DATABASE_URL` | PostgreSQL connection string | Required for all data access | Server | `postgres://sfr:sfrpassword@localhost:5432/sfr?sslmode=disable` |
| `SFR_HTTP_ADDR` | HTTP listen address for `sfr api` | No — default `:8080` | Server | `:8080` |
| `SFR_LOG_LEVEL` | Log verbosity: `debug`, `info`, `warn`, `error` (case-insensitive) | No — default `info` | Server | `info` |
| `SFR_INDEXER_START_LEDGER` | Ledger to start from when no checkpoint row exists | No — default `0` | Server | `1000` |
| `SFR_INDEXER_BATCH_SIZE` | Ledgers requested per `getLedgers` call; must be > 0 | No — default `50` | Server | `50` |
| `SFR_CORS_ALLOWED_ORIGINS` | Comma-separated browser origins allowed by CORS. Unset = any origin accepted (development only) | No | Server | `http://localhost:3000` |
| `SFR_METRICS_ADDR` | Loaded into configuration; no consumer in v1.0.0 | No | Server | `:9090` |
| `SFR_READ_ONLY` | Parsed as a boolean into configuration; no enforcement in v1.0.0 | No | Server | `true` |

### Failure messages

```text
SFR_DATABASE_URL is required for api
SFR_DATABASE_URL is required for indexer
SFR_RPC_URL is required for indexer
SFR_DATABASE_URL is not configured
invalid SFR_INDEXER_START_LEDGER "x": ...
SFR_INDEXER_BATCH_SIZE must be greater than 0
```

`sfr api` and `sfr ingest` exit with code `4` on configuration validation errors.

## Frontend (`NEXT_PUBLIC_*`)

Read by the Next.js build. Embedded into browser bundles — **never** put secrets here.

| Variable | Purpose | Required? | Scope | Example |
|---|---|---|---|---|
| `NEXT_PUBLIC_API_BASE_URL` | API base URL for client and SSR requests | No — default `http://localhost:8080` | Browser + server | `https://api.fleet.example.com` |
| `NEXT_PUBLIC_SFR_API_URL` | Alias checked before `NEXT_PUBLIC_API_BASE_URL` | No | Browser + server | `http://localhost:8080` |
| `NEXT_PUBLIC_NETWORK` | Network name displayed in the header badge | No — default `testnet` | Browser | `testnet` |
| `NEXT_PUBLIC_EXPLORER_BASE_URL` | Block explorer base for address/ledger/tx links; empty disables them | No — default `https://stellar.expert/explorer/testnet` | Browser | `https://stellar.expert/explorer/public` |

Precedence for the API base URL:

```text
NEXT_PUBLIC_SFR_API_URL  →  NEXT_PUBLIC_API_BASE_URL  →  http://localhost:8080
```

## Docker Compose variables

Consumed by `docker-compose.yml` (host-side, with defaults):

| Variable | Default | Purpose |
|---|---|---|
| `POSTGRES_USER` | `sfr` | Database user |
| `POSTGRES_PASSWORD` | `sfrpassword` | Database password (change outside local dev) |
| `POSTGRES_DB` | `sfr` | Database name |
| `POSTGRES_PORT` | `5432` | Host port for PostgreSQL |
| `SFR_HTTP_PORT` | `8080` | Host port for the API |
| `SFR_NETWORK`, `SFR_RPC_URL`, `SFR_DATABASE_URL`, `SFR_LOG_LEVEL`, `SFR_INDEXER_START_LEDGER`, `SFR_INDEXER_BATCH_SIZE`, `SFR_CORS_ALLOWED_ORIGINS` | see `.env.example` | Passed into the `sfr` container |

## Rules

1. Backend variables are server-side; frontend variables are public by construction.
2. Set `SFR_CORS_ALLOWED_ORIGINS` to exact origins in production — never `*`.
3. Require TLS (`sslmode=require` or `verify-full`) for production database connections.
4. `.env` files are for local development only and must not be committed.

Related: [Configuration guide](../development/configuration.md) ·
[Deployment](../development/deployment.md)
