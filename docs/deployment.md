# Soroban Fleet Registry — Deployment & Service Topology Guide

This document describes the deployment architecture, configuration matrix, service topologies, operational lifecycle, health monitoring, failure recovery, and production verification procedures for `soroban-fleet-registry`.

---

## 1. System Architecture & Topology

Soroban Fleet Registry is a read-only observer and deterministic verification engine for Soroban CAP-85 externally managed contract executable fleets.

### Production Topology Overview

```text
                         ┌─────────────────────────┐
                         │          USER           │
                         │    (Browser Client)     │
                         └────────────┬────────────┘
                                      │ HTTPS
                                      ▼
                         ┌─────────────────────────┐
                         │    Next.js Frontend     │
                         │   Web Service (SSR/UI)  │
                         └────────────┬────────────┘
                                      │ HTTPS / Internal API
                                      ▼
                         ┌─────────────────────────┐
                         │       Go REST API       │
                         │  (SFR Server Process)   │
                         └───────┬─────────┬───────┘
                                 │         │
                 Private Network │         │ HTTPS Outbound
                                 ▼         ▼
             ┌──────────────────────┐   ┌────────────────────────┐
             │      PostgreSQL      │   │  Stellar Soroban RPC   │
             │   (v16+ Database)    │   │  (Testnet / Mainnet)   │
             └───────────▲──────────┘   └───────────▲────────────┘
                         │                          │
                         │     ┌──────────────┐     │
                         └─────┤ SFR Indexer  ├─────┘
                               │  Background  │
                               │   Pipeline   │
                               └──────────────┘
```

### Strict Service Boundaries
1. **User / Browser**: Communicates exclusively with the public web frontend or public REST API over HTTPS. Never communicates directly with PostgreSQL or Stellar RPC.
2. **Web Frontend (Next.js)**: Stateless presentation layer. Consumes the Go REST API. Does not access PostgreSQL, Stellar RPC, or store private keys.
3. **REST API (Go)**: Serves fleet summaries, member listings, release history, and on-demand verification requests. Read-only against PostgreSQL; uses Stellar RPC only for live contract executable resolution.
4. **Indexer (Go)**: Background stream processor. Ingests ledger changes from Stellar RPC, parses CAP-85 executable references, records fleet memberships and releases, and updates checkpoints transactionally.
5. **PostgreSQL (16+)**: Authoritative datastore for indexed fleets, members, historical releases, verification logs, and stream checkpoints. Internal network access only.

---

## 2. Environment Configuration Matrix

Configuration is managed via explicit environment variables. Production deployments must inject values via orchestration secrets/config managers rather than committing `.env` files.

| Variable | Local Development | Testnet Deployment | Production Deployment | Required In |
|---|---|---|---|---|
| `SFR_NETWORK` | `testnet` (or `fixture`) | `testnet` | `pubnet` / `mainnet` | Backend / Indexer |
| `SFR_RPC_URL` | `https://soroban-testnet.stellar.org` | `https://soroban-testnet.stellar.org` | Dedicated high-availability RPC endpoint | Indexer & Live Verification |
| `SFR_DATABASE_URL` | `postgres://sfr:sfrpassword@localhost:5432/sfr?sslmode=disable` | Private managed DB URL | Private VPC connection string (`sslmode=verify-full`) | Backend & Indexer |
| `SFR_HTTP_ADDR` | `:8080` | `:8080` (or platform `$PORT`) | `:8080` | Backend API |
| `SFR_LOG_LEVEL` | `debug` or `info` | `info` | `info` or `warn` | Backend & Indexer |
| `SFR_INDEXER_START_LEDGER` | `1000` (or empty to resume) | Network start ledger sequence | Network CAP-85 activation sequence | Indexer |
| `SFR_INDEXER_BATCH_SIZE` | `50` | `50` - `100` | `50` - `200` (depending on RPC limits) | Indexer |
| `SFR_CORS_ALLOWED_ORIGINS` | `http://localhost:3000` | `https://fleet-testnet.example.com` | `https://fleet.example.com` | Backend API |
| `NEXT_PUBLIC_API_BASE_URL` | `http://localhost:8080` | `https://api.fleet-testnet.example.com` | `https://api.fleet.example.com` | Frontend Web UI |
| `NEXT_PUBLIC_NETWORK` | `testnet` | `testnet` | `mainnet` | Frontend Web UI |
| `NEXT_PUBLIC_EXPLORER_BASE_URL`| `https://stellar.expert/explorer/testnet` | `https://stellar.expert/explorer/testnet` | `https://stellar.expert/explorer/public` | Frontend Web UI |

> **Security Rule**: `NEXT_PUBLIC_*` variables are embedded into browser assets at build time. Never expose database credentials, RPC API keys, or private keys through `NEXT_PUBLIC_*` variables.

---

## 3. Local Development Topology

To run the full stack locally:

```bash
# 1. Start PostgreSQL
docker compose up -d postgres

# 2. Apply database migrations
export SFR_DATABASE_URL="postgres://sfr:sfrpassword@localhost:5432/sfr?sslmode=disable"
sfr migrate up

# 3. Start the REST API server
export SFR_HTTP_ADDR=":8080"
export SFR_RPC_URL="https://soroban-testnet.stellar.org"
sfr api

# 4. In a separate terminal, start the indexer (optional for live testnet indexing)
sfr ingest

# 5. In a separate terminal, start the Next.js web application
cd web
npm install
npm run dev
```

The web application is available at `http://localhost:3000`, communicating with the API on `http://localhost:8080`.

---

## 4. Production Service Deployment

