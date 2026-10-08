# Ecosystem Impact Measurement Plan

## 1. Measurement Methodology

To ensure transparent reporting to the Stellar Community Fund without claiming unverified adoption:

1. **Automated Internal Metrics**:
   - Total indexed fleets, active members, releases, and verification events are recorded deterministically in PostgreSQL.
   - The `/v1/health` and planned `/metrics` endpoints report live database counts and checkpoint sequence lag.
2. **External Adoption Verification**:
   - Developer and protocol integrations will be verified through public GitHub Actions workflows utilizing `sfr fleet verify` or documented webhook integrations.
   - User feedback and satisfaction will be collected using standardized follow-up surveys with pilot maintainers.
3. **Audit Trail**:
   - Monthly impact snapshots will be published as open-source reports in the repository under `docs/reports/` for delegate review.

---

## 2. Reporting Schedule Aligned with SCF Tranches

| Review Gate | Timing | Key Metrics Reported | Deliverable |
|---|---|---|---|
| **Tranche 1 Completion** | Month 1.5 | Indexer throughput, test suite pass rates, latency benchmarks, rate-limiting stress test results. | Tranche 1 Technical Milestone Report |
| **Tranche 2 Completion** | Month 3.0 | Number of testnet pilot protocols onboarded, developer feedback survey results, verification runs executed. | Testnet Pilot Validation Report |
| **Tranche 3 Completion** | Month 4.5 | Mainnet fleets indexed, mainnet service uptime percentage, average checkpoint lag, active API consumers. | Final Grant Completion & Mainnet Report |
