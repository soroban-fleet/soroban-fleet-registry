# Contributing to Soroban Fleet Registry

Thank you for your interest in contributing to **Soroban Fleet Registry** (`soroban-fleet-registry`). This project provides an open-source indexing, discovery, resolution, and verification system for Soroban CAP-85 externally managed contract executables.

---

## 1. Architectural Scope & Boundary

`soroban-fleet-registry` is strictly a **read-only observer and deterministic indexer**.

- **Non-Negotiable Scope**:
  - Does NOT sign transactions.
  - Does NOT handle private keys or provide wallet integrations.
  - Does NOT deploy contracts or mutate on-chain state.
  - Does NOT act as a generic blockchain explorer.
- **Protocol Fidelity**: The system adheres strictly to the CAP-85 specification under Stellar Protocol 28. Verification rules (`HEALTHY`, `DRIFT`, `BROKEN_REFERENCE`, `INCOMPLETE`, `UNKNOWN`) must remain deterministic and conservative.

---

## 2. Prerequisites

To build and run all components locally:
- **Go**: `1.26+` (recommended: `1.27.1`)
- **Node.js**: `18.x+` (recommended: `20+` or `24`)
- **npm**: `10+`
- **PostgreSQL**: `16+`
- **Docker**: For containerized database or service testing

---

## 3. Local Setup & Workflow

### 3.1. Clone and Build Backend

```bash
git clone https://github.com/soroban-fleet/soroban-fleet-registry.git
cd soroban-fleet-registry

# Check formatting and vet Go code
make fmt-check
make vet

# Build the CLI binary
make build
./bin/sfr --help
```

### 3.2. Local PostgreSQL & Migrations

```bash
# Start local PostgreSQL via Docker
docker compose up -d postgres

# Configure environment
export SFR_DATABASE_URL="postgres://sfr:sfrpassword@localhost:5432/sfr?sslmode=disable"

# Apply database migrations
./bin/sfr migrate up

# Rollback migrations if testing schema changes
./bin/sfr migrate down
```

### 3.3. Running the Next.js Frontend

```bash
# Install frontend dependencies
make web-install

# Run frontend tests
make web-test

# Run frontend linter
make web-lint

# Build production bundle
make web-build

# Start local Next.js dev server
cd web && npm run dev
```

---

## 4. Testing & Code Quality

All pull requests must pass the existing automated test suite with the race detector enabled.

### Backend Tests
```bash
# Run all unit and integration tests with race detector
go test -v -race -p 1 ./...

# Run end-to-end verification script
./scripts/verify_e2e.sh
```

### Frontend Tests
```bash
cd web
npm test
npm run lint
npm run build
```

---

## 5. Commit & Pull Request Guidelines

### Commit Message Format
We follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:

```text
<type>(<scope>): <subject>
```

**Common Types:**
- `feat`: New feature or capability
- `fix`: Bug fix
- `test`: Adding or updating test cases
- `docs`: Documentation changes
- `chore`: Maintenance, dependencies, or tooling
- `ci`: CI workflow updates
- `style`: Formatting changes (`gofmt`, linter adjustments)

**Common Scopes:**
- `cap85`: CAP-85 decoding, tag resolution
- `ingest`: Ledger ingestion, checkpoints, pipeline
- `fleet`: Fleet metadata and membership
- `verification`: Verification engine and drift checks
- `api`: HTTP REST API endpoints and handlers
- `cli`: `sfr` command-line interface
- `web`: Next.js web application
- `deploy`: Deployment, Docker, environment configuration

### Pull Request Expectations
1. **Focus**: Keep PRs focused on a single logical change. Do not combine unrelated refactors.
2. **Tests**: Include comprehensive unit and integration tests for new behavior or bug fixes.
3. **No Placeholders**: Avoid stub implementations, `TODO` markers in production paths, or mock fallbacks.
4. **Clean Git State**: Do not commit secrets, environment credentials, `.env` files, or temporary artifacts.
