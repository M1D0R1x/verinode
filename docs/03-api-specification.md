# Verinode — API Specification

*External REST surface (client-facing, OpenAPI-schema'd) plus a note on the internal gRPC boundary. Covers Phase 1 core trading API through Phase 3 integration/licensing endpoints. Phase 4 (optional rails) endpoints are called out separately since they're gated.*

## 1. Conventions

- **Base URL:** `https://api.verinode.com/v1/...` (internal-only endpoints under `/internal/...`, not exposed externally)
- **Auth:** OIDC/SAML-issued bearer token for interactive users; scoped API keys for programmatic/ERP access (Phase 3). Every mutating request from an integration must pass an `Idempotency-Key` header.
- **Format:** JSON over HTTPS; all timestamps UTC ISO-8601; all money fields as integer minor units (cents) + ISO currency code, never floats.
- **Errors:** standard problem-details shape — `{ "type": "...", "title": "...", "status": 4xx/5xx, "detail": "...", "trace_id": "..." }`.
- **Versioning:** path-versioned (`/v1`); breaking changes require a new version, additive fields are safe.
- **Every state-changing endpoint** writes an event carrying actor, timestamp, reason, idempotency key, prior state, evidence hash — per the architecture doc's state-machine convention. This is enforced at the service layer, not just documented.
- **`trade_id` is the canonical cross-service key.** It equals the `contracts.id` created at acceptance and is echoed in every response from Contract, Ledger, Telemetry, Delivery/Claims, and Index-ingestion endpoints related to that trade (see architecture doc §1.8, "Canonical trade envelope"). Use it, not ad hoc joins, when reconciling state across services.

## 2. Participant Service

| Method | Path | Purpose | Key request fields | Key response fields |
|---|---|---|---|---|
| POST | `/participants` | Register a legal entity | `legal_name`, `jurisdiction`, `role` (buyer/seller/both), `beneficial_owners[]` | `id`, `kyc_status: pending` |
| GET | `/participants/{id}` | Fetch entity | — | full profile incl. `kyc_status`, `credit_limit` |
| PATCH | `/participants/{id}` | Update entity details | any mutable field | updated profile |
| POST | `/participants/{id}/kyc-check` | Trigger/refresh KYB/KYC + sanctions screening (proxies vendor) | — | `kyc_status`, `sanctions_result`, `checked_at` |
| GET | `/participants/{id}/credit-limit` | Read current credit exposure | — | `limit`, `used`, `available` |
| PATCH | `/internal/participants/{id}/credit-limit` | Compliance sets/adjusts limit (internal only) | `limit`, `reason` | updated limit + audit event id |

## 3. Inventory Service

| Method | Path | Purpose | Key request fields | Key response fields |
|---|---|---|---|---|
| POST | `/inventory/blocks` | Seller lists a capacity block | `grade_id`, `region_bucket`, `start`, `end`, `facility_attested: bool` | `id`, `status: available` |
| GET | `/inventory/blocks` | Search available inventory | query: `grade_id`, `region_bucket`, `start`, `end` | list of blocks |
| GET | `/inventory/blocks/{id}` | Fetch a block | — | full block incl. evidence refs |
| PATCH | `/inventory/blocks/{id}/availability` | Update availability (e.g., reserved, released) | `status`, `reason` | updated block |
| GET | `/grades` | List supported delivery grades | — | grade specs (SKU, topology, benchmark floor) |

## 4. RFQ / Matching Service

| Method | Path | Purpose | Key request fields | Key response fields |
|---|---|---|---|---|
| POST | `/rfqs` | Buyer creates a private RFQ | `grade_id`, `tenor`, `region_bucket`, `window`, `invited_seller_ids[]` | `id`, `status: open` |
| GET | `/rfqs/{id}` | Fetch RFQ (buyer or invited seller only) | — | RFQ + quotes visible to caller |
| POST | `/rfqs/{id}/quotes` | Seller submits a quote | `price`, `currency`, `expires_at`, `block_id` | `quote_id`, `status: quoted` |
| POST | `/rfqs/{id}/quotes/{quoteId}/accept` | Buyer accepts a quote | `idempotency_key` | `status: accepted`, triggers contract generation |
| POST | `/rfqs/{id}/cancel` | Either party cancels before acceptance | `reason` | `status: cancelled` |

## 5. Contract Service

| Method | Path | Purpose | Key request fields | Key response fields |
|---|---|---|---|---|
| POST | `/contracts` | System-generated from an accepted quote (not typically called directly by clients) | `rfq_id`, `quote_id` | `contract_id`, `status: contract_pending` |
| GET | `/contracts/{id}` | Fetch contract + current state | — | full confirmation, `state`, signature status |
| GET | `/contracts/{id}/state` | Lightweight state check | — | `state` enum value only |
| POST | `/internal/contracts/{id}/sign-callback` | Webhook from e-signature provider | provider payload | records signature event, advances state when both parties signed |

**Contract state enum** (mirrors architecture doc):
`DRAFT_RFQ → OPEN → QUOTED → ACCEPTED → CONTRACT_PENDING → FUNDED/SECURED → SCHEDULED → DELIVERY_TEST → LIVE → COMPLETED → SETTLED`, plus `CANCELLED | FAILED_DELIVERY | CURE | SUBSTITUTED | CLAIM_OPEN | DISPUTED | TERMINATED`.

## 6. Collateral / Payment Ledger

| Method | Path | Purpose | Key request fields | Key response fields |
|---|---|---|---|---|
| POST | `/ledger/deposits` | Record a buyer deposit against a contract | `contract_id`, `amount`, `bank_reference` | ledger entry id, updated contract funding status |
| GET | `/ledger/accounts/{participantId}` | View a participant's ledger position | — | balances, pending settlements |
| POST | `/ledger/invoices` | Issue an invoice | `contract_id`, `line_items[]` | `invoice_id`, `status` |
| POST | `/ledger/settlements` | Finalize settlement at contract completion | `contract_id` | settlement record, SLA credits applied if any |

*All ledger writes are double-entry; every endpoint above produces two balanced entries, never one.*

## 7. Telemetry Gateway

| Method | Path | Purpose | Key request fields | Key response fields |
|---|---|---|---|---|
| POST | `/telemetry/agents/register` | Register a supplier's host agent | `supplier_id`, `host_public_key` | `agent_id`, registration status |
| POST | `/telemetry/attestations` | Agent submits a signed hardware/topology report | signed payload (PCI IDs, memory, NVLink topology, driver versions) | accepted/rejected + validation errors |
| POST | `/telemetry/canary-results` | Agent submits standardized benchmark/canary results | signed payload | pass/fail vs. grade's benchmark floor |
| GET | `/telemetry/nodes/{blockId}/health` | Current health/heartbeat status of a delivered block | — | last heartbeat, health summary |

*This gateway never accepts customer workload content — only host/hardware telemetry.*

## 8. Delivery / Claims Engine

| Method | Path | Purpose | Key request fields | Key response fields |
|---|---|---|---|---|
| POST | `/delivery/{contractId}/accept` | Buyer confirms acceptance after health/canary checks pass | — | `state: LIVE` |
| POST | `/delivery/{contractId}/fail` | Delivery test failed | `reason`, `evidence_refs[]` | `state: FAILED_DELIVERY`, triggers cure/substitution workflow |
| POST | `/claims` | Open a claim (either party) | `contract_id`, `type` (outage/degradation/non-delivery/nonpayment), `evidence_refs[]` | `claim_id`, `state: CLAIM_OPEN` |
| GET | `/claims/{id}` | Fetch claim + evidence bundle | — | full claim record |
| POST | `/internal/claims/{id}/resolve` | Ops/arbitration resolves a claim | `resolution`, `remedy` | `state: resolved`, ledger/contract updates as needed |

## 9. Market Data / Index Service

| Method | Path | Purpose | Key request fields | Key response fields |
|---|---|---|---|---|
| GET | `/index/series` | List published series (e.g. `H100-SXM-8XNV-US-WEEK-DEDICATED-USD`) | — | series metadata, methodology version |
| GET | `/index/series/{seriesId}/observations` | Historical observations | query: date range | array of signed observations |
| GET | `/index/series/{seriesId}/latest` | Latest published fix | — | value, unit, observation window, publish time, sequence number, or `insufficient_data: true` |
| POST | `/internal/index/contribute` | Internal-only — trade/quote data flows into the index dataset from core services, not from external callers | — | — |

*Public-facing index endpoints are read-only. There is no external write path into the index — this is the trust-domain separation from the architecture doc.*

## 10. Surveillance (internal only — admin console)

| Method | Path | Purpose |
|---|---|---|
| GET | `/internal/surveillance/flags` | List flagged trades/patterns pending review |
| POST | `/internal/surveillance/flags/{id}/review` | Record a human review decision (cleared / escalated / excluded from index) |

## 11. Reporting

| Method | Path | Purpose |
|---|---|---|
| GET | `/reports/confirmations/{contractId}` | Download signed confirmation PDF |
| GET | `/reports/invoices/{id}` | Download invoice |
| GET | `/reports/sla/{contractId}` | SLA performance report |
| GET | `/reports/audit-export` | Full audit trail export (scoped by role/date range) |

## 12. Phase 3 — Integration/licensing additions

| Method | Path | Purpose |
|---|---|---|
| POST | `/integrations/api-keys` | Enterprise buyer provisions a scoped API key for ERP integration |
| GET | `/licensing/index-feed` | Licensed partner pulls the index feed under a separate scoped agreement |
| GET | `/licensing/delivery-performance` | Licensed partner (exchange, lender) pulls aggregated delivery-performance data |

## 13. Internal gRPC boundary (service-to-service only)

Not externally exposed. Each core service exposes a Protobuf-defined gRPC interface mirroring its REST resource model (e.g., `ContractService.GetContract`, `LedgerService.RecordDeposit`) for use by Temporal workflow activities and other internal services. Keep REST (external) and gRPC (internal) schemas in sync via a shared source-of-truth schema definition, not maintained by hand in two places.
