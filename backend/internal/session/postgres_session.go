package session

import (
	"bluelink-backend/internal/models"
	"bluelink-backend/internal/repository"
	"context"
	"fmt"
	"log"
	"time"
)

// PostgresSessionManager 使用 PostgreSQL 管理 Session
type PostgresSessionManager struct {
	repo          *repository.SessionRepository
	timeout       time.Duration
	idleTimeout   time.Duration
	maxConcurrent int
}

// NewPostgresSessionManager 建立新的 PostgreSQL Session Manager
func NewPostgresSessionManager(repo *repository.SessionRepository) *PostgresSessionManager {
	manager := &PostgresSessionManager{
		repo:          repo,
		timeout:       24 * time.Hour,   // Session 總過期時間
		idleTimeout:   30 * time.Minute, // 閒置過期時間
		maxConcurrent: 3,                // 最大並發 Session 數
	}

	// 啟動清理協程
	go manager.cleanup()

	return manager
}

// Create 建立新的 Session
func (m *PostgresSessionManager) Create(userID int64, walletAddress, role string, ipAddress, userAgent string) (*Session, error) {
	ctx := context.Background()

	// 檢查並發數量
	count, err := m.repo.CountByWalletAddress(ctx, walletAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to count sessions: %w", err)
	}

	// 如果達到最大並發數，刪除最舊的 Session
	if count >= m.maxConcurrent {
		if err := m.repo.DeleteOldestSession(ctx, walletAddress); err != nil {
			return nil, fmt.Errorf("failed to delete oldest session: %w", err)
		}
	}

	// 生成 Session ID
	sessionID, err := generateSecureID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate session ID: %w", err)
	}

	now := time.Now()
	expiresAt := now.Add(m.timeout)

	dbSession := &models.DBSession{
		ID:            sessionID,
		UserID:        userID,
		WalletAddress: walletAddress,
		Role:          role,
		IPAddress:     ipAddress,
		UserAgent:     userAgent,
		CreatedAt:     now,
		LastActiveAt:  now,
		ExpiresAt:     expiresAt,
	}

	// 儲存到資料庫
	if err := m.repo.Create(ctx, dbSession); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return &Session{
		ID:            dbSession.ID,
		UserID:        dbSession.UserID,
		WalletAddress: dbSession.WalletAddress,
		Role:          dbSession.Role,
		CreatedAt:     dbSession.CreatedAt,
		LastActiveAt:  dbSession.LastActiveAt,
		IPAddress:     dbSession.IPAddress,
		UserAgent:     dbSession.UserAgent,
	}, nil
}

// Get 取得 Session
func (m *PostgresSessionManager) Get(sessionID string) (*Session, error) {
	ctx := context.Background()

	dbSession, err := m.repo.GetByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	// 檢查閒置超時
	if time.Since(dbSession.LastActiveAt) > m.idleTimeout {
		m.Delete(sessionID)
		return nil, fmt.Errorf("session expired due to inactivity")
	}

	return &Session{
		ID:            dbSession.ID,
		UserID:        dbSession.UserID,
		WalletAddress: dbSession.WalletAddress,
		Role:          dbSession.Role,
		CreatedAt:     dbSession.CreatedAt,
		LastActiveAt:  dbSession.LastActiveAt,
		IPAddress:     dbSession.IPAddress,
		UserAgent:     dbSession.UserAgent,
	}, nil
}

// Touch 更新 Session 的最後活躍時間
func (m *PostgresSessionManager) Touch(sessionID string) error {
	ctx := context.Background()
	return m.repo.Update(ctx, sessionID)
}

// Delete 刪除特定的 Session
func (m *PostgresSessionManager) Delete(sessionID string) error {
	ctx := context.Background()
	return m.repo.Delete(ctx, sessionID)
}

// DeleteAllUserSessions 刪除特定錢包地址的所有 Session
func (m *PostgresSessionManager) DeleteAllUserSessions(walletAddress string) error {
	ctx := context.Background()
	return m.repo.DeleteByWalletAddress(ctx, walletAddress)
}

// GetUserSessions 取得特定錢包地址的所有 Session
func (m *PostgresSessionManager) GetUserSessions(walletAddress string) ([]*Session, error) {
	ctx := context.Background()

	dbSessions, err := m.repo.GetByWalletAddress(ctx, walletAddress)
	if err != nil {
		return nil, err
	}

	sessions := make([]*Session, 0, len(dbSessions))
	for _, dbSession := range dbSessions {
		// 檢查閒置超時
		if time.Since(dbSession.LastActiveAt) > m.idleTimeout {
			continue
		}

		sessions = append(sessions, &Session{
			ID:            dbSession.ID,
			UserID:        dbSession.UserID,
			WalletAddress: dbSession.WalletAddress,
			Role:          dbSession.Role,
			CreatedAt:     dbSession.CreatedAt,
			LastActiveAt:  dbSession.LastActiveAt,
			IPAddress:     dbSession.IPAddress,
			UserAgent:     dbSession.UserAgent,
		})
	}

	return sessions, nil
}

// cleanup 定期清理過期的 Session
func (m *PostgresSessionManager) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		ctx := context.Background()
		deleted, err := m.repo.DeleteExpired(ctx)
		if err != nil {
			log.Printf("Failed to cleanup expired sessions: %v", err)
			continue
		}

		if deleted > 0 {
			log.Printf("Cleaned up %d expired sessions", deleted)
		}
	}
}
