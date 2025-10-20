package auth

import (
	"bluelink-backend/internal/models"
	"bluelink-backend/internal/services"
	"bluelink-backend/internal/session"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/block-vision/sui-go-sdk/verify"
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

var isProduction = false // TODO: 根據環境設定
type AuthHandler struct {
	userService    *services.UserService
	sessionManager *session.MemorySessionManager
	nonces         map[string]NonceData
	mu             sync.RWMutex
}

type NonceData struct {
	Nonce     string
	Timestamp time.Time
}

type ChallengeRequest struct {
	WalletAddress string `json:"wallet_address" binding:"required"`
}

type ChallengeResponse struct {
	Nonce   string `json:"nonce"`
	Message string `json:"message"`
}

type VerifyRequest struct {
	WalletAddress string `json:"wallet_address" binding:"required"`
	Signature     string `json:"signature" binding:"required"` // Base64 編碼的簽名
	Nonce         string `json:"nonce" binding:"required"`
}

type VerifyResponse struct {
	SessionID     string       `json:"session_id"`
	WalletAddress string       `json:"wallet_address"`
	User          *models.User `json:"user"`
	ExpiresAt     int64        `json:"expires_at"`
}

func NewAuthHandler(userService *services.UserService, sessionManager *session.MemorySessionManager) *AuthHandler {
	handler := &AuthHandler{
		userService:    userService,
		sessionManager: sessionManager,
		nonces:         make(map[string]NonceData),
	}

	// 啟動清理協程
	go handler.cleanupExpiredNonces()

	return handler
}

// GenerateChallenge 產生挑戰訊息
// POST /api/v1/auth/challenge
func (h *AuthHandler) GenerateChallenge(c *gin.Context) {
	var req ChallengeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithDetails(
			http.StatusBadRequest,
			"Invalid request format",
			err.Error(),
		))
		return
	}

	// 生成隨機 nonce
	nonceBytes := make([]byte, 32)
	if _, err := rand.Read(nonceBytes); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to generate nonce",
		})
		return
	}
	nonce := base64.URLEncoding.EncodeToString(nonceBytes)

	// 儲存 nonce
	key := req.WalletAddress
	h.mu.Lock()
	h.nonces[key] = NonceData{
		Nonce:     nonce,
		Timestamp: time.Now(),
	}
	h.mu.Unlock()

	// 構建要簽名的訊息（前端會簽署這個訊息）
	message := fmt.Sprintf("Sign in to BlueLink\nNonce: %s", nonce)

	c.JSON(http.StatusOK, ChallengeResponse{
		Nonce:   nonce,
		Message: message,
	})
}

// VerifySignature 驗證 Sui 錢包簽名
// POST /api/v1/auth/verify
func (h *AuthHandler) VerifySignature(c *gin.Context) {
	var req VerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponseWithDetails(
			http.StatusBadRequest,
			"Invalid request format",
			err.Error(),
		))
		return
	}

	// 1. 驗證 nonce
	key := req.WalletAddress
	h.mu.RLock()
	nonceData, exists := h.nonces[key]
	h.mu.RUnlock()

	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Code:    http.StatusUnauthorized,
			Message: "Nonce not found or expired",
		})
		return
	}

	// 檢查 nonce 是否匹配
	if nonceData.Nonce != req.Nonce {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Code:    http.StatusUnauthorized,
			Message: "Invalid nonce",
		})
		return
	}

	// 檢查是否過期（5 分鐘）
	if time.Since(nonceData.Timestamp) > 5*time.Minute {
		h.mu.Lock()
		delete(h.nonces, key)
		h.mu.Unlock()

		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Code:    http.StatusUnauthorized,
			Message: "Nonce expired",
		})
		return
	}

	// 2. 驗證 Sui 簽名
	message := fmt.Sprintf("Sign in to BlueLink\nNonce: %s", req.Nonce)
	isValid, signerAddress, err := h.verifySuiSignature(req.Signature, message)
	if err != nil || !isValid {
		c.JSON(http.StatusUnauthorized, models.ErrorResponseWithDetails(
			http.StatusUnauthorized,
			"Invalid signature",
			fmt.Sprintf("Verification failed: %v", err),
		))
		return
	}

	// 3. 驗證簽名者地址是否與提供的地址匹配
	if signerAddress != req.WalletAddress {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Code: http.StatusUnauthorized,
			Message: fmt.Sprintf("Address mismatch: provided %s, signed by %s",
				req.WalletAddress, signerAddress),
		})
		return
	}

	// 4. 刪除已使用的 nonce（防止重放攻擊）
	h.mu.Lock()
	delete(h.nonces, key)
	h.mu.Unlock()

	// 5. 取得或建立使用者
	user, err := h.userService.GetByWalletAddress(req.WalletAddress)
	if err != nil {
		// 使用者不存在，建立新使用者
		user, err = h.userService.Create(req.WalletAddress)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse{
				Code:    http.StatusInternalServerError,
				Message: "Failed to create user",
			})
			return
		}
	}

	// 6. 建立 session
	sess, err := h.sessionManager.Create(
		user.ID,
		req.WalletAddress,
		user.Role,
		//user.KYCStatus,
		c.ClientIP(),
		c.Request.UserAgent(),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to create session",
		})
		return
	}

	// 7. 設定 HttpOnly Cookie
	cookie := &http.Cookie{
		Name:     "session_id",
		Value:    sess.ID,
		Path:     "/",
		Domain:   "",
		MaxAge:   int(24 * time.Hour.Seconds()),
		Secure:   isProduction,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(c.Writer, cookie)

	// 8. 回傳成功響應
	c.JSON(http.StatusOK, models.SuccessResponse(
		http.StatusOK,
		"Authentication successful",
		VerifyResponse{
			SessionID:     sess.ID,
			WalletAddress: req.WalletAddress,
			User:          user,
			ExpiresAt:     sess.CreatedAt.Add(24 * time.Hour).Unix(),
		},
	))
}

