import { useMemo, useState } from "react";
import {
  Area,
  AreaChart,
  CartesianGrid,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from "recharts";
import { Card, CardHeader } from "@/shared/ui/Card";
import { Select } from "@/shared/ui/Select";
import { TooltipShell } from "@/shared/charts/ChartTooltip";
import { SrTable } from "@/shared/charts/SrTable";
import { formatIDR, formatIDRCompact } from "@/shared/format";
import { salesLastMonth, salesThisMonth } from "../data";

const THIS_MONTH = "var(--color-spark-sales)";
const LAST_MONTH = "#9CA3AF";
const RANGES = ["This Month", "Last 3 Months", "This Year"] as const;

const MONTH_LABEL = ["Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"];

function dayLabel(day: number) {
  return `${day} ${MONTH_LABEL[4]}`;
}

export function SalesOverview() {
  const [range, setRange] = useState<string>(RANGES[0]);

  // Two months are compared day-over-day, so they share one x scale by day index.
  const data = useMemo(
    () =>
      salesThisMonth.map((p, i) => ({
        day: i + 1,
        thisMonth: p.value,
        lastMonth: salesLastMonth[i]?.value ?? null,
      })),
    [],
  );

  return (
    <Card className="flex h-full flex-col">
      <CardHeader
        title="Sales Overview"
        action={
          <Select
            aria-label="Sales overview range"
            value={range}
            onChange={setRange}
            options={RANGES}
            className="text-xs"
          />
        }
      />

      {/* Legend is always present for 2 series; solid vs dashed carries identity
          a second time, so the pair does not rely on colour alone. */}
      <ul className="flex flex-wrap items-center justify-end gap-5 px-5 pb-1">
        <li className="flex items-center gap-2 text-xs text-ink-secondary">
          <svg width="22" height="8" aria-hidden>
            <line x1="0" y1="4" x2="22" y2="4" stroke={THIS_MONTH} strokeWidth="2.5" strokeLinecap="round" />
          </svg>
          This Month
        </li>
        <li className="flex items-center gap-2 text-xs text-ink-secondary">
          <svg width="22" height="8" aria-hidden>
            <line x1="0" y1="4" x2="22" y2="4" stroke={LAST_MONTH} strokeWidth="2.5" strokeDasharray="4 3" strokeLinecap="round" />
          </svg>
          Last Month
        </li>
      </ul>

      <div className="h-[268px] px-2 pb-3">
        <ResponsiveContainer width="100%" height="100%">
          <AreaChart data={data} margin={{ top: 8, right: 14, bottom: 4, left: 6 }}>
            <defs>
              <linearGradient id="salesFill" x1="0" y1="0" x2="0" y2="1">
                <stop offset="0%" stopColor={THIS_MONTH} stopOpacity={0.22} />
                <stop offset="100%" stopColor={THIS_MONTH} stopOpacity={0.01} />
              </linearGradient>
            </defs>

            <CartesianGrid stroke="#F1F3F6" strokeWidth={1} vertical={false} />

            <XAxis
              dataKey="day"
              ticks={[1, 6, 11, 16, 21, 26, 31]}
              tickFormatter={dayLabel}
              tickLine={false}
              axisLine={false}
              tick={{ fill: "#9CA3AF", fontSize: 12 }}
              dy={8}
            />
            <YAxis
              domain={[0, 200_000_000]}
              ticks={[0, 50_000_000, 100_000_000, 150_000_000, 200_000_000]}
              tickFormatter={(v) => formatIDRCompact(Number(v))}
              tickLine={false}
              axisLine={false}
              tick={{ fill: "#9CA3AF", fontSize: 12 }}
              /* 72px: "Rp 200M" wraps to two lines at anything narrower. */
              width={72}
            />

            <Tooltip
              cursor={{ stroke: "#C9CED6", strokeWidth: 1, strokeDasharray: "4 3" }}
              content={({ active, payload, label }) => {
                if (!active || !payload?.length) return null;
                return (
                  <TooltipShell
                    label={dayLabel(Number(label))}
                    rows={payload.map((p) => ({
                      key: String(p.dataKey),
                      color: p.dataKey === "thisMonth" ? THIS_MONTH : LAST_MONTH,
                      name: p.dataKey === "thisMonth" ? "This Month" : "Last Month",
                      value: formatIDR(Number(p.value)),
                    }))}
                  />
                );
              }}
            />

            <Area
              type="monotone"
              dataKey="lastMonth"
              stroke={LAST_MONTH}
              strokeWidth={2}
              strokeDasharray="5 4"
              fill="none"
              activeDot={{ r: 4, strokeWidth: 2, stroke: "#fff" }}
              isAnimationActive={false}
            />
            <Area
              type="monotone"
              dataKey="thisMonth"
              stroke={THIS_MONTH}
              strokeWidth={2.5}
              fill="url(#salesFill)"
              activeDot={{ r: 5, strokeWidth: 2, stroke: "#fff" }}
              isAnimationActive={false}
            />
          </AreaChart>
        </ResponsiveContainer>
      </div>

      <SrTable
        caption="Sales overview — this month versus last month"
        columns={["Day", "This Month", "Last Month"]}
        rows={data.map((d) => [
          dayLabel(d.day),
          formatIDR(d.thisMonth),
          d.lastMonth == null ? "—" : formatIDR(d.lastMonth),
        ])}
      />
    </Card>
  );
}
