package bonds

import (
	"bluelink-backend/internal/models"
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
