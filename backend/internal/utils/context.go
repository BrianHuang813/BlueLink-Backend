package utils

import (
	"errors"

	"github.com/gin-gonic/gin"
)

// GetWalletAddress 從 context 取得錢包地址
func GetWalletAddress(c *gin.Context) (string, error) {
	wallet, exists := c.Get("WalletAddress")
	if !exists {
		return "", errors.New("wallet address not found in context")
	}
	return wallet.(string), nil
}

// GetUserID 從 context 取得使用者 ID
func GetUserID(c *gin.Context) (int64, error) {
	userID, exists := c.Get("UserID")
	if !exists {
		return 0, errors.New("user ID not found in context")
	}
	return userID.(int64), nil
}

// GetUserRole 從 context 取得使用者角色
func GetUserRole(c *gin.Context) (string, error) {
	role, exists := c.Get("Role")
	if !exists {
		return "", errors.New("user role not found in context")
	}
	return role.(string), nil
}

// GetRequestID 從 context 取得請求 ID
func GetRequestID(c *gin.Context) string {
	requestID, exists := c.Get("RequestID")
	if !exists {
		return ""
	}
	return requestID.(string)
}
