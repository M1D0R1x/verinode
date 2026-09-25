"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { ArrowLeft, ShieldAlert, CheckCircle2, AlertTriangle, RefreshCw, ExternalLink, Scale } from "lucide-react";
import { api } from "@/lib/api";
import { Claim } from "@/lib/types";
import { Badge } from "@/components/ui/badge";

export default function ClaimsAdminPage() {
  const [claims, setClaims] = useState<Claim[]>([]);
  const [loading, setLoading] = useState(true);
  const [updatingId, setUpdatingId] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  const fetchClaims = async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await api.listClaims();
      setClaims(data);
    } catch (err: unknown) {
      if (err instanceof Error) {
        setError(err.message);
      } else {
        setError("Failed to query claims");
      }
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchClaims();
  }, []);

  const handleResolveClaim = async (claimId: string, targetState: "resolved" | "disputed") => {
    try {
      setUpdatingId(claimId);
      await api.resolveClaim(claimId, targetState);
      await fetchClaims();
    } catch (err: unknown) {
      if (err instanceof Error) {
        alert(`Failed to update claim: ${err.message}`);
      }
    } finally {
      setUpdatingId(null);
    }
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4 border-b border-border/80 pb-6">
        <div className="flex items-center gap-4">
          <Link
            href="/admin"
            className="p-2 rounded-lg border border-border bg-surface/50 hover:bg-surface hover:text-white text-muted transition-colors"
          >
            <ArrowLeft className="h-4 w-4" />
          </Link>
          <div>
            <h1 className="text-2xl font-bold tracking-tight text-white">SLA Claims & Dispute Desk</h1>
            <p className="text-xs text-muted mt-0.5">
              Review telemetry canary failures, NCCL bandwidth breach claims, and resolve bilateral contractual remedies.
            </p>
          </div>
        </div>

        <button
          onClick={fetchClaims}
          disabled={loading}
          className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg border border-border bg-surface text-xs font-medium text-muted hover:text-white transition-colors"
        >
          <RefreshCw className={`h-3.5 w-3.5 ${loading ? "animate-spin" : ""}`} />
          <span>Refresh</span>
        </button>
      </div>

      {error && (
        <div className="p-4 rounded-xl border border-danger/30 bg-danger/10 text-danger text-xs flex items-center gap-3">
          <AlertTriangle className="h-4 w-4 shrink-0" />
          <span>{error}</span>
        </div>
      )}

      {/* Claims Review Table */}
      <div className="rounded-xl border border-border bg-surface/30 overflow-hidden">
        <div className="px-6 py-4 border-b border-border/80 flex items-center justify-between">
          <h2 className="text-sm font-semibold text-white">Institutional Claims Queue</h2>
          <Badge variant="outline" className="font-mono text-xs">
            {claims.length} Total Claims
          </Badge>
        </div>

        {loading ? (
          <div className="p-12 text-center text-muted font-mono text-xs">
            Loading claims database...
          </div>
        ) : claims.length === 0 ? (
          <div className="p-12 text-center text-muted font-mono text-xs">
            No active claims or disputes. All physical forward contracts are operating within canary specifications.
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-left text-xs">
              <thead className="bg-surface/60 border-b border-border text-muted font-mono uppercase">
                <tr>
                  <th className="px-6 py-3">Claim ID</th>
                  <th className="px-6 py-3">Trade ID</th>
                  <th className="px-6 py-3">Type</th>
                  <th className="px-6 py-3">Status</th>
                  <th className="px-6 py-3">Filed At</th>
                  <th className="px-6 py-3 text-right">Review Action</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border/60">
                {claims.map((c) => {
                  const isOpen = c.state === "claim_open";
                  const isResolved = c.state === "resolved";
                  const isDisputed = c.state === "disputed";

                  return (
                    <tr key={c.id} className="hover:bg-surface/50 transition-colors">
                      <td className="px-6 py-4 font-mono font-medium text-white">
                        {c.id.slice(0, 8)}...
                      </td>
                      <td className="px-6 py-4 font-mono text-muted">
                        <Link
                          href={`/buyer/contracts/${c.contract_id}`}
                          className="hover:text-primary transition-colors inline-flex items-center gap-1"
                        >
                          <span>{c.contract_id.slice(0, 8)}...</span>
                          <ExternalLink className="h-3 w-3" />
                        </Link>
                      </td>
                      <td className="px-6 py-4">
                        <Badge variant="destructive" className="capitalize font-mono text-[10px]">
                          {c.type.replace("_", " ")}
                        </Badge>
                      </td>
                      <td className="px-6 py-4">
                        <Badge
                          variant={isOpen ? "warning" : isResolved ? "success" : "destructive"}
                          className="capitalize"
                        >
                          {c.state.replace("_", " ")}
                        </Badge>
                      </td>
                      <td className="px-6 py-4 text-muted">
                        {new Date(c.created_at).toLocaleString()}
                      </td>
                      <td className="px-6 py-4 text-right">
                        {isOpen ? (
                          <div className="inline-flex items-center gap-2">
                            <button
                              onClick={() => handleResolveClaim(c.id, "resolved")}
                              disabled={updatingId === c.id}
                              className="inline-flex items-center gap-1 px-2.5 py-1 rounded bg-accent/10 border border-accent/30 text-accent hover:bg-accent hover:text-white transition-colors text-xs font-medium"
                            >
                              <CheckCircle2 className="h-3 w-3" />
                              <span>Grant Remedy</span>
                            </button>
                            <button
                              onClick={() => handleResolveClaim(c.id, "disputed")}
                              disabled={updatingId === c.id}
                              className="inline-flex items-center gap-1 px-2.5 py-1 rounded bg-danger/10 border border-danger/30 text-danger hover:bg-danger hover:text-white transition-colors text-xs font-medium"
                            >
                              <Scale className="h-3 w-3" />
                              <span>Escalate Dispute</span>
                            </button>
                          </div>
                        ) : (
                          <span className="text-xs text-muted font-mono capitalize">
                            Closed ({c.state})
                          </span>
                        )}
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
}
