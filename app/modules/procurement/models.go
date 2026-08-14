package procurement

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type VendorStatus string

const (
	VendorStatusActive   VendorStatus = "Active"
	VendorStatusInactive VendorStatus = "Inactive"
)

func (s VendorStatus) Valid() bool {
	return s == VendorStatusActive || s == VendorStatusInactive
}

type PurchaseStatus string

const (
	PurchaseStatusDraft     PurchaseStatus = "Draft"
	PurchaseStatusConfirmed PurchaseStatus = "Confirmed"
	PurchaseStatusReceived  PurchaseStatus = "Received"
	PurchaseStatusCancelled PurchaseStatus = "Cancelled"
)

func (s PurchaseStatus) Valid() bool {
	switch s {
	case PurchaseStatusDraft, PurchaseStatusConfirmed, PurchaseStatusReceived, PurchaseStatusCancelled:
		return true
	default:
		return false
	}
}

// Vendor is a supplier. Tenant-owned, with a per-tenant unique code so two
// customers can both have a vendor coded "V-001".
type Vendor struct {
	ID       uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	TenantID uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_vendor_tenant_code" json:"tenant_id"`

	Code        string       `gorm:"size:32;not null;uniqueIndex:idx_vendor_tenant_code" json:"code"`
	Name        string       `gorm:"size:200;not null;index" json:"name"`
	Email       string       `gorm:"size:320" json:"email,omitempty"`
	Phone       string       `gorm:"size:40" json:"phone,omitempty"`
	Address     string       `gorm:"size:500" json:"address,omitempty"`
	TaxID       string       `gorm:"size:40" json:"tax_id,omitempty"`
	PaymentTerm int          `gorm:"not null;default:30" json:"payment_term_days"`
	Status      VendorStatus `gorm:"size:16;not null;default:Active;index" json:"status"`
	Notes       string       `gorm:"size:1000" json:"notes,omitempty"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Vendor) TableName() string {
	return "procurement_vendors"
}

// PurchaseOrder mirrors the sales order shape deliberately: money in int64
// minor units, totals derived in one place, per-tenant unique number.
type PurchaseOrder struct {
	ID       uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	TenantID uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_purchase_tenant_number" json:"tenant_id"`

	Number   string    `gorm:"size:32;not null;uniqueIndex:idx_purchase_tenant_number" json:"number"`
	VendorID uuid.UUID `gorm:"type:uuid;not null;index" json:"vendor_id"`
	// VendorName is denormalised so a historical order still reads correctly
	// after the vendor is renamed. The link stays in VendorID.
	VendorName string `gorm:"size:200;not null" json:"vendor_name"`

	Status       PurchaseStatus `gorm:"size:16;not null;default:Draft;index" json:"status"`
	OrderDate    time.Time      `gorm:"not null;index" json:"order_date"`
	ExpectedDate *time.Time     `json:"expected_date,omitempty"`
	ReceivedAt   *time.Time     `json:"received_at,omitempty"`
	Currency     string         `gorm:"size:3;not null;default:IDR" json:"currency"`

	Subtotal      int64 `gorm:"not null;default:0" json:"subtotal"`
	DiscountTotal int64 `gorm:"not null;default:0" json:"discount_total"`
	Total         int64 `gorm:"not null;default:0" json:"total"`

	Notes string `gorm:"size:1000" json:"notes,omitempty"`

	Lines []PurchaseOrderLine `gorm:"foreignKey:PurchaseOrderID;constraint:OnDelete:CASCADE" json:"lines,omitempty"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (PurchaseOrder) TableName() string {
	return "procurement_purchase_orders"
}

// PurchaseOrderLine carries TenantID as well as PurchaseOrderID for the same
// reason the sales line does: the tenant scope then applies uniformly to every
// table, so a query that forgets to join through the order still cannot read
// another tenant's rows.
type PurchaseOrderLine struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	PurchaseOrderID uuid.UUID `gorm:"type:uuid;not null;index" json:"purchase_order_id"`
	TenantID        uuid.UUID `gorm:"type:uuid;not null;index" json:"tenant_id"`

	SKU         string `gorm:"size:64;not null;index" json:"sku"`
	Description string `gorm:"size:200;not null" json:"description"`
	Quantity    int    `gorm:"not null" json:"quantity"`
	UnitPrice   int64  `gorm:"not null" json:"unit_price"`
	LineTotal   int64  `gorm:"not null" json:"line_total"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (PurchaseOrderLine) TableName() string {
	return "procurement_purchase_order_lines"
}

// Recalculate derives the money fields from the lines - the single place that
// decides what a purchase order costs.
func (p *PurchaseOrder) Recalculate() {
	var subtotal int64
	for i := range p.Lines {
		line := &p.Lines[i]
		line.LineTotal = int64(line.Quantity) * line.UnitPrice
		subtotal += line.LineTotal
	}
	p.Subtotal = subtotal

	if p.DiscountTotal > subtotal {
		p.DiscountTotal = subtotal
	}
	if p.DiscountTotal < 0 {
		p.DiscountTotal = 0
	}
	p.Total = subtotal - p.DiscountTotal
}

// Editable reports whether the order may still be changed. Once confirmed it is
// a commitment to a supplier, not a draft.
func (p *PurchaseOrder) Editable() bool {
	return p.Status == PurchaseStatusDraft
}
