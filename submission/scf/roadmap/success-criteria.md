# Roadmap Success Criteria & Verification Gates

The following binary criteria govern the formal completion and approval of each development tranche.

---

## Tranche Review Verification Gates

```text
Tranche 1 Review Gate
  ├── [ ] Rate limiting prevents simulated DoS (HTTP 429 verified in integration test)
  ├── [ ] Prometheus OpenMetrics output matches standard schema
  ├── [ ] Indexer change filtering benchmark confirms zero regression
  ├── [ ] Bounded replay command passes CLI unit tests
  └── [ ] All existing and new tests pass with Go race detector

Tranche 2 Review Gate
  ├── [ ] Documented onboarding of >= 2 external Stellar pilot protocols on testnet
  ├── [ ] Published pilot validation case study in repository docs
  ├── [ ] Live scenario recording CLI operational and verified
  ├── [ ] Web drift drill-down component tested and accessible
  └── [ ] 100% of new frontend and backend tests pass in CI

Tranche 3 Review Gate
  ├── [ ] Public Mainnet indexing service live on stable HTTPS domain
  ├── [ ] Average mainnet indexer lag verified under 5 ledgers
  ├── [ ] 30 consecutive days of mainnet uptime >= 99.5%
  ├── [ ] Automated database backup and restore procedure verified
  └── [ ] Final project completion report published and reviewed by community delegates
```

Each gate requires verifiable, public proof before milestone sign-off.
