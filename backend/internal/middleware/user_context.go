package middleware

import (
	"bluelink-backend/internal/session"
	"net/http"

	"github.com/gin-gonic/gin"
)

// UserContextMiddleware 從session載入使用者基本資訊到 context
func BasicUserContextMiddleware(sessionManager *session.MemorySessionManager) gin.HandlerFunc {
	return func(c *gin.Context) {

		/*
			type UserBasicInfo struct {
				ID            int64  `json:"id"`
				WalletAddress string `json:"wallet_address"`
				Role          string `json:"role"` // "buyer", "issuer", "admin"
			}
		*/

		sessionID, exists := c.Get("SessionID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    http.StatusUnauthorized,
				"message": "Session ID not found in context",
			})
			c.Abort()
			return
		}

		sess, err := sessionManager.Get(sessionID.(string))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
			c.Abort()
			return
		}

		c.Set("UserID", sess.UserID)
		c.Set("Role", sess.Role)
		c.Set("WalletAddress", sess.WalletAddress)

		c.Next()
	}
}

// // FullUserContextMiddleware 需要完整 User 資料時才用
// func FullUserMiddleware(userService *services.UserService) gin.HandlerFunc {
//     return func(c *gin.Context) {
//         userID, exists := c.Get("UserID")
//         if !exists {
//             c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
//             c.Abort()
//             return
//         }

//         // 查詢完整使用者資料
//         user, err := userService.GetByID(userID.(int64))
//         if err != nil {
//             c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
//             c.Abort()
//             return
//         }

//         c.Set("User", user)
//         c.Next()
//     }
// }
