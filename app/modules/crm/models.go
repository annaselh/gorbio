package crm

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CustomerStatus string

const (
	CustomerStatusActive   CustomerStatus = "Active"
	CustomerStatusInactive CustomerStatus = "Inactive"
)

func (s CustomerStatus) Valid() bool {
	return s == CustomerStatusActive || s == CustomerStatusInactive
}

// Customer is tenant-owned master data. It mirrors the procurement Vendor
// deliberately: the two are the same shape of record on opposite sides of the
// ledger, and keeping them symmetrical means one mental model covers both.
type Customer struct {
	ID       uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	TenantID uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_customer_tenant_code" json:"tenant_id"`

	Code    string `gorm:"size:32;not null;uniqueIndex:idx_customer_tenant_code" json:"code"`
	Name    string `gorm:"size:200;not null;index" json:"name"`
	Email   string `gorm:"size:320" json:"email,omitempty"`
	Phone   string `gorm:"size:40" json:"phone,omitempty"`
	Address string `gorm:"size:500" json:"address,omitempty"`
	TaxID   string `gorm:"size:40" json:"tax_id,omitempty"`

	// CreditTermDays is how long this customer has to pay, mirroring the
	// vendor's payment term on the buying side.
	CreditTermDays int            `gorm:"not null;default:30" json:"credit_term_days"`
	Status         CustomerStatus `gorm:"size:16;not null;default:Active;index" json:"status"`
	Notes          string         `gorm:"size:1000" json:"notes,omitempty"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Customer) TableName() string {
	return "crm_customers"
}
