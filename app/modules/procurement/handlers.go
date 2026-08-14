package procurement

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/annaselh/gorbio/modules/base"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type handlers struct {
	service *Service
	audit   *base.AuditService
}

type vendorRequest struct {
	Code            string `json:"code"`
	Name            string `json:"name" binding:"required"`
	Email           string `json:"email"`
	Phone           string `json:"phone"`
	Address         string `json:"address"`
	TaxID           string `json:"tax_id"`
	PaymentTermDays int    `json:"payment_term_days"`
	Notes           string `json:"notes"`
}

type vendorStatusRequest struct {
	Status VendorStatus `json:"status" binding:"required"`
}

type lineRequest struct {
	SKU         string `json:"sku" binding:"required"`
	Description string `json:"description" binding:"required"`
	Quantity    int    `json:"quantity" binding:"required"`
	UnitPrice   int64  `json:"unit_price"`
}

type createPurchaseRequest struct {
	Number       string        `json:"number"`
	VendorID     string        `json:"vendor_id" binding:"required"`
	OrderDate    *time.Time    `json:"order_date"`
	ExpectedDate *time.Time    `json:"expected_date"`
	Currency     string        `json:"currency"`
	Notes        string        `json:"notes"`
	Lines        []lineRequest `json:"lines" binding:"required,min=1"`
}

// updatePurchaseRequest carries the order as it should read after the edit.
// The number is absent on purpose: it is the order's identity, not a field.
type updatePurchaseRequest struct {
	VendorID     string        `json:"vendor_id" binding:"required"`
	OrderDate    *time.Time    `json:"order_date"`
	ExpectedDate *time.Time    `json:"expected_date"`
	Currency     string        `json:"currency"`
	Notes        string        `json:"notes"`
	Lines        []lineRequest `json:"lines" binding:"required,min=1"`
}

type purchaseStatusRequest struct {
	Status PurchaseStatus `json:"status" binding:"required"`
}

// lineInputs converts the wire lines to service inputs. The service is what
// validates them; this only changes their shape.
func lineInputs(requests []lineRequest) []LineInput {
	lines := make([]LineInput, 0, len(requests))
	for _, line := range requests {
		lines = append(lines, LineInput{
			SKU: line.SKU, Description: line.Description,
			Quantity: line.Quantity, UnitPrice: line.UnitPrice,
		})
	}
	return lines
}

func principalOf(c *gin.Context) (base.Principal, bool) {
	principal, ok := base.PrincipalFromContext(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return base.Principal{}, false
	}
	return principal, true
}

// ---------------------------------------------------------------- vendors

func (h *handlers) listVendors(c *gin.Context) {
	principal, ok := principalOf(c)
	if !ok {
		return
	}

	limit, _ := strconv.Atoi(c.Query("limit"))
	offset, _ := strconv.Atoi(c.Query("offset"))

	result, err := h.service.ListVendors(c.Request.Context(), principal.TenantID, VendorFilter{
		Status: VendorStatus(c.Query("status")),
		Search: c.Query("search"),
		Limit:  limit, Offset: offset,
	})
	if err != nil {
		respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result.Vendors, "total": result.Total})
}

func (h *handlers) getVendor(c *gin.Context) {
	principal, ok := principalOf(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid vendor id"})
		return
	}

	vendor, err := h.service.GetVendor(c.Request.Context(), principal.TenantID, id)
	if err != nil {
		respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": vendor})
}

func (h *handlers) createVendor(c *gin.Context) {
	principal, ok := principalOf(c)
	if !ok {
		return
	}
	var request vendorRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid vendor payload"})
		return
	}

	vendor, err := h.service.CreateVendor(c.Request.Context(), principal.TenantID, request.toInput())
	if err != nil {
		respondServiceError(c, err)
		return
	}

	h.record(c, principal, "procurement.vendor_created", "vendor", vendor.ID.String(), map[string]any{
		"code": vendor.Code, "name": vendor.Name,
	})
	c.JSON(http.StatusCreated, gin.H{"data": vendor})
}

func (h *handlers) updateVendor(c *gin.Context) {
	principal, ok := principalOf(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid vendor id"})
		return
	}
	var request vendorRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid vendor payload"})
		return
	}

	vendor, err := h.service.UpdateVendor(c.Request.Context(), principal.TenantID, id, request.toInput())
	if err != nil {
		respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": vendor})
}

func (h *handlers) setVendorStatus(c *gin.Context) {
	principal, ok := principalOf(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid vendor id"})
		return
	}
	var request vendorStatusRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "status is required"})
		return
	}

	vendor, err := h.service.SetVendorStatus(c.Request.Context(), principal.TenantID, id, request.Status)
	if err != nil {
		respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": vendor})
}

func (r vendorRequest) toInput() VendorInput {
	return VendorInput{
		Code: r.Code, Name: r.Name, Email: r.Email, Phone: r.Phone,
		Address: r.Address, TaxID: r.TaxID, PaymentTerm: r.PaymentTermDays, Notes: r.Notes,
	}
}

