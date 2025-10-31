package session

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

type MemorySessionManager struct {
	sessions      map[string]*Session
	userSessions  map[string]map[string]bool // wallet -> sessionIDs
	mu            sync.RWMutex
	timeout       time.Duration
	idleTimeout   time.Duration
	maxConcurrent int
}

type Session struct {
	ID            string    `json:"id"`
	UserID        int64     `json:"user_id"`
	WalletAddress string    `json:"wallet_address"`
	Role          string    `json:"role"`
	CreatedAt     time.Time `json:"created_at"`
	LastActiveAt  time.Time `json:"last_active_at"`
	IPAddress     string    `json:"ip_address"`
	UserAgent     string    `json:"user_agent"`
}

func NewMemorySessionManager() *MemorySessionManager {
	manager := &MemorySessionManager{
		sessions:      make(map[string]*Session),
		userSessions:  make(map[string]map[string]bool),
		timeout:       24 * time.Hour,
		idleTimeout:   30 * time.Minute,
		maxConcurrent: 3,
	}

	// 啟動清理協程
	go manager.cleanup()

	return manager
}

// Create new session
func (m *MemorySessionManager) Create(userID int64, walletAddress, role string, ipAddress, userAgent string) (*Session, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 檢查並發數量
	if userSessions, exists := m.userSessions[walletAddress]; exists {
		if len(userSessions) >= m.maxConcurrent {
			return nil, fmt.Errorf("maximum concurrent sessions reached")
		}
	}

	// 生成 Session ID
	sessionID, err := generateSecureID()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	session := &Session{
		ID:            sessionID,
		UserID:        userID,
		WalletAddress: walletAddress,
		Role:          role,
		CreatedAt:     now,
		LastActiveAt:  now,
		IPAddress:     ipAddress,
		UserAgent:     userAgent,
	}

	// 儲存 session
	m.sessions[sessionID] = session

	// 記錄到用戶 session 列表
	if m.userSessions[walletAddress] == nil {
		m.userSessions[walletAddress] = make(map[string]bool)
	}
	m.userSessions[walletAddress][sessionID] = true

	return session, nil
}

// Get session
func (m *MemorySessionManager) Get(sessionID string) (*Session, error) {
	m.mu.RLock()
	session, exists := m.sessions[sessionID]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("session not found")
	}

	// 檢查是否過期
	if time.Since(session.CreatedAt) > m.timeout {
		m.Delete(sessionID)
		return nil, fmt.Errorf("session expired")
	}

	// 檢查閒置超時
	if time.Since(session.LastActiveAt) > m.idleTimeout {
		m.Delete(sessionID)
		return nil, fmt.Errorf("session expired due to inactivity")
	}

	return session, nil
}

// Touch 更新活躍時間
func (m *MemorySessionManager) Touch(sessionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, exists := m.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found")
	}

	session.LastActiveAt = time.Now()
	return nil
}

// Delete session
func (m *MemorySessionManager) Delete(sessionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, exists := m.sessions[sessionID]
	if !exists {
		return nil
	}

	delete(m.sessions, sessionID)
	delete(m.userSessions[session.WalletAddress], sessionID)

	return nil
}

// DeleteAllUserSessions session
func (m *MemorySessionManager) DeleteAllUserSessions(walletAddress string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	sessionIDs := m.userSessions[walletAddress]
	for sessionID := range sessionIDs {
		delete(m.sessions, sessionID)
	}
	delete(m.userSessions, walletAddress)

	return nil
}

// GetUserSessions session
func (m *MemorySessionManager) GetUserSessions(walletAddress string) ([]*Session, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	sessionIDs := m.userSessions[walletAddress]
	sessions := make([]*Session, 0, len(sessionIDs))

	for sessionID := range sessionIDs {
		if session, exists := m.sessions[sessionID]; exists {
			sessions = append(sessions, session)
		}
	}

	return sessions, nil
}

// cleanup session
func (m *MemorySessionManager) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		m.mu.Lock()
		now := time.Now()

		for sessionID, session := range m.sessions {
			// 清理過期或閒置的 session
			if now.Sub(session.CreatedAt) > m.timeout ||
				now.Sub(session.LastActiveAt) > m.idleTimeout {
				delete(m.sessions, sessionID)
				delete(m.userSessions[session.WalletAddress], sessionID)
			}
		}

		m.mu.Unlock()
	}
}

func generateSecureID() (string, error) {
	// 使用 UUID v4 生成 session ID (36 個字符,符合數據庫 VARCHAR(36) 定義)
	return uuid.New().String(), nil
}
