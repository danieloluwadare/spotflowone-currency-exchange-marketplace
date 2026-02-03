package httptransport

import (
	"encoding/json"
	"net/http"

	"github.com/shopspring/decimal"

	"spotflowone/internal/domain"
	"spotflowone/internal/service"
)

type Handler struct {
	allocationService *service.AllocationService
	orderService      *service.OrderService
}

// NewHandler creates a new HTTP handler with the provided services.
// Returns a pointer to the initialized handler.
func NewHandler(allocationService *service.AllocationService, orderService *service.OrderService) *Handler {
	return &Handler{
		allocationService: allocationService,
		orderService:      orderService,
	}
}

// CreateAllocation handles POST /allocations requests to create a new vendor allocation.
// Parses the JSON request body, validates decimal fields (rate and available),
// and delegates to the allocation service. Returns 201 Created on success,
// 400 Bad Request for invalid input, or 500 Internal Server Error for other failures.
func (h *Handler) CreateAllocation(w http.ResponseWriter, r *http.Request) {
	var req AllocationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	rate, err := decimal.NewFromString(req.Rate)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid rate")
		return
	}
	available, err := decimal.NewFromString(req.Available)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid available amount")
		return
	}

	alloc := domain.Allocation{
		VendorID:      req.VendorID,
		BaseCurrency:  req.BaseCurrency,
		QuoteCurrency: req.QuoteCurrency,
		Rate:          rate,
		Available:     available,
	}

	if err := h.allocationService.AddAllocation(r.Context(), alloc); err != nil {
		status := http.StatusInternalServerError
		if err == service.ErrInvalidInput {
			status = http.StatusBadRequest
		}
		writeError(w, status, err.Error())
		return
	}

	resp := AllocationResponse{
		VendorID:      alloc.VendorID,
		BaseCurrency:  domain.NormalizeCurrency(alloc.BaseCurrency),
		QuoteCurrency: domain.NormalizeCurrency(alloc.QuoteCurrency),
		Rate:          alloc.Rate.String(),
		Available:     alloc.Available.String(),
	}

	writeJSON(w, http.StatusCreated, resp)
}

// CreateOrder handles POST /orders requests to execute a buy or sell order.
// Parses the JSON request body, validates the amount, and delegates to the order service.
// Returns 200 OK for fully filled orders, 206 Partial Content for partially filled orders,
// 400 Bad Request for invalid input, 422 Unprocessable Entity for no liquidity,
// or 500 Internal Server Error for other failures. The response includes fill details,
// blended rate, and market rate comparison.
func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req OrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid amount")
		return
	}

	order := domain.Order{
		Direction:     domain.Direction(req.Direction),
		BaseCurrency:  req.BaseCurrency,
		QuoteCurrency: req.QuoteCurrency,
		Amount:        amount,
	}

	result, err := h.orderService.Execute(r.Context(), order)
	if err != nil {
		status := http.StatusInternalServerError
		switch err {
		case service.ErrInvalidInput:
			status = http.StatusBadRequest
		case service.ErrNoLiquidity:
			status = http.StatusUnprocessableEntity
		}
		writeError(w, status, err.Error())
		return
	}

	fills := make([]FillResponse, 0, len(result.Fills))
	for _, fill := range result.Fills {
		fills = append(fills, FillResponse{
			VendorID: fill.VendorID,
			Rate:     fill.Rate.String(),
			Amount:   fill.Amount.String(),
		})
	}

	blendedRate := result.BlendedRate.String()
	marketRate := (*string)(nil)
	marketSource := (*string)(nil)
	if result.MarketRate.Available {
		value := result.MarketRate.Rate.String()
		source := result.MarketRate.Source
		marketRate = &value
		marketSource = &source
	}

	resp := OrderResponse{
		Status:           string(result.Status),
		RequestedAmount:  result.RequestedAmount.String(),
		FilledAmount:     result.FilledAmount.String(),
		Fills:            fills,
		BlendedRate:      blendedRate,
		MarketRate:       marketRate,
		MarketRateSource: marketSource,
	}

	status := http.StatusOK
	if result.Status == domain.StatusPartiallyFilled {
		status = http.StatusPartialContent
	}

	writeJSON(w, status, resp)
}

// writeJSON writes a JSON response with the given status code and payload.
// Sets the Content-Type header to application/json and encodes the payload as JSON.
func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

// writeError writes an error response as JSON with the given status code and error message.
// Wraps the message in an ErrorResponse struct and delegates to writeJSON.
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, ErrorResponse{Error: message})
}
