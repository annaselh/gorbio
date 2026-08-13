import { BrowserRouter } from "react-router-dom";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { registry } from "@/core/modules";
import { Shell } from "@/core/Shell";
import { AppRouter } from "@/core/AppRouter";

const queryClient = new QueryClient({
  defaultOptions: {
    queries: { staleTime: 30_000, refetchOnWindowFocus: false },
  },
});

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        <Shell menu={registry.menuItems}>
          <AppRouter routes={registry.routes} />
        </Shell>
      </BrowserRouter>
    </QueryClientProvider>
  );
}
