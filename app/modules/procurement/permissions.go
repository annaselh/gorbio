package procurement

// Permission codes owned by this module. Vendors and purchase orders share one
// pair: in practice whoever may raise a purchase order also maintains the
// supplier list, and splitting them would create a role nobody assigns.
const (
	PermissionRead   = "procurement.read"
	PermissionManage = "procurement.manage"
)
