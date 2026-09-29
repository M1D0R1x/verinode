"use client";

import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { ShieldCheck, ArrowRight, LogOut, ChevronDown } from "lucide-react";
import { useState } from "react";
import { cn } from "@/lib/utils";
import { useAuth } from "@/lib/use-auth";

const baseNav = [
  { name: "Buyer", href: "/buyer" },
  { name: "Seller", href: "/seller" },
  { name: "Index", href: "/market-data" },
  { name: "Chain Proofs", href: "/proofs" },
  { name: "Hedge Desk", href: "/hedge" },
];

export function Header() {
  const pathname = usePathname();
  const router = useRouter();
  const { user, logout } = useAuth();
  const [menu, setMenu] = useState(false);

  // Admin is only shown to platform staff.
  const nav = user?.is_platform_staff ? [...baseNav, { name: "Admin", href: "/admin" }] : baseNav;

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
            {nav.map((item) => {
              const isActive =
                pathname === item.href || (item.href !== "/" && pathname.startsWith(item.href));
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
          {user ? (
            <div className="relative">
              <button
                onClick={() => setMenu((v) => !v)}
                onBlur={() => setTimeout(() => setMenu(false), 150)}
                className="flex items-center gap-2 rounded-lg border border-line bg-ink-900 px-3 py-1.5 text-sm text-parchment transition-colors hover:bg-ink-850"
              >
                <span className="flex h-5 w-5 items-center justify-center rounded-full bg-signal/20 text-xs font-semibold text-signal">
                  {(user.company_name || user.email)[0]?.toUpperCase()}
                </span>
                <span className="hidden max-w-[10rem] truncate sm:inline">
                  {user.company_name || user.email}
                </span>
                <ChevronDown className="h-3.5 w-3.5 text-muted" />
              </button>
              {menu && (
                <div className="absolute right-0 mt-2 w-56 rounded-card border border-line bg-ink-850 p-2 shadow-raise">
                  <div className="px-2 py-2">
                    <div className="truncate text-sm text-parchment">{user.email}</div>
                    <div className="mt-1 inline-flex rounded-pill border border-line px-2 py-0.5 text-[10px] text-signal">
                      {user.role}
                    </div>
                  </div>
                  <div className="rule my-1" />
                  <button
                    onClick={() => { logout(); router.push("/"); }}
                    className="flex w-full items-center gap-2 rounded-md px-2 py-2 text-sm text-muted transition-colors hover:bg-ink-800 hover:text-parchment"
                  >
                    <LogOut className="h-4 w-4" /> Sign out
                  </button>
                </div>
              )}
            </div>
          ) : (
            <Link
              href="/login"
              className="hidden items-center gap-2 rounded-lg border border-line bg-ink-900 px-3.5 py-1.5 text-sm font-medium text-parchment transition-colors hover:bg-ink-850 sm:inline-flex"
            >
              <ShieldCheck className="h-3.5 w-3.5 text-signal" /> Sign in
            </Link>
          )}
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
