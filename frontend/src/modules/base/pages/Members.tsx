import { useState } from "react";
import type { FormEvent } from "react";
import { PageHeader } from "@/core/Shell";
import { useAuth } from "@/core/auth";
import { ApiError } from "@/core/apiClient";
import { Card, CardHeader } from "@/shared/ui/Card";
import { Badge } from "@/shared/ui/Badge";
import { Button } from "@/shared/ui/Button";
import { formatDate } from "@/shared/format";
import { cn } from "@/shared/cn";
import {
  useInviteMember,
  useMembers,
  useRoles,
  useUpdateMemberRoles,
  useUpdateMemberStatus,
  type Member,
} from "../members";

const FIELD =
  "rounded-lg border border-hairline bg-surface px-3 py-2 text-sm text-ink placeholder:text-ink-muted focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand disabled:opacity-50";

export default function Members() {
  const { session, hasPermission } = useAuth();
  const canManage = hasPermission("membership.manage");

  const { data: members, isPending, isError } = useMembers();
  const { data: roles } = useRoles();
  const setStatus = useUpdateMemberStatus();
  const setRoles = useUpdateMemberRoles();
  const roleCodes = (roles ?? []).map((role) => role.code);

  const th =
    "px-2.5 py-2.5 text-left text-xs font-medium text-ink-secondary whitespace-nowrap";

  return (
    <>
      <PageHeader
        title="Users & Roles"
        subtitle={`People with access to ${session?.tenant_name ?? "this company"}.`}
      />

      {canManage && <InviteForm roleCodes={(roles ?? []).map((r) => r.code)} />}

      <Card className="mt-5 flex flex-col">
        <CardHeader title="Members" />

        {isPending ? (
          <Message>Loading members…</Message>
        ) : isError ? (
          <Message>Members are unavailable right now.</Message>
        ) : (
          <div className="scrollbar-slim overflow-x-auto">
            <table className="w-full min-w-[720px] border-collapse">
              <thead>
                <tr className="border-y border-hairline-soft bg-hairline-soft/40">
                  <th scope="col" className={th}>Name</th>
                  <th scope="col" className={th}>Email</th>
                  <th scope="col" className={th}>Roles</th>
                  <th scope="col" className={th}>Joined</th>
                  <th scope="col" className={th}>Status</th>
                  {canManage && <th scope="col" className={th}>Actions</th>}
                </tr>
              </thead>
              <tbody className="divide-y divide-hairline-soft">
                {members.map((member) => (
                  <MemberRow
                    key={member.membership_id}
                    member={member}
                    canManage={canManage}
                    isSelf={member.user_id === session?.user_id}
                    roleCodes={roleCodes}
                    onToggleStatus={() =>
                      setStatus.mutate({
                        id: member.membership_id,
                        status:
                          member.status === "active" ? "suspended" : "active",
                      })
                    }
                    onChangeRole={(role) =>
                      setRoles.mutate({ id: member.membership_id, roles: [role] })
                    }
                    pending={setStatus.isPending || setRoles.isPending}
                  />
                ))}
              </tbody>
            </table>
          </div>
        )}

        {/* Both refusals must be visible: the server rejects demoting or
            suspending a tenant's last owner, and silently swallowing that would
            leave the table showing a change that never happened. */}
        {(setStatus.error || setRoles.error) && (
          <p
            role="alert"
            className="border-t border-hairline-soft px-4 py-3 text-sm text-status-critical"
          >
            {(() => {
              const failure = setStatus.error ?? setRoles.error;
              return failure instanceof ApiError
                ? failure.message
                : "Could not update the member.";
            })()}
          </p>
        )}
      </Card>
    </>
  );
}

