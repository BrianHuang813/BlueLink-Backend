package services

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	customizedModel "bluelink-backend/internal/models"
)

// 債券服務層
type BondService struct {
	db *sql.DB
}

func NewBondService(db *sql.DB) *BondService {
	return &BondService{
		db: db,
	}
}

// GetAllBonds 獲取所有活躍債券
func (s *BondService) GetAllBonds(ctx context.Context) ([]*customizedModel.Bond, error) {
	query := `
        SELECT id, on_chain_id, issuer_address, name, description, project_type,
               total_amount, current_amount, interest_rate, maturity_date,
               minimum_purchase, status, impact_metrics, certification_docs,
               created_at, updated_at
        FROM bonds
        WHERE status = 'active'
        ORDER BY created_at DESC
    `

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query bonds: %w", err)
	}
	defer rows.Close() //沒有查詢到變關閉陣列變數，避免記憶體外洩等問題

	var bondsList []*customizedModel.Bond
	for rows.Next() {
		bond := &customizedModel.Bond{}
		err := rows.Scan(
			&bond.ID,
			&bond.OnChainID,
			&bond.Name,
			&bond.IssuerName,
			&bond.IssuerAddress,
			&bond.Description,
			&bond.FaceValue,
			&bond.Currency,
			&bond.MaturityDate,
			&bond.InterestRate,
			&bond.Status,
			&bond.CreatedAt,
			&bond.UpdatedAt,
		)
		if err != nil {
			slog.Error("bonds scanning error", "error", err)
			return nil, fmt.Errorf("failed to scan bond: %w", err)
		}
		bondsList = append(bondsList, bond)
	}

	if err = rows.Err(); err != nil {
		slog.Error("rows iteration error", "error", err)
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return bondsList, nil
}

// /*
// 	債券創建總服務邏輯
// */
// // CreateBond 創建新債券
// func (s *BondService) CreateBond(
// 	ctx context.Context,
// 	issuerAddress string,
// ) (*customizedModel.Bond, string, error) {

// 	// 執行鏈上操作 - 創建債券
// 	txDigest, onChainID, err := s.createBondOnChain(ctx, issuerAddress)
// 	if err != nil {
// 		return nil, "", fmt.Errorf("failed to create bond on chain: %w", err)
// 	}

// 	// 3. 保存到數據庫
// 	bond, err := s.saveBondToDB(ctx, onChainID, issuerAddress)
// 	if err != nil {
// 		// TODO: 這裡應該考慮回滾鏈上操作或記錄異常
// 		return nil, "", fmt.Errorf("failed to save bond to database: %w", err)
// 	}

// 	return bond, txDigest, nil
// }

// // createBondOnChain 在鏈上創建債券
// func (s *BondService) createBondOnChain(
// 	ctx context.Context,
// 	issuerAddress string,
// ) (string, string, error) {
// 	// TODO: 實作與 Sui 智能合約的交互
// 	// 1. 構建交易參數
// 	// 2. 調用合約的 create_bond 函數
// 	// 3. 等待交易確認
// 	// 4. 從事件中提取債券 ID

// 	gasObj := "0x58c103930dc52c0ab86319d99218e301596fda6fd80c4efafd7f4c9df1d0b6d0"
// 	cli := s.suiClient

// 	rsp, err := cli.MoveCall(ctx, models.MoveCallRequest{
// 		Signer:          issuerAddress,
// 		PackageObjectId: "0x7d584c9a27ca4a546e8203b005b0e9ae746c9bec6c8c3c0bc84611bcf4ceab5f",
// 		Module:          "auction",
// 		Function:        "create_bond",
// 		TypeArguments:   []interface{}{},
// 		Arguments: []interface{}{
// 			"0x342e959f8d9d1fa9327a05fd54fefd929bbedad47190bdbb58743d8ba3bd3420",
// 			"0x3fd0fdedb84cf1f59386b6251ba6dd2cb495094da26e0a5a38239acd9d437f96",
// 			"0xb3de4235cb04167b473de806d00ba351e5860500253cf8e62d711e578e1d92ae",
// 			"BlockVision",
// 			"0xc699c6014da947778fe5f740b2e9caf905ca31fb4c81e346f467ae126e3c03f1",
// 		},
// 		Gas:       &gasObj,
// 		GasBudget: "100000000",
// 	})

// 	if err != nil {
// 		fmt.Println(err.Error())
// 		return "", "", err
// 	}

