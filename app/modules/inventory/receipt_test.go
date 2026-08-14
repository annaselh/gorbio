package inventory

import (
	"errors"
	"testing"
)

func TestTotalBySKUSumsRepeatedLines(t *testing.T) {
	totals, order, err := totalBySKU([]Receipt{
		{SKU: "widget-1", Description: "Widget", Quantity: 4},
		{SKU: "BOLT-9", Description: "Bolt", Quantity: 2},
		{SKU: "Widget-1", Description: "Widget again", Quantity: 6},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// One order can name the same SKU on two lines; the receipt is their sum,
	// and case differences are the same SKU because Create upper-cases too.
	if got := totals["WIDGET-1"].Quantity; got != 10 {
		t.Fatalf("repeated lines should sum to 10, got %d", got)
	}
	if got := totals["BOLT-9"].Quantity; got != 2 {
		t.Fatalf("BOLT-9: got %d, want 2", got)
	}
	if len(order) != 2 {
		t.Fatalf("two distinct SKUs, got %d", len(order))
	}
}

func TestTotalBySKULocksInFirstSeenOrder(t *testing.T) {
	// The lock order has to be a function of the lines, not of map iteration,
	// or two concurrent receipts over the same SKUs could deadlock.
	_, order, err := totalBySKU([]Receipt{
		{SKU: "C", Quantity: 1},
		{SKU: "A", Quantity: 1},
		{SKU: "C", Quantity: 1},
		{SKU: "B", Quantity: 1},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for i, want := range []string{"C", "A", "B"} {
		if order[i] != want {
			t.Fatalf("lock order[%d]: got %s, want %s", i, order[i], want)
		}
	}
}

func TestTotalBySKUDefaultsTheNewItemFields(t *testing.T) {
	totals, _, err := totalBySKU([]Receipt{
		{SKU: "NEW-1", Description: "  Fresh item  ", Quantity: 3},
		{SKU: "NEW-2", Quantity: 1},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if totals["NEW-1"].Name != "Fresh item" {
		t.Fatalf("name should come from the trimmed description, got %q", totals["NEW-1"].Name)
	}
	if totals["NEW-1"].Unit != "pcs" {
		t.Fatalf("unit should default to pcs, got %q", totals["NEW-1"].Unit)
	}
	// A line with no description still has to produce a nameable item.
	if totals["NEW-2"].Name != "NEW-2" {
		t.Fatalf("a blank description should fall back to the SKU, got %q", totals["NEW-2"].Name)
	}
}

func TestTotalBySKURejectsUnusableLines(t *testing.T) {
	for name, receipts := range map[string][]Receipt{
		"blank sku":        {{SKU: "   ", Quantity: 1}},
		"zero quantity":    {{SKU: "A", Quantity: 0}},
		"negative arrival": {{SKU: "A", Quantity: -5}},
	} {
		if _, _, err := totalBySKU(receipts); !errors.Is(err, ErrInvalidInput) {
			t.Fatalf("%s: got %v, want ErrInvalidInput", name, err)
		}
	}
}

func TestTotalBySKURejectsANegativeLineHiddenByASum(t *testing.T) {
	// Summing first must not let a negative line pass because a positive one
	// on the same SKU covers it - that would be a withdrawal wearing a
	// receipt's clothes.
	if _, _, err := totalBySKU([]Receipt{
		{SKU: "A", Quantity: 10},
		{SKU: "A", Quantity: -4},
	}); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("got %v, want ErrInvalidInput", err)
	}
}
