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
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/block-vision/sui-go-sdk/sui"
	"github.com/gin-gonic/gin"
)

func main() {
	// for local development initial logo
	logo := `
          _____                    _____            _____                    _____                            _____            _____                    _____                    _____          
         /\    \                  /\    \          /\    \                  /\    \                          /\    \          /\    \                  /\    \                  /\    \         
        /::\    \                /::\____\        /::\____\                /::\    \                        /::\____\        /::\    \                /::\____\                /::\____\        
       /::::\    \              /:::/    /       /:::/    /               /::::\    \                      /:::/    /        \:::\    \              /::::|   |               /:::/    /        
      /::::::\    \            /:::/    /       /:::/    /               /::::::\    \                    /:::/    /          \:::\    \            /:::::|   |              /:::/    /         
     /:::/\:::\    \          /:::/    /       /:::/    /               /:::/\:::\    \                  /:::/    /            \:::\    \          /::::::|   |             /:::/    /          
    /:::/__\:::\    \        /:::/    /       /:::/    /               /:::/__\:::\    \                /:::/    /              \:::\    \        /:::/|::|   |            /:::/____/           
   /::::\   \:::\    \      /:::/    /       /:::/    /               /::::\   \:::\    \              /:::/    /               /::::\    \      /:::/ |::|   |           /::::\    \           
  /::::::\   \:::\    \    /:::/    /       /:::/    /      _____    /::::::\   \:::\    \            /:::/    /       ____    /::::::\    \    /:::/  |::|   | _____    /::::::\____\________  
 /:::/\:::\   \:::\ ___\  /:::/    /       /:::/____/      /\    \  /:::/\:::\   \:::\    \          /:::/    /       /\   \  /:::/\:::\    \  /:::/   |::|   |/\    \  /:::/\:::::::::::\    \ 
/:::/__\:::\   \:::|    |/:::/____/       |:::|    /      /::\____\/:::/__\:::\   \:::\____\        /:::/____/       /::\   \/:::/  \:::\____\/:: /    |::|   /::\____\/:::/  |:::::::::::\____\
\:::\   \:::\  /:::|____|\:::\    \       |:::|____\     /:::/    /\:::\   \:::\   \::/    /        \:::\    \       \:::\  /:::/    \::/    /\::/    /|::|  /:::/    /\::/   |::|~~~|~~~~~     
 \:::\   \:::\/:::/    /  \:::\    \       \:::\    \   /:::/    /  \:::\   \:::\   \/____/          \:::\    \       \:::\/:::/    / \/____/  \/____/ |::| /:::/    /  \/____|::|   |          
  \:::\   \::::::/    /    \:::\    \       \:::\    \ /:::/    /    \:::\   \:::\    \               \:::\    \       \::::::/    /                   |::|/:::/    /         |::|   |          
   \:::\   \::::/    /      \:::\    \       \:::\    /:::/    /      \:::\   \:::\____\               \:::\    \       \::::/____/                    |::::::/    /          |::|   |          
    \:::\  /:::/    /        \:::\    \       \:::\__/:::/    /        \:::\   \::/    /                \:::\    \       \:::\    \                    |:::::/    /           |::|   |          
     \:::\/:::/    /          \:::\    \       \::::::::/    /          \:::\   \/____/                  \:::\    \       \:::\    \                   |::::/    /            |::|   |          
      \::::::/    /            \:::\    \       \::::::/    /            \:::\    \                       \:::\    \       \:::\    \                  /:::/    /             |::|   |          
       \::::/    /              \:::\____\       \::::/    /              \:::\____\                       \:::\____\       \:::\____\                /:::/    /              \::|   |          
        \::/____/                \::/    /        \::/____/                \::/    /                        \::/    /        \::/    /                \::/    /                \:|   |          
         ~~                       \/____/          ~~                       \/____/                          \/____/          \/____/                  \/____/                  \|___|          
	`
	fmt.Println(logo)
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

	// 4. 執行資料庫遷移（開發環境自動執行）
	ctx := context.Background()
	if cfg.Environment == "development" {
		log.Println("Running database migrations...")
		if err := db.Migrate(ctx); err != nil {
			log.Printf("Migration warning: %v", err)
		} else {
			log.Println("Migrations completed")
		}
	}

	// 5. 初始化 Repositories
	userRepo := repository.NewUserRepository(db.DB)
	bondRepo := repository.NewBondRepository(db.DB)
	txRepo := repository.NewTransactionRepository(db.DB)

	// 6. 初始化 Services
	userService := services.NewUserService(userRepo)
	bondService := services.NewBondService(bondRepo)

	// 7. 初始化 Session Manager
	sessionManager := session.NewMemorySessionManager()

	// 8. 初始化 Sui Client
	log.Println("🔄 Initializing Sui client...")
	suiClient := sui.NewSuiClient(cfg.SuiRPCURL)
	log.Println("✅ Sui client initialized")

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
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.RecoveryMiddleware())
	r.Use(middleware.RequestIDMiddleware())
	r.Use(middleware.LoggingMiddleware())

	// 12. 設定路由
	routes.SetupRoutes(r, userService, bondService, sessionManager, cfg)

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

	// 14. 啟動伺服器
	port := cfg.Port
	if port == "" {
		port = ":8080"
	}

	log.Printf("Server starting on port %s", port)
	log.Printf("Environment: %s", cfg.Environment)
	log.Printf("Sui RPC: %s", cfg.SuiRPCURL)

	// 15. 優雅關閉
	go func() {
		if err := r.Run(port); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 16. 等待中斷信號
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	<-shutdownCtx.Done()
	log.Println("Server shutdown complete")
}
