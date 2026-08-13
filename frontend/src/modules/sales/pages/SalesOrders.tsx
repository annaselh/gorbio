import { Plus } from "lucide-react";
import { PageHeader } from "@/core/Shell";
import { Button } from "@/shared/ui/Button";
import { SalesOrdersTable } from "../components/SalesOrdersTable";

export default function SalesOrders() {
  return (
    <>
      <PageHeader
        title="Sales Orders"
        subtitle="All quotations and confirmed orders across your companies."
        actions={
          <Button variant="primary">
            <Plus aria-hidden className="size-4" />
            New Order
          </Button>
        }
      />
      <SalesOrdersTable title="All Orders" pageSize={12} />
    </>
  );
}
