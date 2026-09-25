# Verinode — Phased Task Backlog

*Ties together `01-architecture-and-getting-started.md`, `02-prd-feature-spec.md`, `03-api-specification.md`, `04-data-model-schema.md`, and `05-onchain-optional-rails.md` into an ordered build sequence. Grouped by workstream so it can be parallelized across a small team (or an agent working sequentially).*

## Phase 0 — Concierge pilot (0–6 weeks)

**Goal:** validate demand with 10 manually-brokered deliveries. Minimal engineering; the deliverables here are the foundation every later phase builds on.

- [ ] **[Foundations]** Define `grades` schema v0.1 (single grade: 8x H100 SXM/HGX, 168h) per `04-data-model-schema.md` §3
- [ ] **[Foundations]** Draft the event/state-machine shape used everywhere later (actor, timestamp, reason, idempotency key, prior state, evidence hash, authorization decision)
- [ ] **[Telemetry]** Build a minimal signed-attestation prototype: script reads GPU/topology info from a test host, signs it, validates against a public key
- [ ] **[Ops]** Stand up a real contract repository (not a spreadsheet) with access control for the 10 pilot deals
- [ ] **[Research]** Interview 20–30 buyers and 15–20 suppliers; collect redacted contracts, procurement objections, willingness-to-pay
- [ ] **[Legal]** Commission written commodities-counsel memo (US + first launch jurisdiction) before any code that touches money movement ships

**Exit gate:** 10 completed pilot deliveries; at least 5 buyers show recurring demand a cash hedge doesn't solve; at least 5 sellers will sign standardized delivery/telemetry terms.

---

## Phase 1 — Controlled marketplace (6–16 weeks)

This is the bulk of the build. Suggested internal ordering below; workstreams can run in parallel once the schema and event backbone (first block) land.

### 1.1 Platform foundations
- [ ] Postgres schema migration: `participants`, `beneficial_owners`, `grades`, `inventory_blocks`, `rfqs`, `quotes`, `contracts`, `contract_events`, `event_outbox` (per `04-data-model-schema.md`)
- [ ] Event outbox pattern wired up (Postgres-based; defer Kafka/Redpanda)
- [ ] Temporal cluster + base worker skeleton for the RFQ→sign→fund→deliver→claim workflow
- [ ] RBAC/ABAC scaffolding shared across all services

### 1.2 Participant service
- [ ] `POST/GET/PATCH /participants` (per `03-api-specification.md` §2)
- [ ] KYB/KYC vendor integration + sanctions screening proxy endpoint
- [ ] Credit limit read endpoint + internal admin-only set/adjust endpoint
- [ ] Enforce: no RFQ creation until `kyc_status = approved` and within credit limit

### 1.3 Inventory service
- [ ] `POST/GET/PATCH /inventory/blocks`, `GET /grades`
- [ ] Atomic availability updates (no double-booking)

### 1.4 RFQ/matching service
- [ ] `POST /rfqs`, invited-seller visibility model (private, not broadcast)
- [ ] `POST /rfqs/{id}/quotes`, quote expiry enforcement
- [ ] `POST .../quotes/{id}/accept` → triggers contract generation
- [ ] `POST /rfqs/{id}/cancel`

### 1.5 Contract service
- [ ] Versioned confirmation template engine bound to the delivery-grade spec (no free-text substitution fields)
- [ ] E-signature provider integration + signed-callback handler
- [ ] Full state machine implementation with idempotency-key enforcement at the DB layer (unique constraint per `04-data-model-schema.md` §5)
- [ ] State-machine table-driven tests covering every legal transition and rejecting illegal ones

### 1.6 Collateral / payment ledger
- [ ] Double-entry `ledger_accounts` / `ledger_entries` model + invariant check job (every `transaction_id` sums to zero)
- [ ] `POST /ledger/deposits`, `/ledger/invoices`, `/ledger/settlements`
- [ ] Bank/escrow reconciliation job
- [ ] **Security review checkpoint** before this goes anywhere near production — treat as security-critical from day one

### 1.7 Telemetry gateway
- [ ] Agent registration endpoint + per-host key binding
- [ ] Signed attestation ingestion → ClickHouse, with Postgres pointer/summary on `inventory_blocks`
- [ ] Canary/benchmark result ingestion vs. grade's benchmark floor
- [ ] Reject stale/duplicate/malformed signed payloads

### 1.8 Delivery / claims engine
- [ ] Delivery acceptance flow: credentials check → scheduler check → attestation match → health tests → canary pass → `LIVE`
- [ ] Failure routing: `CURE → SUBSTITUTED →` cash remedy, per replacement hierarchy
- [ ] Claims endpoints + evidence bundle assembly
- [ ] Manual human-audit hook in the workflow (not a side process) — combine telemetry with invoice/serial spot checks

### 1.9 Admin console (core)
- [ ] Participant approval queue
- [ ] Claims review queue
- [ ] Grade ontology management
- [ ] Audit trail viewer

### 1.10 Frontend (Next.js)
- [ ] OIDC/SAML SSO integration
- [ ] Buyer flow: RFQ creation → quote comparison → accept → contract review/sign → delivery status → claims
- [ ] Seller flow: inventory listing → quote response → agent install instructions → delivery status
- [ ] Admin views (or separate admin app — decide topology explicitly, see architecture doc §3 suggestions)

