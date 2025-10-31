package services

import (
	"bluelink-backend/internal/logger"
	"bluelink-backend/internal/models"
	"bluelink-backend/internal/repository"
	"context"
	"errors"
)

// UserService 使用者服務層，處理使用者相關的業務邏輯
type UserService struct {
	repo *repository.UserRepository
}

// NewUserService 建立新的 UserService 實例
func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

// GetByID 根據 ID 取得完整使用者資料
func (s *UserService) GetByID(ctx context.Context, userID int64) (*models.User, error) {
	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		logger.Error("Failed to get user by ID %d: %v", userID, err)
		return nil, err
	}
	if user == nil {
		logger.Warn("User not found: ID=%d", userID)
		return nil, errors.New("user not found")
	}
	return user, nil
}

// GetByWalletAddress 根據錢包地址取得完整使用者資料
func (s *UserService) GetByWalletAddress(ctx context.Context, walletAddress string) (*models.User, error) {
	user, err := s.repo.GetByWalletAddress(ctx, walletAddress)
	if err != nil {
		logger.Error("Failed to get user by wallet address %s: %v", walletAddress, err)
		return nil, err
	}
	if user == nil {
		logger.Warn("User not found: wallet=%s", walletAddress)
		return nil, errors.New("user not found")
	}
	return user, nil
}

// Create 建立新使用者（預設為 buyer）
func (s *UserService) Create(ctx context.Context, walletAddress string) (*models.User, error) {
	return s.CreateWithRole(ctx, walletAddress, "buyer")
}

// CreateWithRole 建立新使用者並指定角色
func (s *UserService) CreateWithRole(ctx context.Context, walletAddress, role string) (*models.User, error) {
	// 驗證角色
	validRoles := map[string]bool{
		"buyer":  true,
		"issuer": true,
		"admin":  false, // admin 不允許自行註冊
	}

	if !validRoles[role] || role == "admin" {
		logger.Warn("Invalid role attempted during registration: %s for wallet %s", role, walletAddress)
		role = "buyer" // 無效角色時預設為 buyer
	}

	user, err := s.repo.CreateWithRole(ctx, walletAddress, role)
	if err != nil {
		logger.Error("Failed to create user with wallet %s and role %s: %v", walletAddress, role, err)
		return nil, err
	}
	logger.Info("User created successfully: ID=%d, wallet=%s, role=%s", user.ID, walletAddress, role)
	return user, nil
}

// Update 更新使用者資料
func (s *UserService) Update(ctx context.Context, user *models.User) error {
	if err := s.repo.Update(ctx, user); err != nil {
		logger.Error("Failed to update user ID %d: %v", user.ID, err)
		return err
	}
	logger.Info("User updated successfully: ID=%d", user.ID)
	return nil
}

// UpdateRole 更新使用者角色（管理員功能）
func (s *UserService) UpdateRole(ctx context.Context, userID int64, newRole string) error {
	validRoles := map[string]bool{
		"buyer":  true,
		"issuer": true,
		"admin":  true,
	}

	if !validRoles[newRole] {
		logger.Warn("Invalid role attempted: %s for user ID %d", newRole, userID)
		return errors.New("invalid role")
	}

	user, err := s.repo.GetByID(ctx, userID)
	if err != nil {
		logger.Error("Failed to get user for role update, ID %d: %v", userID, err)
		return err
	}
	if user == nil {
		logger.Warn("User not found for role update: ID=%d", userID)
		return errors.New("user not found")
	}

	user.Role = newRole
	if err := s.repo.Update(ctx, user); err != nil {
		logger.Error("Failed to update user role, ID %d: %v", userID, err)
		return err
	}

	logger.Info("User role updated: ID=%d, new_role=%s", userID, newRole)
	return nil
}

// Exists 檢查使用者是否存在
func (s *UserService) Exists(ctx context.Context, walletAddress string) (bool, error) {
	user, err := s.repo.GetByWalletAddress(ctx, walletAddress)
	if err != nil {
		logger.Error("Failed to check user existence for wallet %s: %v", walletAddress, err)
		return false, err
	}
	return user != nil, nil
}
