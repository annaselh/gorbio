package integration_test

import (
	"context"
	"testing"
	"time"

	"github.com/annaselh/gorbio/core"
	"github.com/annaselh/gorbio/internal/testdb"
	"github.com/annaselh/gorbio/modules/crm"
	"github.com/annaselh/gorbio/modules/dashboard"
	"github.com/annaselh/gorbio/modules/procurement"
	"github.com/annaselh/gorbio/modules/sales"
	"github.com/google/uuid"
)

// The dashboard is raw SQL against four modules' tables. None of it is covered
// by a unit test, and the customer count in particular casts a uuid column to
// text - a cast the Go compiler cannot check and no fake can catch.
//
// The clock is supplied rather than read: Summary takes `now`, so pinning it to
// a fixed instant makes these deterministic regardless of when they run.

var dashboardNow = time.Date(2026, 6, 15, 12, 0, 0, 0, time.UTC)
var dashboardOrderDate = time.Date(2026, 6, 10, 9, 0, 0, 0, time.UTC)

func TestARenamedCustomerIsCountedOnce(t *testing.T) {
	app := testdb.App(t)
	ctx := context.Background()
	tenantID := testdb.Tenant(t, app.DB)

	metrics := dashboardService(t, app)
	orders, _ := sales.ServiceFromApp(app)
	customers, _ := crm.ServiceFromApp(app)

	customer, err := customers.Create(ctx, tenantID, crm.CustomerInput{Name: "Before Rename"})
	if err != nil {
		t.Fatalf("create customer: %v", err)
	}

	// Two orders against the same CRM record, raised either side of a rename.
	// Counting by name would make this look like two customers; the count
	// coalesces the link first, so the rename is invisible to it.
	confirmSalesOrder(t, ctx, orders, tenantID, sales.CreateOrderInput{
		CustomerID: &customer.ID, OrderDate: dashboardOrderDate,
		Lines: []sales.LineInput{{SKU: "A", Description: "A", Quantity: 1, UnitPrice: 100_000}},
	})
	if _, err := customers.Update(ctx, tenantID, customer.ID, crm.CustomerInput{Name: "After Rename"}); err != nil {
		t.Fatalf("rename customer: %v", err)
	}
	confirmSalesOrder(t, ctx, orders, tenantID, sales.CreateOrderInput{
		CustomerID: &customer.ID, OrderDate: dashboardOrderDate,
		Lines: []sales.LineInput{{SKU: "B", Description: "B", Quantity: 1, UnitPrice: 200_000}},
	})

	summary, err := metrics.Summary(ctx, tenantID, dashboardNow)
	if err != nil {
		t.Fatalf("summary: %v", err)
	}

	if summary.Customers.Value != 1 {
		t.Fatalf("customers: got %d, want 1 - one account, renamed", summary.Customers.Value)
	}
	if summary.Orders.Value != 2 {
		t.Fatalf("orders: got %d, want 2", summary.Orders.Value)
	}
	if summary.Revenue.Value != 300_000 {
		t.Fatalf("revenue: got %d, want 300000", summary.Revenue.Value)
	}
}

