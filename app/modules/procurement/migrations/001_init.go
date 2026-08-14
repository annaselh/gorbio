package migrations

import (
	"fmt"

	"gorm.io/gorm"
)

// Init creates the procurement schema. Kept separate from the module so the
// migration set stays greppable as the module grows.
func Init(db *gorm.DB, models ...any) error {
	if err := db.AutoMigrate(models...); err != nil {
		return fmt.Errorf("migrate procurement schema: %w", err)
	}
	return nil
}
