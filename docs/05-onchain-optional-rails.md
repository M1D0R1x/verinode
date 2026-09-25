# Verinode — On-Chain / Optional Rails

*This doc exists specifically to answer "where's Solana/Arbitrum/Hyperliquid?" — and to make sure nobody reaches for a chain-first design by default while building the other four docs.*

## 1. The core decision, stated plainly

**Verinode is off-chain-first by design, not by omission.** The source strategy explicitly rejects the "GPU futures on chains" thesis:

> *"'GPU futures on chains' is no longer a differentiated thesis and creates the hardest legal and liquidity problems first."*

And settlement is spec'd as bank transfer/invoice **first**, with anything crypto explicitly optional, opt-in, and legally gated:

> *"Settlement: bank transfer/invoice first. Optional stablecoin escrow only for customers who ask for it and after legal review."*

So this isn't an oversight in the architecture — building on Solana/Arbitrum as core infrastructure, or modeling the product after Hyperliquid, would work *against* the strategy, not toward it.

## 2. Why Hyperliquid specifically is the wrong model

Hyperliquid is a perpetuals DEX: continuous, leveraged, cash-settled, retail-tradable synthetic exposure. Every one of those properties is explicitly what Verinode is told **not** to build:

| Hyperliquid-style property | Verinode's actual requirement |
|---|---|
| Perpetual, continuously tradable | Bilateral, non-transferable reservation. No secondary trading until "counsel and market structure support it." |
| Cash-settled synthetic exposure | Physical delivery of an actual GPU node — cash settlement "cannot reserve a machine." |
| Retail leverage / margin | No retail leverage. Mark-to-market margin is explicitly called out as "a derivatives feature and may change the legal perimeter." |
| Anonymous on-chain counterparties | Procurement teams "prefer a known counterparty and clear remedies... anonymous on-chain counterparties and volatile collateral are usually a regression." |

Copying that model isn't just a stylistic mismatch — it directly undermines the legal theory (the US forward exclusion) the whole business depends on, which requires genuine intent to physically deliver, no routine cash close-out, and no marketing as an investment.

## 3. Where a chain *could* legitimately show up — and only there

There are exactly two places in the architecture where a chain is contemplated at all, both in **Phase 4**, both gated:

### 3a. Optional stablecoin escrow
- **Trigger:** a specific customer requests it.
- **Gate:** legal and custody review must clear it first — not a default checkbox in onboarding.
- **Scope:** it replaces one settlement rail (bank/escrow) for that customer's deposits/collateral. It does not change the contract, the state machine, or how delivery is verified.
- **Non-negotiable:** bank transfer/invoice remains the primary path for everyone else, indefinitely.

### 3b. On-chain publication adapter (for the index, once it exists)
This is the one place a chain like Solana or Arbitrum could plausibly appear — and even then, only as a thin mirror, never as the source of truth. From the architecture doc's oracle/publication design:

- The index engine computes **off-chain** from an append-only validated dataset.
- Publication requires **two-person release or threshold signatures**.
- The canonical artifact is a **signed JSON/CSV published over HTTPS and object storage** — this is what legal and every consumer relies on.
- An on-chain adapter, if built, does nothing more than **read that same signed payload** and republish it on a given chain, carrying `series_id, value, unit, observation_window, publish_time, methodology_version, sequence_number, expiration`.
- Adapters must **reject stale, duplicate, wrong-decimal, or unknown-version messages.**
- **No cross-chain "consensus."** Each chain just mirrors the one signed canonical fix. If a chain is unavailable, the legal/operational fallback is always the signed off-chain publication — never "wait for the chain."

If you ever build this adapter, treat "which chain" (Solana vs. Arbitrum vs. anything else) as an implementation detail with no bearing on trust — the adapter is a dumb, verifiable mirror, not infrastructure the business depends on.

## 4. If Phase 4 is triggered — implementation notes per rail

These only apply once the triggers in §3 are actually met. Don't build any of this speculatively.

### 4a. Solana (stablecoin escrow)

