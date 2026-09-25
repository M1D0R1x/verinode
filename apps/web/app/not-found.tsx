import Link from "next/link";
import { ArrowLeft, AlertCircle } from "lucide-react";

export default function NotFound() {
  return (
    <div className="min-h-[70vh] flex items-center justify-center px-4">
      <div className="max-w-md w-full p-8 rounded-2xl bg-surface border border-border text-center space-y-4">
        <div className="h-12 w-12 rounded-xl bg-surfaceSubtle border border-border/60 flex items-center justify-center mx-auto text-accent">
          <AlertCircle className="h-6 w-6" />
        </div>
        <h2 className="text-xl font-bold text-white tracking-tight">Page Not Found</h2>
        <p className="text-xs text-muted">
          The requested institutional reservation, trade envelope, or console route does not exist.
        </p>
        <div className="pt-2">
          <Link
            href="/"
            className="inline-flex items-center gap-1.5 px-4 py-2 rounded-lg bg-primary hover:bg-primary-hover text-white text-xs font-semibold transition-colors"
          >
            <ArrowLeft className="h-3.5 w-3.5" />
            <span>Return to Overview</span>
          </Link>
        </div>
      </div>
    </div>
  );
}
