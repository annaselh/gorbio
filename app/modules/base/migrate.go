package base

import (
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Migrate creates the base identity and tenancy schema. Call it explicitly in
// deployment tooling; it is idempotent for the current schema.
func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&User{}, &Tenant{}, &Membership{}, &Role{}, &Permission{}, &MembershipRole{},
		&RolePermission{}, &Session{}, &PasswordResetToken{}, &EmailVerificationToken{}, &AuditLog{},
	); err != nil {
		return fmt.Errorf("migrate base schema: %w", err)
	}

	return seedSystemAuthorization(db)
}

func seedSystemAuthorization(db *gorm.DB) error {
	roles := []Role{
		{ID: uuid.New(), Code: "owner", Name: "Owner", Description: "Full control within a tenant"},
		{ID: uuid.New(), Code: "admin", Name: "Administrator", Description: "Tenant administration"},
		{ID: uuid.New(), Code: "member", Name: "Member", Description: "Standard tenant member"},
	}
	if err := db.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "code"}}, DoNothing: true}).Create(&roles).Error; err != nil {
		return fmt.Errorf("seed roles: %w", err)
	}

	permissions := []Permission{
		{ID: uuid.New(), Code: "tenant.read", Module: "base", Action: "read", Description: "View tenant details"},
		{ID: uuid.New(), Code: "tenant.manage", Module: "base", Action: "manage", Description: "Manage tenant settings"},
		{ID: uuid.New(), Code: "membership.read", Module: "base", Action: "read", Description: "View tenant members"},
		{ID: uuid.New(), Code: "membership.manage", Module: "base", Action: "manage", Description: "Manage tenant members and roles"},
		{ID: uuid.New(), Code: "sales.read", Module: "sales", Action: "read", Description: "View sales orders"},
		{ID: uuid.New(), Code: "sales.manage", Module: "sales", Action: "manage", Description: "Create and modify sales orders"},
		{ID: uuid.New(), Code: "inventory.read", Module: "inventory", Action: "read", Description: "View stock items and levels"},
		{ID: uuid.New(), Code: "inventory.manage", Module: "inventory", Action: "manage", Description: "Create and adjust stock items"},
	}
	if err := db.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "code"}}, DoNothing: true}).Create(&permissions).Error; err != nil {
		return fmt.Errorf("seed permissions: %w", err)
	}

	var owner, admin, member Role
	if err := db.Where("code = ?", "owner").First(&owner).Error; err != nil {
		return fmt.Errorf("load owner role: %w", err)
	}
	if err := db.Where("code = ?", "admin").First(&admin).Error; err != nil {
		return fmt.Errorf("load admin role: %w", err)
	}
	if err := db.Where("code = ?", "member").First(&member).Error; err != nil {
		return fmt.Errorf("load member role: %w", err)
	}

	var seededPermissions []Permission
	if err := db.Order("code ASC").Find(&seededPermissions).Error; err != nil {
		return fmt.Errorf("load seeded permissions: %w", err)
	}
	assignments := make([]RolePermission, 0, len(seededPermissions)*3)
	for _, permission := range seededPermissions {
		// Owner holds everything; admin holds everything except tenant-level
		// settings; member is read-only across whichever modules are installed.
		assignments = append(assignments, RolePermission{RoleID: owner.ID, PermissionID: permission.ID})
		if permission.Code != "tenant.manage" {
			assignments = append(assignments, RolePermission{RoleID: admin.ID, PermissionID: permission.ID})
		}
		if permission.Action == "read" {
			assignments = append(assignments, RolePermission{RoleID: member.ID, PermissionID: permission.ID})
		}
	}
	if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&assignments).Error; err != nil {
		return fmt.Errorf("seed role permissions: %w", err)
	}
	return nil
}
