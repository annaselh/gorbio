import { Card, CardHeader } from "@/shared/ui/Card";
import { ViewAllButton } from "@/shared/ui/Button";
import { resolveIcon } from "@/shared/icons";
import { formatRelative } from "@/shared/format";
import { activities, DEMO_NOW } from "../data";

export function RecentActivities() {
  return (
    <Card className="flex h-full flex-col">
      <CardHeader title="Recent Activities" action={<ViewAllButton />} />
      <ul className="flex-1 space-y-0.5 px-4 pb-4">
        {activities.map((a) => {
          const Icon = resolveIcon(a.icon);
          return (
            <li key={a.id} className="flex items-start gap-2 py-2.5">
              <span
                className="grid size-7 shrink-0 place-items-center rounded-full"
                style={{ backgroundColor: a.wash }}
              >
                <Icon className="size-4" style={{ color: a.tint }} />
              </span>
              <span className="min-w-0 flex-1">
                <span className="block truncate text-[13px] font-semibold text-ink">
                  {a.title}
                </span>
                <span className="block truncate text-xs text-ink-secondary">
                  {a.detail}
                </span>
              </span>
              <time
                dateTime={a.at}
                className="shrink-0 pt-0.5 text-[11px] whitespace-nowrap text-ink-muted"
              >
                {formatRelative(a.at, DEMO_NOW)}
              </time>
            </li>
          );
        })}
      </ul>
    </Card>
  );
}
