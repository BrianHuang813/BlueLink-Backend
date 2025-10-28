package models

import "time"

// Bond 債券資料模型（對應合約 BondProject）
type Bond struct {
	// 資料庫欄位
	ID        int64  `json:"id" db:"id"`
	OnChainID string `json:"on_chain_id" db:"on_chain_id"` // 對應合約的 UID

	// 發行者資訊
	IssuerAddress string `json:"issuer_address" db:"issuer_address"` // 對應合約的 issuer
	IssuerName    string `json:"issuer_name" db:"issuer_name"`       // 對應合約的 issuer_name
	BondName      string `json:"bond_name" db:"bond_name"`           // 對應合約的 bond_name

	// 金額相關（使用 int64 對應 u64，單位：MIST，1 SUI = 1,000,000,000 MIST）
	TotalAmount    int64 `json:"total_amount" db:"total_amount"`       // 對應 total_amount (募集總額度)
	AmountRaised   int64 `json:"amount_raised" db:"amount_raised"`     // 對應 amount_raised (已募集金額)
	AmountRedeemed int64 `json:"amount_redeemed" db:"amount_redeemed"` // 對應 amount_redeemed (已贖回金額)

	// 代幣相關
	TokensIssued   int64 `json:"tokens_issued" db:"tokens_issued"`     // 對應 tokens_issued (發行的債券代幣數量)
	TokensRedeemed int64 `json:"tokens_redeemed" db:"tokens_redeemed"` // 對應 tokens_redeemed (已贖回代幣數量)

	// 利率和日期
	AnnualInterestRate int64  `json:"annual_interest_rate" db:"annual_interest_rate"` // 對應 annual_interest_rate (basis points / 10000)
	MaturityDate       string `json:"maturity_date" db:"maturity_date"`               // 對應 maturity_date (儲存為字串，格式: YYYY-MM-DD)
	IssueDate          string `json:"issue_date" db:"issue_date"`                     // 對應 issue_date (儲存為字串，格式: YYYY-MM-DD)

	// 狀態
	Active     bool `json:"active" db:"active"`         // 對應 active (債券是否活躍)
	Redeemable bool `json:"redeemable" db:"redeemable"` // 對應 redeemable (是否可贖回)

	// 資金池餘額快照（使用 int64 對應 Balance<SUI>，單位：MIST）
	RaisedFundsBalance    int64 `json:"raised_funds_balance" db:"raised_funds_balance"`       // 對應 raised_funds 的餘額快照
	RedemptionPoolBalance int64 `json:"redemption_pool_balance" db:"redemption_pool_balance"` // 對應 redemption_pool 的餘額快照

	// 資料庫管理欄位
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"` // 軟刪除
}
