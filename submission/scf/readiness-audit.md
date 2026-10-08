# SCF Readiness Audit — Soroban Fleet Registry

## Executive Summary

This readiness audit evaluates `soroban-fleet/soroban-fleet-registry` for submission to the **Stellar Community Fund (SCF) Build Award** program. 

The project has achieved complete technical execution across V1.0.0 (indexer, CAP-85 parser, verification engine, REST API, CLI, web interface, test suite, and documentation site). This audit objectively separates verified achievements from external readiness requirements.

---

## Program & Track Fit Analysis

| Track | Eligibility & Scope | Project Fit Evaluation | Recommendation |
|---|---|---|---|
| **Open Track** | For teams building new core products or infrastructure on Stellar up to $150,000 in XLM across milestone tranches. | **Direct Fit**: SFR provides novel operational infrastructure for CAP-85 externally managed contract executables introduced in Protocol 28 and active on Protocol 29. | **Primary Candidate** |
| **Integration Track** | For projects integrating pre-existing Stellar ecosystem "building blocks" into existing consumer/business products. | **Poor Fit**: SFR is not an external product integrating Stellar blocks; it is developer infrastructure built natively on Stellar RPC and XDR. | Not applicable |
| **RFP Track** | For proposals specifically fulfilling an active SCF Request for Proposals. | **Conditional**: No active RFP directly matches CAP-85 fleet indexing at the current round date. | Not applicable |
| **Public Goods Award** | Award category reserved for proven, established public goods with documented ecosystem usage. | **Future Opportunity**: SFR is open-source (Apache-2.0) and read-only, making it eligible once ecosystem adoption and mainnet public usage are established. | Defer to post-launch |

**Selected Track**: **Open Track** (Build Award).

---

## Program Parameters Verification

- **Program**: Stellar Community Fund (SCF) Build Award
- **Current Round**: SCF #46
- **Published Submission Deadline**: November 8, 2026
- **Award Structure**: Milestone-based funding distributed in payments tied to verifiable tranches (typically 3–4 months execution span, tranches completed within 90 days).
- **Award Ceiling**: Up to $150,000 in XLM (projects must budget based on actual future engineering, infrastructure, and operational costs rather than simply targeting the ceiling).
- **Governance**: Reviewed by verified community delegates and Neural Quorum Governance (NQG).
- **Source of Truth**: [SCF Handbook](https://stellar.gitbook.io/scf-handbook/).

---

## SCF Application Evidence Matrix

| SCF Concern | Evaluation Criteria | Location in Package | Current Status |
|---|---|---|---|
| **Stellar Relevance** | Solves a native Stellar/Soroban requirement using protocol primitives. | [`submission/scf/final/scf-technical-summary.md`](./final/scf-technical-summary.md) | ✅ **VERIFIED** (CAP-85 XDR & LedgerCloseMeta) |
| **Technical Feasibility** | Functional implementation, automated tests, clean architecture. | [`submission/scf/final/scf-technical-summary.md`](./final/scf-technical-summary.md), repo test suites | ✅ **VERIFIED** (Go test race clean, Vitest 47/47) |
| **Differentiation** | Clear boundary vs general explorers and low-level indexers. | [`submission/scf/market/differentiation.md`](./market/differentiation.md) | ✅ **VERIFIED** (Relationship-level indexing) |
| **Security Architecture** | Trust boundaries, attack surfaces, threat modeling. | [`submission/scf/security/threat-model.md`](./security/threat-model.md) | ✅ **VERIFIED** (Full STRIDE analysis) |
| **Monitoring Plan** | Operational metrics, alerting thresholds, runbooks. | [`submission/scf/security/monitoring-plan.md`](./security/monitoring-plan.md) | ✅ **VERIFIED** (Signals, triggers, severity) |
| **Impact Framework** | Measurable outputs, adoption signals, long-term impact. | [`submission/scf/impact/metrics-framework.md`](./impact/metrics-framework.md) | ✅ **VERIFIED** (Baseline vs target model) |
| **Future Roadmap** | Scoped, time-bound tranches requesting funding for future work only. | [`submission/scf/roadmap/tranches.md`](./roadmap/tranches.md) | ✅ **VERIFIED** (Tranches 1, 2, and 3) |
| **Validated Need** | Independent protocol maintainers confirming operational need. | [`submission/scf/validation/validation-log.md`](./validation/validation-log.md) | ❌ **BLOCKED** (0 external interviews completed) |
| **Team Profiles** | Verified personal resumes, public portfolios, and time commitments. | [`submission/scf/team/team-profile.md`](./team/team-profile.md) | ❌ **BLOCKED** (Input required from team) |
| **Cost Budget** | Detailed engineering rates and infrastructure pricing. | [`submission/scf/budget/tranche-budget.md`](./budget/tranche-budget.md) | ❌ **BLOCKED** (Contracting rates required from team) |
| **Public Deployment** | Publicly accessible production instance with TLS and monitoring. | [`submission/scf/deployment/hosting-decision.md`](./deployment/hosting-decision.md) | ❌ **BLOCKED** (Cloud infrastructure not yet provisioned) |

---

## Action Items & Blocker Resolution Roadmap

1. **BLOCKER 1 (Validated Need)**: Execute [`validation/outreach-template.md`](./validation/outreach-template.md) across Stellar Developer Discord (`#soroban-dev`) to secure 3+ external interviews from teams deploying factory contracts.
2. **BLOCKER 2 (Team Profile)**: Maintainer team must supply confirmed legal/professional names, portfolios, and weekly hours in [`team/team-profile.md`](./team/team-profile.md).
3. **BLOCKER 3 (Budget Inputs)**: Maintainer team must confirm exact hourly contractor rates and server quotes in [`budget/assumptions.md`](./budget/assumptions.md).
4. **BLOCKER 4 (Production Hosting)**: Provision cloud container instances and dedicated Soroban RPC node prior to claiming a live public mainnet endpoint (or formally structure as Tranche 3 deliverable).
