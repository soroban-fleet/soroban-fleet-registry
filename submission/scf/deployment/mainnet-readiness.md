# Public Mainnet Service Readiness Assessment

## 1. Production Topology & Trust Zones

The proposed architecture for the public deployment separates public ingress, internal processing, and upstream blockchain RPC:

```text
               +-------------------------------------------+
               |                Public Internet            |
               +-------------------------------------------+
                                     |
                                     | HTTPS (TLS 1.3)
                                     v
               +-------------------------------------------+
               | Zone 1: Public Egress & Reverse Proxy     |
               |                                           |
               |   - Cloudflare CDN / Nginx Reverse Proxy  |
               |   - Rate Limiting (Token Bucket)          |
               |   - DDoS Mitigation & Edge SSL            |
               +-------------------------------------------+
                        |                         |
                        | HTTP (Internal)         | HTTP (Internal)
                        v                         v
       +-----------------------+     +-----------------------+
       | Zone 2: Presentation  |     | Zone 2: API Service   |
       |                       |     |                       |
       |  [ Next.js 14 App ]   |     |  [ Go chi REST API ]  |
       +-----------------------+     +-----------------------+
                        |                         |
                        +------------+------------+
                                     |
                                     | Internal Unix / TCP
                                     v
       +-----------------------------------------------------+
       | Zone 3: Storage & Ingestion (Private Network)       |
       |                                                     |
       |  [ PostgreSQL 16 DB ] <--- [ Indexer Pipeline ]     |
       |  - Automated WAL backups   - Catchup & Checkpoints  |
       +-----------------------------------------------------+
                                     |
                                     | HTTPS / JSON-RPC
                                     v
               +-------------------------------------------+
               | Zone 4: External Stellar Network          |
               |                                           |
               |   - Primary: Dedicated Soroban RPC Node   |
               |   - Backup: Public Stellar Mainnet RPC    |
               +-------------------------------------------+
```

---

## 2. Mainnet Operational Readiness Checklist

| Readiness Domain | Operational Requirement | Status | Gap / Action Plan |
|---|---|---|---|
| **Security Scope** | Zero private keys, no wallet connections, strictly read-only | ✅ **READY** | Enforced by architecture |
| **API Transport** | Enforced HTTPS with valid TLS certificate | ⚠️ **PENDING HOSTING** | Requires cloud reverse proxy provisioning |
| **Abuse Mitigation** | IP token-bucket rate limiting middleware | ⚠️ **IN PROGRESS** | Planned in Tranche 1 (Issue #11) |
| **Database Resiliency** | Managed PostgreSQL 16 with automated point-in-time recovery | ⚠️ **PENDING HOSTING** | Cloud RDS / Managed PostgreSQL needed |
| **RPC Resilience** | Primary dedicated RPC node with automatic secondary failover | ⚠️ **PENDING HOSTING** | Dedicated Soroban RPC provider required |
| **Monitoring** | Automated lag alerts and service availability heartbeats | ⚠️ **IN PROGRESS** | Prometheus exporter planned in Tranche 1 (Issue #12) |
| **Runbooks** | Documented step-by-step incident response procedures | ⚠️ **IN PROGRESS** | Planned in Tranche 3 (Issue #13) |

---

## 3. Current Deployment Status

```text
STATUS: NOT YET DEPLOYED TO PUBLIC MAINNET
BLOCKER: Production cloud infrastructure and dedicated RPC hosting allocation required (funded under Tranche 3).
```

*Note: In accordance with the Do Not Fabricate Evidence standard, no synthetic public mainnet URL is published.*
