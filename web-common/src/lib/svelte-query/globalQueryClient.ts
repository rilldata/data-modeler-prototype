import { QueryClient } from "@tanstack/svelte-query";

export function createQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: {
        refetchOnMount: false,
        refetchOnReconnect: false,
        refetchOnWindowFocus: false,
        retry: false,
        networkMode: "always",
        staleTime: Infinity,
      },
      mutations: {
        networkMode: "always",
      },
    },
  });
}

export const queryClient = createQueryClient();
