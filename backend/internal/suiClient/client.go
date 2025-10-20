package suiClient

import (
	"fmt"

	"github.com/block-vision/sui-go-sdk/sui"
)

// NewSuiClient 創建 Sui 客戶端
func NewSuiClient(rpcUrl string) (*sui.ISuiAPI, error) {
	if rpcUrl == "" {
		return nil, fmt.Errorf("sui RPC URL cannot be empty")
	}

	// 使用傳入的 rpcUrl
	suiClient := sui.NewSuiClient(rpcUrl)

	return &suiClient, nil
}
