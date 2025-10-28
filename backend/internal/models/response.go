package models

import (
	"bluelink-backend/internal/logger"
	"net/http"

	"github.com/gin-gonic/gin"
)

// APIResponse 統一的 API 回應格式
type APIResponse struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
}

// ErrorResponse 錯誤回應格式
type ErrorResponse struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	Details   string `json:"details,omitempty"`
	RequestID string `json:"request_id,omitempty"`
}

// SuccessResponse 成功回應的輔助函數
func SuccessResponse(code int, message string, data interface{}) APIResponse {
	return APIResponse{
		Code:    code,
		Message: message,
		Data:    data,
	}
}

// ErrorResponseWithDetails 錯誤回應的輔助函數
func ErrorResponseWithDetails(code int, message string, details string) ErrorResponse {
	return ErrorResponse{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// RespondWithError 統一的錯誤回應函數（包含日誌記錄）
func RespondWithError(c *gin.Context, code int, message string, err error) {
	requestID := getRequestID(c)

	// 記錄錯誤到日誌
	if err != nil {
		logger.ErrorWithContext(
			requestID,
			c.Request.URL.Path,
			c.Request.Method,
			message,
			err,
		)
	} else {
		logger.WarnWithContext(
			requestID,
			c.Request.URL.Path,
			c.Request.Method,
			message,
		)
	}

	// 返回錯誤回應
	response := ErrorResponse{
		Code:      code,
		Message:   message,
		RequestID: requestID,
	}

	// 只在開發環境顯示詳細錯誤
	if err != nil && gin.Mode() == gin.DebugMode {
		response.Details = err.Error()
	}

	c.JSON(code, response)
}

// RespondWithErrorDetails 統一的錯誤回應函數（包含詳細資訊）
func RespondWithErrorDetails(c *gin.Context, code int, message string, details string) {
	requestID := getRequestID(c)

	// 記錄錯誤到日誌
	logger.WarnWithContext(
		requestID,
		c.Request.URL.Path,
		c.Request.Method,
		message+": "+details,
	)

	// 返回錯誤回應
	c.JSON(code, ErrorResponse{
		Code:      code,
		Message:   message,
		Details:   details,
		RequestID: requestID,
	})
}

// RespondWithSuccess 統一的成功回應函數（包含日誌記錄）
func RespondWithSuccess(c *gin.Context, code int, message string, data any) {
	requestID := getRequestID(c)

	// 記錄成功訊息到日誌（可選）
	if gin.Mode() == gin.DebugMode {
		logger.InfoWithContext(
			requestID,
			c.Request.URL.Path,
			c.Request.Method,
			message,
		)
	}

	// 返回成功回應
	c.JSON(code, APIResponse{
		Code:      code,
		Message:   message,
		Data:      data,
		RequestID: requestID,
	})
}

// RespondBadRequest 400 錯誤
func RespondBadRequest(c *gin.Context, message string, err error) {
	RespondWithError(c, http.StatusBadRequest, message, err)
}

// RespondUnauthorized 401 錯誤
func RespondUnauthorized(c *gin.Context, message string) {
	RespondWithError(c, http.StatusUnauthorized, message, nil)
}

// RespondForbidden 403 錯誤
func RespondForbidden(c *gin.Context, message string) {
	RespondWithError(c, http.StatusForbidden, message, nil)
}

// RespondNotFound 404 錯誤
func RespondNotFound(c *gin.Context, message string) {
	RespondWithError(c, http.StatusNotFound, message, nil)
}

// RespondInternalError 500 錯誤
func RespondInternalError(c *gin.Context, message string, err error) {
	RespondWithError(c, http.StatusInternalServerError, message, err)
}

// RespondTooManyRequests 429 錯誤
func RespondTooManyRequests(c *gin.Context, message string) {
	RespondWithError(c, http.StatusTooManyRequests, message, nil)
}

// getRequestID 從 context 中獲取 request ID
func getRequestID(c *gin.Context) string {
	if requestID, exists := c.Get("request_id"); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return ""
}
