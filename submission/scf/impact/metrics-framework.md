# Ecosystem Impact Metrics Framework

## 1. Metrics Hierarchy

To measure the realistic value delivered by Soroban Fleet Registry, metrics are structured across three hierarchical levels: Direct Outputs, Ecosystem Adoption, and Downstream Outcomes.

```text
Level 1: Direct System Outputs (Controlled directly by SFR infrastructure)
         - Fleets and members indexed
         - Verification runs conducted
         - Release transitions recorded

Level 2: Ecosystem Adoption (Driven by external developers and maintainers)
         - Protocols integrating SFR verification in CI/CD
         - Active API query volume
         - CLI downloads and adoption

Level 3: Downstream Outcomes (Ecosystem-wide benefits)
         - Reduction in manual verification effort
         - Early detection of uncoordinated upgrade drift incidents
```

---

## 2. Core Metrics Definition Table

| Metric | Level | Measurement Source | Frequency | Current Baseline | Target (Tranche 2) | Target (Tranche 3) |
|---|---|---|---|---|---|---|
| **CAP-85 Fleets Indexed** | Output | PostgreSQL `fleets` table | Continuous | 0 (mainnet) | 5 (testnet pilots) | 20+ (mainnet) |
| **Active Fleet Members Indexed** | Output | PostgreSQL `fleet_members` table | Continuous | 0 (mainnet) | 50+ instances | 250+ instances |
| **Verification Runs Conducted** | Output | PostgreSQL `fleet_verifications` | Weekly | 0 (mainnet) | 200+ runs | 1,000+ runs |
| **Detected Drift Incidents** | Output | Verification audits where status = `DRIFT` | Real-time | 0 | Baseline logging | Real-time alerts |
| **External Protocols Integrating SFR** | Adoption | Developer interviews & GitHub CI configs | Monthly | 0 | 2 pilot teams | 5+ protocols |
| **CLI / CI Invocations** | Adoption | Ingestion telemetry / API user agents | Monthly | 0 | 50 monthly runs | 250 monthly runs |
| **Indexer Freshness Lag** | Output | Difference: `latest_ledger - checkpoint` | Continuous | N/A (unhosted) | < 10 ledgers | < 5 ledgers avg |
| **Service Uptime (API/Web)** | Output | External uptime monitor (HTTP 200 checks) | Monthly | N/A (unhosted) | 99.0% (testnet) | 99.5% (mainnet) |
