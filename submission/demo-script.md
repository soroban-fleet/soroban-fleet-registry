# End-to-End Product Demonstration Script

**Target Duration**: 2–4 minutes  
**Format**: Screen capture with terminal and browser walkthrough  
**Tone**: Technical, clear, factual

---

### Scene 1 — Overview & Concept (0:00 – 0:30)
- **Visual**: Soroban Fleet Registry web application landing page (`/fleets`) or docs overview.
- **Narrative**:
  > "Soroban Fleet Registry is an open-source indexing and verification service for Soroban contracts using CAP-85 externally managed executables. When protocols deploy dozens of contract instances sharing a single external reference, maintainers need operational visibility into who belongs to the fleet, which WASM bytecode is currently active, and whether any contract has drifted."

### Scene 2 — Fleet Discovery & Membership (0:30 – 1:00)
- **Visual**: Navigate to a fleet detail page (`/fleets/{owner}/{tag}`).
- **Key Points**:
  - Show Fleet Identity: `(owner_address, tag)`.
  - Highlight total active members and resolved WASM hash.
  - View member list (`/fleets/{owner}/{tag}/members`) displaying individual contract IDs referencing the owner's tag.

### Scene 3 — History & Release Tracking (1:00 – 1:30)
- **Visual**: Navigate to Fleet History (`/fleets/{owner}/{tag}/history`).
- **Key Points**:
  - Show recorded releases where the owner contract updated the tag's WASM hash.
  - Explain how the fleet identity remained constant while members automatically resolve the new executable bytecode.

### Scene 4 — Deterministic Verification: HEALTHY (1:30 – 2:15)
- **Visual**: CLI terminal execution followed by the web verification view (`/fleets/{owner}/{tag}/verify`).
- **Command**:
  ```bash
  ./bin/sfr fleet verify --owner CA... --tag vault-v1
  ```
- **Key Points**:
  - Show verification audit output.
  - Exit code `0` returned for `HEALTHY`.
  - Emphasize that `HEALTHY` requires 100% member resolution, zero mismatches, and complete indexer coverage (`indexed_through >= latest_ledger`).

### Scene 5 — Controlled Drift & Safety Precedence (2:15 – 3:00)
- **Visual**: Run verification against a controlled fixture demonstrating divergence (`DRIFT`).
- **Command**:
  ```bash
  ./bin/sfr fleet verify --owner CA... --tag vault-v1 --expected-wasm <EXPECTED_HASH>
  ```
- **Key Points**:
  - CLI prints `DRIFT` with exit code `1`.
  - Highlight that one contract resolved to a different WASM.
  - Demonstrate `INCOMPLETE` state: when the indexer has not reached the network's latest ledger, the engine outputs `INCOMPLETE` (exit code `3`) rather than falsely asserting healthy status.

### Scene 6 — Architecture & Repository (3:00 – 3:30)
- **Visual**: Show the public GitHub repository (`v1.0.0`), live Docusaurus documentation site, and issue backlog.
- **Key Points**:
  - Explain single repository design: Go indexer/API, PostgreSQL, Next.js web application, CLI binary.
  - Reiterate that V1 is strictly read-only and deploys no custom Soroban contract.
