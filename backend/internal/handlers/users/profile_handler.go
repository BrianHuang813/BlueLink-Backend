package users

import (
	"bluelink-backend/internal/models"
	"bluelink-backend/internal/services"
	"bluelink-backend/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ProfileHandler 處理使用者資料相關的請求
type ProfileHandler struct {
	userService *services.UserService
}

// NewProfileHandler 建立新的 ProfileHandler
func NewProfileHandler(userService *services.UserService) *ProfileHandler {
	return &ProfileHandler{
		userService: userService,
	}
}

// UpdateProfileRequest 更新個人資料的請求格式
type UpdateProfileRequest struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required,email"`
}

// GetProfile 取得當前使用者的基本資料
// GET /api/v1/profile
func (h *ProfileHandler) GetProfile(c *gin.Context) {
	// 從 context 取得使用者資訊
	walletAddress, err := utils.GetWalletAddress(c)
	if err != nil {
		models.RespondUnauthorized(c, err.Error())
		return
	}

	userID, err := utils.GetUserID(c)
	if err != nil {
		models.RespondUnauthorized(c, err.Error())
		return
	}

	role, err := utils.GetUserRole(c)
	if err != nil {
		models.RespondUnauthorized(c, err.Error())
		return
	}

	models.RespondWithSuccess(c, http.StatusOK, "Profile retrieved successfully", gin.H{
		"user_id": userID,
		"wallet":  walletAddress,
		"role":    role,
	})
}

// GetFullProfile 取得當前使用者的完整資料
// GET /api/v1/profile/full
func (h *ProfileHandler) GetFullProfile(c *gin.Context) {
	userID, err := utils.GetUserID(c)
	if err != nil {
		models.RespondUnauthorized(c, err.Error())
		return
	}

	// 從資料庫載入完整使用者資料
	user, err := h.userService.GetByID(c.Request.Context(), userID)
	if err != nil {
		models.RespondInternalError(c, "Failed to get user profile", err)
		return
	}

	models.RespondWithSuccess(c, http.StatusOK, "Full profile retrieved successfully", user)
}

// UpdateProfile 更新使用者資料
// PUT /api/v1/profile
func (h *ProfileHandler) UpdateProfile(c *gin.Context) {
	userID, err := utils.GetUserID(c)
	if err != nil {
		models.RespondUnauthorized(c, err.Error())
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		models.RespondBadRequest(c, "Invalid request format", err)
		return
	}

	// 先取得現有使用者資料
	user, err := h.userService.GetByID(c.Request.Context(), userID)
	if err != nil {
		models.RespondNotFound(c, "User not found")
		return
	}

	// 更新使用者名稱
	user.Name = &req.Name

	// 儲存更新
	if err := h.userService.Update(c.Request.Context(), user); err != nil {
		models.RespondInternalError(c, "Failed to update profile", err)
		return
	}

	models.RespondWithSuccess(c, http.StatusOK, "Profile updated successfully", user)
}
