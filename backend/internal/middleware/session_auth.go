package middleware

import (
	"bluelink-backend/internal/models"
	"bluelink-backend/internal/session"
	"strings"

	"github.com/gin-gonic/gin"
)

/*
   登入流程：
   1. 前端 → POST /auth/challenge (wallet_address)
      後端 → 產生 nonce，儲存到內存，回傳 nonce + message

   2. 前端 → 使用錢包簽署 message
      錢包 → 回傳 signature

   3. 前端 → POST /auth/verify (wallet_address, signature, nonce)
      後端 → 驗證 Sui 簽名
      後端 → 建立 Session，儲存到內存
      後端 → 設定 HttpOnly Cookie (session_id)
      後端 → 回傳 session_id + user 資料

   後續請求：
   前端 → 請求時自動帶上 Cookie (session_id)
   後端 → SessionAuthMiddleware 驗證 session
   後端 → UserContextMiddleware 載入完整 user 資料
   後端 → Handler 從 context 取得 WalletAddress, UserID, User

   登出流程：
   前端 → POST /auth/logout
   後端 → 刪除 session，清除 Cookie

   安全機制：
   - Nonce 5 分鐘過期，使用後立即刪除（防重放攻擊）
   - Session 24 小時過期，30 分鐘閒置自動登出
   - 最多 3 個裝置同時登入
   - HttpOnly Cookie（防 XSS）
   - Secure Cookie（HTTPS only）
   - 記錄 IP 和 UserAgent（審計追蹤）
*/

// SessionAuthMiddleware 驗證 Session
func SessionAuthMiddleware(sessionManager *session.MemorySessionManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID := getSessionID(c)
		if sessionID == "" {
			models.RespondUnauthorized(c, "Missing session")
			c.Abort()
			return
		}

		sess, err := sessionManager.Get(sessionID)
		if err != nil {
			models.RespondUnauthorized(c, "Invalid or expired session")
			c.Abort()
			return
		}

		// 更新活躍時間
		sessionManager.Touch(sessionID)

		// 存入 context (包含所有需要的使用者資訊)
		c.Set("SessionID", sessionID)
		c.Set("WalletAddress", sess.WalletAddress)
		c.Set("UserID", sess.UserID)
		c.Set("Role", sess.Role)

		c.Next()
	}
}

func getSessionID(c *gin.Context) string {
	// 優先從 Cookie
	sessionID, err := c.Cookie("session_id")
	if err == nil && sessionID != "" {
		return sessionID
	}

	// 其次從 Header
	authHeader := c.GetHeader("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}

	return ""
}
