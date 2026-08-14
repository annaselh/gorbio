import { Cell, Pie, PieChart, ResponsiveContainer, Tooltip } from "recharts";
import { Card, CardHeader } from "@/shared/ui/Card";
import { TooltipShell } from "@/shared/charts/ChartTooltip";
import { SrTable } from "@/shared/charts/SrTable";
import { formatIDR, formatIDRCompact } from "@/shared/format";
import { useCashFlow } from "../api";

/**
 * Validated categorical palette (mirrors --color-chart-1..2 in index.css).
 * Literal hex rather than var() because recharts reads these into SVG
 * attributes and the tooltip swatch. Keyed by slice name so the colour follows
 * the meaning, never the array position.
 */
const SLICE_COLOR: Record<string, string> = {
  Income: "#059669",
  Expense: "#EA580C",
};
const FALLBACK_COLOR = "#2563EB";

export function CashFlow() {
  const { data: slices, isPending, isError } = useCashFlow();

  if (isPending || isError) {
    return (
      <Card className="flex h-full flex-col">
        <CardHeader title="Cash Flow" />
        <div className="grid flex-1 place-items-center px-4 py-12">
          <p className="text-sm text-ink-secondary">
            {isPending ? "Loading cash flow…" : "Cash flow is unavailable right now."}
          </p>
        </div>
      </Card>
    );
  }

  const total = slices.reduce((sum, s) => sum + s.value, 0);
  const pct = (v: number) => (total > 0 ? Math.round((v / total) * 100) : 0);
  const colored = slices.map((slice) => ({
    ...slice,
    color: SLICE_COLOR[slice.name] ?? FALLBACK_COLOR,
  }));

  if (total === 0) {
    return (
      <Card className="flex h-full flex-col">
        <CardHeader title="Cash Flow" />
        <div className="grid flex-1 place-items-center px-4 py-12">
          <p className="text-sm text-ink-secondary">
            No confirmed income or spend this month.
          </p>
        </div>
      </Card>
    );
  }

  return (
    <Card className="flex h-full flex-col">
      <CardHeader title="Cash Flow" />

      <div className="flex flex-1 flex-wrap items-center gap-3 px-4 pb-5">
        <div className="relative mx-auto size-[152px] shrink-0">
          <ResponsiveContainer width="100%" height="100%">
            <PieChart>
              <Pie
                data={colored}
                dataKey="value"
                nameKey="name"
                innerRadius="66%"
                outerRadius="100%"
                startAngle={90}
                endAngle={-270}
                /* 2px surface gap + 2px surface ring keeps adjacent fills from
                   touching, which is what makes the pair readable under CVD. */
                paddingAngle={2}
                stroke="var(--color-surface)"
                strokeWidth={2}
                isAnimationActive={false}
              >
                {colored.map((slice) => (
                  <Cell key={slice.name} fill={slice.color} />
                ))}
              </Pie>
              <Tooltip
                content={({ active, payload }) => {
                  if (!active || !payload?.length) return null;
                  const slice = payload[0].payload as (typeof colored)[number];
                  return (
                    <TooltipShell
                      label={slice.name}
                      rows={[
                        {
                          key: slice.name,
                          color: slice.color,
                          name: `${pct(slice.value)}%`,
                          value: formatIDR(slice.value),
                        },
                      ]}
                    />
                  );
                }}
              />
            </PieChart>
          </ResponsiveContainer>

          <div className="pointer-events-none absolute inset-0 grid place-items-center">
            <div className="text-center">
              <p className="text-[11px] text-ink-muted">Net</p>
              <p className="text-sm font-bold text-ink tnum">
                {formatIDRCompact(
                  (colored.find((s) => s.name === "Income")?.value ?? 0) -
                    (colored.find((s) => s.name === "Expense")?.value ?? 0),
                )}
              </p>
            </div>
          </div>
        </div>

        <ul className="mx-auto flex-1 space-y-2.5">
          {colored.map((slice) => (
            <li key={slice.name} className="flex items-center gap-2.5">
              <span
                aria-hidden
                className="size-2.5 shrink-0 rounded-full"
                style={{ backgroundColor: slice.color }}
              />
              <span className="flex-1 text-xs text-ink-secondary">
                {slice.name}
              </span>
              <span className="text-xs font-semibold text-ink tnum">
                {formatIDRCompact(slice.value)}
              </span>
              <span className="w-9 text-right text-xs text-ink-muted tnum">
                {pct(slice.value)}%
              </span>
            </li>
          ))}
        </ul>
      </div>

      <SrTable
        caption="Cash flow breakdown"
        columns={["Category", "Amount", "Share"]}
        rows={colored.map((s) => [s.name, formatIDR(s.value), `${pct(s.value)}%`])}
      />
    </Card>
  );
}
