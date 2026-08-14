import { BrowserRouter } from "react-router-dom";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { registry } from "@/core/modules";
import { Shell } from "@/core/Shell";
import { AppRouter } from "@/core/AppRouter";
import { AuthProvider } from "@/core/auth";
import { AuthGate } from "@/core/AuthGate";
import { ApiError } from "@/core/apiClient";

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 30_000,
      refetchOnWindowFocus: false,
      // Retrying a 4xx cannot help: the request is wrong or the session is
      // gone. Only transient server and network faults are worth a second go.
      retry: (failureCount, error) => {
        if (error instanceof ApiError && error.status < 500) return false;
        return failureCount < 2;
      },
    },
  },
});

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <AuthProvider>
          <AuthGate>
            <Shell menu={registry.menuItems}>
              <AppRouter routes={registry.routes} />
            </Shell>
          </AuthGate>
        </AuthProvider>
      </BrowserRouter>
    </QueryClientProvider>
  );
}
