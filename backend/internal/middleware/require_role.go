package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RequireRoleMiddleware 檢查使用者是否具有指定角色的 middleware
func RequireRoleMiddleware(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("Role")
		if !exists || userRole != role {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    http.StatusForbidden,
				"message": "Forbidden: insufficient permissions",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
