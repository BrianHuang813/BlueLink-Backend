package bonds

type CreateBondRequest struct {
	ID            int64   `json:"id" binding:"required"`
	Name          string  `json:"name" binding:"required"`
	IssuerName    string  `json:"issuer_name" binding:"required"`
	IssuerAddress string  `json:"issuer_address" binding:"required"`
	FaceValue     float64 `json:"face_value" binding:"required,gt=0"`
	Currency      string  `json:"currency" binding:"required"`
}

type BuyBondRequest struct {
	ID           int64   `json:"id" binding:"required"`
	OnChainID    string  `json:"on_chain_id" binding:"required"`
	Name         string  `json:"name" binding:"required"`
	Amount       float64 `json:"amount" binding:"required,gt=0"`
	BuyerAddress string  `json:"buyer_address" binding:"required"`
}

type GetBondByIDRequest struct {
	ID int64 `uri:"id" binding:"required"`
}

type GetAllBondsRequest struct {
}
