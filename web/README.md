# Soroban Fleet Registry — Web Application

The official read-only web application for discovering, inspecting, and verifying Soroban CAP-85 externally managed contract executable fleets.

## Architecture & Principles
- **Read-Only Observer**: Interfaces solely with the Soroban Fleet Registry backend API (`api/openapi.yaml`).
- **No Transaction Signing**: The web application never holds private keys, signs transactions, or executes contract mutations.
- **Protocol Fidelity**: Displays exact backend verification states (`HEALTHY`, `DRIFT`, `BROKEN_REFERENCE`, `INCOMPLETE`, `UNKNOWN`) without reinterpretation.

## Getting Started

### Prerequisites
- Node.js 18+ (tested on Node.js 24)
- npm 10+

### Installation
```bash
npm install
```

### Local Development
```bash
npm run dev
```

### Testing
```bash
npm run test
```

### Production Build & Lint
```bash
npm run lint
npm run build
```
