"use client";

import { RequireAuth } from "@/components/require-auth";

export default function SellerLayout({ children }: { children: React.ReactNode }) {
  return <RequireAuth>{children}</RequireAuth>;
}
