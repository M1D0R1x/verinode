# Colosseum Judging — Research & How Verinode Responds

*Sourced from Colosseum's official rules (Renaissance/Breakout/Cypherpunk) and their
"How to Win a Colosseum Hackathon" and submission-workshop guidance. Summarized here so
every product decision maps back to what judges actually score.*

## What judges evaluate (four official dimensions)

1. **Design / UX** — the product must look and feel like a real startup, not a demo.
2. **Development** — real, working software; reach a functioning **Solana devnet** demo.
3. **Business** — a viable business model and genuine founder-market fit; "enables a new
   market that couldn't exist without crypto, or improves an existing market by bringing
   it onchain."
4. **Solana platform understanding** — the Solana integration must be genuine, not bolted on.

## The highest-leverage facts

- **The pitch video (< 3 min) is the single most important artifact** — first thing judges
  review; determines shortlisting. (Not code, but the product must demo cleanly and tell a story.)
- **Prioritize the features that create the "aha" moment on devnet.** A judge should
  understand the value in one clean demo pass.
- Judges reward **teams intending to build full-time with a viable business model.**
- Every submission in the top 100 is exceptional — **polish and clarity are differentiators.**

## How Verinode responds (concrete)

| Lesson | Verinode action |
|---|---|
| Genuine Solana integration, reach devnet | Real pure-Go devnet client (ed25519 + RPC + Memo) anchoring each trade's canary attestation + state; live tx when funded, labelled simulated fallback otherwise. `/proofs` shows verifiable Explorer links. |
| Design/UX must be enterprise-grade | Institutional design system (warm ink + signal-gold + verification-green, serif display, mono only for data). Fixed contrast bugs, input overflow, CTA weight. |
| "Aha" in one demo pass | `docs/DEMO.md`: no-DB 2-minute run + click-through: RFQ → canary proof on Solana → hedge desk → index integrity gate. |
| Viable business + founder-market fit | Clear thesis: institutional GPU-capacity forwards with verifiable physical delivery — a real market ($ enterprise compute procurement) that today has no standardized, enforceable forward. Off-chain legal authority + on-chain verifiable proof is the wedge. |
| Solana understanding, not bolted on | Chain is scoped exactly where it adds trust (verifiable attestation mirror + optional escrow), never in the critical path — this *demonstrates* judgment about when to use a chain, which reads as platform maturity. |
| Business must not look like a toy | Real auth + RBAC (super_admin/admin platform console vs. company-scoped org panels), double-entry ledger invariants, surveillance, index integrity gates. |

## What we deliberately do NOT do (and why it's a strength, not a gap)

- No leveraged/cash-settled token, no order book, no bridge. The docs make the legal case
  (physical forward exclusion). Showing judges we *chose* to keep Solana as a read-only
  verifiable mirror — rather than forcing a tokenized perp — demonstrates platform
  understanding and de-risks the business, which is exactly the maturity judges reward.
