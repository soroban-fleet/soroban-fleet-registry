# Quick start

Get the repository running locally: database, API, indexer, web app, and this documentation site.

## Repository

```bash
git clone https://github.com/soroban-fleet/soroban-fleet-registry.git
cd soroban-fleet-registry
```

Prerequisites are listed in [Development → Prerequisites](../development/prerequisites.md)
(Go 1.26+, PostgreSQL 16+, Node.js 18+).

## Backend

```bash
# 1. Start PostgreSQL
docker compose up -d postgres

# 2. Apply migrations
export SFR_DATABASE_URL="postgres://sfr:sfrpassword@localhost:5432/sfr?sslmode=disable"
make build
./bin/sfr migrate up

# 3. Start the API server
export SFR_HTTP_ADDR=":8080"
export SFR_RPC_URL="https://soroban-testnet.stellar.org"
./bin/sfr api
```

In a second terminal, start the indexer if you want live testnet indexing:

```bash
export SFR_DATABASE_URL="postgres://sfr:sfrpassword@localhost:5432/sfr?sslmode=disable"
export SFR_RPC_URL="https://soroban-testnet.stellar.org"
./bin/sfr ingest
```

Check that the API answers:

```bash
curl -s http://localhost:8080/health
```

```json
{
  "status": "UP",
  "service": "soroban-fleet-registry",
  "timestamp": "2026-10-07T12:00:00Z"
}
```

## Frontend

```bash
cd web
npm install
npm run dev
```

The app runs at `http://localhost:3000` and calls the API at `http://localhost:8080` by default.
See [Guide: Use the web app](../guides/use-web-app.md).

## CLI

```bash
./bin/sfr fleet list
./bin/sfr fleet inspect --owner <OWNER> --tag <TAG>
./bin/sfr fleet verify --owner <OWNER> --tag <TAG>; echo "exit=$?"
```

Replace `<OWNER>` and `<TAG>` with a real fleet identity. See
[CLI reference](../reference/cli.md) for every command, flag, and exit code.

## Docs

```bash
cd docs-site
npm install
npm run start
```

The documentation site runs at `http://localhost:3000`. Build it with:

```bash
cd docs-site
npm ci
npm run build
```

## Tests

```bash
make test    # go test -v -race -p 1 ./... (needs PostgreSQL)
make vet
cd web && npm test && npm run lint && npm run build
```

Full details: [Development → Testing](../development/testing.md).

## Environment files

Copy the templates instead of inventing values:

```bash
cp .env.example .env          # backend
cp web/.env.example web/.env.local
```

Every variable is documented in [Reference → Environment](../reference/environment.md).
