package domain

import (
	"strings"

	"github.com/shopspring/decimal"
)

type Direction string

const (
	DirectionBuy  Direction = "BUY"
	DirectionSell Direction = "SELL"
)

type ExecutionStatus string

const (
	StatusFilled          ExecutionStatus = "FILLED"
	StatusPartiallyFilled ExecutionStatus = "PARTIALLY_FILLED"
	StatusNoLiquidity     ExecutionStatus = "NO_LIQUIDITY"
)

type Allocation struct {
	VendorID      string
	BaseCurrency  string
	QuoteCurrency string
	Rate          decimal.Decimal
	Available     decimal.Decimal
}

type Order struct {
	Direction     Direction
	BaseCurrency  string
	QuoteCurrency string
	Amount        decimal.Decimal
}

type Fill struct {
	VendorID string
	Rate     decimal.Decimal
	Amount   decimal.Decimal
}

type MarketRate struct {
	Available bool
	Source    string
	Rate      decimal.Decimal
}

type ExecutionResult struct {
	Status          ExecutionStatus
	RequestedAmount decimal.Decimal
	FilledAmount    decimal.Decimal
	Fills           []Fill
	BlendedRate     decimal.Decimal
	MarketRate      MarketRate
}

// NormalizeCurrency normalizes a currency code by trimming whitespace and converting to uppercase.
// Returns the normalized 3-letter currency code.
func NormalizeCurrency(code string) string {
	return strings.ToUpper(strings.TrimSpace(code))
}

// IsValidCurrency validates that a currency code is exactly 3 uppercase letters.
// Returns true if the code is valid, false otherwise.
func IsValidCurrency(code string) bool {
	if len(code) != 3 {
		return false
	}
	for _, r := range code {
		if r < 'A' || r > 'Z' {
			return false
		}
	}
	return true
}
