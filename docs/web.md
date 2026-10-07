# Soroban Fleet Registry — Web Application Development Guide

This guide covers the architecture, environment configuration, local development workflows, testing, production builds, and operational assumptions for the `soroban-fleet-registry` web application located in `web/`.

---

## 1. Architectural Principles & Boundaries

The `web/` application provides the human-facing operational interface for discovering, inspecting, and verifying Soroban CAP-85 externally managed contract executable fleets.

### Strict Read-Only Boundary
- **Observer Only**: The application queries data indexed and verified by the backend indexer API.
- **No Private Keys & No Signing**: The frontend never holds keys, requests wallet connections, prompts signatures, or sends Stellar transactions.
- **Backend as Authoritative Source of Truth**: The application strictly consumes the REST API defined in [`api/openapi.yaml`](file:///home/smog/soroban-fleet/soroban-fleet-registry/api/openapi.yaml). It does not query PostgreSQL directly, query Stellar RPC directly, or re-implement ledger verification logic in TypeScript.
- **Protocol Fidelity**: The verification engine produces 5 exact states:
  - `HEALTHY`: 100% of member contracts resolve to the current tag WASM.
  - `DRIFT`: One or more members reference different WASM hashes.
  - `BROKEN_REFERENCE`: The tag is missing from the owner contract instance storage, or an entry is malformed.
  - `INCOMPLETE`: Indexer lag or unconfirmed ledger state prevents full verification; never displayed as healthy.
  - `UNKNOWN`: Insufficient data to determine state.
  The frontend reflects these statuses verbatim.

---

## 2. Technology Stack

- **Framework**: Next.js 14 (App Router)
- **Language**: TypeScript 5.6+ with strict type checking enabled
- **UI & Components**: React 18 with CSS modules and standard responsive semantic HTML
- **Testing**: Vitest with `@testing-library/react` and JSDOM environment
- **Tooling**: ESLint 8 (Next.js config), PostCSS

---

## 3. Directory Layout

```text
web/
├── app/
│   ├── layout.tsx                     # Root layout with Header, Footer, and navigation
│   ├── page.tsx                       # Dashboard with system health and quick links
│   ├── fleets/
│   │   ├── page.tsx                   # Fleet directory with search & pagination
│   │   └── [owner]/[tag]/
│   │       ├── page.tsx               # Fleet detail overview
│   │       ├── members/page.tsx       # Paginated fleet members and divergence flags
│   │       ├── history/page.tsx       # Timeline of WASM hash upgrades & releases
│   │       └── verify/page.tsx        # Verification view with breakdown & drift inspection
│   ├── contracts/
│   │   └── [contractId]/page.tsx      # Contract detail with fleet membership & drift status
│   ├── not-found.tsx                  # 404 handler
│   └── globals.css                    # Design system styling & responsive tokens
├── components/
│   ├── layout/                        # Header, Footer, Navigation
│   ├── fleet/                         # FleetCard, FleetHeader, FleetStatus, FleetStats,
│   │                                  # FleetMembersTable, FleetHistory, VerificationPanel
│   └── common/                        # AddressDisplay, HashDisplay, LedgerLink,
│                                      # LoadingState, ErrorState, EmptyState
├── lib/
│   ├── api/                           # Typed API client (client.ts, fleets.ts, contracts.ts)
│   ├── types/index.ts                 # Full TypeScript schemas matching openapi.yaml
│   ├── formatting/index.ts            # Stellar address/hash formatters & relative time
│   └── config.ts                      # Runtime configuration & environment variables
└── __tests__/                         # Unit, component, and end-to-end acceptance tests
```

---

## 4. Environment Configuration

The application is configured using environment variables. Default fallbacks are provided for local development.

| Variable | Default Value | Description |
|---|---|---|
| `NEXT_PUBLIC_SFR_API_URL` | `http://127.0.0.1:8080/api/v1` | URL to the SFR backend API root. Client and SSR requests target this endpoint. |
| `NEXT_PUBLIC_EXPLORER_BASE_URL` | `https://stellar.expert/explorer/testnet` | Base URL used by `LedgerLink` to link contract addresses, transactions, and ledgers to a block explorer. |

To override these in development, create a `.env.local` file inside `web/`:

```env
NEXT_PUBLIC_SFR_API_URL=http://localhost:8080/api/v1
NEXT_PUBLIC_EXPLORER_BASE_URL=https://stellar.expert/explorer/testnet
```

---

## 5. Local Development Workflow

### Prerequisites
- Node.js 18.x or newer (recommended: Node.js 24)
- npm 10.x or newer
- Running backend API instance (e.g. `sfr api` on port 8080 or docker-compose)

### Setup & Execution
1. Install dependencies:
   ```bash
   cd web
   npm install
   ```

2. Run the development server:
   ```bash
   npm run dev
   ```
   The application will be accessible at [http://localhost:3000](http://localhost:3000).

3. Fast code formatting & type check:
   ```bash
   npx tsc --noEmit
   ```

---

## 6. Testing Strategy

The web application maintains automated tests using Vitest and Testing Library:

- **Unit Tests**:
  - `formatting.test.ts`: Stellar StrKey address truncation and 32-byte hex WASM hash normalization.
  - `api_client.test.ts`: Error handling, parameter serialization, and typed response unmarshaling.
- **Component Tests**:
  - `components.test.tsx`: Common widgets (`AddressDisplay`, `HashDisplay`, `LedgerLink`, `FleetStatus`).
  - `fleet_list.test.tsx`, `fleet_detail.test.tsx`, `fleet_members.test.tsx`, `fleet_history.test.tsx`, `fleet_verify.test.tsx`: Subview rendering and interactive states.
  - `contract_detail.test.tsx`: Contract inspection, drift badge rendering, and fleet reference links.
- **Acceptance Tests**:
  - `acceptance.test.tsx`: High-level user journeys verifying:
    1. Healthy fleet inspection workflow.
    2. Drifted fleet alert, member divergence identification, and navigation to the members subview.
    3. Incomplete verification state fidelity (ensuring indexer lag is visibly communicated and never rendered as healthy).
    4. Contract detail resolution and parent fleet linking.

Run the test suite:
```bash
cd web
npm test
```

---

## 7. Production Build & Deployment

To create an optimized production build:

```bash
cd web
npm run build
```

To run the production server:

```bash
npm run start
```

### Deployment Considerations
- **Stateless Container**: The built Next.js application is fully stateless and can be containerized or hosted behind reverse proxies (Nginx, Caddy, Cloudflare).
- **Runtime API Connection**: Ensure `NEXT_PUBLIC_SFR_API_URL` points to the public or reverse-proxied URL of the `sfr` API server.
