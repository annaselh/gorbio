package inventory

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const ServiceName = "inventory.stock"

var (
	ErrNotFound      = errors.New("stock item not found")
	ErrDuplicateSKU  = errors.New("stock item with this SKU already exists")
	ErrInvalidInput  = errors.New("invalid stock item")
	ErrInsufficient  = errors.New("insufficient stock")
	errMissingTenant = errors.New("tenant scope is required")
)

// Service holds the unscoped handle. Every method takes the tenant explicitly
// so a caller cannot accidentally run a cross-tenant query - the compiler makes
// the scope impossible to forget.
type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

type CreateItemInput struct {
	SKU          string
	Name         string
	Unit         string
	Quantity     int
	ReorderLevel int
}

type ListFilter struct {
	// LowStockOnly restricts the result to items at or below their reorder
	// level, which is what the dashboard alert widget asks for.
	LowStockOnly bool
}

func (s *Service) scoped(ctx context.Context, tenantID uuid.UUID) (*gorm.DB, error) {
	if tenantID == uuid.Nil {
		return nil, errMissingTenant
	}
	return s.db.WithContext(ctx).Where("tenant_id = ?", tenantID), nil
}

func (s *Service) List(ctx context.Context, tenantID uuid.UUID, filter ListFilter) ([]StockItem, error) {
	query, err := s.scoped(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	if filter.LowStockOnly {
		query = query.Where("quantity <= reorder_level")
	}

	var items []StockItem
	if err := query.Order("name ASC").Find(&items).Error; err != nil {
		return nil, fmt.Errorf("list stock items: %w", err)
	}
	return items, nil
}

func (s *Service) Get(ctx context.Context, tenantID, id uuid.UUID) (*StockItem, error) {
	query, err := s.scoped(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	var item StockItem
	if err := query.Where("id = ?", id).First(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get stock item: %w", err)
	}
	return &item, nil
}

func (s *Service) Create(ctx context.Context, tenantID uuid.UUID, input CreateItemInput) (*StockItem, error) {
	if tenantID == uuid.Nil {
		return nil, errMissingTenant
	}

	sku := strings.ToUpper(strings.TrimSpace(input.SKU))
	name := strings.TrimSpace(input.Name)
	if sku == "" || name == "" {
		return nil, fmt.Errorf("%w: sku and name are required", ErrInvalidInput)
	}
	if input.Quantity < 0 || input.ReorderLevel < 0 {
		return nil, fmt.Errorf("%w: quantity and reorder level must not be negative", ErrInvalidInput)
	}

	unit := strings.TrimSpace(input.Unit)
	if unit == "" {
		unit = "pcs"
	}

	item := StockItem{
		ID: uuid.New(), TenantID: tenantID, SKU: sku, Name: name, Unit: unit,
		Quantity: input.Quantity, ReorderLevel: input.ReorderLevel,
	}
	if err := s.db.WithContext(ctx).Create(&item).Error; err != nil {
		if isUniqueViolation(err) {
			return nil, ErrDuplicateSKU
		}
		return nil, fmt.Errorf("create stock item: %w", err)
	}
	return &item, nil
}

// Adjust applies a signed delta to the on-hand quantity inside a transaction,
// re-reading the row under a write lock so two concurrent adjustments cannot
// both read the same starting value and lose one another's update.
func (s *Service) Adjust(ctx context.Context, tenantID, id uuid.UUID, delta int) (*StockItem, error) {
	if tenantID == uuid.Nil {
		return nil, errMissingTenant
	}

	var item StockItem
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Set("gorm:query_option", "FOR UPDATE").
			Where("tenant_id = ? AND id = ?", tenantID, id).First(&item).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrNotFound
			}
			return fmt.Errorf("load stock item: %w", err)
		}

		updated := item.Quantity + delta
		if updated < 0 {
			return fmt.Errorf("%w: %d on hand, requested %d", ErrInsufficient, item.Quantity, delta)
		}

		if err := tx.Model(&item).Update("quantity", updated).Error; err != nil {
			return fmt.Errorf("update stock quantity: %w", err)
		}
		item.Quantity = updated
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func isUniqueViolation(err error) bool {
	// Matching on message keeps this driver-agnostic; the Postgres code is
	// 23505 and pgx surfaces it in the error text.
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate key") || strings.Contains(msg, "23505") || strings.Contains(msg, "unique constraint")
}
