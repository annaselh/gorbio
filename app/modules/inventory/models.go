package inventory

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// StockItem is tenant-owned. TenantID is not optional and every read must go
// through base.ScopedDB; see modules/base/tenancy.go for why.
type StockItem struct {
	ID       uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	TenantID uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_stock_item_tenant_sku" json:"tenant_id"`

	SKU      string `gorm:"size:64;not null;uniqueIndex:idx_stock_item_tenant_sku" json:"sku"`
	Name     string `gorm:"size:200;not null" json:"name"`
	Unit     string `gorm:"size:24;not null;default:pcs" json:"unit"`
	Quantity int    `gorm:"not null;default:0" json:"quantity"`
	// ReorderLevel drives the dashboard's stock-alert widget: quantity at or
	// below this threshold is surfaced as low stock.
	ReorderLevel int `gorm:"not null;default:0" json:"reorder_level"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (StockItem) TableName() string {
	return "inventory_stock_items"
}

// IsLowStock mirrors the alert rule used by the frontend widget.
func (s StockItem) IsLowStock() bool {
	return s.Quantity <= s.ReorderLevel
}
