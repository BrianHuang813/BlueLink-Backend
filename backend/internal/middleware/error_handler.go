package middleware

import (
	"bluelink-backend/internal/logger"
	"bluelink-backend/internal/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ErrorHandlerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			requestID, _ := c.Get("request_id")
			requestIDStr := ""
			if requestID != nil {
				requestIDStr = requestID.(string)
			}

			// 記錄錯誤到日誌
			logger.ErrorWithContext(
				requestIDStr,
				c.Request.URL.Path,
				c.Request.Method,
				"Request error",
				err.Err,
			)

			response := models.ErrorResponse{
				Code:      http.StatusInternalServerError,
				Message:   "Internal server error",
				RequestID: requestIDStr,
			}

			switch err.Type {
			case gin.ErrorTypeBind:
				response.Code = http.StatusBadRequest
				response.Message = "Invalid request format"
				if gin.Mode() == gin.DebugMode {
					response.Details = err.Error()
				}
			case gin.ErrorTypePublic:
				response.Code = http.StatusInternalServerError
				response.Message = "Internal server error"
			}

			c.JSON(response.Code, response)
		}
	}
}
