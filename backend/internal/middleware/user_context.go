package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// UserContextMiddleware 載入使用者基本資訊到 context
// 注意：這裡只載入輕量級的資訊（UserID, Role），不載入完整的使用者物件
func UserContextMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		walletAddress, exists := c.Get("WalletAddress")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    http.StatusUnauthorized,
				"message": "Wallet address not found in context",
			})
			c.Abort()
			return
		}

		// TODO: 從資料庫查詢使用者基本資訊
		// 這裡只是範例，實際應用需要注入 userService
		// userID, role, err := userService.GetBasicInfo(walletAddress.(string))
		// if err != nil {
		//     c.JSON(http.StatusNotFound, gin.H{
		//         "code":    http.StatusNotFound,
		//         "message": "User not found",
		//     })
		//     c.Abort()
		//     return
		// }

		// 暫時的模擬資料
		// 實際應用應該從資料庫查詢
		userID := int64(1)
		role := "investor"

		// 只存儲輕量級資訊到 context
		c.Set("UserID", userID)
		c.Set("Role", role)
		c.Set("WalletAddress", walletAddress)

		c.Next()
	}
}
