package sales

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OrderStatus string

const (
	OrderStatusDraft     OrderStatus = "Draft"
	OrderStatusConfirmed OrderStatus = "Confirmed"
	OrderStatusCancelled OrderStatus = "Cancelled"
)

func (s OrderStatus) Valid() bool {
	switch s {
	case OrderStatusDraft, OrderStatusConfirmed, OrderStatusCancelled:
		return true
	default:
		return false
	}
}

// Order is tenant-owned; every read must go through the tenant scope.
//
// Money is stored in minor units (cents / rupiah) as int64 rather than a float.
// Binary floating point cannot represent 0.1 exactly, so summing float lines
// drifts, and an ERP that reports a total a customer disputes is worse than one
// that is merely inconvenient to write.
type Order struct {
	ID       uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	TenantID uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_sales_order_tenant_number" json:"tenant_id"`

	Number       string      `gorm:"size:32;not null;uniqueIndex:idx_sales_order_tenant_number" json:"number"`
	CustomerName string      `gorm:"size:200;not null" json:"customer_name"`
	Status       OrderStatus `gorm:"size:16;not null;default:Draft;index" json:"status"`
	OrderDate    time.Time   `gorm:"not null;index" json:"order_date"`
	Currency     string      `gorm:"size:3;not null;default:IDR" json:"currency"`

	// Subtotal is the sum of the lines; DiscountTotal is applied by the
	// sales-discount extension; Total is what the customer owes.
	Subtotal      int64 `gorm:"not null;default:0" json:"subtotal"`
	DiscountTotal int64 `gorm:"not null;default:0" json:"discount_total"`
	Total         int64 `gorm:"not null;default:0" json:"total"`

	Notes string `gorm:"size:1000" json:"notes,omitempty"`

	Lines []OrderLine `gorm:"foreignKey:OrderID;constraint:OnDelete:CASCADE" json:"lines,omitempty"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Order) TableName() string {
	return "sales_orders"
}

// OrderLine carries TenantID as well as OrderID. It is redundant against a
// join, and deliberately so: it lets the tenant scope apply uniformly to every
// table, so a query that forgets to join through the order still cannot read
// another tenant's rows.
type OrderLine struct {
	ID       uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	OrderID  uuid.UUID `gorm:"type:uuid;not null;index" json:"order_id"`
	TenantID uuid.UUID `gorm:"type:uuid;not null;index" json:"tenant_id"`

	SKU         string `gorm:"size:64;not null" json:"sku"`
	Description string `gorm:"size:200;not null" json:"description"`
	Quantity    int    `gorm:"not null" json:"quantity"`
	UnitPrice   int64  `gorm:"not null" json:"unit_price"`
	LineTotal   int64  `gorm:"not null" json:"line_total"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (OrderLine) TableName() string {
	return "sales_order_lines"
}

// Recalculate derives the money fields from the lines. It is the single place
// that decides what an order costs, so the extension that applies a discount
// and the service that adds a line cannot disagree.
func (o *Order) Recalculate() {
	var subtotal int64
	for i := range o.Lines {
		line := &o.Lines[i]
		line.LineTotal = int64(line.Quantity) * line.UnitPrice
		subtotal += line.LineTotal
	}
	o.Subtotal = subtotal

	// A discount can never exceed the subtotal or turn the total negative.
	if o.DiscountTotal > subtotal {
		o.DiscountTotal = subtotal
	}
	if o.DiscountTotal < 0 {
		o.DiscountTotal = 0
	}
	o.Total = subtotal - o.DiscountTotal
}

// Editable reports whether the order may still be changed. A confirmed or
// cancelled order is a financial record, not a draft.
func (o *Order) Editable() bool {
	return o.Status == OrderStatusDraft
}
