# Verinode — Data Model / Database Schema

*Primary store: PostgreSQL (transactional/authoritative). Secondary: ClickHouse (telemetry time-series), object storage (documents/evidence). Redis holds no authoritative data — cache/locks only.*

## 1. Core principles

- **Every table that represents contract/trade state has a companion append-only event table.** The row shows current state; the event table shows how it got there. Never delete or hard-update history rows — only append.
- **Money is integer minor units + currency code.** Never `float`/`numeric` without fixed scale for ledger amounts.
- **Foreign keys are real**, not just convention — this system's whole value proposition is auditability, so referential integrity matters more than in a typical CRUD app.
- **Evidence is stored by reference** (object storage key + content hash), not inline blobs in Postgres.

## 2. Participant domain

### `participants`
| Column | Type | Notes |
|---|---|---|
| `id` | uuid PK | |
| `legal_name` | text | |
| `jurisdiction` | text | ISO country code |
| `role` | enum(`buyer`,`seller`,`both`) | |
| `kyc_status` | enum(`pending`,`approved`,`rejected`,`review`) | |
| `sanctions_result` | jsonb | vendor response snapshot |
| `credit_limit_cents` | bigint | nullable until approved |
| `created_at`, `updated_at` | timestamptz | |

### `beneficial_owners`
| Column | Type | Notes |
|---|---|---|
| `id` | uuid PK | |
| `participant_id` | uuid FK → participants | |
| `name` | text | |
| `ownership_pct` | numeric(5,2) | |

### `participant_credit_events` (append-only)
`id, participant_id FK, delta_cents, reason, actor, created_at`

## 3. Inventory domain

### `grades`
| Column | Type | Notes |
|---|---|---|
| `id` | text PK | e.g. `H100-SXM-8XNV` |
| `gpu_sku` | text | exact SKU, no "equivalent" |
| `min_memory_gb` | int | |
| `topology` | text | SXM/HGX + NVLink/NVSwitch spec |
| `min_healthy_gpu_count` | int | |
| `benchmark_floor` | jsonb | NCCL/all-reduce + workload thresholds |
| `min_cpu_cores`, `min_ram_gb`, `min_nvme_perf` | int | |

### `inventory_blocks`
| Column | Type | Notes |
|---|---|---|
| `id` | uuid PK | |
| `seller_id` | uuid FK → participants | |
| `grade_id` | text FK → grades | |
| `region_bucket` | text | broad region; specific facility attested but access-restricted |
| `facility_ref` | text, restricted-access | not exposed on standard reads |
| `window_start`, `window_end` | timestamptz | |
| `status` | enum(`available`,`reserved`,`delivered`,`released`) | |
| `created_at`, `updated_at` | timestamptz | |

## 4. RFQ / trading domain

### `rfqs`
`id PK, buyer_id FK, grade_id FK, region_bucket, window_start, window_end, status enum(open,quoted,accepted,cancelled,expired), created_at`

### `rfq_invited_sellers`
`rfq_id FK, seller_id FK` — composite PK; controls visibility, not a public order book.

### `quotes`
`id PK, rfq_id FK, seller_id FK, block_id FK → inventory_blocks, price_cents, currency, expires_at, status enum(active,accepted,expired,withdrawn), created_at`

## 5. Contract domain

### `contracts`
| Column | Type | Notes |
|---|---|---|
| `id` | uuid PK | |
| `rfq_id`, `quote_id` | FK | |
| `buyer_id`, `seller_id` | FK → participants | |
| `grade_id` | FK → grades | |
| `template_version` | text | versioned confirmation template |
| `state` | enum, see below | current lifecycle state |
| `signed_pdf_hash` | text | sha256 of final signed document |
| `signed_pdf_ref` | text | object storage key |
| `created_at`, `updated_at` | timestamptz | |

**`state` enum:** `draft_rfq, open, quoted, accepted, contract_pending, funded_secured, scheduled, delivery_test, live, completed, settled, cancelled, failed_delivery, cure, substituted, claim_open, disputed, terminated`

### `contract_events` (append-only — this is the authoritative history)
| Column | Type | Notes |
|---|---|---|
| `id` | uuid PK | |
| `contract_id` | FK | |
| `prior_state`, `new_state` | text | |
| `actor` | text | user/service id |
| `reason` | text | |
| `idempotency_key` | text, unique per contract | prevents duplicate transitions |
| `evidence_hash` | text nullable | |
| `authorization_decision` | jsonb nullable | e.g. compliance sign-off details |
| `created_at` | timestamptz | |

## 6. Ledger domain (double-entry)

### `ledger_accounts`
`id PK, participant_id FK, currency, account_type enum(deposit,payable,receivable,collateral)`

### `ledger_entries` (append-only, never updated)
| Column | Type | Notes |
|---|---|---|
| `id` | uuid PK | |
| `transaction_id` | uuid | groups the balanced pair (or more) of entries |
| `account_id` | FK → ledger_accounts | |
| `contract_id` | FK nullable | |
| `amount_cents` | bigint | positive = debit, negative = credit (pick one convention and enforce it) |
| `currency` | text | |
| `bank_reference` | text nullable | |
| `created_at` | timestamptz | |

