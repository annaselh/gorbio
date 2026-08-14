package salesdiscount

import (
	"errors"
	"net/http"

	"github.com/annaselh/gorbio/modules/base"
	"github.com/annaselh/gorbio/modules/sales"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type handlers struct {
	orders *sales.Service
}

type applyDiscountRequest struct {
	// Amount is an absolute discount in minor currency units, matching how the
	// sales module stores money. Pointer so an omitted field is rejected rather
	// than silently clearing the discount.
	Amount *int64 `json:"amount" binding:"required"`
}

func (h *handlers) applyDiscount(c *gin.Context) {
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

	var request applyDiscountRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "discount amount is required"})
		return
	}

	order, err := h.orders.ApplyDiscount(c.Request.Context(), principal.TenantID, id, *request.Amount)
	if err != nil {
		switch {
		case errors.Is(err, sales.ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "sales order not found"})
		case errors.Is(err, sales.ErrInvalidInput), errors.Is(err, sales.ErrNotEditable):
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not apply discount"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": order})
}
