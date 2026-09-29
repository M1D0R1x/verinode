import Link from "next/link";
import {
  ArrowRight,
  ArrowUpRight,
  ShieldCheck,
  Activity,
  Lock,
  Scale,
  CircleCheck,
  CircleDashed,
} from "lucide-react";

const proofStats = [
  { label: "Standardized tenor", value: "168h", sub: "one-week dedicated block" },
  { label: "AllReduce floor", value: "≥ 400 GB/s", sub: "canary-gated delivery", verify: true },
  { label: "Legal perimeter", value: "Physical forward", sub: "delivery, not cash-settlement" },
  { label: "Workload ingest", value: "Zero", sub: "hardware telemetry only", verify: true },
];

const boundary = [
  {
    model: "Spot capacity",
    promise: "Burst compute, right now",
    risk: "Preemption, variable topology",
    position: "Out of scope — AWS / RunPod lane",
    ours: false,
  },
  {
    model: "Physical forward reservation",
    promise: "Guaranteed future node at a fixed price",
    risk: "Default, mitigated by escrow + canary",
    position: "The Verinode product",
    ours: true,
  },
  {
    model: "Cash-settled derivative",
    promise: "Synthetic payout against an index",
    risk: "Leverage, liquidation, can't train on it",
    position: "Out of scope — CME / Hyperliquid lane",
    ours: false,
  },
  {
    model: "Bespoke enterprise contract",
    promise: "Multi-year custom commitment",
    risk: "Months of legal, opaque SLAs",
    position: "Standardized into one confirmation",
    ours: false,
  },
];

const guarantees = [
  {
    icon: Activity,
    title: "Canary-verified before handover",
    body: "A signed host agent runs an NCCL all-reduce and memory sweep before the block goes live. Below the 400 GB/s floor, the contract routes to cure, substitution, or a full refund — never a silent failure.",
  },
  {
    icon: Lock,
    title: "No customer workload, ever",
    body: "Telemetry reads PCI IDs, DCGM health and NVLink throughput. Weights, training code, prompts and SSH sessions never touch the gateway. It is an architectural invariant, not a policy.",
  },
  {
    icon: Scale,
    title: "Balanced double-entry escrow",
    body: "Every settlement produces debits and credits that sum to zero per transaction, reconciled to regulated bank rails. Optional per-trade on-chain escrow mirrors it — bank settlement stays authoritative.",
  },
];

const steps = [
  {
    n: "01",
    title: "Private RFQ, matched quote",
    body: "A buyer submits a private request against the exact grade and window. Invited suppliers quote firm. On acceptance a versioned confirmation is generated and hashed — no free-text fields.",
  },
  {
    n: "02",
    title: "Cryptographic canary",
    body: "At T-24h the host agent runs the standardized benchmark, signs the report with its ed25519 key, and the gateway checks it against the grade floor before anything is called delivered.",
  },
  {
    n: "03",
    title: "Handover & settlement",
    body: "A passing canary releases credentials for the full 168-hour block. On completion the subledger settles, evidence is archived, and only verified trades feed the published index.",
  },
];

const faqs = [
  {
    q: "Why a physical forward rather than a derivative?",
    a: "A physical forward that results in actual delivery of commercial capacity sits inside the CFTC forward exclusion. Verinode brokers take-or-pay reservation of real hardware for training runs — not retail cash-settled speculation. That legal position depends on genuine intent to deliver, no routine cash close-out, and no investment marketing.",
  },
  {
    q: "How does the canary protect a buyer?",
    a: "Before the delivery window opens, a signed host agent runs an NCCL all-reduce and memory test. If NVLink throughput drops below 400 GB/s or ECC errors appear, the supplier must cure inside the window or the contract cancels with a full escrow refund.",
  },
  {
    q: "Does Verinode ever see training data?",
    a: "Never. The host agent probes hardware only. Customer workloads, weights and code are fully isolated from the telemetry gateway — this is a hard invariant enforced by the data-flow boundary, not a setting.",
  },
  {
    q: "What do Solana, Arbitrum and Hyperliquid do here?",
    a: "Each has a distinct job. Solana anchors settlement and a verifiable proof of every canary attestation and state transition. Arbitrum registers enterprise contract state for counterparties that require EVM-native records. Hyperliquid is the basis-hedging venue where a fixed-price forward can be offset against the floating GPU-hour index. The off-chain PostgreSQL record and signed legal confirmation remain authoritative — the chains carry verifiable mirrors, never a second source of truth.",
  },
];

