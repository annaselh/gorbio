package integration_test

import (
	"context"
	"errors"
	"testing"

	"github.com/annaselh/gorbio/internal/testdb"
	"github.com/annaselh/gorbio/modules/crm"
	"github.com/annaselh/gorbio/modules/procurement"
	"github.com/annaselh/gorbio/modules/sales"
)

// Editing an order replaces its lines, which means a delete and an insert in
// one transaction. Whether the old rows actually go, and whether the totals are
// rewritten to match, is not observable without a database.

func TestUpdateReplacesTheLinesAndTheTotals(t *testing.T) {
	app := testdb.App(t)
	ctx := context.Background()
	tenantID := testdb.Tenant(t, app.DB)
	orders, _ := sales.ServiceFromApp(app)

	order, err := orders.Create(ctx, tenantID, sales.CreateOrderInput{
		CustomerName: "Walk-in",
		Lines: []sales.LineInput{
			{SKU: "OLD-1", Description: "Old one", Quantity: 2, UnitPrice: 100_000},
			{SKU: "OLD-2", Description: "Old two", Quantity: 1, UnitPrice: 50_000},
		},
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if order.Total != 250_000 {
		t.Fatalf("total before edit: got %d, want 250000", order.Total)
	}

	updated, err := orders.Update(ctx, tenantID, order.ID, sales.UpdateOrderInput{
		CustomerName: "Walk-in",
		Lines: []sales.LineInput{
			{SKU: "NEW-1", Description: "New one", Quantity: 3, UnitPrice: 30_000},
		},
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if updated.Total != 90_000 {
		t.Fatalf("total after edit: got %d, want 90000", updated.Total)
	}

	// Re-read rather than trusting the returned struct: the point is what is
	// in the table.
	reloaded, err := orders.Get(ctx, tenantID, order.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if len(reloaded.Lines) != 1 {
		t.Fatalf("expected the old lines to be gone, got %d lines", len(reloaded.Lines))
	}
	if reloaded.Lines[0].SKU != "NEW-1" {
		t.Fatalf("line sku: got %s, want NEW-1", reloaded.Lines[0].SKU)
	}
	if reloaded.Number != order.Number {
		t.Fatalf("the number must not change on edit: %s became %s", order.Number, reloaded.Number)
	}
	if reloaded.Subtotal != 90_000 || reloaded.Total != 90_000 {
		t.Fatalf("stored totals were not rewritten: subtotal %d, total %d", reloaded.Subtotal, reloaded.Total)
	}
}

func TestAConfirmedOrderCannotBeEdited(t *testing.T) {
	app := testdb.App(t)
	ctx := context.Background()
	tenantID := testdb.Tenant(t, app.DB)
	orders, _ := sales.ServiceFromApp(app)

	order := confirmSalesOrder(t, ctx, orders, tenantID, sales.CreateOrderInput{
		CustomerName: "Walk-in",
		Lines:        []sales.LineInput{{SKU: "A", Description: "A", Quantity: 1, UnitPrice: 100_000}},
	})

	_, err := orders.Update(ctx, tenantID, order.ID, sales.UpdateOrderInput{
		CustomerName: "Walk-in",
		Lines:        []sales.LineInput{{SKU: "B", Description: "B", Quantity: 9, UnitPrice: 1}},
	})
	if !errors.Is(err, sales.ErrNotEditable) {
		t.Fatalf("expected ErrNotEditable, got %v", err)
	}

	// The refusal has to leave the order untouched, not half-edited.
	reloaded, err := orders.Get(ctx, tenantID, order.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if len(reloaded.Lines) != 1 || reloaded.Lines[0].SKU != "A" || reloaded.Total != 100_000 {
		t.Fatalf("a rejected edit modified the order: %+v", reloaded.Lines)
	}
}

func TestUpdateTakesTheCustomerNameFromTheCRMRecord(t *testing.T) {
	app := testdb.App(t)
	ctx := context.Background()
	tenantID := testdb.Tenant(t, app.DB)

	orders, _ := sales.ServiceFromApp(app)
	customers, _ := crm.ServiceFromApp(app)

	customer, err := customers.Create(ctx, tenantID, crm.CustomerInput{Name: "Authoritative Name"})
	if err != nil {
		t.Fatalf("create customer: %v", err)
	}

	order, err := orders.Create(ctx, tenantID, sales.CreateOrderInput{
		CustomerName: "Walk-in",
		Lines:        []sales.LineInput{{SKU: "A", Description: "A", Quantity: 1, UnitPrice: 1000}},
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	// Linking a walk-in order to a CRM record on edit: the name the client
	// sends is ignored in favour of the record it points at.
	updated, err := orders.Update(ctx, tenantID, order.ID, sales.UpdateOrderInput{
		CustomerID:   &customer.ID,
		CustomerName: "Something Else Entirely",
		Lines:        []sales.LineInput{{SKU: "A", Description: "A", Quantity: 1, UnitPrice: 1000}},
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if updated.CustomerName != "Authoritative Name" {
		t.Fatalf("customer name: got %q, want the CRM record's name", updated.CustomerName)
	}
	if updated.CustomerID == nil || *updated.CustomerID != customer.ID {
		t.Fatalf("the CRM link was not stored: %+v", updated.CustomerID)
	}
}

func TestAnInactiveCustomerCannotBeLinkedOnEdit(t *testing.T) {
	app := testdb.App(t)
	ctx := context.Background()
	tenantID := testdb.Tenant(t, app.DB)

	orders, _ := sales.ServiceFromApp(app)
	customers, _ := crm.ServiceFromApp(app)

	customer, err := customers.Create(ctx, tenantID, crm.CustomerInput{Name: "Retired Co"})
	if err != nil {
		t.Fatalf("create customer: %v", err)
	}
	if _, err := customers.SetStatus(ctx, tenantID, customer.ID, crm.CustomerStatusInactive); err != nil {
		t.Fatalf("deactivate: %v", err)
	}

	order, err := orders.Create(ctx, tenantID, sales.CreateOrderInput{
		CustomerName: "Walk-in",
		Lines:        []sales.LineInput{{SKU: "A", Description: "A", Quantity: 1, UnitPrice: 1000}},
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	if _, err := orders.Update(ctx, tenantID, order.ID, sales.UpdateOrderInput{
		CustomerID: &customer.ID,
		Lines:      []sales.LineInput{{SKU: "A", Description: "A", Quantity: 1, UnitPrice: 1000}},
	}); err == nil {
		t.Fatal("linking an order to an inactive customer must be refused")
	}
}

func TestAPurchaseOrderEditCannotMoveToAnInactiveVendor(t *testing.T) {
	app := testdb.App(t)
	ctx := context.Background()
	tenantID := testdb.Tenant(t, app.DB)
	purchases, _ := procurement.ServiceFromApp(app)

	active, err := purchases.CreateVendor(ctx, tenantID, procurement.VendorInput{Name: "Active Vendor"})
	if err != nil {
		t.Fatalf("create vendor: %v", err)
	}
	retired, err := purchases.CreateVendor(ctx, tenantID, procurement.VendorInput{Name: "Retired Vendor"})
	if err != nil {
		t.Fatalf("create vendor: %v", err)
	}
	if _, err := purchases.SetVendorStatus(ctx, tenantID, retired.ID, procurement.VendorStatusInactive); err != nil {
		t.Fatalf("deactivate vendor: %v", err)
	}

	order, err := purchases.CreatePurchaseOrder(ctx, tenantID, procurement.CreatePurchaseInput{
		VendorID: active.ID,
		Lines:    []procurement.LineInput{{SKU: "A", Description: "A", Quantity: 1, UnitPrice: 1000}},
	})
	if err != nil {
		t.Fatalf("create purchase order: %v", err)
	}

	// The same rule that stops a draft being raised against a switched-off
	// vendor has to apply when one is edited onto it.
	if _, err := purchases.UpdatePurchaseOrder(ctx, tenantID, order.ID, procurement.UpdatePurchaseInput{
		VendorID: retired.ID,
		Lines:    []procurement.LineInput{{SKU: "A", Description: "A", Quantity: 1, UnitPrice: 1000}},
	}); !errors.Is(err, procurement.ErrVendorInactive) {
		t.Fatalf("expected ErrVendorInactive, got %v", err)
	}
}

func TestAPurchaseOrderEditRedenormalisesTheVendorName(t *testing.T) {
	app := testdb.App(t)
	ctx := context.Background()
	tenantID := testdb.Tenant(t, app.DB)
	purchases, _ := procurement.ServiceFromApp(app)

	first, err := purchases.CreateVendor(ctx, tenantID, procurement.VendorInput{Name: "First Vendor"})
	if err != nil {
		t.Fatalf("create vendor: %v", err)
	}
	second, err := purchases.CreateVendor(ctx, tenantID, procurement.VendorInput{Name: "Second Vendor"})
	if err != nil {
		t.Fatalf("create vendor: %v", err)
	}

	order, err := purchases.CreatePurchaseOrder(ctx, tenantID, procurement.CreatePurchaseInput{
		VendorID: first.ID,
		Lines:    []procurement.LineInput{{SKU: "A", Description: "A", Quantity: 1, UnitPrice: 1000}},
	})
	if err != nil {
		t.Fatalf("create purchase order: %v", err)
	}

	updated, err := purchases.UpdatePurchaseOrder(ctx, tenantID, order.ID, procurement.UpdatePurchaseInput{
		VendorID: second.ID,
		Lines:    []procurement.LineInput{{SKU: "A", Description: "A", Quantity: 2, UnitPrice: 1000}},
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}

	// The denormalised name has to follow the link, or the order would name a
	// supplier it is no longer placed with.
	if updated.VendorName != "Second Vendor" {
		t.Fatalf("vendor name: got %q, want Second Vendor", updated.VendorName)
	}
	if updated.Total != 2000 {
		t.Fatalf("total: got %d, want 2000", updated.Total)
	}
}
