"use client";

import { useState } from "react";
import { api, HedgeQuote } from "@/lib/api";
import { Loader2, TrendingDown, TrendingUp, MinusCircle, Info } from "lucide-react";

function usd(n: number) {
  return new Intl.NumberFormat("en-US", { style: "currency", currency: "USD", maximumFractionDigits: 0 }).format(n);
}

const ACTION_META: Record<string, { label: string; tone: string; icon: typeof TrendingDown }> = {
  SHORT_PERP: { label: "Short the perp", tone: "text-signal", icon: TrendingDown },
  LONG_PERP: { label: "Long the perp", tone: "text-verify", icon: TrendingUp },
  NO_HEDGE: { label: "Hold unhedged", tone: "text-muted", icon: MinusCircle },
};

export default function HedgePage() {
  const [duration, setDuration] = useState(168);
  const [gpus, setGpus] = useState(8);
  const [fixed, setFixed] = useState(2.2);
  const [index, setIndex] = useState(2.05);
  const [quote, setQuote] = useState<HedgeQuote | null>(null);
  const [loading, setLoading] = useState(false);
  const [err, setErr] = useState<string | null>(null);

  async function run() {
    setLoading(true);
    setErr(null);
    try {
      setQuote(
        await api.hedgeQuote({
          duration_hours: duration,
          gpu_count: gpus,
          fixed_rate_hourly: fixed,
          index_mark_price: index,
        })
      );
    } catch (e) {
      setErr(e instanceof Error ? e.message : "Failed to compute hedge");
    } finally {
      setLoading(false);
    }
  }

  const action = quote ? ACTION_META[quote.recommended_action] ?? ACTION_META.NO_HEDGE : null;

  return (
    <div className="mx-auto max-w-5xl px-4 py-16 sm:px-6 lg:px-8">
      <p className="eyebrow">Phase 4 · basis analytics</p>
      <h1 className="mt-3 font-serif text-headline text-parchment">Hedge desk</h1>
      <p className="mt-4 max-w-2xl text-muted">
        A physical forward locks a fixed price; the floating GPU-hour index drifts against it.
        This desk sizes the offsetting Hyperliquid perp position so a supplier or buyer can
        neutralize that basis — advisory analytics against a mature index, never a leveraged
        retail product, and never a substitute for the physical delivery itself.
      </p>

      <div className="mt-10 grid gap-6 lg:grid-cols-[1fr_1.1fr]">
        {/* Inputs */}
        <div className="card card-raise p-6">
          <h2 className="font-serif text-title text-parchment">Position</h2>
          <div className="mt-5 space-y-5">
            {[
              { label: "Tenor (hours)", value: duration, set: setDuration, step: 24, min: 24 },
              { label: "GPU count", value: gpus, set: setGpus, step: 8, min: 8 },
            ].map((f) => (
              <div key={f.label}>
                <label className="eyebrow">{f.label}</label>
                <input
                  type="number"
                  value={f.value}
                  min={f.min}
                  step={f.step}
                  onChange={(e) => f.set(Number(e.target.value))}
                  className="tabular mt-2 w-full rounded-lg border border-line bg-ink-800 px-4 py-2.5 text-sm text-parchment outline-none focus:border-signal"
                />
              </div>
            ))}
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="eyebrow">Fixed rate $/GPU-hr</label>
                <input
                  type="number"
                  value={fixed}
                  step={0.01}
                  onChange={(e) => setFixed(Number(e.target.value))}
                  className="tabular mt-2 w-full rounded-lg border border-line bg-ink-800 px-4 py-2.5 text-sm text-parchment outline-none focus:border-signal"
                />
              </div>
              <div>
                <label className="eyebrow">Index mark $/GPU-hr</label>
                <input
                  type="number"
                  value={index}
                  step={0.01}
                  onChange={(e) => setIndex(Number(e.target.value))}
                  className="tabular mt-2 w-full rounded-lg border border-line bg-ink-800 px-4 py-2.5 text-sm text-parchment outline-none focus:border-signal"
                />
              </div>
            </div>
            <button
              onClick={run}
              disabled={loading}
              className="inline-flex w-full items-center justify-center gap-2 rounded-lg bg-signal px-5 py-2.5 text-sm font-semibold text-ink-950 transition-all hover:bg-signal-bright active:scale-[0.98] disabled:opacity-50"
            >
              {loading ? <Loader2 className="h-4 w-4 animate-spin" /> : null}
              Compute hedge
            </button>
            {err && <p className="text-sm text-alert">{err}</p>}
          </div>
        </div>

        {/* Result */}
        <div className="card p-6">
          {!quote ? (
            <div className="flex h-full min-h-[280px] flex-col items-center justify-center text-center text-muted-soft">
              <Info className="h-6 w-6" />
              <p className="mt-3 text-sm">Enter a position and compute the basis hedge.</p>
            </div>
          ) : (
            <div className="animate-rise">
              <div className="flex items-center justify-between">
                <span className="eyebrow">Recommendation · {quote.market}</span>
                <span className="tabular text-xs text-muted-soft">{quote.total_gpu_hours.toLocaleString()} GPU-hrs</span>
              </div>

              {action && (
                <div className={`mt-4 flex items-center gap-3 ${action.tone}`}>
                  <action.icon className="h-7 w-7" />
                  <span className="font-serif text-2xl">{action.label}</span>
                </div>
              )}

              <div className="mt-6 grid grid-cols-2 gap-px overflow-hidden rounded-xl border border-line bg-line">
                {[
                  { k: "Physical notional", v: usd(quote.physical_contract_usd) },
                  { k: "Floating notional", v: usd(quote.floating_index_usd) },
                  {
                    k: "Basis spread",
                    v: `${quote.basis_spread_usd >= 0 ? "+" : ""}${usd(quote.basis_spread_usd)}`,
                    tone: quote.basis_spread_usd >= 0 ? "text-signal" : "text-verify",
                  },
                  {
                    k: "Basis %",
                    v: `${quote.basis_spread_pct >= 0 ? "+" : ""}${quote.basis_spread_pct.toFixed(2)}%`,
                    tone: quote.basis_spread_pct >= 0 ? "text-signal" : "text-verify",
                  },
                ].map((c) => (
                  <div key={c.k} className="bg-ink-850 p-4">
                    <div className="eyebrow">{c.k}</div>
                    <div className={`tabular mt-1.5 text-lg font-semibold ${c.tone ?? "text-parchment"}`}>{c.v}</div>
                  </div>
                ))}
              </div>

              <div className="mt-4 flex items-center justify-between rounded-lg border border-line bg-ink-900/60 px-4 py-3">
                <span className="text-sm text-muted">Annualized basis</span>
                <span className="tabular text-sm font-semibold text-parchment">
                  {quote.annualized_basis_pct.toFixed(1)}%
                </span>
              </div>

              <p className="mt-4 text-sm leading-relaxed text-muted">{quote.rationale}</p>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
