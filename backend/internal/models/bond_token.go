package models

import "time"

// BondToken 債券代幣 NFT 資料模型（對應合約 BondToken）
type BondToken struct {
	// 資料庫欄位
	ID        int64  `json:"id" db:"id"`
	OnChainID string `json:"on_chain_id" db:"on_chain_id"` // 對應合約的 UID

	// 專案關聯
	ProjectID string `json:"project_id" db:"project_id"` // 對應合約的 project_id

	// 🆕 代幣自包含資訊（從 BondProject 複製而來）
	BondName           string `json:"bond_name" db:"bond_name"`                       // 債券名稱
	TokenImageUrl      string `json:"token_image_url" db:"token_image_url"`           // NFT 圖片 URL
	MaturityDate       int64  `json:"maturity_date" db:"maturity_date"`               // 到期日 (timestamp ms)
	AnnualInterestRate int64  `json:"annual_interest_rate" db:"annual_interest_rate"` // 年利率 (basis points)

	// 代幣資訊
	TokenNumber  int64  `json:"token_number" db:"token_number"`   // 對應 token_number
	Owner        string `json:"owner" db:"owner"`                 // 對應 owner
	Amount       int64  `json:"amount" db:"amount"`               // 投資金額（單位：MIST）
	PurchaseDate int64  `json:"purchase_date" db:"purchase_date"` // 購買日期 (timestamp ms)
	IsRedeemed   bool   `json:"is_redeemed" db:"is_redeemed"`     // 對應 is_redeemed

	// 資料庫管理欄位
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"` // 軟刪除
}
