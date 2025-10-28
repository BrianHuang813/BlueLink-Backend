package services

import (
	"bluelink-backend/internal/logger"
	"bluelink-backend/internal/models"
	"bluelink-backend/internal/repository"
	"context"
)

// BondService 債券服務層
type BondService struct {
	repo *repository.BondRepository
}

// NewBondService 建立新的 BondService 實例
func NewBondService(repo *repository.BondRepository) *BondService {
	return &BondService{repo: repo}
}

// GetAllBonds 獲取所有活躍債券
func (s *BondService) GetAllBonds(ctx context.Context) ([]*models.Bond, error) {
	bonds, err := s.repo.List(ctx, 100, 0)
	if err != nil {
		logger.Error("Failed to get all bonds: %v", err)
		return nil, err
	}
	return bonds, nil
}

// GetBondByID 根據 ID 獲取債券
func (s *BondService) GetBondByID(ctx context.Context, id int64) (*models.Bond, error) {
	bond, err := s.repo.GetByID(ctx, id)
	if err != nil {
		logger.Error("Failed to get bond by ID %d: %v", id, err)
		return nil, err
	}
	return bond, nil
}

// GetBondByOnChainID 根據鏈上 ID 獲取債券
func (s *BondService) GetBondByOnChainID(ctx context.Context, onChainID string) (*models.Bond, error) {
	bond, err := s.repo.GetByOnChainID(ctx, onChainID)
	if err != nil {
		logger.Error("Failed to get bond by on-chain ID %s: %v", onChainID, err)
		return nil, err
	}
	return bond, nil
}

// CreateBond 創建新債券
func (s *BondService) CreateBond(ctx context.Context, bond *models.Bond) error {
	if err := s.repo.Create(ctx, bond); err != nil {
		logger.Error("Failed to create bond %s: %v", bond.BondName, err)
		return err
	}
	logger.Info("Bond created successfully: %s (ID: %d)", bond.BondName, bond.ID)
	return nil
}

// UpdateBondStatus 更新債券狀態
func (s *BondService) UpdateBondStatus(ctx context.Context, id int64, active, redeemable bool) error {
	if err := s.repo.UpdateStatus(ctx, id, active, redeemable); err != nil {
		logger.Error("Failed to update bond status for ID %d: %v", id, err)
		return err
	}
	logger.Info("Bond status updated: ID=%d, active=%v, redeemable=%v", id, active, redeemable)
	return nil
}
