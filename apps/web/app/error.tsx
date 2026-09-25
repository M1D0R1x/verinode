"use client";

import { useEffect } from "react";
import Link from "next/link";
import { AlertTriangle, RefreshCw } from "lucide-react";

export default function ErrorBoundary({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  useEffect(() => {
    console.error("Next.js App Router Error:", error);
  }, [error]);

  return (
    <div className="min-h-[70vh] flex items-center justify-center px-4">
      <div className="max-w-md w-full p-8 rounded-2xl bg-surface border border-rose-900/40 text-center space-y-4">
        <div className="h-12 w-12 rounded-xl bg-rose-950/40 border border-rose-800 flex items-center justify-center mx-auto text-rose-400">
          <AlertTriangle className="h-6 w-6" />
        </div>
        <h2 className="text-xl font-bold text-white tracking-tight">System Encountered an Error</h2>
        <p className="text-xs text-muted">
          {error.message || "An unexpected error occurred while rendering the page."}
        </p>
        <div className="pt-2 flex items-center justify-center gap-3">
          <button
            onClick={() => reset()}
            className="inline-flex items-center gap-1.5 px-4 py-2 rounded-lg bg-surfaceSubtle border border-border hover:border-borderHighlight text-white text-xs font-semibold transition-colors"
          >
            <RefreshCw className="h-3.5 w-3.5" />
            <span>Try Again</span>
          </button>
          <Link
            href="/"
            className="inline-flex items-center gap-1.5 px-4 py-2 rounded-lg bg-primary hover:bg-primary-hover text-white text-xs font-semibold transition-colors"
          >
            <span>Overview</span>
          </Link>
        </div>
      </div>
    </div>
  );
}
