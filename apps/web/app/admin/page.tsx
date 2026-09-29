"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { ShieldAlert, Users, Activity, ScrollText, ArrowRight, Radar } from "lucide-react";
import { api, Participant, SurveillanceFlag } from "@/lib/api";
import { Contract, Claim, ContractEvent } from "@/lib/types";

export default function AdminDashboardPage() {
  const [participants, setParticipants] = useState<Participant[]>([]);
  const [contracts, setContracts] = useState<Contract[]>([]);
  const [claims, setClaims] = useState<Claim[]>([]);
  const [audit, setAudit] = useState<ContractEvent[]>([]);
  const [flags, setFlags] = useState<SurveillanceFlag[]>([]);

  useEffect(() => {
    api.listParticipants().then(setParticipants).catch(() => {});
    api.listContracts().then(setContracts).catch(() => {});
    api.listClaims().then(setClaims).catch(() => {});
    api.listAdminAudit(50).then(setAudit).catch(() => {});
    api.listSurveillanceFlags().then(setFlags).catch(() => {});
  }, []);

  const pendingKYC = participants.filter((p) => !p.kyc_status || p.kyc_status === "pending" || p.kyc_status === "review");
  const openClaims = claims.filter((c) => c.state === "claim_open");
  const liveContracts = contracts.filter((c) => c.state === "live" || c.state === "delivery_test");
  const pendingFlags = flags.filter((f) => f.status === "pending");

  const kpis = [
    { label: "Pending KYC", value: pendingKYC.length, icon: Users, hint: pendingKYC.length ? "Requires compliance sign-off" : "All approved" },
    { label: "Open SLA claims", value: openClaims.length, icon: ShieldAlert, hint: openClaims.length ? "Under dispute review" : "No active claims" },
    { label: "Active allocations", value: liveContracts.length, icon: Activity, hint: `${contracts.length} total contracts` },
    { label: "Audit events", value: audit.length, icon: ScrollText, hint: "Cryptographically sealed" },
  ];

  const desks = [
    { href: "/admin/participants", icon: Users, title: "Participant approval", body: "Review legal entities, sanctions results, and set bilateral credit limits.", cta: `Review ${pendingKYC.length} pending` },
    { href: "/admin/claims", icon: ShieldAlert, title: "SLA claims review", body: "Inspect canary breaches and resolve or escalate delivery disputes.", cta: `Manage ${openClaims.length} claims` },
    { href: "/admin/surveillance", icon: Radar, title: "Surveillance", body: "Wash-trade, related-party and concentration review before index inclusion.", cta: `Review ${pendingFlags.length} alerts` },
    { href: "/admin/audit", icon: ScrollText, title: "Immutable audit", body: "Contract transitions, actor decisions, idempotency keys, evidence digests.", cta: "Inspect stream" },
  ];

  return (
    <div className="space-y-8">
      <div className="grid grid-cols-2 gap-4 lg:grid-cols-4">
        {kpis.map((k) => (
          <div key={k.label} className="card p-5">
            <div className="flex items-center justify-between">
              <span className="eyebrow">{k.label}</span>
              <k.icon className="h-4 w-4 text-signal" />
            </div>
            <div className="tabular mt-2 text-3xl font-semibold text-parchment">{k.value}</div>
            <p className="mt-1 text-xs text-muted-soft">{k.hint}</p>
          </div>
        ))}
      </div>

      <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-4">
        {desks.map((d) => (
          <Link key={d.href} href={d.href} className="group flex flex-col justify-between card p-6 transition-colors hover:border-signal/40">
            <div>
              <div className="flex h-10 w-10 items-center justify-center rounded-xl border border-line bg-ink-850 text-signal">
                <d.icon className="h-5 w-5" />
              </div>
              <h3 className="mt-4 font-serif text-title text-parchment">{d.title}</h3>
              <p className="mt-2 text-xs leading-relaxed text-muted">{d.body}</p>
            </div>
            <div className="mt-5 inline-flex items-center gap-1 text-xs font-medium text-signal">
              {d.cta} <ArrowRight className="h-3.5 w-3.5 transition-transform group-hover:translate-x-0.5" />
            </div>
          </Link>
        ))}
      </div>

      <div className="card overflow-hidden">
        <div className="flex items-center justify-between border-b border-line px-6 py-4">
          <div>
            <h2 className="font-serif text-title text-parchment">Forward contracts</h2>
            <p className="text-xs text-muted-soft">All reservations keyed to the canonical trade_id.</p>
          </div>
          <span className="tabular rounded-pill border border-line px-2.5 py-1 text-xs text-muted">{contracts.length} records</span>
        </div>
        {contracts.length === 0 ? (
          <div className="p-10 text-center text-sm text-muted-soft">
            No contracts yet. Counterparties initiate an RFQ from the buyer portal.
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-left text-xs">
              <thead className="border-b border-line bg-ink-900/60 text-muted-soft">
                <tr>
                  {["Trade ID", "Grade", "Status", "Buyer / Seller", "Executed", ""].map((h) => (
                    <th key={h} className="px-6 py-3 font-medium uppercase tracking-wide">{h}</th>
                  ))}
                </tr>
              </thead>
              <tbody className="divide-y divide-lineSoft">
                {contracts.slice(0, 10).map((c) => (
                  <tr key={c.id} className="transition-colors hover:bg-ink-850/60">
                    <td className="tabular px-6 py-4 text-parchment">
                      <Link href={`/buyer/contracts/${c.id}`} className="hover:text-signal">{c.id.slice(0, 18)}…</Link>
                    </td>
                    <td className="tabular px-6 py-4 text-muted">{c.grade_id}</td>
                    <td className="px-6 py-4">
                      <span className="rounded-pill border border-line px-2 py-0.5 text-xs capitalize text-parchment">
                        {c.state.replace(/_/g, " ")}
                      </span>
                    </td>
                    <td className="tabular px-6 py-4 text-muted-soft">
                      <div>B {c.buyer_id.slice(0, 8)}…</div>
                      <div>S {c.seller_id.slice(0, 8)}…</div>
                    </td>
                    <td className="px-6 py-4 text-muted-soft">{new Date(c.created_at).toLocaleDateString()}</td>
                    <td className="px-6 py-4 text-right">
                      <Link href={`/buyer/contracts/${c.id}`} className="inline-flex items-center gap-1 text-signal hover:text-signal-bright">
                        Inspect <ArrowRight className="h-3 w-3" />
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
