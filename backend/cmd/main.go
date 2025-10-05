package cmd

import (
	"bluelink-backend/internal/config"
	"bluelink-backend/internal/middleware"
	"bluelink-backend/internal/routes"
	"bluelink-backend/internal/sui"
	"fmt"

	"github.com/gin-gonic/gin"
)

// setupGinEngine 設定 Gin 引擎和全域 middleware
func setupGinEngine(config *config.Config) *gin.Engine {
	var r *gin.Engine

	if config.Environment == "production" {
		// 生產環境：完整的安全設定
		gin.SetMode(gin.ReleaseMode)
		r = gin.New()
		r.Use(
			middleware.RecoveryMiddleware(),     // 防止 panic 崩潰
			middleware.RequestIDMiddleware(),    // 追蹤請求
			middleware.LoggingMiddleware(),      // 記錄日誌
			middleware.CORSMiddleware(),         // 跨域處理
			middleware.RateLimitMiddleware(100), // 防止 DDoS（每分鐘 100 次）
			middleware.ErrorHandlerMiddleware(), // 統一錯誤處理
		)
	} else {
		// 開發環境：包含彩色日誌
		r = gin.Default()
		r.Use(
			middleware.RequestIDMiddleware(),
			middleware.CORSMiddleware(),
			middleware.RateLimitMiddleware(200), // 開發環境較寬鬆
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

	// Load environment variables from .env file
	config := config.LoadConfig()

	// Initialize Sui client
	suiClient := sui.NewClient(config.SuiRPCURL)

	// Setup Gin engine
	engine := setupGinEngine(config)

	// TODO: Initialize database connection
	// db, err := initDatabase(config)
	// if err != nil {
	//     log.Fatalf("Failed to initialize database: %v", err)
	// }
	// defer db.Close()

	// TODO: Initialize services
	// userService := services.NewUserService(db)

	// Setup routes (暫時傳 nil 作為 userService，等資料庫設定完成後再更新)
	routes.SetupRoutes(engine, suiClient, nil)

	// Start server
	port := config.Port
	if port == "" {
		port = "8080"
	}
	fmt.Printf("Server running on port %s\n", port)
	if err := engine.Run(":" + port); err != nil {
		fmt.Printf("Failed to run server: %v\n", err)
	}
}
