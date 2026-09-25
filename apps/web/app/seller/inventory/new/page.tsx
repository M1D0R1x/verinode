"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { ArrowLeft, Server, ShieldCheck, CheckCircle2, Plus } from "lucide-react";

export default function NewInventoryBlockPage() {
  const router = useRouter();

  const [gradeId, setGradeId] = useState("H100-SXM-8XNV");
  const [region, setRegion] = useState("US-East");
  const [facilityRef, setFacilityRef] = useState("EQUINIX-DC21-RACK06");
  const [windowStart, setWindowStart] = useState("2026-10-10");
  const [windowEnd, setWindowEnd] = useState("2026-10-17");
  const [agentKey, setAgentKey] = useState("MCowBQYDK2VwAyEAX5qYF8...k9L0vQw");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isSuccess, setIsSuccess] = useState(false);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setIsSubmitting(true);

    setTimeout(() => {
      setIsSubmitting(false);
      setIsSuccess(true);
      setTimeout(() => {
        router.push("/seller");
      }, 1500);
    }, 600);
  };

  return (
    <div className="mx-auto max-w-3xl px-4 py-8 sm:px-6 lg:px-8 space-y-6">
      <Link
        href="/seller"
        className="inline-flex items-center gap-1.5 text-xs text-muted hover:text-white transition-colors"
      >
        <ArrowLeft className="h-3.5 w-3.5" />
        <span>Back to Supplier Console</span>
      </Link>

      <div className="border-b border-border pb-4">
        <h1 className="text-2xl font-bold text-white tracking-tight">List GPU Cluster Capacity Block</h1>
        <p className="text-xs text-muted mt-1">
          List scheduler-visible unbooked capacity to receive matched private forward RFQs.
        </p>
      </div>

      <form onSubmit={handleSubmit} className="space-y-5">
        {/* Grade */}
        <div className="p-4 rounded-xl bg-surface border border-border space-y-2">
          <label className="text-xs font-semibold text-white block">Delivery Grade Specification</label>
          <div className="p-3 rounded-lg border border-primary/50 bg-primary/5 flex items-center justify-between">
            <div className="flex items-center gap-3">
              <Server className="h-5 w-5 text-primary" />
              <div>
                <div className="font-bold text-sm text-white">8x NVIDIA H100 SXM 80GB</div>
                <div className="text-[11px] font-mono text-muted">H100-SXM-8XNV • 640GB VRAM • NVLink 4.0 (900 GB/s)</div>
              </div>
            </div>
            <span className="text-[11px] font-mono text-emerald-400 bg-emerald-950/80 px-2 py-0.5 rounded border border-emerald-800">
              Benchmark Grade
            </span>
          </div>
        </div>

        {/* Region & Facility */}
        <div className="p-4 rounded-xl bg-surface border border-border space-y-4">
          <div>
            <label className="text-xs font-semibold text-white block mb-1.5">Region Bucket</label>
            <select
              value={region}
              onChange={(e) => setRegion(e.target.value)}
              className="w-full px-3 py-2 rounded-lg bg-surfaceSubtle border border-border text-xs text-white focus:outline-none focus:border-primary font-mono"
            >
              <option value="US-East">US-East (Virginia, Ohio, North Carolina)</option>
              <option value="US-West">US-West (Oregon, California, Washington)</option>
              <option value="EU-Central">EU-Central (Frankfurt, Amsterdam, Dublin)</option>
            </select>
          </div>

          <div>
            <label className="text-xs font-semibold text-white block mb-1.5">
              Datacenter Facility Reference (Attested, Access-Restricted)
            </label>
            <input
              type="text"
              value={facilityRef}
              onChange={(e) => setFacilityRef(e.target.value)}
              className="w-full px-3 py-2 rounded-lg bg-surfaceSubtle border border-border text-xs text-white focus:outline-none focus:border-primary font-mono"
              placeholder="e.g. EQUINIX-DC21-RACK06"
              required
            />
            <span className="text-[11px] text-zinc-500 font-mono mt-1 block">
              Confidential to buyers during RFQ; revealed upon contract execution.
            </span>
          </div>
        </div>

        {/* Available Window */}
        <div className="p-4 rounded-xl bg-surface border border-border space-y-3">
          <label className="text-xs font-semibold text-white block">Continuous Availability Window</label>
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div>
              <span className="text-[11px] text-muted block mb-1 font-mono">WINDOW START</span>
              <input
                type="date"
                value={windowStart}
                onChange={(e) => setWindowStart(e.target.value)}
                className="w-full px-3 py-2 rounded-lg bg-surfaceSubtle border border-border text-xs text-white focus:outline-none focus:border-primary font-mono"
                required
              />
            </div>
            <div>
              <span className="text-[11px] text-muted block mb-1 font-mono">WINDOW END</span>
              <input
                type="date"
                value={windowEnd}
                onChange={(e) => setWindowEnd(e.target.value)}
                className="w-full px-3 py-2 rounded-lg bg-surfaceSubtle border border-border text-xs text-white focus:outline-none focus:border-primary font-mono"
                required
              />
            </div>
          </div>
        </div>

        {/* Bound Host Agent Public Key */}
        <div className="p-4 rounded-xl bg-surface border border-border space-y-2">
          <label className="text-xs font-semibold text-white block">Bound Host Agent ed25519 Public Key</label>
          <input
            type="text"
            value={agentKey}
            onChange={(e) => setAgentKey(e.target.value)}
            className="w-full px-3 py-2 rounded-lg bg-surfaceSubtle border border-border text-xs text-white focus:outline-none focus:border-primary font-mono"
            required
          />
          <span className="text-[11px] text-zinc-500 font-mono block">
            The key used by your on-node daemon to sign telemetry reports and canary benchmarks.
          </span>
        </div>

        {/* Submit */}
        <button
          type="submit"
          disabled={isSubmitting || isSuccess}
          className="w-full flex items-center justify-center gap-2 py-3 rounded-lg bg-primary hover:bg-primary-hover disabled:bg-primary/50 text-white font-semibold text-xs transition-all shadow-md"
        >
          {isSuccess ? (
            <>
              <CheckCircle2 className="h-4 w-4 text-white" />
              <span>Capacity Block Registered & Active!</span>
            </>
          ) : isSubmitting ? (
            <span>Registering Block with Inventory Engine...</span>
          ) : (
            <>
              <Plus className="h-4 w-4" />
              <span>List GPU Capacity Block</span>
            </>
          )}
        </button>
      </form>
    </div>
  );
}
