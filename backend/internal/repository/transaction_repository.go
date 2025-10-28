package repository

import (
	"bluelink-backend/internal/models"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

type TransactionRepository struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

// Create 建立交易記錄
func (r *TransactionRepository) Create(ctx context.Context, tx *models.Transaction) error {
	query := `
		INSERT INTO transactions (
			tx_hash, event_type, bond_id, user_id, wallet_address,
			amount, quantity, price, status, block_number, timestamp, metadata, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, created_at
	`

	err := r.db.QueryRowContext(ctx, query,
		tx.TxHash,
		tx.EventType,
		tx.BondID,
		tx.UserID,
		tx.WalletAddress,
		tx.Amount,
		tx.Quantity,
		tx.Price,
		tx.Status,
		tx.BlockNumber,
		tx.Timestamp,
		tx.Metadata,
		time.Now(),
	).Scan(&tx.ID, &tx.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	return nil
}

// GetByTxHash 根據交易哈希查詢
func (r *TransactionRepository) GetByTxHash(ctx context.Context, txHash string) (*models.Transaction, error) {
	tx := &models.Transaction{}

	query := `
		SELECT id, tx_hash, event_type, bond_id, user_id, wallet_address,
		       amount, quantity, price, status, block_number, timestamp, metadata, created_at
		FROM transactions
		WHERE tx_hash = $1
	`

	err := r.db.QueryRowContext(ctx, query, txHash).Scan(
		&tx.ID,
		&tx.TxHash,
		&tx.EventType,
		&tx.BondID,
		&tx.UserID,
		&tx.WalletAddress,
		&tx.Amount,
		&tx.Quantity,
		&tx.Price,
		&tx.Status,
		&tx.BlockNumber,
		&tx.Timestamp,
		&tx.Metadata,
		&tx.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction by hash: %w", err)
	}

	return tx, nil
}

// ListByUser 查詢使用者的交易記錄
func (r *TransactionRepository) ListByUser(ctx context.Context, userID int64, limit, offset int) ([]*models.Transaction, error) {
	query := `
		SELECT id, tx_hash, event_type, bond_id, user_id, wallet_address,
		       amount, quantity, price, status, block_number, timestamp, metadata, created_at
		FROM transactions
		WHERE user_id = $1
		ORDER BY timestamp DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list user transactions: %w", err)
	}
	defer rows.Close()

	return r.scanTransactions(rows)
}

// ListByBond 查詢債券的交易記錄
func (r *TransactionRepository) ListByBond(ctx context.Context, bondID int64, limit, offset int) ([]*models.Transaction, error) {
	query := `
		SELECT id, tx_hash, event_type, bond_id, user_id, wallet_address,
		       amount, quantity, price, status, block_number, timestamp, metadata, created_at
		FROM transactions
		WHERE bond_id = $1
		ORDER BY timestamp DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, bondID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list bond transactions: %w", err)
	}
	defer rows.Close()

	return r.scanTransactions(rows)
}

// UpdateStatus 更新交易狀態
func (r *TransactionRepository) UpdateStatus(ctx context.Context, txHash, status string) error {
	query := `
		UPDATE transactions
		SET status = $1
		WHERE tx_hash = $2
	`

	result, err := r.db.ExecContext(ctx, query, status, txHash)
	if err != nil {
		return fmt.Errorf("failed to update transaction status: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("transaction not found")
	}

	return nil
}

// scanTransactions 掃描交易列表
func (r *TransactionRepository) scanTransactions(rows *sql.Rows) ([]*models.Transaction, error) {
	var transactions []*models.Transaction

	for rows.Next() {
		tx := &models.Transaction{}
		err := rows.Scan(
			&tx.ID,
			&tx.TxHash,
			&tx.EventType,
			&tx.BondID,
			&tx.UserID,
			&tx.WalletAddress,
			&tx.Amount,
			&tx.Quantity,
			&tx.Price,
			&tx.Status,
			&tx.BlockNumber,
			&tx.Timestamp,
			&tx.Metadata,
			&tx.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan transaction: %w", err)
		}
		transactions = append(transactions, tx)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return transactions, nil
}

// GetUserBonds 查詢使用者持倉
func (r *TransactionRepository) GetUserBonds(ctx context.Context, userID int64) ([]*models.UserBondWithDetails, error) {
	query := `
		SELECT 
			ub.id, ub.user_id, ub.bond_id, ub.wallet_address, 
			ub.quantity, ub.average_purchase_price, ub.total_interest_earned,
			ub.created_at, ub.updated_at,
			b.bond_name, b.issuer_name, b.annual_interest_rate, 
			b.maturity_date, b.active, b.redeemable
		FROM user_bonds ub
		JOIN bonds b ON ub.bond_id = b.id
		WHERE ub.user_id = $1 AND ub.quantity > 0 AND b.deleted_at IS NULL
		ORDER BY ub.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user bonds: %w", err)
	}
	defer rows.Close()

	var userBonds []*models.UserBondWithDetails
	for rows.Next() {
		ub := &models.UserBondWithDetails{}

		err := rows.Scan(
			&ub.ID,
			&ub.UserID,
			&ub.BondID,
			&ub.WalletAddress,
			&ub.Quantity,
			&ub.AveragePurchasePrice,
			&ub.TotalInterestEarned,
			&ub.CreatedAt,
			&ub.UpdatedAt,
			&ub.BondName,
			&ub.IssuerName,
			&ub.AnnualInterestRate,
			&ub.MaturityDate,
			&ub.Active,
			&ub.Redeemable,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user bond: %w", err)
		}

		userBonds = append(userBonds, ub)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return userBonds, nil
}

// UpdateUserBond 更新使用者持倉（購買或贖回）
func (r *TransactionRepository) UpdateUserBond(ctx context.Context, userID, bondID int64, quantityChange int64, price float64) error {
	// 使用 UPSERT 來處理持倉更新
	query := `
		INSERT INTO user_bonds (user_id, bond_id, wallet_address, quantity, average_purchase_price, created_at, updated_at)
		SELECT $1, $2, u.wallet_address, $3, $4, $5, $6
		FROM users u WHERE u.id = $1
		ON CONFLICT (user_id, bond_id)
		DO UPDATE SET
			quantity = user_bonds.quantity + $3,
			average_purchase_price = 
				CASE 
					WHEN $3 > 0 THEN -- 購買
						((user_bonds.quantity * COALESCE(user_bonds.average_purchase_price, 0)) + ($3 * $4)) / (user_bonds.quantity + $3)
					ELSE -- 贖回
						user_bonds.average_purchase_price
				END,
			updated_at = $6
	`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query,
		userID,
		bondID,
		quantityChange,
		price,
		now,
		now,
	)

	if err != nil {
		return fmt.Errorf("failed to update user bond: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("failed to update user bond: no rows affected")
	}

	return nil
}

// CreateTransactionWithUserBond 創建交易並更新持倉（事務處理）
func (r *TransactionRepository) CreateTransactionWithUserBond(
	ctx context.Context,
	tx *models.Transaction,
	quantityChange int64,
	price float64,
) error {
	// 開始事務
	dbTx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer dbTx.Rollback()

	// 1. 創建交易記錄
	query := `
		INSERT INTO transactions (
			tx_hash, event_type, bond_id, user_id, wallet_address,
			amount, quantity, price, status, block_number, timestamp, metadata, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, created_at
	`

	err = dbTx.QueryRowContext(ctx, query,
		tx.TxHash,
		tx.EventType,
		tx.BondID,
		tx.UserID,
		tx.WalletAddress,
		tx.Amount,
		tx.Quantity,
		tx.Price,
		tx.Status,
		tx.BlockNumber,
		tx.Timestamp,
		tx.Metadata,
		time.Now(),
	).Scan(&tx.ID, &tx.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create transaction: %w", err)
	}

	// 2. 更新持倉（如果是購買或贖回事件）
	if tx.EventType == models.EventBondPurchased || tx.EventType == models.EventBondRedeemed {
		if tx.UserID == nil || tx.BondID == nil {
			return fmt.Errorf("user_id and bond_id are required for bond purchase/redeem")
		}

		updateQuery := `
			INSERT INTO user_bonds (user_id, bond_id, wallet_address, quantity, average_purchase_price, created_at, updated_at)
			SELECT $1, $2, u.wallet_address, $3, $4, $5, $6
			FROM users u WHERE u.id = $1
			ON CONFLICT (user_id, bond_id)
			DO UPDATE SET
				quantity = user_bonds.quantity + $3,
				average_purchase_price = 
					CASE 
						WHEN $3 > 0 THEN
							((user_bonds.quantity * COALESCE(user_bonds.average_purchase_price, 0)) + ($3 * $4)) / (user_bonds.quantity + $3)
						ELSE
							user_bonds.average_purchase_price
					END,
				updated_at = $6
		`

		now := time.Now()
		_, err = dbTx.ExecContext(ctx, updateQuery,
			*tx.UserID,
			*tx.BondID,
			quantityChange,
			price,
			now,
			now,
		)

		if err != nil {
			return fmt.Errorf("failed to update user bond: %w", err)
		}
	}

	// 提交事務
	if err := dbTx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// MetadataToJSON 將 metadata 轉換為 JSON 字串
func MetadataToJSON(metadata interface{}) (*string, error) {
	if metadata == nil {
		return nil, nil
	}

	bytes, err := json.Marshal(metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	str := string(bytes)
	return &str, nil
}
