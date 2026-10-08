# Deployment

The production topology has four deployable pieces and one external dependency.

```text
Frontend
   ↓
API
   ↓
PostgreSQL

Indexer
   ↓
Stellar RPC
```

Repository reference: [`docs/deployment.md`](https://github.com/soroban-fleet/soroban-fleet-registry/blob/main/docs/deployment.md)

## Topology

| Component | Process | Writes | Reads |
|---|---|---|---|
| Frontend | Next.js (`web/`) | — | API over HTTPS |
| API | `sfr api` | verification audit rows | PostgreSQL, Stellar RPC (live resolution) |
| PostgreSQL | `postgres:16` | fleets, members, releases, checkpoints (via indexer) | — |
| Indexer | `sfr ingest` | fleets, members, releases, checkpoints | Stellar RPC |
| Stellar RPC | external | — | — |

The API and the indexer are separate processes sharing one database. The frontend never touches
PostgreSQL or RPC.

## Startup order

```text
PostgreSQL
    ↓
sfr migrate up        (one-shot)
    ↓
sfr api   +   sfr ingest     (parallel)
    ↓
frontend build + serve
```

The API may start before the indexer has caught up; it serves indexed data and reports
`INCOMPLETE` verification until coverage closes.

## HTTPS

Terminate TLS at the reverse proxy or cloud load balancer (HTTPS frontend → API). Redirect plain
HTTP to HTTPS. The Go server speaks HTTP; do not expose it directly to the internet.

## CORS

Set exact origins — never a wildcard in production:

```bash
SFR_CORS_ALLOWED_ORIGINS=https://fleet.example.com
```

Allowed methods: `GET, OPTIONS`. Allowed headers: `Content-Type, Accept`. Preflight cache: 86400
seconds. If unset, every origin is accepted — acceptable only for local development.

## Private database connection

- PostgreSQL on a private network / VPC; not published publicly.
- Connection string with TLS: `sslmode=require` (or `verify-full`).
- The API's database role can be limited to `SELECT`; only the indexer needs write access.
- Inject credentials through the orchestrator's secret store, not committed `.env` files.

## Containers

Multi-stage `Dockerfile`: builds a static binary with `golang:1.26-alpine`, runs as an
unprivileged `sfr` user on `alpine:3.20`, exposes `8080`, default command `api`.

```bash
docker build -t soroban-fleet-registry:latest .

# API
docker run -d \
  -e SFR_DATABASE_URL='postgres://user:pass@db:5432/sfr?sslmode=require' \
  -e SFR_CORS_ALLOWED_ORIGINS=https://fleet.example.com \
  -p 8080:8080 \
  soroban-fleet-registry:latest api

# Indexer
docker run -d \
  -e SFR_DATABASE_URL='postgres://user:pass@db:5432/sfr?sslmode=require' \
  -e SFR_RPC_URL=https://soroban-testnet.stellar.org \
  soroban-fleet-registry:latest ingest
```

`docker compose up -d --build` starts PostgreSQL (with healthcheck) plus the API service.

## Restart behavior

| Component | Restart behavior |
|---|---|
| PostgreSQL | In-flight transactions roll back; restart API and indexer |
| API | Stateless; restart and health turns `200` immediately |
| Indexer | Resumes at `checkpoint + 1`; replay is idempotent |
| Frontend | Stateless; restart any time |
| Stellar RPC outage | Indexer retries with backoff; checkpoints do not advance; verification reports `INCOMPLETE` |

Details: [Failure and recovery](../architecture/failure-recovery.md) ·
[Operations → Recovery](../operations/recovery.md)

## Configuration checklist

- [ ] `SFR_DATABASE_URL` with TLS and a secrets manager
- [ ] `SFR_RPC_URL` pointing at a reliable RPC endpoint for the target network
- [ ] `SFR_HTTP_ADDR` matching the container port (`:8080`)
- [ ] `SFR_CORS_ALLOWED_ORIGINS` set to the exact frontend origin
- [ ] `SFR_LOG_LEVEL` at `info` or `warn` in production
- [ ] `SFR_INDEXER_START_LEDGER` set to a sensible network ledger for a fresh database
- [ ] `NEXT_PUBLIC_API_BASE_URL` baked in at frontend **build** time
- [ ] TLS terminated at ingress; HTTP redirected

## Health checks

Wire orchestrator probes to `GET /health` (or `GET /v1/health`). It confirms the process is
serving — not that indexing is current. Monitor freshness separately via the checkpoint versus
network ledger: [Operations → Health](../operations/health.md).
