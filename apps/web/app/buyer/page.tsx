import Link from "next/link";
import { Plus, ArrowUpRight, Cpu, ShieldCheck, Activity } from "lucide-react";
import { StatusBadge } from "@/components/ui/status-badge";
import { formatCents } from "@/lib/utils";

// Mock initial data reflecting our database schemas
const mockActiveContracts = [
  {
    id: "trade_c8f921a4-9b2e-4b13-91dc-837264819011",
    grade_id: "H100-SXM-8XNV",
    grade_name: "8x NVIDIA H100 SXM 80GB",
    seller_name: "Nebula Compute Infrastructure LLC",
    state: "live" as const,
    window_start: "2026-09-22T00:00:00Z",
    window_end: "2026-09-29T00:00:00Z",
    total_cents: 3696000, // $36,960.00 ($22/hr * 168h * 8? or $220/hr for 8x node * 168h = $36,960)
    allreduce_gbps: 428.4,
    ecc_errors: 0,
  },
  {
    id: "trade_d4e110b2-7c3a-4a21-88fc-129481726354",
    grade_id: "H100-SXM-8XNV",
    grade_name: "8x NVIDIA H100 SXM 80GB",
    seller_name: "HyperScale Cloud Systems",
    state: "delivery_test" as const,
    window_start: "2026-09-26T00:00:00Z",
    window_end: "2026-10-03T00:00:00Z",
    total_cents: 3528000, // $35,280.00
    allreduce_gbps: 412.0,
    ecc_errors: 0,
  },
];

const mockRFQs = [
  {
    id: "rfq_09a12c44-55ff-4bc1-831e-998877665544",
    grade_id: "H100-SXM-8XNV",
    region_bucket: "US-East (Virginia)",
    window_start: "2026-10-01T00:00:00Z",
    window_end: "2026-10-08T00:00:00Z",
    status: "quoted" as const,
    quotes_count: 3,
    best_quote_cents: 3494400, // $34,944.00
    created_at: "2026-09-24T18:30:00Z",
  },
];

