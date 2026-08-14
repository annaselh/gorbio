package inventory

// Permission codes owned by this module. They are seeded by the base module's
// authorization seed; declaring them here keeps the strings that guard the
// routes in one place instead of scattered literals.
const (
	PermissionRead   = "inventory.read"
	PermissionManage = "inventory.manage"
)
