package repository

import (
	"bluelink-backend/internal/models"
	"context"
	"database/sql"
	"fmt"
	"time"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create 建立新使用者
func (r *UserRepository) Create(ctx context.Context, walletAddress string) (*models.User, error) {
	user := &models.User{
		WalletAddress: walletAddress,
		Role:          "buyer",
		Timezone:      "UTC",
		Language:      "en",
	}

	query := `
		INSERT INTO users (wallet_address, role, timezone, language, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`

	now := time.Now()
	err := r.db.QueryRowContext(ctx, query,
		user.WalletAddress,
		user.Role,
		user.Timezone,
		user.Language,
		now,
		now,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// GetByID 根據 ID 查詢使用者
func (r *UserRepository) GetByID(ctx context.Context, id int64) (*models.User, error) {
	user := &models.User{}

	query := `
		SELECT id, wallet_address, role, institution_name, name, timezone, language, 
		       created_at, updated_at, deleted_at
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.WalletAddress,
		&user.Role,
		&user.InstitutionName,
		&user.Name,
		&user.Timezone,
		&user.Language,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return user, nil
}

// GetByWalletAddress 根據錢包地址查詢使用者
func (r *UserRepository) GetByWalletAddress(ctx context.Context, walletAddress string) (*models.User, error) {
	user := &models.User{}

	query := `
		SELECT id, wallet_address, role, institution_name, name, timezone, language, 
		       created_at, updated_at, deleted_at
		FROM users
		WHERE wallet_address = $1 AND deleted_at IS NULL
	`

	err := r.db.QueryRowContext(ctx, query, walletAddress).Scan(
		&user.ID,
		&user.WalletAddress,
		&user.Role,
		&user.InstitutionName,
		&user.Name,
		&user.Timezone,
		&user.Language,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by wallet address: %w", err)
	}

	return user, nil
}

// Update 更新使用者資料
func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	query := `
		UPDATE users
		SET name = $1, institution_name = $2, timezone = $3, language = $4, updated_at = $5
		WHERE id = $6 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query,
		user.Name,
		user.InstitutionName,
		user.Timezone,
		user.Language,
		time.Now(),
		user.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("user not found or already deleted")
	}

	return nil
}

// Delete 軟刪除使用者
func (r *UserRepository) Delete(ctx context.Context, id int64) error {
	query := `
		UPDATE users
		SET deleted_at = $1
		WHERE id = $2 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("user not found or already deleted")
	}

	return nil
}

// List 查詢所有使用者（分頁）
func (r *UserRepository) List(ctx context.Context, limit, offset int) ([]*models.User, error) {
	query := `
		SELECT id, wallet_address, role, institution_name, name, timezone, language, 
		       created_at, updated_at, deleted_at
		FROM users
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		err := rows.Scan(
			&user.ID,
			&user.WalletAddress,
			&user.Role,
			&user.InstitutionName,
			&user.Name,
			&user.Timezone,
			&user.Language,
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return users, nil
}
