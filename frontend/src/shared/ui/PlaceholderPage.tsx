import { PageHeader } from "@/core/Shell";
import { Card } from "@/shared/ui/Card";
import { resolveIcon } from "@/shared/icons";

/**
 * Honest stand-in for a module that is registered in the menu but has no UI
 * yet. It says so plainly rather than rendering an empty dashboard that looks
 * broken.
 */
export function PlaceholderPage({
  title,
  icon,
  note = "This module is on the roadmap. Its menu entry is registered so the shell can route to it, but no screens have been built yet.",
}: {
  title: string;
  icon?: string;
  note?: string;
}) {
  const Icon = resolveIcon(icon);
  return (
    <>
      <PageHeader title={title} subtitle="Module not implemented yet" />
      <Card className="grid min-h-[320px] place-items-center px-6 py-16 text-center">
        <div className="max-w-md">
          <span className="mx-auto mb-4 grid size-14 place-items-center rounded-2xl bg-hairline-soft">
            <Icon className="size-6 text-ink-muted" />
          </span>
          <p className="text-base font-semibold text-ink">{title}</p>
          <p className="mt-2 text-sm text-ink-secondary">{note}</p>
        </div>
      </Card>
    </>
  );
}
