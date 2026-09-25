# AGENTS.md — Verinode Engineering & System Guidelines

*Instructions, architectural invariants, context pointers, and execution rules for autonomous AI coding agents operating in the Verinode repository.*

---

## 1. Project Identity

**Verinode** is an institutional brokered marketplace for **physically delivered, enterprise GPU-capacity reservations** (initial benchmark grade: `8x NVIDIA H100 SXM/HGX, 168h`), plus the cryptographic telemetry verification and data layer that makes those contracts legally enforceable.

---

## 2. Hard Invariants — Never Violate

These are non-negotiable. Any change touching them requires explicit human sign-off, not just a passing test suite.

1. **Physical forward exclusion.** Verinode trades *non-transferable, bilateral physical reservations*. It is:
    - **NOT** a continuous order book
    - **NOT** a spot market (like AWS or RunPod)
    - **NOT** a cash-settled synthetic derivative/perpetual (like CME or Hyperliquid)

   Do not introduce tradable tokens, leverage, or automatic cash netting for standard availability.

2. **The canonical trade envelope.** Every transaction across all services is keyed to a shared, immutable `trade_id` (`contracts.id`). Reconciliation between the ledger, telemetry, and index must use `trade_id` — never loose timestamp or entity heuristics.

3. **Double-entry balance invariant.** Every financial transaction in `services/internal/ledger` must produce balanced debit and credit entries such that `SUM(amount_cents) == 0`, grouped by `transaction_id`.

4. **Availability vs. consumption.** Contracts are take-or-pay against hardware *availability* (credentials work, topology matches, canary passes) — not usage metering. **Zero customer workload data** (model weights, training code, prompts) may ever be ingested into the Telemetry Gateway.

5. **Data integrity over index vanity.** The Index Engine must return `insufficient_data: true` when minimum contributor counts or volume thresholds are unmet. Never interpolate or fabricate index prices.

6. **Off-chain source of truth.** On-chain rails (Solana Anchor, EVM) are strictly optional, read-only/mirroring adapters gated to Phase 4. The off-chain PostgreSQL database and executed legal confirmations remain the authoritative sources of truth.

---

## 3. Repository Layout

```
verinode/
├── services/               # Go 1.22+ Core Services (Matching, Contract, Ledger, Telemetry)
│   ├── cmd/api-gateway/    # REST / OpenAPI Edge Gateway
│   ├── internal/           # Domain modules (contract, ledger, rfq, participant, inventory)
│   └── migrations/         # PostgreSQL schema migrations
├── apps/web/               # Next.js 14+ Frontend (Buyer, Seller, Admin portals)
├── index-service/          # Python 3.11+ Market Data & Index Engine (Isolated Trust Domain)
├── packages/proto/         # Protobuf definitions for gRPC contracts and cross-service events
├── docs/                   # Authoritative Specifications (PRD, Architecture, Data Model, API)
├── .agents/skills/         # Repository-installed AI agent skills
└── docker-compose.yml      # Local infra (PostgreSQL, ClickHouse, Redis, MinIO, Temporal)
```

### Trust-domain boundary

`index-service/` operates in a **separate trust domain**. It must never have direct write access to core transactional tables — it consumes data only through a read-only database replica or sanitized event-outbox streams.

---

## 4. Skill & Context Pointers

Load and follow the relevant skills below (installed in `.agents/skills/`) before working in a given area of the repository.

### Core backend (Go & data)
| Area | Skill |
|---|---|
| Project layout & codebase design | `golang-project-layout`, `codebase-design` |
| Error handling | `golang-error-handling` |
| Testing & TDD | `golang-testing`, `tdd` |
| Database & outbox pattern | `golang-database` |
| Concurrency & channels | `golang-concurrency` |
| Code style | `golang-code-style` |

### Frontend & UI design (Next.js & React)
| Area | Skill |
|---|---|
| App Router architecture | `nextjs-app-router-patterns` |
| Performance | `vercel-react-best-practices` |
| Design system | `tailwind-design-system` |
| Component design & polish | `shadcn-ui`, `emil-design-eng` |
| Institutional landing pages | `landing-page-design` |

### Phase 4 blockchain & smart contracts *(gated rails — see Invariant 6)*
| Area | Skill |
|---|---|
| Solana & Anchor development | `solana-dev` |
| Solana security audit | `solana-vulnerability-scanner` |
| Solidity / Arbitrum contracts | `solidity-security` |
| Web3 & smart contract testing | `web3-testing` |
| Hyperliquid market & L1 integration | `hyperliquid` |

### Cross-cutting security & process
| Area | Skill |
|---|---|
| Domain modeling | `domain-modeling` |
| Avoiding over-engineering | `karpathy-guidelines` |
| Security & vulnerability auditing | `security-audit`, `bug-bounty` |

---

## 5. Development Workflow & Commands

### Local infrastructure
```bash
# Start PostgreSQL, ClickHouse, Redis, MinIO, Temporal
docker compose up -d

# Verify container health
docker compose ps
```

### Go services
```bash
cd services
go run ./cmd/api-gateway
go test ./... -race -v
go vet ./...
```

### Frontend (Next.js)
```bash
cd apps/web
pnpm install
pnpm dev
pnpm typecheck
```

### Python analytics (index service)
```bash
cd index-service
python -m venv .venv && source .venv/bin/activate
pip install -e ".[dev]"
pytest
ruff check .
```

---

## 6. Verification & Completion Criteria

Before declaring any coding task complete, confirm all of the following:

1. **Zero lint & compilation errors** — `go vet ./...` (Go) and `pnpm typecheck` (TypeScript) both pass cleanly.
2. **Invariant tests pass** — every state-machine transition and double-entry ledger operation has unit test assertions covering both valid paths and rejected invalid paths.
3. **No drift from specs** — all table schemas, endpoints, and event structures strictly match the specifications in [`docs/`](docs/).