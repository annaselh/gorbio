package dashboard

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/annaselh/gorbio/core"
	"github.com/annaselh/gorbio/modules/base"
	"github.com/gin-gonic/gin"
)

// PermissionRead gates the whole dashboard. It reuses tenant.read rather than
// inventing a code of its own: anyone who can see the tenant can see its
// headline figures, and the per-module data behind them is already guarded.
const PermissionRead = base.PermissionTenantRead

type Module struct{}

func NewModule() *Module {
	return &Module{}
}

func (m *Module) Name() string {
	return "dashboard"
}

// Depends lists every module the aggregation reads from, so the registry boots
// them first and fails loudly if one is not installed.
func (m *Module) Depends() []string {
	return []string{"base", "sales", "inventory", "procurement"}
}

func (m *Module) Register(app *core.App) error {
	auth, err := base.AuthFromApp(app)
	if err != nil {
		return err
	}
	audit, err := base.AuditFromApp(app)
	if err != nil {
		return err
	}

	service := NewService(app.DB)
	if err := app.Services.Register(ServiceName, service); err != nil {
		return err
	}

	h := &handlers{service: service, audit: audit}
	read := base.RequirePermission(PermissionRead)

	app.Router.GET("/api/dashboard/summary", auth.RequireAuth(), read, h.summary)
	app.Router.GET("/api/dashboard/sales-series", auth.RequireAuth(), read, h.salesSeries)
	app.Router.GET("/api/dashboard/top-products", auth.RequireAuth(), read, h.topProducts)
	app.Router.GET("/api/dashboard/cash-flow", auth.RequireAuth(), read, h.cashFlow)
	app.Router.GET("/api/dashboard/activities", auth.RequireAuth(), read, h.activities)

	slog.Info("module registered", "module", m.Name())
	return nil
}

func (m *Module) Boot(app *core.App) error {
	slog.Info("module booted", "module", m.Name())
	return nil
}

type handlers struct {
	service *Service
	audit   *base.AuditService
}

func (h *handlers) summary(c *gin.Context) {
	principal, ok := base.PrincipalFromContext(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	summary, err := h.service.Summary(c.Request.Context(), principal.TenantID, time.Now())
	if err != nil {
		slog.Error("dashboard summary failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load summary"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": summary})
}

func (h *handlers) salesSeries(c *gin.Context) {
	principal, ok := base.PrincipalFromContext(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	current, previous, err := h.service.SalesSeries(c.Request.Context(), principal.TenantID, time.Now())
	if err != nil {
		slog.Error("dashboard sales series failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load sales series"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": gin.H{"current": current, "previous": previous}})
}

func (h *handlers) topProducts(c *gin.Context) {
	principal, ok := base.PrincipalFromContext(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	limit, _ := strconv.Atoi(c.Query("limit"))
	products, err := h.service.TopProducts(c.Request.Context(), principal.TenantID, time.Now(), limit)
	if err != nil {
		slog.Error("dashboard top products failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load top products"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": products})
}

func (h *handlers) cashFlow(c *gin.Context) {
	principal, ok := base.PrincipalFromContext(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	slices, err := h.service.CashFlow(c.Request.Context(), principal.TenantID, time.Now())
	if err != nil {
		slog.Error("dashboard cash flow failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load cash flow"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": slices})
}

// activities reads the shared audit trail rather than a table of its own, so
// the feed reflects what actually happened rather than a parallel record that
// can drift from it.
func (h *handlers) activities(c *gin.Context) {
	principal, ok := base.PrincipalFromContext(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	limit, _ := strconv.Atoi(c.Query("limit"))
	entries, err := h.audit.RecentActivity(c.Request.Context(), principal.TenantID, limit)
	if err != nil {
		slog.Error("dashboard activities failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not load activity"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": entries})
}
