package procurement

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

const ServiceName = "procurement.orders"

var (
	ErrVendorNotFound   = errors.New("vendor not found")
	ErrPurchaseNotFound = errors.New("purchase order not found")
	ErrInvalidInput     = errors.New("invalid input")
	ErrNotEditable      = errors.New("only draft purchase orders can be modified")
	ErrDuplicate        = errors.New("a record with this code or number already exists")
	ErrVendorInactive   = errors.New("cannot order from an inactive vendor")
	errMissingTenant    = errors.New("tenant scope is required")
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// ---------------------------------------------------------------- vendors

type VendorInput struct {
	Code        string
	Name        string
	Email       string
	Phone       string
	Address     string
	TaxID       string
	PaymentTerm int
	Notes       string
}

type VendorFilter struct {
	Status VendorStatus
	Search string
	Limit  int
	Offset int
}

type VendorList struct {
	Vendors []Vendor
	Total   int64
}

func (s *Service) scoped(ctx context.Context, tenantID uuid.UUID) (*gorm.DB, error) {
	if tenantID == uuid.Nil {
		return nil, errMissingTenant
	}
	return s.db.WithContext(ctx).Where("tenant_id = ?", tenantID), nil
}

func (s *Service) ListVendors(ctx context.Context, tenantID uuid.UUID, filter VendorFilter) (*VendorList, error) {
	query, err := s.scoped(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	query = query.Model(&Vendor{})

	if filter.Status != "" {
		if !filter.Status.Valid() {
			return nil, fmt.Errorf("%w: unknown vendor status %q", ErrInvalidInput, filter.Status)
		}
		query = query.Where("status = ?", filter.Status)
	}
	if search := strings.TrimSpace(filter.Search); search != "" {
		query = query.Where("name ILIKE ? OR code ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("count vendors: %w", err)
	}

	limit, offset := page(filter.Limit, filter.Offset)
	var vendors []Vendor
	if err := query.Order("name ASC").Limit(limit).Offset(offset).Find(&vendors).Error; err != nil {
		return nil, fmt.Errorf("list vendors: %w", err)
	}
	return &VendorList{Vendors: vendors, Total: total}, nil
}

func (s *Service) GetVendor(ctx context.Context, tenantID, id uuid.UUID) (*Vendor, error) {
	query, err := s.scoped(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	var vendor Vendor
	if err := query.Where("id = ?", id).First(&vendor).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrVendorNotFound
		}
		return nil, fmt.Errorf("get vendor: %w", err)
	}
	return &vendor, nil
}

func (s *Service) CreateVendor(ctx context.Context, tenantID uuid.UUID, input VendorInput) (*Vendor, error) {
	if tenantID == uuid.Nil {
		return nil, errMissingTenant
	}
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, fmt.Errorf("%w: vendor name is required", ErrInvalidInput)
	}
	if input.PaymentTerm < 0 {
		return nil, fmt.Errorf("%w: payment term must not be negative", ErrInvalidInput)
	}

	paymentTerm := input.PaymentTerm
	if paymentTerm == 0 {
		paymentTerm = 30
	}

	vendor := Vendor{
		ID: uuid.New(), TenantID: tenantID, Name: name,
		Email: strings.ToLower(strings.TrimSpace(input.Email)),
		Phone: strings.TrimSpace(input.Phone), Address: strings.TrimSpace(input.Address),
		TaxID: strings.TrimSpace(input.TaxID), PaymentTerm: paymentTerm,
		Status: VendorStatusActive, Notes: strings.TrimSpace(input.Notes),
	}

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		code := strings.ToUpper(strings.TrimSpace(input.Code))
		if code == "" {
			generated, err := nextSequence(tx, tenantID, &Vendor{}, "code", "V-", 4)
			if err != nil {
				return err
			}
			code = generated
		}
		vendor.Code = code

		if err := tx.Create(&vendor).Error; err != nil {
			if isUniqueViolation(err) {
				return ErrDuplicate
			}
			return fmt.Errorf("create vendor: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &vendor, nil
}

func (s *Service) UpdateVendor(ctx context.Context, tenantID, id uuid.UUID, input VendorInput) (*Vendor, error) {
	vendor, err := s.GetVendor(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, fmt.Errorf("%w: vendor name is required", ErrInvalidInput)
	}
	if input.PaymentTerm < 0 {
		return nil, fmt.Errorf("%w: payment term must not be negative", ErrInvalidInput)
	}

	updates := map[string]any{
		"name": name, "email": strings.ToLower(strings.TrimSpace(input.Email)),
		"phone": strings.TrimSpace(input.Phone), "address": strings.TrimSpace(input.Address),
		"tax_id": strings.TrimSpace(input.TaxID), "notes": strings.TrimSpace(input.Notes),
	}
	if input.PaymentTerm > 0 {
		updates["payment_term"] = input.PaymentTerm
	}

	if err := s.db.WithContext(ctx).Model(vendor).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("update vendor: %w", err)
	}
	return s.GetVendor(ctx, tenantID, id)
}

// SetVendorStatus deactivates or reactivates a vendor. Existing purchase orders
// are untouched: they are historical records of what was actually ordered.
func (s *Service) SetVendorStatus(ctx context.Context, tenantID, id uuid.UUID, status VendorStatus) (*Vendor, error) {
	if !status.Valid() {
		return nil, fmt.Errorf("%w: unknown vendor status %q", ErrInvalidInput, status)
	}
	vendor, err := s.GetVendor(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	if err := s.db.WithContext(ctx).Model(vendor).Update("status", status).Error; err != nil {
		return nil, fmt.Errorf("update vendor status: %w", err)
	}
	vendor.Status = status
	return vendor, nil
}

// --------------------------------------------------------- purchase orders

type LineInput struct {
	SKU         string
	Description string
	Quantity    int
	UnitPrice   int64
}

type CreatePurchaseInput struct {
	Number       string
	VendorID     uuid.UUID
	OrderDate    time.Time
	ExpectedDate *time.Time
	Currency     string
	Notes        string
	Lines        []LineInput
}

type PurchaseFilter struct {
	Status   PurchaseStatus
	VendorID uuid.UUID
	Limit    int
	Offset   int
}

type PurchaseList struct {
	Orders []PurchaseOrder
	Total  int64
}

func (s *Service) ListPurchaseOrders(ctx context.Context, tenantID uuid.UUID, filter PurchaseFilter) (*PurchaseList, error) {
	query, err := s.scoped(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	query = query.Model(&PurchaseOrder{})

	if filter.Status != "" {
		if !filter.Status.Valid() {
			return nil, fmt.Errorf("%w: unknown purchase status %q", ErrInvalidInput, filter.Status)
		}
		query = query.Where("status = ?", filter.Status)
	}
	if filter.VendorID != uuid.Nil {
		query = query.Where("vendor_id = ?", filter.VendorID)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("count purchase orders: %w", err)
	}

	limit, offset := page(filter.Limit, filter.Offset)
	var orders []PurchaseOrder
	if err := query.Order("order_date DESC, number DESC").
		Limit(limit).Offset(offset).Find(&orders).Error; err != nil {
		return nil, fmt.Errorf("list purchase orders: %w", err)
	}
	return &PurchaseList{Orders: orders, Total: total}, nil
}

func (s *Service) GetPurchaseOrder(ctx context.Context, tenantID, id uuid.UUID) (*PurchaseOrder, error) {
	query, err := s.scoped(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	var order PurchaseOrder
	if err := query.Preload("Lines", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at ASC")
	}).Where("id = ?", id).First(&order).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPurchaseNotFound
		}
		return nil, fmt.Errorf("get purchase order: %w", err)
	}
	return &order, nil
}

func (s *Service) CreatePurchaseOrder(ctx context.Context, tenantID uuid.UUID, input CreatePurchaseInput) (*PurchaseOrder, error) {
	if tenantID == uuid.Nil {
		return nil, errMissingTenant
	}
	if input.VendorID == uuid.Nil {
		return nil, fmt.Errorf("%w: a vendor is required", ErrInvalidInput)
	}
	if len(input.Lines) == 0 {
		return nil, fmt.Errorf("%w: at least one line is required", ErrInvalidInput)
	}

	orderID := uuid.New()
	lines := make([]PurchaseOrderLine, 0, len(input.Lines))
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
		lines = append(lines, PurchaseOrderLine{
			ID: uuid.New(), PurchaseOrderID: orderID, TenantID: tenantID,
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

	order := PurchaseOrder{
		ID: orderID, TenantID: tenantID, VendorID: input.VendorID,
		Status: PurchaseStatusDraft, OrderDate: orderDate, ExpectedDate: input.ExpectedDate,
		Currency: currency, Notes: strings.TrimSpace(input.Notes), Lines: lines,
	}
	order.Recalculate()

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var vendor Vendor
		if err := tx.Where("tenant_id = ? AND id = ?", tenantID, input.VendorID).
			First(&vendor).Error; err != nil {
			return ErrVendorNotFound
		}
		if vendor.Status != VendorStatusActive {
			return ErrVendorInactive
		}
		order.VendorName = vendor.Name

		number := strings.ToUpper(strings.TrimSpace(input.Number))
		if number == "" {
			generated, err := nextSequence(tx, tenantID, &PurchaseOrder{}, "number", "PO-", 6)
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
			return fmt.Errorf("create purchase order: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &order, nil
}

// UpdatePurchaseStatus moves an order through its lifecycle:
// Draft -> Confirmed -> Received, with cancellation allowed until it is
// received. A received order is terminal because stock has already moved.
func (s *Service) UpdatePurchaseStatus(ctx context.Context, tenantID, id uuid.UUID, status PurchaseStatus) (*PurchaseOrder, error) {
	if tenantID == uuid.Nil {
		return nil, errMissingTenant
	}
	if !status.Valid() {
		return nil, fmt.Errorf("%w: unknown purchase status %q", ErrInvalidInput, status)
	}

	var order PurchaseOrder
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("tenant_id = ? AND id = ?", tenantID, id).First(&order).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrPurchaseNotFound
			}
			return fmt.Errorf("load purchase order: %w", err)
		}

		if err := validatePurchaseTransition(order.Status, status); err != nil {
			return err
		}

		updates := map[string]any{"status": status}
		if status == PurchaseStatusReceived {
			now := time.Now().UTC()
			updates["received_at"] = now
			order.ReceivedAt = &now
		}
		if err := tx.Model(&order).Updates(updates).Error; err != nil {
			return fmt.Errorf("update purchase status: %w", err)
		}
		order.Status = status
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func validatePurchaseTransition(from, to PurchaseStatus) error {
	if from == to {
		return nil
	}
	switch from {
	case PurchaseStatusDraft:
		if to == PurchaseStatusConfirmed || to == PurchaseStatusCancelled {
			return nil
		}
	case PurchaseStatusConfirmed:
		if to == PurchaseStatusReceived || to == PurchaseStatusCancelled {
			return nil
		}
	case PurchaseStatusReceived:
		return fmt.Errorf("%w: a received order is final", ErrNotEditable)
	case PurchaseStatusCancelled:
		return fmt.Errorf("%w: a cancelled order is final", ErrNotEditable)
	}
	return fmt.Errorf("%w: cannot move a %s order to %s", ErrNotEditable, from, to)
}

// ReceivedLines returns the SKUs and quantities of a received order, which the
// caller can hand to the inventory module to increment stock. Procurement does
// not reach into inventory itself: the modules stay independent, and whoever
// wires them decides the policy.
func (s *Service) ReceivedLines(ctx context.Context, tenantID, id uuid.UUID) ([]PurchaseOrderLine, error) {
	order, err := s.GetPurchaseOrder(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	if order.Status != PurchaseStatusReceived {
		return nil, fmt.Errorf("%w: order is %s, not Received", ErrInvalidInput, order.Status)
	}
	return order.Lines, nil
}

// ------------------------------------------------------------------ helpers

func page(limit, offset int) (int, int) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}

// nextSequence allocates PREFIX-000001 per tenant. It reads the current maximum
// inside the caller's transaction; the unique index is what actually guarantees
// no duplicate survives a race.
func nextSequence(tx *gorm.DB, tenantID uuid.UUID, model any, column, prefix string, width int) (string, error) {
	var latest string
	err := tx.Model(model).
		Where("tenant_id = ? AND "+column+" LIKE ?", tenantID, prefix+"%").
		Order(column+" DESC").Limit(1).
		Pluck(column, &latest).Error
	if err != nil {
		return "", fmt.Errorf("read latest %s: %w", column, err)
	}

	sequence := 1
	if latest != "" {
		var parsed int
		if _, err := fmt.Sscanf(latest, prefix+"%d", &parsed); err == nil {
			sequence = parsed + 1
		}
	}
	return fmt.Sprintf("%s%0*d", prefix, width, sequence), nil
}

func isUniqueViolation(err error) bool {
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate key") ||
		strings.Contains(msg, "23505") ||
		strings.Contains(msg, "unique constraint")
}
