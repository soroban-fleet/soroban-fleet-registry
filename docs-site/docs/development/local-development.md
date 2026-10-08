# Local development

The full stack locally: PostgreSQL, migrations, API, indexer, frontend.

## 1. Clone and build

```bash
git clone https://github.com/soroban-fleet/soroban-fleet-registry.git
cd soroban-fleet-registry
make build
./bin/sfr --help
```

## 2. PostgreSQL

```bash
docker compose up -d postgres
```

The container exposes port `5432` with database `sfr`, user `sfr`, password `sfrpassword`
(override with `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB`, `POSTGRES_PORT`).

Check it:

```bash
docker compose ps
```

## 3. Migrations

```bash
export SFR_DATABASE_URL="postgres://sfr:sfrpassword@localhost:5432/sfr?sslmode=disable"
./bin/sfr migrate up
```

Roll back when testing schema changes:

```bash
./bin/sfr migrate down
```

Migrations live in `migrations/` and are applied transactionally.

## 4. API

```bash
export SFR_HTTP_ADDR=":8080"
export SFR_RPC_URL="https://soroban-testnet.stellar.org"
./bin/sfr api
```

Verify:

```bash
curl -s http://localhost:8080/health
```

`sfr api` refuses to start without `SFR_DATABASE_URL` (exit code `4`).

## 5. Indexer

In a second terminal:

```bash
export SFR_DATABASE_URL="postgres://sfr:sfrpassword@localhost:5432/sfr?sslmode=disable"
export SFR_RPC_URL="https://soroban-testnet.stellar.org"
export SFR_INDEXER_START_LEDGER=1000     # first run only; ignored once a checkpoint exists
export SFR_INDEXER_BATCH_SIZE=50
./bin/sfr ingest
```

Watch for `ledger_processed` log lines. Stop with Ctrl-C; restart to confirm it resumes at
`checkpoint + 1`.

Without `SFR_RPC_URL`, ingestion cannot run (exit code `4`). The API can still run without RPC,
but live verification and contract inspection will not resolve.

## 6. Frontend

```bash
cd web
npm install
npm run dev
```

Open `http://localhost:3000`. Default API base is `http://localhost:8080`
(`web/lib/config.ts`). To override:

```bash
echo "NEXT_PUBLIC_API_BASE_URL=http://localhost:8080" > web/.env.local
```

## Suggested working order

```text
docker compose up -d postgres
        ↓
./bin/sfr migrate up
        ↓
./bin/sfr api          (terminal 1)
        ↓
./bin/sfr ingest       (terminal 2, optional for live data)
        ↓
cd web && npm run dev  (terminal 3)
```

## Offline development

Unit and integration tests run against fixtures in `fixtures/` without network access. See
[Fixtures](./fixtures.md). Database-backed tests skip automatically if PostgreSQL is
unreachable.

## Docker for the backend

```bash
docker compose up -d --build
```

Starts PostgreSQL and the `sfr` API service (healthcheck-gated dependency, restart policy
`unless-stopped`). The indexer is not started by default; run the image with `ingest`:

```bash
docker compose run --rm sfr ingest
```

(See [Deployment](./deployment.md) for the production topology.)

## Cleanup

```bash
make clean                 # remove bin/
docker compose down        # stop services (keeps data volume)
docker compose down -v     # stop and delete the data volume
```

Next: [Testing](./testing.md) · [Configuration](./configuration.md)
