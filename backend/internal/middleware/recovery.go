package middleware

import (
	"bluelink-backend/internal/logger"
	"bluelink-backend/internal/models"
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				requestID, _ := c.Get("request_id")
				requestIDStr := ""
				if requestID != nil {
					requestIDStr = requestID.(string)
				}

				// 記錄 panic 詳細資訊和堆疊到日誌
				stackTrace := string(debug.Stack())
				logger.Error("PANIC RECOVERED: RequestID=%s Path=%s Method=%s Error=%v\nStack:\n%s",
					requestIDStr, c.Request.URL.Path, c.Request.Method, err, stackTrace)

				// 返回統一錯誤格式
				response := models.ErrorResponse{
					Code:      http.StatusInternalServerError,
					Message:   "Internal server error",
					RequestID: requestIDStr,
				}

				// 只在開發環境顯示詳細錯誤
				if gin.Mode() == gin.DebugMode {
					response.Details = fmt.Sprintf("%v", err)
				}

				c.JSON(http.StatusInternalServerError, response)
				c.Abort()
			}
		}()
		c.Next()
	}
}