// --------------------------------------------------------- purchase orders

func (h *handlers) listPurchaseOrders(c *gin.Context) {
	principal, ok := principalOf(c)
	if !ok {
		return
	}

	limit, _ := strconv.Atoi(c.Query("limit"))
	offset, _ := strconv.Atoi(c.Query("offset"))

	filter := PurchaseFilter{
		Status: PurchaseStatus(c.Query("status")),
		Limit:  limit, Offset: offset,
	}
	if raw := c.Query("vendor_id"); raw != "" {
		vendorID, err := uuid.Parse(raw)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid vendor id"})
			return
		}
		filter.VendorID = vendorID
	}

	result, err := h.service.ListPurchaseOrders(c.Request.Context(), principal.TenantID, filter)
	if err != nil {
		respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result.Orders, "total": result.Total})
}

func (h *handlers) getPurchaseOrder(c *gin.Context) {
	principal, ok := principalOf(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid purchase order id"})
		return
	}

	order, err := h.service.GetPurchaseOrder(c.Request.Context(), principal.TenantID, id)
	if err != nil {
		respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": order})
}

func (h *handlers) createPurchaseOrder(c *gin.Context) {
	principal, ok := principalOf(c)
	if !ok {
		return
	}
	var request createPurchaseRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid purchase order payload"})
		return
	}
	vendorID, err := uuid.Parse(request.VendorID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid vendor id"})
		return
	}

	input := CreatePurchaseInput{
		Number: request.Number, VendorID: vendorID, ExpectedDate: request.ExpectedDate,
		Currency: request.Currency, Notes: request.Notes,
		Lines: lineInputs(request.Lines),
	}
	if request.OrderDate != nil {
		input.OrderDate = *request.OrderDate
	}

	order, err := h.service.CreatePurchaseOrder(c.Request.Context(), principal.TenantID, input)
	if err != nil {
		respondServiceError(c, err)
		return
	}

	h.record(c, principal, "procurement.order_created", "purchase_order", order.ID.String(), map[string]any{
		"number": order.Number, "vendor": order.VendorName, "total": order.Total,
	})
	c.JSON(http.StatusCreated, gin.H{"data": order})
}

func (h *handlers) updatePurchaseOrder(c *gin.Context) {
	principal, ok := principalOf(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid purchase order id"})
		return
	}

	var request updatePurchaseRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid purchase order payload"})
		return
	}
	vendorID, err := uuid.Parse(request.VendorID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid vendor id"})
		return
	}

	input := UpdatePurchaseInput{
		VendorID: vendorID, ExpectedDate: request.ExpectedDate,
		Currency: request.Currency, Notes: request.Notes,
		Lines: lineInputs(request.Lines),
	}
	if request.OrderDate != nil {
		input.OrderDate = *request.OrderDate
	}

	order, err := h.service.UpdatePurchaseOrder(c.Request.Context(), principal.TenantID, id, input)
	if err != nil {
		respondServiceError(c, err)
		return
	}

	h.record(c, principal, "procurement.order_updated", "purchase_order", order.ID.String(), map[string]any{
		"number": order.Number, "vendor": order.VendorName, "total": order.Total,
	})
	c.JSON(http.StatusOK, gin.H{"data": order})
}

func (h *handlers) updatePurchaseStatus(c *gin.Context) {
	principal, ok := principalOf(c)
	if !ok {
		return
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid purchase order id"})
		return
	}
	var request purchaseStatusRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "status is required"})
		return
	}

	order, err := h.service.UpdatePurchaseStatus(c.Request.Context(), principal.TenantID, id, request.Status)
	if err != nil {
		respondServiceError(c, err)
		return
	}

	h.record(c, principal, "procurement.order_"+strings.ToLower(string(order.Status)),
		"purchase_order", order.ID.String(), map[string]any{
			"number": order.Number, "vendor": order.VendorName, "total": order.Total,
		})
	c.JSON(http.StatusOK, gin.H{"data": order})
}

// record writes an audit row. Auditing must never fail the request it
// describes, so an error is logged rather than returned.
func (h *handlers) record(c *gin.Context, principal base.Principal, action, resourceType, resourceID string, metadata map[string]any) {
	if h.audit == nil {
		return
	}
	if err := h.audit.Record(c.Request.Context(), principal, base.AuditEntry{
		Action: action, ResourceType: resourceType, ResourceID: resourceID, Metadata: metadata,
	}); err != nil {
		slog.Error("audit write failed", "action", action, "error", err)
	}
}

func respondServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrVendorNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "vendor not found"})
	case errors.Is(err, ErrPurchaseNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "purchase order not found"})
	case errors.Is(err, ErrDuplicate):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, ErrInvalidInput), errors.Is(err, ErrNotEditable), errors.Is(err, ErrVendorInactive):
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "procurement request failed"})
	}
}
