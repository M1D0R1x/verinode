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
    <header className="sticky top-0 z-50 w-full border-b border-[#272727] bg-black/90 backdrop-blur-md">
      <div className="mx-auto flex h-16 max-w-7xl items-center justify-between px-4 sm:px-6 lg:px-8">
        {/* Brand */}
        <div className="flex items-center gap-8">
          <Link href="/" className="flex items-center gap-3 group">
            <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-[#181818] border border-[#272727] group-hover:border-zinc-500 transition-colors duration-150 ease-out">
              <Cpu className="h-4 w-4 text-white" />
            </div>
            <div className="flex flex-col">
              <div className="flex items-center gap-2">
                <span className="font-bold text-base tracking-tight text-white">VERINODE</span>
                <span className="inline-flex h-1.5 w-1.5 rounded-full bg-emerald-400" />
              </div>
              <span className="text-xs font-mono tracking-wider text-zinc-500 uppercase">Physical Compute</span>
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
                    "px-3 py-1.5 text-sm font-medium rounded-md transition-colors duration-150 ease-out",
                    isActive
                      ? "text-white bg-[#181818] border border-[#272727]"
                      : "text-zinc-400 hover:text-white hover:bg-[#121212]"
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
          <div className="hidden sm:flex items-center gap-2 px-3 py-1 rounded-full border border-[#272727] bg-[#121212] text-xs font-mono text-zinc-400">
            <ShieldCheck className="h-3.5 w-3.5 text-emerald-400" />
            <span>Physical Delivery Guard</span>
          </div>

          <Link
            href="/buyer/rfqs/new"
            className="inline-flex items-center gap-1.5 px-3.5 py-1.5 rounded-lg bg-white text-zinc-950 hover:bg-zinc-200 text-sm font-semibold shadow-sm transition-all duration-150 ease-out active:scale-[0.98]"
          >
            <span>Create RFQ</span>
            <ArrowUpRight className="h-4 w-4 text-zinc-950" />
          </Link>
        </div>
      </div>
    </header>
  );
}
