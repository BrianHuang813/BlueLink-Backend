package bonds

type CreateBondRequest struct {
	ID            int64   `json:"id" binding:"required"`
	Name          string  `json:"name" binding:"required"`
	IssuerName    string  `json:"issuer_name" binding:"required"`
	IssuerAddress string  `json:"issuer_address" binding:"required"`
	BondImageUrl  string  `json:"bond_image_url" binding:"required"`  // 🆕 專案展示圖片 URL
	TokenImageUrl string  `json:"token_image_url" binding:"required"` // 🆕 NFT 代幣圖片 URL
	MetadataUrl   string  `json:"metadata_url" binding:"required"`    // 🆕 完整元數據 URL
	FaceValue     float64 `json:"face_value" binding:"required,gt=0"`
	Currency      string  `json:"currency" binding:"required"`
}

type BuyBondRequest struct {
	ID           int64   `json:"id" binding:"required"`
	OnChainID    string  `json:"on_chain_id" binding:"required"`
	Name         string  `json:"name" binding:"required"`
	Amount       float64 `json:"amount" binding:"required,gt=0"`
	BuyerAddress string  `json:"buyer_address" binding:"required"`
}

type GetBondByIDRequest struct {
	ID int64 `uri:"id" binding:"required"`
}

type GetAllBondsRequest struct {
}

// 🆕 BondToken 相關請求結構

type GetBondTokenByIDRequest struct {
	ID int64 `uri:"id" binding:"required"`
}

type GetBondTokenByOnChainIDRequest struct {
	OnChainID string `uri:"on_chain_id" binding:"required"`
}

type GetBondTokensByOwnerRequest struct {
	Owner  string `form:"owner" binding:"required"`
	Limit  int    `form:"limit"`
	Offset int    `form:"offset"`
}

type GetBondTokensByProjectRequest struct {
	ProjectID string `form:"project_id" binding:"required"`
	Limit     int    `form:"limit"`
	Offset    int    `form:"offset"`
}

type CalculateRedemptionRequest struct {
	TokenID string `json:"token_id" binding:"required"` // 只需 tokenID，不再需要 projectID
}

// SyncTransactionRequest 同步鏈上交易請求
type SyncTransactionRequest struct {
	TransactionDigest string `json:"transaction_digest" binding:"required"`
	EventType         string `json:"event_type" binding:"required,oneof=bond_created bond_purchased bond_redeemed funds_withdrawn redemption_deposited"`
}
