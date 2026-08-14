package procurement

import (
	"errors"
	"testing"
)

func TestRecalculateDerivesTotals(t *testing.T) {
	order := &PurchaseOrder{Lines: []PurchaseOrderLine{
		{Quantity: 4, UnitPrice: 250_000},
		{Quantity: 1, UnitPrice: 100_000},
	}}

	order.Recalculate()

	if order.Lines[0].LineTotal != 1_000_000 {
		t.Fatalf("line total: got %d, want 1000000", order.Lines[0].LineTotal)
	}
	if order.Subtotal != 1_100_000 || order.Total != 1_100_000 {
		t.Fatalf("subtotal %d / total %d, want 1100000 each", order.Subtotal, order.Total)
	}
}

func TestRecalculateClampsDiscount(t *testing.T) {
	order := &PurchaseOrder{Lines: []PurchaseOrderLine{{Quantity: 1, UnitPrice: 60_000}}}
	order.DiscountTotal = 90_000

	order.Recalculate()

	if order.DiscountTotal != 60_000 {
		t.Fatalf("discount should clamp to the subtotal, got %d", order.DiscountTotal)
	}
	if order.Total != 0 {
		t.Fatalf("total should floor at zero, got %d", order.Total)
	}
}

func TestEditableOnlyInDraft(t *testing.T) {
	if !(&PurchaseOrder{Status: PurchaseStatusDraft}).Editable() {
		t.Fatal("a draft order must be editable")
	}
	for _, status := range []PurchaseStatus{
		PurchaseStatusConfirmed, PurchaseStatusReceived, PurchaseStatusCancelled,
	} {
		if (&PurchaseOrder{Status: status}).Editable() {
			t.Fatalf("a %s order is a commitment and must not be editable", status)
		}
	}
}

func TestPurchaseStatusValid(t *testing.T) {
	for _, status := range []PurchaseStatus{
		PurchaseStatusDraft, PurchaseStatusConfirmed,
		PurchaseStatusReceived, PurchaseStatusCancelled,
	} {
		if !status.Valid() {
			t.Fatalf("%q should be valid", status)
		}
	}
	for _, status := range []PurchaseStatus{"", "Shipped", "draft"} {
		if status.Valid() {
			t.Fatalf("%q should not be accepted", status)
		}
	}
}

func TestVendorStatusValid(t *testing.T) {
	if !VendorStatusActive.Valid() || !VendorStatusInactive.Valid() {
		t.Fatal("both vendor statuses should be valid")
	}
	if VendorStatus("Archived").Valid() {
		t.Fatal("an unknown vendor status must be rejected")
	}
}

func TestValidPurchaseTransitions(t *testing.T) {
	allowed := []struct{ from, to PurchaseStatus }{
		{PurchaseStatusDraft, PurchaseStatusConfirmed},
		{PurchaseStatusDraft, PurchaseStatusCancelled},
		{PurchaseStatusConfirmed, PurchaseStatusReceived},
		{PurchaseStatusConfirmed, PurchaseStatusCancelled},
	}

	for _, tc := range allowed {
		if err := validatePurchaseTransition(tc.from, tc.to); err != nil {
			t.Fatalf("%s -> %s should be allowed, got %v", tc.from, tc.to, err)
		}
	}
}

func TestInvalidPurchaseTransitions(t *testing.T) {
	refused := []struct{ from, to PurchaseStatus }{
		{PurchaseStatusReceived, PurchaseStatusCancelled},
		{PurchaseStatusReceived, PurchaseStatusDraft},
		{PurchaseStatusCancelled, PurchaseStatusConfirmed},
		{PurchaseStatusConfirmed, PurchaseStatusDraft},
		{PurchaseStatusDraft, PurchaseStatusReceived},
	}

	for _, tc := range refused {
		err := validatePurchaseTransition(tc.from, tc.to)
		if err == nil {
			t.Fatalf("%s -> %s should be refused", tc.from, tc.to)
		}
		if !errors.Is(err, ErrNotEditable) {
			t.Fatalf("%s -> %s should report ErrNotEditable, got %v", tc.from, tc.to, err)
		}
	}
}

// Re-submitting the current status is a no-op, not an error: a client retrying
// a request should not get a failure for a change that already landed.
func TestPurchaseTransitionToSameStatusIsAllowed(t *testing.T) {
	for _, status := range []PurchaseStatus{
		PurchaseStatusDraft, PurchaseStatusConfirmed,
		PurchaseStatusReceived, PurchaseStatusCancelled,
	} {
		if err := validatePurchaseTransition(status, status); err != nil {
			t.Fatalf("%s -> %s should be a no-op, got %v", status, status, err)
		}
	}
}

func TestPageDefaultsAndClamps(t *testing.T) {
	if limit, offset := page(0, 0); limit != 50 || offset != 0 {
		t.Fatalf("defaults: got limit %d offset %d, want 50/0", limit, offset)
	}
	if limit, _ := page(500, 0); limit != 50 {
		t.Fatalf("an oversized limit should fall back to 50, got %d", limit)
	}
	if _, offset := page(10, -5); offset != 0 {
		t.Fatalf("a negative offset should clamp to 0, got %d", offset)
	}
}
