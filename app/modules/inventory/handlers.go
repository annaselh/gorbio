package inventory

import (
	"errors"
	"net/http"

	"github.com/annaselh/gorbio/modules/base"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type handlers struct {
	service *Service
}

type createItemRequest struct {
	SKU          string `json:"sku" binding:"required"`
	Name         string `json:"name" binding:"required"`
	Unit         string `json:"unit"`
	Quantity     int    `json:"quantity"`
	ReorderLevel int    `json:"reorder_level"`
}

type adjustRequest struct {
	// Delta is signed: negative issues stock, positive receives it. Pointer so
	// a missing field is rejected rather than silently treated as zero.
	Delta *int `json:"delta" binding:"required"`
}

func (h *handlers) list(c *gin.Context) {
	principal, ok := base.PrincipalFromContext(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	items, err := h.service.List(c.Request.Context(), principal.TenantID, ListFilter{
		LowStockOnly: c.Query("low_stock") == "true",
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list stock items"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items})
}

func (h *handlers) get(c *gin.Context) {
	principal, ok := base.PrincipalFromContext(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid stock item id"})
		return
	}

	item, err := h.service.Get(c.Request.Context(), principal.TenantID, id)
	if err != nil {
		respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": item})
}

func (h *handlers) create(c *gin.Context) {
	principal, ok := base.PrincipalFromContext(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	var request createItemRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid stock item payload"})
		return
	}

	item, err := h.service.Create(c.Request.Context(), principal.TenantID, CreateItemInput{
		SKU: request.SKU, Name: request.Name, Unit: request.Unit,
		Quantity: request.Quantity, ReorderLevel: request.ReorderLevel,
	})
	if err != nil {
		respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": item})
}

func (h *handlers) adjust(c *gin.Context) {
	principal, ok := base.PrincipalFromContext(c)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid stock item id"})
		return
	}

	var request adjustRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "delta is required"})
		return
	}

	item, err := h.service.Adjust(c.Request.Context(), principal.TenantID, id, *request.Delta)
	if err != nil {
		respondServiceError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": item})
}

func respondServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "stock item not found"})
	case errors.Is(err, ErrDuplicateSKU):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, ErrInvalidInput), errors.Is(err, ErrInsufficient):
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "inventory request failed"})
	}
}
