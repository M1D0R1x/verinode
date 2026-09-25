-- Verinode Database Schema Migration: 000001_init_schema.down.sql

DROP TABLE IF EXISTS event_outbox;
DROP TABLE IF EXISTS claim_evidence;
DROP TABLE IF EXISTS claims;
DROP TABLE IF EXISTS delivery_events;
DROP TABLE IF EXISTS settlements;
DROP TABLE IF EXISTS invoice_line_items;
DROP TABLE IF EXISTS invoices;
DROP TABLE IF EXISTS ledger_entries;
DROP TABLE IF EXISTS ledger_accounts;
DROP TABLE IF EXISTS contract_events;
DROP TABLE IF EXISTS contracts;
DROP TABLE IF EXISTS quotes;
DROP TABLE IF EXISTS rfq_invited_sellers;
DROP TABLE IF EXISTS rfqs;
DROP TABLE IF EXISTS inventory_blocks;
DROP TABLE IF EXISTS grades;
DROP TABLE IF EXISTS participant_credit_events;
DROP TABLE IF EXISTS beneficial_owners;
DROP TABLE IF EXISTS participants;
