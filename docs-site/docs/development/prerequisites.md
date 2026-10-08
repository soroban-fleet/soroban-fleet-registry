# Development prerequisites

Versions below come from the repository itself (`go.mod`, `web/package.json`,
`.github/workflows/ci.yml`, `docker-compose.yml`) — not from guesswork.

## Backend

| Tool | Version | Source |
|---|---|---|
| Go | `1.26.7` toolchain; CI uses `1.26` | `go.mod`, `.github/workflows/ci.yml` |
| PostgreSQL | `16+` (`postgres:16-alpine` image) | `docker-compose.yml`, CI service |
| Docker | any version that runs Compose v2 | `docker-compose.yml` (optional but recommended) |

```bash
go version      # expect go1.26.x
psql --version  # or use the Docker container
```

### Verify the Go toolchain

```bash
make fmt-check
make vet
make build
./bin/sfr --help
```

## Frontend

| Tool | Version | Source |
|---|---|---|
| Node.js | `18+` (CI: `20`) | `docs/web.md`, CI `setup-node` |
| npm | `10+` (ships with Node 20) | `CONTRIBUTING.md` |

```bash
node --version
npm --version
```

## Documentation site

| Tool | Version |
|---|---|
| Node.js | `18.0+` (`engines` in `docs-site/package.json`); CI uses `20` |
| npm | ships with Node 20 |

## Optional

| Tool | Used for |
|---|---|
| `gofmt` | formatting checks (`make fmt-check`) |
| `curl` | smoke-testing the API |
| Stellar Testnet RPC access | live indexing and verification (`https://soroban-testnet.stellar.org`) |

## Environment templates

```bash
cp .env.example .env
cp web/.env.example web/.env.local
```

Never commit real values. See [Configuration](./configuration.md).

## Everything at a glance

```bash
# Backend checks (needs Go only)
make fmt-check && make vet && make build

# Full backend test run (needs PostgreSQL)
docker compose up -d postgres
export SFR_DATABASE_URL="postgres://sfr:sfrpassword@localhost:5432/sfr?sslmode=disable"
make test

# Frontend checks (needs Node)
cd web && npm ci && npm run lint && npm test && npm run build

# Docs checks (needs Node)
cd docs-site && npm ci && npm run build
```

Next: [Local development](./local-development.md) · [Testing](./testing.md)
