# Verinode — Skills Guide & Applied Engineering Workflows

*Comprehensive guide to specialized skills installed in this repository and how to apply them across backend, frontend, security, and architecture workstreams.*

---

## 1. Newly Installed Skills (`.agents/skills/` & `.skills/`)

The following specialized engineering skills have been installed directly into the repository and configured for Antigravity:

| Skill | Source | Primary Purpose | How to Apply to Verinode |
|---|---|---|---|
| **`golang-project-layout`** | `samber/cc-skills-golang` | Idiomatic Go package structure and domain separation | Use when structuring `services/internal/{contract,ledger,rfq,telemetry}` to prevent cyclic imports and keep domain models clean. |
| **`golang-error-handling`** | `samber/cc-skills-golang` | Custom error types, wrapping (`fmt.Errorf("%w")`), sentinel errors | Use for standardizing problem-details error responses in the API Gateway and bubbling ledger failure states. |
| **`golang-database`** | `samber/cc-skills-golang` | SQL transactions, connection pooling, prepared statements, migrations | Essential for implementing the PostgreSQL transactional outbox pattern and atomic double-entry ledger debit/credit batches. |
| **`golang-concurrency`** | `samber/cc-skills-golang` | Goroutines, channels, worker pools, select, mutexes | Use for the high-throughput Telemetry Gateway ingestion pipeline and background event outbox polling workers. |
| **`golang-testing`** | `samber/cc-skills-golang` | Table-driven tests, subtests, race detector, mocking | Use to write table-driven test suites asserting all 18 valid contract transitions and asserting all illegal transitions fail. |
| **`golang-code-style`** | `samber/cc-skills-golang` | Idiomatic Go naming, interface design, slice allocations | Enforces clean, production-grade Go code style across all `services/` code. |
| **`vercel-react-best-practices`** | `vercel-labs/agent-skills` | Next.js 14 App Router, Server Components, client boundaries, memoization | Use when building the Buyer and Seller web portals in `apps/web` to avoid unnecessary client re-renders and waterfalls. |

---

## 2. Globally Available High-Impact Skills

These built-in and global Antigravity skills should be referenced during specific development phases:

### Architecture & Domain Design
* **`domain-modeling`**:
  * **When to use:** Whenever updating entities or writing Architecture Decision Records (ADRs).
  * **Application:** Keeps the Canonical Trade Envelope, Grade specifications, and Ledger account types consistent across Go, Python, and TypeScript.
* **`codebase-design`**:
  * **When to use:** When designing interface seams between internal Go packages.
  * **Application:** Ensures deep module interfaces (simple public APIs concealing rich state logic).
* **`karpathy-guidelines`**:
  * **When to use:** Continuous development guide to prevent LLM over-engineering.
  * **Application:** Encourages small, surgical, verifiable diffs and avoids premature abstraction layers.

### Quality, Testing & Security
* **`tdd` (Test-Driven Development)**:
  * **When to use:** Developing the contract state machine and double-entry ledger.
  * **Application:** Write the failing test for an invalid state transition (e.g. `LIVE → CONTRACT_PENDING` must fail) before writing the transition handler.
* **`security-audit` & **`bug-bounty`**:
  * **When to use:** Reviewing API endpoints, authentication middlewares, and telemetry signature verifications.
  * **Application:** Verifies that no IDOR vulnerabilities exist on `/contracts/{id}` or `/rfqs/{id}`, and confirms host telemetry signatures are validated before writing to ClickHouse.
* **`setup-pre-commit`**:
  * **When to use:** Configuring local Git hooks.
  * **Application:** Runs `go vet`, `go fmt`, and `pnpm typecheck` before allowing a commit.

### Frontend Polish & Marketing
* **`emil-design-eng`**:
  * **When to use:** Building the Next.js interactive portals.
  * **Application:** Adds high-grade UI polish, micro-interactions, smooth status badge transitions, and responsive data tables.
* **`landing-page-design`**:
  * **When to use:** Building the public homepage (`verinode.io`).
  * **Application:** Enforces clear typography, conversion layouts, and institutional credibility for enterprise buyers and cloud suppliers.

---

## 3. Concrete Applied Workflows

### Scenario A: Implementing a New Contract State Transition in Go
1. Consult **`domain-modeling`** and [`docs/04-data-model-schema.md`](04-data-model-schema.md) to confirm the valid prior and new states.
2. Follow **`tdd`** and **`golang-testing`**: Create `contract_state_test.go` with a table-driven test matrix defining allowed and disallowed transitions.
3. Apply **`golang-error-handling`**: Define typed sentinel errors (`ErrInvalidStateTransition`, `ErrDuplicateIdempotencyKey`).
4. Implement the transition using **`golang-database`** inside a single database transaction that atomically updates `contracts` and appends to `contract_events` and `event_outbox`.

### Scenario B: Building the Buyer RFQ Creation Interface in Next.js
1. Follow **`vercel-react-best-practices`**: Place page routing in `apps/web/app/(buyer)/rfqs/new/page.tsx` using Server Components for static layouts and Client Components for form state.
2. Apply **`emil-design-eng`**: Add subtle hover states, real-time tenor calculators (e.g. converting 168 hours to days/weeks), and instant validation feedback.
3. Validate forms with Zod types mapped 1:1 to the OpenAPI specification in [`docs/03-api-specification.md`](03-api-specification.md).

### Scenario C: Reviewing Code Before Merging
1. Run **`karpathy-guidelines`** checklist: Is this change minimal? Does it introduce unnecessary dependencies?
2. Run **`security-audit`**: Does this endpoint verify that the authenticated participant matches the entity ID in the path? Are SQL queries parameterized?
3. Run `go test ./... -race` and `pnpm typecheck`.
