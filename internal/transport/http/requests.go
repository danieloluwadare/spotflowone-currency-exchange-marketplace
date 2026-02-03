package httptransport

type AllocationRequest struct {
	VendorID      string `json:"vendorId"`
	BaseCurrency  string `json:"baseCurrency"`
	QuoteCurrency string `json:"quoteCurrency"`
	Rate          string `json:"rate"`
	Available     string `json:"available"`
}

type AllocationResponse struct {
	VendorID      string `json:"vendorId"`
	BaseCurrency  string `json:"baseCurrency"`
	QuoteCurrency string `json:"quoteCurrency"`
	Rate          string `json:"rate"`
	Available     string `json:"available"`
}

type OrderRequest struct {
	Direction     string `json:"direction"`
	BaseCurrency  string `json:"baseCurrency"`
	QuoteCurrency string `json:"quoteCurrency"`
	Amount        string `json:"amount"`
}

type FillResponse struct {
	VendorID string `json:"vendorId"`
	Rate     string `json:"rate"`
	Amount   string `json:"amount"`
}

type OrderResponse struct {
	Status           string         `json:"status"`
	RequestedAmount  string         `json:"requestedAmount"`
	FilledAmount     string         `json:"filledAmount"`
	Fills            []FillResponse `json:"fills"`
	BlendedRate      string         `json:"blendedRate"`
	MarketRate       *string        `json:"marketRate"`
	MarketRateSource *string        `json:"marketRateSource"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
