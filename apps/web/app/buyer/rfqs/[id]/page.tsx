"use client";

import { useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { ArrowLeft, Clock, ShieldCheck, Check, ArrowRight, Server } from "lucide-react";
import { formatCents } from "@/lib/utils";
import { api } from "@/lib/api";

interface MockQuote {
  id: string;
  seller_name: string;
  seller_jurisdiction: string;
  hourly_cents: number;
  total_cents: number;
  expires_in_hours: number;
  facility_type: string;
  allreduce_historical_gbps: number;
}

const mockQuotes: MockQuote[] = [
  {
    id: "quote_88fa12-001",
    seller_name: "CoreWeave Hyperscale Partner LLC",
    seller_jurisdiction: "US (Delaware)",
    hourly_cents: 20800, // $208.00 / hr
    total_cents: 3494400, // $34,944.00 (168h)
    expires_in_hours: 14,
    facility_type: "Tier 3+ Equinix Ashburn",
    allreduce_historical_gbps: 432.1,
  },
  {
    id: "quote_77bc34-002",
    seller_name: "Nebula Compute Infrastructure",
    seller_jurisdiction: "US (Texas)",
    hourly_cents: 21500, // $215.00 / hr
    total_cents: 3612000, // $36,120.00
    expires_in_hours: 22,
    facility_type: "Tier 3 CyrusOne Dallas",
    allreduce_historical_gbps: 426.8,
  },
  {
    id: "quote_66da99-003",
    seller_name: "Vantage High Performance Nodes",
    seller_jurisdiction: "US (Virginia)",
    hourly_cents: 21900, // $219.00 / hr
    total_cents: 3679200, // $36,792.00
    expires_in_hours: 6,
    facility_type: "Tier 3 Digital Realty",
    allreduce_historical_gbps: 418.5,
  },
];

export default function RFQQuotesPage({ params }: { params: { id: string } }) {
  const router = useRouter();
  const [selectedQuoteId, setSelectedQuoteId] = useState<string | null>(null);
  const [isAccepting, setIsAccepting] = useState(false);

  const handleAccept = async (quote: MockQuote) => {
    setSelectedQuoteId(quote.id);
    setIsAccepting(true);

    try {
      const res = await api.acceptQuote(
        params.id,
        quote.id,
        "00000000-0000-0000-0000-000000000001"
      );
      if (res && res.contract_id) {
        router.push(`/buyer/contracts/${res.contract_id}`);
        return;
      }
    } catch {
      // Graceful fallback for offline demo
    }

    setTimeout(() => {
      setIsAccepting(false);
      router.push(`/buyer/contracts/trade_c8f921a4-9b2e-4b13-91dc-837264819011`);
    }, 1000);
  };

  return (
    <div className="mx-auto max-w-5xl px-4 py-8 sm:px-6 lg:px-8 space-y-6">
      <Link
        href="/buyer"
        className="inline-flex items-center gap-1.5 text-xs text-muted hover:text-white transition-colors"
      >
        <ArrowLeft className="h-3.5 w-3.5" />
        <span>Back to Buyer Console</span>
      </Link>

      {/* RFQ Header */}
      <div className="p-5 rounded-xl bg-surface border border-border flex flex-col md:flex-row md:items-center justify-between gap-4">
        <div>
          <div className="flex items-center gap-2">
            <span className="font-mono text-xs text-muted">RFQ ID: {params.id.slice(0, 16)}...</span>
            <span className="px-2 py-0.5 rounded-full text-[11px] font-mono bg-purple-950 text-purple-300 border border-purple-800">
              Quotes Received (3)
            </span>
          </div>
          <h1 className="text-xl font-bold text-white tracking-tight mt-1">
            8x NVIDIA H100 SXM 80GB (168-Hour Block)
          </h1>
          <p className="text-xs text-muted font-mono mt-0.5">
            Region: US-East • Window: Oct 05 – Oct 12, 2026 • Benchmark: NCCL $\ge$ 400 GB/s
          </p>
        </div>

        <div className="text-left md:text-right font-mono text-xs">
          <span className="text-muted block text-[10px]">TIME TO EXPIRATION</span>
          <span className="text-amber-400 font-bold flex items-center md:justify-end gap-1 mt-0.5">
            <Clock className="h-3.5 w-3.5" />
            <span>06h 42m remaining</span>
          </span>
        </div>
      </div>

      {/* QUOTES MATRIX */}
      <div className="space-y-3">
        <h2 className="text-sm font-bold text-white uppercase tracking-wider font-mono">
          Received Binding Quotes ({mockQuotes.length})
        </h2>

        <div className="space-y-3">
          {mockQuotes.map((q, idx) => {
            const isBest = idx === 0;
            return (
              <div
                key={q.id}
                className={`p-5 rounded-xl border transition-all ${
                  isBest
                    ? "bg-surface border-primary/60 shadow-lg shadow-primary/5"
                    : "bg-surface border-border hover:border-borderHighlight"
                }`}
              >
                <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
                  <div className="space-y-1.5">
                    <div className="flex items-center gap-2">
                      <span className="font-bold text-white text-base">{q.seller_name}</span>
                      {isBest && (
                        <span className="px-2 py-0.5 rounded text-[10px] font-mono bg-emerald-950 text-emerald-300 border border-emerald-800 font-semibold">
                          Best Offer
                        </span>
                      )}
                    </div>

                    <div className="flex flex-wrap items-center gap-4 text-xs font-mono text-muted">
                      <span>Jurisdiction: <strong className="text-zinc-300">{q.seller_jurisdiction}</strong></span>
                      <span>Facility: <strong className="text-zinc-300">{q.facility_type}</strong></span>
                      <span className="text-emerald-400">
                        Historical Telemetry: <strong>{q.allreduce_historical_gbps} GB/s</strong>
                      </span>
                    </div>
                  </div>

                  <div className="flex items-center justify-between sm:justify-end gap-6 pt-3 sm:pt-0 border-t sm:border-t-0 border-border/60">
                    <div className="text-left sm:text-right font-mono">
                      <span className="text-lg font-bold text-white block">
                        {formatCents(q.total_cents)}
                      </span>
                      <span className="text-xs text-muted block">
                        ${(q.hourly_cents / 100).toFixed(2)}/hr (${(q.hourly_cents / 100 / 8).toFixed(2)}/GPU-hr)
                      </span>
                    </div>

                    <button
                      onClick={() => handleAccept(q)}
                      disabled={isAccepting}
                      className="inline-flex items-center gap-1.5 px-4 py-2.5 rounded-lg bg-primary hover:bg-primary-hover disabled:bg-primary/50 text-white font-medium text-xs transition-all shadow-sm shrink-0"
                    >
                      {isAccepting && selectedQuoteId === q.id ? (
                        <span>Binding Contract...</span>
                      ) : (
                        <>
                          <span>Accept & Sign</span>
                          <ArrowRight className="h-4 w-4" />
                        </>
                      )}
                    </button>
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      </div>
    </div>
  );
}
