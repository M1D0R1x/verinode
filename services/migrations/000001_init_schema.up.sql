-- Verinode Database Schema Migration: 000001_init_schema.up.sql
-- Implements Phase 1 core tables per docs/04-data-model-schema.md

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ==============================================================================
-- 1. PARTICIPANT DOMAIN
-- ==============================================================================

CREATE TABLE participants (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    legal_name TEXT NOT NULL,
    jurisdiction TEXT NOT NULL, -- ISO-3166 country code
    role TEXT NOT NULL CHECK (role IN ('buyer', 'seller', 'both')),
    kyc_status TEXT NOT NULL DEFAULT 'pending' CHECK (kyc_status IN ('pending', 'approved', 'rejected', 'review')),
    sanctions_result JSONB,
    credit_limit_cents BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE beneficial_owners (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    participant_id UUID NOT NULL REFERENCES participants(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    ownership_pct NUMERIC(5,2) NOT NULL CHECK (ownership_pct > 0 AND ownership_pct <= 100),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE participant_credit_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    participant_id UUID NOT NULL REFERENCES participants(id),
    delta_cents BIGINT NOT NULL,
    reason TEXT NOT NULL,
    actor TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_beneficial_owners_participant ON beneficial_owners(participant_id);
CREATE INDEX idx_credit_events_participant ON participant_credit_events(participant_id, created_at);

-- ==============================================================================
-- 2. INVENTORY & GRADES DOMAIN
-- ==============================================================================

CREATE TABLE grades (
    id TEXT PRIMARY KEY, -- e.g. 'H100-SXM-8XNV'
    gpu_sku TEXT NOT NULL,
    min_memory_gb INT NOT NULL,
    topology TEXT NOT NULL,
    min_healthy_gpu_count INT NOT NULL,
    benchmark_floor JSONB NOT NULL,
    min_cpu_cores INT NOT NULL,
    min_ram_gb INT NOT NULL,
    min_nvme_perf INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE inventory_blocks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    seller_id UUID NOT NULL REFERENCES participants(id),
    grade_id TEXT NOT NULL REFERENCES grades(id),
    region_bucket TEXT NOT NULL,
    facility_ref TEXT, -- Access-restricted facility code/hash
    window_start TIMESTAMPTZ NOT NULL,
    window_end TIMESTAMPTZ NOT NULL,
    status TEXT NOT NULL DEFAULT 'available' CHECK (status IN ('available', 'reserved', 'delivered', 'released')),
    latest_attestation_ref TEXT,
    last_heartbeat_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_inventory_window CHECK (window_end > window_start)
);

CREATE INDEX idx_inventory_blocks_search ON inventory_blocks(grade_id, region_bucket, status, window_start, window_end);

-- ==============================================================================
-- 3. RFQ & TRADING DOMAIN
-- ==============================================================================

CREATE TABLE rfqs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    buyer_id UUID NOT NULL REFERENCES participants(id),
    grade_id TEXT NOT NULL REFERENCES grades(id),
    region_bucket TEXT NOT NULL,
    window_start TIMESTAMPTZ NOT NULL,
    window_end TIMESTAMPTZ NOT NULL,
    status TEXT NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'quoted', 'accepted', 'cancelled', 'expired')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_rfq_window CHECK (window_end > window_start)
);

CREATE TABLE rfq_invited_sellers (
    rfq_id UUID NOT NULL REFERENCES rfqs(id) ON DELETE CASCADE,
    seller_id UUID NOT NULL REFERENCES participants(id) ON DELETE CASCADE,
    PRIMARY KEY (rfq_id, seller_id)
);

CREATE TABLE quotes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    rfq_id UUID NOT NULL REFERENCES rfqs(id) ON DELETE CASCADE,
    seller_id UUID NOT NULL REFERENCES participants(id),
    block_id UUID NOT NULL REFERENCES inventory_blocks(id),
    price_cents BIGINT NOT NULL CHECK (price_cents > 0),
    currency TEXT NOT NULL DEFAULT 'USD',
    expires_at TIMESTAMPTZ NOT NULL,
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'accepted', 'expired', 'withdrawn')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_rfqs_buyer ON rfqs(buyer_id, status);
CREATE INDEX idx_quotes_rfq ON quotes(rfq_id, status);

-- ==============================================================================
-- 4. CONTRACT DOMAIN (SOURCE OF TRUTH & STATE MACHINE)
-- ==============================================================================

