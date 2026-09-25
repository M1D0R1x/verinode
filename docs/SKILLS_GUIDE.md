# Verinode — Skills Guide & Applied Engineering Workflows

*Comprehensive catalog of all specialized skills installed in `.agents/skills/` (and `.skills/`), organized by domain with direct application patterns to Verinode.*

---

## 1. Complete Catalog of Repository Skills

### A. Core Backend & Database Skills (Go 1.22+ & PostgreSQL)
| Skill | Primary Focus | Application in Verinode |
|---|---|---|
| **`golang-project-layout`** | Standard Go package structure | Structuring `services/internal/{contract,ledger,rfq,participant,inventory,telemetry}` to prevent cyclic dependencies. |
| **`golang-error-handling`** | Sentinel errors & error wrapping | RFC 7807 problem details in API Gateway, structured errors for invalid state hops and double-entry imbalances. |
| **`golang-database`** | Transactions & connection pooling | PostgreSQL transactional outbox pattern and atomic double-entry balance insertions (`SUM(amount_cents) == 0`). |
| **`golang-concurrency`** | Goroutines, channels, worker pools | High-throughput telemetry ingestion pipeline and outbox background publisher workers. |
| **`golang-testing`** | Table-driven tests & race detection | Testing all legal and illegal contract state transitions with `go test -race ./...`. |
| **`golang-code-style`** | Clean naming & memory patterns | Enforcing idiomatic Go standards across all core microservices. |

### B. Frontend & UI Engineering Skills (Next.js 14+ & Tailwind)
| Skill | Primary Focus | Application in Verinode |
|---|---|---|
| **`nextjs-app-router-patterns`** | App Router routing & data fetching | Layout hierarchy for Buyer Portal (`/buyer`), Seller Portal (`/seller`), and Admin Desk (`/admin`). |
| **`vercel-react-best-practices`** | Performance & Server Components | Zero-waterfall data fetching, client-server component boundaries, fast SSR for market data. |
| **`tailwind-design-system`** | Tokenized CSS & responsive design | Clean design system tokens for institutional compute dashboard styling. |
| **`shadcn-ui`** | Reusable accessible UI primitives | Building data tables, modals, RFQ comparison drawers, and contract sign sheets. |

### C. Phase 4 Blockchain & Smart Contracts (Gated Rails)
| Skill | Primary Focus | Application in Verinode |
|---|---|---|
| **`solana-dev`** *(Solana Foundation)* | Rust & Anchor smart contracts | Writing the optional Phase 4 stablecoin escrow program with isolated per-trade PDAs. |
| **`solana-vulnerability-scanner`** *(Trail of Bits)* | Solana smart contract security | Static analysis, account ownership validation, and PDA collision prevention for the escrow program. |
| **`solidity-security`** | EVM smart contract vulnerabilities | Reentrancy, access control, and integer overflow audits for optional Arbitrum / EVM escrow adapters. |
| **`web3-testing`** | Foundry, Hardhat & mock RPC testing | Writing invariant tests and property-based test suites for smart contracts. |
| **`hyperliquid`** | Hyperliquid L1 & SDK integrations | Evaluating hedge venue integrations and read-only pricing feeds for Phase 4 secondary market research. |

---

## 2. Practical Execution Workflows

### Scenario 1: Building the Next.js Buyer RFQ Dashboard
1. Load **`nextjs-app-router-patterns`**: Set up route groups `app/(buyer)/rfqs/page.tsx` and `app/(buyer)/rfqs/[id]/page.tsx`.
2. Load **`shadcn-ui`** & **`tailwind-design-system`**: Generate accessible tables for listing quotes with firm pricing in cents and expiry countdowns.
3. Apply **`vercel-react-best-practices`**: Use Server Actions or React Query mutations with optimistic UI updates for quote acceptance.

### Scenario 2: Developing the Go Double-Entry Ledger
1. Load **`golang-database`**: Ensure all ledger entries within a `transaction_id` execute in a single `tx.BeginTx` with serializable or read-committed isolation.
2. Load **`golang-testing`**: Write table-driven unit tests verifying that unbalanced transactions (`debits != credits`) fail at the database constraint and application layer.
3. Load **`golang-error-handling`**: Return typed `ErrLedgerUnbalanced` with debit/credit details.

### Scenario 3: Developing the Phase 4 Solana Escrow (When Triggered)
1. Load **`solana-dev`**: Scaffold an Anchor program with instructions: `initialize_escrow`, `deposit_funds`, `release_funds`, `dispute_funds`.
2. Verify with **`solana-vulnerability-scanner`**: Audit signer verification and assert that funds can only be released to the buyer/seller designated in the PDA seed (`trade_id`).
3. Enforce **`AGENTS.md` Invariant 6**: Confirm that off-chain Postgres remains the source of truth, and the Solana program only mirrors the confirmed state transition.
