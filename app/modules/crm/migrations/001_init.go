package migrations

import (
	"fmt"

	"gorm.io/gorm"
)

// Init creates the CRM schema. Kept separate from the module so the migration
// set stays greppable as the module grows.
func Init(db *gorm.DB, models ...any) error {
	if err := db.AutoMigrate(models...); err != nil {
		return fmt.Errorf("migrate crm schema: %w", err)
	}
	return nil
}
