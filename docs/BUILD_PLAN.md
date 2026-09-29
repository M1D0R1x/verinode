# Verinode — Colosseum Completion Build Plan

Goal: a full working product through Phase 4, submittable to Colosseum (Solana hackathon).
Solana is the headline chain: must produce real, clickable devnet transactions.

## Baseline (verified 2026-09-29)
- Go: build + vet + test PASS. Web: tsc PASS. Python index: 8/8 PASS. On `main`, clean.
- Phases 0–2 largely built. Chain adapters exist but are simulation stubs, NOT wired to API/UI.
- Toolchains present: Go 1.27, Node 26, pnpm 12. MISSING: Rust/cargo, solana CLI, anchor, foundry.

## Decisions
- Solana devnet integration = REAL via @solana/web3.js (no Rust toolchain needed to submit).
  Anchor program source kept as canonical on-chain spec + buildable when toolchain present.
  Mirror anchors canonical trade-envelope hash + attestation digest on devnet (memo/PDA), returns real signature + Explorer URL.
- Arbitrum (Sepolia) + Hyperliquid (testnet): real HTTP/RPC-capable Go adapters with a `SIMULATED` fallback
  so the product always runs even without keys. Wired into the API gateway + UI.
- Chains remain optional/read-only mirrors per Invariant 6. Off-chain Postgres stays authoritative.

## Workstreams
1. [ ] Solana devnet mirror service (JS): keypair gen + airdrop, submit real tx, expose signature+explorer URL. Go adapter calls it (or Go native web3 via RPC memo).
2. [ ] API gateway: /internal/chain/mirror, /v1/chain/proofs/{tradeId}, /v1/hedge/quote endpoints, wired to contract lifecycle transitions.
3. [ ] Hyperliquid hedge: index-driven hedge quote endpoint + basis analytics, wired to marketdata latest fix.
4. [ ] Phase 3 gaps: API keys/rate limit/idempotency, automated credit adjust, replacement routing, licensing feed endpoints.
5. [ ] Phase 4: stablecoin escrow flow (Solana PDA mirror), on-chain index publication adapter (real devnet), HL HIP-3 gating.
6. [ ] Frontend redesign: real visual identity (kill mono overuse, palette, hierarchy, remove skill-tag comments), add Web3 proofs page + Hedge desk + Chain settlement drawer.
7. [ ] End-to-end demo seed + README/DEMO script for judges (one command up, clickable devnet link).
8. [ ] Verify: go build/vet/test, tsc, pytest, and a live devnet tx captured.

## Progress log
- 2026-09-29: baseline verified; plan written; starting Solana devnet JS mirror + toolchain probe.
