import { useState } from "react";
import type { FormEvent } from "react";
import { useAuth } from "./auth";
import { Button } from "@/shared/ui/Button";
import { cn } from "@/shared/cn";

const FIELD_CLASS = cn(
  "w-full rounded-lg border border-hairline bg-surface px-3 py-2 text-sm text-ink",
  "placeholder:text-ink-secondary",
  "focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand",
  "disabled:cursor-not-allowed disabled:opacity-50",
);

export function LoginPage() {
  const { login, isLoggingIn, loginError } = useAuth();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [tenantSlug, setTenantSlug] = useState("");

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    try {
      await login({
        email,
        password,
        // Optional: the backend picks the earliest active membership when the
        // slug is blank, which is the common single-tenant case.
        tenant_slug: tenantSlug.trim() || undefined,
      });
    } catch {
      // Surfaced through loginError; nothing to do here.
    }
  }

  return (
    <main className="grid min-h-dvh place-items-center bg-canvas px-4">
      <div className="w-full max-w-sm">
        <div className="mb-8 text-center">
          <h1 className="text-xl font-semibold text-ink">Sign in to Orbio</h1>
          <p className="mt-1 text-sm text-ink-secondary">
            Enter your credentials to continue.
          </p>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4" noValidate>
          <div className="space-y-1.5">
            <label htmlFor="email" className="text-sm font-medium text-ink">
              Email
            </label>
            <input
              id="email"
              type="email"
              autoComplete="username"
              required
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              disabled={isLoggingIn}
              className={FIELD_CLASS}
              placeholder="you@company.com"
            />
          </div>

          <div className="space-y-1.5">
            <label htmlFor="password" className="text-sm font-medium text-ink">
              Password
            </label>
            <input
              id="password"
              type="password"
              autoComplete="current-password"
              required
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              disabled={isLoggingIn}
              className={FIELD_CLASS}
            />
          </div>

          <div className="space-y-1.5">
            <label htmlFor="tenant" className="text-sm font-medium text-ink">
              Company{" "}
              <span className="font-normal text-ink-secondary">(optional)</span>
            </label>
            <input
              id="tenant"
              type="text"
              autoComplete="organization"
              value={tenantSlug}
              onChange={(e) => setTenantSlug(e.target.value)}
              disabled={isLoggingIn}
              className={FIELD_CLASS}
              placeholder="acme"
            />
          </div>

          {loginError ? (
            <p
              role="alert"
              className="rounded-lg border border-hairline bg-hairline-soft px-3 py-2 text-sm text-ink"
            >
              {loginError}
            </p>
          ) : null}

          <Button
            type="submit"
            variant="primary"
            disabled={isLoggingIn}
            className="w-full"
          >
            {isLoggingIn ? "Signing in…" : "Sign in"}
          </Button>
        </form>
      </div>
    </main>
  );
}
