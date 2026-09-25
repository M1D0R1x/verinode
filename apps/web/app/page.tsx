import Link from "next/link";
import {
  ShieldCheck,
  Cpu,
  ArrowRight,
  CheckCircle2,
  FileText,
  Lock,
  Activity,
  Layers,
  BarChart3,
  Server,
  Terminal,
} from "lucide-react";

export default function HomePage() {
  return (
    <div className="flex flex-col min-h-screen">
      {/* HERO SECTION */}
      <section className="relative overflow-hidden pt-20 pb-24 md:pt-28 md:pb-32 bg-grid border-b border-border/80">
        {/* Glow backdrop */}
        <div className="absolute top-1/4 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[600px] h-[350px] bg-primary/10 blur-[140px] pointer-events-none rounded-full" />

        <div className="relative mx-auto max-w-7xl px-4 sm:px-6 lg:px-8 text-center">
          <div className="inline-flex items-center gap-2 px-3 py-1.5 rounded-full border border-border bg-surface text-xs font-mono text-zinc-300 mb-6 shadow-sm">
            <span className="h-2 w-2 rounded-full bg-accent animate-pulse" />
            <span>Benchmark Grade: 8x H100 SXM 80GB (168-Hour Block)</span>
          </div>

          <h1 className="text-4xl sm:text-5xl md:text-6xl font-extrabold tracking-tight text-white max-w-4xl mx-auto leading-[1.15]">
            Physically Delivered GPU Capacity.{" "}
            <span className="bg-gradient-to-r from-blue-400 via-indigo-300 to-emerald-400 bg-clip-text text-transparent">
              Legally & Cryptographically Enforceable.
            </span>
          </h1>

          <p className="mt-6 text-base sm:text-lg text-muted max-w-2xl mx-auto leading-relaxed">
            Trade bilateral, non-transferable physical reservations for verified enterprise GPU nodes. Built on standardized forward contracts, host-level ed25519 hardware telemetry, and institutional banking escrow.
          </p>

          <div className="mt-10 flex flex-wrap items-center justify-center gap-4">
            <Link
              href="/buyer/rfqs/new"
              className="inline-flex items-center gap-2 px-6 py-3 rounded-lg bg-primary hover:bg-primary-hover text-white font-medium text-sm transition-all shadow-lg shadow-primary/20"
            >
              <span>Submit Private RFQ</span>
              <ArrowRight className="h-4 w-4" />
            </Link>

            <Link
              href="/seller"
              className="inline-flex items-center gap-2 px-6 py-3 rounded-lg bg-surface hover:bg-surfaceSubtle border border-border hover:border-borderHighlight text-white font-medium text-sm transition-all"
            >
              <span>List GPU Capacity</span>
            </Link>
          </div>

          {/* Quick Metrics Bar */}
          <div className="mt-16 grid grid-cols-2 md:grid-cols-4 gap-4 max-w-4xl mx-auto text-left">
            <div className="p-4 rounded-xl bg-surface/80 border border-border">
              <span className="text-xs text-muted font-mono uppercase">Delivery Unit</span>
              <p className="text-lg font-bold text-white mt-1">168 Hours (1 Wk)</p>
              <span className="text-[11px] text-zinc-400">Continuous take-or-pay</span>
            </div>

            <div className="p-4 rounded-xl bg-surface/80 border border-border">
              <span className="text-xs text-muted font-mono uppercase">Interconnect Floor</span>
              <p className="text-lg font-bold text-emerald-400 mt-1">$\ge$ 400 GB/s</p>
              <span className="text-[11px] text-zinc-400">NCCL All-Reduce synthetic</span>
            </div>

            <div className="p-4 rounded-xl bg-surface/80 border border-border">
              <span className="text-xs text-muted font-mono uppercase">Regulatory Posture</span>
              <p className="text-lg font-bold text-blue-400 mt-1">Forward Exclusion</p>
              <span className="text-[11px] text-zinc-400">CFTC physical commodity</span>
            </div>

            <div className="p-4 rounded-xl bg-surface/80 border border-border">
              <span className="text-xs text-muted font-mono uppercase">Host Telemetry</span>
              <p className="text-lg font-bold text-purple-400 mt-1">ed25519 Signed</p>
              <span className="text-[11px] text-zinc-400">Zero customer data ingested</span>
            </div>
          </div>
        </div>
      </section>

      {/* THE 4 OFFERINGS COMPARISON */}
      <section className="py-20 bg-background border-b border-border">
        <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
          <div className="text-center max-w-3xl mx-auto mb-14">
            <h2 className="text-2xl sm:text-3xl font-bold text-white tracking-tight">
              The Product Boundary: Why Physical Delivery Matters
            </h2>
            <p className="mt-3 text-sm text-muted">
              Verinode builds exactly one product: standardized physical forward reservations. We do not conflate spot computing with cash derivatives.
            </p>
          </div>

          <div className="overflow-x-auto">
            <table className="w-full text-left border-collapse text-xs sm:text-sm">
              <thead>
                <tr className="border-b border-border text-muted uppercase font-mono text-[11px]">
                  <th className="py-3 px-4">Offering Type</th>
                  <th className="py-3 px-4">Core Promise</th>
                  <th className="py-3 px-4">Primary Risk</th>
                  <th className="py-3 px-4">Verinode Lane</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border/60">
                <tr className="hover:bg-surface/30 transition-colors">
                  <td className="py-4 px-4 font-semibold text-zinc-300">Spot Capacity Marketplace</td>
                  <td className="py-4 px-4 text-zinc-400">Compute available now or soon</td>
                  <td className="py-4 px-4 text-zinc-400">Interruption, preemption, variable topology</td>
                  <td className="py-4 px-4">
                    <span className="inline-flex px-2 py-0.5 rounded text-[11px] font-mono bg-zinc-800 text-zinc-400">
                      Not Built (AWS/RunPod)
                    </span>
                  </td>
                </tr>

                <tr className="bg-primary/5 hover:bg-primary/10 transition-colors border-l-2 border-l-primary">
                  <td className="py-4 px-4 font-bold text-white flex items-center gap-2">
                    <CheckCircle2 className="h-4 w-4 text-accent" />
                    <span>Physical Forward Reservation</span>
                  </td>
                  <td className="py-4 px-4 text-white font-medium">Specified future node at fixed price, verified delivery</td>
                  <td className="py-4 px-4 text-zinc-300">Delivery delay & counterparty default</td>
                  <td className="py-4 px-4">
                    <span className="inline-flex px-2 py-0.5 rounded text-[11px] font-mono bg-emerald-950 text-emerald-300 border border-emerald-800 font-bold">
                      VERINODE CORE MOAT
                    </span>
                  </td>
                </tr>

                <tr className="hover:bg-surface/30 transition-colors">
                  <td className="py-4 px-4 font-semibold text-zinc-300">Financial Derivative (Cash-Settled)</td>
                  <td className="py-4 px-4 text-zinc-400">Cash payout based on a price index</td>
                  <td className="py-4 px-4 text-zinc-400">Leverage, basis risk, speculative licensing</td>
                  <td className="py-4 px-4">
                    <span className="inline-flex px-2 py-0.5 rounded text-[11px] font-mono bg-zinc-800 text-zinc-400">
                      Not Built (CME / ICE)
                    </span>
                  </td>
                </tr>

                <tr className="hover:bg-surface/30 transition-colors">
                  <td className="py-4 px-4 font-semibold text-zinc-300">Enterprise Bilateral Contract</td>
                  <td className="py-4 px-4 text-zinc-400">Bespoke SLA, enterprise support, tailored liability</td>
                  <td className="py-4 px-4 text-zinc-400">Slow procurement, non-standard terms</td>
                  <td className="py-4 px-4">
                    <span className="inline-flex px-2 py-0.5 rounded text-[11px] font-mono bg-zinc-800 text-zinc-300">
                      Master Confirmation Wrapper
                    </span>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </section>

      {/* STANDARDIZED GRADE SPECIFICATION */}
      <section id="grades" className="py-20 bg-surface/30 border-b border-border">
        <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
          <div className="flex flex-col md:flex-row md:items-end justify-between mb-12">
            <div>
              <span className="text-xs font-mono text-primary uppercase tracking-wider">Delivery Grade Ontology</span>
              <h2 className="text-2xl sm:text-3xl font-bold text-white tracking-tight mt-1">
                Benchmark Spec: H100-SXM-8XNV
              </h2>
            </div>
            <p className="mt-3 md:mt-0 text-xs text-muted max-w-md">
              Every contract specifies an exact hardware grade with zero permissible "equivalent" substitutions. If it doesn't match the benchmark floor, it isn't delivered.
            </p>
          </div>

          <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
            {/* Spec Details Card */}
            <div className="lg:col-span-2 p-6 rounded-2xl bg-surface border border-border space-y-6">
              <div className="flex items-center justify-between border-b border-border/80 pb-4">
                <div className="flex items-center gap-3">
                  <div className="p-2.5 rounded-lg bg-surfaceSubtle border border-border">
                    <Server className="h-6 w-6 text-primary" />
                  </div>
                  <div>
                    <h3 className="font-bold text-lg text-white">8x NVIDIA H100 SXM5 80GB</h3>
                    <p className="text-xs text-muted font-mono">Series: H100-SXM-8XNV-US-WEEK-DEDICATED-USD</p>
                  </div>
                </div>
                <span className="px-2.5 py-1 rounded-full text-xs font-mono bg-emerald-950/80 text-emerald-300 border border-emerald-800">
                  Standard Benchmark
                </span>
              </div>

              <div className="grid grid-cols-2 sm:grid-cols-3 gap-4 text-xs font-mono">
                <div className="p-3 rounded-lg bg-surfaceSubtle border border-border/60">
                  <span className="text-muted block text-[10px]">TOTAL VRAM</span>
                  <span className="text-white text-base font-bold mt-0.5 block">640 GB HBM3</span>
                  <span className="text-zinc-500 text-[10px]">8x 80GB GPUs</span>
                </div>

                <div className="p-3 rounded-lg bg-surfaceSubtle border border-border/60">
                  <span className="text-muted block text-[10px]">INTERCONNECT</span>
                  <span className="text-white text-base font-bold mt-0.5 block">NVLink 4.0</span>
                  <span className="text-zinc-500 text-[10px]">900 GB/s bidirectional</span>
                </div>

                <div className="p-3 rounded-lg bg-surfaceSubtle border border-border/60">
                  <span className="text-muted block text-[10px]">BENCHMARK FLOOR</span>
                  <span className="text-emerald-400 text-base font-bold mt-0.5 block">400.0 GB/s</span>
                  <span className="text-zinc-500 text-[10px]">NCCL All-Reduce</span>
                </div>

                <div className="p-3 rounded-lg bg-surfaceSubtle border border-border/60">
                  <span className="text-muted block text-[10px]">HOST CPU CORES</span>
                  <span className="text-white text-base font-bold mt-0.5 block">112 Cores</span>
                  <span className="text-zinc-500 text-[10px]">Dual Intel Xeon / AMD EPYC</span>
                </div>

                <div className="p-3 rounded-lg bg-surfaceSubtle border border-border/60">
                  <span className="text-muted block text-[10px]">SYSTEM RAM</span>
                  <span className="text-white text-base font-bold mt-0.5 block">1,024 GB</span>
                  <span className="text-zinc-500 text-[10px]">DDR5 ECC registered</span>
                </div>

                <div className="p-3 rounded-lg bg-surfaceSubtle border border-border/60">
                  <span className="text-muted block text-[10px]">NVME SCRATCH IOPS</span>
                  <span className="text-white text-base font-bold mt-0.5 block">100,000 IOPS</span>
                  <span className="text-zinc-500 text-[10px]">Direct local scratch storage</span>
                </div>
              </div>

              <div className="p-4 rounded-xl bg-surfaceSubtle/50 border border-border text-xs text-muted space-y-1.5">
                <span className="font-semibold text-zinc-300 block">Hardware Acceptance Rule:</span>
                <p>
                  Delivery only reaches <span className="font-mono text-emerald-400">LIVE</span> state when SSH/IPMI credentials verify, scheduler shows dedicated capacity, attestation matches SXM5 NVLink topology, and the synthetic NCCL all-reduce canary completes $\ge$ 400 GB/s with 0 unrecovered ECC errors.
                </p>
              </div>
            </div>

            {/* Quick Action Sidecard */}
            <div className="p-6 rounded-2xl bg-surface border border-border flex flex-col justify-between">
              <div>
                <div className="flex items-center gap-2 text-xs font-mono text-primary mb-3">
                  <Activity className="h-4 w-4" />
                  <span>Market Availability</span>
                </div>
                <h3 className="font-bold text-white text-lg">Reserve an H100 Node</h3>
                <p className="text-xs text-muted mt-2 leading-relaxed">
                  Submit a private RFQ to vetted suppliers. Review firm quotes, execute tamper-evident confirmations, and fund via escrow.
                </p>

                <div className="mt-6 space-y-3 font-mono text-xs">
                  <div className="flex justify-between py-1.5 border-b border-border/60">
                    <span className="text-muted">Standard Tenor:</span>
                    <span className="text-white font-medium">168 Hours (7 Days)</span>
                  </div>
                  <div className="flex justify-between py-1.5 border-b border-border/60">
                    <span className="text-muted">Regions:</span>
                    <span className="text-white font-medium">US-East, US-West, EU</span>
                  </div>
                  <div className="flex justify-between py-1.5 border-b border-border/60">
                    <span className="text-muted">Tenancy:</span>
                    <span className="text-white font-medium">100% Bare-Metal Dedicated</span>
                  </div>
                </div>
              </div>

              <div className="mt-8 pt-4 border-t border-border">
                <Link
                  href="/buyer/rfqs/new"
                  className="w-full flex items-center justify-center gap-2 py-3 rounded-lg bg-primary hover:bg-primary-hover text-white text-xs font-semibold shadow-sm transition-all"
                >
                  <span>Create Forward RFQ</span>
                  <ArrowRight className="h-4 w-4" />
                </Link>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* HOW VERINODE WORKS - 4 PILLARS */}
      <section className="py-20 bg-background border-b border-border">
        <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
          <div className="text-center max-w-2xl mx-auto mb-16">
            <span className="text-xs font-mono text-accent uppercase tracking-wider">End-to-End Architecture</span>
            <h2 className="text-2xl sm:text-3xl font-bold text-white tracking-tight mt-1">
              Engineered for Institutional Reliability
            </h2>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
            <div className="p-6 rounded-xl bg-surface border border-border hover:border-primary/50 transition-all group">
              <div className="h-10 w-10 rounded-lg bg-surfaceSubtle border border-border flex items-center justify-center text-primary mb-4 group-hover:scale-105 transition-transform">
                <FileText className="h-5 w-5" />
              </div>
              <h3 className="font-bold text-white text-sm">1. Standardized Contract</h3>
              <p className="text-xs text-muted mt-2 leading-relaxed">
                Versioned confirmations executed under master agreements. Strict parameters eliminate ambiguous SLAs.
              </p>
            </div>

            <div className="p-6 rounded-xl bg-surface border border-border hover:border-primary/50 transition-all group">
              <div className="h-10 w-10 rounded-lg bg-surfaceSubtle border border-border flex items-center justify-center text-accent mb-4 group-hover:scale-105 transition-transform">
                <Lock className="h-5 w-5" />
              </div>
              <h3 className="font-bold text-white text-sm">2. Balanced Escrow</h3>
              <p className="text-xs text-muted mt-2 leading-relaxed">
                Double-entry financial subledger reconciled against bank statements. Funds release only upon verified delivery.
              </p>
            </div>

            <div className="p-6 rounded-xl bg-surface border border-border hover:border-primary/50 transition-all group">
              <div className="h-10 w-10 rounded-lg bg-surfaceSubtle border border-border flex items-center justify-center text-purple-400 mb-4 group-hover:scale-105 transition-transform">
                <Terminal className="h-5 w-5" />
              </div>
              <h3 className="font-bold text-white text-sm">3. Cryptographic Proof</h3>
              <p className="text-xs text-muted mt-2 leading-relaxed">
                Host agent signs hardware telemetry with ed25519. Synthetic canaries verify bandwidth before tenant login.
              </p>
            </div>

            <div className="p-6 rounded-xl bg-surface border border-border hover:border-primary/50 transition-all group">
              <div className="h-10 w-10 rounded-lg bg-surfaceSubtle border border-border flex items-center justify-center text-blue-400 mb-4 group-hover:scale-105 transition-transform">
                <BarChart3 className="h-5 w-5" />
              </div>
              <h3 className="font-bold text-white text-sm">4. Independent Index</h3>
              <p className="text-xs text-muted mt-2 leading-relaxed">
                Isolated trust domain publishes volume-weighted fixes. Enforces strict `insufficient_data` integrity gates.
              </p>
            </div>
          </div>
        </div>
      </section>

      {/* CTA SECTION */}
      <section className="py-20 bg-surface/50">
        <div className="mx-auto max-w-5xl px-4 sm:px-6 lg:px-8 text-center">
          <h2 className="text-3xl font-extrabold text-white tracking-tight">
            Ready to secure guaranteed compute capacity?
          </h2>
          <p className="mt-4 text-sm text-muted max-w-xl mx-auto">
            Join leading AI teams and GPU clouds using Verinode to trade standardized physical compute reservations.
          </p>
          <div className="mt-8 flex flex-wrap justify-center gap-4">
            <Link
              href="/buyer/rfqs/new"
              className="px-6 py-3 rounded-lg bg-primary hover:bg-primary-hover text-white font-medium text-sm transition-all"
            >
              Submit an RFQ
            </Link>
            <Link
              href="/seller"
              className="px-6 py-3 rounded-lg bg-surface border border-border hover:border-borderHighlight text-white font-medium text-sm transition-all"
            >
              Onboard as Supplier
            </Link>
          </div>
        </div>
      </section>
    </div>
  );
}
