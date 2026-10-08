# Funded Development Tranches — Soroban Fleet Registry

The proposed funding roadmap starts from the completed V1.0.0 codebase and requests funding strictly for future engineering, validation, and infrastructure deliverables across three phased tranches.

```text
V1.0.0 (Completed Foundation)
  ├── CAP-85 Parser & DB
  ├── Indexer & Verifier Engine
  └── REST API, CLI & Web App
             |
             v
Tranche 1: Production Hardening & Operations (Weeks 1–6)
  ├── History Archive Backfill (#16) & Change Filtering Optimization (#2)
  ├── Rate Limiting Middleware (#11) & WASM TTL Cache (#5)
  └── Prometheus Metrics Exporter (#12) & Bounded Replay Tooling (#4)
             |
             v
Tranche 2: Testnet Ecosystem Validation & Tooling (Weeks 7–12)
  ├── 2+ Pilot Protocol Integrations & Feedback Implementation
  ├── Live Scenario Recording CLI (#14) & Protocol Migration Harness (#15)
  └── Web Member Drift Drill-down Modal (#8) & Export Tooling (#9)
             |
             v
Tranche 3: Mainnet Public Service Launch (Weeks 13–18)
  ├── High-Availability Mainnet Indexer & API Deployment
  ├── Production TLS, CDN Caching, and DDoS Protection
  ├── Mainnet Uptime SLA & Published Incident Runbooks (#13)
  └── Final Ecosystem Documentation & Compliance Audit Generator (#10)
```

---

## Tranche Breakdown

### Tranche 1: Production Hardening & Operational Observability
- **Duration**: Weeks 1–6 (1.5 months)
- **Focus**: Hardening indexing performance, implementing the threat model mitigations, and adding production telemetry.
- **Key Backlog Alignment**: Issues #2, #4, #5, #11, #12.

### Tranche 2: Testnet Ecosystem Validation & Tooling
- **Duration**: Weeks 7–12 (1.5 months)
- **Focus**: Live pilot onboarding with Soroban protocol maintainers, user-experience improvements, and scenario testing.
- **Key Backlog Alignment**: Issues #8, #9, #14, #15.

### Tranche 3: Mainnet Public Service Launch & Ecosystem Hardening
- **Duration**: Weeks 13–18 (1.5 months)
- **Focus**: Launching a dedicated, high-availability public indexing and verification service on Stellar Mainnet.
- **Key Backlog Alignment**: Issues #3, #6, #7, #10, #13, #16.
