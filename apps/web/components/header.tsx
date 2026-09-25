"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { ShieldCheck, Cpu, ArrowUpRight } from "lucide-react";
import { cn } from "@/lib/utils";

export function Header() {
  const pathname = usePathname();

  const navItems = [
    { name: "Grades & Specs", href: "/#grades" },
    { name: "Buyer Portal", href: "/buyer" },
    { name: "Seller Portal", href: "/seller" },
    { name: "Admin Console", href: "/admin" },
    { name: "Index & Data", href: "/#index" },
  ];

  return (
    <header className="sticky top-0 z-50 w-full border-b border-border/80 bg-background/80 backdrop-blur-md">
      <div className="mx-auto flex h-16 max-w-7xl items-center justify-between px-4 sm:px-6 lg:px-8">
        {/* Brand */}
        <div className="flex items-center gap-6">
          <Link href="/" className="flex items-center gap-2.5 group">
            <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-surface border border-border group-hover:border-primary/50 transition-colors">
              <Cpu className="h-5 w-5 text-primary" />
            </div>
            <div className="flex flex-col">
              <div className="flex items-center gap-1.5">
                <span className="font-semibold text-base tracking-tight text-white">Verinode</span>
                <span className="inline-flex h-2 w-2 rounded-full bg-accent animate-pulse" />
              </div>
              <span className="text-[10px] font-mono tracking-wider text-muted uppercase">Verified Compute</span>
            </div>
          </Link>

          {/* Navigation Links */}
          <nav className="hidden md:flex items-center gap-1">
            {navItems.map((item) => {
              const isActive = pathname === item.href || (item.href !== "/" && pathname.startsWith(item.href));
              return (
                <Link
                  key={item.name}
                  href={item.href}
                  className={cn(
                    "px-3 py-1.5 text-sm font-medium rounded-md transition-colors",
                    isActive
                      ? "text-white bg-surface border border-border"
                      : "text-muted hover:text-white hover:bg-surface/50"
                  )}
                >
                  {item.name}
                </Link>
              );
            })}
          </nav>
        </div>

        {/* Action CTAs */}
        <div className="flex items-center gap-3">
          <div className="hidden sm:flex items-center gap-2 px-3 py-1 rounded-full border border-border/60 bg-surface/50 text-xs font-mono text-muted">
            <ShieldCheck className="h-3.5 w-3.5 text-accent" />
            <span>Physical Delivery Guard</span>
          </div>

          <Link
            href="/buyer/rfqs/new"
            className="inline-flex items-center gap-1.5 px-4 py-2 rounded-lg bg-primary hover:bg-primary-hover text-white text-xs sm:text-sm font-medium shadow-sm transition-all"
          >
            <span>Create RFQ</span>
            <ArrowUpRight className="h-4 w-4" />
          </Link>
        </div>
      </div>
    </header>
  );
}
