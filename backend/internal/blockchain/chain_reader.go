package blockchain

import (
	"bluelink-backend/internal/logger"
	"bluelink-backend/internal/models"
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	suiModels "github.com/block-vision/sui-go-sdk/models"
	"github.com/block-vision/sui-go-sdk/sui"
)

// ChainReader 從 Sui 鏈上讀取對象數據
type ChainReader struct {
	suiClient sui.ISuiAPI
	packageID string
}

// NewChainReader 創建鏈上數據讀取器
func NewChainReader(suiClient sui.ISuiAPI, packageID string) *ChainReader {
	return &ChainReader{
		suiClient: suiClient,
		packageID: packageID,
	}
}

// BondProjectOnChain 鏈上 BondProject 對象的數據結構
type BondProjectOnChain struct {
	ObjectID           string
	BondName           string
	Issuer             string
	IssuerName         string
	BondImageUrl       string
	TokenImageUrl      string
	MetadataUrl        string
	TotalAmount        int64
	AmountRaised       int64
	AmountRedeemed     int64
	TokensIssued       int64
	TokensRedeemed     int64
	AnnualInterestRate int64
	MaturityDate       int64 // timestamp in milliseconds
	IssueDate          int64 // timestamp in milliseconds
	Active             bool
	Redeemable         bool
}

