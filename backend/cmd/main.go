package main

import (
	"bluelink-backend/internal/blockchain"
	"bluelink-backend/internal/config"
	"bluelink-backend/internal/database"
	"bluelink-backend/internal/middleware"
	"bluelink-backend/internal/repository"
	"bluelink-backend/internal/routes"
	"bluelink-backend/internal/services"
	"bluelink-backend/internal/session"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/block-vision/sui-go-sdk/sui"
	"github.com/gin-gonic/gin"
)

func main() {

	fmt.Println("Starting BlueLink Backend Application...")

	// 1. 載入配置
	cfg := config.LoadConfig()

	// 2. 設定 Gin 模式
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 3. 連接資料庫
	log.Println("Connecting to database...")
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
	log.Println("Database connected")

	// 4. 執行資料庫遷移
	ctx := context.Background()
	log.Println("Running database migrations...")
	if err := db.Migrate(ctx); err != nil {
		log.Fatalf("Migration failed: %v", err)
	} else {
		log.Println("Migrations completed successfully")
	}

	// 5. 初始化 Repositories
	userRepo := repository.NewUserRepository(db.DB)
	bondRepo := repository.NewBondRepository(db.DB)
	bondTokenRepo := repository.NewBondTokenRepository(db.DB)
	txRepo := repository.NewTransactionRepository(db.DB)
	sessionRepo := repository.NewSessionRepository(db.DB)
	nonceRepo := repository.NewNonceRepository(db.DB)

	// 6. 初始化 Services
	userService := services.NewUserService(userRepo)
	bondService := services.NewBondService(bondRepo)
	bondTokenService := services.NewBondTokenService(bondTokenRepo)

	// 7. 初始化 Session Manager（使用 PostgreSQL）
	sessionManager := session.NewPostgresSessionManager(sessionRepo)
	log.Println("✅ Using PostgreSQL Session Manager (persistent sessions)")

	// 8. 初始化 Sui Client
	log.Println("Initializing Sui client...")
	suiClient := sui.NewSuiClient(cfg.SuiRPCURL)
	log.Println("Sui client initialized")

	// 9. 初始化並啟動區塊鏈事件監聽器
	if cfg.SuiPackageID != "" {
		log.Println("Starting blockchain event listener...")
		eventListener := blockchain.NewEventListener(
			suiClient,
			txRepo,
			bondRepo,
			userRepo,
			cfg.SuiPackageID,
		)

		if err := eventListener.Start(ctx); err != nil {
			log.Printf("Failed to start event listener: %v", err)
		} else {
			defer eventListener.Stop()
			log.Println("Event listener started")
		}
	} else {
		log.Println("SUI_PACKAGE_ID not set, skipping event listener")
	}

	// 10. 初始化 Gin Router
	r := gin.Default()

	// 11. 全域中間件
	r.Use(middleware.CORSMiddleware(cfg.CORSAllowedOrigins))
	r.Use(middleware.RecoveryMiddleware())
	r.Use(middleware.RequestIDMiddleware())
	r.Use(middleware.LoggingMiddleware())

	// 12. 設定路由
	routes.SetupRoutes(r, userService, bondService, bondTokenService, sessionManager, nonceRepo, cfg)

	// 13. 健康檢查路由
	r.GET("/health", func(c *gin.Context) {
		if err := db.HealthCheck(ctx); err != nil {
			c.JSON(500, gin.H{
				"status":   "unhealthy",
				"database": "disconnected",
				"error":    err.Error(),
			})
			return
		}

		c.JSON(200, gin.H{
			"status":   "healthy",
			"database": "connected",
			"sui":      "connected",
			"version":  "1.0.0",
		})
	})

	// 14. 建立 HTTP Server（使用原生 net/http 以支援優雅關閉）
	srv := &http.Server{
		Addr:           cfg.Port,
		Handler:        r,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	log.Printf("Server starting on port %s", cfg.Port)
	log.Printf("Environment: %s", cfg.Environment)
	log.Printf("Sui RPC: %s", cfg.SuiRPCURL)
	log.Printf("Database: %s@%s:%s/%s", cfg.DBUser, cfg.DBHost, cfg.DBPort, cfg.DBName)

	// 15. 在 goroutine 中啟動伺服器
	go func() {
		log.Printf("Server is ready and listening on %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 16. 等待中斷信號（優雅關閉）
	quit := make(chan os.Signal, 1)
	// 監聽 SIGINT (Ctrl+C) 和 SIGTERM (kill)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 阻塞直到收到信號
	sig := <-quit
	log.Printf("\nReceived signal: %v", sig)
	log.Println("Initiating graceful shutdown...")

	// 17. 建立關閉 context（5 秒超時）
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 18. 優雅地關閉伺服器
	// Shutdown 會：
	// 1. 停止接收新請求
	// 2. 等待現有請求完成（最多等待 5 秒）
	// 3. 關閉所有連線
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server gracefully stopped")
}
