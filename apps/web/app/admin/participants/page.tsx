"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { ArrowLeft, CheckCircle2, XCircle, ShieldCheck, AlertCircle, RefreshCw, UserCheck } from "lucide-react";
import { api, Participant } from "@/lib/api";
import { Badge } from "@/components/ui/badge";

export default function ParticipantsAdminPage() {
  const [participants, setParticipants] = useState<Participant[]>([]);
  const [loading, setLoading] = useState(true);
  const [updatingId, setUpdatingId] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  const fetchParticipants = async () => {
    try {
      setLoading(true);
      setError(null);
      const data = await api.listParticipants();
      setParticipants(data);
    } catch (err: unknown) {
      if (err instanceof Error) {
        setError(err.message);
      } else {
        setError("Failed to load participants");
      }
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchParticipants();
  }, []);

  const handleUpdateStatus = async (id: string, status: "approved" | "rejected") => {
    try {
      setUpdatingId(id);
      await api.updateParticipantKYC(id, status);
      await fetchParticipants();
    } catch (err: unknown) {
      if (err instanceof Error) {
        alert(`Failed to update KYC: ${err.message}`);
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
            <h1 className="text-2xl font-bold tracking-tight text-white">Participant KYC & Counterparty Desk</h1>
            <p className="text-xs text-muted mt-0.5">
              Review institutional entity verification, sanctions checks, and approve bilateral credit limits.
            </p>
          </div>
        </div>

        <button
          onClick={fetchParticipants}
          disabled={loading}
          className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg border border-border bg-surface text-xs font-medium text-muted hover:text-white transition-colors"
        >
          <RefreshCw className={`h-3.5 w-3.5 ${loading ? "animate-spin" : ""}`} />
          <span>Refresh</span>
        </button>
      </div>

      {error && (
        <div className="p-4 rounded-xl border border-danger/30 bg-danger/10 text-danger text-xs flex items-center gap-3">
          <AlertCircle className="h-4 w-4 shrink-0" />
          <span>{error}</span>
        </div>
      )}

      {/* Participants Table */}
      <div className="rounded-xl border border-border bg-surface/30 overflow-hidden">
        <div className="px-6 py-4 border-b border-border/80 flex items-center justify-between">
          <h2 className="text-sm font-semibold text-white">Registered Counterparties</h2>
          <Badge variant="outline" className="font-mono text-xs">
            {participants.length} Entities
          </Badge>
        </div>

        {loading ? (
          <div className="p-12 text-center text-muted font-mono text-xs">
            Loading counterparty database...
          </div>
        ) : participants.length === 0 ? (
          <div className="p-12 text-center text-muted font-mono text-xs">
            No participants registered.
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-left text-xs">
              <thead className="bg-surface/60 border-b border-border text-muted font-mono uppercase">
                <tr>
                  <th className="px-6 py-3">Entity Name</th>
                  <th className="px-6 py-3">Jurisdiction</th>
                  <th className="px-6 py-3">Role</th>
                  <th className="px-6 py-3">KYC Status</th>
                  <th className="px-6 py-3">Credit Limit</th>
                  <th className="px-6 py-3 text-right">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border/60">
                {participants.map((p) => {
                  const isApproved = p.kyc_status === "approved";
                  const isRejected = p.kyc_status === "rejected";
                  const isPending = !p.kyc_status || p.kyc_status === "pending" || p.kyc_status === "review";

                  return (
                    <tr key={p.id} className="hover:bg-surface/50 transition-colors">
                      <td className="px-6 py-4">
                        <div className="font-semibold text-white">{p.legal_name}</div>
                        <div className="text-[10px] font-mono text-muted">{p.id}</div>
                      </td>
                      <td className="px-6 py-4 text-muted">{p.jurisdiction}</td>
                      <td className="px-6 py-4">
                        <Badge variant="outline" className="capitalize font-mono">
                          {p.role}
                        </Badge>
                      </td>
                      <td className="px-6 py-4">
                        <Badge
                          variant={isApproved ? "success" : isRejected ? "destructive" : "warning"}
                          className="capitalize"
                        >
                          {p.kyc_status || "pending"}
                        </Badge>
                      </td>
                      <td className="px-6 py-4 font-mono text-muted">
                        {p.credit_limit_cents
                          ? `$${(p.credit_limit_cents / 100).toLocaleString()}`
                          : "$500,000"}
                      </td>
                      <td className="px-6 py-4 text-right">
                        <div className="inline-flex items-center gap-2">
                          {isPending && p.id && (
                            <>
                              <button
                                onClick={() => handleUpdateStatus(p.id!, "approved")}
                                disabled={updatingId === p.id}
                                className="inline-flex items-center gap-1 px-2.5 py-1 rounded bg-accent/10 border border-accent/30 text-accent hover:bg-accent hover:text-white transition-colors text-xs font-medium"
                              >
                                <CheckCircle2 className="h-3 w-3" />
                                <span>Approve</span>
                              </button>
                              <button
                                onClick={() => handleUpdateStatus(p.id!, "rejected")}
                                disabled={updatingId === p.id}
                                className="inline-flex items-center gap-1 px-2.5 py-1 rounded bg-danger/10 border border-danger/30 text-danger hover:bg-danger hover:text-white transition-colors text-xs font-medium"
                              >
                                <XCircle className="h-3 w-3" />
                                <span>Reject</span>
                              </button>
                            </>
                          )}
                          {isApproved && (
                            <span className="inline-flex items-center gap-1 text-accent text-xs font-mono">
                              <ShieldCheck className="h-3.5 w-3.5" />
                              <span>Verified</span>
                            </span>
                          )}
                          {isRejected && (
                            <span className="inline-flex items-center gap-1 text-danger text-xs font-mono">
                              <XCircle className="h-3.5 w-3.5" />
                              <span>Disqualified</span>
                            </span>
                          )}
                        </div>
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