function MemberRow({
  member,
  canManage,
  isSelf,
  roleCodes,
  onToggleStatus,
  onChangeRole,
  pending,
}: {
  member: Member;
  canManage: boolean;
  isSelf: boolean;
  roleCodes: string[];
  onToggleStatus: () => void;
  onChangeRole: (role: string) => void;
  pending: boolean;
}) {
  const active = member.status === "active";
  // The server refuses to let anyone edit their own membership, so editing
  // controls are read-only for the signed-in user rather than showing a control
  // whose every use would fail.
  const editable = canManage && !isSelf;

  return (
    <tr className="transition-colors hover:bg-hairline-soft/50">
      <th scope="row" className="px-2.5 py-3 text-left text-xs font-medium text-ink">
        {member.display_name}
        {isSelf && <span className="ml-1.5 text-ink-muted">(you)</span>}
      </th>
      <td className="max-w-[220px] truncate px-2.5 py-3 text-xs text-ink-secondary">
        {member.email}
        {!member.email_verified && (
          <span className="ml-1.5 text-[11px] text-ink-muted">unverified</span>
        )}
      </td>
      <td className="px-2.5 py-3">
        {editable ? (
          <select
            value={member.roles[0] ?? ""}
            onChange={(e) => onChangeRole(e.target.value)}
            disabled={pending}
            aria-label={`Role for ${member.display_name}`}
            className="rounded-lg border border-hairline bg-surface px-2 py-1 text-xs text-ink focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand disabled:opacity-50"
          >
            {roleCodes.map((code) => (
              <option key={code} value={code}>{code}</option>
            ))}
          </select>
        ) : (
          <span className="flex flex-wrap gap-1">
            {member.roles.map((role) => (
              <Badge key={role} tone="neutral">{role}</Badge>
            ))}
          </span>
        )}
      </td>
      <td className="px-2.5 py-3 text-xs whitespace-nowrap text-ink-secondary tnum">
        {formatDate(member.joined_at)}
      </td>
      <td className="px-2.5 py-3">
        <Badge tone={active ? "good" : "critical"}>
          {active ? "Active" : "Suspended"}
        </Badge>
      </td>
      {canManage && (
        <td className="px-2.5 py-3">
          {/* The server refuses to let anyone change their own membership or
              remove the last owner; disabling here just avoids a pointless
              round trip for the self case. */}
          <Button
            variant="outline"
            className={cn("px-2.5 py-1 text-xs", isSelf && "invisible")}
            disabled={isSelf || pending}
            onClick={onToggleStatus}
          >
            {active ? "Suspend" : "Reactivate"}
          </Button>
        </td>
      )}
    </tr>
  );
}

function InviteForm({ roleCodes }: { roleCodes: string[] }) {
  const invite = useInviteMember();
  const [email, setEmail] = useState("");
  const [displayName, setDisplayName] = useState("");
  const [role, setRole] = useState("member");

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    invite.mutate(
      { email, display_name: displayName, roles: [role] },
      {
        onSuccess: () => {
          setEmail("");
          setDisplayName("");
        },
      },
    );
  }

  return (
    <Card>
      <CardHeader title="Invite a member" />
      <form onSubmit={handleSubmit} className="flex flex-wrap items-end gap-3 px-4 pb-4">
        <label className="flex flex-1 min-w-[180px] flex-col gap-1.5">
          <span className="text-xs font-medium text-ink">Email</span>
          <input
            type="email"
            required
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            disabled={invite.isPending}
            className={FIELD}
            placeholder="colleague@company.com"
          />
        </label>
        <label className="flex flex-1 min-w-[160px] flex-col gap-1.5">
          <span className="text-xs font-medium text-ink">Name</span>
          <input
            type="text"
            required
            value={displayName}
            onChange={(e) => setDisplayName(e.target.value)}
            disabled={invite.isPending}
            className={FIELD}
            placeholder="Jane Smith"
          />
        </label>
        <label className="flex flex-col gap-1.5">
          <span className="text-xs font-medium text-ink">Role</span>
          <select
            value={role}
            onChange={(e) => setRole(e.target.value)}
            disabled={invite.isPending}
            className={FIELD}
          >
            {(roleCodes.length > 0 ? roleCodes : ["member"]).map((code) => (
              <option key={code} value={code}>{code}</option>
            ))}
          </select>
        </label>
        <Button type="submit" variant="primary" disabled={invite.isPending}>
          {invite.isPending ? "Sending…" : "Send invite"}
        </Button>
      </form>

      {invite.isSuccess && (
        <p className="px-4 pb-4 text-sm text-ink-secondary">
          Invite sent. They will receive a link to set their password.
        </p>
      )}
      {invite.error ? (
        <p role="alert" className="px-4 pb-4 text-sm text-status-critical">
          {invite.error instanceof ApiError
            ? invite.error.message
            : "Could not send the invite."}
        </p>
      ) : null}
    </Card>
  );
}

function Message({ children }: { children: React.ReactNode }) {
  return (
    <div className="grid place-items-center px-4 py-12">
      <p className="text-sm text-ink-secondary">{children}</p>
    </div>
  );
}
