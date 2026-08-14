package crm

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const ServiceName = "crm.customers"

var (
	ErrNotFound      = errors.New("customer not found")
	ErrInvalidInput  = errors.New("invalid customer")
	ErrDuplicate     = errors.New("a customer with this code already exists")
	errMissingTenant = errors.New("tenant scope is required")
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

type CustomerInput struct {
	Code           string
	Name           string
	Email          string
	Phone          string
	Address        string
	TaxID          string
	CreditTermDays int
	Notes          string
}

type ListFilter struct {
	Status CustomerStatus
	Search string
	Limit  int
	Offset int
}

type ListResult struct {
	Customers []Customer
	Total     int64
}

func (s *Service) scoped(ctx context.Context, tenantID uuid.UUID) (*gorm.DB, error) {
	if tenantID == uuid.Nil {
		return nil, errMissingTenant
	}
	return s.db.WithContext(ctx).Where("tenant_id = ?", tenantID), nil
}

func (s *Service) List(ctx context.Context, tenantID uuid.UUID, filter ListFilter) (*ListResult, error) {
	query, err := s.scoped(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	query = query.Model(&Customer{})

	if filter.Status != "" {
		if !filter.Status.Valid() {
			return nil, fmt.Errorf("%w: unknown customer status %q", ErrInvalidInput, filter.Status)
		}
		query = query.Where("status = ?", filter.Status)
	}
	if search := strings.TrimSpace(filter.Search); search != "" {
		query = query.Where("name ILIKE ? OR code ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("count customers: %w", err)
	}

	limit, offset := page(filter.Limit, filter.Offset)
	var customers []Customer
	if err := query.Order("name ASC").Limit(limit).Offset(offset).Find(&customers).Error; err != nil {
		return nil, fmt.Errorf("list customers: %w", err)
	}
	return &ListResult{Customers: customers, Total: total}, nil
}

func (s *Service) Get(ctx context.Context, tenantID, id uuid.UUID) (*Customer, error) {
	query, err := s.scoped(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	var customer Customer
	if err := query.Where("id = ?", id).First(&customer).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get customer: %w", err)
	}
	return &customer, nil
}

func (s *Service) Create(ctx context.Context, tenantID uuid.UUID, input CustomerInput) (*Customer, error) {
	if tenantID == uuid.Nil {
		return nil, errMissingTenant
	}
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, fmt.Errorf("%w: customer name is required", ErrInvalidInput)
	}
	if input.CreditTermDays < 0 {
		return nil, fmt.Errorf("%w: credit term must not be negative", ErrInvalidInput)
	}

	creditTerm := input.CreditTermDays
	if creditTerm == 0 {
		creditTerm = 30
	}

	customer := Customer{
		ID: uuid.New(), TenantID: tenantID, Name: name,
		Email: strings.ToLower(strings.TrimSpace(input.Email)),
		Phone: strings.TrimSpace(input.Phone), Address: strings.TrimSpace(input.Address),
		TaxID: strings.TrimSpace(input.TaxID), CreditTermDays: creditTerm,
		Status: CustomerStatusActive, Notes: strings.TrimSpace(input.Notes),
	}

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		code := strings.ToUpper(strings.TrimSpace(input.Code))
		if code == "" {
			generated, err := nextCustomerCode(tx, tenantID)
			if err != nil {
				return err
			}
			code = generated
		}
		customer.Code = code

		if err := tx.Create(&customer).Error; err != nil {
			if isUniqueViolation(err) {
				return ErrDuplicate
			}
			return fmt.Errorf("create customer: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &customer, nil
}

func (s *Service) Update(ctx context.Context, tenantID, id uuid.UUID, input CustomerInput) (*Customer, error) {
	customer, err := s.Get(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, fmt.Errorf("%w: customer name is required", ErrInvalidInput)
	}
	if input.CreditTermDays < 0 {
		return nil, fmt.Errorf("%w: credit term must not be negative", ErrInvalidInput)
	}

	updates := map[string]any{
		"name": name, "email": strings.ToLower(strings.TrimSpace(input.Email)),
		"phone": strings.TrimSpace(input.Phone), "address": strings.TrimSpace(input.Address),
		"tax_id": strings.TrimSpace(input.TaxID), "notes": strings.TrimSpace(input.Notes),
	}
	if input.CreditTermDays > 0 {
		updates["credit_term_days"] = input.CreditTermDays
	}

	if err := s.db.WithContext(ctx).Model(customer).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("update customer: %w", err)
	}
	return s.Get(ctx, tenantID, id)
}

// SetStatus deactivates or reactivates a customer. Existing sales orders are
// untouched: they record who actually bought.
func (s *Service) SetStatus(ctx context.Context, tenantID, id uuid.UUID, status CustomerStatus) (*Customer, error) {
	if !status.Valid() {
		return nil, fmt.Errorf("%w: unknown customer status %q", ErrInvalidInput, status)
	}
	customer, err := s.Get(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	if err := s.db.WithContext(ctx).Model(customer).Update("status", status).Error; err != nil {
		return nil, fmt.Errorf("update customer status: %w", err)
	}
	customer.Status = status
	return customer, nil
}

func page(limit, offset int) (int, int) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}

// nextCustomerCode allocates C-0001 per tenant. The unique index on
// (tenant_id, code) is what guarantees no duplicate survives a race; this read
// is only an allocator.
func nextCustomerCode(tx *gorm.DB, tenantID uuid.UUID) (string, error) {
	var latest string
	err := tx.Model(&Customer{}).
		Where("tenant_id = ? AND code LIKE 'C-%'", tenantID).
		Order("code DESC").Limit(1).
		Pluck("code", &latest).Error
	if err != nil {
		return "", fmt.Errorf("read latest customer code: %w", err)
	}

	sequence := 1
	if latest != "" {
		var parsed int
		if _, err := fmt.Sscanf(latest, "C-%d", &parsed); err == nil {
			sequence = parsed + 1
		}
	}
	return fmt.Sprintf("C-%04d", sequence), nil
}

func isUniqueViolation(err error) bool {
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate key") ||
		strings.Contains(msg, "23505") ||
		strings.Contains(msg, "unique constraint")
}
