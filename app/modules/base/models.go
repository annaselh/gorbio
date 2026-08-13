package base

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserStatus string

const (
	UserStatusPending  UserStatus = "pending"
	UserStatusActive   UserStatus = "active"
	UserStatusDisabled UserStatus = "disabled"
)

type TenantStatus string

const (
	TenantStatusActive    TenantStatus = "active"
	TenantStatusSuspended TenantStatus = "suspended"
)

type MembershipStatus string

const (
	MembershipStatusActive    MembershipStatus = "active"
	MembershipStatusSuspended MembershipStatus = "suspended"
)

// User is global: one person may belong to multiple tenants.
type User struct {
	ID                  uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Email               string         `gorm:"size:320;not null;uniqueIndex" json:"email"`
	DisplayName         string         `gorm:"size:160;not null" json:"display_name"`
	PasswordHash        string         `gorm:"not null" json:"-"`
	Status              UserStatus     `gorm:"size:16;not null;default:pending" json:"status"`
	EmailVerifiedAt     *time.Time     `json:"email_verified_at,omitempty"`
	PasswordChangedAt   *time.Time     `json:"password_changed_at,omitempty"`
	LastLoginAt         *time.Time     `json:"last_login_at,omitempty"`
	FailedLoginAttempts uint           `gorm:"not null;default:0" json:"-"`
	LockedUntil         *time.Time     `json:"-"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"-"`
}

type Tenant struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Slug      string         `gorm:"size:80;not null;uniqueIndex" json:"slug"`
	Name      string         `gorm:"size:160;not null" json:"name"`
	Status    TenantStatus   `gorm:"size:16;not null;default:active" json:"status"`
	Plan      string         `gorm:"size:40;not null;default:starter" json:"plan"`
	Settings  []byte         `gorm:"type:jsonb;not null;default:'{}'" json:"settings"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type Membership struct {
	ID        uuid.UUID        `gorm:"type:uuid;primaryKey" json:"id"`
	UserID    uuid.UUID        `gorm:"type:uuid;not null;uniqueIndex:idx_membership_user_tenant" json:"user_id"`
	TenantID  uuid.UUID        `gorm:"type:uuid;not null;uniqueIndex:idx_membership_user_tenant;index" json:"tenant_id"`
	Status    MembershipStatus `gorm:"size:16;not null;default:active" json:"status"`
	JoinedAt  time.Time        `gorm:"not null" json:"joined_at"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
	User      User             `gorm:"constraint:OnDelete:CASCADE" json:"-"`
	Tenant    Tenant           `gorm:"constraint:OnDelete:CASCADE" json:"-"`
}

type Role struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Code        string    `gorm:"size:100;not null;uniqueIndex" json:"code"`
	Name        string    `gorm:"size:120;not null" json:"name"`
	Description string    `gorm:"size:500" json:"description"`
	IsSystem    bool      `gorm:"not null;default:true" json:"is_system"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Permission struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Code        string    `gorm:"size:150;not null;uniqueIndex" json:"code"`
	Module      string    `gorm:"size:80;not null;index" json:"module"`
	Action      string    `gorm:"size:80;not null" json:"action"`
	Description string    `gorm:"size:500" json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type MembershipRole struct {
	MembershipID uuid.UUID `gorm:"type:uuid;primaryKey"`
	RoleID       uuid.UUID `gorm:"type:uuid;primaryKey"`
	CreatedAt    time.Time
}

type RolePermission struct {
	RoleID       uuid.UUID `gorm:"type:uuid;primaryKey"`
	PermissionID uuid.UUID `gorm:"type:uuid;primaryKey"`
	CreatedAt    time.Time
}

// Session stores only a hash of the opaque browser token.
type Session struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	UserID         uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	ActiveTenantID *uuid.UUID `gorm:"type:uuid;index" json:"active_tenant_id,omitempty"`
	TokenHash      string     `gorm:"size:64;not null;uniqueIndex" json:"-"`
	ExpiresAt      time.Time  `gorm:"not null;index" json:"expires_at"`
	LastUsedAt     time.Time  `gorm:"not null" json:"last_used_at"`
	RevokedAt      *time.Time `gorm:"index" json:"revoked_at,omitempty"`
	IPAddress      string     `gorm:"size:64" json:"ip_address,omitempty"`
	UserAgent      string     `gorm:"size:512" json:"user_agent,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type PasswordResetToken struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null;index"`
	TokenHash string     `gorm:"size:64;not null;uniqueIndex"`
	ExpiresAt time.Time  `gorm:"not null;index"`
	UsedAt    *time.Time `gorm:"index"`
	CreatedAt time.Time
}

type EmailVerificationToken struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null;index"`
	TokenHash string     `gorm:"size:64;not null;uniqueIndex"`
	ExpiresAt time.Time  `gorm:"not null;index"`
	UsedAt    *time.Time `gorm:"index"`
	CreatedAt time.Time
}

type AuditLog struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	TenantID     *uuid.UUID `gorm:"type:uuid;index" json:"tenant_id,omitempty"`
	ActorUserID  *uuid.UUID `gorm:"type:uuid;index" json:"actor_user_id,omitempty"`
	Action       string     `gorm:"size:150;not null;index" json:"action"`
	ResourceType string     `gorm:"size:100;not null;index" json:"resource_type"`
	ResourceID   string     `gorm:"size:100;index" json:"resource_id,omitempty"`
	Metadata     []byte     `gorm:"type:jsonb;not null;default:'{}'" json:"metadata"`
	IPAddress    string     `gorm:"size:64" json:"ip_address,omitempty"`
	UserAgent    string     `gorm:"size:512" json:"user_agent,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}
