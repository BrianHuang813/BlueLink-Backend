package services

import (
	"bluelink-backend/internal/logger"
	"bluelink-backend/internal/models"
	"bluelink-backend/internal/repository"
	"context"
)

// BondTokenService 債券代幣服務層
type BondTokenService struct {
	repo *repository.BondTokenRepository
}

// NewBondTokenService 建立新的 BondTokenService 實例
func NewBondTokenService(repo *repository.BondTokenRepository) *BondTokenService {
	return &BondTokenService{repo: repo}
}

// GetBondTokenByID 根據 ID 獲取債券代幣
func (s *BondTokenService) GetBondTokenByID(ctx context.Context, id int64) (*models.BondToken, error) {
	token, err := s.repo.GetByID(ctx, id)
	if err != nil {
		logger.Error("Failed to get bond token by ID %d: %v", id, err)
		return nil, err
	}
	return token, nil
}

// GetBondTokenByOnChainID 根據鏈上 ID 獲取債券代幣
func (s *BondTokenService) GetBondTokenByOnChainID(ctx context.Context, onChainID string) (*models.BondToken, error) {
	token, err := s.repo.GetByOnChainID(ctx, onChainID)
	if err != nil {
		logger.Error("Failed to get bond token by on-chain ID %s: %v", onChainID, err)
		return nil, err
	}
	return token, nil
}

// GetBondTokensByOwner 根據擁有者地址獲取債券代幣列表
func (s *BondTokenService) GetBondTokensByOwner(ctx context.Context, owner string, limit, offset int) ([]*models.BondToken, error) {
	if limit <= 0 {
		limit = 100
	}

	tokens, err := s.repo.GetByOwner(ctx, owner, limit, offset)
	if err != nil {
		logger.Error("Failed to get bond tokens by owner %s: %v", owner, err)
		return nil, err
	}
	return tokens, nil
}

// GetBondTokensByProjectID 根據專案 ID 獲取債券代幣列表
func (s *BondTokenService) GetBondTokensByProjectID(ctx context.Context, projectID string, limit, offset int) ([]*models.BondToken, error) {
	if limit <= 0 {
		limit = 100
	}

	tokens, err := s.repo.GetByProjectID(ctx, projectID, limit, offset)
	if err != nil {
		logger.Error("Failed to get bond tokens by project ID %s: %v", projectID, err)
		return nil, err
	}
	return tokens, nil
}

// CreateBondToken 創建新債券代幣
func (s *BondTokenService) CreateBondToken(ctx context.Context, token *models.BondToken) error {
	if err := s.repo.Create(ctx, token); err != nil {
		logger.Error("Failed to create bond token for project %s: %v", token.ProjectID, err)
		return err
	}
	logger.Info("Bond token created successfully: Project=%s, TokenNumber=%d, Owner=%s", 
		token.ProjectID, token.TokenNumber, token.Owner)
	return nil
}

// UpdateBondTokenRedeemed 更新債券代幣贖回狀態
func (s *BondTokenService) UpdateBondTokenRedeemed(ctx context.Context, id int64, isRedeemed bool) error {
	if err := s.repo.UpdateRedeemed(ctx, id, isRedeemed); err != nil {
		logger.Error("Failed to update bond token redeemed status for ID %d: %v", id, err)
		return err
	}
	logger.Info("Bond token redeemed status updated: ID=%d, isRedeemed=%v", id, isRedeemed)
	return nil
}
