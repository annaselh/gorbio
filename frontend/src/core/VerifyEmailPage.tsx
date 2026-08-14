import { useEffect, useRef } from "react";
import { useMutation } from "@tanstack/react-query";
import { api, ApiError } from "./apiClient";
import { Button } from "@/shared/ui/Button";
import { AuthLayout, AuthMessage } from "./AuthLayout";

export function VerifyEmailPage({
  token,
  onDone,
}: {
  token: string;
  onDone: () => void;
}) {
  const verify = useMutation({
    mutationFn: () => api.post<null>("/auth/email/verify", { token }),
  });

  // Verification needs no input, so submit as soon as the page opens. The ref
  // guards against StrictMode's double effect consuming the single-use token
  // twice, where the second attempt would fail as already used.
  const { mutate } = verify;
  const submitted = useRef(false);
  useEffect(() => {
    if (submitted.current) return;
    submitted.current = true;
    mutate();
  }, [mutate]);

  if (verify.isPending || verify.isIdle) {
    return <AuthLayout title="Verifying your email…" />;
  }

  if (verify.isSuccess) {
    return (
      <AuthLayout
        title="Email verified"
        subtitle="Your address is confirmed and your account is active."
      >
        <Button variant="primary" onClick={onDone} className="w-full">
          Continue
        </Button>
      </AuthLayout>
    );
  }

  const message =
    verify.error instanceof ApiError
      ? verify.error.message
      : "Could not verify this address. Please try again.";

  return (
    <AuthLayout title="Verification failed">
      <AuthMessage tone="alert">{message}</AuthMessage>
      <Button variant="outline" onClick={onDone} className="w-full">
        Back to sign in
      </Button>
    </AuthLayout>
  );
}
