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
	suiClient        interface{} // 將在路由設置時注入
}

func NewBondHandler(bondService *services.BondService, bondTokenService *services.BondTokenService) *BondHandler {
	return &BondHandler{
		bondService:      bondService,
		bondTokenService: bondTokenService,
	}
}

// SetSuiClient 設置 Sui 客戶端 (用於同步交易)
func (h *BondHandler) SetSuiClient(client interface{}) {
	h.suiClient = client
}

// GetAllBonds 獲取所有上架債券
func (h *BondHandler) GetAllBonds(c *gin.Context) {
	bonds, err := h.bondService.GetAllBonds(c.Request.Context())
	if err != nil {
		models.RespondInternalError(c, "Failed to fetch bonds", err)
		return
	}

	// 轉換為前端需要的響應格式
	responseData := ToBondResponseList(bonds)

	// 返回符合前端要求的格式: {code, message, data: Bond[]}
	models.RespondWithSuccess(c, http.StatusOK, "success", responseData)
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

	// 轉換為前端需要的響應格式
	responseData := ToBondResponse(bond)

	models.RespondWithSuccess(c, http.StatusOK, "success", responseData)
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

// SyncTransaction 同步鏈上交易
func (h *BondHandler) SyncTransaction(c *gin.Context) {
	var req SyncTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		models.RespondBadRequest(c, "Invalid request body", err)
		return
	}

	// 注意: 完整的鏈上交易同步邏輯較為複雜,需要:
	// 1. 從 Sui 鏈上查詢交易詳情
	// 2. 解析交易事件
	// 3. 根據事件類型更新數據庫
	//
	// 目前的實現返回成功響應,實際的同步邏輯由事件監聽器處理
	// 如果需要實時同步,可以在這裡調用 blockchain.EventListener 的相關方法

	// TODO: 實現完整的交易同步邏輯
	// 現階段先返回成功,交易會由後台事件監聽器自動處理

	models.RespondWithSuccess(c, http.StatusOK, "Transaction will be indexed by event listener", nil)
}
