package blockchain

import (
	"bluelink-backend/internal/logger"
	"bluelink-backend/internal/models"
	"bluelink-backend/internal/repository"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	suiModels "github.com/block-vision/sui-go-sdk/models"
	"github.com/block-vision/sui-go-sdk/sui"
)

// EventListener Sui 區塊鏈事件監聽器
type EventListener struct {
	suiClient  sui.ISuiAPI                       // Sui 區塊鏈客戶端（查詢事件）
	txRepo     *repository.TransactionRepository // 交易 Repository
	bondRepo   *repository.BondRepository        // 債券 Repository
	userRepo   *repository.UserRepository        // 使用者 Repository
	packageID  string                            // 合約地址（過濾事件用）
	stopChan   chan struct{}                     // 停止信號通道
	isRunning  bool                              // 運行狀態
	lastCursor *suiModels.EventId                // 游標（追蹤查詢進度）
}

// NewEventListener 創建事件監聽器
func NewEventListener(
	suiClient sui.ISuiAPI,
	txRepo *repository.TransactionRepository,
	bondRepo *repository.BondRepository,
	userRepo *repository.UserRepository,
	packageID string,
) *EventListener {
	return &EventListener{
		suiClient:  suiClient,
		txRepo:     txRepo,
		bondRepo:   bondRepo,
		userRepo:   userRepo,
		packageID:  packageID,
		stopChan:   make(chan struct{}),
		isRunning:  false,
		lastCursor: nil,
	}
}

// Start 啟動事件監聽
func (el *EventListener) Start(ctx context.Context) error {
	if el.isRunning {
		return fmt.Errorf("event listener is already running")
	}

	el.isRunning = true
	logger.Info("Starting blockchain event listener...")

	// 使用輪詢方式查詢事件
	go el.pollEvents(ctx)

	return nil
}

// Stop 停止事件監聽
func (el *EventListener) Stop() {
	if !el.isRunning {
		return
	}

	logger.Info("Stopping blockchain event listener...")
	close(el.stopChan)
	el.isRunning = false
}

// pollEvents 輪詢查詢事件
func (el *EventListener) pollEvents(ctx context.Context) {
	ticker := time.NewTicker(15 * time.Second) // 每 5 秒查詢一次
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info("Context cancelled, stopping event listener")
			return
		case <-el.stopChan:
			logger.Info("Stop signal received")
			return
		case <-ticker.C:
			if err := el.queryAndProcessEvents(ctx); err != nil {
				logger.Error("Error querying events: %v", err)
			}
		}
	}
}

// queryAndProcessEvents 查詢並處理事件
func (el *EventListener) queryAndProcessEvents(ctx context.Context) error {
	// 使用 SuiX_QueryEvents API 查詢事件
	// 注意：Sui 節點不再支援 Package 過濾器，改用 MoveModule 過濾器
	// 這裡假設模組名為 "bond_project"，請根據實際合約調整
	request := suiModels.SuiXQueryEventsRequest{
		SuiEventFilter: suiModels.EventFilterByMoveModule{
			MoveModule: suiModels.MoveModule{
				Package: el.packageID,
				Module:  "blue_link", // 請根據你的合約模組名稱調整
			},
		},
		Cursor:          el.lastCursor,
		Limit:           50, // 每次查詢最多 50 個事件
		DescendingOrder: false,
	}

	// 查詢事件
	response, err := el.suiClient.SuiXQueryEvents(ctx, request)
	if err != nil {
		return fmt.Errorf("failed to query events: %w", err)
	}

	// 沒有新事件
	if len(response.Data) == 0 {
		return nil
	}

	logger.Info("Found %d new events", len(response.Data))

	// 處理每個事件
	for _, event := range response.Data {
		if err := el.handleEvent(ctx, event); err != nil {
			logger.Error("Error handling event %s: %v", event.Id.TxDigest, err)
			continue
		}
	}

	// 更新游標
	if response.HasNextPage {
		el.lastCursor = &response.NextCursor
	}

	return nil
}

// handleEvent 處理單個事件
func (el *EventListener) handleEvent(ctx context.Context, event suiModels.SuiEventResponse) error {
	logger.Debug("📬 Processing event type: %s from tx: %s", event.Type, event.Id.TxDigest)

	// 根據事件類型分發處理
	switch {
	case containsString(event.Type, "BondProjectCreated"):
		return el.handleBondProjectCreated(ctx, event)
	case containsString(event.Type, "BondTokensPurchased"):
		return el.handleBondTokensPurchased(ctx, event)
	case containsString(event.Type, "BondTokenRedeemed"):
		return el.handleBondTokenRedeemed(ctx, event)
	case containsString(event.Type, "RedemptionFundsDeposited"):
		return el.handleRedemptionFundsDeposited(ctx, event)
	case containsString(event.Type, "FundsWithdrawn"):
		return el.handleFundsWithdrawn(ctx, event)
	case containsString(event.Type, "SalePaused"):
		return el.handleSalePaused(ctx, event)
	case containsString(event.Type, "SaleResumed"):
		return el.handleSaleResumed(ctx, event)
	default:
		logger.Warn("Unknown event type: %s", event.Type)
		return nil
	}
}