// Logout 登出當前會話
// POST /api/v1/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	sessionID, exists := c.Get("SessionID")
	if !exists {
		c.JSON(http.StatusOK, models.SuccessResponse(
			http.StatusOK,
			"Already logged out",
			nil,
		))
		return
	}

	// 刪除 session
	if err := h.sessionManager.Delete(sessionID.(string)); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to logout",
		})
		return
	}

	// 清除 Cookie
	c.SetCookie("session_id", "", -1, "/", "", true, true)

	c.JSON(http.StatusOK, models.SuccessResponse(
		http.StatusOK,
		"Logged out successfully",
		nil,
	))
}

// LogoutAll 登出所有裝置
// POST /api/v1/auth/logout-all
func (h *AuthHandler) LogoutAll(c *gin.Context) {
	walletAddress, exists := c.Get("WalletAddress")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Code:    http.StatusUnauthorized,
			Message: "Unauthorized",
		})
		return
	}

	// 刪除所有 session
	if err := h.sessionManager.DeleteAllUserSessions(walletAddress.(string)); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to logout all sessions",
		})
		return
	}

	// 清除當前 Cookie
	c.SetCookie("session_id", "", -1, "/", "", true, true)

	c.JSON(http.StatusOK, models.SuccessResponse(
		http.StatusOK,
		"Logged out from all devices",
		nil,
	))
}

// GetActiveSessions 取得所有活躍會話
// GET /api/v1/auth/sessions
func (h *AuthHandler) GetActiveSessions(c *gin.Context) {
	walletAddress, exists := c.Get("WalletAddress")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Code:    http.StatusUnauthorized,
			Message: "Unauthorized",
		})
		return
	}

	sessions, err := h.sessionManager.GetUserSessions(walletAddress.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to get sessions",
		})
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(
		http.StatusOK,
		"Sessions retrieved successfully",
		sessions,
	))
}

// verifySuiSignature 使用 SDK 的 verify 套件驗證 Sui 簽名
func (h *AuthHandler) verifySuiSignature(signatureB64, message string) (bool, string, error) {
	// 1. 解碼 Base64 簽名
	signatureBytes, err := base64.StdEncoding.DecodeString(signatureB64)
	if err != nil {
		return false, "", fmt.Errorf("failed to decode signature: %w", err)
	}

	// 2. 準備訊息（轉為 bytes）
	messageBytes := []byte(message)

	// 3. 使用 SDK 的 VerifyPersonalMessageSignature
	// 注意：如果不需要 zkLogin，options 可以傳 nil
	signerAddress, pass, err := verify.VerifyPersonalMessageSignature(
		messageBytes,
		signatureBytes,
		nil, // zkLogin options (一般錢包不需要，傳 nil)
	)
	if err != nil {
		return false, "", fmt.Errorf("signature verification failed: %w", err)
	}

	if !pass {
		return false, "", fmt.Errorf("signature verification failed: invalid signature")
	}

	// 4. 回傳驗證結果和簽名者地址
	return true, signerAddress, nil
}

// cleanupExpiredNonces 定期清理過期的 nonce
func (h *AuthHandler) cleanupExpiredNonces() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		h.mu.Lock()
		now := time.Now()
		for key, data := range h.nonces {
			if now.Sub(data.Timestamp) > 5*time.Minute {
				delete(h.nonces, key)
			}
		}
		h.mu.Unlock()
	}
}
