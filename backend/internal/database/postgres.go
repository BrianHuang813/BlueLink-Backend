package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

type PostgresDB struct {
	DB *sql.DB
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// NewPostgresDB 建立 PostgreSQL 連接
func NewPostgresDB(cfg DBConfig) (*PostgresDB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// 設定連接池
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(10 * time.Minute)

	// 測試連接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresDB{DB: db}, nil
}

// Close 關閉資料庫連接
func (p *PostgresDB) Close() error {
	return p.DB.Close()
}

// HealthCheck 健康檢查
func (p *PostgresDB) HealthCheck(ctx context.Context) error {
	return p.DB.PingContext(ctx)
}

// BeginTx 開始交易
func (p *PostgresDB) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return p.DB.BeginTx(ctx, nil)
}
