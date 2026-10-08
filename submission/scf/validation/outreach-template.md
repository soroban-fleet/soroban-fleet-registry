# Ecosystem Outreach Template

This template is designed for direct outreach to protocol maintainers and Soroban developers across the Stellar Developer Discord, GitHub issues, and developer forums.

---

### Outreach Message

**Subject**: Seeking technical feedback: CAP-85 multi-instance contract verification & indexing

Hi [Name/Team],

I’m working on **Soroban Fleet Registry**, an open-source indexing and verification tool for Soroban contracts utilizing CAP-85 externally managed executables (`CONTRACT_EXECUTABLE_EXTERNAL_REF`).

We built an open-source service (Go indexer + REST API + CLI + Next.js UI) that discovers all contract instances pointing to a given `(owner, tag)`, resolves their active WASM bytecode, and verifies fleet integrity across deterministic safety states (`HEALTHY`, `DRIFT`, `INCOMPLETE`, `BROKEN_REFERENCE`).

We are currently gathering objective engineering feedback from teams deploying multi-instance or factory contracts on Stellar:

1. **Discovery**: How does your team currently keep track of all contract instances deployed by your protocol that share a base implementation?
2. **Verification**: When executing an upgrade or updating a shared executable reference, what process do you use to verify that every active instance resolves to the expected bytecode hash?
3. **Tooling Gap**: Do you maintain your own custom indexer/scripts for this, or would a read-only, open-source verification API and CLI add value to your release process?

Our code is open source under Apache-2.0, with an early v1.0.0 release and documentation available here:
- Repository: https://github.com/soroban-fleet/soroban-fleet-registry
- Documentation: https://soroban-fleet.github.io/soroban-fleet-registry/

We are strictly seeking honest technical feedback on whether this operational problem aligns with your current workflow—not seeking endorsements or promotion. If you're open to sharing your thoughts, please let us know if we may reference your feedback in our open-source notes.

Thank you for your time and insights,

[Maintainer Name / Organization]
Soroban Fleet Registry
