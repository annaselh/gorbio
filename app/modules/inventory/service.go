package inventory

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
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

// Receipt is one incoming line of goods. Description and Unit are consulted
// only when the SKU has no stock item yet and one has to be created.
type Receipt struct {
	SKU         string
	Description string
	Unit        string
	Quantity    int
}

// ReceiveTx adds incoming quantities to stock inside the caller's transaction.
//
// It takes the transaction rather than opening its own so the stock movement
// commits with the business event that caused it. A purchase order marked
// Received and the stock it brought in must not be able to disagree: if either
// write fails, neither should survive.
//
// A SKU with no stock item yet is created at the received quantity. Purchase
// order lines carry free-text SKUs with no foreign key into inventory, so
// refusing would make a perfectly ordinary receipt impossible to record without
// hand-creating the item first.
func (s *Service) ReceiveTx(ctx context.Context, tx *gorm.DB, tenantID uuid.UUID, receipts []Receipt) error {
	if tenantID == uuid.Nil {
		return errMissingTenant
	}
	if tx == nil {
		return fmt.Errorf("%w: a transaction is required", ErrInvalidInput)
	}

	// Two lines of the same order may name the same SKU. Summing them first
	// means one locked row per SKU instead of one per line.
	totals, order, err := totalBySKU(receipts)
	if err != nil {
		return err
	}

	for _, sku := range order {
		incoming := totals[sku]

		var item StockItem
		err := tx.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("tenant_id = ? AND sku = ?", tenantID, sku).First(&item).Error
		switch {
		case err == nil:
			if err := tx.WithContext(ctx).Model(&item).
				Update("quantity", item.Quantity+incoming.Quantity).Error; err != nil {
				return fmt.Errorf("receive stock for %s: %w", sku, err)
			}
		case errors.Is(err, gorm.ErrRecordNotFound):
			created := StockItem{
				ID: uuid.New(), TenantID: tenantID, SKU: sku,
				Name:     incoming.Name,
				Unit:     incoming.Unit,
				Quantity: incoming.Quantity,
			}
			if err := tx.WithContext(ctx).Create(&created).Error; err != nil {
				return fmt.Errorf("create stock item %s on receipt: %w", sku, err)
			}
		default:
			return fmt.Errorf("load stock item %s: %w", sku, err)
		}
	}
	return nil
}

// incoming is the summed arrival for one SKU, carrying the fields needed to
// open a stock item if none exists.
type incoming struct {
	Quantity int
	Name     string
	Unit     string
}

// totalBySKU validates and sums receipts, returning the SKUs in first-seen
// order so the rows are always locked in a deterministic sequence - two
// concurrent receipts touching the same pair of SKUs cannot deadlock by
// approaching them from opposite ends.
func totalBySKU(receipts []Receipt) (map[string]incoming, []string, error) {
	totals := make(map[string]incoming, len(receipts))
	order := make([]string, 0, len(receipts))

	for _, receipt := range receipts {
		sku := strings.ToUpper(strings.TrimSpace(receipt.SKU))
		if sku == "" {
			return nil, nil, fmt.Errorf("%w: a received line has no sku", ErrInvalidInput)
		}
		if receipt.Quantity <= 0 {
			return nil, nil, fmt.Errorf("%w: received quantity for %s must be positive", ErrInvalidInput, sku)
		}

		existing, seen := totals[sku]
		if !seen {
			order = append(order, sku)
			existing.Name = strings.TrimSpace(receipt.Description)
			if existing.Name == "" {
				existing.Name = sku
			}
			existing.Unit = strings.TrimSpace(receipt.Unit)
			if existing.Unit == "" {
				existing.Unit = "pcs"
			}
		}
		existing.Quantity += receipt.Quantity
		totals[sku] = existing
	}
	return totals, order, nil
}

func isUniqueViolation(err error) bool {
	// Matching on message keeps this driver-agnostic; the Postgres code is
	// 23505 and pgx surfaces it in the error text.
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "duplicate key") || strings.Contains(msg, "23505") || strings.Contains(msg, "unique constraint")
}
