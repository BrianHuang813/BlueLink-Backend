package routes

import (
	"bluelink-backend/internal/handlers/auth"
	"bluelink-backend/internal/handlers/users"
	"bluelink-backend/internal/middleware"
	"bluelink-backend/internal/services"
	"bluelink-backend/internal/session"
	"bluelink-backend/internal/sui"

	"github.com/gin-gonic/gin"
)

// SetupRoutes 設定所有路由
func SetupRoutes(
	r *gin.Engine,
	suiClient *sui.Client,
	userService *services.UserService,
	sessionManager *session.MemorySessionManager,
) {
	// 初始化 handlers
	authHandler := auth.NewAuthHandler(userService, sessionManager)
	profileHandler := users.NewProfileHandler(userService)

	// API v1
	public := r.Group("/api/v1")

	// ===== Public Routes =====
	{
		// Health check
		public.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status":  "ok",
				"message": "BlueLink Backend is running",
			})
		})

		// 認證端點
		authGroup := public.Group("/auth")
		{
			authGroup.POST("/challenge", authHandler.GenerateChallenge) // 取得挑戰訊息
			authGroup.POST("/verify", authHandler.VerifySignature)      // 驗證簽名並登入
		}

		// 公開資訊
		public.GET("/bonds", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "List all public bonds"})
		})
	}

	// ===== Protected Routes (Session Authentication)=====
	protected := public.Group("/")
	protected.Use(
		middleware.SessionAuthMiddleware(sessionManager),      // 驗證 Session
		middleware.BasicUserContextMiddleware(sessionManager), // 載入用戶資訊
	)
	{
		// 認證管理
		authGroup := protected.Group("/auth")
		{
			authGroup.POST("/logout", authHandler.Logout)
			authGroup.POST("/logout-all", authHandler.LogoutAll)
			authGroup.GET("/sessions", authHandler.GetActiveSessions)
		}

		// 使用者資料
		protected.GET("/profile", profileHandler.GetProfile)
		protected.GET("/profile/full", profileHandler.GetFullProfile)
		protected.PUT("/profile", profileHandler.UpdateProfile)

		// 債券操作
		protected.POST("/bonds", func(c *gin.Context) {
			c.JSON(201, gin.H{"message": "Create bond"})
		})

		protected.GET("/my-bonds", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "Get my bonds"})
		})
	}
}