### 4.1. Database Setup
1. Provision a PostgreSQL 16+ instance (AWS RDS, GCP Cloud SQL, or Supabase).
2. Ensure database connections require SSL/TLS (`sslmode=require` or `sslmode=verify-full`).
3. Run migrations as a one-time deployment step or pre-startup init container:
   ```bash
   sfr migrate up
   ```

### 4.2. Backend / API Service
The backend container is built from the multi-stage [`Dockerfile`](file:///home/smog/soroban-fleet/soroban-fleet-registry/Dockerfile) and runs as an unprivileged user:
```bash
docker build -t soroban-fleet-registry:latest .
docker run -d \
  -e SFR_NETWORK=testnet \
  -e SFR_RPC_URL=https://soroban-testnet.stellar.org \
  -e SFR_DATABASE_URL=postgres://user:pass@db-host:5432/sfr?sslmode=require \
  -e SFR_HTTP_ADDR=:8080 \
  -e SFR_CORS_ALLOWED_ORIGINS=https://fleet.example.com \
  -p 8080:8080 \
  soroban-fleet-registry:latest api
```

### 4.3. Indexer Service
Deploy the indexer container with the command override `ingest`:
```bash
docker run -d \
  -e SFR_NETWORK=testnet \
  -e SFR_RPC_URL=https://soroban-testnet.stellar.org \
  -e SFR_DATABASE_URL=postgres://user:pass@db-host:5432/sfr?sslmode=require \
  -e SFR_INDEXER_BATCH_SIZE=50 \
  soroban-fleet-registry:latest ingest
```

### 4.4. Frontend Service
The web application is deployed to Node.js hosting (Vercel, AWS ECS, Cloudflare Pages, Kubernetes):
```bash
cd web
npm ci
NEXT_PUBLIC_API_BASE_URL=https://api.fleet.example.com \
NEXT_PUBLIC_NETWORK=testnet \
NEXT_PUBLIC_EXPLORER_BASE_URL=https://stellar.expert/explorer/testnet \
npm run build
npm run start
```

---

## 5. Service Startup Order & Orchestration

The system components must be orchestrated in the following dependency order:

```text
PostgreSQL (Port 5432)
       │
       ▼
Migrations Up (sfr migrate up)
       │
       ├─────────────────────────┐
       ▼                         ▼
Go REST API (sfr api)     Go Indexer (sfr ingest)
       │                         │
       ▼                         ▼
Next.js Frontend (web)     Stellar RPC Ingestion
```

- **Independent Health**: The Go REST API can start and serve requests even if the indexer has not yet caught up to the latest ledger.
- **Graceful Indexing Lag**: When the indexer is lagging, verification returns `status: INCOMPLETE` along with `indexer_coverage` metadata. The API does not crash or pretend the fleet is healthy.

---

## 6. Health Checks vs. Indexer Readiness

The system explicitly distinguishes **service liveness** from **indexer synchronization**:

### Service Health (`GET /health` or `GET /v1/health`)
- **HTTP 200**: Indicates that the HTTP server process is running and accepting network connections.
- Response:
  ```json
  {
    "status": "UP",
    "service": "soroban-fleet-registry",
    "timestamp": "2026-10-07T12:00:00Z"
  }
  ```

### Indexer Readiness & Freshness
- Indexer freshness is surfaced on a per-fleet basis via `GET /v1/fleets/{owner}/{tag}/verify`:
  ```json
  {
    "status": "INCOMPLETE",
    "indexer_coverage": {
      "latest_indexed_ledger": 1050,
      "network_latest_ledger": 5071169,
      "ledger_gap": 5070119,
      "is_caught_up": false
    }
  }
  ```
- **Safety Principle**: A healthy HTTP API never converts a ledger gap into a false `HEALTHY` verification verdict.

---

## 7. Failure Handling & Disaster Recovery

### 7.1. Database Failure
- **Behavior**: Database query failures return HTTP 500 with structured JSON error payload `{ "error": { "code": "INTERNAL_ERROR", "message": "Failed to list fleets" } }`.
- **Safety**: No database connection strings, stack traces, or schema internals are leaked in response bodies.

### 7.2. Stellar RPC Outage
- **Indexer Behavior**: Network errors or RPC timeouts trigger exponential backoff logging warnings without crashing the indexer process. Checkpoints are only advanced upon successful commit of processed ledger blocks.
- **API Behavior**: Fleet verification marks status as `INCOMPLETE` or `UNKNOWN` if live resolution fails. Live resolution errors do not corrupt indexed database state.

### 7.3. Indexer Process Interruption & Replay
- The ingestion pipeline processes ledger transactions in atomic database transactions (`BeginTx`).
- The checkpoint sequence in `indexer_checkpoints` is updated inside the same transaction as the fleet entity changes.
- If the process terminates abruptly during ledger $N$, the transaction rolls back. Upon restart, the indexer resumes from the last successfully committed checkpoint, replaying ledger $N$ idempotently without duplicate records.

---

## 8. Security Boundaries & CORS

1. **Read-Only Guarantees**:
   - The web frontend and API contain zero transaction-signing logic, private keys, or wallet connectors.
   - Database user roles for the REST API can be granted strictly `SELECT` permissions.
2. **CORS Policy**:
   - Configured via `SFR_CORS_ALLOWED_ORIGINS`.
   - In production, specify exact origins (e.g. `SFR_CORS_ALLOWED_ORIGINS=https://fleet.example.com`). Wildcards (`*`) must not be used in secure environments.
3. **Transport Encryption**:
   - Production ingress must terminate TLS 1.3/1.2 at reverse proxies or cloud load balancers. Plain HTTP traffic must be permanently redirected to HTTPS.
