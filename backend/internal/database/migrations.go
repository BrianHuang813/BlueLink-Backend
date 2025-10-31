package database

import (
	"context"
	"database/sql"
	"fmt"
)

// Migration 資料庫遷移結構
type Migration struct {
	Version     int
	Description string
	Up          string
	Down        string
}

// GetMigrations 返回所有遷移腳本
func GetMigrations() []Migration {
	return []Migration{
		{
			Version:     1,
			Description: "Create users table",
			Up: `
				CREATE TABLE IF NOT EXISTS users (
					id BIGSERIAL PRIMARY KEY,
					wallet_address VARCHAR(66) UNIQUE NOT NULL,
					role VARCHAR(20) NOT NULL DEFAULT 'buyer',
					institution_name VARCHAR(255),
					name VARCHAR(255),
					timezone VARCHAR(50) DEFAULT 'UTC',
					language VARCHAR(10) DEFAULT 'en',
					created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
					updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
					deleted_at TIMESTAMP,
					CONSTRAINT chk_role CHECK (role IN ('buyer', 'issuer', 'admin'))
				);

				CREATE INDEX idx_users_wallet_address ON users(wallet_address);
				CREATE INDEX idx_users_role ON users(role);
				CREATE INDEX idx_users_deleted_at ON users(deleted_at);
			`,
			Down: `DROP TABLE IF EXISTS users;`,
		},
		{
			Version:     2,
			Description: "Create sessions table",
			Up: `
				CREATE TABLE IF NOT EXISTS sessions (
					id VARCHAR(36) PRIMARY KEY,
					user_id BIGINT NOT NULL,
					wallet_address VARCHAR(66) NOT NULL,
					role VARCHAR(20) NOT NULL,
					ip_address VARCHAR(45),
					user_agent TEXT,
					created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
					last_active_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
					expires_at TIMESTAMP NOT NULL,
					FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
				);

				CREATE INDEX idx_sessions_user_id ON sessions(user_id);
				CREATE INDEX idx_sessions_wallet_address ON sessions(wallet_address);
				CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);
			`,
			Down: `DROP TABLE IF EXISTS sessions;`,
		},
		{
			Version:     3,
			Description: "Create bonds table",
			Up: `
				CREATE TABLE IF NOT EXISTS bonds (
					id BIGSERIAL PRIMARY KEY,
					on_chain_id VARCHAR(66) UNIQUE NOT NULL,
					
					-- 發行者資訊
					issuer_address VARCHAR(66) NOT NULL,
					issuer_name VARCHAR(255) NOT NULL,
					bond_name VARCHAR(255) NOT NULL,
					
					-- 金額相關（使用 BIGINT 對應合約的 u64，單位：MIST）
					total_amount BIGINT NOT NULL DEFAULT 0,
					amount_raised BIGINT NOT NULL DEFAULT 0,
					amount_redeemed BIGINT NOT NULL DEFAULT 0,
					
					-- 代幣相關
					tokens_issued BIGINT NOT NULL DEFAULT 0,
					tokens_redeemed BIGINT NOT NULL DEFAULT 0,
					
					-- 利率和日期（利率使用 DECIMAL，日期使用 VARCHAR 儲存）
					annual_interest_rate BIGINT NOT NULL,
					maturity_date VARCHAR(10) NOT NULL,
					issue_date VARCHAR(10) NOT NULL,
					
					-- 狀態
					active BOOLEAN NOT NULL DEFAULT true,
					redeemable BOOLEAN NOT NULL DEFAULT false,
					
					-- 資金池餘額快照（使用 BIGINT，單位：MIST）
					raised_funds_balance BIGINT NOT NULL DEFAULT 0,
					redemption_pool_balance BIGINT NOT NULL DEFAULT 0,
					
					-- 資料庫管理欄位
					created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
					updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
					deleted_at TIMESTAMP
				);

				CREATE INDEX idx_bonds_on_chain_id ON bonds(on_chain_id);
				CREATE INDEX idx_bonds_issuer_address ON bonds(issuer_address);
				CREATE INDEX idx_bonds_active ON bonds(active);
				CREATE INDEX idx_bonds_maturity_date ON bonds(maturity_date);
				CREATE INDEX idx_bonds_deleted_at ON bonds(deleted_at);
			`,
			Down: `DROP TABLE IF EXISTS bonds;`,
		},
		{
			Version:     4,
			Description: "Create transactions table",
			Up: `
				CREATE TABLE IF NOT EXISTS transactions (
					id BIGSERIAL PRIMARY KEY,
					tx_hash VARCHAR(66) UNIQUE NOT NULL,
					event_type VARCHAR(50) NOT NULL,
					bond_id BIGINT,
					user_id BIGINT,
					wallet_address VARCHAR(66) NOT NULL,
					amount DECIMAL(20, 2),
					quantity BIGINT,
					price DECIMAL(20, 2),
					status VARCHAR(20) NOT NULL DEFAULT 'pending',
					block_number BIGINT,
					timestamp TIMESTAMP NOT NULL,
					metadata JSONB,
					created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
					FOREIGN KEY (bond_id) REFERENCES bonds(id) ON DELETE SET NULL,
					FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL,
					CONSTRAINT chk_event_type CHECK (event_type IN ('bond_created', 'bond_purchased', 'bond_redeemed', 'interest_paid', 'bond_transferred')),
					CONSTRAINT chk_tx_status CHECK (status IN ('pending', 'confirmed', 'failed'))
				);

				CREATE INDEX idx_transactions_tx_hash ON transactions(tx_hash);
				CREATE INDEX idx_transactions_event_type ON transactions(event_type);
				CREATE INDEX idx_transactions_bond_id ON transactions(bond_id);
				CREATE INDEX idx_transactions_user_id ON transactions(user_id);
				CREATE INDEX idx_transactions_wallet_address ON transactions(wallet_address);
				CREATE INDEX idx_transactions_timestamp ON transactions(timestamp);
				CREATE INDEX idx_transactions_status ON transactions(status);
			`,
			Down: `DROP TABLE IF EXISTS transactions;`,
		},
		{
			Version:     5,
			Description: "Create user_bonds table (持倉記錄)",
			Up: `
				CREATE TABLE IF NOT EXISTS user_bonds (
					id BIGSERIAL PRIMARY KEY,
					user_id BIGINT NOT NULL,
					bond_id BIGINT NOT NULL,
					wallet_address VARCHAR(66) NOT NULL,
					quantity BIGINT NOT NULL DEFAULT 0,
					average_purchase_price DECIMAL(20, 2),
					total_interest_earned DECIMAL(20, 2) DEFAULT 0,
					created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
					updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
					FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
					FOREIGN KEY (bond_id) REFERENCES bonds(id) ON DELETE CASCADE,
					UNIQUE (user_id, bond_id)
				);

				CREATE INDEX idx_user_bonds_user_id ON user_bonds(user_id);
				CREATE INDEX idx_user_bonds_bond_id ON user_bonds(bond_id);
				CREATE INDEX idx_user_bonds_wallet_address ON user_bonds(wallet_address);
			`,
			Down: `DROP TABLE IF EXISTS user_bonds;`,
		},
		{
			Version:     6,
			Description: "Create schema_migrations table",
			Up: `
				CREATE TABLE IF NOT EXISTS schema_migrations (
					version INT PRIMARY KEY,
					description VARCHAR(255) NOT NULL,
					applied_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
				);
			`,
			Down: `DROP TABLE IF EXISTS schema_migrations;`,
		},
		{
			Version:     7,
			Description: "Add image and metadata URLs to bonds table",
			Up: `
				ALTER TABLE bonds
				ADD COLUMN IF NOT EXISTS bond_image_url VARCHAR(500),
				ADD COLUMN IF NOT EXISTS token_image_url VARCHAR(500),
				ADD COLUMN IF NOT EXISTS metadata_url VARCHAR(500);

				CREATE INDEX IF NOT EXISTS idx_bonds_bond_image_url ON bonds(bond_image_url);
				CREATE INDEX IF NOT EXISTS idx_bonds_token_image_url ON bonds(token_image_url);
			`,
			Down: `
				DROP INDEX IF EXISTS idx_bonds_bond_image_url;
				DROP INDEX IF EXISTS idx_bonds_token_image_url;
				ALTER TABLE bonds
				DROP COLUMN IF EXISTS bond_image_url,
				DROP COLUMN IF EXISTS token_image_url,
				DROP COLUMN IF EXISTS metadata_url;
			`,
		},
		{
			Version:     8,
			Description: "Create bond_tokens table",
			Up: `
				CREATE TABLE IF NOT EXISTS bond_tokens (
					id BIGSERIAL PRIMARY KEY,
					on_chain_id VARCHAR(66) UNIQUE NOT NULL,
					
					-- 專案關聯
					project_id VARCHAR(66) NOT NULL,
					
					-- 🆕 代幣自包含資訊（從 BondProject 複製而來）
					bond_name VARCHAR(255) NOT NULL,
					token_image_url VARCHAR(500) NOT NULL,
					maturity_date BIGINT NOT NULL,
					annual_interest_rate BIGINT NOT NULL,
					
					-- 代幣資訊
					token_number BIGINT NOT NULL,
					owner VARCHAR(66) NOT NULL,
					amount BIGINT NOT NULL,
					purchase_date BIGINT NOT NULL,
					is_redeemed BOOLEAN NOT NULL DEFAULT false,
					
					-- 資料庫管理欄位
					created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
					updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
					deleted_at TIMESTAMP
				);

				CREATE INDEX IF NOT EXISTS idx_bond_tokens_on_chain_id ON bond_tokens(on_chain_id);
				CREATE INDEX IF NOT EXISTS idx_bond_tokens_project_id ON bond_tokens(project_id);
				CREATE INDEX IF NOT EXISTS idx_bond_tokens_owner ON bond_tokens(owner);
				CREATE INDEX IF NOT EXISTS idx_bond_tokens_token_number ON bond_tokens(token_number);
				CREATE INDEX IF NOT EXISTS idx_bond_tokens_maturity_date ON bond_tokens(maturity_date);
				CREATE INDEX IF NOT EXISTS idx_bond_tokens_is_redeemed ON bond_tokens(is_redeemed);
				CREATE INDEX IF NOT EXISTS idx_bond_tokens_deleted_at ON bond_tokens(deleted_at);
			`,
			Down: `DROP TABLE IF EXISTS bond_tokens;`,
		},
		{
			Version:     9,
			Description: "Create nonces table",
			Up: `
				CREATE TABLE IF NOT EXISTS nonces (
					id BIGSERIAL PRIMARY KEY,
					wallet_address VARCHAR(66) UNIQUE NOT NULL,
					nonce VARCHAR(255) NOT NULL,
					created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
					expires_at TIMESTAMP NOT NULL
				);

				CREATE INDEX IF NOT EXISTS idx_nonces_wallet_address ON nonces(wallet_address);
				CREATE INDEX IF NOT EXISTS idx_nonces_expires_at ON nonces(expires_at);
			`,
			Down: `DROP TABLE IF EXISTS nonces;`,
		},
	}
}

