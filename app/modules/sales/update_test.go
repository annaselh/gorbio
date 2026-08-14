package sales

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

// Update validates before it opens a transaction, so every case here is
// reachable with a nil database handle - the same property the create tests
// rely on.

func TestUpdateRejectsMissingTenant(t *testing.T) {
	service := NewService(nil)

	_, err := service.Update(context.Background(), uuid.Nil, uuid.New(), UpdateOrderInput{
		CustomerName: "Walk-in",
		Lines:        []LineInput{{SKU: "A", Description: "A", Quantity: 1, UnitPrice: 1}},
	})

	if err == nil {
		t.Fatal("a nil tenant must be refused before any query runs")
	}
}

func TestUpdateRejectsEmptyLines(t *testing.T) {
	service := NewService(nil)

	// An edit that removes every line would leave an order for nothing. The
	// way to cancel an order is the status, not an empty body.
	_, err := service.Update(context.Background(), uuid.New(), uuid.New(), UpdateOrderInput{
		CustomerName: "Walk-in",
	})

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestUpdateValidatesLinesLikeCreate(t *testing.T) {
	service := NewService(nil)
	cases := map[string]LineInput{
		"blank sku":         {SKU: "  ", Description: "A", Quantity: 1, UnitPrice: 1},
		"blank description": {SKU: "A", Description: " ", Quantity: 1, UnitPrice: 1},
		"zero quantity":     {SKU: "A", Description: "A", Quantity: 0, UnitPrice: 1},
		"negative price":    {SKU: "A", Description: "A", Quantity: 1, UnitPrice: -1},
	}

	for name, line := range cases {
		_, err := service.Update(context.Background(), uuid.New(), uuid.New(), UpdateOrderInput{
			CustomerName: "Walk-in", Lines: []LineInput{line},
		})
		if !errors.Is(err, ErrInvalidInput) {
			t.Fatalf("%s: expected ErrInvalidInput, got %v", name, err)
		}
	}
}

func TestUpdateResolvesTheLinkedCustomerName(t *testing.T) {
	service := NewService(nil).WithCustomerResolver(
		func(context.Context, uuid.UUID, uuid.UUID) (string, error) {
			return "Resolved Name", nil
		},
	)
	customerID := uuid.New()

	// The name the client sent is ignored in favour of the CRM record, exactly
	// as on create; the assertion is on customerName because Update itself
	// needs a database to get any further.
	name, err := service.customerName(context.Background(), uuid.New(), &customerID, "Whatever The Client Typed")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "Resolved Name" {
		t.Fatalf("the CRM record must win, got %q", name)
	}
}

func TestUpdateRejectsLinkedCustomerWithoutResolver(t *testing.T) {
	service := NewService(nil)
	customerID := uuid.New()

	_, err := service.Update(context.Background(), uuid.New(), uuid.New(), UpdateOrderInput{
		CustomerID: &customerID,
		Lines:      []LineInput{{SKU: "A", Description: "A", Quantity: 1, UnitPrice: 1}},
	})

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestBuildLinesAttachesEveryLineToTheOrderAndTenant(t *testing.T) {
	tenantID, orderID := uuid.New(), uuid.New()

	lines, err := buildLines(tenantID, orderID, []LineInput{
		{SKU: "  widget-1 ", Description: " Widget ", Quantity: 2, UnitPrice: 500},
		{SKU: "bolt-9", Description: "Bolt", Quantity: 1, UnitPrice: 250},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for i, line := range lines {
		// The tenant on the line is what stops a query that forgets to join
		// through the order from reading another tenant's rows.
		if line.TenantID != tenantID || line.OrderID != orderID {
			t.Fatalf("line %d is not attached to the order and tenant", i)
		}
		if line.ID == uuid.Nil {
			t.Fatalf("line %d has no id", i)
		}
	}
	if lines[0].SKU != "WIDGET-1" {
		t.Fatalf("sku should be trimmed and upper-cased, got %q", lines[0].SKU)
	}
	if lines[0].Description != "Widget" {
		t.Fatalf("description should be trimmed, got %q", lines[0].Description)
	}
}

func TestUpdateKeepsADiscountClampedToTheNewSubtotal(t *testing.T) {
	// Editing the lines down must not leave a discount larger than the order it
	// discounts. Recalculate owns that rule; this pins the case Update relies
	// on, where the extension's discount outlives the lines it was sized for.
	order := &Order{
		DiscountTotal: 90_000,
		Lines:         []OrderLine{{Quantity: 1, UnitPrice: 40_000}},
	}

	order.Recalculate()

	if order.DiscountTotal != 40_000 {
		t.Fatalf("discount should clamp to the new subtotal, got %d", order.DiscountTotal)
	}
	if order.Total != 0 {
		t.Fatalf("total should floor at zero, got %d", order.Total)
	}
}

func TestOrderDefaultsFillInWhatTheClientOmitted(t *testing.T) {
	if got := currencyOrDefault("  "); got != "IDR" {
		t.Fatalf("blank currency should default to IDR, got %q", got)
	}
	if got := currencyOrDefault(" usd "); got != "USD" {
		t.Fatalf("currency should be trimmed and upper-cased, got %q", got)
	}

	if orderDateOrNow(time.Time{}).IsZero() {
		t.Fatal("a zero order date should become now, not stay zero")
	}
	stated := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	if !orderDateOrNow(stated).Equal(stated) {
		t.Fatal("a stated order date must be kept")
	}
}