// 	// 暫時返回模擬數據
// 	txDigest := "0x" + fmt.Sprintf("%064d", time.Now().Unix())
// 	onChainBondID := "bond_" + fmt.Sprintf("%d", time.Now().Unix())

// 	return txDigest, onChainBondID, nil
// }

// // saveBondToDB 將債券信息保存到數據庫
// func (s *BondService) saveBondToDB(
// 	ctx context.Context,
// 	onChainBondID string,
// 	issuerAddress string,
// ) (*customizedModel.Bond, error) {
// 	query := `
//         INSERT INTO bonds (
//             on_chain_id, issuer_address, name, description, project_type,
//             total_amount, current_amount, interest_rate, maturity_date,
//             minimum_purchase, status, impact_metrics, certification_docs,
//             created_at, updated_at
//         ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
//         RETURNING id, on_chain_id, issuer_address, name, description, project_type,
//                   total_amount, current_amount, interest_rate, maturity_date,
//                   minimum_purchase, status, created_at, updated_at
//     `

// 	now := time.Now()
// 	bond := &models.Bond{}

// 	// TODO
// 	err := s.db.QueryRowContext(
// 		ctx,
// 		query,
// 		onChainBondID,
// 		issuerAddress,
// 		req.Name,
// 		req.Description,
// 		req.ProjectType,
// 		req.TotalAmount,
// 		0, // current_amount 初始為 0
// 		req.InterestRate,
// 		req.MaturityDate,
// 		req.MinimumPurchase,
// 		"active",
// 		req.ImpactMetrics,
// 		req.CertificationDocs,
// 		now,
// 		now,
// 	).Scan(
// 		&bond.ID,
// 		&bond.OnChainID,
// 		&bond.Name,
// 		&bond.IssuerName,
// 		&bond.IssuerAddress,
// 		&bond.Description,
// 		&bond.FaceValue,
// 		&bond.Currency,
// 		&bond.MaturityDate,
// 		&bond.InterestRate,
// 		&bond.Status,
// 		&bond.CreatedAt,
// 		&bond.UpdatedAt,
// 	)

// 	if err != nil {
// 		return nil, err
// 	}

// 	return bond, nil
// }

// // BuyBond 購買債券（包含鏈上操作）
// func (s *BondService) BuyBond(
// 	ctx context.Context,
// 	buyerAddress string,
// 	bondID string,
// 	amount uint64,
// ) (*customizedModel.BondPurchase, string, error) {
// 	// 1. 查找並驗證債券
// 	bond, err := s.getBondByOnChainID(ctx, bondID)
// 	if err != nil {
// 		return nil, "", err
// 	}

// 	// 2. 驗證購買條件
// 	if err := s.validatePurchase(bond, amount); err != nil {
// 		return nil, "", err
// 	}

// 	// 3. 執行鏈上購買操作
// 	txDigest, err := s.buyBondOnChain(ctx, buyerAddress, bondID, amount)
// 	if err != nil {
// 		return nil, "", fmt.Errorf("failed to buy bond on chain: %w", err)
// 	}

// 	// 4. 更新數據庫（使用事務）
// 	purchase, err := s.recordPurchase(ctx, bond.ID, buyerAddress, amount, txDigest)
// 	if err != nil {
// 		// TODO: 記錄鏈上成功但數據庫失敗的情況
// 		return nil, "", fmt.Errorf("failed to record purchase: %w", err)
// 	}

// 	return purchase, txDigest, nil
// }

// // getBondByOnChainID 根據鏈上 ID 獲取債券
// func (s *BondService) getBondByOnChainID(ctx context.Context, onChainID string) (*models.Bond, error) {
// 	query := `
//         SELECT id, on_chain_id, issuer_address, name, description, project_type,
//                total_amount, current_amount, interest_rate, maturity_date,
//                minimum_purchase, status, created_at, updated_at
//         FROM bonds
//         WHERE on_chain_id = $1 AND status = 'active'
//     `

// 	bond := &models.Bond{}
// 	err := s.db.QueryRowContext(ctx, query, onChainID).Scan(
// 		&bond.ID,
// 		&bond.OnChainID,
// 		&bond.IssuerAddress,
// 		&bond.Name,
// 		&bond.Description,
// 		&bond.ProjectType,
// 		&bond.TotalAmount,
// 		&bond.CurrentAmount,
// 		&bond.InterestRate,
// 		&bond.MaturityDate,
// 		&bond.MinimumPurchase,
// 		&bond.Status,
// 		&bond.CreatedAt,
// 		&bond.UpdatedAt,
// 	)

