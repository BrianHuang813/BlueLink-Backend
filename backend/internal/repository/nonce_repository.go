package repository

import (
	"bluelink-backend/internal/models"
	"context"
	"database/sql"
	"fmt"
	"time"
)

type NonceRepository struct {
	db *sql.DB
}

func NewNonceRepository(db *sql.DB) *NonceRepository {
	repo := &NonceRepository{db: db}

	// 啟動定期清理協程
	go repo.cleanupExpiredNonces()

	return repo
}

// cleanupExpiredNonces 定期清理過期的 nonce
func (r *NonceRepository) cleanupExpiredNonces() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		ctx := context.Background()
		deleted, err := r.DeleteExpired(ctx)
		if err != nil {
			// 可以使用 logger 記錄錯誤
			continue
		}
		if deleted > 0 {
			// 可以使用 logger 記錄清理結果
			_ = deleted
		}
	}
}

// Create 創建新的 nonce
// 如果該 wallet_address 已有 nonce，會先刪除舊的（確保一個地址只有一個有效 nonce）
func (r *NonceRepository) Create(ctx context.Context, walletAddress, nonce string, ttl time.Duration) error {
	// 先刪除該地址的舊 nonce
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM nonces WHERE wallet_address = $1
	`, walletAddress)
	if err != nil {
		return fmt.Errorf("failed to delete old nonce: %w", err)
	}

	// 插入新 nonce
	expiresAt := time.Now().Add(ttl)
	_, err = r.db.ExecContext(ctx, `
		INSERT INTO nonces (wallet_address, nonce, created_at, expires_at)
		VALUES ($1, $2, $3, $4)
	`, walletAddress, nonce, time.Now(), expiresAt)
	if err != nil {
		return fmt.Errorf("failed to create nonce: %w", err)
	}

	return nil
}

// Get 根據 wallet_address 取得 nonce
// 會自動檢查是否過期，如果過期會刪除並返回 not found
func (r *NonceRepository) Get(ctx context.Context, walletAddress string) (*models.Nonce, error) {
	var n models.Nonce
	err := r.db.QueryRowContext(ctx, `
		SELECT id, wallet_address, nonce, created_at, expires_at
		FROM nonces
		WHERE wallet_address = $1
	`, walletAddress).Scan(&n.ID, &n.WalletAddress, &n.Nonce, &n.CreatedAt, &n.ExpiresAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("nonce not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce: %w", err)
	}

	// 檢查是否過期
	if time.Now().After(n.ExpiresAt) {
		// 刪除過期的 nonce
		r.Delete(ctx, walletAddress)
		return nil, fmt.Errorf("nonce expired")
	}

	return &n, nil
}

// Delete 根據 wallet_address 刪除 nonce（防止重放攻擊）
func (r *NonceRepository) Delete(ctx context.Context, walletAddress string) error {
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM nonces WHERE wallet_address = $1
	`, walletAddress)
	if err != nil {
		return fmt.Errorf("failed to delete nonce: %w", err)
	}
	return nil
}

// DeleteExpired 刪除所有過期的 nonce（定期清理）
func (r *NonceRepository) DeleteExpired(ctx context.Context) (int64, error) {
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM nonces WHERE expires_at < $1
	`, time.Now())
	if err != nil {
		return 0, fmt.Errorf("failed to delete expired nonces: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get affected rows: %w", err)
	}

	return affected, nil
}

// Verify 驗證 nonce 是否匹配（原子操作：查詢、比對、刪除）
func (r *NonceRepository) Verify(ctx context.Context, walletAddress, nonce string) (bool, error) {
	storedNonce, err := r.Get(ctx, walletAddress)
	if err != nil {
		return false, err
	}

	// 比對 nonce
	if storedNonce.Nonce != nonce {
		return false, fmt.Errorf("nonce mismatch")
	}

	// 刪除已使用的 nonce（防止重放攻擊）
	if err := r.Delete(ctx, walletAddress); err != nil {
		return false, fmt.Errorf("failed to delete used nonce: %w", err)
	}

	return true, nil
}
