package middleware

import (
	"log"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				requestID, _ := c.Get("RequestID")

				// 記錄詳細錯誤和堆疊
				log.Printf("PANIC: %v\nStack: %s\nRequestID: %v\nPath: %s\nMethod: %s",
					err, debug.Stack(), requestID, c.Request.URL.Path, c.Request.Method)

				c.JSON(http.StatusInternalServerError, gin.H{
					"code":       http.StatusInternalServerError,
					"message":    "Internal server error",
					"request_id": requestID,
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}