- Scope it as a **small, minimal Rust/Anchor program** — resist the urge to make it a general-purpose escrow framework. Smaller surface area is easier to audit and reason about.
- Use **per-trade PDAs** (program-derived addresses) so each reservation's escrow is isolated — no shared pooled balance across trades.
- Model **explicit state transitions** (funded → released → disputed → resolved) mirroring the off-chain contract state machine, not a parallel state machine that can drift from it.
- Include **timeouts** so funds don't get stuck if a counterparty goes silent, and **multisig for disputed-fund release** rather than a single admin key.
- **Cap balances per PDA** — bound the blast radius of any single bug or compromised key.
- Keep **identities, commercial terms, and telemetry off-chain**. The chain only ever sees a funding amount and a release/dispute instruction, never who the parties are or what they agreed to.
- If using Pyth for any on-chain price reference (e.g., for a stablecoin's peg), follow Pyth's Solana guidance on feed identity, ownership, and staleness checks — a stale or spoofed feed on a financial-adjacent contract is a real loss vector.
- Solana's Confidential Transfer extension can hide transfer amounts, but it does **not** solve metadata privacy generally (counterparties, timing, and PDA linkage can still leak information) — don't treat it as a full privacy solution.

### 4b. Arbitrum (only for enterprise EVM integration)

- Only build this if a specific enterprise customer **requires** EVM-native integration — it is not a default rail.
- If needed, evaluate a **permissioned Arbitrum Orbit chain** rather than deploying to public Arbitrum. Orbit gives configurable chains and data availability, but it adds real ongoing burden: sequencer operation, governance, and reconciliation against the off-chain canonical record.
- Arbitrum does **not** make a contract more private or more legally enforceable than a signed off-chain agreement — normal e-signature and a contract repository remain the actual legal instrument. Anchor a hash on-chain only if a specific customer values that as evidence, not as a default practice.

### 4c. Hyperliquid (only as a later hedge-distribution venue)

- Hyperliquid's HIP-3 lets a "deployer" define a market and its oracle — and that deployer **bears responsibility** for leverage and settlement risk on that market. Hyperliquid's own documentation warns explicitly about unsuitable or manipulable indices and about slashing risk for deployers.
- **Do not deploy a HIP-3 market on a self-published, thin GPU index.** That's exactly the oracle/manipulation scenario Hyperliquid's own docs warn deployers about — a low-volume, single-contributor-adjacent index is trivially easier to move than a deep one.
- Trigger for even considering this: the index has **independent observations, demonstrated manipulation resistance, and committed market makers** (i.e., Phase 2/3 governance maturity, not Phase 1 volume).
- A third-party oracle vendor (e.g., Pyth) can supply HIP-3 infrastructure, but **no oracle vendor can manufacture independent price observations that don't exist** — infrastructure isn't a substitute for real, deep, arm's-length trading activity.

### 4d. Oracle/publication design (applies whether or not any chain is ever used)

This is really a hardening of the Phase 2 index design (see `02-prd-feature-spec.md` §2, Phase 2), described here because it's the piece any chain adapter would eventually read from:

- Aggregate from **authenticated completed trades, executable quotes, and independent benchmarks** — never a single contributor's self-reported number.
- Use a **deterministic methodology** (documented, versioned) rather than case-by-case judgment calls at publication time.
- Enforce **contributor caps**, **confidence bands**, **threshold signatures** for release, **heartbeats**, and **staleness/deviation guards**.
- Maintain an explicit **"no settlement" / "insufficient data" state** as a first-class output, never a fabricated fallback value.
- Pyth's multi-publisher aggregation with confidence intervals is a reasonable model to study for the *shape* of this design (many independent publishers, an aggregated value, and an explicit confidence/uncertainty band) — not because Verinode needs to use Pyth, but because the pattern of "aggregate many independent signed inputs into a value plus a confidence measure, with an explicit low-confidence state" is exactly right here.

### 4e. No bridging, ever

**Do not bridge collateral across Solana, Arbitrum, and Hyperliquid**, even if more than one of these is eventually adopted for different customers. Each venue reconciles independently against the one signed canonical trade record (`trade_id`, per the architecture doc §1.8 and data model doc §12) and keeps its balances purpose-limited to that venue. A cross-chain bridge is both a security liability (bridge failures are one of the most common ways crypto systems lose funds) and an unnecessary one here — nothing about the product requires moving value between chains.

## 5. What this means for engineering, concretely

- **Do not** put any chain, wallet, or token in the critical path of RFQ → contract → funding → delivery → settlement. That entire flow is bank/invoice + signed documents + the Postgres event log.
- **Do not** build wallet-based identity/auth for participants. Participant identity is KYB/KYC-backed legal-entity identity (see `04-data-model-schema.md` §2), not a wallet address.
- **Do not** treat "we could tokenize the reservation" as a feature request — the docs are explicit that tokenization "does not reduce the [regulatory] perimeter" and adds custody, AML, sanctions, money-transmission, tax, and smart-contract risk on top of the existing regulatory surface, for no delivery benefit.
- **Do** build the append-only, signed-publication pattern for the index (§3b above) in a chain-agnostic way from the start, so that *if* a chain adapter is ever approved, it's a thin add-on, not a rearchitecture.
- **Do** keep collateral/custody logic buildable-and-swappable: use a third-party bank/escrow rail underneath the ledger abstraction (`04-data-model-schema.md` §6) so that adding stablecoin escrow later for one customer doesn't require touching the ledger's core double-entry model.

## 6. Phase-4-specific risk additions

These extend the risk register in `01-architecture-and-getting-started.md` §1.9 — only relevant if any rail in §4 is actually built:

| Risk | Severity | Control |
|---|---|---|
| Smart-contract/bridge failure | Medium initially, rises with adoption | Minimal audited contracts, capped balances per PDA, pause controls, avoid bridges entirely (§4e) |
| Stablecoin depeg/freeze | Medium initially | Bank-rail settlement remains available as fallback for every customer; exposure limits; rapid settlement if a depeg event starts |
| Oracle/HIP-3 deployer liability | Critical if Hyperliquid is used | Never deploy on a thin/self-published index; require the Phase 2 governance maturity bar (§4c) before considering it |

## 7. Summary table

| Component | Chain involvement | Phase | Status |
|---|---|---|---|
| RFQ, matching, contracts | None | 1 | Core, off-chain |
| Collateral/payment ledger | None (bank/escrow only) | 1 | Core, off-chain |
| Telemetry/attestation | None (signed reports, off-chain) | 1 | Core, off-chain |
| Delivery/claims | None | 1 | Core, off-chain |
| Index calculation & publication | None — signed off-chain fix is canonical | 2 | Core, off-chain |
| Stablecoin escrow | Small Rust/Anchor program on Solana, per-trade PDAs, customer-triggered, legally gated (§4a) | 4 | Not built by default |
| On-chain index mirror (any chain incl. Solana/Arbitrum) | Thin adapter reading the signed off-chain fix (§3b) | 4 | Not built by default |
| Enterprise EVM integration | Permissioned Arbitrum Orbit chain, only if a customer requires it (§4b) | 4 | Not built by default |
| Perpetual/leveraged hedge product (Hyperliquid HIP-3) | Only after independent index observations, manipulation resistance, and committed market makers exist (§4c) | 4 | Not built by default — never as a Phase 1 default, never on a thin index |
