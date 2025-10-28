package models

import "time"

// Transaction 交易記錄
type Transaction struct {
	ID            int64     `json:"id" db:"id"`
	TxHash        string    `json:"tx_hash" db:"tx_hash"`
	EventType     string    `json:"event_type" db:"event_type"`
	BondID        *int64    `json:"bond_id,omitempty" db:"bond_id"`
	UserID        *int64    `json:"user_id,omitempty" db:"user_id"`
	WalletAddress string    `json:"wallet_address" db:"wallet_address"`
	Amount        *float64  `json:"amount,omitempty" db:"amount"`
	Quantity      *int64    `json:"quantity,omitempty" db:"quantity"`
	Price         *float64  `json:"price,omitempty" db:"price"`
	Status        string    `json:"status" db:"status"`
	BlockNumber   *int64    `json:"block_number,omitempty" db:"block_number"`
	Timestamp     time.Time `json:"timestamp" db:"timestamp"`
	Metadata      *string   `json:"metadata,omitempty" db:"metadata"` // JSONB
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

// EventType 常量
const (
	EventBondCreated     = "bond_created"
	EventBondPurchased   = "bond_purchased"
	EventBondRedeemed    = "bond_redeemed"
	EventInterestPaid    = "interest_paid"
	EventBondTransferred = "bond_transferred"
)

// TransactionStatus 常量
const (
	TxStatusPending   = "pending"
	TxStatusConfirmed = "confirmed"
	TxStatusFailed    = "failed"
)

// UserBond 使用者持倉記錄
type UserBond struct {
	ID                   int64     `json:"id" db:"id"`
	UserID               int64     `json:"user_id" db:"user_id"`
	BondID               int64     `json:"bond_id" db:"bond_id"`
	WalletAddress        string    `json:"wallet_address" db:"wallet_address"`
	Quantity             int64     `json:"quantity" db:"quantity"`
	AveragePurchasePrice *float64  `json:"average_purchase_price,omitempty" db:"average_purchase_price"`
	TotalInterestEarned  float64   `json:"total_interest_earned" db:"total_interest_earned"`
	CreatedAt            time.Time `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time `json:"updated_at" db:"updated_at"`
}

// UserBondWithDetails 持倉詳情（含債券資訊）
type UserBondWithDetails struct {
	UserBond
	BondName           string  `json:"bond_name" db:"bond_name"`
	IssuerName         string  `json:"issuer_name" db:"issuer_name"`
	AnnualInterestRate float64 `json:"annual_interest_rate" db:"annual_interest_rate"`
	MaturityDate       string  `json:"maturity_date" db:"maturity_date"`
	Active             bool    `json:"active" db:"active"`
	Redeemable         bool    `json:"redeemable" db:"redeemable"`
}
