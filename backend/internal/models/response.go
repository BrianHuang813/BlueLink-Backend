package models

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