// handleBondProjectCreated 處理債券專案創建事件
// Event: BondProjectCreated { id, issuer, issuer_name, bond_name, total_amount, annual_interest_rate, maturity_date, issue_date }
func (el *EventListener) handleBondProjectCreated(ctx context.Context, event suiModels.SuiEventResponse) error {
	// 檢查是否已處理
	existing, err := el.txRepo.GetByTxHash(ctx, event.Id.TxDigest)
	if err != nil {
		return fmt.Errorf("failed to check existing transaction: %w", err)
	}
	if existing != nil {
		logger.Debug("Transaction %s already processed, skipping", event.Id.TxDigest)
		return nil
	}

	// 解析事件數據
	bondID, _ := event.ParsedJson["id"].(string)
	bondName, _ := event.ParsedJson["bond_name"].(string)
	issuerName, _ := event.ParsedJson["issuer_name"].(string)
	issuerAddress := event.Sender

	// 查詢或創建發行者
	user, err := el.userRepo.GetByWalletAddress(ctx, issuerAddress)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		user, err = el.userRepo.Create(ctx, issuerAddress)
		if err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}
	}

	// 檢查債券是否已存在
	bond, err := el.bondRepo.GetByOnChainID(ctx, bondID)
	if err != nil {
		return fmt.Errorf("failed to get bond: %w", err)
	}

	if bond == nil {
		// 解析合約數據（合約使用 u64，這裡使用 int64）
		totalAmount := getInt64OrDefault(event.ParsedJson, "total_amount", 0)
		annualInterestRate := getInt64OrDefault(event.ParsedJson, "annual_interest_rate", 0) / 10000 // 轉換 basis points 到百分比
		maturityDateTimestamp := getInt64OrDefault(event.ParsedJson, "maturity_date", 0)
		issueDateTimestamp := getInt64OrDefault(event.ParsedJson, "issue_date", 0)

		// 從 Unix timestamp (ms) 轉換為日期字串 (YYYY-MM-DD)
		maturityTime := time.Unix(0, maturityDateTimestamp*int64(time.Millisecond))
		maturityDateStr := maturityTime.Format("2006-01-02")

		issueTime := time.Unix(0, issueDateTimestamp*int64(time.Millisecond))
		issueDateStr := issueTime.Format("2006-01-02")

		// 創建債券記錄（對應合約的 BondProject 結構）
		bond = &models.Bond{
			OnChainID:             bondID,
			IssuerAddress:         issuerAddress,
			IssuerName:            issuerName,
			BondName:              bondName,
			TotalAmount:           totalAmount,
			AmountRaised:          0, // 初始為 0
			AmountRedeemed:        0, // 初始為 0
			TokensIssued:          0, // 初始為 0
			TokensRedeemed:        0, // 初始為 0
			AnnualInterestRate:    annualInterestRate,
			MaturityDate:          maturityDateStr,
			IssueDate:             issueDateStr,
			Active:                true,  // 新創建的債券預設為 active
			Redeemable:            false, // 初始不可贖回
			RaisedFundsBalance:    0,     // 初始餘額為 0
			RedemptionPoolBalance: 0,     // 初始餘額為 0
		}

		if err := el.bondRepo.Create(ctx, bond); err != nil {
			return fmt.Errorf("failed to create bond: %w", err)
		}
	}

	// 創建交易記錄
	metadata, _ := MetadataToJSON(event.ParsedJson)
	timestamp := parseTimestamp(event.TimestampMs)

	tx := &models.Transaction{
		TxHash:        event.Id.TxDigest,
		EventType:     models.EventBondCreated,
		BondID:        &bond.ID,
		UserID:        &user.ID,
		WalletAddress: issuerAddress,
		Status:        models.TxStatusConfirmed,
		Timestamp:     timestamp,
		Metadata:      metadata,
	}

	if err := el.txRepo.Create(ctx, tx); err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	logger.Info("✅ Bond project created: %s (%s) by %s", bondName, bondID, issuerName)
	return nil
}

