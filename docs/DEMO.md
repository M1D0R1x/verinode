# Verinode — Demo & Judge Quickstart

**Verinode is an institutional desk for physically delivered GPU-capacity forward
reservations** — standardized bilateral contracts, independent cryptographic canary
verification, and balanced double-entry escrow, with **optional on-chain audit
mirrors on Solana**. Chains are read-only proofs; the off-chain PostgreSQL record and
signed legal confirmation are always authoritative (Invariant 6).

This document is the fastest path for a judge to run the product and see the Solana
integration produce a verifiable proof.

---

## 0. TL;DR (no database, ~2 min)

The API gateway runs standalone (repositories are nil-safe), so you can exercise the
whole chain + hedge surface with zero infra:

```bash
# 1. API gateway
cd services && go run ./cmd/api-gateway      # :8080

# 2. Web app (new terminal)
cd apps/web && pnpm install && pnpm dev       # :3000
```

Open <http://localhost:3000> and visit:

- **/** — the marketplace (institutional landing).
- **/proofs** — anchor a canonical trade across Solana + Arbitrum and get verifiable
  proofs. Runs in **simulated** mode out of the box; **live devnet** once a keypair is
  funded (see §3).
- **/hedge** — the Hyperliquid basis-hedge desk: sizes the offsetting perp against a
  physical forward.
- **/market-data** — the signed benchmark index (Invariant 5: `insufficient_data`
  is a first-class response, never fabricated).

---

## 1. Full stack with infrastructure

```bash
cp .env.example .env
docker compose up -d                 # Postgres, ClickHouse, Redis, MinIO, Temporal
cd services && go run ./cmd/migrate  # applies migrations 0001–0004
go run ./cmd/seed-demo               # turnkey demo dataset (participants→contract→index)
go run ./cmd/api-gateway             # :8080
```

Frontend: `cd apps/web && pnpm install && pnpm dev`.

---

## 2. What to click (2-minute demo script)

1. **Landing → "Submit an allocation RFQ"** — the physical-forward product and the
   four-model boundary that keeps it out of derivative territory.
2. **/market-data** — a volume-weighted median index with confidence bands; force
   `insufficient_data` by publishing with too few contributors to see the integrity gate.
3. **/proofs** — paste a `trade_id` and click **Anchor mirror**. Each rail returns a
   state anchor + canary attestation with an Explorer link. Note the **Invariant 6**
   banner: this is a mirror, not the source of truth.
4. **/hedge** — enter a fixed rate above the index mark → **Short the perp** with a
   sized basis; move it within 0.5% → **Hold unhedged**. This is the Hyperliquid hedge
   idea, as basis analytics, never a leveraged retail product.
5. **/admin/surveillance** — wash-trade / related-party flags reviewed before anything
   feeds the index.

---

## 3. Arming a REAL Solana devnet transaction

The integration is live-capable today; it needs a funded signer (the public faucet is
IP-rate-limited, so this is a one-time human step):

```bash
cd services
go run ./cmd/devnet-prove          # prints the signer pubkey; persists a keypair
# -> fund that pubkey with ~0.05 devnet SOL at https://faucet.solana.com
go run ./cmd/devnet-prove          # now prints a live signature + Explorer URL
```

Then export the keypair so the gateway submits real txs from **/proofs**:

```bash
export SOLANA_SETTLEMENT_KEYPAIR=$(cat services/.devnet-keypair.hex)
go run ./cmd/api-gateway
```

Arbitrum (`ARBITRUM_SETTLEMENT_KEY`) and Hyperliquid (`HYPERLIQUID_DEPLOYER_KEY`)
follow the same pattern; all default to a labelled simulated proof when unset, so the
product never breaks in a demo.

---

## 4. Verifying the invariants (what makes this defensible)

```bash
cd services && go test ./...        # state machine, double-entry ledger, chain, phase3
cd ../index-service && pytest       # index calculation + insufficient-data gate
```

- **Invariant 1** — physical forward only: no order book, no tradable token, no cash netting.
- **Invariant 3** — every ledger `transaction_id` group sums to zero (fuzz + unit tested).
- **Invariant 5** — index returns `insufficient_data` rather than a fabricated number;
  the Hyperliquid oracle refuses to publish below threshold.
- **Invariant 6** — chains are optional read-only mirrors; bank/invoice settlement is primary.

---

## 5. Architecture at a glance

```
Next.js web ── Go API gateway ── Postgres (authoritative) + ClickHouse + Redis + MinIO
                     │
                     ├── Temporal workflow: RFQ→sign→fund→attest→deliver→settle
                     ├── Python index service (separated trust domain, signed fixes)
                     └── chain.Orchestrator (Phase 4, optional):
                            Solana devnet anchor  ·  Arbitrum Sepolia mirror  ·  Hyperliquid hedge/oracle
```

See `docs/01-architecture-and-getting-started.md` for the full design and
`docs/05-onchain-optional-rails.md` for exactly why the chains are scoped the way they are.
