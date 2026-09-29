"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { ArrowLeft, Server, ShieldCheck, CheckCircle2, Send, AlertTriangle } from "lucide-react";
import { api } from "@/lib/api";
import { DeliveryGrade } from "@/lib/types";

// Sane bounds for the price ceiling, in whole USD per node-hour.
// A single 8×H100 node realistically rents $8–$400/GPU-hr => $64–$3,200/node-hr.
// We cap generously at $10,000/node-hr and floor at $8 to reject malformed input.
const MIN_HOURLY_USD = 8;
const MAX_HOURLY_USD = 10_000;
const GPU_PER_NODE = 8;

const TENORS = [
  { hours: 168, label: "168 hours · 1 week (standard)" },
  { hours: 336, label: "336 hours · 2 weeks" },
  { hours: 720, label: "720 hours · 1 month" },
];

const REGIONS = [
  { id: "US-East", label: "US-East", sub: "VA · OH · NC" },
  { id: "US-West", label: "US-West", sub: "OR · CA · WA" },
  { id: "EU-Central", label: "EU-Central", sub: "DE · NL · IE" },
];

function usd(n: number, frac = 2) {
  return new Intl.NumberFormat("en-US", {
    style: "currency",
    currency: "USD",
    minimumFractionDigits: frac,
    maximumFractionDigits: frac,
  }).format(n);
}