func TestWalkInsAndLinkedCustomersAreCountedTogether(t *testing.T) {
	app := testdb.App(t)
	ctx := context.Background()
	tenantID := testdb.Tenant(t, app.DB)

	metrics := dashboardService(t, app)
	orders, _ := sales.ServiceFromApp(app)
	customers, _ := crm.ServiceFromApp(app)

	customer, err := customers.Create(ctx, tenantID, crm.CustomerInput{Name: "Linked Co"})
	if err != nil {
		t.Fatalf("create customer: %v", err)
	}

	confirmSalesOrder(t, ctx, orders, tenantID, sales.CreateOrderInput{
		CustomerID: &customer.ID, OrderDate: dashboardOrderDate,
		Lines: []sales.LineInput{{SKU: "A", Description: "A", Quantity: 1, UnitPrice: 100_000}},
	})
	// Two orders from the same walk-in name are one customer; a third name is
	// another. The COALESCE falls back to the name only where there is no link.
	for range 2 {
		confirmSalesOrder(t, ctx, orders, tenantID, sales.CreateOrderInput{
			CustomerName: "Repeat Walk-in", OrderDate: dashboardOrderDate,
			Lines: []sales.LineInput{{SKU: "A", Description: "A", Quantity: 1, UnitPrice: 50_000}},
		})
	}
	confirmSalesOrder(t, ctx, orders, tenantID, sales.CreateOrderInput{
		CustomerName: "One-off Walk-in", OrderDate: dashboardOrderDate,
		Lines: []sales.LineInput{{SKU: "A", Description: "A", Quantity: 1, UnitPrice: 50_000}},
	})

	summary, err := metrics.Summary(ctx, tenantID, dashboardNow)
	if err != nil {
		t.Fatalf("summary: %v", err)
	}
	if summary.Customers.Value != 3 {
		t.Fatalf("customers: got %d, want 3 (one linked, two distinct walk-in names)", summary.Customers.Value)
	}
}

func TestOnlyConfirmedSalesCountAsRevenue(t *testing.T) {
	app := testdb.App(t)
	ctx := context.Background()
	tenantID := testdb.Tenant(t, app.DB)

	metrics := dashboardService(t, app)
	orders, _ := sales.ServiceFromApp(app)

	// A draft is a proposal and a cancelled order never happened; neither is
	// revenue.
	if _, err := orders.Create(ctx, tenantID, sales.CreateOrderInput{
		CustomerName: "Draft Customer", OrderDate: dashboardOrderDate,
		Lines: []sales.LineInput{{SKU: "A", Description: "A", Quantity: 1, UnitPrice: 999_000}},
	}); err != nil {
		t.Fatalf("create draft: %v", err)
	}

	cancelled, err := orders.Create(ctx, tenantID, sales.CreateOrderInput{
		CustomerName: "Cancelled Customer", OrderDate: dashboardOrderDate,
		Lines: []sales.LineInput{{SKU: "A", Description: "A", Quantity: 1, UnitPrice: 888_000}},
	})
	if err != nil {
		t.Fatalf("create order to cancel: %v", err)
	}
	if _, err := orders.UpdateStatus(ctx, tenantID, cancelled.ID, sales.OrderStatusCancelled); err != nil {
		t.Fatalf("cancel: %v", err)
	}

	confirmSalesOrder(t, ctx, orders, tenantID, sales.CreateOrderInput{
		CustomerName: "Real Customer", OrderDate: dashboardOrderDate,
		Lines: []sales.LineInput{{SKU: "A", Description: "A", Quantity: 2, UnitPrice: 125_000}},
	})

	summary, err := metrics.Summary(ctx, tenantID, dashboardNow)
	if err != nil {
		t.Fatalf("summary: %v", err)
	}
	if summary.Revenue.Value != 250_000 {
		t.Fatalf("revenue: got %d, want 250000 - only the confirmed order counts", summary.Revenue.Value)
	}
	if summary.Orders.Value != 1 {
		t.Fatalf("orders: got %d, want 1", summary.Orders.Value)
	}
	if summary.Customers.Value != 1 {
		t.Fatalf("customers: got %d, want 1", summary.Customers.Value)
	}
}

