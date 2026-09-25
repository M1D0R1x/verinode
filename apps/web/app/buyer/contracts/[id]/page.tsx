"use client";

import { useState, useEffect } from "react";
import Link from "next/link";
import {
  ArrowLeft,
  Server,
  ShieldCheck,
  CheckCircle2,
  FileCheck,
  Download,
  AlertTriangle,
  Cpu,
  Activity,
  Lock,
} from "lucide-react";
import { StatusBadge } from "@/components/ui/status-badge";
import { StateTimeline } from "@/components/ui/state-timeline";
import { ContractState } from "@/lib/types";
import { formatCents } from "@/lib/utils";
import { api } from "@/lib/api";

export default function ContractDetailPage({ params }: { params: { id: string } }) {
  const [contractState, setContractState] = useState<ContractState>("live");
  const [isDisputing, setIsDisputing] = useState(false);
  const [transitionNotice, setTransitionNotice] = useState<string | null>(null);

  const tradeId = params.id;
  const gradeName = "8x NVIDIA H100 SXM 80GB (168-Hour Block)";
  const sellerName = "Nebula Compute Infrastructure LLC";
  const totalCents = 3696000; // $36,960.00
  const pdfHash = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855";

  useEffect(() => {
    api
      .getContract(tradeId)
      .then((c) => {
        if (c && c.state) {
          setContractState(c.state);
        }
      })
      .catch(() => {
        // Fallback to initial state if offline
      });
  }, [tradeId]);

  const handleOpenClaim = async () => {
    if (confirm("Open an SLA performance claim with cryptographic telemetry evidence bundle?")) {
      try {
        const res = await api.validateTransition({
          current_state: contractState,
          next_state: "claim_open",
          actor: "usr_buyer_institutional",
          reason: "NCCL bandwidth dip below 400 GB/s benchmark floor",
          idempotency_key: `claim_${Date.now()}`,
        });
        if (res.valid) {
          setContractState("claim_open");
          setIsDisputing(true);
          setTransitionNotice("Transition verified and logged by Go State Machine (RFC 7807 compliant).");
        }
      } catch (err: unknown) {
        // Fallback update if gateway offline
        setContractState("claim_open");
        setIsDisputing(true);
        setTransitionNotice("Claim opened locally (API gateway offline).");
      }
    }
  };

  return (
    <div className="mx-auto max-w-6xl px-4 py-8 sm:px-6 lg:px-8 space-y-6">
      <Link
        href="/buyer"
        className="inline-flex items-center gap-1.5 text-xs text-muted hover:text-white transition-colors"
      >
        <ArrowLeft className="h-3.5 w-3.5" />
        <span>Back to Buyer Console</span>
      </Link>

      {/* Contract Header */}
      <div className="p-6 rounded-2xl bg-surface border border-border space-y-4">
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 border-b border-border/60 pb-4">
          <div>
            <div className="flex items-center gap-3">
              <h1 className="text-xl font-bold text-white tracking-tight">{gradeName}</h1>
              <StatusBadge state={contractState} />
            </div>
            <p className="text-xs font-mono text-muted mt-1">
              Canonical Trade Envelope ID (<span className="text-zinc-300">trade_id</span>): {tradeId}
            </p>
          </div>

          <div className="flex items-center gap-3">
            <button
              onClick={() => alert(`Signed PDF SHA-256 Hash: ${pdfHash}\nRetrieved from object storage.`)}
              className="inline-flex items-center gap-1.5 px-3.5 py-2 rounded-lg bg-surfaceSubtle border border-border hover:border-borderHighlight text-white text-xs font-medium transition-colors"
            >
              <Download className="h-3.5 w-3.5 text-muted" />
              <span>Signed Master Confirmation (PDF)</span>
            </button>

            {contractState === "live" && (
              <button
                onClick={handleOpenClaim}
                className="inline-flex items-center gap-1.5 px-3.5 py-2 rounded-lg bg-rose-950/50 border border-rose-800 text-rose-300 hover:bg-rose-900/50 text-xs font-medium transition-colors"
              >
                <AlertTriangle className="h-3.5 w-3.5" />
                <span>Open SLA Claim</span>
              </button>
            )}
          </div>
        </div>

        {/* State Machine Lifecycle Progression */}
        <div className="pt-2">
          <span className="text-xs font-mono text-muted uppercase tracking-wider block mb-2">
            Contract Lifecycle State Progression
          </span>
          <StateTimeline currentState={contractState} />
        </div>

        {transitionNotice && (
          <div className="p-3 rounded-lg bg-emerald-950/40 border border-emerald-800 text-emerald-300 text-xs font-mono flex items-center gap-2">
            <CheckCircle2 className="h-4 w-4 shrink-0 text-emerald-400" />
            <span>{transitionNotice}</span>
          </div>
        )}
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* TELEMETRY & HARDWARE HEALTH */}
        <div className="lg:col-span-2 space-y-6">
          <div className="p-6 rounded-2xl bg-surface border border-border space-y-5">
            <div className="flex items-center justify-between border-b border-border/60 pb-3">
              <div className="flex items-center gap-2">
                <Activity className="h-4 w-4 text-accent" />
                <h2 className="text-sm font-bold text-white uppercase tracking-wider font-mono">
                  Cryptographic Hardware Telemetry & Canary Status
                </h2>
              </div>
              <span className="text-[11px] font-mono text-emerald-400 flex items-center gap-1">
                <span className="h-2 w-2 rounded-full bg-emerald-400 animate-pulse" />
                <span>Verified Active</span>
              </span>
            </div>

            {/* Metrics Grid */}
            <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 text-xs font-mono">
              <div className="p-3 rounded-xl bg-surfaceSubtle border border-border/60">
                <span className="text-muted block text-[10px]">ALL-REDUCE BANDWIDTH</span>
                <span className="text-lg font-bold text-emerald-400 mt-0.5 block">428.4 GB/s</span>
                <span className="text-[10px] text-zinc-500">Floor: 400.0 GB/s</span>
              </div>

              <div className="p-3 rounded-xl bg-surfaceSubtle border border-border/60">
                <span className="text-muted block text-[10px]">HEALTHY GPUS</span>
                <span className="text-lg font-bold text-white mt-0.5 block">8 / 8 Online</span>
                <span className="text-[10px] text-zinc-500">SXM5 Topology</span>
              </div>

              <div className="p-3 rounded-xl bg-surfaceSubtle border border-border/60">
                <span className="text-muted block text-[10px]">NVLINK TOPOLOGY</span>
                <span className="text-lg font-bold text-white mt-0.5 block">NVSwitch 4.0</span>
                <span className="text-[10px] text-zinc-500">Full 900 GB/s Mesh</span>
              </div>

              <div className="p-3 rounded-xl bg-surfaceSubtle border border-border/60">
                <span className="text-muted block text-[10px]">ECC MEMORY ERRORS</span>
                <span className="text-lg font-bold text-emerald-400 mt-0.5 block">0 Detected</span>
                <span className="text-[10px] text-zinc-500">100% Clean Pass</span>
              </div>
            </div>

            {/* Signed Hardware Attestation Proof Box */}
            <div className="p-4 rounded-xl bg-surfaceSubtle border border-border space-y-2 text-xs font-mono">
              <div className="flex items-center justify-between text-muted text-[11px]">
                <span className="flex items-center gap-1.5 text-accent font-semibold">
                  <ShieldCheck className="h-3.5 w-3.5" />
                  <span>Supplier ed25519 Host Agent Attestation</span>
                </span>
                <span>Seq #8,421</span>
              </div>
              <div className="p-2.5 rounded bg-background border border-border/60 text-[11px] text-zinc-400 break-all space-y-1">
                <div><span className="text-muted">Agent Public Key:</span> MCowBQYDK2VwAyEAX5qYF8...k9L0vQw</div>
                <div><span className="text-muted">Report Digest (SHA-256):</span> a7d9e410c841bb82...88319fbc</div>
                <div><span className="text-muted">Telemetry Heartbeat:</span> 3 seconds ago (ClickHouse time-series)</div>
              </div>
            </div>

            <div className="flex items-center gap-2 p-3 rounded-lg bg-blue-950/30 border border-blue-900/40 text-blue-300 text-xs">
              <Lock className="h-4 w-4 shrink-0" />
              <span>
                <strong>Invariant 4 Active:</strong> Zero customer workload data (model weights, code, prompts) is ingested. Only DCGM & NCCL synthetic telemetry is recorded.
              </span>
            </div>
          </div>
        </div>

        {/* FINANCIAL / SETTLEMENT DETAILS */}
        <div className="space-y-6">
          <div className="p-6 rounded-2xl bg-surface border border-border space-y-4">
            <h2 className="text-sm font-bold text-white uppercase tracking-wider font-mono">
              Financial Escrow & Ledger
            </h2>

            <div className="space-y-3 text-xs font-mono">
              <div className="flex justify-between py-1.5 border-b border-border/60">
                <span className="text-muted">Supplier Entity:</span>
                <span className="text-white font-medium">{sellerName}</span>
              </div>
              <div className="flex justify-between py-1.5 border-b border-border/60">
                <span className="text-muted">Tenor:</span>
                <span className="text-white font-medium">168h Continuous</span>
              </div>
              <div className="flex justify-between py-1.5 border-b border-border/60">
                <span className="text-muted">Total Escrow Value:</span>
                <span className="text-white font-bold">{formatCents(totalCents)}</span>
              </div>
              <div className="flex justify-between py-1.5 border-b border-border/60">
                <span className="text-muted">Escrow Status:</span>
                <span className="text-emerald-400 font-bold">Funds Secured (Bank Fedwire)</span>
              </div>
            </div>

            <div className="p-3.5 rounded-xl bg-surfaceSubtle border border-border text-[11px] text-muted space-y-1">
              <span className="font-semibold text-zinc-300 block">Double-Entry Accounting Invariant:</span>
              <p>
                Escrow debits and seller payout credits balance to zero (Sum of amount_cents == 0). Payout executes automatically upon window completion.
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
