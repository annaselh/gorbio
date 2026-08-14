import { PageHeader } from "@/core/Shell";
import { SalesOrdersTable } from "../components/SalesOrdersTable";

export default function SalesOrders() {
  return (
    <>
      <PageHeader
        title="Sales Orders"
        subtitle="All quotations and confirmed orders across your companies."
      />
      {/* The create button lives in the card header beside the table it
          affects, rather than in the page header away from it. */}
      <SalesOrdersTable title="All Orders" pageSize={12} showActions />
    </>
  );
}
