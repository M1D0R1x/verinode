"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { api } from "@/lib/api";
import { useAuth } from "@/lib/use-auth";
import { Loader2, ShieldCheck, ArrowRight, KeyRound } from "lucide-react";

type Mode = "login" | "register";

export default function LoginPage() {
  const router = useRouter();
  const { login, register } = useAuth();
  const [mode, setMode] = useState<Mode>("login");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [company, setCompany] = useState("");
  const [err, setErr] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);
  const [demos, setDemos] = useState<{ role: string; email: string; password: string }[]>([]);

  useEffect(() => {
    api.demoCredentials().then((d) => setDemos(d.credentials)).catch(() => setDemos([]));
  }, []);

  async function submit(e: React.FormEvent) {
    e.preventDefault();
    setBusy(true);
    setErr(null);
    try {
      const user = mode === "login" ? await login(email, password) : await register(company, email, password);
      router.push(user.is_platform_staff ? "/admin" : "/buyer");
    } catch (e) {
      setErr(e instanceof Error ? e.message.replace(/^\[\d+\]\s*/, "") : "Authentication failed");
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="mx-auto grid min-h-[calc(100vh-4rem)] max-w-6xl grid-cols-1 items-center gap-12 px-4 py-16 sm:px-6 lg:grid-cols-2 lg:px-8">
      {/* Left — brand story */}
      <div className="hidden lg:block">
        <p className="eyebrow">Institutional access</p>
        <h1 className="mt-3 font-serif text-headline text-parchment">
          One desk for verified GPU forwards.
        </h1>
        <p className="mt-4 max-w-md text-muted">
          Platform staff administer participants, surveillance and settlement. Registered
          companies get a scoped desk to run their own RFQs, contracts and claims — never
          another company&apos;s data.
        </p>
        <div className="mt-8 space-y-3">
          {[
            ["Platform staff", "super_admin · admin — the compliance & ops console"],
            ["Company desk", "company_admin · trader · viewer — scoped to your entity"],
          ].map(([k, v]) => (
            <div key={k} className="flex items-start gap-3 rounded-card border border-line bg-ink-900/60 p-4">
              <ShieldCheck className="mt-0.5 h-4 w-4 shrink-0 text-signal" />
              <div>
                <div className="text-sm font-medium text-parchment">{k}</div>
                <div className="text-xs text-muted">{v}</div>
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* Right — form */}
      <div className="mx-auto w-full max-w-md">
        <div className="card card-raise p-7">
          <div className="flex rounded-lg border border-line bg-ink-900 p-1">
            {(["login", "register"] as Mode[]).map((m) => (
              <button
                key={m}
                onClick={() => { setMode(m); setErr(null); }}
                className={`flex-1 rounded-md px-3 py-2 text-sm font-medium capitalize transition-colors ${
                  mode === m ? "bg-signal text-ink-950" : "text-muted hover:text-parchment"
                }`}
              >
                {m === "login" ? "Sign in" : "Register company"}
              </button>
            ))}
          </div>

          <form onSubmit={submit} className="mt-6 space-y-4">
            {mode === "register" && (
              <div>
                <label className="text-sm font-medium text-parchment">Company legal name</label>
                <input
                  value={company}
                  onChange={(e) => setCompany(e.target.value)}
                  required
                  className="mt-2 w-full rounded-lg border border-line bg-ink-800 px-3 py-2.5 text-sm text-parchment outline-none focus:border-signal"
                  placeholder="Acme Compute Inc."
                />
              </div>
            )}
            <div>
              <label className="text-sm font-medium text-parchment">Work email</label>
              <input
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
                autoComplete="email"
                className="mt-2 w-full rounded-lg border border-line bg-ink-800 px-3 py-2.5 text-sm text-parchment outline-none focus:border-signal"
                placeholder="you@company.com"
              />
            </div>
            <div>
              <label className="text-sm font-medium text-parchment">Password</label>
              <input
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
                autoComplete={mode === "login" ? "current-password" : "new-password"}
                className="mt-2 w-full rounded-lg border border-line bg-ink-800 px-3 py-2.5 text-sm text-parchment outline-none focus:border-signal"
                placeholder={mode === "register" ? "at least 8 characters" : "••••••••"}
              />
            </div>

            {err && <p className="text-sm text-alert">{err}</p>}

            <button
              type="submit"
              disabled={busy}
              className="group flex w-full items-center justify-center gap-2 rounded-xl bg-signal px-5 py-3 text-sm font-semibold text-ink-950 shadow-glow transition-all hover:bg-signal-bright active:scale-[0.99] disabled:opacity-50"
            >
              {busy ? <Loader2 className="h-4 w-4 animate-spin" /> : <ArrowRight className="h-4 w-4" />}
              {mode === "login" ? "Sign in" : "Create company desk"}
            </button>
          </form>
        </div>

        {/* Demo credentials */}
        {demos.length > 0 && (
          <div className="mt-5 rounded-card border border-line bg-ink-900/60 p-4">
            <div className="flex items-center gap-1.5 text-xs text-muted">
              <KeyRound className="h-3.5 w-3.5 text-signal" /> Demo logins — click to fill
            </div>
            <div className="mt-3 space-y-1.5">
              {demos.map((d) => (
                <button
                  key={d.email}
                  onClick={() => { setMode("login"); setEmail(d.email); setPassword(d.password); }}
                  className="flex w-full items-center justify-between rounded-md px-2 py-1.5 text-left text-xs transition-colors hover:bg-ink-850"
                >
                  <span className="tabular text-muted">{d.email}</span>
                  <span className="rounded-pill border border-line px-2 py-0.5 text-[10px] text-signal">{d.role}</span>
                </button>
              ))}
            </div>
          </div>
        )}

        <p className="mt-4 text-center text-xs text-muted-soft">
          <Link href="/" className="hover:text-parchment">← Back to the marketplace</Link>
        </p>
      </div>
    </div>
  );
}
