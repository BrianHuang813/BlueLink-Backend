package auth

import (
	"bluelink-backend/internal/models"
	"bluelink-backend/internal/repository"
	"bluelink-backend/internal/services"
	"bluelink-backend/internal/session"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	suiModels "github.com/block-vision/sui-go-sdk/models"
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

type AuthHandler struct {
	userService    *services.UserService
	sessionManager session.SessionManager
	nonceRepo      *repository.NonceRepository
	isProduction   bool // 從配置讀取的環境標誌
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
	Role          string `json:"role"` // 可選：使用者角色 (buyer/issuer)，首次註冊時使用
}

type VerifyResponse struct {
	SessionID     string       `json:"session_id"`
	WalletAddress string       `json:"wallet_address"`
	User          *models.User `json:"user"`
	ExpiresAt     int64        `json:"expires_at"`
}

func NewAuthHandler(userService *services.UserService, sessionManager session.SessionManager, nonceRepo *repository.NonceRepository, isProduction bool) *AuthHandler {
	return &AuthHandler{
		userService:    userService,
		sessionManager: sessionManager,
		nonceRepo:      nonceRepo,
		isProduction:   isProduction,
	}
}

// GenerateChallenge 產生挑戰訊息
// POST /api/v1/auth/challenge
func (h *AuthHandler) GenerateChallenge(c *gin.Context) {
	var req ChallengeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		models.RespondBadRequest(c, "Invalid request format", err)
		return
	}

	// 生成隨機 nonce（32 bytes base64 編碼）
	nonceBytes := make([]byte, 32)
	if _, err := rand.Read(nonceBytes); err != nil {
		models.RespondInternalError(c, "Failed to generate nonce", err)
		return
	}
	nonce := base64.URLEncoding.EncodeToString(nonceBytes)

	// 儲存 nonce 到資料庫，使用 wallet_address 作為 KEY，TTL 設置為 10 分鐘
	if err := h.nonceRepo.Create(c.Request.Context(), req.WalletAddress, nonce, 10*time.Minute); err != nil {
		models.RespondInternalError(c, "Failed to store nonce", err)
		return
	}

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
	// 安全檢查：確保 handler 已正確初始化
	if h == nil {
		fmt.Printf("[HANDLER ERROR] handler is nil\n")
		models.RespondInternalError(c, "Handler not initialized", fmt.Errorf("handler is nil"))
		return
	}
	if h.userService == nil {
		fmt.Printf("[HANDLER ERROR] userService is nil\n")
		models.RespondInternalError(c, "User service not initialized", fmt.Errorf("userService is nil"))
		return
	}
	if h.sessionManager == nil {
		fmt.Printf("[HANDLER ERROR] sessionManager is nil\n")
		models.RespondInternalError(c, "Session manager not initialized", fmt.Errorf("sessionManager is nil"))
		return
	}
	if h.nonceRepo == nil {
		fmt.Printf("[HANDLER ERROR] nonceRepo is nil\n")
		models.RespondInternalError(c, "Nonce repository not initialized", fmt.Errorf("nonceRepo is nil"))
		return
	}

	var req VerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		models.RespondBadRequest(c, "Invalid request format", err)
		return
	}

	// 1. 驗證 nonce（使用 wallet_address 作為 KEY 查找）
	// Verify 方法會自動檢查過期、比對 nonce、並在驗證成功後刪除（防止重放攻擊）
	isValid, err := h.nonceRepo.Verify(c.Request.Context(), req.WalletAddress, req.Nonce)
	if err != nil || !isValid {
		// 記錄詳細錯誤
		fmt.Printf("[NONCE ERROR] wallet=%s, error=%v\n", req.WalletAddress, err)
		models.RespondUnauthorized(c, fmt.Sprintf("Nonce verification failed: %v", err))
		return
	}
	fmt.Printf("[NONCE OK] wallet=%s verified successfully\n", req.WalletAddress)

	// 2. 驗證 Sui 簽名
	// 構建相同的 message，確保與 challenge 時的格式一致
	message := fmt.Sprintf("Sign in to BlueLink\nNonce: %s", req.Nonce)
	fmt.Printf("[SIG VERIFY] message=%s, signature_len=%d\n", message, len(req.Signature))
	isSigValid, signerAddress, err := h.verifySuiSignature(req.Signature, message)
	if err != nil || !isSigValid {
		fmt.Printf("[SIG ERROR] error=%v, valid=%v\n", err, isSigValid)
		models.RespondWithErrorDetails(c, http.StatusUnauthorized, "Invalid signature",
			fmt.Sprintf("Verification failed: %v", err))
		return
	}
	fmt.Printf("[SIG OK] signer=%s\n", signerAddress)

	// 3. 驗證簽名者地址是否與提供的地址匹配
	if signerAddress != req.WalletAddress {
		models.RespondWithErrorDetails(c, http.StatusUnauthorized, "Address mismatch",
			fmt.Sprintf("Provided %s, signed by %s", req.WalletAddress, signerAddress))
		return
	}

	// 5. 取得或建立使用者
	user, err := h.userService.GetByWalletAddress(c.Request.Context(), req.WalletAddress)
	if err != nil {
		// 使用者不存在，建立新使用者
		// 如果前端提供了角色，使用該角色;否則預設為 buyer
		role := req.Role
		if role == "" {
			role = "buyer" // 預設角色
		}

		fmt.Printf("[USER CREATE] wallet=%s, role=%s\n", req.WalletAddress, role)
		user, err = h.userService.CreateWithRole(c.Request.Context(), req.WalletAddress, role)
		if err != nil {
			fmt.Printf("[USER ERROR] failed to create user: %v\n", err)
			models.RespondInternalError(c, "Failed to create user", err)
			return
		}

		// 確保 user 不是 nil
		if user == nil {
			fmt.Printf("[USER ERROR] user is nil after creation\n")
			models.RespondInternalError(c, "Failed to create user", fmt.Errorf("user is nil"))
			return
		}
	}

	// 額外的安全檢查：確保 user 不是 nil
	if user == nil {
		fmt.Printf("[USER ERROR] user is nil after get/create\n")
		models.RespondInternalError(c, "User data is invalid", fmt.Errorf("user is nil"))
		return
	}

	fmt.Printf("[USER OK] user_id=%d, wallet=%s\n", user.ID, user.WalletAddress)

	// 6. 建立 session
	fmt.Printf("[SESSION CREATE] user_id=%d, wallet=%s\n", user.ID, req.WalletAddress)
	sess, err := h.sessionManager.Create(
		user.ID,
		req.WalletAddress,
		user.Role,
		//user.KYCStatus,
		c.ClientIP(),
		c.Request.UserAgent(),
	)
	if err != nil {
		fmt.Printf("[SESSION ERROR] failed to create session: %v\n", err)
		models.RespondInternalError(c, "Failed to create session", err)
		return
	}
	fmt.Printf("[SESSION OK] session_id=%s\n", sess.ID)

	// 7. 設定 HttpOnly Cookie
	cookie := &http.Cookie{
		Name:     "session_id",
		Value:    sess.ID,
		Path:     "/",
		Domain:   "",
		MaxAge:   int(24 * time.Hour.Seconds()),
		Secure:   h.isProduction,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(c.Writer, cookie)

	// 8. 回傳成功響應
	models.RespondWithSuccess(c, http.StatusOK, "Authentication successful", VerifyResponse{
		SessionID:     sess.ID,
		WalletAddress: req.WalletAddress,
		User:          user,
		ExpiresAt:     sess.CreatedAt.Add(24 * time.Hour).Unix(),
	})
}

