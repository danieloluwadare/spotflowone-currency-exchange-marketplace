package rates

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/shopspring/decimal"

	"spotflowone/internal/domain"
)

const defaultFrankfurterBaseURL = "https://api.frankfurter.app"

type FrankfurterClient struct {
	baseURL string
	client  *http.Client
}

// NewFrankfurterClient creates a new Frankfurter API client with default configuration.
// The client uses a 4-second timeout for HTTP requests. Returns a pointer to the initialized client.
func NewFrankfurterClient() *FrankfurterClient {
	return &FrankfurterClient{
		baseURL: defaultFrankfurterBaseURL,
		client: &http.Client{
			Timeout: 4 * time.Second,
		},
	}
}

type frankfurterResponse struct {
	Rates map[string]decimal.Decimal `json:"rates"`
}

// GetRate fetches the current exchange rate from the Frankfurter API for the given currency pair.
// Normalizes currency codes, makes an HTTP GET request, and parses the JSON response.
// Returns a MarketRate with Available=true and Source="LIVE" on success.
// Returns an error if the request fails, the response status is not 2xx, or the rate is not found.
func (c *FrankfurterClient) GetRate(ctx context.Context, baseCurrency, quoteCurrency string) (domain.MarketRate, error) {
	baseCurrency = domain.NormalizeCurrency(baseCurrency)
	quoteCurrency = domain.NormalizeCurrency(quoteCurrency)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/latest?from=%s&to=%s", c.baseURL, baseCurrency, quoteCurrency), nil)
	if err != nil {
		return domain.MarketRate{}, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return domain.MarketRate{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return domain.MarketRate{}, fmt.Errorf("rate provider status: %d", resp.StatusCode)
	}

	var payload frankfurterResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return domain.MarketRate{}, err
	}

	rate, ok := payload.Rates[quoteCurrency]
	if !ok {
		return domain.MarketRate{}, errors.New("rate not found in response")
	}

	return domain.MarketRate{
		Available: true,
		Source:    "LIVE",
		Rate:      rate,
	}, nil
}
