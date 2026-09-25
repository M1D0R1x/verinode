import Link from "next/link";
import { Cpu, ShieldCheck } from "lucide-react";

export function Footer() {
  return (
    <footer className="border-t border-[#272727] bg-[#121212] text-xs text-zinc-400">
      <div className="mx-auto max-w-7xl px-4 py-12 sm:px-6 lg:px-8">
        <div className="grid grid-cols-1 md:grid-cols-4 gap-8">
          <div className="space-y-3">
            <div className="flex items-center gap-2">
              <Cpu className="h-4 w-4 text-white" />
              <span className="font-bold text-white text-sm">VERINODE</span>
            </div>
            <p className="text-zinc-400 leading-relaxed text-xs">
              Institutional marketplace for physically delivered enterprise GPU reservations and cryptographic telemetry verification.
            </p>
            <div className="flex items-center gap-2 text-xs text-emerald-400 font-mono">
              <ShieldCheck className="h-3.5 w-3.5" />
              <span>Physical Forward Exclusion</span>
            </div>
          </div>

          <div>
            <h4 className="font-semibold text-white text-xs uppercase tracking-wider mb-3 font-mono">Marketplace</h4>
            <ul className="space-y-2">
              <li><Link href="/#grades" className="hover:text-white transition-colors duration-150">Grade Ontology</Link></li>
              <li><Link href="/buyer/rfqs/new" className="hover:text-white transition-colors duration-150">Submit Private RFQ</Link></li>
              <li><Link href="/seller" className="hover:text-white transition-colors duration-150">List Cluster Capacity</Link></li>
              <li><Link href="/#index" className="hover:text-white transition-colors duration-150">Market Data Index</Link></li>
            </ul>
          </div>

          <div>
            <h4 className="font-semibold text-white text-xs uppercase tracking-wider mb-3 font-mono">Specifications</h4>
            <ul className="space-y-2">
              <li><span className="text-zinc-300">8x H100 SXM 80GB (168h block)</span></li>
              <li><span className="text-zinc-300">NCCL AllReduce floor 400 GB/s</span></li>
              <li><span className="text-zinc-300">Host ed25519 attestations</span></li>
              <li><span className="text-zinc-300">Zero customer data ingestion</span></li>
            </ul>
          </div>

          <div>
            <h4 className="font-semibold text-white text-xs uppercase tracking-wider mb-3 font-mono">Institutional Trust</h4>
            <p className="text-xs text-zinc-400 leading-relaxed">
              Standardized physical capacity delivery. Off-chain PostgreSQL legal authority reconciled to double-entry escrow subledger.
            </p>
            <div className="mt-4 font-mono text-xs text-zinc-500">
              © {new Date().getFullYear()} Verinode Inc. All rights reserved.
            </div>
          </div>
        </div>
      </div>
    </footer>
  );
}
