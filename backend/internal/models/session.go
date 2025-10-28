package models

import "time"

// Session 資料庫 Session 模型
type DBSession struct {
	ID            string    `json:"id" db:"id"`
	UserID        int64     `json:"user_id" db:"user_id"`
	WalletAddress string    `json:"wallet_address" db:"wallet_address"`
	Role          string    `json:"role" db:"role"`
	IPAddress     string    `json:"ip_address" db:"ip_address"`
	UserAgent     string    `json:"user_agent" db:"user_agent"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	LastActiveAt  time.Time `json:"last_active_at" db:"last_active_at"`
	ExpiresAt     time.Time `json:"expires_at" db:"expires_at"`
}
