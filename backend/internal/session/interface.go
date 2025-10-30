package session

// SessionManager 定義 Session 管理介面
type SessionManager interface {
	// Create 建立新的 Session
	Create(userID int64, walletAddress, role string, ipAddress, userAgent string) (*Session, error)

	// Get 取得 Session
	Get(sessionID string) (*Session, error)

	// Touch 更新 Session 的最後活躍時間
	Touch(sessionID string) error

	// Delete 刪除特定的 Session
	Delete(sessionID string) error

	// DeleteAllUserSessions 刪除特定錢包地址的所有 Session
	DeleteAllUserSessions(walletAddress string) error

	// GetUserSessions 取得特定錢包地址的所有 Session
	GetUserSessions(walletAddress string) ([]*Session, error)
}
