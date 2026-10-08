# Production Hosting Strategy Evaluation & Decision

## 1. Context & Objectives

Soroban Fleet Registry requires a reliable hosting environment to execute as a public read-only indexer and verification service. In accordance with the Stellar Maintainer Playbook, this evaluation compares realistic hosting strategies against operational criteria rather than selecting infrastructure based on prestige.

---

## 2. Infrastructure Strategy Comparison

### Strategy A: Managed Cloud Architecture (Managed DB + Container Services)
- **Architecture**:
  - Presentation: Vercel or Cloudflare Pages (Next.js frontend).
  - API & Indexer: Managed Container Runner (e.g., Fly.io, AWS ECS, or Render).
  - Persistence: Managed PostgreSQL 16 (AWS RDS or Supabase/Neon with PITR backups).
  - Upstream RPC: Dedicated Soroban RPC provider (e.g., Blockdaemon, NowNodes, or SDF public node with commercial fallback).
- **Estimated Monthly Cost**:
  - Next.js Web Frontend: $20 / mo (Pro team plan with custom domain & CDN)
  - Go Backend & Continuous Indexer (2 vCPU, 4GB RAM): $40 / mo
  - Managed PostgreSQL 16 (2 vCPU, 4GB RAM, 50GB SSD, automated daily backups): $60 / mo
  - Dedicated / High-Rate-Limit Soroban RPC Endpoint: $150–$250 / mo
  - Egress, Domain, Cloudflare SSL/WAF: $20 / mo
  - **Total Estimated Cost**: ~$290 – $390 / month (~$3,500 – $4,680 / year)
- **Failure Modes & Resiliency**:
  - Automatic container restarts on memory/process panic.
  - Independent database failover and automated point-in-time recovery (RPO < 5 minutes).
  - Zero server OS patching burden on maintainers.
- **Complexity**: Low to moderate. Standard CI/CD automated deployment via Docker.

### Strategy B: Single Dedicated VPS Host (Self-Hosted Docker Compose)
- **Architecture**:
  - Single large cloud instance (e.g., Hetzner / DigitalOcean Droplet: 8 vCPU, 16GB RAM, 160GB NVMe SSD).
  - Runs Next.js, Go API, Indexer daemon, and PostgreSQL 16 via Docker Compose.
  - Nginx reverse proxy with Let's Encrypt SSL and local WAL archiving to object storage (S3).
- **Estimated Monthly Cost**:
  - Dedicated Cloud Droplet / VPS: $50 – $80 / mo
  - Dedicated Soroban RPC Node / Tier: $150 – $250 / mo
  - S3 / Object Storage Backups: $10 / mo
  - **Total Estimated Cost**: ~$210 – $340 / month (~$2,520 – $4,080 / year)
- **Failure Modes & Resiliency**:
  - Single point of failure: host kernel crash or disk failure takes down API, DB, and Indexer simultaneously.
  - Requires manual server hardening, OS patching, firewall maintenance, and failover automation.
- **Complexity**: High operational burden for maintainers.

---

## 3. Upstream Stellar RPC Strategy Evaluation

| Option | Reliability | Rate Limits | Latency | Operational Burden | Evaluation & Recommendation |
|---|---|---|---|---|---|
| **Public SDF RPC (`soroban-rpc.stellar.org`)** | Medium | Strict (rate-limited bursts) | Moderate | Zero | **Acceptable for local development and non-critical testing**, but prone to throttling during continuous batch ingestion. |
| **Dedicated Managed RPC (e.g., Commercial Provider)** | High (99.9% SLA) | High (unmetered or high RPS tier) | Low | Low | **Recommended for production public service**. Ensures reliable batch catchup without 429 throttling. |
| **Self-Hosted Soroban RPC Node** | Full control | Unlimited internal | Lowest | Very High (requires running full Stellar Core + Soroban RPC node + history archive ingestion) | **Not recommended for V1**; prohibitive maintenance burden and high infrastructure costs ($500+/mo). |

---

## 4. Production Decision & Recommendation

### Selected Strategy: **Strategy A (Managed Cloud Architecture) with Commercial Managed RPC**
- **Rationale**: Separating the PostgreSQL database into a managed instance with automated backups eliminates single-server data loss risk. Offloading the frontend to an edge CDN ensures fast global response times and DDoS protection while keeping maintainer operational overhead focused on indexing logic.
- **Implementation Gate**: Provisioning will be funded and executed in **Tranche 3 (Milestone 3.1)** of the SCF grant.

---

## 5. Deployment Authorization Status

```text
STATUS: BLOCKED (Awaiting Tranche 3 Infrastructure Allocation)
MISSING INPUTS:
1. Production cloud hosting account and billing authorization.
2. Commercial Soroban RPC dedicated endpoint credentials.
3. Production domain assignment (e.g., api.sorobanfleet.org).
```

*In strict compliance with the Do Not Fabricate Evidence standard, no active public mainnet URL is claimed.*
