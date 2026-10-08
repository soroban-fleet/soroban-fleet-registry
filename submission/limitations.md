# System Limitations

Soroban Fleet Registry maintains clear operational boundaries to ensure deterministic safety and avoid overstating capabilities:

1. **Read-Only Operation**: V1 does not manage private keys, connect to wallets, sign transactions, deploy contracts, or initiate contract upgrades. All administrative state mutations must occur directly via authorized on-chain transactions submitted to Stellar.
2. **Dependence on Stellar Ingestion Sources**: Verification freshness and member discovery rely strictly on accessible Stellar JSON-RPC and ledger data sources. If the indexer is paused or behind the network's current ledger sequence, fleet verification will report `INCOMPLETE` rather than asserting healthy status.
3. **No General-Purpose Block Explorer**: The service purposefully does not index arbitrary Stellar transactions, payment operations, token balances, or unrelated contract invocations. It strictly indices CAP-85 externally managed contract references and owner executable tag entries.
4. **No Financial Economics**: The project possesses no native token, staking pools, protocol fees, or custom smart contracts. It is an open-source indexing and verification service.