// Logout 登出當前會話
// POST /api/v1/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	sessionID, exists := c.Get("SessionID")
	if !exists {
		models.RespondWithSuccess(c, http.StatusOK, "Already logged out", nil)
		return
	}

	// 刪除 session
	if err := h.sessionManager.Delete(sessionID.(string)); err != nil {
		models.RespondInternalError(c, "Failed to logout", err)
		return
	}

	// 清除 Cookie
	c.SetCookie("session_id", "", -1, "/", "", true, true)

	models.RespondWithSuccess(c, http.StatusOK, "Logged out successfully", nil)
}

// LogoutAll 登出所有裝置
// POST /api/v1/auth/logout-all
func (h *AuthHandler) LogoutAll(c *gin.Context) {
	walletAddress, exists := c.Get("WalletAddress")
	if !exists {
		models.RespondUnauthorized(c, "Unauthorized")
		return
	}

	// 刪除所有 session
	if err := h.sessionManager.DeleteAllUserSessions(walletAddress.(string)); err != nil {
		models.RespondInternalError(c, "Failed to logout all sessions", err)
		return
	}

	// 清除當前 Cookie
	c.SetCookie("session_id", "", -1, "/", "", true, true)

	models.RespondWithSuccess(c, http.StatusOK, "Logged out from all devices", nil)
}

