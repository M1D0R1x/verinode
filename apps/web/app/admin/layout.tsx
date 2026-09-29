"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { LayoutDashboard, Users, ShieldAlert, Radar, ScrollText } from "lucide-react";
import { RequireAuth } from "@/components/require-auth";
import { cn } from "@/lib/utils";

const adminNav = [
  { name: "Overview", href: "/admin", icon: LayoutDashboard },
  { name: "Participants", href: "/admin/participants", icon: Users },
  { name: "Claims", href: "/admin/claims", icon: ShieldAlert },
  { name: "Surveillance", href: "/admin/surveillance", icon: Radar },
  { name: "Audit", href: "/admin/audit", icon: ScrollText },
];

export default function AdminLayout({ children }: { children: React.ReactNode }) {
  const pathname = usePathname();

  return (
    <RequireAuth staff>
      <div className="mx-auto max-w-6xl px-4 py-10 sm:px-6 lg:px-8">
        <div className="flex items-center justify-between border-b border-line pb-6">
          <div>
            <p className="eyebrow">Platform · compliance & operations</p>
            <h1 className="mt-2 font-serif text-headline text-parchment">Clearing desk</h1>
          </div>
          <span className="inline-flex items-center gap-2 rounded-pill border border-verify/40 bg-verify-wash px-3 py-1.5 text-xs text-verify">
            <span className="h-1.5 w-1.5 rounded-full bg-verify animate-pulseDot" /> Core engine online
          </span>
        </div>

        <nav className="mt-6 flex flex-wrap gap-1.5">
          {adminNav.map((item) => {
            const active = pathname === item.href;
            return (
              <Link
                key={item.href}
                href={item.href}
                className={cn(
                  "inline-flex items-center gap-2 rounded-lg px-3.5 py-2 text-sm font-medium transition-colors",
                  active
                    ? "bg-signal text-ink-950"
                    : "border border-line bg-ink-900 text-muted hover:bg-ink-850 hover:text-parchment"
                )}
              >
                <item.icon className="h-4 w-4" />
                {item.name}
              </Link>
            );
          })}
        </nav>

        <div className="mt-8">{children}</div>
      </div>
    </RequireAuth>
  );
}
