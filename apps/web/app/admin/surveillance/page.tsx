"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { 
  ArrowLeft, 
  ShieldAlert, 
  CheckCircle2, 
  AlertTriangle, 
  RefreshCw, 
  Sliders, 
  Eye, 
  XCircle,
  FileCheck,
  Scale
} from "lucide-react";
import { api, SurveillanceFlag } from "@/lib/api";
import { Badge } from "@/components/ui/badge";

export default function SurveillanceAdminPage() {
  const [flags, setFlags] = useState<SurveillanceFlag[]>([]);
  const [loading, setLoading] = useState(true);
  const [statusFilter, setStatusFilter] = useState<string>("all");
  const [updatingId, setUpdatingId] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  const fetchFlags = async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await api.listSurveillanceFlags(statusFilter === "all" ? undefined : statusFilter);
      setFlags(data);
    } catch (err: unknown) {
      if (err instanceof Error) {
        setError(err.message);
      } else {
        setError("Failed to query surveillance flags");
      }
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchFlags();
  }, [statusFilter]);

  const handleReview = async (
    flagId: string, 
    status: "reviewed" | "dismissed" | "escalated"
  ) => {
    const reviewer = "compliance-officer@verinode.com";
    const note = prompt(`Enter resolution rationale for flag (${status}):`, `Audit confirmed under IOSCO benchmark principle 7: ${status}`);
    if (!note) return;

    try {
      setUpdatingId(flagId);
      await api.reviewSurveillanceFlag(flagId, reviewer, note, status);
      await fetchFlags();
    } catch (err: unknown) {
      if (err instanceof Error) {
        alert(`Failed to review flag: ${err.message}`);
      }
    } finally {
      setUpdatingId(null);
    }
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-[#272727] pb-6">
        <div className="flex items-center gap-4">
          <Link
            href="/admin"
            className="p-2 rounded-lg border border-[#272727] bg-[#121212] hover:bg-[#181818] hover:text-white text-zinc-400 transition-colors active:scale-[0.98]"
          >
            <ArrowLeft className="h-4 w-4" />
          </Link>
          <div>
            <div className="flex items-center gap-2">
              <span className="text-xs font-mono text-zinc-500 uppercase">Phase 2 Surveillance Engine</span>
              <span className="inline-flex h-1.5 w-1.5 rounded-full bg-amber-400" />
            </div>
            <h1 className="text-xl sm:text-2xl font-bold tracking-tight text-white">
              Anti-Manipulation & Index Surveillance Queue
            </h1>
          </div>
        </div>

        <div className="flex items-center gap-3">
          <button
            onClick={() => fetchFlags()}
            disabled={loading}
            className="flex items-center gap-2 px-3 py-1.5 text-xs font-medium rounded-lg border border-[#272727] bg-[#121212] hover:bg-[#181818] text-zinc-300 transition-colors active:scale-[0.98]"
          >
            <RefreshCw className={`h-3.5 w-3.5 ${loading ? "animate-spin" : ""}`} />
            Refresh Queue
          </button>
        </div>
      </div>

      {/* Filter Tabs */}
      <div className="flex items-center gap-2 border-b border-[#272727] pb-3 text-xs font-mono">
        <span className="text-zinc-500 mr-2">FILTER STATUS:</span>
        {["all", "pending", "reviewed", "dismissed", "escalated"].map((st) => (
          <button
            key={st}
            onClick={() => setStatusFilter(st)}
            className={`px-2.5 py-1 rounded-md transition-colors ${
              statusFilter === st
                ? "bg-white text-black font-semibold"
                : "text-zinc-400 hover:text-white hover:bg-[#181818]"
            }`}
          >
            {st.toUpperCase()}
          </button>
        ))}
      </div>

      {error && (
        <div className="rounded-lg border border-red-500/30 bg-red-500/10 p-4 text-xs font-mono text-red-400">
          {error}
        </div>
      )}

      {/* Flags List */}
      <div className="space-y-4">
        {loading ? (
          <div className="rounded-xl border border-[#272727] bg-[#121212] p-8 text-center text-sm font-mono text-zinc-500">
            Scanning market data surveillance audit trail...
          </div>
        ) : flags.length === 0 ? (
          <div className="rounded-xl border border-[#272727] bg-[#121212] p-12 text-center space-y-3">
            <CheckCircle2 className="h-8 w-8 text-emerald-400 mx-auto" />
            <h3 className="text-sm font-bold text-white">No Surveillance Flags Found</h3>
            <p className="text-xs text-zinc-400 max-w-sm mx-auto">
              All trades, contributions, and participants meet market conduct and concentration standards.
            </p>
          </div>
        ) : (
          flags.map((flag) => (
            <div
              key={flag.id}
              className="rounded-xl border border-[#272727] bg-[#121212] p-5 space-y-4 hover:border-zinc-700 transition-colors"
            >
              <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 border-b border-[#272727] pb-3">
                <div className="flex items-center gap-3">
                  <div className="flex h-7 w-7 items-center justify-center rounded-md bg-amber-500/10 border border-amber-500/30 text-amber-400">
                    <ShieldAlert className="h-4 w-4" />
                  </div>
                  <div>
                    <div className="flex items-center gap-2">
                      <span className="font-mono text-sm font-bold uppercase text-white">
                        FLAG: {flag.flag_type.replace("_", " ")}
                      </span>
                      <span className="px-2 py-0.5 rounded text-[10px] font-mono uppercase bg-amber-500/20 text-amber-300 border border-amber-500/40">
                        {flag.severity}
                      </span>
                    </div>
                    <div className="text-xs font-mono text-zinc-500">
                      Subject: {flag.subject_type} • ID: {flag.subject_id}
                    </div>
                  </div>
                </div>

                <div className="flex items-center gap-2">
                  <span className={`px-2.5 py-1 rounded-full text-xs font-mono uppercase ${
                    flag.status === "pending"
                      ? "bg-amber-500/10 text-amber-400 border border-amber-500/30"
                      : flag.status === "reviewed"
                      ? "bg-emerald-500/10 text-emerald-400 border border-emerald-500/30"
                      : "bg-zinc-800 text-zinc-400 border border-zinc-700"
                  }`}>
                    {flag.status}
                  </span>
                </div>
              </div>

              {/* Details JSON / Notes */}
              <div className="rounded-lg border border-[#272727] bg-black p-3 text-xs font-mono space-y-2">
                <div className="text-zinc-500 uppercase text-[10px]">Surveillance Rule Engine Telemetry:</div>
                <div className="text-zinc-300 leading-relaxed">
                  {typeof flag.details === "object" ? JSON.stringify(flag.details, null, 2) : String(flag.details)}
                </div>
                {flag.resolution && (
                  <div className="border-t border-[#272727] pt-2 text-emerald-400">
                    <strong>Audit Resolution:</strong> {flag.resolution} (by {flag.reviewed_by})
                  </div>
                )}
              </div>

              {/* Action Buttons */}
              {flag.status === "pending" && (
                <div className="flex flex-wrap items-center justify-end gap-2 pt-2">
                  <button
                    onClick={() => handleReview(flag.id, "dismissed")}
                    disabled={updatingId === flag.id}
                    className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium rounded-md border border-[#272727] bg-[#181818] hover:bg-[#222222] text-zinc-300 transition-colors active:scale-[0.98]"
                  >
                    <XCircle className="h-3.5 w-3.5 text-zinc-400" />
                    Dismiss (False Positive)
                  </button>

                  <button
                    onClick={() => handleReview(flag.id, "reviewed")}
                    disabled={updatingId === flag.id}
                    className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium rounded-md border border-emerald-500/30 bg-emerald-500/10 hover:bg-emerald-500/20 text-emerald-300 transition-colors active:scale-[0.98]"
                  >
                    <CheckCircle2 className="h-3.5 w-3.5 text-emerald-400" />
                    Clear & Authorize for Fix
                  </button>

                  <button
                    onClick={() => handleReview(flag.id, "escalated")}
                    disabled={updatingId === flag.id}
                    className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium rounded-md border border-red-500/30 bg-red-500/10 hover:bg-red-500/20 text-red-300 transition-colors active:scale-[0.98]"
                  >
                    <AlertTriangle className="h-3.5 w-3.5 text-red-400" />
                    Escalate to Oversight Committee
                  </button>
                </div>
              )}
            </div>
          ))
        )}
      </div>
    </div>
  );
}
