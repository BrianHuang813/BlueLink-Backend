package routes

import (
	"bluelink-backend/internal/handlers/auth"
	"bluelink-backend/internal/handlers/users"
	"bluelink-backend/internal/middleware"
	"bluelink-backend/internal/services"
	"bluelink-backend/internal/sui"

	"github.com/gin-gonic/gin"
)

// SetupRoutes 設定所有路由
func SetupRoutes(r *gin.Engine, suiClient *sui.Client, userService *services.UserService) {
	// 初始化 handlers
	authHandler := auth.NewAuthHandler(userService)
	profileHandler := users.NewProfileHandler(userService)

	// API v1
	v1 := r.Group("/api/v1")

	// ===== Public Routes（不需要驗證）=====
	{
		// Health check
		v1.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status":  "ok",
				"message": "BlueLink Backend is running",
			})
		})

		// 錢包驗證相關
		authGroup := v1.Group("/auth")
		{
			authGroup.POST("/challenge", authHandler.GenerateChallenge) // 取得挑戰訊息
			authGroup.POST("/verify", authHandler.VerifySignature)      // 驗證簽名
		}

		// 公開的債券資訊
		v1.GET("/bonds", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "List all public bonds"})
		})

		v1.GET("/bonds/:id", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "Get bond detail"})
		})
	}

	// ===== Protected Routes（需要錢包驗證）=====
	protected := v1.Group("/")
	protected.Use(
		middleware.WalletAuthMiddleware(),
		middleware.UserContextMiddleware(),
	)
	{
		// 使用者資料相關
		protected.GET("/profile", profileHandler.GetProfile)          // 基本資料
		protected.GET("/profile/full", profileHandler.GetFullProfile) // 完整資料
		protected.PUT("/profile", profileHandler.UpdateProfile)       // 更新資料

		// 債券操作（需要驗證）
		protected.POST("/bonds", func(c *gin.Context) {
			c.JSON(201, gin.H{"message": "Create new bond"})
		})

		protected.POST("/bonds/:id/purchase", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "Purchase bond"})
		})

		// 使用者的債券持有記錄
		protected.GET("/my-bonds", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "Get my bonds"})
		})

		// 使用者的交易歷史
		protected.GET("/transactions", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "Get my transactions"})
		})
	}
}
