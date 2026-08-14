import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { api } from "@/core/apiClient";

/** Mirrors Member in app/modules/base/members.go. */
export interface Member {
  membership_id: string;
  user_id: string;
  email: string;
  display_name: string;
  status: "active" | "suspended";
  roles: string[];
  joined_at: string;
  last_login_at?: string;
  email_verified: boolean;
}

export interface Role {
  id: string;
  code: string;
  name: string;
  description: string;
}

const membersKey = ["members"] as const;
const rolesKey = ["roles"] as const;

export function useMembers() {
  return useQuery({
    queryKey: membersKey,
    queryFn: () => api.get<{ data: Member[] }>("/members"),
    select: (response) => response.data ?? [],
  });
}

export function useRoles() {
  return useQuery({
    queryKey: rolesKey,
    queryFn: () => api.get<{ data: Role[] }>("/roles"),
    select: (response) => response.data ?? [],
    // The role list is seeded at migration time and effectively static.
    staleTime: 30 * 60_000,
  });
}

export function useInviteMember() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: {
      email: string;
      display_name: string;
      roles: string[];
    }) => api.post<{ data: Member }>("/members", input),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: membersKey }),
  });
}

export function useUpdateMemberRoles() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, roles }: { id: string; roles: string[] }) =>
      api.put<{ data: Member }>(`/members/${id}/roles`, { roles }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: membersKey }),
  });
}

export function useUpdateMemberStatus() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, status }: { id: string; status: Member["status"] }) =>
      api.put<{ data: Member }>(`/members/${id}/status`, { status }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: membersKey }),
  });
}
