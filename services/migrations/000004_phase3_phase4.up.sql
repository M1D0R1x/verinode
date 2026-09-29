-- Verinode Database Schema Migration: 000004_phase3_phase4.up.sql
-- Phase 3 (integrations/licensing/automation) + Phase 4 (optional on-chain escrow mirror)
-- per docs/02 §Phase3-4, docs/05 (on-chain optional rails), docs/06 backlog.

-- ==============================================================================
-- PHASE 3.1 — Programmatic API keys (scoped, rate-limited, ERP integration)
-- ==============================================================================

CREATE TABLE IF NOT EXISTS api_keys (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    participant_id UUID REFERENCES participants(id) ON DELETE CASCADE,
    key_prefix TEXT NOT NULL,            -- first 8 chars, shown in UI
    key_hash TEXT NOT NULL,              -- sha256 of the full secret; secret shown once
    scopes TEXT[] NOT NULL DEFAULT '{}', -- e.g. {trading:read, index:read}
    rate_limit_per_min INT NOT NULL DEFAULT 120,
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'revoked')),
    last_used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_api_keys_prefix ON api_keys(key_prefix);
CREATE INDEX IF NOT EXISTS idx_api_keys_participant ON api_keys(participant_id, status);

-- ==============================================================================
-- PHASE 3.2 — Automated credit-limit adjustments (bounded, fully logged)
-- Extends participant_credit_events with the triggering data + automation flag.
-- ==============================================================================

ALTER TABLE participant_credit_events
    ADD COLUMN IF NOT EXISTS automated BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS trigger_data JSONB;

-- ==============================================================================
-- PHASE 3.3 — Replacement / substitution routing within the cure window
-- ==============================================================================

CREATE TABLE IF NOT EXISTS substitution_offers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    contract_id UUID NOT NULL REFERENCES contracts(id) ON DELETE CASCADE,
    replacement_block_id UUID REFERENCES inventory_blocks(id),
    offered_grade_id TEXT NOT NULL REFERENCES grades(id),
    kind TEXT NOT NULL CHECK (kind IN ('same_grade', 'superior_grade', 'cash_remedy')),
    status TEXT NOT NULL DEFAULT 'offered' CHECK (status IN ('offered', 'accepted', 'declined', 'expired')),
    cure_deadline TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_substitution_offers_contract ON substitution_offers(contract_id, status);

-- ==============================================================================
-- PHASE 4 — Optional stablecoin escrow mirror (Solana per-trade PDA)
-- Never authoritative (Invariant 6); bank/invoice remains primary. This table
-- only records the on-chain mirror linkage + capped balance, not custody truth.
-- ==============================================================================

CREATE TABLE IF NOT EXISTS onchain_escrows (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    contract_id UUID NOT NULL REFERENCES contracts(id) ON DELETE CASCADE,
    chain TEXT NOT NULL CHECK (chain IN ('solana', 'arbitrum')),
    pda_or_address TEXT NOT NULL,
    escrow_state TEXT NOT NULL DEFAULT 'funded' CHECK (escrow_state IN ('funded', 'released', 'disputed', 'resolved')),
    amount_cents BIGINT NOT NULL CHECK (amount_cents >= 0),
    cap_cents BIGINT NOT NULL,           -- per-PDA balance cap (blast-radius bound)
    last_tx_signature TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_escrow_cap CHECK (amount_cents <= cap_cents)
);

CREATE INDEX IF NOT EXISTS idx_onchain_escrows_contract ON onchain_escrows(contract_id, chain);
