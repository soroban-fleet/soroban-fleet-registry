# Local Development & Operating Modes

This document details how to set up, build, test, and run `soroban-fleet-registry`.

## Prerequisites

- Go `1.26+` (or `1.27+`)
- PostgreSQL `16+`
- Docker (optional, for containerized deployment)

## Operating Modes

`sfr` operates in three distinct modes:

### 1. Fixture Mode
Runs unit and integration tests against deterministic local fixtures in `fixtures/` without contacting an external Stellar network. Used for offline development and CI.

```bash
go test -v ./fixtures/... ./integration/...
```

### 2. Historical Ingestion Mode
Continuously connects to a Stellar RPC node and ingests ledgers from a configured start sequence into PostgreSQL.

```bash
export SFR_DATABASE_URL="postgres://sfr:sfrpassword@localhost:5432/sfr?sslmode=disable"
export SFR_RPC_URL="https://soroban-testnet.stellar.org"
export SFR_INDEXER_START_LEDGER="1000"

# Apply migrations
sfr migrate up

# Run ingestion
sfr ingest
```

### 3. Live RPC & API Mode
Serves the HTTP REST API and CLI queries using the indexed PostgreSQL database and live RPC for verification.

```bash
export SFR_DATABASE_URL="postgres://sfr:sfrpassword@localhost:5432/sfr?sslmode=disable"
export SFR_RPC_URL="https://soroban-testnet.stellar.org"
export SFR_HTTP_ADDR=":8080"

sfr api
```

## Running Tests

Run all unit and integration tests:

```bash
make test
```

Run code formatting and vet checks:

```bash
make fmt-check
make vet
```

Build the CLI binary:

```bash
make build
./bin/sfr --help
```

## Running PostgreSQL with Docker

```bash
docker compose up -d postgres
```

Run database migrations:

```bash
export SFR_DATABASE_URL="postgres://sfr:sfrpassword@localhost:5432/sfr?sslmode=disable"
./bin/sfr migrate up
```
