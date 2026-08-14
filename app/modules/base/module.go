package base

import "github.com/annaselh/gorbio/core"

type Module struct{}

func NewModule() *Module {
	return &Module{}
}

func (m *Module) Name() string {
	return "base"
}

func (m *Module) Depends() []string {
	return nil
}

func (m *Module) Migrate(app *core.App) error {
	return Migrate(app.DB)
}

func (m *Module) Register(app *core.App) error {
	auth := NewAuthService(app.DB, app.Settings, app.Mailer)
	if err := app.Services.Register(AuthServiceName, auth); err != nil {
		return err
	}

	// Published so any module can write to the shared audit trail without
	// importing base's models.
	if err := app.Services.Register(AuditServiceName, NewAuditService(app.DB)); err != nil {
		return err
	}

	app.Router.POST("/api/auth/login", auth.loginHandler)
	app.Router.POST("/api/auth/logout", auth.RequireAuth(), auth.logoutHandler)
	app.Router.GET("/api/auth/me", auth.RequireAuth(), auth.meHandler)

	// Recovery endpoints are unauthenticated by nature: the caller is someone
	// who cannot sign in. They carry their own rate limiting and answer
	// neutrally so they cannot be used to enumerate accounts.
	app.Router.POST("/api/auth/password/forgot", auth.forgotPasswordHandler)
	app.Router.POST("/api/auth/password/reset", auth.resetPasswordHandler)
	app.Router.POST("/api/auth/password/change", auth.RequireAuth(), auth.changePasswordHandler)
	app.Router.POST("/api/auth/email/verify", auth.verifyEmailHandler)
	app.Router.POST("/api/auth/email/resend", auth.RequireAuth(), auth.resendVerificationHandler)

	// Tenant administration.
	app.Router.GET("/api/members",
		auth.RequireAuth(), RequirePermission(PermissionMembershipRead), auth.listMembersHandler)
	app.Router.GET("/api/roles",
		auth.RequireAuth(), RequirePermission(PermissionMembershipRead), auth.listRolesHandler)
	app.Router.POST("/api/members",
		auth.RequireAuth(), RequirePermission(PermissionMembershipManage), auth.inviteMemberHandler)
	app.Router.PUT("/api/members/:id/roles",
		auth.RequireAuth(), RequirePermission(PermissionMembershipManage), auth.updateMemberRolesHandler)
	app.Router.PUT("/api/members/:id/status",
		auth.RequireAuth(), RequirePermission(PermissionMembershipManage), auth.updateMemberStatusHandler)
	return nil
}

func (m *Module) Boot(app *core.App) error {
	return nil
}