CREATE TABLE contracts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(), -- Globally unique trade_id
    rfq_id UUID REFERENCES rfqs(id),
    quote_id UUID REFERENCES quotes(id),
    buyer_id UUID NOT NULL REFERENCES participants(id),
    seller_id UUID NOT NULL REFERENCES participants(id),
    grade_id TEXT NOT NULL REFERENCES grades(id),
    template_version TEXT NOT NULL,
    state TEXT NOT NULL CHECK (state IN (
        'draft_rfq', 'open', 'quoted', 'accepted', 'contract_pending',
        'funded_secured', 'scheduled', 'delivery_test', 'live', 'completed', 'settled',
        'cancelled', 'failed_delivery', 'cure', 'substituted', 'claim_open', 'disputed', 'terminated'
    )),
    signed_pdf_hash TEXT, -- SHA-256 hash of final signed PDF confirmation
    signed_pdf_ref TEXT,  -- Object storage key
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE contract_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    contract_id UUID NOT NULL REFERENCES contracts(id) ON DELETE CASCADE,
    prior_state TEXT NOT NULL,
    new_state TEXT NOT NULL,
    actor TEXT NOT NULL,
    reason TEXT NOT NULL,
    idempotency_key TEXT NOT NULL,
    evidence_hash TEXT,
    authorization_decision JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_contract_events_idempotency UNIQUE (contract_id, idempotency_key)
);

CREATE INDEX idx_contracts_state ON contracts(state);
CREATE INDEX idx_contracts_parties ON contracts(buyer_id, seller_id);
CREATE INDEX idx_contract_events_contract ON contract_events(contract_id, created_at);

-- ==============================================================================
-- 5. DOUBLE-ENTRY LEDGER DOMAIN
-- ==============================================================================

CREATE TABLE ledger_accounts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    participant_id UUID NOT NULL REFERENCES participants(id),
    currency TEXT NOT NULL DEFAULT 'USD',
    account_type TEXT NOT NULL CHECK (account_type IN ('deposit', 'payable', 'receivable', 'collateral')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE ledger_entries (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    transaction_id UUID NOT NULL, -- Groups balanced pair/set of entries (SUM == 0)
    account_id UUID NOT NULL REFERENCES ledger_accounts(id),
    contract_id UUID REFERENCES contracts(id),
    amount_cents BIGINT NOT NULL, -- Positive = Debit, Negative = Credit
    currency TEXT NOT NULL DEFAULT 'USD',
    bank_reference TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE invoices (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    contract_id UUID NOT NULL REFERENCES contracts(id),
    buyer_id UUID NOT NULL REFERENCES participants(id),
    seller_id UUID NOT NULL REFERENCES participants(id),
    total_cents BIGINT NOT NULL CHECK (total_cents >= 0),
    currency TEXT NOT NULL DEFAULT 'USD',
    status TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'issued', 'paid', 'void')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE invoice_line_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    invoice_id UUID NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
    description TEXT NOT NULL,
    amount_cents BIGINT NOT NULL
);

CREATE TABLE settlements (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    contract_id UUID NOT NULL REFERENCES contracts(id),
    total_payout_cents BIGINT NOT NULL,
    sla_credit_cents BIGINT NOT NULL DEFAULT 0,
    currency TEXT NOT NULL DEFAULT 'USD',
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'completed', 'disputed')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ledger_entries_tx ON ledger_entries(transaction_id);
CREATE INDEX idx_ledger_entries_account ON ledger_entries(account_id);
CREATE INDEX idx_ledger_entries_contract ON ledger_entries(contract_id);

-- ==============================================================================
-- 6. DELIVERY & CLAIMS DOMAIN
-- ==============================================================================

CREATE TABLE delivery_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    contract_id UUID NOT NULL REFERENCES contracts(id) ON DELETE CASCADE,
    event_type TEXT NOT NULL CHECK (event_type IN ('credentials_verified', 'health_test', 'canary_result', 'accepted', 'failed')),
    evidence_refs JSONB NOT NULL DEFAULT '[]',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE claims (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    contract_id UUID NOT NULL REFERENCES contracts(id),
    opened_by UUID NOT NULL REFERENCES participants(id),
    type TEXT NOT NULL CHECK (type IN ('outage', 'degradation', 'non_delivery', 'nonpayment')),
    state TEXT NOT NULL DEFAULT 'claim_open' CHECK (state IN ('claim_open', 'resolved', 'disputed')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMPTZ
);

CREATE TABLE claim_evidence (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    claim_id UUID NOT NULL REFERENCES claims(id) ON DELETE CASCADE,
    evidence_ref TEXT NOT NULL,
    evidence_type TEXT NOT NULL,
    added_by TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_delivery_events_contract ON delivery_events(contract_id, created_at);
CREATE INDEX idx_claims_contract ON claims(contract_id, state);

-- ==============================================================================
-- 7. TRANSACTIONAL EVENT OUTBOX
-- ==============================================================================

CREATE TABLE event_outbox (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    aggregate_type TEXT NOT NULL,
    aggregate_id TEXT NOT NULL,
    event_type TEXT NOT NULL,
    payload JSONB NOT NULL,
    published BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_event_outbox_unpublished ON event_outbox(published, created_at) WHERE published = FALSE;
