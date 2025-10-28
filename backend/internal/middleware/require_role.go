package middleware

import (
	"bluelink-backend/internal/models"

	"github.com/gin-gonic/gin"
)

// RequireRoleMiddleware 檢查使用者是否具有指定角色的 middleware
func RequireRoleMiddleware(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("Role")
		if !exists || userRole != role {
			models.RespondForbidden(c, "Forbidden: insufficient permissions")
			c.Abort()
			return
		}
		c.Next()
	}
}
