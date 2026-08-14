import { PageHeader } from "@/core/Shell";
import { Slot } from "@/core/Slot";
import { useAuth } from "@/core/auth";
import { StatCard, StatCardSkeleton } from "../components/StatCard";
import { SalesOverview } from "../components/SalesOverview";
import { TopProducts } from "../components/TopProducts";
import { RecentActivities } from "../components/RecentActivities";
import { CashFlow } from "../components/CashFlow";
import { useDashboardSummary, useSalesSeries } from "../api";
import { buildKpiCards } from "../kpis";

/** Shared by both widget rows so the two grids stay column-aligned. */
const ROW =
  "mt-5 grid grid-cols-1 gap-5 lg:grid-cols-2 " +
  "xl:grid-cols-[minmax(0,1.85fr)_minmax(0,1.2fr)_minmax(0,1.05fr)]";

/** At the 2-column breakpoint the third card takes the full width. */
const LAST = "lg:col-span-2 xl:col-span-1";

export default function Dashboard() {
  const { session } = useAuth();
  const summary = useDashboardSummary();
  const series = useSalesSeries();

  const cards =
    summary.data && series.data
      ? buildKpiCards(summary.data, series.data.current)
      : [];

  const firstName = session?.display_name?.split(" ")[0] ?? "there";

  return (
    <>
      <PageHeader
        title="Dashboard"
        subtitle={`Welcome back, ${firstName}! Here's what's happening this month.`}
      />

      <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 xl:grid-cols-4">
        {cards.length > 0
          ? cards.map((card) => <StatCard key={card.id} kpi={card} />)
          : // Four skeletons rather than one spinner: the row keeps its shape,
            // so the widgets below do not jump when the figures arrive.
            Array.from({ length: 4 }, (_, i) => <StatCardSkeleton key={i} />)}
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