### 1.11 Cross-cutting
- [ ] Full audit log verification: every mutation has a corresponding event, nothing editable without one
- [ ] SOC 2 readiness checklist started
- [ ] No customer workload content flows into telemetry (verify with a data-flow review, not just code review)

**Exit gate:** two-sided executable quotes clear after fees; default/collateral requirements don't erase supplier economics; counsel confirms a viable launch structure.

---

## Phase 2 — Index/data (months 4–8)

- [ ] **[Schema]** `index_series`, `index_observations`, `index_contributions` tables in a separately-permissioned schema (per `04-data-model-schema.md` §9)
- [ ] **[Ingestion]** Read-only, scoped pipe from core trading data into the index dataset — no direct write access from Index service back into trading tables
- [ ] **[Contributor portal]** Submission UI/API tagged by input-hierarchy tier (completed trade > matched trade > firm two-sided quote > firm one-sided quote > public list price)
- [ ] **[Calculation engine]** Volume-weighted/robust median calculation; contributor caps; venue caps; related-party aggregation
- [ ] **[Thresholds]** Hard-coded minimum-contributor / minimum-notional / concentration gates — `insufficient_data` is a first-class response, never a fabricated fallback
- [ ] **[Publication]** Two-person release or threshold-signature release flow; signed JSON/CSV artifact as the canonical output (see `05-onchain-optional-rails.md` §3b for what may consume this later)
- [ ] **[Shadow mode]** Run a full observation cycle computed-but-unpublished before any public release
- [ ] **[Surveillance]** Wash-trade, related-party clustering, spoofing, end-window-marking detection wired into the review queue *before* the first real index calculation, not after
- [ ] **[Governance]** Conflict register, contributor code of conduct, corrections policy, versioned methodology doc, IOSCO Principles mapping

**Exit gate:** don't call it a "benchmark" publicly until governance controls above are actually operating, not just documented.

---

## Phase 3 — Scale and integrations (months 7–12)

- [ ] API key provisioning + rate limiting + idempotency enforcement for ERP integrations (`03-api-specification.md` §12)
- [ ] Automated credit-limit adjustment logic (bounded, fully logged)
- [ ] Automated replacement/substitution routing within the cure window
- [ ] Multi-region resiliency: define RTO/RPO targets, then build and drill failover
- [ ] Additional grade (H200/B200) onboarding — gated on the same liquidity thresholds as the index, not launched speculatively
- [ ] Exchange/data licensing API surface, access-scoped separately from the trading API

---

## Phase 4 — Optional rails (gated, not default)

**Do not start any item below without an explicit customer request and legal/custody clearance — see `05-onchain-optional-rails.md` for the full rationale.**

- [ ] Stablecoin escrow integration (only for the requesting customer's collateral flow; bank/invoice remains primary for everyone else)
  - [ ] Minimal Rust/Anchor program, per-trade PDAs, explicit funded→released→disputed→resolved states mirroring the off-chain contract machine
  - [ ] Timeouts on stuck funds; multisig for disputed-fund release
  - [ ] Per-PDA balance caps; identities/terms/telemetry kept off-chain
  - [ ] If referencing a price feed (e.g., stablecoin peg), validate feed identity/ownership/staleness before trusting it
- [ ] On-chain publication adapter for the index (thin mirror of the signed off-chain fix; reject stale/duplicate/wrong-version messages; no cross-chain consensus)
- [ ] Enterprise EVM integration — only if a specific customer requires it
  - [ ] Evaluate permissioned Arbitrum Orbit vs. public Arbitrum; document sequencer/governance/reconciliation burden before committing
- [ ] Hyperliquid HIP-3 hedge-venue evaluation — only after independent index observations, manipulation resistance, and committed market makers exist
  - [ ] Explicitly reject deployment on a thin/self-published index regardless of business pressure to move faster
- [ ] No-bridging enforcement: each adopted rail reconciles independently against the canonical `trade_id` record; add an automated check (or at minimum a documented manual reconciliation process) that flags any attempt to move value directly between two chains
- [ ] Regulated derivatives partnership discussions (Verinode does not build a cash-settled derivative itself)

---

## Ongoing / cross-phase workstreams

- [ ] Security release gates before any production milestone: threat model, no workload content in telemetry, segregated benchmark/market ops, least privilege, signed builds/SBOM, dependency scanning, external pen test, backup/restore drills, key-rotation and oracle-compromise runbooks
- [ ] CI: GitHub Actions, Renovate, Semgrep, CodeQL, Trivy, Syft/Grype running on every PR
- [ ] Observability: OpenTelemetry + Prometheus/Grafana + Sentry + paging wired in from Phase 1, not bolted on later
- [ ] **[Phase 1 hardening, worth doing early]** Model collateral, default, and settlement states in TLA+/PlusCal before finalizing the ledger and contract state machine — cheap way to catch illegal-transition or double-spend-style bugs before they're load-bearing
- [ ] **[Phase 1 hardening]** Property-based tests (proptest-style in Go, or Hypothesis if any Python-side logic needs it) on the double-entry ledger invariant (every `transaction_id` group sums to zero) and on state-machine transition legality — in addition to the table-driven tests in §1.5
- [ ] **[Phase 4 only, if any Solidity/Anchor code is written]** Slither (Solidity static analysis), Foundry invariants, and a bug-bounty program before any chain code handles real customer funds
