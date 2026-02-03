package service

import (
	"context"
	"sort"

	"github.com/shopspring/decimal"

	"spotflowone/internal/domain"
	"spotflowone/internal/external/rates"
	"spotflowone/internal/repo"
)

type OrderService struct {
	repo         repo.AllocationRepository
	rateProvider rates.Provider
}

// NewOrderService creates a new order service with the provided repository and rate provider.
// Returns a pointer to the initialized service.
func NewOrderService(repo repo.AllocationRepository, rateProvider rates.Provider) *OrderService {
	return &OrderService{
		repo:         repo,
		rateProvider: rateProvider,
	}
}

// Execute processes an order by matching it against available allocations using best-rate-first logic.
// For BUY orders, allocations are sorted by lowest rate first. For SELL orders, highest rate first.
// The order is filled across multiple vendors until the requested amount is met or liquidity is exhausted.
// Returns an ExecutionResult containing fill details, blended rate, and market rate comparison.
// Returns ErrInvalidInput if the order is invalid, or ErrNoLiquidity if no allocations are available.
func (s *OrderService) Execute(ctx context.Context, order domain.Order) (domain.ExecutionResult, error) {
	order.BaseCurrency = domain.NormalizeCurrency(order.BaseCurrency)
	order.QuoteCurrency = domain.NormalizeCurrency(order.QuoteCurrency)

	if order.Direction != domain.DirectionBuy && order.Direction != domain.DirectionSell {
		return domain.ExecutionResult{}, ErrInvalidInput
	}
	if !domain.IsValidCurrency(order.BaseCurrency) || !domain.IsValidCurrency(order.QuoteCurrency) {
		return domain.ExecutionResult{}, ErrInvalidInput
	}
	if order.BaseCurrency == order.QuoteCurrency {
		return domain.ExecutionResult{}, ErrInvalidInput
	}
	if order.Amount.Cmp(decimal.Zero) <= 0 {
		return domain.ExecutionResult{}, ErrInvalidInput
	}

	allocations, err := s.repo.ListByPair(ctx, order.BaseCurrency, order.QuoteCurrency)
	if err != nil {
		return domain.ExecutionResult{}, err
	}
	if len(allocations) == 0 {
		return domain.ExecutionResult{}, ErrNoLiquidity
	}

	sort.Slice(allocations, func(i, j int) bool {
		if order.Direction == domain.DirectionBuy {
			return allocations[i].Rate.Cmp(allocations[j].Rate) < 0
		}
		return allocations[i].Rate.Cmp(allocations[j].Rate) > 0
	})

	remaining := order.Amount
	fills := make([]domain.Fill, 0)
	totalBase := decimal.Zero
	totalQuote := decimal.Zero

	for _, alloc := range allocations {
		if remaining.Cmp(decimal.Zero) <= 0 {
			break
		}
		if alloc.Available.Cmp(decimal.Zero) <= 0 {
			continue
		}

		fillAmount := minDecimal(alloc.Available, remaining)
		if fillAmount.Cmp(decimal.Zero) <= 0 {
			continue
		}

		fills = append(fills, domain.Fill{
			VendorID: alloc.VendorID,
			Rate:     alloc.Rate,
			Amount:   fillAmount,
		})

		totalBase = totalBase.Add(fillAmount)
		totalQuote = totalQuote.Add(fillAmount.Mul(alloc.Rate))
		remaining = remaining.Sub(fillAmount)
	}

	if len(fills) == 0 {
		return domain.ExecutionResult{}, ErrNoLiquidity
	}

	if err := s.repo.ApplyFills(ctx, order.BaseCurrency, order.QuoteCurrency, fills); err != nil {
		return domain.ExecutionResult{}, err
	}

	status := domain.StatusFilled
	if remaining.Cmp(decimal.Zero) > 0 {
		status = domain.StatusPartiallyFilled
	}

	blendedRate := totalQuote.Div(totalBase)

	marketRate := domain.MarketRate{Available: false}
	if s.rateProvider != nil {
		rate, rateErr := s.rateProvider.GetRate(ctx, order.BaseCurrency, order.QuoteCurrency)
		if rateErr == nil {
			marketRate = rate
		}
	}

	return domain.ExecutionResult{
		Status:          status,
		RequestedAmount: order.Amount,
		FilledAmount:    totalBase,
		Fills:           fills,
		BlendedRate:     blendedRate,
		MarketRate:      marketRate,
	}, nil
}

// minDecimal returns the smaller of two decimal values.
// Returns a if a <= b, otherwise returns b.
func minDecimal(a, b decimal.Decimal) decimal.Decimal {
	if a.Cmp(b) <= 0 {
		return a
	}
	return b
}
