import { useId } from "react";
import { Area, AreaChart, ResponsiveContainer, Tooltip, YAxis } from "recharts";
import { TooltipShell } from "./ChartTooltip";

export interface SparkPoint {
  date: string;
  value: number;
}

/**
 * Single-series micro-chart inside a stat tile. One series means no legend —
 * the tile's own title names it. It still gets a hover layer, because it has a
 * plot to hover.
 */
export function Sparkline({
  data,
  color,
  seriesName,
  formatValue,
  formatLabel,
}: {
  data: SparkPoint[];
  color: string;
  seriesName: string;
  formatValue: (v: number) => string;
  formatLabel: (d: string) => string;
}) {
  const gradientId = useId().replace(/:/g, "");

  // Pad the domain so the curve floats clear of the tile's bottom edge.
  const values = data.map((d) => d.value);
  const min = Math.min(...values);
  const max = Math.max(...values);
  const pad = (max - min) * 0.35 || 1;

  return (
    <ResponsiveContainer width="100%" height="100%">
      <AreaChart data={data} margin={{ top: 4, right: 0, bottom: 0, left: 0 }}>
        <defs>
          <linearGradient id={gradientId} x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stopColor={color} stopOpacity={0.22} />
            <stop offset="100%" stopColor={color} stopOpacity={0} />
          </linearGradient>
        </defs>
        <YAxis hide domain={[min - pad, max + pad]} />
        <Tooltip
          cursor={{ stroke: "#c9ced6", strokeWidth: 1 }}
          content={({ active, payload, label }) => {
            if (!active || !payload?.length) return null;
            const point = payload[0].payload as SparkPoint;
            return (
              <TooltipShell
                label={formatLabel(point.date ?? String(label))}
                rows={[
                  {
                    key: seriesName,
                    color,
                    name: seriesName,
                    value: formatValue(Number(payload[0].value)),
                  },
                ]}
              />
            );
          }}
        />
        <Area
          type="monotone"
          dataKey="value"
          stroke={color}
          strokeWidth={2}
          fill={`url(#${gradientId})`}
          activeDot={{ r: 4, strokeWidth: 2, stroke: "#fff" }}
          isAnimationActive={false}
        />
      </AreaChart>
    </ResponsiveContainer>
  );
}