*Invariant enforced at the application layer and checked by a reconciliation job: every `transaction_id` group sums to zero per currency.*

### `invoices`, `invoice_line_items`, `settlements` — standard supporting tables FK'd to `contracts` and `ledger_entries`.

## 7. Telemetry domain

### `telemetry_agents` (Postgres — identity/registration only)
`id PK, supplier_id FK → participants, host_public_key, registered_at, status enum(active,revoked)`

### `telemetry_attestations` (ClickHouse — high volume time-series)
`agent_id, block_id, timestamp, pci_ids, memory_gb, nvlink_topology, driver_version, signature, sequence_number`

### `canary_results` (ClickHouse)
`agent_id, block_id, timestamp, test_name, pass_bool, metrics jsonb, signature`

*Postgres holds a lightweight pointer/summary (`latest_attestation_ref`, `last_heartbeat_at`) on `inventory_blocks` for fast reads; ClickHouse holds the full raw series.*

## 8. Delivery / claims domain

### `delivery_events` (append-only)
`id PK, contract_id FK, event_type enum(credentials_verified,health_test,canary_result,accepted,failed), evidence_refs jsonb, created_at`

### `claims`
`id PK, contract_id FK, opened_by FK → participants, type enum(outage,degradation,non_delivery,nonpayment), state enum(claim_open,resolved,disputed), created_at, resolved_at`

### `claim_evidence`
`id PK, claim_id FK, evidence_ref, evidence_type, added_by, created_at`

## 9. Index domain (separate schema/permissions — different trust boundary)

### `index_series`
`id PK (e.g. H100-SXM-8XNV-US-WEEK-DEDICATED-USD), gpu_model, form, topology, region_bucket, tenor, tenancy, currency, methodology_version`

### `index_observations` (append-only, signed)
| Column | Type | Notes |
|---|---|---|
| `id` | uuid PK | |
| `series_id` | FK | |
| `value` | numeric | |
| `unit` | text | USD/GPU-hour |
| `observation_window_start`, `observation_window_end` | timestamptz | |
| `publish_time` | timestamptz | |
| `sequence_number` | bigint, monotonic per series | |
| `contributor_count` | int | must meet series minimum or row is `insufficient_data = true` |
| `insufficient_data` | boolean | |
| `signature` | text | |

### `index_contributions` (raw input, restricted access — feeds the calculation, never exposed directly)
`id, series_id, tier enum(completed_trade, matched_trade_pending, firm_two_sided_quote, firm_one_sided_quote, public_list_price), contract_id nullable, price, notional, contributor_id, related_party_flag, created_at`

## 10. Surveillance domain

### `surveillance_flags`
`id PK, subject_type enum(contract,contribution,participant), subject_id, flag_type enum(wash_trade,related_party,concentration,spoofing,end_window_marking), status enum(pending,reviewed), reviewed_by, resolution, created_at`

## 11. Reporting / audit

### `event_outbox` (Phase 1 default; Kafka/Redpanda topics take over this role once volume justifies it)
`id PK, aggregate_type, aggregate_id, event_type, payload jsonb, published boolean, created_at`

*This is the mechanism every domain event above flows through for downstream consumers (index ingestion, reporting exports, notifications) — a single, consistent event backbone rather than ad hoc triggers per table.*

## 12. Canonical trade envelope

`contracts.id` doubles as the system-wide `trade_id`. Every table below that references `contract_id` is a shard of one logical, signed record — this is what the API spec calls the canonical envelope, and what reconciliation (ledger vs. telemetry vs. index vs. any future on-chain mirror) is always checked against:

| Canonical concept | Backing table(s) |
|---|---|
| `ContractSpec` | `contracts` + `grades` |
| `Order/Trade` | `rfqs` + `quotes` |
| `Allocation` | `inventory_blocks` (the reserved row) |
| `DeliveryEvidence` | `delivery_events` + `telemetry_attestations` + `canary_results` |
| `Obligation` | `contracts.state` + open rows in `claims` |
| `SettlementInstruction` | `ledger_entries` + `settlements` |

Never reconcile two of these tables against each other directly (e.g., matching a `ledger_entries` row to a `telemetry_attestations` row by timestamp guessing) — always join through `contract_id`/`trade_id`. This is also the rule that keeps any Phase 4 chain adapter honest: it mirrors this one signed record, it doesn't become a second source of truth.

## 13. Indexing/performance notes

- Index `contracts.state`, `contracts.buyer_id`, `contracts.seller_id` for dashboard queries.
- Index `contract_events.contract_id, created_at` for fast history reconstruction.
- Partition `telemetry_attestations` and `canary_results` in ClickHouse by month + `block_id`.
- Unique constraint on `(contract_id, idempotency_key)` in `contract_events` to enforce transition idempotency at the DB layer, not just application logic.