func TestConfirmedAndReceivedPurchasesBothCountAsSpend(t *testing.T) {
	app := testdb.App(t)
	ctx := context.Background()
	tenantID := testdb.Tenant(t, app.DB)

	metrics := dashboardService(t, app)
	purchases, _ := procurement.ServiceFromApp(app)

	vendor, err := purchases.CreateVendor(ctx, tenantID, procurement.VendorInput{Name: "Spend Vendor"})
	if err != nil {
		t.Fatalf("create vendor: %v", err)
	}

	// A draft purchase is not yet a commitment; confirmed and received both
	// are, so both count.
	if _, err := purchases.CreatePurchaseOrder(ctx, tenantID, procurement.CreatePurchaseInput{
		VendorID: vendor.ID, OrderDate: dashboardOrderDate,
		Lines: []procurement.LineInput{{SKU: "P", Description: "P", Quantity: 1, UnitPrice: 700_000}},
	}); err != nil {
		t.Fatalf("create draft purchase: %v", err)
	}

	for _, amount := range []int64{100_000, 50_000} {
		order, err := purchases.CreatePurchaseOrder(ctx, tenantID, procurement.CreatePurchaseInput{
			VendorID: vendor.ID, OrderDate: dashboardOrderDate,
			Lines: []procurement.LineInput{{SKU: "P", Description: "P", Quantity: 1, UnitPrice: amount}},
		})
		if err != nil {
			t.Fatalf("create purchase: %v", err)
		}
		if _, err := purchases.UpdatePurchaseStatus(ctx, tenantID, order.ID, procurement.PurchaseStatusConfirmed); err != nil {
			t.Fatalf("confirm purchase: %v", err)
		}
		if amount == 50_000 {
			if _, err := purchases.UpdatePurchaseStatus(ctx, tenantID, order.ID, procurement.PurchaseStatusReceived); err != nil {
				t.Fatalf("receive purchase: %v", err)
			}
		}
	}

	summary, err := metrics.Summary(ctx, tenantID, dashboardNow)
	if err != nil {
		t.Fatalf("summary: %v", err)
	}
	if summary.Purchases.Value != 150_000 {
		t.Fatalf("purchases: got %d, want 150000 - the draft must not count", summary.Purchases.Value)
	}
	if summary.GrossMargin.Value != -150_000 {
		t.Fatalf("gross margin: got %d, want -150000 (no revenue, 150000 spend)", summary.GrossMargin.Value)
	}
}

func TestTheDashboardNeverReadsAnotherTenantsOrders(t *testing.T) {
	app := testdb.App(t)
	ctx := context.Background()
	mine := testdb.Tenant(t, app.DB)
	theirs := testdb.Tenant(t, app.DB)

	metrics := dashboardService(t, app)
	orders, _ := sales.ServiceFromApp(app)

	confirmSalesOrder(t, ctx, orders, theirs, sales.CreateOrderInput{
		CustomerName: "Their Customer", OrderDate: dashboardOrderDate,
		Lines: []sales.LineInput{{SKU: "A", Description: "A", Quantity: 1, UnitPrice: 500_000}},
	})

	summary, err := metrics.Summary(ctx, mine, dashboardNow)
	if err != nil {
		t.Fatalf("summary: %v", err)
	}
	if summary.Revenue.Value != 0 || summary.Orders.Value != 0 || summary.Customers.Value != 0 {
		t.Fatalf("another tenant's orders leaked into the summary: %+v", summary)
	}
}

// ------------------------------------------------------------------ helpers

func dashboardService(t *testing.T, app *core.App) *dashboard.Service {
	t.Helper()
	service, err := core.ServiceAs[*dashboard.Service](app, dashboard.ServiceName)
	if err != nil {
		t.Fatalf("resolve dashboard service: %v", err)
	}
	return service
}

func confirmSalesOrder(
	t *testing.T,
	ctx context.Context,
	orders *sales.Service,
	tenantID uuid.UUID,
	input sales.CreateOrderInput,
) *sales.Order {
	t.Helper()

	order, err := orders.Create(ctx, tenantID, input)
	if err != nil {
		t.Fatalf("create sales order: %v", err)
	}
	confirmed, err := orders.UpdateStatus(ctx, tenantID, order.ID, sales.OrderStatusConfirmed)
	if err != nil {
		t.Fatalf("confirm sales order: %v", err)
	}
	return confirmed
}
