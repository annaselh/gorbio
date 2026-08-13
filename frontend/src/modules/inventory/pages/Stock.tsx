import { PageHeader } from "@/core/Shell";
import { StockAlert } from "../components/StockAlert";

export default function Stock() {
  return (
    <>
      <PageHeader
        title="Inventory"
        subtitle="Stock levels and replenishment alerts."
      />
      <div className="grid grid-cols-1 gap-5 lg:grid-cols-3">
        <StockAlert />
      </div>
    </>
  );
}