export default function BuyerDashboardPage() {
  return (
    <div className="mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8 space-y-8">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-border pb-6">
        <div>
          <div className="flex items-center gap-2">
            <h1 className="text-2xl sm:text-3xl font-bold text-white tracking-tight">Buyer Console</h1>
            <span className="px-2 py-0.5 rounded text-[11px] font-mono bg-blue-950 text-blue-300 border border-blue-800">
              Verified Entity
            </span>
          </div>
          <p className="text-xs text-muted mt-1">
            Manage your GPU capacity reservations, evaluate private RFQ quotes, and monitor live node telemetry.
          </p>
        </div>

        <Link
          href="/buyer/rfqs/new"
          className="inline-flex items-center justify-center gap-2 px-4 py-2.5 rounded-lg bg-primary hover:bg-primary-hover text-white text-xs sm:text-sm font-medium shadow-sm transition-all shrink-0"
        >
          <Plus className="h-4 w-4" />
          <span>New Forward RFQ</span>
        </Link>
      </div>

      {/* ACTIVE DELIVERED RESERVATIONS */}
      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <h2 className="text-base font-bold text-white flex items-center gap-2">
            <Activity className="h-4 w-4 text-accent" />
            <span>Active Reservations & Delivery Monitor</span>
          </h2>
          <span className="text-xs font-mono text-muted">{mockActiveContracts.length} Active Contracts</span>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {mockActiveContracts.map((c) => (
            <div
              key={c.id}
              className="p-5 rounded-xl bg-surface border border-border hover:border-borderHighlight transition-all space-y-4"
            >
              <div className="flex items-start justify-between">
                <div>
                  <h3 className="font-bold text-white text-sm">{c.grade_name}</h3>
                  <p className="text-xs text-muted mt-0.5 font-mono">{c.seller_name}</p>
                </div>
                <StatusBadge state={c.state} />
              </div>

              {/* Hardware Telemetry Snapshot */}
              <div className="grid grid-cols-3 gap-2 p-3 rounded-lg bg-surfaceSubtle border border-border/60 text-xs font-mono">
                <div>
                  <span className="text-muted block text-[10px]">ALL-REDUCE</span>
                  <span className="text-emerald-400 font-bold">{c.allreduce_gbps} GB/s</span>
                </div>
                <div>
                  <span className="text-muted block text-[10px]">ECC ERRORS</span>
                  <span className="text-white font-bold">{c.ecc_errors} (Clean)</span>
                </div>
                <div>
                  <span className="text-muted block text-[10px]">TOTAL VALUE</span>
                  <span className="text-white font-bold">{formatCents(c.total_cents)}</span>
                </div>
              </div>

              <div className="flex items-center justify-between text-xs pt-2 border-t border-border/60">
                <span className="text-muted font-mono text-[11px]">
                  Window: {new Date(c.window_start).toLocaleDateString()} – {new Date(c.window_end).toLocaleDateString()}
                </span>
                <Link
                  href={`/buyer/contracts/${c.id}`}
                  className="inline-flex items-center gap-1 text-primary hover:text-blue-400 font-medium text-xs transition-colors"
                >
                  <span>Telemetry & Details</span>
                  <ArrowUpRight className="h-3.5 w-3.5" />
                </Link>
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* OPEN / PENDING RFQS */}
      <div className="space-y-4 pt-4">
        <div className="flex items-center justify-between">
          <h2 className="text-base font-bold text-white flex items-center gap-2">
            <Cpu className="h-4 w-4 text-primary" />
            <span>Open Forward RFQs & Quote Pipeline</span>
          </h2>
          <span className="text-xs font-mono text-muted">{mockRFQs.length} Pending RFQ</span>
        </div>

        <div className="rounded-xl border border-border bg-surface overflow-hidden">
          <table className="w-full text-left text-xs sm:text-sm">
            <thead className="border-b border-border bg-surfaceSubtle text-muted uppercase font-mono text-[10px]">
              <tr>
                <th className="py-3 px-4">RFQ Identifier</th>
                <th className="py-3 px-4">Requested Grade</th>
                <th className="py-3 px-4">Region Bucket</th>
                <th className="py-3 px-4">Window Tenor</th>
                <th className="py-3 px-4">Quotes In</th>
                <th className="py-3 px-4">Best Offer</th>
                <th className="py-3 px-4 text-right">Action</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border/60">
              {mockRFQs.map((rfq) => (
                <tr key={rfq.id} className="hover:bg-surfaceSubtle/50 transition-colors">
                  <td className="py-3.5 px-4 font-mono text-xs text-zinc-300">
                    {rfq.id.slice(0, 16)}...
                  </td>
                  <td className="py-3.5 px-4 font-medium text-white">
                    {rfq.grade_id}
                  </td>
                  <td className="py-3.5 px-4 text-muted">
                    {rfq.region_bucket}
                  </td>
                  <td className="py-3.5 px-4 text-zinc-300 font-mono text-xs">
                    7 Days (168h)
                  </td>
                  <td className="py-3.5 px-4">
                    <span className="px-2 py-0.5 rounded-full text-[11px] font-mono bg-purple-950 text-purple-300 border border-purple-800">
                      {rfq.quotes_count} Received
                    </span>
                  </td>
                  <td className="py-3.5 px-4 font-mono font-bold text-emerald-400">
                    {formatCents(rfq.best_quote_cents)}
                  </td>
                  <td className="py-3.5 px-4 text-right">
                    <Link
                      href={`/buyer/rfqs/${rfq.id}`}
                      className="inline-flex items-center gap-1 px-3 py-1.5 rounded-lg bg-surface border border-border hover:border-borderHighlight text-white text-xs font-medium transition-all"
                    >
                      <span>Compare Quotes</span>
                      <ArrowUpRight className="h-3.5 w-3.5" />
                    </Link>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}
