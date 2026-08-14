package integration_test

import (
	"context"
	"testing"

	"github.com/annaselh/gorbio/internal/testdb"
	"github.com/annaselh/gorbio/modules/inventory"
	"github.com/annaselh/gorbio/modules/procurement"
	"github.com/google/uuid"
)

// Receiving a purchase order is the one path in this codebase where one module
// writes through another's service inside a shared transaction. None of it is
// reachable without a database: the hook, the row locks it takes, and the
// upsert it performs all live below the validation the unit tests can see.

func TestReceivingAPurchaseOrderRaisesStock(t *testing.T) {
	app := testdb.App(t)
	ctx := context.Background()
	tenantID := testdb.Tenant(t, app.DB)

	stock, err := inventory.ServiceFromApp(app)
	if err != nil {
		t.Fatalf("resolve inventory service: %v", err)
	}
	orders, err := procurement.ServiceFromApp(app)
	if err != nil {
		t.Fatalf("resolve procurement service: %v", err)
	}

	item, err := stock.Create(ctx, tenantID, inventory.CreateItemInput{
		SKU: "WIDGET-1", Name: "Widget", Quantity: 5,
	})
	if err != nil {
		t.Fatalf("create stock item: %v", err)
	}

	order := receivedOrder(t, ctx, orders, tenantID, []procurement.LineInput{
		{SKU: "WIDGET-1", Description: "Widget", Quantity: 12, UnitPrice: 1000},
	})
	if order.Status != procurement.PurchaseStatusReceived {
		t.Fatalf("order status: got %s, want Received", order.Status)
	}

	after, err := stock.Get(ctx, tenantID, item.ID)
	if err != nil {
		t.Fatalf("get stock item: %v", err)
	}
	if after.Quantity != 17 {
		t.Fatalf("stock after receipt: got %d, want 17 (5 on hand + 12 received)", after.Quantity)
	}
}

func TestReceivingOpensAStockItemForAnUnknownSKU(t *testing.T) {
	app := testdb.App(t)
	ctx := context.Background()
	tenantID := testdb.Tenant(t, app.DB)

	stock, _ := inventory.ServiceFromApp(app)
	orders, _ := procurement.ServiceFromApp(app)

	// Purchase lines carry free-text SKUs with no foreign key into inventory,
	// so a receipt against a SKU nobody has stocked before has to work.
	receivedOrder(t, ctx, orders, tenantID, []procurement.LineInput{
		{SKU: "brand-new", Description: "Brand new item", Quantity: 4, UnitPrice: 2500},
	})

	items, err := stock.List(ctx, tenantID, inventory.ListFilter{})
	if err != nil {
		t.Fatalf("list stock: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected one stock item to be opened, got %d", len(items))
	}
	if items[0].SKU != "BRAND-NEW" {
		t.Fatalf("sku: got %q, want BRAND-NEW", items[0].SKU)
	}
	if items[0].Quantity != 4 {
		t.Fatalf("quantity: got %d, want 4", items[0].Quantity)
	}
	if items[0].Name != "Brand new item" {
		t.Fatalf("name should come from the line description, got %q", items[0].Name)
	}
	if items[0].Unit != "pcs" {
		t.Fatalf("unit should default to pcs, got %q", items[0].Unit)
	}
}

func TestReceivingTheSameOrderTwiceDoesNotDoubleStock(t *testing.T) {
	app := testdb.App(t)
	ctx := context.Background()
	tenantID := testdb.Tenant(t, app.DB)

	stock, _ := inventory.ServiceFromApp(app)
	orders, _ := procurement.ServiceFromApp(app)

	order := receivedOrder(t, ctx, orders, tenantID, []procurement.LineInput{
		{SKU: "REPEAT-1", Description: "Repeat", Quantity: 7, UnitPrice: 1000},
	})

	// Received -> Received is a permitted no-op transition, so a client that
	// resends the status - a double-click, a retried request - must not book
	// the delivery a second time.
	if _, err := orders.UpdatePurchaseStatus(ctx, tenantID, order.ID, procurement.PurchaseStatusReceived); err != nil {
		t.Fatalf("re-sending Received should be accepted as a no-op: %v", err)
	}

	items, err := stock.List(ctx, tenantID, inventory.ListFilter{})
	if err != nil {
		t.Fatalf("list stock: %v", err)
	}
	if len(items) != 1 || items[0].Quantity != 7 {
		t.Fatalf("stock should still be 7 after a repeated receipt, got %+v", items)
	}
}

