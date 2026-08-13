package salesdiscount

import "github.com/gin-gonic/gin"

func applyDiscount(c *gin.Context) {
	c.JSON(200, gin.H{
		"extension": "discount",
		"data":      []string{},
	})
}
