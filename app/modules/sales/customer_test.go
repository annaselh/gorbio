package sales

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
)

// Create's customer handling is worth testing without a database: the branch
// that decides whether the CRM record or the client's free text wins is the
// part that can silently put the wrong name on an order.

func TestCreateRejectsLinkedCustomerWithoutResolver(t *testing.T) {
	service := NewService(nil)
	customerID := uuid.New()

	_, err := service.Create(context.Background(), uuid.New(), CreateOrderInput{
		CustomerID: &customerID,
		Lines:      []LineInput{{SKU: "A", Description: "A", Quantity: 1, UnitPrice: 1}},
	})

	if err == nil {
		t.Fatal("a customer id with no CRM module installed must be refused")
	}
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestCreatePropagatesResolverFailure(t *testing.T) {
	sentinel := errors.New("customer not found")
	service := NewService(nil).WithCustomerResolver(
		func(context.Context, uuid.UUID, uuid.UUID) (string, error) {
			return "", sentinel
		},
	)
	customerID := uuid.New()

	_, err := service.Create(context.Background(), uuid.New(), CreateOrderInput{
		CustomerID: &customerID,
		Lines:      []LineInput{{SKU: "A", Description: "A", Quantity: 1, UnitPrice: 1}},
	})

	if !errors.Is(err, sentinel) {
		t.Fatalf("resolver failure should surface to the caller, got %v", err)
	}
}

// A blank name with no linked customer is the one case that must fail: an order
// has to say who bought.
func TestCreateRequiresACustomer(t *testing.T) {
	service := NewService(nil)

	_, err := service.Create(context.Background(), uuid.New(), CreateOrderInput{
		CustomerName: "   ",
		Lines:        []LineInput{{SKU: "A", Description: "A", Quantity: 1, UnitPrice: 1}},
	})

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestCreateRejectsEmptyLines(t *testing.T) {
	service := NewService(nil)

	_, err := service.Create(context.Background(), uuid.New(), CreateOrderInput{
		CustomerName: "Walk-in",
	})

	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("an order with no lines must be refused, got %v", err)
	}
}

func TestCreateRejectsMissingTenant(t *testing.T) {
	service := NewService(nil)

	_, err := service.Create(context.Background(), uuid.Nil, CreateOrderInput{
		CustomerName: "Walk-in",
		Lines:        []LineInput{{SKU: "A", Description: "A", Quantity: 1, UnitPrice: 1}},
	})

	if err == nil {
		t.Fatal("a nil tenant must be refused before any query runs")
	}
}

// Line validation runs before the database is touched, so these cases are
// reachable with a nil handle.
func TestCreateValidatesLines(t *testing.T) {
	service := NewService(nil)
	cases := map[string]LineInput{
		"blank sku":         {SKU: "  ", Description: "A", Quantity: 1, UnitPrice: 1},
		"blank description": {SKU: "A", Description: " ", Quantity: 1, UnitPrice: 1},
		"zero quantity":     {SKU: "A", Description: "A", Quantity: 0, UnitPrice: 1},
		"negative quantity": {SKU: "A", Description: "A", Quantity: -2, UnitPrice: 1},
		"negative price":    {SKU: "A", Description: "A", Quantity: 1, UnitPrice: -1},
	}

	for name, line := range cases {
		_, err := service.Create(context.Background(), uuid.New(), CreateOrderInput{
			CustomerName: "Walk-in", Lines: []LineInput{line},
		})
		if !errors.Is(err, ErrInvalidInput) {
			t.Fatalf("%s: expected ErrInvalidInput, got %v", name, err)
		}
	}
}
