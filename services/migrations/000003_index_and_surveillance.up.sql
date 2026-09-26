-- Verinode Database Schema Migration: 000003_index_and_surveillance.up.sql
-- Implements Phase 2 Index & Surveillance tables per docs/04-data-model-schema.md §9 and §10

-- ==============================================================================
-- 1. INDEX DOMAIN (Separate trust & read surface)
-- ==============================================================================

CREATE TABLE IF NOT EXISTS index_series (
    id TEXT PRIMARY KEY, -- e.g. 'H100-SXM-8XNV-US-WEEK-DEDICATED-USD'
    gpu_model TEXT NOT NULL,
    form TEXT NOT NULL,
    topology TEXT NOT NULL,
    region_bucket TEXT NOT NULL,
    tenor TEXT NOT NULL, -- '168h'
    tenancy TEXT NOT NULL, -- 'dedicated'
    currency TEXT NOT NULL DEFAULT 'USD',
    methodology_version TEXT NOT NULL DEFAULT 'v1.0.0-institutional',
    min_contributors INT NOT NULL DEFAULT 3,
    min_notional_usd NUMERIC NOT NULL DEFAULT 50000.0,
    max_contributor_weight NUMERIC NOT NULL DEFAULT 0.35,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS index_observations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    series_id TEXT NOT NULL REFERENCES index_series(id),
    value NUMERIC, -- NULL if insufficient_data = TRUE per Invariant 5
    unit TEXT NOT NULL DEFAULT 'USD_PER_NODE_HOUR',
    observation_window_start TIMESTAMPTZ NOT NULL,
    observation_window_end TIMESTAMPTZ NOT NULL,
    publish_time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    sequence_number BIGINT NOT NULL,
    contributor_count INT NOT NULL,
    observation_count INT NOT NULL DEFAULT 0,
    total_notional_usd NUMERIC NOT NULL DEFAULT 0,
    confidence_interval_low NUMERIC,
    confidence_interval_high NUMERIC,
    insufficient_data BOOLEAN NOT NULL DEFAULT FALSE,
    reason TEXT,
    signature TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_series_sequence UNIQUE (series_id, sequence_number)
);

CREATE TABLE IF NOT EXISTS index_contributions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    series_id TEXT NOT NULL REFERENCES index_series(id),
    tier TEXT NOT NULL CHECK (tier IN ('completed_trade', 'matched_trade_pending', 'firm_two_sided_quote', 'firm_one_sided_quote', 'public_list_price')),
    contract_id UUID REFERENCES contracts(id) ON DELETE SET NULL,
    hourly_price_usd NUMERIC NOT NULL,
    duration_hours INT NOT NULL,
    notional_usd NUMERIC NOT NULL,
    contributor_id UUID NOT NULL REFERENCES participants(id),
    related_party_flag BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_index_observations_series ON index_observations(series_id, publish_time DESC);
CREATE INDEX IF NOT EXISTS idx_index_contributions_series ON index_contributions(series_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_index_contributions_contract ON index_contributions(contract_id);

-- ==============================================================================
-- 2. SURVEILLANCE DOMAIN
-- ==============================================================================

CREATE TABLE IF NOT EXISTS surveillance_flags (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    subject_type TEXT NOT NULL CHECK (subject_type IN ('contract', 'contribution', 'participant')),
    subject_id TEXT NOT NULL,
    flag_type TEXT NOT NULL CHECK (flag_type IN ('wash_trade', 'related_party', 'concentration', 'spoofing', 'end_window_marking')),
    severity TEXT NOT NULL DEFAULT 'warning' CHECK (severity IN ('info', 'warning', 'critical')),
    details JSONB NOT NULL DEFAULT '{}'::jsonb,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'reviewed', 'dismissed', 'escalated')),
    reviewed_by TEXT,
    resolution TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_surveillance_flags_status ON surveillance_flags(status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_surveillance_flags_subject ON surveillance_flags(subject_type, subject_id);