func TestReceivingSumsRepeatedSKUsOnOneOrder(t *testing.T) {
	app := testdb.App(t)
	ctx := context.Background()
	tenantID := testdb.Tenant(t, app.DB)

	stock, _ := inventory.ServiceFromApp(app)
	orders, _ := procurement.ServiceFromApp(app)

	// One order may name the same SKU on two lines, at different prices. The
	// stock movement is their sum.
	receivedOrder(t, ctx, orders, tenantID, []procurement.LineInput{
		{SKU: "SUM-1", Description: "First batch", Quantity: 3, UnitPrice: 1000},
		{SKU: "OTHER", Description: "Other", Quantity: 1, UnitPrice: 500},
		{SKU: "sum-1", Description: "Second batch", Quantity: 4, UnitPrice: 1200},
	})

	items, err := stock.List(ctx, tenantID, inventory.ListFilter{})
	if err != nil {
		t.Fatalf("list stock: %v", err)
	}

	quantities := map[string]int{}
	for _, item := range items {
		quantities[item.SKU] = item.Quantity
	}
	if quantities["SUM-1"] != 7 {
		t.Fatalf("SUM-1: got %d, want 7 (3 + 4, matched case-insensitively)", quantities["SUM-1"])
	}
	if quantities["OTHER"] != 1 {
		t.Fatalf("OTHER: got %d, want 1", quantities["OTHER"])
	}
}

func TestReceivingDoesNotTouchAnotherTenantsStock(t *testing.T) {
	app := testdb.App(t)
	ctx := context.Background()
	mine := testdb.Tenant(t, app.DB)
	theirs := testdb.Tenant(t, app.DB)

	stock, _ := inventory.ServiceFromApp(app)
	orders, _ := procurement.ServiceFromApp(app)

	// The same SKU string in two tenants is two different items.
	theirItem, err := stock.Create(ctx, theirs, inventory.CreateItemInput{
		SKU: "SHARED-SKU", Name: "Theirs", Quantity: 100,
	})
	if err != nil {
		t.Fatalf("create the other tenant's item: %v", err)
	}

	receivedOrder(t, ctx, orders, mine, []procurement.LineInput{
		{SKU: "SHARED-SKU", Description: "Mine", Quantity: 9, UnitPrice: 1000},
	})

	after, err := stock.Get(ctx, theirs, theirItem.ID)
	if err != nil {
		t.Fatalf("get the other tenant's item: %v", err)
	}
	if after.Quantity != 100 {
		t.Fatalf("another tenant's stock moved: got %d, want 100", after.Quantity)
	}
}

func TestConfirmingAnOrderDoesNotMoveStock(t *testing.T) {
	app := testdb.App(t)
	ctx := context.Background()
	tenantID := testdb.Tenant(t, app.DB)

	stock, _ := inventory.ServiceFromApp(app)
	orders, _ := procurement.ServiceFromApp(app)

	order := draftOrder(t, ctx, orders, tenantID, []procurement.LineInput{
		{SKU: "NOT-YET", Description: "Not yet", Quantity: 5, UnitPrice: 1000},
	})
	if _, err := orders.UpdatePurchaseStatus(ctx, tenantID, order.ID, procurement.PurchaseStatusConfirmed); err != nil {
		t.Fatalf("confirm: %v", err)
	}

	// Confirming is a commitment to buy, not a delivery.
	items, err := stock.List(ctx, tenantID, inventory.ListFilter{})
	if err != nil {
		t.Fatalf("list stock: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("confirming must not move stock, got %+v", items)
	}
}

// ------------------------------------------------------------------ helpers

func draftOrder(
	t *testing.T,
	ctx context.Context,
	orders *procurement.Service,
	tenantID uuid.UUID,
	lines []procurement.LineInput,
) *procurement.PurchaseOrder {
	t.Helper()

	vendor, err := orders.CreateVendor(ctx, tenantID, procurement.VendorInput{
		Name: "Integration Vendor",
	})
	if err != nil {
		t.Fatalf("create vendor: %v", err)
	}

	order, err := orders.CreatePurchaseOrder(ctx, tenantID, procurement.CreatePurchaseInput{
		VendorID: vendor.ID, Lines: lines,
	})
	if err != nil {
		t.Fatalf("create purchase order: %v", err)
	}
	return order
}

func receivedOrder(
	t *testing.T,
	ctx context.Context,
	orders *procurement.Service,
	tenantID uuid.UUID,
	lines []procurement.LineInput,
) *procurement.PurchaseOrder {
	t.Helper()

	order := draftOrder(t, ctx, orders, tenantID, lines)
	if _, err := orders.UpdatePurchaseStatus(ctx, tenantID, order.ID, procurement.PurchaseStatusConfirmed); err != nil {
		t.Fatalf("confirm: %v", err)
	}
	received, err := orders.UpdatePurchaseStatus(ctx, tenantID, order.ID, procurement.PurchaseStatusReceived)
	if err != nil {
		t.Fatalf("receive: %v", err)
	}
	return received
}