// handleBondTokensPurchased 處理債券代幣購買事件
// Event: BondTokensPurchased { project_id, buyer, token_id, amount }
func (el *EventListener) handleBondTokensPurchased(ctx context.Context, event suiModels.SuiEventResponse) error {
	// 檢查是否已處理
	existing, err := el.txRepo.GetByTxHash(ctx, event.Id.TxDigest)
	if err != nil {
		return err
	}
	if existing != nil {
		return nil
	}

	// 解析事件數據
	projectID, _ := event.ParsedJson["project_id"].(string)
	buyerAddress := event.Sender
	tokenID, _ := event.ParsedJson["token_id"].(string)
	amount := getFloat64OrDefault(event.ParsedJson, "amount", 0)

	// 查詢債券
	bond, err := el.bondRepo.GetByOnChainID(ctx, projectID)
	if err != nil || bond == nil {
		return fmt.Errorf("bond not found: %s", projectID)
	}

	// 查詢或創建使用者
	user, err := el.userRepo.GetByWalletAddress(ctx, buyerAddress)
	if err != nil {
		return err
	}
	if user == nil {
		user, err = el.userRepo.Create(ctx, buyerAddress)
		if err != nil {
			return err
		}
	}

	// 創建交易記錄並更新持倉
	metadata, _ := MetadataToJSON(event.ParsedJson)
	timestamp := parseTimestamp(event.TimestampMs)

	// 注意：合約中每次購買創建一個 NFT，數量為 1
	quantity := int64(1)
	price := amount // 購買金額就是價格

	tx := &models.Transaction{
		TxHash:        event.Id.TxDigest,
		EventType:     models.EventBondPurchased,
		BondID:        &bond.ID,
		UserID:        &user.ID,
		WalletAddress: buyerAddress,
		Quantity:      &quantity,
		Price:         &price,
		Amount:        &amount,
		Status:        models.TxStatusConfirmed,
		Timestamp:     timestamp,
		Metadata:      metadata,
	}

	// 使用事務創建交易並更新持倉
	if err := el.txRepo.CreateTransactionWithUserBond(ctx, tx, quantity, price); err != nil {
		return fmt.Errorf("failed to create transaction with user bond: %w", err)
	}

	logger.Info("✅ Bond purchased: %s bought token %s of %s (amount: %.2f SUI)",
		buyerAddress, tokenID, bond.BondName, amount)
	return nil
}

// handleBondTokenRedeemed 處理債券代幣贖回事件
// Event: BondTokenRedeemed { project_id, token_id, redeemer, redemption_amount }
func (el *EventListener) handleBondTokenRedeemed(ctx context.Context, event suiModels.SuiEventResponse) error {
	existing, err := el.txRepo.GetByTxHash(ctx, event.Id.TxDigest)
	if err != nil {
		return err
	}
	if existing != nil {
		return nil
	}

	projectID, _ := event.ParsedJson["project_id"].(string)
	redeemerAddress := event.Sender
	tokenID, _ := event.ParsedJson["token_id"].(string)
	redemptionAmount := getFloat64OrDefault(event.ParsedJson, "redemption_amount", 0)

	bond, err := el.bondRepo.GetByOnChainID(ctx, projectID)
	if err != nil || bond == nil {
		return fmt.Errorf("bond not found: %s", projectID)
	}

	user, err := el.userRepo.GetByWalletAddress(ctx, redeemerAddress)
	if err != nil || user == nil {
		return fmt.Errorf("user not found: %s", redeemerAddress)
	}

	metadata, _ := MetadataToJSON(event.ParsedJson)
	timestamp := parseTimestamp(event.TimestampMs)

	quantity := int64(1) // 贖回一個 NFT
	tx := &models.Transaction{
		TxHash:        event.Id.TxDigest,
		EventType:     models.EventBondRedeemed,
		BondID:        &bond.ID,
		UserID:        &user.ID,
		WalletAddress: redeemerAddress,
		Quantity:      &quantity,
		Amount:        &redemptionAmount,
		Status:        models.TxStatusConfirmed,
		Timestamp:     timestamp,
		Metadata:      metadata,
	}

	// 贖回時減少持倉
	if err := el.txRepo.CreateTransactionWithUserBond(ctx, tx, -quantity, 0); err != nil {
		return fmt.Errorf("failed to create transaction with user bond: %w", err)
	}

	logger.Info("✅ Bond redeemed: %s redeemed token %s (amount: %.2f SUI)",
		redeemerAddress, tokenID, redemptionAmount)
	return nil
}

