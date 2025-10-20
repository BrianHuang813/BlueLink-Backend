package models

type Bond struct {
	ID            int64   `json:"id" db:"id"`
	OnChainID     string  `json:"on_chain_id" db:"on_chain_id"`
	Name          string  `json:"name" db:"name"`
	IssuerName    string  `json:"issuer" db:"issuer"`
	IssuerAddress string  `json:"issuer_address" db:"issuer_address"`
	Description   string  `json:"description" db:"description"`
	FaceValue     float64 `json:"face_value" db:"face_value"`
	Currency      string  `json:"currency" db:"currency"`
	MaturityDate  string  `json:"maturity_date" db:"maturity_date"`
	InterestRate  float64 `json:"interest_rate" db:"interest_rate"`
	Status        string  `json:"status" db:"status"` // e.g., "active", "matured", "defaulted"
	CreatedAt     string  `json:"created_at" db:"created_at"`
	UpdatedAt     string  `json:"updated_at" db:"updated_at"`
	DeletedAt     *string `json:"deleted_at,omitempty" db:"deleted_at"` // 軟刪除時間戳記
}
