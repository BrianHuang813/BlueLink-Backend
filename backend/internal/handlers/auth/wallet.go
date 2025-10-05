package auth

import (
	"bluelink-backend/internal/models"
	"bluelink-backend/internal/services"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// AuthHandler 處理錢包驗證相關的請求
type AuthHandler struct {
	userService *services.UserService
	challenges  map[string]ChallengeData
	mu          sync.RWMutex
}

// ChallengeData 儲存挑戰訊息資料
type ChallengeData struct {
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

// ChallengeRequest 請求挑戰訊息的請求格式
type ChallengeRequest struct {
	WalletAddress string `json:"wallet_address" binding:"required"`
}

// ChallengeResponse 挑戰訊息的回應格式
type ChallengeResponse struct {
	Message   string `json:"message"`
	Challenge string `json:"challenge"`
}

// VerifyRequest 驗證簽名的請求格式
type VerifyRequest struct {
	WalletAddress string `json:"wallet_address" binding:"required"`
	Signature     string `json:"signature" binding:"required"`
	Message       string `json:"message" binding:"required"`
}

// VerifyResponse 驗證成功的回應格式
type VerifyResponse struct {
	Token         string       `json:"token"`
	WalletAddress string       `json:"wallet_address"`
	User          *models.User `json:"user,omitempty"`
}

// NewAuthHandler 建立新的 AuthHandler
func NewAuthHandler(userService *services.UserService) *AuthHandler {
	handler := &AuthHandler{
		userService: userService,
		challenges:  make(map[string]ChallengeData),
	}

	// 啟動清理過期挑戰的協程
	go handler.cleanupExpiredChallenges()

	return handler
}

// GenerateChallenge 產生挑戰訊息供前端簽署
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

	// 產生隨機挑戰碼
	challengeBytes := make([]byte, 32)
	if _, err := rand.Read(challengeBytes); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Code:    http.StatusInternalServerError,
			Message: "Failed to generate challenge",
		})
		return
	}
	challenge := base64.StdEncoding.EncodeToString(challengeBytes)

	// 產生要簽署的訊息
	timestamp := time.Now().Unix()
	message := fmt.Sprintf("Login to BlueLink\nWallet: %s\nChallenge: %s\nTimestamp: %d",
		req.WalletAddress, challenge, timestamp)

	// 儲存挑戰（5 分鐘有效）
	h.mu.Lock()
	h.challenges[req.WalletAddress] = ChallengeData{
		Message:   message,
		Timestamp: time.Now(),
	}
	h.mu.Unlock()

	c.JSON(http.StatusOK, models.SuccessResponse(
		http.StatusOK,
		"Challenge generated successfully",
		ChallengeResponse{
			Message:   message,
			Challenge: challenge,
		},
	))
}

// VerifySignature 驗證錢包簽名
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

	// 檢查挑戰是否存在
	h.mu.RLock()
	challengeData, exists := h.challenges[req.WalletAddress]
	h.mu.RUnlock()

	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Code:    http.StatusUnauthorized,
			Message: "Challenge not found or expired",
		})
		return
	}

	// 檢查挑戰是否過期（5 分鐘）
	if time.Since(challengeData.Timestamp) > 5*time.Minute {
		h.mu.Lock()
		delete(h.challenges, req.WalletAddress)
		h.mu.Unlock()

		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Code:    http.StatusUnauthorized,
			Message: "Challenge expired",
		})
		return
	}

	// 驗證訊息是否匹配
	if challengeData.Message != req.Message {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Code:    http.StatusUnauthorized,
			Message: "Message mismatch",
		})
		return
	}

	// TODO: 這裡應該驗證簽名
	// 實際應用中，應該使用 Sui SDK 驗證簽名
	// isValid := verifySuiSignature(req.WalletAddress, req.Signature, req.Message)
	// if !isValid {
	//     c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid signature"})
	//     return
	// }

	// 驗證成功，刪除挑戰
	h.mu.Lock()
	delete(h.challenges, req.WalletAddress)
	h.mu.Unlock()

	// 檢查使用者是否存在，不存在則建立
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

	// 產生認證 token（這裡簡化處理，實際應使用 JWT）
	token := h.generateAuthToken(req.WalletAddress)

	c.JSON(http.StatusOK, models.SuccessResponse(
		http.StatusOK,
		"Authentication successful",
		VerifyResponse{
			Token:         token,
			WalletAddress: req.WalletAddress,
			User:          user,
		},
	))
}

// generateAuthToken 產生認證 token
// 實際應用應使用 JWT 或其他安全的 token 方案
func (h *AuthHandler) generateAuthToken(walletAddress string) string {
	timestamp := time.Now().Unix()
	return fmt.Sprintf("%d:%s", timestamp, walletAddress)
}

// cleanupExpiredChallenges 定期清理過期的挑戰
func (h *AuthHandler) cleanupExpiredChallenges() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		h.mu.Lock()
		for wallet, data := range h.challenges {
			if time.Since(data.Timestamp) > 5*time.Minute {
				delete(h.challenges, wallet)
			}
		}
		h.mu.Unlock()
	}
}