// 	if err != nil {
// 		if err == sql.ErrNoRows {
// 			return nil, errors.New("bond not found or inactive")
// 		}
// 		return nil, fmt.Errorf("failed to query bond: %w", err)
// 	}

// 	return bond, nil
// }

// // validatePurchase 驗證購買條件
// func (s *BondService) validatePurchase(bond *models.Bond, amount uint64) error {
// 	// 檢查最小購買金額
// 	if amount < bond.MinimumPurchase {
// 		return fmt.Errorf("amount %d is below minimum purchase %d", amount, bond.MinimumPurchase)
// 	}

// 	// 檢查剩餘額度
// 	remainingAmount := bond.TotalAmount - bond.CurrentAmount
// 	if amount > remainingAmount {
// 		return fmt.Errorf("insufficient bond capacity: requested %d, available %d", amount, remainingAmount)
// 	}

// 	// 檢查是否已到期
// 	if time.Now().After(bond.MaturityDate) {
// 		return errors.New("bond has matured")
// 	}

// 	return nil
// }

// // buyBondOnChain 在鏈上執行購買操作
// func (s *BondService) buyBondOnChain(
// 	ctx context.Context,
// 	buyerAddress string,
// 	bondID string,
// 	amount uint64,
// ) (string, error) {
// 	// TODO: 實作與 Sui 智能合約的交互
// 	// 1. 構建交易參數
// 	// 2. 調用合約的 buy_bond 函數
// 	// 3. 處理 SUI 代幣轉移
// 	// 4. 等待交易確認

// 	// 暫時返回模擬數據
// 	txDigest := "0x" + fmt.Sprintf("%064d", time.Now().Unix())

// 	return txDigest, nil
// }

// // recordPurchase 記錄購買交易（使用數據庫事務）
// func (s *BondService) recordPurchase(
// 	ctx context.Context,
// 	bondID int64,
// 	buyerAddress string,
// 	amount uint64,
// 	txDigest string,
// ) (*customizedModel.BondPurchase, error) {
// 	tx, err := s.db.BeginTx(ctx, nil)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to begin transaction: %w", err)
// 	}
// 	defer tx.Rollback()

// 	// 1. 創建購買記錄
// 	insertQuery := `
//         INSERT INTO bond_purchases (
//             bond_id, buyer_address, amount, purchase_price, transaction_id, status, created_at
//         ) VALUES ($1, $2, $3, $4, $5, $6, $7)
//         RETURNING id, bond_id, buyer_address, amount, purchase_price, transaction_id, status, created_at
//     `

// 	purchase := &models.BondPurchase{}
// 	now := time.Now()
// 	purchasePrice := amount // 假設 1:1，實際可能需要計算

// 	err = tx.QueryRowContext(
// 		ctx,
// 		insertQuery,
// 		bondID,
// 		buyerAddress,
// 		amount,
// 		purchasePrice,
// 		txDigest,
// 		"active",
// 		now,
// 	).Scan(
// 		&purchase.ID,
// 		&purchase.BondID,
// 		&purchase.BuyerAddress,
// 		&purchase.Amount,
// 		&purchase.PurchasePrice,
// 		&purchase.TransactionID,
// 		&purchase.Status,
// 		&purchase.CreatedAt,
// 	)

// 	if err != nil {
// 		return nil, fmt.Errorf("failed to insert purchase: %w", err)
// 	}

// 	// 2. 更新債券的已募集金額
// 	updateQuery := `
//         UPDATE bonds
//         SET current_amount = current_amount + $1,
//             updated_at = $2
//         WHERE id = $3
//     `

// 	result, err := tx.ExecContext(ctx, updateQuery, amount, now, bondID)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to update bond amount: %w", err)
// 	}

// 	rowsAffected, err := result.RowsAffected()
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to get rows affected: %w", err)
// 	}

// 	if rowsAffected == 0 {
// 		return nil, errors.New("bond not found during update")
// 	}

// 	// 提交事務
// 	if err := tx.Commit(); err != nil {
// 		return nil, fmt.Errorf("failed to commit transaction: %w", err)
// 	}

// 	return purchase, nil
// }
