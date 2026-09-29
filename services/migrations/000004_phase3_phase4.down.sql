-- Verinode Database Schema Migration: 000004_phase3_phase4.down.sql

DROP TABLE IF EXISTS onchain_escrows;
DROP TABLE IF EXISTS substitution_offers;
ALTER TABLE participant_credit_events DROP COLUMN IF EXISTS trigger_data;
ALTER TABLE participant_credit_events DROP COLUMN IF EXISTS automated;
DROP TABLE IF EXISTS api_keys;
