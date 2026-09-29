"use client";

import { useEffect } from "react";
import { useRouter, usePathname } from "next/navigation";
import { Loader2, ShieldAlert } from "lucide-react";
import { useAuth } from "@/lib/use-auth";

/**
 * RequireAuth gates a client route. `staff` requires platform staff (super_admin/admin);
 * otherwise any authenticated principal passes. Unauthenticated users are redirected to
 * /login with a return path. It waits for auth hydration to finish before deciding, so a
 * freshly-logged-in user is never bounced by the initial loading state.
 */
export function RequireAuth({
  children,
  staff = false,
}: {
  children: React.ReactNode;
  staff?: boolean;
}) {
  const { user, loading } = useAuth();
  const router = useRouter();
  const pathname = usePathname();

  useEffect(() => {
    if (loading) return;
    if (!user) {
      router.replace(`/login?next=${encodeURIComponent(pathname)}`);
    }
  }, [user, loading, router, pathname]);

  if (loading) {
    return (
      <div className="flex min-h-[60vh] items-center justify-center">
        <Loader2 className="h-6 w-6 animate-spin text-signal" />
      </div>
    );
  }

  if (!user) {
    return (
      <div className="flex min-h-[60vh] items-center justify-center">
        <Loader2 className="h-6 w-6 animate-spin text-signal" />
      </div>
    );
  }

  if (staff && !user.is_platform_staff) {
    return (
      <div className="mx-auto flex min-h-[60vh] max-w-md flex-col items-center justify-center px-6 text-center">
        <ShieldAlert className="h-8 w-8 text-alert" />
        <h2 className="mt-4 font-serif text-title text-parchment">Platform console</h2>
        <p className="mt-2 text-sm text-muted">
          This area is restricted to Verinode platform staff (super_admin / admin). Your
          company desk is under Buyer and Seller. Surveillance, participant approval and
          cross-company audit are platform trust-domain functions.
        </p>
      </div>
    );
  }

  return <>{children}</>;
}
