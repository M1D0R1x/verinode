import Link from "next/link";
import { ShieldCheck } from "lucide-react";

const columns = [
  {
    title: "Marketplace",
    links: [
      { label: "Grade ontology", href: "/#grades" },
      { label: "Submit private RFQ", href: "/buyer/rfqs/new" },
      { label: "List cluster capacity", href: "/seller" },
      { label: "Benchmark index", href: "/market-data" },
    ],
  },
  {
    title: "Verification",
    links: [
      { label: "On-chain proofs", href: "/proofs" },
      { label: "Hedge desk", href: "/hedge" },
      { label: "Admin console", href: "/admin" },
      { label: "Surveillance", href: "/admin/surveillance" },
    ],
  },
];

const specs = [
  "8× H100 SXM 80GB · 168h block",
  "NCCL all-reduce floor 400 GB/s",
  "ed25519 host attestations",
  "Zero customer-data ingestion",
];

export function Footer() {
  return (
    <footer className="border-t border-line bg-ink-900/60">
      <div className="mx-auto max-w-6xl px-4 py-14 sm:px-6 lg:px-8">
        <div className="grid grid-cols-1 gap-10 md:grid-cols-4">
          <div className="space-y-4">
            <div className="flex items-center gap-2.5">
              <span className="flex h-8 w-8 items-center justify-center rounded-lg bg-signal text-ink-950">
                <span className="font-serif text-lg font-semibold leading-none">V</span>
              </span>
              <span className="font-serif text-lg text-parchment">Verinode</span>
            </div>
            <p className="text-sm leading-relaxed text-muted">
              An institutional desk for physically delivered enterprise GPU-capacity
              reservations and cryptographic telemetry verification.
            </p>
            <div className="inline-flex items-center gap-2 text-xs text-signal">
              <ShieldCheck className="h-3.5 w-3.5" /> Physical forward exclusion
            </div>
          </div>

          {columns.map((col) => (
            <div key={col.title}>
              <h4 className="eyebrow mb-4">{col.title}</h4>
              <ul className="space-y-2.5 text-sm">
                {col.links.map((l) => (
                  <li key={l.label}>
                    <Link href={l.href} className="text-muted transition-colors hover:text-parchment">
                      {l.label}
                    </Link>
                  </li>
                ))}
              </ul>
            </div>
          ))}

          <div>
            <h4 className="eyebrow mb-4">Delivery grade</h4>
            <ul className="space-y-2.5 text-sm text-muted">
              {specs.map((s) => (
                <li key={s} className="tabular text-xs">{s}</li>
              ))}
            </ul>
          </div>
        </div>

        <div className="rule mt-12" />
        <div className="mt-6 flex flex-col gap-2 text-xs text-muted-soft sm:flex-row sm:items-center sm:justify-between">
          <span>© {new Date().getFullYear()} Verinode Inc. Off-chain PostgreSQL is the legal authority.</span>
          <span className="tabular">Bank/invoice settlement primary · chain rails optional & read-only</span>
        </div>
      </div>
    </footer>
  );
}
