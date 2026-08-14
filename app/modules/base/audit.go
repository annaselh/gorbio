package base

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/annaselh/gorbio/core"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const AuditServiceName = "base.audit"

// AuditService lets any module record a business event against the shared audit
// trail. It is published in the service container so a module does not have to
// import base's models to write one.
type AuditService struct {
	db *gorm.DB
}

func NewAuditService(db *gorm.DB) *AuditService {
	return &AuditService{db: db}
}

// AuditFromApp resolves the audit service.
func AuditFromApp(app *core.App) (*AuditService, error) {
	return core.ServiceAs[*AuditService](app, AuditServiceName)
}

// AuditEntry describes one recorded event. Action is a dotted verb phrase such
// as "sales.order_confirmed"; the prefix is the owning module.
type AuditEntry struct {
	Action       string
	ResourceType string
	ResourceID   string
	// Metadata is serialised to JSON. Keep it small and free of secrets - audit
	// rows are read by anyone with membership.read.
	Metadata map[string]any
}

// Record writes an audit row for the principal's tenant.
//
// Auditing must never fail the operation it describes, so a write error is
// returned for the caller to log rather than propagated as a business failure.
// Use RecordTx when the row should commit atomically with the change.
func (s *AuditService) Record(ctx context.Context, principal Principal, entry AuditEntry) error {
	row, err := buildAuditLog(principal, entry)
	if err != nil {
		return err
	}
	return s.db.WithContext(ctx).Create(row).Error
}

// RecordTx writes the audit row inside an existing transaction, so the record
// commits or rolls back together with the change it describes.
func (s *AuditService) RecordTx(ctx context.Context, tx *gorm.DB, principal Principal, entry AuditEntry) error {
	row, err := buildAuditLog(principal, entry)
	if err != nil {
		return err
	}
	return tx.WithContext(ctx).Create(row).Error
}

func buildAuditLog(principal Principal, entry AuditEntry) (*AuditLog, error) {
	metadata := []byte("{}")
	if len(entry.Metadata) > 0 {
		encoded, err := json.Marshal(entry.Metadata)
		if err != nil {
			return nil, fmt.Errorf("encode audit metadata: %w", err)
		}
		metadata = encoded
	}

	actor := principal.UserID
	tenant := principal.TenantID
	return &AuditLog{
		ID:           uuid.New(),
		TenantID:     &tenant,
		ActorUserID:  &actor,
		Action:       entry.Action,
		ResourceType: entry.ResourceType,
		ResourceID:   entry.ResourceID,
		Metadata:     metadata,
	}, nil
}

// ActivityEntry is one row of the dashboard activity feed: an audit row joined
// to the display name of whoever caused it.
type ActivityEntry struct {
	ID           uuid.UUID `json:"id"`
	Action       string    `json:"action"`
	ResourceType string    `json:"resource_type"`
	ResourceID   string    `json:"resource_id,omitempty"`
	ActorName    string    `json:"actor_name,omitempty"`
	Metadata     []byte    `json:"metadata"`
	CreatedAt    time.Time `json:"created_at"`
}

// RecentActivity returns the newest audit rows for a tenant. Authentication
// events are excluded: a feed of "someone signed in" crowds out the business
// events the dashboard is meant to surface.
func (s *AuditService) RecentActivity(ctx context.Context, tenantID uuid.UUID, limit int) ([]ActivityEntry, error) {
	if tenantID == uuid.Nil {
		return nil, fmt.Errorf("tenant scope is required")
	}
	if limit <= 0 || limit > 100 {
		limit = 10
	}

	var entries []ActivityEntry
	err := s.db.WithContext(ctx).
		Table("audit_logs").
		Select("audit_logs.id, audit_logs.action, audit_logs.resource_type, audit_logs.resource_id, audit_logs.metadata, audit_logs.created_at, users.display_name AS actor_name").
		Joins("LEFT JOIN users ON users.id = audit_logs.actor_user_id").
		Where("audit_logs.tenant_id = ? AND audit_logs.action NOT LIKE 'auth.%'", tenantID).
		Order("audit_logs.created_at DESC").
		Limit(limit).
		Scan(&entries).Error
	if err != nil {
		return nil, fmt.Errorf("load recent activity: %w", err)
	}
	return entries, nil
}
