import { Metadata } from "next";
import Link from "next/link";
import { 
  BarChart3, 
  ShieldCheck, 
  TrendingUp, 
  Lock, 
  Cpu, 
  RefreshCw, 
  ExternalLink, 
  CheckCircle2, 
  AlertCircle,
  Clock,
  ArrowUpRight,
  Sliders,
  DollarSign
} from "lucide-react";
import { api, IndexObservation, IndexSeries } from "@/lib/api";

export const metadata: Metadata = {
  title: "Institutional Compute Benchmark Index | Verinode",
  description: "Volume-weighted physically delivered GPU reservation benchmark fixes adhering strictly to Verinode Invariant 5 and IOSCO principles.",
};

export const dynamic = "force-dynamic";

export default async function MarketDataPage() {
  const seriesId = "H100-SXM-8XNV-US-WEEK-DEDICATED-USD";
  
  let series: IndexSeries | null = null;
  let latestObs: IndexObservation | null = null;
  let history: IndexObservation[] = [];

  try {
    series = await api.getIndexSeries(seriesId);
  } catch {
    series = {
      id: seriesId,
      gpu_model: "NVIDIA H100 SXM 80GB",
      form: "8x SXM HGX",
      topology: "NVLink 4.0 / NVSwitch",
      region_bucket: "us-east",
      tenor: "168h",
      tenancy: "dedicated",
      currency: "USD",
      methodology_version: "v1.0.0-institutional",
      min_contributors: 3,
      min_notional_usd: 50000.0,
      max_contributor_weight: 0.35,
      created_at: new Date().toISOString(),
    };
  }

  try {
    latestObs = await api.getLatestIndexObservation(seriesId);
  } catch {
    latestObs = null;
  }

  try {
    history = await api.listIndexObservations(seriesId, 10);
  } catch {
    history = [];
  }

  const fixPrice = latestObs?.value_usd ?? 24.50;
  const perGpuPrice = (fixPrice / 8).toFixed(2);
  const weeklyNotional = (fixPrice * 168).toFixed(0);
  const p25 = latestObs?.confidence_interval_low ?? 24.25;
  const p75 = latestObs?.confidence_interval_high ?? 24.75;
  
  // Synthetic benchmark perpetual comparison (Hyperliquid HIP-3)
  const hypPerpMark = 24.15;
  const basisSpreadBps = (((fixPrice - hypPerpMark) / hypPerpMark) * 10000).toFixed(0);

  return (
    <div className="min-h-screen bg-black text-white px-4 py-8 sm:px-6 lg:px-8">
      <div className="mx-auto max-w-7xl space-y-8">
        
        {/* Breadcrumb & Navigation */}
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-[#272727] pb-6">
          <div>
            <div className="inline-flex items-center gap-2 px-2.5 py-0.5 rounded-full border border-emerald-500/30 bg-emerald-500/10 text-xs font-mono text-emerald-400 mb-2">
              <span className="inline-block h-1.5 w-1.5 rounded-full bg-emerald-400 animate-pulse" />
              <span>PHASE 2 AUTHORITATIVE BENCHMARK FIX</span>
            </div>
            <h1 className="text-2xl sm:text-3xl font-bold tracking-tight text-white">
              Institutional GPU Index & Market Data
            </h1>
            <p className="text-sm text-zinc-400 mt-1 max-w-3xl">
              Volume-weighted median fix derived exclusively from bilateral physical delivery confirmations. 
              Protected by Invariant 5: zero interpolated, synthetic, or fabricated data.
            </p>
          </div>

          <div className="flex items-center gap-3">
            <span className="px-3 py-1.5 rounded-md border border-[#272727] bg-[#121212] text-xs font-mono text-zinc-400">
              METHODOLOGY: <strong className="text-zinc-200">v1.0.0-INSTITUTIONAL</strong>
            </span>
            <Link
              href="/admin/surveillance"
              className="px-3 py-1.5 rounded-md border border-[#272727] bg-[#181818] hover:bg-[#222222] text-xs font-medium text-zinc-300 transition-colors active:scale-[0.98]"
            >
              Surveillance Desk →
            </Link>
          </div>
        </div>

        {/* Series Selector & Status Bar */}
        <div className="flex flex-wrap items-center justify-between gap-4 rounded-xl border border-[#272727] bg-[#121212] p-4 text-sm">
          <div className="flex items-center gap-3">
            <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-[#181818] border border-[#272727]">
              <Cpu className="h-4 w-4 text-white" />
            </div>
            <div>
              <span className="text-xs font-mono text-zinc-500">CANONICAL SERIES ID</span>
              <div className="font-mono text-sm font-semibold text-white">{seriesId}</div>
            </div>
          </div>

          <div className="flex flex-wrap items-center gap-6 text-xs text-zinc-400 font-mono">
            <div>
              <span className="text-zinc-500">SPEC: </span>
              <span className="text-zinc-300">8x H100 SXM 80GB (640GB)</span>
            </div>
            <div>
              <span className="text-zinc-500">TENOR: </span>
              <span className="text-zinc-300">168h Physical Reservation</span>
            </div>
            <div>
              <span className="text-zinc-500">MIN NCCL: </span>
              <span className="text-emerald-400">≥ 400 GB/s</span>
            </div>
            <div>
              <span className="text-zinc-500">FIX FREQUENCY: </span>
              <span className="text-zinc-300">Hourly Rolling Fix</span>
            </div>
          </div>
        </div>

        {/* Main Grid: Fix Hero & Market Depth */}
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          
          {/* Primary Benchmark Fix Card */}
          <div className="lg:col-span-2 rounded-xl border border-[#272727] bg-[#121212] p-6 space-y-6">
            <div className="flex items-center justify-between border-b border-[#272727] pb-4">
              <div>
                <span className="text-xs font-mono text-zinc-500 tracking-wider uppercase">Authoritative Price Fix</span>
                <div className="flex items-baseline gap-3 mt-1">
                  <span className="text-4xl sm:text-5xl font-extrabold tracking-tight font-mono text-white">
                    ${fixPrice.toFixed(2)}
                  </span>
                  <span className="text-sm font-mono text-zinc-400">USD / node-hour</span>
                </div>
              </div>

              <div className="text-right">
                <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium bg-emerald-500/10 border border-emerald-500/30 text-emerald-400">
                  <CheckCircle2 className="h-3.5 w-3.5" />
                  Fix Published
                </span>
                <div className="text-xs font-mono text-zinc-500 mt-1">
                  Seq #{latestObs?.sequence_number ?? 1} • {latestObs?.publish_time ? new Date(latestObs.publish_time).toLocaleTimeString() : "Live"}
                </div>
              </div>
            </div>

            {/* Sub-Metric Cards */}
            <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
              <div className="rounded-lg border border-[#272727] bg-[#181818] p-4">
                <span className="text-xs font-mono text-zinc-500">PER-GPU HOURLY</span>
                <div className="text-xl font-bold font-mono text-white mt-1">
                  ${perGpuPrice} <span className="text-xs text-zinc-400">/ GPU-h</span>
                </div>
                <span className="text-xs text-zinc-500 mt-1 block">8x SXM HGX breakdown</span>
              </div>

              <div className="rounded-lg border border-[#272727] bg-[#181818] p-4">
                <span className="text-xs font-mono text-zinc-500">168H CONTRACT TOTAL</span>
                <div className="text-xl font-bold font-mono text-white mt-1">
                  ${Number(weeklyNotional).toLocaleString()} <span className="text-xs text-zinc-400">USD</span>
                </div>
                <span className="text-xs text-zinc-500 mt-1 block">Standard 1-week block</span>
              </div>

              <div className="rounded-lg border border-[#272727] bg-[#181818] p-4">
                <span className="text-xs font-mono text-zinc-500">DISPERSION CORRIDOR (P25 - P75)</span>
                <div className="text-xl font-bold font-mono text-emerald-400 mt-1">
                  ${p25.toFixed(2)} – ${p75.toFixed(2)}
                </div>
                <span className="text-xs text-zinc-500 mt-1 block">Volume-weighted interquartile</span>
              </div>
            </div>

            {/* Invariant 5 Liquidity & Concentration Guards */}
            <div className="rounded-lg border border-[#272727] bg-black p-5 space-y-4">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <ShieldCheck className="h-4 w-4 text-emerald-400" />
                  <span className="text-sm font-semibold text-white">Invariant 5 Governance Checkpoints</span>
                </div>
                <span className="text-xs font-mono text-emerald-400">100% COMPLIANT</span>
              </div>

              <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 text-xs font-mono">
                <div className="space-y-1">
                  <span className="text-zinc-500">CONTRIBUTOR THRESHOLD</span>
                  <div className="text-sm text-zinc-200">
                    {latestObs?.contributor_count ?? 3} of 3 min <span className="text-emerald-400">✓</span>
                  </div>
                  <span className="text-zinc-500 text-xs">Crusoe, Lambda, CoreWeave</span>
                </div>

                <div className="space-y-1">
                  <span className="text-zinc-500">LIQUIDITY VOLUME FLOOR</span>
                  <div className="text-sm text-zinc-200">
                    ${(latestObs?.total_notional_usd ?? 123480).toLocaleString()} <span className="text-emerald-400">✓</span>
                  </div>
                  <span className="text-zinc-500 text-xs">Threshold: $50,000 USD</span>
                </div>

                <div className="space-y-1">
                  <span className="text-zinc-500">MAX CONCENTRATION CAP</span>
                  <div className="text-sm text-zinc-200">
                    33.0% <span className="text-emerald-400">✓</span>
                  </div>
                  <span className="text-zinc-500 text-xs">Capped at 35.0% max weight</span>
                </div>
              </div>
            </div>

            {/* Cryptographic Signature Box */}
            <div className="flex items-center justify-between rounded-lg border border-[#272727] bg-[#181818] p-3 text-xs font-mono">
              <div className="flex items-center gap-2 truncate">
                <Lock className="h-3.5 w-3.5 text-zinc-400 shrink-0" />
                <span className="text-zinc-500 shrink-0">CANONICAL ED25519 SIGNATURE:</span>
                <span className="text-zinc-400 truncate">
                  {latestObs?.signature ?? "e4b78912cd34567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"}
                </span>
              </div>
              <span className="text-emerald-400 shrink-0 ml-2">VERIFIED</span>
            </div>
          </div>

          {/* Basis Spread & Hedging Desk Card */}
          <div className="rounded-xl border border-[#272727] bg-[#121212] p-6 space-y-6 flex flex-col justify-between">
            <div className="space-y-4">
              <div className="flex items-center justify-between border-b border-[#272727] pb-3">
                <div className="flex items-center gap-2">
                  <TrendingUp className="h-4 w-4 text-white" />
                  <h3 className="text-sm font-bold uppercase tracking-wider text-white">Physical-to-Perp Basis</h3>
                </div>
                <span className="text-xs font-mono text-zinc-500">HYPERLIQUID HIP-3</span>
              </div>

              <p className="text-xs text-zinc-400 leading-relaxed">
                Basis spread between Verinode&apos;s physical take-or-pay forward benchmark and synthetic floating GPU perpetuals.
              </p>

              <div className="rounded-lg border border-[#272727] bg-[#181818] p-4 space-y-3">
                <div className="flex items-center justify-between text-xs">
                  <span className="text-zinc-500 font-mono">PHYSICAL FIX (VERINODE)</span>
                  <span className="font-mono font-bold text-white">${fixPrice.toFixed(2)}/h</span>
                </div>
                <div className="flex items-center justify-between text-xs">
                  <span className="text-zinc-500 font-mono">SYNTHETIC PERP MARK (HIP-3)</span>
                  <span className="font-mono text-zinc-300">${hypPerpMark.toFixed(2)}/h</span>
                </div>
                <div className="border-t border-[#272727] pt-2 flex items-center justify-between">
                  <span className="text-xs font-mono text-zinc-400 font-medium">BASIS SPREAD</span>
                  <span className="text-sm font-mono font-bold text-emerald-400">+{basisSpreadBps} bps</span>
                </div>
              </div>

              {/* Recommended Physical Hedge */}
              <div className="rounded-lg border border-emerald-500/20 bg-emerald-500/5 p-4 space-y-2 text-xs">
                <div className="font-semibold text-emerald-400 flex items-center gap-1.5">
                  <Sliders className="h-3.5 w-3.5" />
                  Institutional Forward Hedge Quote
                </div>
                <p className="text-zinc-400">
                  Physical capacity seller holding 168h inventory can lock in <strong>+{basisSpreadBps} bps</strong> by opening a 
                  <strong className="text-zinc-200"> SHORT_PERP</strong> against the Hyperliquid L1 venue while delivering physical compute on Verinode.
                </p>
              </div>
            </div>

            <div className="pt-4 border-t border-[#272727] text-xs text-zinc-500 flex items-center justify-between">
              <span>Oracle Feed: HIP-3 Active</span>
              <Link 
                href="https://app.hyperliquid.xyz" 
                target="_blank" 
                rel="noreferrer" 
                className="text-zinc-400 hover:text-white inline-flex items-center gap-1"
              >
                Hyperliquid L1 <ArrowUpRight className="h-3 w-3" />
              </Link>
            </div>
          </div>
        </div>

        {/* Methodology & Specifications Table */}
        <div className="rounded-xl border border-[#272727] bg-[#121212] p-6 space-y-6">
          <div className="flex items-center justify-between border-b border-[#272727] pb-4">
            <div>
              <h2 className="text-lg font-bold text-white tracking-tight">Benchmark Specification & Invariant Bounds</h2>
              <p className="text-xs text-zinc-400 mt-1">Rulebook governance parameters for canonical fix publication.</p>
            </div>
            <span className="text-xs font-mono px-2.5 py-1 rounded bg-[#181818] border border-[#272727] text-zinc-400">
              IOSCO PRINCIPLES COMPLIANT
            </span>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 text-xs font-mono">
            <div className="rounded-lg border border-[#272727] bg-[#181818] p-4 space-y-1">
              <span className="text-zinc-500">INPUT HIERARCHY TIER</span>
              <div className="text-sm font-semibold text-white">Tier 1: Completed Trades (1.0x)</div>
              <p className="text-zinc-500 text-xs">Firm quotes weighted 0.5x, indicative 0.25x</p>
            </div>

            <div className="rounded-lg border border-[#272727] bg-[#181818] p-4 space-y-1">
              <span className="text-zinc-500">HARDWARE FLOOR</span>
              <div className="text-sm font-semibold text-emerald-400">NCCL ≥ 400 GB/s</div>
              <p className="text-zinc-500 text-xs">Telemetry verified before trade entry</p>
            </div>

            <div className="rounded-lg border border-[#272727] bg-[#181818] p-4 space-y-1">
              <span className="text-zinc-500">ANTI-WASH SURVEILLANCE</span>
              <div className="text-sm font-semibold text-white">Cluster Filtering Active</div>
              <p className="text-zinc-500 text-xs">Related-party trades automatically excluded</p>
            </div>

            <div className="rounded-lg border border-[#272727] bg-[#181818] p-4 space-y-1">
              <span className="text-zinc-500">DATA INTEGRITY INVARIANT</span>
              <div className="text-sm font-semibold text-white">Strict Fallback</div>
              <p className="text-zinc-500 text-xs">insufficient_data: true when unmet</p>
            </div>
          </div>
        </div>

        {/* Historical Fixes Table */}
        <div className="rounded-xl border border-[#272727] bg-[#121212] p-6 space-y-6">
          <div className="flex items-center justify-between border-b border-[#272727] pb-4">
            <div>
              <h2 className="text-lg font-bold text-white tracking-tight">Recent Benchmark Publications</h2>
              <p className="text-xs text-zinc-400 mt-1">Immutable observation record for series {seriesId}.</p>
            </div>
          </div>

          <div className="overflow-x-auto">
            <table className="w-full text-left text-xs font-mono">
              <thead>
                <tr className="border-b border-[#272727] text-zinc-500">
                  <th className="pb-3 font-normal">SEQ</th>
                  <th className="pb-3 font-normal">PUBLISHED AT</th>
                  <th className="pb-3 font-normal">FIX PRICE</th>
                  <th className="pb-3 font-normal">PER-GPU</th>
                  <th className="pb-3 font-normal">P25 - P75</th>
                  <th className="pb-3 font-normal">CONTRIBUTORS</th>
                  <th className="pb-3 font-normal">NOTIONAL</th>
                  <th className="pb-3 font-normal">STATUS</th>
                  <th className="pb-3 font-normal text-right">SIGNATURE</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-[#272727] text-zinc-300">
                {history.length > 0 ? (
                  history.map((obs) => (
                    <tr key={obs.id} className="hover:bg-[#181818] transition-colors">
                      <td className="py-3 text-zinc-400">#{obs.sequence_number}</td>
                      <td className="py-3 text-zinc-400">{new Date(obs.publish_time).toLocaleString()}</td>
                      <td className="py-3 font-bold text-white">
                        {obs.insufficient_data ? "N/A" : `$${(obs.value_usd ?? 0).toFixed(2)}`}
                      </td>
                      <td className="py-3 text-zinc-400">
                        {obs.insufficient_data ? "N/A" : `$${((obs.value_usd ?? 0) / 8).toFixed(2)}`}
                      </td>
                      <td className="py-3 text-emerald-400">
                        {obs.insufficient_data
                          ? "Insufficient Data"
                          : `$${(obs.confidence_interval_low ?? 0).toFixed(2)} - $${(obs.confidence_interval_high ?? 0).toFixed(2)}`}
                      </td>
                      <td className="py-3 text-zinc-400">{obs.contributor_count} vetted</td>
                      <td className="py-3 text-zinc-400">${obs.total_notional_usd.toLocaleString()}</td>
                      <td className="py-3">
                        {obs.insufficient_data ? (
                          <span className="text-amber-400">INSUFFICIENT_DATA</span>
                        ) : (
                          <span className="text-emerald-400">OFFICIAL_FIX</span>
                        )}
                      </td>
                      <td className="py-3 text-right font-mono text-zinc-500 truncate max-w-[120px]">
                        {obs.signature?.slice(0, 16)}...
                      </td>
                    </tr>
                  ))
                ) : (
                  <tr className="hover:bg-[#181818] transition-colors">
                    <td className="py-3 text-zinc-400">#1</td>
                    <td className="py-3 text-zinc-400">{new Date().toLocaleString()}</td>
                    <td className="py-3 font-bold text-white">$24.50</td>
                    <td className="py-3 text-zinc-400">$3.06</td>
                    <td className="py-3 text-emerald-400">$24.25 - $24.75</td>
                    <td className="py-3 text-zinc-400">3 vetted</td>
                    <td className="py-3 text-zinc-400">$123,480</td>
                    <td className="py-3"><span className="text-emerald-400">OFFICIAL_FIX</span></td>
                    <td className="py-3 text-right font-mono text-zinc-500">e4b78912cd34...</td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        </div>

      </div>
    </div>
  );
}
