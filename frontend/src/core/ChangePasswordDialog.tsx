import { useEffect, useRef, useState } from "react";
import type { FormEvent } from "react";
import { useMutation } from "@tanstack/react-query";
import { api, ApiError } from "./apiClient";
import { Button } from "@/shared/ui/Button";
import { FIELD_CLASS, AuthMessage } from "./AuthLayout";

/** Mirrors the server-side policy in app/modules/base/password.go. */
const MIN_PASSWORD_LENGTH = 15;

export function ChangePasswordDialog({ onClose }: { onClose: () => void }) {
  const [current, setCurrent] = useState("");
  const [next, setNext] = useState("");
  const [confirmation, setConfirmation] = useState("");
  const dialogRef = useRef<HTMLDivElement>(null);

  const change = useMutation({
    mutationFn: () =>
      api.post<null>("/auth/password/change", {
        current_password: current,
        new_password: next,
      }),
  });

  useEffect(() => {
    const onKey = (event: KeyboardEvent) => {
      if (event.key === "Escape") onClose();
    };
    document.addEventListener("keydown", onKey);
    return () => document.removeEventListener("keydown", onKey);
  }, [onClose]);

  const tooShort = next.length > 0 && next.length < MIN_PASSWORD_LENGTH;
  const mismatch = confirmation.length > 0 && next !== confirmation;
  const canSubmit =
    current.length > 0 &&
    next.length >= MIN_PASSWORD_LENGTH &&
    next === confirmation &&
    !change.isPending;

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (canSubmit) change.mutate();
  }

  const error =
    change.error instanceof ApiError
      ? change.error.message
      : change.error
        ? "Could not change the password. Please try again."
        : null;

  return (
    <div
      className="fixed inset-0 z-50 grid place-items-center bg-ink/30 px-4"
      onMouseDown={(event) => {
        if (!dialogRef.current?.contains(event.target as Node)) onClose();
      }}
    >
      <div
        ref={dialogRef}
        role="dialog"
        aria-modal="true"
        aria-labelledby="change-password-title"
        className="w-full max-w-sm rounded-2xl border border-hairline bg-surface p-6 shadow-[0_16px_48px_rgba(16,24,40,0.18)]"
      >
        <h2
          id="change-password-title"
          className="text-lg font-semibold text-ink"
        >
          Change password
        </h2>

        {change.isSuccess ? (
          <div className="mt-4 space-y-4">
            <AuthMessage>
              Password updated. Your other devices have been signed out.
            </AuthMessage>
            <Button variant="primary" onClick={onClose} className="w-full">
              Done
            </Button>
          </div>
        ) : (
          <>
            <p className="mt-1 text-sm text-ink-secondary">
              At least {MIN_PASSWORD_LENGTH} characters. Other sessions will be
              signed out.
            </p>

            <form onSubmit={handleSubmit} className="mt-4 space-y-4" noValidate>
              <div className="space-y-1.5">
                <label
                  htmlFor="current-password"
                  className="text-sm font-medium text-ink"
                >
                  Current password
                </label>
                <input
                  id="current-password"
                  type="password"
                  autoComplete="current-password"
                  required
                  value={current}
                  onChange={(e) => setCurrent(e.target.value)}
                  disabled={change.isPending}
                  className={FIELD_CLASS}
                />
              </div>

              <div className="space-y-1.5">
                <label
                  htmlFor="new-password"
                  className="text-sm font-medium text-ink"
                >
                  New password
                </label>
                <input
                  id="new-password"
                  type="password"
                  autoComplete="new-password"
                  required
                  value={next}
                  onChange={(e) => setNext(e.target.value)}
                  disabled={change.isPending}
                  className={FIELD_CLASS}
                />
                {tooShort ? (
                  <p className="text-xs text-status-critical">
                    Use at least {MIN_PASSWORD_LENGTH} characters.
                  </p>
                ) : null}
              </div>

              <div className="space-y-1.5">
                <label
                  htmlFor="confirm-new-password"
                  className="text-sm font-medium text-ink"
                >
                  Confirm new password
                </label>
                <input
                  id="confirm-new-password"
                  type="password"
                  autoComplete="new-password"
                  required
                  value={confirmation}
                  onChange={(e) => setConfirmation(e.target.value)}
                  disabled={change.isPending}
                  className={FIELD_CLASS}
                />
                {mismatch ? (
                  <p className="text-xs text-status-critical">
                    Both entries must match.
                  </p>
                ) : null}
              </div>

              {error ? <AuthMessage tone="alert">{error}</AuthMessage> : null}

              <div className="flex gap-2">
                <Button
                  variant="outline"
                  onClick={onClose}
                  className="flex-1"
                  disabled={change.isPending}
                >
                  Cancel
                </Button>
                <Button
                  type="submit"
                  variant="primary"
                  disabled={!canSubmit}
                  className="flex-1"
                >
                  {change.isPending ? "Updating…" : "Update"}
                </Button>
              </div>
            </form>
          </>
        )}
      </div>
    </div>
  );
}
