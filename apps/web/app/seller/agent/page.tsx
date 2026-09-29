"use client";

import Link from "next/link";
import { ArrowLeft, Terminal, ShieldCheck } from "lucide-react";
import { RequireAuth } from "@/components/require-auth";

const steps = [
  {
    n: "1",
    title: "Install the telemetry agent",
    body: "Runs as a systemd daemon on Linux (Ubuntu 22.04+ / Rocky 9, NVIDIA driver ≥ 535.129.03).",
    code: "curl -fsSL https://get.verinode.io/agent/install.sh | sudo bash",
  },
  {
    n: "2",
    title: "Generate the host ed25519 signing key",
    body: "Each node holds a unique keypair in a hardware-backed enclave or protected key directory. Its public key is bound to your supplier identity at registration.",
    code: "sudo verinode-agent keygen --out /etc/verinode/agent.key",
  },
  {
    n: "3",
    title: "Run the synthetic canary benchmark",
    body: "A standardized NCCL all-reduce proves the NVLink interconnect clears the grade floor (≥ 400.0 GB/s) before the block can go live.",
    code: "sudo verinode-agent canary --grade H100-SXM-8XNV",
    output: [
      { ok: true, text: "NCCL all-reduce: 428.4 GB/s  (floor 400.0 GB/s) PASS" },
      { ok: true, text: "NVLink NVSwitch mesh: 900 GB/s PASS" },
      { ok: true, text: "ECC unrecovered errors: 0 PASS" },
    ],
  },
  {
    n: "4",
    title: "Enable the heartbeat service",
    body: "The agent then emits signed hardware attestations on a heartbeat so the platform can verify liveness through the delivery window.",
    code: "sudo systemctl enable --now verinode-agent",
  },
];

export default function AgentSetupPage() {
  return (
    <RequireAuth>
      <div className="mx-auto max-w-3xl px-4 py-10 sm:px-6 lg:px-8">
        <Link href="/seller" className="inline-flex items-center gap-1.5 text-sm text-muted transition-colors hover:text-parchment">
          <ArrowLeft className="h-4 w-4" /> Back to supplier console
        </Link>

        <div className="mt-6 border-b border-line pb-6">
          <div className="flex items-center gap-2">
            <Terminal className="h-5 w-5 text-signal" />
            <p className="eyebrow">Supplier onboarding</p>
          </div>
          <h1 className="mt-2 font-serif text-headline text-parchment">Host telemetry agent</h1>
          <p className="mt-2 max-w-2xl text-sm text-muted">
            The agent is what makes a reservation <em>verifiable</em>: it runs on your delivery node
            and produces cryptographically signed hardware attestations and canary proofs that the
            platform checks against the contracted grade before delivery is certified.
          </p>
        </div>

        <div className="mt-6 flex items-start gap-3 rounded-card border border-verify/30 bg-verify-wash/50 p-4">
          <ShieldCheck className="mt-0.5 h-5 w-5 shrink-0 text-verify" />
          <div>
            <div className="text-sm font-medium text-parchment">Invariant 4 — zero workload ingestion</div>
            <p className="mt-1 text-xs leading-relaxed text-muted">
              The agent samples only host-level NVML/DCGM metrics (PCI IDs, thermals, NVLink mesh
              health) and synthetic NCCL benchmarks. It never inspects filesystems, memory, model
              weights, or tenant workloads.
            </p>
          </div>
        </div>

        <div className="mt-8 space-y-4">
          {steps.map((s) => (
            <div key={s.n} className="card p-6">
              <div className="flex items-center gap-3">
                <span className="tabular flex h-7 w-7 items-center justify-center rounded-full bg-signal/15 text-sm font-semibold text-signal">
                  {s.n}
                </span>
                <h3 className="font-serif text-title text-parchment">{s.title}</h3>
              </div>
              <p className="mt-2 text-sm text-muted">{s.body}</p>
              <div className="tabular mt-3 rounded-lg border border-line bg-ink-950 p-3.5 text-xs text-parchment">
                <code>{s.code}</code>
                {s.output && (
                  <div className="mt-3 space-y-1">
                    {s.output.map((o) => (
                      <div key={o.text} className="text-verify">✓ {o.text}</div>
                    ))}
                  </div>
                )}
              </div>
            </div>
          ))}
        </div>
      </div>
    </RequireAuth>
  );
}
