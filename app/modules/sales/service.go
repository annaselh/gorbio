package sales

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const ServiceName = "sales.orders"

var (
	ErrNotFound      = errors.New("sales order not found")
	ErrInvalidInput  = errors.New("invalid sales order")
	ErrNotEditable   = errors.New("only draft orders can be modified")
	ErrDuplicate     = errors.New("a sales order with this number already exists")
	errMissingTenant = errors.New("tenant scope is required")
)

type Service struct {
	db *gorm.DB
	// resolveCustomer is nil when no CRM module is installed; Create then falls
	// back to the free-text customer name.
	resolveCustomer CustomerResolver
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// WithCustomerResolver attaches a CRM lookup. Called during wiring, before any
// request is served.
func (s *Service) WithCustomerResolver(resolver CustomerResolver) *Service {
	s.resolveCustomer = resolver
	return s
}

type LineInput struct {
	SKU         string
	Description string
	Quantity    int
	UnitPrice   int64
}

type CreateOrderInput struct {
	// Number is optional; a sequential SO-XXXXXX is generated when blank.
	Number string
	// CustomerID is optional. When set, the name is resolved from the CRM
	// record so the two cannot disagree; when absent, CustomerName is taken
	// verbatim, which keeps walk-in sales possible without a CRM entry.
	CustomerID   *uuid.UUID
	CustomerName string
	OrderDate    time.Time
	Currency     string
	Notes        string
	Lines        []LineInput
}

// CustomerResolver lets the sales module accept a CRM customer id without
// importing the CRM module. Whoever wires the app supplies the lookup, so the
// two modules stay independent and sales still works with CRM uninstalled.
type CustomerResolver func(ctx context.Context, tenantID, customerID uuid.UUID) (name string, err error)

type ListFilter struct {
	Status   OrderStatus
	Customer string
	Limit    int
	Offset   int
}

type ListResult struct {
	Orders []Order
	Total  int64
}

func (s *Service) scoped(ctx context.Context, tenantID uuid.UUID) (*gorm.DB, error) {
	if tenantID == uuid.Nil {
		return nil, errMissingTenant
	}
	return s.db.WithContext(ctx).Where("tenant_id = ?", tenantID), nil
}

// List returns a page of orders plus the unpaginated count, which the table UI
// needs to render pagination without a second round trip.
func (s *Service) List(ctx context.Context, tenantID uuid.UUID, filter ListFilter) (*ListResult, error) {
	query, err := s.scoped(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	query = query.Model(&Order{})

	if filter.Status != "" {
		if !filter.Status.Valid() {
			return nil, fmt.Errorf("%w: unknown status %q", ErrInvalidInput, filter.Status)
		}
		query = query.Where("status = ?", filter.Status)
	}
	if customer := strings.TrimSpace(filter.Customer); customer != "" {
		query = query.Where("customer_name ILIKE ?", "%"+customer+"%")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("count sales orders: %w", err)
	}

	limit := filter.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	var orders []Order
	if err := query.
		Order("order_date DESC, number DESC").
		Limit(limit).Offset(offset).
		Find(&orders).Error; err != nil {
		return nil, fmt.Errorf("list sales orders: %w", err)
	}

	return &ListResult{Orders: orders, Total: total}, nil
}

// Get returns one order with its lines.
func (s *Service) Get(ctx context.Context, tenantID, id uuid.UUID) (*Order, error) {
	query, err := s.scoped(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	var order Order
	if err := query.Preload("Lines", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at ASC")
	}).Where("id = ?", id).First(&order).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get sales order: %w", err)
	}
	return &order, nil
}

func (s *Service) Create(ctx context.Context, tenantID uuid.UUID, input CreateOrderInput) (*Order, error) {
	if tenantID == uuid.Nil {
		return nil, errMissingTenant
	}

	customer := strings.TrimSpace(input.CustomerName)
	// A linked customer wins over any name the client sent: the CRM record is
	// the authority, and letting the two diverge would put a name on the order
	// that no longer matches the account it points at.
	if input.CustomerID != nil {
		if s.resolveCustomer == nil {
			return nil, fmt.Errorf("%w: customer records are not available", ErrInvalidInput)
		}
		resolved, err := s.resolveCustomer(ctx, tenantID, *input.CustomerID)
		if err != nil {
			return nil, err
		}
		customer = resolved
	}
	if customer == "" {
		return nil, fmt.Errorf("%w: customer name is required", ErrInvalidInput)
	}
	if len(input.Lines) == 0 {
		return nil, fmt.Errorf("%w: at least one line is required", ErrInvalidInput)
	}

	orderID := uuid.New()
	lines := make([]OrderLine, 0, len(input.Lines))
	for index, line := range input.Lines {
		sku := strings.ToUpper(strings.TrimSpace(line.SKU))
		description := strings.TrimSpace(line.Description)
		if sku == "" || description == "" {
			return nil, fmt.Errorf("%w: line %d needs a SKU and description", ErrInvalidInput, index+1)
		}
		if line.Quantity <= 0 {
			return nil, fmt.Errorf("%w: line %d quantity must be positive", ErrInvalidInput, index+1)
		}
		if line.UnitPrice < 0 {
			return nil, fmt.Errorf("%w: line %d unit price must not be negative", ErrInvalidInput, index+1)
		}
		lines = append(lines, OrderLine{
			ID: uuid.New(), OrderID: orderID, TenantID: tenantID,
			SKU: sku, Description: description,
			Quantity: line.Quantity, UnitPrice: line.UnitPrice,
		})
	}

	orderDate := input.OrderDate
	if orderDate.IsZero() {
		orderDate = time.Now().UTC()
	}
	currency := strings.ToUpper(strings.TrimSpace(input.Currency))
	if currency == "" {
		currency = "IDR"
	}

	order := Order{
		ID: orderID, TenantID: tenantID,
		CustomerID: input.CustomerID, CustomerName: customer,
		Status: OrderStatusDraft, OrderDate: orderDate, Currency: currency,
		Notes: strings.TrimSpace(input.Notes), Lines: lines,
	}
	order.Recalculate()

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		number := strings.ToUpper(strings.TrimSpace(input.Number))
		if number == "" {
			generated, err := nextOrderNumber(tx, tenantID)
			if err != nil {
				return err
			}
			number = generated
		}
		order.Number = number

		if err := tx.Create(&order).Error; err != nil {
			if isUniqueViolation(err) {
				return ErrDuplicate
			}
			return fmt.Errorf("create sales order: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// UpdateStatus moves an order through its lifecycle. Draft may be confirmed or
// cancelled; a confirmed order may still be cancelled; a cancelled order is
// terminal.
func (s *Service) UpdateStatus(ctx context.Context, tenantID, id uuid.UUID, status OrderStatus) (*Order, error) {
	if tenantID == uuid.Nil {
		return nil, errMissingTenant
	}
	if !status.Valid() {
		return nil, fmt.Errorf("%w: unknown status %q", ErrInvalidInput, status)
	}

	var order Order
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("tenant_id = ? AND id = ?", tenantID, id).First(&order).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrNotFound
			}
			return fmt.Errorf("load sales order: %w", err)
		}

		if order.Status == OrderStatusCancelled {
			return fmt.Errorf("%w: a cancelled order is final", ErrNotEditable)
		}
		if order.Status == OrderStatusConfirmed && status == OrderStatusDraft {
			return fmt.Errorf("%w: a confirmed order cannot return to draft", ErrNotEditable)
		}

		if err := tx.Model(&order).Update("status", status).Error; err != nil {
			return fmt.Errorf("update sales order status: %w", err)
		}
		order.Status = status
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// ApplyDiscount sets an absolute discount in minor units. It is exported for
// the sales-discount extension, which owns the HTTP surface for it - the module
// itself has no opinion on how a discount is decided.
func (s *Service) ApplyDiscount(ctx context.Context, tenantID, id uuid.UUID, discount int64) (*Order, error) {
	if tenantID == uuid.Nil {
		return nil, errMissingTenant
	}
	if discount < 0 {
		return nil, fmt.Errorf("%w: discount must not be negative", ErrInvalidInput)
	}

	var order Order
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("tenant_id = ? AND id = ?", tenantID, id).First(&order).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrNotFound
			}
			return fmt.Errorf("load sales order: %w", err)
		}
		if !order.Editable() {
			return fmt.Errorf("%w: order is %s", ErrNotEditable, order.Status)
		}

		if err := tx.Where("order_id = ?", order.ID).Find(&order.Lines).Error; err != nil {
			return fmt.Errorf("load order lines: %w", err)
		}

		order.DiscountTotal = discount
		order.Recalculate()

		if err := tx.Model(&order).Updates(map[string]any{
			"discount_total": order.DiscountTotal,
			"subtotal":       order.Subtotal,
			"total":          order.Total,
		}).Error; err != nil {
			return fmt.Errorf("apply discount: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// nextOrderNumber allocates SO-000001, SO-000002, ... per tenant. It reads the
// current maximum inside the caller's transaction; the unique index on
// (tenant_id, number) is what actually guarantees no duplicate survives a race.
func nextOrderNumber(tx *gorm.DB, tenantID uuid.UUID) (string, error) {
	var latest string
	err := tx.Model(&Order{}).
		Where("tenant_id = ? AND number LIKE 'SO-%'", tenantID).
		Order("number DESC").Limit(1).
		Pluck("number", &latest).Error
	if err != nil {
		return "", fmt.Errorf("read latest order number: %w", err)
	}

	sequence := 1
	if latest != "" {
		var parsed int
		if _, err := fmt.Sscanf(latest, "SO-%d", &parsed); err == nil {
			sequence = parsed + 1
		}
	}
	return fmt.Sprintf("SO-%06d", sequence), nil
}

func isUniqueViolation(err error) bool {
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate key") ||
		strings.Contains(msg, "23505") ||
		strings.Contains(msg, "unique constraint")
}
