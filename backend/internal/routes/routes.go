package routes

import (
	"bluelink-backend/internal/config"
	"bluelink-backend/internal/handlers/auth"
	"bluelink-backend/internal/handlers/bonds"
	"bluelink-backend/internal/handlers/users"
	"bluelink-backend/internal/middleware"
	"bluelink-backend/internal/models"
	"bluelink-backend/internal/services"
	"bluelink-backend/internal/session"
	"net/http"

	"github.com/gin-gonic/gin"
)

// SetupRoutes 設定所有路由
func SetupRoutes(
	r *gin.Engine,
	userService *services.UserService,
	bondService *services.BondService,
	sessionManager *session.MemorySessionManager,
	cfg *config.Config,
) {
	// 判斷是否為生產環境
	isProduction := cfg.Environment == "production"

	// 初始化 handlers
	authHandler := auth.NewAuthHandler(userService, sessionManager, isProduction)
	profileHandler := users.NewProfileHandler(userService)
	bondHandler := bonds.NewBondHandler(bondService)

	// API v1
	v1 := r.Group("/api/v1")

	// ===== 公開路由（不需要任何驗證）=====
	public := v1.Group("/")
	{
		// 後端狀況確認
		public.GET("/health", func(c *gin.Context) {
			models.RespondWithSuccess(c, http.StatusOK, "Service is healthy", gin.H{
				"status": "ok",
			})
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
		// SessionAuth - 驗證 Cookie 中的 session_id 並載入使用者資訊到 Context
		middleware.SessionAuthMiddleware(sessionManager),
	)
	{
		// User 相關
		protected.GET("/profile", profileHandler.GetProfile)
		protected.PUT("/profile", profileHandler.UpdateProfile)
		protected.GET("/profile/full", profileHandler.GetFullProfile)

		// Session 管理（有額外的速率限制）
		sessionGroup := protected.Group("/sessions")
		sessionGroup.Use(middleware.RateLimitMiddleware(30)) // 每分鐘 30 次
		{
			sessionGroup.GET("", authHandler.GetActiveSessions)            // 取得所有 session
			sessionGroup.DELETE("/:session_id", authHandler.RevokeSession) // 撤銷特定 session
		}

		// Bond 相關
		protected.GET("/bonds", bondHandler.GetAllBonds)
	}

	// ===== 4. 管理員路由（需要 Session + 管理員權限）=====
	admin := v1.Group("/admin")
	admin.Use(
		middleware.SessionAuthMiddleware(sessionManager),
		middleware.RequireRoleMiddleware("admin"),
	)
	{
		// TODO: 管理員功能路由
		// admin.GET("/users", adminHandler.GetAllUsers)
		// admin.PUT("/users/:id/role", adminHandler.UpdateUserRole)
	}
}
