import { useState } from "react";
import type { FormEvent } from "react";
import { useMutation } from "@tanstack/react-query";
import { api, ApiError } from "./apiClient";
import { Button } from "@/shared/ui/Button";
import { AuthLayout, FIELD_CLASS, AuthMessage } from "./AuthLayout";

export function ForgotPasswordPage({ onBack }: { onBack: () => void }) {
  const [email, setEmail] = useState("");

  const request = useMutation({
    mutationFn: (address: string) =>
      api.post<{ message: string }>("/auth/password/forgot", {
        email: address,
      }),
  });

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    request.mutate(email);
  }

  // The server answers 202 whether or not the address exists, so the UI must
  // show the same confirmation either way - anything else would leak which
  // addresses are registered.
  if (request.isSuccess) {
    return (
      <AuthLayout
        title="Check your inbox"
        subtitle="If that address has an account, a reset link is on its way."
      >
        <AuthMessage>
          The link expires in 1 hour and can be used once.
        </AuthMessage>
        <Button variant="outline" onClick={onBack} className="w-full">
          Back to sign in
        </Button>
      </AuthLayout>
    );
  }

  const error =
    request.error instanceof ApiError
      ? request.error.message
      : request.error
        ? "Could not send the reset link. Please try again."
        : null;

  return (
    <AuthLayout
      title="Reset your password"
      subtitle="We'll email you a link to choose a new one."
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
            disabled={request.isPending}
            className={FIELD_CLASS}
            placeholder="you@company.com"
          />
        </div>

        {error ? <AuthMessage tone="alert">{error}</AuthMessage> : null}

        <Button
          type="submit"
          variant="primary"
          disabled={request.isPending}
          className="w-full"
        >
          {request.isPending ? "Sending…" : "Send reset link"}
        </Button>
        <Button variant="link" onClick={onBack} className="w-full">
          Back to sign in
        </Button>
      </form>
    </AuthLayout>
  );
}
