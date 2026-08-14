package sales

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/annaselh/gorbio/modules/base"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type handlers struct {
	service *Service
}

type lineRequest struct {
	SKU         string `json:"sku" binding:"required"`
	Description string `json:"description" binding:"required"`
	Quantity    int    `json:"quantity" binding:"required"`
	UnitPrice   int64  `json:"unit_price"`
}

type createOrderRequest struct {
	Number       string        `json:"number"`
	CustomerName string        `json:"customer_name" binding:"required"`
	OrderDate    *time.Time    `json:"order_date"`
	Currency     string        `json:"currency"`
	Notes        string        `json:"notes"`
	Lines        []lineRequest `json:"lines" binding:"required,min=1"`
}

type updateStatusRequest struct {
	Status OrderStatus `json:"status" binding:"required"`
}

func (h *handlers) list(c *gin.Context) {
	principal, ok := base.PrincipalFromContext(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	limit, _ := strconv.Atoi(c.Query("limit"))
	offset, _ := strconv.Atoi(c.Query("offset"))

	result, err := h.service.List(c.Request.Context(), principal.TenantID, ListFilter{
		Status:   OrderStatus(c.Query("status")),
		Customer: c.Query("customer"),
		Limit:    limit,
		Offset:   offset,
	})
	if err != nil {
		respondServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": result.Orders, "total": result.Total})
}

func (h *handlers) get(c *gin.Context) {
	principal, ok := base.PrincipalFromContext(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sales order id"})
		return
	}

	order, err := h.service.Get(c.Request.Context(), principal.TenantID, id)
	if err != nil {
		respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": order})
}

func (h *handlers) create(c *gin.Context) {
	principal, ok := base.PrincipalFromContext(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	var request createOrderRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sales order payload"})
		return
	}

	lines := make([]LineInput, 0, len(request.Lines))
	for _, line := range request.Lines {
		lines = append(lines, LineInput{
			SKU: line.SKU, Description: line.Description,
			Quantity: line.Quantity, UnitPrice: line.UnitPrice,
		})
	}

	input := CreateOrderInput{
		Number: request.Number, CustomerName: request.CustomerName,
		Currency: request.Currency, Notes: request.Notes, Lines: lines,
	}
	if request.OrderDate != nil {
		input.OrderDate = *request.OrderDate
	}

	order, err := h.service.Create(c.Request.Context(), principal.TenantID, input)
	if err != nil {
		respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": order})
}

func (h *handlers) updateStatus(c *gin.Context) {
	principal, ok := base.PrincipalFromContext(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sales order id"})
		return
	}

	var request updateStatusRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "status is required"})
		return
	}

	order, err := h.service.UpdateStatus(c.Request.Context(), principal.TenantID, id, request.Status)
	if err != nil {
		respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": order})
}

func respondServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "sales order not found"})
	case errors.Is(err, ErrDuplicate):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, ErrInvalidInput), errors.Is(err, ErrNotEditable):
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "sales request failed"})
	}
}