// GetActiveSessions 取得使用者的所有活躍會話
// GET /api/v1/sessions
func (h *AuthHandler) GetActiveSessions(c *gin.Context) {
	walletAddress, exists := c.Get("WalletAddress")
	if !exists {
		models.RespondUnauthorized(c, "Unauthorized")
		return
	}

	sessions, err := h.sessionManager.GetUserSessions(walletAddress.(string))
	if err != nil {
		models.RespondInternalError(c, "Failed to get sessions", err)
		return
	}

	models.RespondWithSuccess(c, http.StatusOK, "Sessions retrieved successfully", sessions)
}

// RevokeSession 撤銷指定的會話
// DELETE /api/v1/sessions/:session_id
func (h *AuthHandler) RevokeSession(c *gin.Context) {
	sessionIDToRevoke := c.Param("session_id")
	if sessionIDToRevoke == "" {
		models.RespondBadRequest(c, "Session ID is required", nil)
		return
	}

	walletAddress, exists := c.Get("WalletAddress")
	if !exists {
		models.RespondUnauthorized(c, "Unauthorized")
		return
	}

	// 驗證該 session 是否屬於當前使用者
	session, err := h.sessionManager.Get(sessionIDToRevoke)
	if err != nil || session == nil {
		models.RespondNotFound(c, "Session not found")
		return
	}

	if session.WalletAddress != walletAddress.(string) {
		models.RespondForbidden(c, "Cannot revoke another user's session")
		return
	}

	// 刪除 session
	if err := h.sessionManager.Delete(sessionIDToRevoke); err != nil {
		models.RespondInternalError(c, "Failed to revoke session", err)
		return
	}

	models.RespondWithSuccess(c, http.StatusOK, "Session revoked successfully", nil)
}

// verifySuiSignature 使用 SDK 的 models 套件驗證 Sui 簽名
func (h *AuthHandler) verifySuiSignature(signatureB64, message string) (bool, string, error) {
	// 使用 SDK 的 VerifyPersonalMessage 函數
	// 這個函數接受:
	// 1. message: 原始訊息字符串(不需要 base64 編碼)
	// 2. signature: base64 編碼的序列化簽名 [flag_byte][signature_bytes][pubkey_bytes]
	// 返回: (signer_address, pass, error)

	signerAddress, pass, err := suiModels.VerifyPersonalMessage(message, signatureB64)
	if err != nil {
		return false, "", fmt.Errorf("signature verification failed: %w", err)
	}

	if !pass {
		return false, "", fmt.Errorf("signature verification failed: invalid signature")
	}

	// 回傳驗證結果和簽名者地址
	return true, signerAddress, nil
}
