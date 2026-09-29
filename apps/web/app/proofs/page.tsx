"use client";

import { useEffect, useState } from "react";
import { api, ChainStatus, ProofBundle } from "@/lib/api";
import { ArrowUpRight, Loader2, ShieldCheck, Link2, Copy, Check } from "lucide-react";

const CHAIN_META: Record<string, { label: string; cluster: string; tone: string }> = {
  solana: { label: "Solana", cluster: "devnet", tone: "text-signal" },
  arbitrum: { label: "Arbitrum", cluster: "sepolia", tone: "text-parchment" },
  hyperliquid: { label: "Hyperliquid", cluster: "testnet", tone: "text-verify" },
};

function ModePill({ mode }: { mode: string }) {
  const live = mode.startsWith("live");
  return (
    <span
      className={`inline-flex items-center gap-1.5 rounded-pill border px-2.5 py-0.5 text-xs ${
        live
          ? "border-verify/40 bg-verify-wash text-verify"
          : "border-line bg-ink-850 text-muted"
      }`}
    >
      <span className={`h-1.5 w-1.5 rounded-full ${live ? "bg-verify" : "bg-muted-soft"}`} />
      {live ? "Live" : "Simulated"}
    </span>
  );
}

function CopyHash({ value }: { value: string }) {
  const [copied, setCopied] = useState(false);
  return (
    <button
      onClick={() => {
        navigator.clipboard.writeText(value);
        setCopied(true);
        setTimeout(() => setCopied(false), 1200);
      }}
      className="inline-flex items-center gap-1.5 text-muted transition-colors hover:text-parchment"
    >
      <span className="tabular text-xs">{value.length > 28 ? `${value.slice(0, 14)}…${value.slice(-8)}` : value}</span>
      {copied ? <Check className="h-3.5 w-3.5 text-verify" /> : <Copy className="h-3.5 w-3.5" />}
    </button>
  );
}

