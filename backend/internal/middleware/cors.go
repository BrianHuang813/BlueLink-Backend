package middleware

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CORSMiddleware(allowedOrigins []string) gin.HandlerFunc {
	config := cors.Config{
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization", "X-Request-ID"},
		ExposeHeaders:    []string{"Content-Length", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}

	// 如果設定為 "*"，則允許所有來源
	if len(allowedOrigins) == 1 && allowedOrigins[0] == "*" {
		config.AllowAllOrigins = true
		// 注意：當 AllowAllOrigins = true 時，AllowCredentials 必須為 false
		config.AllowCredentials = false
	} else {
		config.AllowOrigins = allowedOrigins
	}

	return cors.New(config)
}
