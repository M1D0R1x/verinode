# Verinode — Product Requirements / Feature Spec

*Companion to `01-architecture-and-getting-started.md`. Covers all phases (0 concierge pilot → 4 optional rails), not just v1, so the backlog has an end state to build toward.*

## 1. Personas

| Persona | Wants | Primary surfaces |
|---|---|---|
| **Buyer** (AI startup, enterprise, inference platform) | Reserve a specific, verified GPU node for a defined window, with recourse if it fails | Web app: RFQ creation, contract review, delivery status, claims |
| **Seller** (GPU cloud, neo-cloud, hyperscaler reseller) | List capacity, get pre-sold utilization, get paid reliably | Web app: inventory listing, quote response, telemetry agent install |
| **Ops/Admin** (internal, human-in-the-loop desk) | Broker RFQs manually early on, approve participants, resolve claims | Admin console |
| **Compliance/Risk** | Enforce KYB/KYC, sanctions, credit limits, and the "no retail leverage / no transferability" rules that protect the forward-exclusion legal position | Admin console, surveillance dashboard |
| **Index contributor** (later phase) | Submit trade/quote data toward a published series | Contributor portal |
| **Data licensee / exchange partner** (later phase) | Consume the published index or delivery-performance data | Public index API, licensing agreements |

## 2. Feature set by phase

### Phase 0 — Concierge pilot (0–6 weeks, mostly non-engineering)
No production trading system yet. Engineering deliverables are foundational only.

| Feature | Description | Acceptance criteria |
|---|---|---|
| Grade schema v0.1 | Define the single initial delivery grade (8x H100 SXM/HGX, 168h) as structured data, not prose | Grade is representable as a typed record every later service can reference |
| Event model prototype | Draft the append-only event shape used by every service later (actor, timestamp, reason, idempotency key, prior state, evidence hash, authorization decision) | A sample RFQ→contract→delivery flow can be represented as an ordered event log |
| Telemetry prototype | Minimal script that reads GPU/topology info from a test host and produces a signed report | Signed report validates against a known public key |
| Contract repository | Real (not spreadsheet) storage for the 10 manually-brokered pilot deals, with controlled access | Every pilot contract + evidence is retrievable with an audit trail of who viewed/edited it |

**Out of scope for Phase 0:** automated matching, payments UI, public index, on-chain anything.

### Phase 1 — Controlled marketplace (6–16 weeks)
This is the core product. Every feature below maps to a service in the architecture doc.

| Feature | User story | Acceptance criteria |
|---|---|---|
| Participant onboarding | As a buyer/seller, I register my legal entity and complete KYB/KYC so I can trade | Entity cannot submit/receive RFQs until KYC status = `approved`; sanctions screening result stored with timestamp |
| Credit/permissions | As compliance, I set a credit limit per buyer entity | RFQ creation is blocked if it would exceed the buyer's available credit |
| Inventory listing | As a seller, I list a capacity block against the grade schema with real availability | A listed block cannot be double-booked; availability updates are atomic |
| RFQ creation | As a buyer, I submit a private RFQ specifying grade, tenor, region bucket, and window | RFQ is visible only to invited/matched sellers, not broadcast publicly |
| Quoting | As a seller, I respond to an RFQ with a firm price and expiry | Expired quotes cannot be accepted; only one quote per seller per RFQ is active at a time |
| Accept & contract generation | As a buyer, I accept a quote and get a binding confirmation under the master agreement | Confirmation is generated from a versioned template; every required field from the delivery-grade spec is populated, none free-text |
| E-signature | As both parties, I sign the confirmation | Contract cannot enter `FUNDED` state until both signatures are recorded, with a hash of the final signed PDF stored |
| Deposit/invoice | As a buyer, I pay the deposit; as a seller, I see funds are secured before committing capacity | Ledger entry is double-entry balanced; contract cannot move to `SCHEDULED` until funding confirms against the bank/escrow statement |
| Telemetry agent registration | As a seller, I install the agent on the delivered host so it can be verified | Agent identity is bound to a registered supplier identity and a per-host signing key |
| Attestation & health checks | As the platform, I verify the delivered node matches the contracted grade before calling it "delivered" | Delivery only reaches `LIVE` when: credentials work, scheduler shows reserved capacity, attestation matches expected hardware, health tests pass, canary container completes |
| Delivery failure handling | As a buyer, if the node fails acceptance tests, I get a remedy per the contract | System correctly routes to `CURE`, `SUBSTITUTED`, or a cash remedy per the replacement hierarchy — never silent failure |
| Claims/disputes | As either party, I can open a claim with evidence if delivery or performance falls short | Claim carries an evidence bundle (attestation history, health-test logs) sufficient for independent review/arbitration |
| Admin console (core) | As ops, I can approve participants, review flagged claims, and manage the grade ontology | Every admin action is itself an audited event |
| Full audit log | As compliance, I can reconstruct any contract's full history | Every state transition is replayable from immutable events; no field can be edited without a new event |

**Non-functional requirements for Phase 1:** SOC 2 readiness begins; RBAC/ABAC enforced on every admin action; no customer workload content ever enters telemetry.

