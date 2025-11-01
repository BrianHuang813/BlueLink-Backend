package bonds

import (
	"bluelink-backend/internal/models"
	"time"
)

// BondResponse 債券響應格式 (符合前端 API 規範)
type BondResponse struct {
	ID                 int64  `json:"id"`
	OnChainID          string `json:"on_chain_id"`
	IssuerAddress      string `json:"issuer_address"`
	IssuerName         string `json:"issuer_name"`
	BondName           string `json:"bond_name"`
	BondImageURL       string `json:"bond_image_url"`
	TokenImageURL      string `json:"token_image_url"`
	TotalAmount        int64  `json:"total_amount"`         // MIST 單位
	AmountRaised       int64  `json:"amount_raised"`        // MIST 單位
	AmountRedeemed     int64  `json:"amount_redeemed"`      // MIST 單位
	TokensIssued       int64  `json:"tokens_issued"`        // 已發行代幣數量
	TokensRedeemed     int64  `json:"tokens_redeemed"`      // 已贖回代幣數量
	AnnualInterestRate int64  `json:"annual_interest_rate"` // 基點 (5% = 500)
	MaturityDate       string `json:"maturity_date"`        // ISO 8601 格式
	IssueDate          string `json:"issue_date"`           // ISO 8601 格式
	Active             bool   `json:"active"`
	Redeemable         bool   `json:"redeemable"`
	MetadataURL        string `json:"metadata_url"`
	CreatedAt          string `json:"created_at"` // ISO 8601 格式
	UpdatedAt          string `json:"updated_at"` // ISO 8601 格式
}

// ToBondResponse 將 Bond 模型轉換為 API 響應格式
func ToBondResponse(bond *models.Bond) *BondResponse {
	if bond == nil {
		return nil
	}

	// 將日期字符串 (YYYY-MM-DD) 轉換為 ISO 8601 格式 (YYYY-MM-DDTHH:mm:ssZ)
	maturityDate := formatDateToISO8601(bond.MaturityDate)
	issueDate := formatDateToISO8601(bond.IssueDate)

	return &BondResponse{
		ID:                 bond.ID,
		OnChainID:          bond.OnChainID,
		IssuerAddress:      bond.IssuerAddress,
		IssuerName:         bond.IssuerName,
		BondName:           bond.BondName,
		BondImageURL:       bond.BondImageUrl,
		TokenImageURL:      bond.TokenImageUrl,
		TotalAmount:        bond.TotalAmount,
		AmountRaised:       bond.AmountRaised,
		AmountRedeemed:     bond.AmountRedeemed,
		TokensIssued:       bond.TokensIssued,
		TokensRedeemed:     bond.TokensRedeemed,
		AnnualInterestRate: bond.AnnualInterestRate,
		MaturityDate:       maturityDate,
		IssueDate:          issueDate,
		Active:             bond.Active,
		Redeemable:         bond.Redeemable,
		MetadataURL:        bond.MetadataUrl,
		CreatedAt:          bond.CreatedAt.Format(time.RFC3339),
		UpdatedAt:          bond.UpdatedAt.Format(time.RFC3339),
	}
}

// ToBondResponseList 將 Bond 列表轉換為響應格式列表
func ToBondResponseList(bonds []*models.Bond) []*BondResponse {
	if bonds == nil {
		return []*BondResponse{}
	}

	responses := make([]*BondResponse, 0, len(bonds))
	for _, bond := range bonds {
		responses = append(responses, ToBondResponse(bond))
	}
	return responses
}

// formatDateToISO8601 將日期字符串 (YYYY-MM-DD) 轉換為 ISO 8601 格式
func formatDateToISO8601(dateStr string) string {
	if dateStr == "" {
		return ""
	}

	// 嘗試解析日期字符串
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		// 如果解析失敗,返回原始字符串
		return dateStr
	}

	// 轉換為 ISO 8601 格式 (UTC 時區,午夜時間)
	return t.UTC().Format(time.RFC3339)
}
