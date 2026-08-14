import { useState } from "react";
import type { ReactNode } from "react";
import { useLocation, useNavigate, useSearchParams } from "react-router-dom";
import { useAuth } from "./auth";
import { LoginPage } from "./LoginPage";
import { ForgotPasswordPage } from "./ForgotPasswordPage";
import { ResetPasswordPage } from "./ResetPasswordPage";
import { VerifyEmailPage } from "./VerifyEmailPage";

/**
 * Decides between the unauthenticated screens and the authenticated shell.
 *
 * Gating here rather than per route means a new module cannot accidentally
 * publish an unguarded page: every registry route renders inside this subtree.
 * The server enforces the same rule independently - this only shapes the UI.
 */
export function AuthGate({ children }: { children: ReactNode }) {
  const { session, isReady } = useAuth();
  const { pathname } = useLocation();
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const [showForgot, setShowForgot] = useState(false);

  const token = searchParams.get("token") ?? "";
  const returnToSignIn = () => {
    setShowForgot(false);
    navigate("/", { replace: true });
  };

  // Token links arrive from email and must work whether or not a session
  // exists - the recipient may be signed in elsewhere, or not at all.
  if (pathname === "/reset-password" && token) {
    return <ResetPasswordPage token={token} onDone={returnToSignIn} />;
  }
  if (pathname === "/verify-email" && token) {
    return <VerifyEmailPage token={token} onDone={returnToSignIn} />;
  }

  if (!isReady) {
    return (
      <div className="grid min-h-dvh place-items-center bg-canvas">
        <p className="text-sm text-ink-secondary">Loading…</p>
      </div>
    );
  }

  if (!session) {
    return showForgot ? (
      <ForgotPasswordPage onBack={() => setShowForgot(false)} />
    ) : (
      <LoginPage onForgotPassword={() => setShowForgot(true)} />
    );
  }

  return <>{children}</>;
}