// GetBondProjectFromTransaction 從交易中提取並讀取 BondProject 對象
func (cr *ChainReader) GetBondProjectFromTransaction(ctx context.Context, txDigest string) (*BondProjectOnChain, error) {
	// 1. 獲取交易詳情
	txResp, err := cr.suiClient.SuiGetTransactionBlock(ctx, suiModels.SuiGetTransactionBlockRequest{
		Digest: txDigest,
		Options: suiModels.SuiTransactionBlockOptions{
			ShowEffects:       true,
			ShowObjectChanges: true,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	// 2. 從 ObjectChanges 中找到創建的 BondProject 對象
	bondProjectID := ""
	if len(txResp.ObjectChanges) > 0 {
		for _, change := range txResp.ObjectChanges {
			// 檢查是否為創建操作
			if change.Type == "created" {
				// 檢查對象類型是否為 BondProject
				objectType := change.GetObjectChangeAddressOwner()
				if strings.Contains(objectType, "::blue_link::BondProject") ||
					strings.Contains(change.ObjectType, "BondProject") {
					bondProjectID = change.ObjectId
					logger.Info("Found BondProject object: %s", bondProjectID)
					break
				}
			}
		}
	}

	if bondProjectID == "" {
		return nil, fmt.Errorf("no BondProject object found in transaction")
	}

	// 3. 從鏈上讀取完整的 BondProject 對象數據
	return cr.GetBondProjectByID(ctx, bondProjectID)
}

// GetBondProjectByID 根據對象 ID 讀取 BondProject 數據
func (cr *ChainReader) GetBondProjectByID(ctx context.Context, objectID string) (*BondProjectOnChain, error) {
	// 調用 sui_getObject
	resp, err := cr.suiClient.SuiGetObject(ctx, suiModels.SuiGetObjectRequest{
		ObjectId: objectID,
		Options: suiModels.SuiObjectDataOptions{
			ShowContent: true,
			ShowType:    true,
			ShowOwner:   true,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get object %s: %w", objectID, err)
	}

	// 檢查對象是否存在
	if resp.Data == nil {
		return nil, fmt.Errorf("object %s not found", objectID)
	}

	// 解析對象內容
	content := resp.Data.Content
	if content == nil {
		return nil, fmt.Errorf("object %s has no content", objectID)
	}

	// 檢查是否為 MoveObject
	if content.DataType != "moveObject" {
		return nil, fmt.Errorf("object %s is not a move object", objectID)
	}

	// 解析 fields
	fields := content.Fields
	if fields == nil {
		return nil, fmt.Errorf("object %s has no fields", objectID)
	}

	// 提取數據
	bondProject := &BondProjectOnChain{
		ObjectID:           objectID,
		BondName:           getStringField(fields, "bond_name"),
		Issuer:             getStringField(fields, "issuer"),
		IssuerName:         getStringField(fields, "issuer_name"),
		BondImageUrl:       getStringField(fields, "bond_image_url"),
		TokenImageUrl:      getStringField(fields, "token_image_url"),
		MetadataUrl:        getStringField(fields, "metadata_url"),
		TotalAmount:        getInt64Field(fields, "total_amount"),
		AmountRaised:       getInt64Field(fields, "amount_raised"),
		AmountRedeemed:     getInt64Field(fields, "amount_redeemed"),
		TokensIssued:       getInt64Field(fields, "tokens_issued"),
		TokensRedeemed:     getInt64Field(fields, "tokens_redeemed"),
		AnnualInterestRate: getInt64Field(fields, "annual_interest_rate"),
		MaturityDate:       getInt64Field(fields, "maturity_date"),
		IssueDate:          getInt64Field(fields, "issue_date"),
		Active:             getBoolField(fields, "active"),
		Redeemable:         getBoolField(fields, "redeemable"),
	}

	// 詳細的調試日誌
	logger.Info("📊 Bond data from chain: %s", bondProject.BondName)
	logger.Info("   🆔 Object ID: %s", objectID)
	logger.Info("   👤 Issuer: %s", bondProject.Issuer)
	logger.Info("   🏢 Issuer Name: %s", bondProject.IssuerName)
	logger.Info("   💰 Total Amount: %d MIST (%.2f SUI)",
		bondProject.TotalAmount,
		float64(bondProject.TotalAmount)/1e9)
	logger.Info("   📈 Amount Raised: %d MIST (%.2f SUI)",
		bondProject.AmountRaised,
		float64(bondProject.AmountRaised)/1e9)
	logger.Info("   📊 Annual Interest Rate: %d (%.2f%%)",
		bondProject.AnnualInterestRate,
		float64(bondProject.AnnualInterestRate)/100)
	logger.Info("   📅 Issue Date: %d", bondProject.IssueDate)
	logger.Info("   📅 Maturity Date: %d", bondProject.MaturityDate)
	logger.Info("   ✅ Active: %v, Redeemable: %v", bondProject.Active, bondProject.Redeemable)

	return bondProject, nil
}

// ToBondModel 將鏈上數據轉換為數據庫模型
func (bp *BondProjectOnChain) ToBondModel() *models.Bond {
	// 從 Unix timestamp (ms) 轉換為日期字串 (YYYY-MM-DD)
	maturityTime := time.Unix(0, bp.MaturityDate*int64(time.Millisecond))
	maturityDateStr := maturityTime.UTC().Format("2006-01-02")

	issueTime := time.Unix(0, bp.IssueDate*int64(time.Millisecond))
	issueDateStr := issueTime.UTC().Format("2006-01-02")

	return &models.Bond{
		OnChainID:          bp.ObjectID,
		IssuerAddress:      bp.Issuer,
		IssuerName:         bp.IssuerName,
		BondName:           bp.BondName,
		BondImageUrl:       bp.BondImageUrl,
		TokenImageUrl:      bp.TokenImageUrl,
		MetadataUrl:        bp.MetadataUrl,
		TotalAmount:        bp.TotalAmount,
		AmountRaised:       bp.AmountRaised,
		AmountRedeemed:     bp.AmountRedeemed,
		TokensIssued:       bp.TokensIssued,
		TokensRedeemed:     bp.TokensRedeemed,
		AnnualInterestRate: bp.AnnualInterestRate,
		MaturityDate:       maturityDateStr,
		IssueDate:          issueDateStr,
		Active:             bp.Active,
		Redeemable:         bp.Redeemable,
		// 初始餘額設為 0，後續由事件更新
		RaisedFundsBalance:    0,
		RedemptionPoolBalance: 0,
	}
}

// 輔助函數：安全地從 fields map 中提取字符串
func getStringField(fields map[string]interface{}, key string) string {
	if val, ok := fields[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// 輔助函數：安全地從 fields map 中提取 int64
func getInt64Field(fields map[string]interface{}, key string) int64 {
	if val, ok := fields[key]; ok {
		switch v := val.(type) {
		case string:
			// 處理字符串格式的數字
			if num, err := strconv.ParseInt(v, 10, 64); err == nil {
				return num
			}
		case float64:
			return int64(v)
		case int64:
			return v
		case int:
			return int64(v)
		case uint64:
			return int64(v)
		}
	}

	logger.Warn("⚠️ Failed to parse field '%s' as int64, value: %v (type: %T)", key, fields[key], fields[key])
	return 0
}

// 輔助函數：安全地從 fields map 中提取布爾值
func getBoolField(fields map[string]interface{}, key string) bool {
	if val, ok := fields[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return false
}
