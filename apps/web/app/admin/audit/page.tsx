"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { ArrowLeft, Layers, RefreshCw, Search, ArrowRight, ShieldCheck, Hash } from "lucide-react";
import { api } from "@/lib/api";
import { ContractEvent } from "@/lib/types";
import { Badge } from "@/components/ui/badge";

export default function AdminAuditPage() {
  const [events, setEvents] = useState<ContractEvent[]>([]);
  const [loading, setLoading] = useState(true);
  const [filterQuery, setFilterQuery] = useState("");
  const [error, setError] = useState<string | null>(null);

  const fetchAuditEvents = async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await api.listAdminAudit(100);
      setEvents(data);
    } catch (err: unknown) {
      if (err instanceof Error) {
        setError(err.message);
      } else {
        setError("Failed to query audit stream");
      }
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchAuditEvents();
  }, []);

  const filtered = events.filter((e) => {
    if (!filterQuery) return true;
    const q = filterQuery.toLowerCase();
    return (
      e.contract_id.toLowerCase().includes(q) ||
      e.actor.toLowerCase().includes(q) ||
      e.reason.toLowerCase().includes(q) ||
      e.new_state.toLowerCase().includes(q)
    );
  });

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
            <h1 className="text-2xl font-bold tracking-tight text-white">Immutable Contract Audit Trail</h1>
            <p className="text-xs text-muted mt-0.5">
              Append-only state machine ledger verifying transitions, actor decisions, and SHA-256 evidence digests.
            </p>
          </div>
        </div>

        <button
          onClick={fetchAuditEvents}
          disabled={loading}
          className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg border border-border bg-surface text-xs font-medium text-muted hover:text-white transition-colors"
        >
          <RefreshCw className={`h-3.5 w-3.5 ${loading ? "animate-spin" : ""}`} />
          <span>Refresh</span>
        </button>
      </div>

      {/* Search and Filters */}
      <div className="flex items-center gap-3">
        <div className="relative flex-1 max-w-md">
          <Search className="absolute left-3 top-2.5 h-4 w-4 text-muted" />
          <input
            type="text"
            placeholder="Search by Trade ID, Actor, or State..."
            value={filterQuery}
            onChange={(e) => setFilterQuery(e.target.value)}
            className="w-full bg-surface border border-border rounded-lg pl-9 pr-4 py-2 text-xs text-white placeholder-muted focus:outline-none focus:border-primary font-mono"
          />
        </div>
        <span className="text-xs font-mono text-muted">
          Showing {filtered.length} of {events.length} events
        </span>
      </div>

      {/* Audit Log Table */}
      <div className="rounded-xl border border-border bg-surface/30 overflow-hidden">
        <div className="px-6 py-4 border-b border-border/80 flex items-center justify-between">
          <div className="flex items-center gap-2">
            <ShieldCheck className="h-4 w-4 text-accent" />
            <h2 className="text-sm font-semibold text-white">Cryptographically Keyed Event Log</h2>
          </div>
          <Badge variant="outline" className="font-mono text-xs">
            PostgreSQL Append-Only
          </Badge>
        </div>

        {loading ? (
          <div className="p-12 text-center text-muted font-mono text-xs">
            Loading immutable audit trail...
          </div>
        ) : filtered.length === 0 ? (
          <div className="p-12 text-center text-muted font-mono text-xs">
            No audit events found matching query.
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-left text-xs">
              <thead className="bg-surface/60 border-b border-border text-muted font-mono uppercase">
                <tr>
                  <th className="px-6 py-3">Timestamp</th>
                  <th className="px-6 py-3">Canonical Trade ID</th>
                  <th className="px-6 py-3">Transition</th>
                  <th className="px-6 py-3">Actor</th>
                  <th className="px-6 py-3">Reason & Context</th>
                  <th className="px-6 py-3">Idempotency Key</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border/60">
                {filtered.map((e) => (
                  <tr key={e.id} className="hover:bg-surface/50 transition-colors">
                    <td className="px-6 py-4 text-muted whitespace-nowrap">
                      {new Date(e.created_at).toLocaleString()}
                    </td>
                    <td className="px-6 py-4 font-mono font-medium text-white">
                      <Link
                        href={`/buyer/contracts/${e.contract_id}`}
                        className="hover:text-primary transition-colors inline-flex items-center gap-1"
                      >
                        <span>{e.contract_id.slice(0, 8)}...</span>
                      </Link>
                    </td>
                    <td className="px-6 py-4">
                      <div className="inline-flex items-center gap-1.5 font-mono text-[11px]">
                        <span className="text-muted">{e.prior_state}</span>
                        <ArrowRight className="h-3 w-3 text-primary" />
                        <span className="text-white font-semibold">{e.new_state}</span>
                      </div>
                    </td>
                    <td className="px-6 py-4 font-mono text-muted">{e.actor}</td>
                    <td className="px-6 py-4">
                      <div className="text-white max-w-xs truncate">{e.reason}</div>
                      {e.evidence_hash && (
                        <div className="text-[10px] font-mono text-muted flex items-center gap-1 mt-0.5">
                          <Hash className="h-2.5 w-2.5" />
                          <span>{e.evidence_hash.slice(0, 16)}...</span>
                        </div>
                      )}
                    </td>
                    <td className="px-6 py-4 font-mono text-muted text-[10px]">
                      {e.idempotency_key}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
}
