package services

import (
	"bluelink-backend/internal/blockchain"
	"bluelink-backend/internal/logger"
	"bluelink-backend/internal/repository"
	"context"
	"fmt"

	"github.com/block-vision/sui-go-sdk/sui"
)

// SyncService 同步服務
type SyncService struct {
	chainReader *blockchain.ChainReader
	bondRepo    *repository.BondRepository
	userRepo    *repository.UserRepository
	txRepo      *repository.TransactionRepository
}

// NewSyncService 創建同步服務
func NewSyncService(
	suiClient sui.ISuiAPI,
	packageID string,
	bondRepo *repository.BondRepository,
	userRepo *repository.UserRepository,
	txRepo *repository.TransactionRepository,
) *SyncService {
	return &SyncService{
		chainReader: blockchain.NewChainReader(suiClient, packageID),
		bondRepo:    bondRepo,
		userRepo:    userRepo,
		txRepo:      txRepo,
	}
}

// SyncBondCreated 同步債券創建事件
func (s *SyncService) SyncBondCreated(ctx context.Context, txDigest string) error {
	logger.Info("🔄 Syncing bond created transaction: %s", txDigest)

	// 檢查是否已處理
	existing, err := s.txRepo.GetByTxHash(ctx, txDigest)
	if err != nil {
		return fmt.Errorf("failed to check existing transaction: %w", err)
	}
	if existing != nil {
		logger.Info("Transaction %s already processed", txDigest)
		return nil
	}

	// 從鏈上讀取 BondProject 數據
	bondData, err := s.chainReader.GetBondProjectFromTransaction(ctx, txDigest)
	if err != nil {
		return fmt.Errorf("failed to get bond project from chain: %w", err)
	}

	// 查詢或創建發行者
	user, err := s.userRepo.GetByWalletAddress(ctx, bondData.Issuer)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		user, err = s.userRepo.Create(ctx, bondData.Issuer)
		if err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}
	}

	// 檢查債券是否已存在
	existingBond, err := s.bondRepo.GetByOnChainID(ctx, bondData.ObjectID)
	if err != nil {
		return fmt.Errorf("failed to get bond: %w", err)
	}

	var bond = existingBond
	if bond == nil {
		// 創建新債券
		bond = bondData.ToBondModel()
		if err := s.bondRepo.Create(ctx, bond); err != nil {
			return fmt.Errorf("failed to create bond: %w", err)
		}

		logger.Info("✅ Bond synced successfully:")
		logger.Info("   📋 Name: %s", bond.BondName)
		logger.Info("   🆔 ID: %s", bond.OnChainID)
		logger.Info("   💰 Total Amount: %d MIST (%.2f SUI)",
			bond.TotalAmount,
			float64(bond.TotalAmount)/1e9)
	}

	return nil
}

// SyncBondPurchased 同步債券購買事件
func (s *SyncService) SyncBondPurchased(ctx context.Context, txDigest string) error {
	logger.Info("🔄 Syncing bond purchased transaction: %s", txDigest)
	// TODO: 實現購買事件同步邏輯
	return nil
}

// SyncBondRedeemed 同步債券贖回事件
func (s *SyncService) SyncBondRedeemed(ctx context.Context, txDigest string) error {
	logger.Info("🔄 Syncing bond redeemed transaction: %s", txDigest)
	// TODO: 實現贖回事件同步邏輯
	return nil
}
