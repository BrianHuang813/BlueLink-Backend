package bonds

import (
	"bluelink-backend/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type BondHandler struct {
	bondService *services.BondService
}

func NewBondHandler(bondService *services.BondService) *BondHandler {
	return &BondHandler{
		bondService: bondService,
	}
}

// // Create 創建新債券
// func (h *BondHandler) Create(c *gin.Context) {
// 	var req CreateBondRequest
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{
// 			"error": "Invalid request body",
// 		})
// 		return
// 	}

// 	// 從 context 獲取用戶地址
// 	issuerAddress, addressExists := c.Get("wallet_address")
// 	issuerRole, roleExists := c.Get("role")
// 	if !addressExists || !roleExists {
// 		c.JSON(http.StatusUnauthorized, gin.H{
// 			"error": "User not authorized to create bonds",
// 		})
// 		return
// 	}
// 	if issuerRole != "issuer" {
// 		c.JSON(http.StatusUnauthorized, gin.H{
// 			"error": "User not authorized to create bonds",
// 		})
// 		return
// 	}

// 	// 創建債券（on-chain)
// 	bond, txDigest, err := h.bondService.CreateBond(
// 		c.Request.Context(),
// 		issuerAddress.(string),
// 	)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{
// 			"error": err.Error(),
// 		})
// 		return
// 	}

// 	c.JSON(http.StatusCreated, gin.H{
// 		"bond":      bond,
// 		"tx_digest": txDigest,
// 	})
// }

// // BuyBond 購買債券
// func (h *BondHandler) BuyBond(c *gin.Context) {
// 	var req BuyBondRequest
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{
// 			"error": "Invalid request body",
// 		})
// 		return
// 	}

// 	// 從 context 獲取用戶地址
// 	userAddress, exists := c.Get("wallet_address")
// 	if !exists {
// 		c.JSON(http.StatusUnauthorized, gin.H{
// 			"error": "User not authenticated",
// 		})
// 		return
// 	}

// 	// 購買債券（on-chain）
// 	purchase, txDigest, err := h.bondService.BuyBond(
// 		c.Request.Context(),
// 		userAddress.(string),
// 		req.ID,
// 		req.Amount,
// 	)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{
// 			"error": err.Error(),
// 		})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{
// 		"purchase":  purchase,
// 		"tx_digest": txDigest,
// 	})
// }

// GetList 獲取所有債券列表
func (h *BondHandler) GetAllList(c *gin.Context) {
	bonds, err := h.bondService.GetAllBonds(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch bonds",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"bonds": bonds,
	})
}
