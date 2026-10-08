# Contributing: Workflow

Source of truth: [`CONTRIBUTING.md`](https://github.com/soroban-fleet/soroban-fleet-registry/blob/main/CONTRIBUTING.md)
in the repository root. This page summarizes the workflow.

## Scope before you start

SFR is a **read-only observer and deterministic indexer**. Contributions that add transaction
signing, wallet handling, on-chain mutations, or a project token are out of scope for V1. See
[Soroban boundary](../architecture/soroban-boundary.md).

## Local setup

```bash
git clone https://github.com/soroban-fleet/soroban-fleet-registry.git
cd soroban-fleet-registry

make fmt-check
make vet
make build
./bin/sfr --help

docker compose up -d postgres
export SFR_DATABASE_URL="postgres://sfr:sfrpassword@localhost:5432/sfr?sslmode=disable"
./bin/sfr migrate up
```

Frontend and docs:

```bash
cd web && npm ci && npm run dev          # terminal A
cd docs-site && npm ci && npm run start  # terminal B
```

Full detail: [Local development](../development/local-development.md).

## The loop

```text
1. Find or open an issue
2. Branch from main
3. Implement (backend: Go; frontend: TypeScript; docs: Markdown)
4. Add or update tests
5. Run the checks
6. Open a pull request
7. Address review
```

## Checks to run before pushing

```bash
# Backend
make fmt-check && make vet && make test
./scripts/verify_e2e.sh

# Frontend
cd web && npm run lint && npm test && npm run build

# Documentation
cd docs-site && npm run build
```

CI runs the same checks on every pull request:
[Testing → Continuous integration](../development/testing.md#continuous-integration).

## Commit messages

[Conventional Commits](https://www.conventionalcommits.org/): `<type>(<scope>): <subject>`

| Type | Use |
|---|---|
| `feat` | new capability |
| `fix` | bug fix |
| `test` | tests only |
| `docs` | documentation |
| `chore` | maintenance, tooling |
| `ci` | workflow changes |
| `style` | formatting |

Scopes used in this repository: `cap85`, `ingest`, `fleet`, `verification`, `api`, `cli`, `web`,
`deploy`, `docs`, `repo`.

One logical change per commit. Do not commit secrets, `.env` files, or build output.

## Where changes usually land

| Area | Paths |
|---|---|
| Protocol decoding | `internal/cap85/` |
| Ingestion | `internal/ingest/` |
| Fleet/membership | `internal/fleet/` |
| Verification rules | `internal/verification/` |
| REST API | `internal/api/`, `api/openapi.yaml` |
| CLI | `cmd/sfr/` |
| Frontend | `web/` |
| Schema | `migrations/` |
| Documentation | `docs-site/` |

Next: [Issues](./issues.md) · [Pull requests](./pull-requests.md) ·
[Security](./security.md)
