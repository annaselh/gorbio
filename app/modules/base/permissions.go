package base

// Permission codes owned by this module. Declaring them keeps the strings that
// guard the routes in one place instead of scattered literals.
const (
	PermissionTenantRead       = "tenant.read"
	PermissionTenantManage     = "tenant.manage"
	PermissionMembershipRead   = "membership.read"
	PermissionMembershipManage = "membership.manage"
)
