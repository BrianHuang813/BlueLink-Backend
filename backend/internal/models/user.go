package models

import "time"

// User 使用者資料模型
type User struct {
	ID            int64     `json:"id" db:"id"`
	WalletAddress string    `json:"wallet_address" db:"wallet_address"`
	Name          string    `json:"name,omitempty" db:"name"`
	Email         string    `json:"email,omitempty" db:"email"`
	Role          string    `json:"role" db:"role"` // "investor", "issuer", "admin"
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

// UserBasicInfo 使用者基本資訊（用於 middleware 和 context）
type UserBasicInfo struct {
	ID            int64  `json:"id"`
	WalletAddress string `json:"wallet_address"`
	Role          string `json:"role"`
}
