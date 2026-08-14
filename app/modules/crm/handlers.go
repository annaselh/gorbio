package crm

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/annaselh/gorbio/modules/base"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type handlers struct {
	service *Service
	audit   *base.AuditService
}

type customerRequest struct {
	Code           string `json:"code"`
	Name           string `json:"name" binding:"required"`
	Email          string `json:"email"`
	Phone          string `json:"phone"`
	Address        string `json:"address"`
	TaxID          string `json:"tax_id"`
	CreditTermDays int    `json:"credit_term_days"`
	Notes          string `json:"notes"`
}

type statusRequest struct {
	Status CustomerStatus `json:"status" binding:"required"`
}

func (r customerRequest) toInput() CustomerInput {
	return CustomerInput{
		Code: r.Code, Name: r.Name, Email: r.Email, Phone: r.Phone,
		Address: r.Address, TaxID: r.TaxID,
		CreditTermDays: r.CreditTermDays, Notes: r.Notes,
	}
}

func principalOf(c *gin.Context) (base.Principal, bool) {
	principal, ok := base.PrincipalFromContext(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return base.Principal{}, false
	}
	return principal, true
}

func (h *handlers) list(c *gin.Context) {
	principal, ok := principalOf(c)
	if !ok {
		return
	}

	limit, _ := strconv.Atoi(c.Query("limit"))
	offset, _ := strconv.Atoi(c.Query("offset"))

	result, err := h.service.List(c.Request.Context(), principal.TenantID, ListFilter{
		Status: CustomerStatus(c.Query("status")),
		Search: c.Query("search"),
		Limit:  limit, Offset: offset,
	})
	if err != nil {
		respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result.Customers, "total": result.Total})
}

func (h *handlers) get(c *gin.Context) {
	principal, ok := principalOf(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid customer id"})
		return
	}

	customer, err := h.service.Get(c.Request.Context(), principal.TenantID, id)
	if err != nil {
		respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": customer})
}

func (h *handlers) create(c *gin.Context) {
	principal, ok := principalOf(c)
	if !ok {
		return
	}
	var request customerRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid customer payload"})
		return
	}

	customer, err := h.service.Create(c.Request.Context(), principal.TenantID, request.toInput())
	if err != nil {
		respondServiceError(c, err)
		return
	}

	h.record(c, principal, "crm.customer_created", customer.ID.String(), map[string]any{
		"code": customer.Code, "name": customer.Name,
	})
	c.JSON(http.StatusCreated, gin.H{"data": customer})
}

func (h *handlers) update(c *gin.Context) {
	principal, ok := principalOf(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid customer id"})
		return
	}
	var request customerRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid customer payload"})
		return
	}

	customer, err := h.service.Update(c.Request.Context(), principal.TenantID, id, request.toInput())
	if err != nil {
		respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": customer})
}

func (h *handlers) setStatus(c *gin.Context) {
	principal, ok := principalOf(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid customer id"})
		return
	}
	var request statusRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "status is required"})
		return
	}

	customer, err := h.service.SetStatus(c.Request.Context(), principal.TenantID, id, request.Status)
	if err != nil {
		respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": customer})
}

// record writes an audit row. Auditing must never fail the request it
// describes, so an error is logged rather than returned.
func (h *handlers) record(c *gin.Context, principal base.Principal, action, resourceID string, metadata map[string]any) {
	if h.audit == nil {
		return
	}
	if err := h.audit.Record(c.Request.Context(), principal, base.AuditEntry{
		Action: action, ResourceType: "customer", ResourceID: resourceID, Metadata: metadata,
	}); err != nil {
		slog.Error("audit write failed", "action", action, "error", err)
	}
}

func respondServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "customer not found"})
	case errors.Is(err, ErrDuplicate):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, ErrInvalidInput):
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "customer request failed"})
	}
}
