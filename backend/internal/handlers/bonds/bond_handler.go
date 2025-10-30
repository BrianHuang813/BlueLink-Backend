package bonds

import (
	"bluelink-backend/internal/models"
	"bluelink-backend/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type BondHandler struct {
	bondService      *services.BondService
	bondTokenService *services.BondTokenService
}

func NewBondHandler(bondService *services.BondService, bondTokenService *services.BondTokenService) *BondHandler {
	return &BondHandler{
		bondService:      bondService,
		bondTokenService: bondTokenService,
	}
}

// GetAllBonds 獲取所有上架債券
func (h *BondHandler) GetAllBonds(c *gin.Context) {
	bonds, err := h.bondService.GetAllBonds(c.Request.Context())
	if err != nil {
		models.RespondInternalError(c, "Failed to fetch bonds", err)
		return
	}

	models.RespondWithSuccess(c, http.StatusOK, "Bonds retrieved successfully", gin.H{
		"bonds": bonds,
	})
}

// GetBondByID 根據 ID 獲取債券詳情
func (h *BondHandler) GetBondByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		models.RespondBadRequest(c, "Invalid bond ID", err)
		return
	}

	bond, err := h.bondService.GetBondByID(c.Request.Context(), id)
	if err != nil {
		models.RespondInternalError(c, "Failed to fetch bond", err)
		return
	}

	if bond == nil {
		models.RespondNotFound(c, "Bond not found")
		return
	}

	models.RespondWithSuccess(c, http.StatusOK, "Bond retrieved successfully", gin.H{
		"bond": bond,
	})
}

// 🆕 GetBondTokenByID 根據 ID 獲取債券代幣詳情
func (h *BondHandler) GetBondTokenByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		models.RespondBadRequest(c, "Invalid bond token ID", err)
		return
	}

	token, err := h.bondTokenService.GetBondTokenByID(c.Request.Context(), id)
	if err != nil {
		models.RespondInternalError(c, "Failed to fetch bond token", err)
		return
	}

	if token == nil {
		models.RespondNotFound(c, "Bond token not found")
		return
	}

	models.RespondWithSuccess(c, http.StatusOK, "Bond token retrieved successfully", gin.H{
		"bond_token": token,
	})
}

// 🆕 GetBondTokenByOnChainID 根據鏈上 ID 獲取債券代幣詳情
func (h *BondHandler) GetBondTokenByOnChainID(c *gin.Context) {
	onChainID := c.Param("on_chain_id")
	if onChainID == "" {
		models.RespondBadRequest(c, "On-chain ID is required", nil)
		return
	}

	token, err := h.bondTokenService.GetBondTokenByOnChainID(c.Request.Context(), onChainID)
	if err != nil {
		models.RespondInternalError(c, "Failed to fetch bond token", err)
		return
	}

	if token == nil {
		models.RespondNotFound(c, "Bond token not found")
		return
	}

	models.RespondWithSuccess(c, http.StatusOK, "Bond token retrieved successfully", gin.H{
		"bond_token": token,
	})
}

// 🆕 GetBondTokensByOwner 根據擁有者地址獲取債券代幣列表
func (h *BondHandler) GetBondTokensByOwner(c *gin.Context) {
	var req GetBondTokensByOwnerRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		models.RespondBadRequest(c, "Invalid request parameters", err)
		return
	}

	if req.Limit <= 0 {
		req.Limit = 100
	}

	tokens, err := h.bondTokenService.GetBondTokensByOwner(c.Request.Context(), req.Owner, req.Limit, req.Offset)
	if err != nil {
		models.RespondInternalError(c, "Failed to fetch bond tokens", err)
		return
	}

	models.RespondWithSuccess(c, http.StatusOK, "Bond tokens retrieved successfully", gin.H{
		"bond_tokens": tokens,
		"count":       len(tokens),
	})
}

// 🆕 GetBondTokensByProject 根據專案 ID 獲取債券代幣列表
func (h *BondHandler) GetBondTokensByProject(c *gin.Context) {
	var req GetBondTokensByProjectRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		models.RespondBadRequest(c, "Invalid request parameters", err)
		return
	}

	if req.Limit <= 0 {
		req.Limit = 100
	}

	tokens, err := h.bondTokenService.GetBondTokensByProjectID(c.Request.Context(), req.ProjectID, req.Limit, req.Offset)
	if err != nil {
		models.RespondInternalError(c, "Failed to fetch bond tokens", err)
		return
	}

	models.RespondWithSuccess(c, http.StatusOK, "Bond tokens retrieved successfully", gin.H{
		"bond_tokens": tokens,
		"count":       len(tokens),
	})
}
