package models

import "time"

// User 使用者資料模型 (TODO: blacklist, KYC, 2FA, lastLoginAt, LastLoginIP...)
type User struct {
	ID            int64      `json:"id" db:"id"`
	WalletAddress string     `json:"wallet_address" db:"wallet_address"`
	Name          string     `json:"name,omitempty" db:"name"`
	Role          string     `json:"role" db:"role"` // "buyer", "issuer", "admin"
	Timezone      string     `json:"timezone,omitempty" db:"timezone"`
	Language      string     `json:"language,omitempty" db:"language"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt     *time.Time `json:"deleted_at,omitempty" db:"deleted_at"` // 軟刪除時間戳記

	// // ===== KYC/AML =====
	// KYCStatus     string    `json:"kyc_status" db:"kyc_status"`         // "pending", "verified", "rejected"
	// KYCLevel      int       `json:"kyc_level" db:"kyc_level"`           // 0: 未認證, 1: 基本, 2: 進階
	// KYCVerifiedAt *time.Time `json:"kyc_verified_at,omitempty" db:"kyc_verified_at"`
	// Country       string    `json:"country,omitempty" db:"country"`     // 國家/地區（合規需求）

	// // ===== 交易限制 =====
	// DailyLimit    *float64  `json:"daily_limit,omitempty" db:"daily_limit"`     // 每日交易限額
	// MonthlyLimit  *float64  `json:"monthly_limit,omitempty" db:"monthly_limit"` // 每月交易限額
	// IsBlacklisted bool      `json:"is_blacklisted" db:"is_blacklisted"`         // 黑名單

	// // ===== 安全相關 =====
	// TwoFactorEnabled bool   `json:"two_factor_enabled" db:"two_factor_enabled"` // 是否啟用 2FA
	// LastLoginAt      *time.Time `json:"last_login_at,omitempty" db:"last_login_at"`
	// LastLoginIP      string `json:"last_login_ip,omitempty" db:"last_login_ip"`
}

// UserProfile 前端顯示用使用者資料
type UserProfile struct {
	ID            int64  `json:"id" db:"id"`
	WalletAddress string `json:"wallet_address" db:"wallet_address"`
	Name          string `json:"name,omitempty" db:"name"`
	Role          string `json:"role" db:"role"` // "buyer", "issuer", "admin"
	Timezoen      string `json:"timezone,omitempty" db:"timezone"`
	Language      string `json:"language,omitempty" db:"language"`
}

// UserBasicInfo 最小必要資訊
type UserBasicInfo struct {
	ID            int64  `json:"id"`
	WalletAddress string `json:"wallet_address"`
	Role          string `json:"role"` // "buyer", "issuer", "admin"
}
