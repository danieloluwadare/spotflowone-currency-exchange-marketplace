package seed

import (
	"github.com/shopspring/decimal"

	"spotflowone/internal/domain"
)

// Allocations returns the seed data for vendor allocations as specified in the requirements.
// Includes allocations for USD/EUR and USD/GBP pairs from vendors A through E.
// Returns a slice of Allocation structs with pre-configured rates and available amounts.
func Allocations() []domain.Allocation {
	return []domain.Allocation{
		{
			VendorID:      "VENDOR_A",
			BaseCurrency:  "USD",
			QuoteCurrency: "EUR",
			Rate:          decimal.RequireFromString("1.25"),
			Available:     decimal.RequireFromString("5000"),
		},
		{
			VendorID:      "VENDOR_B",
			BaseCurrency:  "USD",
			QuoteCurrency: "EUR",
			Rate:          decimal.RequireFromString("1.22"),
			Available:     decimal.RequireFromString("3000"),
		},
		{
			VendorID:      "VENDOR_C",
			BaseCurrency:  "USD",
			QuoteCurrency: "EUR",
			Rate:          decimal.RequireFromString("1.28"),
			Available:     decimal.RequireFromString("10000"),
		},
		{
			VendorID:      "VENDOR_D",
			BaseCurrency:  "USD",
			QuoteCurrency: "GBP",
			Rate:          decimal.RequireFromString("0.86"),
			Available:     decimal.RequireFromString("8000"),
		},
		{
			VendorID:      "VENDOR_E",
			BaseCurrency:  "USD",
			QuoteCurrency: "GBP",
			Rate:          decimal.RequireFromString("0.84"),
			Available:     decimal.RequireFromString("6000"),
		},
	}
}
