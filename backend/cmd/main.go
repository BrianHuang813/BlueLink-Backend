package cmd

import (
	"bluelink-backend/internal/config"
	"bluelink-backend/internal/middleware"
	"bluelink-backend/internal/routes"
	"bluelink-backend/internal/services"
	"bluelink-backend/internal/session"
	"fmt"
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
)

/*
	請求進入
	↓
	1️⃣ RecoveryMiddleware()        ← 最外層：捕捉 panic
	↓
	2️⃣ RequestIDMiddleware()       ← 生成請求 ID
	↓
	3️⃣ LoggingMiddleware()         ← 記錄請求（使用 RequestID）
	↓
	4️⃣ CORSMiddleware()            ← 處理跨域（OPTIONS 請求在這裡結束）
	↓
	5️⃣ RateLimitMiddleware()       ← 限制請求頻率
	↓
	6️⃣ ErrorHandlerMiddleware()    ← 統一錯誤處理
	↓
	【路由匹配】
	↓
	【路由級別 Middleware】
	↓
	【Handler】
	↓
	響應返回（依序經過所有 Middleware）
*/

// setupGinEngine 設定 Gin 引擎和全域 middleware
func setupGinEngine(config *config.Config) *gin.Engine {
	var r *gin.Engine

	if config.Environment == "production" {
		// 生產環境：完整的安全設定
		gin.SetMode(gin.ReleaseMode)
		r = gin.New() // 不使用 gin.Default()，手動配置

		// ===== 全域 Middleware =====
		r.Use(
			// 最外層，捕捉所有 panic
			middleware.RecoveryMiddleware(),

			// 為每個請求分配唯一 ID
			middleware.RequestIDMiddleware(),

			// Logging - 記錄請求
			middleware.LoggingMiddleware(),

			// CORS - 處理跨域請求（OPTIONS 請求在這裡就返回）
			middleware.CORSMiddleware(),

			// RateLimit - 限制請求頻率
			middleware.RateLimitMiddleware(100), // 每分鐘 100 次

			// ErrorHandler - 統一錯誤格式
			middleware.ErrorHandlerMiddleware(),
		)
	} else {
		// 開發環境：較寬鬆的設定
		r = gin.Default() // 內建 Logger + Recovery

		r.Use(
			middleware.RequestIDMiddleware(),
			middleware.CORSMiddleware(),
			middleware.RateLimitMiddleware(1000), // 開發環境很寬鬆
			middleware.ErrorHandlerMiddleware(),
		)
	}

	return r
}

func main() {
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
	fmt.Println("Starting BlueLink Backend Application...")
	fmt.Println(logo)

	opts := &slog.HandlerOptions{
		Level: slog.LevelDebug, // 設定日誌級別，只有高於或等於 Debug 的日誌會被記錄
	}
	handler := slog.NewJSONHandler(os.Stdout, opts)

	slog.SetDefault(slog.New(handler))

	// 載入設定
	config := config.LoadConfig()

	// 設定 Gin engine
	engine := setupGinEngine(config)

	// TODO: Create databse initial function
	// 初始化資料庫
	db, err := initDatabase(config)
	if err != nil {
		slog.Error("Failed to initialize the database.", "error", err)
	}
	defer db.Close()

	// 初始化服務層
	userService := services.NewUserService(db)
	bondService := services.NewBondService(db)

	// 初始化 Session 管理器
	sessionManager := session.NewMemorySessionManager()

	// 設定路由
	routes.SetupRoutes(engine, userService, bondService, sessionManager)

	// 啟動伺服器
	port := config.Port
	if port == "" {
		port = "8080"
	}
	slog.Info("Server starts running", "port", port)

	if err := engine.Run(":" + port); err != nil {
		slog.Error("Failed to initialize the server.", "error", err)
	}
}
