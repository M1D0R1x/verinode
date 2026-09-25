"use client";

import Link from "next/link";
import { Plus, Server, Activity, ShieldCheck, ArrowUpRight, Terminal } from "lucide-react";
import { formatCents } from "@/lib/utils";

const mockInventory = [
  {
    id: "block_01h8a912-3344",
    grade_id: "H100-SXM-8XNV",
    region_bucket: "US-East (Virginia)",
    facility_ref: "EQUINIX-DC21-RACK04",
    window_start: "2026-10-01T00:00:00Z",
    window_end: "2026-10-08T00:00:00Z",
    status: "available" as const,
    agent_status: "active" as const,
    last_heartbeat: "2m ago",
  },
  {
    id: "block_02h8a912-5566",
    grade_id: "H100-SXM-8XNV",
    region_bucket: "US-East (Virginia)",
    facility_ref: "EQUINIX-DC21-RACK05",
    window_start: "2026-09-22T00:00:00Z",
    window_end: "2026-09-29T00:00:00Z",
    status: "delivered" as const,
    agent_status: "active" as const,
    last_heartbeat: "5s ago",
  },
];

const mockIncomingRFQs = [
  {
    id: "rfq_09a12c44-55ff-4bc1-831e-998877665544",
    buyer_entity: "Anthropic Partner Lab Inc.",
    grade_id: "H100-SXM-8XNV",
    region_bucket: "US-East",
    window: "Oct 05 – Oct 12 (168h)",
    target_hourly: 21000,
    created_at: "3h ago",
  },
];

export default function SellerDashboardPage() {
  return (
    <div className="mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8 space-y-8">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-border pb-6">
        <div>
          <div className="flex items-center gap-2">
            <h1 className="text-2xl sm:text-3xl font-bold text-white tracking-tight">Supplier Console</h1>
            <span className="px-2 py-0.5 rounded text-[11px] font-mono bg-emerald-950 text-emerald-300 border border-emerald-800">
              Vetted Tier-1 Provider
            </span>
          </div>
          <p className="text-xs text-muted mt-1">
            List GPU cluster availability, respond to private RFQs, and monitor host telemetry agent attestations.
          </p>
        </div>

        <div className="flex items-center gap-3">
          <Link
            href="/seller/agent"
            className="inline-flex items-center gap-1.5 px-3.5 py-2 rounded-lg bg-surfaceSubtle border border-border hover:border-borderHighlight text-white text-xs font-medium transition-colors"
          >
            <Terminal className="h-3.5 w-3.5 text-muted" />
            <span>Agent Setup</span>
          </Link>

          <Link
            href="/seller/inventory/new"
            className="inline-flex items-center gap-1.5 px-4 py-2 rounded-lg bg-primary hover:bg-primary-hover text-white text-xs font-medium transition-all shadow-sm"
          >
            <Plus className="h-4 w-4" />
            <span>List Capacity Block</span>
          </Link>
        </div>
      </div>

      {/* INCOMING RFQS AWAITING QUOTE */}
      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <h2 className="text-base font-bold text-white flex items-center gap-2">
            <Server className="h-4 w-4 text-primary" />
            <span>Incoming Private RFQs Awaiting Your Quote</span>
          </h2>
          <span className="text-xs font-mono text-muted">{mockIncomingRFQs.length} Action Needed</span>
        </div>

        <div className="rounded-xl border border-border bg-surface overflow-hidden">
          <table className="w-full text-left text-xs sm:text-sm">
            <thead className="border-b border-border bg-surfaceSubtle text-muted uppercase font-mono text-[10px]">
              <tr>
                <th className="py-3 px-4">Buyer Entity</th>
                <th className="py-3 px-4">Requested Grade</th>
                <th className="py-3 px-4">Delivery Window</th>
                <th className="py-3 px-4">Target Rate</th>
                <th className="py-3 px-4">Received</th>
                <th className="py-3 px-4 text-right">Action</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border/60">
              {mockIncomingRFQs.map((rfq) => (
                <tr key={rfq.id} className="hover:bg-surfaceSubtle/50 transition-colors">
                  <td className="py-3.5 px-4 font-semibold text-white">
                    {rfq.buyer_entity}
                  </td>
                  <td className="py-3.5 px-4 font-mono text-xs text-zinc-300">
                    {rfq.grade_id}
                  </td>
                  <td className="py-3.5 px-4 text-muted font-mono text-xs">
                    {rfq.window}
                  </td>
                  <td className="py-3.5 px-4 font-mono text-zinc-300 font-bold">
                    ~${(rfq.target_hourly / 100).toFixed(2)}/hr
                  </td>
                  <td className="py-3.5 px-4 text-muted text-xs">
                    {rfq.created_at}
                  </td>
                  <td className="py-3.5 px-4 text-right">
                    <button
                      onClick={() => alert(`Submit Quote Modal for RFQ: ${rfq.id}`)}
                      className="inline-flex items-center gap-1 px-3 py-1.5 rounded-lg bg-primary hover:bg-primary-hover text-white text-xs font-medium transition-all"
                    >
                      <span>Submit Quote</span>
                      <ArrowUpRight className="h-3.5 w-3.5" />
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      {/* INVENTORY BLOCKS */}
      <div className="space-y-4 pt-4">
        <div className="flex items-center justify-between">
          <h2 className="text-base font-bold text-white flex items-center gap-2">
            <Activity className="h-4 w-4 text-accent" />
            <span>Listed GPU Blocks & Telemetry Agent Status</span>
          </h2>
          <span className="text-xs font-mono text-muted">{mockInventory.length} Registered Blocks</span>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {mockInventory.map((item) => (
            <div
              key={item.id}
              className="p-5 rounded-xl bg-surface border border-border space-y-4"
            >
              <div className="flex items-start justify-between">
                <div>
                  <h3 className="font-bold text-white text-sm">8x NVIDIA H100 SXM 80GB</h3>
                  <p className="text-xs font-mono text-muted mt-0.5">{item.id}</p>
                </div>
                <span
                  className={`px-2.5 py-1 rounded-full text-xs font-mono font-medium border ${
                    item.status === "available"
                      ? "bg-blue-950 text-blue-300 border-blue-800"
                      : "bg-emerald-950 text-emerald-300 border-emerald-800"
                  }`}
                >
                  {item.status.toUpperCase()}
                </span>
              </div>

              <div className="grid grid-cols-2 gap-2 p-3 rounded-lg bg-surfaceSubtle border border-border/60 text-xs font-mono">
                <div>
                  <span className="text-muted block text-[10px]">REGION & FACILITY</span>
                  <span className="text-zinc-300">{item.region_bucket}</span>
                  <span className="text-[10px] text-zinc-500 block truncate">{item.facility_ref}</span>
                </div>
                <div>
                  <span className="text-muted block text-[10px]">HOST AGENT HEARTBEAT</span>
                  <span className="text-emerald-400 font-bold flex items-center gap-1">
                    <span className="h-2 w-2 rounded-full bg-emerald-400 animate-pulse" />
                    <span>Active ({item.last_heartbeat})</span>
                  </span>
                  <span className="text-[10px] text-zinc-500 block">ed25519 Signed</span>
                </div>
              </div>

              <div className="flex items-center justify-between text-xs pt-1 border-t border-border/60 text-muted font-mono">
                <span>
                  Window: {new Date(item.window_start).toLocaleDateString()} – {new Date(item.window_end).toLocaleDateString()}
                </span>
                <Link
                  href="/seller/agent"
                  className="text-primary hover:text-blue-400 text-xs font-medium transition-colors"
                >
                  Agent Logs →
                </Link>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
