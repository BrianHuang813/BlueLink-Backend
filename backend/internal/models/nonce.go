package models

import "time"

// Nonce 用於認證挑戰的一次性隨機數
type Nonce struct {
	ID            int64     `json:"id" db:"id"`
	WalletAddress string    `json:"wallet_address" db:"wallet_address"` // 用作存儲的 key
	Nonce         string    `json:"nonce" db:"nonce"`                   // Base64 編碼的隨機數
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	ExpiresAt     time.Time `json:"expires_at" db:"expires_at"`
}
