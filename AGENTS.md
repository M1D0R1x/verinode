# AGENTS.md — Verinode Engineering & System Guidelines

*Instructions, architectural invariants, context pointers, and execution rules for autonomous AI coding agents operating in the Verinode repository.*

---

## 1. Project Identity & Non-Negotiable Invariants

**Verinode** is an institutional brokered marketplace for **physically delivered, enterprise GPU-capacity reservations** (initial benchmark grade: `8x NVIDIA H100 SXM/HGX, 168h`), plus the cryptographic telemetry verification and data layer that makes those contracts legally enforceable.

### Hard Invariants (Never Violate)
1. **Physical Forward Exclusion:** Verinode trades *non-transferable, bilateral physical reservations*. It is **NOT** a continuous order book, **NOT** a spot market (AWS/RunPod), and **NOT** a cash-settled synthetic derivative/perpetual (CME/Hyperliquid). Do not introduce tradable tokens, leverage, or automatic cash netting for standard availability.
2. **The Canonical Trade Envelope:** Every transaction across all services is keyed to a shared, immutable `trade_id` (`contracts.id`). Reconciliations between the ledger, telemetry, and index must use `trade_id`, never loose timestamp or entity heuristics.
3. **Double-Entry Balance Invariant:** Every financial transaction in `services/internal/ledger` must produce balanced debit and credit entries such that `SUM(amount_cents) == 0` grouped by `transaction_id`.
4. **Availability vs. Consumption:** Contracts are take-or-pay against hardware *availability* (credentials work, topology matches, canary passes), not usage metering. **Zero customer workload data** (model weights, training code, prompts) may ever be ingested into the Telemetry Gateway.
5. **Data Integrity Over Index Vanity:** The Index Engine must return `insufficient_data: true` when minimum contributor counts or volume thresholds are unmet. Never interpolate or fabricate index prices.
6. **Off-Chain Source of Truth:** On-chain rails (Solana Anchor, EVM) are strictly optional, read-only/mirroring adapters gated to Phase 4. The off-chain PostgreSQL database and executed legal confirmations are the authoritative sources of truth.

---

## 2. Repository Layout & Architecture Boundaries

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

### Trust-Domain Boundary
The `index-service/` operates in a **separate trust domain**. It must never have direct write access to core transactional tables. It consumes data through a read-only database replica or sanitized event outbox streams.

---

## 3. Skill & Context Pointers

When performing tasks in this repository, load and follow these specialized skills:

| Trigger / Task Domain | Relevant Skill Location |
|---|---|
| Designing Go project structure & packages | `.agents/skills/golang-project-layout` & `codebase-design` |
| Writing Go error handling & wrapping | `.agents/skills/golang-error-handling` |
| Writing Go unit/integration tests | `.agents/skills/golang-testing` & `tdd` |
| Writing Go database queries & transactions | `.agents/skills/golang-database` |
| Go concurrency, channels, and locks | `.agents/skills/golang-concurrency` |
| Building Next.js / React UI components | `.agents/skills/vercel-react-best-practices` & `emil-design-eng` |
| Designing institutional landing pages | `landing-page-design` |
| Modeling domain entities, invariants, & ADRs | `domain-modeling` |
| Preventing over-engineering & code sprawl | `karpathy-guidelines` |
| Security auditing & financial vulnerability testing | `security-audit` & `bug-bounty` |

---

## 4. Development Workflow & Commands

### Local Infrastructure
```bash
# Start PostgreSQL, ClickHouse, Redis, MinIO, Temporal
docker compose up -d

# Verify container health
docker compose ps
```

### Go Services
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

### Python Analytics (Index Service)
```bash
cd index-service
python -m venv .venv && source .venv/bin/activate
pip install -e ".[dev]"
pytest
ruff check .
```

---

## 5. Verification & Completion Criteria

Before declaring any coding task complete:
1. **Zero Lint & Compilation Errors:** `go vet ./...` (Go) and `pnpm typecheck` (TypeScript) must pass cleanly.
2. **Invariant Tests:** All state machine transitions and double-entry ledger operations must have unit test assertions verifying both valid paths and rejected invalid paths.
3. **No Drift from Specs:** All table schemas, endpoints, and event structures must strictly align with the specifications in [`docs/`](docs/).
