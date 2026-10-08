# Public Service Launch & Operational SLA Plan

## 1. Target Service Level Agreement (SLA)

Upon completion of Milestone 3.1, the public Mainnet instance of Soroban Fleet Registry will commit to the following operational objectives:

- **Service Availability Target**: 99.5% uptime on public HTTP endpoints (`api.sorobanfleet.org` and `app.sorobanfleet.org`), measured via external monitoring.
- **Indexer Freshness Objective**: Average ledger checkpoint lag < 5 ledgers during normal network conditions.
- **Recovery Time Objective (RTO)**: Maximum 1 hour to restore database and ingestion services following a server node failure.
- **Recovery Point Objective (RPO)**: Zero state loss of consensus data (Stellar ledger is the permanent source of truth; database checkpoints can be safely resynchronized).

---

## 2. Infrastructure Architecture & Data Boundaries

| Boundary Classification | Infrastructure Component | Description |
|---|---|---|
| **Public Surface** | Web App UI (`/fleets`, `/contracts`) | Static Next.js frontend cached via CDN. |
| **Public Surface** | REST API (`/v1/fleets`, `/v1/health`) | Read-only JSON API protected by rate limiting. |
| **Private Surface** | PostgreSQL Database | Private VPC network; strictly accessible to backend and indexer processes. |
| **Observable Surface** | Prometheus Metrics (`/metrics`) | Authenticated endpoint reporting lag, error rates, and throughput. |
| **Secret Surface** | Database passwords & RPC API keys | Stored in encrypted cloud secret manager; never logged or exposed. |
