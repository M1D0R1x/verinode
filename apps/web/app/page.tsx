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
  HelpCircle,
} from "lucide-react";

export default function HomePage() {
  return (
    <div className="flex flex-col min-h-screen bg-black text-[#EDEDED]">
      {/* 1. HERO SECTION */}
      <section className="relative pt-20 pb-24 md:pt-28 md:pb-32 border-b border-[#272727]">
        <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
          <div className="text-center max-w-4xl mx-auto space-y-6">
            {/* Status Pill */}
            <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full border border-[#272727] bg-[#121212] text-xs font-mono text-zinc-300">
              <span className="h-1.5 w-1.5 rounded-full bg-emerald-400" />
              <span>Benchmark Grade: 8x H100 SXM 80GB (168-Hour Block)</span>
            </div>

            {/* Main Headline with subtle left-to-right gradient (#FFFFFF -> #9B9B9B) per landing-page-design B5 */}
            <h1 className="text-4xl sm:text-5xl md:text-6xl font-extrabold tracking-tight text-transparent bg-clip-text bg-gradient-to-r from-white via-zinc-100 to-[#9B9B9B] leading-[1.1]">
              Physical GPU capacity. Guaranteed delivery.
            </h1>

            {/* Subheadline: clarify what it is, outcome plus audience */}
            <p className="text-base sm:text-lg text-zinc-400 max-w-2xl mx-auto leading-relaxed">
              Bilateral, non-transferable capacity forward reservations for verified enterprise GPU clusters. Standardized contracts, ed25519 host canary verification, and balanced double-entry escrow.
            </p>

            {/* CTAs: Emil Kowalski active state + landing-page-design button sizes */}
            <div className="pt-2 flex flex-wrap items-center justify-center gap-4">
              <Link
                href="/buyer/rfqs/new"
                className="inline-flex items-center gap-2 px-4 py-2.5 rounded-lg bg-white text-zinc-950 hover:bg-zinc-200 font-semibold text-sm transition-colors duration-150 ease-out active:scale-[0.98] shadow-sm"
              >
                <span>Submit Allocation RFQ</span>
                <ArrowRight className="h-4 w-4 text-zinc-950" />
              </Link>

              <Link
                href="/seller"
                className="inline-flex items-center gap-2 px-4 py-2.5 rounded-lg bg-[#181818] hover:bg-[#222222] border border-[#272727] text-white font-semibold text-sm transition-colors duration-150 ease-out active:scale-[0.98]"
              >
                <span>List Hardware Capacity</span>
              </Link>
            </div>

            {/* Proof Signal Bar */}
            <div className="pt-8 grid grid-cols-2 md:grid-cols-4 gap-3 max-w-4xl mx-auto text-left text-xs font-mono">
              <div className="p-3.5 rounded-xl bg-[#121212] border border-[#272727]">
                <span className="text-zinc-500 uppercase block text-xs">Standardized Tenor</span>
                <p className="text-base font-bold text-white mt-1">168 Hours (1 Wk)</p>
                <span className="text-zinc-400 text-xs">Continuous take-or-pay</span>
              </div>

              <div className="p-3.5 rounded-xl bg-[#121212] border border-[#272727]">
                <span className="text-zinc-500 uppercase block text-xs">NCCL Floor Gate</span>
                <p className="text-base font-bold text-emerald-400 mt-1">&ge; 400 GB/s</p>
                <span className="text-zinc-400 text-xs">Pre-delivery canary pass</span>
              </div>

              <div className="p-3.5 rounded-xl bg-[#121212] border border-[#272727]">
                <span className="text-zinc-500 uppercase block text-xs">Legal Perimeter</span>
                <p className="text-base font-bold text-white mt-1">Forward Exclusion</p>
                <span className="text-zinc-400 text-xs">CFTC physical commodity</span>
              </div>

              <div className="p-3.5 rounded-xl bg-[#121212] border border-[#272727]">
                <span className="text-zinc-500 uppercase block text-xs">Zero Workload Ingest</span>
                <p className="text-base font-bold text-emerald-400 mt-1">ed25519 Signed</p>
                <span className="text-zinc-400 text-xs">No code or weights touched</span>
              </div>
            </div>
          </div>

          {/* Hero Visual: Institutional Terminal Preview */}
          <div className="mt-14 max-w-4xl mx-auto rounded-xl border border-[#272727] bg-[#121212] overflow-hidden shadow-2xl">
            {/* Terminal Header */}
            <div className="px-4 py-3 bg-[#181818] border-b border-[#272727] flex items-center justify-between">
              <div className="flex items-center gap-2">
                <span className="h-2.5 w-2.5 rounded-full bg-zinc-600" />
                <span className="h-2.5 w-2.5 rounded-full bg-zinc-600" />
                <span className="h-2.5 w-2.5 rounded-full bg-zinc-600" />
                <span className="ml-2 text-xs font-mono text-zinc-400">
                  VERINODE CONTRACT EXECUTION DESK — TRADE #c37610f3
                </span>
              </div>
              <div className="flex items-center gap-2 text-xs font-mono text-emerald-400">
                <span className="h-1.5 w-1.5 rounded-full bg-emerald-400 animate-pulse" />
                <span>CANARY PASS VERIFIED</span>
              </div>
            </div>

            {/* Terminal Body */}
            <div className="p-6 font-mono text-xs space-y-4">
              <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                <div className="p-3 rounded-lg bg-[#000000] border border-[#272727] space-y-1">
                  <span className="text-zinc-500 block text-xs">BENCHMARK SPECIFICATION</span>
                  <div className="text-white font-semibold">8x NVIDIA H100 SXM5</div>
                  <div className="text-zinc-400">640GB HBM3 | NVLink 4.0 900 GB/s</div>
                </div>

                <div className="p-3 rounded-lg bg-[#000000] border border-[#272727] space-y-1">
                  <span className="text-zinc-500 block text-xs">CANARY BENCHMARK RESULT</span>
                  <div className="text-emerald-400 font-semibold">405.2 GB/s AllReduce</div>
                  <div className="text-zinc-400">Floor &ge; 400.0 GB/s (100% Passed)</div>
                </div>

                <div className="p-3 rounded-lg bg-[#000000] border border-[#272727] space-y-1">
                  <span className="text-zinc-500 block text-xs">ESCROW BALANCE (INVARIANT 3)</span>
                  <div className="text-white font-semibold">$29,568.00 USD</div>
                  <div className="text-emerald-400">Subledger Balanced (Sum == 0)</div>
                </div>
              </div>

              {/* Legal Confirmation Snippet */}
              <div className="p-3 rounded-lg bg-[#000000] border border-[#272727] text-zinc-400 flex flex-col sm:flex-row sm:items-center justify-between gap-2">
                <div className="flex items-center gap-2">
                  <FileText className="h-4 w-4 text-zinc-300" />
                  <span>Master Legal Confirmation SHA-256:</span>
                  <span className="text-white font-mono font-bold">4adec53cc851db49c10c317c5fc507e31...</span>
                </div>
                <Link
                  href="/buyer"
                  className="text-xs text-white hover:underline flex items-center gap-1 font-semibold"
                >
                  <span>Open Buyer Desk</span>
                  <ArrowRight className="h-3 w-3" />
                </Link>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* 2. TAGLINE REVEAL SECTION (Mandatory per landing-page-design B11) */}
      <section className="py-20 border-b border-[#272727] bg-[#121212]">
        <div className="mx-auto max-w-5xl px-4 sm:px-6 lg:px-8 text-center space-y-4">
          <span className="text-xs font-mono text-zinc-500 uppercase tracking-widest block">
            The Institutional Standard
          </span>
          <p className="text-2xl sm:text-3xl md:text-4xl font-bold tracking-tight text-white leading-snug">
            Compute is not an abstract token. It is physical silicon in a data center with thermal and network limits. We make future access legally and cryptographically enforceable.
          </p>
        </div>
      </section>

      {/* 3. PROBLEM TO SOLUTION: THE 4 OFFERINGS BOUNDARY */}
      <section className="py-24 border-b border-[#272727] bg-black">
        <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8 space-y-12">
          <div className="text-center max-w-3xl mx-auto space-y-3">
            <h2 className="text-2xl sm:text-3xl font-bold text-white tracking-tight">
              The Product Boundary: Why Physical Delivery Matters
            </h2>
            <p className="text-sm text-zinc-400">
              Verinode builds exactly one product: standardized physical forward reservations. We do not build continuous spot markets or cash-settled synthetic derivatives.
            </p>
          </div>

          <div className="overflow-x-auto rounded-xl border border-[#272727] bg-[#121212]">
            <table className="w-full text-left border-collapse text-xs sm:text-sm">
              <thead>
                <tr className="border-b border-[#272727] text-zinc-400 uppercase font-mono text-xs bg-[#181818]">
                  <th className="py-3.5 px-4 font-semibold">Offering Model</th>
                  <th className="py-3.5 px-4 font-semibold">Platform Promise</th>
                  <th className="py-3.5 px-4 font-semibold">Counterparty Risk</th>
                  <th className="py-3.5 px-4 font-semibold">Verinode Position</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-[#272727] font-mono text-xs">
                <tr className="hover:bg-[#181818] transition-colors duration-150">
                  <td className="py-4 px-4 font-semibold text-white">Spot Capacity Market</td>
                  <td className="py-4 px-4 text-zinc-400">Burst compute right now</td>
                  <td className="py-4 px-4 text-zinc-400">Preemption, sudden price spikes, degraded topology</td>
                  <td className="py-4 px-4 text-zinc-500">Excluded (AWS, RunPod lane)</td>
                </tr>
                <tr className="hover:bg-[#181818] transition-colors duration-150 bg-[#181818]/60">
                  <td className="py-4 px-4 font-bold text-emerald-400 flex items-center gap-1.5">
                    <CheckCircle2 className="h-4 w-4 text-emerald-400 shrink-0" />
                    Physical Forward Reservation
                  </td>
                  <td className="py-4 px-4 text-white font-medium">Guaranteed future node delivery at fixed price</td>
                  <td className="py-4 px-4 text-zinc-300">Default risk mitigated by bank escrow and canary test</td>
                  <td className="py-4 px-4 text-emerald-400 font-bold">Verinode Core Moat (Invariant 1)</td>
                </tr>
                <tr className="hover:bg-[#181818] transition-colors duration-150">
                  <td className="py-4 px-4 font-semibold text-white">Cash-Settled Synthetic Derivative</td>
                  <td className="py-4 px-4 text-zinc-400">Synthetic cash payout against an index</td>
                  <td className="py-4 px-4 text-zinc-400">Retail leverage, liquidations, cannot run model training</td>
                  <td className="py-4 px-4 text-zinc-500">Excluded (CME, Hyperliquid lane)</td>
                </tr>
                <tr className="hover:bg-[#181818] transition-colors duration-150">
                  <td className="py-4 px-4 font-semibold text-white">Bespoke Enterprise Contract</td>
                  <td className="py-4 px-4 text-zinc-400">Multi-year custom commitments</td>
                  <td className="py-4 px-4 text-zinc-400">Months of legal review, opaque SLAs with zero remedies</td>
                  <td className="py-4 px-4 text-zinc-300">Standardized into Master Confirmation</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </section>

      {/* 4. BENEFITS (Outcome-driven per landing-page-design A2 & A5) */}
      <section className="py-24 border-b border-[#272727] bg-[#121212]">
        <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8 space-y-12">
          <div className="text-center max-w-3xl mx-auto space-y-3">
            <span className="text-xs font-mono text-zinc-500 uppercase tracking-widest">
              Institutional Guarantees
            </span>
            <h2 className="text-2xl sm:text-3xl font-bold text-white tracking-tight">
              Engineered for Enterprise Training Teams and Sovereign Compute Funds
            </h2>
            <p className="text-sm text-zinc-400">
              Every feature solves an acute contract or performance risk that currently plagues the enterprise GPU market.
            </p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            <div className="p-6 rounded-xl bg-black border border-[#272727] space-y-3">
              <div className="h-9 w-9 rounded-lg bg-[#181818] border border-[#272727] flex items-center justify-center text-white">
                <Activity className="h-5 w-5 text-emerald-400" />
              </div>
              <h3 className="font-bold text-base text-white">
                Synthetic canary verification before handover
              </h3>
              <p className="text-xs text-zinc-400 leading-relaxed">
                A host daemon executes an automated NCCL AllReduce test 24 hours prior to delivery start. If bandwidth drops below 400 GB/s, cure remedies or full refunds trigger automatically.
              </p>
            </div>

            <div className="p-6 rounded-xl bg-black border border-[#272727] space-y-3">
              <div className="h-9 w-9 rounded-lg bg-[#181818] border border-[#272727] flex items-center justify-center text-white">
                <Lock className="h-5 w-5 text-zinc-300" />
              </div>
              <h3 className="font-bold text-base text-white">
                Zero customer workload ingestion
              </h3>
              <p className="text-xs text-zinc-400 leading-relaxed">
                Telemetry probes evaluate hardware health, PCI IDs, and NVLink mesh throughput only. Model weights, code, and training prompts never touch our gateway.
              </p>
            </div>

            <div className="p-6 rounded-xl bg-black border border-[#272727] space-y-3">
              <div className="h-9 w-9 rounded-lg bg-[#181818] border border-[#272727] flex items-center justify-center text-white">
                <BarChart3 className="h-5 w-5 text-emerald-400" />
              </div>
              <h3 className="font-bold text-base text-white">
                Balanced double-entry subledger
              </h3>
              <p className="text-xs text-zinc-400 leading-relaxed">
                Escrow funds are isolated in a double-entry subledger where debits and credits strictly sum to zero. Regulated bank transfers and take-or-pay settlements protect both parties.
              </p>
            </div>
          </div>
        </div>
      </section>

      {/* 5. HOW IT WORKS (3 steps per landing-page-design A2) */}
      <section className="py-24 border-b border-[#272727] bg-black">
        <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8 space-y-12">
          <div className="text-center max-w-3xl mx-auto space-y-3">
            <span className="text-xs font-mono text-zinc-500 uppercase tracking-widest">
              Execution Lifecycle
            </span>
            <h2 className="text-2xl sm:text-3xl font-bold text-white tracking-tight">
              From Private Bilateral RFQ to Certified Delivery in 3 Steps
            </h2>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-6 font-mono text-xs">
            <div className="p-6 rounded-xl bg-[#121212] border border-[#272727] space-y-3">
              <span className="text-xs text-zinc-500">STEP 01</span>
              <h3 className="text-sm font-bold text-white font-sans">Submit RFQ & Lock Terms</h3>
              <p className="text-zinc-400 font-sans leading-relaxed">
                Buyer submits private request for designated hardware grade and window. Tier-1 suppliers quote competitive rates. Once matched, the bilateral confirmation generates a deterministic SHA-256 digest.
              </p>
            </div>

            <div className="p-6 rounded-xl bg-[#121212] border border-[#272727] space-y-3">
              <span className="text-xs text-zinc-500">STEP 02</span>
              <h3 className="text-sm font-bold text-white font-sans">Automated Canary Verification</h3>
              <p className="text-zinc-400 font-sans leading-relaxed">
                At T-24h before delivery, host agent executes synthetic NCCL AllReduce and memory tests. The report is ed25519 signed and evaluated by our gateway against the 400 GB/s benchmark floor.
              </p>
            </div>

            <div className="p-6 rounded-xl bg-[#121212] border border-[#272727] space-y-3">
              <span className="text-xs text-zinc-500">STEP 03</span>
              <h3 className="text-sm font-bold text-white font-sans">Handover & Escrow Settlement</h3>
              <p className="text-zinc-400 font-sans leading-relaxed">
                Passing canary unlocks access credentials for the buyer. Hardware runs continuously for the full 168-hour block. Subledger automatically settles escrow payout upon completion.
              </p>
            </div>
          </div>
        </div>
      </section>

      {/* 6. STANDARDIZED BENCHMARK GRADE SPEC */}
      <section id="grades" className="py-24 border-b border-[#272727] bg-[#121212]">
        <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8 space-y-12">
          <div className="text-center max-w-3xl mx-auto space-y-3">
            <span className="text-xs font-mono text-zinc-500 uppercase tracking-widest">
              Standardized Physical Commodity
            </span>
            <h2 className="text-2xl sm:text-3xl font-bold text-white tracking-tight">
              Grade Specification: H100-SXM-8XNV
            </h2>
            <p className="text-sm text-zinc-400">
              Like Light Sweet Crude on NYMEX or 5,000 Bushels on CBOT, compute requires a deterministic physical grade definition to establish forward liquidity.
            </p>
          </div>

          <div className="p-8 rounded-xl bg-black border border-[#272727] max-w-4xl mx-auto space-y-6">
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-[#272727] pb-4">
              <div>
                <h3 className="text-lg font-bold text-white">8x NVIDIA H100 SXM5 (80GB HBM3)</h3>
                <p className="text-xs text-zinc-500 font-mono mt-0.5">GRADE IDENTIFIER: H100-SXM-8XNV</p>
              </div>
              <span className="text-xs font-mono px-3 py-1 rounded bg-emerald-950/50 border border-emerald-800 text-emerald-400 w-fit">
                Benchmark Grade Active
              </span>
            </div>

            <div className="grid grid-cols-2 sm:grid-cols-4 gap-4 text-xs font-mono">
              <div className="p-3 rounded-lg bg-[#121212] border border-[#272727]">
                <span className="text-zinc-500 block text-xs">VRAM POOL</span>
                <span className="text-sm font-bold text-white mt-1 block">640 GB</span>
                <span className="text-zinc-500 text-xs">HBM3 3.35 TB/s</span>
              </div>
              <div className="p-3 rounded-lg bg-[#121212] border border-[#272727]">
                <span className="text-zinc-500 block text-xs">NVLINK MESH</span>
                <span className="text-sm font-bold text-white mt-1 block">900 GB/s</span>
                <span className="text-zinc-500 text-xs">NVSwitch 4.0</span>
              </div>
              <div className="p-3 rounded-lg bg-[#121212] border border-[#272727]">
                <span className="text-zinc-500 block text-xs">SYSTEM HOST</span>
                <span className="text-sm font-bold text-white mt-1 block">112 Cores</span>
                <span className="text-zinc-500 text-xs">1,024 GB DDR5</span>
              </div>
              <div className="p-3 rounded-lg bg-[#121212] border border-[#272727]">
                <span className="text-zinc-500 block text-xs">ALLREDUCE FLOOR</span>
                <span className="text-sm font-bold text-emerald-400 mt-1 block">&ge; 400 GB/s</span>
                <span className="text-zinc-500 text-xs">Synthetic Canary</span>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* 7. INSTITUTIONAL FAQ (6 questions per landing-page-design A2 & A4) */}
      <section className="py-24 border-b border-[#272727] bg-black">
        <div className="mx-auto max-w-4xl px-4 sm:px-6 lg:px-8 space-y-12">
          <div className="text-center space-y-3">
            <span className="text-xs font-mono text-zinc-500 uppercase tracking-widest">
              Frequently Asked Questions
            </span>
            <h2 className="text-2xl sm:text-3xl font-bold text-white tracking-tight">
              Legal, Verification, and Settlement Clarity
            </h2>
          </div>

          <div className="space-y-4 text-xs font-mono">
            <div className="p-5 rounded-xl bg-[#121212] border border-[#272727] space-y-2">
              <h3 className="text-sm font-bold text-white font-sans">
                Why is Verinode structured as a physical forward rather than a derivative?
              </h3>
              <p className="text-zinc-400 font-sans text-xs leading-relaxed">
                Under CFTC commodity regulations, physical forward contracts that result in actual physical delivery of commercial capacity are exempt from swap-dealer registration. Verinode facilitates take-or-pay capacity reservation for training runs, not retail cash-settled speculation.
              </p>
            </div>

            <div className="p-5 rounded-xl bg-[#121212] border border-[#272727] space-y-2">
              <h3 className="text-sm font-bold text-white font-sans">
                How does the hardware canary test protect buyers?
              </h3>
              <p className="text-zinc-400 font-sans text-xs leading-relaxed">
                Prior to delivery start, a cryptographic host agent runs synthetic NCCL AllReduce and memory tests. If the cluster suffers degraded NVLink topology (dropping below 400 GB/s) or unrecovered ECC errors, the supplier must cure or the contract cancels with a 100% escrow refund.
              </p>
            </div>

            <div className="p-5 rounded-xl bg-[#121212] border border-[#272727] space-y-2">
              <h3 className="text-sm font-bold text-white font-sans">
                Does Verinode ever touch customer training data or model weights?
              </h3>
              <p className="text-zinc-400 font-sans text-xs leading-relaxed">
                Never (Invariant 4). Our host agent probes hardware metrics only (PCI IDs, DCGM health, synthetic bandwidth). Customer workloads, weights, training code, and SSH sessions are completely isolated from our telemetry gateway.
              </p>
            </div>

            <div className="p-5 rounded-xl bg-[#121212] border border-[#272727] space-y-2">
              <h3 className="text-sm font-bold text-white font-sans">
                How is collateral and payment escrow managed?
              </h3>
              <p className="text-zinc-400 font-sans text-xs leading-relaxed">
                All transactions execute through our balanced double-entry subledger where credits and debits strictly balance to zero. Escrow deposits are held via institutional banking rails (Fedwire) with automatic payout to suppliers upon successful completion.
              </p>
            </div>

            <div className="p-5 rounded-xl bg-[#121212] border border-[#272727] space-y-2">
              <h3 className="text-sm font-bold text-white font-sans">
                What role do Solana, Arbitrum, and Hyperliquid play?
              </h3>
              <p className="text-zinc-400 font-sans text-xs leading-relaxed">
                Blockchains are optional Phase 4 public audit mirrors (Invariant 6). The off-chain PostgreSQL database and executed legal confirmations remain authoritative. We mirror state transitions and canary attestation proofs to Solana Devnet and Arbitrum Sepolia for public verification.
              </p>
            </div>
          </div>
        </div>
      </section>

      {/* 8. FINAL CTA (Risk Reversal per landing-page-design A2 & A4) */}
      <section className="py-24 bg-[#121212]">
        <div className="mx-auto max-w-4xl px-4 sm:px-6 lg:px-8 text-center space-y-6">
          <h2 className="text-3xl sm:text-4xl font-extrabold text-white tracking-tight">
            Reserve Enterprise GPU Capacity with Legal Recourse
          </h2>
          <p className="text-sm sm:text-base text-zinc-400 max-w-xl mx-auto leading-relaxed">
            Eliminate spot preemption and opaque brokers. Execute standardized forward reservations backed by independent hardware telemetry proofs.
          </p>

          <div className="pt-2 flex flex-wrap items-center justify-center gap-4">
            <Link
              href="/buyer/rfqs/new"
              className="inline-flex items-center gap-2 px-5 py-2.5 rounded-lg bg-white text-zinc-950 hover:bg-zinc-200 font-semibold text-sm transition-colors duration-150 ease-out active:scale-[0.98] shadow-sm"
            >
              <span>Submit Private RFQ</span>
              <ArrowRight className="h-4 w-4 text-zinc-950" />
            </Link>

            <Link
              href="/seller"
              className="inline-flex items-center gap-2 px-5 py-2.5 rounded-lg bg-black hover:bg-[#181818] border border-[#272727] text-white font-semibold text-sm transition-colors duration-150 ease-out active:scale-[0.98]"
            >
              <span>List Cluster Inventory</span>
            </Link>
          </div>
        </div>
      </section>
    </div>
  );
}
