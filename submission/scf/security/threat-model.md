# Threat Model — Soroban Fleet Registry

## 1. System Architecture & Boundaries

Soroban Fleet Registry is an open-source, read-only indexing and verification service.

```text
               +-------------------------------------------+
               |                Public Internet            |
               +-------------------------------------------+
                                     |
                                     | HTTPS (User requests)
                                     v
               +-------------------------------------------+
               | Trust Boundary 1: Ingress / Presentation  |
               |                                           |
               |   [ Next.js 14 Web Application ]          |
               +-------------------------------------------+
                                     |
                                     | HTTP/JSON (Read-only API requests)
                                     v
               +-------------------------------------------+
               | Trust Boundary 2: API & Verifier Service  |
               |                                           |
               |   [ Go chi REST API & Verifier Engine ]   |
               +-------------------------------------------+
                      |                             |
                      | SQL (Internal Network)      | JSON-RPC (TLS)
                      v                             v
+-------------------------------+  +-------------------------------+
| Trust Boundary 3: Persistence |  | Trust Boundary 4: Blockchain  |
|                               |  |                               |
|   [ PostgreSQL 16 DB ]        |  |   [ Stellar Soroban RPC ]     |
|   - fleets, members, releases |  |   - getLatestLedger           |
|   - indexer_checkpoints       |  |   - getLedgers, entries       |
+-------------------------------+  +-------------------------------+
```

---

## 2. Core Architectural Trust Assumptions

1. **Read-Only Scope**: The service holds no private keys, signs no transactions, and possesses no administrative credentials capable of modifying on-chain contract state.
2. **Blockchain as Truth**: Soroban RPC and ledger metadata are trusted as canonical historical representations of the Stellar network.
3. **Internal Database Privacy**: The PostgreSQL database runs on an isolated internal network interface and is never exposed directly to the public internet.
4. **Deterministic Computation**: Given identical ledger sequences and RPC data, the verification engine produces identical status outputs.
