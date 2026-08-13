import { Card } from "@/shared/ui/Card";
import { Sparkline } from "@/shared/charts/Sparkline";
import { SrTable } from "@/shared/charts/SrTable";
import { resolveIcon } from "@/shared/icons";
import { formatNumber, formatPercent } from "@/shared/format";
import type { Kpi } from "../data";

export function StatCard({ kpi }: { kpi: Kpi }) {
  const Icon = resolveIcon(kpi.icon);
  const up = kpi.delta >= 0;

  return (
    <Card className="overflow-hidden">
      <div className="px-5 pt-5">
        <div className="flex items-start gap-3.5">
          <span
            className="grid size-11 shrink-0 place-items-center rounded-xl"
            style={{ backgroundColor: kpi.wash }}
          >
            <Icon className="size-5" style={{ color: kpi.tint }} />
          </span>
          <div className="min-w-0">
            <p className="text-sm font-medium text-ink-secondary">{kpi.label}</p>
            {/* 20px + tight tracking keeps "Rp 1.250.000.000" — the widest
                value in the set — on one line at the 4-up card width. */}
            <p className="mt-1 text-[20px] leading-tight font-bold tracking-[-0.015em] text-ink tnum">
              {kpi.value}
            </p>
          </div>
        </div>

        <p className="mt-3 flex items-center gap-1.5 text-xs">
          <span
            className="font-semibold tnum"
            style={{ color: up ? "var(--color-status-good)" : "var(--color-status-critical)" }}
          >
            {formatPercent(kpi.delta)}
          </span>
          <span className="text-ink-muted">vs last month</span>
        </p>
      </div>

      <div className="mt-2 h-16">
        <Sparkline
          data={kpi.spark}
          color={kpi.tint}
          seriesName={kpi.label}
          formatValue={(v) => formatNumber(v)}
          formatLabel={(d) => d}
        />
      </div>

      <SrTable
        caption={`${kpi.label} trend`}
        columns={["Date", kpi.label]}
        rows={kpi.spark.map((p) => [p.date, formatNumber(p.value)])}
      />
    </Card>
  );
}
