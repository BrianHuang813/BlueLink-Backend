package routes

import (
	"bluelink-backend/internal/handlers/auth"
	"bluelink-backend/internal/handlers/bonds"
	"bluelink-backend/internal/handlers/users"
	"bluelink-backend/internal/middleware"
	"bluelink-backend/internal/services"
	"bluelink-backend/internal/session"

	"github.com/gin-gonic/gin"
)

// SetupRoutes 設定所有路由
func SetupRoutes(
	r *gin.Engine,
	userService *services.UserService,
	bondService *services.BondService,
	sessionManager *session.MemorySessionManager,
) {
	// 初始化 handlers
	authHandler := auth.NewAuthHandler(userService, sessionManager)
	profileHandler := users.NewProfileHandler(userService)
	bondHandler := bonds.NewBondHandler(bondService)

	// API v1
	v1 := r.Group("/api/v1")

	// ===== 公開路由（不需要任何驗證）=====
	public := v1.Group("/")
	{
		// 後端狀況確認
		public.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok"})
		})
	}

	// ===== 認證路由（不需要 Session，但需要限流）=====
	authGroup := v1.Group("/auth")
	authGroup.Use(
		// 認證專用的更嚴格限流（
		middleware.RateLimitMiddleware(20), // 每分鐘 20 次
	)
	{
		// 前端請求 nonce
		authGroup.POST("/challenge", authHandler.GenerateChallenge)

		// 前端提交簽名（這裡會建立 Session）
		authGroup.POST("/verify", authHandler.VerifySignature)

		// 登出（需要 Session）
		authGroup.POST(
			"/logout",
			middleware.SessionAuthMiddleware(sessionManager),
			authHandler.Logout,
		)

		authGroup.POST(
			"/logout-all",
			middleware.SessionAuthMiddleware(sessionManager),
			authHandler.LogoutAll,
		)
	}

	// ===== 3. 受保護路由（需要 Session）=====
	protected := v1.Group("/")
	protected.Use(
		// 7️⃣ SessionAuth - 驗證 Cookie 中的 session_id
		middleware.SessionAuthMiddleware(sessionManager),

		// 8️⃣ UserContext - 從 Session 載入使用者資訊到 Context
		middleware.BasicUserContextMiddleware(sessionManager),
	)
	{
		// 使用者相關
		protected.GET("/profile", profileHandler.GetProfile)
		protected.PUT("/profile", profileHandler.UpdateProfile)
		protected.GET("/profile/full", profileHandler.GetFullProfile)

		// Session 管理
		protected.GET("/auth/sessions", authHandler.GetActiveSessions)

		// 債券相關（之後實作）
		protected.GET("/bonds", bondHandler.GetAllList)
		// protected.POST("bonds/redeem", bondHandler.RedeemBond)
		// protected.GET("bonds/owned", bondHandler.GetOwnedBonds)
	}

	// ===== 4. 管理員路由（需要 Session + 管理員權限）=====
	admin := v1.Group("/admin")
	admin.Use(
		middleware.SessionAuthMiddleware(sessionManager),
		middleware.BasicUserContextMiddleware(sessionManager),
		middleware.RequireRoleMiddleware("admin"),
	)
	{
		// 管理員功能（之後實作）
		// admin.GET("/users", adminHandler.GetAllUsers)
		// admin.PUT("/users/:id/role", adminHandler.UpdateUserRole)
	}
}
