package migrate

import (
	"bluelink-backend/internal/config"
	"bluelink-backend/internal/database"
	"context"
	"fmt"
	"log"
	"os"
)

func main() {
	// 載入配置
	cfg := config.LoadConfig()

	// 連接資料庫
	dbConfig := database.DBConfig{
		Host:     cfg.DBHost,
		Port:     cfg.DBPort,
		User:     cfg.DBUser,
		Password: cfg.DBPassword,
		DBName:   cfg.DBName,
		SSLMode:  cfg.DBSSLMode,
	}

	db, err := database.NewPostgresDB(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("✅ Database connection established")

	ctx := context.Background()

	// 檢查命令行參數
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	command := os.Args[1]

	switch command {
	case "migrate":
		// 執行遷移
		log.Println("🔄 Running database migrations...")
		if err := db.Migrate(ctx); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
		log.Println("✅ Migrations completed successfully!")

	case "rollback":
		// 回滾最後一次遷移
		log.Println("🔄 Rolling back last migration...")
		if err := db.Rollback(ctx); err != nil {
			log.Fatalf("Rollback failed: %v", err)
		}
		log.Println("✅ Rollback completed successfully!")

	case "status":
		// 查看遷移狀態
		log.Println("📊 Checking migration status...")
		printMigrationStatus(ctx, db)

	case "reset":
		// 重置資料庫（危險操作！）
		log.Println("⚠️  WARNING: This will drop all tables!")
		fmt.Print("Type 'yes' to confirm: ")
		var confirm string
		fmt.Scanln(&confirm)

		if confirm == "yes" {
			if err := resetDatabase(ctx, db); err != nil {
				log.Fatalf("Reset failed: %v", err)
			}
			log.Println("✅ Database reset completed!")
		} else {
			log.Println("❌ Reset cancelled")
		}

	default:
		log.Printf("Unknown command: %s\n", command)
		printUsage()
	}
}

func printUsage() {
	fmt.Println(`
Database Migration Tool

Usage:
  go run cmd/migrate/main.go <command>

Commands:
  migrate   - Run all pending migrations
  rollback  - Rollback the last migration
  status    - Show migration status
  reset     - Drop all tables and reset database (DANGEROUS!)

Examples:
  go run cmd/migrate/main.go migrate
  go run cmd/migrate/main.go rollback
  go run cmd/migrate/main.go status`)
}

func printMigrationStatus(ctx context.Context, db *database.PostgresDB) {
	rows, err := db.DB.QueryContext(ctx, `
		SELECT version, description, applied_at 
		FROM schema_migrations 
		ORDER BY version
	`)
	if err != nil {
		log.Printf("Failed to get migration status: %v", err)
		return
	}
	defer rows.Close()

	fmt.Println("\n Applied Migrations:")
	fmt.Println("╔═══════╦═══════════════════════════════════════╦═════════════════════╗")
	fmt.Println("║Version║ Description                           ║ Applied At          ║")
	fmt.Println("╠═══════╬═══════════════════════════════════════╬═════════════════════╣")

	for rows.Next() {
		var version int
		var description string
		var appliedAt string
		if err := rows.Scan(&version, &description, &appliedAt); err != nil {
			log.Printf("Failed to scan row: %v", err)
			continue
		}
		fmt.Printf("║ %-5d ║ %-37s ║ %-19s ║\n", version, description, appliedAt[:19])
	}

	fmt.Println("╚═══════╩═══════════════════════════════════════╩═════════════════════╝")
}

func resetDatabase(ctx context.Context, db *database.PostgresDB) error {
	// 刪除所有表
	tables := []string{
		"user_bonds",
		"transactions",
		"sessions",
		"bonds",
		"users",
		"schema_migrations",
	}

	for _, table := range tables {
		_, err := db.DB.ExecContext(ctx, fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", table))
		if err != nil {
			return fmt.Errorf("failed to drop table %s: %w", table, err)
		}
		log.Printf("✓ Dropped table: %s", table)
	}

	return nil
}
