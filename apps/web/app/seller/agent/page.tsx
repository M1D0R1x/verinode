import Link from "next/link";
import { ArrowLeft, Terminal, ShieldCheck, CheckCircle2, Copy } from "lucide-react";

export default function AgentSetupPage() {
  return (
    <div className="mx-auto max-w-4xl px-4 py-8 sm:px-6 lg:px-8 space-y-6">
      <Link
        href="/seller"
        className="inline-flex items-center gap-1.5 text-xs text-muted hover:text-white transition-colors"
      >
        <ArrowLeft className="h-3.5 w-3.5" />
        <span>Back to Supplier Console</span>
      </Link>

      <div className="border-b border-border pb-4">
        <div className="flex items-center gap-2">
          <Terminal className="h-5 w-5 text-accent" />
          <h1 className="text-2xl font-bold text-white tracking-tight">Host Telemetry Agent Installation</h1>
        </div>
        <p className="text-xs text-muted mt-1">
          Install the Verinode Host Agent on your delivery node to produce cryptographically signed hardware attestations and canary proofs.
        </p>
      </div>

      <div className="p-4 rounded-xl bg-blue-950/30 border border-blue-900/50 flex items-start gap-3 text-xs text-blue-200">
        <ShieldCheck className="h-5 w-5 text-blue-400 shrink-0 mt-0.5" />
        <div>
          <strong className="block font-semibold">Invariant 4: Zero Workload Ingestion Guarantee</strong>
          <p className="text-blue-300/90 mt-0.5 leading-relaxed">
            The telemetry agent strictly samples host-level NVML/DCGM hardware metrics (PCI IDs, thermals, NVLink mesh health) and synthetic NCCL benchmarks. It never inspects file systems, memory buffers, model weights, or tenant workloads.
          </p>
        </div>
      </div>

      <div className="space-y-6">
        {/* Step 1 */}
        <div className="p-6 rounded-2xl bg-surface border border-border space-y-3">
          <div className="flex items-center gap-2 font-bold text-sm text-white">
            <span className="flex h-6 w-6 items-center justify-center rounded-full bg-primary/20 text-primary text-xs font-mono">1</span>
            <span>Download & Install Telemetry Agent Binary</span>
          </div>
          <p className="text-xs text-muted">
            The agent runs as a standalone daemon or systemd service on Linux (Ubuntu 22.04+ / Rocky 9 with NVIDIA Driver $\ge$ 535.129.03).
          </p>
          <div className="p-3.5 rounded-lg bg-background border border-border font-mono text-xs text-zinc-300 relative">
            <code>curl -fsSL https://get.verinode.io/agent/install.sh | sudo bash</code>
          </div>
        </div>

        {/* Step 2 */}
        <div className="p-6 rounded-2xl bg-surface border border-border space-y-3">
          <div className="flex items-center gap-2 font-bold text-sm text-white">
            <span className="flex h-6 w-6 items-center justify-center rounded-full bg-primary/20 text-primary text-xs font-mono">2</span>
            <span>Generate Host ed25519 Signing Key</span>
          </div>
          <p className="text-xs text-muted">
            Each physical node generates a unique ed25519 keypair stored in a hardware-backed enclave or protected key directory.
          </p>
          <div className="p-3.5 rounded-lg bg-background border border-border font-mono text-xs text-zinc-300">
            <code>sudo verinode-agent keygen --out /etc/verinode/agent.key</code>
          </div>
        </div>

        {/* Step 3 */}
        <div className="p-6 rounded-2xl bg-surface border border-border space-y-3">
          <div className="flex items-center gap-2 font-bold text-sm text-white">
            <span className="flex h-6 w-6 items-center justify-center rounded-full bg-primary/20 text-primary text-xs font-mono">3</span>
            <span>Execute Synthetic Benchmark Canary Test</span>
          </div>
          <p className="text-xs text-muted">
            Execute the standardized NCCL all-reduce synthetic benchmark to prove the NVLink interconnect satisfies the benchmark floor ($\ge$ 400.0 GB/s).
          </p>
          <div className="p-3.5 rounded-lg bg-background border border-border font-mono text-xs text-zinc-300 space-y-1">
            <div className="text-zinc-500"># Run NCCL all-reduce canary benchmark across all 8 SXM GPUs</div>
            <div><code>sudo verinode-agent canary --grade H100-SXM-8XNV</code></div>
            <div className="text-emerald-400 mt-2">✓ NCCL All-Reduce Bandwidth: 428.4 GB/s (Floor: 400.0 GB/s PASS)</div>
            <div className="text-emerald-400">✓ NVLink NVSwitch Mesh: 900 GB/s PASS</div>
            <div className="text-emerald-400">✓ ECC Memory Errors: 0 Clean PASS</div>
          </div>
        </div>

        {/* Step 4 */}
        <div className="p-6 rounded-2xl bg-surface border border-border space-y-3">
          <div className="flex items-center gap-2 font-bold text-sm text-white">
            <span className="flex h-6 w-6 items-center justify-center rounded-full bg-primary/20 text-primary text-xs font-mono">4</span>
            <span>Start Background Telemetry Heartbeat</span>
          </div>
          <div className="p-3.5 rounded-lg bg-background border border-border font-mono text-xs text-zinc-300">
            <code>sudo systemctl enable --now verinode-agent</code>
          </div>
        </div>
      </div>
    </div>
  );
}