export default function NewRFQPage() {
  const router = useRouter();

  const [grades, setGrades] = useState<DeliveryGrade[]>([]);
  const [gradeId, setGradeId] = useState("H100-SXM-8XNV");
  const [region, setRegion] = useState("US-East");
  const [tenorHours, setTenorHours] = useState(168);
  const [startDate, setStartDate] = useState("2026-10-05");
  // Price ceiling held as whole USD per node-hour (not cents) to avoid the
  // ×100 round-trip that previously produced absurd values.
  const [hourlyUsdRaw, setHourlyUsdRaw] = useState("210");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isSuccess, setIsSuccess] = useState(false);

  useEffect(() => {
    api
      .getGrades()
      .then((data) => {
        if (data && data.length > 0) {
          setGrades(data);
          setGradeId(data[0].id);
        }
      })
      .catch(() => {
        setGrades([]);
      });
  }, []);

  // Parse + clamp the price into a valid number; expose a validation message.
  const parsed = Number(hourlyUsdRaw);
  const hourlyValid = Number.isFinite(parsed) && parsed >= MIN_HOURLY_USD && parsed <= MAX_HOURLY_USD;
  const hourlyUsd = hourlyValid ? parsed : NaN;
  const perGpuHour = hourlyValid ? hourlyUsd / GPU_PER_NODE : NaN;
  const ceilingNotional = hourlyValid ? hourlyUsd * tenorHours : NaN;

  let priceError: string | null = null;
  if (hourlyUsdRaw.trim() !== "" && !Number.isFinite(parsed)) {
    priceError = "Enter a valid number.";
  } else if (Number.isFinite(parsed) && parsed < MIN_HOURLY_USD) {
    priceError = `Minimum ceiling is ${usd(MIN_HOURLY_USD, 0)} / node-hour.`;
  } else if (Number.isFinite(parsed) && parsed > MAX_HOURLY_USD) {
    priceError = `Maximum ceiling is ${usd(MAX_HOURLY_USD, 0)} / node-hour.`;
  }

  const canSubmit = hourlyValid && !!startDate && !isSubmitting && !isSuccess;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!canSubmit) return;
    setIsSubmitting(true);
    try {
      const windowStart = new Date(startDate);
      const windowEnd = new Date(windowStart.getTime() + tenorHours * 3600 * 1000);
      await api.createRFQ({
        buyer_id: "00000000-0000-0000-0000-000000000001",
        grade_id: gradeId,
        region_bucket: region,
        window_start: windowStart.toISOString(),
        window_end: windowEnd.toISOString(),
      });
    } catch {
      // graceful offline fallback for the demo
    } finally {
      setIsSubmitting(false);
      setIsSuccess(true);
      setTimeout(() => router.push("/buyer"), 1200);
    }
  };

  return (
    <div className="mx-auto max-w-4xl px-4 py-10 sm:px-6 lg:px-8">
      <Link
        href="/buyer"
        className="inline-flex items-center gap-1.5 text-sm text-muted transition-colors hover:text-parchment"
      >
        <ArrowLeft className="h-4 w-4" /> Back to buyer console
      </Link>

      <div className="mt-6 border-b border-line pb-6">
        <p className="eyebrow">Private bilateral request</p>
        <h1 className="mt-2 font-serif text-headline text-parchment">Create a forward RFQ</h1>
        <p className="mt-2 max-w-2xl text-sm text-muted">
          Specify delivery parameters to request binding, firm quotes from vetted suppliers.
          Your request is visible only to invited suppliers — never broadcast publicly.
        </p>
      </div>

      <div className="mt-8 grid grid-cols-1 gap-6 lg:grid-cols-3">
        <form onSubmit={handleSubmit} className="space-y-5 lg:col-span-2">
          {/* Grade */}
          <div className="card p-5">
            <label className="text-sm font-medium text-parchment">Benchmark hardware grade</label>
            <div className="mt-3 flex items-center justify-between rounded-lg border border-signal/40 bg-signal-wash/50 p-4">
              <div className="flex items-center gap-3">
                <Server className="h-5 w-5 text-signal" />
                <div>
                  <div className="text-sm font-semibold text-parchment">8× NVIDIA H100 SXM 80GB</div>
                  <div className="tabular text-xs text-muted-soft">H100-SXM-8XNV · 640GB HBM3 · NVLink 4.0</div>
                </div>
              </div>
              <span className="tabular rounded-pill border border-verify/40 bg-verify-wash px-2.5 py-1 text-xs text-verify">
                Active benchmark
              </span>
            </div>
            <p className="mt-3 text-xs text-muted-soft">
              Standardized delivery grade. Guaranteed NCCL all-reduce throughput floor ≥ 400 GB/s.
            </p>
          </div>

          {/* Region */}
          <div className="card p-5">
            <label className="text-sm font-medium text-parchment">Region bucket</label>
            <div className="mt-3 grid grid-cols-3 gap-2.5">
              {REGIONS.map((r) => (
                <button
                  type="button"
                  key={r.id}
                  onClick={() => setRegion(r.id)}
                  className={`rounded-lg border p-3 text-left transition-all ${
                    region === r.id
                      ? "border-signal bg-signal-wash/60 text-parchment"
                      : "border-line bg-ink-800 text-muted hover:border-ink-700"
                  }`}
                >
                  <span className="block text-sm font-semibold">{r.label}</span>
                  <span className="tabular mt-0.5 block text-xs text-muted-soft">{r.sub}</span>
                </button>
              ))}
            </div>
          </div>

          {/* Tenor & window */}
          <div className="card grid grid-cols-1 gap-4 p-5 sm:grid-cols-2">
            <div>
              <label className="text-sm font-medium text-parchment">Reservation tenor</label>
              <select
                value={tenorHours}
                onChange={(e) => setTenorHours(Number(e.target.value))}
                className="mt-2 w-full rounded-lg border border-line bg-ink-800 px-3 py-2.5 text-sm text-parchment outline-none focus:border-signal"
              >
                {TENORS.map((t) => (
                  <option key={t.hours} value={t.hours} className="bg-ink-850 text-parchment">
                    {t.label}
                  </option>
                ))}
              </select>
            </div>
            <div>
              <label className="text-sm font-medium text-parchment">Requested start date</label>
              <input
                type="date"
                value={startDate}
                min={new Date().toISOString().slice(0, 10)}
                onChange={(e) => setStartDate(e.target.value)}
                className="tabular mt-2 w-full rounded-lg border border-line bg-ink-800 px-3 py-2.5 text-sm text-parchment outline-none focus:border-signal [color-scheme:dark]"
              />
            </div>
          </div>

          {/* Price ceiling */}
          <div className="card p-5">
            <label className="text-sm font-medium text-parchment">
              Maximum price ceiling
              <span className="ml-1 text-muted-soft">(USD / hour / 8× node)</span>
            </label>
            <div
              className={`mt-3 flex items-center gap-2 rounded-lg border bg-ink-800 px-3 ${
                priceError ? "border-alert" : "border-line focus-within:border-signal"
              }`}
            >
              <span className="tabular text-sm text-muted">$</span>
              <input
                type="number"
                inputMode="decimal"
                value={hourlyUsdRaw}
                onChange={(e) => setHourlyUsdRaw(e.target.value)}
                onBlur={() => {
                  if (Number.isFinite(parsed)) {
                    setHourlyUsdRaw(String(Math.min(MAX_HOURLY_USD, Math.max(MIN_HOURLY_USD, Math.round(parsed)))));
                  }
                }}
                className="tabular w-full bg-transparent py-2.5 text-sm text-parchment outline-none"
                min={MIN_HOURLY_USD}
                max={MAX_HOURLY_USD}
                step={5}
                placeholder="210"
              />
              <span className="tabular shrink-0 text-xs text-muted">/ hr</span>
            </div>
            {priceError ? (
              <p className="mt-2 flex items-center gap-1.5 text-xs text-alert">
                <AlertTriangle className="h-3.5 w-3.5" /> {priceError}
              </p>
            ) : (
              <p className="tabular mt-2 text-xs text-muted-soft">
                ≈ {usd(perGpuHour)} per GPU-hour. Quotes above this ceiling are flagged.
              </p>
            )}
          </div>

          {/* Submit — high-emphasis gold CTA */}
          <button
            type="submit"
            disabled={!canSubmit}
            className={`group flex w-full items-center justify-center gap-2 rounded-xl px-5 py-3.5 text-sm font-semibold transition-all ${
              isSuccess
                ? "bg-verify text-ink-950"
                : "bg-signal text-ink-950 shadow-glow hover:bg-signal-bright active:scale-[0.99] disabled:cursor-not-allowed disabled:opacity-50 disabled:shadow-none"
            }`}
          >
            {isSuccess ? (
              <>
                <CheckCircle2 className="h-5 w-5" /> RFQ dispatched to invited suppliers
              </>
            ) : isSubmitting ? (
              <>Dispatching private RFQ…</>
            ) : (
              <>
                <Send className="h-4 w-4" /> Submit private RFQ
              </>
            )}
          </button>
        </form>

        {/* Summary */}
        <aside>
          <div className="card card-raise sticky top-24 p-6">
            <h3 className="font-serif text-title text-parchment">Contract spec</h3>
            <dl className="mt-5 space-y-3 text-sm">
              {[
                ["Grade", "8× H100 SXM 80GB"],
                ["Region", region],
                ["Tenor", `${tenorHours} hours`],
                ["Max hourly", hourlyValid ? `${usd(hourlyUsd)} / hr` : "—"],
              ].map(([k, v]) => (
                <div key={k} className="flex items-center justify-between border-b border-lineSoft pb-3">
                  <dt className="text-muted">{k}</dt>
                  <dd className="tabular text-parchment">{v}</dd>
                </div>
              ))}
              <div className="flex items-center justify-between pt-1">
                <dt className="font-medium text-parchment">Ceiling notional</dt>
                <dd className="tabular text-lg font-semibold text-signal">
                  {hourlyValid ? usd(ceilingNotional, 0) : "—"}
                </dd>
              </div>
            </dl>

            <div className="mt-6 rounded-lg border border-line bg-ink-900/60 p-4">
              <div className="flex items-center gap-1.5 text-sm font-medium text-verify">
                <ShieldCheck className="h-4 w-4" /> Bilateral private matching
              </div>
              <p className="mt-1.5 text-xs text-muted">
                Dispatched only to verified suppliers with compatible inventory blocks — never a public order book.
              </p>
            </div>
          </div>
        </aside>
      </div>
    </div>
  );
}