// handleRedemptionFundsDeposited 處理贖回資金存入事件
// Event: RedemptionFundsDeposited { project_id, issuer, amount }
func (el *EventListener) handleRedemptionFundsDeposited(ctx context.Context, event suiModels.SuiEventResponse) error {
	existing, err := el.txRepo.GetByTxHash(ctx, event.Id.TxDigest)
	if err != nil {
		return err
	}
	if existing != nil {
		return nil
	}

	projectID, _ := event.ParsedJson["project_id"].(string)
	issuerAddress := event.Sender
	amount := getFloat64OrDefault(event.ParsedJson, "amount", 0)

	bond, err := el.bondRepo.GetByOnChainID(ctx, projectID)
	if err != nil || bond == nil {
		return fmt.Errorf("bond not found: %s", projectID)
	}

	user, err := el.userRepo.GetByWalletAddress(ctx, issuerAddress)
	if err != nil || user == nil {
		// 如果發行者不存在，創建一個
		user, err = el.userRepo.Create(ctx, issuerAddress)
		if err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}
	}

	metadata, _ := MetadataToJSON(event.ParsedJson)
	timestamp := parseTimestamp(event.TimestampMs)

	tx := &models.Transaction{
		TxHash:        event.Id.TxDigest,
		EventType:     "redemption_funds_deposited", // 新增的事件類型
		BondID:        &bond.ID,
		UserID:        &user.ID,
		WalletAddress: issuerAddress,
		Amount:        &amount,
		Status:        models.TxStatusConfirmed,
		Timestamp:     timestamp,
		Metadata:      metadata,
	}

	if err := el.txRepo.Create(ctx, tx); err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	logger.Info("✅ Redemption funds deposited: %s deposited %.2f SUI to %s",
		issuerAddress, amount, bond.BondName)
	return nil
}

// handleFundsWithdrawn 處理資金提取事件
// Event: FundsWithdrawn { project_id, withdrawer, amount }
func (el *EventListener) handleFundsWithdrawn(ctx context.Context, event suiModels.SuiEventResponse) error {
	existing, err := el.txRepo.GetByTxHash(ctx, event.Id.TxDigest)
	if err != nil {
		return err
	}
	if existing != nil {
		return nil
	}

	projectID, _ := event.ParsedJson["project_id"].(string)
	withdrawerAddress := event.Sender
	amount := getFloat64OrDefault(event.ParsedJson, "amount", 0)

	bond, err := el.bondRepo.GetByOnChainID(ctx, projectID)
	if err != nil || bond == nil {
		return fmt.Errorf("bond not found: %s", projectID)
	}

	user, err := el.userRepo.GetByWalletAddress(ctx, withdrawerAddress)
	if err != nil || user == nil {
		return fmt.Errorf("user not found: %s", withdrawerAddress)
	}

	metadata, _ := MetadataToJSON(event.ParsedJson)
	timestamp := parseTimestamp(event.TimestampMs)

	tx := &models.Transaction{
		TxHash:        event.Id.TxDigest,
		EventType:     "funds_withdrawn", // 新增的事件類型
		BondID:        &bond.ID,
		UserID:        &user.ID,
		WalletAddress: withdrawerAddress,
		Amount:        &amount,
		Status:        models.TxStatusConfirmed,
		Timestamp:     timestamp,
		Metadata:      metadata,
	}

	if err := el.txRepo.Create(ctx, tx); err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	logger.Info("Funds withdrawn: %s withdrew %.2f SUI from %s",
		withdrawerAddress, amount, bond.BondName)
	return nil
}

// handleSalePaused 處理銷售暫停事件
// Event: SalePaused { project_id, paused_by }
func (el *EventListener) handleSalePaused(ctx context.Context, event suiModels.SuiEventResponse) error {
	logger.Info("Sale paused: project %v paused by %s",
		event.ParsedJson["project_id"],
		event.Sender)
	return nil
}

// handleSaleResumed 處理銷售恢復事件
// Event: SaleResumed { project_id, resumed_by }
func (el *EventListener) handleSaleResumed(ctx context.Context, event suiModels.SuiEventResponse) error {
	logger.Info("Sale resumed: project %v resumed by %s",
		event.ParsedJson["project_id"],
		event.Sender)
	return nil
}

// 輔助函數
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && s[len(s)-len(substr):] == substr
}

func getFloat64OrDefault(m map[string]interface{}, key string, defaultValue float64) float64 {
	if v, ok := m[key].(float64); ok {
		return v
	}
	if v, ok := m[key].(int); ok {
		return float64(v)
	}
	if v, ok := m[key].(int64); ok {
		return float64(v)
	}
	return defaultValue
}

func getInt64OrDefault(m map[string]interface{}, key string, defaultValue int64) int64 {
	if v, ok := m[key].(int64); ok {
		return v
	}
	if v, ok := m[key].(int); ok {
		return int64(v)
	}
	if v, ok := m[key].(float64); ok {
		return int64(v)
	}
	// 處理字串格式的數字
	if v, ok := m[key].(string); ok {
		if parsed, err := strconv.ParseInt(v, 10, 64); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func parseTimestamp(timestampMs string) time.Time {
	ms, err := strconv.ParseInt(timestampMs, 10, 64)
	if err != nil {
		return time.Now()
	}
	return time.Unix(0, ms*int64(time.Millisecond))
}

func MetadataToJSON(data interface{}) (*string, error) {
	if data == nil {
		return nil, nil
	}
	bytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	str := string(bytes)
	return &str, nil
}
