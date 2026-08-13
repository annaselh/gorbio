import { useState } from "react";
import { SlidersHorizontal } from "lucide-react";
import { PageHeader } from "@/core/Shell";
import { Slot } from "@/core/Slot";
import { Button } from "@/shared/ui/Button";
import { Select } from "@/shared/ui/Select";
import { StatCard } from "../components/StatCard";
import { SalesOverview } from "../components/SalesOverview";
import { TopProducts } from "../components/TopProducts";
import { RecentActivities } from "../components/RecentActivities";
import { CashFlow } from "../components/CashFlow";
import { kpis } from "../data";

/** Shared by both widget rows so the two grids stay column-aligned. */
const ROW =
  "mt-5 grid grid-cols-1 gap-5 lg:grid-cols-2 " +
  "xl:grid-cols-[minmax(0,1.85fr)_minmax(0,1.2fr)_minmax(0,1.05fr)]";

/** At the 2-column breakpoint the third card takes the full width. */
const LAST = "lg:col-span-2 xl:col-span-1";

const PERIODS = [
  "May 1 – May 31, 2024",
  "Apr 1 – Apr 30, 2024",
  "Q2 2024",
  "Year to date",
] as const;

export default function Dashboard() {
  const [period, setPeriod] = useState<string>(PERIODS[0]);

  return (
    <>
      <PageHeader
        title="Dashboard"
        subtitle="Welcome back, John Doe! Here's what's happening with your business."
        actions={
          <>
            {/* Filters sit in one row above the charts. */}
            <Select
              aria-label="Reporting period"
              value={period}
              onChange={setPeriod}
              options={PERIODS}
            />
            <Button variant="outline">
              <SlidersHorizontal aria-hidden className="size-4" />
              Customize
            </Button>
          </>
        }
      />

      <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 xl:grid-cols-4">
        {kpis.map((k) => (
          <StatCard key={k.id} kpi={k} />
        ))}
      </div>

      {/* Column ratios come from the mockup (~46 / 28 / 25), which 12-col
          spans can't express without starving the two narrow cards. */}
      <div className={ROW}>
        <div>
          <SalesOverview />
        </div>
        <div>
          <TopProducts />
        </div>
        <div className={LAST}>
          <RecentActivities />
        </div>
      </div>

      <div className={ROW}>
        {/* Filled by the sales module — base never imports it. */}
        <div>
          <Slot name="dashboard.wide" />
        </div>
        <div>
          <CashFlow />
        </div>
        {/* Filled by the inventory module. */}
        <div className={LAST}>
          <Slot name="dashboard.aside" />
        </div>
      </div>
    </>
  );
}
