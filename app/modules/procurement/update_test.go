package procurement

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

// UpdatePurchaseOrder validates before it opens a transaction, so these cases
// are reachable with a nil database handle.

func TestUpdatePurchaseOrderRejectsMissingTenant(t *testing.T) {
	service := NewService(nil)

	_, err := service.UpdatePurchaseOrder(context.Background(), uuid.Nil, uuid.New(), UpdatePurchaseInput{
		VendorID: uuid.New(),
		Lines:    []LineInput{{SKU: "A", Description: "A", Quantity: 1, UnitPrice: 1}},
	})

	if err == nil {
		t.Fatal("a nil tenant must be refused before any query runs")
	}
}

func TestUpdatePurchaseOrderRequiresAVendor(t *testing.T) {
	service := NewService(nil)

	// An order has to say who is being bought from, on an edit as much as on
	// the original.
	_, err := service.UpdatePurchaseOrder(context.Background(), uuid.New(), uuid.New(), UpdatePurchaseInput{
		Lines: []LineInput{{SKU: "A", Description: "A", Quantity: 1, UnitPrice: 1}},
	})

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestUpdatePurchaseOrderRejectsEmptyLines(t *testing.T) {
	service := NewService(nil)

	_, err := service.UpdatePurchaseOrder(context.Background(), uuid.New(), uuid.New(), UpdatePurchaseInput{
		VendorID: uuid.New(),
	})

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestUpdatePurchaseOrderValidatesLines(t *testing.T) {
	service := NewService(nil)
	cases := map[string]LineInput{
		"blank sku":         {SKU: "  ", Description: "A", Quantity: 1, UnitPrice: 1},
		"blank description": {SKU: "A", Description: " ", Quantity: 1, UnitPrice: 1},
		"zero quantity":     {SKU: "A", Description: "A", Quantity: 0, UnitPrice: 1},
		"negative price":    {SKU: "A", Description: "A", Quantity: 1, UnitPrice: -1},
	}

	for name, line := range cases {
		_, err := service.UpdatePurchaseOrder(context.Background(), uuid.New(), uuid.New(), UpdatePurchaseInput{
			VendorID: uuid.New(), Lines: []LineInput{line},
		})
		if !errors.Is(err, ErrInvalidInput) {
			t.Fatalf("%s: expected ErrInvalidInput, got %v", name, err)
		}
	}
}

func TestBuildPurchaseLinesAttachesEveryLineToTheOrderAndTenant(t *testing.T) {
	tenantID, orderID := uuid.New(), uuid.New()

	lines, err := buildPurchaseLines(tenantID, orderID, []LineInput{
		{SKU: " widget-1 ", Description: " Widget ", Quantity: 2, UnitPrice: 500},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if lines[0].TenantID != tenantID || lines[0].PurchaseOrderID != orderID {
		t.Fatal("the line is not attached to the order and tenant")
	}
	if lines[0].SKU != "WIDGET-1" {
		t.Fatalf("sku should be trimmed and upper-cased, got %q", lines[0].SKU)
	}
	if lines[0].Description != "Widget" {
		t.Fatalf("description should be trimmed, got %q", lines[0].Description)
	}
}

func TestPurchaseDefaultsFillInWhatTheClientOmitted(t *testing.T) {
	if got := currencyOrDefault(" "); got != "IDR" {
		t.Fatalf("blank currency should default to IDR, got %q", got)
	}
	if orderDateOrNow(time.Time{}).IsZero() {
		t.Fatal("a zero order date should become now, not stay zero")
	}
}
