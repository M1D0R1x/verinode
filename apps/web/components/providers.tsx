"use client";

import { AuthProvider } from "@/lib/use-auth";

export function Providers({ children }: { children: React.ReactNode }) {
  return <AuthProvider>{children}</AuthProvider>;
}
