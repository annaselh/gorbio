package sales

import "testing"

func orderWithLines(lines ...OrderLine) *Order {
	return &Order{Lines: lines}
}

func TestRecalculateDerivesLineAndOrderTotals(t *testing.T) {
	order := orderWithLines(
		OrderLine{Quantity: 3, UnitPrice: 150_000},
		OrderLine{Quantity: 2, UnitPrice: 25_000},
	)

	order.Recalculate()

	if order.Lines[0].LineTotal != 450_000 {
		t.Fatalf("first line total: got %d, want 450000", order.Lines[0].LineTotal)
	}
	if order.Subtotal != 500_000 {
		t.Fatalf("subtotal: got %d, want 500000", order.Subtotal)
	}
	if order.Total != 500_000 {
		t.Fatalf("total without discount should equal subtotal, got %d", order.Total)
	}
}

func TestRecalculateAppliesDiscount(t *testing.T) {
	order := orderWithLines(OrderLine{Quantity: 1, UnitPrice: 100_000})
	order.DiscountTotal = 30_000

	order.Recalculate()

	if order.Total != 70_000 {
		t.Fatalf("total: got %d, want 70000", order.Total)
	}
}

// A discount larger than the order must not produce a negative total - the
// customer is never owed money by placing an order.
func TestRecalculateClampsExcessiveDiscount(t *testing.T) {
	order := orderWithLines(OrderLine{Quantity: 1, UnitPrice: 50_000})
	order.DiscountTotal = 80_000

	order.Recalculate()

	if order.DiscountTotal != 50_000 {
		t.Fatalf("discount should clamp to the subtotal, got %d", order.DiscountTotal)
	}
	if order.Total != 0 {
		t.Fatalf("total should floor at zero, got %d", order.Total)
	}
}

func TestRecalculateClampsNegativeDiscount(t *testing.T) {
	order := orderWithLines(OrderLine{Quantity: 1, UnitPrice: 50_000})
	order.DiscountTotal = -10_000

	order.Recalculate()

	if order.DiscountTotal != 0 {
		t.Fatalf("a negative discount must clamp to zero, got %d", order.DiscountTotal)
	}
	if order.Total != 50_000 {
		t.Fatalf("total: got %d, want 50000", order.Total)
	}
}

func TestRecalculateIsIdempotent(t *testing.T) {
	order := orderWithLines(OrderLine{Quantity: 2, UnitPrice: 75_000})

	order.Recalculate()
	first := *order
	order.Recalculate()

	if order.Subtotal != first.Subtotal || order.Total != first.Total {
		t.Fatal("recalculating twice must not change the result")
	}
}

func TestRecalculateEmptyOrder(t *testing.T) {
	order := orderWithLines()
	order.Recalculate()

	if order.Subtotal != 0 || order.Total != 0 {
		t.Fatalf("an order with no lines costs nothing, got subtotal %d total %d",
			order.Subtotal, order.Total)
	}
}

// Money is held in minor units precisely so totals stay exact; this guards the
// int64 choice against a future refactor to float.
func TestRecalculateKeepsExactTotalsAcrossManyLines(t *testing.T) {
	lines := make([]OrderLine, 100)
	for i := range lines {
		lines[i] = OrderLine{Quantity: 1, UnitPrice: 10}
	}
	order := orderWithLines(lines...)

	order.Recalculate()

	if order.Subtotal != 1000 {
		t.Fatalf("100 lines of 10 must total exactly 1000, got %d", order.Subtotal)
	}
}

func TestOrderStatusValid(t *testing.T) {
	for _, status := range []OrderStatus{OrderStatusDraft, OrderStatusConfirmed, OrderStatusCancelled} {
		if !status.Valid() {
			t.Fatalf("%q should be a valid status", status)
		}
	}
	for _, status := range []OrderStatus{"", "Shipped", "draft"} {
		if status.Valid() {
			t.Fatalf("%q should not be accepted as a status", status)
		}
	}
}

func TestEditableOnlyInDraft(t *testing.T) {
	if !(&Order{Status: OrderStatusDraft}).Editable() {
		t.Fatal("a draft order must be editable")
	}
	for _, status := range []OrderStatus{OrderStatusConfirmed, OrderStatusCancelled} {
		if (&Order{Status: status}).Editable() {
			t.Fatalf("a %s order is a financial record and must not be editable", status)
		}
	}
}
