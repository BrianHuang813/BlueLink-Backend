package services

import (
	"bluelink-backend/internal/models"
	"database/sql"
	"errors"
	"time"
)

// UserService 使用者服務層，處理使用者相關的業務邏輯
type UserService struct {
	db *sql.DB
}

// NewUserService 建立新的 UserService 實例
func NewUserService(db *sql.DB) *UserService {
	return &UserService{db: db}
}

// GetByID 根據 ID 取得完整使用者資料
func (s *UserService) GetByID(userID int64) (*models.User, error) {
	query := `
		SELECT id, wallet_address, name, email, role, created_at, updated_at 
		FROM users 
		WHERE id = $1
	`

	user := &models.User{}
	err := s.db.QueryRow(query, userID).Scan(
		&user.ID,
		&user.WalletAddress,
		&user.Name,
		&user.Role,
		&user.Timezone,
		&user.Language,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return user, nil
}

// GetByWalletAddress 根據錢包地址取得完整使用者資料
func (s *UserService) GetByWalletAddress(walletAddress string) (*models.User, error) {
	query := `
		SELECT id, wallet_address, name, email, role, created_at, updated_at 
		FROM users 
		WHERE wallet_address = $1
	`

	user := &models.User{}
	err := s.db.QueryRow(query, walletAddress).Scan(
		&user.ID,
		&user.WalletAddress,
		&user.Name,
		&user.Role,
		&user.Timezone,
		&user.Language,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return user, nil
}

// Create 建立新使用者
func (s *UserService) Create(walletAddress string) (*models.User, error) {
	query := `
		INSERT INTO users (wallet_address, role, created_at, updated_at) 
		VALUES ($1, 'investor', $2, $3) 
		RETURNING id, wallet_address, role, created_at, updated_at
	`

	now := time.Now()
	user := &models.User{}

	err := s.db.QueryRow(query, walletAddress, now, now).Scan(
		&user.ID,
		&user.WalletAddress,
		&user.Role,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return user, nil
}

// Update 更新使用者資料
func (s *UserService) Update(name string) error {
	query := `
		UPDATE users 
		SET name = $1
	`

	result, err := s.db.Exec(query, name, time.Now())
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("user not found")
	}

	return nil
}

// UpdateRole 更新使用者角色（管理員功能）
func (s *UserService) UpdateRole(userID int64, newRole string) error {
	// 驗證角色是否有效
	validRoles := map[string]bool{
		"investor": true,
		"issuer":   true,
		"admin":    true,
	}

	if !validRoles[newRole] {
		return errors.New("invalid role")
	}

	query := `
		UPDATE users 
		SET role = $1, updated_at = $2 
		WHERE id = $3
	`

	result, err := s.db.Exec(query, newRole, time.Now(), userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("user not found")
	}

	return nil
}

// Exists 檢查使用者是否存在
func (s *UserService) Exists(walletAddress string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE wallet_address = $1)`

	var exists bool
	err := s.db.QueryRow(query, walletAddress).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}
