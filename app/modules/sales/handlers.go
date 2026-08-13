package sales

import (
	"github.com/gin-gonic/gin"
)

func listOrders(c *gin.Context) {
	c.JSON(200, gin.H{
		"module": "sales",
		"data":   []string{},
	})
}