**Phase 1 exit KPIs (in addition to the design-partner advance gate — at least 5 buyers and 5 sellers showing recurring demand/willingness to sign, per `06-phased-task-backlog.md` Phase 0 exit gate):** 20+ completed reservation blocks; on-time delivery start rate above 98%; at least one repeat buyer; positive contribution margin after make-good costs. No leverage, no bearer-token collateral, no cash settlement anywhere in this phase — bank/invoice only.

### Phase 2 — Index/data (months 4–8)

| Feature | User story | Acceptance criteria |
|---|---|---|
| Benchmark database | As the index team, I need trade/quote data isolated from live trading permissions | Index service reads from a scoped, validated dataset — never direct write access to core trading tables |
| Contributor portal | As a data contributor, I submit trades/quotes toward a series | Submissions are tagged by input-hierarchy tier (completed verified trade > matched trade > firm two-sided quote > firm one-sided quote > public list price) |
| Index calculation engine | As the platform, I calculate a series using volume-weighted/robust median, never a simple mean | Calculation enforces contributor caps, venue caps, related-party aggregation, and minimum-contributor/minimum-notional thresholds |
| Insufficient-data handling | As a consumer of the index, I should never see a fabricated number | If thresholds aren't met, the API returns "insufficient data," and any prior value shown as stale is explicitly labeled stale, never silently carried forward |
| Signed publication | As a data licensee, I need to trust the published fix | Publication requires two-person release or threshold signatures; signed JSON/CSV over HTTPS is the canonical artifact |
| Shadow index mode | As the platform, I want to validate the index before anyone relies on it | A full observation cycle runs in shadow (computed, not published) before any public release |
| Surveillance | As compliance, I need to catch wash trades and self-dealing before they poison the index | System flags related-party clustering, circular rebates, spoofing, and end-window marking for human review before inclusion |

**Gate to exit Phase 2:** don't call anything a "benchmark" until governance (conflict register, contributor code of conduct, corrections policy, versioned methodology) is actually in place — this is a product gate, not just a legal one.

### Phase 3 — Scale and integrations (months 7–12)

| Feature | User story | Acceptance criteria |
|---|---|---|
| ERP/API integrations | As an enterprise buyer, I want RFQs and invoices to flow into my procurement system | Programmatic API access (API keys, not just SSO sessions) with rate limits and idempotency keys enforced |
| Automated credit limits | As compliance, I want credit limits to adjust based on payment history without manual review for every case | Automated adjustments are bounded and every change is logged with the triggering data |
| Replacement routing automation | As a buyer whose node fails, I want an automatic substitute offer instead of a manual scramble | System proposes a same-grade or buyer-approved-superior-grade replacement within the cure window automatically when inventory allows |
| Multi-region resiliency | As the platform, I need to survive a regional outage without losing contract state | Documented RTO/RPO targets met in a failover drill |
| Additional grades (H200/B200) | As the business, I only add a grade once it can sustain real liquidity | New grade launch is gated on the same liquidity thresholds used for the index (below) |
| Exchange/data licensing | As a partner exchange or lender, I consume Verinode's delivery-performance data or index | Licensing API is versioned and access-scoped separately from the trading API |

### Phase 4 — Optional rails (only after customer demand + legal/custody clearance)

See the dedicated **`05-onchain-optional-rails.md`** doc for full detail. Summary:

| Feature | Trigger to build | Non-negotiable constraint |
|---|---|---|
| Stablecoin escrow | Customer explicitly requests it, and legal/custody review clears it | Never the default settlement path; bank/invoice remains primary |
| On-chain publication adapter | Only after the off-chain signed index is stable and governed | Adapter mirrors the signed off-chain fix; it is never itself the authoritative record |
| Regulated derivatives partnership | Only via a partner exchange/FCM | Verinode does not build or operate a cash-settled derivative itself |

## 3. Explicit non-goals (all phases)

- **Don't conflate the four distinct GPU-compute offerings.** Verinode builds exactly one of them — the physical reservation/forward. Spot marketplaces (AWS/GCP/Azure spot, neo-clouds) and cash-settled financial derivatives (CME/ICE/Nodal) are explicitly out of scope; enterprise bilateral contracts are supported only as the legal wrapper around a reservation, not as a separate bespoke-everything product line. See `01-architecture-and-getting-started.md` §1.2 for the full boundary table.
- No continuously tradable retail contract, no order book (RFQ only).
- No on-chain perpetual, no synthetic/leveraged retail product.
- No automatic cash netting for ordinary performance issues (that would undercut the physical-forward legal position).
- No secondary trading of confirmations until counsel and market structure support it.
- No mixed-card/"H100 equivalent" substitutions — the grade is exact or it isn't delivered.

## 4. Non-functional requirements (cross-cutting)

- **Auditability:** every state transition immutable and replayable, across all phases.
- **Data isolation:** telemetry/index/surveillance kept in separate trust domains from trading and money movement, enforced by permissions, not convention.
- **No fabricated data:** applies to telemetry (no location claims without corroboration) and to the index (no numbers below data-sufficiency thresholds).
- **Security:** least-privilege, hardware-backed production credentials, signed builds/SBOM, external pen test before enterprise production (see architecture doc §2.4 / security release gates).
