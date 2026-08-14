import { useState } from "react";
import { Plus } from "lucide-react";
import { PageHeader } from "@/core/Shell";
import { useAuth } from "@/core/auth";
import { Card, CardHeader } from "@/shared/ui/Card";
import { Badge } from "@/shared/ui/Badge";
import { Button } from "@/shared/ui/Button";
import { Pagination } from "@/shared/ui/Pagination";
import {
  CUSTOMER_STATUS_TONE,
  useCustomers,
  useSetCustomerStatus,
  type Customer,
} from "../data";
import { CustomerDialog } from "../components/CustomerDialog";

const PAGE_SIZE = 10;

export default function Customers() {
  const [page, setPage] = useState(1);
  const [search, setSearch] = useState("");
  const [creating, setCreating] = useState(false);
  const [editing, setEditing] = useState<Customer | null>(null);
  const { hasPermission } = useAuth();
  const canManage = hasPermission("crm.manage");
  const setStatus = useSetCustomerStatus();

  const { data, isPending, isError, isPlaceholderData } = useCustomers({
    limit: PAGE_SIZE,
    offset: (page - 1) * PAGE_SIZE,
    search: search.trim() || undefined,
  });

  const rows = data?.data ?? [];
  const total = data?.total ?? 0;
  const pageCount = Math.max(1, Math.ceil(total / PAGE_SIZE));

  const th =
    "px-2.5 py-2.5 text-left text-xs font-medium text-ink-secondary whitespace-nowrap";

  return (
    <>
      <PageHeader
        title="Customers"
        subtitle="Accounts you raise sales orders for."
      />

      <Card className="flex flex-col">
        <CardHeader
          title="All Customers"
          action={
            <span className="flex items-center gap-2">
              <input
                type="search"
                value={search}
                onChange={(e) => {
                  setSearch(e.target.value);
                  setPage(1);
                }}
                placeholder="Search name or code"
                aria-label="Search customers"
                className="rounded-lg border border-hairline bg-surface px-3 py-1.5 text-xs text-ink placeholder:text-ink-muted focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand"
              />
              {canManage && (
                <Button
                  variant="primary"
                  className="px-3 py-1.5 text-xs"
                  onClick={() => setCreating(true)}
                >
                  <Plus aria-hidden className="size-3.5" />
                  New customer
                </Button>
              )}
            </span>
          }
        />

        {isPending ? (
          <Message>Loading customers…</Message>
        ) : isError ? (
          <Message>Customers are unavailable right now.</Message>
        ) : rows.length === 0 ? (
          <Message>
            {search ? "No customers match that search." : "No customers yet."}
          </Message>
        ) : (
          <>
            <div
              className="scrollbar-slim overflow-x-auto"
              aria-busy={isPlaceholderData}
              style={{ opacity: isPlaceholderData ? 0.6 : 1 }}
            >
              <table className="w-full min-w-[640px] border-collapse">
                <thead>
                  <tr className="border-y border-hairline-soft bg-hairline-soft/40">
                    <th scope="col" className={th}>Code</th>
                    <th scope="col" className={th}>Name</th>
                    <th scope="col" className={th}>Email</th>
                    <th scope="col" className={th}>Credit</th>
                    <th scope="col" className={th}>Status</th>
                    {canManage && <th scope="col" className={th}>Actions</th>}
                  </tr>
                </thead>
                <tbody className="divide-y divide-hairline-soft">
                  {rows.map((customer) => (
                    <tr
                      key={customer.id}
                      className="transition-colors hover:bg-hairline-soft/50"
                    >
                      <th
                        scope="row"
                        className="px-2.5 py-3 text-left text-xs font-medium whitespace-nowrap text-ink"
                      >
                        {customer.code}
                      </th>
                      <td className="max-w-[220px] truncate px-2.5 py-3 text-xs text-ink">
                        {customer.name}
                      </td>
                      <td className="max-w-[200px] truncate px-2.5 py-3 text-xs text-ink-secondary">
                        {customer.email || "—"}
                      </td>
                      <td className="px-2.5 py-3 text-xs whitespace-nowrap text-ink-secondary tnum">
                        {customer.credit_term_days} days
                      </td>
                      <td className="px-2.5 py-3">
                        <Badge tone={CUSTOMER_STATUS_TONE[customer.status]}>
                          {customer.status}
                        </Badge>
                      </td>
                      {canManage && (
                        <td className="px-2.5 py-3">
                          <span className="flex gap-1.5">
                            <Button
                              variant="outline"
                              className="px-2.5 py-1 text-xs"
                              onClick={() => setEditing(customer)}
                            >
                              Edit
                            </Button>
                            <Button
                              variant="ghost"
                              className="px-2.5 py-1 text-xs"
                              disabled={setStatus.isPending}
                              onClick={() =>
                                setStatus.mutate({
                                  id: customer.id,
                                  status:
                                    customer.status === "Active"
                                      ? "Inactive"
                                      : "Active",
                                })
                              }
                            >
                              {customer.status === "Active"
                                ? "Deactivate"
                                : "Activate"}
                            </Button>
                          </span>
                        </td>
                      )}
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>

            <div className="border-t border-hairline-soft">
              <Pagination
                page={page}
                pageCount={pageCount}
                total={total}
                pageSize={PAGE_SIZE}
                onPageChange={(p) => setPage(Math.min(Math.max(p, 1), pageCount))}
              />
            </div>
          </>
        )}
      </Card>

      {creating && <CustomerDialog onClose={() => setCreating(false)} />}
      {editing && (
        <CustomerDialog customer={editing} onClose={() => setEditing(null)} />
      )}
    </>
  );
}

function Message({ children }: { children: React.ReactNode }) {
  return (
    <div className="grid place-items-center px-4 py-12">
      <p className="text-sm text-ink-secondary">{children}</p>
    </div>
  );
}
