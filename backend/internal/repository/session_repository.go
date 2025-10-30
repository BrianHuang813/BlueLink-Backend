package repository

import (
	"bluelink-backend/internal/models"
	"context"
	"database/sql"
	"fmt"
	"time"
)

// SessionRepository 處理 Session 的資料庫操作
type SessionRepository struct {
	db *sql.DB
}

// NewSessionRepository 建立新的 SessionRepository
func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

// Create 建立新的 Session
func (r *SessionRepository) Create(ctx context.Context, session *models.DBSession) error {
	query := `
		INSERT INTO sessions (
			id, user_id, wallet_address, role, 
			ip_address, user_agent, created_at, 
			last_active_at, expires_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.ExecContext(
		ctx, query,
		session.ID,
		session.UserID,
		session.WalletAddress,
		session.Role,
		session.IPAddress,
		session.UserAgent,
		session.CreatedAt,
		session.LastActiveAt,
		session.ExpiresAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	return nil
}

// GetByID 根據 ID 取得 Session
func (r *SessionRepository) GetByID(ctx context.Context, sessionID string) (*models.DBSession, error) {
	query := `
		SELECT 
			id, user_id, wallet_address, role,
			ip_address, user_agent, created_at,
			last_active_at, expires_at
		FROM sessions
		WHERE id = $1 AND expires_at > NOW()
	`

	var session models.DBSession
	err := r.db.QueryRowContext(ctx, query, sessionID).Scan(
		&session.ID,
		&session.UserID,
		&session.WalletAddress,
		&session.Role,
		&session.IPAddress,
		&session.UserAgent,
		&session.CreatedAt,
		&session.LastActiveAt,
		&session.ExpiresAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("session not found or expired")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return &session, nil
}

// Update 更新 Session 的最後活躍時間
func (r *SessionRepository) Update(ctx context.Context, sessionID string) error {
	query := `
		UPDATE sessions
		SET last_active_at = $1
		WHERE id = $2
	`

	result, err := r.db.ExecContext(ctx, query, time.Now(), sessionID)
	if err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("session not found")
	}

	return nil
}

// Delete 刪除特定的 Session
func (r *SessionRepository) Delete(ctx context.Context, sessionID string) error {
	query := `DELETE FROM sessions WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, sessionID)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}

// DeleteByWalletAddress 刪除特定錢包地址的所有 Session
func (r *SessionRepository) DeleteByWalletAddress(ctx context.Context, walletAddress string) error {
	query := `DELETE FROM sessions WHERE wallet_address = $1`

	_, err := r.db.ExecContext(ctx, query, walletAddress)
	if err != nil {
		return fmt.Errorf("failed to delete user sessions: %w", err)
	}

	return nil
}

// GetByWalletAddress 取得特定錢包地址的所有 Session
func (r *SessionRepository) GetByWalletAddress(ctx context.Context, walletAddress string) ([]*models.DBSession, error) {
	query := `
		SELECT 
			id, user_id, wallet_address, role,
			ip_address, user_agent, created_at,
			last_active_at, expires_at
		FROM sessions
		WHERE wallet_address = $1 AND expires_at > NOW()
		ORDER BY last_active_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, walletAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get user sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*models.DBSession
	for rows.Next() {
		var session models.DBSession
		err := rows.Scan(
			&session.ID,
			&session.UserID,
			&session.WalletAddress,
			&session.Role,
			&session.IPAddress,
			&session.UserAgent,
			&session.CreatedAt,
			&session.LastActiveAt,
			&session.ExpiresAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}
		sessions = append(sessions, &session)
	}

	return sessions, nil
}

// CountByWalletAddress 計算特定錢包地址的 Session 數量
func (r *SessionRepository) CountByWalletAddress(ctx context.Context, walletAddress string) (int, error) {
	query := `
		SELECT COUNT(*) 
		FROM sessions 
		WHERE wallet_address = $1 AND expires_at > NOW()
	`

	var count int
	err := r.db.QueryRowContext(ctx, query, walletAddress).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count sessions: %w", err)
	}

	return count, nil
}

// DeleteExpired 清理過期的 Session
func (r *SessionRepository) DeleteExpired(ctx context.Context) (int64, error) {
	query := `DELETE FROM sessions WHERE expires_at <= NOW()`

	result, err := r.db.ExecContext(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("failed to delete expired sessions: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rows, nil
}

// DeleteOldestSession 刪除特定錢包地址最舊的 Session
func (r *SessionRepository) DeleteOldestSession(ctx context.Context, walletAddress string) error {
	query := `
		DELETE FROM sessions 
		WHERE id = (
			SELECT id 
			FROM sessions 
			WHERE wallet_address = $1 
			ORDER BY last_active_at ASC 
			LIMIT 1
		)
	`

	_, err := r.db.ExecContext(ctx, query, walletAddress)
	if err != nil {
		return fmt.Errorf("failed to delete oldest session: %w", err)
	}

	return nil
}
