# Testing

Every command below exists in the repository. Nothing here is aspirational.

## Backend

| Command | What it does |
|---|---|
| `go test ./...` | Run all Go tests |
| `make test` | `go test -v -race -p 1 ./...` — race detector, serialized packages |
| `make vet` | `go vet ./...` |
| `make fmt-check` | Fails if any file needs `gofmt` |
| `make build` | Builds `bin/sfr` |

```bash
export SFR_DATABASE_URL="postgres://sfr:sfrpassword@localhost:5432/sfr?sslmode=disable"
make fmt-check
make vet
make test
```

`-p 1` serializes package execution, which keeps database-backed tests from colliding on the
shared PostgreSQL instance.

### Test layout

| Package | Covers |
|---|---|
| `internal/cap85` | XDR decoding, tag validation, resolver behavior |
| `internal/ingest` | change extraction, ledger processing, checkpoints, idempotent replay |
| `internal/fleet` | repository, membership lifecycle, service queries |
| `internal/verification` | classification rules for all five statuses |
| `internal/api` | handlers, pagination, errors, CORS, health |
| `internal/config` | environment parsing and validation |
| `internal/rpc` | RPC client behavior |
| `migrations` | up/down migration cycle |
| `fixtures` | fixture loading and integrity |
| `integration` | end-to-end pipeline scenarios |

### Database requirement

Tests that touch PostgreSQL open `SFR_DATABASE_URL`. If the variable is unset or the database is
unreachable, those tests **skip** rather than fail (`t.Skipf`). Unit tests without a database
always run.

```bash
docker compose up -d postgres
export SFR_DATABASE_URL="postgres://sfr:sfrpassword@localhost:5432/sfr?sslmode=disable"
go test -v -race -p 1 ./...
```

### Single package

```bash
go test -v -race ./internal/verification/
go test -run TestPipeline_IdempotentReplayAndResume ./integration/ -v
```

## End-to-end script

```bash
./scripts/verify_e2e.sh
```

Runs, in order: `gofmt` check → `go vet` → full race-detector test suite → `make build` →
`./bin/sfr --help`. Use it as the pre-push check for backend changes.

CI additionally runs a migration smoke check:

```bash
./bin/sfr migrate down
./bin/sfr migrate up
./bin/sfr --help
```

## Frontend

From `web/`:

| Command | What it does |
|---|---|
| `npm run test` | `vitest run` — unit, component, and acceptance suites |
| `npm run lint` | `next lint` |
| `npm run build` | Production build (also type-checks) |
| `npx tsc --noEmit` | Type check without emitting |

Suites live in `web/__tests__/`: formatting, API client, components, fleet list/detail/members/
history/verify pages, contract detail, smoke, and acceptance journeys covering healthy, drifted,
incomplete, and contract-inspection flows.

```bash
cd web
npm ci
npm run lint
npm run test
npm run build
```

Or via make: `make web-install`, `make web-test`, `make web-lint`, `make web-build`.

## Documentation site

```bash
cd docs-site
npm ci
npm run build
```

`npm run build` fails on broken internal links, broken anchors, and broken Markdown links.

## Continuous integration

`.github/workflows/ci.yml` runs two jobs on every push and pull request to `main`:

| Job | Steps |
|---|---|
| **Build, Vet, and Test** | Go 1.26 setup → `gofmt` check → `go vet` → `go test -v -race -p 1 ./...` (PostgreSQL 16 service) → migration smoke checks |
| **Frontend Lint, Test, and Build** | Node 20 setup → `npm ci` → `npm run lint` → `npm run test` → `npm run build` |

A third job, **Documentation**, runs `npm ci` and `npm run build` for `docs-site/`.

## What is not tested

There is no live-network test suite in CI. Live Testnet checks are manual:

```bash
export SFR_RPC_URL=https://soroban-testnet.stellar.org
./bin/sfr ingest
./bin/sfr fleet verify --owner <OWNER> --tag <TAG>
```
