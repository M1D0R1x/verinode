import Link from "next/link";
import { ShieldAlert, Users, FileText, Activity, ArrowRight, CheckCircle2, AlertTriangle, Layers } from "lucide-react";
import { api, Participant } from "@/lib/api";
import { Contract, Claim, ContractEvent } from "@/lib/types";
import { Badge } from "@/components/ui/badge";

export const dynamic = "force-dynamic";

export default async function AdminDashboardPage() {
  let participants: Participant[] = [];
  let contracts: Contract[] = [];
  let claims: Claim[] = [];
  let auditEvents: ContractEvent[] = [];

  try {
    participants = await api.listParticipants();
  } catch {
    // fallback if unseeded
  }

  try {
    contracts = await api.listContracts();
  } catch {
    // fallback
  }

  try {
    claims = await api.listClaims();
  } catch {
    // fallback
  }

  try {
    auditEvents = await api.listAdminAudit(50);
  } catch {
    // fallback
  }

  const pendingKYC = participants.filter((p) => !p.kyc_status || p.kyc_status === "pending" || p.kyc_status === "review");
  const openClaims = claims.filter((c) => c.state === "claim_open");
  const liveContracts = contracts.filter((c) => c.state === "live" || c.state === "delivery_test");

  return (
    <div className="space-y-8">
      {/* Page Header */}
      <div className="border-b border-border/80 pb-6">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
          <div>
            <div className="inline-flex items-center gap-2 px-2.5 py-0.5 rounded-full border border-primary/30 bg-primary/10 text-xs font-mono text-primary mb-2">
              <span>INSTITUTIONAL CLEARING & COMPLIANCE DESK</span>
            </div>
            <h1 className="text-2xl sm:text-3xl font-bold tracking-tight text-white">
              Operations & Risk Control
            </h1>
            <p className="text-sm text-muted mt-1">
              Bilateral counterparty onboarding, telemetry claims dispute desk, and immutable audit logs.
            </p>
          </div>
          <div className="flex items-center gap-2">
            <span className="inline-flex h-2.5 w-2.5 rounded-full bg-accent animate-pulse" />
            <span className="text-xs font-mono text-muted">Core Engine Online</span>
          </div>
        </div>
      </div>

      {/* KPI Summary Cards */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <div className="p-5 rounded-xl border border-border bg-surface/50 space-y-2">
          <div className="flex items-center justify-between text-muted">
            <span className="text-xs font-mono uppercase">Pending KYC</span>
            <Users className="h-4 w-4 text-warning" />
          </div>
          <div className="text-3xl font-bold font-mono text-white">{pendingKYC.length}</div>
          <p className="text-xs text-muted">
            {pendingKYC.length > 0 ? "Requires compliance sign-off" : "All participants approved"}
          </p>
        </div>

        <div className="p-5 rounded-xl border border-border bg-surface/50 space-y-2">
          <div className="flex items-center justify-between text-muted">
            <span className="text-xs font-mono uppercase">Open SLA Claims</span>
            <ShieldAlert className="h-4 w-4 text-danger" />
          </div>
          <div className="text-3xl font-bold font-mono text-white">{openClaims.length}</div>
          <p className="text-xs text-muted">
            {openClaims.length > 0 ? "Under telemetry dispute review" : "No active delivery claims"}
          </p>
        </div>

        <div className="p-5 rounded-xl border border-border bg-surface/50 space-y-2">
          <div className="flex items-center justify-between text-muted">
            <span className="text-xs font-mono uppercase">Active Allocations</span>
            <Activity className="h-4 w-4 text-accent" />
          </div>
          <div className="text-3xl font-bold font-mono text-white">{liveContracts.length}</div>
          <p className="text-xs text-muted">
            {contracts.length} total executed forward contracts
          </p>
        </div>

        <div className="p-5 rounded-xl border border-border bg-surface/50 space-y-2">
          <div className="flex items-center justify-between text-muted">
            <span className="text-xs font-mono uppercase">Audit Log Events</span>
            <FileText className="h-4 w-4 text-primary" />
          </div>
          <div className="text-3xl font-bold font-mono text-white">{auditEvents.length}</div>
          <p className="text-xs text-muted">
            Cryptographically sealed events
          </p>
        </div>
      </div>

      {/* Main Action Desks */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        <Link
          href="/admin/participants"
          className="group p-6 rounded-xl border border-border bg-surface/40 hover:bg-surface hover:border-primary/50 transition-all flex flex-col justify-between"
        >
          <div className="space-y-3">
            <div className="h-10 w-10 rounded-lg bg-warning/10 border border-warning/30 flex items-center justify-center text-warning">
              <Users className="h-5 w-5" />
            </div>
            <div>
              <h3 className="font-semibold text-white group-hover:text-primary transition-colors">
                Participant Approval Desk
              </h3>
              <p className="text-xs text-muted mt-1 leading-relaxed">
                Review legal entities, jurisdictions, sanctions screening results, and assign bilateral institutional credit limits.
              </p>
            </div>
          </div>
          <div className="mt-6 flex items-center text-xs font-medium text-primary gap-1">
            <span>Review {pendingKYC.length} Pending</span>
            <ArrowRight className="h-3.5 w-3.5 group-hover:translate-x-1 transition-transform" />
          </div>
        </Link>

        <Link
          href="/admin/claims"
          className="group p-6 rounded-xl border border-border bg-surface/40 hover:bg-surface hover:border-primary/50 transition-all flex flex-col justify-between"
        >
          <div className="space-y-3">
            <div className="h-10 w-10 rounded-lg bg-danger/10 border border-danger/30 flex items-center justify-center text-danger">
              <ShieldAlert className="h-5 w-5" />
            </div>
            <div>
              <h3 className="font-semibold text-white group-hover:text-primary transition-colors">
                SLA Claims Review Desk
              </h3>
              <p className="text-xs text-muted mt-1 leading-relaxed">
                Inspect canary test breaches, NCCL bandwidth shortfalls, and resolve or escalate delivery disputes under Master Legal Confirmations.
              </p>
            </div>
          </div>
          <div className="mt-6 flex items-center text-xs font-medium text-danger gap-1">
            <span>Manage {openClaims.length} Claims</span>
            <ArrowRight className="h-3.5 w-3.5 group-hover:translate-x-1 transition-transform" />
          </div>
        </Link>

        <Link
          href="/admin/audit"
          className="group p-6 rounded-xl border border-border bg-surface/40 hover:bg-surface hover:border-primary/50 transition-all flex flex-col justify-between"
        >
          <div className="space-y-3">
            <div className="h-10 w-10 rounded-lg bg-primary/10 border border-primary/30 flex items-center justify-center text-primary">
              <Layers className="h-5 w-5" />
            </div>
            <div>
              <h3 className="font-semibold text-white group-hover:text-primary transition-colors">
                Immutable Audit Trail
              </h3>
              <p className="text-xs text-muted mt-1 leading-relaxed">
                Query contract transition events, actor authorization decisions, idempotency keys, and SHA-256 evidence digests.
              </p>
            </div>
          </div>
          <div className="mt-6 flex items-center text-xs font-medium text-primary gap-1">
            <span>Inspect Audit Stream</span>
            <ArrowRight className="h-3.5 w-3.5 group-hover:translate-x-1 transition-transform" />
          </div>
        </Link>
      </div>

      {/* Recent Forward Contracts Table */}
      <div className="rounded-xl border border-border bg-surface/30 overflow-hidden">
        <div className="px-6 py-4 border-b border-border/80 flex items-center justify-between">
          <div>
            <h2 className="text-base font-semibold text-white">Institutional Forward Contracts</h2>
            <p className="text-xs text-muted">All physical capacity reservations keyed to canonical trade_id.</p>
          </div>
          <Badge variant="outline" className="font-mono text-xs">
            {contracts.length} Records
          </Badge>
        </div>

        {contracts.length === 0 ? (
          <div className="p-8 text-center text-muted text-sm font-mono">
            No contracts spawned yet. Counterparties can initiate an RFQ via the Buyer Portal.
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-left text-xs">
              <thead className="bg-surface/60 border-b border-border text-muted font-mono uppercase">
                <tr>
                  <th className="px-6 py-3">Canonical Trade ID</th>
                  <th className="px-6 py-3">Grade</th>
                  <th className="px-6 py-3">Status</th>
                  <th className="px-6 py-3">Buyer / Seller</th>
                  <th className="px-6 py-3">Executed At</th>
                  <th className="px-6 py-3 text-right">Confirmation</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border/60">
                {contracts.slice(0, 10).map((c) => (
                  <tr key={c.id} className="hover:bg-surface/50 transition-colors">
                    <td className="px-6 py-4 font-mono font-medium text-white">
                      <Link href={`/buyer/contracts/${c.id}`} className="hover:text-primary transition-colors">
                        {c.id}
                      </Link>
                    </td>
                    <td className="px-6 py-4 font-mono text-muted">{c.grade_id}</td>
                    <td className="px-6 py-4">
                      <Badge variant="default" className="capitalize">
                        {c.state.replace("_", " ")}
                      </Badge>
                    </td>
                    <td className="px-6 py-4 font-mono text-muted">
                      <div>B: {c.buyer_id.slice(0, 8)}...</div>
                      <div>S: {c.seller_id.slice(0, 8)}...</div>
                    </td>
                    <td className="px-6 py-4 text-muted">
                      {new Date(c.created_at).toLocaleDateString()}
                    </td>
                    <td className="px-6 py-4 text-right">
                      <Link
                        href={`/buyer/contracts/${c.id}`}
                        className="inline-flex items-center gap-1 text-primary hover:text-primary-hover font-medium"
                      >
                        <span>Inspect</span>
                        <ArrowRight className="h-3 w-3" />
                      </Link>
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