// Migrate 執行資料庫遷移
func (p *PostgresDB) Migrate(ctx context.Context) error {
	// 先建立 schema_migrations 表
	migrations := GetMigrations()

	// 找到 schema_migrations 的遷移並先執行
	for _, m := range migrations {
		if m.Version == 6 {
			if _, err := p.DB.ExecContext(ctx, m.Up); err != nil {
				return fmt.Errorf("failed to create schema_migrations table: %w", err)
			}
			break
		}
	}

	// 獲取已應用的遷移
	appliedVersions, err := p.getAppliedMigrations(ctx)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// 執行未應用的遷移
	for _, migration := range migrations {
		if migration.Version == 6 {
			continue // 已經執行過了
		}

		if appliedVersions[migration.Version] {
			continue // 已應用
		}

		fmt.Printf("Applying migration %d: %s\n", migration.Version, migration.Description)

		// 開始事務
		tx, err := p.DB.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}

		// 執行遷移
		if _, err := tx.ExecContext(ctx, migration.Up); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to apply migration %d: %w", migration.Version, err)
		}

		// 記錄遷移
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO schema_migrations (version, description) VALUES ($1, $2)`,
			migration.Version, migration.Description,
		); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to record migration %d: %w", migration.Version, err)
		}

		// 提交事務
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %d: %w", migration.Version, err)
		}

		fmt.Printf("✓ Migration %d applied successfully\n", migration.Version)
	}

	return nil
}

// getAppliedMigrations 獲取已應用的遷移版本
func (p *PostgresDB) getAppliedMigrations(ctx context.Context) (map[int]bool, error) {
	rows, err := p.DB.QueryContext(ctx, `SELECT version FROM schema_migrations`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[int]bool)
	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		applied[version] = true
	}

	return applied, rows.Err()
}

// Rollback 回滾最後一次遷移
func (p *PostgresDB) Rollback(ctx context.Context) error {
	// 獲取最後應用的遷移
	var lastVersion int
	err := p.DB.QueryRowContext(ctx,
		`SELECT version FROM schema_migrations ORDER BY version DESC LIMIT 1`,
	).Scan(&lastVersion)
	if err == sql.ErrNoRows {
		return fmt.Errorf("no migrations to rollback")
	}
	if err != nil {
		return fmt.Errorf("failed to get last migration: %w", err)
	}

	// 找到對應的遷移
	migrations := GetMigrations()
	var targetMigration *Migration
	for _, m := range migrations {
		if m.Version == lastVersion {
			targetMigration = &m
			break
		}
	}

	if targetMigration == nil {
		return fmt.Errorf("migration %d not found", lastVersion)
	}

	fmt.Printf("Rolling back migration %d: %s\n", lastVersion, targetMigration.Description)

	// 開始事務
	tx, err := p.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// 執行回滾
	if _, err := tx.ExecContext(ctx, targetMigration.Down); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to rollback migration %d: %w", lastVersion, err)
	}

	// 刪除遷移記錄
	if _, err := tx.ExecContext(ctx,
		`DELETE FROM schema_migrations WHERE version = $1`,
		lastVersion,
	); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete migration record %d: %w", lastVersion, err)
	}

	// 提交事務
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit rollback %d: %w", lastVersion, err)
	}

	fmt.Printf("✓ Migration %d rolled back successfully\n", lastVersion)
	return nil
}
