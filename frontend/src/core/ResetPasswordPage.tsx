import { useState } from "react";
import type { FormEvent } from "react";
import { useMutation } from "@tanstack/react-query";
import { api, ApiError } from "./apiClient";
import { Button } from "@/shared/ui/Button";
import { AuthLayout, FIELD_CLASS, AuthMessage } from "./AuthLayout";

/** Mirrors the server-side policy in app/modules/base/password.go. */
const MIN_PASSWORD_LENGTH = 15;

export function ResetPasswordPage({
  token,
  onDone,
}: {
  token: string;
  onDone: () => void;
}) {
  const [password, setPassword] = useState("");
  const [confirmation, setConfirmation] = useState("");

  const reset = useMutation({
    mutationFn: () => api.post<null>("/auth/password/reset", { token, password }),
  });

  // Checked here only to save a round trip; the server is the authority.
  const tooShort = password.length > 0 && password.length < MIN_PASSWORD_LENGTH;
  const mismatch = confirmation.length > 0 && password !== confirmation;
  const canSubmit =
    password.length >= MIN_PASSWORD_LENGTH &&
    password === confirmation &&
    !reset.isPending;

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (canSubmit) reset.mutate();
  }

  if (reset.isSuccess) {
    return (
      <AuthLayout
        title="Password updated"
        subtitle="Every other session has been signed out."
      >
        <Button variant="primary" onClick={onDone} className="w-full">
          Sign in
        </Button>
      </AuthLayout>
    );
  }

  const error =
    reset.error instanceof ApiError
      ? reset.error.message
      : reset.error
        ? "Could not reset the password. Please try again."
        : null;

  return (
    <AuthLayout
      title="Choose a new password"
      subtitle={`At least ${MIN_PASSWORD_LENGTH} characters.`}
    >
      <form onSubmit={handleSubmit} className="space-y-4" noValidate>
        <div className="space-y-1.5">
          <label htmlFor="password" className="text-sm font-medium text-ink">
            New password
          </label>
          <input
            id="password"
            type="password"
            autoComplete="new-password"
            required
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            disabled={reset.isPending}
            className={FIELD_CLASS}
          />
          {tooShort ? (
            <p className="text-xs text-status-critical">
              Use at least {MIN_PASSWORD_LENGTH} characters.
            </p>
          ) : null}
        </div>

        <div className="space-y-1.5">
          <label htmlFor="confirm" className="text-sm font-medium text-ink">
            Confirm password
          </label>
          <input
            id="confirm"
            type="password"
            autoComplete="new-password"
            required
            value={confirmation}
            onChange={(e) => setConfirmation(e.target.value)}
            disabled={reset.isPending}
            className={FIELD_CLASS}
          />
          {mismatch ? (
            <p className="text-xs text-status-critical">
              Both entries must match.
            </p>
          ) : null}
        </div>

        {error ? <AuthMessage tone="alert">{error}</AuthMessage> : null}

        <Button
          type="submit"
          variant="primary"
          disabled={!canSubmit}
          className="w-full"
        >
          {reset.isPending ? "Updating…" : "Update password"}
        </Button>
      </form>
    </AuthLayout>
  );
}
