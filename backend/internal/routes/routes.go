package routes

import (
	"bluelink-backend/internal/config"
	"bluelink-backend/internal/handlers/auth"
	"bluelink-backend/internal/handlers/bonds"
	"bluelink-backend/internal/handlers/users"
	"bluelink-backend/internal/middleware"
	"bluelink-backend/internal/models"
	"bluelink-backend/internal/repository"
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
	bondTokenService *services.BondTokenService,
	syncService *services.SyncService,
	sessionManager session.SessionManager,
	nonceRepo *repository.NonceRepository,
	cfg *config.Config,
) {
	// 判斷是否為生產環境
	isProduction := cfg.Environment == "production"

	// 初始化 handlers
	authHandler := auth.NewAuthHandler(userService, sessionManager, nonceRepo, isProduction)
	profileHandler := users.NewProfileHandler(userService)
	bondHandler := bonds.NewBondHandler(bondService, bondTokenService, syncService)

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

	// ===== 債券市場公開路由（不需要認證）=====
	bondsPublic := v1.Group("/bonds")
	{
		// 獲取所有債券 - 公開訪問
		bondsPublic.GET("", bondHandler.GetAllBonds)
		bondsPublic.GET("/:id", bondHandler.GetBondByID)

		// 同步鏈上交易 - 需要認證
		bondsPublic.POST("/sync",
			middleware.SessionAuthMiddleware(sessionManager),
			bondHandler.SyncTransaction,
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

		// 🆕 BondToken 相關
		protected.GET("/bond-tokens/:id", bondHandler.GetBondTokenByID)
		protected.GET("/bond-tokens/on-chain/:on_chain_id", bondHandler.GetBondTokenByOnChainID)
		protected.GET("/bond-tokens/owner", bondHandler.GetBondTokensByOwner)     // Query: ?owner=0x...&limit=10&offset=0
		protected.GET("/bond-tokens/project", bondHandler.GetBondTokensByProject) // Query: ?project_id=0x...&limit=10&offset=0
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
