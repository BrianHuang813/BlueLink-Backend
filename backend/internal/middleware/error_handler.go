package middleware

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ErrorResponse struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	Details   string `json:"details,omitempty"`
	RequestID string `json:"request_id,omitempty"`
}

func ErrorHandlerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			requestID, _ := c.Get("request_id")

			log.Printf("Request error: %v, Path: %s, Method: %s, RequestID: %v",
				err.Error(), c.Request.URL.Path, c.Request.Method, requestID)

			response := ErrorResponse{
				Code:    http.StatusInternalServerError,
				Message: "Internal server error",
			}

			if requestID != nil {
				response.RequestID = requestID.(string)
			}

			switch err.Type {
			case gin.ErrorTypeBind:
				response.Code = http.StatusBadRequest
				response.Message = "Invalid request format"
				response.Details = err.Error()
			case gin.ErrorTypePublic:
				response.Code = http.StatusInternalServerError
				response.Message = "Internal server error"
			}

			c.JSON(response.Code, response)
		}
	}
}