export default function ProofsPage() {
  const [status, setStatus] = useState<ChainStatus | null>(null);
  const [tradeId, setTradeId] = useState("c37610f3-4aea-414f-868c-5937e97e822d");
  const [bundle, setBundle] = useState<ProofBundle | null>(null);
  const [loading, setLoading] = useState(false);
  const [err, setErr] = useState<string | null>(null);

  useEffect(() => {
    api.getChainStatus().then(setStatus).catch(() => setStatus(null));
  }, []);

  async function mirror() {
    setLoading(true);
    setErr(null);
    try {
      setBundle(await api.mirrorContract(tradeId.trim()));
    } catch (e) {
      setErr(e instanceof Error ? e.message : "Failed to mirror");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="mx-auto max-w-5xl px-4 py-16 sm:px-6 lg:px-8">
      <p className="eyebrow">Phase 4 · optional on-chain audit</p>
      <h1 className="mt-3 font-serif text-headline text-parchment">On-chain proof mirror</h1>
      <p className="mt-4 max-w-2xl text-muted">
        The off-chain PostgreSQL record and signed legal confirmation are authoritative.
        These rails anchor a tamper-evident mirror of each canonical trade&apos;s state and
        canary attestation so anyone can verify it independently — never a second source of truth.
      </p>

      {/* Rail status */}
      <div className="mt-8 grid gap-4 sm:grid-cols-3">
        {status
          ? (["solana", "arbitrum", "hyperliquid"] as const).map((c) => (
              <div key={c} className="card p-5">
                <div className="flex items-center justify-between">
                  <span className={`font-serif text-lg ${CHAIN_META[c].tone}`}>{CHAIN_META[c].label}</span>
                  <ModePill mode={status.modes[c]} />
                </div>
                <div className="tabular mt-2 text-xs text-muted-soft">{status.clusters[c]}</div>
              </div>
            ))
          : (["solana", "arbitrum", "hyperliquid"] as const).map((c) => (
              <div key={c} className="card p-5 opacity-60">
                <span className={`font-serif text-lg ${CHAIN_META[c].tone}`}>{CHAIN_META[c].label}</span>
                <div className="tabular mt-2 text-xs text-muted-soft">gateway offline</div>
              </div>
            ))}
      </div>

      {status && (
        <p className="mt-4 flex items-center gap-2 text-xs text-muted-soft">
          <ShieldCheck className="h-3.5 w-3.5 text-signal" /> {status.authority}
        </p>
      )}

      {/* Mirror control */}
      <div className="card card-raise mt-10 p-6">
        <label className="eyebrow">Canonical trade_id</label>
        <div className="mt-3 flex flex-col gap-3 sm:flex-row">
          <input
            value={tradeId}
            onChange={(e) => setTradeId(e.target.value)}
            className="tabular flex-1 rounded-lg border border-line bg-ink-800 px-4 py-2.5 text-sm text-parchment outline-none focus:border-signal"
            placeholder="contract UUID"
          />
          <button
            onClick={mirror}
            disabled={loading || !tradeId.trim()}
            className="inline-flex items-center justify-center gap-2 rounded-lg bg-signal px-5 py-2.5 text-sm font-semibold text-ink-950 transition-all hover:bg-signal-bright active:scale-[0.98] disabled:opacity-50"
          >
            {loading ? <Loader2 className="h-4 w-4 animate-spin" /> : <Link2 className="h-4 w-4" />}
            Anchor mirror
          </button>
        </div>
        {err && <p className="mt-3 text-sm text-alert">{err}</p>}
      </div>

      {/* Proof bundle */}
      {bundle && (
        <div className="mt-8 space-y-4 animate-rise">
          <div className="flex items-center justify-between">
            <h2 className="font-serif text-title text-parchment">Proof bundle</h2>
            <span className="tabular text-xs text-muted-soft">
              mirrored {new Date(bundle.mirrored_at).toLocaleTimeString()}
            </span>
          </div>

          {bundle.rails.map((rail) => (
            <div key={rail.chain} className="card p-6">
              <div className="flex items-center justify-between">
                <span className={`font-serif text-lg ${CHAIN_META[rail.chain]?.tone ?? "text-parchment"}`}>
                  {CHAIN_META[rail.chain]?.label ?? rail.chain}
                </span>
                <ModePill mode={rail.mode} />
              </div>

              {rail.error ? (
                <p className="mt-3 text-sm text-alert">{rail.error}</p>
              ) : (
                <dl className="mt-4 grid gap-3 sm:grid-cols-2">
                  {rail.state_tx && (
                    <div>
                      <dt className="eyebrow">State anchor</dt>
                      <dd className="mt-1"><CopyHash value={rail.state_tx} /></dd>
                    </div>
                  )}
                  {rail.attestation_tx && (
                    <div>
                      <dt className="eyebrow">Canary attestation</dt>
                      <dd className="mt-1"><CopyHash value={rail.attestation_tx} /></dd>
                    </div>
                  )}
                  {rail.reference && (
                    <div>
                      <dt className="eyebrow">{rail.chain === "solana" ? "PDA" : "Contract"}</dt>
                      <dd className="mt-1"><CopyHash value={rail.reference} /></dd>
                    </div>
                  )}
                  {rail.explorer_url && (
                    <div className="sm:col-span-2">
                      <a
                        href={rail.explorer_url}
                        target="_blank"
                        rel="noreferrer"
                        className="inline-flex items-center gap-1.5 text-sm font-medium text-signal hover:text-signal-bright"
                      >
                        Verify on block explorer <ArrowUpRight className="h-4 w-4" />
                      </a>
                    </div>
                  )}
                </dl>
              )}
            </div>
          ))}

          <p className="rounded-card border border-line bg-ink-900/60 p-4 text-xs text-muted-soft">{bundle.note}</p>
        </div>
      )}
    </div>
  );
}
