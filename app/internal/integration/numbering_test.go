package integration_test

import (
	"context"
	"sync"
	"testing"

	"github.com/annaselh/gorbio/internal/testdb"
	"github.com/annaselh/gorbio/modules/procurement"
	"github.com/annaselh/gorbio/modules/sales"
	"github.com/google/uuid"
)

// Sequential numbering reads the current maximum and adds one. The unique index
// is what actually guarantees no duplicate survives a race, and neither half of
// that arrangement exists without a database.

func TestOrderNumbersStartAtOneForEachTenant(t *testing.T) {
	app := testdb.App(t)
	ctx := context.Background()
	orders, _ := sales.ServiceFromApp(app)

	// Two tenants each get SO-000001. A global sequence would leak the size of
	// one tenant's ledger to another.
	first := testdb.Tenant(t, app.DB)
	second := testdb.Tenant(t, app.DB)

	for _, tenantID := range []uuid.UUID{first, second} {
		order, err := orders.Create(ctx, tenantID, sales.CreateOrderInput{
			CustomerName: "Walk-in",
			Lines:        []sales.LineInput{{SKU: "A", Description: "A", Quantity: 1, UnitPrice: 1000}},
		})
		if err != nil {
			t.Fatalf("create order: %v", err)
		}
		if order.Number != "SO-000001" {
			t.Fatalf("first order for a tenant: got %s, want SO-000001", order.Number)
		}
	}
}

func TestOrderNumbersIncrementWithinATenant(t *testing.T) {
	app := testdb.App(t)
	ctx := context.Background()
	tenantID := testdb.Tenant(t, app.DB)
	orders, _ := sales.ServiceFromApp(app)

	want := []string{"SO-000001", "SO-000002", "SO-000003"}
	for i, expected := range want {
		order, err := orders.Create(ctx, tenantID, sales.CreateOrderInput{
			CustomerName: "Walk-in",
			Lines:        []sales.LineInput{{SKU: "A", Description: "A", Quantity: 1, UnitPrice: 1000}},
		})
		if err != nil {
			t.Fatalf("create order %d: %v", i+1, err)
		}
		if order.Number != expected {
			t.Fatalf("order %d: got %s, want %s", i+1, order.Number, expected)
		}
	}
}

func TestPurchaseAndSalesSequencesAreIndependent(t *testing.T) {
	app := testdb.App(t)
	ctx := context.Background()
	tenantID := testdb.Tenant(t, app.DB)

	orders, _ := sales.ServiceFromApp(app)
	purchases, _ := procurement.ServiceFromApp(app)

	if _, err := orders.Create(ctx, tenantID, sales.CreateOrderInput{
		CustomerName: "Walk-in",
		Lines:        []sales.LineInput{{SKU: "A", Description: "A", Quantity: 1, UnitPrice: 1000}},
	}); err != nil {
		t.Fatalf("create sales order: %v", err)
	}

	vendor, err := purchases.CreateVendor(ctx, tenantID, procurement.VendorInput{Name: "Vendor"})
	if err != nil {
		t.Fatalf("create vendor: %v", err)
	}
	// The vendor sequence is its own too, so the first vendor is V-0001.
	if vendor.Code != "V-0001" {
		t.Fatalf("vendor code: got %s, want V-0001", vendor.Code)
	}

	order, err := purchases.CreatePurchaseOrder(ctx, tenantID, procurement.CreatePurchaseInput{
		VendorID: vendor.ID,
		Lines:    []procurement.LineInput{{SKU: "A", Description: "A", Quantity: 1, UnitPrice: 1000}},
	})
	if err != nil {
		t.Fatalf("create purchase order: %v", err)
	}
	if order.Number != "PO-000001" {
		t.Fatalf("purchase number: got %s, want PO-000001 - a sales order must not advance it", order.Number)
	}
}

// TestConcurrentOrdersNeverShareANumber is the case the unique index exists
// for. Two transactions can read the same maximum before either commits; what
// must never happen is both keeping the number.
func TestConcurrentOrdersNeverShareANumber(t *testing.T) {
	app := testdb.App(t)
	ctx := context.Background()
	tenantID := testdb.Tenant(t, app.DB)
	orders, _ := sales.ServiceFromApp(app)

	const attempts = 8

	var wg sync.WaitGroup
	var mu sync.Mutex
	numbers := make([]string, 0, attempts)
	failures := 0

	for range attempts {
		wg.Add(1)
		go func() {
			defer wg.Done()
			order, err := orders.Create(ctx, tenantID, sales.CreateOrderInput{
				CustomerName: "Walk-in",
				Lines:        []sales.LineInput{{SKU: "A", Description: "A", Quantity: 1, UnitPrice: 1000}},
			})

			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				// A loser of the race is refused, not silently given a
				// duplicate. That is the acceptable outcome; a duplicate is
				// not.
				failures++
				return
			}
			numbers = append(numbers, order.Number)
		}()
	}
	wg.Wait()

	seen := make(map[string]bool, len(numbers))
	for _, number := range numbers {
		if seen[number] {
			t.Fatalf("two orders were given the number %s", number)
		}
		seen[number] = true
	}

	if len(numbers) == 0 {
		t.Fatal("every concurrent create failed; the sequence is unusable under contention")
	}
	// Reported rather than asserted: how many collide depends on timing, and
	// pinning it would make this test flaky for no gain. It is worth seeing,
	// because a high number means callers need a retry they do not have.
	t.Logf("%d of %d concurrent creates succeeded, %d were refused as duplicates",
		len(numbers), attempts, failures)
}
