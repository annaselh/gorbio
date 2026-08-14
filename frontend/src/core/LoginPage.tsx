import { useState } from "react";
import type { FormEvent } from "react";
import { useAuth } from "./auth";
import { Button } from "@/shared/ui/Button";
import { AuthLayout, FIELD_CLASS, AuthMessage } from "./AuthLayout";

export function LoginPage({
  onForgotPassword,
}: {
  onForgotPassword: () => void;
}) {
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
    <AuthLayout
      title="Sign in to Orbio"
      subtitle="Enter your credentials to continue."
    >
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
          <div className="flex items-baseline justify-between gap-2">
            <label htmlFor="password" className="text-sm font-medium text-ink">
              Password
            </label>
            <Button
              variant="link"
              onClick={onForgotPassword}
              className="text-xs"
            >
              Forgot password?
            </Button>
          </div>
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
          <AuthMessage tone="alert">{loginError}</AuthMessage>
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
    </AuthLayout>
  );
}
