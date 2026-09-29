"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { ShieldCheck, ArrowRight } from "lucide-react";
import { cn } from "@/lib/utils";

const navItems = [
  { name: "Buyer", href: "/buyer" },
  { name: "Seller", href: "/seller" },
  { name: "Index", href: "/market-data" },
  { name: "Chain Proofs", href: "/proofs" },
  { name: "Hedge Desk", href: "/hedge" },
  { name: "Admin", href: "/admin" },
];

export function Header() {
  const pathname = usePathname();

  return (
    <header className="sticky top-0 z-50 w-full border-b border-line bg-ink-950/85 backdrop-blur-xl">
      <div className="mx-auto flex h-16 max-w-6xl items-center justify-between px-4 sm:px-6 lg:px-8">
        <div className="flex items-center gap-9">
          <Link href="/" className="group flex items-center gap-2.5">
            <span className="flex h-8 w-8 items-center justify-center rounded-lg bg-signal text-ink-950 shadow-glow">
              <span className="font-serif text-lg font-semibold leading-none">V</span>
            </span>
            <span className="font-serif text-lg tracking-tight text-parchment">Verinode</span>
          </Link>

          <nav className="hidden items-center gap-1 md:flex">
            {navItems.map((item) => {
              const isActive =
                pathname === item.href ||
                (item.href !== "/" && pathname.startsWith(item.href));
              return (
                <Link
                  key={item.name}
                  href={item.href}
                  className={cn(
                    "rounded-md px-3 py-1.5 text-sm font-medium transition-colors",
                    isActive
                      ? "bg-ink-850 text-parchment"
                      : "text-muted hover:bg-ink-900 hover:text-parchment"
                  )}
                >
                  {item.name}
                </Link>
              );
            })}
          </nav>
        </div>

        <div className="flex items-center gap-3">
          <span className="hidden items-center gap-2 rounded-pill border border-line bg-ink-900 px-3 py-1 text-xs text-muted lg:inline-flex">
            <ShieldCheck className="h-3.5 w-3.5 text-signal" />
            Physical forward exclusion
          </span>
          <Link
            href="/buyer/rfqs/new"
            className="group inline-flex items-center gap-1.5 rounded-lg bg-signal px-3.5 py-1.5 text-sm font-semibold text-ink-950 transition-all hover:bg-signal-bright active:scale-[0.98]"
          >
            Create RFQ
            <ArrowRight className="h-4 w-4 transition-transform group-hover:translate-x-0.5" />
          </Link>
        </div>
      </div>
    </header>
  );
}
