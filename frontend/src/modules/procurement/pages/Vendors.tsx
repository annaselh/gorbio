import { useState } from "react";
import { PageHeader } from "@/core/Shell";
import { Card, CardHeader } from "@/shared/ui/Card";
import { Badge } from "@/shared/ui/Badge";
import { Pagination } from "@/shared/ui/Pagination";
import { VENDOR_STATUS_TONE, useVendors } from "../data";

const PAGE_SIZE = 10;

export default function Vendors() {
  const [page, setPage] = useState(1);
  const [search, setSearch] = useState("");

  const { data, isPending, isError, isPlaceholderData } = useVendors({
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
        title="Vendors"
        subtitle="Suppliers you raise purchase orders against."
      />

      <Card className="flex flex-col">
        <CardHeader
          title="All Vendors"
          action={
            <input
              type="search"
              value={search}
              onChange={(e) => {
                setSearch(e.target.value);
                setPage(1);
              }}
              placeholder="Search name or code"
              aria-label="Search vendors"
              className="rounded-lg border border-hairline bg-surface px-3 py-1.5 text-xs text-ink placeholder:text-ink-muted focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-brand"
            />
          }
        />

        {isPending ? (
          <Message>Loading vendors…</Message>
        ) : isError ? (
          <Message>Vendors are unavailable right now.</Message>
        ) : rows.length === 0 ? (
          <Message>{search ? "No vendors match that search." : "No vendors yet."}</Message>
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
                    <th scope="col" className={th}>Terms</th>
                    <th scope="col" className={th}>Status</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-hairline-soft">
                  {rows.map((vendor) => (
                    <tr
                      key={vendor.id}
                      className="transition-colors hover:bg-hairline-soft/50"
                    >
                      <th
                        scope="row"
                        className="px-2.5 py-3 text-left text-xs font-medium whitespace-nowrap text-ink"
                      >
                        {vendor.code}
                      </th>
                      <td className="max-w-[220px] truncate px-2.5 py-3 text-xs text-ink">
                        {vendor.name}
                      </td>
                      <td className="max-w-[200px] truncate px-2.5 py-3 text-xs text-ink-secondary">
                        {vendor.email || "—"}
                      </td>
                      <td className="px-2.5 py-3 text-xs whitespace-nowrap text-ink-secondary tnum">
                        {vendor.payment_term_days} days
                      </td>
                      <td className="px-2.5 py-3">
                        <Badge tone={VENDOR_STATUS_TONE[vendor.status]}>
                          {vendor.status}
                        </Badge>
                      </td>
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
