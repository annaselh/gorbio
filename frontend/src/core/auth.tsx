import { createContext, useCallback, useContext, useMemo } from "react";
import type { ReactNode } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { api, ApiError } from "./apiClient";

/** Mirrors the payload of GET /api/auth/me in app/modules/base/handlers.go. */
export interface Session {
  user_id: string;
  email: string;
  display_name: string;
  tenant_id: string;
  tenant_slug: string;
  tenant_name: string;
  permissions: string[];
}

export interface LoginCredentials {
  email: string;
  password: string;
  tenant_slug?: string;
}

interface AuthContextValue {
  session: Session | null;
  isLoading: boolean;
  /** True once the session probe has settled, however it settled. */
  isReady: boolean;
  login: (credentials: LoginCredentials) => Promise<void>;
  logout: () => Promise<void>;
  isLoggingIn: boolean;
  loginError: string | null;
  hasPermission: (code: string) => boolean;
}

const AuthContext = createContext<AuthContextValue | null>(null);

const SESSION_KEY = ["auth", "session"] as const;

export function AuthProvider({ children }: { children: ReactNode }) {
  const queryClient = useQueryClient();

  const sessionQuery = useQuery({
    queryKey: SESSION_KEY,
    queryFn: () => api.get<Session>("/auth/me"),
    // A 401 here is the normal "not signed in" answer, not a transient fault:
    // retrying it just delays the login screen.
    retry: false,
    staleTime: 5 * 60_000,
  });

  const loginMutation = useMutation({
    mutationFn: (credentials: LoginCredentials) =>
      api.post<unknown>("/auth/login", credentials),
    onSuccess: async () => {
      // The login response carries only ids; refetch /auth/me for the
      // permission list so the UI has a single source of truth.
      await queryClient.invalidateQueries({ queryKey: SESSION_KEY });
    },
  });

  const logoutMutation = useMutation({
    mutationFn: () => api.post<null>("/auth/logout"),
    onSettled: () => {
      // Clear everything: cached module data belongs to the session that just
      // ended and must not leak into the next sign-in.
      queryClient.clear();
    },
  });

  const login = useCallback(
    async (credentials: LoginCredentials) => {
      await loginMutation.mutateAsync(credentials);
    },
    [loginMutation],
  );

  const logout = useCallback(async () => {
    await logoutMutation.mutateAsync();
  }, [logoutMutation]);

  const session = sessionQuery.data ?? null;

  const hasPermission = useCallback(
    (code: string) => session?.permissions.includes(code) ?? false,
    [session],
  );

  const loginError = loginMutation.error
    ? loginMutation.error instanceof ApiError
      ? loginMutation.error.message
      : "Unable to sign in. Please try again."
    : null;

  const value = useMemo<AuthContextValue>(
    () => ({
      session,
      isLoading: sessionQuery.isLoading,
      isReady: !sessionQuery.isPending,
      login,
      logout,
      isLoggingIn: loginMutation.isPending,
      loginError,
      hasPermission,
    }),
    [
      session,
      sessionQuery.isLoading,
      sessionQuery.isPending,
      login,
      logout,
      loginMutation.isPending,
      loginError,
      hasPermission,
    ],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth(): AuthContextValue {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("useAuth must be used inside <AuthProvider>");
  }
  return context;
}
