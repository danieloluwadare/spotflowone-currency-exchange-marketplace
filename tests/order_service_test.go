package tests

import (
	"context"
	"testing"

	"github.com/shopspring/decimal"

	"spotflowone/internal/domain"
	"spotflowone/internal/external/rates"
	"spotflowone/internal/repo"
	"spotflowone/internal/service"
)

type fakeRateProvider struct {
	rate domain.MarketRate
	err  error
}

func (f fakeRateProvider) GetRate(ctx context.Context, baseCurrency, quoteCurrency string) (domain.MarketRate, error) {
	return f.rate, f.err
}

func TestBuyOrderBestRateFirst(t *testing.T) {
	allocationRepo := repo.NewMemoryAllocationRepository()
	seedAllocations(t, allocationRepo)

	orderService := service.NewOrderService(allocationRepo, fakeRateProvider{
		rate: domain.MarketRate{Available: true, Source: "LIVE", Rate: decimal.RequireFromString("1.24")},
	})

	result, err := orderService.Execute(context.Background(), domain.Order{
		Direction:     domain.DirectionBuy,
		BaseCurrency:  "USD",
		QuoteCurrency: "EUR",
		Amount:        decimal.RequireFromString("7000"),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Status != domain.StatusFilled {
		t.Fatalf("expected status FILLED, got %s", result.Status)
	}

	if len(result.Fills) != 2 {
		t.Fatalf("expected 2 fills, got %d", len(result.Fills))
	}

	if result.Fills[0].VendorID != "VENDOR_B" || result.Fills[0].Amount.String() != "3000" {
		t.Fatalf("expected first fill to be VENDOR_B 3000, got %+v", result.Fills[0])
	}

	if result.Fills[1].VendorID != "VENDOR_A" || result.Fills[1].Amount.String() != "4000" {
		t.Fatalf("expected second fill to be VENDOR_A 4000, got %+v", result.Fills[1])
	}

	expectedRate := decimal.RequireFromString("8660").Div(decimal.RequireFromString("7000"))
	if result.BlendedRate.Cmp(expectedRate) != 0 {
		t.Fatalf("expected blended rate %s, got %s", expectedRate, result.BlendedRate)
	}
}

func TestSellOrderBestRateFirst(t *testing.T) {
	allocationRepo := repo.NewMemoryAllocationRepository()
	seedAllocations(t, allocationRepo)

	orderService := service.NewOrderService(allocationRepo, rates.NewCachedProvider(nil))

	result, err := orderService.Execute(context.Background(), domain.Order{
		Direction:     domain.DirectionSell,
		BaseCurrency:  "USD",
		QuoteCurrency: "EUR",
		Amount:        decimal.RequireFromString("6000"),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Fills[0].VendorID != "VENDOR_C" {
		t.Fatalf("expected first fill to be VENDOR_C for sell, got %+v", result.Fills[0])
	}
}

func TestPartialFill(t *testing.T) {
	allocationRepo := repo.NewMemoryAllocationRepository()
	seedAllocations(t, allocationRepo)

	orderService := service.NewOrderService(allocationRepo, rates.NewCachedProvider(nil))

	result, err := orderService.Execute(context.Background(), domain.Order{
		Direction:     domain.DirectionBuy,
		BaseCurrency:  "USD",
		QuoteCurrency: "EUR",
		Amount:        decimal.RequireFromString("25000"),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Status != domain.StatusPartiallyFilled {
		t.Fatalf("expected status PARTIALLY_FILLED, got %s", result.Status)
	}

	expectedFilled := decimal.RequireFromString("18000")
	if result.FilledAmount.Cmp(expectedFilled) != 0 {
		t.Fatalf("expected filled amount %s, got %s", expectedFilled, result.FilledAmount)
	}
}

func TestNoLiquidity(t *testing.T) {
	allocationRepo := repo.NewMemoryAllocationRepository()
	orderService := service.NewOrderService(allocationRepo, rates.NewCachedProvider(nil))

	_, err := orderService.Execute(context.Background(), domain.Order{
		Direction:     domain.DirectionBuy,
		BaseCurrency:  "USD",
		QuoteCurrency: "JPY",
		Amount:        decimal.RequireFromString("1000"),
	})
	if err != service.ErrNoLiquidity {
		t.Fatalf("expected ErrNoLiquidity, got %v", err)
	}
}

func seedAllocations(t *testing.T, repo *repo.MemoryAllocationRepository) {
	t.Helper()

	allocations := []domain.Allocation{
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
	}

	for _, alloc := range allocations {
		if err := repo.Add(context.Background(), alloc); err != nil {
			t.Fatalf("failed to seed allocation: %v", err)
		}
	}
}
