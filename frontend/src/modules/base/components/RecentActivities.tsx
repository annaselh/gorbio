import { Card, CardHeader } from "@/shared/ui/Card";
import { resolveIcon } from "@/shared/icons";
import { formatRelative } from "@/shared/format";
import { describeAction, moduleOfAction, useActivities } from "../api";

/**
 * Icon and colour per originating module. The backend sends the audit action
 * only; mapping it to a look belongs here, alongside the rest of the palette.
 */
const MODULE_STYLE: Record<string, { icon: string; tint: string; wash: string }> = {
  sales: { icon: "wallet", tint: "#EA580C", wash: "#FFF2E7" },
  procurement: { icon: "purchase", tint: "#2563EB", wash: "#E8F1FE" },
  inventory: { icon: "inventory", tint: "#7C3AED", wash: "#F1EBFE" },
  membership: { icon: "customers", tint: "#059669", wash: "#E6F6EF" },
  base: { icon: "customers", tint: "#059669", wash: "#E6F6EF" },
};
const FALLBACK_STYLE = { icon: "accounting", tint: "#D97706", wash: "#FEF3C7" };

export function RecentActivities() {
  const { data: entries, isPending, isError } = useActivities(5);

  return (
    <Card className="flex h-full flex-col">
      <CardHeader title="Recent Activities" />

      {isPending ? (
        <Message>Loading activity…</Message>
      ) : isError ? (
        <Message>Activity is unavailable right now.</Message>
      ) : entries.length === 0 ? (
        <Message>Nothing has happened yet.</Message>
      ) : (
        <ul className="flex-1 space-y-0.5 px-4 pb-4">
          {entries.map((entry) => {
            const style = MODULE_STYLE[moduleOfAction(entry.action)] ?? FALLBACK_STYLE;
            const Icon = resolveIcon(style.icon);
            return (
              <li key={entry.id} className="flex items-start gap-2 py-2.5">
                <span
                  className="grid size-7 shrink-0 place-items-center rounded-full"
                  style={{ backgroundColor: style.wash }}
                >
                  <Icon className="size-4" style={{ color: style.tint }} />
                </span>
                <span className="min-w-0 flex-1">
                  <span className="block truncate text-[13px] font-semibold text-ink">
                    {describeAction(entry.action)}
                  </span>
                  <span className="block truncate text-xs text-ink-secondary">
                    {entry.actor_name
                      ? `by ${entry.actor_name}`
                      : entry.resource_type}
                  </span>
                </span>
                <time
                  dateTime={entry.created_at}
                  className="shrink-0 pt-0.5 text-[11px] whitespace-nowrap text-ink-muted"
                >
                  {formatRelative(entry.created_at)}
                </time>
              </li>
            );
          })}
        </ul>
      )}
    </Card>
  );
}

function Message({ children }: { children: React.ReactNode }) {
  return (
    <div className="grid flex-1 place-items-center px-4 py-10">
      <p className="text-sm text-ink-secondary">{children}</p>
    </div>
  );
}
