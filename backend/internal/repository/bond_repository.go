package repository

import (
	"bluelink-backend/internal/models"
	"context"
	"database/sql"
	"fmt"
	"time"
)

type BondRepository struct {
	db *sql.DB
}

func NewBondRepository(db *sql.DB) *BondRepository {
	return &BondRepository{db: db}
}

// Create 建立新債券
func (r *BondRepository) Create(ctx context.Context, bond *models.Bond) error {
	query := `
		INSERT INTO bonds (
			on_chain_id, issuer_address, issuer_name, bond_name,
			total_amount, amount_raised, amount_redeemed,
			tokens_issued, tokens_redeemed,
			annual_interest_rate, maturity_date, issue_date,
			active, redeemable,
			raised_funds_balance, redemption_pool_balance,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
		RETURNING id, created_at, updated_at
	`

	now := time.Now()

	err := r.db.QueryRowContext(ctx, query,
		bond.OnChainID,
		bond.IssuerAddress,
		bond.IssuerName,
		bond.BondName,
		bond.TotalAmount,
		bond.AmountRaised,
		bond.AmountRedeemed,
		bond.TokensIssued,
		bond.TokensRedeemed,
		bond.AnnualInterestRate,
		bond.MaturityDate,
		bond.IssueDate,
		bond.Active,
		bond.Redeemable,
		bond.RaisedFundsBalance,
		bond.RedemptionPoolBalance,
		now,
		now,
	).Scan(&bond.ID, &bond.CreatedAt, &bond.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create bond: %w", err)
	}

	return nil
}

// GetByID 根據 ID 查詢債券
func (r *BondRepository) GetByID(ctx context.Context, id int64) (*models.Bond, error) {
	bond := &models.Bond{}

	query := `
		SELECT id, on_chain_id, issuer_address, issuer_name, bond_name,
		       total_amount, amount_raised, amount_redeemed,
		       tokens_issued, tokens_redeemed,
		       annual_interest_rate, maturity_date, issue_date,
		       active, redeemable,
		       raised_funds_balance, redemption_pool_balance,
		       created_at, updated_at, deleted_at
		FROM bonds
		WHERE id = $1 AND deleted_at IS NULL
	`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&bond.ID,
		&bond.OnChainID,
		&bond.IssuerAddress,
		&bond.IssuerName,
		&bond.BondName,
		&bond.TotalAmount,
		&bond.AmountRaised,
		&bond.AmountRedeemed,
		&bond.TokensIssued,
		&bond.TokensRedeemed,
		&bond.AnnualInterestRate,
		&bond.MaturityDate,
		&bond.IssueDate,
		&bond.Active,
		&bond.Redeemable,
		&bond.RaisedFundsBalance,
		&bond.RedemptionPoolBalance,
		&bond.CreatedAt,
		&bond.UpdatedAt,
		&bond.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get bond by ID: %w", err)
	}

	return bond, nil
}

// GetByOnChainID 根據鏈上 ID 查詢債券
func (r *BondRepository) GetByOnChainID(ctx context.Context, onChainID string) (*models.Bond, error) {
	bond := &models.Bond{}

	query := `
		SELECT id, on_chain_id, issuer_address, issuer_name, bond_name,
		       total_amount, amount_raised, amount_redeemed,
		       tokens_issued, tokens_redeemed,
		       annual_interest_rate, maturity_date, issue_date,
		       active, redeemable,
		       raised_funds_balance, redemption_pool_balance,
		       created_at, updated_at, deleted_at
		FROM bonds
		WHERE on_chain_id = $1 AND deleted_at IS NULL
	`

	err := r.db.QueryRowContext(ctx, query, onChainID).Scan(
		&bond.ID,
		&bond.OnChainID,
		&bond.IssuerAddress,
		&bond.IssuerName,
		&bond.BondName,
		&bond.TotalAmount,
		&bond.AmountRaised,
		&bond.AmountRedeemed,
		&bond.TokensIssued,
		&bond.TokensRedeemed,
		&bond.AnnualInterestRate,
		&bond.MaturityDate,
		&bond.IssueDate,
		&bond.Active,
		&bond.Redeemable,
		&bond.RaisedFundsBalance,
		&bond.RedemptionPoolBalance,
		&bond.CreatedAt,
		&bond.UpdatedAt,
		&bond.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get bond by on-chain ID: %w", err)
	}

	return bond, nil
}

// List 查詢所有債券（分頁）
func (r *BondRepository) List(ctx context.Context, limit, offset int) ([]*models.Bond, error) {
	query := `
		SELECT id, on_chain_id, issuer_address, issuer_name, bond_name,
		       total_amount, amount_raised, amount_redeemed,
		       tokens_issued, tokens_redeemed,
		       annual_interest_rate, maturity_date, issue_date,
		       active, redeemable,
		       raised_funds_balance, redemption_pool_balance,
		       created_at, updated_at, deleted_at
		FROM bonds
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list bonds: %w", err)
	}
	defer rows.Close()

	var bonds []*models.Bond
	for rows.Next() {
		bond := &models.Bond{}

		err := rows.Scan(
			&bond.ID,
			&bond.OnChainID,
			&bond.IssuerAddress,
			&bond.IssuerName,
			&bond.BondName,
			&bond.TotalAmount,
			&bond.AmountRaised,
			&bond.AmountRedeemed,
			&bond.TokensIssued,
			&bond.TokensRedeemed,
			&bond.AnnualInterestRate,
			&bond.MaturityDate,
			&bond.IssueDate,
			&bond.Active,
			&bond.Redeemable,
			&bond.RaisedFundsBalance,
			&bond.RedemptionPoolBalance,
			&bond.CreatedAt,
			&bond.UpdatedAt,
			&bond.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan bond: %w", err)
		}

		bonds = append(bonds, bond)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return bonds, nil
}

// UpdateStatus 更新債券狀態（active/redeemable）
func (r *BondRepository) UpdateStatus(ctx context.Context, id int64, active, redeemable bool) error {
	query := `
		UPDATE bonds
		SET active = $1, redeemable = $2, updated_at = $3
		WHERE id = $4 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, active, redeemable, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update bond status: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("bond not found or already deleted")
	}

	return nil
}

// Delete 軟刪除債券
func (r *BondRepository) Delete(ctx context.Context, id int64) error {
	query := `
		UPDATE bonds
		SET deleted_at = $1
		WHERE id = $2 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to delete bond: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("bond not found or already deleted")
	}

	return nil
}
