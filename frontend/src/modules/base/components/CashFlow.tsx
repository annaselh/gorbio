import { useState } from "react";
import { Cell, Pie, PieChart, ResponsiveContainer, Tooltip } from "recharts";
import { Card, CardHeader } from "@/shared/ui/Card";
import { Select } from "@/shared/ui/Select";
import { TooltipShell } from "@/shared/charts/ChartTooltip";
import { SrTable } from "@/shared/charts/SrTable";
import { formatIDR, formatIDRCompact } from "@/shared/format";
import { cashFlow } from "../data";

const RANGES = ["This Month", "Last 3 Months", "This Year"] as const;

export function CashFlow() {
  const [range, setRange] = useState<string>(RANGES[0]);
  const total = cashFlow.reduce((sum, s) => sum + s.value, 0);
  const pct = (v: number) => Math.round((v / total) * 100);

  return (
    <Card className="flex h-full flex-col">
      <CardHeader
        title="Cash Flow"
        action={
          <Select
            aria-label="Cash flow range"
            value={range}
            onChange={setRange}
            options={RANGES}
            className="text-xs"
          />
        }
      />

      <div className="flex flex-1 flex-wrap items-center gap-3 px-4 pb-5">
        <div className="relative mx-auto size-[152px] shrink-0">
          <ResponsiveContainer width="100%" height="100%">
            <PieChart>
              <Pie
                data={cashFlow}
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
                {cashFlow.map((s) => (
                  <Cell key={s.name} fill={s.color} />
                ))}
              </Pie>
              <Tooltip
                content={({ active, payload }) => {
                  if (!active || !payload?.length) return null;
                  const slice = payload[0].payload as (typeof cashFlow)[number];
                  return (
                    <TooltipShell
                      label="Cash Flow"
                      rows={[
                        {
                          key: slice.name,
                          color: slice.color,
                          name: slice.name,
                          value: `${formatIDR(slice.value)} · ${pct(slice.value)}%`,
                        },
                      ]}
                    />
                  );
                }}
              />
            </PieChart>
          </ResponsiveContainer>

          <div className="pointer-events-none absolute inset-0 grid place-items-center text-center">
            <div>
              <p className="text-[19px] font-bold text-ink tnum">
                {formatIDRCompact(total)}
              </p>
              <p className="text-xs text-ink-muted">Total</p>
            </div>
          </div>
        </div>

        {/* Direct labels: every slice carries name + value + share in text. */}
        <ul className="min-w-[132px] flex-1 space-y-3.5">
          {cashFlow.map((s) => (
            <li key={s.name} className="flex items-start gap-2.5">
              <span
                aria-hidden
                className="mt-1 size-2.5 shrink-0 rounded-full"
                style={{ backgroundColor: s.color }}
              />
              <span className="min-w-0 flex-1">
                <span className="block text-xs text-ink-secondary">
                  {s.name}
                </span>
                {/* nowrap: "Rp 850.000.000" must not break across two lines */}
                <span className="block text-xs font-semibold whitespace-nowrap text-ink tnum">
                  {formatIDR(s.value)}
                </span>
              </span>
              <span className="shrink-0 text-xs font-medium text-ink-secondary tnum">
                {pct(s.value)}%
              </span>
            </li>
          ))}
        </ul>
      </div>

      <SrTable
        caption="Cash flow breakdown"
        columns={["Category", "Amount", "Share"]}
        rows={cashFlow.map((s) => [s.name, formatIDR(s.value), `${pct(s.value)}%`])}
      />
    </Card>
  );
}
