package middleware

import (
	"aimodels-prices/database"

	"github.com/gin-gonic/gin"
)

func Database() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("db", database.DB)
		c.Next()
	}
}
