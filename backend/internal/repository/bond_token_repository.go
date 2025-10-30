package repository

import (
	"bluelink-backend/internal/models"
	"context"
	"database/sql"
	"fmt"
	"time"
)

type BondTokenRepository struct {
	db *sql.DB
}

func NewBondTokenRepository(db *sql.DB) *BondTokenRepository {
	return &BondTokenRepository{db: db}
}

// Create 建立新債券代幣
func (r *BondTokenRepository) Create(ctx context.Context, token *models.BondToken) error {
	query := `
		INSERT INTO bond_tokens (
			on_chain_id, project_id,
			bond_name, token_image_url, maturity_date, annual_interest_rate,
			token_number, owner, amount, purchase_date, is_redeemed,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, created_at, updated_at
	`

	now := time.Now()

	err := r.db.QueryRowContext(ctx, query,
		token.OnChainID,
		token.ProjectID,
		token.BondName,
		token.TokenImageUrl,
		token.MaturityDate,
		token.AnnualInterestRate,
		token.TokenNumber,
		token.Owner,
		token.Amount,
		token.PurchaseDate,
		token.IsRedeemed,
		now,
		now,
	).Scan(&token.ID, &token.CreatedAt, &token.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create bond token: %w", err)
	}

	return nil
}

// GetByID 根據 ID 查詢債券代幣
func (r *BondTokenRepository) GetByID(ctx context.Context, id int64) (*models.BondToken, error) {
	token := &models.BondToken{}

	query := `
		SELECT id, on_chain_id, project_id,
		       bond_name, token_image_url, maturity_date, annual_interest_rate,
		       token_number, owner, amount, purchase_date, is_redeemed,
		       created_at, updated_at, deleted_at
		FROM bond_tokens
		WHERE id = $1 AND deleted_at IS NULL
	`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&token.ID,
		&token.OnChainID,
		&token.ProjectID,
		&token.BondName,
		&token.TokenImageUrl,
		&token.MaturityDate,
		&token.AnnualInterestRate,
		&token.TokenNumber,
		&token.Owner,
		&token.Amount,
		&token.PurchaseDate,
		&token.IsRedeemed,
		&token.CreatedAt,
		&token.UpdatedAt,
		&token.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get bond token by ID: %w", err)
	}

	return token, nil
}

// GetByOnChainID 根據鏈上 ID 查詢債券代幣
func (r *BondTokenRepository) GetByOnChainID(ctx context.Context, onChainID string) (*models.BondToken, error) {
	token := &models.BondToken{}

	query := `
		SELECT id, on_chain_id, project_id,
		       bond_name, token_image_url, maturity_date, annual_interest_rate,
		       token_number, owner, amount, purchase_date, is_redeemed,
		       created_at, updated_at, deleted_at
		FROM bond_tokens
		WHERE on_chain_id = $1 AND deleted_at IS NULL
	`

	err := r.db.QueryRowContext(ctx, query, onChainID).Scan(
		&token.ID,
		&token.OnChainID,
		&token.ProjectID,
		&token.BondName,
		&token.TokenImageUrl,
		&token.MaturityDate,
		&token.AnnualInterestRate,
		&token.TokenNumber,
		&token.Owner,
		&token.Amount,
		&token.PurchaseDate,
		&token.IsRedeemed,
		&token.CreatedAt,
		&token.UpdatedAt,
		&token.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get bond token by on-chain ID: %w", err)
	}

	return token, nil
}

// GetByOwner 根據擁有者地址查詢債券代幣
func (r *BondTokenRepository) GetByOwner(ctx context.Context, owner string, limit, offset int) ([]*models.BondToken, error) {
	query := `
		SELECT id, on_chain_id, project_id,
		       bond_name, token_image_url, maturity_date, annual_interest_rate,
		       token_number, owner, amount, purchase_date, is_redeemed,
		       created_at, updated_at, deleted_at
		FROM bond_tokens
		WHERE owner = $1 AND deleted_at IS NULL
		ORDER BY purchase_date DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, owner, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list bond tokens by owner: %w", err)
	}
	defer rows.Close()

	var tokens []*models.BondToken
	for rows.Next() {
		token := &models.BondToken{}

		err := rows.Scan(
			&token.ID,
			&token.OnChainID,
			&token.ProjectID,
			&token.BondName,
			&token.TokenImageUrl,
			&token.MaturityDate,
			&token.AnnualInterestRate,
			&token.TokenNumber,
			&token.Owner,
			&token.Amount,
			&token.PurchaseDate,
			&token.IsRedeemed,
			&token.CreatedAt,
			&token.UpdatedAt,
			&token.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan bond token: %w", err)
		}

		tokens = append(tokens, token)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return tokens, nil
}

// GetByProjectID 根據專案 ID 查詢債券代幣
func (r *BondTokenRepository) GetByProjectID(ctx context.Context, projectID string, limit, offset int) ([]*models.BondToken, error) {
	query := `
		SELECT id, on_chain_id, project_id,
		       bond_name, token_image_url, maturity_date, annual_interest_rate,
		       token_number, owner, amount, purchase_date, is_redeemed,
		       created_at, updated_at, deleted_at
		FROM bond_tokens
		WHERE project_id = $1 AND deleted_at IS NULL
		ORDER BY token_number ASC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, projectID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list bond tokens by project: %w", err)
	}
	defer rows.Close()

	var tokens []*models.BondToken
	for rows.Next() {
		token := &models.BondToken{}

		err := rows.Scan(
			&token.ID,
			&token.OnChainID,
			&token.ProjectID,
			&token.BondName,
			&token.TokenImageUrl,
			&token.MaturityDate,
			&token.AnnualInterestRate,
			&token.TokenNumber,
			&token.Owner,
			&token.Amount,
			&token.PurchaseDate,
			&token.IsRedeemed,
			&token.CreatedAt,
			&token.UpdatedAt,
			&token.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan bond token: %w", err)
		}

		tokens = append(tokens, token)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return tokens, nil
}

// UpdateRedeemed 更新代幣贖回狀態
func (r *BondTokenRepository) UpdateRedeemed(ctx context.Context, id int64, isRedeemed bool) error {
	query := `
		UPDATE bond_tokens
		SET is_redeemed = $1, updated_at = $2
		WHERE id = $3 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, isRedeemed, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update bond token redeemed status: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("bond token not found or already deleted")
	}

	return nil
}

// Delete 軟刪除債券代幣
func (r *BondTokenRepository) Delete(ctx context.Context, id int64) error {
	query := `
		UPDATE bond_tokens
		SET deleted_at = $1
		WHERE id = $2 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to delete bond token: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("bond token not found or already deleted")
	}

	return nil
}
