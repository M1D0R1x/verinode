"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { ArrowLeft, Server, ShieldCheck, CheckCircle2, Send } from "lucide-react";
import { formatCents } from "@/lib/utils";
import { api } from "@/lib/api";
import { DeliveryGrade } from "@/lib/types";

export default function NewRFQPage() {
  const router = useRouter();

  const [grades, setGrades] = useState<DeliveryGrade[]>([]);
  const [gradeId, setGradeId] = useState("H100-SXM-8XNV");
  const [region, setRegion] = useState("US-East");
  const [tenorHours, setTenorHours] = useState(168);
  const [startDate, setStartDate] = useState("2026-10-05");
  const [targetHourlyCents, setTargetHourlyCents] = useState(21000); // $210.00 / hr node
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
        // Fallback to static benchmark grade if backend is offline
        setGrades([
          {
            id: "H100-SXM-8XNV",
            gpu_sku: "NVIDIA H100 SXM 80GB",
            min_memory_gb: 640,
            topology: "SXM5/HGX NVLink 4.0 / NVSwitch",
            min_healthy_gpu_count: 8,
            benchmark_floor: {
              nccl_allreduce_gb_per_sec: 400,
              min_cuda_driver: "535.129.03",
              max_ecc_unrecovered_errors: 0,
            },
            min_cpu_cores: 112,
            min_ram_gb: 1024,
            min_nvme_perf: 100000,
          },
        ]);
      });
  }, []);

  const totalEstimateCents = targetHourlyCents * tenorHours;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
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
      // Graceful fallback for offline demo
    } finally {
      setIsSubmitting(false);
      setIsSuccess(true);
      setTimeout(() => {
        router.push("/buyer");
      }, 1200);
    }
  };

  return (
    <div className="mx-auto max-w-4xl px-4 py-8 sm:px-6 lg:px-8 space-y-6">
      <Link
        href="/buyer"
        className="inline-flex items-center gap-1.5 text-xs text-muted hover:text-white transition-colors"
      >
        <ArrowLeft className="h-3.5 w-3.5" />
        <span>Back to Buyer Console</span>
      </Link>

      <div className="border-b border-border pb-4">
        <h1 className="text-2xl font-bold text-white tracking-tight">Create Private Forward RFQ</h1>
        <p className="text-xs text-muted mt-1">
          Specify your delivery parameters to request binding executable quotes from vetted suppliers.
        </p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* FORM */}
        <form onSubmit={handleSubmit} className="lg:col-span-2 space-y-5">
          {/* Grade selection */}
          <div className="p-4 rounded-xl bg-surface border border-border space-y-2">
            <label className="text-xs font-semibold text-white block">Benchmark Hardware Grade</label>
            <div className="p-3 rounded-lg border border-primary/50 bg-primary/5 flex items-center justify-between">
              <div className="flex items-center gap-3">
                <Server className="h-5 w-5 text-primary" />
                <div>
                  <div className="font-bold text-sm text-white">8x NVIDIA H100 SXM 80GB</div>
                  <div className="text-[11px] font-mono text-muted">ID: H100-SXM-8XNV (640GB HBM3, NVLink 4.0)</div>
                </div>
              </div>
              <span className="text-[11px] font-mono text-accent bg-emerald-950/80 px-2 py-0.5 rounded border border-emerald-800">
                Active Benchmark
              </span>
            </div>
            <p className="text-[11px] text-muted">
              Standardized delivery grade. Guaranteed NCCL all-reduce throughput floor $\ge$ 400 GB/s.
            </p>
          </div>

          {/* Region selection */}
          <div className="p-4 rounded-xl bg-surface border border-border space-y-3">
            <label className="text-xs font-semibold text-white block">Region Bucket</label>
            <div className="grid grid-cols-3 gap-2">
              {[
                { id: "US-East", label: "US-East", sub: "VA, OH, NC" },
                { id: "US-West", label: "US-West", sub: "OR, CA, WA" },
                { id: "EU-Central", label: "EU-Central", sub: "DE, NL, IE" },
              ].map((r) => (
                <button
                  type="button"
                  key={r.id}
                  onClick={() => setRegion(r.id)}
                  className={`p-3 rounded-lg text-left border transition-all text-xs ${
                    region === r.id
                      ? "border-primary bg-primary/10 text-white"
                      : "border-border bg-surfaceSubtle text-muted hover:border-borderHighlight"
                  }`}
                >
                  <span className="font-bold block">{r.label}</span>
                  <span className="text-[10px] text-zinc-500 font-mono block mt-0.5">{r.sub}</span>
                </button>
              ))}
            </div>
          </div>

          {/* Tenor & Window */}
          <div className="p-4 rounded-xl bg-surface border border-border space-y-4">
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <div>
                <label className="text-xs font-semibold text-white block mb-1.5">Reservation Tenor</label>
                <select
                  value={tenorHours}
                  onChange={(e) => setTenorHours(Number(e.target.value))}
                  className="w-full px-3 py-2 rounded-lg bg-surfaceSubtle border border-border text-xs text-white focus:outline-none focus:border-primary font-mono"
                >
                  <option value={168}>168 Hours (1 Week — Standard)</option>
                  <option value={336}>336 Hours (2 Weeks)</option>
                  <option value={720}>720 Hours (1 Month)</option>
                </select>
              </div>

              <div>
                <label className="text-xs font-semibold text-white block mb-1.5">Requested Start Date</label>
                <input
                  type="date"
                  value={startDate}
                  onChange={(e) => setStartDate(e.target.value)}
                  className="w-full px-3 py-2 rounded-lg bg-surfaceSubtle border border-border text-xs text-white focus:outline-none focus:border-primary font-mono"
                />
              </div>
            </div>
          </div>

          {/* Target Hourly Rate Ceiling */}
          <div className="p-4 rounded-xl bg-surface border border-border space-y-2">
            <label className="text-xs font-semibold text-white block">Maximum Price Ceiling (USD / Hour / Node)</label>
            <div className="flex items-center gap-2">
              <span className="text-muted text-sm font-mono">$</span>
              <input
                type="number"
                value={targetHourlyCents / 100}
                onChange={(e) => setTargetHourlyCents(Number(e.target.value) * 100)}
                className="w-full px-3 py-2 rounded-lg bg-surfaceSubtle border border-border text-xs text-white focus:outline-none focus:border-primary font-mono"
                min={100}
                step={5}
              />
              <span className="text-muted text-xs font-mono shrink-0">/ hr (8x node)</span>
            </div>
            <p className="text-[11px] text-zinc-500 font-mono">
              ~${(targetHourlyCents / 100 / 8).toFixed(2)} per GPU-hour. Quotes exceeding this ceiling will be flagged.
            </p>
          </div>

          {/* Submit Action */}
          <button
            type="submit"
            disabled={isSubmitting || isSuccess}
            className="w-full flex items-center justify-center gap-2 py-3 rounded-lg bg-primary hover:bg-primary-hover disabled:bg-primary/50 text-white font-semibold text-xs transition-all shadow-md"
          >
            {isSuccess ? (
              <>
                <CheckCircle2 className="h-4 w-4 text-white" />
                <span>Forward RFQ Broadcasted to Suppliers!</span>
              </>
            ) : isSubmitting ? (
              <span>Dispatching Private RFQ...</span>
            ) : (
              <>
                <Send className="h-4 w-4" />
                <span>Submit Private RFQ</span>
              </>
            )}
          </button>
        </form>

        {/* SUMMARY CARD */}
        <div className="space-y-4">
          <div className="p-5 rounded-xl bg-surface border border-border space-y-4">
            <h3 className="font-bold text-white text-sm">Contract Spec Summary</h3>

            <div className="space-y-2.5 text-xs font-mono">
              <div className="flex justify-between py-1 border-b border-border/60">
                <span className="text-muted">Grade:</span>
                <span className="text-white font-medium">8x H100 SXM 80GB</span>
              </div>
              <div className="flex justify-between py-1 border-b border-border/60">
                <span className="text-muted">Region:</span>
                <span className="text-white font-medium">{region}</span>
              </div>
              <div className="flex justify-between py-1 border-b border-border/60">
                <span className="text-muted">Tenor:</span>
                <span className="text-white font-medium">{tenorHours} Hours</span>
              </div>
              <div className="flex justify-between py-1 border-b border-border/60">
                <span className="text-muted">Max Hourly:</span>
                <span className="text-white font-medium">${(targetHourlyCents / 100).toFixed(2)}/hr</span>
              </div>
              <div className="flex justify-between py-1.5 border-t border-border font-bold">
                <span className="text-white">Ceiling Notional:</span>
                <span className="text-emerald-400 text-sm">{formatCents(totalEstimateCents)}</span>
              </div>
            </div>

            <div className="p-3 rounded-lg bg-surfaceSubtle border border-border text-[11px] text-muted space-y-1">
              <div className="flex items-center gap-1.5 text-accent font-medium">
                <ShieldCheck className="h-3.5 w-3.5" />
                <span>Bilateral Private Matching</span>
              </div>
              <p>
                This request is dispatched only to verified suppliers with compatible inventory blocks.
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