export default function HomePage() {
  return (
    <div className="flex flex-col">
      {/* Hero */}
      <section className="relative overflow-hidden border-b border-line">
        <div className="pointer-events-none absolute inset-0 bg-grid-faint [background-size:56px_56px] opacity-40" />
        <div className="relative mx-auto max-w-6xl px-4 pt-24 pb-28 sm:px-6 lg:px-8">
          <div className="max-w-3xl animate-rise">
            <div className="inline-flex items-center gap-2.5 rounded-pill border border-line bg-ink-900 px-3.5 py-1.5">
              <span className="h-1.5 w-1.5 rounded-full bg-verify animate-pulseDot" />
              <span className="tabular text-xs text-muted">
                Benchmark grade · 8× H100 SXM 80GB · 168h dedicated
              </span>
            </div>

            <h1 className="mt-7 font-serif text-display text-parchment">
              Reserve the exact machine.{" "}
              <span className="italic text-signal">Prove it was delivered.</span>
            </h1>

            <p className="mt-6 max-w-2xl text-lg leading-relaxed text-muted">
              Verinode is an institutional desk for GPU compute sold spot and forward —
              buyers lock price, sellers lock revenue. Standardized bilateral contracts,
              independent cryptographic canary verification, and balanced double-entry
              escrow settled on Solana, so a cluster booked eight weeks out is one you can
              hold your supplier to.
            </p>

            <div className="mt-9 flex flex-wrap items-center gap-3">
              <Link
                href="/buyer/rfqs/new"
                className="group inline-flex items-center gap-2 rounded-xl bg-signal px-5 py-3 text-sm font-semibold text-ink-950 transition-all duration-150 hover:bg-signal-bright active:scale-[0.98]"
              >
                Submit an allocation RFQ
                <ArrowRight className="h-4 w-4 transition-transform group-hover:translate-x-0.5" />
              </Link>
              <Link
                href="/seller"
                className="inline-flex items-center gap-2 rounded-xl border border-line bg-ink-850 px-5 py-3 text-sm font-semibold text-parchment transition-colors hover:bg-ink-800"
              >
                List cluster capacity
              </Link>
              <Link
                href="/market-data"
                className="inline-flex items-center gap-1.5 px-2 py-3 text-sm font-medium text-muted transition-colors hover:text-parchment"
              >
                View the benchmark index
                <ArrowUpRight className="h-4 w-4" />
              </Link>
            </div>
          </div>

          {/* Live confirmation card */}
          <div className="mt-16 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
            {proofStats.map((s) => (
              <div key={s.label} className="card card-raise p-5">
                <div className="eyebrow">{s.label}</div>
                <div
                  className={`mt-2 tabular text-2xl font-semibold ${
                    s.verify ? "text-verify" : "text-parchment"
                  }`}
                >
                  {s.value}
                </div>
                <div className="mt-1 text-xs text-muted-soft">{s.sub}</div>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Tagline reveal */}
      <section className="border-b border-line bg-ink-900/60">
        <div className="mx-auto max-w-4xl px-4 py-24 text-center sm:px-6 lg:px-8">
          <p className="eyebrow">The institutional standard</p>
          <p className="mt-5 font-serif text-headline text-parchment">
            Compute is not an abstract token. It is physical silicon in a room with
            thermal and network limits. We make future access to it legally and
            cryptographically <span className="italic text-signal">enforceable</span>.
          </p>
        </div>
      </section>

      {/* Product boundary */}
      <section className="border-b border-line">
        <div className="mx-auto max-w-6xl px-4 py-24 sm:px-6 lg:px-8">
          <div className="max-w-2xl">
            <p className="eyebrow">The product boundary</p>
            <h2 className="mt-3 font-serif text-headline text-parchment">
              One product. Not the other three.
            </h2>
            <p className="mt-4 text-muted">
              Conflating these four models in code is how the legal position gets lost.
              Verinode builds exactly one — the physical reservation — and refuses the rest.
            </p>
          </div>

          <div className="mt-12 overflow-hidden rounded-card border border-line">
            {boundary.map((row, i) => (
              <div
                key={row.model}
                className={`grid grid-cols-1 gap-2 px-6 py-5 md:grid-cols-[1.1fr_1.4fr_1.4fr_1.2fr] md:gap-6 ${
                  i !== 0 ? "border-t border-line" : ""
                } ${row.ours ? "bg-signal-wash/60" : "bg-ink-850/40"}`}
              >
                <div className="flex items-center gap-2">
                  {row.ours ? (
                    <CircleCheck className="h-4 w-4 shrink-0 text-signal" />
                  ) : (
                    <CircleDashed className="h-4 w-4 shrink-0 text-muted-soft" />
                  )}
                  <span
                    className={`text-sm font-semibold ${
                      row.ours ? "text-signal-bright" : "text-parchment"
                    }`}
                  >
                    {row.model}
                  </span>
                </div>
                <div className="text-sm text-muted">{row.promise}</div>
                <div className="text-sm text-muted">{row.risk}</div>
                <div
                  className={`text-sm ${
                    row.ours ? "font-medium text-signal" : "text-muted-soft"
                  }`}
                >
                  {row.position}
                </div>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Guarantees */}
      <section className="border-b border-line bg-ink-900/60">
        <div className="mx-auto max-w-6xl px-4 py-24 sm:px-6 lg:px-8">
          <div className="max-w-2xl">
            <p className="eyebrow">Institutional guarantees</p>
            <h2 className="mt-3 font-serif text-headline text-parchment">
              Built for training teams and compute funds.
            </h2>
          </div>

          <div className="mt-12 grid gap-5 md:grid-cols-3">
            {guarantees.map((g) => (
              <div key={g.title} className="card p-7 transition-colors hover:border-ink-700">
                <div className="flex h-10 w-10 items-center justify-center rounded-xl border border-line bg-ink-850">
                  <g.icon className="h-5 w-5 text-signal" />
                </div>
                <h3 className="mt-5 font-serif text-title text-parchment">{g.title}</h3>
                <p className="mt-3 text-sm leading-relaxed text-muted">{g.body}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* How it works */}
      <section className="border-b border-line">
        <div className="mx-auto max-w-6xl px-4 py-24 sm:px-6 lg:px-8">
          <div className="max-w-2xl">
            <p className="eyebrow">Execution lifecycle</p>
            <h2 className="mt-3 font-serif text-headline text-parchment">
              From private RFQ to certified delivery.
            </h2>
          </div>

          <div className="mt-12 grid gap-5 md:grid-cols-3">
            {steps.map((s) => (
              <div key={s.n} className="relative card p-7">
                <span className="tabular text-3xl font-semibold text-ink-700">{s.n}</span>
                <h3 className="mt-3 font-serif text-title text-parchment">{s.title}</h3>
                <p className="mt-3 text-sm leading-relaxed text-muted">{s.body}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Grade spec */}
      <section id="grades" className="border-b border-line bg-ink-900/60">
        <div className="mx-auto max-w-5xl px-4 py-24 sm:px-6 lg:px-8">
          <div className="max-w-2xl">
            <p className="eyebrow">Standardized physical commodity</p>
            <h2 className="mt-3 font-serif text-headline text-parchment">
              Grade H100-SXM-8XNV
            </h2>
            <p className="mt-4 text-muted">
              Like a delivery grade on any physical exchange, forward liquidity needs a
              deterministic definition of what is delivered. This is ours.
            </p>
          </div>

          <div className="card card-raise mt-10 p-8">
            <div className="flex flex-col gap-4 border-b border-line pb-6 sm:flex-row sm:items-center sm:justify-between">
              <div>
                <h3 className="font-serif text-2xl text-parchment">8× NVIDIA H100 SXM5 · 80GB HBM3</h3>
                <p className="tabular mt-1 text-xs text-muted-soft">GRADE ID · H100-SXM-8XNV</p>
              </div>
              <span className="inline-flex w-fit items-center gap-2 rounded-pill border border-verify/40 bg-verify-wash px-3 py-1.5 text-xs font-medium text-verify">
                <span className="h-1.5 w-1.5 rounded-full bg-verify" /> Benchmark grade active
              </span>
            </div>

            <div className="mt-6 grid grid-cols-2 gap-px overflow-hidden rounded-xl border border-line bg-line sm:grid-cols-4">
              {[
                { k: "VRAM pool", v: "640 GB", s: "HBM3 · 3.35 TB/s" },
                { k: "NVLink mesh", v: "900 GB/s", s: "NVSwitch 4.0" },
                { k: "Host", v: "112 cores", s: "1,024 GB DDR5" },
                { k: "AllReduce floor", v: "≥ 400 GB/s", s: "synthetic canary", verify: true },
              ].map((c) => (
                <div key={c.k} className="bg-ink-850 p-5">
                  <div className="eyebrow">{c.k}</div>
                  <div className={`tabular mt-2 text-lg font-semibold ${c.verify ? "text-verify" : "text-parchment"}`}>
                    {c.v}
                  </div>
                  <div className="mt-0.5 text-xs text-muted-soft">{c.s}</div>
                </div>
              ))}
            </div>
          </div>
        </div>
      </section>

      {/* FAQ */}
      <section className="border-b border-line">
        <div className="mx-auto max-w-3xl px-4 py-24 sm:px-6 lg:px-8">
          <div className="text-center">
            <p className="eyebrow">Legal, verification & settlement</p>
            <h2 className="mt-3 font-serif text-headline text-parchment">Clarity, on the record.</h2>
          </div>

          <div className="mt-12 divide-y divide-line">
            {faqs.map((f) => (
              <details key={f.q} className="group py-5">
                <summary className="flex cursor-pointer list-none items-center justify-between gap-4">
                  <span className="font-serif text-lg text-parchment">{f.q}</span>
                  <span className="text-signal transition-transform group-open:rotate-45">
                    <ArrowUpRight className="h-5 w-5 rotate-45 group-open:rotate-0" />
                  </span>
                </summary>
                <p className="mt-3 text-sm leading-relaxed text-muted">{f.a}</p>
              </details>
            ))}
          </div>
        </div>
      </section>

      {/* CTA */}
      <section className="bg-ink-900/60">
        <div className="mx-auto max-w-3xl px-4 py-24 text-center sm:px-6 lg:px-8">
          <ShieldCheck className="mx-auto h-8 w-8 text-signal" />
          <h2 className="mt-6 font-serif text-headline text-parchment">
            Reserve capacity with recourse.
          </h2>
          <p className="mx-auto mt-4 max-w-xl text-muted">
            Trade spot preemption and opaque brokers for a standardized forward backed by
            independent hardware proofs and balanced escrow.
          </p>
          <div className="mt-9 flex flex-wrap items-center justify-center gap-3">
            <Link
              href="/buyer/rfqs/new"
              className="group inline-flex items-center gap-2 rounded-xl bg-signal px-5 py-3 text-sm font-semibold text-ink-950 transition-all hover:bg-signal-bright active:scale-[0.98]"
            >
              Submit a private RFQ
              <ArrowRight className="h-4 w-4 transition-transform group-hover:translate-x-0.5" />
            </Link>
            <Link
              href="/seller"
              className="inline-flex items-center gap-2 rounded-xl border border-line bg-ink-850 px-5 py-3 text-sm font-semibold text-parchment transition-colors hover:bg-ink-800"
            >
              List cluster inventory
            </Link>
          </div>
        </div>
      </section>
    </div>
  );
}
